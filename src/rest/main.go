package main

import (
    "log"
    "net/http"    
    "github.com/emicklei/go-restful"
    "github.com/emicklei/go-restful/swagger"
)

type CatError struct {
    HttpStatusCode int  `json:"http_status_code"`
    HttpStatus string   `json:"http_status"`
    Message string      `json:"message"`
}

func WriteCatciergeErrorString(response *restful.Response, httpStatus int, message string) {
    if message == "" {
        message = http.StatusText(httpStatus)
    }

    response.AddHeader("Content-Type", "application/json")
    response.WriteEntity(&CatError{HttpStatusCode: httpStatus, HttpStatus: http.StatusText(httpStatus), Message: message})
}

func Returns200(b *restful.RouteBuilder) {
    b.Returns(http.StatusOK, http.StatusText(http.StatusOK), nil)
}

func Returns400(b *restful.RouteBuilder) {
    b.Returns(http.StatusBadRequest,
              http.StatusText(http.StatusBadRequest),
              CatError{})
}

func Returns404(b *restful.RouteBuilder) {
    b.Returns(http.StatusNotFound,
              http.StatusText(http.StatusNotFound),
              CatError{})
}

func Returns500(b *restful.RouteBuilder) {
    b.Returns(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), CatError{})
}

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
