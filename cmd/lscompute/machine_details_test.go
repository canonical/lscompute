package main

import (
	"fmt"

	"github.com/canonical/lscompute/pkg/machine"
	"github.com/canonical/lscompute/pkg/machine/cpu"
	"github.com/canonical/lscompute/pkg/machine/device/pci"
	"github.com/canonical/lscompute/pkg/machine/disk"
	"github.com/canonical/lscompute/pkg/machine/memory"
	"github.com/canonical/lscompute/pkg/machine/types"
)

func machineInfoForExamples() *machine.MachineInfo {
	return &machine.MachineInfo{
		Cpus: []cpu.CpuInfo{{
			Architecture:   "amd64",
			ManufacturerId: "GenuineIntel",
			Flags:          []string{"fpu", "vme", "de"},
		}},
		Memory: memory.MemoryInfo{
			TotalRam:  67012501504,
			TotalSwap: 0,
		},
		Disk: map[string]disk.DirInfo{
			"/var/lib/snapd/snaps": {
				Total: 1006451294208,
				Avail: 943543738368,
			},
		},
		Devices: []any{
			pci.Device{
				Bus:         pci.BusName,
				Slot:        "0000:00:00.0",
				BusNumber:   0x0,
				DeviceClass: 0x600,
				VendorId:    0x8086,
				DeviceId:    0x4637,
				SubvendorId: new(types.HexInt(0x103C)),
				SubdeviceId: new(types.HexInt(0x89C6)),
			},
			pci.Device{
				Bus:         pci.BusName,
				Slot:        "0000:00:02.0",
				BusNumber:   0x0,
				DeviceClass: 0x300,
				VendorId:    0x8086,
				DeviceId:    0x9B41,
				SubvendorId: new(types.HexInt(0x1028)),
				SubdeviceId: new(types.HexInt(0x962)),
				AdditionalProperties: map[string]string{
					"vram": "14477950976",
				},
			},
			pci.Device{
				Bus:         pci.BusName,
				Slot:        "0000:01:00.0",
				BusNumber:   0x1,
				DeviceClass: 0x300,
				VendorId:    0x10DE,
				DeviceId:    0x1B06,
				SubvendorId: new(types.HexInt(0x10DE)),
				SubdeviceId: new(types.HexInt(0x1B06)),
				AdditionalProperties: map[string]string{
					"vram":               "11811160064",
					"compute-capability": "6.1",
				},
			},
			pci.Device{
				Bus:         pci.BusName,
				Slot:        "0000:03:00.0",
				BusNumber:   0x3,
				DeviceClass: 0x300,
				VendorId:    0x1002,
				DeviceId:    0x73E1,
				SubvendorId: new(types.HexInt(0x103C)),
				SubdeviceId: new(types.HexInt(0x89C6)),
				AdditionalProperties: map[string]string{
					"microarchitecture": "gfx1032",
					"vram":              "8573157376",
				},
			},
		},
	}
}

func Example_marshalJson() {
	machineInfo := machineInfoForExamples()
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
	machineInfo := machineInfoForExamples()
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
	//   total-swap: 0
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
