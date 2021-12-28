package jenkins

import (
	"context"
	"errors"
	"fmt"
	"github.com/bndr/gojenkins"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

// Jenkins connection object
type Jenkins struct {
	Instance    *gojenkins.Jenkins
	Server      string
	JenkinsUser string
	Token       string
	Context     context.Context
}

// Config is focused in the configuration json file
type Config struct {
	Server         string `mapstructure: Server`
	JenkinsUser    string `mapstructure: JenkinsUser`
	Token          string `mapstructure: Token`
	ConfigPath     string
	ConfigFileName string
	ConfigFullPath string
}

// SetConfigPath set the default config path
//
// Args:
//
// Returns
//	string or error
func (j *Config) SetConfigPath(path string) {
	dir, file := filepath.Split(path)
	j.ConfigPath = dir
	j.ConfigFileName = file
	j.ConfigFullPath = j.ConfigPath + j.ConfigFileName
}

// CheckIfExists check if file exists
//
// Args:
//	path - string
//
// Returns
//	error
func (j *Config) CheckIfExists() error {
	var err error
	if _, err = os.Stat(j.ConfigFullPath); err == nil {
		return nil

	}
	return err
}

// LoadConfig read the JSON configuration from specified file
//
// Example file:
//
// $HOME/.config/jenkinsctl/config.json
//
//
// Args:
//
// Returns
//	nil or error
func (j *Config) LoadConfig() (config Config, err error) {
	viper.AddConfigPath(j.ConfigPath)
	viper.SetConfigName(j.ConfigFileName)
	viper.SetConfigType("json")
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}

// PluginsShow show all plugins installed and enabled
//
// Returns
//	nil or error
func (j *Jenkins) PluginsShow() {
	p, _ := j.Instance.GetPlugins(j.Context, 1)

	if len(p.Raw.Plugins) > 0 {
		fmt.Printf("Plugins Activated and Enabled 🚀\n")
		for _, p := range p.Raw.Plugins {
			if len(p.LongName) > 0 && p.Active && p.Enabled {
				fmt.Printf("    %s - %s ✅\n", p.LongName, p.Version)
			}
		}
	}
}

// DeleteJob will delete a job
//
// Args:
//	jobName - job name
//
// Returns:
//	error or nil
func (j *Jenkins) DeleteJob(jobName string) error {
	job, err := j.Instance.GetJob(j.Context, jobName)
	if err != nil {
		return err
	}

	_, err = job.Delete(j.Context)

	return err
}

// JobGetConfig get the configuration from job
//
// Args:
//	jobName - job name
//
// Returns:
//	error or nil
func (j *Jenkins) JobGetConfig(jobName string) error {
	job, err := j.Instance.GetJob(j.Context, jobName)
	if err != nil {
		return err
	}
	config, _ := job.GetConfig(j.Context)
	fmt.Println(config)
	return nil

}

// ShowBuildQueue show the Build Queue
//
// Args:
//
// Returns
//
func (j *Jenkins) ShowBuildQueue() error {
	queue, _ := j.Instance.GetQueue(j.Context)
	totalTasks := 0
	for i, item := range queue.Raw.Items {
		fmt.Printf("Name: %s\n", item.Task.Name)
		fmt.Printf("ID: %d\n", item.ID)
		j.ShowStatus(item.Task.Color)
		fmt.Printf("Pending: %v\n", item.Pending)
		fmt.Printf("Stuck: %v\n", item.Stuck)

		fmt.Printf("Why: %s\n", item.Why)
		fmt.Printf("URL: %s\n", item.Task.URL)
		fmt.Printf("\n")
		totalTasks = i + 1
	}
	fmt.Printf("Number of tasks in the build queue: %d\n", totalTasks)

	return nil
}

// ShowStatus will show the statys of object
// TIP: Meaning of collors:
// https://github.com/jenkinsci/jenkins/blob/5e9b451a11926e5b42d4a94612ca566de058f494/core/src/main/java/hudson/model/BallColor.java#L56
func (j *Jenkins) ShowStatus(object string) {
	switch object {
	case "blue":
		fmt.Printf("Status: ✅ Success\n")
		break
	case "red":
		fmt.Printf("Status: ❌ Failed\n")
		break
	case "red_anime", "blue_anime", "yellow_anime", "gray_anime", "notbuild_anime":
		fmt.Printf("Status: ⏳ In Progress\n")
		break
	case "notbuilt":
		fmt.Printf("Status: 🚧 Not Build\n")
		break
	default:
		if len(object) > 0 {
			fmt.Printf("Status: %s\n", object)
		}
	}
}

// GetLastCompletedBuild get last completed build
//
// Args:
//	jobName - Job Name
//
// Returns:
//	error or nil
func (j *Jenkins) GetLastCompletedBuild(jobName string) error {
	fmt.Printf("⏳ Collecting job information...\n")
	job, err := j.Instance.GetJob(j.Context, jobName)
	if err != nil {
		return errors.New("❌ unable to find the specific job")
	}
	build, err := job.GetLastCompletedBuild(j.Context)
	if err != nil {
		return errors.New("❌ unable to find the last completed build job")
	}

	if len(build.Job.Raw.LastBuild.URL) > 0 {
		fmt.Printf("✅ Last completed build Number: %d\n", build.Job.Raw.LastBuild.Number)
		fmt.Printf("✅ Last completed build URL: %s\n", build.Job.Raw.LastBuild.URL)
		fmt.Printf("✅ Parameters: %s\n", build.GetParameters())
	} else {
		fmt.Printf("No last completed build available for job: %s", jobName)
	}
	return nil
}

// CreateView will create a view
//
// Args:
//	viewname - view name
//	viewType - view type
//
// Returns
//	error or nil
func (j *Jenkins) CreateView(viewName string, viewType string) error {
	fmt.Printf("%s\n", viewType)
	_, err := j.Instance.CreateView(j.Context, viewName, viewType)
	if err != nil {
		return err
	}

	fmt.Printf("✅ View created: %s\n", viewName)
	return nil
}

// DownloadArtifacts will download artifacts
//
// Args:
//	jobName - job name
//	buildID - build ID
//	pathToSave - path to save artifact
//
// Returns:
//	error or nil
func (j *Jenkins) DownloadArtifacts(jobName string, buildID int64, pathToSave string) error {
	job, err := j.Instance.GetJob(j.Context, jobName)
	if err != nil {
		return errors.New("❌ unable to find the job")
	}
	build, err := job.GetBuild(j.Context, buildID)
	if err != nil {
		return errors.New("❌ unable to find the specific build id")
	}
	artifacts := build.GetArtifacts()

	if len(artifacts) <= 0 {
		fmt.Printf("No artifacts available for download\n")
		return nil
	}

	for _, a := range artifacts {
		fmt.Printf("Saving artifact %s in %s\n", a.FileName, pathToSave)
		_, err := a.SaveToDir(j.Context, pathToSave)
		if err != nil {
			return errors.New("❌ unable to download artifact")
		}
	}
	return nil
}

// GetLastUnstableBuild will get last unstable build
//
// Args:
//	jobName - Job Name
//
// Returns:
//	error or nil
func (j *Jenkins) GetLastUnstableBuild(jobName string) error {
	fmt.Printf("⏳ Collecting job information...\n")
	job, err := j.Instance.GetJob(j.Context, jobName)
	if err != nil {
		return errors.New("❌ unable to find the specific job")
	}
	build, err := job.GetLastBuild(j.Context)
	if err != nil {
		return errors.New("❌ unable to find the last unstable build job")
	}

	if len(build.Job.Raw.LastBuild.URL) > 0 {
		fmt.Printf("Last unstable build Number: %d\n", build.Job.Raw.LastBuild.Number)
		fmt.Printf("Last unstable build URL: %s\n", build.Job.Raw.LastBuild.URL)
		fmt.Printf("Parameters: %s\n", build.GetParameters())
	} else {
		fmt.Printf("No last unstable build available for job: %s", jobName)
	}
	return nil
}

// GetLastStableBuild will get last stable build
//
// Args:
//	jobName - Job Name
//
// Returns:
//	error or nil
func (j *Jenkins) GetLastStableBuild(jobName string) error {
	fmt.Printf("⏳ Collecting job information...\n")
	job, err := j.Instance.GetJob(j.Context, jobName)
	if err != nil {
		return errors.New("❌ unable to find the specific job")
	}
	build, err := job.GetLastStableBuild(j.Context)
	if err != nil {
		return errors.New("❌ unable to find the last stable build job")
	}

	if len(build.Job.Raw.LastBuild.URL) > 0 {
		fmt.Printf("✅ Last stable build Number: %d\n", build.Job.Raw.LastBuild.Number)
		fmt.Printf("✅ Last stable build URL: %s\n", build.Job.Raw.LastBuild.URL)
		fmt.Printf("✅ Parameters: %s\n", build.GetParameters())
	} else {
		fmt.Printf("No last stable build available for job: %s", jobName)
	}
	return nil
}

// GetLastBuild will get last build
//
// Args:
//	jobName - Job Name
//
// Returns:
//	error or nil
func (j *Jenkins) GetLastBuild(jobName string) error {
	fmt.Printf("⏳ Collecting job information...\n")
	job, err := j.Instance.GetJob(j.Context, jobName)
	if err != nil {
		return errors.New("❌ unable to find the specific job")
	}
	build, err := job.GetLastBuild(j.Context)
	if err != nil {
		return errors.New("❌ unable to find the last build job")
	}

	if len(build.Job.Raw.LastBuild.URL) > 0 {
		fmt.Printf("✅ Last build Number: %d\n", build.Job.Raw.LastBuild.Number)
		fmt.Printf("✅ Last build URL: %s\n", build.Job.Raw.LastBuild.URL)
		fmt.Printf("✅ Parameters: %s\n", build.GetParameters())
	} else {
		fmt.Printf("No last build available for job: %s", jobName)
	}
	return nil
}

// GetLastFailedBuild will get last failed build
//
// Args:
//	jobName - Job Name
//
// Returns:
//	error or nil
func (j *Jenkins) GetLastFailedBuild(jobName string) error {
	fmt.Printf("⏳ Collecting job information...\n")
	jobObj, err := j.Instance.GetJob(j.Context, jobName)
	if err != nil {
		return errors.New("❌ unable to find the specific job")
	}
	build, err := jobObj.GetLastFailedBuild(j.Context)
	if err != nil {
		return errors.New("❌ unable to get the last successful build")
	}
	if len(build.Job.Raw.LastBuild.URL) > 0 {
		fmt.Printf("Last Failed build Number: %d\n", build.Job.Raw.LastBuild.Number)
		fmt.Printf("Last Failed build URL: %s\n", build.Job.Raw.LastBuild.URL)
		fmt.Printf("Parameters: %s\n", build.GetParameters())
	} else {
		fmt.Printf("No last failed build available for job")
	}
	return nil
}

// AddJobToView will add a specific job to a view
func (j *Jenkins) AddJobToView(viewName string, jobName string) error {

	view, _ := j.Instance.GetView(j.Context, viewName)
	_, err := view.AddJob(j.Context, jobName)
	if err != nil {
		return err
	}

	return nil

}

// GetLastSuccessfulBuild will get last failed build
//
// Args:
//	jobName - Job Name
//
// Returns:
//	error or nil
func (j *Jenkins) GetLastSuccessfulBuild(jobName string) error {
	fmt.Printf("⏳ Collecting job information...\n")
	jobObj, err := j.Instance.GetJob(j.Context, jobName)
	if err != nil {
		return errors.New("❌ unable to find the specific job")
	}
	build, err := jobObj.GetLastSuccessfulBuild(j.Context)
	if err != nil {
		return errors.New("❌ unable to get the last successful build")
	}
	if len(build.Job.Raw.LastBuild.URL) > 0 {
		fmt.Printf("✅ Last Successful build Number: %d\n", build.Job.Raw.LastBuild.Number)
		fmt.Printf("✅ Last Successful build URL: %s\n", build.Job.Raw.LastBuild.URL)
		fmt.Printf("✅ Parameters: %s\n", build.GetParameters())
	} else {
		fmt.Printf("No last successful build available for job")
	}
	return nil
}

// ShowAllJobs will show all jobs
//
// Args:
//
// Returns:
//	error or nil
func (j *Jenkins) ShowAllJobs() error {
	jobs, err := j.Instance.GetAllJobs(j.Context)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		fmt.Printf("✅ %s\n", job.Raw.Name)
		j.ShowStatus(job.Raw.Color)
		fmt.Printf("%s\n", job.Raw.Description)
		fmt.Printf("%s\n", job.Raw.URL)
		fmt.Printf("\n")
	}
	return nil
}

// ShowViews will show all views
//
// Args:
//
// Returns:
// 	error or nil
func (j *Jenkins) ShowViews() error {
	views, err := j.Instance.GetView(j.Context, "All")
	if err != nil {
		return err
	}

	for _, view := range views.Raw.Jobs {
		fmt.Printf("✅ %s\n", view.Name)
		fmt.Printf("%s\n", view.Url)
		fmt.Printf("\n")
	}
	return nil
}

// getFileAsString
func getFileAsString(path string) (string, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

// CreateJob will create a job based on XML specification
//
// Args:
//	xmlFile	- Job described in XML format
//	jobName - Job Name
//
// Returns:
//	error or nil
func (j *Jenkins) CreateJob(xmlFile string, jobName string) error {
	jobData, err := getFileAsString(xmlFile)
	if err != nil {
		return err
	}

	_, err = j.Instance.CreateJob(j.Context, jobData, jobName)
	return err
}

// ShowNodes show all plugins installed and enabled
//
// Args:
//	showStatus - show only the
//
// Returns
//	code return, nil or error
func (j *Jenkins) ShowNodes(showStatus string) ([]string, error) {
	var hosts []string

	nodes, err := j.Instance.GetAllNodes(j.Context)
	if err != nil {
		return hosts, err
	}
	for _, node := range nodes {
		// Fetch Node Data
		node.Poll(j.Context)

		switch showStatus {

		case "offline":
			if node.Raw.Offline || node.Raw.TemporarilyOffline {
				fmt.Printf("❌ %s - offline\n", node.GetName())
				fmt.Printf("Reason: %s\n\n", node.Raw.OfflineCauseReason)
			}
			hosts = append(hosts, node.GetName())

		case "online":
			if !node.Raw.Offline {
				fmt.Printf("✅ %s - online\n", node.GetName())
			}
			if node.Raw.Idle {
				fmt.Printf("😴 %s - idle\n", node.GetName())
			}
			hosts = append(hosts, node.GetName())
		}
	}
	return hosts, nil
}

// Init will initilialize connection with jenkins server
//
// Args:
//
// Returns
//
func (j *Jenkins) Init(config Config) error {
	j.JenkinsUser = config.JenkinsUser
	j.Server = config.Server
	j.Token = config.Token
	j.Context = context.Background()

	j.Instance = gojenkins.CreateJenkins(
		nil,
		j.Server,
		j.JenkinsUser,
		j.Token)

	_, err := j.Instance.Init(j.Context)
	return err
}

// ServerInfo will show information regarding the server
//
// Args:
//
func (j *Jenkins) ServerInfo() error {
	j.Instance.Info(j.Context)
	fmt.Printf("✅ Connected with: %s\n", j.JenkinsUser)
	fmt.Printf("✅ Server: %s\n", j.Server)
	fmt.Printf("✅ Version: %s\n", j.Instance.Version)

	return nil
}

// serverReachable will do validation if the jenkins server
// is reachable
//
// Args:
//	string - Jenkins url
//
// Returns
//	nil or error
func serverReachable(url string) error {
	_, err := http.Get(url)
	if err != nil {
		return err
	}
	return nil

}
