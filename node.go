package main

// Nodes

type Computers struct {
	BusyExecutors  int            `json:"busyExecutors"`
	Computers      []nodeResponse `json:"computer"`
	DisplayName    string         `json:"displayName"`
	TotalExecutors int            `json:"totalExecutors"`
}

type Node struct {
	Raw       *nodeResponse
	Jenkins   *Jenkins
	Requester *Requester
	Base      string
}

type nodeResponse struct {
	Actions             []interface{} `json:"actions"`
	DisplayName         string        `json:"displayName"`
	Executors           []struct{}    `json:"executors"`
	Icon                string        `json:"icon"`
	IconClassName       string        `json:"iconClassName"`
	Idle                bool          `json:"idle"`
	JnlpAgent           bool          `json:"jnlpAgent"`
	LaunchSupported     bool          `json:"launchSupported"`
	LoadStatistics      struct{}      `json:"loadStatistics"`
	ManualLaunchAllowed bool          `json:"manualLaunchAllowed"`
	MonitorData         struct {
		Hudson_NodeMonitors_ArchitectureMonitor interface{} `json:"hudson.node_monitors.ArchitectureMonitor"`
		Hudson_NodeMonitors_ClockMonitor        interface{} `json:"hudson.node_monitors.ClockMonitor"`
		Hudson_NodeMonitors_DiskSpaceMonitor    interface{} `json:"hudson.node_monitors.DiskSpaceMonitor"`
		Hudson_NodeMonitors_ResponseTimeMonitor struct {
			Average float64 `json:"average"`
		} `json:"hudson.node_monitors.ResponseTimeMonitor"`
		Hudson_NodeMonitors_SwapSpaceMonitor      interface{} `json:"hudson.node_monitors.SwapSpaceMonitor"`
		Hudson_NodeMonitors_TemporarySpaceMonitor interface{} `json:"hudson.node_monitors.TemporarySpaceMonitor"`
	} `json:"monitorData"`
	NumExecutors       float64       `json:"numExecutors"`
	Offline            bool          `json:"offline"`
	OfflineCause       struct{}      `json:"offlineCause"`
	OfflineCauseReason string        `json:"offlineCauseReason"`
	OneOffExecutors    []interface{} `json:"oneOffExecutors"`
	TemporarilyOffline bool          `json:"temporarilyOffline"`
}

func (n *Node) Info() *nodeResponse {
	return n.Raw
}

func (n *Node) GetName() string {
	return n.Raw.DisplayName
}

func (n *Node) Delete() bool {
	resp := n.Requester.Post(n.Base+"/doDelete", nil, nil, nil)
	return resp.StatusCode == 200
}

func (n *Node) IsOnline() bool {
	n.Poll()
	return !n.Raw.Offline
}

func (n *Node) IsTemporarilyOffline() bool {
	n.Poll()
	return n.Raw.TemporarilyOffline
}

func (n *Node) IsIdle() bool {
	n.Poll()
	return n.Raw.Idle
}

func (n *Node) IsJnlpAgent() bool {
	n.Poll()
	return n.Raw.JnlpAgent
}

func (n *Node) SetOnline() {
	n.Poll()
	if n.Raw.Offline && !n.Raw.TemporarilyOffline {
		panic("Node is Permanently offline, can't bring it up")
	}

	if n.Raw.Offline && n.Raw.TemporarilyOffline {
		n.ToggleTemporarilyOffline()
	}
}

func (n *Node) SetOffline() {
	if !n.Raw.Offline {
		n.ToggleTemporarilyOffline()
	}
}

func (n *Node) ToggleTemporarilyOffline(options ...interface{}) {
	state_before := n.IsTemporarilyOffline()
	qr := map[string]string{"offlineMessage": "requested from gojenkins"}
	if len(options) > 0 {
		qr["offlineMessage"] = options[0].(string)
	}
	n.Requester.Get(n.Base+"/toggleOffline", nil, qr)
	if state_before == n.IsTemporarilyOffline() {
		panic("Node state not changed")
	}
}

func (n *Node) Poll() int {
	n.Requester.Get(n.Base, n.Raw, nil)
	return n.Requester.LastResponse.StatusCode
}
