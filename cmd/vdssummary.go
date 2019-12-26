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

var vdsSummaryCmd = &cobra.Command{
	Use:   "vdsSummary",
	Short: "vds summary for prtg",
	Long:  `Provides basic vds status for PRTG monitoring`,
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

		name, err := flags.GetString("name")
		if err != nil {
			app.SensorWarn(err, true)
		}
		js, err := flags.GetBool("json")
		if err != nil {
			app.SensorWarn(err, true)
			return
		}
		err = c.VdsSummary(name, oid, js)
		if err != nil {
			app.SensorWarn(fmt.Errorf("get summary error: %v", err), true)

		}
		if !c.Cached {
			_ = c.Logout()
		}
	},
}

func init() {
	rootCmd.AddCommand(vdsSummaryCmd)
}
