package cpu

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/canonical/lscompute/pkg/machine/constants"
	"golang.org/x/sys/unix"
)

// hostMachineArch returns the kernel's machine architecture string (e.g. "x86_64",
// "aarch64"), matching the output of `uname -m`. It reads /proc/sys/kernel/arch when
// available and falls back to the uname(2) syscall on older kernels that do not export
// the sysctl (Linux < 6.1).
func hostMachineArch() (string, error) {
	data, err := os.ReadFile("/proc/sys/kernel/arch")
	if err == nil {
		return strings.TrimSpace(string(data)), nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return "", fmt.Errorf("reading /proc/sys/kernel/arch: %v", err)
	}

	var uname unix.Utsname
	if err := unix.Uname(&uname); err != nil {
		return "", fmt.Errorf("uname syscall: %w", err)
	}
	return unix.ByteSliceToString(uname.Machine[:]), nil
}

// debianArchitecture translates the kernel architecture (as reported by `uname -m` /
// /proc/sys/kernel/arch) to the debian architecture.
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
