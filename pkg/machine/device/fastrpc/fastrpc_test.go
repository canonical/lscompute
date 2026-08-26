package fastrpc

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/canonical/lscompute/pkg/machine/host"
)

func TestDevicesDetectsFastRPCNodes(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{
		"FASTRPC-adsp",
		"fastrpc-cdsp1",
		"fastrpc-cdsp1-secure",
		"fastrpc-gdsp2-secure",
		"not-fastrpc-adsp",
		"fastrpc-unknown0",
	} {
		writeFile(t, filepath.Join(root, "dev", name), "")
	}

	bus := NewBus(host.Fake(root), Options{})
	got, warnings, err := bus.Devices()
	if err != nil {
		t.Fatalf("Devices() error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("Devices() warnings = %v, want none", warnings)
	}

	want := []Device{
		{Bus: BusName, Domain: ADSPDomain},
		{Bus: BusName, Domain: CDSPDomain, Index: 1},
		{Bus: BusName, Domain: CDSPDomain, Index: 1, Secure: true},
		{Bus: BusName, Domain: GDSPDomain, Index: 2, Secure: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Devices() = %#v, want %#v", got, want)
	}
}

func TestDevicesMissingDevDirectory(t *testing.T) {
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

func TestDevicesReadDirError(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "dev"), "not a directory")

	bus := NewBus(host.Fake(root), Options{})
	_, _, err := bus.Devices()
	if err == nil {
		t.Fatal("Devices() error = nil, want an error when dev is not a directory")
	}
}

func TestParseFastRPCDeviceName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     Device
		wantOkay bool
	}{
		{
			name:     "default adsp",
			input:    "fastrpc-adsp",
			want:     Device{Domain: ADSPDomain},
			wantOkay: true,
		},
		{
			name:     "indexed cdsp",
			input:    "fastrpc-cdsp1",
			want:     Device{Domain: CDSPDomain, Index: 1},
			wantOkay: true,
		},
		{
			name:     "secure indexed cdsp",
			input:    "fastrpc-cdsp1-secure",
			want:     Device{Domain: CDSPDomain, Index: 1, Secure: true},
			wantOkay: true,
		},
		{
			name:     "case insensitive",
			input:    "FASTRPC-MDSP2-SECURE",
			want:     Device{Domain: MDSPDomain, Index: 2, Secure: true},
			wantOkay: true,
		},
		{
			name:  "missing prefix",
			input: "adsp",
		},
		{
			name:  "unknown domain",
			input: "fastrpc-unknown0",
		},
		{
			name:  "malformed secure suffix",
			input: "fastrpc-adsp-secure-extra",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, okay := parseFastRPCDeviceName(tc.input)
			if okay != tc.wantOkay {
				t.Fatalf("parseFastRPCDeviceName(%q) okay = %v, want %v", tc.input, okay, tc.wantOkay)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("parseFastRPCDeviceName(%q) = %#v, want %#v", tc.input, got, tc.want)
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
