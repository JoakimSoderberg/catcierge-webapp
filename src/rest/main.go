package main

import (
    "log"
    "net/http"    
    "github.com/emicklei/go-restful"
    "github.com/emicklei/go-restful/swagger"
)

func main() {
    wsContainer := restful.NewContainer()

    // TODO: Enable to configure directory where data is unzipped.

    // TODO: Replace with mongodb
    cr := CatEventResource {events: map[string]CatEvent{}}

    cr.Register(wsContainer)

    // Swagger documentation.
    // TODO: Set ports and all via command line options
    config := swagger.Config {
        WebServices:    wsContainer.RegisteredWebServices(), // you control what services are visible
        WebServicesUrl: "http://192.168.99.100:8080",
        ApiPath:        "/apidocs/swagger.json",

        // Optionally, specifiy where the UI is located
        SwaggerPath:     "/apidocs/",
        SwaggerFilePath: "/usr/local/lib/node_modules/swagger-ui/dist"}

    swagger.RegisterSwaggerService(config, wsContainer)

    log.Printf("start listening on port 8080")
    server := &http.Server{Addr: ":8080", Handler: wsContainer}
    log.Fatal(server.ListenAndServe())
}
