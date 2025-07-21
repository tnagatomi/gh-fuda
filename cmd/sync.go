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
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/tnagatomi/gh-fuda/executor"
	"github.com/tnagatomi/gh-fuda/option"
	"github.com/tnagatomi/gh-fuda/parser"

	"github.com/spf13/cobra"
)

// NewSyncCmd represents the sync command
func NewSyncCmd(in io.Reader, out io.Writer) *cobra.Command {
	var syncCmd = &cobra.Command{
		Use:   "sync",
		Short: "Sync the labels in the specified repositories with the specified labels",
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonPath != "" && labels != "" {
				return errors.New("--labels (-l) and --json cannot be used together")
			}
			if jsonPath == "" && labels == "" {
				return errors.New("either --labels (-l) or --json must be specified")
			}

			var labelList []option.Label
			var err error
			if jsonPath != "" {
				labelList, err = parser.LabelFromJSON(jsonPath)
				if err != nil {
					return fmt.Errorf("failed to parse JSON file: %v", err)
				}
			} else {
				labelList, err = parser.Label(labels)
				if err != nil {
					return fmt.Errorf("failed to parse labels option: %v", err)
				}
			}

			repoList, err := parser.Repo(repos)
			if err != nil {
				return fmt.Errorf("failed to parse repos option: %v", err)
			}

			client, err := api.NewHTTPClient(api.ClientOptions{})
			if err != nil {
				return fmt.Errorf("failed to create gh http client: %v", err)
			}

			e, err := executor.NewExecutor(client, dryRun)
			if err != nil {
				return fmt.Errorf("failed to create executor: %v", err)
			}

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
	syncCmd := NewSyncCmd(os.Stdin, os.Stdout)
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().StringVarP(&labels, "labels", "l", "", "Specify the labels to set in the format of 'label1:color1:description1[,label2:color2:description2,...]' (description can be omitted)")
	syncCmd.Flags().StringVar(&jsonPath, "json", "", "Specify the path to a JSON file containing labels to sync")
	syncCmd.Flags().BoolVar(&force, "force", false, "Do not prompt for confirmation")
}
