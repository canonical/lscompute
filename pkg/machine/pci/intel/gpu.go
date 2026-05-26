package intel

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/canonical/lscompute/pkg/machine/host"
	"github.com/canonical/lscompute/pkg/machine/types"
)

const clInfoTimeout = 10 * time.Second

func gpuProperties(h host.Host, pciDevice types.PciDevice) (map[string]string, error) {
	properties := make(map[string]string)

	vRamVal, err := vRam(h, pciDevice)
	if err != nil {
		return nil, fmt.Errorf("looking up vram: %v", err)
	}
	if vRamVal != nil {
		properties["vram"] = strconv.FormatUint(*vRamVal, 10)
	}

	return properties, nil
}

func vRam(h host.Host, device types.PciDevice) (*uint64, error) {
	/*
		For GPU vRAM information use clinfo. `clinfo --json` reports a field
		`CL_DEVICE_GLOBAL_MEM_SIZE` which corresponds to the installed hardware's vRAM.
	*/
	ctx, cancel := context.WithTimeout(context.Background(), clInfoTimeout)
	defer cancel()

	data, err := h.RunCommand(ctx, "clinfo", nil, "--json")
	if err != nil {
		return nil, fmt.Errorf("executing clinfo: %v", err)
	}

	clinfo, err := parseClinfoJson(data)
	if err != nil {
		return nil, fmt.Errorf("parsing clinfo output: %w", err)
	}
	if len(clinfo.Devices) == 0 {
		return nil, fmt.Errorf("clinfo: no devices found")
	}
	if len(clinfo.Devices[0].Online) == 0 {
		return nil, fmt.Errorf("clinfo: no online devices found")
	}

	var vramValue *uint64 = nil
	// Search for the device with a matching PCI address
	for _, clInfoDevice := range clinfo.Devices[0].Online {
		if strings.Contains(clInfoDevice.ClDevicePciBusInfoKhr, device.Slot) {
			vram := clInfoDevice.ClDeviceGlobalMemSize
			vramValue = &vram
		}
	}
	return vramValue, nil
}

func parseClinfoJson(clinfoJson []byte) (types.Clinfo, error) {
	clinfo := types.Clinfo{}
	err := json.Unmarshal(clinfoJson, &clinfo)
	return clinfo, err
}
