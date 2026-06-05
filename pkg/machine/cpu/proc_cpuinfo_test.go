package cpu

import (
	"strings"
	"testing"
)

// Minimal two-core amd64 /proc/cpuinfo fixture.
const amd64CpuInfoFixture = `processor	: 0
vendor_id	: GenuineIntel
model name	: Intel(R) Core(TM) i7-6500U CPU @ 2.50GHz
flags		: fpu vme sse sse2 avx

processor	: 1
vendor_id	: GenuineIntel
model name	: Intel(R) Core(TM) i7-6500U CPU @ 2.50GHz
flags		: fpu vme sse sse2 avx
`

// Minimal two-core arm64 /proc/cpuinfo fixture (Raspberry Pi 5 style).
const arm64CpuInfoFixture = `processor	: 0
BogoMIPS	: 108.00
Features	: fp asimd aes sha1 sha2
CPU implementer	: 0x41
CPU architecture: 8
CPU variant	: 0x4
CPU part	: 0xd0b
CPU revision	: 1

processor	: 1
BogoMIPS	: 108.00
Features	: fp asimd aes sha1 sha2
CPU implementer	: 0x41
CPU architecture: 8
CPU variant	: 0x4
CPU part	: 0xd0b
CPU revision	: 1
`

// Minimal two-core riscv64 /proc/cpuinfo fixture (SiFive P550 Premier style).
const riscv64CpuInfoFixture = `processor	: 0
hart		: 3
isa		: rv64imafdch_zicsr_zifencei_zba_zbb_sscofpmf
mmu		: sv48
mvendorid	: 0x489
marchid		: 0x8000000000000008
mimpid		: 0x6220425

processor	: 1
hart		: 0
isa		: rv64imafdch_zicsr_zifencei_zba_zbb_sscofpmf
mmu		: sv48
mvendorid	: 0x489
marchid		: 0x8000000000000008
mimpid		: 0x6220425
`

func TestParseProcCpuInfoAmd64(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantCount     int
		wantProcessor int64
		wantVendor    string
		wantBrand     string
		wantFlags     []string
	}{
		{
			name:          "two cores",
			input:         amd64CpuInfoFixture,
			wantCount:     2,
			wantProcessor: 0,
			wantVendor:    "GenuineIntel",
			wantBrand:     "Intel(R) Core(TM) i7-6500U CPU @ 2.50GHz",
			wantFlags:     []string{"fpu", "vme", "sse", "sse2", "avx"},
		},
		{
			name: "single core",
			input: `processor	: 0
vendor_id	: AuthenticAMD
model name	: AMD EPYC 7742
flags		: sse sse2 avx2
`,
			wantCount:     1,
			wantProcessor: 0,
			wantVendor:    "AuthenticAMD",
			wantBrand:     "AMD EPYC 7742",
			wantFlags:     []string{"sse", "sse2", "avx2"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseProcCpuInfoAmd64(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != tc.wantCount {
				t.Fatalf("expected %d CPUs, got %d", tc.wantCount, len(got))
			}
			first := got[0]
			if first.Processor != tc.wantProcessor {
				t.Errorf("Processor: got %d, want %d", first.Processor, tc.wantProcessor)
			}
			if first.ManufacturerId != tc.wantVendor {
				t.Errorf("ManufacturerId: got %q, want %q", first.ManufacturerId, tc.wantVendor)
			}
			if first.BrandString != tc.wantBrand {
				t.Errorf("BrandString: got %q, want %q", first.BrandString, tc.wantBrand)
			}
			for _, flag := range tc.wantFlags {
				found := false
				for _, f := range first.Flags {
					if f == flag {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected flag %q in Flags %v", flag, first.Flags)
				}
			}
			if first.Architecture != Amd64 {
				t.Errorf("Architecture: got %q, want %q", first.Architecture, Amd64)
			}
		})
	}
}

func TestParseProcCpuInfoArm64(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		wantCount         int
		wantProcessor     int64
		wantBogoMips      float64
		wantImplementerId uint64
		wantPartNumber    uint64
		wantRevision      uint64
		wantFeatures      []string
	}{
		{
			name:              "two cores raspberry pi 5 style",
			input:             arm64CpuInfoFixture,
			wantCount:         2,
			wantProcessor:     0,
			wantBogoMips:      108.00,
			wantImplementerId: 0x41,
			wantPartNumber:    0xd0b,
			wantRevision:      1,
			wantFeatures:      []string{"fp", "asimd", "aes"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseProcCpuInfoArm64(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != tc.wantCount {
				t.Fatalf("expected %d CPUs, got %d", tc.wantCount, len(got))
			}
			first := got[0]
			if first.Processor != tc.wantProcessor {
				t.Errorf("Processor: got %d, want %d", first.Processor, tc.wantProcessor)
			}
			if first.BogoMips != tc.wantBogoMips {
				t.Errorf("BogoMips: got %f, want %f", first.BogoMips, tc.wantBogoMips)
			}
			if first.ImplementerId != tc.wantImplementerId {
				t.Errorf("ImplementerId: got 0x%02x, want 0x%02x", first.ImplementerId, tc.wantImplementerId)
			}
			if first.PartNumber != tc.wantPartNumber {
				t.Errorf("PartNumber: got 0x%03x, want 0x%03x", first.PartNumber, tc.wantPartNumber)
			}
			if first.Revision != tc.wantRevision {
				t.Errorf("Revision: got %d, want %d", first.Revision, tc.wantRevision)
			}
			for _, feat := range tc.wantFeatures {
				found := false
				for _, f := range first.Features {
					if f == feat {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected feature %q in Features %v", feat, first.Features)
				}
			}
			if first.Architecture != Arm64 {
				t.Errorf("Architecture: got %q, want %q", first.Architecture, Arm64)
			}
		})
	}
}

func TestParseProcCpuInfo_Dispatch(t *testing.T) {
	t.Run("routes amd64", func(t *testing.T) {
		got, err := parseProcCpuInfo(amd64CpuInfoFixture, Amd64)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) == 0 {
			t.Fatal("expected CPUs, got none")
		}
		if got[0].Architecture != Amd64 {
			t.Errorf("expected amd64 architecture, got %q", got[0].Architecture)
		}
	})

	t.Run("routes arm64", func(t *testing.T) {
		got, err := parseProcCpuInfo(arm64CpuInfoFixture, Arm64)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) == 0 {
			t.Fatal("expected CPUs, got none")
		}
		if got[0].Architecture != Arm64 {
			t.Errorf("expected arm64 architecture, got %q", got[0].Architecture)
		}
	})

	t.Run("routes riscv64", func(t *testing.T) {
		got, err := parseProcCpuInfo(riscv64CpuInfoFixture, Riscv64)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) == 0 {
			t.Fatal("expected CPUs, got none")
		}
		if got[0].Architecture != Riscv64 {
			t.Errorf("expected riscv64 architecture, got %q", got[0].Architecture)
		}
	})

	t.Run("unknown architecture returns error", func(t *testing.T) {
		_, err := parseProcCpuInfo("processor\t: 0\n", "s390x")
		if err == nil {
			t.Fatal("expected error for unsupported architecture, got nil")
		}
		if !strings.Contains(err.Error(), "unsupported architecture") {
			t.Errorf("expected 'unsupported architecture' in error, got %q", err.Error())
		}
	})
}
