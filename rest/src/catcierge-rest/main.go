package main

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"
	"gopkg.in/alecthomas/kingpin.v2"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var (
	DefaultPort = "8080"
	app         = kingpin.New(os.Args[0], "A REST API Server for the Catcierge project.")
	port        = app.Flag("port", "Listen port for the web server.").
			Short('p').
			Default(DefaultPort).
			HintOptions("80", "443", DefaultPort).
			Int()
	swaggerUrl = app.Flag("swagger-url", "(Optional) URL to an external swagger documentation browser service. Use this if you don't want to self host Swagger UI using --no-host-swagger-ui").
			PlaceHolder("URL").
			String()
	hostSwaggerUi = app.Flag("host-swagger-ui", "(Default) If we should host the Swagger UI API browser. Turn off using --no-host-swagger-ui.").
			Default("true").
			Bool()
	swaggerFilePath = app.Flag("swagger-ui-dist", "Path to Swagger-UI files (http://swagger.io/swagger-ui/). Use 'npm install -g swagger-ui' to install").
			Default("/usr/local/lib/node_modules/swagger-ui/dist").
			String()
	swaggerPath = app.Flag("swagger-path", "URL that we should host Swagger-UI under.").
			Default("/apidocs/").
			String()
	swaggerFileName = app.Flag("swagger-file", "Name of the swagger JSON file hosted under 'swagger-path'").
			Default("swagger.json").
			String()
	useSSL   = app.Flag("ssl", "Run the server in HTTPS").Bool()
	sslCert  = app.Flag("ssl-cert", "Path to the SSL cert").String()
	sslKey   = app.Flag("ssl-key", "Path to the SSL key file").String()
	mongoUrl = app.Flag("mongo-url", "Url to MongoDB instance. mongodb://host:port").Default("mongodb://localhost").OverrideDefaultFromEnvar("MONGO_URL").String()
)

var server *http.Server
var serverScheme string

func main() {

	app.HelpFlag.Short('h')
	kingpin.MustParse(app.Parse(os.Args[1:]))

	wsContainer := restful.NewContainer()

	cr := NewCatEventResource(DialMongo(*mongoUrl))
	cr.Register(wsContainer)

	// TODO: Add support for getting JSON schemas for everything.
	// TODO: Add heartbeat support, so we can notif if catcierge is down

	// Swagger documentation.
	config := swagger.Config{
		WebServices: wsContainer.RegisteredWebServices(),
		//WebServicesUrl: *swaggerUrl,
		ApiPath: filepath.Join(*swaggerPath, *swaggerFileName),
	}

	if *hostSwaggerUi {
		log.Printf("Hosting Swagger UI under %v\n", config.ApiPath)
		config.SwaggerPath = *swaggerPath
		config.SwaggerFilePath = *swaggerFilePath
	} else {
		log.Printf("Using external Swagger UI at %v\n", config.WebServicesUrl)
	}

	swagger.RegisterSwaggerService(config, wsContainer)

	log.Printf("Start listening on port %v", *port)
	server = &http.Server{Addr: fmt.Sprintf(":%v", *port), Handler: wsContainer}

	if *useSSL {
		log.Printf("Using SSL")
		serverScheme = "https"
		log.Fatal(server.ListenAndServeTLS(*sslCert, *sslKey))
	} else {
		serverScheme = "http"
		log.Fatal(server.ListenAndServe())
	}
}
