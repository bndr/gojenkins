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
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAPIRequest_Basic(t *testing.T) {
	ar := NewAPIRequest("GET", "/api/json", nil)

	assert.Equal(t, "GET", ar.Method)
	assert.Equal(t, "/api/json", ar.Endpoint)
	assert.Nil(t, ar.Payload)
	assert.NotNil(t, ar.Headers)
	assert.Equal(t, "", ar.Suffix)
}

func TestNewAPIRequest_WithPayload(t *testing.T) {
	payload := strings.NewReader("test payload")
	ar := NewAPIRequest("POST", "/job/test/build", payload)

	assert.Equal(t, "POST", ar.Method)
	assert.Equal(t, "/job/test/build", ar.Endpoint)
	assert.NotNil(t, ar.Payload)

	// Verify payload content
	content, err := io.ReadAll(ar.Payload)
	assert.NoError(t, err)
	assert.Equal(t, "test payload", string(content))
}

func TestNewAPIRequest_DifferentMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			ar := NewAPIRequest(method, "/test", nil)
			assert.Equal(t, method, ar.Method)
		})
	}
}

func TestAPIRequest_SetHeader(t *testing.T) {
	ar := NewAPIRequest("GET", "/test", nil)

	// SetHeader should return the APIRequest for chaining
	result := ar.SetHeader("Content-Type", "application/json")
	assert.Same(t, ar, result)
	assert.Equal(t, "application/json", ar.Headers.Get("Content-Type"))
}

func TestAPIRequest_SetHeader_Multiple(t *testing.T) {
	ar := NewAPIRequest("GET", "/test", nil)

	ar.SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer token").
		SetHeader("X-Custom-Header", "custom-value")

	assert.Equal(t, "application/json", ar.Headers.Get("Content-Type"))
	assert.Equal(t, "Bearer token", ar.Headers.Get("Authorization"))
	assert.Equal(t, "custom-value", ar.Headers.Get("X-Custom-Header"))
}

func TestAPIRequest_SetHeader_Overwrite(t *testing.T) {
	ar := NewAPIRequest("GET", "/test", nil)

	ar.SetHeader("Content-Type", "application/xml")
	ar.SetHeader("Content-Type", "application/json")

	assert.Equal(t, "application/json", ar.Headers.Get("Content-Type"))
}

func TestReadRawResponse_Success(t *testing.T) {
	requester := &Requester{}

	response := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBufferString("raw response content")),
	}

	var result string
	resp, err := requester.ReadRawResponse(response, &result)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "raw response content", result)
}

func TestReadRawResponse_EmptyBody(t *testing.T) {
	requester := &Requester{}

	response := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBufferString("")),
	}

	var result string
	resp, err := requester.ReadRawResponse(response, &result)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "", result)
}

func TestReadRawResponse_InvalidType(t *testing.T) {
	requester := &Requester{}

	response := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBufferString("content")),
	}

	var result int // Wrong type
	_, err := requester.ReadRawResponse(response, &result)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not cast responseStruct to *string")
}

func TestReadJSONResponse_Success(t *testing.T) {
	requester := &Requester{}

	jsonBody := `{"name": "test-job", "color": "blue"}`
	response := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBufferString(jsonBody)),
	}

	var result struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}
	resp, err := requester.ReadJSONResponse(response, &result)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test-job", result.Name)
	assert.Equal(t, "blue", result.Color)
}

func TestReadJSONResponse_EmptyBody(t *testing.T) {
	requester := &Requester{}

	response := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBufferString("")),
	}

	var result struct {
		Name string `json:"name"`
	}
	resp, err := requester.ReadJSONResponse(response, &result)

	// Empty body doesn't cause an error, just doesn't populate the struct
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "", result.Name)
}

func TestReadJSONResponse_ComplexStructure(t *testing.T) {
	requester := &Requester{}

	jsonBody := `{
		"name": "test-job",
		"builds": [
			{"number": 1, "url": "http://jenkins/job/test/1/"},
			{"number": 2, "url": "http://jenkins/job/test/2/"}
		],
		"lastBuild": {"number": 2, "url": "http://jenkins/job/test/2/"}
	}`
	response := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBufferString(jsonBody)),
	}

	var result struct {
		Name   string `json:"name"`
		Builds []struct {
			Number int64  `json:"number"`
			URL    string `json:"url"`
		} `json:"builds"`
		LastBuild struct {
			Number int64  `json:"number"`
			URL    string `json:"url"`
		} `json:"lastBuild"`
	}
	resp, err := requester.ReadJSONResponse(response, &result)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test-job", result.Name)
	assert.Len(t, result.Builds, 2)
	assert.Equal(t, int64(1), result.Builds[0].Number)
	assert.Equal(t, int64(2), result.LastBuild.Number)
}

func TestRequester_SetClient(t *testing.T) {
	requester := &Requester{}
	customClient := &http.Client{}

	result := requester.SetClient(customClient)

	assert.Same(t, requester, result)
	assert.Same(t, customClient, requester.Client)
}

func TestRequester_Fields(t *testing.T) {
	requester := &Requester{
		Base: "http://localhost:8080",
		BasicAuth: &BasicAuth{
			Username: "admin",
			Password: "password",
		},
		SslVerify: true,
	}

	assert.Equal(t, "http://localhost:8080", requester.Base)
	assert.Equal(t, "admin", requester.BasicAuth.Username)
	assert.Equal(t, "password", requester.BasicAuth.Password)
	assert.True(t, requester.SslVerify)
}
