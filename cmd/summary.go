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
	"github.com/vmware/govmomi/property"
	"log"
	"net/url"
	"time"

	"github.com/mutl3y/PRTG_VMware/VMware"
	"github.com/spf13/cobra"
)

// summaryCmd represents the summary command
var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "vm summary for a single machine",
	Long: `Property filter examples
	self=vm-25
	name=*vm42
	[name=*vm42,tag=prod]
`,
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

		snapAge, err := flags.GetDuration("snapAge")
		if err != nil {
			fmt.Println(err)
		}

		js, err := flags.GetBool("json")
		if err != nil {
			fmt.Println(err)
		}

		filter, err := flags.GetStringToString("propertyFilter")

		c, err := VMware.NewClient(u, user, pww)
		if err != nil {
			log.Fatal(err)
		}

		f := property.Filter{}
		for k, v := range filter {
			f[k] = v
		}
		lim, err := limitStruct(flags)
		if err != nil {
			log.Fatal(err)
		}
		err = c.VmSummary(f, &lim, snapAge, js)
		if err != nil {
			fmt.Println(err)
		}

	},
}

func init() {
	rootCmd.AddCommand(summaryCmd)
	summaryCmd.Flags().BoolP("json", "j", false, "pretty print json version of vmware data")
	//summaryCmd.Flags().StringP("ptype", "t", "self", "managed object property type. eg self or name")
	//summaryCmd.Flags().StringP("psearch", "U", "prtgUtil", "managed object property")
	summaryCmd.Flags().DurationP("snapAge", "P", (7*24)*time.Hour, "ignore snapshots younger than")
	summaryCmd.Flags().StringToStringP("propertyFilter", "F", map[string]string{"name": "*1"}, "vmware property filter")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// summaryCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// summaryCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
