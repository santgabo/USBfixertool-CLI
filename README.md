# USB Fixer CLI

USB Fixer CLI is a macOS command-line tool for inspecting, verifying, repairing, wiping, and formatting USB drives or external disks using native system utilities.

The goal is to provide a clearer and safer workflow than Disk Utility when a USB drive is corrupted, fails to mount, or needs to be prepared again from the terminal.

> Note: this README explains how to use the CLI. Installation is intentionally not documented yet.

## What It Does

- Lists external or removable disks detected by macOS.
- Inspects detailed information for a disk or volume.
- Verifies or repairs disks and volumes with `diskutil`.
- Runs filesystem-specific repair tools with `fsck` when requested.
- Formats whole USB disks with an explicit filesystem and partition scheme.
- Wipes partition metadata before reformatting.
- Shows planned operations with `--dry-run` before changing anything.

## Requirements

- macOS.
- Access to `diskutil`, which is included with macOS.
- Sufficient permissions for repair, wipe, or format operations when required by the system.

## Quick Start

First, identify the correct disk:

```bash
usbfix list
```

Then inspect it before running any destructive operation:

```bash
usbfix inspect --disk disk4
```

Verify a volume:

```bash
usbfix repair --volume disk4s1 --verify-only
```

Repair a volume:

```bash
usbfix repair --volume disk4s1 --repair
```

Preview a format operation without running it:

```bash
usbfix --dry-run format --disk disk4 --fs exfat --scheme mbr --label USBBOOT
```

When you are completely sure about the target disk, confirm a destructive operation explicitly:

```bash
usbfix format --disk disk4 --fs exfat --scheme mbr --label USBBOOT --confirm "ERASE disk4"
```

## Commands

### `usbfix list`

Lists candidate disks. By default, internal disks are filtered out and only external or removable media are shown.

```bash
usbfix list
usbfix list --removable-only
usbfix list --external-only
usbfix list --all
```

Useful options:

- `--all`: include internal and non-removable disks.
- `--removable-only`: show only removable media.
- `--external-only`: show only external disks.
- `--output table|json|plain`: change the output format.

### `usbfix inspect`

Shows detailed information for a whole disk or a volume. You must pass exactly one target: `--disk` or `--volume`.

```bash
usbfix inspect --disk disk4
usbfix inspect --volume disk4s1
usbfix inspect --disk /dev/disk4 --show-raw
```

Useful options:

- `--disk diskN`: inspect a whole disk.
- `--volume diskNsM`: inspect a volume or partition.
- `--show-raw`: include raw `diskutil info` output.

### `usbfix repair`

Verifies or repairs a disk or volume. If `--repair` is not provided, the default mode is verification.

```bash
usbfix repair --volume disk4s1
usbfix repair --volume disk4s1 --repair
usbfix repair --volume disk4s1 --repair --fsck
usbfix repair --disk disk4 --repair --force-unmount --mount-after
```

Useful options:

- `--disk diskN`: target a whole disk.
- `--volume diskNsM`: target a volume.
- `--verify-only`: verify without modifying.
- `--repair`: run repair actions.
- `--fsck`: allow tools such as `fsck_exfat` or `fsck_msdos`.
- `--force-unmount`: force unmount before repairing.
- `--mount-after`: mount after the operation.
- `--no-mount-after`: do not mount after the operation.
- `--retry N`: retry unmount or repair actions.
- `--timeout 60s`: limit the operation duration.

### `usbfix format`

Erases and formats a whole USB disk. This command is destructive and only accepts whole-disk identifiers such as `disk4`, not partitions such as `disk4s1`.

```bash
usbfix --dry-run format --disk disk4 --fs exfat --scheme mbr --label USBBOOT
usbfix format --disk disk4 --fs exfat --scheme mbr --label USBBOOT --confirm "ERASE disk4"
```

Required options:

- `--disk diskN`: whole disk to format.
- `--fs exfat|fat32|apfs|hfs+`: filesystem.
- `--scheme mbr|gpt|apm`: partition scheme.
- `--label LABEL`: volume label.

Useful options:

- `--case-sensitive`: use a case-sensitive variant when supported by the filesystem.
- `--force-unmount`: force unmount before formatting.
- `--mount-after`: mount after formatting.
- `--no-mount-after`: unmount after formatting.

### `usbfix wipe`

Wipes partition metadata or overwrites the first megabyte of the disk before reformatting. This command is also destructive.

```bash
usbfix --dry-run wipe --disk disk4 --quick --scheme mbr
usbfix wipe --disk disk4 --quick --scheme gpt --confirm "ERASE disk4"
usbfix wipe --disk disk4 --zero-first-mb --force-unmount --confirm "ERASE disk4"
```

Useful options:

- `--quick`: quickly wipe partition metadata with `diskutil`.
- `--scheme mbr|gpt|apm`: scheme used by quick wipe mode.
- `--zero-first-mb`: overwrite the first megabyte with zeros.
- `--force-unmount`: force unmount before wiping.

You must choose at least one wipe mode: `--quick` or `--zero-first-mb`.

### `usbfix version`

Prints binary version information.

```bash
usbfix version
```

## Global Flags

Global flags can be used before the command:

```bash
usbfix --output json list
usbfix --dry-run repair --volume disk4s1 --repair
```

General options:

- `--output table|json|plain`: output format.
- `--quiet`: suppress non-essential diagnostics.
- `--verbose`: print verbose diagnostics.
- `--no-color`: disable colored output.
- `--log-file PATH`: write diagnostics to a log file.
- `--interactive`: prompt for missing values and confirmations.
- `--non-interactive`: fail when a required value is missing.

Safety options:

- `--dry-run`: print the planned operation without modifying disks.
- `--confirm "ERASE diskN"`: confirm destructive operations.
- `--yes --i-know-this-erases-data`: confirm without a prompt, explicitly.
- `--allow-internal`: allow operations against internal disks.
- `--allow-non-removable`: allow operations against non-removable disks.
- `--require-removable`: require the target to report removable media.
- `--whole-disk-only`: reject volume targets.
- `--snapshot-info PATH`: save `diskutil info` output before a destructive operation.

## Safety

USB Fixer CLI avoids dangerous decisions by default:

- It does not choose a filesystem, partition scheme, or label automatically.
- It rejects partition identifiers when a command requires a whole disk.
- It inspects disk information before wiping or formatting.
- It refuses internal disks unless `--allow-internal` is provided.
- It requires textual confirmation for destructive operations.
- It supports `--dry-run` so you can review the plan before executing.

The confirmation phrase always follows this format:

```text
ERASE diskN
```

For example:

```bash
usbfix format --disk disk4 --fs exfat --scheme mbr --label USBBOOT --confirm "ERASE disk4"
```

## Output Formats

The default format is `table`, intended for human-readable output.

```bash
usbfix list --output table
usbfix list --output json
usbfix list --output plain
```

Use `json` for scripts or automation:

```bash
usbfix --output json inspect --disk disk4
```

## Common Examples

List external or removable disks:

```bash
usbfix list
```

Inspect a USB drive before changing it:

```bash
usbfix inspect --disk disk4
```

Repair an exFAT partition and allow `fsck`:

```bash
usbfix repair --volume disk4s1 --repair --fsck
```

Format a USB drive as exFAT for broad compatibility:

```bash
usbfix format --disk disk4 --fs exfat --scheme mbr --label USBBOOT --confirm "ERASE disk4"
```

Prepare a disk by wiping partition metadata:

```bash
usbfix wipe --disk disk4 --quick --scheme gpt --confirm "ERASE disk4"
```

Preview a destructive action without running it:

```bash
usbfix --dry-run wipe --disk disk4 --quick --scheme mbr
```

## Exit Codes

- `0`: success.
- `1`: general runtime error.
- `2`: usage error, invalid arguments, or invalid flags.
- `3`: operation cancelled.
- `126`: permission denied.
- `127`: required command not found.
- `130`: interrupted with `Ctrl-C`.

## Project Status

This project currently focuses on macOS and USB or external drives. It does not attempt deleted-file recovery or forensic data recovery; its scope is inspecting, repairing, wiping, and preparing disks from the terminal with explicit safety checks.
