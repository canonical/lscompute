package apusys

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/canonical/lscompute/pkg/machine/host"
)

func TestDevicesDetectsAPUSYSDevice(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "dev", "apusys"), "")
	writeFile(t, filepath.Join(root, compatiblePath), "mediatek,mt8195-evb\x00mediatek,mt8195\x00")

	bus := NewBus(host.Fake(root), Options{})
	got, warnings, err := bus.Devices()
	if err != nil {
		t.Fatalf("Devices() error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("Devices() warnings = %v, want none", warnings)
	}

	want := []Device{
		{
			Bus:        BusName,
			Type:       "mdla",
			SocID:      "mt8195",
			VendorName: "mediatek",
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Devices() = %#v, want %#v", got, want)
	}
}

func TestDevicesDetectsAPUSYSDeviceWithoutCompatible(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "dev", "apusys0"), "")

	bus := NewBus(host.Fake(root), Options{})
	got, warnings, err := bus.Devices()
	if err != nil {
		t.Fatalf("Devices() error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("Devices() warnings = %v, want none", warnings)
	}

	want := []Device{
		{Bus: BusName, Type: "mdla"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Devices() = %#v, want %#v", got, want)
	}
}

func TestDevicesMissingAPUSYSDevice(t *testing.T) {
	bus := NewBus(host.Fake(t.TempDir()), Options{})
	got, warnings, err := bus.Devices()
	if err != nil {
		t.Fatalf("Devices() error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("Devices() warnings = %v, want none", warnings)
	}
	if len(got) != 0 {
		t.Fatalf("Devices() = %#v, want no devices", got)
	}
}

func TestDevicesIgnoresLegacyMDLADevfreqNode(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "sys", "devices", "platform", "soc", "soc:mdla_devfreq", "devfreq", "soc:mdla_devfreq"), "")

	bus := NewBus(host.Fake(root), Options{})
	got, warnings, err := bus.Devices()
	if err != nil {
		t.Fatalf("Devices() error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("Devices() warnings = %v, want none", warnings)
	}
	if len(got) != 0 {
		t.Fatalf("Devices() = %#v, want no devices", got)
	}
}

func TestParseCompatible(t *testing.T) {
	tests := []struct {
		name       string
		raw        string
		wantVendor string
		wantSocID  string
	}{
		{
			name:       "board and soc compatibles",
			raw:        "mediatek,mt8195-evb\x00mediatek,mt8195\x00",
			wantVendor: "mediatek",
			wantSocID:  "mt8195",
		},
		{
			name:       "uppercase model",
			raw:        "mediatek,MT8188\x00",
			wantVendor: "mediatek",
			wantSocID:  "mt8188",
		},
		{
			name:       "ignores malformed entries",
			raw:        "not-a-compatible\x00mediatek,mt8370\x00",
			wantVendor: "mediatek",
			wantSocID:  "mt8370",
		},
		{
			name: "empty",
			raw:  "\x00",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotVendor, gotSocID := parseCompatible([]byte(tc.raw))
			if gotVendor != tc.wantVendor || gotSocID != tc.wantSocID {
				t.Fatalf("parseCompatible() = (%q, %q), want (%q, %q)", gotVendor, gotSocID, tc.wantVendor, tc.wantSocID)
			}
		})
	}
}

func writeFile(t *testing.T, path, data string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error: %v", err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}
}
