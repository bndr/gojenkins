package gojenkins

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestJenkins(handler http.Handler) (*Jenkins, *httptest.Server) {
	server := httptest.NewServer(handler)
	jenkins := CreateJenkins(server.Client(), server.URL, "admin", "admin")
	jenkins.initLoggers()
	return jenkins, server
}

func TestCreateUserEncodesFormValues(t *testing.T) {
	var gotBody string
	var gotContentType string

	jenkins, server := newTestJenkins(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "crumbIssuer") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"crumbRequestField":"Jenkins-Crumb","crumb":"abc"}`))
			return
		}
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		gotBody = string(body)
		gotContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx := context.Background()
	user, err := jenkins.CreateUser(ctx, "user&name", "p@ss=word", "Full Name", "a+b@example.com")
	require.NoError(t, err)
	assert.Equal(t, "user&name", user.UserName)
	assert.Contains(t, gotContentType, "application/x-www-form-urlencoded")

	values, err := url.ParseQuery(gotBody)
	require.NoError(t, err)
	assert.Equal(t, "user&name", values.Get("username"))
	assert.Equal(t, "p@ss=word", values.Get("password1"))
	assert.Equal(t, "p@ss=word", values.Get("password2"))
	assert.Equal(t, "Full Name", values.Get("fullname"))
	assert.Equal(t, "a+b@example.com", values.Get("email"))
}

func TestDeleteUserUsesPathEscapeAndContext(t *testing.T) {
	var gotPath string

	jenkins, server := newTestJenkins(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "crumbIssuer") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		gotPath = r.URL.EscapedPath()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx := context.Background()
	err := jenkins.DeleteUser(ctx, "user name")
	require.NoError(t, err)
	assert.Equal(t, "/securityRealm/user/user%20name/doDelete", gotPath)
}

func TestGetUserPopulatesFieldsAndChecksStatus(t *testing.T) {
	jenkins, server := newTestJenkins(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/user/alice/api/json", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(UserResponse{
			ID:          "alice",
			FullName:    "Alice Example",
			AbsoluteURL: "http://localhost/user/alice",
			Description: "tester",
		})
	}))
	defer server.Close()

	ctx := context.Background()
	user, err := jenkins.GetUser(ctx, "alice")
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, "alice", user.UserName)
	assert.Equal(t, "alice", user.ID)
	assert.Equal(t, "Alice Example", user.FullName)
	assert.Equal(t, "tester", user.Raw.Description)
}

func TestGetUserNotFound(t *testing.T) {
	jenkins, server := newTestJenkins(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	_, err := jenkins.GetUser(context.Background(), "missing")
	require.Error(t, err)
	var userErr *ErrUser
	require.ErrorAs(t, err, &userErr)
}

func TestGetAllUsersDerivesMissingIDs(t *testing.T) {
	jenkins, server := newTestJenkins(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/asynchPeople/api/json", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"_class":"hudson.model.View$AsynchPeople",
			"users":[{
				"lastChange":1,
				"project":{"_class":"hudson.model.FreeStyleProject","name":"job","url":"http://localhost/job/job/"},
				"user":{"absoluteUrl":"http://localhost/user/bob","fullName":"Bob"}
			}]
		}`))
	}))
	defer server.Close()

	allUsers, err := jenkins.GetAllUsers(context.Background())
	require.NoError(t, err)
	require.Len(t, allUsers.Raw.Users, 1)
	assert.Equal(t, "bob", allUsers.Raw.Users[0].User.ID)
	assert.Equal(t, "Bob", allUsers.Raw.Users[0].User.FullName)
}

func TestUserDeleteUsesContext(t *testing.T) {
	var deleted bool
	jenkins, server := newTestJenkins(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "crumbIssuer") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if strings.Contains(r.URL.Path, "doDelete") {
			deleted = true
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	user := User{Jenkins: jenkins, UserName: "carol"}
	require.NoError(t, user.Delete(context.Background()))
	assert.True(t, deleted)
}

func TestGenerateAPITokenEncodesName(t *testing.T) {
	var gotBody string
	jenkins, server := newTestJenkins(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "crumbIssuer") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		gotBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(APITokenGenerateResponse{
			Status: "ok",
			Data: APIToken{
				Name:  "token=name",
				UUID:  "uuid-1",
				Value: "secret",
			},
		})
	}))
	defer server.Close()

	token, err := jenkins.GenerateAPIToken(context.Background(), "token=name")
	require.NoError(t, err)
	assert.Equal(t, "secret", token.Value)
	assert.Equal(t, jenkins, token.Jenkins)

	values, err := url.ParseQuery(gotBody)
	require.NoError(t, err)
	assert.Equal(t, "token=name", values.Get("newTokenName"))
}

func TestSetCrumbDoesNotPanicWhenCrumbIssuerFails(t *testing.T) {
	jenkins, server := newTestJenkins(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "crumbIssuer") {
			http.Error(w, "gone", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ar := NewAPIRequest("POST", "/safeRestart", nil)
	err := jenkins.Requester.SetCrumb(context.Background(), ar)
	require.NoError(t, err)
}

func TestReadJSONResponseHandlesEmptyBody(t *testing.T) {
	jenkins, server := newTestJenkins(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "crumbIssuer") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	resp, err := jenkins.Requester.Post(context.Background(), "/safeRestart", strings.NewReader(""), nil, nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadJSONResponseReturnsDecodeError(t *testing.T) {
	jenkins, server := newTestJenkins(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":`))
	}))
	defer server.Close()

	var raw UserResponse
	_, err := jenkins.Requester.GetJSON(context.Background(), "/user/bad", &raw, nil)
	require.Error(t, err)
}
