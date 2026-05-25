package usb

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/canonical/lscompute/pkg/machine/types"
)

// usbIdsSearchPaths lists candidate paths for the usb.ids database, in priority order.
var usbIdsSearchPaths = []string{
	"/usr/share/misc/usb.ids",
	"/usr/share/hwdata/usb.ids",
	"/usr/share/usb.ids",
}

// usbIdEntry holds the names resolved from the USB IDs database.
type usbIdEntry struct {
	VendorName  string
	ProductName string
}

// lookupUsbIds looks up the human-readable vendor and product names for the
// given IDs from the system's usb.ids database file.
// Both names may be empty if the IDs are not found.
func lookupUsbIds(vendorId, productId types.HexInt) (usbIdEntry, error) {
	path, err := findUsbIdsFile()
	if err != nil {
		return usbIdEntry{}, err
	}

	f, err := os.Open(path)
	if err != nil {
		return usbIdEntry{}, fmt.Errorf("opening usb.ids: %w", err)
	}
	defer f.Close()

	vendorHex := fmt.Sprintf("%04x", uint64(vendorId))
	productHex := fmt.Sprintf("%04x", uint64(productId))

	var result usbIdEntry
	var inVendor bool

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Lines starting with a tab are product entries (or interface entries
		// with two tabs — we ignore those).
		if strings.HasPrefix(line, "\t\t") {
			// Interface entry — skip
			continue
		}
		if strings.HasPrefix(line, "\t") {
			if !inVendor {
				continue
			}
			// Product line: "\tPPPP  Product Name"
			rest := strings.TrimPrefix(line, "\t")
			id, name, ok := splitIdName(rest)
			if ok && id == productHex {
				result.ProductName = name
				// Both names found — we're done.
				if result.VendorName != "" {
					return result, nil
				}
			}
			continue
		}

		// Unindented line: a new vendor entry.
		// If we were in the right vendor block and already have the vendor
		// name, a new unindented line means we've left that block.
		if inVendor {
			// Left the vendor block without finding the product — still return
			// what we have.
			return result, nil
		}

		id, name, ok := splitIdName(line)
		if ok && id == vendorHex {
			result.VendorName = name
			inVendor = true
		}
	}

	if err := scanner.Err(); err != nil {
		return usbIdEntry{}, fmt.Errorf("reading usb.ids: %w", err)
	}

	return result, nil
}

// splitIdName splits a line of the form "XXXX  Some Name" into id and name.
func splitIdName(line string) (id, name string, ok bool) {
	// The separator between ID and name is two or more spaces (or a tab).
	// Use Fields-based splitting: first field is the ID, rest is the name.
	fields := strings.SplitN(line, "  ", 2)
	if len(fields) < 2 {
		return "", "", false
	}
	return strings.TrimSpace(fields[0]), strings.TrimSpace(fields[1]), true
}

// findUsbIdsFile returns the path of the first usb.ids file found on the system.
func findUsbIdsFile() (string, error) {
	for _, path := range usbIdsSearchPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("usb.ids database not found (searched: %s)", strings.Join(usbIdsSearchPaths, ", "))
}

