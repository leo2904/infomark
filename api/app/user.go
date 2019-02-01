// InfoMark - a platform for managing courses with
//            distributing exercise sheets and testing exercise submissions
// Copyright (C) 2019  ComputerGraphics Tuebingen
// Authors: Patrick Wieschollek
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package app

import (
  "context"
  "net/http"
  "strconv"

  "github.com/cgtuebingen/infomark-backend/api/auth"
  "github.com/cgtuebingen/infomark-backend/model"
  "github.com/go-chi/chi"
  "github.com/go-chi/render"
)

// UserStore specifies required database queries for user management.
type UserStore interface {
  Get(userID int64) (*model.User, error)
  Update(p *model.User) error
  GetAll() ([]model.User, error)
  Create(p *model.User) (*model.User, error)
}

// UserResource specifies user management handler.
type UserResource struct {
  UserStore UserStore
}

// NewUserResource create and returns a UserResource.
func NewUserResource(userStore UserStore) *UserResource {
  return &UserResource{
    UserStore: userStore,
  }
}

// .............................................................................

// userRequest is the request payload for user management.
type userRequest struct {
  *model.User
  ProtectedID   int64  `json:"id"`
  PlainPassword string `json:"plain_password"`
}

// userResponse is the response payload for user management.
type userResponse struct {
  *model.User
}

// newUserResponse creates a response from a user model.
func newUserResponse(p *model.User) *userResponse {
  return &userResponse{
    User: p,
  }
}

// newUserListResponse creates a response from a list of user models.
func newUserListResponse(users []model.User) []render.Renderer {
  // https://stackoverflow.com/a/36463641/7443104
  list := []render.Renderer{}
  for k := range users {
    list = append(list, newUserResponse(&users[k]))
  }

  return list
}

// Bind preprocesses a userRequest.
func (d *userRequest) Bind(r *http.Request) error {
  // Sending the id via request-body is invalid.
  // The id should be submitted in the url.
  d.ProtectedID = 0

  // Encrypt plain password
  hash, err := auth.HashPassword(d.PlainPassword)
  d.EncryptedPassword = hash

  return err
}

// Render post-processes a userResponse.
func (u *userResponse) Render(w http.ResponseWriter, r *http.Request) error {
  // nothing to hide
  return nil
}

// bindValidate jointly binds data from json request and validates the model.
func (rs *UserResource) bindValidate(w http.ResponseWriter, r *http.Request) (*userRequest, error) {
  // get user from middle-ware context
  data := &userRequest{User: r.Context().Value("user").(*model.User)}

  // parse JSON request into struct
  if err := render.Bind(r, data); err != nil {
    return nil, err
  }

  // validate final model
  if err := data.User.Validate(); err != nil {
    return nil, err
  }

  return data, nil
}

// Index is the enpoint for retrieving all users.
func (rs *UserResource) Index(w http.ResponseWriter, r *http.Request) {
  // fetch collection of users from database
  users, err := rs.UserStore.GetAll()

  // render JSON reponse
  if err = render.RenderList(w, r, newUserListResponse(users)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }
}

// Get is the enpoint for retrieving a specific user.
func (rs *UserResource) Get(w http.ResponseWriter, r *http.Request) {
  // `user` is retrieved via middle-ware
  user := r.Context().Value("user").(*model.User)

  // render JSON reponse
  if err := render.Render(w, r, newUserResponse(user)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }
}

// Patch is the endpoint fro updating a specific user with given id.
func (rs *UserResource) Patch(w http.ResponseWriter, r *http.Request) {

  data, err := rs.bindValidate(w, r)
  if err != nil {
    render.Render(w, r, ErrBadRequest)
    return
  }

  // update database entry
  if err := rs.UserStore.Update(data.User); err != nil {
    render.Render(w, r, ErrInternalServerError)
    return
  }

  render.Status(r, http.StatusNoContent)
}

// .............................................................................
// Context middleware is used to load an User object from
// the URL parameter `userID` passed through as the request. In case
// the User could not be found, we stop here and return a 404.
// We do NOT check whether the user is authorized to get this user.
func (d *UserResource) Context(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // TODO: check permission if inquirer of request is allowed to access this user
    // Should be done via another middleware
    var user_id int64
    var err error

    // try to get id from URL
    if user_id, err = strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64); err != nil {
      render.Render(w, r, ErrNotFound)
      return
    }

    // find specific user in database
    user, err := d.UserStore.Get(user_id)
    if err != nil {
      render.Render(w, r, ErrNotFound)
      return
    }

    // serve next
    ctx := context.WithValue(r.Context(), "user", user)
    next.ServeHTTP(w, r.WithContext(ctx))
  })
}
