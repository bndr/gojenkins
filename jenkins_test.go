package gojenkins

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	jenkins *Jenkins
	queueID int64
)

func createTestJobs(ctx context.Context) ([]*Job, error) {
	job_data := getFileAsString("job.xml")
	job1ID := "Job1_test"
	job2ID := "job2_test"
	job1, err := jenkins.CreateJob(ctx, job_data, job1ID)
	if err != nil {
		return nil, err
	}
	job2, err := jenkins.CreateJob(ctx, job_data, job2ID)
	if err != nil {
		return nil, err
	}

	return []*Job{job1, job2}, nil
}

func deleteJobs(ctx context.Context, j []*Job) error {
	errorsArr := []error{}
	for _, j := range j {
		ok, err := j.Delete(ctx)
		if !ok || err != nil {
			errorsArr = append(errorsArr, err)
		}
	}
	if len(errorsArr) > 0 {
		return errors.New("one or more jobs failed to delete")
	}
	return nil
}

func init() {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	jenkins = CreateJenkins(nil, "http://localhost:5000", "aidaleuc", "11d3433d67057c2e26bc273d1045b1e771")
	_, err := jenkins.Init(ctx)

	// More or less fail all other tests if we can't connect.
	if err != nil {
		panic(err)
	}

	jobs, err := jenkins.GetAllJobs(ctx)
	if err != nil {
		panic(err)
	}
	if err = deleteJobs(ctx, jobs); err != nil {
		panic(err)
	}

	views, err := jenkins.GetAllViews(ctx)
	if err != nil {
		panic(err)
	}
	for _, view := range views {
		if view.Base == "/view/all" {
			continue
		}
		if err := view.Delete(ctx); err != nil {
			panic(err)
		}
	}

	nodes, err := jenkins.GetAllNodes(ctx)
	if err != nil {
		panic(err)
	}
	for _, node := range nodes {
		if node.Raw.DisplayName == "Built-In Node" {
			continue
		}
		ok, err := node.Delete(ctx)
		if !ok || err != nil {
			panic("failed to delete node")
		}
	}
}

func TestCreateJobs(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}

	job1ID := "Job1_test"
	job2ID := "job2_test"
	ctx := context.Background()
	jobs, err := createTestJobs(ctx)
	if err != nil {
		t.Fatalf("failed to create jobs: %v", err)
	}
	job1 := jobs[0]
	defer job1.Delete(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, job1)
	assert.Equal(t, "Some Job Description", job1.GetDescription())
	assert.Equal(t, job1ID, job1.GetName())

	job2 := jobs[1]
	defer job2.Delete(ctx)
	assert.NotNil(t, job2)
	assert.Equal(t, "Some Job Description", job2.GetDescription())
	assert.Equal(t, job2ID, job2.GetName())
}

func TestCreateNodeSSHBasic(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	id1 := "ssh_test"
	ctx := context.Background()

	sshLauncher := DefaultSSHLauncher()
	node1, _ := jenkins.CreateNode(ctx, id1, 1, "Node 1 Description", "/var/lib/jenkins", "", sshLauncher)
	defer node1.Delete(ctx)
	ok, err := node1.IsJnlpAgent(ctx)
	if err != nil {
		t.Fatalf("failed to query jenkins about jnlp secreet")
	}
	assert.Equal(t, id1, node1.GetName())

	// Assert it is a ssh node!
	assert.Equal(t, ok, false)
}

func TestCreateNodeSSHAdvanced(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	id1 := "ssh_test"
	ctx := context.Background()

	host := "127.0.0.1"
	port := 26
	credential := ""
	timeouts := 25
	jvmOptions := "woop"
	javaPath := "home/bin/java"
	suffixPrefix := "worpp"
	sshLauncher := NewSSHLauncher(host, port, credential, timeouts, timeouts, timeouts, jvmOptions, javaPath, suffixPrefix, suffixPrefix)
	node1, err := jenkins.CreateNode(ctx, id1, 1, "Node 1 Description", "/var/lib/jenkins", "", sshLauncher)
	defer node1.Delete(ctx)
	if err != nil {
		t.Fatal("failed to create node")
	}

	config, err := node1.GetLauncherConfig(ctx)
	if err != nil {
		t.Fatal("failed to get config")
	}

	actualLauncher, ok := config.Launcher.Launcher.(*SSHLauncher)
	if !ok {
		t.Fatal("wrong type")
	}

	assert.Equal(t, actualLauncher.Class, SSHLauncherClass)
	assert.Equal(t, actualLauncher.Port, port)
	assert.Equal(t, actualLauncher.CredentialsId, credential)
	assert.Equal(t, actualLauncher.RetryWaitTime, timeouts)
	assert.Equal(t, actualLauncher.MaxNumRetries, timeouts)
	assert.Equal(t, actualLauncher.LaunchTimeoutSeconds, timeouts)
	assert.Equal(t, actualLauncher.JvmOptions, jvmOptions)
	assert.Equal(t, actualLauncher.JavaPath, javaPath)
	assert.Equal(t, actualLauncher.PrefixStartSlaveCmd, suffixPrefix)
	assert.Equal(t, actualLauncher.SuffixStartSlaveCmd, suffixPrefix)
}

func TestUpdateNodeSSH(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}

	// Setup
	id1 := "ssh_test"
	newNodeName := "ssh_test_new"
	ctx := context.Background()
	host := "127.0.0.1"
	port := 26
	credential := ""
	timeouts := 25
	jvmOptions := "woop"
	javaPath := "home/bin/java"
	description := "GORB"
	label := "donkey kong"
	remoteFS := "C:\\_jenkins"
	suffixPrefix := "worpp"
	executors := 15
	sshLauncher := NewSSHLauncher(host, port, credential, timeouts, timeouts, timeouts, jvmOptions, javaPath, suffixPrefix, suffixPrefix)
	node1, err := jenkins.CreateNode(ctx, id1, 1, "Node 1 Description", "/var/lib/jenkins", "", DefaultJNLPLauncher())
	if err != nil {
		t.Fatal("failed to create node")
	}

	// ACT
	newNode, err := node1.UpdateNode(ctx, newNodeName, executors, description, remoteFS, label, sshLauncher)
	defer newNode.Delete(ctx)
	if err != nil {
		t.Fatal("failed to update node")
	}
	config, err := newNode.GetLauncherConfig(ctx)
	if err != nil {
		t.Fatal("failed to get config")
	}

	newNodeInfo, err := newNode.Info(ctx)
	if err != nil {
		t.Fatal("failed to get node info")
	}

	actualLauncher, ok := config.Launcher.Launcher.(*SSHLauncher)
	if !ok {
		t.Fatal("wrong type")
	}

	// ASSERT
	// Test we updated all the things
	assert.Equal(t, newNode.GetName(), newNodeName)
	assert.Equal(t, len(newNodeInfo.Executors), executors)
	assert.Equal(t, config.Description, description)
	assert.Equal(t, config.NumExecutors, executors)
	assert.Equal(t, config.Label, label)
	assert.Equal(t, config.Name, newNodeName)
	assert.Equal(t, config.RemoteFS, remoteFS)
	assert.Equal(t, actualLauncher.Class, SSHLauncherClass)
	assert.Equal(t, actualLauncher.Port, port)
	assert.Equal(t, actualLauncher.CredentialsId, credential)
	assert.Equal(t, actualLauncher.RetryWaitTime, timeouts)
	assert.Equal(t, actualLauncher.MaxNumRetries, timeouts)
	assert.Equal(t, actualLauncher.LaunchTimeoutSeconds, timeouts)
	assert.Equal(t, actualLauncher.JvmOptions, jvmOptions)
	assert.Equal(t, actualLauncher.JavaPath, javaPath)
	assert.Equal(t, actualLauncher.PrefixStartSlaveCmd, suffixPrefix)
	assert.Equal(t, actualLauncher.SuffixStartSlaveCmd, suffixPrefix)
}

func TestGetNodeConfig(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	sshLauncher := DefaultSSHLauncher()
	node1, err := jenkins.CreateNode(ctx, "whatev", 1, "Node 1 Description", "/var/lib/jenkins", "", sshLauncher)
	defer node1.Delete(ctx)
	if err != nil {
		t.Fatal(err)
	}

	sshConfig, err := node1.GetLauncherConfig(ctx)
	if err != nil {
		t.Fatalf("error retrieving node conifg %v", err)
	}

	node2, err := jenkins.CreateNode(ctx, "whatev-2", 1, "", "C:\\_jenkins", "boop", nil)
	if err != nil {
		t.Fatalf("error retrieving node config %v", err)
	}

	jnlpConfig, err := node2.GetLauncherConfig(ctx)
	if err != nil {
		t.Fatalf("error retrieving config")
	}
	assert.NotNil(t, sshConfig.Launcher)
	assert.Equal(t, sshConfig.Launcher.Class, SSHLauncherClass)
	assert.NotNil(t, jnlpConfig.Launcher)
	assert.Equal(t, jnlpConfig.Launcher.Class, JNLPLauncherClass)
}

func TestCreateJNLPNodeBasic(t *testing.T) {
	ctx := context.Background()

	name := "whatever"
	executors := 5
	description := "A basic node"
	fs := "/var/jenkins"
	label := "tay tor"

	node, err := jenkins.CreateNode(ctx, name, executors, description, fs, label, nil)
	defer node.Delete(ctx)
	if err != nil {
		t.Fatal(err)
	}

	jnlpNode, err := node.IsJnlpAgent(ctx)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, node.GetName(), name)
	assert.Equal(t, jnlpNode, true)
}

func TestCreateJNLPNodeAdvanced(t *testing.T) {
	ctx := context.Background()

	name := "whatever"
	executors := 5
	description := "A basic node"
	fs := "/var/jenkins"
	label := "tay tor"

	launcher := NewJNLPLauncher(true, &WorkDirSettings{
		Disabled:               true,
		InternalDir:            "my-fake-remoting-dir",
		FailIfWorkDirIsMissing: true,
	})

	node, err := jenkins.CreateNode(ctx, name, executors, description, fs, label, launcher)
	defer node.Delete(ctx)
	if err != nil {
		t.Fatal(err)
	}

	jnlpNode, err := node.IsJnlpAgent(ctx)
	if err != nil {
		t.Fatal(err)
	}

	launcherConfig, err := node.GetLauncherConfig(ctx)
	if err != nil {
		t.Fatal(err)
	}

	jnlpLauncherConfig := launcherConfig.Launcher.Launcher.(*JNLPLauncher)
	assert.Equal(t, node.GetName(), name)
	assert.Equal(t, jnlpNode, true)
	assert.Equal(t, launcherConfig.NumExecutors, executors)
	assert.Equal(t, launcherConfig.Description, description)
	assert.Equal(t, launcherConfig.Label, label)
	assert.Equal(t, launcherConfig.RemoteFS, fs)
	assert.Equal(t, jnlpLauncherConfig.WebSocket, true)
	assert.Equal(t, jnlpLauncherConfig.WorkDirSettings.Disabled, true)
	assert.Equal(t, jnlpLauncherConfig.WorkDirSettings.InternalDir, "my-fake-remoting-dir")
	assert.Equal(t, jnlpLauncherConfig.WorkDirSettings.FailIfWorkDirIsMissing, true)
}

func TestUpdateNodeJNLP(t *testing.T) {
	ctx := context.Background()

	name := "whatever"
	executors := 5
	description := "A basic node"
	fs := "/var/jenkins"
	label := "tay tor"

	node, err := jenkins.CreateNode(ctx, "who-cares", 1, "", "C:\\_jenkins", "", nil)
	defer node.Delete(ctx)
	if err != nil {
		t.Fatal(err)
	}

	launcher := NewJNLPLauncher(true, &WorkDirSettings{
		Disabled:               true,
		InternalDir:            "my-fake-remoting-dir",
		FailIfWorkDirIsMissing: true,
	})

	newNode, err := node.UpdateNode(ctx, name, executors, description, fs, label, launcher)
	defer newNode.Delete(ctx)
	if err != nil {
		t.Fatal(err)
	}

	jnlpNode, err := newNode.IsJnlpAgent(ctx)
	if err != nil {
		t.Fatal(err)
	}

	launcherConfig, err := newNode.GetLauncherConfig(ctx)
	if err != nil {
		t.Fatal(err)
	}

	jnlpLauncherConfig := launcherConfig.Launcher.Launcher.(*JNLPLauncher)
	assert.Equal(t, newNode.GetName(), name)
	assert.Equal(t, jnlpNode, true)
	assert.Equal(t, launcherConfig.NumExecutors, executors)
	assert.Equal(t, launcherConfig.Description, description)
	assert.Equal(t, launcherConfig.Label, label)
	assert.Equal(t, launcherConfig.RemoteFS, fs)
	assert.Equal(t, jnlpLauncherConfig.WebSocket, true)
	assert.Equal(t, jnlpLauncherConfig.WorkDirSettings.Disabled, true)
	assert.Equal(t, jnlpLauncherConfig.WorkDirSettings.InternalDir, "my-fake-remoting-dir")
	assert.Equal(t, jnlpLauncherConfig.WorkDirSettings.FailIfWorkDirIsMissing, true)
}

func TestGetJNLPSecret(t *testing.T) {
	ctx := context.Background()
	node, err := jenkins.CreateNode(ctx, "who-cares", 1, "", "C:\\_jenkins", "", nil)
	defer node.Delete(ctx)
	if err != nil {
		t.Fatal(err)
	}
	secret, err := node.GetJNLPSecret(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, secret)
	assert.NotEqual(t, secret, "")
}

func TestDeleteNodes(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	id := "node4_test"

	ctx := context.Background()
	node, _ := jenkins.DeleteNode(ctx, id)
	assert.NotNil(t, node)
}

func TestCreateBuilds(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	if _, err := createTestJobs(ctx); err != nil {
		t.Fatalf("failed to create test jobs")
	}

	jobs, _ := jenkins.GetAllJobs(ctx)
	defer deleteJobs(ctx, jobs)
	for _, item := range jobs {
		queueID, _ = item.InvokeSimple(ctx, map[string]string{"params1": "param1"})
		item.Poll(ctx)
		isQueued, _ := item.IsQueued(ctx)
		assert.Equal(t, true, isQueued)
		time.Sleep(10 * time.Second)
		builds, _ := item.GetAllBuildIds(ctx)

		assert.True(t, (len(builds) > 0))

	}
}

func TestGetQueueItem(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	task, err := jenkins.GetQueueItem(ctx, queueID)
	if err != nil {
		t.Fatal(err)
	}
	if task.Raw == nil || task.Raw.ID != queueID {
		t.Fatal()
	}
}

func TestParseBuildHistory(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	r, err := os.Open("_tests/build_history.txt")
	if err != nil {
		panic(err)
	}
	history := parseBuildHistory(r)
	assert.True(t, len(history) == 3)
}

func TestCreateViews(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()

	jobs, err := createTestJobs(ctx)
	if err != nil {
		t.Fatalf("failed to create jobs")
	}
	defer deleteJobs(ctx, jobs)
	list_view, err := jenkins.CreateView(ctx, "test_list_view", LIST_VIEW)
	defer list_view.Delete(ctx)

	assert.Nil(t, err)
	assert.Equal(t, "test_list_view", list_view.GetName())
	assert.Equal(t, "", list_view.GetDescription())
	assert.Equal(t, 0, len(list_view.GetJobs()))

	my_view, err := jenkins.CreateView(ctx, "test_my_view", MY_VIEW)
	defer my_view.Delete(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "test_my_view", my_view.GetName())
	assert.Equal(t, "", my_view.GetDescription())
	assert.Equal(t, 2, len(my_view.GetJobs()))

}

func TestGetAllJobs(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}

	ctx := context.Background()
	testJobs, err := createTestJobs(ctx)
	defer deleteJobs(ctx, testJobs)
	if err != nil {
		t.Fatalf("failed to create test jobs")
	}
	jobs, _ := jenkins.GetAllJobs(ctx)
	assert.Equal(t, 2, len(jobs))
	assert.Equal(t, jobs[0].Raw.Color, "notbuilt")
}

func TestGetAllNodes(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()

	sshLauncher := DefaultSSHLauncher()
	node1, _ := jenkins.CreateNode(ctx, "whatever-id", 1, "Node 1 Description", "/var/lib/jenkins", "", sshLauncher)
	defer node1.Delete(ctx)
	nodes, _ := jenkins.GetAllNodes(ctx)
	assert.Equal(t, 2, len(nodes))
	assert.Equal(t, nodes[0].GetName(), "Built-In Node")
}

func TestGetAllBuilds(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()

	createdJobs, err := createTestJobs(ctx)
	defer deleteJobs(ctx, createdJobs)

	if err != nil {
		t.Fatalf("failed to create test jobs")
	}

	jobs, _ := jenkins.GetAllJobs(ctx)
	for _, item := range jobs {
		queueID, _ = item.InvokeSimple(ctx, map[string]string{"params1": "param1"})
		item.Poll(ctx)
		isQueued, _ := item.IsQueued(ctx)
		assert.Equal(t, true, isQueued)
		time.Sleep(10 * time.Second)
		builds, _ := item.GetAllBuildIds(ctx)

		assert.True(t, (len(builds) > 0))

	}
	builds, _ := jenkins.GetAllBuildIds(ctx, "Job1_test")
	for _, b := range builds {
		build, _ := jenkins.GetBuild(ctx, "Job1_test", b.Number)
		assert.Equal(t, "SUCCESS", build.GetResult())
	}
	assert.Equal(t, 1, len(builds))
}

func TestBuildMethods(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}

	ctx := context.Background()
	jobs, err := createTestJobs(ctx)
	defer deleteJobs(ctx, jobs)
	if err != nil {
		t.Fatalf("failed")
	}
	job, _ := jenkins.GetJob(ctx, "Job1_test")
	build, _ := job.GetLastBuild(ctx)
	params := build.GetParameters()
	assert.Equal(t, "params1", params[0].Name)
}

func TestGetSingleJob(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	job, _ := jenkins.GetJob(ctx, "Job1_test")
	isRunning, _ := job.IsRunning(ctx)
	config, err := job.GetConfig(ctx)
	assert.Nil(t, err)
	assert.Equal(t, false, isRunning)
	assert.Contains(t, config, "<project>")
}

func TestEnableDisableJob(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	job, _ := jenkins.GetJob(ctx, "Job1_test")
	result, _ := job.Disable(ctx)
	assert.Equal(t, true, result)
	result, _ = job.Enable(ctx)
	assert.Equal(t, true, result)
}

func TestCopyDeleteJob(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	job, _ := jenkins.GetJob(ctx, "Job1_test")
	jobCopy, _ := job.Copy(ctx, "Job1_test_copy")
	assert.Equal(t, jobCopy.GetName(), "Job1_test_copy")
	jobDelete, _ := job.Delete(ctx)
	assert.Equal(t, true, jobDelete)
}

func TestGetPlugins(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	plugins, _ := jenkins.GetPlugins(ctx, 3)
	assert.Equal(t, 10, plugins.Count())
}

func TestGetViews(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	views, _ := jenkins.GetAllViews(ctx)
	assert.Equal(t, len(views), 3)
	assert.Equal(t, len(views[0].Raw.Jobs), 2)
}

func TestGetSingleView(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	view, _ := jenkins.GetView(ctx, "All")
	view2, _ := jenkins.GetView(ctx, "test_list_view")
	assert.Equal(t, len(view.Raw.Jobs), 2)
	assert.Equal(t, len(view2.Raw.Jobs), 0)
	assert.Equal(t, view2.Raw.Name, "test_list_view")
}

func TestCreateFolder(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	folder1ID := "folder1_test"
	folder2ID := "folder2_test"

	folder1, err := jenkins.CreateFolder(ctx, folder1ID)
	assert.Nil(t, err)
	assert.NotNil(t, folder1)
	assert.Equal(t, folder1ID, folder1.GetName())

	folder2, err := jenkins.CreateFolder(ctx, folder2ID, folder1ID)
	assert.Nil(t, err)
	assert.NotNil(t, folder2)
	assert.Equal(t, folder2ID, folder2.GetName())
}

func TestCreateJobInFolder(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	jobName := "Job_test"
	job_data := getFileAsString("job.xml")

	job1, err := jenkins.CreateJobInFolder(ctx, job_data, jobName, "folder1_test")
	assert.Nil(t, err)
	assert.NotNil(t, job1)
	assert.Equal(t, "Some Job Description", job1.GetDescription())
	assert.Equal(t, jobName, job1.GetName())

	job2, err := jenkins.CreateJobInFolder(ctx, job_data, jobName, "folder1_test", "folder2_test")
	assert.Nil(t, err)
	assert.NotNil(t, job2)
	assert.Equal(t, "Some Job Description", job2.GetDescription())
	assert.Equal(t, jobName, job2.GetName())
}

func TestGetFolder(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	folder1ID := "folder1_test"
	folder2ID := "folder2_test"

	folder1, err := jenkins.GetFolder(ctx, folder1ID)
	assert.Nil(t, err)
	assert.NotNil(t, folder1)
	assert.Equal(t, folder1ID, folder1.GetName())

	folder2, err := jenkins.GetFolder(ctx, folder2ID, folder1ID)
	assert.Nil(t, err)
	assert.NotNil(t, folder2)
	assert.Equal(t, folder2ID, folder2.GetName())
}
func TestInstallPlugin(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()

	err := jenkins.InstallPlugin(ctx, "packer", "1.4")

	assert.Nil(t, err, "Could not install plugin")
}

func TestConcurrentRequests(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	for i := 0; i <= 16; i++ {
		go func() {
			jenkins.GetAllJobs(ctx)
			jenkins.GetAllViews(ctx)
			jenkins.GetAllNodes(ctx)
		}()
	}
}

func getFileAsString(path string) string {
	buf, err := ioutil.ReadFile("_tests/" + path)
	if err != nil {
		panic(err)
	}

	return string(buf)
}
