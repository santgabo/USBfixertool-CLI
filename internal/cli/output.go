package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/santgabo/usbfixertool-cli/internal/usbfix"
)

func validateOutputFormat(format string) error {
	switch format {
	case "table", "json", "plain":
		return nil
	default:
		return usbfix.Usagef("unsupported output format %q; choose table, json, or plain", format)
	}
}

func writeJSON(w io.Writer, value any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}

func writeDiskList(w io.Writer, format string, disks []usbfix.DiskEntry) error {
	switch format {
	case "json":
		return writeJSON(w, disks)
	case "plain":
		for _, disk := range disks {
			if _, err := fmt.Fprintf(w, "%s\t%s\t%s\n", disk.Identifier, disk.DeviceLocation, disk.Size); err != nil {
				return err
			}
		}
		return nil
	default:
		tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
		fmt.Fprintln(tw, "IDENTIFIER\tLOCATION\tREMOVABLE\tINTERNAL\tSIZE\tNAME")
		for _, disk := range disks {
			fmt.Fprintf(tw, "%s\t%s\t%t\t%t\t%s\t%s\n", disk.Identifier, disk.DeviceLocation, disk.Removable, disk.Internal, disk.Size, firstText(disk.Name, "-"))
		}
		return tw.Flush()
	}
}

func writeDeviceInfo(w io.Writer, format string, info usbfix.DeviceInfo) error {
	switch format {
	case "json":
		return writeJSON(w, info)
	case "plain":
		lines := []string{
			"identifier=" + info.Identifier,
			"device_path=" + info.DevicePath,
			"whole=" + fmt.Sprint(info.Whole),
			"internal=" + fmt.Sprint(info.Internal),
			"external=" + fmt.Sprint(info.External),
			"removable=" + fmt.Sprint(info.Removable),
		}
		for _, line := range lines {
			fmt.Fprintln(w, line)
		}
		return nil
	default:
		tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
		rows := [][2]string{
			{"Identifier", info.Identifier},
			{"Device Path", info.DevicePath},
			{"Name", info.Name},
			{"Volume Name", info.VolumeName},
			{"File System", info.FileSystem},
			{"Content", info.Content},
			{"Size", info.Size},
			{"Whole", fmt.Sprint(info.Whole)},
			{"Whole Disk", info.WholeDiskIdentifier},
			{"Location", info.DeviceLocation},
			{"Protocol", info.Protocol},
			{"Internal", fmt.Sprint(info.Internal)},
			{"External", fmt.Sprint(info.External)},
			{"Removable", fmt.Sprint(info.Removable)},
			{"Mounted", fmt.Sprint(info.Mounted)},
			{"Mount Point", info.MountPoint},
		}
		for _, row := range rows {
			if strings.TrimSpace(row[1]) == "" {
				continue
			}
			fmt.Fprintf(tw, "%s:\t%s\n", row[0], row[1])
		}
		if info.Raw != "" {
			fmt.Fprintln(tw)
			fmt.Fprintln(tw, "Raw diskutil info:")
			fmt.Fprintln(tw, info.Raw)
		}
		return tw.Flush()
	}
}

func firstText(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
