package nvidia

import (
	"testing"

	"github.com/canonical/lscompute/pkg/machine/host"
)

func TestVRam(t *testing.T) {
	expected := uint64(4096 * 1024 * 1024)
	tests := []struct {
		name      string
		testInput string
		expected  *uint64
		err       error
		shouldErr bool
	}{
		{
			name:      "converts MiB to bytes",
			testInput: "4096 MiB",
			expected:  &expected,
		},
		{
			name:      "returns nil for unavailable VRAM",
			testInput: "[N/A]",
			expected:  nil,
		},
		{
			name:      "reports parsing errors",
			testInput: "not-a-number MiB",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseVramAmount(tt.testInput)
			if tt.shouldErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.expected == nil {
				if got != nil {
					t.Fatalf("expected nil, got %d", *got)
				}
				return
			}
			if got == nil {
				t.Fatalf("expected non-nil result")
			}
			if *got != *tt.expected {
				t.Fatalf("expected %d, got %d", *tt.expected, *got)
			}
		})
	}
}

const i5gtxMachineRoot = "../../../../../test_data/machines/i5-3570k+arc-a580+gtx1080ti/machine-root"
const gtxSlot = "0000:01:00.0"

// TestParseVramAmount_KiBAndGiB exercises the KiB and GiB unit conversion paths.
func TestParseVramAmount_KiBAndGiB(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  uint64
	}{
		{"KiB", "1024 KiB", 1024 * 1024},
		{"GiB", "2 GiB", 2 * 1024 * 1024 * 1024},
		{"no unit", "4096", 4096},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseVramAmount(tc.input)
			if err != nil {
				t.Fatalf("parseVramAmount(%q) error: %v", tc.input, err)
			}
			if got == nil {
				t.Fatalf("parseVramAmount(%q) = nil, want %d", tc.input, tc.want)
			}
			if *got != tc.want {
				t.Errorf("parseVramAmount(%q) = %d, want %d", tc.input, *got, tc.want)
			}
		})
	}
}

func TestVRam_Nvidia(t *testing.T) {
	h := host.Fake(i5gtxMachineRoot)
	got, err := vRam(h, gtxSlot)
	if err != nil {
		t.Fatalf("vRam() unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("vRam() returned nil, expected non-nil")
	}
	// Fixture: "11264 MiB" → 11264 * 1024 * 1024 = 11811160064
	const want = uint64(11264 * 1024 * 1024)
	if *got != want {
		t.Errorf("vRam() = %d, want %d", *got, want)
	}
}

func TestVRam_Nvidia_MissingFixture(t *testing.T) {
	h := host.Fake(t.TempDir())
	_, err := vRam(h, gtxSlot)
	if err == nil {
		t.Fatal("expected error for missing nvidia-smi fixture, got nil")
	}
}

func TestComputeCapability(t *testing.T) {
	h := host.Fake(i5gtxMachineRoot)
	got, err := computeCapability(h, gtxSlot)
	if err != nil {
		t.Fatalf("computeCapability() unexpected error: %v", err)
	}
	// Fixture has "6.1"
	if got != "6.1" {
		t.Errorf("computeCapability() = %q, want %q", got, "6.1")
	}
}

func TestComputeCapability_MissingFixture(t *testing.T) {
	h := host.Fake(t.TempDir())
	_, err := computeCapability(h, gtxSlot)
	if err == nil {
		t.Fatal("expected error for missing nvidia-smi fixture, got nil")
	}
}

func TestGpuProperties_Nvidia(t *testing.T) {
	h := host.Fake(i5gtxMachineRoot)
	props, err := gpuProperties(h, gtxSlot)
	if err != nil {
		t.Fatalf("gpuProperties() unexpected error: %v", err)
	}
	if v, ok := props["vram"]; !ok || v == "" {
		t.Errorf("expected 'vram' property, got %v", props)
	}
	if v, ok := props["compute-capability"]; !ok || v == "" {
		t.Errorf("expected 'compute-capability' property, got %v", props)
	}
}

func TestAdditionalProperties_NvidiaGpu(t *testing.T) {
	h := host.Fake(i5gtxMachineRoot)
	props, err := AdditionalProperties(h, gtxSlot, true)
	if err != nil {
		t.Fatalf("AdditionalProperties(isGpu=true) error: %v", err)
	}
	if props == nil {
		t.Fatal("expected non-nil properties for GPU")
	}
}

func TestAdditionalProperties_NvidiaNotGpu(t *testing.T) {
	h := host.Fake(i5gtxMachineRoot)
	props, err := AdditionalProperties(h, gtxSlot, false)
	if err != nil {
		t.Fatalf("AdditionalProperties(isGpu=false) error: %v", err)
	}
	if props != nil {
		t.Errorf("expected nil properties for non-GPU, got %v", props)
	}
}
