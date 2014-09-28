package main

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
)

// Request Methods

type Request struct {
	Base      string
	BasicAuth *BasicAuth
	Headers   http.Header
	Client    *http.Client
}

func (r *Request) Create() {

}

func (r *Request) Post() {

}

func (r *Request) PostXML() {

}

func (r *Request) Get() {

}

func (r *Request) SetHeaders() {

}

func (r *Request) SetClient() {

}

func (r *Request) Do(method string) {
	var url string
	var payload io.Reader

	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		panic(err)
	}

	if c.BasicAuth != nil {
		req.SetBasicAuth(c.BasicAuth.Username, c.BasicAuth.Password)
	}

	if c.Headers != nil {
		for k, _ := range c.Headers {
			req.Header.Set(k, c.Headers.Get(k))
		}
	}

	// Rewrite
	cookies, _ := cookiejar.New(nil)

	// Skip verify by default?
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
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

}
