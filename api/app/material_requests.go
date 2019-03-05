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
	"errors"
	"net/http"

	"github.com/cgtuebingen/infomark-backend/model"
)

// MaterialRequest is the request payload for Material management.
type MaterialRequest struct {
	*model.Material
	ProtectedID int64 `json:"id"`
}

// Bind preprocesses a MaterialRequest.
func (body *MaterialRequest) Bind(r *http.Request) error {

	if body.Material == nil {
		return errors.New("missing \"materials\" data")
	}

	// Sending the id via request-body is invalid.
	// The id should be submitted in the url.
	body.ProtectedID = 0

	return body.Material.Validate()
}