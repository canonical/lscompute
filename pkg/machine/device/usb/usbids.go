package usb

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"strings"

	"github.com/canonical/lscompute/pkg/machine/host"
	"github.com/canonical/lscompute/pkg/machine/types"
)

// usbIdsSearchPaths lists candidate paths for the usb.ids database, in priority order.
// Paths follow io/fs convention (no leading slash).
var usbIdsSearchPaths = []string{
	"usr/share/misc/usb.ids",
	"usr/share/hwdata/usb.ids",
	"usr/share/usb.ids",
}

// usbIdEntry holds the names resolved from the USB IDs database.
type usbIdEntry struct {
	VendorName  string
	ProductName string
}

// lookupUsbIds looks up the human-readable vendor and product names for the
// given IDs from the usb.ids database file.
// Both names may be empty if the IDs are not found.
func lookupUsbIds(h host.Host, vendorId, productId types.HexInt) (usbIdEntry, error) {
	path, err := findUsbIdsFile(h)
	if err != nil {
		return usbIdEntry{}, err
	}

	data, err := fs.ReadFile(h.FS(), path)
	if err != nil {
		return usbIdEntry{}, fmt.Errorf("opening usb.ids: %w", err)
	}

	vendorHex := fmt.Sprintf("%04x", uint64(vendorId))
	productHex := fmt.Sprintf("%04x", uint64(productId))

	var result usbIdEntry
	var inVendor bool

	scanner := bufio.NewScanner(bytes.NewReader(data))
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
	fields := strings.SplitN(line, "  ", 2)
	if len(fields) < 2 {
		return "", "", false
	}
	return strings.TrimSpace(fields[0]), strings.TrimSpace(fields[1]), true
}

// findUsbIdsFile returns the io/fs path of the first usb.ids file found via h.FS().
func findUsbIdsFile(h host.Host) (string, error) {
	for _, path := range usbIdsSearchPaths {
		if _, err := fs.Stat(h.FS(), path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("usb.ids database not found (searched: %s)", strings.Join(usbIdsSearchPaths, ", "))
}

// lookupFriendlyNames resolves human-readable vendor and product names for a
// device from the usb.ids database and populates the device's FriendlyNames.
// Errors are returned so the caller can emit them as warnings.
func lookupFriendlyNames(h host.Host, device Device) (Device, error) {
	entry, err := lookupUsbIds(h, device.VendorId, device.ProductId)
	if err != nil {
		return device, err
	}
	if entry.VendorName != "" {
		name := entry.VendorName
		device.VendorName = &name
	}
	if entry.ProductName != "" {
		name := entry.ProductName
		device.ProductName = &name
	}
	return device, nil
}
