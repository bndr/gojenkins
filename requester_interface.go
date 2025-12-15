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

// JenkinsRequester defines the interface for making Jenkins API requests.
// This interface enables mocking for unit tests.
type JenkinsRequester interface {
	GetJSON(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error)
	Post(ctx context.Context, endpoint string, payload io.Reader, response interface{}, query map[string]string) (*http.Response, error)
	PostXML(ctx context.Context, endpoint string, xml string, response interface{}, query map[string]string) (*http.Response, error)
	PostJSON(ctx context.Context, endpoint string, payload io.Reader, response interface{}, query map[string]string) (*http.Response, error)
	PostFiles(ctx context.Context, endpoint string, payload io.Reader, response interface{}, query map[string]string, files []string) (*http.Response, error)
	Get(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error)
	GetXML(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error)
}

// Ensure Requester implements JenkinsRequester
var _ JenkinsRequester = (*Requester)(nil)
