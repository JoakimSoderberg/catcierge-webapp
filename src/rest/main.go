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

    ws.Path("/events").
        Doc("Manage events").
        Consumes(restful.MIME_XML, restful.MIME_JSON).
        Produces(restful.MIME_JSON, restful.MIME_XML)

    ws.Route(ws.GET("/{event-id}").To(u.findEvent).
        Doc("Get an event").
        Operation("findEvent").
        Param(ws.PathParameter("event-id", "identifier of the event").DataType("string")).
        Writes(CatEvent{}))

    container.Add(ws)
}

func (u CatEventResource) findEvent(request *restful.Request, response *restful.Response) {
    id := request.PathParameter("event-id")
    usr := u.events[id]
    if len(usr.Id) == 0 {
        response.AddHeader("Content-Type", "text/plain")
        response.WriteErrorString(http.StatusNotFound, "404: Event could not be found.")
        return
    }
    response.WriteEntity(usr)
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
        SwaggerFilePath: "/usr/local/lib/node_modules/swagger-ui/dist"}

    swagger.RegisterSwaggerService(config, wsContainer)

    log.Printf("start listening on localhost:8080")
    server := &http.Server{Addr: ":8080", Handler: wsContainer}
    log.Fatal(server.ListenAndServe())
}
