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
	"github.com/mutl3y/PRTG_VMware/app"

	"github.com/spf13/cobra"
)

// hssummaryCmd represents the hssummary command
var hssummaryCmd = &cobra.Command{
	Use:   "hssummary",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
		err = c.HostSummary(name, oid, &lim, js)
		if err != nil {
			app.SensorWarn(fmt.Errorf("get summary error: %v", err), true)

		}
		if !c.Cached {
			c.Logout()
		}
	},
}

func init() {
	rootCmd.AddCommand(hssummaryCmd)
}
