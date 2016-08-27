package main

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
	"log"
	"net/http"
	"os"
//	"path/filepath"
//	"strings"
	"time"
    "strconv"
)

type CatEventV1Time struct {
    time.Time
}

func (self *CatEventV1Time) UnmarshalJSON(b []byte) (err error) {
    s := string(b)

    // Get rid of the quotes "" around the value.
    s = s[1:len(s)-1]

    t, err := time.Parse(time.RFC3339Nano, s)

    // The first version of the catcierge event JSON uses
    // the wrong format for the timezone without ':' so if we
    // fail to parse we attempt again
    if err != nil {
        t, err = time.Parse("2006-01-02T15:04:05.999999999Z0700", s)
        if err != nil {
            // This will be parsed as UTC which is incorrect...
            // But some dates are in this format.
            t, err = time.Parse("2006-01-02 15:04:05", s)
            return err
        }
    }
    self.Time = t
    return
}

// TODO: Break this up into several structs instead.
// http://attilaolah.eu/2014/09/10/json-and-struct-composition-in-go/
type CatEventDataV1 struct {
	ID                  string    `json:"id"`
	EventJSONVersion    string    `json:"event_json_version"`
	CatciergeType       string    `json:"catcierge_type"`
	Description         string    `json:"description"`
	Start               CatEventV1Time `json:"start"`
	End                 string    `json:"end"`
	TimeGenerated       CatEventV1Time `json:"time_generated"`
	Timezone            string    `json:"timezone"`
	TimezoneUtcOffset   string    `json:"timezone_utc_offset"`
	GitHash             string    `json:"git_hash"`
	GitHashShort        string    `json:"git_hash_short"`
	GitTainted          int       `json:"git_tainted"`
	MatchGroupCount     int       `json:"match_group_count"`
	MatchGroupDirection string    `json:"match_group_direction"`
	MatchGroupMaxCount  int       `json:"match_group_max_count"`
	MatchGroupSuccess   int       `json:"match_group_success"`
	Rootpath            string    `json:"rootpath"`
	State               string    `json:"state"`
	PrevState           string    `json:"prev_state"`
	Version             string    `json:"version"`
	Matches             []struct {
		ID              string    `json:"id"`
		Description     string    `json:"description"`
		Directon        string    `json:"direction"`
		Filename        string    `json:"filename"`
		Path            string    `json:"path"`
		Result          float32   `json:"result"`
		Success         int       `json:"success"`
		Time            CatEventV1Time `json:"time"`
		IsFalsePositive bool      `json:"is_false_positive"`
		StepCount       int       `json:"step_count"`
		Steps           []struct {
			Active      int    `json:"active"`
			Description string `json:"description"`
			Filename    string `json:"filename"`
			Name        string `json:"name"`
			Path        string `json:"path"`
		} `json:"steps"`
	} `json:"matches"`
	Settings struct {
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
		LockoutErrorDelay float32 `json:"lockout_error_delay"`
		LockoutMethod     int    `json:"lockout_method"`
		LockoutTime       int    `json:"lockout_time"`
		Matcher           string `json:"matcher"`
		Matchtime         int    `json:"matchtime"`
		NoFinalDecision   int    `json:"no_final_decision"`
		OkMatchesNeeded   int    `json:"ok_matches_needed"`
	} `json:"settings"`
}

type CatEvent struct {
	ID   bson.ObjectId `json:"id" bson:"_id"`
	Name string        `json:"name"`
	Data CatEventDataV1  `json:"data" bson:"data"`
	Tags []string      `json:"tags" bson:"tags"`
    Missing bool       `json:"missing" bson:"missing"`
}

type CatEventResource struct {
	// TODO: Replace with MongoDB
	events map[string]CatEvent
}

type CatEventListResponse struct {
    // Make these part of a more general struct that can be reused for all lists somehow
    Count int `json:"count"`
    Offset int `json:"offset"`
    Limit int `json:"limit"`

    Items []CatEvent `json:"items"`
}
// TODO: Add a new resource for listing all images.

func (ev CatEventResource) Register(container *restful.Container) {
	ws := new(restful.WebService)

	ws.Path("/events").
		Doc("Manage events").
		Consumes(restful.MIME_XML, restful.MIME_JSON, "application/zip").
		Produces(restful.MIME_JSON, restful.MIME_XML)

    // TODO: Add pagination and stuff. Items should be shown under an "items" field.
    // TODO: Make constants for default values here
	ws.Route(ws.GET("/").To(ev.listEvents).
		Doc("Get all events").
        Param(ws.QueryParameter("offset", "Offset into the list").DataType("int").DefaultValue("0")).
        Param(ws.QueryParameter("limit", "Number of items to return").DataType("int").DefaultValue("10")).
		Returns(http.StatusOK, http.StatusText(http.StatusOK), []CatEvent{}).
		Do(ReturnsError(http.StatusInternalServerError)))

	ws.Route(ws.GET("/{event-id}").To(ev.getEvent).
		Doc("Get an event").
		Param(ws.PathParameter("event-id", "identifier of the event").DataType("string")).
		Do(ReturnsStatus(http.StatusOK, "", CatEvent{}),
			ReturnsError(http.StatusNotFound),
			ReturnsError(http.StatusInternalServerError)).
		Writes(CatEvent{}))

	ws.Route(ws.POST("").To(ev.createEvent).
		Doc("Create an event based on an event ZIP file").
		Do(ReturnsError(http.StatusBadRequest),
			ReturnsError(http.StatusInternalServerError)))

	container.Add(ws)
}

// TODO: Make this a more generic list function that can be reused.
func (ev CatEventResource) listEvents(request *restful.Request, response *restful.Response) {
    var l = CatEventListResponse{}
    offset, _ := strconv.Atoi(request.QueryParameter("offset"))
    limit, err := strconv.Atoi(request.QueryParameter("limit"))
    if err != nil {
        limit = 10
    }

    l.Offset = offset
    l.Limit = limit
    l.Items = make([]CatEvent, len(ev.events))

    i := 0
    for _, v := range ev.events {
        l.Items[i] = v
        i++
    }

    response.WriteEntity(l)
}

func (ev CatEventResource) getEvent(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("event-id")
	event, ok := ev.events[id]

	if !ok {
		WriteCatciergeErrorString(response, http.StatusNotFound, fmt.Sprintf("Event '%v' could not be found", id))
		return
	}

	response.WriteEntity(event)
}

var (
	unzipPath = app.Flag("unzip-path", "Unzip the uploaded event files in this path.").
		Short('u').
		Default("/go/src/app/events/").
		String()
)

// curl --verbose --header "Content-Type: application/zip" --data-binary @file.zip http://awesome
func (ev *CatEventResource) createEvent(request *restful.Request, response *restful.Response) {

	// TODO: Check file size so it is not too big. Maybe as a filter?

	log.Printf("Received file\n")

	// Save the ZIP on the filesystem temporarily.
	tmpfile, err := ioutil.TempFile("/tmp/", "event")
	if err != nil {
		log.Printf("Failed to get temp file name")
		WriteCatciergeErrorString(response, http.StatusInternalServerError, "")
		return
	}

	defer os.Remove(tmpfile.Name())

	content, err := ioutil.ReadAll(request.Request.Body)
	if err != nil {
		log.Printf("Failed to read body\n")
		WriteCatciergeErrorString(response, http.StatusInternalServerError, "")
		return
	}

    log.Printf("ZIP file size = %+v\n", len(content))

	if _, err := tmpfile.Write(content); err != nil {
		log.Printf("Failed to write to tmpfile %v", tmpfile.Name())
		WriteCatciergeErrorString(response, http.StatusInternalServerError, "")
		return
	}
	if err := tmpfile.Close(); err != nil {
		WriteCatciergeErrorString(response, http.StatusInternalServerError, "")
		return
	}

	// Unzip the file to the output directory.
	eventData, err := UnzipEvent(tmpfile.Name(), *unzipPath)
    if err != nil {
		log.Printf("Failed to unzip file %v to %v: %s", tmpfile.Name(), *unzipPath, err)
		WriteCatciergeErrorString(response, http.StatusInternalServerError, "")
		return
	}

    // TODO: Replace with MongoDB
    ev.events[eventData.ID] = CatEvent{ID: bson.ObjectIdHex(eventData.ID[0:24]), Data: *eventData}

	response.WriteHeader(http.StatusCreated)

	log.Printf("Successfully unpacked event\n")
}
