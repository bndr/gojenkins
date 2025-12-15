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

import "context"

// Label represents a Jenkins label used to group nodes.
type Label struct {
	Raw     *LabelResponse
	Jenkins *Jenkins
	Base    string
}

// MODE represents the usage mode for a node.
type MODE string

// Node usage mode constants.
const (
	NORMAL    MODE = "NORMAL"
	EXCLUSIVE      = "EXCLUSIVE"
)

// LabelNode represents a node associated with a label.
type LabelNode struct {
	NodeName        string `json:"nodeName"`
	NodeDescription string `json:"nodeDescription"`
	NumExecutors    int64  `json:"numExecutors"`
	Mode            string `json:"mode"`
	Class           string `json:"_class"`
}

// LabelResponse represents the JSON response from the Jenkins API for a label.
type LabelResponse struct {
	Name           string      `json:"name"`
	Description    string      `json:"description"`
	Nodes          []LabelNode `json:"nodes"`
	Offline        bool        `json:"offline"`
	IdleExecutors  int64       `json:"idleExecutors"`
	BusyExecutors  int64       `json:"busyExecutors"`
	TotalExecutors int64       `json:"totalExecutors"`
}

// GetName returns the name of the label.
func (l *Label) GetName() string {
	return l.Raw.Name
}

// GetNodes returns all nodes associated with the label.
func (l *Label) GetNodes() []LabelNode {
	return l.Raw.Nodes
}

// Poll fetches the latest label data from Jenkins.
func (l *Label) Poll(ctx context.Context) (int, error) {
	response, err := l.Jenkins.Requester.GetJSON(ctx, l.Base, l.Raw, nil)
	if err != nil {
		return 0, err
	}
	return response.StatusCode, nil
}
