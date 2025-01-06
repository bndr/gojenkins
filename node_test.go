package gojenkins

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type nodeTC struct {
	launcher     Launcher
	name         string
	numExecutors int
	label        string
	description  string
	remoteFs     string
}

var (
	nodeTCs = []nodeTC{
		{
			launcher:     createMockJNLPLauncher(),
			name:         "CreateJNLPLauncher",
			numExecutors: 1,
			label:        "foo bar",
			description:  "mock",
			remoteFs:     "C:\\_jenkins",
		},
		{
			launcher:     createMockSSHLauncher(),
			name:         "CreateSSHLauncher",
			numExecutors: 1,
			label:        "foo bar",
			description:  "mock",
			remoteFs:     "C:\\_jenkins",
		},
		{
			launcher:     nil,
			name:         "CreateNilLauncher",
			numExecutors: 10,
			label:        "foo",
			description:  "mock",
			remoteFs:     "mock",
		},
		{
			launcher:     DefaultSSHLauncher(),
			name:         "CreateDefaultSSHLauncher",
			numExecutors: 1,
			label:        "",
			description:  "",
			remoteFs:     "",
		},
		{
			launcher:     DefaultJNLPLauncher(),
			name:         "CreateDefaultJNLPLauncher",
			numExecutors: 1,
			label:        "foo bar",
			description:  "mock",
			remoteFs:     "C:\\_jenkins",
		},
	}
)

func TestCreateNode(t *testing.T) {
	ctx := GetTestContext()

	for idx, tc := range nodeTCs {
		t.Run(tc.name, func(t *testing.T) {
			nodeName := fmt.Sprintf("node_%d", idx)
			node, err := J.CreateNode(ctx, nodeName, tc.numExecutors, tc.description, tc.remoteFs, tc.label, tc.launcher)
			assert.NoError(t, err)
			assert.NotNil(t, node)
			defer node.Delete(ctx)

			slaveConfig, err := node.GetSlaveConfig(ctx)
			assert.NoError(t, err)
			assert.NotNil(t, slaveConfig)
			assert.Equal(t, nodeName, node.GetName())
			assert.Equal(t, tc.description, slaveConfig.Description)
			assert.Equal(t, tc.label, slaveConfig.Label)
			assert.Equal(t, tc.numExecutors, slaveConfig.NumExecutors)
			assert.Equal(t, tc.remoteFs, slaveConfig.RemoteFS)
			checkLauncher(t, tc.launcher, slaveConfig.Launcher.Launcher)
		})
	}
}

func TestUpdateNode(t *testing.T) {
	ctx := GetTestContext()

	for idx, tc := range nodeTCs {
		t.Run(tc.name, func(t *testing.T) {
			nodeName := fmt.Sprintf("node_%d", idx)
			node, err := createMockNode(ctx, nodeName, nil)
			defer node.Delete(ctx)

			assert.NoError(t, err)
			assert.NotNil(t, node)

			newNodeName := fmt.Sprintf("node_new_%d", idx)
			newNode, err := node.UpdateNode(ctx, newNodeName, tc.numExecutors, tc.description, tc.remoteFs, tc.label, tc.launcher)
			defer newNode.Delete(ctx)
			assert.NoError(t, err)
			assert.NotNil(t, newNode)

			slaveConfig, err := newNode.GetSlaveConfig(ctx)
			assert.NoError(t, err)
			assert.NotNil(t, slaveConfig)
			assert.Equal(t, newNodeName, newNode.GetName())
			assert.Equal(t, tc.description, slaveConfig.Description)
			assert.Equal(t, tc.label, slaveConfig.Label)
			assert.Equal(t, tc.numExecutors, slaveConfig.NumExecutors)
			assert.Equal(t, tc.remoteFs, slaveConfig.RemoteFS)
			checkLauncher(t, tc.launcher, slaveConfig.Launcher.Launcher)
		})
	}
}

func TestGetJNLPSecret(t *testing.T) {
	type tc struct {
		launcher Launcher
		expected error
		name     string
	}

	tcs := []tc{
		{
			launcher: DefaultJNLPLauncher(),
			expected: nil,
			name:     "JNLPLauncherReturnsSecret",
		},
		{
			launcher: DefaultSSHLauncher(),
			expected: errNotJnlpAgent,
			name:     "SSHLauncherReturnsError",
		},
	}

	ctx := GetTestContext()
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			node, err := createMockNode(ctx, "mock", tc.launcher)
			defer node.Delete(ctx)
			assert.NoError(t, err)
			assert.NotNil(t, node)

			secret, err := node.GetJNLPSecret(ctx)
			assert.ErrorIs(t, err, tc.expected)

			if err == nil {
				assert.NotEmpty(t, secret)
			}
		})
	}

}

func TestDeleteNodes(t *testing.T) {
	ctx := GetTestContext()

	nodeName := "mock"
	node, err := createMockNode(ctx, nodeName, nil)
	require.NoError(t, err)
	require.NotNil(t, node)

	ok, err := node.Delete(ctx)
	assert.NoError(t, err, "error occurred deleting node")
	assert.True(t, ok, "node failed to delete")

	_, err = J.GetNode(ctx, nodeName)
	assert.ErrorIs(t, err, ErrNoNodeFound)
}

func TestGetAllNodes(t *testing.T) {
	ctx := GetTestContext()
	node1, err := createMockNode(ctx, "test_1", DefaultJNLPLauncher())
	defer node1.Delete(ctx)
	require.NoError(t, err)
	node2, err := createMockNode(ctx, "test_2", DefaultJNLPLauncher())
	defer node2.Delete(ctx)
	require.NoError(t, err)
	nodes, _ := J.GetAllNodes(ctx)
	require.Equal(t, 3, len(nodes))
	require.Equal(t, nodes[0].GetName(), "Built-In Node")
}

func checkLauncher(t *testing.T, expectedLauncher Launcher, actualLauncher Launcher) {
	t.Helper()

	// If nil launcher is passed the default is to fall back to JNLP
	if expectedLauncher == nil {
		expectedLauncher = DefaultJNLPLauncher()
	}
	launcherClass := actualLauncher.GetClass()
	expectedLauncherClass := expectedLauncher.GetClass()
	assert.Equal(t, expectedLauncherClass, launcherClass)

	switch launcherClass {
	case JNLPLauncherClass:
		jnlpExpected, ok := expectedLauncher.(*JNLPLauncher)
		assert.True(t, ok)
		jnlpActual, ok := actualLauncher.(*JNLPLauncher)
		assert.True(t, ok)
		compareJNLPLauncher(t, jnlpExpected, jnlpActual)
	case SSHLauncherClass:
		sshExpected, ok := expectedLauncher.(*SSHLauncher)
		assert.True(t, ok)
		sshActual, ok := actualLauncher.(*SSHLauncher)
		assert.True(t, ok)
		compareSSHLauncher(t, sshExpected, sshActual)
	default:
		t.Errorf("unrecognized launcher class %s", launcherClass)
	}

}

func compareSSHLauncher(t *testing.T, expected *SSHLauncher, actual *SSHLauncher) {
	t.Helper()

	assert.Equal(t, expected.Class, SSHLauncherClass)
	assert.Equal(t, expected.Port, actual.Port)
	assert.Equal(t, expected.CredentialsId, actual.CredentialsId)
	assert.Equal(t, expected.RetryWaitTime, actual.RetryWaitTime)
	assert.Equal(t, expected.MaxNumRetries, actual.MaxNumRetries)
	assert.Equal(t, expected.LaunchTimeoutSeconds, actual.LaunchTimeoutSeconds)
	assert.Equal(t, expected.JvmOptions, actual.JvmOptions)
	assert.Equal(t, expected.JavaPath, actual.JavaPath)
	assert.Equal(t, expected.PrefixStartSlaveCmd, actual.PrefixStartSlaveCmd)
	assert.Equal(t, expected.SuffixStartSlaveCmd, actual.SuffixStartSlaveCmd)
}

func compareJNLPLauncher(t *testing.T, expected *JNLPLauncher, actual *JNLPLauncher) {
	t.Helper()
	assert.Equal(t, expected.WebSocket, actual.WebSocket)
	assert.Equal(t, expected.WorkDirSettings.Disabled, actual.WorkDirSettings.Disabled)
	assert.Equal(t, expected.WorkDirSettings.InternalDir, actual.WorkDirSettings.InternalDir)
	assert.Equal(t, expected.WorkDirSettings.FailIfWorkDirIsMissing, actual.WorkDirSettings.FailIfWorkDirIsMissing)
}

func createMockSSHLauncher() *SSHLauncher {
	host := "127.0.0.1"
	port := 26
	credential := ""
	timeouts := 25
	jvmOptions := "woop"
	javaPath := "home/bin/java"
	suffixPrefix := "worpp"

	return NewSSHLauncher(host,
		port,
		credential,
		timeouts,
		timeouts,
		timeouts,
		jvmOptions,
		javaPath,
		suffixPrefix,
		suffixPrefix)
}

func createMockJNLPLauncher() *JNLPLauncher {
	return NewJNLPLauncher(true, &WorkDirSettings{
		Disabled:               true,
		InternalDir:            "/mock",
		FailIfWorkDirIsMissing: true,
	})
}

func createMockNode(ctx context.Context, name string, l Launcher) (*Node, error) {
	return J.CreateNode(ctx, name, 1, "Mock test node", "C:\\_jenkins", "best_label", l)
}
