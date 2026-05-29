package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/santgabo/usbfixertool-cli/internal/usbfix"
)

func executeRoot(t *testing.T, runner *fakeRunner, stdin string, args ...string) (string, string, error) {
	t.Helper()

	cmd := NewRootCommand(Dependencies{Runner: runner, Stdin: strings.NewReader(stdin)})
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

func TestRootRejectsConflictingInteractivityFlags(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{}
	_, _, err := executeRoot(t, runner, "", "--interactive", "--non-interactive", "list")
	if err == nil {
		t.Fatal("expected conflicting flags error")
	}
	if code := usbfix.ExitCode(err); code != usbfix.ExitUsage {
		t.Fatalf("exit code = %d, want %d; err = %v", code, usbfix.ExitUsage, err)
	}
	if len(runner.calls) != 0 {
		t.Fatalf("expected no diskutil calls, got %#v", runner.calls)
	}
}

func TestListDefaultFiltersInternalAndNonPortableDisks(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{responses: map[string]string{
		"diskutil list":            diskutilListFixture(),
		"diskutil info /dev/disk0": internalDiskInfo(),
		"diskutil info /dev/disk4": externalDiskInfo(),
		"diskutil info /dev/disk5": nonPortableDiskInfo(),
	}}

	stdout, stderr, err := executeRoot(t, runner, "", "--output", "plain", "list")
	if err != nil {
		t.Fatalf("Execute returned error: %v\nstderr:\n%s", err, stderr)
	}
	if !strings.Contains(stdout, "disk4\tExternal\t31.0 GB\tUSBBOOT") {
		t.Fatalf("expected external disk with volume label in output:\n%s", stdout)
	}
	if strings.Contains(stdout, "disk0") || strings.Contains(stdout, "disk5") {
		t.Fatalf("expected internal and non-portable disks to be filtered:\n%s", stdout)
	}
}

func TestListTableIncludesVolumeLabel(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{responses: map[string]string{
		"diskutil list":            diskutilListFixture(),
		"diskutil info /dev/disk4": externalDiskInfo(),
	}}

	stdout, stderr, err := executeRoot(t, runner, "", "list")
	if err != nil {
		t.Fatalf("Execute returned error: %v\nstderr:\n%s", err, stderr)
	}
	for _, want := range []string{"LABEL", "USBBOOT", "IDENTIFIER", "NAME"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected %q in table output:\n%s", want, stdout)
		}
	}
}

func TestListJSONIncludesVolumeLabel(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{responses: map[string]string{
		"diskutil list":            diskutilListFixture(),
		"diskutil info /dev/disk4": externalDiskInfo(),
	}}

	stdout, stderr, err := executeRoot(t, runner, "", "--output", "json", "list")
	if err != nil {
		t.Fatalf("Execute returned error: %v\nstderr:\n%s", err, stderr)
	}
	var disks []usbfix.DiskEntry
	if err := json.Unmarshal([]byte(stdout), &disks); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, stdout)
	}
	if len(disks) != 1 {
		t.Fatalf("got %d disks, want 1: %#v", len(disks), disks)
	}
	if disks[0].VolumeLabel != "USBBOOT" {
		t.Fatalf("volumeLabel = %q, want USBBOOT", disks[0].VolumeLabel)
	}
}

func TestListAllIncludesInternalDisks(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{responses: map[string]string{
		"diskutil list":            diskutilListFixture(),
		"diskutil info /dev/disk0": internalDiskInfo(),
		"diskutil info /dev/disk4": externalDiskInfo(),
		"diskutil info /dev/disk5": nonPortableDiskInfo(),
	}}

	stdout, stderr, err := executeRoot(t, runner, "", "--output", "plain", "list", "--all")
	if err != nil {
		t.Fatalf("Execute returned error: %v\nstderr:\n%s", err, stderr)
	}
	for _, want := range []string{"disk0\tInternal", "disk4\tExternal", "disk5\t"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected %q in output:\n%s", want, stdout)
		}
	}
}

func TestInspectRequiresExactlyOneTarget(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{}
	_, _, err := executeRoot(t, runner, "", "inspect")
	if err == nil {
		t.Fatal("expected target validation error")
	}
	if code := usbfix.ExitCode(err); code != usbfix.ExitUsage {
		t.Fatalf("exit code = %d, want %d; err = %v", code, usbfix.ExitUsage, err)
	}
	if len(runner.calls) != 0 {
		t.Fatalf("expected no diskutil info calls, got %#v", runner.calls)
	}
}

func TestInspectJSONOmitsRawInfoByDefault(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{responses: map[string]string{
		"diskutil info /dev/disk4": externalDiskInfo(),
	}}

	stdout, stderr, err := executeRoot(t, runner, "", "--output", "json", "inspect", "--disk", "disk4")
	if err != nil {
		t.Fatalf("Execute returned error: %v\nstderr:\n%s", err, stderr)
	}
	var info usbfix.DeviceInfo
	if err := json.Unmarshal([]byte(stdout), &info); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, stdout)
	}
	if info.Identifier != "disk4" || !info.External || !info.Removable {
		t.Fatalf("unexpected inspected info: %#v", info)
	}
	if info.Raw != "" {
		t.Fatalf("raw info should be omitted by default, got %q", info.Raw)
	}
}

func TestRepairDryRunPlansVolumeFsckWithoutRunningIt(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{}
	stdout, stderr, err := executeRoot(t, runner, "", "--dry-run", "repair", "--volume", "disk4s1", "--repair", "--fsck", "--mount-after")
	if err != nil {
		t.Fatalf("Execute returned error: %v\nstderr:\n%s", err, stderr)
	}
	for _, want := range []string{
		"DRY-RUN: diskutil repairVolume /dev/disk4s1",
		"DRY-RUN: fsck_* /dev/disk4s1",
		"DRY-RUN: diskutil mount /dev/disk4s1",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected %q in output:\n%s", want, stdout)
		}
	}
	if len(runner.calls) != 0 {
		t.Fatalf("dry-run should not execute repair commands, calls = %#v", runner.calls)
	}
}

func TestRepairVolumeWithFsckExecutesExpectedCommands(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{responses: map[string]string{
		"diskutil repairVolume /dev/disk4s1": "volume repaired\n",
		"diskutil info /dev/disk4s1":         exfatVolumeInfo(),
		"fsck_exfat -y /dev/disk4s1":         "fsck repaired\n",
	}}

	stdout, stderr, err := executeRoot(t, runner, "", "repair", "--volume", "disk4s1", "--repair", "--fsck")
	if err != nil {
		t.Fatalf("Execute returned error: %v\nstderr:\n%s", err, stderr)
	}
	if !strings.Contains(stdout, "volume repaired\nfsck repaired\n") {
		t.Fatalf("unexpected repair output:\n%s", stdout)
	}
	wantCalls := []string{
		"diskutil repairVolume /dev/disk4s1",
		"diskutil info /dev/disk4s1",
		"fsck_exfat -y /dev/disk4s1",
	}
	if strings.Join(runner.calls, "|") != strings.Join(wantCalls, "|") {
		t.Fatalf("calls = %#v, want %#v", runner.calls, wantCalls)
	}
}

func TestWipeDryRunPlansQuickAndZeroFirstMB(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{responses: map[string]string{
		"diskutil info /dev/disk4": externalDiskInfo(),
	}}

	stdout, stderr, err := executeRoot(t, runner, "", "--dry-run", "wipe", "--disk", "disk4", "--quick", "--scheme", "mbr", "--zero-first-mb", "--force-unmount")
	if err != nil {
		t.Fatalf("Execute returned error: %v\nstderr:\n%s", err, stderr)
	}
	for _, want := range []string{
		"DRY-RUN: diskutil unmountDisk force /dev/disk4",
		"DRY-RUN: dd if=/dev/zero of=/dev/rdisk4 bs=1m count=1",
		"DRY-RUN: diskutil eraseDisk free EMPTY MBR /dev/disk4",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected %q in output:\n%s", want, stdout)
		}
	}
	for _, call := range runner.calls {
		if strings.Contains(call, "eraseDisk") || strings.HasPrefix(call, "dd ") {
			t.Fatalf("dry-run should not execute destructive calls, calls = %#v", runner.calls)
		}
	}
}

func TestFormatRefusesInternalDiskBeforeConfirmation(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{responses: map[string]string{
		"diskutil info /dev/disk0": internalDiskInfo(),
	}}

	_, _, err := executeRoot(t, runner, "", "format", "--disk", "disk0", "--fs", "exfat", "--scheme", "mbr", "--label", "USBBOOT", "--confirm", "ERASE disk0")
	if err == nil {
		t.Fatal("expected internal disk safety error")
	}
	if code := usbfix.ExitCode(err); code != usbfix.ExitRuntime {
		t.Fatalf("exit code = %d, want %d; err = %v", code, usbfix.ExitRuntime, err)
	}
	for _, call := range runner.calls {
		if strings.Contains(call, "eraseDisk") {
			t.Fatalf("internal disk should not be erased, calls = %#v", runner.calls)
		}
	}
}

func diskutilListFixture() string {
	return `/dev/disk0 (internal, physical):
   #:                       TYPE NAME                    SIZE       IDENTIFIER
   0:      GUID_partition_scheme                        *500.0 GB   disk0
   1:                        EFI EFI                     209.7 MB   disk0s1
/dev/disk4 (external, physical):
   #:                       TYPE NAME                    SIZE       IDENTIFIER
   0:      FDisk_partition_scheme                        *31.0 GB   disk4
   1:             Windows_FAT_32 USBBOOT                 31.0 GB    disk4s1
/dev/disk5 (disk image):
   #:                       TYPE NAME                    SIZE       IDENTIFIER
   0:      GUID_partition_scheme                        *4.0 GB     disk5
`
}

func internalDiskInfo() string {
	return `   Device Identifier:         disk0
   Device Node:               /dev/disk0
   Whole:                     Yes
   Part of Whole:             disk0
   Device / Media Name:       APPLE SSD
   Disk Size:                 500.0 GB (500000000000 Bytes)
   Device Location:           Internal
   Removable Media:           Fixed
   Internal:                  Yes
`
}

func nonPortableDiskInfo() string {
	return `   Device Identifier:         disk5
   Device Node:               /dev/disk5
   Whole:                     Yes
   Part of Whole:             disk5
   Device / Media Name:       Disk Image
   Disk Size:                 4.0 GB (4000000000 Bytes)
   Removable Media:           Fixed
   Internal:                  No
`
}

func exfatVolumeInfo() string {
	return `   Device Identifier:         disk4s1
   Device Node:               /dev/disk4s1
   Whole:                     No
   Part of Whole:             disk4
   Volume Name:               USBBOOT
   File System Personality:   ExFAT
   Mount Point:               /Volumes/USBBOOT
   Mounted:                   Yes
   Device Location:           External
   Removable Media:           Removable
   Internal:                  No
`
}
