package apusys

import (
	"bytes"
	"errors"
	"io/fs"
	"strings"

	"github.com/canonical/lscompute/pkg/machine/host"
)

const BusName = "apusys"

const (
	apusysDevicePattern = "dev/apusys*"
	compatiblePath      = "sys/firmware/devicetree/base/compatible"
)

// Device represents a MediaTek APUSYS MDLA NPU detected on the system.
type Device struct {
	Bus        string
	Type       string
	SocID      string
	VendorName string
}

// apusys is the APUSYS bus implementation.
type apusys struct {
	host host.Host
	opts Options
}

// Options holds APUSYS-specific scanner configuration.
type Options struct {
	// e.g. FriendlyName
}

// NewBus returns an APUSYS bus configured with the given options.
func NewBus(targetHost host.Host, opts Options) *apusys {
	return &apusys{host: targetHost, opts: opts}
}

// Devices discovers MediaTek APUSYS NPUs exposed through an APUSYS device node.
func (bus *apusys) Devices() ([]Device, []string, error) {
	present, err := hasAPUSYSDevice(bus.host.FS())
	if err != nil {
		return nil, nil, err
	}
	if !present {
		return nil, nil, nil
	}

	compatible, err := readCompatible(bus.host.FS())
	if err != nil {
		return nil, nil, err
	}
	vendorName, socID := parseCompatible(compatible)

	return []Device{
		{
			Bus:        BusName,
			Type:       "mdla",
			SocID:      socID,
			VendorName: vendorName,
		},
	}, nil, nil
}

func hasAPUSYSDevice(fsys fs.FS) (bool, error) {
	matches, err := fs.Glob(fsys, apusysDevicePattern)
	if err != nil {
		return false, err
	}
	return len(matches) > 0, nil
}

func readCompatible(fsys fs.FS) ([]byte, error) {
	data, err := fs.ReadFile(fsys, compatiblePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return data, nil
}

// parseCompatible extracts the vendor and SoC model from the device-tree
// compatible property, whose values are NUL-separated strings such as
// "mediatek,mt8195-evb" and "mediatek,mt8195".
func parseCompatible(data []byte) (vendorName, socID string) {
	var fallbackVendor, fallbackSoCID string

	for _, raw := range bytes.Split(data, []byte{0}) {
		compatible := strings.TrimSpace(string(raw))
		vendor, model, ok := strings.Cut(compatible, ",")
		if !ok {
			continue
		}

		vendor = strings.TrimSpace(vendor)
		model = socIDFromModel(model)
		if vendor == "" || model == "" {
			continue
		}

		if fallbackVendor == "" {
			fallbackVendor = vendor
			fallbackSoCID = model
		}
		if strings.EqualFold(vendor, "mediatek") {
			return vendor, model
		}
	}

	return fallbackVendor, fallbackSoCID
}

func socIDFromModel(model string) string {
	model = strings.ToLower(strings.TrimSpace(model))
	if !strings.HasPrefix(model, "mt") {
		return model
	}

	end := 2
	for end < len(model) && model[end] >= '0' && model[end] <= '9' {
		end++
	}
	if end == 2 {
		return ""
	}
	return model[:end]
}
