// Copyright 2015 Vadim Kravcenko
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
	"fmt"
	"strconv"
)

var (
	errJnlpSecret   = errors.New("failed to query jenkins for jnlp secret")
	errNotJnlpAgent = errors.New("agent is not a jnlp agent")
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

// Jenkins slave configuration. This is different than the data returned by the rest api.
// the rest api gives general information about the node, but this gives detailed information about the nodes
// actual configuration
type Slave struct {
	XMLName        xml.Name         `xml:"slave"`
	Name           string           `xml:"name"`
	Description    string           `xml:"description"`
	RemoteFS       string           `xml:"remoteFS"`
	NumExecutors   int              `xml:"numExecutors"`
	Mode           MODE             `xml:"mode"`
	Launcher       *CustomLauncher  `xml:"launcher"`
	Label          string           `xml:"label"`
	NodeProperties *NodeProperties  `xml:"nodeProperties,omitempty"`
}

// NodeOption is a function type for configuring a Node
type NodeOption func(*nodeConfig)

// nodeConfig holds configuration options for creating/updating nodes
type nodeConfig struct {
	name           string
	numExecutors   int
	description    string
	remoteFS       string
	label          string
	launcher       Launcher
	nodeProperties []NodeProperty
}

// WithNodeProperties sets the node properties
func WithNodeProperties(props ...NodeProperty) NodeOption {
	return func(c *nodeConfig) {
		c.nodeProperties = append(c.nodeProperties, props...)
	}
}

// WithLauncher sets the launcher for the node
func WithLauncher(launcher Launcher) NodeOption {
	return func(c *nodeConfig) {
		c.launcher = launcher
	}
}

// WithNumExecutors sets the number of executors
func WithNumExecutors(num int) NodeOption {
	return func(c *nodeConfig) {
		c.numExecutors = num
	}
}

// WithDescription sets the node description
func WithDescription(desc string) NodeOption {
	return func(c *nodeConfig) {
		c.description = desc
	}
}

// WithRemoteFS sets the remote filesystem root
func WithRemoteFS(fs string) NodeOption {
	return func(c *nodeConfig) {
		c.remoteFS = fs
	}
}

// WithLabel sets the node label
func WithLabel(label string) NodeOption {
	return func(c *nodeConfig) {
		c.label = label
	}
}

// GetConfig returns the launcher configuration for a given node.
// Only supports SSH and JNLP launchers.
func (n *Node) GetSlaveConfig(ctx context.Context) (*Slave, error) {
	// Gets the node configuration with launcher information.
	var sl Slave
	_, err := n.Jenkins.Requester.GetXML(ctx, n.Base+"/config.xml", &sl, nil)
	if err != nil {
		return nil, err
	}

	// Return the go struct.
	return &sl, nil
}

/*
Updates a Jenkins node with a new configuration
*/
// UpdateNode updates an existing Jenkins node.
//
// Deprecated: Use UpdateNodeV2 instead. UpdateNodeV2 provides a more flexible API
// with functional options and support for node properties (environment variables,
// tool locations, etc.).
//
// Example migration:
//   // Old way:
//   updated, err := node.UpdateNode(ctx, "new-name", 4, "updated", "/path", "label", launcher)
//
//   // New way:
//   updated, err := node.UpdateNodeV2(ctx, "new-name",
//       WithNumExecutors(4),
//       WithDescription("updated"),
//       WithRemoteFS("/path"),
//       WithLabel("label"),
//       WithLauncher(launcher),
//   )
func (n *Node) UpdateNode(ctx context.Context, name string, numExecutors int, description string, remoteFS string, label string, launchOptions Launcher) (*Node, error) {
	if launchOptions == nil {
		launchOptions = DefaultJNLPLauncher()
	}
	// Request to update the node. Uses a custom launcher for options specific to the node update.
	updateNodeRequest := &Slave{
		Name:         name,
		NumExecutors: numExecutors,
		Description:  description,
		RemoteFS:     remoteFS,
		Label:        label,
		Mode:         NORMAL,
		Launcher: &CustomLauncher{
			Class:    launchOptions.GetClass(),
			Launcher: launchOptions,
		},
	}

	// Converts the go struct to xml to send to Jenkins
	xmlBytes, err := xml.Marshal(updateNodeRequest)
	if err != nil {
		return nil, err
	}

	// Post an XML request.
	resp, err := n.Jenkins.Requester.PostXML(ctx, n.Base+"/config.xml", string(xmlBytes), nil, nil)
	if err != nil {
		return nil, err
	}

	// Get the updated node!
	newNode, err := n.Jenkins.GetNode(ctx, name)
	if err != nil {
		return nil, err
	}

	// Make sure the launcher was updated correctly
	// since the launchers are plugin based it is possible that Jenkins will return succcess
	// even when the launcher was not updated
	slaveConfig, err := newNode.GetSlaveConfig(ctx)
	if err != nil {
		return nil, err
	}

	// If the user requested a launch option to be set but the returned config
	// was inaccurate return an error
	if slaveConfig.Launcher == nil && launchOptions != nil {
		return nil, fmt.Errorf("failed to set launcher config. Config was null")
	}

	// Check for success status code.
	if resp.StatusCode < 400 {
		_, err := newNode.Poll(ctx)
		if err != nil {
			return nil, err
		}
		return newNode, nil
	}

	// If the response indicated non success throw an error.
	return nil, errors.New(strconv.Itoa(resp.StatusCode))
}

/*
UpdateNodeV2 updates a Jenkins node with the given name and options.
This version uses functional options pattern for better extensibility and maintainability.
It uses XML API to support setting node properties.
*/
func (n *Node) UpdateNodeV2(ctx context.Context, name string, options ...NodeOption) (*Node, error) {
	// Set defaults
	config := &nodeConfig{
		name:         name,
		numExecutors: 1,
		description:  "",
		remoteFS:     "/var/jenkins",
		label:        "",
		launcher:     DefaultJNLPLauncher(),
	}

	// Apply options
	for _, opt := range options {
		opt(config)
	}

	// Build the Slave XML structure
	updateNodeRequest := &Slave{
		Name:         config.name,
		NumExecutors: config.numExecutors,
		Description:  config.description,
		RemoteFS:     config.remoteFS,
		Label:        config.label,
		Mode:         NORMAL,
		Launcher: &CustomLauncher{
			Class:    config.launcher.GetClass(),
			Launcher: config.launcher,
		},
	}

	// Add node properties if provided
	if len(config.nodeProperties) > 0 {
		updateNodeRequest.NodeProperties = &NodeProperties{
			Properties: config.nodeProperties,
		}
	}

	// Converts the go struct to xml to send to Jenkins
	xmlBytes, err := xml.Marshal(updateNodeRequest)
	if err != nil {
		return nil, err
	}

	// Post an XML request.
	resp, err := n.Jenkins.Requester.PostXML(ctx, n.Base+"/config.xml", string(xmlBytes), nil, nil)
	if err != nil {
		return nil, err
	}

	// Get the updated node!
	newNode, err := n.Jenkins.GetNode(ctx, name)
	if err != nil {
		return nil, err
	}

	// Make sure the launcher was updated correctly
	slaveConfig, err := newNode.GetSlaveConfig(ctx)
	if err != nil {
		return nil, err
	}

	// If the user requested a launch option to be set but the returned config was inaccurate return an error
	if slaveConfig.Launcher == nil && config.launcher != nil {
		return nil, fmt.Errorf("failed to set launcher config. Config was null")
	}

	// Check for success status code.
	if resp.StatusCode < 400 {
		_, err := newNode.Poll(ctx)
		if err != nil {
			return nil, err
		}
		return newNode, nil
	}

	// If the response indicated non success throw an error.
	return nil, errors.New(strconv.Itoa(resp.StatusCode))
}

type jnlpSecret struct {
	Root            xml.Name `xml:"jnlp"`
	ApplicationDesc struct {
		Argument []string `xml:"argument"`
	} `xml:"application-desc"`
}

// Retrieves a JNLP secret for a JNLP node.
func (n *Node) GetJNLPSecret(ctx context.Context) (string, error) {
	jnlpAgent, err := n.IsJnlpAgent(ctx)
	if err != nil {
		return "", err
	}
	if !jnlpAgent {
		return "", errNotJnlpAgent
	}
	var jnlpResponse jnlpSecret
	jnlpAgentEndpoint := fmt.Sprintf("%s/%s", n.Base, "jenkins-agent.jnlp")
	_, err = n.Jenkins.Requester.GetXML(ctx, jnlpAgentEndpoint, &jnlpResponse, nil)

	if err != nil {
		return "", fmt.Errorf("%w for node %s Error details: %v", errJnlpSecret, n.GetName(), err)
	}

	// Make sure we are not going to index into a zero length array and panic.
	if len(jnlpResponse.ApplicationDesc.Argument) == 0 {
		return "", fmt.Errorf("%w for node %s : empty response", errJnlpSecret, n.GetName())
	}

	secret := jnlpResponse.ApplicationDesc.Argument[0]
	return secret, nil
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
