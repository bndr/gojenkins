package main

// Nodes

type Node struct {
	Raw *nodeResponse
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

func (n *Node) Exists() {

}

func (n *Node) Delete() {

}

func (n *Node) Disable() {

}

func (n *Node) Enable() {

}

func (n *Node) Create() {

}

func (n *Node) IsOnline() {

}

func (n *Node) IsTemporarilyOffline() {

}

func (n *Node) IsIdle() {

}

func (n *Node) SetOnline() {

}

func (n *Node) SetOffline() {

}

func (n *Node) SetTemporarilyOffline() {

}
