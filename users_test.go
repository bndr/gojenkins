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

// TestGetUserSuccess tests successful user retrieval
func TestGetUserSuccess(t *testing.T) {
	mock := &MockRequester{
		GetJSONFunc: func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
			user, ok := response.(*userResponse)
			assert.True(t, ok)
			user.ID = "testuser"
			user.FullName = "Test User"
			user.Property = append(user.Property, struct {
				Class   string `json:"_class"`
				Address string `json:"address"`
			}{
				Class:   "hudson.tasks.Mailer$UserProperty",
				Address: "test@example.com",
			})
			return &http.Response{StatusCode: http.StatusOK}, nil
		},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	user, err := jenkins.GetUser(context.Background(), "testuser")

	assert.NoError(t, err)
	assert.Equal(t, "testuser", user.UserName)
	assert.Equal(t, "Test User", user.FullName)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, jenkins, user.Jenkins)
	assert.Equal(t, "/user/testuser", mock.lastEndpoint)
}

// TestGetUserEscapesUsername tests that user IDs are escaped in the API path
func TestGetUserEscapesUsername(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: http.StatusOK},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	_, err := jenkins.GetUser(context.Background(), "folder/user")

	assert.NoError(t, err)
	assert.Equal(t, "/user/folder%2Fuser", mock.lastEndpoint)
}

// TestGetUserWithoutEmail tests successful user retrieval without a Mailer property
func TestGetUserWithoutEmail(t *testing.T) {
	mock := &MockRequester{
		GetJSONFunc: func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
			user, ok := response.(*userResponse)
			assert.True(t, ok)
			user.ID = "noemail"
			user.FullName = "No Email"
			return &http.Response{StatusCode: http.StatusOK}, nil
		},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	user, err := jenkins.GetUser(context.Background(), "noemail")

	assert.NoError(t, err)
	assert.Equal(t, "noemail", user.UserName)
	assert.Equal(t, "No Email", user.FullName)
	assert.Empty(t, user.Email)
}

// TestGetUserError tests user retrieval with HTTP error
func TestGetUserError(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: http.StatusNotFound},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	_, err := jenkins.GetUser(context.Background(), "missing")

	assert.Error(t, err)
	errUser, ok := err.(*ErrUser)
	assert.True(t, ok)
	assert.Contains(t, errUser.Message, "404")
}

// TestGetUserNetworkError tests user retrieval with network error
func TestGetUserNetworkError(t *testing.T) {
	mock := &MockRequester{
		err: errors.New("connection refused"),
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	_, err := jenkins.GetUser(context.Background(), "testuser")

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
