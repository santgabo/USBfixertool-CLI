# USB Fixer CLI

[![skills.sh](https://skills.sh/b/santgabo/USBfixertool-CLI)](https://skills.sh/b/santgabo/USBfixertool-CLI)

USB Fixer CLI is a macOS command-line tool written in Go for inspecting, verifying, repairing, wiping, and formatting USB drives or external disks using native system utilities safely and transparently.

The tool acts as a safer, terminal-friendly alternative to the native Disk Utility application, protecting users against accidental data loss through rigorous verification layers.

---

## Table of Contents
- [Installation](#installation)
- [Agent Skills (AI)](#agent-skills-ai)
- [Key Features](#key-features)
- [Safety \& Protective Mechanisms](#safety--protective-mechanisms)
- [Guided Quick Start Workflow](#guided-quick-start-workflow)
- [Command Reference](#command-reference)
- [Global Flags](#global-flags)
- [Output Formats](#output-formats)
- [Exit Codes](#exit-codes)
- [Project Status \& Scope](#project-status--scope)

---

## Installation

### Via Homebrew (Recommended)
You can easily install USB Fixer CLI via Homebrew by running:

```bash
brew install santgabo/tap/usbfix
```

### From Source
If you prefer to build from source, ensure you have Go 1.23 or newer installed:

```bash
# Clone the repository
git clone https://github.com/santgabo/USBfixertool-CLI.git
cd USBfixertool-CLI

# Build the binary
go build -o usbfix ./cmd/usbfix

# (Optional) Install the binary in your GOPATH bin directory
go install ./cmd/usbfix
```

### Requirements
- macOS
- Access to `diskutil` (pre-installed on macOS)
- Sufficient permissions (some repair, wipe, or format operations may prompt for sudo or require elevated permissions)

---

## Agent Skills (AI)

This repository publishes **`usbfix-cli-safety`**, an agent skill that teaches AI assistants to operate `usbfix` safely — including mandatory confirmation before touching internal disks.

Install with [skills.sh](https://skills.sh):

```bash
# List available skills in this repo
npx skills add santgabo/USBfixertool-CLI --list

# Install for Cursor (project-local)
npx skills add santgabo/USBfixertool-CLI --skill usbfix-cli-safety -a cursor -y

# Install globally (available in all projects)
npx skills add santgabo/USBfixertool-CLI --skill usbfix-cli-safety -g -a cursor -y
```

The skill source lives at [`skills/usbfix-cli-safety/SKILL.md`](skills/usbfix-cli-safety/SKILL.md). After installation, agents read it before running any `usbfix` command that could modify disks.

---

## Key Features

- **Disk Identification**: Lists external or removable disks, filtering out internal system disks by default.
- **Deep Inspection**: Displays granular metadata for targeted disks or partitions.
- **Smart Repair**: Verifies and repairs disks or individual volumes using `diskutil` and filesystem-specific tools (`fsck`).
- **Secure Formatting**: Safely formats whole drives using explicit schemes, filesystems, and volume labels.
- **Metadata Wiping**: Destroys leftover partition layouts/metadata or zeros the boot sector before a reformat.
- **Dry-Run Preview**: Supports dry-run execution (`--dry-run`) across all modifying operations to show actions before performing them.

---

## Safety & Protective Mechanisms

Disk operations are high-stakes. USB Fixer CLI enforces strict safety mechanisms:

1. **No Dangerous Defaults**: The tool never makes assumptions about filesystems, partition schemes, or volume labels. Missing required values will trigger an interactive prompt (if `--interactive` is active) or fail immediately.
2. **Explicit Target Scoping**: Whole-disk destructive actions (like `format` or `wipe`) will refuse partition identifiers (e.g. `disk4s1`) and demand a whole disk identifier (e.g. `disk4`).
3. **Internal Disk Lock**: Operations targeting internal drives are blocked automatically. To override this, you must explicitly supply `--allow-internal` after verifying you've targeted the correct drive.
4. **Textual Confirmation**: Destructive operations cannot be executed blindly. They require typing or passing a specific confirmation phrase:
   ```text
   ERASE diskN
   ```
   *Example:* `--confirm "ERASE disk4"`

---

## Guided Quick Start Workflow

Follow this standard workflow to safely format or repair a USB drive:

### Step 1: Identify the target drive
List external and removable drives connected to your macOS machine:
```bash
usbfix list
```

### Step 2: Inspect the drive metadata
Retrieve detailed details to make sure you have the right disk:
```bash
usbfix inspect --disk disk4
```

### Step 3: Verify the drive health (Safe/Read-only)
Run a verify-only repair command to inspect for corruption:
```bash
usbfix repair --volume disk4s1 --verify-only
```

### Step 4: Preview a destructive action (Dry Run)
Inspect exactly what commands would execute behind the scenes:
```bash
usbfix --dry-run format --disk disk4 --fs exfat --scheme mbr --label USBBOOT
```

### Step 5: Execute and Confirm
Once absolutely certain of the target disk, perform the destructive format:
```bash
usbfix format --disk disk4 --fs exfat --scheme mbr --label USBBOOT --confirm "ERASE disk4"
```

---

## Command Reference

### `usbfix list`
Lists all candidate drives. By default, it hides internal disks and displays only removable/external media.

```bash
usbfix list [flags]
```

**Useful Flags:**
- `--all`: Includes internal and non-removable system disks.
- `--removable-only`: Displays only removable media.
- `--external-only`: Displays only external drives.
- `--output <table|json|plain>`: Selects output format.

---

### `usbfix inspect`
Inspects and outputs detailed system information for a single disk or volume partition. You must pass exactly one target: `--disk` or `--volume`.

```bash
usbfix inspect --disk diskN
usbfix inspect --volume diskNsM
```

**Useful Flags:**
- `--disk <diskN>`: Specifies whole disk to inspect.
- `--volume <diskNsM>`: Specifies volume/partition partition to inspect.
- `--show-raw`: Prints the raw `diskutil info` plist response along with structured output.

---

### `usbfix repair`
Verifies or repairs the target volume/disk. By default, it operates in verification-only mode unless `--repair` is specified.

```bash
usbfix repair --volume diskNsM [flags]
```

**Useful Flags:**
- `--disk <diskN>`: Targets a whole disk.
- `--volume <diskNsM>`: Targets a single volume partition.
- `--verify-only`: Verifies integrity without modifying any data (default).
- `--repair`: Allows writing repairs to the disk.
- `--fsck`: Authorizes running filesystem-specific repair helpers (such as `fsck_exfat` or `fsck_msdos`).
- `--force-unmount`: Force unmounts the volume before repairing.
- `--mount-after` / `--no-mount-after`: Controls whether the target is mounted back after the operation.
- `--retry <N>`: Retries unmounting or repair steps up to N times.
- `--timeout <duration>`: Maximum duration allowed for the operation (e.g., `--timeout 60s`).

---

### `usbfix format`
Erases and formats a whole USB drive. This command is destructive and only accepts whole-disk targets.

```bash
usbfix format --disk diskN --fs <filesystem> --scheme <scheme> --label <label> [flags]
```

**Required Flags:**
- `--disk <diskN>`: Whole disk identifier to format.
- `--fs <exfat|fat32|apfs|hfs+>`: Target filesystem.
- `--scheme <mbr|gpt|apm>`: Partition layout scheme.
- `--label <LABEL>`: Volume label name.

**Useful Flags:**
- `--case-sensitive`: Uses case-sensitive volume format variants (if supported).
- `--force-unmount`: Forcefully unmounts volumes before formatting.
- `--mount-after` / `--no-mount-after`: Controls disk mounting behavior post-format.

---

### `usbfix wipe`
Wipes partition structures, layouts, or overwrites the primary partition sector to prepare a problematic disk for clean reformatting.

```bash
usbfix wipe --disk diskN <mode-flag> [flags]
```

**Mode Flags (At least one is required):**
- `--quick`: Quickly removes partition headers/metadata with `diskutil`.
- `--scheme <mbr|gpt|apm>`: Scheme utilized for quick wiping.
- `--zero-first-mb`: Overwrites the initial megabyte of the drive with zeros using raw disk streams.

**Useful Flags:**
- `--force-unmount`: Force unmounts all drive volumes before wiping.

---

### `usbfix version`
Prints the binary version, git commit, build date, and target architecture.

```bash
usbfix version
```

---

## Global Flags

Global flags can be placed before or after any command to modify overall CLI behavior.

```bash
usbfix --output json list
usbfix --dry-run repair --volume disk4s1 --repair
```

### General & Output Flags
- `--output <table|json|plain>`: Selects console format for command results.
- `--quiet`: Silences all diagnostics and non-essential progress lines from standard error.
- `--verbose`: Prints highly detailed diagnostic logs.
- `--no-color`: Disables colored outputs in terminals.
- `--log-file <path>`: Simultaneously redirects all background execution logs to a file.
- `--interactive`: Prompts user interactively for any missing parameters or approvals (default).
- `--non-interactive`: Strictly fails execution if a value is missing or confirmation is required without prompting.

### Safety Override Flags
- `--dry-run`: Performs a full dry-run simulation of the subcommand.
- `--confirm <phrase>`: Automates verification bypass by providing the confirmation string `"ERASE diskN"`.
- `--yes --i-know-this-erases-data`: Explicit command-line confirmation bypass.
- `--allow-internal`: Explicitly permits destructive commands to proceed on macOS internal/system disks.
- `--allow-non-removable`: Allows actions on fixed external/internal hardware.
- `--require-removable`: Restricts operations strictly to drives identifying as removable media.
- `--whole-disk-only`: Rejects any volume partition target identifiers.
- `--snapshot-info <path>`: Backs up a snapshot of disk metadata to a file prior to execution.

---

## Output Formats

`usbfix` is built for both humans and scripts.

1. **Table (Default)**: Best for terminal readability.
2. **JSON**: Stable and easily parser-friendly, ideal for scripts:
   ```bash
   usbfix --output json inspect --disk disk4
   ```
3. **Plain**: Emits simple space-delimited text.

---

## Exit Codes

| Code | Meaning |
| :---: | :--- |
| **`0`** | Success |
| **`1`** | General runtime error |
| **`2`** | Usage error (invalid arguments or flags) |
| **`3`** | Operation cancelled by user |
| **`126`** | Permission denied (run with elevated privileges if needed) |
| **`127`** | Required system utility not found (e.g. `diskutil`) |
| **`130`** | Execution interrupted with `Ctrl-C` |

---

## Project Status & Scope

This project focuses on macOS-specific external and USB disk management. It does not perform digital forensics, deleted file recovery, or reverse-engineering. Its primary purpose is to safely format, clean, repair, and diagnose USB flash drives and portable storage devices from the comfort of your terminal.
