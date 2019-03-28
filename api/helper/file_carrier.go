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

package helper

import (
  "crypto/sha256"
  "errors"
  "fmt"
  "io"
  "mime/multipart"
  "net/http"
  "os"
  "strconv"

  "github.com/spf13/viper"
)

// FaileCarrier is a unified way to handle uploads and downloads of different
// files.

type FileCategory int32

const (
  AvatarCategory      FileCategory = 0
  SheetCategory       FileCategory = 1
  PublicTestCategory  FileCategory = 2
  PrivateTestCategory FileCategory = 3
  MaterialCategory    FileCategory = 4
  SubmissionCategory  FileCategory = 5
)

type FileManager interface {
  WriteToBody(w http.ResponseWriter) error
  WriteToDisk(req multipart.File) error
  GetContentType() (string, error)
  Path(fallback bool) bool
  Delete() error
  Exists() bool
}

// FileHandle represents all information for file being uploaded or downloaded.
type FileHandle struct {
  Category   FileCategory
  ID         int64    // an unique identifier (e.g. from database)
  Extensions []string // some extra flag (ugly but works)
}

// NewAvatarFileHandle will handle user avatars. We support jpg only.
func NewAvatarFileHandle(userID int64) *FileHandle {
  return &FileHandle{
    Category:   AvatarCategory,
    ID:         userID,
    Extensions: []string{"jpg", "jpeg", "png"},
  }
}

// NewSheetFileHandle will handle exercise sheets (zip files).
func NewSheetFileHandle(ID int64) *FileHandle {
  return &FileHandle{
    Category:   SheetCategory,
    ID:         ID,
    Extensions: []string{"zip"},
  }
}

// NewPublicTestFileHandle will handle the testing framework for
// public unit tests (zip files).
func NewPublicTestFileHandle(ID int64) *FileHandle {
  return &FileHandle{
    Category:   PublicTestCategory,
    ID:         ID,
    Extensions: []string{"zip"},
  }
}

// NewPrivateTestFileHandle will handle the testing framework for
// private unit tests (zip files).
func NewPrivateTestFileHandle(ID int64) *FileHandle {
  return &FileHandle{
    Category:   PrivateTestCategory,
    ID:         ID,
    Extensions: []string{"zip"},
  }
}

// NewMaterialFileHandle will handle course slides or extra material (zip files).
func NewMaterialFileHandle(ID int64) *FileHandle {
  return &FileHandle{
    Category:   MaterialCategory,
    ID:         ID,
    Extensions: []string{"zip", "pdf"},
  }
}

// NewSubmissionFileHandle will handle homework/exercise submissiosn (zip files).
func NewSubmissionFileHandle(ID int64) *FileHandle {
  return &FileHandle{
    Category:   SubmissionCategory,
    ID:         ID,
    Extensions: []string{"zip"},
  }
}

// Path returns a path without checking if it exists.
func (f *FileHandle) Sha256() (string, error) {

  hnd, err := os.Open(f.Path())
  if err != nil {
    return "", err
  }
  defer hnd.Close()

  h := sha256.New()
  if _, err := io.Copy(h, hnd); err != nil {
    return "", err
  }

  return fmt.Sprintf("%x", h.Sum(nil)), nil

}

func (f *FileHandle) Path() string {
  switch f.Category {
  case AvatarCategory:

    for _, ext := range f.Extensions {
      path := fmt.Sprintf("%s/avatars/%s.%s", viper.GetString("uploads_dir"), strconv.FormatInt(f.ID, 10), ext)
      if fileExists(path) {
        return path
      }
    }
    return ""

  case SheetCategory:
    return fmt.Sprintf("%s/sheets/%s.zip", viper.GetString("uploads_dir"), strconv.FormatInt(f.ID, 10))

  case PublicTestCategory:
    return fmt.Sprintf("%s/tasks/%s-public.zip", viper.GetString("uploads_dir"), strconv.FormatInt(f.ID, 10))

  case PrivateTestCategory:
    return fmt.Sprintf("%s/tasks/%s-private.zip", viper.GetString("uploads_dir"), strconv.FormatInt(f.ID, 10))

  case MaterialCategory:

    for _, ext := range f.Extensions {
      path := fmt.Sprintf("%s/materials/%s.%s", viper.GetString("uploads_dir"), strconv.FormatInt(f.ID, 10), ext)
      if fileExists(path) {
        return path
      }
    }
    return ""

  case SubmissionCategory:
    return fmt.Sprintf("%s/submissions/%s.zip", viper.GetString("uploads_dir"), strconv.FormatInt(f.ID, 10))
  }
  return ""
}

// Exists checks if a file really exists.
func fileExists(path string) bool {
  if _, err := os.Stat(path); os.IsNotExist(err) {
    return false
  }

  return true
}

// Exists checks if a file really exists.
func (f *FileHandle) Exists() bool {
  return fileExists(f.Path())
}

// Delete deletes a file from disk.
func (f *FileHandle) Delete() error {
  return os.Remove(f.Path())
}

// GetContentType tries to predict the content type without reading the entire
// file. There are some issues with this function as it cannot distinguish
// between zip and octstream.
func (f *FileHandle) GetContentType() (string, error) {

  // Only the first 512 bytes are used to sniff the content type.
  buffer := make([]byte, 512)

  file, err := os.Open(f.Path())
  if err != nil {
    return "", err
  }
  defer file.Close()

  _, err = file.Read(buffer)
  if err != nil {
    return "", err
  }

  // Use the net/http package's handy DectectContentType function. Always returns a valid
  // content-type by returning "application/octet-stream" if no others seemed to match.
  contentType := http.DetectContentType(buffer)

  return contentType, nil
}

// WriteToBody will write a file from disk to the http reponse (download process)
func (f *FileHandle) WriteToBody(w http.ResponseWriter) error {

  // check if file exists
  file, err := os.Open(f.Path())
  if err != nil {
    return err
  }
  defer file.Close()

  // prepare header
  fileType, err := f.GetContentType()
  if err != nil {
    return err
  }
  w.Header().Set("Content-Type", fileType)

  // return file
  _, err = io.Copy(w, file)
  if err != nil {
    return err
  }

  return nil
}

// WriteToDisk will save uploads from a http request to the directory specified
// in the config.
func (f *FileHandle) WriteToDisk(r *http.Request, fieldName string) error {

  // receive data from post request
  if err := r.ParseMultipartForm(32 << 20); err != nil {
    return err
  }

  // we are interested in the field "avatar_data"
  file, handler, err := r.FormFile(fieldName)
  if err != nil {
    return err
  }
  defer file.Close()

  givenContentType := handler.Header["Content-Type"][0]

  path := f.Path()

  switch f.Category {
  case AvatarCategory:

    switch givenContentType {
    case "image/jpeg", "image/jpg", "image/png":

      if givenContentType == "image/png" {
        path = fmt.Sprintf("%s/avatars/%s.png", viper.GetString("uploads_dir"), strconv.FormatInt(f.ID, 10))
      } else {
        path = fmt.Sprintf("%s/avatars/%s.jpg", viper.GetString("uploads_dir"), strconv.FormatInt(f.ID, 10))
      }

    default:
      return errors.New(fmt.Sprintf("We support JPG/JPEG/PNG files only. But %s was given", givenContentType))

    }
  case SheetCategory,
    PublicTestCategory,
    PrivateTestCategory,
    SubmissionCategory:
    switch givenContentType {
    case "application/zip", "application/octet-stream":

    default:
      return errors.New(fmt.Sprintf("We support ZIP files only. But %s was given", givenContentType))

    }
  case MaterialCategory:
    switch givenContentType {
    case "application/zip", "application/octet-stream", "application/pdf":

      if givenContentType == "application/pdf" {
        path = fmt.Sprintf("%s/materials/%s.pdf", viper.GetString("uploads_dir"), strconv.FormatInt(f.ID, 10))

      } else {
        path = fmt.Sprintf("%s/materials/%s.zip", viper.GetString("uploads_dir"), strconv.FormatInt(f.ID, 10))

      }

    default:
      return errors.New(fmt.Sprintf("We support ZIP files only. But %s was given", givenContentType))

    }
  }
  // try to open new file
  hnd, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
  if err != nil {
    return err
  }
  defer hnd.Close()

  // copy file from request
  _, err = io.Copy(hnd, file)
  return err

}
