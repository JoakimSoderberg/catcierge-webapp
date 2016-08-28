package main

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	DefaultPageOffset = 0
	DefaultPageLimit  = 10
)

type CatEventTimeV1 struct {
	time.Time
}

func (self *CatEventTimeV1) UnmarshalJSON(b []byte) (err error) {
	s := string(b)

	// Get rid of the quotes "" around the value.
	s = s[1 : len(s)-1]

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

type CatEventHeader struct {
	ID               string `json:"id"`
	EventJSONVersion string `json:"event_json_version"`
	Version          string `json:"version"`
	GitHash          string `json:"git_hash"`
	GitHashShort     string `json:"git_hash_short"`
	GitTainted       int    `json:"git_tainted"`
}

type CatEventHaarMatcherSettingsV1 struct {
	Cascade       string `json:"cascade"`
	EqHistogram   int    `json:"eq_histogram"`
	InDirection   string `json:"in_direction"`
	MinSizeHeight int    `json:"min_size_height"`
	MinSizeWidth  int    `json:"min_size_width"`
	NoMatchIsFail int    `json:"no_match_is_fail"`
	PreyMethod    string `json:"prey_method"`
	PreySteps     int    `json:"prey_steps"`
}

type CatEventSettingsV1 struct {
	HaarMatcher       CatEventHaarMatcherSettingsV1 `json:"haar_matcher"`
	LockoutError      int                           `json:"lockout_error"`
	LockoutErrorDelay float32                       `json:"lockout_error_delay"`
	LockoutMethod     int                           `json:"lockout_method"`
	LockoutTime       int                           `json:"lockout_time"`
	Matcher           string                        `json:"matcher"`
	Matchtime         int                           `json:"matchtime"`
	NoFinalDecision   int                           `json:"no_final_decision"`
	OkMatchesNeeded   int                           `json:"ok_matches_needed"`
}

type CatEventMatchStepV1 struct {
	Active      int    `json:"active"`
	Description string `json:"description"`
	Filename    string `json:"filename"`
	Name        string `json:"name"`
	Path        string `json:"path"`
}

type CatEventMatchV1 struct {
	ID              string                `json:"id"`
	Description     string                `json:"description"`
	Directon        string                `json:"direction"`
	Filename        string                `json:"filename"`
	Path            string                `json:"path"`
	Result          float32               `json:"result"`
	Success         int                   `json:"success"`
	Time            CatEventTimeV1        `json:"time"`
	IsFalsePositive bool                  `json:"is_false_positive"`
	StepCount       int                   `json:"step_count"`
	Steps           []CatEventMatchStepV1 `json:"steps"`
}

// TODO: Break this up into several structs instead.
// http://attilaolah.eu/2014/09/10/json-and-struct-composition-in-go/
type CatEventDataV1 struct {
	CatEventHeader
	State               string             `json:"state"`
	PrevState           string             `json:"prev_state"`
	CatciergeType       string             `json:"catcierge_type"`
	Description         string             `json:"description"`
	Start               CatEventTimeV1     `json:"start"`
	End                 CatEventTimeV1     `json:"end"`
	TimeGenerated       CatEventTimeV1     `json:"time_generated"`
	Timezone            string             `json:"timezone"`
	TimezoneUtcOffset   string             `json:"timezone_utc_offset"`
	Rootpath            string             `json:"rootpath"`
	MatchGroupCount     int                `json:"match_group_count"`
	MatchGroupDirection string             `json:"match_group_direction"`
	MatchGroupMaxCount  int                `json:"match_group_max_count"`
	MatchGroupSuccess   int                `json:"match_group_success"`
	Matches             []CatEventMatchV1  `json:"matches"`
	Settings            CatEventSettingsV1 `json:"settings"`
}

type CatEvent struct {
	ID      bson.ObjectId  `json:"id" bson:"_id"`
	Name    string         `json:"name" bson:"name"`
	Data    CatEventDataV1 `json:"data" bson:"data"`
	Tags    []string       `json:"tags" bson:"tags"`
	Missing bool           `json:"missing" bson:"missing"`
}

type CatEventResource struct {
	// TODO: Replace with MongoDB
	events map[string]CatEvent
}

type CatEventListResponse struct {
	ListResponseHeader
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
		Returns(http.StatusOK, http.StatusText(http.StatusOK), []CatEvent{}).
		Do(AddListResponseParams(ws),
			ReturnsError(http.StatusInternalServerError)))

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

func (ev CatEventResource) listEvents(request *restful.Request, response *restful.Response) {
	var l = CatEventListResponse{}
	l.getListResponseParams(request)

	l.Items = make([]CatEvent, len(ev.events))
	i := 0
	for _, v := range ev.events {
		l.Items[i] = v
		i++
	}

	l.Count = len(l.Items)

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
		log.Printf("Failed to create temp file for unzipping: %s", err)
		WriteCatciergeErrorString(response, http.StatusInternalServerError, "")
		return
	}

	defer os.Remove(tmpfile.Name())

	content, err := ioutil.ReadAll(request.Request.Body)
	if err != nil {
		log.Printf("Failed to read body: %s\n", err)
		WriteCatciergeErrorString(response, http.StatusInternalServerError, "")
		return
	}

	log.Printf("ZIP file size = %+v\n", len(content))

	if _, err := tmpfile.Write(content); err != nil {
		log.Printf("Failed to write to tmpfile %v: %s", tmpfile.Name(), err)
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
