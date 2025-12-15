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

func TestPlugins_Count(t *testing.T) {
	plugins := &Plugins{
		Raw: &PluginResponse{
			Plugins: []Plugin{
				{ShortName: "git", LongName: "Git Plugin"},
				{ShortName: "maven", LongName: "Maven Integration Plugin"},
				{ShortName: "workflow-aggregator", LongName: "Pipeline"},
			},
		},
	}

	assert.Equal(t, 3, plugins.Count())
}

func TestPlugins_Count_Empty(t *testing.T) {
	plugins := &Plugins{
		Raw: &PluginResponse{
			Plugins: []Plugin{},
		},
	}

	assert.Equal(t, 0, plugins.Count())
}

func TestPlugins_Contains_ByLongName(t *testing.T) {
	plugins := &Plugins{
		Raw: &PluginResponse{
			Plugins: []Plugin{
				{ShortName: "git", LongName: "Git Plugin", Version: "4.11.0"},
				{ShortName: "maven", LongName: "Maven Integration Plugin", Version: "3.18"},
			},
		},
	}

	plugin := plugins.Contains("Git Plugin")
	assert.NotNil(t, plugin)
	assert.Equal(t, "git", plugin.ShortName)
	assert.Equal(t, "4.11.0", plugin.Version)
}

func TestPlugins_Contains_ByShortName(t *testing.T) {
	plugins := &Plugins{
		Raw: &PluginResponse{
			Plugins: []Plugin{
				{ShortName: "git", LongName: "Git Plugin", Version: "4.11.0"},
				{ShortName: "maven", LongName: "Maven Integration Plugin", Version: "3.18"},
			},
		},
	}

	plugin := plugins.Contains("maven")
	assert.NotNil(t, plugin)
	assert.Equal(t, "Maven Integration Plugin", plugin.LongName)
	assert.Equal(t, "3.18", plugin.Version)
}

func TestPlugins_Contains_NotFound(t *testing.T) {
	plugins := &Plugins{
		Raw: &PluginResponse{
			Plugins: []Plugin{
				{ShortName: "git", LongName: "Git Plugin"},
			},
		},
	}

	plugin := plugins.Contains("nonexistent-plugin")
	assert.Nil(t, plugin)
}

func TestPlugins_Poll_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	plugins := &Plugins{
		Jenkins: jenkins,
		Raw:     &PluginResponse{},
		Base:    "/pluginManager",
		Depth:   1,
	}

	status, err := plugins.Poll(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 200, status)
}

func TestPlugins_Poll_Error(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).err = assert.AnError

	plugins := &Plugins{
		Jenkins: jenkins,
		Raw:     &PluginResponse{},
		Base:    "/pluginManager",
		Depth:   1,
	}

	_, err := plugins.Poll(context.Background())
	assert.Error(t, err)
}
