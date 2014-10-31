package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type Artifact struct {
	Jenkins  *Jenkins
	Build    *Build
	FileName string
	Path     string
}

func (a Artifact) GetData() ([]byte, error) {
	var data string
	a.Jenkins.Requester.Get(a.Path, &data, nil)
	code := a.Jenkins.Requester.LastResponse.StatusCode
	if code != 200 {
		Error.Printf("Jenkins responded with StatusCode: %d", code)
		return nil, errors.New("Could not get File Contents")
	}
	return []byte(data), nil
}

func (a Artifact) Save(path string) {
	Info.Printf("Saving artifact @ %s to %s", a.Path, path)
	data, err := a.GetData()

	if err != nil {
		Error.Println("No Data received, not saving file.")
		return
	}

	if _, err = os.Stat(path); err == nil {
		Warning.Println("Local Copy already exists, Overwriting...")
	}

	err = ioutil.WriteFile(path, data, 0644)
	a.validateDownload(path)

	if err != nil {
		Error.Println(err.Error())
	}
}

func (a Artifact) SaveToDir(dir string) {

}

func (a Artifact) validateDownload(path string) bool {
	localHash := a.getMD5local(path)
	fp := Fingerprint{Jenkins: a.Jenkins, Base: "/fingerprint/", Id: localHash, Raw: new(fingerPrintResponse)}

	if !fp.ValidateForBuild(a.FileName, a.Build) {
		Error.Println("Fingerprint of the downloaded artifact could not be verified")
		return false
	}
	return true
}

// Get Local MD5
func (a Artifact) getMD5local(path string) string {
	h := md5.New()
	localFile, err := os.Open(path)
	if err != nil {
		return ""
	}
	buffer := make([]byte, 2^20)
	n, err := localFile.Read(buffer)
	defer localFile.Close()
	for err == nil {
		io.WriteString(h, string(buffer[0:n]))
		n, err = localFile.Read(buffer)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
