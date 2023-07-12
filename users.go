package gojenkins

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"encoding/json"
	"errors"
)

const (
	createUserContext = "/securityRealm/createAccountByAdmin"
)

// User is a Jenkins account
type User struct {
	Jenkins  *Jenkins
	UserName string
	FullName string
	Email    string
}

// ErrUser occurs when there is error creating or revoking Jenkins users
type ErrUser struct {
	Message string
}

func (e *ErrUser) Error() string {
	return e.Message
}

// CreateUser creates a new Jenkins account
func (j *Jenkins) CreateUser(ctx context.Context, userName, password, fullName, email string) (User, error) {
	user := User{
		// Set Jenkins client pointer to be able to delete user later
		Jenkins:  j,
		UserName: userName,
		FullName: fullName,
		Email:    email,
	}
	payload := "username=" + userName + "&password1=" + password + "&password2=" + password + "&fullname=" + fullName + "&email=" + email
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

// DeleteUser deletes a Jenkins account
func (j *Jenkins) DeleteUser(ctx context.Context, userName string) error {
	deleteContext := "/securityRealm/user/" + userName + "/doDelete"
	payload := "Submit=Yes"
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

// Delete deletes a Jenkins account
func (u *User) Delete() error {
	return u.Jenkins.DeleteUser(context.Background(), u.UserName)
}

func (j *Jenkins) GetUser(ctx context.Context, userName string) (User, error) {
	getUserContext := "/securityRealm/user/" + userName + "/api/json"
	response, err := j.Requester.Get(ctx, getUserContext, nil, nil)
	if err != nil {
		return User{}, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return User{}, errors.New(fmt.Sprintf("error retrieving user. Status is %d", response.StatusCode))
	}

	var userAPIResp struct {
		UserName string `json:"id"`
		FullName string `json:"fullName"`
		Email    string `json:"email"`
	}

	err = json.NewDecoder(response.Body).Decode(&userAPIResp)
	if err != nil {
		return User{}, err
	}

	user := User{
		Jenkins:  j,
		UserName: userAPIResp.UserName,
		FullName: userAPIResp.FullName,
		Email:    userAPIResp.Email,
	}

	return user, nil
}


// GetAllUsers retrieves information about all Jenkins users
func (j *Jenkins) GetAllUsers(ctx context.Context) ([]User, error) {
	getAllUsersContext := "/jenkins/asynchPeople/"
	response, err := j.Requester.Get(ctx, getAllUsersContext, nil, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("error retrieving users. Status is %d", response.StatusCode))
	}

	var usersAPIResp []struct {
		UserName string `json:"id"`
		FullName string `json:"fullName"`
		Email    string `json:"email"`
	}

	err = json.NewDecoder(response.Body).Decode(&usersAPIResp)
	if err != nil {
		return nil, err
	}

	users := make([]User, len(usersAPIResp))
	for i, userAPIResp := range usersAPIResp {
		users[i] = User{
			Jenkins:  j,
			UserName: userAPIResp.UserName,
			FullName: userAPIResp.FullName,
			Email:    userAPIResp.Email,
		}
	}

	return users, nil
}