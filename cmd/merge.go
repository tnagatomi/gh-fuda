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
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/tnagatomi/gh-fuda/executor"
	"github.com/tnagatomi/gh-fuda/parser"
)

var (
	fromLabel   string
	toLabel     string
	skipConfirm bool
)

// NewMergeCmd represents the merge command
func NewMergeCmd(in io.Reader, out io.Writer) *cobra.Command {
	var mergeCmd = &cobra.Command{
		Use:   "merge",
		Short: "Merge a source label into a target label across repositories",
		Long: `Merge a source label into a target label across repositories.

This command:
1. Adds the target label to all issues, PRs, and discussions that have the source label
2. Removes the source label from those items
3. Deletes the source label from the repository

Both the source (--from) and target (--to) labels must exist in each repository.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if fromLabel == toLabel {
				return fmt.Errorf("source and target labels must be different")
			}

			e, err := executor.NewExecutor(dryRun)
			if err != nil {
				return fmt.Errorf("failed to create executor: %v", err)
			}

			repoList, err := parser.Repo(repos)
			if err != nil {
				return fmt.Errorf("failed to parse repos option: %v", err)
			}

			if !dryRun && !skipConfirm {
				confirmed, err := confirm(in, out)
				if err != nil {
					return fmt.Errorf("failed to confirm execution: %v", err)
				}
				if !confirmed {
					_, _ = fmt.Fprintf(out, "Canceled execution\n")
					return nil
				}
			}

			err = e.Merge(out, repoList, fromLabel, toLabel)
			if err != nil {
				return fmt.Errorf("failed to merge labels: %v", err)
			}

			return nil
		},
	}
	return mergeCmd
}

func init() {
	mergeCmd := NewMergeCmd(os.Stdin, os.Stdout)
	rootCmd.AddCommand(mergeCmd)

	mergeCmd.Flags().StringVar(&fromLabel, "from", "", "Source label to merge from (will be deleted)")
	mergeCmd.Flags().StringVar(&toLabel, "to", "", "Target label to merge into")
	mergeCmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Do not prompt for confirmation")

	err := mergeCmd.MarkFlagRequired("from")
	if err != nil {
		fmt.Printf("Failed to mark flag required: %v\n", err)
	}
	err = mergeCmd.MarkFlagRequired("to")
	if err != nil {
		fmt.Printf("Failed to mark flag required: %v\n", err)
	}
}
