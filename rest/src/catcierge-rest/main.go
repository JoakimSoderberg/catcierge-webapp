package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"labix.org/v2/mgo"

	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"
	"gopkg.in/alecthomas/kingpin.v2"
)

// DefaultPort Default port to run the Webserver on.
const DefaultPort = "8080"

type catSettings struct {
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

func configureFlags(app *kingpin.Application) *catSettings {
	c := &catSettings{}

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

	app.Flag("event-path", "Unzip the uploaded event files in this path.").
		Short('u').
		Default("/go/src/app/events/").
		StringVar(&c.eventPath)

	app.HelpFlag.Short('h')

	return c
}

func setupSwagger(container *restful.Container, settings *catSettings) {
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

// WrapContext Wraps the given Handler and injects a context into each request.
func WrapContext(handler http.Handler, ev *CatEventResource) http.Handler {
	// Create a new context and inject our catevent resource into it.
	ctx := NewContext(context.Background(), ev)
	wrapped := func(w http.ResponseWriter, req *http.Request) {
		handler.ServeHTTP(w, req.WithContext(ctx))
	}
	return http.HandlerFunc(wrapped)
}

func main() {

	// Parse command line flags.
	app := kingpin.New(os.Args[0], "A REST API Server for the Catcierge project.")
	settings := configureFlags(app)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	// Connect to MongoDB.
	db := DialMongo(settings.mongoURL)
	defer db.Close()

	// Setup Go-restful and create the REST resources.
	wsContainer := restful.NewContainer()
	ev := NewCatEventResource(db, settings)
	ev.Register(wsContainer)

	// TODO: Add support for getting JSON schemas for everything.
	// TODO: Add heartbeat support, so we can notify if catcierge is down
	setupSwagger(wsContainer, settings)

	log.Printf("Start listening on port %v", settings.port)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%v", settings.port),
		Handler: WrapContext(wsContainer, ev)}

	if settings.useSSL {
		log.Printf("Using SSL")
		settings.serverScheme = "https"
		log.Fatal(server.ListenAndServeTLS(settings.sslCert, settings.sslKey))
	} else {
		settings.serverScheme = "http"
		log.Fatal(server.ListenAndServe())
	}
}
