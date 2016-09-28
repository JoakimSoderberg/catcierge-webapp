package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"

	restful "github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

// DefaultPort Default port to run the Webserver on.
const DefaultPort = "8080"

// CatSettings represents the command line settings for the app.
type CatSettings struct {
	serverScheme    string
	port            int
	swaggerURL      string
	hostSwaggerUI   bool
	swaggerFilePath string
	swaggerPath     string
	swaggerFileName string
	useSSL          bool
	sslCert         string
	sslKey          string
	mongoURL        string
	eventPath       string
}

var settingsKey key

// AddContext adds the CatSettings to the request context.
func (settings *CatSettings) AddContext(c *context.Context) context.Context {
	return context.WithValue(*c, settingsKey, settings)
}

// FromSettingsContext returns the catSettings in ctx, if any.
func FromSettingsContext(ctx context.Context) (*CatSettings, bool) {
	ev, ok := ctx.Value(eventsKey).(*CatSettings)
	return ev, ok
}

func configureFlags(app *kingpin.Application) *CatSettings {
	c := &CatSettings{}

	app.Flag("port", "Listen port for the web server.").
		Short('p').
		Default(DefaultPort).
		HintOptions("80", "443", DefaultPort).
		IntVar(&c.port)

	app.Flag("swagger-url", "(Optional) URL to an external swagger documentation browser service. Use this if you don't want to self host Swagger UI using --no-host-swagger-ui").
		PlaceHolder("URL").
		StringVar(&c.swaggerURL)

	app.Flag("host-swagger-ui", "(Default) If we should host the Swagger UI API browser. Turn off using --no-host-swagger-ui.").
		Default("true").
		BoolVar(&c.hostSwaggerUI)

	app.Flag("swagger-ui-dist", "Path to Swagger-UI files (http://swagger.io/swagger-ui/). Use 'npm install -g swagger-ui' to install").
		Default("/usr/local/lib/node_modules/swagger-ui/dist").
		StringVar(&c.swaggerFilePath)

	app.Flag("swagger-path", "URL that we should host Swagger-UI under.").
		Default("/apidocs/").
		StringVar(&c.swaggerPath)

	app.Flag("swagger-file", "Name of the swagger JSON file hosted under 'swagger-path'").
		Default("swagger.json").
		StringVar(&c.swaggerFileName)

	app.Flag("ssl", "Run the server in HTTPS").BoolVar(&c.useSSL)
	app.Flag("ssl-cert", "Path to the SSL cert").StringVar(&c.sslCert)
	app.Flag("ssl-key", "Path to the SSL key file").StringVar(&c.sslKey)

	app.Flag("mongo-url", "Url to MongoDB instance. mongodb://host:port").
		Default("mongodb://localhost").
		OverrideDefaultFromEnvar("MONGO_URL").
		StringVar(&c.mongoURL)

	app.Flag("event-path", "Path to where the event data should be stored.").
		Short('u').
		Default("/go/src/app/events/").
		StringVar(&c.eventPath)

	app.HelpFlag.Short('h')

	return c
}

func setupSwagger(container *restful.Container, settings *CatSettings) {
	// Swagger documentation.
	config := swagger.Config{
		WebServices: container.RegisteredWebServices(),
		ApiPath:     filepath.Join(settings.swaggerPath, settings.swaggerFileName),
	}

	if settings.hostSwaggerUI {
		log.Printf("Hosting Swagger UI under %v\n", config.ApiPath)
		config.SwaggerPath = settings.swaggerPath
		config.SwaggerFilePath = settings.swaggerFilePath
	} else {
		log.Printf("Using external Swagger UI at %v\n", config.WebServicesUrl)
	}

	swagger.RegisterSwaggerService(config, container)
}

// DialMongo Dials the MongoDB instance.
func DialMongo(mongoURL string) *mgo.Session {
	log.Printf("Attempting to dial %s", mongoURL)

	session, err := mgo.Dial(mongoURL)
	if err != nil {
		log.Printf("Failed to dial MongoDB: %s", mongoURL)
		panic(err)
	}

	return session
}

// WrapContexts Wraps the given Handler and injects a context into each request.
func WrapContexts(handler http.Handler, resources []CatciergeContextAdder) http.Handler {
	c := context.Background()

	for _, r := range resources {
		c = r.AddContext(&c)
	}

	return GetWrappedContextHTTPHandler(handler, c)
}

// GetAuthenticationStateFromToken gets an authentication state based on a given token. We can later
// use this state in the handlers to verify if we're authenticated or not, and which user is logged in if
// that is the case.
func GetAuthenticationStateFromToken(req *http.Request, tokenStr string) (*AuthenticationState, error) {
	users, ok := FromUsersContext(req.Context())
	if !ok {
		return nil, errors.New("Failed to get users resource from context in basic authentication")
	}

	var token AccessToken
	var authState AuthenticationState

	// If the access token is found in the database, get the logged in user.
	err := users.session.DB("catcierge").C("tokens").Find(bson.M{"token": tokenStr}).One(&token)
	if err != nil {
		return &authState, fmt.Errorf("No such token '%s'", tokenStr)
	}

	// TODO: Include this in the above query instead by doing magic with the AuthenticationState type
	err = users.session.DB("catcierge").C("users").FindId(token.UserID).One(&authState.User)
	if err != nil {
		return &authState, fmt.Errorf("Invalid token '%s'", tokenStr)
	}

	authState.IsAuthenticated = true
	return &authState, nil
}

func basicTokenAuthenticate(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	var i int
	var authState *AuthenticationState
	var err error

	rawTokenStr := req.Request.Header.Get("Authorization")
	if rawTokenStr == "" {
		goto skip
	}

	i = strings.Index(rawTokenStr, "token ")
	if i == -1 {
		goto skip
	}

	authState, err = GetAuthenticationStateFromToken(req.Request, rawTokenStr[i:])
	if authState == nil {
		if err != nil {
			log.Printf("Failed to inject AuthenticationState into request: %s", err)
		}
		WriteCatciergeErrorString(resp, http.StatusInternalServerError, "")
		return
	}

skip:
	// Add the Authentication state to the HTTP request context.
	ctx := req.Request.Context()
	ctx = authState.AddContext(&ctx)
	req.Request = req.Request.WithContext(ctx)

	chain.ProcessFilter(req, resp)
}

func main() {

	// TODO: Move all of this into a separate package that simply takes a config.
	// TODO: Read config file.

	// Parse command line flags.
	app := kingpin.New(os.Args[0], "A REST API Server for the Catcierge project.")
	settings := configureFlags(app)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	// Connect to MongoDB.
	db := DialMongo(settings.mongoURL)
	defer db.Close()

	// Setup Go-restful and create the REST resources.
	wsContainer := restful.NewContainer()

	// TODO: Add more filters, such as session authentication.
	wsContainer.Filter(basicTokenAuthenticate)

	events := NewEventsResource(db, settings)
	events.Register(wsContainer)

	accounts := NewAccountResource(db, settings)
	accounts.Register(wsContainer)

	users := NewUserResource(db, settings)
	users.Register(wsContainer)

	tokens := NewAccessTokensResource(db, settings)
	tokens.Register(wsContainer)

	// TODO: Add support for getting JSON schemas for everything.
	// TODO: Add heartbeat support, so we can notify if catcierge is down
	setupSwagger(wsContainer, settings)

	// Accept and respond in JSON unless told otherwise.
	restful.DefaultRequestContentType(restful.MIME_JSON)
	restful.DefaultResponseContentType(restful.MIME_JSON)

	// Faster router.
	restful.DefaultContainer.Router(restful.CurlyRouter{})
	// No need to access body more than once.
	restful.SetCacheReadEntity(false)

	log.Printf("Start listening on port %v", settings.port)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%v", settings.port),
		Handler: WrapContexts(wsContainer, []CatciergeContextAdder{events, accounts, users, settings, tokens})}

	// Handle interrupts.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for _ = range c {
			// sig is a ^C, handle it
			// TODO: Do any cleanup on interrupt.
			log.Fatal("Teardown finished, forcing exit\n")
		}
	}()

	if settings.useSSL {
		log.Printf("Using SSL")
		// TODO: Add redirect handler for HTTP to HTTPS (or simply a 404 with "Use HTTPS")
		settings.serverScheme = "https"
		log.Fatal(server.ListenAndServeTLS(settings.sslCert, settings.sslKey))
	} else {
		settings.serverScheme = "http"
		log.Fatal(server.ListenAndServe())
	}
}
