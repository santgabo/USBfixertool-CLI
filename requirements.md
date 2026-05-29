# USB Fixer CLI Requirements

## Goal

Build a Go CLI for inspecting, repairing, and formatting corrupted USB drives on macOS. The tool should provide safer and more flexible workflows than the native Disk Utility UI, while avoiding dangerous defaults.

The CLI must never assume a filesystem, partition scheme, or volume label by default. If required values are missing, it should either prompt interactively or fail in non-interactive mode.

## Target Platform

- macOS.
- Uses native disk tools such as `diskutil` and filesystem-specific repair tools where appropriate.
- Initial scope is USB/external removable drives.

## Proposed Commands

```bash
usbfix list
usbfix inspect --disk disk4
usbfix repair --volume disk4s1
usbfix format --disk disk4 --fs exfat --scheme mbr --label USBBOOT
usbfix wipe --disk disk4
```

### `list`

List candidate external/removable disks and volumes.

Useful flags:

- `--output table|json|plain`
- `--all` to include internal or non-removable disks in the listing.
- `--removable-only` to show only removable media.
- `--external-only` to show only external disks.

### `inspect`

Show detailed information about a disk or volume before any destructive action.

Useful flags:

- `--disk diskN`
- `--volume diskNsM`
- `--output table|json|plain`
- `--show-raw` to include raw `diskutil` output for debugging.

### `repair`

Attempt to verify or repair a corrupted USB volume.

Useful flags:

- `--disk diskN`
- `--volume diskNsM`
- `--verify-only` to inspect without modifying.
- `--repair` to run repair actions.
- `--fsck` to allow filesystem-specific tools such as `fsck_exfat` or `fsck_msdos`.
- `--force-unmount`
- `--mount-after`
- `--no-mount-after`
- `--retry N` for unmount or repair retries.
- `--timeout duration`, for example `--timeout 60s`.

### `format`

Erase and format a whole USB disk.

Required values:

- Target disk.
- Filesystem.
- Partition scheme.
- Volume label.

Useful flags:

- `--disk diskN`
- `--fs exfat|fat32|apfs|hfs+`
- `--scheme mbr|gpt|apm`
- `--label LABEL`
- `--allocation-unit SIZE`
- `--cluster-size SIZE`
- `--case-sensitive`
- `--force-unmount`
- `--mount-after`
- `--no-mount-after`

The `format` command must reject partition identifiers such as `disk4s1`; it should only accept whole disks such as `disk4`.

### `wipe`

Erase partition metadata or perform a stronger cleanup step before reformatting.

Useful flags:

- `--disk diskN`
- `--scheme mbr|gpt|apm`
- `--quick`
- `--zero-first-mb`
- `--force-unmount`

This command should be treated as destructive and require confirmation.

## Interactive Behavior

The CLI should support both interactive and scriptable use.

Global flags:

- `--interactive`
- `--non-interactive`

Behavior:

- In interactive mode, ask for missing required values.
- In non-interactive mode, fail if a required value is missing.
- Do not silently default to `exfat`, `mbr`, or any label.
- For destructive commands, show a summary before continuing.

Example destructive confirmation:

```text
This operation will erase all content on /dev/disk4.
Type exactly "ERASE disk4" to continue:
```

## Safety Requirements

The CLI must prioritize data safety.

Required checks:

- Normalize disk identifiers, accepting both `disk4` and `/dev/disk4`.
- Validate whole-disk identifiers with a pattern like `diskN`.
- Reject partition identifiers for whole-disk operations.
- Verify that the target exists before running any operation.
- Inspect `diskutil info` before destructive operations.
- Abort by default if the disk is internal.
- Abort by default if the disk is not a whole disk when whole-disk operation is required.
- Prefer removable or external media for destructive operations.
- Print a target summary before destructive operations.

Useful safety flags:

- `--dry-run`
- `--confirm "ERASE diskN"`
- `--yes`
- `--allow-internal`
- `--allow-non-removable`
- `--require-removable`
- `--whole-disk-only`
- `--snapshot-info PATH`

`--yes` should not bypass all safety by itself. If supported, it should require an additional explicit destructive intent flag, such as:

```bash
--yes --i-know-this-erases-data
```

## Output Requirements

Global flags:

- `--output table|json|plain`
- `--quiet`
- `--verbose`
- `--no-color`
- `--log-file PATH`

Output rules:

- Human diagnostics, warnings, progress, and errors go to stderr.
- Command results go to stdout.
- JSON output must be stable enough for scripts.
- Errors should be concise and actionable.

## Label Rules

Labels should be validated according to the selected filesystem.

Initial rule from `format-exfat.sh`:

- exFAT labels should be limited to 11 characters.
- Prefer uppercase letters, numbers, `_`, and `-` for compatibility.

Future implementation should define filesystem-specific validation for FAT32, APFS, and HFS+.

## Exit Codes

Use predictable Unix-style exit codes.

- `0`: Success.
- `1`: General runtime error.
- `2`: Usage error or invalid arguments.
- `3`: Operation cancelled by user.
- `64-78`: Optional BSD `sysexits`-style specific failures.
- `126`: Permission denied.
- `127`: Required system command not found.
- `130`: Interrupted by Ctrl-C.

## Implementation Notes

- Use Go.
- Prefer Cobra for command and flag structure.
- Consider Viper if config files or environment variable support become useful.
- Keep command output testable by writing through command output streams instead of direct `os.Stdout` or `os.Stderr`.
- Use context cancellation and signal handling for long-running disk operations.
- Do not call `os.Exit` deep inside command logic; return errors and let the entrypoint map them to exit codes.

## Initial Non-Goals

- Linux and Windows support.
- Recovery of deleted files.
- Full forensic data recovery.
- Automatically choosing the best filesystem without user input.
- Formatting internal disks by default.
