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

package shared

type SubmissionAMQPWorkerRequest struct {
	SubmissionID      int64  `json:"submission_id"`
	AccessToken       string `json:"access_token"`
	FrameworkFileURL  string `json:"framework_file_url"`
	SubmissionFileURL string `json:"submission_file_url"`
	ResultEndpointURL string `json:"result_endpoint_url"`
	DockerImage       string `json:"docker_image"`
	Sha256            string `json:"sha_256"`
}

type SubmissionWorkerResponse struct {
	Log    string `json:"log"`
	Status int    `json:"status"`
}
