package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
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
}

type Response struct {
}

func (r *Requester) Post(endpoint string, payload io.Reader, responseStruct interface{}, querystring map[string]string) *http.Response {
	r.SetHeader("Content-Type", "application/x-www-form-urlencoded")

	return r.Do("POST", endpoint, payload, &responseStruct, querystring)
}

func (r *Requester) PostXML(endpoint string, xml string, responseStruct interface{}, options ...interface{}) *http.Response {
	payload := bytes.NewBuffer([]byte(xml))
	r.SetHeader("Content-Type", "text/xml")
	return r.Do("XML", endpoint, payload, &responseStruct, options)
}

func (r *Requester) Get(endpoint string, responseStruct interface{}, querystring map[string]string) *http.Response {
	r.SetHeader("Content-Type", "application/json")
	return r.Do("GET", endpoint, nil, responseStruct, querystring)
}

func (r *Requester) GetXML(endpoint string, responseStruct interface{}, querystring map[string]string) *http.Response {
	r.SetHeader("Content-Type", "application/json")
	return r.Do("XML", endpoint, nil, responseStruct, querystring)
}

func (r *Requester) SetHeader(key string, value string) *Requester {
	r.Headers.Add(key, value)
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
	url := r.Base + endpoint + "api/json"
	for _, o := range options {
		switch v := o.(type) {
		case map[string]string:
			url += r.parseQueryString(v)
		}
	}
	fmt.Printf("%s\n\n", url)
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		panic(err)
	}

	if r.BasicAuth != nil {
		req.SetBasicAuth(r.BasicAuth.Username, r.BasicAuth.Password)
	}

	if r.Headers != nil {
		for k, _ := range r.Headers {
			req.Header.Set(k, r.Headers.Get(k))
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
