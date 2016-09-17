package main

import (
	"context"
	"net/http"

	"labix.org/v2/mgo"

	restful "github.com/emicklei/go-restful"
)

// AccountsResource Account
type AccountsResource struct {
	CatciergeResource
}

// Account representation.
type Account struct {
	Name string `json:"name"`
}

// AccountListResponse A response returned when listing the accounts.
type AccountListResponse struct {
	ListResponseHeader
	Items []Account `json:"items"`
}

var accountsKey key

// NewAccountsContext Returns a new Context containing the CatEventResource value.
func NewAccountsContext(ctx context.Context, ev *AccountsResource) context.Context {
	return context.WithValue(ctx, accountsKey, ev)
}

// FromAccountsContext returns the CatEventResource in ctx, if any.
func FromAccountsContext(ctx context.Context) (*AccountsResource, bool) {
	ev, ok := ctx.Value(accountsKey).(*AccountsResource)
	return ev, ok
}

// WrapContext Wraps the given Handler and injects a context into each request containing the AccountsResource.
func (ac *AccountsResource) WrapContext(handler http.Handler, c *context.Context) http.Handler {
	// Create a new context and inject our accounts resource into it.
	ctx := NewAccountsContext(*c, ac)
	wrapped := func(w http.ResponseWriter, req *http.Request) {
		handler.ServeHTTP(w, req.WithContext(ctx))
	}
	return http.HandlerFunc(wrapped)
}

// NewAccountResource create a new AccountsResource
func NewAccountResource(session *mgo.Session, settings *catSettings) *AccountsResource {
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

	container.Add(ws)
}

func (ac *AccountsResource) listAccounts(req *restful.Request, resp *restful.Response) {
	// TODO: Check if user is logged in and list accounts based on access.
	// TODO: If not logged in, still list public accounts.
}

func (ac *AccountsResource) getAccount(req *restful.Request, resp *restful.Response) {

}
