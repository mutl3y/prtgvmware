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

Supports Markdown, ReStructured Text Docs and man page formats
`,
	Run: func(cmd *cobra.Command, args []string) {
		flags := cmd.Flags()
		header := &doc.GenManHeader{
			Title: "prtg_client_util",
		}

		path, err := flags.GetString("folder")
		if err != nil {
			fmt.Printf("failed getting type %v", err)
		}

		fmt.Println("Creating documentation in ", path)

		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.Mkdir(path, os.ModePerm); err != nil {
				log.Fatalf("failed creating docs folder %v", err)
			}

		}

		op, err := flags.GetString("type")
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
		default:
			err := doc.GenMarkdownTree(rootCmd, path)
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(genDocsCmd)
	genDocsCmd.Flags().String("type", "markdown", "markdown,rest or man")
	genDocsCmd.Flags().StringP("folder", "f", "./docs", "folder to create docs in")
	_ = genDocsCmd.MarkFlagDirname("folder")
}
