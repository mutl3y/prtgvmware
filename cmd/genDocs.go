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
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"log"
	"os"
)

// genDocsCmd represents the genDocs command
var genDocsCmd = &cobra.Command{
	Use:   "genDocs",
	Short: "Create documentation for app",
	Long: `Create documentation for app

Supports Markdown, Rest and man page formats
`,
	Run: func(cmd *cobra.Command, args []string) {
		flags := cmd.Flags()
		header := &doc.GenManHeader{
			Title: "prtg_client_util",
		}

		path, err := flags.GetString("Folder")
		if err != nil {
			fmt.Printf("failed getting type %v", err)
		}

		fmt.Println("Creating documentation in ", path)

		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.Mkdir(path, os.ModePerm); err != nil {
				log.Fatalf("failed creating docs folder %v", err)
			}

		}

		op, err := flags.GetString("Type")
		if err != nil {
			fmt.Printf("failed getting type %v", err)
		}

		switch op {
		case "man":
			err := doc.GenManTree(rootCmd, header, path)
			if err != nil {
				log.Fatal(err)
			}
		case "rest":
			err := doc.GenReSTTree(rootCmd, path)
			if err != nil {
				log.Fatal(err)
			}
		case "markdown":
			err := doc.GenMarkdownTree(rootCmd, path)
			if err != nil {
				log.Fatal(err)
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(genDocsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// genDocsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	genDocsCmd.Flags().StringP("Type", "T", "markdown", "markdown,rest or man")
	genDocsCmd.Flags().StringP("Folder", "f", "./docs", "folder to create docs in")
	_ = genDocsCmd.MarkFlagDirname("Folder")
}
