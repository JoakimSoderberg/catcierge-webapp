package main

import (
	"context"
	"net/http"

	restful "github.com/emicklei/go-restful"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

// AuthenticationState is used to store Authentication state for
// a single HTTP request. If the request is authenticated and if
// so what user is logged in (our owns the token that was used to login).
type AuthenticationState struct {
	IsAuthenticated bool  // If this request is authenticated.
	User            *User // The logged in user if any.
}

var authStateKey key

// FromAuthStateContext returns the CatEventResource in ctx, if any.
func FromAuthStateContext(ctx context.Context) (*AuthenticationState, bool) {
	ev, ok := ctx.Value(authStateKey).(*AuthenticationState)
	return ev, ok
}

// AddContext add AccountsResource to the context.
func (ac *AuthenticationState) AddContext(c *context.Context) context.Context {
	return context.WithValue(*c, authStateKey, ac)
}

// NewAuthenticationState create a new AuthenticationState
func NewAuthenticationState(isAuthenticated bool, user *User) *AuthenticationState {
	return &AuthenticationState{IsAuthenticated: isAuthenticated, User: user}
}

// AccessToken is used to authenticate to the API with.
type AccessToken struct {
	ID        bson.ObjectId `json:"id" bson:"_id"`
	Token     string        `json:"token" bson:"token"`
	UserID    bson.ObjectId `json:"user_id" bson:"user_id"`
	AccountID bson.ObjectId `json:"account_id" bson:"account_id"`
}

// AccessTokenResource resource
type AccessTokenResource struct {
	CatciergeResource
}

// TODO: Add "clients" to database"
/*
User – a user who has a name, password hash and a salt.
Client – a client application which requests access on behalf of a user, has a name and a secret code.
AccessToken – token (type of bearer), issued to the client application, limited by time.
RefreshToken – another type of token allows you to request a new bearer-token without re-request a password from the user.
*/

// AccessTokenListResponse A response returned when listing the access tokens.
type AccessTokenListResponse struct {
	ListResponseHeader
	Items []AccessToken `json:"items"`
}

var accessTokenKey key

// FromAcessTokenContext returns the CatEventResource in ctx, if any.
func FromAcessTokenContext(ctx context.Context) (*AccessTokenResource, bool) {
	ev, ok := ctx.Value(accessTokenKey).(*AccessTokenResource)
	return ev, ok
}

// AddContext add AccountsResource to the context.
func (ac *AccessTokenResource) AddContext(c *context.Context) context.Context {
	return context.WithValue(*c, accessTokenKey, ac)
}

// NewAccessTokensResource create a new AccessTokenResource
func NewAccessTokensResource(session *mgo.Session, settings *CatSettings) *AccessTokenResource {
	return &AccessTokenResource{CatciergeResource{session: session, settings: settings}}
}

// Register AccountsResource resource end points.
func (ac AccessTokenResource) Register(container *restful.Container) {
	ws := new(restful.WebService)

	tokenID := ws.PathParameter("token-id", "Access token ID").DataType("string")

	ws.Path("/tokens").
		Doc("Mange API access tokens").
		Consumes("application/json").
		Produces(restful.MIME_JSON, restful.MIME_XML)

	ws.Route(ws.GET("/").To(ac.listAccessTokens).
		Doc("List access tokens").
		Returns(http.StatusOK, http.StatusText(http.StatusOK), AccessTokenListResponse{}).
		Do(AddListRequestParams(ws),
			ReturnsError(http.StatusInternalServerError)).
		Writes(AccountListResponse{}))

	ws.Route(ws.GET("/{token-id}").To(ac.getAccessToken).
		Doc("Get a single access token").
		Param(tokenID).
		Do(ReturnsStatus(http.StatusOK, "", Account{}),
			ReturnsError(http.StatusNotFound),
			ReturnsError(http.StatusInternalServerError)).
		Writes(Account{}))

	ws.Route(ws.POST("").To(ac.createAccessToken).
		Doc("Create an access token").
		Do(ReturnsStatus(http.StatusOK, "", Account{}),
			ReturnsError(http.StatusBadRequest),
			ReturnsError(http.StatusUnauthorized),
			ReturnsError(http.StatusConflict),
			ReturnsError(http.StatusInternalServerError)))

	container.Add(ws)
}

func (ev *AccessTokenResource) listAccessTokens(request *restful.Request, response *restful.Response) {
	// TODO: We can only list access tokens for the currently logged in user.
}

func (ev *AccessTokenResource) getAccessToken(request *restful.Request, response *restful.Response) {
}

func (ev *AccessTokenResource) createAccessToken(request *restful.Request, response *restful.Response) {

}
