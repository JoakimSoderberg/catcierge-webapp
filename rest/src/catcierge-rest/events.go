package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"

	restful "github.com/emicklei/go-restful"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

// DefaultPageOffset The default page offset for pagination.
const DefaultPageOffset = 0

// DefaultPageLimit The default page limit for pagination.
const DefaultPageLimit = 10

// DefaultMaxEventSize The default max ZIP size for a cat event.
const DefaultMaxEventSize = 2 * MB

// CatEvent Cat event.
type CatEvent struct {
	ID      bson.ObjectId  `json:"id" bson:"_id"`
	Name    string         `json:"name" bson:"name"`
	Data    CatEventDataV1 `json:"data" bson:"data"`
	Tags    []string       `json:"tags" bson:"tags"`
	Missing bool           `json:"missing" bson:"missing"`
}

// FillResponse This will fill a CatEvent struct with URLs based on the request origin
// as well as the Path specified in the JSON.
func (c *CatEvent) FillResponse(request *restful.Request) {
	d := &c.Data
	for mi := range d.Matches {
		m := &d.Matches[mi]
		m.Ref = ReverseURL(request.Request, path.Join(request.Request.URL.String(), m.Path))

		for si := range m.Steps {
			s := &m.Steps[si]
			s.Ref = ReverseURL(request.Request, path.Join(request.Request.URL.String(), s.Path))
		}
	}
}

// CatEventsResource A REST resource representing the CatEvents.
type CatEventsResource struct {
	CatciergeResource
}

// CatEventListResponse A response returned when listing the CatEventResource.
type CatEventListResponse struct {
	ListResponseHeader
	Items []CatEvent `json:"items"`
}

var eventsKey key

// FromEventsContext returns the CatEventResource in ctx, if any.
func FromEventsContext(ctx context.Context) (*CatEventsResource, bool) {
	ev, ok := ctx.Value(eventsKey).(*CatEventsResource)
	return ev, ok
}

// AddContext appends the CatEventsResource to the request context.
func (ev *CatEventsResource) AddContext(c *context.Context) context.Context {
	return context.WithValue(*c, eventsKey, ev)
}

// NewEventsResource Create a new CatEventResource instance.
func NewEventsResource(session *mgo.Session, settings *CatSettings) *CatEventsResource {
	return &CatEventsResource{CatciergeResource{session: session, settings: settings}}
}

// TODO: Add a new resource for listing all images.

// Register Registers the resource endpoints for a CatEventResource.
func (ev CatEventsResource) Register(container *restful.Container) {
	ws := new(restful.WebService)

	eventID := ws.PathParameter("event-id", "identifier of the event").DataType("string")

	ws.Path("/events").
		Doc("Manage events").
		Consumes("application/zip").
		Produces(restful.MIME_JSON, restful.MIME_XML)

	ws.Route(ws.GET("/").To(ev.listEvents).
		Doc("Get all events").
		Returns(http.StatusOK, http.StatusText(http.StatusOK), CatEventListResponse{}).
		Do(AddListRequestParams(ws),
			ReturnsError(http.StatusInternalServerError)).
		Writes(CatEventListResponse{}))

	ws.Route(ws.GET("/{event-id}").To(ev.getEvent).
		Doc("Get an event").
		Param(eventID).
		Do(ReturnsStatus(http.StatusOK, "", CatEvent{}),
			ReturnsError(http.StatusNotFound),
			ReturnsError(http.StatusInternalServerError)).
		Writes(CatEvent{}))

	// Static images.
	ws.Route(ws.GET("/{event-id}/{subpath:*}").To(ev.eventStaticFiles).
		Doc("Get static files for an event such as images").
		Param(eventID).
		Do(ReturnsStatus(http.StatusOK, "", CatEvent{}),
			ReturnsError(http.StatusNotFound),
			ReturnsError(http.StatusInternalServerError)))

	ws.Route(ws.POST("").To(ev.createEvent).
		Doc("Create an event based on an event ZIP file").
		Do(ReturnsStatus(http.StatusOK, "", CatEvent{}),
			ReturnsError(http.StatusBadRequest),
			ReturnsError(http.StatusUnauthorized),
			ReturnsError(http.StatusConflict),
			ReturnsError(http.StatusInternalServerError)))

	container.Add(ws)
}

func (ev *CatEventsResource) eventStaticFiles(req *restful.Request, resp *restful.Response) {
	fullPath := path.Join(ev.settings.eventPath, req.PathParameter("event-id"), req.PathParameter("subpath"))
	log.Printf("GET %s", fullPath)
	http.ServeFile(resp.ResponseWriter, req.Request, fullPath)
}

// List events. Supports pagination.
func (ev *CatEventsResource) listEvents(request *restful.Request, response *restful.Response) {
	var l = CatEventListResponse{}
	l.getListResponseParams(request)

	count, err := ev.session.DB("catcierge").C("events").Count()
	if err != nil {
		WriteCatciergeErrorString(response, http.StatusInternalServerError, fmt.Sprintf("Failed to get event count"))
	}
	l.Count = count

	err = ev.session.DB("catcierge").C("events").Find(nil).Skip(l.Offset).Limit(l.Limit).Sort("start").All(&l.Items)
	if err != nil {
		log.Printf("Failed to list items: %s", err)
		WriteCatciergeErrorString(response, http.StatusInternalServerError, fmt.Sprintf("Failed to list events"))
		return
	}

	for i := range l.Items {
		l.Items[i].FillResponse(request)
	}

	response.WriteEntity(l)
}

// Gets a single event.
func (ev *CatEventsResource) getEvent(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("event-id")
	oid := bson.ObjectIdHex(id[0:24])

	account, ok := request.PathParameters()["account-name"]
	if ok {
		// TODO: Change the query based on if we have an accout name or not
		WriteCatciergeErrorString(response, http.StatusNotFound, fmt.Sprintf("Account '%s' could not be found", account))
		return
	}

	catEvent := CatEvent{}

	if err := ev.session.DB("catcierge").C("events").FindId(oid).One(&catEvent); err != nil {
		WriteCatciergeErrorString(response, http.StatusNotFound, fmt.Sprintf("Event '%s' could not be found", id))
		return
	}

	response.WriteEntity(catEvent)
}

// Create a new catcierge event by uploading a ZIP file.
func (ev *CatEventsResource) createEvent(request *restful.Request, response *restful.Response) {
	// TODO: Copy the db session for this more costly operation.
	fileSize := ByteSize(request.Request.ContentLength)
	// TODO: Add check if user is logged in.
	if fileSize >= DefaultMaxEventSize {
		msg := fmt.Sprintf("Max file size allowed %s but got %s", DefaultMaxEventSize, fileSize)
		log.Println(msg)
		WriteCatciergeErrorString(response, http.StatusBadRequest, msg)
		return
	}

	log.Printf("Received file of size %s\n", fileSize)

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
	eventHeader, eventData, err := UnzipEvent(tmpfile.Name(), ev.settings.eventPath)
	if err != nil {
		log.Printf("Failed to unzip file %v to %v: %s", tmpfile.Name(), ev.settings.eventPath, err)
		extra := ""
		status := http.StatusInternalServerError

		switch err.(type) {
		case *CatJSONHeaderError:
			extra = fmt.Sprintf("Failed to parse the JSON header %s", err)
			status = http.StatusBadRequest
			break
		case *CatJSONError:
			extra = fmt.Sprintf("Failed to parse JSON for event %s. Expecting format %s: %s", eventHeader.ID, eventHeader.EventJSONVersion, err)
			status = http.StatusBadRequest
		case *CatJSONVersionError:
			extra = fmt.Sprintf("Failed to parse JSON for event %s. Event JSON version %s is not supported.", eventHeader.ID, eventHeader.EventJSONVersion)
			status = http.StatusBadRequest
		default:
			break
		}
		WriteCatciergeErrorString(response, status, extra)
		return
	}

	// Create the event in MongoDB.
	catEvent := CatEvent{ID: bson.ObjectIdHex(eventData.ID[0:24]), Data: *eventData}

	if err := ev.session.DB("catcierge").C("events").Insert(catEvent); err != nil {
		log.Printf("Failed to insert event in database: %s", err)
		if mgo.IsDup(err) {
			// TODO: Return a link to the existing resource in this error.
			WriteCatciergeErrorString(response, http.StatusConflict,
				fmt.Sprintf("An event with this ID already exists: %s", eventData.ID))
		} else {
			WriteCatciergeErrorString(response, http.StatusInternalServerError, "")
		}
		return
	}

	catEvent.FillResponse(request)
	response.WriteHeaderAndEntity(http.StatusCreated, catEvent)

	log.Printf("Successfully unpacked event %s\n", eventData.ID)
}
