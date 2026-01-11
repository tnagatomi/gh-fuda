/*
Copyright © 2025 Takayuki Nagatomi <tnagatomi@okweird.net>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tnagatomi/gh-fuda/executor"
	"github.com/tnagatomi/gh-fuda/parser"
)

// NewCreateCmd initialize the create command
func NewCreateCmd() *cobra.Command {
	var force bool

	var createCmd = &cobra.Command{
		Use:   "create",
		Short: "Create specified labels to the specified repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			labelList, err := parseLabelsInput(labels, jsonPath, yamlPath)
			if err != nil {
				return err
			}

			repoList, err := parser.Repo(repos)
			if err != nil {
				return fmt.Errorf("failed to parse repos option: %v", err)
			}

			e, err := executor.NewExecutor(dryRun)
			if err != nil {
				return fmt.Errorf("failed to create executor: %v", err)
			}

			out := cmd.OutOrStdout()
			err = e.Create(out, repoList, labelList, force)
			if err != nil {
				return fmt.Errorf("failed to create labels: %v", err)
			}

			return nil
		},
	}

	createCmd.Flags().BoolVarP(&force, "force", "f", false, "Update the label color and description if label already exists")

	return createCmd
}

func init() {
	createCmd := NewCreateCmd()
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringVarP(&labels, "labels", "l", "", "Specify the labels to create in the format of 'label1:color1:description1[,label2:color2:description2,...]' (description can be omitted)")
	createCmd.Flags().StringVar(&jsonPath, "json", "", "Specify the path to a JSON file containing labels to create")
	createCmd.Flags().StringVar(&yamlPath, "yaml", "", "Specify the path to a YAML file containing labels to create")
}
