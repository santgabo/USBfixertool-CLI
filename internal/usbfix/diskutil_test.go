package usbfix

import "testing"

func TestParseDiskutilInfo(t *testing.T) {
	t.Parallel()

	raw := `   Device Identifier:         disk4
   Device Node:               /dev/disk4
   Whole:                     Yes
   Part of Whole:             disk4
   Device / Media Name:       USB Flash Disk
   Disk Size:                 31.0 GB (31000000000 Bytes)
   Device Location:           External
   Removable Media:           Removable
   Internal:                  No
`

	info := ParseDiskutilInfo(raw)
	if info.Identifier != "disk4" {
		t.Fatalf("identifier = %q, want disk4", info.Identifier)
	}
	if !info.Whole {
		t.Fatal("expected whole disk")
	}
	if !info.External {
		t.Fatal("expected external disk")
	}
	if !info.Removable {
		t.Fatal("expected removable disk")
	}
	if info.Internal {
		t.Fatal("did not expect internal disk")
	}
}

func TestParseDiskutilList(t *testing.T) {
	t.Parallel()

	raw := `/dev/disk4 (external, physical):
   #:                       TYPE NAME                    SIZE       IDENTIFIER
   0:      FDisk_partition_scheme                        *31.0 GB   disk4
   1:             Windows_FAT_32 USBBOOT                 31.0 GB    disk4s1
`

	disks := ParseDiskutilList(raw)
	if len(disks) != 1 {
		t.Fatalf("got %d disks, want 1", len(disks))
	}
	if disks[0].Identifier != "disk4" {
		t.Fatalf("disk identifier = %q, want disk4", disks[0].Identifier)
	}
	if !disks[0].External {
		t.Fatal("expected external disk")
	}
	if disks[0].Size != "31.0 GB" {
		t.Fatalf("size = %q, want 31.0 GB", disks[0].Size)
	}
	if len(disks[0].Volumes) != 1 || disks[0].Volumes[0].Identifier != "disk4s1" {
		t.Fatalf("volumes = %#v, want disk4s1", disks[0].Volumes)
	}
}

func TestParseDiskutilInfoMountedNotApplicableIsFalse(t *testing.T) {
	t.Parallel()

	raw := `   Device Identifier:         disk4
   Device Node:               /dev/disk4
   Whole:                     Yes
   Mount Point:               Not applicable
   Mounted:                   No
`

	info := ParseDiskutilInfo(raw)
	if info.Mounted {
		t.Fatalf("Mounted = true, want false for Not applicable mount point")
	}
}

func TestParseDiskutilListKeepsVolumeNamesWithSpaces(t *testing.T) {
	t.Parallel()

	raw := `/dev/disk7 (external, physical):
   #:                       TYPE NAME                    SIZE       IDENTIFIER
   0:      GUID_partition_scheme                        *64.0 GB    disk7
   1:                        EFI EFI                     209.7 MB   disk7s1
   2:                  Apple_HFS USB Backup Drive        63.7 GB    disk7s2
`

	disks := ParseDiskutilList(raw)
	if len(disks) != 1 {
		t.Fatalf("got %d disks, want 1", len(disks))
	}
	if len(disks[0].Volumes) != 2 {
		t.Fatalf("got %d volumes, want 2: %#v", len(disks[0].Volumes), disks[0].Volumes)
	}
	if got := disks[0].Volumes[1].Name; got != "USB Backup Drive" {
		t.Fatalf("volume name = %q, want USB Backup Drive", got)
	}
	if got := disks[0].Volumes[1].Size; got != "63.7 GB" {
		t.Fatalf("volume size = %q, want 63.7 GB", got)
	}
}
