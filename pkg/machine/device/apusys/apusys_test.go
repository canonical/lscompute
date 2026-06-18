package apusys

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/canonical/lscompute/pkg/machine/host"
)

func TestDevicesDetectsMDLADevfreq(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, mdlaDevfreqDir)
	if err := os.MkdirAll(filepath.Join(path, "power"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error: %v", err)
	}
	writeFile(t, filepath.Join(path, "cur_freq"), "700000000\n")
	writeFile(t, filepath.Join(path, "available_frequencies"), "273000000 550000000 700000000\n")
	writeFile(t, filepath.Join(path, "power/runtime_status"), "active\n")
	writeFile(t, filepath.Join(root, socIDPath), "jep106:0426:8195\n")

	bus := NewBus(host.Fake(root), Options{})
	got, warnings, err := bus.Devices()
	if err != nil {
		t.Fatalf("Devices() error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("Devices() warnings = %v, want none", warnings)
	}

	want := []any{
		Device{
			Bus:             BusName,
			Type:            "mdla",
			SocID:           "jep106:0426:8195",
			SocInfo: SocInfo{
				ChipModel:       "MT8195 / MT8395",
				ProductFamily:   "Kompanio 1200 / Genio 1200",
				NPUArchitecture: "4.0 TOPS (Dual-core APU)",
			},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Devices() = %#v, want %#v", got, want)
	}
}

func TestDevicesDetectsMDLADevfreqWithoutSoCID(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, mdlaDevfreqDir), 0o755); err != nil {
		t.Fatalf("MkdirAll() error: %v", err)
	}

	bus := NewBus(host.Fake(root), Options{})
	got, warnings, err := bus.Devices()
	if err != nil {
		t.Fatalf("Devices() error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("Devices() warnings = %v, want none", warnings)
	}

	want := []any{
		Device{Bus: BusName, Type: "mdla"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Devices() = %#v, want %#v", got, want)
	}
}

func TestDevicesMissingMDLADevfreq(t *testing.T) {
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

func TestDevicesIgnoresNonDirectoryMDLADevfreq(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, mdlaDevfreqDir), "not a directory\n")

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

func TestSoCIDInfo(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want Device
	}{
		{
			name: "jep106 mt8195",
			raw:  "jep106:0426:8195",
			want: Device{
				SocID:           "jep106:0426:8195",
				SocInfo: SocInfo{
					ChipModel:       "MT8195 / MT8395",
					ProductFamily:   "Kompanio 1200 / Genio 1200",
					NPUArchitecture: "4.0 TOPS (Dual-core APU)",
				},
			},
		},
		{
			name: "mt8188",
			raw:  "8188",
			want: Device{
				SocID:           "8188",
				SocInfo: SocInfo{
					ChipModel:       "MT8188 / MT8390",
					ProductFamily:   "Kompanio 520 / Genio 700",
					NPUArchitecture: "2.0 TOPS (Single-core APU)",
				},
			},
		},
		{
			name: "mt8370",
			raw:  "0x8370",
			want: Device{
				SocID:           "0x8370",
				SocInfo: SocInfo{
					ChipModel:       "MT8370",
					ProductFamily:   "Genio 510",
					NPUArchitecture: "1.0 TOPS (Single-core APU)",
				},
			},
		},
		{
			name: "mt8365",
			raw:  "mt8365",
			want: Device{
				SocID:           "mt8365",
				SocInfo: SocInfo{
					ChipModel:       "MT8365",
					ProductFamily:   "Genio 350",
					NPUArchitecture: "0.5 TOPS (Single-core APU)",
				},
			},
		},
		{
			name: "mt8192",
			raw:  "8192",
			want: Device{
				SocID:           "8192",
				SocInfo: SocInfo{
					ChipModel:       "MT8192",
					ProductFamily:   "Kompanio 828",
					NPUArchitecture: "Similar to MT8195 (NPU present)",
				},
			},
		},
		{
			name: "mt8186",
			raw:  "8186",
			want: Device{
				SocID:           "8186",
				SocInfo: SocInfo{
					ChipModel:       "MT8186",
					ProductFamily:   "Kompanio 528",
					NPUArchitecture: "Entry-level NPU",
				},
			},
		},
		{
			name: "unknown",
			raw:  "jep106:0426:9999",
			want: Device{SocID: "jep106:0426:9999"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Device{SocID: tc.raw}
			addSoCInfo(&got)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("addSoCInfo() = %#v, want %#v", got, tc.want)
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
