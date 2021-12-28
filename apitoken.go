package gojenkins

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

const (
	apiTokenBaseContext   = "/me/descriptorByName/jenkins.security.ApiTokenProperty"
	generateAPITokenURL   = apiTokenBaseContext + "/generateNewToken"
	revokeAPITokenURL     = apiTokenBaseContext + "/revoke"
	revokeAllAPITokensURL = apiTokenBaseContext + "/revokeAll"
)

// APIToken is a Jenkins API token to be created for the user instantiated with the Jenkins client
type APIToken struct {
	Jenkins *Jenkins
	Name    string `json:"tokenName"`
	UUID    string `json:"tokenUuid"`
	Value   string `json:"tokenValue"`
}

// APITokenGenerateResponse is the response given by Jenkins when an API token is created
type APITokenGenerateResponse struct {
	Status string   `json:"status"`
	Data   APIToken `json:"data"`
}

// ErrAPIToken occurs when there is error creating or revoking API tokens
type ErrAPIToken struct {
	Message string
}

func (e *ErrAPIToken) Error() string {
	return e.Message
}

// GenerateAPIToken creates a new API token for the Jenkins client user
func (j *Jenkins) GenerateAPIToken(ctx context.Context, tokenName string) (APIToken, error) {
	payload := "newTokenName=" + tokenName
	apiTokenResponse := &APITokenGenerateResponse{}
	response, err := j.Requester.Post(ctx, generateAPITokenURL, strings.NewReader(payload), apiTokenResponse, nil)
	if err != nil {
		return apiTokenResponse.Data, err
	}
	if response.StatusCode != http.StatusOK {
		return apiTokenResponse.Data, &ErrAPIToken{
			Message: fmt.Sprintf("error creating API token. Status is %d", response.StatusCode),
		}
	}
	apiToken := apiTokenResponse.Data
	// Set Jenkins client pointer to be able to revoke token later
	apiToken.Jenkins = j
	return apiToken, nil
}

// RevokeAPIToken revokes an API token
func (j *Jenkins) RevokeAPIToken(ctx context.Context, tokenUuid string) error {
	payload := "tokenUuid=" + tokenUuid
	response, err := j.Requester.Post(ctx, revokeAPITokenURL, strings.NewReader(payload), nil, nil)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return &ErrAPIToken{
			Message: fmt.Sprintf("error revoking API token. Status is %d", response.StatusCode),
		}
	}
	return nil
}

// RevokeAllAPITokens revokes all API tokens for the Jenkins client user
func (j *Jenkins) RevokeAllAPITokens(ctx context.Context) error {
	response, err := j.Requester.Post(ctx, revokeAllAPITokensURL, nil, nil, nil)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return &ErrAPIToken{
			Message: fmt.Sprintf("error revoking all API tokens. Status is %d", response.StatusCode),
		}
	}
	return nil
}

// Revoke revokes an API token
func (a *APIToken) Revoke() error {
	return a.Jenkins.RevokeAPIToken(context.Background(), a.UUID)
}
