package pci

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"strings"

	"github.com/canonical/lscompute/pkg/machine/host"
	"github.com/canonical/lscompute/pkg/machine/types"
)

// pciIdsSearchPaths lists candidate paths for the pci.ids database, in priority order.
// Paths follow io/fs convention (no leading slash).
var pciIdsSearchPaths = []string{
	"usr/share/misc/pci.ids",
	"usr/share/hwdata/pci.ids",
	"usr/share/pci.ids",
}

// pciIdEntry holds the names resolved from the PCI IDs database.
type pciIdEntry struct {
	VendorName    string
	DeviceName    string
	SubvendorName string
	SubdeviceName string
}

// lookupPciIds looks up the human-readable vendor, device, subvendor and
// subdevice names for the given IDs from the pci.ids database file.
// Any name may be empty if the corresponding ID is not found.
func lookupPciIds(h host.Host, vendorId, deviceId types.HexInt, subvendorId, subdeviceId *types.HexInt) (pciIdEntry, error) {
	path, err := findPciIdsFile(h)
	if err != nil {
		return pciIdEntry{}, err
	}

	data, err := fs.ReadFile(h.FS(), path)
	if err != nil {
		return pciIdEntry{}, fmt.Errorf("opening pci.ids: %w", err)
	}

	vendorHex := fmt.Sprintf("%04x", uint64(vendorId))
	deviceHex := fmt.Sprintf("%04x", uint64(deviceId))

	subvendorHex := ""
	subdeviceHex := ""
	if subvendorId != nil {
		subvendorHex = fmt.Sprintf("%04x", uint64(*subvendorId))
	}
	if subdeviceId != nil {
		subdeviceHex = fmt.Sprintf("%04x", uint64(*subdeviceId))
	}

	var result pciIdEntry

	// Parsing state
	var inTargetVendor bool // inside the vendor block for vendorId
	var inTargetDevice bool // inside the device block for deviceId (implies inTargetVendor)
	var currentVendorId string

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Skip class definitions (lines starting with "C ")
		if strings.HasPrefix(line, "C ") {
			// Once we hit class definitions the vendor list is done.
			break
		}

		// Three-tab lines don't exist in pci.ids; two-tab lines are subsystems.
		if strings.HasPrefix(line, "\t\t") {
			// Subsystem line: "\t\tSSVV SSDD  Subsystem Name"
			if inTargetDevice && subvendorId != nil && subdeviceId != nil {
				rest := strings.TrimPrefix(line, "\t\t")
				sv, sd, name, ok := splitSubsystemLine(rest)
				if ok && sv == subvendorHex && sd == subdeviceHex {
					result.SubdeviceName = name
				}
			}
			continue
		}

		if strings.HasPrefix(line, "\t") {
			// Device line: "\tDDDD  Device Name"
			if inTargetVendor {
				rest := strings.TrimPrefix(line, "\t")
				id, name, ok := splitPciIdName(rest)
				if ok && id == deviceHex {
					result.DeviceName = name
					inTargetDevice = true
				} else {
					// Moved to a different device entry
					if inTargetDevice {
						inTargetDevice = false
					}
				}
			}
			continue
		}

		// Unindented line: a new vendor entry (or end of previous one).
		inTargetDevice = false
		if inTargetVendor {
			// We've left the target vendor block.
			inTargetVendor = false
		}

		id, name, ok := splitPciIdName(line)
		if !ok {
			continue
		}
		currentVendorId = id

		if id == vendorHex {
			result.VendorName = name
			inTargetVendor = true
		}

		// Check whether this vendor is also the subvendor we're looking for.
		if subvendorId != nil && currentVendorId == subvendorHex {
			result.SubvendorName = name
		}
	}

	if err := scanner.Err(); err != nil {
		return pciIdEntry{}, fmt.Errorf("reading pci.ids: %w", err)
	}

	return result, nil
}

// splitPciIdName splits a line of the form "XXXX  Some Name" into id and name.
func splitPciIdName(line string) (id, name string, ok bool) {
	fields := strings.SplitN(line, "  ", 2)
	if len(fields) < 2 {
		return "", "", false
	}
	return strings.TrimSpace(fields[0]), strings.TrimSpace(fields[1]), true
}

// splitSubsystemLine splits a subsystem line of the form "SVVV SDDD  Name" into
// subvendor id, subdevice id and name.
func splitSubsystemLine(line string) (subvendor, subdevice, name string, ok bool) {
	// Format: "svvv  sddd  Subsystem Name"
	// The two IDs are separated by a space; the name follows two spaces.
	fields := strings.SplitN(line, "  ", 2)
	if len(fields) < 2 {
		return "", "", "", false
	}
	ids := strings.Fields(fields[0])
	if len(ids) < 2 {
		return "", "", "", false
	}
	return ids[0], ids[1], strings.TrimSpace(fields[1]), true
}

// findPciIdsFile returns the io/fs path of the first pci.ids file found via h.FS().
func findPciIdsFile(h host.Host) (string, error) {
	for _, path := range pciIdsSearchPaths {
		if _, err := fs.Stat(h.FS(), path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("pci.ids database not found (searched: %s)", strings.Join(pciIdsSearchPaths, ", "))
}
