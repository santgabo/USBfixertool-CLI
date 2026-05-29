---
name: usbfix-cli-safety
description: "Safely operate USB Fixer CLI (usbfix) on macOS disks. Use whenever an AI agent needs to list, inspect, verify, repair, format, or wipe disks with usbfix — including automation, scripting, or troubleshooting USB/external drives. Always use this skill before running any usbfix command that could modify disks. Critical for internal disk operations: the agent must obtain explicit user confirmation in chat before passing --allow-internal or running destructive commands against internal disks. Also triggers for diskutil-adjacent workflows, 'format my USB', 'wipe disk', 'repair external drive', or any request involving disk4/diskN identifiers."
---

# USB Fixer CLI — Safe Agent Operations

You are helping a user manage macOS disks with `usbfix`. Disk operations can destroy data permanently. Your job is to be cautious, transparent, and never bypass safety checks on the user's behalf.

## Core Principles

1. **Identify before acting** — never guess which disk is the target.
2. **Preview before modifying** — use `--dry-run` for any destructive command.
3. **Confirm before erasing** — destructive commands require explicit confirmation.
4. **Protect internal disks** — internal disks need explicit user consent in chat before any modifying operation.
5. **Prefer read-only first** — `list`, `inspect`, and `repair --verify-only` are safe starting points.

## Command Safety Tiers

| Tier | Commands | Agent behavior |
|------|----------|----------------|
| Read-only | `list`, `inspect`, `version` | Safe to run without extra confirmation |
| Verify-only | `repair --verify-only` | Safe; no data changes |
| Modifying | `repair --repair`, `format`, `wipe` | Requires full safety workflow (below) |

## Mandatory Workflow for Modifying Operations

Follow these steps in order. Do not skip steps.

### Step 1: List candidate disks

```bash
usbfix list
# or for JSON parsing:
usbfix --output json list
```

If the user mentions a specific disk identifier, still verify it exists and matches their intent.

Use `usbfix list --all` only when you need to see internal disks — and treat anything internal as high-risk.

### Step 2: Inspect the target

```bash
usbfix inspect --disk diskN
# or for a volume:
usbfix inspect --volume diskNsM
```

Check these fields in the output:

- **Internal**: `true` → STOP. Do not proceed without explicit user consent (see Internal Disk Rules).
- **Removable Media**: prefer `true` for USB operations.
- **Device Location**: `External` vs `Internal`.
- **Whole disk**: required for `format` and `wipe` (use `diskN`, not `diskNsM`).

Present a brief summary to the user: disk ID, size, location, internal/removable status, and current filesystem.

### Step 3: Dry-run preview

For `format`, `wipe`, or `repair --repair`:

```bash
usbfix --dry-run <command> [flags...]
```

Show the user what would happen. Wait for approval before proceeding.

### Step 4: Execute with explicit confirmation

Destructive commands (`format`, `wipe`) require the confirmation phrase:

```text
ERASE diskN
```

Pass it non-interactively:

```bash
usbfix format --disk disk4 --fs exfat --scheme mbr --label MYUSB --confirm "ERASE disk4"
usbfix wipe --disk disk4 --quick --scheme gpt --confirm "ERASE disk4"
```

**Never** use `--yes --i-know-this-erases-data` unless the user explicitly asked you to skip confirmation prompts.

## Internal Disk Rules (Critical)

`usbfix` blocks internal disks by default. The CLI will refuse with:

```text
refusing to operate on internal disk disk0; pass --allow-internal only if you are certain
```

When the target is internal (`Internal: true`):

1. **Stop immediately** — do not run modifying commands.
2. **Explain the risk clearly** to the user: internal disks hold the OS, apps, and personal data.
3. **Ask for explicit consent in chat**, for example:
   > "This operation targets disk0, which is your internal drive. This could destroy your macOS installation and all data. Do you explicitly authorize operating on this internal disk?"
4. **Wait for an unambiguous yes** before proceeding.
5. Only after consent, add `--allow-internal` and the required `--confirm "ERASE diskN"` phrase.
6. Even with consent, still run `--dry-run` first and show the plan.

**Never** pass `--allow-internal` proactively. **Never** assume the user meant an internal disk when they said "format my drive."

## External / USB Operations

For external or removable disks, the standard workflow applies:

```bash
# 1. Identify
usbfix list

# 2. Inspect
usbfix inspect --disk disk4

# 3. Preview
usbfix --dry-run format --disk disk4 --fs exfat --scheme mbr --label USBBOOT

# 4. Execute (after user approval)
usbfix format --disk disk4 --fs exfat --scheme mbr --label USBBOOT --confirm "ERASE disk4"
```

## Repair Workflow

Repairs are lower risk than format/wipe but can still modify data:

```bash
# Always verify first
usbfix repair --volume disk4s1 --verify-only

# Only repair after showing verify results and getting approval
usbfix --dry-run repair --volume disk4s1 --repair
usbfix repair --volume disk4s1 --repair
```

For whole-disk repair on internal disks, apply the Internal Disk Rules.

## Flags Reference

### Safe global flags

| Flag | Purpose |
|------|---------|
| `--dry-run` | Preview without changes |
| `--output json` | Machine-readable output for parsing |
| `--non-interactive` | Fail instead of prompting (use with `--confirm`) |

### Destructive / bypass flags (require user approval)

| Flag | Risk | Agent rule |
|------|------|------------|
| `--confirm "ERASE diskN"` | Required for format/wipe | Use exact phrase after user approval |
| `--allow-internal` | Permits internal disk ops | Only after explicit chat consent |
| `--allow-non-removable` | Permits fixed disks | Only after user confirms intent |
| `--yes --i-know-this-erases-data` | Skips confirmation | Avoid unless user explicitly requests |
| `--force-unmount` | Force unmount | Explain why needed; use with care |

## What NOT to Do

- Do not run `format` or `wipe` without prior `inspect`.
- Do not pass `--allow-internal` to "make it work" without asking the user.
- Do not use `diskutil eraseDisk` directly — use `usbfix` so safety checks apply.
- Do not target `diskNsM` (partitions) for `format` or `wipe` — whole disk only (`diskN`).
- Do not chain destructive commands without re-inspecting if the disk layout may have changed.
- Do not assume `disk4` is the USB — always verify via `list` + `inspect`.

## Interpreting Exit Codes

| Code | Meaning | Agent action |
|------|---------|--------------|
| 0 | Success | Report result |
| 2 | Usage error | Fix flags/args, do not retry blindly |
| 3 | Cancelled | User or confirmation rejected — stop |
| 126 | Permission denied | Tell user elevated permissions may be needed |
| 130 | Interrupted | Operation was cancelled mid-run |

## Reporting Results to the User

After any operation, summarize:

1. Which disk was targeted and its properties (external/internal, size).
2. What command ran (or would run in dry-run).
3. Outcome (success, error, or preview).
4. Next recommended step, if any.

For JSON output, parse stdout (not stderr) for structured results.

## Example: User asks to format a USB drive

```
User: "Format my USB as exFAT"

Agent:
1. usbfix list                          → find external/removable disk
2. usbfix inspect --disk disk4          → confirm external, show summary
3. Ask user to confirm disk4 is correct
4. usbfix --dry-run format ...          → show plan
5. Ask user to approve
6. usbfix format ... --confirm "ERASE disk4"
```

## Example: User asks to wipe internal disk (must refuse without consent)

```
User: "Wipe disk0"

Agent:
1. usbfix inspect --disk disk0          → Internal: true
2. Explain: disk0 is the internal drive; wiping destroys macOS and all data
3. Ask: "Do you explicitly authorize wiping your internal disk disk0?"
4. If NO → stop
5. If YES → dry-run first, then wipe with --allow-internal --confirm "ERASE disk0"
```
