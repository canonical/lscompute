package pci

import (
	"errors"
	"testing"

	"github.com/canonical/lscompute/pkg/machine/constants"
	"github.com/canonical/lscompute/pkg/machine/host"
	"github.com/canonical/lscompute/pkg/machine/types"
)

// TestAdditionalProperties_UnknownVendor verifies that an unrecognised vendor
// returns ErrorVendorNotSupported.
func TestAdditionalProperties_UnknownVendor(t *testing.T) {
	h := host.Fake(t.TempDir())
	dev := Device{VendorId: types.HexInt(0xffff), DeviceClass: types.HexInt(0x0300)}
	_, err := additionalProperties(h, dev)
	if !errors.Is(err, ErrorVendorNotSupported) {
		t.Errorf("expected ErrorVendorNotSupported, got %v", err)
	}
}

// TestAdditionalProperties_NvidiaNotGpu verifies that a non-GPU NVIDIA device
// returns nil properties without error.
func TestAdditionalProperties_NvidiaNotGpu(t *testing.T) {
	h := host.Fake(t.TempDir())
	dev := Device{
		VendorId:    types.HexInt(constants.PciVendorNvidia),
		DeviceClass: types.HexInt(0x0200), // network — not a GPU
	}
	props, err := additionalProperties(h, dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if props != nil {
		t.Errorf("expected nil properties for non-GPU device, got %v", props)
	}
}

// TestAdditionalProperties_IntelNotGpu verifies that a non-GPU Intel device
// returns nil properties without error.
func TestAdditionalProperties_IntelNotGpu(t *testing.T) {
	h := host.Fake(t.TempDir())
	dev := Device{
		VendorId:    types.HexInt(constants.PciVendorIntel),
		DeviceClass: types.HexInt(0x0c03), // USB host — not a GPU
	}
	props, err := additionalProperties(h, dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if props != nil {
		t.Errorf("expected nil properties for non-GPU device, got %v", props)
	}
}

// TestAdditionalProperties_AmdNotGpu verifies that a non-GPU AMD device
// returns nil properties without error.
func TestAdditionalProperties_AmdNotGpu(t *testing.T) {
	h := host.Fake(t.TempDir())
	dev := Device{
		VendorId:    types.HexInt(constants.PciVendorAmd),
		DeviceClass: types.HexInt(0x0200), // network — not a GPU
	}
	props, err := additionalProperties(h, dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if props != nil {
		t.Errorf("expected nil properties for non-GPU device, got %v", props)
	}
}

// TestAddAdditionalProperties_UnknownVendorSilent verifies that unknown-vendor
// devices produce no warnings (ErrorVendorNotSupported is swallowed silently).
func TestAddAdditionalProperties_UnknownVendorSilent(t *testing.T) {
	h := host.Fake(t.TempDir())
	devices := []Device{
		{VendorId: types.HexInt(0xffff), DeviceClass: types.HexInt(0x0300)},
		{VendorId: types.HexInt(0x1234), DeviceClass: types.HexInt(0x0200)},
	}
	_, warnings := addAdditionalProperties(h, devices)
	if len(warnings) != 0 {
		t.Errorf("expected no warnings for unknown-vendor devices, got: %v", warnings)
	}
}

// TestAddAdditionalProperties_GpuError verifies that GPU lookup failures for
// supported vendors are collected as warnings (not returned as errors).
func TestAddAdditionalProperties_GpuError(t *testing.T) {
	// The fake host has no nvidia-smi / clinfo / AMD sysfs fixtures, so
	// looking up properties for a GPU will fail → should become a warning.
	h := host.Fake(t.TempDir())
	devices := []Device{
		{
			VendorId:    types.HexInt(constants.PciVendorNvidia),
			DeviceClass: types.HexInt(0x0300), // GPU
			Slot:        "0000:01:00.0",
		},
	}
	_, warnings := addAdditionalProperties(h, devices)
	if len(warnings) == 0 {
		t.Error("expected at least one warning for failed GPU property lookup, got none")
	}
}
