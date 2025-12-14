module github.com/dougsland/gojenkins/cli/jenkinsctl

go 1.16

require (
	github.com/dougsland/gojenkins/cli/jenkinsctl/jenkins v0.0.0
	github.com/spf13/cobra v1.1.3
)

replace github.com/dougsland/gojenkins/cli/jenkinsctl/jenkins => ./jenkins

replace github.com/bndr/gojenkins => ../..
