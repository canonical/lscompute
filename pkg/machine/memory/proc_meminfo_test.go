package memory

import (
	"testing"
)

func TestParseProcMemInfo(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantTotalRam  uint64
		wantTotalSwap uint64
	}{
		{
			name: "typical kB values with swap",
			input: `MemTotal:       16330752 kB
MemFree:         1234567 kB
SwapTotal:       2097152 kB
SwapFree:        2097152 kB
`,
			wantTotalRam:  16330752 * 1024,
			wantTotalSwap: 2097152 * 1024,
		},
		{
			name: "zero swap",
			input: `MemTotal:       8000000 kB
MemFree:         400000 kB
SwapTotal:       0 kB
`,
			wantTotalRam:  8000000 * 1024,
			wantTotalSwap: 0,
		},
		{
			name: "unknown keys are ignored",
			input: `MemTotal:       4096000 kB
SomeUnknownKey: 9999 kB
SwapTotal:      1048576 kB
`,
			wantTotalRam:  4096000 * 1024,
			wantTotalSwap: 1048576 * 1024,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseProcMemInfo(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.TotalRam != tc.wantTotalRam {
				t.Errorf("TotalRam: got %d, want %d", got.TotalRam, tc.wantTotalRam)
			}
			if got.TotalSwap != tc.wantTotalSwap {
				t.Errorf("TotalSwap: got %d, want %d", got.TotalSwap, tc.wantTotalSwap)
			}
		})
	}
}

func TestProcStringToBytes(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{
			name:  "kB suffix",
			input: "16330752 kB",
			want:  16330752 * 1024,
		},
		{
			name:  "kB suffix with extra whitespace",
			input: "  8192 kB  ",
			want:  8192 * 1024,
		},
		{
			name:  "plain bytes no suffix",
			input: "1048576",
			want:  1048576,
		},
		{
			name:    "invalid value",
			input:   "notanumber kB",
			wantErr: true,
		},
		{
			name:    "invalid plain value",
			input:   "abc",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := procStringToBytes(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %d, want %d", got, tc.want)
			}
		})
	}
}
