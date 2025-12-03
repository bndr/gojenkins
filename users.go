package gojenkins

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

const (
	createUserContext = "/securityRealm/createAccountByAdmin"
	listUserContext   = "/securityRealm"
)

type securityRealmUserList struct {
	Users securityRealmUsers `json:"users"`
}

type securityRealmUsers struct {
	Users []securityRealmUserNode `json:"users"`
}

type securityRealmUserNode struct {
	User securityRealmUser `json:"user"`
}

type securityRealmUser struct {
	ID         string                    `json:"id"`
	FullName   string                    `json:"fullName"`
	Properties []securityRealmUserConfig `json:"property"`
}

type securityRealmUserConfig struct {
	Class   string `json:"_class"`
	Address string `json:"address"`
}

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

// ListUsers returns all Jenkins accounts that the security realm exposes.
func (j *Jenkins) ListUsers(ctx context.Context) ([]User, error) {
	resp := new(securityRealmUserList)
	query := map[string]string{"depth": "1"}
	_, err := j.Requester.GetJSON(ctx, listUserContext, resp, query)
	if err != nil {
		return nil, err
	}

	if len(resp.Users.Users) == 0 {
		return []User{}, nil
	}

	users := make([]User, 0, len(resp.Users.Users))
	for _, item := range resp.Users.Users {
		user := User{
			Jenkins:  j,
			UserName: item.User.ID,
			FullName: item.User.FullName,
			Email:    extractUserEmail(item.User.Properties),
		}
		users = append(users, user)
	}

	return users, nil
}

func extractUserEmail(props []securityRealmUserConfig) string {
	for _, prop := range props {
		if prop.Class == "hudson.tasks.Mailer$UserProperty" && prop.Address != "" {
			return prop.Address
		}
	}
	return ""
}
