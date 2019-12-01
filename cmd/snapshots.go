/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

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
	"github.com/mutl3y/PRTG_VMware/VMware"
	"github.com/vmware/govmomi/property"
	"log"
	"net/url"
	"time"

	"github.com/spf13/cobra"
)

// snapshotsCmd represents the snapshots command
var snapshotsCmd = &cobra.Command{
	Use:   "snapshots",
	Short: "snapshots for many vm's",
	Long: `Property filter examples
	self=vm-25
	name=*vm42
	tag=prod`,
	Run: func(cmd *cobra.Command, args []string) {
		flags := cmd.Flags()
		u := &url.URL{}
		urls, err := flags.GetString("url")
		if err != nil {
			fmt.Println(err)
		}
		user, err := flags.GetString("username")
		if err != nil {
			fmt.Println(err)
		}
		pww, err := flags.GetString("password")
		if err != nil {
			fmt.Println(err)
		}

		u, _ = u.Parse(urls)

		Age, err := flags.GetDuration("snapAge")
		if err != nil {
			fmt.Println(err)
		}

		js, err := flags.GetBool("json")
		if err != nil {
			fmt.Println(err)
		}

		c, err := VMware.NewClient(u, user, pww)
		if err != nil {
			VMware.SensorWarn(fmt.Errorf("API connection error: %v", err), true)
			return
		}
		f := property.Filter{}
		name, err := flags.GetString("Name")
		if err != nil {
			log.Fatal(err)
		}
		tags, err := flags.GetStringSlice("Tags")
		if err != nil {
			log.Fatal(err)
		}

		if name != "" {
			f["name"] = name
		} else {
			f["name"] = "*"
		}

		lim, err := limitStruct(flags)
		if err != nil {
			log.Fatal(err)
		}

		err = c.SnapShotsOlderThan(f, tags, &lim, Age, js)
		if err != nil {
			VMware.SensorWarn(fmt.Errorf("get snapshots error: %v", err), true)

		}

	},
}

func init() {
	rootCmd.AddCommand(snapshotsCmd)
	snapshotsCmd.Flags().BoolP("json", "j", false, "pretty print json version of vmware data")
	//summaryCmd.Flags().StringP("ptype", "t", "self", "managed object property type. eg self or name")
	//summaryCmd.Flags().StringP("psearch", "U", "prtgUtil", "managed object property")
	snapshotsCmd.Flags().DurationP("snapAge", "A", (7*24)*time.Hour, "ignore snapshots younger than")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// snapshotsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// snapshotsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
