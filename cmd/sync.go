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

// NewSyncCmd represents the sync command
func NewSyncCmd() *cobra.Command {
	var syncCmd = &cobra.Command{
		Use:   "sync",
		Short: "Sync the labels in the specified repositories with the specified labels",
		RunE: func(cmd *cobra.Command, args []string) error {
			labelList, err := parseLabelsInput(labels, jsonPath, yamlPath)
			if err != nil {
				return err
			}

			repoList, err := parser.Repo(repos)
			if err != nil {
				return fmt.Errorf("failed to parse repos option: %v", err)
			}

			in := cmd.InOrStdin()
			out := cmd.OutOrStdout()

			if !dryRun && !force {
				confirmed, err := confirm(in, out)
				if err != nil {
					return fmt.Errorf("failed to confirm execution: %v", err)
				}
				if !confirmed {
					_, _ = fmt.Fprintf(out, "Canceled execution\n")
					return nil
				}
			}

			e, err := executor.NewExecutor(dryRun)
			if err != nil {
				return fmt.Errorf("failed to create executor: %v", err)
			}

			err = e.Sync(out, repoList, labelList)
			if err != nil {
				return fmt.Errorf("failed to sync labels: %v", err)
			}

			return nil
		},
	}
	return syncCmd
}

func init() {
	syncCmd := NewSyncCmd()
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().StringVarP(&labels, "labels", "l", "", "Specify the labels to set in the format of 'label1:color1:description1[,label2:color2:description2,...]' (description can be omitted)")
	syncCmd.Flags().StringVar(&jsonPath, "json", "", "Specify the path to a JSON file containing labels to sync")
	syncCmd.Flags().StringVar(&yamlPath, "yaml", "", "Specify the path to a YAML file containing labels to sync")
	syncCmd.Flags().BoolVar(&force, "force", false, "Do not prompt for confirmation")
}
