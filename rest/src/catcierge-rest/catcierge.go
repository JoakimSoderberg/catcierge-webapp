package main

import "labix.org/v2/mgo"

// CatciergeResource base resource for the entire API.
type CatciergeResource struct {
	// MongoDB session.
	session  *mgo.Session
	settings *catSettings
}
