package usbfix

import "testing"

func TestNormalizeWholeDiskIdentifier(t *testing.T) {
	t.Parallel()

	got, err := NormalizeWholeDiskIdentifier("/dev/disk4")
	if err != nil {
		t.Fatalf("NormalizeWholeDiskIdentifier returned error: %v", err)
	}
	if got != "disk4" {
		t.Fatalf("got %q, want disk4", got)
	}
}

func TestNormalizeWholeDiskIdentifierRejectsPartition(t *testing.T) {
	t.Parallel()

	if _, err := NormalizeWholeDiskIdentifier("disk4s1"); err == nil {
		t.Fatal("expected partition identifier to be rejected")
	}
}

func TestNormalizeVolumeIdentifier(t *testing.T) {
	t.Parallel()

	got, err := NormalizeVolumeIdentifier("/dev/disk4s1")
	if err != nil {
		t.Fatalf("NormalizeVolumeIdentifier returned error: %v", err)
	}
	if got != "disk4s1" {
		t.Fatalf("got %q, want disk4s1", got)
	}
}

func TestNormalizeIdentifierRejectsEmptyAndUnexpectedNames(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{"", "   ", "/dev/rdisk4", "disk", "diskx", "disk4p1"} {
		if _, err := NormalizeIdentifier(raw); err == nil {
			t.Fatalf("NormalizeIdentifier(%q) returned nil error", raw)
		}
	}
}

func TestDevicePathHelpers(t *testing.T) {
	t.Parallel()

	if got := DevicePath("disk4"); got != "/dev/disk4" {
		t.Fatalf("DevicePath = %q, want /dev/disk4", got)
	}
	if got := RawDevicePath("disk4"); got != "/dev/rdisk4" {
		t.Fatalf("RawDevicePath = %q, want /dev/rdisk4", got)
	}
}
