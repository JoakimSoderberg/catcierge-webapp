package main

// TODO: Implement users resource
// TODO: Implement API tokens for users

import (
	"context"
	"net/http"

	"labix.org/v2/mgo"

	restful "github.com/emicklei/go-restful"
)

// UsersResource User
type UsersResource struct {
	CatciergeResource
}

// User representation.
type User struct {
	Name string `json:"name"`
}

// UserListResponse A response returned when listing the Users.
type UserListResponse struct {
	ListResponseHeader
	Items []User `json:"items"`
}

var usersKey key

// FromUsersContext returns the CatEventResource in ctx, if any.
func FromUsersContext(ctx context.Context) (*UsersResource, bool) {
	ev, ok := ctx.Value(usersKey).(*UsersResource)
	return ev, ok
}

// AddContext add UsersResource to the context.
func (ac *UsersResource) AddContext(c *context.Context) context.Context {
	return context.WithValue(*c, usersKey, ac)
}

// NewUserResource create a new UsersResource
func NewUserResource(session *mgo.Session, settings *CatSettings) *UsersResource {
	return &UsersResource{CatciergeResource{session: session, settings: settings}}
}

// Register UsersResource resource end points.
func (ac UsersResource) Register(container *restful.Container) {
	ws := new(restful.WebService)

	userID := ws.PathParameter("user-id", "User ID").DataType("string")

	ws.Path("/users").
		Doc("Mange users").
		Consumes("application/json").
		Produces(restful.MIME_JSON, restful.MIME_XML)

	ws.Route(ws.GET("/").To(ac.listUsers).
		Doc("List users").
		Returns(http.StatusOK, http.StatusText(http.StatusOK), []CatEvent{}).
		Do(AddListRequestParams(ws),
			ReturnsError(http.StatusInternalServerError)).
		Writes(UserListResponse{}))

	ws.Route(ws.GET("/{user-id}").To(ac.getUser).
		Doc("Get a user").
		Param(userID).
		Do(ReturnsStatus(http.StatusOK, "", User{}),
			ReturnsError(http.StatusNotFound),
			ReturnsError(http.StatusInternalServerError)).
		Writes(User{}))

	container.Add(ws)
}

func (ac *UsersResource) listUsers(req *restful.Request, resp *restful.Response) {
	// TODO: Check if user is logged in and list Users based on access.
	// TODO: If not logged in, still list public Users.
}

func (ac *UsersResource) getUser(req *restful.Request, resp *restful.Response) {

}
