package cpu

import (
	"path/filepath"
	"testing"

	"github.com/canonical/lscompute/pkg/machine/host"
)

const xps13MachineRoot = "../../../test_data/machines/xps13-9350/machine-root"
const rpiMachineRoot = "../../../test_data/machines/raspberry-pi-5/machine-root"

func machineHost(t *testing.T, root string) host.Host {
	t.Helper()
	abs, err := filepath.Abs(root)
	if err != nil {
		t.Fatalf("resolving machine root: %v", err)
	}
	return host.Fake(abs)
}

// TestInfo_Amd64 exercises the full Info() pipeline on an x86_64 machine fixture.
func TestInfo_Amd64(t *testing.T) {
	h := machineHost(t, xps13MachineRoot)
	cpus, err := Info(h)
	if err != nil {
		t.Fatalf("Info() error: %v", err)
	}
	if len(cpus) == 0 {
		t.Fatal("expected at least one CPU, got none")
	}
	for _, c := range cpus {
		if c.Architecture != Amd64 {
			t.Errorf("Architecture = %q, want %q", c.Architecture, Amd64)
		}
		if c.ManufacturerId == "" {
			t.Errorf("ManufacturerId is empty for amd64 CPU")
		}
	}
}

// TestInfo_Arm64 exercises the full Info() pipeline on an aarch64 machine fixture.
func TestInfo_Arm64(t *testing.T) {
	h := machineHost(t, rpiMachineRoot)
	cpus, err := Info(h)
	if err != nil {
		t.Fatalf("Info() error: %v", err)
	}
	if len(cpus) == 0 {
		t.Fatal("expected at least one CPU, got none")
	}
	for _, c := range cpus {
		if c.Architecture != Arm64 {
			t.Errorf("Architecture = %q, want %q", c.Architecture, Arm64)
		}
	}
}

// TestInfo_MissingCpuInfo verifies that Info returns an error when proc/cpuinfo is absent.
func TestInfo_MissingCpuInfo(t *testing.T) {
	// Empty host — proc/cpuinfo is missing → Info must return an error.
	h := host.Fake(t.TempDir())
	_, err := Info(h)
	if err == nil {
		t.Fatal("expected error for missing proc/cpuinfo, got nil")
	}
}

// TestInfoFromRawData_UnsupportedArch verifies infoFromRawData rejects unknown architectures.
func TestInfoFromRawData_UnsupportedArch(t *testing.T) {
	_, err := infoFromRawData("processor\t: 0\n", "mips64")
	if err == nil {
		t.Fatal("expected error for unsupported architecture, got nil")
	}
}

// TestInfoFromRawData_Amd64 verifies infoFromRawData parses amd64 data correctly.
func TestInfoFromRawData_Amd64(t *testing.T) {
	cpus, err := infoFromRawData(amd64CpuInfoFixture, "x86_64")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cpus) == 0 {
		t.Fatal("expected at least one CPU")
	}
	if cpus[0].Architecture != Amd64 {
		t.Errorf("Architecture = %q, want %q", cpus[0].Architecture, Amd64)
	}
}

// TestInfoFromRawData_Arm64 verifies infoFromRawData parses arm64 data correctly.
func TestInfoFromRawData_Arm64(t *testing.T) {
	cpus, err := infoFromRawData(arm64CpuInfoFixture, "aarch64")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cpus) == 0 {
		t.Fatal("expected at least one CPU")
	}
	if cpus[0].Architecture != Arm64 {
		t.Errorf("Architecture = %q, want %q", cpus[0].Architecture, Arm64)
	}
}

// TestInfoFromRawData_EmptyCpuInfo verifies infoFromRawData fail-fast behavior
// when proc/cpuinfo content has no CPU entries.
func TestInfoFromRawData_EmptyCpuInfo(t *testing.T) {
	_, err := infoFromRawData("\n\n", "x86_64")
	if err == nil {
		t.Fatal("expected error for empty cpuinfo, got nil")
	}
}

// TestUniqueCpuInfo_Deduplication verifies that identical CPU entries are collapsed.
func TestUniqueCpuInfo_Deduplication(t *testing.T) {
	// Two identical cores (same flags/vendor) — should be collapsed to one.
	core := procCpuInfo{
		Architecture:   Amd64,
		ManufacturerId: "GenuineIntel",
		Flags:          []string{"sse", "sse2"},
	}
	result, err := uniqueCpuInfo([]procCpuInfo{core, core})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 unique CPU after dedup, got %d", len(result))
	}
}

// TestUniqueCpuInfo_DifferentKeepsAll verifies that distinct CPUs are all kept.
func TestUniqueCpuInfo_DifferentKeepsAll(t *testing.T) {
	a := procCpuInfo{Architecture: Amd64, ManufacturerId: "GenuineIntel", Flags: []string{"sse"}}
	b := procCpuInfo{Architecture: Amd64, ManufacturerId: "AuthenticAMD", Flags: []string{"sse"}}
	result, err := uniqueCpuInfo([]procCpuInfo{a, b})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 CPUs (both distinct), got %d", len(result))
	}
}

// TestIsDuplicate verifies the deep-equality check used by uniqueCpuInfo.
func TestIsDuplicate(t *testing.T) {
	a := procCpuInfo{Architecture: Amd64, ManufacturerId: "GenuineIntel"}
	b := procCpuInfo{Architecture: Amd64, ManufacturerId: "GenuineIntel"}
	c := procCpuInfo{Architecture: Amd64, ManufacturerId: "AuthenticAMD"}

	if !isDuplicate(a, b) {
		t.Error("isDuplicate(a, b): expected true for identical structs")
	}
	if isDuplicate(a, c) {
		t.Error("isDuplicate(a, c): expected false for different structs")
	}
}

// TestCpuInfoFromProc_UnsupportedArch verifies that an unknown architecture returns an error.
func TestCpuInfoFromProc_UnsupportedArch(t *testing.T) {
	_, err := cpuInfoFromProc([]procCpuInfo{{Architecture: "mips64"}})
	if err == nil {
		t.Fatal("expected error for unsupported architecture, got nil")
	}
}

// TestCpuInfoFromProc_Amd64 spot-checks the amd64 field mapping.
func TestCpuInfoFromProc_Amd64(t *testing.T) {
	pci := procCpuInfo{
		Architecture:   Amd64,
		ManufacturerId: "GenuineIntel",
		Flags:          []string{"sse", "avx"},
	}
	result, err := cpuInfoFromProc([]procCpuInfo{pci})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].ManufacturerId != "GenuineIntel" {
		t.Errorf("ManufacturerId = %q, want %q", result[0].ManufacturerId, "GenuineIntel")
	}
	if len(result[0].Flags) != 2 {
		t.Errorf("Flags len = %d, want 2", len(result[0].Flags))
	}
}

// TestCpuInfoFromProc_Arm64 spot-checks the arm64 field mapping.
func TestCpuInfoFromProc_Arm64(t *testing.T) {
	pci := procCpuInfo{
		Architecture:  Arm64,
		ImplementerId: 0x41,
		PartNumber:    0xd0b,
		Features:      []string{"fp", "asimd"},
	}
	result, err := cpuInfoFromProc([]procCpuInfo{pci})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].Architecture != Arm64 {
		t.Errorf("Architecture = %q, want %q", result[0].Architecture, Arm64)
	}
	if int(result[0].ImplementerId) != 0x41 {
		t.Errorf("ImplementerId = %#x, want 0x41", int(result[0].ImplementerId))
	}
}
