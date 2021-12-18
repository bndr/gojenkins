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

// pluginsCmd represents the plugins command
var pluginsCmd = &cobra.Command{
	Use:   "plugins",
	Short: "Commands related to plugins",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("❌ requires at least one argument")
		}
		return nil
	},
}

var hasPlugin = &cobra.Command{
	Use:   "hasplugin",
	Short: "check if plugin is installed on the server",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("❌ requires at least one argument [PLUGIN_NAME]")
		}

		ret, err := jenkinsMod.Instance.HasPlugin(jenkinsMod.Context, args[0])
		if err != nil {
			fmt.Printf("error cannot install the plugin: %s - %s\n", args[0], err)
			os.Exit(1)
		}

		if ret == nil {
			fmt.Printf("Plugin %s NOT installed\n", args[0])
		} else {
			fmt.Printf("Plugin %s installed\n", args[0])
		}
		return nil
	},
}

var installPlugin = &cobra.Command{
	Use:   "install",
	Short: "install a plugin",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return errors.New("❌ requires at least two arguments [PLUGIN_NAME] [VERSION]")
		}

		err := jenkinsMod.Instance.InstallPlugin(jenkinsMod.Context, args[0], args[1])
		if err != nil {
			fmt.Printf("error cannot install the plugin: %s - %s\n", args[0], err)
			os.Exit(1)
		}
		fmt.Printf("Plugin %s installed\n", args[0])
		return nil
	},
}

// Plugins Command
var getInfo = &cobra.Command{
	Use:   "listall",
	Short: "get all plugins active and enabled",
	RunE: func(cmd *cobra.Command, args []string) error {
		jenkinsMod.PluginsShow()
		return nil
	},
}

var unInstallPlugin = &cobra.Command{
	Use:   "uninstall",
	Short: "uninstall a plugin",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("❌ requires at least one argument [PLUGIN_NAME]")
		}

		_, err := jenkinsMod.Instance.HasPlugin(jenkinsMod.Context, args[0])
		if err != nil {
			fmt.Printf("cannot find plugin %s in the server - %s\n", args[0], err)
			os.Exit(1)
		}

		err = jenkinsMod.Instance.UninstallPlugin(jenkinsMod.Context, args[0])
		if err != nil {
			fmt.Printf("error cannot uninstall the plugin: %s - %s\n", args[0], err)
			os.Exit(1)
		}
		fmt.Printf("Plugin %s uninstalled\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pluginsCmd)
	pluginsCmd.AddCommand(unInstallPlugin)
	pluginsCmd.AddCommand(installPlugin)
	pluginsCmd.AddCommand(hasPlugin)
	pluginsCmd.AddCommand(getInfo)
}
