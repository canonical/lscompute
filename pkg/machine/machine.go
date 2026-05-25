package machine

import (
	"fmt"
	"os"

	"github.com/canonical/lscompute/pkg/machine/cpu"
	"github.com/canonical/lscompute/pkg/machine/disk"
	"github.com/canonical/lscompute/pkg/machine/memory"
	"github.com/canonical/lscompute/pkg/machine/types"
)

func Get(friendlyNames bool) (*types.MachineInfo, []string, error) {
	var machineInfo types.MachineInfo

	memoryInfo, err := memory.Info()
	if err != nil {
		return nil, nil, fmt.Errorf("getting memory info: %v", err)
	}
	machineInfo.Memory = memoryInfo

	cpus, err := cpu.Info()
	if err != nil {
		return nil, nil, fmt.Errorf("getting cpu info: %v", err)
	}
	machineInfo.Cpus = cpus

	diskInfo, err := disk.Info()
	if err != nil {
		return nil, nil, fmt.Errorf("getting disk info: %v", err)
	}
	machineInfo.Disk = diskInfo

	devices, warnings, err := Devices(friendlyNames)
	if err != nil {
		return nil, nil, fmt.Errorf("getting devices: %v", err)
	}
	machineInfo.Devices = devices

	return &machineInfo, warnings, nil
}

// GetFromRawData is a test helper
func GetFromRawData(device string, friendlyNames bool, testDir string) (*types.MachineInfo, error) {
	var hwInfo types.MachineInfo

	devicePath := testDir + "/machines/" + device + "/"

	// memory
	procMemInfo, err := os.ReadFile(devicePath + "meminfo.txt")
	if err != nil {
		return nil, err
	}
	memInfo, err := memory.InfoFromRawData(string(procMemInfo))
	if err != nil {
		return nil, err
	}
	hwInfo.Memory = memInfo

	// disk
	dfInfo, err := os.ReadFile(devicePath + "disk.txt")
	if err != nil {
		return nil, err
	}
	diskInfo, err := disk.InfoFromRawData(string(dfInfo))
	if err != nil {
		return nil, err
	}
	hwInfo.Disk = diskInfo

	// cpu
	unameMachine, err := os.ReadFile(devicePath + "uname-m.txt")
	if err != nil {
		return nil, err
	}
	procCpuInfo, err := os.ReadFile(devicePath + "cpuinfo.txt")
	if err != nil {
		return nil, err
	}
	cpuInfo, err := cpu.InfoFromRawData(string(procCpuInfo), string(unameMachine))
	if err != nil {
		return nil, err
	}
	hwInfo.Cpus = cpuInfo

	// pci
	//pciData, err := os.ReadFile(devicePath + "lspci.txt")
	//if err != nil {
	//	return nil, err
	//}
	//pciDevices, warnings, err := pci.DevicesFromRawData(string(pciData), friendlyNames)
	//if len(warnings) > 0 {
	//	fmt.Printf("Warnings: %v\n", warnings)
	//}
	//if err != nil {
	//	return nil, err
	//}
	//hwInfo.PciDevices = pciDevices
	//
	//// Additional properties - we append these directly from a file, as we can not run the vendor specific tools on the machine
	//addPropsFile := devicePath + "additional-properties.json"
	//_, err = os.Stat(addPropsFile)
	//if err != nil {
	//	if os.IsNotExist(err) {
	//		// File does not exist. Skipping additional properties
	//	} else {
	//		return nil, fmt.Errorf("error checking file '%s': %v\n", addPropsFile, err)
	//	}
	//} else {
	//	var addProps map[string]map[string]string
	//	addPropsData, err := os.ReadFile(devicePath + "additional-properties.json")
	//	if err != nil {
	//		return nil, err
	//	}
	//	err = json.Unmarshal(addPropsData, &addProps)
	//	if err != nil {
	//		return nil, err
	//	}
	//	for i, pciDevice := range hwInfo.PciDevices {
	//		if val, ok := addProps[pciDevice.Slot]; ok {
	//			hwInfo.Devices[i].AdditionalProperties = val
	//		}
	//	}
	//}

	return &hwInfo, nil
}
