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
	"fmt"
	"github.com/dougsland/jenkinsctl/jenkins"
	"github.com/spf13/cobra"
	"os"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "jenkinsctl",
	Short:   "A client for jenkins",
	Version: "v0.0.1",
	Long:    `Client for jenkins, manage resources by the jenkins`,
}

var jenkinsMod jenkins.Jenkins
var jenkinsConfig jenkins.Config
var configFile string

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "", "", "Path to config file")
}

func initConfig() {
	dirname, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if configFile != "" {
		jenkinsConfig.SetConfigPath(configFile)
	} else {
		jenkinsConfig.SetConfigPath(dirname + "/.config/jenkinsctl/config.json")
	}

	config, err := jenkinsConfig.LoadConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	jenkinsMod = jenkins.Jenkins{}
	err = jenkinsMod.Init(config)
	if err != nil {
		fmt.Println("❌ jenkins server unreachable: " + jenkinsMod.Server)
		os.Exit(1)
	}

}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
