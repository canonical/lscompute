package cpu

import "github.com/canonical/lscompute/pkg/machine/types"

type CpuInfo struct {
	Architecture string `json:"architecture" yaml:"architecture"`

	// amd64
	ManufacturerId string   `json:"manufacturer-id,omitempty" yaml:"manufacturer-id,omitempty"`
	Flags          []string `json:"flags,omitempty" yaml:"flags,omitempty"`

	// arm64
	ImplementerId types.HexInt `json:"implementer-id,omitempty" yaml:"implementer-id,omitempty"`
	PartNumber    types.HexInt `json:"part-number,omitempty" yaml:"part-number,omitempty"`
	Features      []string     `json:"features,omitempty" yaml:"features,omitempty"`
}

// procCpuInfo contains general information about a system CPU found in /proc/cpuinfo.
type procCpuInfo struct {
	Processor    int64 // %d - kernel defines it as long long
	Architecture string

	// amd64
	ManufacturerId string
	BrandString    string
	Flags          []string

	// arm64
	ModelName     *string  // %s
	BogoMips      float64  // %lu.%02lu
	Features      []string // space separated strings
	ImplementerId uint64   // 0x%02x
	//Architecture  uint64   // constant int
	Variant    uint64 // 0x%x
	PartNumber uint64 // 0x%03x
	Revision   uint64 // %d

}
