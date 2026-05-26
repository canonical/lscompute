package amd

import (
	"github.com/canonical/lscompute/pkg/machine/host"
	"github.com/canonical/lscompute/pkg/machine/types"
	"testing"
)

func TestVRam(t *testing.T) {
	tests := []struct {
		name        string
		device      types.PciDevice
		machineRoot string
		expected    uint64
		shouldErr   bool
	}{
		{name: "valid vram read hp-zbook", device: types.PciDevice{Slot: "0000:03:00.0"}, machineRoot: "../../../../test_data/machines/hp-zbook-i712850HX+RadeonPROW6600M/machine-root", expected: 8573157376, shouldErr: false},
		{name: "invalid path hp-zbook", device: types.PciDevice{Slot: "9999:99:99.9"}, machineRoot: "../../../../test_data/machines/hp-zbook-i712850HX+RadeonPROW6600M/machine-root", shouldErr: true},
		{name: "valid vram read lenovo", device: types.PciDevice{Slot: "0000:c4:00.0"}, machineRoot: "../../../../test_data/machines/lenovo-thinkpad-p16s/machine-root", expected: 8589934592, shouldErr: false},
		{name: "invalid path lenovo", device: types.PciDevice{Slot: "9999:99:99.9"}, machineRoot: "../../../../test_data/machines/lenovo-thinkpad-p16s/machine-root", shouldErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := host.Fake(tt.machineRoot)
			got, err := vRam(h, tt.device)
			if tt.shouldErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatalf("expected non-nil vram value")
			}
			if *got != tt.expected {
				t.Fatalf("expected %d, got %d", tt.expected, *got)
			}
		})
	}
}
func TestGetAmdGpuPciSlot(t *testing.T) {
	tests := []struct {
		name, input, machineRoot, expected, errContains string
		shouldErr                                       bool
	}{
		{name: "valid hp-zbook render 129", input: "drm_render_minor 129", machineRoot: "../../../../test_data/machines/hp-zbook-i712850HX+RadeonPROW6600M/machine-root", expected: "0000:03:00.0"},
		{name: "invalid format hp-zbook missing value", input: "drm_render_minor", machineRoot: "../../../../test_data/machines/hp-zbook-i712850HX+RadeonPROW6600M/machine-root", shouldErr: true, errContains: "unexpected format for drm_render_minor"},
		{name: "invalid format hp-zbook too many parts", input: "drm_render_minor 128 extra", machineRoot: "../../../../test_data/machines/hp-zbook-i712850HX+RadeonPROW6600M/machine-root", shouldErr: true, errContains: "unexpected format for drm_render_minor"},
		{name: "invalid symlink hp-zbook", input: "drm_render_minor 999", machineRoot: "../../../../test_data/machines/hp-zbook-i712850HX+RadeonPROW6600M/machine-root", shouldErr: true},
		{name: "valid lenovo render 128", input: "drm_render_minor 128", machineRoot: "../../../../test_data/machines/lenovo-thinkpad-p16s/machine-root", expected: "0000:c4:00.0"},
		{name: "invalid format lenovo missing value", input: "drm_render_minor", machineRoot: "../../../../test_data/machines/lenovo-thinkpad-p16s/machine-root", shouldErr: true, errContains: "unexpected format for drm_render_minor"},
		{name: "invalid format lenovo too many parts", input: "drm_render_minor 128 extra", machineRoot: "../../../../test_data/machines/lenovo-thinkpad-p16s/machine-root", shouldErr: true, errContains: "unexpected format for drm_render_minor"},
		{name: "invalid symlink lenovo", input: "drm_render_minor 999", machineRoot: "../../../../test_data/machines/lenovo-thinkpad-p16s/machine-root", shouldErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := host.Fake(tt.machineRoot)
			got, err := getAmdGpuPciSlot(h, tt.input)
			if tt.shouldErr {
				if err == nil {
					t.Fatalf("expected error, got nil (result: %q)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
func TestGetGfxTargetVersion(t *testing.T) {
	tests := []struct {
		name, input, expected, errContains string
		expectFailure                      bool
	}{
		{name: "valid", input: "gfx_target_version 110502", expected: "gfx1152"},
		{name: "zero value", input: "gfx_target_version 0", errContains: "gfx_target_version is invalid", expectFailure: true},
		{name: "missing value", input: "gfx_target_version", errContains: "unexpected format", expectFailure: true},
		{name: "non-numeric major", input: "gfx_target_version ab1234", errContains: "invalid syntax", expectFailure: true},
		{name: "non-numeric minor", input: "gfx_target_version 12ab34", errContains: "invalid syntax", expectFailure: true},
		{name: "non-numeric revision", input: "gfx_target_version 1234ab", errContains: "invalid syntax", expectFailure: true},
		{name: "too short", input: "gfx_target_version 12345", errContains: "unexpected format", expectFailure: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseGfxTargetVersion(tt.input)
			t.Logf("input=%q got=%q err=%v", tt.input, got, err)
			if tt.expectFailure {
				if err == nil {
					t.Fatalf("expected error, got nil (result: %q)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
func TestGpuProperties(t *testing.T) {
	tests := []struct {
		name, machineRoot                    string
		device                               types.PciDevice
		shouldErr, checkVram, checkMicroArch bool
	}{
		{name: "hp-zbook AMD GPU", device: types.PciDevice{Slot: "0000:03:00.0"}, machineRoot: "../../../../test_data/machines/hp-zbook-i712850HX+RadeonPROW6600M/machine-root", checkVram: true, checkMicroArch: true},
		{name: "hp-zbook invalid slot", device: types.PciDevice{Slot: "9999:99:99.9"}, machineRoot: "../../../../test_data/machines/hp-zbook-i712850HX+RadeonPROW6600M/machine-root", shouldErr: true},
		{name: "lenovo AMD GPU", device: types.PciDevice{Slot: "0000:c4:00.0"}, machineRoot: "../../../../test_data/machines/lenovo-thinkpad-p16s/machine-root", checkVram: true, checkMicroArch: true},
		{name: "lenovo invalid slot", device: types.PciDevice{Slot: "9999:99:99.9"}, machineRoot: "../../../../test_data/machines/lenovo-thinkpad-p16s/machine-root", shouldErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := host.Fake(tt.machineRoot)
			props, err := gpuProperties(h, tt.device)
			if tt.shouldErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if v, ok := props["vram"]; !ok || v == "" {
				t.Fatalf("expected vram property")
			}
		})
	}
}
func TestGfxArchitecture(t *testing.T) {
	tests := []struct {
		name, machineRoot, expected, errContains string
		device                                   types.PciDevice
		shouldErr                                bool
	}{
		{name: "valid hp-zbook", device: types.PciDevice{Slot: "0000:03:00.0"}, machineRoot: "../../../../test_data/machines/hp-zbook-i712850HX+RadeonPROW6600M/machine-root", expected: "gfx1032"},
		{name: "invalid nodes dir", device: types.PciDevice{Slot: "0000:03:00.0"}, machineRoot: "/nonexistent/path/", shouldErr: true},
		{name: "no match hp-zbook", device: types.PciDevice{Slot: "9999:99:99.9"}, machineRoot: "../../../../test_data/machines/hp-zbook-i712850HX+RadeonPROW6600M/machine-root", shouldErr: true},
		{name: "valid lenovo", device: types.PciDevice{Slot: "0000:c4:00.0"}, machineRoot: "../../../../test_data/machines/lenovo-thinkpad-p16s/machine-root", expected: "gfx1152"},
		{name: "no match lenovo", device: types.PciDevice{Slot: "9999:99:99.9"}, machineRoot: "../../../../test_data/machines/lenovo-thinkpad-p16s/machine-root", shouldErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := host.Fake(tt.machineRoot)
			got, err := gfxArchitecture(h, tt.device)
			if tt.shouldErr {
				if err == nil {
					t.Fatalf("expected error, got nil (result: %q)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
