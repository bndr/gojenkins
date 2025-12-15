package gojenkins

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestErrUser verifies ErrUser implements error interface
func TestErrUser(t *testing.T) {
	err := &ErrUser{Message: "user creation failed"}

	assert.Error(t, err)
	assert.Equal(t, "user creation failed", err.Error())
}

// TestCreateUserSuccess tests successful user creation
func TestCreateUserSuccess(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: http.StatusOK},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	ctx := context.Background()
	user, err := jenkins.CreateUser(ctx, "newuser", "password123", "New User", "new@example.com")

	assert.NoError(t, err)
	assert.Equal(t, "newuser", user.UserName)
	assert.Equal(t, "New User", user.FullName)
	assert.Equal(t, "new@example.com", user.Email)
	assert.Equal(t, jenkins, user.Jenkins)
	assert.Contains(t, mock.lastEndpoint, "/securityRealm/createAccountByAdmin")
}

// TestCreateUserError tests user creation with HTTP error
func TestCreateUserError(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: http.StatusForbidden},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	ctx := context.Background()
	_, err := jenkins.CreateUser(ctx, "newuser", "password123", "New User", "new@example.com")

	assert.Error(t, err)
	errUser, ok := err.(*ErrUser)
	assert.True(t, ok)
	assert.Contains(t, errUser.Message, "403")
}

// TestCreateUserNetworkError tests user creation with network error
func TestCreateUserNetworkError(t *testing.T) {
	mock := &MockRequester{
		err: errors.New("connection refused"),
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	ctx := context.Background()
	_, err := jenkins.CreateUser(ctx, "newuser", "password123", "New User", "new@example.com")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

// TestDeleteUserSuccess tests successful user deletion
func TestDeleteUserSuccess(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: http.StatusOK},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	ctx := context.Background()
	err := jenkins.DeleteUser(ctx, "olduser")

	assert.NoError(t, err)
	assert.Contains(t, mock.lastEndpoint, "/securityRealm/user/olduser/doDelete")
}

// TestDeleteUserError tests user deletion with HTTP error
func TestDeleteUserError(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: http.StatusNotFound},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	ctx := context.Background()
	err := jenkins.DeleteUser(ctx, "nonexistent")

	assert.Error(t, err)
	errUser, ok := err.(*ErrUser)
	assert.True(t, ok)
	assert.Contains(t, errUser.Message, "404")
}

// TestDeleteUserNetworkError tests user deletion with network error
func TestDeleteUserNetworkError(t *testing.T) {
	mock := &MockRequester{
		err: errors.New("timeout"),
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	ctx := context.Background()
	err := jenkins.DeleteUser(ctx, "testuser")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

// TestUserDeleteMethod tests the Delete method on User struct
func TestUserDeleteMethod(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: http.StatusOK},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	user := User{
		Jenkins:  jenkins,
		UserName: "testuser",
		FullName: "Test User",
		Email:    "test@example.com",
	}

	err := user.Delete()

	assert.NoError(t, err)
	assert.Contains(t, mock.lastEndpoint, "/securityRealm/user/testuser/doDelete")
}

// TestUserDeleteMethodError tests the Delete method with error
func TestUserDeleteMethodError(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: http.StatusInternalServerError},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	user := User{
		Jenkins:  jenkins,
		UserName: "testuser",
	}

	err := user.Delete()

	assert.Error(t, err)
}

// TestCreateUserContext tests the constant value
func TestCreateUserContext(t *testing.T) {
	assert.Equal(t, "/securityRealm/createAccountByAdmin", createUserContext)
}
