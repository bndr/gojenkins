package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
)

// Request Methods

type Requester struct {
	Base      string
	BasicAuth *BasicAuth
	Headers   http.Header
	Client    *http.Client
	SslVerify bool
}

type Response struct {
}

func (r *Requester) Post(endpoint string, payload io.Reader, responseStruct interface{}) {
	r.SetHeader("Content-Type", "application/x-www-form-urlencoded")
	r.Do("POST", endpoint, payload, &responseStruct)
}

func (r *Requester) PostXML(endpoint string, xml string, responseStruct interface{}) {
	payload := bytes.NewBuffer([]byte(xml))
	r.SetHeader("Content-Type", "text/xml")
	r.Do("XML", endpoint, payload, &responseStruct)
}

func (r *Requester) Get(endpoint string, responseStruct interface{}) {
	r.SetHeader("Content-Type", "application/json")
	r.Do("GET", endpoint, nil, responseStruct)
}

func (r *Requester) SetHeader(key string, value string) *Requester {
	r.Headers.Add(key, value)
	return r
}

func (r *Requester) SetClient(client *http.Client) *Requester {
	r.Client = client
	return r
}

func (r *Requester) SetQuery(querystring map[string]string) *Requester {
	// TODO
	return r
}

func (r *Requester) Do(method string, endpoint string, payload io.Reader, responseStruct interface{}) *http.Response {
	url := r.Base + endpoint

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

	// Rewrite
	cookies, _ := cookiejar.New(nil)

	// Skip verify by default?
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !r.SslVerify},
	}

	client := &http.Client{
		Transport: tr,
		Jar:       cookies,
	}
	req.Close = true
	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	if resp.StatusCode >= 400 {
		panic(err)
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(responseStruct)
	if err != nil {
		panic(err)
	}
	return resp
}
