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
	"context"
	"io"
	"net/http"
)

// MockRequester is a mock implementation of JenkinsRequester for testing.
type MockRequester struct {
	// Default response and error for simple tests
	response     *http.Response
	err          error
	lastEndpoint string

	// Function fields allow customizing behavior per test
	GetJSONFunc   func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error)
	PostFunc      func(ctx context.Context, endpoint string, payload io.Reader, response interface{}, query map[string]string) (*http.Response, error)
	PostXMLFunc   func(ctx context.Context, endpoint string, xml string, response interface{}, query map[string]string) (*http.Response, error)
	PostJSONFunc  func(ctx context.Context, endpoint string, payload io.Reader, response interface{}, query map[string]string) (*http.Response, error)
	PostFilesFunc func(ctx context.Context, endpoint string, payload io.Reader, response interface{}, query map[string]string, files []string) (*http.Response, error)
	GetFunc       func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error)
	GetXMLFunc    func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error)
}

// GetJSON implements JenkinsRequester.
func (m *MockRequester) GetJSON(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
	m.lastEndpoint = endpoint
	if m.GetJSONFunc != nil {
		return m.GetJSONFunc(ctx, endpoint, response, query)
	}
	if m.err != nil {
		return nil, m.err
	}
	if m.response != nil {
		return m.response, nil
	}
	return &http.Response{StatusCode: 200}, nil
}

// Post implements JenkinsRequester.
func (m *MockRequester) Post(ctx context.Context, endpoint string, payload io.Reader, response interface{}, query map[string]string) (*http.Response, error) {
	m.lastEndpoint = endpoint
	if m.PostFunc != nil {
		return m.PostFunc(ctx, endpoint, payload, response, query)
	}
	if m.err != nil {
		return nil, m.err
	}
	if m.response != nil {
		return m.response, nil
	}
	return &http.Response{StatusCode: 200}, nil
}

// PostXML implements JenkinsRequester.
func (m *MockRequester) PostXML(ctx context.Context, endpoint string, xml string, response interface{}, query map[string]string) (*http.Response, error) {
	m.lastEndpoint = endpoint
	if m.PostXMLFunc != nil {
		return m.PostXMLFunc(ctx, endpoint, xml, response, query)
	}
	if m.err != nil {
		return nil, m.err
	}
	if m.response != nil {
		return m.response, nil
	}
	return &http.Response{StatusCode: 200}, nil
}

// PostJSON implements JenkinsRequester.
func (m *MockRequester) PostJSON(ctx context.Context, endpoint string, payload io.Reader, response interface{}, query map[string]string) (*http.Response, error) {
	m.lastEndpoint = endpoint
	if m.PostJSONFunc != nil {
		return m.PostJSONFunc(ctx, endpoint, payload, response, query)
	}
	if m.err != nil {
		return nil, m.err
	}
	if m.response != nil {
		return m.response, nil
	}
	return &http.Response{StatusCode: 200}, nil
}

// PostFiles implements JenkinsRequester.
func (m *MockRequester) PostFiles(ctx context.Context, endpoint string, payload io.Reader, response interface{}, query map[string]string, files []string) (*http.Response, error) {
	m.lastEndpoint = endpoint
	if m.PostFilesFunc != nil {
		return m.PostFilesFunc(ctx, endpoint, payload, response, query, files)
	}
	if m.err != nil {
		return nil, m.err
	}
	if m.response != nil {
		return m.response, nil
	}
	return &http.Response{StatusCode: 200}, nil
}

// Get implements JenkinsRequester.
func (m *MockRequester) Get(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
	m.lastEndpoint = endpoint
	if m.GetFunc != nil {
		return m.GetFunc(ctx, endpoint, response, query)
	}
	if m.err != nil {
		return nil, m.err
	}
	if m.response != nil {
		return m.response, nil
	}
	return &http.Response{StatusCode: 200}, nil
}

// GetXML implements JenkinsRequester.
func (m *MockRequester) GetXML(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
	m.lastEndpoint = endpoint
	if m.GetXMLFunc != nil {
		return m.GetXMLFunc(ctx, endpoint, response, query)
	}
	if m.err != nil {
		return nil, m.err
	}
	if m.response != nil {
		return m.response, nil
	}
	return &http.Response{StatusCode: 200}, nil
}

// Ensure MockRequester implements JenkinsRequester
var _ JenkinsRequester = (*MockRequester)(nil)

// newMockJenkins creates a Jenkins instance with a mock requester for testing.
func newMockJenkins() *Jenkins {
	return &Jenkins{
		Server:    "http://localhost:8080",
		Requester: &MockRequester{},
		Raw:       &ExecutorResponse{},
	}
}
