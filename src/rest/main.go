package main

import (
    "fmt"
    "time"
    "log"
    "net/http"
    "archive/zip"
    "os"
    "io"
    "io/ioutil"
    "strings"
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

func (ev CatEventResource) Register(container *restful.Container) {
    ws := new(restful.WebService)

    ws.Path("/events").
        Doc("Manage events").
        Consumes(restful.MIME_XML, restful.MIME_JSON).
        Produces(restful.MIME_JSON, restful.MIME_XML)

    ws.Route(ws.GET("/").To(ev.listEvents).
        Doc("Get all events").
        Returns(http.StatusOK, http.StatusText(http.StatusOK), []CatEvent{}).
        Do(Returns500))

    ws.Route(ws.GET("/{event-id}").To(ev.findEvent).
        Doc("Get an event").
        Param(ws.PathParameter("event-id", "identifier of the event").DataType("string")).
        Returns(http.StatusOK, http.StatusText(http.StatusOK), CatEvent{}).
        Do(Returns404, Returns500).
        Writes(CatEvent{}))

    ws.Route(ws.POST("").To(ev.createEvent).
        Doc("Create an event based on an event ZIP file").
        Do(Returns400, Returns500))

    container.Add(ws)
}

func (ev CatEventResource) listEvents(request *restful.Request, response *restful.Response) {
    response.WriteEntity(ev.events)
}

func (ev CatEventResource) findEvent(request *restful.Request, response *restful.Response) {
    id := request.PathParameter("event-id")
    event, ok := ev.events[id]

    if !ok {
        WriteCatciergeErrorString(response, http.StatusNotFound, fmt.Sprintf("Event '%v' could not be found", id))
        return
    }

    response.WriteEntity(event)
}

// curl --verbose --header "Content-Type: application/zip" --data-binary @file.zip http://awesome
func (ev *CatEventResource) createEvent(request *restful.Request, response *restful.Response) {
    //sr := User{Id: request.PathParameter("event-id")}

    // TODO: Check file size so it is not too big. Maybe as a filter?

    // Save the ZIP on the filesystem temporarily.
    tmpfile, err := ioutil.TempFile("/tmp/", "event")
    if err != nil {
        WriteCatciergeErrorString(response, http.StatusInternalServerError, "")
        return
    }

    defer os.Remove(tmpfile.Name())

    content, err := ioutil.ReadAll(request.Request.Body)
    if err != nil {
        WriteCatciergeErrorString(response, http.StatusInternalServerError, "")
        return
    }

    if _, err := tmpfile.Write(content); err != nil {
        WriteCatciergeErrorString(response, http.StatusInternalServerError, "")
        return
    }
    if err := tmpfile.Close(); err != nil {
        WriteCatciergeErrorString(response, http.StatusInternalServerError, "")
        return
    }

    // Unzip the file to the output directory.
    tmpfileNoExt := strings.TrimSuffix(tmpfile.Name(), filepath.Ext(tmpfile.Name()))
    // TODO: Make this path configurable.
    destDir := filepath.Join("/go/src/app/events/", tmpfileNoExt)

    if err := Unzip(tmpfile.Name(), destDir); err != nil {
        WriteCatciergeErrorString(response, http.StatusInternalServerError, "")
        return
    }

    //response.WriteHeaderAndEntity(http.StatusCreated, usr)
}

// TODO: Move to separate file
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
    wsContainer := restful.NewContainer()

    // TODO: Enable to configure directory where data is unzipped.

    // TODO: Replace with mongodb
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
