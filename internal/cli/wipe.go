package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/santgabo/usbfixertool-cli/internal/usbfix"
	"github.com/spf13/cobra"
)

type wipeOptions struct {
	disk         string
	scheme       string
	quick        bool
	zeroFirstMB  bool
	forceUnmount bool
}

func newWipeCommand(cfg *config) *cobra.Command {
	opts := &wipeOptions{}
	cmd := &cobra.Command{
		Use:   "wipe",
		Short: "Erase partition metadata before reformatting",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.diskutil.CheckAvailable(); err != nil {
				return err
			}
			if !opts.quick && !opts.zeroFirstMB {
				return usbfix.Usagef("choose at least one wipe mode: --quick or --zero-first-mb")
			}

			var err error
			opts.disk, err = requireValue(cmd, cfg, opts.disk, "disk", "Target whole disk, for example disk4")
			if err != nil {
				return err
			}
			disk, err := usbfix.NormalizeWholeDiskIdentifier(opts.disk)
			if err != nil {
				return err
			}

			scheme := ""
			if opts.quick {
				opts.scheme, err = requireValue(cmd, cfg, opts.scheme, "scheme", "Partition scheme for quick wipe (mbr, gpt, apm)")
				if err != nil {
					return err
				}
				scheme, err = usbfix.NormalizeScheme(opts.scheme)
				if err != nil {
					return err
				}
			}

			info, err := ensureWholeDiskSafety(cmd, cfg, disk)
			if err != nil {
				return err
			}
			if !cfg.quiet {
				printTargetSummary(cmd.ErrOrStderr(), disk, info, "Wipe target summary")
				if scheme != "" {
					fmt.Fprintf(cmd.ErrOrStderr(), "Partition scheme: %s\n", scheme)
				}
			}
			if err := confirmErase(cmd, cfg, disk); err != nil {
				return err
			}

			result := operationResult{Target: disk, Mode: "wipe", DryRun: cfg.dryRun}
			if cfg.dryRun {
				appendWipeDryRun(&result, opts, scheme, disk)
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
			if opts.zeroFirstMB {
				out, err := cfg.diskutil.ZeroFirstMB(ctx, disk)
				if err != nil {
					return err
				}
				result.Actions = append(result.Actions, actionResult{Command: actionCommand("dd", "if=/dev/zero", "of="+usbfix.RawDevicePath(disk), "bs=1m", "count=1"), Output: string(out)})
			}
			if opts.quick {
				out, err := cfg.diskutil.FreeDisk(ctx, scheme, disk)
				if err != nil {
					return err
				}
				result.Actions = append(result.Actions, actionResult{Command: actionCommand("diskutil", "eraseDisk", "free", "EMPTY", scheme, usbfix.DevicePath(disk)), Output: string(out)})
			}
			return renderOperationResult(cmd.OutOrStdout(), cfg.output, result)
		},
	}
	cmd.Flags().StringVar(&opts.disk, "disk", "", "whole disk identifier, for example disk4")
	cmd.Flags().StringVar(&opts.scheme, "scheme", "", "partition scheme: mbr, gpt, or apm")
	cmd.Flags().BoolVar(&opts.quick, "quick", false, "quickly erase partition metadata with diskutil")
	cmd.Flags().BoolVar(&opts.zeroFirstMB, "zero-first-mb", false, "overwrite the first megabyte with zeros")
	cmd.Flags().BoolVar(&opts.forceUnmount, "force-unmount", false, "force unmount before wiping")
	return cmd
}

func appendWipeDryRun(result *operationResult, opts *wipeOptions, scheme, disk string) {
	if opts.forceUnmount {
		appendDryRun(result, actionCommand("diskutil", "unmountDisk", "force", usbfix.DevicePath(disk)))
	}
	if opts.zeroFirstMB {
		appendDryRun(result, actionCommand("dd", "if=/dev/zero", "of="+usbfix.RawDevicePath(disk), "bs=1m", "count=1"))
	}
	if opts.quick {
		appendDryRun(result, actionCommand("diskutil", "eraseDisk", "free", "EMPTY", scheme, usbfix.DevicePath(disk)))
	}
}
