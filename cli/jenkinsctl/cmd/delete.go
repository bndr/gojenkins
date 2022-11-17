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

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a resource from Jenkins",
}

var deleteNode = &cobra.Command{
	Use:   "node",
	Short: "delete a node",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("⏳ Deleting the node %s...\n", args[0])
		_, err := jenkinsMod.Instance.DeleteNode(jenkinsMod.Context, args[0])
		if err != nil {
			fmt.Printf("unable to find the node: %s - err: %s \n", args[0], err)
			os.Exit(1)
		}
		fmt.Printf("Deleted node: %s\n", args[0])
	},
}

var deleteJob = &cobra.Command{
	Use:   "job",
	Short: "delete a job",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("❌ requires at least one argument")
		}

		fmt.Printf("⏳ Deleting the job %s...\n", args[0])
		err := jenkinsMod.DeleteJob(args[0])
		if err != nil {
			fmt.Printf("unable to find the job: %s - err: %s \n", args[0], err)
			os.Exit(1)
		}
		fmt.Printf("Deleted job: %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.AddCommand(deleteJob)
	deleteCmd.AddCommand(deleteNode)
}
