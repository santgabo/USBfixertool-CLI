package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/santgabo/usbfixertool-cli/internal/usbfix"
	"github.com/spf13/cobra"
)

type formatOptions struct {
	disk           string
	filesystem     string
	scheme         string
	label          string
	allocationUnit string
	clusterSize    string
	caseSensitive  bool
	forceUnmount   bool
	mountAfter     bool
	noMountAfter   bool
}

func newFormatCommand(cfg *config) *cobra.Command {
	opts := &formatOptions{}
	cmd := &cobra.Command{
		Use:   "format",
		Short: "Erase and format a whole USB disk",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.diskutil.CheckAvailable(); err != nil {
				return err
			}
			if opts.mountAfter && opts.noMountAfter {
				return usbfix.Usagef("--mount-after and --no-mount-after cannot be used together")
			}
			if opts.allocationUnit != "" {
				return usbfix.Usagef("--allocation-unit is not implemented yet")
			}
			if opts.clusterSize != "" {
				return usbfix.Usagef("--cluster-size is not implemented yet")
			}

			var err error
			opts.disk, err = requireValue(cmd, cfg, opts.disk, "disk", "Target whole disk, for example disk4")
			if err != nil {
				return err
			}
			opts.filesystem, err = requireValue(cmd, cfg, opts.filesystem, "fs", "Filesystem (exfat, fat32, apfs, hfs+)")
			if err != nil {
				return err
			}
			opts.scheme, err = requireValue(cmd, cfg, opts.scheme, "scheme", "Partition scheme (mbr, gpt, apm)")
			if err != nil {
				return err
			}
			opts.label, err = requireValue(cmd, cfg, opts.label, "label", "Volume label")
			if err != nil {
				return err
			}

			disk, err := usbfix.NormalizeWholeDiskIdentifier(opts.disk)
			if err != nil {
				return err
			}
			fs, err := usbfix.NormalizeFilesystem(opts.filesystem, opts.caseSensitive)
			if err != nil {
				return err
			}
			scheme, err := usbfix.NormalizeScheme(opts.scheme)
			if err != nil {
				return err
			}
			labelValidation, err := usbfix.ValidateLabel(opts.filesystem, opts.label)
			if err != nil {
				return err
			}

			info, err := ensureWholeDiskSafety(cmd, cfg, disk)
			if err != nil {
				return err
			}
			if !cfg.quiet {
				for _, warning := range labelValidation.Warnings {
					fmt.Fprintf(cmd.ErrOrStderr(), "warning: %s\n", warning)
				}
				printTargetSummary(cmd.ErrOrStderr(), disk, info, "Format target summary")
				fmt.Fprintf(cmd.ErrOrStderr(), "New filesystem : %s\n", fs)
				fmt.Fprintf(cmd.ErrOrStderr(), "New scheme     : %s\n", scheme)
				fmt.Fprintf(cmd.ErrOrStderr(), "New label      : %s\n", opts.label)
			}
			if err := confirmErase(cmd, cfg, disk); err != nil {
				return err
			}

			result := operationResult{Target: disk, Mode: "format", DryRun: cfg.dryRun}
			if cfg.dryRun {
				appendFormatDryRun(&result, opts, fs, scheme, disk)
				return renderOperationResult(cmd.OutOrStdout(), cfg.output, result)
			}

			ctx := cmd.Context()
			if opts.forceUnmount {
				out, err := runWithRetries(ctx, 1, time.Second, func(ctx context.Context) ([]byte, error) {
					return cfg.diskutil.UnmountDisk(ctx, disk, true)
				})
				if err != nil {
					return err
				}
				result.Actions = append(result.Actions, actionResult{Command: actionCommand("diskutil", "unmountDisk", "force", usbfix.DevicePath(disk)), Output: string(out)})
			}

			out, err := cfg.diskutil.EraseDisk(ctx, fs, opts.label, scheme, disk)
			if err != nil {
				return err
			}
			result.Actions = append(result.Actions, actionResult{Command: actionCommand("diskutil", "eraseDisk", fs, opts.label, scheme, usbfix.DevicePath(disk)), Output: string(out)})

			if opts.noMountAfter {
				out, err := cfg.diskutil.UnmountDisk(ctx, disk, false)
				if err != nil {
					return err
				}
				result.Actions = append(result.Actions, actionResult{Command: actionCommand("diskutil", "unmountDisk", usbfix.DevicePath(disk)), Output: string(out)})
			}

			return renderOperationResult(cmd.OutOrStdout(), cfg.output, result)
		},
	}
	cmd.Flags().StringVar(&opts.disk, "disk", "", "whole disk identifier, for example disk4")
	cmd.Flags().StringVar(&opts.filesystem, "fs", "", "filesystem: exfat, fat32, apfs, or hfs+")
	cmd.Flags().StringVar(&opts.scheme, "scheme", "", "partition scheme: mbr, gpt, or apm")
	cmd.Flags().StringVar(&opts.label, "label", "", "volume label")
	cmd.Flags().StringVar(&opts.allocationUnit, "allocation-unit", "", "allocation unit size")
	cmd.Flags().StringVar(&opts.clusterSize, "cluster-size", "", "cluster size")
	cmd.Flags().BoolVar(&opts.caseSensitive, "case-sensitive", false, "use a case-sensitive variant when supported")
	cmd.Flags().BoolVar(&opts.forceUnmount, "force-unmount", false, "force unmount before formatting")
	cmd.Flags().BoolVar(&opts.mountAfter, "mount-after", false, "mount after formatting")
	cmd.Flags().BoolVar(&opts.noMountAfter, "no-mount-after", false, "unmount after formatting")
	return cmd
}

func appendFormatDryRun(result *operationResult, opts *formatOptions, fs, scheme, disk string) {
	if opts.forceUnmount {
		appendDryRun(result, actionCommand("diskutil", "unmountDisk", "force", usbfix.DevicePath(disk)))
	}
	appendDryRun(result, actionCommand("diskutil", "eraseDisk", fs, opts.label, scheme, usbfix.DevicePath(disk)))
	if opts.noMountAfter {
		appendDryRun(result, actionCommand("diskutil", "unmountDisk", usbfix.DevicePath(disk)))
	}
}
