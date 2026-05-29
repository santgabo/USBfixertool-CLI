package usbfix

import "testing"

func TestNormalizeFilesystemAliases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		raw           string
		caseSensitive bool
		want          string
	}{
		{name: "exfat", raw: " exfat ", want: "ExFAT"},
		{name: "fat32 alias", raw: "ms-dos", want: "MS-DOS FAT32"},
		{name: "apfs case sensitive", raw: "apfs", caseSensitive: true, want: "Case-sensitive APFS"},
		{name: "hfs alias", raw: "hfsplus", want: "Journaled HFS+"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NormalizeFilesystem(tt.raw, tt.caseSensitive)
			if err != nil {
				t.Fatalf("NormalizeFilesystem returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeSchemeAliases(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"mbr":   "MBR",
		"fdisk": "MBR",
		"guid":  "GPT",
		"apple": "APM",
	}
	for raw, want := range tests {
		raw, want := raw, want
		t.Run(raw, func(t *testing.T) {
			t.Parallel()

			got, err := NormalizeScheme(raw)
			if err != nil {
				t.Fatalf("NormalizeScheme returned error: %v", err)
			}
			if got != want {
				t.Fatalf("got %q, want %q", got, want)
			}
		})
	}
}

func TestValidateExFATLabelLength(t *testing.T) {
	t.Parallel()

	if _, err := ValidateLabel("exfat", "TOO-LONG-USB"); err == nil {
		t.Fatal("expected long exFAT label to be rejected")
	}
}

func TestValidateExFATLabelCompatibilityWarning(t *testing.T) {
	t.Parallel()

	result, err := ValidateLabel("exfat", "usbboot")
	if err != nil {
		t.Fatalf("ValidateLabel returned error: %v", err)
	}
	if len(result.Warnings) != 1 {
		t.Fatalf("got %d warnings, want 1", len(result.Warnings))
	}
}

func TestValidateAPFSLabelAllowsLongerHumanName(t *testing.T) {
	t.Parallel()

	result, err := ValidateLabel("apfs", "My Install Media")
	if err != nil {
		t.Fatalf("ValidateLabel returned error: %v", err)
	}
	if len(result.Warnings) != 0 {
		t.Fatalf("warnings = %#v, want none", result.Warnings)
	}
}
