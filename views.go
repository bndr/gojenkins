package gojenkins

type View struct {
	Raw     *viewResponse
	Jenkins *Jenkins
	Base    string
}

type viewResponse struct {
	Description string        `json:"description"`
	Jobs        []job         `json:"jobs"`
	Name        string        `json:"name"`
	Property    []interface{} `json:"property"`
	URL         string        `json:"url"`
}

var (
	LIST_VIEW      = "hudson.model.ListView"
	NESTED_VIEW    = "hudson.plugins.nested_view.NestedView"
	MY_VIEW        = "hudson.model.MyView"
	DASHBOARD_VIEW = "hudson.plugins.view.dashboard.Dashboard"
	PIPELINE_VIEW  = "au.com.centrumsystems.hudson.plugin.buildpipeline.BuildPipelineView"
)

// Returns True if successfully added Job, otherwise false
func (v *View) AddJob(name string) bool {
	url := "/addJobToView"
	qr := map[string]string{"name": name}
	resp := v.Jenkins.Requester.Post(v.Base+url, nil, nil, qr)
	return resp.StatusCode == 200
}

// Returns True if successfully deleted Job, otherwise false
func (v *View) DeleteJob(name string) bool {
	url := "/removeJobFromView"
	qr := map[string]string{"name": name}
	resp := v.Jenkins.Requester.Post(v.Base+url, nil, nil, qr)
	return resp.StatusCode == 200
}

func (v *View) Poll() int {
	v.Jenkins.Requester.GetJSON(v.Base, v.Raw, nil)
	return v.Jenkins.Requester.LastResponse.StatusCode
}
