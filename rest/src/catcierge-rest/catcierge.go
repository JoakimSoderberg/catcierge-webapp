package main

import (
	"context"

	"labix.org/v2/mgo"
)

// CatciergeResource base resource for the entire API.
type CatciergeResource struct {
	// MongoDB session.
	session  *mgo.Session
	settings *CatSettings
}

// CatciergeContextAdder Implementers of this interface will add a new context to a context parent.
type CatciergeContextAdder interface {
	AddContext(c *context.Context) context.Context
}
