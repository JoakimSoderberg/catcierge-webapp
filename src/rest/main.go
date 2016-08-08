package main

import (
    "fmt"
    "time"
    "log"
    "net/http"
//  "strconv"
    "archive/zip" // TODO: Add unzip support
    "os"
    "io"
    "path/filepath"
    "github.com/emicklei/go-restful"
    "github.com/emicklei/go-restful/swagger"
    "labix.org/v2/mgo/bson"
)

type CatEventData struct {
    ID                  string `json:"id"`
    EventJSONVersion    string `json:"event_json_version"`
    CatciergeType       string `json:"catcierge_type"`
    Description         string `json:"description"`
    Start               time.Time `json:"start"`
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
    Rootpath            string `json:"rootpath"`
    State               string `json:"state"`
    PrevState           string `json:"prev_state"`
    Version             string `json:"version"`
    Matches             []struct {
        ID          string `json:"id"`
        Description string `json:"description"`
        Directon    string `json:"direction"`
        Filename    string `json:"filename"`
        Path        string `json:"path"`
        Result      int    `json:"result"`
        Success     int    `json:"success"`
        Time        time.Time `json:"time"`
        IsFalsePositive bool `json:"is_false_positive"`
        StepCount   int    `json:"step_count"`
        Steps       []struct {
            Active      int    `json:"active"`
            Description string `json:"description"`
            Filename    string `json:"filename"`
            Name        string `json:"name"`
            Path        string `json:"path"`
        } `json:"steps"`
    } `json:"matches"`
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
}

type CatEvent struct {
    ID bson.ObjectId    `bson:"_id"`
    Name string
    Data CatEventData   `bson:"data"`
    Tags []string       `bson:"tags"`
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
    event := u.events[id]

    if len(event.ID) == 0 {
        response.AddHeader("Content-Type", "text/plain")
        response.WriteErrorString(http.StatusNotFound, "404: Event could not be found.")
        return
    }
    response.WriteEntity(event)
}

func Unzip(src, dest string) error {
    r, err := zip.OpenReader(src)
    if err != nil {
        return err
    }
    defer func() {
        if err := r.Close(); err != nil {
            panic(err)
        }
    }()

    os.MkdirAll(dest, 0755)

    // Closure to address file descriptors issue with all the deferred .Close() methods
    extractAndWriteFile := func(f *zip.File) error {
        rc, err := f.Open()
        if err != nil {
            return err
        }
        defer func() {
            if err := rc.Close(); err != nil {
                panic(err)
            }
        }()

        path := filepath.Join(dest, f.Name)

        if f.FileInfo().IsDir() {
            os.MkdirAll(path, f.Mode())
        } else {
            f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
            if err != nil {
                return err
            }
            defer func() {
                if err := f.Close(); err != nil {
                    panic(err)
                }
            }()

            _, err = io.Copy(f, rc)
            if err != nil {
                return err
            }
        }
        return nil
    }

    for _, f := range r.File {
        err := extractAndWriteFile(f)
        if err != nil {
            return err
        }
    }

    return nil
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
