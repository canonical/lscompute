package nvidia

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/canonical/lscompute/pkg/machine/host"
	"github.com/canonical/lscompute/pkg/machine/types"
)

const nvidiaSmiTimeout = 30 * time.Second

func gpuProperties(h host.Host, pciDevice types.PciDevice) (map[string]string, error) {
	properties := make(map[string]string)

	vRamVal, err := vRam(h, pciDevice)
	if err != nil {
		return nil, fmt.Errorf("looking up vram: %v", err)
	}
	if vRamVal != nil {
		properties["vram"] = strconv.FormatUint(*vRamVal, 10)
	}

	ccVal, err := computeCapability(h, pciDevice)
	if err != nil {
		return nil, fmt.Errorf("looking up compute capability: %v", err)
	}
	if ccVal != "" {
		properties["compute-capability"] = ccVal
	}

	return properties, nil
}

func vRam(h host.Host, device types.PciDevice) (*uint64, error) {
	/*
		Nvidia: LANG=C nvidia-smi --query-gpu=memory.total --format=csv,noheader,nounits

		$ nvidia-smi --id=00000000:01:00.0 --query-gpu=memory.total --format=csv,noheader
		4096 MiB
	*/
	ctx, cancel := context.WithTimeout(context.Background(), nvidiaSmiTimeout)
	defer cancel()
	output, err := h.RunCommand(ctx, "nvidia-smi", []string{"LANG=C"},
		"--id="+device.Slot, "--query-gpu=memory.total", "--format=csv,noheader")
	if err != nil {
		return nil, fmt.Errorf("executing nvidia-smi: %v", err)
	}
	return parseVramAmount(strings.TrimSpace(string(output)))
}

func parseVramAmount(smiOutputString string) (*uint64, error) {
	if smiOutputString == "[N/A]" {
		return nil, nil
	}

	valueStr, unit, hasUnit := strings.Cut(smiOutputString, " ")
	vramValue, err := strconv.ParseUint(valueStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parsing nvidia-smi output: %v", err)
	}

	if hasUnit {
		switch unit {
		case "KiB":
			vramValue = vramValue * 1024
		case "MiB":
			vramValue = vramValue * 1024 * 1024
		case "GiB":
			vramValue = vramValue * 1024 * 1024 * 1024
		}
	}

	return &vramValue, nil
}

func computeCapability(h host.Host, device types.PciDevice) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), nvidiaSmiTimeout)
	defer cancel()
	output, err := h.RunCommand(ctx, "nvidia-smi", []string{"LANG=C"},
		"--id="+device.Slot, "--query-gpu=compute_cap", "--format=csv,noheader")
	if err != nil {
		return "", fmt.Errorf("executing nvidia-smi: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}
