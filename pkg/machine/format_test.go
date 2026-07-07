package machine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/canonical/lscompute/pkg/machine/host"
)

var testMachines = []string{
	"xps13-9350",
	"i5-3570k+arc-a580+gtx1080ti",
	"raspberry-pi-5+hailo-8",
}

func machineInfoFromTestData(t *testing.T, machineName string) *MachineInfo {
	t.Helper()
	machineRoot := filepath.Join("..", "..", "test_data", "machines", machineName, "machine-root")

	pciIdsPath := filepath.Join(machineRoot, "usr", "share", "misc", "pci.ids")
	_, pciIdsErr := os.Stat(pciIdsPath)
	friendlyNames := pciIdsErr == nil

	info, _, err := Get(host.Fake(machineRoot), friendlyNames)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}
	return info
}

func TestMarshal_JSON(t *testing.T) {
	for _, machineName := range testMachines {
		t.Run(machineName, func(t *testing.T) {
			got, err := Marshal(machineInfoFromTestData(t, machineName), FormatJSON)
			if err != nil {
				t.Fatalf("Marshal(json) failed: %v", err)
			}

			var decoded MachineInfo
			if err := json.Unmarshal(got, &decoded); err != nil {
				t.Fatalf("output is not valid JSON: %v", err)
			}

			if !strings.Contains(string(got), "\n  \"memory\"") {
				t.Errorf("expected indented JSON, got:\n%s", got)
			}
			if !strings.Contains(string(got), "total-ram") {
				t.Errorf("expected kebab-cased keys, got:\n%s", got)
			}
		})
	}
}

func TestMarshal_Plain(t *testing.T) {
	for _, machineName := range testMachines {
		t.Run(machineName, func(t *testing.T) {
			got, err := Marshal(machineInfoFromTestData(t, machineName), FormatPlain)
			if err != nil {
				t.Fatalf("Marshal(plain) failed: %v", err)
			}
			out := string(got)

			for _, section := range []string{"CPUs:", "Memory:", "Disk:", "Devices:"} {
				if !strings.Contains(out, section) {
					t.Errorf("plain output missing section %q, got:\n%s", section, out)
				}
			}
			if !strings.Contains(out, "total-ram:") {
				t.Errorf("expected total-ram in output, got:\n%s", out)
			}
			if !strings.Contains(out, "bus: pci") {
				t.Errorf("expected pci device rendered, got:\n%s", out)
			}
		})
	}
}

func TestMarshal_Amd64MultiGpu(t *testing.T) {
	info := machineInfoFromTestData(t, "i5-3570k+arc-a580+gtx1080ti")

	for _, format := range []Format{FormatJSON, FormatPlain} {
		got, err := Marshal(info, format)
		if err != nil {
			t.Fatalf("Marshal(%s) failed: %v", format, err)
		}
		for _, want := range []string{"compute-capability", "vram", "manufacturer-id"} {
			if !strings.Contains(string(got), want) {
				t.Errorf("format %s: expected %q in output, got:\n%s", format, want, got)
			}
		}
	}
}

func TestMarshal_Arm64Npu(t *testing.T) {
	info := machineInfoFromTestData(t, "raspberry-pi-5+hailo-8")

	if len(info.Cpus) == 0 || info.Cpus[0].Architecture != "arm64" {
		t.Fatalf("expected arm64 CPU, got %+v", info.Cpus)
	}

	for _, format := range []Format{FormatJSON, FormatPlain} {
		got, err := Marshal(info, format)
		if err != nil {
			t.Fatalf("Marshal(%s) failed: %v", format, err)
		}
		for _, want := range []string{"part-number", "features", "Hailo"} {
			if !strings.Contains(string(got), want) {
				t.Errorf("format %s: expected %q in output, got:\n%s", format, want, got)
			}
		}
	}
}

func TestMarshal_UnknownFormat(t *testing.T) {
	_, err := Marshal(machineInfoFromTestData(t, "xps13-9350"), Format("xml"))
	if err == nil {
		t.Fatal("expected error for unknown format, got nil")
	}
	if !strings.Contains(err.Error(), "unknown format") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestFmtBytes(t *testing.T) {
	cases := []struct {
		in   uint64
		want string
	}{
		{0, "0"},
		{500 * (1 << 20), "500.0MiB"},
		{2 * (1 << 30), "2.0GiB"},
		{3 * (1 << 40), "3.0TiB"},
	}
	for _, c := range cases {
		if got := fmtBytes(c.in); got != c.want {
			t.Errorf("fmtBytes(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}
