package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/cookiejar"
)

type BasicAuth struct {
	Username string
	Password string
}

type TokenAuth struct {
	Username string
	Token    string
}

type Jenkins struct {
	Server    string
	Version   string
	Requester *Requester
}

// Jenkins

func (j *Jenkins) Init() *Jenkins {

	// Skip SSL Verification?
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !j.Requester.SslVerify},
	}

	cookies, _ := cookiejar.New(nil)

	client := &http.Client{
		Transport: tr,
		Jar:       cookies,
	}
	if j.Requester.Client == nil {
		j.Requester.Client = client
	}

	// Check Connection
	resp := new(ExecutorResponse)
	raw := j.Requester.Do("GET", "api/json", nil, resp)
	j.Version = raw.Header.Get("X-Jenkins")
	if resp == nil {
		panic("Connection Failed, Please verify that the host and credentials are correct.")
	}
	return j
}

func (j *Jenkins) Info() {

}

func (j *Jenkins) CreateNode() {

}

func (j *Jenkins) CreateBuild() {

}

func (j *Jenkins) CreateJob() {

}

func (j *Jenkins) GetNode() {

}

func (j *Jenkins) GetBuild() {

}

func (j *Jenkins) GetJob() {

}

func (j *Jenkins) GetAllNodes() {

}

func (j *Jenkins) GetAllBuilds() {

}

func (j *Jenkins) GetAllJobs() {

}

func CreateJenkins(base string, username string, password string) *Jenkins {
	j := &Jenkins{}
	j.Server = base
	j.Requester = &Requester{Base: base, SslVerify: false, Headers: http.Header{}}
	j.Requester.BasicAuth = &BasicAuth{Username: username, Password: password}
	return j
}

func main() {
	j := CreateJenkins("http://localhost:8081/", "admin", "admin").Init()
	fmt.Printf("%#v\n", j)
}
