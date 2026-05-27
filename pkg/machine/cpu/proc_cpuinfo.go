package cpu

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/canonical/lscompute/pkg/machine/constants"
)

func parseProcCpuInfo(cpuInfoString string, architecture string) ([]procCpuInfo, error) {
	switch architecture {

	case constants.Amd64:
		cpuInfo, err := parseProcCpuInfoAmd64(cpuInfoString)
		if err != nil {
			return nil, fmt.Errorf("amd64: %w", err)
		}
		return cpuInfo, nil

	case constants.Arm64:
		cpuInfo, err := parseProcCpuInfoArm64(cpuInfoString)
		if err != nil {
			return nil, fmt.Errorf("arm64: %w", err)
		}
		return cpuInfo, nil

	default:
		return nil, fmt.Errorf("unsupported architecture: %s", architecture)

	}
}

func parseProcCpuInfoAmd64(cpuInfoString string) ([]procCpuInfo, error) {
	var parsedCpus []procCpuInfo

	lines := strings.Split(cpuInfoString, "\n")
	cpuIndex := -1

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.SplitN(line, ":", 2)
		if len(fields) != 2 {
			return nil, fmt.Errorf("malformed cpuinfo line: %q", line)
		}
		key := strings.TrimSpace(fields[0]) // remove \t between key and colon
		value := strings.TrimSpace(fields[1])

		// New cpu block
		if key == "processor" {
			newCpu := procCpuInfo{}
			newCpu.Architecture = constants.Amd64
			parsedCpus = append(parsedCpus, newCpu)
			cpuIndex = len(parsedCpus) - 1
		}

		if cpuIndex < 0 {
			return nil, fmt.Errorf("field %q encountered before first processor", key)
		}

		switch key {
		case "processor":
			processorIndex, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return nil, err
			}
			parsedCpus[cpuIndex].Processor = processorIndex
		case "vendor_id":
			parsedCpus[cpuIndex].ManufacturerId = value

		case "flags":
			flags := strings.Fields(value)
			parsedCpus[cpuIndex].Flags = append(parsedCpus[cpuIndex].Flags, flags...)

		case "model name":
			parsedCpus[cpuIndex].BrandString = value
		}
	}

	return parsedCpus, nil
}

func parseProcCpuInfoArm64(cpuInfoString string) ([]procCpuInfo, error) {
	var parsedCpus []procCpuInfo

	lines := strings.Split(cpuInfoString, "\n")
	cpuIndex := -1

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.SplitN(line, ":", 2)
		if len(fields) != 2 {
			return nil, fmt.Errorf("malformed cpuinfo line: %q", line)
		}
		key := strings.TrimSpace(fields[0]) // remove \t between key and colon
		value := strings.TrimSpace(fields[1])

		// New cpu block
		if key == "processor" {
			newCpu := procCpuInfo{}
			newCpu.Architecture = constants.Arm64
			parsedCpus = append(parsedCpus, newCpu)
			cpuIndex = len(parsedCpus) - 1
		}

		if cpuIndex < 0 {
			return nil, fmt.Errorf("field %q encountered before first processor", key)
		}

		switch key {

		// Formatting strings above the following cases are from https://github.com/torvalds/linux/blob/master/arch/arm64/kernel/cpuinfo.c
		// "processor\t: %d\n"
		case "processor":
			processorIndex, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return nil, err
			}
			parsedCpus[cpuIndex].Processor = processorIndex

		// "model name\t: ARMv8 Processor rev %d (%s)\n"
		case "model name":
			modelName := strings.TrimSpace(value)
			parsedCpus[cpuIndex].ModelName = &modelName

		// BogoMIPS\t: %lu.%02lu\n
		case "BogoMIPS", "bogomips":
			bogoMips, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, err
			}
			parsedCpus[cpuIndex].BogoMips = bogoMips

		// "Features\t:"+" %s"
		case "Features":
			flags := strings.Fields(value)
			parsedCpus[cpuIndex].Features = append(parsedCpus[cpuIndex].Features, flags...)

		// "CPU implementer\t: 0x%02x\n"
		case "CPU implementer":
			implementer, err := strconv.ParseUint(value, 0, 8) // use base 0 to allow parser to detect and remove 0x prefix
			if err != nil {
				return nil, err
			}
			parsedCpus[cpuIndex].ImplementerId = implementer

		// "CPU architecture: 8\n"
		case "CPU architecture":
			//architecture, err := strconv.ParseUint(value, 10, 64)
			//if err != nil {
			//	return nil, err
			//}
			parsedCpus[cpuIndex].Architecture = constants.Arm64

		// "CPU variant\t: 0x%x\n"
		case "CPU variant":
			variant, err := strconv.ParseUint(value, 0, 64)
			if err != nil {
				return nil, err
			}
			parsedCpus[cpuIndex].Variant = variant

		// "CPU part\t: 0x%03x\n"
		case "CPU part":
			part, err := strconv.ParseUint(value, 0, 16)
			if err != nil {
				return nil, err
			}
			parsedCpus[cpuIndex].PartNumber = part

		// "CPU revision\t: %d\n\n"
		case "CPU revision":
			revision, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return nil, err
			}
			parsedCpus[cpuIndex].Revision = revision
		}
	}

	return parsedCpus, nil
}
