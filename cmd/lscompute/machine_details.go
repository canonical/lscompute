package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/canonical/lscompute/pkg/machine"
	"github.com/canonical/lscompute/pkg/machine/device/fastrpc"
	"github.com/canonical/lscompute/pkg/machine/device/pci"
	"github.com/canonical/lscompute/pkg/machine/device/usb"
	"github.com/canonical/lscompute/pkg/machine/types"
	"go.yaml.in/yaml/v4"
)

const (
	// FormatPlain is a human-readable YAML output.
	FormatPlain string = "plain"
	// FormatJSON is a machine-readable, indented JSON output with kebab-cased keys.
	FormatJSON string = "json"
)

type MachineDetails struct {
	Cpus    []CpuDetails          `json:"cpus,omitempty" yaml:"cpus,omitempty"`
	Memory  MemoryDetails         `json:"memory,omitempty" yaml:"memory,omitempty"`
	Disk    map[string]DirDetails `json:"disk,omitempty" yaml:"disk,omitempty"`
	Devices []any                 `json:"devices,omitempty" yaml:"devices,omitempty"`
}

type CpuDetails struct {
	Architecture string `json:"architecture" yaml:"architecture"`

	// amd64
	ManufacturerId string   `json:"manufacturer-id,omitempty" yaml:"manufacturer-id,omitempty"`
	Flags          []string `json:"flags,omitempty" yaml:"flags,flow,omitempty"`

	// arm64
	ImplementerId types.HexInt `json:"implementer-id,omitempty" yaml:"implementer-id,omitempty"`
	PartNumber    types.HexInt `json:"part-number,omitempty" yaml:"part-number,omitempty"`
	Features      []string     `json:"features,omitempty" yaml:"features,omitempty"`

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

type DirDetails struct {
	Total uint64 `json:"total" yaml:"total"`
	Avail uint64 `json:"avail" yaml:"avail"`
}

func (d DirDetails) MarshalYAML() (any, error) {
	return struct {
		Total any `yaml:"total"`
		Avail any `yaml:"avail"`
	}{
		Total: FormatBytes(d.Total),
		Avail: FormatBytes(d.Avail),
	}, nil
}

type PciDeviceDetails struct {
	Bus                  string                      `json:"bus" yaml:"bus"`
	Slot                 string                      `json:"slot,omitempty" yaml:"slot,omitempty"`
	BusNumber            any                         `json:"bus-number,omitempty" yaml:"bus-number,omitempty"`
	DeviceClass          types.HexInt                `json:"device-class,omitempty" yaml:"device-class,omitempty"`
	VendorId             types.HexInt                `json:"vendor-id,omitempty" yaml:"vendor-id,omitempty"`
	DeviceId             types.HexInt                `json:"device-id,omitempty" yaml:"device-id,omitempty"`
	SubvendorId          types.HexInt                `json:"subvendor-id,omitempty" yaml:"subvendor-id,omitempty"`
	SubdeviceId          types.HexInt                `json:"subdevice-id,omitempty" yaml:"subdevice-id,omitempty"`
	VendorName           string                      `json:"vendor-name,omitempty" yaml:"vendor-name,omitempty"`
	DeviceName           string                      `json:"device-name,omitempty" yaml:"device-name,omitempty"`
	SubvendorName        string                      `json:"subvendor-name,omitempty" yaml:"subvendor-name,omitempty"`
	SubdeviceName        string                      `json:"subdevice-name,omitempty" yaml:"subdevice-name,omitempty"`
	AdditionalProperties *AdditionalDeviceProperties `json:"additional-properties,omitempty" yaml:"additional-properties,omitempty"`
}

type UsbDeviceDetails struct {
	Bus                  string                      `json:"bus" yaml:"bus"`
	BusNumber            any                         `json:"bus-number,omitempty" yaml:"bus-number,omitempty"`
	DeviceNumber         int                         `json:"device-number,omitempty" yaml:"device-number,omitempty"`
	VendorId             types.HexInt                `json:"vendor-id,omitempty" yaml:"vendor-id,omitempty"`
	ProductId            types.HexInt                `json:"product-id,omitempty" yaml:"product-id,omitempty"`
	VendorName           string                      `json:"vendor-name,omitempty" yaml:"vendor-name,omitempty"`
	ProductName          string                      `json:"product-name,omitempty" yaml:"product-name,omitempty"`
	AdditionalProperties *AdditionalDeviceProperties `json:"additional-properties,omitempty" yaml:"additional-properties,omitempty"`
}

type FastRPCDeviceDetails struct {
	Bus                  string                      `json:"bus" yaml:"bus"`
	Domain               string                      `json:"domain,omitempty" yaml:"domain,omitempty"`
	Index                int                         `json:"index,omitempty" yaml:"index,omitempty"`
	Secure               bool                        `json:"secure,omitempty" yaml:"secure,omitempty"`
	AdditionalProperties *AdditionalDeviceProperties `json:"additional-properties,omitempty" yaml:"additional-properties,omitempty"`
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

type UsbAdditionalDeviceProperties struct {
	Properties map[string]string
}

type FastRPCAdditionalDeviceProperties struct {
	Properties map[string]string
}
type AdditionalDeviceProperties struct {
	pci     *PciAdditionalDeviceProperties
	usb     *UsbAdditionalDeviceProperties
	fastrpc *FastRPCAdditionalDeviceProperties
}

func (a AdditionalDeviceProperties) MarshalJSON() ([]byte, error) {
	switch {
	case a.pci != nil:
		return json.Marshal(a.pci)
	case a.usb != nil:
		return json.Marshal(a.usb.Properties)
	case a.fastrpc != nil:
		return json.Marshal(a.fastrpc.Properties)
	default:
		return json.Marshal(struct{}{})
	}
}

func (a AdditionalDeviceProperties) MarshalYAML() (any, error) {
	switch {
	case a.pci != nil:
		return a.pci.MarshalYAML()
	case a.usb != nil:
		return a.usb.Properties, nil
	case a.fastrpc != nil:
		return a.fastrpc.Properties, nil
	default:
		return struct{}{}, nil
	}
}

func NewMachineDetails(info *machine.MachineInfo) *MachineDetails {
	if info == nil {
		return nil
	}

	v := &MachineDetails{
		Memory: MemoryDetails(info.Memory),
	}

	if info.Devices != nil {
		v.Devices = make([]any, len(info.Devices))
		for i, d := range info.Devices {
			switch typed := d.(type) {
			case pci.Device:
				v.Devices[i] = PciDeviceDetails{
					Bus:                  typed.Bus,
					Slot:                 typed.Slot,
					BusNumber:            typed.BusNumber,
					DeviceClass:          typed.DeviceClass,
					VendorId:             typed.VendorId,
					DeviceId:             typed.DeviceId,
					SubvendorId:          derefHexInt(typed.SubvendorId),
					SubdeviceId:          derefHexInt(typed.SubdeviceId),
					VendorName:           derefString(typed.VendorName),
					DeviceName:           derefString(typed.DeviceName),
					SubvendorName:        derefString(typed.SubvendorName),
					SubdeviceName:        derefString(typed.SubdeviceName),
					AdditionalProperties: newAdditionalDeviceProperties(typed.Bus, typed.AdditionalProperties),
				}
			case usb.Device:
				v.Devices[i] = UsbDeviceDetails{
					Bus:                  typed.Bus,
					BusNumber:            typed.BusNumber,
					DeviceNumber:         typed.DeviceNumber,
					VendorId:             typed.VendorId,
					ProductId:            typed.ProductId,
					VendorName:           derefString(typed.VendorName),
					ProductName:          derefString(typed.ProductName),
					AdditionalProperties: newAdditionalDeviceProperties(typed.Bus, typed.AdditionalProperties),
				}
			case fastrpc.Device:
				v.Devices[i] = FastRPCDeviceDetails{
					Bus:                  typed.Bus,
					Domain:               string(typed.Domain),
					Index:                typed.Index,
					Secure:               typed.Secure,
					AdditionalProperties: newAdditionalDeviceProperties(typed.Bus, typed.AdditionalProperties),
				}
			default:
				continue
			}
		}
	}

	if info.Cpus != nil {
		v.Cpus = make([]CpuDetails, len(info.Cpus))
		for i, c := range info.Cpus {
			v.Cpus[i] = CpuDetails(c)
		}
	}

	if info.Disk != nil {
		v.Disk = make(map[string]DirDetails, len(info.Disk))
		for path, d := range info.Disk {
			v.Disk[path] = DirDetails(d)
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

func newAdditionalDeviceProperties(bus string, props map[string]string) *AdditionalDeviceProperties {
	if len(props) == 0 {
		return nil
	}

	switch bus {
	case "pci":
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
		return &AdditionalDeviceProperties{pci: ap}
	case "usb":
		return &AdditionalDeviceProperties{usb: &UsbAdditionalDeviceProperties{Properties: props}}
	case "fastrpc":
		return &AdditionalDeviceProperties{fastrpc: &FastRPCAdditionalDeviceProperties{Properties: props}}
	default:
		return nil
	}
}

// derefString returns the pointed-to string, or "" when the pointer is nil.
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// derefHexInt returns the pointed-to value, or the zero value when the
// pointer is nil.
func derefHexInt(h *types.HexInt) types.HexInt {
	if h == nil {
		return 0
	}
	return *h
}
