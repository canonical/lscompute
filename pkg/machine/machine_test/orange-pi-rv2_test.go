package machine_test

import (
	"github.com/canonical/lscompute/pkg/machine"
	"github.com/canonical/lscompute/pkg/machine/cpu"
	"github.com/canonical/lscompute/pkg/machine/device/pci"
	"github.com/canonical/lscompute/pkg/machine/disk"
	"github.com/canonical/lscompute/pkg/machine/memory"
)

// OrangePiRv2 is the expected machine info for the "orange-pi-rv2" test fixture.
var OrangePiRv2 = register("orange-pi-rv2", machine.Machine{
	CPUs: []cpu.CPU{
		{
			Architecture: "riscv64",
			Isa:          []string{"rv64imafdcv", "zicbom", "zicboz", "zicntr", "zicond", "zicsr", "zifencei", "zihintpause", "zihpm", "zfh", "zfhmin", "zca", "zcd", "zba", "zbb", "zbc", "zbs", "zkt", "zve32f", "zve32x", "zve64d", "zve64f", "zve64x", "zvfh", "zvfhmin", "zvkt", "sscofpmf", "sstc", "svinval", "svnapot", "svpbmt"},
		},
	},
	Memory: memory.Memory{TotalRam: 8323276800, TotalSwap: 0},
	Disk: []disk.Disk{
		{MountPoint: new("/"), Path: "/var/lib/snapd/snaps", Total: 62270910464, Available: 48009973760},
	},
	PCIDevices: []pci.PCIDevice{
		{
			Bus:         "pci",
			Slot:        "0001:00:00.0",
			BusNumber:   0x0,
			DeviceClass: 0x604,
			VendorId:    0x201F,
			DeviceId:    0x1,
			SubvendorId: new(uint64(0x0)),
			SubdeviceId: new(uint64(0x0)),
		},
		{
			Bus:                  "pci",
			Slot:                 "0001:01:00.0",
			BusNumber:            0x1,
			DeviceClass:          0x108,
			ProgrammingInterface: new(uint8(2)),
			VendorId:             0x15B7,
			DeviceId:             0x5041,
			SubvendorId:          new(uint64(0x15B7)),
			SubdeviceId:          new(uint64(0x5041)),
		},
		{
			Bus:         "pci",
			Slot:        "0002:00:00.0",
			BusNumber:   0x0,
			DeviceClass: 0x604,
			VendorId:    0x201F,
			DeviceId:    0x1,
			SubvendorId: new(uint64(0x0)),
			SubdeviceId: new(uint64(0x0)),
		},
	},
})
