package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "v0.1.1"
	commit  = "none"
	date    = "unknown"
)

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "usbfix %s (commit %s, built %s)\n", version, commit, date)
			return err
		},
	}
}
