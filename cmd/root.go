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
	"github.com/spf13/pflag"
	"log"
	"net/url"
	"time"
)

// rootCmd represents the base command when called without any sub-commands
var rootCmd = &cobra.Command{
	Use:     "prtgvmware",
	Short:   "VMware sensors for prtg",
	Version: "v1.0.7", //todo make sure you update this
	Long: `advanced sensors for VMware

this app exposes all the common stats for vm's, Hypervisor's, VDS & Datastore's

to use autodiscovery you need to generate template using tags for each set of objects you want to monitor
`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)

	}

}

func init() {
	//cobra.OnInitialize(initConfig)
	//rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "prtgvmware.yaml", "config file ")
	//rootCmd.PersistentFlags().BoolP("saveconfig", "s", false, "save parameters to file")

	rootCmd.PersistentFlags().StringP("username", "u", "", "vcenter username")
	rootCmd.PersistentFlags().StringP("password", "p", "", "vcenter password")
	rootCmd.PersistentFlags().StringP("url", "U", "", "url for vcenter api")
	rootCmd.PersistentFlags().String("msgWarn", "", "message to use if warning value exceeded (used with snapshots)")
	rootCmd.PersistentFlags().String("msgError", "", "message to use if error value exceeded (used with snapshots)")
	//rootCmd.PersistentFlags().Float64("MinWarn", 0, "")
	//rootCmd.PersistentFlags().Float64("MinErr", 0, "")
	rootCmd.PersistentFlags().Float64("maxWarn", 1, "greater than equal this will trigger a warning response (used with snapshots)")
	rootCmd.PersistentFlags().Float64("maxErr", 0, "greater than equal this will trigger a error response (used with snapshots)")

	rootCmd.PersistentFlags().StringP("name", "n", "", "name of vm, supports *partofname*")
	rootCmd.PersistentFlags().StringP("oid", "i", "", "exact id of an object e.g. vm-12, vds-81, host-9, datastore-10 ")
	rootCmd.PersistentFlags().StringSliceP("tags", "t", []string{}, "slice of tags to include")
	rootCmd.PersistentFlags().DurationP("snapAge", "a", (7*24)*time.Hour, "ignore snapshots younger than")
	rootCmd.PersistentFlags().BoolP("json", "j", false, "pretty print json version of vmware data")
	rootCmd.PersistentFlags().BoolP("cachedCreds", "c", false, "disable cached connection")

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
//		// Search config in home directory with name "prtgvmware" (without extension).
//		viper.AddConfigPath(home)
//		viper.SetConfigType("yml")
//		viper.SetConfigName("prtgvmware")
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
	warnMsg string
	errMsg  string
	MinWarn float64
	maxWarn float64
	MinErr  float64
	maxErr  float64
)

func limitStruct(flags *pflag.FlagSet) (lim app.LimitsStruct, err error) {

	warnMsg, err = flags.GetString("msgWarn")
	if err != nil {
		return
	}
	errMsg, err = flags.GetString("msgError")
	if err != nil {
		return
	}
	//MinWarn, err = flags.GetFloat64("MinWarn")
	//if err != nil {
	//	return
	//}
	maxWarn, err = flags.GetFloat64("maxWarn")
	if err != nil {
		return
	}
	//MinErr, err = flags.GetFloat64("MinErr")
	//if err != nil {
	//	return
	//}
	maxErr, err = flags.GetFloat64("maxErr")
	if err != nil {
		return
	}
	lim = app.LimitsStruct{
		MinWarn: fmt.Sprintf("%v", MinWarn),
		MaxWarn: fmt.Sprintf("%v", maxWarn),
		WarnMsg: warnMsg,
		MinErr:  fmt.Sprintf("%v", MinErr),
		MaxErr:  fmt.Sprintf("%v", maxErr),
		ErrMsg:  errMsg,
	}
	return lim, nil
}

func login(flags *pflag.FlagSet) (c app.Client, err error) {
	u := &url.URL{}
	urls, err := flags.GetString("url")
	if err != nil {
		return
	}
	user, err := flags.GetString("username")
	if err != nil {
		return
	}
	pww, err := flags.GetString("password")
	if err != nil {
		return
	}

	useCached, err := flags.GetBool("cachedCreds")
	if err != nil {
		return
	}
	u, _ = u.Parse(urls)
	c, err = app.NewClient(u, user, pww, !useCached)
	if err != nil {
		return
	}
	return
}
