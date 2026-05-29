package usbfix

import (
	"regexp"
	"strings"
)

var compatibleLabelPattern = regexp.MustCompile(`^[A-Z0-9_-]+$`)

type LabelValidation struct {
	Warnings []string
}

func NormalizeFilesystem(raw string, caseSensitive bool) (string, error) {
	fs := strings.ToLower(strings.TrimSpace(raw))
	switch fs {
	case "exfat":
		return "ExFAT", nil
	case "fat32", "msdos", "ms-dos", "ms-dos fat32":
		return "MS-DOS FAT32", nil
	case "apfs":
		if caseSensitive {
			return "Case-sensitive APFS", nil
		}
		return "APFS", nil
	case "hfs+", "hfsplus", "jhfs+":
		if caseSensitive {
			return "Case-sensitive Journaled HFS+", nil
		}
		return "Journaled HFS+", nil
	default:
		return "", Usagef("unsupported filesystem %q; choose exfat, fat32, apfs, or hfs+", raw)
	}
}

func FilesystemKey(raw string) (string, error) {
	fs := strings.ToLower(strings.TrimSpace(raw))
	switch fs {
	case "exfat":
		return "exfat", nil
	case "fat32", "msdos", "ms-dos", "ms-dos fat32":
		return "fat32", nil
	case "apfs":
		return "apfs", nil
	case "hfs+", "hfsplus", "jhfs+":
		return "hfs+", nil
	default:
		return "", Usagef("unsupported filesystem %q; choose exfat, fat32, apfs, or hfs+", raw)
	}
}

func NormalizeScheme(raw string) (string, error) {
	scheme := strings.ToLower(strings.TrimSpace(raw))
	switch scheme {
	case "mbr", "fdisk":
		return "MBR", nil
	case "gpt", "guid":
		return "GPT", nil
	case "apm", "apple":
		return "APM", nil
	default:
		return "", Usagef("unsupported partition scheme %q; choose mbr, gpt, or apm", raw)
	}
}

func ValidateLabel(filesystem, label string) (LabelValidation, error) {
	label = strings.TrimSpace(label)
	if label == "" {
		return LabelValidation{}, Usagef("volume label is required")
	}

	key, err := FilesystemKey(filesystem)
	if err != nil {
		return LabelValidation{}, err
	}

	result := LabelValidation{}
	switch key {
	case "exfat", "fat32":
		if len(label) > 11 {
			return result, Usagef("%s labels must be 11 characters or fewer", filesystem)
		}
		if !compatibleLabelPattern.MatchString(label) {
			result.Warnings = append(result.Warnings, "label is accepted, but uppercase letters, numbers, '_' and '-' are safest for exFAT/FAT32 compatibility")
		}
	case "apfs", "hfs+":
		if len(label) > 255 {
			return result, Usagef("%s labels must be 255 characters or fewer", filesystem)
		}
	}

	return result, nil
}
