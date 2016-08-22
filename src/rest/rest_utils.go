package main

import (
    "net/http"   
    "github.com/emicklei/go-restful"
)

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
