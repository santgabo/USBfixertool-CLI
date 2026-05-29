---
name: usbfix-cli
description: "Operate USB Fixer CLI (usbfix) on macOS — list, inspect, verify, repair, format, and wipe USB drives or external disks. Use whenever an AI agent needs to help with usbfix commands, disk troubleshooting, USB formatting, partition repair, diskutil-adjacent workflows, or scripting usbfix with JSON output. Always use this skill before running any usbfix command. Includes safety rules for destructive operations and internal disks. Triggers for 'format my USB', 'repair external drive', 'wipe disk', 'list disks', disk4/diskN identifiers, or any usbfix automation."
---

# USB Fixer CLI (`usbfix`)

`usbfix` is a macOS CLI for inspecting, verifying, repairing, formatting, and wiping USB drives and external disks. It wraps `diskutil` and filesystem tools with safety checks and structured output.

**Platform:** macOS only. Requires `diskutil` (pre-installed).

## Quick Reference

| Command | Purpose | Modifies data? |
|---------|---------|----------------|
| `list` | Show candidate disks | No |
| `inspect` | Detailed disk/volume metadata | No |
| `repair` | Verify or repair filesystem | Only with `--repair` |
| `format` | Erase and format whole disk | Yes (destructive) |
| `wipe` | Remove partition layout / zero boot sector | Yes (destructive) |
| `version` | Print binary version info | No |

## Standard Workflows

### Identify a USB drive

```bash
usbfix list
usbfix list --removable-only
usbfix list --external-only
usbfix --output json list
```

Use `--all` only when internal disks must be visible (high risk).

### Inspect before acting

```bash
usbfix inspect --disk diskN
usbfix inspect --volume diskNsM
usbfix inspect --disk diskN --show-raw
```

Key fields: **Internal**, **Removable Media**, **Device Location**, filesystem, size.

### Verify disk health (safe)

```bash
usbfix repair --volume diskNsM --verify-only
usbfix repair --disk diskN --verify-only
```

### Repair a volume

```bash
usbfix repair --volume diskNsM --verify-only   # always verify first
usbfix --dry-run repair --volume diskNsM --repair
usbfix repair --volume diskNsM --repair
usbfix repair --volume diskNsM --repair --fsck  # run fsck helpers when needed
```

### Format a whole USB drive

Requires whole disk (`diskN`), not a partition (`diskNsM`).

```bash
usbfix --dry-run format --disk disk4 --fs exfat --scheme mbr --label USBBOOT
usbfix format --disk disk4 --fs exfat --scheme mbr --label USBBOOT --confirm "ERASE disk4"
```

**Required flags:** `--disk`, `--fs`, `--scheme`, `--label`.

**Filesystems:** `exfat`, `fat32`, `apfs`, `hfs+`
**Schemes:** `mbr`, `gpt`, `apm`

### Wipe partition metadata

```bash
usbfix --dry-run wipe --disk disk4 --quick --scheme gpt
usbfix wipe --disk disk4 --quick --scheme gpt --confirm "ERASE disk4"

usbfix wipe --disk disk4 --zero-first-mb --confirm "ERASE disk4"
```

At least one mode flag is required: `--quick --scheme <scheme>` or `--zero-first-mb`.

## Global Flags

Place before or after any subcommand:

```bash
usbfix --output json list
usbfix --dry-run format --disk disk4 --fs exfat --scheme mbr --label USB
```

| Flag | Purpose |
|------|---------|
| `--output table\|json\|plain` | Output format (default: `table`) |
| `--quiet` | Suppress stderr diagnostics |
| `--verbose` | Verbose diagnostics |
| `--no-color` | Disable colors |
| `--log-file <path>` | Write diagnostics to file |
| `--interactive` | Prompt for missing values (default) |
| `--non-interactive` | Fail if a value or confirmation is missing |
| `--dry-run` | Preview commands without executing |
| `--confirm "ERASE diskN"` | Bypass interactive confirmation for format/wipe |
| `--allow-internal` | Permit operations on internal disks |
| `--allow-non-removable` | Permit fixed/non-removable disks |
| `--require-removable` | Restrict to removable media only |
| `--whole-disk-only` | Reject volume/partition targets |
| `--snapshot-info <path>` | Save disk metadata snapshot before execution |
| `--force-unmount` | Force unmount before modifying |
| `--mount-after` / `--no-mount-after` | Control remount after operation |

## Command Details

### `list`

```bash
usbfix list [--all] [--removable-only] [--external-only]
```

Default: hides internal disks, shows external/removable candidates.

### `inspect`

Exactly one target required:

```bash
usbfix inspect --disk diskN
usbfix inspect --volume diskNsM
```

### `repair`

Exactly one target required:

```bash
usbfix repair --disk diskN [flags]
usbfix repair --volume diskNsM [flags]
```

| Flag | Purpose |
|------|---------|
| `--verify-only` | Check only, no writes (default behavior) |
| `--repair` | Apply repairs |
| `--fsck` | Run filesystem-specific repair tools |
| `--retry <N>` | Retry unmount/repair steps |
| `--timeout <duration>` | Max operation time (e.g. `60s`) |

### `format`

Whole disk only. All of `--disk`, `--fs`, `--scheme`, `--label` are required.

```bash
usbfix format --disk diskN --fs exfat --scheme mbr --label MYUSB [flags]
```

Optional: `--case-sensitive`, `--force-unmount`, `--mount-after`, `--no-mount-after`.

### `wipe`

Whole disk only.

```bash
usbfix wipe --disk diskN --quick --scheme gpt
usbfix wipe --disk diskN --zero-first-mb
```

### `version`

```bash
usbfix version
```

## Output Formats

- **table** — human-readable terminal output (default)
- **json** — structured stdout for scripts; parse stdout, not stderr
- **plain** — simple space-delimited text

```bash
usbfix --output json inspect --disk disk4
usbfix --output json list --removable-only
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General runtime error |
| 2 | Usage error (invalid flags/args) |
| 3 | Operation cancelled |
| 126 | Permission denied |
| 127 | Required utility not found |
| 130 | Interrupted (Ctrl-C) |

## Safety Rules

Disk operations can destroy data permanently. Apply these rules for **all modifying operations** (`repair --repair`, `format`, `wipe`).

### Core principles

1. **Identify before acting** — run `list` + `inspect`; never guess disk IDs.
2. **Preview before modifying** — use `--dry-run` for destructive commands.
3. **Confirm before erasing** — format/wipe require `--confirm "ERASE diskN"`.
4. **Prefer read-only first** — start with `list`, `inspect`, `repair --verify-only`.
5. **Use usbfix, not raw diskutil** — safety checks apply only through `usbfix`.

### Mandatory workflow for modifying operations

1. `usbfix list` — find the target
2. `usbfix inspect --disk diskN` — verify external/removable, show summary to user
3. Ask user to confirm the correct disk
4. `usbfix --dry-run <command> ...` — show plan
5. Execute with `--confirm "ERASE diskN"` after approval

### Internal disk rules (critical)

`usbfix` blocks internal disks by default:

```text
refusing to operate on internal disk disk0; pass --allow-internal only if you are certain
```

When `Internal: true`:

1. **Stop** — do not run modifying commands.
2. **Explain the risk** — internal disks hold macOS and all user data.
3. **Ask for explicit consent in chat** before proceeding.
4. Only after consent: `--dry-run` first, then `--allow-internal --confirm "ERASE diskN"`.

**Never** pass `--allow-internal` proactively.

### Destructive flag rules

| Flag | Agent rule |
|------|------------|
| `--confirm "ERASE diskN"` | Required for format/wipe; use exact phrase |
| `--allow-internal` | Only after explicit chat consent |
| `--allow-non-removable` | Only after user confirms intent |
| `--yes --i-know-this-erases-data` | Avoid unless user explicitly requests |
| `--force-unmount` | Explain why; use with care |

### What NOT to do

- Do not `format` or `wipe` without prior `inspect`.
- Do not target partitions (`diskNsM`) for `format` or `wipe` — whole disk only.
- Do not assume `disk4` is the USB — always verify via `list` + `inspect`.
- Do not chain destructive commands without re-inspecting if layout may have changed.

## Reporting Results

After any operation, summarize:

1. Target disk and properties (external/internal, size, filesystem).
2. Command executed (or dry-run preview).
3. Outcome (success, error, or cancelled).
4. Recommended next step, if any.

## Examples

### User: "Format my USB as exFAT"

```
1. usbfix list
2. usbfix inspect --disk disk4          → confirm external/removable
3. Ask user to confirm disk4
4. usbfix --dry-run format --disk disk4 --fs exfat --scheme mbr --label USBBOOT
5. usbfix format ... --confirm "ERASE disk4"
```

### User: "Check if my USB is healthy"

```
1. usbfix list
2. usbfix inspect --volume disk4s1
3. usbfix repair --volume disk4s1 --verify-only
```

### User: "Wipe disk0"

```
1. usbfix inspect --disk disk0          → Internal: true
2. Warn: internal drive; ask explicit authorization
3. If approved: dry-run, then wipe with --allow-internal --confirm "ERASE disk0"
```

### Scripting with JSON

```bash
usbfix --output json --non-interactive list --removable-only
usbfix --output json inspect --disk disk4
```

Use `--non-interactive` with `--confirm` for automation; never automate internal disk operations without explicit user authorization.
