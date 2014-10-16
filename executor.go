package main

type executor struct {
	Raw       *executorResponse
	Requester *Requester
}
type executorResponse struct {
	AssignedLabels []struct{}  `json:"assignedLabels"`
	Description    interface{} `json:"description"`
	Jobs           []struct {
		Color string `json:"color"`
		Name  string `json:"name"`
		URL   string `json:"url"`
	} `json:"jobs"`
	Mode            string   `json:"mode"`
	NodeDescription string   `json:"nodeDescription"`
	NodeName        string   `json:"nodeName"`
	NumExecutors    float64  `json:"numExecutors"`
	OverallLoad     struct{} `json:"overallLoad"`
	PrimaryView     struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"primaryView"`
	QuietingDown   bool     `json:"quietingDown"`
	SlaveAgentPort float64  `json:"slaveAgentPort"`
	UnlabeledLoad  struct{} `json:"unlabeledLoad"`
	UseCrumbs      bool     `json:"useCrumbs"`
	UseSecurity    bool     `json:"useSecurity"`
	Views          []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"views"`
}
