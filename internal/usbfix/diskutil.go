package usbfix

import (
	"context"
	"regexp"
	"strings"
)

type Diskutil struct {
	Runner SystemRunner
}

type DeviceInfo struct {
	Identifier          string `json:"identifier"`
	DevicePath          string `json:"devicePath"`
	Name                string `json:"name,omitempty"`
	VolumeName          string `json:"volumeName,omitempty"`
	FileSystem          string `json:"fileSystem,omitempty"`
	Content             string `json:"content,omitempty"`
	Size                string `json:"size,omitempty"`
	WholeDiskIdentifier string `json:"wholeDiskIdentifier,omitempty"`
	MountPoint          string `json:"mountPoint,omitempty"`
	DeviceLocation      string `json:"deviceLocation,omitempty"`
	Protocol            string `json:"protocol,omitempty"`
	Whole               bool   `json:"whole"`
	Internal            bool   `json:"internal"`
	External            bool   `json:"external"`
	Removable           bool   `json:"removable"`
	Mounted             bool   `json:"mounted"`
	Raw                 string `json:"raw,omitempty"`
}

type DiskEntry struct {
	Identifier     string        `json:"identifier"`
	DevicePath     string        `json:"devicePath"`
	Size           string        `json:"size,omitempty"`
	Name           string        `json:"name,omitempty"`
	DeviceLocation string        `json:"deviceLocation,omitempty"`
	Whole          bool          `json:"whole"`
	Internal       bool          `json:"internal"`
	External       bool          `json:"external"`
	Removable      bool          `json:"removable"`
	Volumes        []VolumeEntry `json:"volumes,omitempty"`
}

type VolumeEntry struct {
	Identifier string `json:"identifier"`
	Name       string `json:"name,omitempty"`
	Type       string `json:"type,omitempty"`
	Size       string `json:"size,omitempty"`
}

func NewDiskutil(runner SystemRunner) *Diskutil {
	if runner == nil {
		runner = ExecRunner{}
	}
	return &Diskutil{Runner: runner}
}

func (d *Diskutil) CheckAvailable() error {
	if _, err := d.Runner.LookPath("diskutil"); err != nil {
		return CommandNotFoundf("required command not found: diskutil")
	}
	return nil
}

func (d *Diskutil) Info(ctx context.Context, rawIdentifier string) (DeviceInfo, error) {
	id, err := NormalizeIdentifier(rawIdentifier)
	if err != nil {
		return DeviceInfo{}, err
	}
	out, err := d.Runner.Run(ctx, "diskutil", "info", DevicePath(id))
	if err != nil {
		return DeviceInfo{}, err
	}
	info := ParseDiskutilInfo(string(out))
	if info.Identifier == "" {
		info.Identifier = id
	}
	if info.DevicePath == "" {
		info.DevicePath = DevicePath(id)
	}
	return info, nil
}

func (d *Diskutil) List(ctx context.Context) ([]DiskEntry, error) {
	out, err := d.Runner.Run(ctx, "diskutil", "list")
	if err != nil {
		return nil, err
	}
	disks := ParseDiskutilList(string(out))
	for i := range disks {
		info, err := d.Info(ctx, disks[i].Identifier)
		if err != nil {
			continue
		}
		disks[i].Name = firstNonEmpty(info.Name, disks[i].Name)
		disks[i].DeviceLocation = info.DeviceLocation
		disks[i].Whole = info.Whole
		disks[i].Internal = info.Internal
		disks[i].External = info.External
		disks[i].Removable = info.Removable
	}
	return disks, nil
}

func (d *Diskutil) VerifyDisk(ctx context.Context, disk string) ([]byte, error) {
	return d.Runner.Run(ctx, "diskutil", "verifyDisk", DevicePath(disk))
}

func (d *Diskutil) RepairDisk(ctx context.Context, disk string) ([]byte, error) {
	return d.Runner.Run(ctx, "diskutil", "repairDisk", DevicePath(disk))
}

func (d *Diskutil) VerifyVolume(ctx context.Context, volume string) ([]byte, error) {
	return d.Runner.Run(ctx, "diskutil", "verifyVolume", DevicePath(volume))
}

func (d *Diskutil) RepairVolume(ctx context.Context, volume string) ([]byte, error) {
	return d.Runner.Run(ctx, "diskutil", "repairVolume", DevicePath(volume))
}

func (d *Diskutil) UnmountDisk(ctx context.Context, disk string, force bool) ([]byte, error) {
	args := []string{"unmountDisk"}
	if force {
		args = append(args, "force")
	}
	args = append(args, DevicePath(disk))
	return d.Runner.Run(ctx, "diskutil", args...)
}

func (d *Diskutil) UnmountVolume(ctx context.Context, volume string, force bool) ([]byte, error) {
	args := []string{"unmount"}
	if force {
		args = append(args, "force")
	}
	args = append(args, DevicePath(volume))
	return d.Runner.Run(ctx, "diskutil", args...)
}

func (d *Diskutil) MountDisk(ctx context.Context, disk string) ([]byte, error) {
	return d.Runner.Run(ctx, "diskutil", "mountDisk", DevicePath(disk))
}

func (d *Diskutil) MountVolume(ctx context.Context, volume string) ([]byte, error) {
	return d.Runner.Run(ctx, "diskutil", "mount", DevicePath(volume))
}

func (d *Diskutil) EraseDisk(ctx context.Context, filesystem, label, scheme, disk string) ([]byte, error) {
	return d.Runner.Run(ctx, "diskutil", "eraseDisk", filesystem, label, scheme, DevicePath(disk))
}

func (d *Diskutil) FreeDisk(ctx context.Context, scheme, disk string) ([]byte, error) {
	return d.Runner.Run(ctx, "diskutil", "eraseDisk", "free", "EMPTY", scheme, DevicePath(disk))
}

func (d *Diskutil) ZeroFirstMB(ctx context.Context, disk string) ([]byte, error) {
	return d.Runner.Run(ctx, "dd", "if=/dev/zero", "of="+RawDevicePath(disk), "bs=1m", "count=1")
}

func (d *Diskutil) Fsck(ctx context.Context, filesystem, volume string, repair bool) ([]byte, error) {
	fs := strings.ToLower(filesystem)
	switch {
	case strings.Contains(fs, "exfat"):
		flag := "-n"
		if repair {
			flag = "-y"
		}
		return d.Runner.Run(ctx, "fsck_exfat", flag, DevicePath(volume))
	case strings.Contains(fs, "fat"), strings.Contains(fs, "ms-dos"), strings.Contains(fs, "dos"):
		flag := "-n"
		if repair {
			flag = "-y"
		}
		return d.Runner.Run(ctx, "fsck_msdos", flag, DevicePath(volume))
	default:
		return nil, Usagef("no filesystem-specific fsck adapter for %q yet", filesystem)
	}
}

func ParseDiskutilInfo(raw string) DeviceInfo {
	fields := parseColonFields(raw)
	info := DeviceInfo{
		Identifier:          fields["Device Identifier"],
		DevicePath:          fields["Device Node"],
		Name:                fields["Device / Media Name"],
		VolumeName:          fields["Volume Name"],
		FileSystem:          firstNonEmpty(fields["File System Personality"], fields["File System"]),
		Content:             firstNonEmpty(fields["Content (IOContent)"], fields["Content"]),
		Size:                fields["Disk Size"],
		WholeDiskIdentifier: fields["Part of Whole"],
		MountPoint:          fields["Mount Point"],
		DeviceLocation:      fields["Device Location"],
		Protocol:            fields["Protocol"],
		Raw:                 raw,
	}
	info.Whole = parseBool(fields["Whole"])
	info.Internal = parseBool(fields["Internal"]) || strings.EqualFold(info.DeviceLocation, "Internal")
	info.External = strings.EqualFold(info.DeviceLocation, "External")
	info.Removable = parseRemovable(fields["Removable Media"])
	info.Mounted = parseBool(fields["Mounted"]) || (info.MountPoint != "" && !strings.EqualFold(info.MountPoint, "Not applicable"))
	return info
}

func parseColonFields(raw string) map[string]string {
	fields := make(map[string]string)
	for _, line := range strings.Split(raw, "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		fields[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	return fields
}

func parseBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "yes", "true", "1", "mounted":
		return true
	default:
		return false
	}
}

func parseRemovable(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "yes", "true", "1", "removable":
		return true
	default:
		return false
	}
}

var diskHeaderPattern = regexp.MustCompile(`^/dev/(disk[0-9]+)\s+\(([^)]*)\):`)

func ParseDiskutilList(raw string) []DiskEntry {
	var disks []DiskEntry
	var current *DiskEntry

	for _, line := range strings.Split(raw, "\n") {
		if match := diskHeaderPattern.FindStringSubmatch(line); match != nil {
			meta := strings.ToLower(match[2])
			disks = append(disks, DiskEntry{
				Identifier: match[1],
				DevicePath: DevicePath(match[1]),
				Whole:      true,
				Internal:   strings.Contains(meta, "internal"),
				External:   strings.Contains(meta, "external"),
			})
			current = &disks[len(disks)-1]
			continue
		}
		if current == nil {
			continue
		}
		entry := parseListRow(line)
		if entry.Identifier == "" {
			continue
		}
		if entry.Identifier == current.Identifier {
			current.Size = entry.Size
			current.Name = entry.Name
			continue
		}
		current.Volumes = append(current.Volumes, VolumeEntry(entry))
	}

	return disks
}

func parseListRow(line string) VolumeEntry {
	fields := strings.Fields(line)
	if len(fields) < 4 || !strings.HasSuffix(fields[0], ":") {
		return VolumeEntry{}
	}
	id := fields[len(fields)-1]
	if !anyDiskPattern.MatchString(id) {
		return VolumeEntry{}
	}

	sizeStart := len(fields) - 3
	if sizeStart < 2 {
		return VolumeEntry{Identifier: id}
	}
	size := strings.Join(fields[sizeStart:len(fields)-1], " ")
	nameFields := fields[2:sizeStart]
	name := strings.Join(nameFields, " ")
	name = strings.TrimPrefix(name, "*")

	return VolumeEntry{
		Identifier: id,
		Type:       fields[1],
		Name:       name,
		Size:       strings.TrimPrefix(size, "*"),
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" && !strings.EqualFold(strings.TrimSpace(value), "Not applicable") {
			return value
		}
	}
	return ""
}
