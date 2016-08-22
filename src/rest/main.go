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
	DEFAULT_PORT = "8080"
	app          = kingpin.New(os.Args[0], "A REST API Server for the Catcierge project.")
	port         = app.Flag("port", "Listen port for the web server.").
			Short('p').
			Default(DEFAULT_PORT).
			HintOptions("80", "443", DEFAULT_PORT).
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
)

func main() {

	app.HelpFlag.Short('h')
	app.Parse(os.Args[1:])

	wsContainer := restful.NewContainer()

	// TODO: Replace with mongodb
	cr := CatEventResource{events: map[string]CatEvent{}}

	cr.Register(wsContainer)

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
	server := &http.Server{Addr: fmt.Sprintf(":%v", *port), Handler: wsContainer}
	log.Fatal(server.ListenAndServe())
}
