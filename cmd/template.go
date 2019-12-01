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
	"github.com/spf13/cobra"
	"log"
)

// templateCmd represents the template command
var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "generate prtg template for autodiscovery",
	Long: `
use this to support autodiscovery using VMware tags

`,
	Run: func(cmd *cobra.Command, args []string) {

		flags := cmd.Flags()
		tags, err := flags.GetStringSlice("Tags")
		if err != nil {
			log.Fatal(err)
		}
		snapAge, err := flags.GetDuration("snapAge")
		if err != nil {
			log.Fatal(err)
		}
		err = VMware.GenTemplate(tags, snapAge)
		fmt.Println(err)
	},
}

func init() {
	rootCmd.AddCommand(templateCmd)
	err := rootCmd.MarkPersistentFlagRequired("Tags")
	if err != nil {
		log.Fatalf("%v", err)
	}

}
