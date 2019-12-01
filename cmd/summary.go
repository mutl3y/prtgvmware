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
	"net/url"
	"time"

	"github.com/mutl3y/PRTG_VMware/VMware"
	"github.com/spf13/cobra"
)

// summaryCmd represents the summary command
var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "vm summary for a single machine",
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

		snapAge, err := flags.GetDuration("snapAge")
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

		name, err := flags.GetString("Name")
		if err != nil {
			VMware.SensorWarn(err, true)
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

		extraSensors, err := flags.GetStringSlice("vmmetrics")

		err = c.VmSummary(name, moid, &lim, snapAge, js, extraSensors)
		if err != nil {
			VMware.SensorWarn(fmt.Errorf("get summary error: %v", err), true)

		}

	},
}

func init() {
	rootCmd.AddCommand(summaryCmd)
	summaryCmd.Flags().BoolP("json", "j", false, "pretty print json version of vmware data")
	summaryCmd.Flags().DurationP("snapAge", "P", (7*24)*time.Hour, "ignore snapshots younger than")
	summaryCmd.Flags().StringSlice("vmmetrics", []string{}, "include additonal vm metrics, I.E. cpu.ready.summation")

}
