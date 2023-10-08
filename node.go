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
	"encoding/xml"
	"errors"
	"log"
)

type LauncherClass string

const (
	JNLPLauncherClass LauncherClass = "hudson.slaves.JNLPLauncher"
	SSHLauncherClass  LauncherClass = "hudson.plugins.sshslaves.SSHLauncher"
)

// Nodes

type Computers struct {
	BusyExecutors  int             `json:"busyExecutors"`
	Computers      []*NodeResponse `json:"computer"`
	DisplayName    string          `json:"displayName"`
	TotalExecutors int             `json:"totalExecutors"`
}

type Node struct {
	Raw     *NodeResponse
	Jenkins *Jenkins
	Base    string
}

type NodeResponse struct {
	Class       string        `json:"_class"`
	Actions     []interface{} `json:"actions"`
	DisplayName string        `json:"displayName"`
	Executors   []struct {
		CurrentExecutable struct {
			Number    int    `json:"number"`
			URL       string `json:"url"`
			SubBuilds []struct {
				Abort             bool        `json:"abort"`
				Build             interface{} `json:"build"`
				BuildNumber       int         `json:"buildNumber"`
				Duration          string      `json:"duration"`
				Icon              string      `json:"icon"`
				JobName           string      `json:"jobName"`
				ParentBuildNumber int         `json:"parentBuildNumber"`
				ParentJobName     string      `json:"parentJobName"`
				PhaseName         string      `json:"phaseName"`
				Result            string      `json:"result"`
				Retry             bool        `json:"retry"`
				URL               string      `json:"url"`
			} `json:"subBuilds"`
		} `json:"currentExecutable"`
	} `json:"executors"`
	Icon                string   `json:"icon"`
	IconClassName       string   `json:"iconClassName"`
	Idle                bool     `json:"idle"`
	JnlpAgent           bool     `json:"jnlpAgent"`
	LaunchSupported     bool     `json:"launchSupported"`
	LoadStatistics      struct{} `json:"loadStatistics"`
	ManualLaunchAllowed bool     `json:"manualLaunchAllowed"`
	MonitorData         struct {
		Hudson_NodeMonitors_ArchitectureMonitor interface{} `json:"hudson.node_monitors.ArchitectureMonitor"`
		Hudson_NodeMonitors_ClockMonitor        interface{} `json:"hudson.node_monitors.ClockMonitor"`
		Hudson_NodeMonitors_DiskSpaceMonitor    interface{} `json:"hudson.node_monitors.DiskSpaceMonitor"`
		Hudson_NodeMonitors_ResponseTimeMonitor struct {
			Average int64 `json:"average"`
		} `json:"hudson.node_monitors.ResponseTimeMonitor"`
		Hudson_NodeMonitors_SwapSpaceMonitor      interface{} `json:"hudson.node_monitors.SwapSpaceMonitor"`
		Hudson_NodeMonitors_TemporarySpaceMonitor interface{} `json:"hudson.node_monitors.TemporarySpaceMonitor"`
	} `json:"monitorData"`
	NumExecutors       int64         `json:"numExecutors"`
	Offline            bool          `json:"offline"`
	OfflineCause       struct{}      `json:"offlineCause"`
	OfflineCauseReason string        `json:"offlineCauseReason"`
	OneOffExecutors    []interface{} `json:"oneOffExecutors"`
	TemporarilyOffline bool          `json:"temporarilyOffline"`
}

type Launcher interface{}

type CustomLauncher struct {
	XMLName  xml.Name `xml:"launcher"`
	Launcher Launcher `xml:"-"`
	Class    string   `xml:"class,attr"`
}

type SSHLauncher struct {
	XMLName              xml.Name `xml:"launcher"`
	Plugin               string   `xml:"plugin,attr"`
	Host                 string   `xml:"host"`
	Port                 int      `xml:"port"`
	CredentialsId        string   `xml:"credentialsId"`
	LaunchTimeoutSeconds int      `xml:"launchTimeoutSeconds"`
	MaxNumRetries        int      `xml:"maxNumRetries"`
	RetryWaitTime        int      `xml:"retryWaitTime"`
}

type WorkDirSettings struct {
	Disabled               bool   `xml:"disabled"`
	InternalDir            string `xml:"internalDir"`
	FailIfWorkDirIsMissing bool   `xml:"failIfWorkDirIsMissing"`
}
type JNLPLauncher struct {
	XMLName         xml.Name         `xml:"launcher"`
	WorkDirSettings *WorkDirSettings `xml:"workDirSettings"`
	WebSocket       bool             `xml:"webSocket"`
}

type Slave struct {
	XMLName        xml.Name        `xml:"slave"`
	Name           string          `xml:"name"`
	Description    string          `xml:"description"`
	RemoteFS       string          `xml:"remoteFS"`
	NumExecutors   int             `xml:"numExecutors"`
	Mode           MODE            `xml:"mode"`
	Launcher       *CustomLauncher `xml:"launcher"`
	Label          string          `xml:"label"`
	NodeProperties string          `xml:"nodeProperties"`
}

func (c *CustomLauncher) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	c.Class = ""
	for _, attr := range start.Attr {
		if attr.Name.Local == "class" {
			c.Class = attr.Value
			break
		}
	}
	switch c.Class {
	case string(SSHLauncherClass):
		var sshLauncher SSHLauncher
		if err := d.DecodeElement(&sshLauncher, &start); err != nil {
			return err
		}
		c.Launcher = sshLauncher
	case string(JNLPLauncherClass):
		var jnlpLauncher JNLPLauncher
		if err := d.DecodeElement(&jnlpLauncher, &start); err != nil {
			return err
		}
		c.Launcher = jnlpLauncher
	default:
		return errors.New("unknown launcher class")
	}
	return nil
}

func (c *CustomLauncher) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	switch l := c.Launcher.(type) {
	case *SSHLauncher:
		start.Name.Local = "launcher"
		start.Attr = []xml.Attr{{Name: xml.Name{Local: "class"}, Value: string(SSHLauncherClass)}}
		return e.EncodeElement(l, start)
	case *JNLPLauncher:
		start.Name.Local = "launcher"
		start.Attr = []xml.Attr{{Name: xml.Name{Local: "class"}, Value: string(JNLPLauncherClass)}}
		return e.EncodeElement(l, start)
	default:
		return errors.New("unsupported launcher type")
	}
}

// GetConfig returns the launcher configuration for a given node.
// Only supports SSH and JNLP launchers.
func (n *Node) GetLauncherConfig(ctx context.Context) (*Slave, error) {
	var sl Slave
	_, err := n.Jenkins.Requester.Get(ctx, n.Base+"/config.xml", &sl, nil)
	if err != nil {
		return nil, err
	}

	return &sl, nil
}

func (n *Node) UpdateNode(ctx context.Context, s *Slave) error {
	xmlBytes, err := xml.Marshal(s)
	if err != nil {
		return err
	}

	resp, err := n.Jenkins.Requester.PostXML(ctx, n.Base+"/config.xml", string(xmlBytes), nil, nil)
	if err != nil {
		return err
	}
	log.Print(resp.StatusCode)

	return nil
}

func (n *Node) Info(ctx context.Context) (*NodeResponse, error) {
	_, err := n.Poll(ctx)
	if err != nil {
		return nil, err
	}
	return n.Raw, nil
}

func (n *Node) GetName() string {
	return n.Raw.DisplayName
}

func (n *Node) Delete(ctx context.Context) (bool, error) {
	resp, err := n.Jenkins.Requester.Post(ctx, n.Base+"/doDelete", nil, nil, nil)
	if err != nil {
		return false, err
	}
	return resp.StatusCode == 200, nil
}

func (n *Node) IsOnline(ctx context.Context) (bool, error) {
	_, err := n.Poll(ctx)
	if err != nil {
		return false, err
	}
	return !n.Raw.Offline, nil
}

func (n *Node) IsTemporarilyOffline(ctx context.Context) (bool, error) {
	_, err := n.Poll(ctx)
	if err != nil {
		return false, err
	}
	return n.Raw.TemporarilyOffline, nil
}

func (n *Node) IsIdle(ctx context.Context) (bool, error) {
	_, err := n.Poll(ctx)
	if err != nil {
		return false, err
	}
	return n.Raw.Idle, nil
}

func (n *Node) IsJnlpAgent(ctx context.Context) (bool, error) {
	_, err := n.Poll(ctx)
	if err != nil {
		return false, err
	}
	return n.Raw.JnlpAgent, nil
}

func (n *Node) SetOnline(ctx context.Context) (bool, error) {
	_, err := n.Poll(ctx)

	if err != nil {
		return false, err
	}

	if n.Raw.Offline && !n.Raw.TemporarilyOffline {
		return false, errors.New("Node is Permanently offline, can't bring it up")
	}

	if n.Raw.Offline && n.Raw.TemporarilyOffline {
		return n.ToggleTemporarilyOffline(ctx)
	}

	return true, nil
}

func (n *Node) SetOffline(ctx context.Context, options ...interface{}) (bool, error) {
	if !n.Raw.Offline {
		return n.ToggleTemporarilyOffline(ctx, options...)
	}
	return false, errors.New("Node already Offline")
}

func (n *Node) ToggleTemporarilyOffline(ctx context.Context, options ...interface{}) (bool, error) {
	state_before, err := n.IsTemporarilyOffline(ctx)
	if err != nil {
		return false, err
	}
	qr := map[string]string{"offlineMessage": "requested from gojenkins"}
	if len(options) > 0 {
		qr["offlineMessage"] = options[0].(string)
	}
	_, err = n.Jenkins.Requester.Post(ctx, n.Base+"/toggleOffline", nil, nil, qr)
	if err != nil {
		return false, err
	}
	new_state, err := n.IsTemporarilyOffline(ctx)
	if err != nil {
		return false, err
	}
	if state_before == new_state {
		return false, errors.New("Node state not changed")
	}
	return true, nil
}

func (n *Node) Poll(ctx context.Context) (int, error) {
	response, err := n.Jenkins.Requester.GetJSON(ctx, n.Base, n.Raw, nil)
	if err != nil {
		return 0, err
	}
	return response.StatusCode, nil
}

func (n *Node) LaunchNodeBySSH(ctx context.Context) (int, error) {
	qr := map[string]string{
		"json":   "",
		"Submit": "Launch slave agent",
	}
	response, err := n.Jenkins.Requester.Post(ctx, n.Base+"/launchSlaveAgent", nil, nil, qr)
	if err != nil {
		return 0, err
	}
	return response.StatusCode, nil
}

func (n *Node) Disconnect(ctx context.Context) (int, error) {
	qr := map[string]string{
		"offlineMessage": "",
		"json":           makeJson(map[string]string{"offlineMessage": ""}),
		"Submit":         "Yes",
	}
	response, err := n.Jenkins.Requester.Post(ctx, n.Base+"/doDisconnect", nil, nil, qr)
	if err != nil {
		return 0, err
	}
	return response.StatusCode, nil
}

func (n *Node) GetLogText(ctx context.Context) (string, error) {
	var log string

	_, err := n.Jenkins.Requester.Post(ctx, n.Base+"/log", nil, nil, nil)
	if err != nil {
		return "", err
	}

	qr := map[string]string{"start": "0"}
	_, err = n.Jenkins.Requester.GetJSON(ctx, n.Base+"/logText/progressiveHtml/", &log, qr)
	if err != nil {
		return "", nil
	}

	return log, nil
}
