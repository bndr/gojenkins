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

func TestNode_GetName(t *testing.T) {
	node := &Node{
		Raw: &NodeResponse{
			DisplayName: "test-agent",
		},
	}

	assert.Equal(t, "test-agent", node.GetName())
}

func TestNode_IsOnline_True(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	node := &Node{
		Jenkins: jenkins,
		Raw: &NodeResponse{
			Offline: false,
		},
		Base: "/computer/test-agent",
	}

	online, err := node.IsOnline(context.Background())
	assert.NoError(t, err)
	assert.True(t, online)
}

func TestNode_IsOnline_False(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	node := &Node{
		Jenkins: jenkins,
		Raw: &NodeResponse{
			Offline: true,
		},
		Base: "/computer/test-agent",
	}

	online, err := node.IsOnline(context.Background())
	assert.NoError(t, err)
	assert.False(t, online)
}

func TestNode_IsTemporarilyOffline_True(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	node := &Node{
		Jenkins: jenkins,
		Raw: &NodeResponse{
			TemporarilyOffline: true,
		},
		Base: "/computer/test-agent",
	}

	tempOffline, err := node.IsTemporarilyOffline(context.Background())
	assert.NoError(t, err)
	assert.True(t, tempOffline)
}

func TestNode_IsTemporarilyOffline_False(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	node := &Node{
		Jenkins: jenkins,
		Raw: &NodeResponse{
			TemporarilyOffline: false,
		},
		Base: "/computer/test-agent",
	}

	tempOffline, err := node.IsTemporarilyOffline(context.Background())
	assert.NoError(t, err)
	assert.False(t, tempOffline)
}

func TestNode_IsIdle_True(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	node := &Node{
		Jenkins: jenkins,
		Raw: &NodeResponse{
			Idle: true,
		},
		Base: "/computer/test-agent",
	}

	idle, err := node.IsIdle(context.Background())
	assert.NoError(t, err)
	assert.True(t, idle)
}

func TestNode_IsIdle_False(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	node := &Node{
		Jenkins: jenkins,
		Raw: &NodeResponse{
			Idle: false,
		},
		Base: "/computer/test-agent",
	}

	idle, err := node.IsIdle(context.Background())
	assert.NoError(t, err)
	assert.False(t, idle)
}

func TestNode_IsJnlpAgent_True(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	node := &Node{
		Jenkins: jenkins,
		Raw: &NodeResponse{
			JnlpAgent: true,
		},
		Base: "/computer/test-agent",
	}

	isJnlp, err := node.IsJnlpAgent(context.Background())
	assert.NoError(t, err)
	assert.True(t, isJnlp)
}

func TestNode_IsJnlpAgent_False(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	node := &Node{
		Jenkins: jenkins,
		Raw: &NodeResponse{
			JnlpAgent: false,
		},
		Base: "/computer/test-agent",
	}

	isJnlp, err := node.IsJnlpAgent(context.Background())
	assert.NoError(t, err)
	assert.False(t, isJnlp)
}

func TestNode_Delete_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	node := &Node{
		Jenkins: jenkins,
		Raw:     &NodeResponse{},
		Base:    "/computer/test-agent",
	}

	success, err := node.Delete(context.Background())
	assert.NoError(t, err)
	assert.True(t, success)

	// Verify the correct endpoint was called
	mock := jenkins.Requester.(*MockRequester)
	assert.Equal(t, "/computer/test-agent/doDelete", mock.lastEndpoint)
}

func TestNode_Delete_Failure(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 404,
	}

	node := &Node{
		Jenkins: jenkins,
		Raw:     &NodeResponse{},
		Base:    "/computer/test-agent",
	}

	success, err := node.Delete(context.Background())
	assert.NoError(t, err)
	assert.False(t, success)
}

func TestNode_Poll_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	node := &Node{
		Jenkins: jenkins,
		Raw:     &NodeResponse{},
		Base:    "/computer/test-agent",
	}

	status, err := node.Poll(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 200, status)
}

func TestNode_Poll_Error(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).err = assert.AnError

	node := &Node{
		Jenkins: jenkins,
		Raw:     &NodeResponse{},
		Base:    "/computer/test-agent",
	}

	_, err := node.Poll(context.Background())
	assert.Error(t, err)
}

func TestNode_Info_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	rawResponse := &NodeResponse{
		DisplayName: "test-agent",
		Idle:        true,
		Offline:     false,
	}
	node := &Node{
		Jenkins: jenkins,
		Raw:     rawResponse,
		Base:    "/computer/test-agent",
	}

	info, err := node.Info(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, rawResponse, info)
}

func TestNode_Info_Error(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).err = assert.AnError

	node := &Node{
		Jenkins: jenkins,
		Raw:     &NodeResponse{},
		Base:    "/computer/test-agent",
	}

	_, err := node.Info(context.Background())
	assert.Error(t, err)
}

func TestNode_SetOnline_AlreadyOnline(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	node := &Node{
		Jenkins: jenkins,
		Raw: &NodeResponse{
			Offline:            false,
			TemporarilyOffline: false,
		},
		Base: "/computer/test-agent",
	}

	success, err := node.SetOnline(context.Background())
	assert.NoError(t, err)
	assert.True(t, success)
}

func TestNode_SetOnline_PermanentlyOffline(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	node := &Node{
		Jenkins: jenkins,
		Raw: &NodeResponse{
			Offline:            true,
			TemporarilyOffline: false,
		},
		Base: "/computer/test-agent",
	}

	success, err := node.SetOnline(context.Background())
	assert.Error(t, err)
	assert.False(t, success)
	assert.Contains(t, err.Error(), "Permanently offline")
}

func TestNode_SetOffline_AlreadyOffline(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	node := &Node{
		Jenkins: jenkins,
		Raw: &NodeResponse{
			Offline: true,
		},
		Base: "/computer/test-agent",
	}

	success, err := node.SetOffline(context.Background())
	assert.Error(t, err)
	assert.False(t, success)
	assert.Contains(t, err.Error(), "already Offline")
}

func TestNode_LaunchNodeBySSH_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	node := &Node{
		Jenkins: jenkins,
		Raw:     &NodeResponse{},
		Base:    "/computer/test-agent",
	}

	status, err := node.LaunchNodeBySSH(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 200, status)

	// Verify the correct endpoint was called
	mock := jenkins.Requester.(*MockRequester)
	assert.Equal(t, "/computer/test-agent/launchSlaveAgent", mock.lastEndpoint)
}

func TestNode_Disconnect_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	node := &Node{
		Jenkins: jenkins,
		Raw:     &NodeResponse{},
		Base:    "/computer/test-agent",
	}

	status, err := node.Disconnect(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 200, status)

	// Verify the correct endpoint was called
	mock := jenkins.Requester.(*MockRequester)
	assert.Equal(t, "/computer/test-agent/doDisconnect", mock.lastEndpoint)
}

func TestNode_ToggleTemporarilyOffline_Success(t *testing.T) {
	toggleCount := 0
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}
	jenkins.Requester.(*MockRequester).GetJSONFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		if nr, ok := response.(*NodeResponse); ok {
			// First call: offline=false, after toggle: offline=true
			nr.TemporarilyOffline = toggleCount > 0
			toggleCount++
		}
		return &http.Response{StatusCode: 200}, nil
	}

	node := &Node{
		Jenkins: jenkins,
		Raw: &NodeResponse{
			TemporarilyOffline: false,
		},
		Base: "/computer/test-agent",
	}

	success, err := node.ToggleTemporarilyOffline(context.Background())
	assert.NoError(t, err)
	assert.True(t, success)
}

func TestNode_ToggleTemporarilyOffline_WithMessage(t *testing.T) {
	toggleCount := 0
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}
	jenkins.Requester.(*MockRequester).GetJSONFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		if nr, ok := response.(*NodeResponse); ok {
			nr.TemporarilyOffline = toggleCount > 0
			toggleCount++
		}
		return &http.Response{StatusCode: 200}, nil
	}

	node := &Node{
		Jenkins: jenkins,
		Raw: &NodeResponse{
			TemporarilyOffline: false,
		},
		Base: "/computer/test-agent",
	}

	success, err := node.ToggleTemporarilyOffline(context.Background(), "Maintenance mode")
	assert.NoError(t, err)
	assert.True(t, success)
}

func TestNode_ToggleTemporarilyOffline_StateNotChanged(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}
	jenkins.Requester.(*MockRequester).GetJSONFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		if nr, ok := response.(*NodeResponse); ok {
			// State doesn't change
			nr.TemporarilyOffline = false
		}
		return &http.Response{StatusCode: 200}, nil
	}

	node := &Node{
		Jenkins: jenkins,
		Raw: &NodeResponse{
			TemporarilyOffline: false,
		},
		Base: "/computer/test-agent",
	}

	success, err := node.ToggleTemporarilyOffline(context.Background())
	assert.Error(t, err)
	assert.False(t, success)
	assert.Contains(t, err.Error(), "state not changed")
}

func TestNode_GetLogText_Success(t *testing.T) {
	expectedLog := "Agent connected\nAgent launched successfully"

	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}
	jenkins.Requester.(*MockRequester).GetJSONFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		if strPtr, ok := response.(*string); ok {
			*strPtr = expectedLog
		}
		return &http.Response{StatusCode: 200}, nil
	}

	node := &Node{
		Jenkins: jenkins,
		Raw:     &NodeResponse{},
		Base:    "/computer/test-agent",
	}

	log, err := node.GetLogText(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, expectedLog, log)
}

func TestNode_GetLogText_Error(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).err = assert.AnError

	node := &Node{
		Jenkins: jenkins,
		Raw:     &NodeResponse{},
		Base:    "/computer/test-agent",
	}

	log, err := node.GetLogText(context.Background())
	assert.Error(t, err)
	assert.Empty(t, log)
}
