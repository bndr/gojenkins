// Copyright 2015 Vadim Kravcenko
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package gojenkins

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	createUserContext = "/securityRealm/createAccountByAdmin"
)

// User represents a Jenkins user account.
type User struct {
	Jenkins  *Jenkins
	UserName string
	FullName string
	Email    string
	ID       string
	Base     string
	Raw      *UserResponse
}

// Users is an alias for User kept for backward compatibility.
// Deprecated: use User instead.
type Users = User

// UserResponse is the JSON payload returned by the Jenkins user API.
type UserResponse struct {
	Class       string `json:"_class"`
	AbsoluteURL string `json:"absoluteUrl"`
	Description string `json:"description"`
	FullName    string `json:"fullName"`
	ID          string `json:"id"`
}

// PeopleUser is a single entry from the Jenkins people/asynchPeople API.
type PeopleUser struct {
	AbsoluteURL string `json:"absoluteUrl"`
	FullName    string `json:"fullName"`
	ID          string `json:"id"`
}

// PeopleProject is the project associated with a people API entry.
type PeopleProject struct {
	Class string `json:"_class"`
	Name  string `json:"name"`
	URL   string `json:"url"`
}

// PeopleEntry is one row in the Jenkins people/asynchPeople listing.
type PeopleEntry struct {
	LastChange int64         `json:"lastChange"`
	Project    PeopleProject `json:"project"`
	User       PeopleUser    `json:"user"`
}

// AllUserResponse is the JSON payload returned by /asynchPeople/.
type AllUserResponse struct {
	Class string        `json:"_class"`
	Users []PeopleEntry `json:"users"`
}

// AllUsers is a collection of Jenkins users returned by GetAllUsers.
type AllUsers struct {
	Jenkins *Jenkins
	Base    string
	Raw     *AllUserResponse
}

// ErrUser is returned when a user API operation fails.
type ErrUser struct {
	Message string
}

func (e *ErrUser) Error() string {
	return e.Message
}

func userAPIPath(userName string) string {
	return "/user/" + url.PathEscape(userName)
}

func deleteUserPath(userName string) string {
	return "/securityRealm/user/" + url.PathEscape(userName) + "/doDelete"
}

// CreateUser creates a new Jenkins account.
func (j *Jenkins) CreateUser(ctx context.Context, userName, password, fullName, email string) (User, error) {
	user := User{
		Jenkins:  j,
		UserName: userName,
		FullName: fullName,
		Email:    email,
		Base:     userAPIPath(userName),
		Raw:      new(UserResponse),
	}

	data := url.Values{}
	data.Set("username", userName)
	data.Set("password1", password)
	data.Set("password2", password)
	data.Set("fullname", fullName)
	data.Set("email", email)

	response, err := j.Requester.Post(ctx, createUserContext, bytes.NewBufferString(data.Encode()), nil, nil)
	if err != nil {
		return user, err
	}
	if response.StatusCode != http.StatusOK {
		return user, &ErrUser{
			Message: fmt.Sprintf("error creating user. Status is %d", response.StatusCode),
		}
	}
	return user, nil
}

// DeleteUser deletes a Jenkins account.
func (j *Jenkins) DeleteUser(ctx context.Context, userName string) error {
	data := url.Values{}
	data.Set("Submit", "Yes")

	response, err := j.Requester.Post(ctx, deleteUserPath(userName), bytes.NewBufferString(data.Encode()), nil, nil)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return &ErrUser{
			Message: fmt.Sprintf("error deleting user. Status is %d", response.StatusCode),
		}
	}
	return nil
}

// Delete deletes a Jenkins account.
func (u *User) Delete(ctx context.Context) error {
	return u.Jenkins.DeleteUser(ctx, u.UserName)
}

// GetUser retrieves information about a Jenkins user.
func (j *Jenkins) GetUser(ctx context.Context, userName string) (*User, error) {
	user := &User{
		Jenkins:  j,
		UserName: userName,
		Raw:      new(UserResponse),
		Base:     userAPIPath(userName),
	}

	status, err := user.Poll(ctx)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, &ErrUser{
			Message: fmt.Sprintf("user not found. Status is %d", status),
		}
	}
	return user, nil
}

// GetAllUsers retrieves information about all Jenkins users.
// This operation may take a long time on large Jenkins instances.
func (j *Jenkins) GetAllUsers(ctx context.Context) (*AllUsers, error) {
	allUsers := &AllUsers{Jenkins: j, Raw: new(AllUserResponse), Base: "/asynchPeople/"}

	status, err := allUsers.Poll(ctx)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, errors.New(strconv.Itoa(status))
	}
	return allUsers, nil
}

// Poll retrieves user information and updates exported fields from the response.
func (u *User) Poll(ctx context.Context) (int, error) {
	response, err := u.Jenkins.Requester.GetJSON(ctx, u.Base, u.Raw, nil)
	if err != nil {
		return 0, err
	}
	if u.Raw != nil {
		u.ID = u.Raw.ID
		u.FullName = u.Raw.FullName
		if u.UserName == "" {
			u.UserName = u.Raw.ID
		}
	}
	return response.StatusCode, nil
}

// Poll retrieves all user information.
func (u *AllUsers) Poll(ctx context.Context) (int, error) {
	response, err := u.Jenkins.Requester.GetJSON(ctx, u.Base, u.Raw, nil)
	if err != nil {
		return 0, err
	}
	if u.Raw != nil {
		for i := range u.Raw.Users {
			entry := &u.Raw.Users[i]
			if entry.User.ID == "" {
				entry.User.ID = userIDFromAbsoluteURL(entry.User.AbsoluteURL)
			}
		}
	}
	return response.StatusCode, nil
}

func userIDFromAbsoluteURL(absoluteURL string) string {
	absoluteURL = strings.TrimSuffix(absoluteURL, "/")
	if absoluteURL == "" {
		return ""
	}
	parts := strings.Split(absoluteURL, "/")
	return parts[len(parts)-1]
}
