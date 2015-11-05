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
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Request Methods

type Requester struct {
	Base         string
	BasicAuth    *BasicAuth
	Headers      http.Header
	Client       *http.Client
	SslVerify    bool
	LastResponse *http.Response
	Suffix       string
}

func (r *Requester) PostJSON(endpoint string, payload io.Reader, responseStruct interface{}, querystring map[string]string) (*http.Response, error) {
	r.SetHeader("Content-Type", "application/x-www-form-urlencoded")
	r.Suffix = "api/json"
	return r.Do("POST", endpoint, payload, &responseStruct, querystring)
}

func (r *Requester) Post(endpoint string, payload io.Reader, responseStruct interface{}, querystring map[string]string) (*http.Response, error) {
	r.SetHeader("Content-Type", "application/x-www-form-urlencoded")
	r.Suffix = ""
	return r.Do("POST", endpoint, payload, &responseStruct, querystring)
}

func (r *Requester) PostFiles(endpoint string, payload io.Reader, responseStruct interface{}, querystring map[string]string, files []string) (*http.Response, error) {
	return r.Do("POST", endpoint, payload, &responseStruct, querystring, files)
}

func (r *Requester) PostXML(endpoint string, xml string, responseStruct interface{}, querystring map[string]string) (*http.Response, error) {
	payload := bytes.NewBuffer([]byte(xml))
	r.SetHeader("Content-Type", "application/xml")
	r.Suffix = ""
	return r.Do("POST", endpoint, payload, &responseStruct, querystring)
}

func (r *Requester) GetJSON(endpoint string, responseStruct interface{}, querystring map[string]string) (*http.Response, error) {
	r.SetHeader("Content-Type", "application/json")
	r.Suffix = "api/json"
	return r.Do("GET", endpoint, nil, responseStruct, querystring)
}

func (r *Requester) GetXML(endpoint string, responseStruct interface{}, querystring map[string]string) (*http.Response, error) {
	r.SetHeader("Content-Type", "application/xml")
	r.Suffix = ""
	return r.Do("GET", endpoint, nil, responseStruct, querystring)
}

func (r *Requester) Get(endpoint string, responseStruct interface{}, querystring map[string]string) (*http.Response, error) {
	r.Suffix = ""
	return r.Do("GET", endpoint, nil, responseStruct, querystring)
}

func (r *Requester) SetHeader(key string, value string) *Requester {
	r.Headers.Set(key, value)
	return r
}

func (r *Requester) SetClient(client *http.Client) *Requester {
	r.Client = client
	return r
}

func (r *Requester) parseQueryString(queries map[string]string) string {
	output := ""
	delimiter := "?"
	for k, v := range queries {
		output += delimiter + k + "=" + v
		delimiter = "&"
	}
	return output
}

func (r *Requester) Do(method string, endpoint string, payload io.Reader, responseStruct interface{}, options ...interface{}) (*http.Response, error) {
	if !strings.HasSuffix(endpoint, "/") && method != "POST" {
		endpoint += "/"
	}

	fileUpload := false
	var files []string
	URL, err := url.Parse(r.Base + endpoint + r.Suffix)

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
			break
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
				return nil, err
			}
			defer fileData.Close()
		}
		var params map[string]string
		json.NewDecoder(payload).Decode(&params)
		for key, val := range params {
			if err = writer.WriteField(key, val); err != nil {
				return nil, err
			}
		}
		if err = writer.Close(); err != nil {
			return nil, err
		}
		req, err = http.NewRequest(method, URL.String(), body)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())
	} else {

		req, err = http.NewRequest(method, URL.String(), payload)
		if err != nil {
			return nil, err
		}
	}

	if r.BasicAuth != nil {
		req.SetBasicAuth(r.BasicAuth.Username, r.BasicAuth.Password)
	}

	if r.Headers != nil {
		for k := range r.Headers {
			req.Header.Add(k, r.Headers.Get(k))
		}
	}

	r.LastResponse, err = r.Client.Do(req)

	if err != nil {
		return nil, err
	}

	errorText := r.LastResponse.Header.Get("X-Error")

	if errorText != "" {
		return nil, errors.New(errorText)
	}

	switch responseStruct.(type) {
	case *string:
		rawResponse, err := r.ReadRawResponse(responseStruct)
		if err != nil {
			return nil, err
		}
		return rawResponse, nil
	default:
		jsonResponse := r.ReadJSONResponse(responseStruct)
		return jsonResponse, nil
	}
}

func (r *Requester) ReadRawResponse(responseStruct interface{}) (*http.Response, error) {
	defer r.LastResponse.Body.Close()

	content, err := ioutil.ReadAll(r.LastResponse.Body)
	if str, ok := responseStruct.(*string); ok {
		*str = string(content)
	}

	if err != nil {
		return nil, err
	}

	return r.LastResponse, nil
}

func (r *Requester) ReadJSONResponse(responseStruct interface{}) *http.Response {
	defer r.LastResponse.Body.Close()
	json.NewDecoder(r.LastResponse.Body).Decode(responseStruct)
	return r.LastResponse
}
