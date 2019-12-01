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
	"net/url"

	"github.com/spf13/cobra"
)

// dssummaryCmd represents the dssummary command
var dssummaryCmd = &cobra.Command{
	Use:   "dssummary",
	Short: "ds summary for a single datastore",
	Long: `
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

		js, err := flags.GetBool("json")
		if err != nil {
			fmt.Println(err)
		}

		c, err := VMware.NewClient(u, user, pww)
		if err != nil {
			VMware.SensorWarn(fmt.Errorf("API connection error: %v", err), true)
			return
		}

		moid, err := flags.GetString("Moid")
		if err != nil {
			VMware.SensorWarn(err, true)
			return
		}

		lim, err := limitStruct(flags)
		if err != nil {

			VMware.SensorWarn(err, true)
			return
		}
		name, err := flags.GetString("Name")
		if err != nil {
			VMware.SensorWarn(err, true)
		}

		err = c.DsSummary(name, moid, &lim, js)
		if err != nil {
			VMware.SensorWarn(fmt.Errorf("get summary error: %v", err), true)

		}

	},
}

func init() {
	rootCmd.AddCommand(dssummaryCmd)
	dssummaryCmd.Flags().BoolP("json", "j", false, "pretty print json version of vmware data")

}
