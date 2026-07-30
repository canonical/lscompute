package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/canonical/lscompute/pkg/machine"
	"go.yaml.in/yaml/v4"
)

const (
	// FormatPlain is a human-readable YAML output.
	FormatPlain string = "plain"
	// FormatJSON is a machine-readable, indented JSON output with kebab-cased keys.
	FormatJSON string = "json"
)

type MachineDetails struct {
	Cpus    []CpuDetails  `json:"cpus,omitempty" yaml:"cpus,omitempty"`
	Memory  MemoryDetails `json:"memory,omitempty" yaml:"memory,omitempty"`
	Disk    []DiskDetails `json:"disks,omitempty" yaml:"disks,omitempty"`
	Devices []any         `json:"devices,omitempty" yaml:"devices,omitempty"`
}

type CpuDetails struct {
	Architecture string `json:"architecture" yaml:"architecture"`

	// amd64
	ManufacturerId string   `json:"manufacturer-id,omitempty" yaml:"manufacturer-id,omitempty"`
	Flags          []string `json:"flags,omitempty" yaml:"flags,flow,omitempty"`

	// arm64
	ImplementerId HexInt   `json:"implementer-id,omitempty" yaml:"implementer-id,omitempty"`
	PartNumber    HexInt   `json:"part-number,omitempty" yaml:"part-number,omitempty"`
	Features      []string `json:"features,omitempty" yaml:"features,omitempty"`

	// riscv64
	Isa []string `json:"isa,omitempty" yaml:"isa,omitempty"`
}

type MemoryDetails struct {
	TotalRam  uint64 `json:"total-ram" yaml:"total-ram"`
	TotalSwap uint64 `json:"total-swap" yaml:"total-swap"`
}

func (m MemoryDetails) MarshalYAML() (any, error) {
	return struct {
		TotalRam  any `yaml:"total-ram"`
		TotalSwap any `yaml:"total-swap"`
	}{
		TotalRam:  FormatBytes(m.TotalRam),
		TotalSwap: FormatBytes(m.TotalSwap),
	}, nil
}

type DiskDetails struct {
	MountPoint *string `json:"mount-point,omitempty" yaml:"mount-point,omitempty"`
	Path       string  `json:"path" yaml:"path"`
	Total      uint64  `json:"total" yaml:"total"`
	Avail      uint64  `json:"avail" yaml:"avail"`
}

func (d DiskDetails) MarshalYAML() (any, error) {
	return struct {
		MountPoint *string `yaml:"mount-point,omitempty"`
		Path       string  `yaml:"path"`
		Total      any     `yaml:"total"`
		Avail      any     `yaml:"avail"`
	}{
		MountPoint: d.MountPoint,
		Path:       d.Path,
		Total:      FormatBytes(d.Total),
		Avail:      FormatBytes(d.Avail),
	}, nil
}

type PciDeviceDetails struct {
	Bus                  string                         `json:"bus" yaml:"bus"`
	Slot                 string                         `json:"slot,omitempty" yaml:"slot,omitempty"`
	BusNumber            any                            `json:"bus-number,omitempty" yaml:"bus-number,omitempty"`
	DeviceClass          HexInt                         `json:"device-class,omitempty" yaml:"device-class,omitempty"`
	ProgrammingInterface uint8                          `json:"programming-interface,omitempty" yaml:"programming-interface,omitempty"`
	VendorId             HexInt                         `json:"vendor-id,omitempty" yaml:"vendor-id,omitempty"`
	DeviceId             HexInt                         `json:"device-id,omitempty" yaml:"device-id,omitempty"`
	SubvendorId          HexInt                         `json:"subvendor-id,omitempty" yaml:"subvendor-id,omitempty"`
	SubdeviceId          HexInt                         `json:"subdevice-id,omitempty" yaml:"subdevice-id,omitempty"`
	VendorName           string                         `json:"vendor-name,omitempty" yaml:"vendor-name,omitempty"`
	DeviceName           string                         `json:"device-name,omitempty" yaml:"device-name,omitempty"`
	SubvendorName        string                         `json:"subvendor-name,omitempty" yaml:"subvendor-name,omitempty"`
	SubdeviceName        string                         `json:"subdevice-name,omitempty" yaml:"subdevice-name,omitempty"`
	AdditionalProperties *PciAdditionalDeviceProperties `json:"additional-properties,omitempty" yaml:"additional-properties,omitempty"`
}

type UsbDeviceDetails struct {
	Bus                  string            `json:"bus" yaml:"bus"`
	BusNumber            any               `json:"bus-number,omitempty" yaml:"bus-number,omitempty"`
	DeviceNumber         uint8             `json:"device-number,omitempty" yaml:"device-number,omitempty"`
	VendorId             HexInt            `json:"vendor-id,omitempty" yaml:"vendor-id,omitempty"`
	ProductId            HexInt            `json:"product-id,omitempty" yaml:"product-id,omitempty"`
	VendorName           string            `json:"vendor-name,omitempty" yaml:"vendor-name,omitempty"`
	ProductName          string            `json:"product-name,omitempty" yaml:"product-name,omitempty"`
	AdditionalProperties map[string]string `json:"additional-properties,omitempty" yaml:"additional-properties,omitempty"`
}

type FastRPCDeviceDetails struct {
	Bus                  string            `json:"bus" yaml:"bus"`
	Domain               string            `json:"domain,omitempty" yaml:"domain,omitempty"`
	Index                int               `json:"index,omitempty" yaml:"index,omitempty"`
	Secure               bool              `json:"secure,omitempty" yaml:"secure,omitempty"`
	AdditionalProperties map[string]string `json:"additional-properties,omitempty" yaml:"additional-properties,omitempty"`
}

type PciAdditionalDeviceProperties struct {
	Microarchitecture string `json:"microarchitecture,omitempty" yaml:"microarchitecture,omitempty"`
	Vram              uint64 `json:"vram,omitempty" yaml:"vram,omitempty"`
	ComputeCapability string `json:"compute-capability,omitempty" yaml:"compute-capability,omitempty"`
}

func (a PciAdditionalDeviceProperties) MarshalYAML() (any, error) {
	return struct {
		Microarchitecture string `yaml:"microarchitecture,omitempty"`
		Vram              any    `yaml:"vram,omitempty"`
		ComputeCapability string `yaml:"compute-capability,omitempty"`
	}{
		Microarchitecture: a.Microarchitecture,
		Vram:              FormatBytes(a.Vram),
		ComputeCapability: a.ComputeCapability,
	}, nil
}

func NewMachineDetails(info *machine.Machine) *MachineDetails {
	if info == nil {
		return nil
	}

	v := &MachineDetails{
		Memory: MemoryDetails(info.Memory),
	}

	// Combine all devices into a single slice
	totalDevices := len(info.PCIDevices) + len(info.USBDevices) + len(info.FastRPCDevices)
	v.Devices = make([]any, 0, totalDevices)

	// Add PCI devices
	for _, d := range info.PCIDevices {
		var programmingInterface uint8
		if d.ProgrammingInterface != nil {
			programmingInterface = *d.ProgrammingInterface
		}
		var subvendorId, subdeviceId uint16
		if d.SubvendorId != nil {
			subvendorId = *d.SubvendorId
		}
		if d.SubdeviceId != nil {
			subdeviceId = *d.SubdeviceId
		}
		v.Devices = append(v.Devices, PciDeviceDetails{
			Bus:                  d.Bus,
			Slot:                 d.Slot,
			BusNumber:            HexInt(d.BusNumber),
			DeviceClass:          HexInt(d.DeviceClass),
			ProgrammingInterface: programmingInterface,
			VendorId:             HexInt(d.VendorId),
			DeviceId:             HexInt(d.DeviceId),
			SubvendorId:          HexInt(subvendorId),
			SubdeviceId:          HexInt(subdeviceId),
			VendorName:           d.FriendlyNames.VendorName,
			DeviceName:           d.FriendlyNames.DeviceName,
			SubvendorName:        d.FriendlyNames.SubvendorName,
			SubdeviceName:        d.FriendlyNames.SubdeviceName,
			AdditionalProperties: newPciAdditionalDeviceProperties(d.AdditionalProperties),
		})
	}

	// Add USB devices
	for _, d := range info.USBDevices {
		v.Devices = append(v.Devices, UsbDeviceDetails{
			Bus:                  d.Bus,
			BusNumber:            HexInt(d.BusNumber),
			DeviceNumber:         d.DeviceNumber,
			VendorId:             HexInt(d.VendorId),
			ProductId:            HexInt(d.ProductId),
			VendorName:           d.FriendlyNames.VendorName,
			ProductName:          d.FriendlyNames.ProductName,
			AdditionalProperties: d.AdditionalProperties,
		})
	}

	// Add FastRPC devices
	for _, d := range info.FastRPCDevices {
		v.Devices = append(v.Devices, FastRPCDeviceDetails{
			Bus:                  d.Bus,
			Domain:               string(d.Domain),
			Index:                d.Index,
			Secure:               d.Secure,
			AdditionalProperties: d.AdditionalProperties,
		})
	}

	if info.CPUs != nil {
		v.Cpus = make([]CpuDetails, len(info.CPUs))
		for i, c := range info.CPUs {
			v.Cpus[i] = CpuDetails{
				Architecture:   c.Architecture,
				ManufacturerId: c.ManufacturerId,
				Flags:          c.Flags,
				ImplementerId:  HexInt(c.ImplementerId),
				PartNumber:     HexInt(c.PartNumber),
				Features:       c.Features,
				Isa:            c.Isa,
			}
		}
	}

	if info.Disk != nil {
		v.Disk = make([]DiskDetails, 0, len(info.Disk))
		for _, d := range info.Disk {
			v.Disk = append(v.Disk, DiskDetails{
				MountPoint: d.MountPoint,
				Path:       d.Path,
				Total:      d.Total,
				Avail:      d.Available,
			})
		}
	}

	return v
}

func (m *MachineDetails) Marshal(f string) ([]byte, error) {
	switch f {
	case FormatJSON:
		jsonString, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return nil, err
		}
		jsonString = append(jsonString, '\n')
		return jsonString, nil
	case FormatPlain:
		return m.marshalPlain()
	default:
		return nil, fmt.Errorf("unknown format %q (choices: plain, json)", f)
	}
}

func (m *MachineDetails) marshalPlain() ([]byte, error) {
	var b bytes.Buffer
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(2)
	if err := enc.Encode(m); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func FormatBytes(b uint64) any {
	const (
		mib = 1024 * 1024
		gib = 1024 * mib
		tib = 1024 * gib
	)
	switch {
	case b >= tib:
		return fmt.Sprintf("%.1fT", float64(b)/tib)
	case b >= gib:
		return fmt.Sprintf("%.1fG", float64(b)/gib)
	case b >= mib:
		return fmt.Sprintf("%.1fM", float64(b)/mib)
	default:
		return b
	}
}

func newPciAdditionalDeviceProperties(props map[string]string) *PciAdditionalDeviceProperties {
	if len(props) == 0 {
		return nil
	}

	ap := &PciAdditionalDeviceProperties{
		Microarchitecture: props["microarchitecture"],
		ComputeCapability: props["compute-capability"],
	}
	if v, ok := props["vram"]; ok {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			ap.Vram = n
		}
	}
	if *ap == (PciAdditionalDeviceProperties{}) {
		return nil
	}
	return ap
}

// HexInt Custom type to handle hex values
type HexInt int

// UnmarshalYAML parses a hex string into an int
func (hi *HexInt) UnmarshalYAML(value *yaml.Node) error {
	// Ignore empty string
	if value.Value == "" {
		return nil
	}

	if value.Kind != yaml.ScalarNode {
		return fmt.Errorf("expected a scalar node, got %v", value.LongTag())
	}

	// Strip 0x prefix if it exists
	hexString := strings.TrimPrefix(value.Value, "0x")

	// Parse the hex string to int
	parsed, err := strconv.ParseInt(hexString, 16, 64)
	if err != nil {
		return fmt.Errorf("failed to parse hex value %s: %w", value.Value, err)
	}

	*hi = HexInt(parsed)
	return nil
}

func (hi HexInt) MarshalYAML() (interface{}, error) {
	return fmt.Sprintf("0x%X", hi), nil
}

func (hi *HexInt) UnmarshalJSON(data []byte) error {
	// Remove quotes
	hexString := strings.Trim(string(data), "\"")

	// Ignore empty string
	if hexString == "" {
		return nil
	}

	// Remove "0x" prefix if present
	hexString = strings.TrimPrefix(hexString, "0x")

	// Parse as base 16 integer
	val, err := strconv.ParseInt(hexString, 16, 64)
	if err != nil {
		return fmt.Errorf("failed to parse hex value %s: %w", hexString, err)
	}
	*hi = HexInt(val)
	return nil
}

func (hi HexInt) MarshalJSON() ([]byte, error) {
	hexString := fmt.Sprintf("\"0x%X\"", int(hi))
	return []byte(hexString), nil
}
