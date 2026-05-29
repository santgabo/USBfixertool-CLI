package cli

import (
	"context"
	"strings"
	"time"

	"github.com/santgabo/usbfixertool-cli/internal/usbfix"
	"github.com/spf13/cobra"
)

type repairOptions struct {
	disk         string
	volume       string
	verifyOnly   bool
	repair       bool
	fsck         bool
	forceUnmount bool
	mountAfter   bool
	noMountAfter bool
	retry        int
	timeout      time.Duration
}

func newRepairCommand(cfg *config) *cobra.Command {
	opts := &repairOptions{retry: 1}
	cmd := &cobra.Command{
		Use:   "repair",
		Short: "Verify or repair a disk or volume",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.diskutil.CheckAvailable(); err != nil {
				return err
			}
			if (opts.disk == "") == (opts.volume == "") {
				return usbfix.Usagef("provide exactly one of --disk or --volume")
			}
			if opts.verifyOnly && opts.repair {
				return usbfix.Usagef("--verify-only and --repair cannot be used together")
			}
			if opts.mountAfter && opts.noMountAfter {
				return usbfix.Usagef("--mount-after and --no-mount-after cannot be used together")
			}
			if opts.retry < 1 {
				return usbfix.Usagef("--retry must be at least 1")
			}
			if cfg.wholeDiskOnly && opts.volume != "" {
				return usbfix.Usagef("--whole-disk-only rejects volume target %q", opts.volume)
			}
			if opts.fsck && opts.disk != "" {
				return usbfix.Usagef("--fsck requires --volume because filesystem-specific tools run on volumes")
			}

			ctx := cmd.Context()
			if opts.timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, opts.timeout)
				defer cancel()
			}

			target, isVolume, err := normalizeRepairTarget(opts)
			if err != nil {
				return err
			}
			mode := "verify"
			if opts.repair {
				mode = "repair"
			}
			result := operationResult{Target: target, Mode: mode, DryRun: cfg.dryRun}

			if cfg.dryRun {
				appendRepairDryRun(&result, opts, target, isVolume)
				return renderOperationResult(cmd.OutOrStdout(), cfg.output, result)
			}

			if opts.forceUnmount {
				out, err := runWithRetries(ctx, opts.retry, time.Second, func(ctx context.Context) ([]byte, error) {
					if isVolume {
						return cfg.diskutil.UnmountVolume(ctx, target, true)
					}
					return cfg.diskutil.UnmountDisk(ctx, target, true)
				})
				if err != nil {
					return err
				}
				result.Actions = append(result.Actions, actionResult{Command: repairUnmountCommand(target, isVolume), Output: string(out)})
			}

			if opts.repair {
				out, err := runWithRetries(ctx, opts.retry, time.Second, func(ctx context.Context) ([]byte, error) {
					if isVolume {
						return cfg.diskutil.RepairVolume(ctx, target)
					}
					return cfg.diskutil.RepairDisk(ctx, target)
				})
				if err != nil {
					return err
				}
				result.Actions = append(result.Actions, actionResult{Command: repairDiskutilCommand(target, isVolume, true), Output: string(out)})
			} else {
				out, err := runWithRetries(ctx, opts.retry, time.Second, func(ctx context.Context) ([]byte, error) {
					if isVolume {
						return cfg.diskutil.VerifyVolume(ctx, target)
					}
					return cfg.diskutil.VerifyDisk(ctx, target)
				})
				if err != nil {
					return err
				}
				result.Actions = append(result.Actions, actionResult{Command: repairDiskutilCommand(target, isVolume, false), Output: string(out)})
			}

			if opts.fsck {
				info, err := cfg.diskutil.Info(ctx, target)
				if err != nil {
					return err
				}
				out, err := runWithRetries(ctx, opts.retry, time.Second, func(ctx context.Context) ([]byte, error) {
					return cfg.diskutil.Fsck(ctx, info.FileSystem, target, opts.repair)
				})
				if err != nil {
					return err
				}
				fsckName := "fsck"
				if strings.Contains(strings.ToLower(info.FileSystem), "exfat") {
					fsckName = "fsck_exfat"
				}
				if strings.Contains(strings.ToLower(info.FileSystem), "fat") {
					fsckName = "fsck_msdos"
				}
				flag := "-n"
				if opts.repair {
					flag = "-y"
				}
				result.Actions = append(result.Actions, actionResult{Command: actionCommand(fsckName, flag, usbfix.DevicePath(target)), Output: string(out)})
			}

			if opts.mountAfter {
				out, err := mountRepairTarget(ctx, cfg, target, isVolume)
				if err != nil {
					return err
				}
				result.Actions = append(result.Actions, actionResult{Command: repairMountCommand(target, isVolume), Output: string(out)})
			}

			return renderOperationResult(cmd.OutOrStdout(), cfg.output, result)
		},
	}
	cmd.Flags().StringVar(&opts.disk, "disk", "", "whole disk identifier, for example disk4")
	cmd.Flags().StringVar(&opts.volume, "volume", "", "volume identifier, for example disk4s1")
	cmd.Flags().BoolVar(&opts.verifyOnly, "verify-only", false, "verify without modifying")
	cmd.Flags().BoolVar(&opts.repair, "repair", false, "run repair actions")
	cmd.Flags().BoolVar(&opts.fsck, "fsck", false, "allow filesystem-specific fsck tools")
	cmd.Flags().BoolVar(&opts.forceUnmount, "force-unmount", false, "force unmount before repairing a whole disk")
	cmd.Flags().BoolVar(&opts.mountAfter, "mount-after", false, "mount the disk after the operation")
	cmd.Flags().BoolVar(&opts.noMountAfter, "no-mount-after", false, "do not mount after the operation")
	cmd.Flags().IntVar(&opts.retry, "retry", 1, "number of retries for unmount or repair actions")
	cmd.Flags().DurationVar(&opts.timeout, "timeout", 0, "operation timeout, for example 60s")
	return cmd
}

func normalizeRepairTarget(opts *repairOptions) (string, bool, error) {
	if opts.volume != "" {
		id, err := usbfix.NormalizeVolumeIdentifier(opts.volume)
		return id, true, err
	}
	id, err := usbfix.NormalizeWholeDiskIdentifier(opts.disk)
	return id, false, err
}

func appendRepairDryRun(result *operationResult, opts *repairOptions, target string, isVolume bool) {
	if opts.forceUnmount {
		appendDryRun(result, repairUnmountCommand(target, isVolume))
	}
	appendDryRun(result, repairDiskutilCommand(target, isVolume, opts.repair))
	if opts.fsck && isVolume {
		appendDryRun(result, actionCommand("fsck_*", usbfix.DevicePath(target)))
	}
	if opts.mountAfter {
		appendDryRun(result, repairMountCommand(target, isVolume))
	}
}

func repairDiskutilCommand(target string, isVolume bool, repair bool) string {
	action := "verifyDisk"
	if isVolume {
		action = "verifyVolume"
	}
	if repair && isVolume {
		action = "repairVolume"
	}
	if repair && !isVolume {
		action = "repairDisk"
	}
	return actionCommand("diskutil", action, usbfix.DevicePath(target))
}

func repairUnmountCommand(target string, isVolume bool) string {
	if isVolume {
		return actionCommand("diskutil", "unmount", "force", usbfix.DevicePath(target))
	}
	return actionCommand("diskutil", "unmountDisk", "force", usbfix.DevicePath(target))
}

func repairMountCommand(target string, isVolume bool) string {
	if isVolume {
		return actionCommand("diskutil", "mount", usbfix.DevicePath(target))
	}
	return actionCommand("diskutil", "mountDisk", usbfix.DevicePath(target))
}

func mountRepairTarget(ctx context.Context, cfg *config, target string, isVolume bool) ([]byte, error) {
	if isVolume {
		return cfg.diskutil.MountVolume(ctx, target)
	}
	return cfg.diskutil.MountDisk(ctx, target)
}
