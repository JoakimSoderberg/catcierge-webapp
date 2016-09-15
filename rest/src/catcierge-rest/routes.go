package main

import (
	"net/http"

	restful "github.com/emicklei/go-restful"
)

type resourceLocation struct {
	URL   string
	Route *restful.RouteBuilder
}

// Register all routes for the REST API.
func (cr CatciergeResource) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ev := cr.events
	ac := cr.accounts

	eventID := ws.PathParameter("event-id", "Identifier of the event").DataType("string")
	accountName := ws.PathParameter("account-name", "Account name").DataType("string")

	ws.Path("/").
		Doc("Catcierge web application").
		Produces(restful.MIME_JSON, restful.MIME_XML)

	ws.Route(ws.GET("/accounts").To(ac.accountsList).
		Doc("List accounts"))

	ws.Route(ws.GET("/accounts/{account-name}").To(ac.accountsList).
		Param(accountName).
		Doc("Account resource"))

	eventsList := new(restful.RouteBuilder).
		Method("GET").
		Doc("List events").
		Param(accountName).
		Returns(http.StatusOK, "", []CatEvent{}).
		Do(AddListRequestParams(ws),
			ReturnsError(http.StatusInternalServerError)).
		Writes(CatEventListResponse{})
	for _, loc := range []resourceLocation{
		{URL: "/events", Route: eventsList.To(ev.listEvents)},
		{URL: "/accounts/{account-name}/events", Route: (*eventsList).To(ev.listEvents).Param(accountName)}} {
		ws.Route(loc.Route.Path(loc.URL))
	}

	eventsItem := new(restful.RouteBuilder).
		Method("GET").
		Doc("Get an event").
		Param(eventID).
		Do(ReturnsStatus(http.StatusOK, "", CatEvent{}),
			ReturnsError(http.StatusNotFound),
			ReturnsError(http.StatusInternalServerError)).
		Writes(CatEvent{})
	for _, loc := range []resourceLocation{
		{URL: "/events/{event-id}", Route: eventsItem.To(ev.getEvent)},
		{URL: "/accounts/{account-name}/events/{event-id}", Route: (*eventsItem).To(ev.getEvent).Param(accountName)}} {
		ws.Route(loc.Route.Path(loc.URL))
	}

	// Static images.
	eventStatic := new(restful.RouteBuilder).
		Method("GET").
		Doc("Get static files for an event such as images").
		Param(accountName).
		Param(eventID).
		Do(ReturnsStatus(http.StatusOK, "", nil),
			ReturnsError(http.StatusNotFound),
			ReturnsError(http.StatusInternalServerError))
	for _, loc := range []resourceLocation{
		{URL: "/events/{event-id}/{subpath:*}", Route: eventStatic.To(ev.eventStaticFiles)},
		{URL: "/accounts/{account-name}/events/{event-id}/{subpath:*}", Route: (*eventStatic).To(ev.eventStaticFiles).Param(accountName)}} {
		ws.Route(loc.Route.Path(loc.URL))
	}

	// We don't allow creating events in the global /events/ resource.
	ws.Route(ws.POST("/accounts/{account-name}/events").To(ev.createEvent).
		Consumes("application/zip").
		Produces("application/json").
		Param(accountName).
		Doc("Create an event based on an event ZIP file").
		Do(ReturnsError(http.StatusBadRequest),
			ReturnsError(http.StatusConflict),
			ReturnsError(http.StatusInternalServerError)))

	container.Add(ws)
}
