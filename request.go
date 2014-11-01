// Copyright 2014 Vadim Kravcenko
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
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
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

func (r *Requester) Post(endpoint string, payload io.Reader, responseStruct interface{}, querystring map[string]string) *http.Response {
	r.SetHeader("Content-Type", "application/x-www-form-urlencoded")
	r.Suffix = "api/json"
	return r.Do("POST", endpoint, payload, &responseStruct, querystring)
}

func (r *Requester) PostFiles(endpoint string, payload io.Reader, responseStruct interface{}, querystring map[string]string, files []string) *http.Response {
	return r.Do("POST", endpoint, payload, &responseStruct, querystring, files)
}

func (r *Requester) PostXML(endpoint string, xml string, responseStruct interface{}, querystring map[string]string) *http.Response {
	payload := bytes.NewBuffer([]byte(xml))
	r.SetHeader("Content-Type", "application/xml")
	r.Suffix = ""
	return r.Do("POST", endpoint, payload, &responseStruct, querystring)
}

func (r *Requester) GetJSON(endpoint string, responseStruct interface{}, querystring map[string]string) *http.Response {
	r.SetHeader("Content-Type", "application/json")
	r.Suffix = "api/json"
	return r.Do("GET", endpoint, nil, responseStruct, querystring)
}

func (r *Requester) GetXML(endpoint string, responseStruct interface{}, querystring map[string]string) *http.Response {
	r.SetHeader("Content-Type", "application/json")
	r.Suffix = "api/json"
	return r.Do("XML", endpoint, nil, responseStruct, querystring)
}

func (r *Requester) Get(endpoint string, responseStruct interface{}, querystring map[string]string) *http.Response {
	r.Suffix = ""
	return r.Do("XML", endpoint, nil, responseStruct, querystring)
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

func (r *Requester) Do(method string, endpoint string, payload io.Reader, responseStruct interface{}, options ...interface{}) *http.Response {
	if !strings.HasSuffix(endpoint, "/") {
		endpoint += "/"
	}
	fileUpload := false
	var files []string
	url := r.Base + endpoint + r.Suffix
	for _, o := range options {
		switch v := o.(type) {
		case map[string]string:
			url += r.parseQueryString(v)
			break
		case []string:
			fileUpload = true
			files = v
		}
	}
	var req *http.Request
	var err error
	if fileUpload {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		for _, file := range files {
			fileData, err := os.Open(file)
			if err != nil {
				Error.Println(err.Error())
				return nil
			}

			part, err := writer.CreateFormFile("file", filepath.Base(file))
			if err != nil {
				Error.Println(err.Error())
			}
			_, err = io.Copy(part, fileData)
			defer fileData.Close()
		}
		var params map[string]string
		json.NewDecoder(payload).Decode(&params)
		for key, val := range params {
			_ = writer.WriteField(key, val)
		}
		err = writer.Close()
		req, err = http.NewRequest(method, url, body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
	} else {

		req, err = http.NewRequest(method, url, payload)
		if err != nil {
			Error.Println(err.Error())
		}
	}

	if r.BasicAuth != nil {
		req.SetBasicAuth(r.BasicAuth.Username, r.BasicAuth.Password)
	}

	if r.Headers != nil {
		for k, _ := range r.Headers {
			req.Header.Add(k, r.Headers.Get(k))
		}
	}

	r.LastResponse, err = r.Client.Do(req)

	if err != nil {
		panic(err)
	}

	defer r.LastResponse.Body.Close()

	if method == "XML" {
		content, err := ioutil.ReadAll(r.LastResponse.Body)
		if str, ok := responseStruct.(*string); ok {
			*str = string(content)
		}
		if err != nil {
			panic(err)
		}
		return r.LastResponse
	}
	json.NewDecoder(r.LastResponse.Body).Decode(responseStruct)
	return r.LastResponse
}
