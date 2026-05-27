package intel

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/lscompute/pkg/machine/host"
)

var clinfoFiles = []string{
	"../../../../test_data/clinfo/intel-arc-a580.json",
	"../../../../test_data/clinfo/intel-arc-a580-inside-snap.json",
	"../../../../test_data/clinfo/intel-arc-b580.json",
	"../../../../test_data/clinfo/no-devices.json",
}

func TestParseClinfo(t *testing.T) {
	for _, clinfoFile := range clinfoFiles {
		t.Run(clinfoFile, func(t *testing.T) {
			clinfoJson, err := os.ReadFile(clinfoFile)
			if err != nil {
				t.Fatal(err)
			}
			clinfo, err := parseClinfoJson(clinfoJson)
			if err != nil {
				t.Fatal(err)
			}
			if len(clinfo.Devices) > 0 {
				if len(clinfo.Devices[0].Online) > 0 {
					t.Logf("Global memory: %d", clinfo.Devices[0].Online[0].ClDeviceGlobalMemSize)
				}
			}
		})
	}
}

const i5MachineRoot = "../../../../test_data/machines/i5-3570k+arc-a580+gtx1080ti/machine-root"

func requireIntelFixture(t *testing.T) {
	t.Helper()
	path := filepath.Join(i5MachineRoot, "run", "clinfo.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skipf("fixture not present yet: %s", path)
	}
}

// TestVRam_Intel verifies vRam returns the correct value for a known Intel GPU
// slot from the clinfo fixture.
func TestVRam_Intel(t *testing.T) {
	requireIntelFixture(t)
	h := host.Fake(i5MachineRoot)
	// The clinfo fixture reports the Arc A580 at PCI-E, 0000:03:00.0
	got, err := vRam(h, "0000:03:00.0")
	if err != nil {
		t.Fatalf("vRam() unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("vRam() returned nil, expected non-nil")
	}
	// Expected: 8096681984 bytes as reported in the clinfo fixture.
	const want = uint64(8096681984)
	if *got != want {
		t.Errorf("vRam() = %d, want %d", *got, want)
	}
}

func TestVRam_Intel_NoMatch(t *testing.T) {
	requireIntelFixture(t)
	// A slot that doesn't appear in the clinfo output → nil, no error.
	h := host.Fake(i5MachineRoot)
	got, err := vRam(h, "9999:99:99.9")
	if err != nil {
		t.Fatalf("vRam() unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("vRam() = %d, expected nil for unknown slot", *got)
	}
}

func TestVRam_Intel_MissingClinfo(t *testing.T) {
	// No run/clinfo.json in empty host → error.
	h := host.Fake(t.TempDir())
	_, err := vRam(h, "0000:03:00.0")
	if err == nil {
		t.Fatal("expected error for missing clinfo.json, got nil")
	}
}

func TestVRam_Intel_EmptyDevices(t *testing.T) {
	// A clinfo.json with a devices array but no online entries → error.
	dir := t.TempDir()
	runDir := dir + "/run"
	if err := os.MkdirAll(runDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(runDir+"/clinfo.json", []byte(`{"devices":[]}`), 0644); err != nil {
		t.Fatal(err)
	}
	h := host.Fake(dir)
	_, err := vRam(h, "0000:03:00.0")
	if err == nil {
		t.Fatal("expected error for empty devices list, got nil")
	}
}

func TestGpuProperties_Intel(t *testing.T) {
	requireIntelFixture(t)
	h := host.Fake(i5MachineRoot)
	props, err := gpuProperties(h, "0000:03:00.0")
	if err != nil {
		t.Fatalf("gpuProperties() unexpected error: %v", err)
	}
	vram, ok := props["vram"]
	if !ok || vram == "" {
		t.Errorf("expected 'vram' property, got props=%v", props)
	}
}

func TestAdditionalProperties_IntelGpu(t *testing.T) {
	requireIntelFixture(t)
	h := host.Fake(i5MachineRoot)
	props, err := AdditionalProperties(h, "0000:03:00.0", true)
	if err != nil {
		t.Fatalf("AdditionalProperties(isGpu=true) error: %v", err)
	}
	if props == nil {
		t.Fatal("expected non-nil properties for GPU")
	}
}

func TestAdditionalProperties_IntelNotGpu(t *testing.T) {
	requireIntelFixture(t)
	h := host.Fake(i5MachineRoot)
	props, err := AdditionalProperties(h, "0000:03:00.0", false)
	if err != nil {
		t.Fatalf("AdditionalProperties(isGpu=false) error: %v", err)
	}
	if props != nil {
		t.Errorf("expected nil properties for non-GPU, got %v", props)
	}
}
