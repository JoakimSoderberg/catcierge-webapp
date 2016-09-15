package main

import "labix.org/v2/mgo"

type CatciergeResource struct {
	session  *mgo.Session
	settings *catSettings
	events   *CatEventResource
	accounts *Accounts
}

func NewCatciergeResource(session *mgo.Session, settings *catSettings) *CatciergeResource {
	return &CatciergeResource{
		session:  session,
		settings: settings}
}
