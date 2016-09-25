package main

import (
	"context"
	"net/http"

	"labix.org/v2/mgo"

	restful "github.com/emicklei/go-restful"
	bson "labix.org/v2/mgo/bson"
)

// AccountsResource Account
type AccountsResource struct {
	CatciergeResource
}

// Account representation.
type Account struct {
	ID    bson.ObjectId   `json:"id" bson:"_id"`
	Name  string          `json:"name"`
	Users []bson.ObjectId `json:"users" bson:"users"`
}

// AccountListResponse A response returned when listing the accounts.
type AccountListResponse struct {
	ListResponseHeader
	Items []Account `json:"items"`
}

var accountsKey key

// FromAccountsContext returns the CatEventResource in ctx, if any.
func FromAccountsContext(ctx context.Context) (*AccountsResource, bool) {
	ev, ok := ctx.Value(accountsKey).(*AccountsResource)
	return ev, ok
}

// AddContext add AccountsResource to the context.
func (ac *AccountsResource) AddContext(c *context.Context) context.Context {
	return context.WithValue(*c, accountsKey, ac)
}

// NewAccountResource create a new AccountsResource
func NewAccountResource(session *mgo.Session, settings *CatSettings) *AccountsResource {
	return &AccountsResource{CatciergeResource{session: session, settings: settings}}
}

// Register AccountsResource resource end points.
func (ac AccountsResource) Register(container *restful.Container) {
	ws := new(restful.WebService)

	accountName := ws.PathParameter("account-name", "Account name").DataType("string")

	ws.Path("/accounts").
		Doc("Mange accounts").
		Consumes("application/json").
		Produces(restful.MIME_JSON, restful.MIME_XML)

	ws.Route(ws.GET("/").To(ac.listAccounts).
		Doc("List accounts").
		Returns(http.StatusOK, http.StatusText(http.StatusOK), []CatEvent{}).
		Do(AddListRequestParams(ws),
			ReturnsError(http.StatusInternalServerError)).
		Writes(AccountListResponse{}))

	ws.Route(ws.GET("/{account-name}").To(ac.getAccount).
		Doc("Get an account").
		Param(accountName).
		Do(ReturnsStatus(http.StatusOK, "", Account{}),
			ReturnsError(http.StatusNotFound),
			ReturnsError(http.StatusInternalServerError)).
		Writes(Account{}))

	ws.Route(ws.POST("").To(ac.createAccount).
		Doc("Create an account").
		Do(ReturnsStatus(http.StatusOK, "", Account{}),
			ReturnsError(http.StatusBadRequest),
			ReturnsError(http.StatusUnauthorized),
			ReturnsError(http.StatusConflict),
			ReturnsError(http.StatusInternalServerError)))

	container.Add(ws)
}

func (ac *AccountsResource) listAccounts(req *restful.Request, resp *restful.Response) {
	// TODO: Check if user is logged in and list accounts based on access.
	// TODO: If not logged in, still list public accounts.
}

func (ac *AccountsResource) getAccount(req *restful.Request, resp *restful.Response) {

}

func (ac *AccountsResource) createAccount(req *restful.Request, resp *restful.Response) {
	// TODO: To create an account one has to be:
	// TODO: An authenticated user
	// TODO:
}
