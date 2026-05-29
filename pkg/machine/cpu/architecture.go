package cpu

import (
	"fmt"
	"strings"

	"golang.org/x/sys/unix"
)

// hostMachineArchFallback uses the uname(2) syscall to get the machine architecture.
// This is only reachable when proc/sys/kernel/arch is not present (Linux < 6.1).
// Fake hosts must always provide that file, so this fallback only runs on real hosts.
func hostMachineArchFallback() (string, error) {
	var uname unix.Utsname
	if err := unix.Uname(&uname); err != nil {
		return "", fmt.Errorf("uname syscall: %w", err)
	}
	return strings.TrimSpace(unix.ByteSliceToString(uname.Machine[:])), nil
}

// debianArchitecture translates the kernel architecture (as reported by `uname -m` /
// /proc/sys/kernel/arch) to the debian architecture.
// Based on lookup table from snapd: https://github.com/canonical/snapd/blob/master/arch/arch.go
func debianArchitecture(unameArch string) (string, error) {
	// Trim whitespace
	unameArch = strings.TrimSpace(unameArch)

	lookupTable := map[string]string{
		// uname:  debian
		"aarch64": Arm64,
		"armv7l":  Armhf,
		"armv8l":  Arm64,
		"i686":    I386,
		"ppc":     Powerpc,
		"ppc64":   Ppc64,
		"ppc64le": Ppc64el,
		"riscv64": Riscv64,
		"s390x":   S390x,
		"x86_64":  Amd64,
	}

	debArch, ok := lookupTable[unameArch]
	if !ok {
		return "", fmt.Errorf("unsupported architecture: %s", unameArch)
	}
	return debArch, nil
}
