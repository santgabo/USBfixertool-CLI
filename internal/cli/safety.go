package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/santgabo/usbfixertool-cli/internal/usbfix"
	"github.com/spf13/cobra"
)

func ensureWholeDiskSafety(cmd *cobra.Command, cfg *config, rawDisk string) (usbfix.DeviceInfo, error) {
	disk, err := usbfix.NormalizeWholeDiskIdentifier(rawDisk)
	if err != nil {
		return usbfix.DeviceInfo{}, err
	}
	info, err := cfg.diskutil.Info(cmd.Context(), disk)
	if err != nil {
		return usbfix.DeviceInfo{}, err
	}
	if !info.Whole {
		return usbfix.DeviceInfo{}, usbfix.Runtimef("%s is not a whole disk; use diskN, not diskNsM", usbfix.DevicePath(disk))
	}
	if info.Internal && !cfg.allowInternal {
		return usbfix.DeviceInfo{}, usbfix.Runtimef("refusing to operate on internal disk %s; pass --allow-internal only if you are certain", usbfix.DevicePath(disk))
	}
	if cfg.requireRemovable && !info.Removable {
		return usbfix.DeviceInfo{}, usbfix.Runtimef("refusing to operate on %s because it is not reported as removable", usbfix.DevicePath(disk))
	}
	if !cfg.allowNonRemovable && !info.Removable && !info.External {
		return usbfix.DeviceInfo{}, usbfix.Runtimef("refusing to operate on %s because it is neither removable nor external; pass --allow-non-removable only if intended", usbfix.DevicePath(disk))
	}
	if cfg.snapshotInfoPath != "" {
		if err := os.WriteFile(cfg.snapshotInfoPath, []byte(info.Raw), 0o600); err != nil {
			return usbfix.DeviceInfo{}, usbfix.NewError(usbfix.ExitRuntime, "failed to write snapshot info", err)
		}
	}
	return info, nil
}

func printTargetSummary(w io.Writer, disk string, info usbfix.DeviceInfo, title string) {
	fmt.Fprintln(w, "------------------------------------------------------------")
	fmt.Fprintf(w, "%s\n", title)
	fmt.Fprintf(w, "Target disk    : %s\n", usbfix.DevicePath(disk))
	fmt.Fprintf(w, "Whole disk     : %t\n", info.Whole)
	fmt.Fprintf(w, "Device Location: %s\n", firstText(info.DeviceLocation, "unknown"))
	fmt.Fprintf(w, "Internal       : %t\n", info.Internal)
	fmt.Fprintf(w, "Removable Media: %t\n", info.Removable)
	if info.Size != "" {
		fmt.Fprintf(w, "Disk Size      : %s\n", info.Size)
	}
	fmt.Fprintln(w, "------------------------------------------------------------")
}

func confirmErase(cmd *cobra.Command, cfg *config, disk string) error {
	if cfg.dryRun {
		return nil
	}
	expected := "ERASE " + disk
	if cfg.confirm != "" {
		if cfg.confirm == expected {
			return nil
		}
		return usbfix.Cancelledf("confirmation did not match; expected %q", expected)
	}
	if cfg.yes {
		if !cfg.destructiveIntent {
			return usbfix.Usagef("--yes requires --i-know-this-erases-data for destructive operations")
		}
		return nil
	}
	if cfg.interactive {
		fmt.Fprintf(cmd.ErrOrStderr(), "This operation will erase all content on %s.\n", usbfix.DevicePath(disk))
		fmt.Fprintf(cmd.ErrOrStderr(), "Type exactly %q to continue: ", expected)
		reader := bufio.NewReader(cmd.InOrStdin())
		response, err := reader.ReadString('\n')
		if err != nil && len(response) == 0 {
			return usbfix.NewError(usbfix.ExitCancelled, "confirmation prompt failed", err)
		}
		if strings.TrimSpace(response) != expected {
			return usbfix.Cancelledf("operation cancelled by user")
		}
		return nil
	}
	return usbfix.Cancelledf("confirmation required; rerun with --confirm %q or --interactive", expected)
}

func requireValue(cmd *cobra.Command, cfg *config, value, name, prompt string) (string, error) {
	value = strings.TrimSpace(value)
	if value != "" {
		return value, nil
	}
	if cfg.interactive {
		fmt.Fprintf(cmd.ErrOrStderr(), "%s: ", prompt)
		reader := bufio.NewReader(cmd.InOrStdin())
		response, err := reader.ReadString('\n')
		if err != nil && len(response) == 0 {
			return "", usbfix.NewError(usbfix.ExitUsage, "failed to read "+name, err)
		}
		response = strings.TrimSpace(response)
		if response == "" {
			return "", usbfix.Usagef("%s is required", name)
		}
		return response, nil
	}
	return "", usbfix.Usagef("%s is required; pass --%s or use --interactive", name, name)
}
