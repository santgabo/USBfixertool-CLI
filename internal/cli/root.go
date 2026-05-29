package cli

import (
	"io"
	"os"

	"github.com/santgabo/usbfixertool-cli/internal/usbfix"
	"github.com/spf13/cobra"
)

type Dependencies struct {
	Runner usbfix.SystemRunner
	Stdin  io.Reader
}

type config struct {
	runner   usbfix.SystemRunner
	diskutil *usbfix.Diskutil
	stdin    io.Reader

	output         string
	quiet          bool
	verbose        bool
	noColor        bool
	logFile        string
	interactive    bool
	nonInteractive bool

	dryRun            bool
	confirm           string
	yes               bool
	destructiveIntent bool
	allowInternal     bool
	allowNonRemovable bool
	requireRemovable  bool
	wholeDiskOnly     bool
	snapshotInfoPath  string
}

func NewRootCommand(deps Dependencies) *cobra.Command {
	if deps.Runner == nil {
		deps.Runner = usbfix.ExecRunner{}
	}
	if deps.Stdin == nil {
		deps.Stdin = os.Stdin
	}

	cfg := &config{
		runner:   deps.Runner,
		diskutil: usbfix.NewDiskutil(deps.Runner),
		stdin:    deps.Stdin,
		output:   "table",
	}

	cmd := &cobra.Command{
		Use:           "usbfix",
		Short:         "Inspect, repair, and format USB drives on macOS",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputFormat(cfg.output); err != nil {
				return err
			}
			if cfg.interactive && cfg.nonInteractive {
				return usbfix.Usagef("--interactive and --non-interactive cannot be used together")
			}
			return nil
		},
	}
	cmd.SetIn(deps.Stdin)
	cmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		return usbfix.NewError(usbfix.ExitUsage, err.Error(), nil)
	})

	flags := cmd.PersistentFlags()
	flags.StringVar(&cfg.output, "output", "table", "output format: table, json, or plain")
	flags.BoolVar(&cfg.quiet, "quiet", false, "suppress non-essential diagnostics")
	flags.BoolVar(&cfg.verbose, "verbose", false, "print verbose diagnostics")
	flags.BoolVar(&cfg.noColor, "no-color", false, "disable colored output")
	flags.StringVar(&cfg.logFile, "log-file", "", "write diagnostics to a log file")
	flags.BoolVar(&cfg.interactive, "interactive", false, "prompt for missing values and confirmations")
	flags.BoolVar(&cfg.nonInteractive, "non-interactive", false, "fail instead of prompting")

	flags.BoolVar(&cfg.dryRun, "dry-run", false, "print the planned operation without changing disks")
	flags.StringVar(&cfg.confirm, "confirm", "", "explicit confirmation phrase, for example \"ERASE disk4\"")
	flags.BoolVar(&cfg.yes, "yes", false, "answer yes to confirmation prompts; requires --i-know-this-erases-data for destructive operations")
	flags.BoolVar(&cfg.destructiveIntent, "i-know-this-erases-data", false, "acknowledge destructive intent when using --yes")
	flags.BoolVar(&cfg.allowInternal, "allow-internal", false, "allow destructive operations against internal disks")
	flags.BoolVar(&cfg.allowNonRemovable, "allow-non-removable", false, "allow destructive operations against disks that are not removable or external")
	flags.BoolVar(&cfg.requireRemovable, "require-removable", false, "require the target disk to report removable media")
	flags.BoolVar(&cfg.wholeDiskOnly, "whole-disk-only", false, "require whole-disk targets for operations that accept either disks or volumes")
	flags.StringVar(&cfg.snapshotInfoPath, "snapshot-info", "", "write diskutil info output to PATH before a destructive operation")

	cmd.AddCommand(
		newListCommand(cfg),
		newInspectCommand(cfg),
		newRepairCommand(cfg),
		newFormatCommand(cfg),
		newWipeCommand(cfg),
		newVersionCommand(),
	)

	return cmd
}
