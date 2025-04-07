package gojenkins

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	JenkinsPort     = "8080"
	jenkinsUsername = "admin"
	jenkinsPassword = "admin"
)

type ContainerizedTest struct {
	Jenkins     *Jenkins
	CleanupFunc func() error
}

func GetTestContext() context.Context {
	ctx := context.Background()
	timeoutCtx, _ := context.WithTimeout(ctx, time.Minute)
	return timeoutCtx
}

func Setup(ctx context.Context, username string, password string) (*ContainerizedTest, error) {
	timeoutContext, _ := context.WithTimeout(ctx, time.Second*30)
	jenkinsContainer, err := CreateJenkinsContainer(timeoutContext, username, password)
	cleanup := func() error {
		return testcontainers.TerminateContainer(jenkinsContainer)
	}

	if err != nil {
		cleanup()
		return nil, err
	}

	jenkinsEndpoint, err := jenkinsContainer.Endpoint(timeoutContext, "")
	if err != nil {
		cleanup()
		return nil, err
	}

	j := CreateJenkins(nil, fmt.Sprintf("http://%s", jenkinsEndpoint), username, password)

	j, err = j.Init(timeoutContext)
	if err != nil {
		cleanup()
		return nil, err
	}
	return &ContainerizedTest{
		Jenkins:     j,
		CleanupFunc: cleanup,
	}, nil
}

func CreateJenkinsContainer(ctx context.Context, username string, password string) (testcontainers.Container, error) {

	dockerFile := testcontainers.FromDockerfile{
		Context:    ".",
		Dockerfile: "Dockerfile",
		KeepImage:  true,
		Repo:       "test-image",
		Tag:        "jenkins-test",
	}

	jenkinsOpts := []string{
		fmt.Sprintf("--argumentsRealm.roles.user=%s", username),
		fmt.Sprintf(" --argumentsRealm.passwd.admin=%s", password),
		"--argumentsRealm.roles.admin=admin",
	}
	req := testcontainers.ContainerRequest{
		FromDockerfile: dockerFile,
		ExposedPorts:   []string{fmt.Sprintf("%s/tcp", JenkinsPort)},
		Env: map[string]string{
			"JAVA_OPTS":    "-Djenkins.install.runSetupWizard=false",
			"JENKINS_OPTS": strings.Join(jenkinsOpts, " "),
		},
		WaitingFor: wait.ForLog("Jenkins is fully up and running"),
	}
	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}
