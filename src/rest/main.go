package main

import (
    "fmt"
    "time"
    "log"
    "net/http"
//  "strconv"
//  "archive/zip" // TODO: Add unzip support

    "github.com/emicklei/go-restful"
    "github.com/emicklei/go-restful/swagger"
)

type CatEvent struct {
    ID                  string `json:"id"`
    EventJSONVersion    string `json:"event_json_version"`
    CatciergeType       string `json:"catcierge_type"`
    Description         string `json:"description"`
    Start                time.Time `json:"start"`
    End                 string `json:"end"`
    TimeGenerated       time.Time `json:"time_generated"`
    Timezone            string `json:"timezone"`
    TimezoneUtcOffset   string `json:"timezone_utc_offset"`
    GitHash             string `json:"git_hash"`
    GitHashShort        string `json:"git_hash_short"`
    GitTainted          int    `json:"git_tainted"`
    MatchGroupCount     int    `json:"match_group_count"`
    MatchGroupDirection string `json:"match_group_direction"`
    MatchGroupMaxCount  int    `json:"match_group_max_count"`
    MatchGroupSuccess   int    `json:"match_group_success"`
    Matches             []struct {
        Description string `json:"description"`
        Directon    string `json:"direction"`
        Filename    string `json:"filename"`
        ID          string `json:"id"`
        Path        string `json:"path"`
        Result      int    `json:"result"`
        StepCount   int    `json:"step_count"`
        Steps       []struct {
            Active      int    `json:"active"`
            Description string `json:"description"`
            Filename    string `json:"filename"`
            Name        string `json:"name"`
            Path        string `json:"path"`
        } `json:"steps"`
        Success int    `json:"success"`
        Time    time.Time `json:"time"`
    } `json:"matches"`
    Rootpath  string `json:"rootpath"`
    Settings  struct {
        HaarMatcher struct {
            Cascade       string `json:"cascade"`
            EqHistogram   int    `json:"eq_histogram"`
            InDirection   string `json:"in_direction"`
            MinSizeHeight int    `json:"min_size_height"`
            MinSizeWidth  int    `json:"min_size_width"`
            NoMatchIsFail int    `json:"no_match_is_fail"`
            PreyMethod    string `json:"prey_method"`
            PreySteps     int    `json:"prey_steps"`
        } `json:"haar_matcher"`
        LockoutError      int    `json:"lockout_error"`
        LockoutErrorDelay int    `json:"lockout_error_delay"`
        LockoutMethod     int    `json:"lockout_method"`
        LockoutTime       int    `json:"lockout_time"`
        Matcher           string `json:"matcher"`
        Matchtime         int    `json:"matchtime"`
        NoFinalDecision   int    `json:"no_final_decision"`
        OkMatchesNeeded   int    `json:"ok_matches_needed"`
    } `json:"settings"`
    State             string `json:"state"`
    PrevState         string `json:"prev_state"`
    Version           string `json:"version"`
}

/*
type CatEvent struct {
    Id string
    Start_time time.Time
}*/

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
    event := u.events[id]
    if len(event.ID) == 0 {
        response.AddHeader("Content-Type", "text/plain")
        response.WriteErrorString(http.StatusNotFound, "404: Event could not be found.")
        return
    }
    response.WriteEntity(event)
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
        ApiPath:        "/apidocs/swagger.json",

        // Optionally, specifiy where the UI is located
        SwaggerPath:     "/apidocs/",
        SwaggerFilePath: "/usr/local/lib/node_modules/swagger-ui/dist"}

    swagger.RegisterSwaggerService(config, wsContainer)

    log.Printf("start listening on localhost:8080")
    server := &http.Server{Addr: ":8080", Handler: wsContainer}
    log.Fatal(server.ListenAndServe())
}
