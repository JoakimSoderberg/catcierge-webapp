package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"
	"gopkg.in/alecthomas/kingpin.v2"
)

// DefaultPort Default port to run the Webserver on.
const DefaultPort = "8080"

var server *http.Server
var serverScheme string

type catSettings struct {
	server          *http.Server
	serverScheme    string
	port            int
	swaggerUrl      string
	hostSwaggerUi   bool
	swaggerFilePath string
	swaggerPath     string
	swaggerFileName string
	useSSL          bool
	sslCert         string
	sslKey          string
	mongoUrl        string
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
		StringVar(&c.swaggerUrl)

	app.Flag("host-swagger-ui", "(Default) If we should host the Swagger UI API browser. Turn off using --no-host-swagger-ui.").
		Default("true").
		BoolVar(&c.hostSwaggerUi)

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
		StringVar(&c.mongoUrl)

	app.Flag("event-path", "Unzip the uploaded event files in this path.").
		Short('u').
		Default("/go/src/app/events/").
		StringVar(&c.eventPath)

	app.HelpFlag.Short('h')

	return c
}

func main() {
	app := kingpin.New(os.Args[0], "A REST API Server for the Catcierge project.")
	settings := configureFlags(app)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	wsContainer := restful.NewContainer()

	db := DialMongo(settings.mongoUrl)
	defer db.Close()

	cr := NewCatEventResource(db, settings)
	cr.Register(wsContainer)

	// TODO: Add support for getting JSON schemas for everything.
	// TODO: Add heartbeat support, so we can notif if catcierge is down

	// Swagger documentation.
	config := swagger.Config{
		WebServices: wsContainer.RegisteredWebServices(),
		ApiPath:     filepath.Join(settings.swaggerPath, settings.swaggerFileName),
	}

	if settings.hostSwaggerUi {
		log.Printf("Hosting Swagger UI under %v\n", config.ApiPath)
		config.SwaggerPath = settings.swaggerPath
		config.SwaggerFilePath = settings.swaggerFilePath
	} else {
		log.Printf("Using external Swagger UI at %v\n", config.WebServicesUrl)
	}

	swagger.RegisterSwaggerService(config, wsContainer)

	log.Printf("Start listening on port %v", settings.port)
	server = &http.Server{Addr: fmt.Sprintf(":%v", settings.port), Handler: wsContainer}

	if settings.useSSL {
		log.Printf("Using SSL")
		settings.serverScheme = "https"
		log.Fatal(server.ListenAndServeTLS(settings.sslCert, settings.sslKey))
	} else {
		settings.serverScheme = "http"
		log.Fatal(server.ListenAndServe())
	}
}
