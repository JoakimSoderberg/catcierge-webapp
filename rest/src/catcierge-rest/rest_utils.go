package main

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/emicklei/go-restful"
)

type CatError struct {
	HTTPStatusCode int    `json:"http_status_code"`
	HTTPStatus     string `json:"http_status"`
	Message        string `json:"message"`
}

// ListResponseHeader Used for all list resource responses to return pagination
// and other general information.
type ListResponseHeader struct {
	Count  int `json:"count"`
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

// ReverseUrl Based on an incoming HTTP request gets the full URL with hostname.
func ReverseUrl(request *http.Request, fullPath string) string {
	ev, ok := FromEventContext(request.Context())
	serverScheme := "http"
	if ok {
		serverScheme = ev.settings.serverScheme
	}

	revURL := url.URL{
		Host:   request.Host,
		Path:   strings.Trim(fullPath, "/"),
		Scheme: serverScheme}
	return revURL.String()
}

func (l ListResponseHeader) getListResponseParams(request *restful.Request) {

	offset, err := strconv.Atoi(request.QueryParameter("offset"))
	if err != nil {
		offset = DefaultPageOffset
	}

	limit, err := strconv.Atoi(request.QueryParameter("limit"))
	if err != nil {
		limit = DefaultPageLimit
	}

	l.Offset = offset
	l.Limit = limit
}

// AddListRequestParams Sets parameters for list pagination for a resource.
func AddListRequestParams(ws *restful.WebService) func(b *restful.RouteBuilder) {
	return func(b *restful.RouteBuilder) {
		b.Param(ws.QueryParameter("offset", "Offset into the list").
			DataType("int").DefaultValue(string(DefaultPageOffset)))

		b.Param(ws.QueryParameter("limit", "Number of items to return").
			DataType("int").DefaultValue(string(DefaultPageLimit)))
	}
}

// WriteCatciergeErrorString Writes an error message as a response.
func WriteCatciergeErrorString(response *restful.Response, httpStatus int, message string) {
	if message == "" {
		message = http.StatusText(httpStatus)
	}

	// TODO: Set correct content type based on what was requested by user.
	response.AddHeader("Content-Type", "application/json")
	response.WriteHeaderAndEntity(httpStatus,
		&CatError{
			HTTPStatusCode: httpStatus,
			HTTPStatus:     http.StatusText(httpStatus),
			Message:        message})
}

// ReturnsStatus Helper for specifying what HTTP statuses a restful.RouteBuilder returns.
func ReturnsStatus(httpStatus int, message string, model interface{}) func(b *restful.RouteBuilder) {
	return func(b *restful.RouteBuilder) {
		if message == "" {
			message = http.StatusText(httpStatus)
		}
		b.Returns(httpStatus, message, model)
	}
}

// ReturnsError A more concise wrapper for error status codes for ReturnsStatus.
func ReturnsError(httpStatus int) func(b *restful.RouteBuilder) {
	return ReturnsStatus(httpStatus, "", CatError{})
}
