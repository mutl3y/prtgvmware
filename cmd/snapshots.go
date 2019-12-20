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
	"github.com/vmware/govmomi/property"
)

// snapshotsCmd represents the snapshots command
var snapshotsCmd = &cobra.Command{
	Use:   "snapshots",
	Short: "snapshots for many vm's",
	Long: `queries a count of snapshot's that are older than specified
snapAge is 7 days by default`,
	Run: func(cmd *cobra.Command, args []string) {
		flags := cmd.Flags()
		c, err := login(flags)
		if err != nil {
			app.SensorWarn(err, true)
			return
		}
		f := property.Filter{}
		name, err := flags.GetString("name")
		if err != nil {
			app.SensorWarn(err, true)
			return
		}
		tags, err := flags.GetStringSlice("tags")
		if err != nil {
			app.SensorWarn(err, true)
			return
		}
		age, err := flags.GetDuration("snapAge")
		if err != nil {
			app.SensorWarn(err, true)
			return
		}

		js, err := flags.GetBool("json")
		if err != nil {
			app.SensorWarn(err, true)
			return
		}
		if name != "" {
			f["name"] = name
		} else {
			f["name"] = "*"
		}

		lim, err := limitStruct(flags)
		if err != nil {
			app.SensorWarn(err, true)
			return
		}

		err = c.SnapShotsOlderThan(f, tags, &lim, age, js)
		if err != nil {
			app.SensorWarn(fmt.Errorf("get snapshots error: %v", err), true)
			return
		}
		if !c.Cached {
			c.Logout()
		}
	},
}

func init() {
	rootCmd.AddCommand(snapshotsCmd)
}
