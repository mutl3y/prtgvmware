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
	"github.com/mutl3y/prtgvmware/app"
	"github.com/spf13/cobra"
)

// dynamicTemplatesCmd represents the dynamicTemplates command
var dynamicTemplatesCmd = &cobra.Command{
	Use:   "dynamicTemplates",
	Short: "generate prtg template for autodiscovery",
	Long: `use this to support autodiscovery using VMware tags

run this regually via cron or task scheduler and copy template to devicetemplates folder for use by autodiscovery
`,
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
