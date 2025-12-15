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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// APIRequest represents an HTTP request to the Jenkins API.
type APIRequest struct {
	Method   string
	Endpoint string
	Payload  io.Reader
	Headers  http.Header
	Suffix   string
}

// SetHeader sets a header on the API request.
func (ar *APIRequest) SetHeader(key string, value string) *APIRequest {
	ar.Headers.Set(key, value)
	return ar
}

// NewAPIRequest creates a new API request with the specified method, endpoint, and payload.
func NewAPIRequest(method string, endpoint string, payload io.Reader) *APIRequest {
	var headers = http.Header{}
	var suffix string
	ar := &APIRequest{method, endpoint, payload, headers, suffix}
	return ar
}

// Requester handles HTTP requests to the Jenkins API.
type Requester struct {
	Base      string
	BasicAuth *BasicAuth
	Client    *http.Client
	CACert    []byte
	SslVerify bool
}

// SetCrumb fetches and sets the CSRF crumb token on the request.
func (r *Requester) SetCrumb(ctx context.Context, ar *APIRequest) error {
	crumbData := map[string]string{}
	response, _ := r.GetJSON(ctx, "/crumbIssuer/api/json", &crumbData, nil)

	if response.StatusCode == 200 && crumbData["crumbRequestField"] != "" {
		ar.SetHeader(crumbData["crumbRequestField"], crumbData["crumb"])
		ar.SetHeader("Cookie", response.Header.Get("set-cookie"))
	}

	return nil
}

// PostJSON sends a POST request with JSON content type.
func (r *Requester) PostJSON(ctx context.Context, endpoint string, payload io.Reader, responseStruct interface{}, querystring map[string]string) (*http.Response, error) {
	ar := NewAPIRequest("POST", endpoint, payload)
	if err := r.SetCrumb(ctx, ar); err != nil {
		return nil, err
	}
	ar.SetHeader("Content-Type", "application/x-www-form-urlencoded")
	ar.Suffix = "api/json"
	return r.Do(ctx, ar, &responseStruct, querystring)
}

// Post sends a POST request with form-urlencoded content type.
func (r *Requester) Post(ctx context.Context, endpoint string, payload io.Reader, responseStruct interface{}, querystring map[string]string) (*http.Response, error) {
	ar := NewAPIRequest("POST", endpoint, payload)
	if err := r.SetCrumb(ctx, ar); err != nil {
		return nil, err
	}
	ar.SetHeader("Content-Type", "application/x-www-form-urlencoded")
	ar.Suffix = ""
	return r.Do(ctx, ar, &responseStruct, querystring)
}

// PostFiles sends a POST request with file attachments.
func (r *Requester) PostFiles(ctx context.Context, endpoint string, payload io.Reader, responseStruct interface{}, querystring map[string]string, files []string) (*http.Response, error) {
	ar := NewAPIRequest("POST", endpoint, payload)
	if err := r.SetCrumb(ctx, ar); err != nil {
		return nil, err
	}
	return r.Do(ctx, ar, &responseStruct, querystring, files)
}

// PostXML sends a POST request with XML content.
func (r *Requester) PostXML(ctx context.Context, endpoint string, xml string, responseStruct interface{}, querystring map[string]string) (*http.Response, error) {
	payload := bytes.NewBuffer([]byte(xml))
	ar := NewAPIRequest("POST", endpoint, payload)
	if err := r.SetCrumb(ctx, ar); err != nil {
		return nil, err
	}
	ar.SetHeader("Content-Type", "application/xml;charset=utf-8")
	ar.Suffix = ""
	return r.Do(ctx, ar, &responseStruct, querystring)
}

// GetJSON sends a GET request and expects a JSON response.
func (r *Requester) GetJSON(ctx context.Context, endpoint string, responseStruct interface{}, query map[string]string) (*http.Response, error) {
	ar := NewAPIRequest("GET", endpoint, nil)
	ar.SetHeader("Content-Type", "application/json")
	ar.Suffix = "api/json"
	return r.Do(ctx, ar, &responseStruct, query)
}

// GetXML sends a GET request and expects an XML response.
func (r *Requester) GetXML(ctx context.Context, endpoint string, responseStruct interface{}, query map[string]string) (*http.Response, error) {
	ar := NewAPIRequest("GET", endpoint, nil)
	ar.SetHeader("Content-Type", "application/xml")
	ar.Suffix = ""
	return r.Do(ctx, ar, responseStruct, query)
}

// Get sends a GET request to the specified endpoint.
func (r *Requester) Get(ctx context.Context, endpoint string, responseStruct interface{}, querystring map[string]string) (*http.Response, error) {
	ar := NewAPIRequest("GET", endpoint, nil)
	ar.Suffix = ""
	return r.Do(ctx, ar, responseStruct, querystring)
}

// SetClient sets the HTTP client to use for requests.
func (r *Requester) SetClient(client *http.Client) *Requester {
	r.Client = client
	return r
}

// Do executes the API request and returns the HTTP response.
func (r *Requester) Do(ctx context.Context, ar *APIRequest, responseStruct interface{}, options ...interface{}) (*http.Response, error) {
	if !strings.HasSuffix(ar.Endpoint, "/") && ar.Method != "POST" {
		ar.Endpoint += "/"
	}

	fileUpload := false
	var files []string
	URL, err := url.Parse(r.Base + ar.Endpoint + ar.Suffix)

	if err != nil {
		return nil, err
	}

	for _, o := range options {
		switch v := o.(type) {
		case map[string]string:

			querystring := make(url.Values)
			for key, val := range v {
				querystring.Set(key, val)
			}

			URL.RawQuery = querystring.Encode()
		case []string:
			fileUpload = true
			files = v
		}
	}
	var req *http.Request

	if fileUpload {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		for _, file := range files {
			fileData, err := os.Open(file)
			if err != nil {
				Error.Println(err.Error())
				return nil, err
			}

			part, err := writer.CreateFormFile("file", filepath.Base(file))
			if err != nil {
				Error.Println(err.Error())
				return nil, err
			}
			if _, err = io.Copy(part, fileData); err != nil {
				_ = fileData.Close()
				return nil, err
			}
			_ = fileData.Close()
		}
		var params map[string]string
		if ar.Payload != nil {
			if err := json.NewDecoder(ar.Payload).Decode(&params); err != nil {
				// Ignore decode errors - payload may not be JSON
				params = nil
			}
		}
		for key, val := range params {
			if err = writer.WriteField(key, val); err != nil {
				return nil, err
			}
		}
		if err = writer.Close(); err != nil {
			return nil, err
		}
		req, err = http.NewRequestWithContext(ctx, ar.Method, URL.String(), body)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())
	} else {

		req, err = http.NewRequestWithContext(ctx, ar.Method, URL.String(), ar.Payload)
		if err != nil {
			return nil, err
		}
	}

	if r.BasicAuth != nil {
		req.SetBasicAuth(r.BasicAuth.Username, r.BasicAuth.Password)
	}

	for k := range ar.Headers {
		req.Header.Add(k, ar.Headers.Get(k))
	}

	if response, err := r.Client.Do(req); err != nil {
		return nil, err
	} else {
		if v := ctx.Value("debug"); v != nil {
			dump, err := httputil.DumpResponse(response, true)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("DEBUG %q\n", dump)
		}
		errorText := response.Header.Get("X-Error")
		if errorText != "" {
			return nil, errors.New(errorText)
		}
		switch responseStruct.(type) {
		case *string:
			return r.ReadRawResponse(response, responseStruct)
		default:
			return r.ReadJSONResponse(response, responseStruct)
		}

	}

}

// ReadRawResponse reads the response body as a raw string.
func (r *Requester) ReadRawResponse(response *http.Response, responseStruct interface{}) (*http.Response, error) {
	defer func() { _ = response.Body.Close() }()

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if str, ok := responseStruct.(*string); ok {
		*str = string(content)
	} else {
		return nil, fmt.Errorf("could not cast responseStruct to *string")
	}

	return response, nil
}

// ReadJSONResponse reads the response body as JSON and decodes it into responseStruct.
func (r *Requester) ReadJSONResponse(response *http.Response, responseStruct interface{}) (*http.Response, error) {
	defer func() { _ = response.Body.Close() }()

	if err := json.NewDecoder(response.Body).Decode(responseStruct); err != nil && err != io.EOF {
		return response, err
	}
	return response, nil
}
