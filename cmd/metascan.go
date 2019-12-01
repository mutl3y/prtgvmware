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
	"github.com/spf13/cobra"
	"github.com/vmware/govmomi/property"
	"log"
	"net/url"
)

// metascanCmd represents the metascan command
var metascanCmd = &cobra.Command{
	Use:   "metascan",
	Short: "returns sensors for vmware vcenter instances by name or tag",
	Long:  `used for autodiscovery of vmware sensors`,
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

		c, err := VMware.NewClient(u, user, pww)
		if err != nil {
			log.Fatal(err)
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

		scanTypes, err := flags.GetStringSlice("Type")
		if err != nil {
			log.Fatal(err)
		}
		if name != "" {
			f["name"] = "*"
		}
		snapAge, err := flags.GetDuration("snapAge")
		if err != nil {
			log.Fatal(err)
		}
		tagMap := VMware.NewTagMap()

		err = c.Metascan(tags, tagMap, scanTypes, snapAge)
		if err != nil {
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(metascanCmd)
	metascanCmd.Flags().StringSliceP("Type", "T", []string{"vm"}, "what to scan for")
}
