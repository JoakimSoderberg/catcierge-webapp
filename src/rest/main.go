package main

import (
    "fmt"
    "time"
    "log"
    "net/http"
//  "strconv"

    "github.com/emicklei/go-restful"
    "github.com/emicklei/go-restful/swagger"
)

type CatEvent struct {
    Id string
    Start_time time.Time
}

type CatEventResource struct {
    // TODO: Replace with MongoDB
    events map[string]CatEvent
}

func (u CatEventResource) Register(container *restful.Container) {
    ws := new(restful.WebService)

    ws.Path("/events")
      .Doc("Manage events")
      .Consumes(restful.MIME_XML, restful.MIME_JSON)
      .Produces()
}

func main() {
    fmt.Println("hello worldsss");

    wsContainer := restful.NewContainer()

    cr := CatEventResource {map[string]CatEvent{}}

    cr.Register(wsContainer)

    // Swagger documentation.
    // TODO: Set ports and all via command line options
    config := swagger.Config {
        WebServices:    wsContainer.RegisteredWebServices(), // you control what services are visible
        WebServicesUrl: "http://localhost:8080",
        ApiPath:        "/apidocs.json",

        // Optionally, specifiy where the UI is located
        SwaggerPath:     "/apidocs/",
        SwaggerFilePath: "/swagger-ui-dist"}

    swagger.RegisterSwaggerService(config, wsContainer)

    log.Printf("start listening on localhost:8080")
    server := &http.Server{Addr: ":8080", Handler: wsContainer}
    log.Fatal(server.ListenAndServe())
}
