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
	"github.com/spf13/cobra"
	"os"
)

var enableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable a resource in Jenkins",
}

// enableCmd represents the enable command
var enableJobCmd = &cobra.Command{
	Use:   "job",
	Short: "Enable job",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("❌ requires at least one argument")
		}
		fmt.Printf("⏳ Enabling job %s...\n", args[0])

		job, err := jenkinsMod.Instance.GetJob(jenkinsMod.Context, args[0])
		if err != nil {
			fmt.Printf("unable to find the job: %s - err: %s \n", args[0], err)
			os.Exit(1)
		}
		job.Enable(jenkinsMod.Context)
		fmt.Printf("job %s enabled..\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(enableCmd)
	enableCmd.AddCommand(enableJobCmd)
}
