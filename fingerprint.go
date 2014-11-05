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

func (f Fingerprint) Valid() bool {
	if f.Poll() != 200 || f.Raw.Hash != f.Id {
		Info.Printf("Jenkins says %s is Invalid or the Status is unknown", f.Id)
		return false
	}
	return true
}

func (f Fingerprint) ValidateForBuild(filename string, build *Build) bool {
	if f.Valid() {
		return true
	}

	if f.Raw.FileName != filename {
		return false
	}
	if build != nil && f.Raw.Original.Name == build.Job.GetName() &&
		f.Raw.Original.Number == build.GetBuildNumber() {
		return true
	}
	return false
}

func (f Fingerprint) GetInfo() *fingerPrintResponse {
	resp := f.Poll()
	if resp == 200 {
		return f.Raw
	}
	Error.Println("Jenkins returned status code for Fingerprint %s: %d", f.Id, resp)
	return nil
}

func (f Fingerprint) Poll() int {
	f.Jenkins.Requester.GetJSON(f.Base+f.Id, f.Raw, nil)
	return f.Jenkins.Requester.LastResponse.StatusCode
}
