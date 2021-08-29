/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// getCmd represents the show command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a resource from Jenkins",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("❌ requires at least one argument")
		}
		return nil
	},
}

// Connection Command
var connectionInfo = &cobra.Command{
	Use:   "connection",
	Short: "get connection info",
	Run: func(cmd *cobra.Command, args []string) {
		err := jenkinsMod.ServerInfo()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

// Plugins Command
var pluginsInfo = &cobra.Command{
	Use:   "plugins",
	Short: "get all plugins active and enabled",
	RunE: func(cmd *cobra.Command, args []string) error {
		jenkinsMod.PluginsShow()
		return nil
	},
}

var viewsInfo = &cobra.Command{
	Use:   "views",
	Short: "get all views",
	Run: func(cmd *cobra.Command, args []string) {
		err := jenkinsMod.ShowViews()
		if err != nil {
			fmt.Println("❌ cannot get all views")
			os.Exit(1)
		}
	},
}

// Build Command
var build = &cobra.Command{
	Use:   "build",
	Short: "build related commands",
}

var buildQueue = &cobra.Command{
	Use:   "queue",
	Short: "get build queue",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("⏳ Collecting build queue information...\n")
		err := jenkinsMod.ShowBuildQueue()
		if err != nil {
			fmt.Println("❌ cannot collect build queue")
			os.Exit(1)
		}
	},
}

// Job Commands
var job = &cobra.Command{
	Use:   "job",
	Short: "job related commands",
}

var jobLastUnstableBuild = &cobra.Command{
	Use:   "lastunstablebuild",
	Short: "get last unstable build from a job",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("❌ requires at least one argument [JOB NAME]")
			os.Exit(1)
		}
		err := jenkinsMod.GetLastUnstableBuild(args[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var jobLastStableBuild = &cobra.Command{
	Use:   "laststablebuild",
	Short: "get last stable build from a job",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("❌ requires at least one argument [JOB NAME]")
			os.Exit(1)
		}
		err := jenkinsMod.GetLastStableBuild(args[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var jobLastFailedBuild = &cobra.Command{
	Use:   "lastfailedbuild",
	Short: "get last failed build from a job",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("❌ requires at least one argument [JOB NAME]")
			os.Exit(1)
		}
		err := jenkinsMod.GetLastFailedBuild(args[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var jobLastCompletedBuild = &cobra.Command{
	Use:   "lastcompletedbuild",
	Short: "get last successful build from a job",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("❌ requires at least one argument [JOB NAME]")
			os.Exit(1)
		}
		err := jenkinsMod.GetLastCompletedBuild(args[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var jobGetLastSuccessfulBuild = &cobra.Command{
	Use:   "lastsuccessfulbuild",
	Short: "get last successful build from a job",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("❌ requires at least one argument [JOB NAME]")
			os.Exit(1)
		}
		err := jenkinsMod.GetLastSuccessfulBuild(args[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var jobGetLastBuild = &cobra.Command{
	Use:   "lastbuild",
	Short: "get last build from a job",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("❌ requires at least one argument [JOB NAME]")
			os.Exit(1)
		}
		err := jenkinsMod.GetLastBuild(args[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var jobAll = &cobra.Command{
	Use:   "all",
	Short: "get all jobs",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("⏳ Collecting all job(s) information...\n")
		err := jenkinsMod.ShowAllJobs()
		if err != nil {
			fmt.Printf("❌ unable to find any job. err: %s \n", err)
			os.Exit(1)
		}
	},
}

var jobConfig = &cobra.Command{
	Use:   "config",
	Short: "get job config",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("❌ requires at least one argument [JOB NAME]")
			os.Exit(1)
		}
		err := jenkinsMod.JobGetConfig(args[0])
		if err != nil {
			fmt.Printf("❌ unable to find the job: %s - err: %s \n", args[0], err)
			os.Exit(1)
		}
	},
}

// Node Commands
var nodes = &cobra.Command{
	Use:   "nodes",
	Short: "nodes related commands",
}

var nodesOffline = &cobra.Command{
	Use:   "offline",
	Short: "get nodes offline",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("⏳ Collecting node(s) information...\n")
		hosts, err := jenkinsMod.ShowNodes("offline")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// We must exit as failure in case we have nodes offline
		if len(hosts) > 0 {
			os.Exit(1)
		}
	},
}

var nodesOnline = &cobra.Command{
	Use:   "online",
	Short: "get nodes online",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("⏳ Collecting node(s) information...\n")
		_, err := jenkinsMod.ShowNodes("online")
		if err != nil {
			fmt.Printf("❌ unable to find nodes - err: %s \n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	// get
	getCmd.AddCommand(connectionInfo)
	getCmd.AddCommand(pluginsInfo)
	getCmd.AddCommand(viewsInfo)
	getCmd.AddCommand(nodes)
	getCmd.AddCommand(build)
	getCmd.AddCommand(job)

	// nodes
	nodes.AddCommand(nodesOffline)
	nodes.AddCommand(nodesOnline)

	// build
	build.AddCommand(buildQueue)

	// job
	job.AddCommand(jobConfig)
	job.AddCommand(jobAll)
	job.AddCommand(jobGetLastBuild)
	job.AddCommand(jobGetLastSuccessfulBuild)
	job.AddCommand(jobLastCompletedBuild)
	job.AddCommand(jobLastFailedBuild)
	job.AddCommand(jobLastStableBuild)
	job.AddCommand(jobLastUnstableBuild)
}
