package main

import (
	"context"
	"net/http"

	"labix.org/v2/mgo"
)

// CatciergeResource base resource for the entire API.
type CatciergeResource struct {
	// MongoDB session.
	session  *mgo.Session
	settings *catSettings
}

// CatciergeContextWrapper is an interface that wraps a HTTP handler with a catcierge resource.
type CatciergeContextWrapper interface {
	// WrapContext Wraps the given Handler and injects a context into each request.
	WrapContext(handler http.Handler, c *context.Context) http.Handler
}
