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
  "errors"
  "net/http"
  "strconv"

  "github.com/cgtuebingen/infomark-backend/api/helper"
  "github.com/cgtuebingen/infomark-backend/auth/authenticate"
  "github.com/cgtuebingen/infomark-backend/auth/authorize"
  "github.com/cgtuebingen/infomark-backend/email"
  "github.com/cgtuebingen/infomark-backend/model"
  "github.com/go-chi/chi"
  "github.com/go-chi/render"
)

// CourseResource specifies course management handler.
type CourseResource struct {
  Stores *Stores
}

// NewCourseResource create and returns a CourseResource.
func NewCourseResource(stores *Stores) *CourseResource {
  return &CourseResource{
    Stores: stores,
  }
}

// .............................................................................

// courseResponse is the response payload for course management.
type courseResponse struct {
  *model.Course
}

// Render post-processes a courseResponse.
func (body *courseResponse) Render(w http.ResponseWriter, r *http.Request) error {
  return nil
}

// newCourseResponse creates a response from a course model.
func (rs *CourseResource) newCourseResponse(p *model.Course) *courseResponse {
  return &courseResponse{
    Course: p,
  }
}

// newCourseListResponse creates a response from a list of course models.
func (rs *CourseResource) newCourseListResponse(courses []model.Course) []render.Renderer {
  // https://stackoverflow.com/a/36463641/7443104
  list := []render.Renderer{}
  for k := range courses {
    list = append(list, rs.newCourseResponse(&courses[k]))
  }
  return list
}

type SheetPointsResponse struct {
  SheetPoints *model.SheetPoints `json:"sheet_points"`
}

func (body *SheetPointsResponse) Render(w http.ResponseWriter, r *http.Request) error {
  return nil
}

func newSheetPointsResponse(p *model.SheetPoints) *SheetPointsResponse {
  return &SheetPointsResponse{
    SheetPoints: p,
  }
}

// newCourseListResponse creates a response from a list of course models.
func newSheetPointsListResponse(collection []model.SheetPoints) []render.Renderer {
  list := []render.Renderer{}
  for k := range collection {
    list = append(list, newSheetPointsResponse(&collection[k]))
  }

  return list
}

// .............................................................................
type groupBidsResponse struct {
  *model.GroupBid
}

// Render post-processes a groupBidsResponse.
func (body *groupBidsResponse) Render(w http.ResponseWriter, r *http.Request) error {
  return nil
}

// newCourseResponse creates a response from a course model.
func newGroupBidsResponse(p *model.GroupBid) *groupBidsResponse {
  return &groupBidsResponse{
    GroupBid: p,
  }
}

func newGroupBidsListResponse(collection []model.GroupBid) []render.Renderer {
  list := []render.Renderer{}
  for k := range collection {
    list = append(list, newGroupBidsResponse(&collection[k]))
  }

  return list
}

// .............................................................................

// courseResponse is the response payload for course management.
type enrollmentResponse struct {
  Role int64       `json:"role"`
  User *model.User `json:"user"`
}

// Render post-processes a courseResponse.
func (body *enrollmentResponse) Render(w http.ResponseWriter, r *http.Request) error {
  return nil
}

// newCourseResponse creates a response from a course model.
func (rs *CourseResource) newEnrollmentResponse(p *model.UserCourse) *enrollmentResponse {

  return &enrollmentResponse{
    Role: p.Role,
    User: &model.User{
      ID:        p.UserID,
      FirstName: p.FirstName,
      LastName:  p.LastName,
      // AvatarPath:    p.AvatarPath,
      Email:         p.Email,
      StudentNumber: p.StudentNumber,
      Semester:      p.Semester,
      Subject:       p.Subject,
      Language:      p.Language,
    },
  }
}

func (rs *CourseResource) newEnrollmentListResponse(enrollments []model.UserCourse) []render.Renderer {
  list := []render.Renderer{}
  for k := range enrollments {
    list = append(list, rs.newEnrollmentResponse(&enrollments[k]))
  }

  return list
}

// .............................................................................

// IndexHandler is the enpoint for retrieving all courses if claim.root is true.
func (rs *CourseResource) IndexHandler(w http.ResponseWriter, r *http.Request) {
  // fetch collection of courses from database
  courses, err := rs.Stores.Course.GetAll()

  // render JSON reponse
  if err = render.RenderList(w, r, rs.newCourseListResponse(courses)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }
}

// CreateHandler is the enpoint for retrieving all courses if claim.root is true.
func (rs *CourseResource) CreateHandler(w http.ResponseWriter, r *http.Request) {
  // start from empty Request
  data := &courseRequest{}

  // parse JSON request into struct
  if err := render.Bind(r, data); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  // create course entry in database
  newCourse, err := rs.Stores.Course.Create(data.Course)
  if err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

  render.Status(r, http.StatusCreated)

  // return course information of created entry
  if err := render.Render(w, r, rs.newCourseResponse(newCourse)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

}

// GetHandler is the enpoint for retrieving a specific course.
func (rs *CourseResource) GetHandler(w http.ResponseWriter, r *http.Request) {
  // `course` is retrieved via middle-ware
  course, ok := r.Context().Value("course").(*model.Course)
  if !ok {
    render.Render(w, r, ErrInternalServerErrorWithDetails(errors.New("course context is missing")))
    return
  }

  // render JSON reponse
  if err := render.Render(w, r, rs.newCourseResponse(course)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

  render.Status(r, http.StatusOK)
}

// PatchHandler is the endpoint fro updating a specific course with given id.
func (rs *CourseResource) EditHandler(w http.ResponseWriter, r *http.Request) {
  // start from empty Request
  data := &courseRequest{
    Course: r.Context().Value("course").(*model.Course),
  }

  // parse JSON request into struct
  if err := render.Bind(r, data); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  // update database entry
  if err := rs.Stores.Course.Update(data.Course); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  // TODO(patwie): change StatusNoContent
  render.Status(r, http.StatusNoContent)
}

func (rs *CourseResource) DeleteHandler(w http.ResponseWriter, r *http.Request) {
  course := r.Context().Value("course").(*model.Course)

  // Warning: There is more to do! Currently we just dis-enroll all students,
  // remove all sheets and delete the course it self FROM THE DATABASE.
  // This does not remove gradings and the sheets or touches any file!

  // update database entry
  if err := rs.Stores.Course.Delete(course.ID); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  render.Status(r, http.StatusNoContent)
}

// IndexEnrollmentsHandler lists all enrolled users. The query can be refined by several
// URLparameters.
func (rs *CourseResource) IndexEnrollmentsHandler(w http.ResponseWriter, r *http.Request) {
  // /courses/1/enrollments?roles=0,1
  course := r.Context().Value("course").(*model.Course)

  // extract filters
  filterRoles := helper.StringArrayFromUrl(r, "roles", []string{"0", "1", "2"})
  filterFirstName := helper.StringFromUrl(r, "first_name", "%%")
  filterLastName := helper.StringFromUrl(r, "last_name", "%%")
  filterEmail := helper.StringFromUrl(r, "email", "%%")
  filterSubject := helper.StringFromUrl(r, "subject", "%%")
  filterLanguage := helper.StringFromUrl(r, "language", "%%")

  givenRole := r.Context().Value("course_role").(authorize.CourseRole)

  if givenRole == authorize.STUDENT {
    // students cannot query other students
    filterRoles = []string{"1", "2"}
  }

  enrolledUsers, err := rs.Stores.Course.EnrolledUsers(course,
    filterRoles, filterFirstName, filterLastName, filterEmail,
    filterSubject, filterLanguage,
  )
  if err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  // users and tutors will not see student number
  for k, _ := range enrolledUsers {
    enrolledUsers[k].StudentNumber = ""
  }

  // render JSON reponse
  if err = render.RenderList(w, r, rs.newEnrollmentListResponse(enrolledUsers)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

  render.Status(r, http.StatusOK)
}

// GetEnrollmentsHandler
// URL /api/v1/course/{course_id}/enrollments/{user_id}
// METHOD: GET
func (rs *CourseResource) GetEnrollmentsHandler(w http.ResponseWriter, r *http.Request) {
  // /courses/1/enrollments?roles=0,1
  course := r.Context().Value("course").(*model.Course)
  user := r.Context().Value("user").(*model.User)

  // find role in the course
  courseRole, err := rs.Stores.Course.RoleInCourse(user.ID, course.ID)
  if err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  resp := &enrollmentResponse{
    Role: int64(courseRole),
    User: user,
  }

  // render JSON reponse
  if err = render.Render(w, r, resp); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

  render.Status(r, http.StatusOK)
}

// ChangeRole
// URL /api/v1/course/{course_id}/enrollments/{user_id}
// METHOD: PUT
func (rs *CourseResource) ChangeRole(w http.ResponseWriter, r *http.Request) {
  // /courses/1/enrollments?roles=0,1

  course := r.Context().Value("course").(*model.Course)
  user := r.Context().Value("user").(*model.User)

  data := &changeRoleInCourseRequest{}
  // parse JSON request into struct
  if err := render.Bind(r, data); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  // update database entry
  if err := rs.Stores.Course.UpdateRole(course.ID, user.ID, data.Role); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  render.Status(r, http.StatusOK)
}

// EnrollHandler will enroll the current identity into the given course
func (rs *CourseResource) EnrollHandler(w http.ResponseWriter, r *http.Request) {
  course := r.Context().Value("course").(*model.Course)
  accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)

  // update database entry
  if err := rs.Stores.Course.Enroll(course.ID, accessClaims.LoginID); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  user, err := rs.Stores.User.Get(accessClaims.LoginID)
  if err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  resp := &enrollmentResponse{
    Role: 0,
    User: user,
  }

  render.Status(r, http.StatusCreated)

  if err := render.Render(w, r, resp); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

}

// DisenrollHandler will disenroll the current identity into the given course
func (rs *CourseResource) DisenrollHandler(w http.ResponseWriter, r *http.Request) {
  course := r.Context().Value("course").(*model.Course)
  accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)

  givenRole := r.Context().Value("course_role").(authorize.CourseRole)

  if givenRole == authorize.TUTOR {
    render.Render(w, r, ErrBadRequestWithDetails(errors.New("tutors cannot disenroll from a course")))
    return
  }

  // update database entry
  if err := rs.Stores.Course.Disenroll(course.ID, accessClaims.LoginID); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  render.Status(r, http.StatusNoContent)
}

// SendEmailHandler will send email to the entire course filtered by role.
func (rs *CourseResource) SendEmailHandler(w http.ResponseWriter, r *http.Request) {

  course := r.Context().Value("course").(*model.Course)

  accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)
  accessUser, _ := rs.Stores.User.Get(accessClaims.LoginID)

  data := &EmailRequest{}

  // parse JSON request into struct
  if err := render.Bind(r, data); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  // extract filters
  filterRoles := helper.StringArrayFromUrl(r, "roles", []string{"0", "1", "2"})
  filterFirstName := "%%"
  filterLastName := "%%"
  filterEmail := "%%"
  filterSubject := "%%"
  filterLanguage := "%%"

  recipients, err := rs.Stores.Course.EnrolledUsers(course,
    filterRoles, filterFirstName, filterLastName, filterEmail,
    filterSubject, filterLanguage,
  )

  if err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  for _, recipient := range recipients {
    // add sender identity
    msg := email.NewEmailFromUser(
      recipient.Email,
      data.Subject,
      data.Body,
      accessUser,
    )

    if err := email.DefaultMail.Send(msg); err != nil {
      render.Render(w, r, ErrInternalServerErrorWithDetails(err))
      return
    }
  }

}

// PointsHandler returns the point for the identity in a given course. This is
// intented to serve data for a plot.
func (rs *CourseResource) PointsHandler(w http.ResponseWriter, r *http.Request) {
  course := r.Context().Value("course").(*model.Course)
  accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)

  sheetPoints, err := rs.Stores.Course.PointsForUser(accessClaims.LoginID, course.ID)
  if err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  // resp := &SheetPointsResponse{SheetPoints: sheetPoints}

  if err := render.RenderList(w, r, newSheetPointsListResponse(sheetPoints)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

  render.Status(r, http.StatusOK)
}

// BidsHandler  Show your group bids or if allowed show all bids.
// url: /course/{course_id}/bids
// mwethod: GET
func (rs *CourseResource) BidsHandler(w http.ResponseWriter, r *http.Request) {
  course := r.Context().Value("course").(*model.Course)
  accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)

  givenRole := r.Context().Value("course_role").(authorize.CourseRole)

  var bids []model.GroupBid
  var err error

  if givenRole == authorize.TUTOR {
    // tutors see nothing
    render.Render(w, r, ErrBadRequestWithDetails(errors.New("tutors cannot have bids for a group in a course")))
    return

  }

  if givenRole == authorize.STUDENT {
    // students only see their own bids
    bids, err = rs.Stores.Group.GetBidsForCourseForUser(course.ID, accessClaims.LoginID)
  } else {
    // admins see all (to later setup the bid)
    bids, err = rs.Stores.Group.GetBidsForCourse(course.ID)
  }

  if err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  if err := render.RenderList(w, r, newGroupBidsListResponse(bids)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

  render.Status(r, http.StatusOK)
}

// .............................................................................
// Context middleware is used to load an Course object from
// the URL parameter `courseID` passed through as the request. In case
// the Course could not be found, we stop here and return a 404.
// We do NOT check whether the course is authorized to get this course.
func (rs *CourseResource) Context(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // TODO: check permission if inquirer of request is allowed to access this course
    // Should be done via another middleware
    var courseID int64
    var err error

    // try to get id from URL
    if courseID, err = strconv.ParseInt(chi.URLParam(r, "courseID"), 10, 64); err != nil {
      render.Render(w, r, ErrNotFound)
      return
    }

    // find specific course in database
    course, err := rs.Stores.Course.Get(courseID)
    if err != nil {
      render.Render(w, r, ErrNotFound)
      return
    }

    // serve next
    ctx := context.WithValue(r.Context(), "course", course)
    next.ServeHTTP(w, r.WithContext(ctx))
  })
}

func (rs *CourseResource) RoleContext(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    course := r.Context().Value("course").(*model.Course)
    accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)

    // find role in the course
    courseRole, err := rs.Stores.Course.RoleInCourse(accessClaims.LoginID, course.ID)
    if err != nil {
      render.Render(w, r, ErrBadRequestWithDetails(err))
      return
    }

    // serve next
    ctx := context.WithValue(r.Context(), "course_role", courseRole)
    next.ServeHTTP(w, r.WithContext(ctx))
  })
}
