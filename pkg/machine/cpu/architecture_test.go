package cpu

import (
	"testing"
)

func TestDebianArchitecture(t *testing.T) {
	cases := []struct {
		uname   string
		want    string
		wantErr bool
	}{
		{"aarch64", Arm64, false},
		{"armv7l", Armhf, false},
		{"armv8l", Arm64, false},
		{"i686", I386, false},
		{"ppc", Powerpc, false},
		{"ppc64", Ppc64, false},
		{"ppc64le", Ppc64el, false},
		{"riscv64", Riscv64, false},
		{"s390x", S390x, false},
		{"x86_64", Amd64, false},
		// Whitespace must be trimmed.
		{"  x86_64  ", Amd64, false},
		// Unsupported arch → error.
		{"mips64", "", true},
		{"loongarch64", "", true},
		{"", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.uname, func(t *testing.T) {
			got, err := debianArchitecture(tc.uname)
			if tc.wantErr {
				if err == nil {
					t.Errorf("debianArchitecture(%q): expected error, got %q", tc.uname, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("debianArchitecture(%q): unexpected error: %v", tc.uname, err)
			}
			if got != tc.want {
				t.Errorf("debianArchitecture(%q) = %q, want %q", tc.uname, got, tc.want)
			}
		})
	}
}
