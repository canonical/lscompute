package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/canonical/lscompute/pkg/machine"
	"github.com/canonical/lscompute/pkg/machine/host"
)

func machineInfoFromTestData(machineName string) (*machine.MachineInfo, error) {
	machineRoot := filepath.Join("..", "..", "test_data", "machines", machineName, "machine-root")

	pciIdsPath := filepath.Join(machineRoot, "usr", "share", "misc", "pci.ids")
	_, pciIdsErr := os.Stat(pciIdsPath)
	friendlyNames := pciIdsErr == nil

	info, _, err := machine.Get(host.Fake(machineRoot), friendlyNames)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func Example_marshalJson() {
	machineName := "dummy-machine"
	machineInfo, err := machineInfoFromTestData(machineName)
	if err != nil {
		fmt.Printf("Get() failed: %v", err)
		return
	}
	testOutput, err := NewMachineDetails(machineInfo).Marshal(FormatJSON)
	if err != nil {
		fmt.Printf("Marshal() failed: %v", err)
		return
	}
	fmt.Println(string(testOutput))
	// Output:
	// {
	//   "cpus": [
	//     {
	//       "architecture": "amd64",
	//       "manufacturer-id": "GenuineIntel",
	//       "flags": [
	//         "fpu",
	//         "vme",
	//         "de"
	//       ]
	//     }
	//   ],
	//   "memory": {
	//     "total-ram": 67012501504,
	//     "total-swap": 0
	//   },
	//   "disk": {
	//     "/var/lib/snapd/snaps": {
	//       "total": 1006451294208,
	//       "avail": 943543738368
	//     }
	//   },
	//   "devices": [
	//     {
	//       "bus": "pci",
	//       "slot": "0000:00:00.0",
	//       "bus-number": "0x0",
	//       "device-class": "0x600",
	//       "vendor-id": "0x8086",
	//       "device-id": "0x4637",
	//       "subvendor-id": "0x103C",
	//       "subdevice-id": "0x89C6"
	//     },
	//     {
	//       "bus": "pci",
	//       "slot": "0000:00:02.0",
	//       "bus-number": "0x0",
	//       "device-class": "0x300",
	//       "vendor-id": "0x8086",
	//       "device-id": "0x9B41",
	//       "subvendor-id": "0x1028",
	//       "subdevice-id": "0x962",
	//       "additional-properties": {
	//         "vram": 14477950976
	//       }
	//     },
	//     {
	//       "bus": "pci",
	//       "slot": "0000:01:00.0",
	//       "bus-number": "0x1",
	//       "device-class": "0x300",
	//       "vendor-id": "0x10DE",
	//       "device-id": "0x1B06",
	//       "subvendor-id": "0x10DE",
	//       "subdevice-id": "0x1B06",
	//       "additional-properties": {
	//         "vram": 11811160064,
	//         "compute-capability": "6.1"
	//       }
	//     },
	//     {
	//       "bus": "pci",
	//       "slot": "0000:03:00.0",
	//       "bus-number": "0x3",
	//       "device-class": "0x300",
	//       "vendor-id": "0x1002",
	//       "device-id": "0x73E1",
	//       "subvendor-id": "0x103C",
	//       "subdevice-id": "0x89C6",
	//       "additional-properties": {
	//         "microarchitecture": "gfx1032",
	//         "vram": 8573157376
	//       }
	//     }
	//   ]
	// }

}

func Example_marshalPlain() {
	machineName := "dummy-machine"
	machineInfo, err := machineInfoFromTestData(machineName)
	if err != nil {
		fmt.Printf("Get() failed: %v", err)
		return
	}
	testOutput, err := NewMachineDetails(machineInfo).Marshal(FormatPlain)
	if err != nil {
		fmt.Printf("Marshal() failed: %v", err)
		return
	}
	fmt.Println(string(testOutput))
	// Output:
	// cpus:
	//   - architecture: amd64
	//     manufacturer-id: GenuineIntel
	//     flags: [fpu, vme, de]
	// memory:
	//   total-ram: 62.4G
	//   total-swap: "0"
	// disk:
	//   /var/lib/snapd/snaps:
	//     total: 937.3G
	//     avail: 878.7G
	// devices:
	//   - bus: pci
	//     slot: '0000:00:00.0'
	//     bus-number: "0x0"
	//     device-class: "0x600"
	//     vendor-id: "0x8086"
	//     device-id: "0x4637"
	//     subvendor-id: "0x103C"
	//     subdevice-id: "0x89C6"
	//   - bus: pci
	//     slot: '0000:00:02.0'
	//     bus-number: "0x0"
	//     device-class: "0x300"
	//     vendor-id: "0x8086"
	//     device-id: "0x9B41"
	//     subvendor-id: "0x1028"
	//     subdevice-id: "0x962"
	//     additional-properties:
	//       vram: 13.5G
	//   - bus: pci
	//     slot: '0000:01:00.0'
	//     bus-number: "0x1"
	//     device-class: "0x300"
	//     vendor-id: "0x10DE"
	//     device-id: "0x1B06"
	//     subvendor-id: "0x10DE"
	//     subdevice-id: "0x1B06"
	//     additional-properties:
	//       vram: 11.0G
	//       compute-capability: "6.1"
	//   - bus: pci
	//     slot: '0000:03:00.0'
	//     bus-number: "0x3"
	//     device-class: "0x300"
	//     vendor-id: "0x1002"
	//     device-id: "0x73E1"
	//     subvendor-id: "0x103C"
	//     subdevice-id: "0x89C6"
	//     additional-properties:
	//       microarchitecture: gfx1032
	//       vram: 8.0G
}
