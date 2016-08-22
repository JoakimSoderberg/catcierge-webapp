package main

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type CatEventData struct {
	ID                  string    `json:"id"`
	EventJSONVersion    string    `json:"event_json_version"`
	CatciergeType       string    `json:"catcierge_type"`
	Description         string    `json:"description"`
	Start               time.Time `json:"start"`
	End                 string    `json:"end"`
	TimeGenerated       time.Time `json:"time_generated"`
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
		Result          int       `json:"result"`
		Success         int       `json:"success"`
		Time            time.Time `json:"time"`
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
	ID   bson.ObjectId `json:"id" bson:"_id"`
	Name string        `json:"name"`
	Data CatEventData  `json:"data" bson:"data"`
	Tags []string      `json:"tags" bson:"tags"`
}

type CatEventResource struct {
	// TODO: Replace with MongoDB
	events map[string]CatEvent
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

func (ev CatEventResource) listEvents(request *restful.Request, response *restful.Response) {
	response.WriteEntity(ev.events)
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
	destDir := filepath.Join(*unzipPath, tmpfileNoExt)

	if err := Unzip(tmpfile.Name(), destDir); err != nil {
		WriteCatciergeErrorString(response, http.StatusInternalServerError, "")
		return
	}

	//response.WriteHeaderAndEntity(http.StatusCreated, usr)
}
