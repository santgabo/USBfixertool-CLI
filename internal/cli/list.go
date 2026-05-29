package cli

import (
	"github.com/santgabo/usbfixertool-cli/internal/usbfix"
	"github.com/spf13/cobra"
)

type listOptions struct {
	all           bool
	removableOnly bool
	externalOnly  bool
}

func newListCommand(cfg *config) *cobra.Command {
	opts := &listOptions{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List candidate external or removable disks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.diskutil.CheckAvailable(); err != nil {
				return err
			}
			disks, err := cfg.diskutil.List(cmd.Context())
			if err != nil {
				return err
			}
			disks = filterDisks(disks, opts)
			return writeDiskList(cmd.OutOrStdout(), cfg.output, disks)
		},
	}
	cmd.Flags().BoolVar(&opts.all, "all", false, "include internal and non-removable disks")
	cmd.Flags().BoolVar(&opts.removableOnly, "removable-only", false, "show only removable media")
	cmd.Flags().BoolVar(&opts.externalOnly, "external-only", false, "show only external disks")
	return cmd
}

func filterDisks(disks []usbfix.DiskEntry, opts *listOptions) []usbfix.DiskEntry {
	filtered := make([]usbfix.DiskEntry, 0, len(disks))
	for _, disk := range disks {
		if !opts.all {
			if disk.Internal {
				continue
			}
			if !disk.External && !disk.Removable {
				continue
			}
		}
		if opts.removableOnly && !disk.Removable {
			continue
		}
		if opts.externalOnly && !disk.External {
			continue
		}
		filtered = append(filtered, disk)
	}
	return filtered
}
