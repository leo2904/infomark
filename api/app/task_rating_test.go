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
  "encoding/json"
  "net/http"
  "testing"

  "github.com/cgtuebingen/infomark-backend/database"
  "github.com/cgtuebingen/infomark-backend/email"
  "github.com/franela/goblin"
)

func TestTaskRating(t *testing.T) {
  g := goblin.Goblin(t)
  email.DefaultMail = email.VoidMail

  tape := &Tape{}

  var stores *Stores

  g.Describe("TaskRating", func() {

    g.BeforeEach(func() {
      tape.BeforeEach()
      stores = NewStores(tape.DB)
    })

    g.It("Should get own rating", func() {
      userID := int64(112)
      taskID := int64(1)

      givenRating, err := stores.Task.GetRatingOfTaskByUser(taskID, userID)
      g.Assert(err).Equal(nil)

      w := tape.GetWithClaims("/api/v1/courses/1/tasks/1/ratings", userID, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      task_rating_actual := &TaskRatingResponse{}
      err = json.NewDecoder(w.Body).Decode(task_rating_actual)
      g.Assert(err).Equal(nil)

      g.Assert(task_rating_actual.OwnRating).Equal(givenRating.Rating)
      g.Assert(task_rating_actual.TaskID).Equal(taskID)

      // update rating (mock had rating 2)
      w = tape.PostWithClaims("/api/v1/courses/1/tasks/1/ratings", H{"rating": 4}, userID, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      // new query
      w = tape.GetWithClaims("/api/v1/courses/1/tasks/1/ratings", userID, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      task_rating_actual2 := &TaskRatingResponse{}
      err = json.NewDecoder(w.Body).Decode(task_rating_actual2)
      g.Assert(err).Equal(nil)

      g.Assert(task_rating_actual2.OwnRating).Equal(4)
      g.Assert(task_rating_actual2.TaskID).Equal(taskID)
    })

    g.It("Should create own rating", func() {
      userID := int64(112)
      taskID := int64(1)

      // delete and create (see mock.py)
      prevRatingModel, err := stores.Task.GetRatingOfTaskByUser(taskID, userID)
      g.Assert(err).Equal(nil)
      database.Delete(tape.DB, "task_ratings", prevRatingModel.ID)

      w := tape.GetWithClaims("/api/v1/courses/1/tasks/1/ratings", userID, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      task_rating_actual3 := &TaskRatingResponse{}
      err = json.NewDecoder(w.Body).Decode(task_rating_actual3)
      g.Assert(err).Equal(nil)

      g.Assert(task_rating_actual3.OwnRating).Equal(0)
      g.Assert(task_rating_actual3.TaskID).Equal(taskID)

      // update rating (mock had rating 2)
      w = tape.PostWithClaims("/api/v1/courses/1/tasks/1/ratings", H{"rating": 4}, userID, false)
      g.Assert(w.Code).Equal(http.StatusCreated)

      // new query
      w = tape.GetWithClaims("/api/v1/courses/1/tasks/1/ratings", userID, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      task_rating_actual2 := &TaskRatingResponse{}
      err = json.NewDecoder(w.Body).Decode(task_rating_actual2)
      g.Assert(err).Equal(nil)

      g.Assert(task_rating_actual2.OwnRating).Equal(4)
      g.Assert(task_rating_actual2.TaskID).Equal(taskID)
    })

    g.AfterEach(func() {
      tape.AfterEach()
    })

  })

}