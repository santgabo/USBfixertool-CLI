package cli

import (
	"github.com/santgabo/usbfixertool-cli/internal/usbfix"
	"github.com/spf13/cobra"
)

type inspectOptions struct {
	disk    string
	volume  string
	showRaw bool
}

func newInspectCommand(cfg *config) *cobra.Command {
	opts := &inspectOptions{}
	cmd := &cobra.Command{
		Use:   "inspect",
		Short: "Show detailed disk or volume information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.diskutil.CheckAvailable(); err != nil {
				return err
			}
			if (opts.disk == "") == (opts.volume == "") {
				return usbfix.Usagef("provide exactly one of --disk or --volume")
			}

			var id string
			var err error
			if opts.disk != "" {
				id, err = usbfix.NormalizeWholeDiskIdentifier(opts.disk)
			} else {
				id, err = usbfix.NormalizeVolumeIdentifier(opts.volume)
			}
			if err != nil {
				return err
			}

			info, err := cfg.diskutil.Info(cmd.Context(), id)
			if err != nil {
				return err
			}
			if !opts.showRaw {
				info.Raw = ""
			}
			return writeDeviceInfo(cmd.OutOrStdout(), cfg.output, info)
		},
	}
	cmd.Flags().StringVar(&opts.disk, "disk", "", "whole disk identifier, for example disk4")
	cmd.Flags().StringVar(&opts.volume, "volume", "", "volume identifier, for example disk4s1")
	cmd.Flags().BoolVar(&opts.showRaw, "show-raw", false, "include raw diskutil info output")
	return cmd
}
