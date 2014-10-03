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

func (r *Requester) Post(endpoint string, payload io.Reader) {
	r.SetHeader("Content-Type", "application/x-www-form-urlencoded")
	r.Do("POST", endpoint, payload)
}

func (r *Requester) PostXML(endpoint string, xml string) {
	payload := bytes.NewBuffer([]byte(xml))
	r.SetHeader("Content-Type", "text/xml")
	r.Do("XML", endpoint, payload)
}

func (r *Requester) Get(endpoint string) {
	r.Do("GET", endpoint)
}

func (r *Requester) SetHeader(key string, value string) {
	r.Headers.Add(key, value)
}

func (r *Requester) SetClient(client *http.Client) {
	r.Client = client
}

func (r *Requester) Do(method string, endpoint string, payload io.Reader) *Response {
	var url string

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

	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	if resp.StatusCode >= 400 {
		panic(err)
	}

	defer resp.Body.Close()

	var response interface{}

	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		panic(err)
	}
	return &Response{}
}
