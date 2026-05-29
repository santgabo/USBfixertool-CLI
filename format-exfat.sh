#!/usr/bin/env bash
#
# format-exfat.sh
# Reformats an external disk as exFAT (MBR scheme) on macOS.
# Includes a safety check that aborts if the disk is internal.
#
# Usage:
#   ./format-exfat.sh <disk_identifier> [volume_label]
# Example:
#   ./format-exfat.sh disk4 USBBOOT

set -euo pipefail

usage() {
    echo "Usage: $0 <disk_identifier> [volume_label]"
    echo "Example: $0 disk4 USBBOOT"
    exit 1
}

# --- Argument validation --------------------------------------------------
if [[ $# -lt 1 || $# -gt 2 ]]; then
    usage
fi

RAW_INPUT="$1"
LABEL="${2:-EXFAT}"

# Normalize the identifier: accepts "disk4" or "/dev/disk4"
DISK_ID="${RAW_INPUT#/dev/}"
DISK_DEV="/dev/${DISK_ID}"

# The identifier must be of type diskN (without slices: diskNsM is not valid here)
if [[ ! "$DISK_ID" =~ ^disk[0-9]+$ ]]; then
    echo "ERROR: Invalid identifier '$RAW_INPUT'. Must be of type diskN (e.g. disk4)."
    exit 1
fi

# exFAT label allows up to 11 characters in uppercase/numbers/_-
if [[ ${#LABEL} -gt 11 ]]; then
    echo "ERROR: Label '$LABEL' exceeds 11 characters."
    exit 1
fi

# --- Verify that the disk exists ------------------------------------------
if ! diskutil info "$DISK_DEV" >/dev/null 2>&1; then
    echo "ERROR: Disk '$DISK_DEV' not found."
    exit 1
fi

# --- SAFETY: abort if the disk is internal -------------------------------
INFO="$(diskutil info "$DISK_DEV")"

DEVICE_LOCATION="$(echo "$INFO" | awk -F: '/Device Location/ {gsub(/^[ \t]+/, "", $2); print $2}')"
INTERNAL_FLAG="$(echo  "$INFO" | awk -F: '/^[ \t]*Internal:/ {gsub(/^[ \t]+/, "", $2); print $2}')"
REMOVABLE_FLAG="$(echo "$INFO" | awk -F: '/Removable Media/ {gsub(/^[ \t]+/, "", $2); print $2}')"
WHOLE_FLAG="$(echo     "$INFO" | awk -F: '/^[ \t]*Whole:/ {gsub(/^[ \t]+/, "", $2); print $2}')"

echo "------------------------------------------------------------"
echo "Target disk    : $DISK_DEV"
echo "Whole disk     : $WHOLE_FLAG"
echo "Device Location: $DEVICE_LOCATION"
echo "Internal       : $INTERNAL_FLAG"
echo "Removable Media: $REMOVABLE_FLAG"
echo "New label      : $LABEL"
echo "------------------------------------------------------------"

if [[ "$WHOLE_FLAG" != "Yes" ]]; then
    echo "ERROR: '$DISK_DEV' is not a whole disk. Point to diskN, not a partition."
    exit 1
fi

if [[ "$INTERNAL_FLAG" == "Yes" || "$DEVICE_LOCATION" == "Internal" ]]; then
    echo "ABORTED: '$DISK_DEV' is an INTERNAL disk. Will not format for safety reasons."
    exit 2
fi

# --- Explicit user confirmation -------------------------------------------
echo
echo "This operation will ERASE ALL content on $DISK_DEV."
read -r -p "Type exactly 'ERASE' to continue: " CONFIRM
if [[ "$CONFIRM" != "ERASE" ]]; then
    echo "Operation cancelled by user."
    exit 3
fi

# --- Unmount and format ---------------------------------------------------
echo "Unmounting $DISK_DEV ..."
diskutil unmountDisk force "$DISK_DEV"

echo "Reformatting $DISK_DEV as exFAT (MBR) with label '$LABEL' ..."
diskutil eraseDisk ExFAT "$LABEL" MBR "$DISK_DEV"

echo
echo "Result:"
diskutil list "$DISK_DEV"

echo "OK: $DISK_DEV has been formatted as exFAT."
