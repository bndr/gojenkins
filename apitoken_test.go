package gojenkins

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestErrAPIToken verifies ErrAPIToken implements error interface
func TestErrAPIToken(t *testing.T) {
	err := &ErrAPIToken{Message: "token creation failed"}

	assert.Error(t, err)
	assert.Equal(t, "token creation failed", err.Error())
}

// TestGenerateAPITokenSuccess tests successful token generation
func TestGenerateAPITokenSuccess(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: http.StatusOK},
		PostFunc: func(ctx context.Context, endpoint string, payload io.Reader, response interface{}, query map[string]string) (*http.Response, error) {
			if apiResp, ok := response.(*APITokenGenerateResponse); ok {
				apiResp.Status = "ok"
				apiResp.Data = APIToken{
					Name:  "my-token",
					UUID:  "generated-uuid",
					Value: "generated-value",
				}
			}
			return &http.Response{StatusCode: http.StatusOK}, nil
		},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	ctx := context.Background()
	token, err := jenkins.GenerateAPIToken(ctx, "my-token")

	assert.NoError(t, err)
	assert.Equal(t, "my-token", token.Name)
	assert.Equal(t, "generated-uuid", token.UUID)
	assert.Equal(t, "generated-value", token.Value)
	assert.Equal(t, jenkins, token.Jenkins)
}

// TestGenerateAPITokenError tests token generation with HTTP error
func TestGenerateAPITokenError(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: http.StatusForbidden},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	ctx := context.Background()
	_, err := jenkins.GenerateAPIToken(ctx, "my-token")

	assert.Error(t, err)
	errToken, ok := err.(*ErrAPIToken)
	assert.True(t, ok)
	assert.Contains(t, errToken.Message, "403")
}

// TestGenerateAPITokenNetworkError tests token generation with network error
func TestGenerateAPITokenNetworkError(t *testing.T) {
	mock := &MockRequester{
		err: errors.New("connection refused"),
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	ctx := context.Background()
	_, err := jenkins.GenerateAPIToken(ctx, "my-token")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

// TestRevokeAPITokenSuccess tests successful token revocation
func TestRevokeAPITokenSuccess(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: http.StatusOK},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	ctx := context.Background()
	err := jenkins.RevokeAPIToken(ctx, "token-uuid-123")

	assert.NoError(t, err)
	assert.Contains(t, mock.lastEndpoint, "/revoke")
}

// TestRevokeAPITokenError tests token revocation with HTTP error
func TestRevokeAPITokenError(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: http.StatusNotFound},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	ctx := context.Background()
	err := jenkins.RevokeAPIToken(ctx, "nonexistent-uuid")

	assert.Error(t, err)
	errToken, ok := err.(*ErrAPIToken)
	assert.True(t, ok)
	assert.Contains(t, errToken.Message, "404")
}

// TestRevokeAPITokenNetworkError tests token revocation with network error
func TestRevokeAPITokenNetworkError(t *testing.T) {
	mock := &MockRequester{
		err: errors.New("timeout"),
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	ctx := context.Background()
	err := jenkins.RevokeAPIToken(ctx, "token-uuid")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

// TestRevokeAllAPITokensSuccess tests revoking all tokens
func TestRevokeAllAPITokensSuccess(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: http.StatusOK},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	ctx := context.Background()
	err := jenkins.RevokeAllAPITokens(ctx)

	assert.NoError(t, err)
	assert.Contains(t, mock.lastEndpoint, "/revokeAll")
}

// TestRevokeAllAPITokensError tests revoking all tokens with error
func TestRevokeAllAPITokensError(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: http.StatusForbidden},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	ctx := context.Background()
	err := jenkins.RevokeAllAPITokens(ctx)

	assert.Error(t, err)
	errToken, ok := err.(*ErrAPIToken)
	assert.True(t, ok)
	assert.Contains(t, errToken.Message, "403")
}

// TestRevokeAllAPITokensNetworkError tests revoking all tokens with network error
func TestRevokeAllAPITokensNetworkError(t *testing.T) {
	mock := &MockRequester{
		err: errors.New("network error"),
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	ctx := context.Background()
	err := jenkins.RevokeAllAPITokens(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "network error")
}

// TestAPITokenRevokeMethod tests the Revoke method on APIToken struct
func TestAPITokenRevokeMethod(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: http.StatusOK},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	token := APIToken{
		Jenkins: jenkins,
		Name:    "test-token",
		UUID:    "token-uuid-456",
		Value:   "token-value",
	}

	err := token.Revoke()

	assert.NoError(t, err)
	assert.Contains(t, mock.lastEndpoint, "/revoke")
}

// TestAPITokenRevokeMethodError tests the Revoke method with error
func TestAPITokenRevokeMethodError(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: http.StatusInternalServerError},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	token := APIToken{
		Jenkins: jenkins,
		UUID:    "token-uuid-456",
	}

	err := token.Revoke()

	assert.Error(t, err)
}

// TestAPITokenConstants tests the API token URL constants
func TestAPITokenConstants(t *testing.T) {
	assert.Equal(t, "/me/descriptorByName/jenkins.security.ApiTokenProperty", apiTokenBaseContext)
	assert.Equal(t, apiTokenBaseContext+"/generateNewToken", generateAPITokenURL)
	assert.Equal(t, apiTokenBaseContext+"/revoke", revokeAPITokenURL)
	assert.Equal(t, apiTokenBaseContext+"/revokeAll", revokeAllAPITokensURL)
}

// TestAPITokenJSONMarshaling tests JSON field tags
func TestAPITokenJSONMarshaling(t *testing.T) {
	token := APIToken{
		Name:  "test-token",
		UUID:  "uuid-123",
		Value: "value-456",
	}

	data, err := json.Marshal(token)
	assert.NoError(t, err)

	jsonStr := string(data)
	assert.Contains(t, jsonStr, "tokenName")
	assert.Contains(t, jsonStr, "tokenUuid")
	assert.Contains(t, jsonStr, "tokenValue")
}

// TestAPITokenGenerateResponseJSONUnmarshaling tests JSON unmarshaling
func TestAPITokenGenerateResponseJSONUnmarshaling(t *testing.T) {
	jsonData := `{
		"status": "ok",
		"data": {
			"tokenName": "my-api-token",
			"tokenUuid": "abc-123",
			"tokenValue": "secret-value"
		}
	}`

	var response APITokenGenerateResponse
	err := json.Unmarshal([]byte(jsonData), &response)

	assert.NoError(t, err)
	assert.Equal(t, "ok", response.Status)
	assert.Equal(t, "my-api-token", response.Data.Name)
	assert.Equal(t, "abc-123", response.Data.UUID)
	assert.Equal(t, "secret-value", response.Data.Value)
}
