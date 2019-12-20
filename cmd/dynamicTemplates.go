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
	"github.com/mutl3y/PRTG_VMware/app"
	"github.com/spf13/cobra"
)

// dynamicTemplatesCmd represents the dynamicTemplates command
var dynamicTemplatesCmd = &cobra.Command{
	Use:   "dynamicTemplates",
	Short: "A brief description of your command",
	Long: `use this to support autodiscovery using VMware tags

this is to get around the limitations of the metadata scan option not adding new items after first 
discovery as it will not trigger the PRTG sensor tracking code

this should be run regularly and the file that is created should be copied to the 
device templates folder so its picked up by autodiscovery on next invocation

if you prefer the metascan option you will need to delete all your sensors 
from the device for the metascan to find anything new`,
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
		}
		snapAge, err := flags.GetDuration("snapAge")
		if err != nil {
			app.SensorWarn(err, true)
		}
		tplate, err := flags.GetString("template")
		if err != nil {
			app.SensorWarn(err, true)
		}

		err = c.DynTemplate(tags, snapAge, tplate)
		if err != nil {
			app.SensorWarn(err, true)
		}
	},
}

func init() {

	rootCmd.AddCommand(dynamicTemplatesCmd)
	dynamicTemplatesCmd.Flags().StringP("template", "f", "prtgvmware", "filename to save template as, adds .odt, only needed if using multiple sensors")

}
