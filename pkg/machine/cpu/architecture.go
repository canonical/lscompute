package cpu

import (
	"fmt"
	"strings"

	"github.com/canonical/lscompute/pkg/machine/constants"
	"golang.org/x/sys/unix"
)

func hostUnameMachine() (string, error) {
	var uname unix.Utsname
	if err := unix.Uname(&uname); err != nil {
		return "", fmt.Errorf("uname syscall: %w", err)
	}
	return unix.ByteSliceToString(uname.Machine[:]), nil
}

// debianArchitecture translates the kernel architecture as reported by uname() to the debian architecture
// Based on lookup table from snapd: https://github.com/canonical/snapd/blob/master/arch/arch.go
func debianArchitecture(unameArch string) (string, error) {
	// Trim whitespace
	unameArch = strings.TrimSpace(unameArch)

	lookupTable := map[string]string{
		// uname:  debian
		"aarch64": constants.Arm64,
		"armv7l":  constants.Armhf,
		"armv8l":  constants.Arm64,
		"i686":    constants.I386,
		"ppc":     constants.Powerpc,
		"ppc64":   constants.Ppc64,
		"ppc64le": constants.Ppc64el,
		"riscv64": constants.Riscv64,
		"s390x":   constants.S390x,
		"x86_64":  constants.Amd64,
	}

	debArch, ok := lookupTable[unameArch]
	if !ok {
		return "", fmt.Errorf("unsupported architecture: %s", unameArch)
	}
	return debArch, nil
}
