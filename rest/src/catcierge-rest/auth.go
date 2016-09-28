package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	restful "github.com/emicklei/go-restful"
	"github.com/satori/go.uuid"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

// AuthenticationState is used to store Authentication state for
// a single HTTP request. If the request is authenticated and if
// so what user is logged in (our owns the token that was used to login).
type AuthenticationState struct {
	IsAuthenticated bool     // If this request is authenticated.
	User            *User    // The logged in user if any.
	Account         *Account // The account the user is logged in to.
}

var authStateKey key

// FromAuthStateContext returns the CatEventResource in ctx, if any.
func FromAuthStateContext(ctx context.Context) (*AuthenticationState, bool) {
	ev, ok := ctx.Value(authStateKey).(*AuthenticationState)
	return ev, ok
}

// AddContext add AccountsResource to the context.
func (at *AuthenticationState) AddContext(c *context.Context) context.Context {
	return context.WithValue(*c, authStateKey, at)
}

// NewAuthenticationState create a new AuthenticationState
func NewAuthenticationState(isAuthenticated bool, user *User) *AuthenticationState {
	return &AuthenticationState{IsAuthenticated: isAuthenticated, User: user}
}

// AccessToken is used to authenticate to the API with.
type AccessToken struct {
	ID        bson.ObjectId `json:"id" bson:"_id"`
	Name      string        `json:"name" bson:"name"`
	Token     string        `json:"token" bson:"token"`
	UserID    bson.ObjectId `json:"user_id" bson:"user_id"`
	AccountID bson.ObjectId `json:"account_id" bson:"account_id"`
}

// AccessTokenPublic Public version of an access token only shows the name.
// For listing via the API.
type AccessTokenPublic struct {
	Name  string `json:"name" bson:"name"`
	Token string `json:"token" bson:"token"`
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
	Items []AccessTokenPublic `json:"items"`
}

var accessTokenKey key

// FromAcessTokenContext returns the CatEventResource in ctx, if any.
func FromAcessTokenContext(ctx context.Context) (*AccessTokenResource, bool) {
	ev, ok := ctx.Value(accessTokenKey).(*AccessTokenResource)
	return ev, ok
}

// AddContext add AccountsResource to the context.
func (at *AccessTokenResource) AddContext(c *context.Context) context.Context {
	return context.WithValue(*c, accessTokenKey, at)
}

// NewAccessTokensResource create a new AccessTokenResource
func NewAccessTokensResource(session *mgo.Session, settings *CatSettings) *AccessTokenResource {
	return &AccessTokenResource{CatciergeResource{session: session, settings: settings}}
}

// Register AccountsResource resource end points.
func (at AccessTokenResource) Register(container *restful.Container) {
	ws := new(restful.WebService)

	tokenID := ws.PathParameter("token-name", "Access token name").DataType("string")

	ws.Path("/tokens").
		Doc("Mange API access tokens").
		Consumes("application/json").
		Produces(restful.MIME_JSON, restful.MIME_XML)

	ws.Route(ws.GET("/").To(at.listAccessTokens).
		Doc("List access tokens").
		Returns(http.StatusOK, http.StatusText(http.StatusOK), AccessTokenListResponse{}).
		Do(AddListRequestParams(ws),
			ReturnsError(http.StatusInternalServerError)).
		Writes(AccessTokenListResponse{}))

	ws.Route(ws.GET("/{token-name}").To(at.getAccessToken).
		Doc("Get a single access token").
		Param(tokenID).
		Do(ReturnsStatus(http.StatusOK, "", AccessTokenPublic{}),
			ReturnsError(http.StatusNotFound),
			ReturnsError(http.StatusInternalServerError)).
		Writes(AccessTokenPublic{}))

	ws.Route(ws.POST("").To(at.createAccessToken).
		Doc("Create an access token").
		Param(ws.BodyParameter("name", "Name of token").DataType("string")).
		Do(ReturnsStatus(http.StatusOK, "", AccessTokenPublic{}),
			ReturnsError(http.StatusBadRequest),
			ReturnsError(http.StatusUnauthorized),
			ReturnsError(http.StatusConflict),
			ReturnsError(http.StatusInternalServerError)))

	container.Add(ws)
}

// IsAuthorizedForAccesTokens Checks if the request has the correct authorization to handle Access Tokens.
func IsAuthorizedForAccesTokens(request *restful.Request, response *restful.Response) (*AuthenticationState, error) {
	authState, ok := FromAuthStateContext(request.Request.Context())
	if !ok {
		WriteCatciergeErrorString(response, http.StatusInternalServerError, "")
		return nil, errors.New("Failed to get authentication state from request context")
	}

	if !authState.IsAuthenticated {
		WriteCatciergeErrorString(response, http.StatusUnauthorized,
			"You must be logged in to create an Access token")
		return authState, errors.New("Unauthenticated user")
	}

	if authState.Account == nil {
		WriteCatciergeErrorString(response, http.StatusUnauthorized,
			"You must be logged in to an account to be able to create an Access token")
		return authState, errors.New("User not member of any account")
	}

	return authState, nil
}

func (at *AccessTokenResource) listAccessTokens(request *restful.Request, response *restful.Response) {
	_, err := IsAuthorizedForAccesTokens(request, response)
	if err != nil {
		log.Printf("%s", err)
		return
	}

	var l = AccessTokenListResponse{}
	l.getListResponseParams(request)

	// TODO: Copy db session.
	// TODO: We can only list access tokens for the currently logged in user.
	err = at.session.DB("catcierge").C("tokens").Find(nil).Skip(l.Offset).Limit(l.Limit).All(&l.Items)
	if err != nil {
		WriteCatciergeErrorString(response, http.StatusInternalServerError, fmt.Sprintf("Failed to list Access Tokens"))
	}

	response.WriteEntity(l)
}

func (at *AccessTokenResource) getAccessToken(request *restful.Request, response *restful.Response) {
	_, err := IsAuthorizedForAccesTokens(request, response)
	if err != nil {
		log.Printf("%s", err)
		return
	}

	name := request.PathParameter("name")

	var token AccessTokenPublic

	err = at.session.DB("catcierge").C("tokens").Find(bson.M{"name": name}).One(&token)
	if err != nil {
		WriteCatciergeErrorString(response, http.StatusNotFound,
			fmt.Sprintf("No access token with name '%s' found", name))
		return
	}

	response.WriteEntity(&token)
}

func (at *AccessTokenResource) createAccessToken(request *restful.Request, response *restful.Response) {
	authState, err := IsAuthorizedForAccesTokens(request, response)
	if err != nil {
		log.Printf("%s", err)
		return
	}

	// TODO: Make helper function to get path or return error.
	name, err := request.BodyParameter("name")
	if err != nil {
		WriteCatciergeErrorString(response, http.StatusBadRequest, "Missing 'name' in request")
		return
	}

	tokenStr := uuid.NewV4().String()

	token := AccessToken{
		Name:      name,
		Token:     tokenStr, // TODO: Force to be unique in the database.
		UserID:    authState.User.ID,
		AccountID: authState.Account.ID}

	// TODO: Retry below if mgo.IsDup(err) on the Token.
	if err := at.session.DB("catcierge").C("tokens").Insert(&token); err != nil {
		if mgo.IsDup(err) {
			nameCount, _ := at.session.DB("catcierge").C("tokens").Find(bson.M{"name": name}).Count()
			if nameCount == 0 {
				// TODO: Retry adding using a newly generated token
			}
			WriteCatciergeErrorString(response, http.StatusConflict,
				fmt.Sprintf("A token with the name '%s' already exists", name))
			return
		}
	}

	response.WriteEntity(AccessTokenPublic{Name: name, Token: tokenStr})
}
