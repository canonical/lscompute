package cpu

import (
	"errors"
	"fmt"
	"io/fs"
	"reflect"
	"slices"
	"strings"

	"github.com/canonical/lscompute/pkg/machine/constants"
	"github.com/canonical/lscompute/pkg/machine/host"
	"github.com/canonical/lscompute/pkg/machine/types"
)

func Info(h host.Host) ([]types.CpuInfo, error) {
	procCpuData, err := fs.ReadFile(h.FS(), "proc/cpuinfo")
	if err != nil {
		return nil, fmt.Errorf("reading proc/cpuinfo: %w", err)
	}

	archData, err := machineArch(h)
	if err != nil {
		return nil, fmt.Errorf("getting machine architecture: %w", err)
	}

	cpus, err := infoFromRawData(string(procCpuData), archData)
	if err != nil {
		return nil, fmt.Errorf("parsing cpu data: %w", err)
	}

	return cpus, nil
}

// machineArch returns the kernel machine architecture string (e.g. "x86_64").
// It reads proc/sys/kernel/arch via the host FS (available on Linux 6.1+).
// On older kernels the file does not exist; for the real host the architecture.go
// fallback uses uname(2). For a fake host the file must be present.
func machineArch(h host.Host) (string, error) {
	data, err := fs.ReadFile(h.FS(), "proc/sys/kernel/arch")
	if err == nil {
		return strings.TrimSpace(string(data)), nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return "", fmt.Errorf("reading proc/sys/kernel/arch: %w", err)
	}
	// File not present — fall back to uname(2). This only works on a real host;
	// fake hosts must always provide the file.
	return hostMachineArchFallback()
}

func infoFromRawData(procCpuInfoData string, uname string) ([]types.CpuInfo, error) {
	architecture, err := debianArchitecture(uname)
	if err != nil {
		return nil, fmt.Errorf("translating architecture: %w", err)
	}

	machineprocCpuInfo, err := parseProcCpuInfo(procCpuInfoData, architecture)
	if err != nil {
		return nil, fmt.Errorf("parsing cpuinfo: %w", err)
	}
	if len(machineprocCpuInfo) == 0 {
		return nil, fmt.Errorf("parsing cpuinfo: no cpu entries found")
	}

	cpus, err := uniqueCpuInfo(machineprocCpuInfo)
	if err != nil {
		return nil, fmt.Errorf("filtering cpu info: %w", err)
	}
	if len(cpus) == 0 {
		return nil, fmt.Errorf("filtering cpu info: no cpu info entries produced")
	}

	return cpus, nil
}

func uniqueCpuInfo(procCpus []procCpuInfo) ([]types.CpuInfo, error) {
	// Set processor index to 0 to only check other fields for uniqueness
	for i := range procCpus {
		procCpus[i].Processor = 0
	}

	procCpus = slices.CompactFunc(procCpus, isDuplicate)

	cpuInfos, err := cpuInfoFromProc(procCpus)
	if err != nil {
		return nil, fmt.Errorf("converting cpu info: %w", err)
	}
	return cpuInfos, nil
}

func isDuplicate(a procCpuInfo, b procCpuInfo) bool {
	return reflect.DeepEqual(a, b)
}

func cpuInfoFromProc(procCpus []procCpuInfo) ([]types.CpuInfo, error) {
	var cpuInfos []types.CpuInfo
	for _, procCpu := range procCpus {
		var cpuInfo types.CpuInfo
		if procCpu.Architecture == constants.Amd64 {
			cpuInfo.Architecture = procCpu.Architecture
			cpuInfo.ManufacturerId = procCpu.ManufacturerId
			cpuInfo.Flags = procCpu.Flags
		} else if procCpu.Architecture == constants.Arm64 {
			cpuInfo.Architecture = procCpu.Architecture
			cpuInfo.ImplementerId = types.HexInt(procCpu.ImplementerId)
			cpuInfo.PartNumber = types.HexInt(procCpu.PartNumber)
			cpuInfo.Features = procCpu.Features
		} else {
			return nil, fmt.Errorf("unsupported architecture: %s", procCpu.Architecture)
		}
		cpuInfos = append(cpuInfos, cpuInfo)
	}
	return cpuInfos, nil
}
