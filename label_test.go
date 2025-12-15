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
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLabel_GetName(t *testing.T) {
	label := &Label{
		Raw: &LabelResponse{
			Name: "linux",
		},
	}

	assert.Equal(t, "linux", label.GetName())
}

func TestLabel_GetName_Empty(t *testing.T) {
	label := &Label{
		Raw: &LabelResponse{
			Name: "",
		},
	}

	assert.Equal(t, "", label.GetName())
}

func TestLabel_GetNodes(t *testing.T) {
	nodes := []LabelNode{
		{NodeName: "agent-1", NumExecutors: 4, Mode: "NORMAL"},
		{NodeName: "agent-2", NumExecutors: 2, Mode: "EXCLUSIVE"},
	}
	label := &Label{
		Raw: &LabelResponse{
			Name:  "linux",
			Nodes: nodes,
		},
	}

	result := label.GetNodes()
	assert.Equal(t, 2, len(result))
	assert.Equal(t, "agent-1", result[0].NodeName)
	assert.Equal(t, int64(4), result[0].NumExecutors)
	assert.Equal(t, "NORMAL", result[0].Mode)
	assert.Equal(t, "agent-2", result[1].NodeName)
	assert.Equal(t, int64(2), result[1].NumExecutors)
	assert.Equal(t, "EXCLUSIVE", result[1].Mode)
}

func TestLabel_GetNodes_Empty(t *testing.T) {
	label := &Label{
		Raw: &LabelResponse{
			Name:  "unused-label",
			Nodes: []LabelNode{},
		},
	}

	result := label.GetNodes()
	assert.Equal(t, 0, len(result))
}

func TestLabel_Poll_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	label := &Label{
		Jenkins: jenkins,
		Raw:     &LabelResponse{},
		Base:    "/label/linux",
	}

	status, err := label.Poll(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 200, status)
}

func TestLabel_Poll_Error(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).err = assert.AnError

	label := &Label{
		Jenkins: jenkins,
		Raw:     &LabelResponse{},
		Base:    "/label/linux",
	}

	_, err := label.Poll(context.Background())
	assert.Error(t, err)
}

func TestMODE_Constants(t *testing.T) {
	assert.Equal(t, MODE("NORMAL"), NORMAL)
	assert.Equal(t, string(EXCLUSIVE), "EXCLUSIVE")
}
