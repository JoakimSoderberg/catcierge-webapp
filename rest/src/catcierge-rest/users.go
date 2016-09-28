package main

// TODO: Implement API tokens for users

import (
	"context"
	"net/http"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"

	restful "github.com/emicklei/go-restful"
)

// UsersResource User
type UsersResource struct {
	CatciergeResource
}

// User representation.
type User struct {
	ID        bson.ObjectId `json:"id" bson:"_id"` // User ID in MongoDB.
	Name      string        `json:"name"`          // Full name.
	AvatarURL string        `json:"avatar_url"`    // Avatar image URL.
	Email     string        `json:"email"`         // E-mail.
	Nickname  string        `json:"nickname"`      // Nickname or username.
	Provider  string        `json:"provider"`      // Login provider (if any).
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
func (us *UsersResource) AddContext(c *context.Context) context.Context {
	return context.WithValue(*c, usersKey, us)
}

// NewUserResource create a new UsersResource
func NewUserResource(session *mgo.Session, settings *CatSettings) *UsersResource {
	return &UsersResource{CatciergeResource{session: session, settings: settings}}
}

// Register UsersResource resource end points.
func (us UsersResource) Register(container *restful.Container) {
	ws := new(restful.WebService)

	userID := ws.PathParameter("user-id", "User ID").DataType("string")

	ws.Path("/users").
		Doc("Mange users").
		Consumes("application/json").
		Produces(restful.MIME_JSON, restful.MIME_XML)

	ws.Route(ws.GET("/").To(us.listUsers).
		Doc("List users").
		Returns(http.StatusOK, http.StatusText(http.StatusOK), []CatEvent{}).
		Do(AddListRequestParams(ws),
			ReturnsError(http.StatusInternalServerError)).
		Writes(UserListResponse{}))

	ws.Route(ws.GET("/{user-id}").To(us.getUser).
		Doc("Get a user").
		Param(userID).
		Do(ReturnsStatus(http.StatusOK, "", User{}),
			ReturnsError(http.StatusNotFound),
			ReturnsError(http.StatusInternalServerError)).
		Writes(User{}))

	container.Add(ws)
}

func (us *UsersResource) listUsers(req *restful.Request, resp *restful.Response) {
	// TODO: Check if user is logged in and list Users based on access.
	// TODO: If not logged in, still list public Users.
}

func (us *UsersResource) getUser(req *restful.Request, resp *restful.Response) {

}
