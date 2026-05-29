package usbfix

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	wholeDiskPattern = regexp.MustCompile(`^disk[0-9]+$`)
	volumePattern    = regexp.MustCompile(`^disk[0-9]+s[0-9]+$`)
	anyDiskPattern   = regexp.MustCompile(`^disk[0-9]+(s[0-9]+)?$`)
)

func NormalizeIdentifier(raw string) (string, error) {
	id := strings.TrimSpace(raw)
	id = strings.TrimPrefix(id, "/dev/")
	if id == "" {
		return "", Usagef("disk identifier is required")
	}
	if !anyDiskPattern.MatchString(id) {
		return "", Usagef("invalid disk identifier %q; expected diskN or diskNsM", raw)
	}
	return id, nil
}

func NormalizeWholeDiskIdentifier(raw string) (string, error) {
	id, err := NormalizeIdentifier(raw)
	if err != nil {
		return "", err
	}
	if !wholeDiskPattern.MatchString(id) {
		return "", Usagef("invalid whole-disk identifier %q; use diskN, not a partition such as diskNsM", raw)
	}
	return id, nil
}

func NormalizeVolumeIdentifier(raw string) (string, error) {
	id, err := NormalizeIdentifier(raw)
	if err != nil {
		return "", err
	}
	if !volumePattern.MatchString(id) {
		return "", Usagef("invalid volume identifier %q; expected a partition such as disk4s1", raw)
	}
	return id, nil
}

func DevicePath(identifier string) string {
	return fmt.Sprintf("/dev/%s", identifier)
}

func RawDevicePath(identifier string) string {
	return fmt.Sprintf("/dev/r%s", identifier)
}

func IsWholeDiskIdentifier(identifier string) bool {
	return wholeDiskPattern.MatchString(identifier)
}

func IsVolumeIdentifier(identifier string) bool {
	return volumePattern.MatchString(identifier)
}
