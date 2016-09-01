package main

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
	"path"
	"strings"
	"labix.org/v2/mgo"
)

const (
	DefaultPageOffset = 0
	DefaultPageLimit  = 10
)

var (
	eventPath = app.Flag("event-path", "Unzip the uploaded event files in this path.").
		Short('u').
		Default("/go/src/app/events/").
		String()
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
	ID               string `json:"id" bson:"id"`
	EventJSONVersion string `json:"event_json_version" bson:"event_json_version"`
	Version          string `json:"version" bson:"version"`
	GitHash          string `json:"git_hash" bson:"git_hash"`
	GitHashShort     string `json:"git_hash_short" bson:"git_hash_short"`
	GitTainted       int    `json:"git_tainted" bson:"git_tainted"`
}

type CatEventHaarMatcherSettingsV1 struct {
	Cascade       string `json:"cascade" bson:"cascade"`
	EqHistogram   int    `json:"eq_histogram" bson:"eq_histogram"`
	InDirection   string `json:"in_direction" bson:"in_direction"`
	MinSizeHeight int    `json:"min_size_height" bson:"min_size_height"`
	MinSizeWidth  int    `json:"min_size_width" bson:"min_size_width"`
	NoMatchIsFail int    `json:"no_match_is_fail" bson:"no_match_is_fail"`
	PreyMethod    string `json:"prey_method" bson:"prey_method"`
	PreySteps     int    `json:"prey_steps" bson:"prey_steps"`
}

type CatEventSettingsV1 struct {
	HaarMatcher       CatEventHaarMatcherSettingsV1 `json:"haar_matcher" bson:"haar_matcher"`
	LockoutError      int                           `json:"lockout_error" bson:"lockout_error"`
	LockoutErrorDelay float32                       `json:"lockout_error_delay" bson:"lockout_error_delay"`
	LockoutMethod     int                           `json:"lockout_method" bson:"lockout_method"`
	LockoutTime       int                           `json:"lockout_time" bson:"lockout_time"`
	Matcher           string                        `json:"matcher" bson:"matcher"`
	Matchtime         int                           `json:"matchtime" bson:"matchtime"`
	NoFinalDecision   int                           `json:"no_final_decision" bson:"no_final_decision"`
	OkMatchesNeeded   int                           `json:"ok_matches_needed" bson:"ok_matches_needed"`
}

type CatEventMatchStepV1 struct {
	Active      int    `json:"active" bson:"active"`
	Description string `json:"description" bson:"description"`
	Filename    string `json:"filename" bson:"filename"`
	Name        string `json:"name" bson:"name"`
	Path        string `json:"path" bson:"path"`
	Ref 		string `json:"ref",omitempty`
}

type CatEventMatchV1 struct {
	ID              string                `json:"id" bson:"id"`
	Description     string                `json:"description" bson:"description"`
	Directon        string                `json:"direction" bson:"direction"`
	Filename        string                `json:"filename" bson:"filename"`
	Path            string                `json:"path" bson:"path"`
	Ref 			string                `json:"ref",omitempty`
	Result          float32               `json:"result" bson:"result"`
	Success         int                   `json:"success" bson:"success"`
	Time            CatEventTimeV1        `json:"time" bson:"time"`
	IsFalsePositive bool                  `json:"is_false_positive" bson:"is_false_positive"`
	StepCount       int                   `json:"step_count" bson:"step_count"`
	Steps           []CatEventMatchStepV1 `json:"steps" bson:"steps"`
}

type CatEventDataV1 struct {
	CatEventHeader
	State               string             `json:"state" bson:"state"`
	PrevState           string             `json:"prev_state" bson:"prev_state"`
	CatciergeType       string             `json:"catcierge_type" bson:"catcierge_type"`
	Description         string             `json:"description" bson:"description"`
	Start               CatEventTimeV1     `json:"start" bson:"start"`
	End                 CatEventTimeV1     `json:"end" bson:"end"`
	TimeGenerated       CatEventTimeV1     `json:"time_generated" bson:"time_generated"`
	Timezone            string             `json:"timezone" bson:"timezone"`
	TimezoneUtcOffset   string             `json:"timezone_utc_offset" bson:"timezone_utc_offset"`
	Rootpath            string             `json:"rootpath" bson:"rootpath"`
	MatchGroupCount     int                `json:"match_group_count" bson:"match_group_count"`
	MatchGroupDirection string             `json:"match_group_direction" bson:"match_group_direction"`
	MatchGroupMaxCount  int                `json:"match_group_max_count" bson:"match_group_max_count"`
	MatchGroupSuccess   int                `json:"match_group_success" bson:"match_group_success"`
	Matches             []CatEventMatchV1  `json:"matches" bson:"matches"`
	Settings            CatEventSettingsV1 `json:"settings" bson:"settings"`
}

type CatEvent struct {
	ID      bson.ObjectId  `json:"id" bson:"_id"`
	Name    string         `json:"name" bson:"name"`
	Data    CatEventDataV1 `json:"data" bson:"data"`
	Tags    []string       `json:"tags" bson:"tags"`
	Missing bool           `json:"missing" bson:"missing"`
}

type CatEventResource struct {
	// MongoDB session.
	session *mgo.Session
}

type CatEventListResponse struct {
	ListResponseHeader
	Items []CatEvent `json:"items"`
}

func DialMongo(mongoUrl string) *mgo.Session {
	log.Printf("Attempting to dial %s", mongoUrl)

	session, err := mgo.Dial(mongoUrl)
	if err != nil {
		log.Printf("Failed to dial MongoDB")
		panic(err)
	}

	return session
}

func NewCatEventResource(session *mgo.Session) *CatEventResource {
	return &CatEventResource{session: session}
}

// TODO: Add a new resource for listing all images.

func (ev CatEventResource) Register(container *restful.Container) {
	ws := new(restful.WebService)

	ws.Path("/events").
		Doc("Manage events").
		Consumes("application/zip").
		Produces(restful.MIME_JSON, restful.MIME_XML)

	ws.Route(ws.GET("/").To(ev.listEvents).
		Doc("Get all events").
		Returns(http.StatusOK, http.StatusText(http.StatusOK), []CatEvent{}).
		Do(AddListResponseParams(ws),
			ReturnsError(http.StatusInternalServerError)).
		Writes(CatEventListResponse{}))

	ws.Route(ws.GET("/{event-id}").To(ev.getEvent).
		Doc("Get an event").
		Param(ws.PathParameter("event-id", "identifier of the event").DataType("string")).
		Do(ReturnsStatus(http.StatusOK, "", CatEvent{}),
			ReturnsError(http.StatusNotFound),
			ReturnsError(http.StatusInternalServerError)).
		Writes(CatEvent{}))

	// Static images.
	ws.Route(ws.GET("/{event-id}/{subpath:*}").To(eventStaticFiles).
		Doc("Get static files for an event such as images").
		Param(ws.PathParameter("event-id", "identifier of the event").DataType("string")).
		Do(ReturnsStatus(http.StatusOK, "", CatEvent{}),
			ReturnsError(http.StatusNotFound),
			ReturnsError(http.StatusInternalServerError)))

	ws.Route(ws.POST("").To(ev.createEvent).
		Doc("Create an event based on an event ZIP file").
		Do(ReturnsError(http.StatusBadRequest),
			ReturnsError(http.StatusInternalServerError)))

	container.Add(ws)
}

func eventStaticFiles(req *restful.Request, resp *restful.Response) {
	fullPath := path.Join(*eventPath, req.PathParameter("event-id"), req.PathParameter("subpath"))
	log.Printf("GET %s", fullPath)
	http.ServeFile(resp.ResponseWriter, req.Request, fullPath)
}

func ReverseUrl(request *http.Request, fullPath string) string {
	revUrl := url.URL{Host: request.Host, Path: strings.Trim(fullPath, "/"), Scheme: serverScheme}
	return revUrl.String()
}

func (ev CatEventResource) listEvents(request *restful.Request, response *restful.Response) {
	var l = CatEventListResponse{}
	l.getListResponseParams(request)

	/*
	// TODO: Replace with MongoDB query
	l.Items = make([]CatEvent, len(ev.events))

	i := 0
	for k := range ev.events {
		v := ev.events[k]
		c := &v.Data
		for mi := range c.Matches {
			// TODO: Make generic function
			m := &c.Matches[mi]
			m.Ref = ReverseUrl(request.Request, path.Join(request.Request.URL.String(), m.Path))

			for si := range m.Steps {
				s := &m.Steps[si]
				s.Ref = ReverseUrl(request.Request, path.Join(request.Request.URL.String(), s.Path))
			}
		}

		l.Items[i] = ev.events[k]
		i++
	}
	*/
	count, err := ev.session.DB("catcierge").C("events").Count()
	if err != nil {
		WriteCatciergeErrorString(response, http.StatusInternalServerError, fmt.Sprintf("Failed to get event count"))
	}
	l.Count = count

	err = ev.session.DB("catcierge").C("events").Find(nil).Skip(l.Offset).Limit(l.Limit).All(&l.Items)
	if err != nil {
		log.Printf("Failed to list items: %s", err)
		WriteCatciergeErrorString(response, http.StatusInternalServerError, fmt.Sprintf("Failed to list events"))
		return
	}

	response.WriteEntity(l)
}

func (ev CatEventResource) getEvent(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("event-id")
	//vent, ok := ev.events[id]
	oid := bson.ObjectIdHex(id[0:24])
	catEvent := CatEvent{}

	if err := ev.session.DB("catcierge").C("events").FindId(oid).One(&catEvent); err != nil {
		WriteCatciergeErrorString(response, http.StatusNotFound, fmt.Sprintf("Event '%s' could not be found", id))
		return
	}

	response.WriteEntity(catEvent)
}

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
	eventData, err := UnzipEvent(tmpfile.Name(), *eventPath)
	if err != nil {
		log.Printf("Failed to unzip file %v to %v: %s", tmpfile.Name(), *eventPath, err)
		WriteCatciergeErrorString(response, http.StatusInternalServerError, "")
		return
	}

	// Create the event in MongoDB.
	catEvent := CatEvent{ID: bson.ObjectIdHex(eventData.ID[0:24]), Data: *eventData}
	ev.session.DB("catcierge").C("events").Insert(catEvent)

	response.WriteHeader(http.StatusCreated)
	response.WriteEntity(catEvent) // TODO: Should return with Ref links.

	log.Printf("Successfully unpacked event\n")
}
