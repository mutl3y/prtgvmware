/*
Copyright © 2019 mutl3y

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
	"github.com/spf13/pflag"
	"os"
)

var cfgFile = "PRTG_VMware.yml"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "PRTG_VMware",
	Short: "VMware sensors for prtg",
	Long: `advanced sensors for VMware

this app exposes all the common stats for vm's, Hypervisors, Vcenter & Datastores'

if you find I have missed one that's available via an api drop me a mail or register a request ion github
`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	//save, err := rootCmd.Flags().GetBool("saveconfig")
	//if err != nil {
	//	log.Fatalf("failed saving config %v", err)
	//}
	//if save {
	//	fmt.Println("writing config file")
	//	if err := viper.WriteConfigAs(cfgFile); err != nil {
	//		fmt.Println(err)
	//	}
	//}
}

func init() {
	//cobra.OnInitialize(initConfig)
	//rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "PRTG_VMware.yaml", "config file ")
	//rootCmd.PersistentFlags().BoolP("saveconfig", "s", false, "save parameters to file")

	rootCmd.PersistentFlags().StringP("username", "u", "", "vcenter username")
	rootCmd.PersistentFlags().StringP("password", "p", "", "vcenter password")
	rootCmd.PersistentFlags().StringP("url", "U", "", "url for vcenter api")
	rootCmd.PersistentFlags().String("WarnMsg", "", "")
	rootCmd.PersistentFlags().String("ErrMsg", "", "")
	rootCmd.PersistentFlags().Float64("MinWarn", 0, "")
	rootCmd.PersistentFlags().Float64("MaxWarn", 0, "")
	rootCmd.PersistentFlags().Float64("MinErr", 0, "")
	rootCmd.PersistentFlags().Float64("MaxErr", 0, "")
}

// initConfig reads in config file and ENV variables if set.
//func initConfig() {
//	if cfgFile != "" {
//		// Use config file from the flag.
//		viper.SetConfigFile(cfgFile)
//	} else {
//		// Find home directory.
//		home, err := homedir.Dir()
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		// Search config in home directory with name "PRTG_VMware" (without extension).
//		viper.AddConfigPath(home)
//		viper.SetConfigType("yml")
//		viper.SetConfigName("PRTG_VMware")
//	}
//
//	viper.AutomaticEnv() // read in environment variables that match
//
//	// If a config file is found, read it in.
//	if err := viper.ReadInConfig(); err == nil {
//		fmt.Println("Using config file:", viper.ConfigFileUsed())
//	}
//
//}

var (
	WarnMsg, ErrMsg                  string
	MinWarn, MaxWarn, MinErr, MaxErr float64
)

func limitStruct(flags *pflag.FlagSet) (lim VMware.LimitsStruct, err error) {

	WarnMsg, err = flags.GetString("WarnMsg")
	if err != nil {
		return
	}
	ErrMsg, err = flags.GetString("ErrMsg")
	if err != nil {
		return
	}
	MinWarn, err = flags.GetFloat64("MinWarn")
	if err != nil {
		return
	}
	MaxWarn, err = flags.GetFloat64("MaxWarn")
	if err != nil {
		return
	}
	MinErr, err = flags.GetFloat64("MinErr")
	if err != nil {
		return
	}
	MaxErr, err = flags.GetFloat64("MaxErr")
	if err != nil {
		return
	}
	lim = VMware.LimitsStruct{
		MinWarn: MinWarn,
		MaxWarn: MaxWarn,
		WarnMsg: WarnMsg,
		MinErr:  MinErr,
		MaxErr:  MaxErr,
		ErrMsg:  ErrMsg,
	}
	return lim, nil
}
