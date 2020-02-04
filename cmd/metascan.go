/*Copyright © 2019.  mutl3y
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
	"github.com/mutl3y/prtgvmware/app"
	"github.com/spf13/cobra"
)

// metascanCmd represents the metascan command
var metascanCmd = &cobra.Command{
	Use:   "metascan",
	Short: "returns prtg sensors for autodiscovery",
	Long:  `used for autodiscovery of vmware sensors`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := rootCmd.MarkPersistentFlagRequired("tags")
		return err
	},
	Run: func(cmd *cobra.Command, args []string) {

		flags := cmd.Flags()
		c, err := login(flags)
		if err != nil {
			app.SensorWarn(err, true)
			return
		}
		tags, err := flags.GetStringSlice("tags")
		if err != nil {
			app.SensorWarn(err, true)
			return
		}

		snapAge, err := flags.GetDuration("snapAge")
		if err != nil {
			app.SensorWarn(err, true)
			return
		}
		tagMap := app.NewTagMap()

		err = c.Metascan(tags, tagMap, snapAge)
		if err != nil {
			app.SensorWarn(err, true)
			return
		}

		_ = c.Logout()

	},
}

func init() {
	rootCmd.AddCommand(metascanCmd)
}
