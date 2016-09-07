package main

import (
	"github.com/emicklei/go-restful"
	"net/http"
	"strconv"
    "strings"
    "net/url"
)

type CatError struct {
	HttpStatusCode int    `json:"http_status_code"`
	HttpStatus     string `json:"http_status"`
	Message        string `json:"message"`
}

// Used for all list resource responses to return pagination
// and other general information.
type ListResponseHeader struct {
	Count  int `json:"count"`
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

func ReverseUrl(request *http.Request, fullPath string) string {
    revUrl := url.URL{Host: request.Host, Path: strings.Trim(fullPath, "/"), Scheme: serverScheme}
    return revUrl.String()
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

func AddListResponseParams(ws *restful.WebService) func(b *restful.RouteBuilder) {
	return func(b *restful.RouteBuilder) {
		b.Param(ws.QueryParameter("offset", "Offset into the list").DataType("int").DefaultValue(string(DefaultPageOffset))).
		  Param(ws.QueryParameter("limit", "Number of items to return").DataType("int").DefaultValue(string(DefaultPageLimit)))
	}
}

func WriteCatciergeErrorString(response *restful.Response, httpStatus int, message string) {
	if message == "" {
		message = http.StatusText(httpStatus)
	}

	response.AddHeader("Content-Type", "application/json")
	response.WriteEntity(&CatError{HttpStatusCode: httpStatus, HttpStatus: http.StatusText(httpStatus), Message: message})
}

func ReturnsStatus(httpStatus int, message string, model interface{}) func(b *restful.RouteBuilder) {
	return func(b *restful.RouteBuilder) {
		if message == "" {
			message = http.StatusText(httpStatus)
		}
		b.Returns(httpStatus, message, model)
	}
}

func ReturnsError(httpStatus int) func(b *restful.RouteBuilder) {
	return ReturnsStatus(httpStatus, "", CatError{})
}
