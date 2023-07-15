package gojenkins

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

const (
	createUserContext = "/securityRealm/createAccountByAdmin"
)

type User struct {
	Jenkins  *Jenkins
	UserName string
	FullName string
	Email    string
	Raw      *UserResponse
}

type Users struct {
	Jenkins  *Jenkins
	UserName string
	FullName string
	Email    string
	ID       string
	Base     string
	Raw      *UserResponse
}

type UserResponse struct {
	Class       string `json:"_class"`
	AbsoluteURL string `json:"absoluteUrl"`
	Description string `json:"description"`
	FullName    string `json:"fullName"`
	ID          string `json:"id"`
}

type AllUserResponse struct {
	Class string `json:"_class"`
	Users []struct {
		LastChange int64 `json:"lastChange"`
		Project    struct {
			Class string `json:"_class"`
			Name  string `json:"name"`
			URL   string `json:"url"`
		} `json:"project"`
		User struct {
			AbsoluteURL string `json:"absoluteUrl"`
			FullName    string `json:"fullName"`
		} `json:"user"`
	} `json:"users"`
}

type AllUsers struct {
	Jenkins *Jenkins
	Base    string
	Raw     *AllUserResponse
}

type ErrUser struct {
	Message string
}

func (e *ErrUser) Error() string {
	return e.Message
}

// CreateUser creates a new Jenkins account.
func (j *Jenkins) CreateUser(ctx context.Context, userName, password, fullName, email string) (User, error) {
	user := User{
		Jenkins:  j,
		UserName: userName,
		FullName: fullName,
		Email:    email,
	}

	// Create the payload string
	payload := fmt.Sprintf("username=%s&password1=%s&password2=%s&fullname=%s&email=%s", userName, password, password, fullName, email)

	// Send the POST request to create the user
	response, err := j.Requester.Post(ctx, createUserContext, strings.NewReader(payload), nil, nil)
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
	deleteContext := "/securityRealm/user/" + userName + "/doDelete"
	payload := "Submit=Yes"

	// Send the POST request to delete the user
	response, err := j.Requester.Post(ctx, deleteContext, strings.NewReader(payload), nil, nil)
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
func (u *User) Delete() error {
	return u.Jenkins.DeleteUser(context.Background(), u.UserName)
}

// GetUser retrieves information about a Jenkins user.
func (j *Jenkins) GetUser(ctx context.Context, userName string) (*Users, error) {
	userInfo := Users{Jenkins: j, Raw: new(UserResponse), Base: "/user/" + userName}

	// Poll for user information
	_, err := userInfo.Poll(ctx)
	if err != nil {
		return nil, err
	}
	return &userInfo, nil
}

// GetAllUsers retrieves information about all Jenkins users.
// This operation may take a lot of time.
func (j *Jenkins) GetAllUsers(ctx context.Context) (*AllUsers, error) {
	allUsers := AllUsers{Jenkins: j, Raw: new(AllUserResponse), Base: "/asynchPeople/"}

	// Poll for all user information
	_, err := allUsers.Poll(ctx)
	if err != nil {
		return nil, err
	}
	return &allUsers, nil
}

// Poll retrieves user information.
func (u *Users) Poll(ctx context.Context) (int, error) {
	response, err := u.Jenkins.Requester.GetJSON(ctx, u.Base, u.Raw, nil)
	if err != nil {
		return 0, err
	}
	return response.StatusCode, nil
}

// Poll retrieves all user information.
func (u *AllUsers) Poll(ctx context.Context) (int, error) {
	response, err := u.Jenkins.Requester.GetJSON(ctx, u.Base, u.Raw, nil)
	if err != nil {
		return 0, err
	}
	return response.StatusCode, nil
}
