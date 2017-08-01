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
	"errors"
	"fmt"
)

// Fingerprint represents a Fingerprint
type Fingerprint struct {
	Jenkins *Jenkins
	Base    string
	Id      string
	Raw     *fingerPrintResponse
}

type fingerPrintResponse struct {
	FileName string `json:"fileName"`
	Hash     string `json:"hash"`
	Original struct {
		Name   string
		Number int64
	} `json:"original"`
	Timestamp int64 `json:"timestamp"`
	Usage     []struct {
		Name   string `json:"name"`
		Ranges struct {
			Ranges []struct {
				End   int64 `json:"end"`
				Start int64 `json:"start"`
			} `json:"ranges"`
		} `json:"ranges"`
	} `json:"usage"`
}

// Valid returns if a fingerprint is valid
func (f Fingerprint) Valid() (bool, error) {
	status, err := f.Poll()

	if err != nil {
		return false, err
	}

	if status != 200 || f.Raw.Hash != f.Id {
		return false, fmt.Errorf("Jenkins says %s is Invalid or the Status is unknown", f.Id)
	}
	return true, nil
}

// ValidateForBuild validates the fingerprint for a build
func (f Fingerprint) ValidateForBuild(filename string, build *Build) (bool, error) {
	valid, err := f.Valid()
	if err != nil {
		return false, err
	}

	if valid {
		return true, nil
	}

	if f.Raw.FileName != filename {
		return false, errors.New("Filename does not Match")
	}
	if build != nil && f.Raw.Original.Name == build.Job.GetName() &&
		f.Raw.Original.Number == build.GetBuildNumber() {
		return true, nil
	}
	return false, nil
}

// GetInfo gets a fingerprint's info
func (f Fingerprint) GetInfo() (*fingerPrintResponse, error) {
	_, err := f.Poll()
	if err != nil {
		return nil, err
	}
	return f.Raw, nil
}

// Poll polls a fingerprint
func (f Fingerprint) Poll() (int, error) {
	response, err := f.Jenkins.Requester.GetJSON(f.Base+f.Id, f.Raw, nil)
	if err != nil {
		return 0, err
	}
	return response.StatusCode, nil
}
