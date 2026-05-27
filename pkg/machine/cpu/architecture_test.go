package cpu

import (
	"testing"

	"github.com/canonical/lscompute/pkg/machine/constants"
)

func TestDebianArchitecture(t *testing.T) {
	cases := []struct {
		uname   string
		want    string
		wantErr bool
	}{
		{"aarch64", constants.Arm64, false},
		{"armv7l", constants.Armhf, false},
		{"armv8l", constants.Arm64, false},
		{"i686", constants.I386, false},
		{"ppc", constants.Powerpc, false},
		{"ppc64", constants.Ppc64, false},
		{"ppc64le", constants.Ppc64el, false},
		{"riscv64", constants.Riscv64, false},
		{"s390x", constants.S390x, false},
		{"x86_64", constants.Amd64, false},
		// Whitespace must be trimmed.
		{"  x86_64  ", constants.Amd64, false},
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
