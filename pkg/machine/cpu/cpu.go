package cpu

import (
	"fmt"
	"io/fs"
	"reflect"
	"slices"
	"strings"

	"github.com/canonical/lscompute/pkg/machine/host"
	"github.com/canonical/lscompute/pkg/machine/types"
)

func Info(h host.Host) ([]CpuInfo, error) {
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
// If reading that file fails for any reason, it falls back to uname(2).
func machineArch(h host.Host) (string, error) {
	data, err := fs.ReadFile(h.FS(), "proc/sys/kernel/arch")
	if err == nil {
		return strings.TrimSpace(string(data)), nil
	}
	// Any read error (missing or unreadable file) falls back to uname(2).
	return hostMachineArchFallback()
}

func infoFromRawData(procCpuInfoData string, uname string) ([]CpuInfo, error) {
	architecture, err := debianArchitecture(uname)
	if err != nil {
		return nil, fmt.Errorf("translating architecture: %w", err)
	}

	machineProcCpuInfo, err := parseProcCpuInfo(procCpuInfoData, architecture)
	if err != nil {
		return nil, fmt.Errorf("parsing cpuinfo: %w", err)
	}
	if len(machineProcCpuInfo) == 0 {
		return nil, fmt.Errorf("parsing cpuinfo: no cpu entries found")
	}

	cpus, err := uniqueCpuInfo(machineProcCpuInfo)
	if err != nil {
		return nil, fmt.Errorf("filtering cpu info: %w", err)
	}
	if len(cpus) == 0 {
		return nil, fmt.Errorf("filtering cpu info: no cpu info entries produced")
	}

	return cpus, nil
}

func uniqueCpuInfo(procCpus []procCpuInfo) ([]CpuInfo, error) {
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

func cpuInfoFromProc(procCpus []procCpuInfo) ([]CpuInfo, error) {
	var cpuInfos []CpuInfo
	for _, procCpu := range procCpus {
		var cpuInfo CpuInfo
		if procCpu.Architecture == Amd64 {
			cpuInfo.Architecture = procCpu.Architecture
			cpuInfo.ManufacturerId = procCpu.ManufacturerId
			cpuInfo.Flags = procCpu.Flags
		} else if procCpu.Architecture == Arm64 {
			cpuInfo.Architecture = procCpu.Architecture
			cpuInfo.ImplementerId = types.HexInt(procCpu.ImplementerId)
			cpuInfo.PartNumber = types.HexInt(procCpu.PartNumber)
			cpuInfo.Features = procCpu.Features
		} else if procCpu.Architecture == Riscv64 {
			cpuInfo.Architecture = procCpu.Architecture
			cpuInfo.Isa = procCpu.Isa
		} else {
			return nil, fmt.Errorf("unsupported architecture: %s", procCpu.Architecture)
		}
		cpuInfos = append(cpuInfos, cpuInfo)
	}
	return cpuInfos, nil
}
