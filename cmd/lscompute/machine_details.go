package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/canonical/lscompute/pkg/machine"
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
	Flags          []string `json:"flags,omitempty" yaml:"flags,omitempty"`

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

func (cpu CpuDetails) MarshalYAML() (any, error) {
	// Marshal all fields the standard way, then render only the flags
	// sequence inline (flow style). The local alias prevents this method
	// from recursing when the struct is encoded.
	type cpuDetails CpuDetails
	var node yaml.Node
	if err := node.Encode(cpuDetails(cpu)); err != nil {
		return nil, err
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == "flags" {
			node.Content[i+1].Style = yaml.FlowStyle
		}
	}
	return &node, nil
}

func NewMachineDetails(info *machine.MachineInfo) *MachineDetails {
	if info == nil {
		return nil
	}

	v := &MachineDetails{
		Memory:  MemoryDetails(info.Memory),
		Devices: info.Devices,
	}

	if info.Cpus != nil {
		v.Cpus = make([]CpuDetails, len(info.Cpus))
		for i, c := range info.Cpus {
			v.Cpus[i] = CpuDetails{
				Architecture:   c.Architecture,
				ManufacturerId: c.ManufacturerId,
				Flags:          c.Flags,
				ImplementerId:  c.ImplementerId,
				PartNumber:     c.PartNumber,
				Features:       c.Features,
				Isa:            c.Isa,
			}
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

func (m MachineDetails) Marshal(f string) ([]byte, error) {
	switch f {
	case FormatJSON:
		jsonString, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return nil, err
		}
		jsonString = append(jsonString, '\n')
		return jsonString, nil
	case FormatPlain:
		return marshalPlain(&m)
	default:
		return nil, fmt.Errorf("unknown format %q (choices: plain, json)", f)
	}
}

func marshalPlain(m *MachineDetails) ([]byte, error) {
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
