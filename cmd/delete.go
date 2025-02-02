/*
Copyright © 2024 Takayuki Nagatomi <tommyt6073@gmail.com>

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
	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/tnagatomi/gh-fuda/executor"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// NewDeleteCmd represents the delete command
func NewDeleteCmd(in io.Reader, out io.Writer) *cobra.Command {
	var deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete specified labels from the specified repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := api.NewHTTPClient(api.ClientOptions{})
			if err != nil {
				return fmt.Errorf("failed to create gh http client: %v", err)
			}

			e, err := executor.NewExecutor(client, dryRun)
			if err != nil {
				return fmt.Errorf("failed to create exector: %v", err)
			}

			if !dryRun && !force {
				confirmed, err := confirm(in, out)
				if err != nil {
					return fmt.Errorf("failed to confirm execution: %v", err)
				}
				if !confirmed {
					fmt.Fprintf(out, "Canceled execution\n")
					return nil
				}
			}

			err = e.Delete(out, repos, labels)
			if err != nil {
				return fmt.Errorf("failed to delete labels: %v", err)
			}

			return nil
		},
	}
	return deleteCmd
}

func init() {
	deleteCmd := NewDeleteCmd(os.Stdin, os.Stdout)
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().StringVarP(&labels, "labels", "l", "", "Specify the labels to delete in the format of 'label1[,label2,...]'")
	deleteCmd.Flags().BoolVar(&force, "force", false, "Do not prompt for confirmation")

	err := deleteCmd.MarkFlagRequired("labels")
	if err != nil {
		fmt.Printf("Failed to mark flag required: %v\n", err)
	}
}
