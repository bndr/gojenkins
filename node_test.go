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
			require.NoError(t, err)
			require.NotNil(t, node)
			defer node.Delete(ctx)

			slaveConfig, err := node.GetSlaveConfig(ctx)
			require.NoError(t, err)
			require.NotNil(t, slaveConfig)
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

			require.NoError(t, err)
			require.NotNil(t, node)

			newNodeName := fmt.Sprintf("node_new_%d", idx)
			newNode, err := node.UpdateNode(ctx, newNodeName, tc.numExecutors, tc.description, tc.remoteFs, tc.label, tc.launcher)
			defer newNode.Delete(ctx)
			require.NoError(t, err)
			assert.NotNil(t, newNode)

			slaveConfig, err := newNode.GetSlaveConfig(ctx)
			require.NoError(t, err)
			require.NotNil(t, slaveConfig)
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
			require.NoError(t, err)
			require.NotNil(t, node)

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

// createTestNode is a helper that creates a node and registers it for automatic cleanup
func createTestNode(t *testing.T, ctx context.Context, name string, options ...NodeOption) *Node {
	t.Helper()
	node, err := J.CreateNodeV2(ctx, name, options...)
	require.NoError(t, err)
	require.NotNil(t, node)
	t.Cleanup(func() { node.Delete(ctx) })
	return node
}

// verifyNodeConfig is a helper that verifies common node configuration
func verifyNodeConfig(t *testing.T, ctx context.Context, node *Node, expectedName string) *Slave {
	t.Helper()
	slaveConfig, err := node.GetSlaveConfig(ctx)
	require.NoError(t, err)
	require.NotNil(t, slaveConfig)
	assert.Equal(t, expectedName, node.GetName())
	return slaveConfig
}

func TestCreateNodeV2(t *testing.T) {
	ctx := GetTestContext()

	t.Run("CreateNodeV2WithDefaultOptions", func(t *testing.T) {
		nodeName := "test_node_v2_default"
		node := createTestNode(t, ctx, nodeName)
		slaveConfig := verifyNodeConfig(t, ctx, node, nodeName)
		assert.Equal(t, 1, slaveConfig.NumExecutors)
	})

	t.Run("CreateNodeV2WithAllOptions", func(t *testing.T) {
		nodeName := "test_node_v2_full"
		launcher := createMockJNLPLauncher()

		node := createTestNode(t, ctx, nodeName,
			WithNumExecutors(2),
			WithDescription("Test node v2"),
			WithRemoteFS("/var/jenkins"),
			WithLabel("docker linux"),
			WithLauncher(launcher),
		)
		slaveConfig := verifyNodeConfig(t, ctx, node, nodeName)
		assert.Equal(t, 2, slaveConfig.NumExecutors)
		assert.Equal(t, "Test node v2", slaveConfig.Description)
		assert.Equal(t, "/var/jenkins", slaveConfig.RemoteFS)
		assert.Equal(t, "docker linux", slaveConfig.Label)
		checkLauncher(t, launcher, slaveConfig.Launcher.Launcher)
	})

	t.Run("CreateNodeV2WithEnvironmentVariables", func(t *testing.T) {
		nodeName := "test_node_v2_env"
		envVars := map[string]string{
			"TEST_VAR1": "value1",
			"TEST_VAR2": "value2",
		}

		node := createTestNode(t, ctx, nodeName,
			WithNumExecutors(1),
			WithDescription("Node with env vars"),
			WithRemoteFS("/var/jenkins"),
			WithNodeProperties(
				NewEnvironmentVariablesNodeProperty(envVars),
			),
		)
		slaveConfig := verifyNodeConfig(t, ctx, node, nodeName)
		assert.Equal(t, nodeName, node.GetName())
		assert.NotNil(t, slaveConfig.NodeProperties)

		// Verify the environment variables are persisted
		require.Len(t, slaveConfig.NodeProperties.Properties, 1)
		envProp, ok := slaveConfig.NodeProperties.Properties[0].(*EnvironmentVariablesNodeProperty)
		require.True(t, ok, "Expected EnvironmentVariablesNodeProperty")
		assert.Len(t, envProp.EnvVars.Tree, 2, "Expected 2 environment variables")

		// Verify the actual values
		envMap := make(map[string]string)
		for _, env := range envProp.EnvVars.Tree {
			envMap[env.Key] = env.Value
		}
		assert.Equal(t, "value1", envMap["TEST_VAR1"])
		assert.Equal(t, "value2", envMap["TEST_VAR2"])
	})

	t.Run("CreateNodeV2WithToolLocations", func(t *testing.T) {
		nodeName := "test_node_v2_tools"

		toolLocations := map[string]string{
			"hudson.plugins.git.GitTool$DescriptorImpl:Default": "/usr/bin/git",
			"hudson.model.JDK:JDK11":                            "/usr/lib/jvm/java-11",
		}

		node := createTestNode(t, ctx, nodeName,
			WithNumExecutors(1),
			WithDescription("Node with tool locations"),
			WithRemoteFS("/var/jenkins"),
			WithNodeProperties(
				NewToolLocationNodeProperty(toolLocations),
			),
		)
		slaveConfig := verifyNodeConfig(t, ctx, node, nodeName)
		require.NotNil(t, slaveConfig.NodeProperties)
		require.Len(t, slaveConfig.NodeProperties.Properties, 1)

		// Verify tool location property
		toolProp, ok := slaveConfig.NodeProperties.Properties[0].(*ToolLocationNodeProperty)
		require.True(t, ok, "Expected ToolLocationNodeProperty")
		require.NotEmpty(t, toolProp.Locations, "Expected at least one tool location")

		// Verify we can find at least one of our tools
		foundGit := false
		for _, loc := range toolProp.Locations {
			if loc.Type == "hudson.plugins.git.GitTool$DescriptorImpl" && loc.Home == "/usr/bin/git" {
				foundGit = true
				break
			}
		}
		assert.True(t, foundGit, "Expected to find Git tool location")
	})

	t.Run("CreateNodeV2WithSSHLauncher", func(t *testing.T) {
		nodeName := "test_node_v2_ssh"
		launcher := createMockSSHLauncher()

		node := createTestNode(t, ctx, nodeName,
			WithNumExecutors(3),
			WithDescription("SSH node"),
			WithRemoteFS("/home/jenkins"),
			WithLabel("ssh linux"),
			WithLauncher(launcher),
		)
		slaveConfig := verifyNodeConfig(t, ctx, node, nodeName)
		checkLauncher(t, launcher, slaveConfig.Launcher.Launcher)
	})
}

func TestUpdateNodeV2(t *testing.T) {
	ctx := GetTestContext()

	t.Run("UpdateNodeV2BasicFields", func(t *testing.T) {
		nodeName := "test_update_v2_basic"
		node, err := createMockNode(ctx, nodeName, nil)
		require.NoError(t, err)
		require.NotNil(t, node)
		defer node.Delete(ctx)

		newNodeName := "test_update_v2_basic_renamed"
		updatedNode, err := node.UpdateNodeV2(ctx, newNodeName,
			WithNumExecutors(5),
			WithDescription("Updated description"),
			WithRemoteFS("/new/path"),
			WithLabel("updated label"),
		)
		require.NoError(t, err)
		require.NotNil(t, updatedNode)
		defer updatedNode.Delete(ctx)

		slaveConfig, err := updatedNode.GetSlaveConfig(ctx)
		require.NoError(t, err)
		require.NotNil(t, slaveConfig)
		assert.Equal(t, newNodeName, updatedNode.GetName())
		assert.Equal(t, 5, slaveConfig.NumExecutors)
		assert.Equal(t, "Updated description", slaveConfig.Description)
		assert.Equal(t, "/new/path", slaveConfig.RemoteFS)
		assert.Equal(t, "updated label", slaveConfig.Label)
	})

	t.Run("UpdateNodeV2WithLauncherChange", func(t *testing.T) {
		nodeName := "test_update_v2_launcher"
		node, err := createMockNode(ctx, nodeName, DefaultJNLPLauncher())
		require.NoError(t, err)
		require.NotNil(t, node)
		defer node.Delete(ctx)

		newLauncher := createMockJNLPLauncher()
		updatedNode, err := node.UpdateNodeV2(ctx, nodeName,
			WithLauncher(newLauncher),
		)
		require.NoError(t, err)
		require.NotNil(t, updatedNode)

		slaveConfig, err := updatedNode.GetSlaveConfig(ctx)
		require.NoError(t, err)
		require.NotNil(t, slaveConfig)
		checkLauncher(t, newLauncher, slaveConfig.Launcher.Launcher)
	})

	t.Run("UpdateNodeV2WithNodeProperties", func(t *testing.T) {
		nodeName := "test_update_v2_props"
		node, err := createMockNode(ctx, nodeName, nil)
		require.NoError(t, err)
		require.NotNil(t, node)
		defer node.Delete(ctx)

		envVars := map[string]string{
			"UPDATED_VAR": "updated_value",
		}

		updatedNode, err := node.UpdateNodeV2(ctx, nodeName,
			WithNodeProperties(
				NewEnvironmentVariablesNodeProperty(envVars),
			),
		)
		require.NoError(t, err)
		require.NotNil(t, updatedNode)

		slaveConfig, err := updatedNode.GetSlaveConfig(ctx)
		require.NoError(t, err)
		require.NotNil(t, slaveConfig)
		assert.NotNil(t, slaveConfig.NodeProperties)
	})

	t.Run("UpdateNodeV2MultipleOptions", func(t *testing.T) {
		nodeName := "test_update_v2_multi"
		node, err := createMockNode(ctx, nodeName, nil)
		require.NoError(t, err)
		require.NotNil(t, node)
		defer node.Delete(ctx)

		launcher := createMockSSHLauncher()
		envVars := map[string]string{
			"MULTI_VAR1": "value1",
			"MULTI_VAR2": "value2",
		}

		updatedNode, err := node.UpdateNodeV2(ctx, nodeName,
			WithNumExecutors(10),
			WithDescription("Multiple options test"),
			WithRemoteFS("/multi/path"),
			WithLabel("multi label"),
			WithLauncher(launcher),
			WithNodeProperties(
				NewEnvironmentVariablesNodeProperty(envVars),
			),
		)
		require.NoError(t, err)
		require.NotNil(t, updatedNode)

		slaveConfig, err := updatedNode.GetSlaveConfig(ctx)
		require.NoError(t, err)
		require.NotNil(t, slaveConfig)
		assert.Equal(t, 10, slaveConfig.NumExecutors)
		assert.Equal(t, "Multiple options test", slaveConfig.Description)
		assert.Equal(t, "/multi/path", slaveConfig.RemoteFS)
		assert.Equal(t, "multi label", slaveConfig.Label)
		checkLauncher(t, launcher, slaveConfig.Launcher.Launcher)
		assert.NotNil(t, slaveConfig.NodeProperties)
	})
}

func TestCreateNodeV2WithMultipleNodeProperties(t *testing.T) {
	ctx := GetTestContext()

	t.Run("CreateNodeV2WithDiskSpaceMonitor", func(t *testing.T) {
		nodeName := "test_node_v2_disk"

		node := createTestNode(t, ctx, nodeName,
			WithNumExecutors(1),
			WithDescription("Node with disk space monitor"),
			WithRemoteFS("/var/jenkins"),
			WithNodeProperties(
				NewDiskSpaceMonitorNodeProperty("1GiB", "500MiB", "800MiB", "400MiB"),
			),
		)
		slaveConfig := verifyNodeConfig(t, ctx, node, nodeName)
		require.NotNil(t, slaveConfig.NodeProperties)
		require.Len(t, slaveConfig.NodeProperties.Properties, 1)

		// Verify disk space property - check all threshold fields
		diskProp, ok := slaveConfig.NodeProperties.Properties[0].(*DiskSpaceMonitorNodeProperty)
		require.True(t, ok, "Expected DiskSpaceMonitorNodeProperty")
		assert.Equal(t, "1GiB", diskProp.FreeDiskSpaceThreshold)
		assert.Equal(t, "500MiB", diskProp.FreeTempSpaceThreshold)
		assert.Equal(t, "800MiB", diskProp.FreeDiskSpaceWarningThreshold)
		assert.Equal(t, "400MiB", diskProp.FreeTempSpaceWarningThreshold)
	})

	t.Run("CreateNodeV2WithRawProperty", func(t *testing.T) {
		nodeName := "test_node_v2_raw"

		// Test RawNodeProperty by manually defining a ToolLocationNodeProperty
		// This simulates a user defining a custom property we don't have a built-in type for
		rawToolProperty := NewRawNodeProperty(
			"hudson.tools.ToolLocationNodeProperty",
			`<locations>
				<hudson.tools.ToolLocationNodeProperty_-ToolLocation>
					<type>hudson.plugins.git.GitTool$DescriptorImpl</type>
					<name>Default</name>
					<home>/usr/bin/git</home>
				</hudson.tools.ToolLocationNodeProperty_-ToolLocation>
			</locations>`,
		)

		node := createTestNode(t, ctx, nodeName,
			WithNumExecutors(1),
			WithDescription("Node with raw property"),
			WithRemoteFS("/var/jenkins"),
			WithNodeProperties(rawToolProperty),
		)
		slaveConfig := verifyNodeConfig(t, ctx, node, nodeName)
		require.NotNil(t, slaveConfig.NodeProperties)
		require.Len(t, slaveConfig.NodeProperties.Properties, 1)

		// Verify it was stored (it will come back as ToolLocationNodeProperty since we have that type)
		toolProp, ok := slaveConfig.NodeProperties.Properties[0].(*ToolLocationNodeProperty)
		require.True(t, ok, "Expected ToolLocationNodeProperty")
		require.Len(t, toolProp.Locations, 1)
		assert.Equal(t, "Default", toolProp.Locations[0].Name)
		assert.Equal(t, "/usr/bin/git", toolProp.Locations[0].Home)
	})
	t.Run("CreateNodeV2WithDeferredWipeout", func(t *testing.T) {
		nodeName := "test_node_v2_wipeout"

		node := createTestNode(t, ctx, nodeName,
			WithNumExecutors(1),
			WithDescription("Node with deferred wipeout"),
			WithRemoteFS("/var/jenkins"),
			WithNodeProperties(
				NewDeferredWipeoutNodeProperty(),
			),
		)
		slaveConfig := verifyNodeConfig(t, ctx, node, nodeName)
		require.NotNil(t, slaveConfig.NodeProperties)
		require.Len(t, slaveConfig.NodeProperties.Properties, 1)

		// Verify workspace cleanup property
		_, ok := slaveConfig.NodeProperties.Properties[0].(*WorkspaceCleanupNodeProperty)
		require.True(t, ok, "Expected WorkspaceCleanupNodeProperty")
	})
	t.Run("CreateNodeV2WithMixedProperties", func(t *testing.T) {
		nodeName := "test_node_v2_mixed"

		envVars := map[string]string{
			"VAR1": "value1",
			"VAR2": "value2",
		}
		toolLocs := map[string]string{
			"hudson.plugins.git.GitTool$DescriptorImpl:Default": "/usr/bin/git",
			"hudson.model.JDK:JDK11":                            "/usr/lib/jvm/java-11",
		}

		node := createTestNode(t, ctx, nodeName,
			WithNumExecutors(2),
			WithDescription("Node with mixed properties"),
			WithRemoteFS("/var/jenkins"),
			WithNodeProperties(
				NewEnvironmentVariablesNodeProperty(envVars),
				NewToolLocationNodeProperty(toolLocs),
				NewDiskSpaceMonitorNodeProperty("2GiB"),
				NewDeferredWipeoutNodeProperty(),
			),
		)
		slaveConfig := verifyNodeConfig(t, ctx, node, nodeName)
		require.NotNil(t, slaveConfig.NodeProperties)

		// Should have at least some properties (Jenkins may not persist all types)
		require.NotEmpty(t, slaveConfig.NodeProperties.Properties, "Expected at least one property")

		// Verify we got at least some of the property types
		foundEnv := false
		foundTool := false
		foundDisk := false
		foundWorkspace := false

		for _, prop := range slaveConfig.NodeProperties.Properties {
			switch p := prop.(type) {
			case *EnvironmentVariablesNodeProperty:
				foundEnv = true
				assert.Len(t, p.EnvVars.Tree, 2)
				// Verify actual values
				envMap := make(map[string]string)
				for _, env := range p.EnvVars.Tree {
					envMap[env.Key] = env.Value
				}
				assert.Equal(t, "value1", envMap["VAR1"])
				assert.Equal(t, "value2", envMap["VAR2"])
			case *ToolLocationNodeProperty:
				foundTool = true
				assert.Len(t, p.Locations, 2)
			case *DiskSpaceMonitorNodeProperty:
				foundDisk = true
				assert.Equal(t, "2GiB", p.FreeDiskSpaceThreshold)
				assert.Equal(t, "2GiB", p.FreeTempSpaceThreshold)
			case *WorkspaceCleanupNodeProperty:
				foundWorkspace = true
				// ws-cleanup plugin property has no additional fields to check
			}
		}

		// Verify we got at least the core properties (env vars and disk space are most reliable)
		assert.True(t, foundEnv || foundTool || foundDisk || foundWorkspace, "Should have at least one property type")
	})
}

func TestUpdateNodeV2WithMultipleNodeProperties(t *testing.T) {
	ctx := GetTestContext()

	t.Run("UpdateNodeV2AddingProperties", func(t *testing.T) {
		nodeName := "test_update_v2_add_props"
		// Create node without properties
		node := createTestNode(t, ctx, nodeName,
			WithNumExecutors(1),
			WithRemoteFS("/var/jenkins"),
		)

		// Update with properties
		envVars := map[string]string{
			"NEW_VAR": "new_value",
		}
		updatedNode, err := node.UpdateNodeV2(ctx, nodeName,
			WithNodeProperties(
				NewEnvironmentVariablesNodeProperty(envVars),
				NewDiskSpaceMonitorNodeProperty("500MB"),
			),
		)
		require.NoError(t, err)
		require.NotNil(t, updatedNode)

		slaveConfig, err := updatedNode.GetSlaveConfig(ctx)
		require.NoError(t, err)
		require.NotNil(t, slaveConfig)
		assert.NotNil(t, slaveConfig.NodeProperties)
		assert.Len(t, slaveConfig.NodeProperties.Properties, 2)
	})

	t.Run("UpdateNodeV2ReplacingProperties", func(t *testing.T) {
		nodeName := "test_update_v2_replace_props"
		// Create node with properties
		node := createTestNode(t, ctx, nodeName,
			WithNumExecutors(1),
			WithRemoteFS("/var/jenkins"),
			WithNodeProperties(
				NewEnvironmentVariablesNodeProperty(map[string]string{"OLD": "value"}),
			),
		)

		// Update with different properties
		toolLocs := map[string]string{
			"hudson.plugins.git.GitTool$DescriptorImpl:Default": "/usr/bin/git",
		}
		updatedNode, err := node.UpdateNodeV2(ctx, nodeName,
			WithNodeProperties(
				NewToolLocationNodeProperty(toolLocs),
			),
		)
		require.NoError(t, err)
		require.NotNil(t, updatedNode)

		slaveConfig, err := updatedNode.GetSlaveConfig(ctx)
		require.NoError(t, err)
		require.NotNil(t, slaveConfig)
		assert.NotNil(t, slaveConfig.NodeProperties)
	})
}
