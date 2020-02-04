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
	"fmt"
	"github.com/mutl3y/prtgvmware/app"
	"github.com/spf13/cobra"
)

// summaryCmd represents the summary command
var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "vm summary for a single machine",
	Long: `queries vm performance metrics and outputs in PRTG format

Can be further expanded by adding additional VmWare performance counters

counters included by default are 
"cpu.readiness.average", "cpu.usage.average",
"datastore.datastoreNormalReadLatency.latest", "datastore.datastoreNormalWriteLatency.latest",
"datastore.datastoreReadIops.latest", "datastore.datastoreWriteIops.latest",
"disk.read.average", "disk.write.average", "disk.usage.average",
"mem.active.average", "mem.consumed.average", "mem.usage.average",
"net.bytesRx.average", "net.bytesTx.average", "net.usage.average",
`,
	Run: func(cmd *cobra.Command, args []string) {
		flags := cmd.Flags()
		c, err := login(flags)
		if err != nil {
			app.SensorWarn(err, true)
			return
		}
		name, err := flags.GetString("name")
		if err != nil {
			app.SensorWarn(err, true)
			return
		}
		oid, err := flags.GetString("oid")
		if err != nil {
			app.SensorWarn(err, true)
			return
		}

		if name == "" && oid == "" {
			app.SensorWarn(fmt.Errorf("you need to provide a name or managed object id"), true)
			return
		}

		snapAge, err := flags.GetDuration("snapAge")
		if err != nil {
			app.SensorWarn(err, true)
			return
		}

		lim, err := limitStruct(flags)
		if err != nil {
			app.SensorWarn(err, true)
			return
		}

		extraSensors, err := flags.GetStringSlice("vmMetrics")
		if err != nil {
			app.SensorWarn(err, true)
			return
		}
		js, err := flags.GetBool("json")
		if err != nil {
			app.SensorWarn(err, true)
			return
		}

		err = c.VMSummary(name, oid, &lim, snapAge, js, extraSensors)
		if err != nil {
			app.SensorWarn(err, true)

		}
		//if !c.Cached {
		//	c.Logout()
		//}
	},
}

func init() {
	rootCmd.AddCommand(summaryCmd)
	summaryCmd.Flags().StringSlice("vmMetrics", []string{}, "include additional vm metrics, I.E. cpu.ready.summation")
}
