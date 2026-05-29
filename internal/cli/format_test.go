package cli

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/santgabo/usbfixertool-cli/internal/usbfix"
)

type fakeRunner struct {
	responses map[string]string
	calls     []string
}

func (f *fakeRunner) LookPath(file string) (string, error) {
	if file == "diskutil" {
		return "/usr/sbin/diskutil", nil
	}
	if file == "dd" {
		return "/bin/dd", nil
	}
	return "", fmt.Errorf("unexpected LookPath(%q)", file)
}

func (f *fakeRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	command := usbfix.CommandString(name, args...)
	f.calls = append(f.calls, command)
	if response, ok := f.responses[command]; ok {
		return []byte(response), nil
	}
	return nil, fmt.Errorf("unexpected command: %s", command)
}

func TestFormatRejectsPartitionIdentifier(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{}
	cmd := NewRootCommand(Dependencies{Runner: runner, Stdin: strings.NewReader("")})
	var out, stderr bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{
		"format",
		"--disk", "disk4s1",
		"--fs", "exfat",
		"--scheme", "mbr",
		"--label", "USBBOOT",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if code := usbfix.ExitCode(err); code != usbfix.ExitUsage {
		t.Fatalf("exit code = %d, want %d; err = %v", code, usbfix.ExitUsage, err)
	}
	if len(runner.calls) != 0 {
		t.Fatalf("expected no diskutil calls, got %#v", runner.calls)
	}
}

func TestFormatDryRunDoesNotEraseDisk(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{responses: map[string]string{
		"diskutil info /dev/disk4": externalDiskInfo(),
	}}
	cmd := NewRootCommand(Dependencies{Runner: runner, Stdin: strings.NewReader("")})
	var out, stderr bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{
		"--dry-run",
		"--output", "json",
		"format",
		"--disk", "disk4",
		"--fs", "exfat",
		"--scheme", "mbr",
		"--label", "USBBOOT",
		"--force-unmount",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v\nstderr:\n%s", err, stderr.String())
	}
	if !strings.Contains(out.String(), `"dryRun": true`) {
		t.Fatalf("dry-run JSON missing from output:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "diskutil eraseDisk ExFAT USBBOOT MBR /dev/disk4") {
		t.Fatalf("planned erase command missing from output:\n%s", out.String())
	}
	for _, call := range runner.calls {
		if strings.Contains(call, "eraseDisk") {
			t.Fatalf("dry run should not execute eraseDisk, calls = %#v", runner.calls)
		}
	}
}

func TestFormatRequiresConfirmation(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{responses: map[string]string{
		"diskutil info /dev/disk4": externalDiskInfo(),
	}}
	cmd := NewRootCommand(Dependencies{Runner: runner, Stdin: strings.NewReader("")})
	var out, stderr bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{
		"format",
		"--disk", "disk4",
		"--fs", "exfat",
		"--scheme", "mbr",
		"--label", "USBBOOT",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirmation error")
	}
	if code := usbfix.ExitCode(err); code != usbfix.ExitCancelled {
		t.Fatalf("exit code = %d, want %d; err = %v", code, usbfix.ExitCancelled, err)
	}
	for _, call := range runner.calls {
		if strings.Contains(call, "eraseDisk") {
			t.Fatalf("missing confirmation should not execute eraseDisk, calls = %#v", runner.calls)
		}
	}
}

func externalDiskInfo() string {
	return `   Device Identifier:         disk4
   Device Node:               /dev/disk4
   Whole:                     Yes
   Part of Whole:             disk4
   Device / Media Name:       USB Flash Disk
   Disk Size:                 31.0 GB (31000000000 Bytes)
   Device Location:           External
   Removable Media:           Removable
   Internal:                  No
`
}
