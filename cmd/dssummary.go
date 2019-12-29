/*
 * Copyright © 2019.  mutl3y
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package cmd

import (
	"fmt"
	"github.com/mutl3y/prtgvmware/app"
	"github.com/spf13/cobra"
)

// dssummaryCmd represents the dssummary command
var dssummaryCmd = &cobra.Command{
	Use:   "dsSummary",
	Short: "summary for a single datastore",
	Long: `
queries datastore summary metrics and outputs in PRTG format
`,
	Run: func(cmd *cobra.Command, args []string) {
		flags := cmd.Flags()
		c, err := login(flags)
		if err != nil {
			app.SensorWarn(err, true)
			return
		}
		oid, err := flags.GetString("oid")
		if err != nil {
			app.SensorWarn(err, true)
			return
		}

		lim, err := limitStruct(flags)
		if err != nil {
			app.SensorWarn(err, true)
			return
		}
		if lim.MinWarn < "20" {
			lim.MinWarn = "20"
		}
		if lim.MinErr < "10" {
			lim.MinErr = "20"
		}

		name, err := flags.GetString("name")
		if err != nil {
			app.SensorWarn(err, true)
		}
		if name == "" && oid == "" {
			app.SensorWarn(fmt.Errorf("you need to provide a name or managed object id"), true)
			return
		}
		js, err := flags.GetBool("json")
		if err != nil {
			app.SensorWarn(err, true)
			return
		}
		err = c.DsSummary(name, oid, &lim, js)
		if err != nil {
			app.SensorWarn(err, true)

		}
		if !c.Cached {
			_ = c.Logout()
		}
	},
}

func init() {
	rootCmd.AddCommand(dssummaryCmd)
	dssummaryCmd.Flags().BoolP("json", "j", false, "pretty print json version of vmware data")

}
