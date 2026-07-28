package machine_test

import (
	"github.com/canonical/lscompute/pkg/machine"
	"github.com/canonical/lscompute/pkg/machine/cpu"
	"github.com/canonical/lscompute/pkg/machine/device/pci"
	"github.com/canonical/lscompute/pkg/machine/disk"
	"github.com/canonical/lscompute/pkg/machine/memory"
)

// RaspberryPi5Hailo8 is the expected machine info for the "raspberry-pi-5+hailo-8" test fixture.
var RaspberryPi5Hailo8 = register("raspberry-pi-5+hailo-8", machine.Machine{
	CPUs: []cpu.CPU{
		{
			Architecture:  "arm64",
			ImplementerId: 0x41,
			PartNumber:    0xD0B,
			Features:      []string{"fp", "asimd", "evtstrm", "aes", "pmull", "sha1", "sha2", "crc32", "atomics", "fphp", "asimdhp", "cpuid", "asimdrdm", "lrcpc", "dcpop", "asimddp"},
		},
	},
	Memory: memory.Memory{TotalRam: 8317198336, TotalSwap: 1073737728},
	Disk: []disk.Disk{
		{MountPoint: "/var/lib/snapd/snaps", Total: 30777720832, Available: 16953257984},
	},
	PCIDevices: []pci.PCIDevice{
		{
			Bus:           "pci",
			Slot:          "0001:00:00.0",
			BusNumber:     0x0,
			DeviceClass:   0x604,
			VendorId:      0x14E4,
			DeviceId:      0x2712,
			FriendlyNames: pci.FriendlyNames{VendorName: "Broadcom Inc. and subsidiaries", DeviceName: "BCM2712 PCIe Bridge"},
		},
		{
			Bus:           "pci",
			Slot:          "0001:01:00.0",
			BusNumber:     0x1,
			DeviceClass:   0xB40,
			VendorId:      0x1E60,
			DeviceId:      0x2864,
			SubvendorId:   new(uint64(0x1E60)),
			SubdeviceId:   new(uint64(0x2864)),
			FriendlyNames: pci.FriendlyNames{VendorName: "Hailo Technologies Ltd.", DeviceName: "Hailo-8 AI Processor", SubvendorName: "Hailo Technologies Ltd."},
		},
		{
			Bus:           "pci",
			Slot:          "0002:00:00.0",
			BusNumber:     0x0,
			DeviceClass:   0x604,
			VendorId:      0x14E4,
			DeviceId:      0x2712,
			FriendlyNames: pci.FriendlyNames{VendorName: "Broadcom Inc. and subsidiaries", DeviceName: "BCM2712 PCIe Bridge"},
		},
		{
			Bus:           "pci",
			Slot:          "0002:01:00.0",
			BusNumber:     0x1,
			DeviceClass:   0x200,
			VendorId:      0x1DE4,
			DeviceId:      0x1,
			FriendlyNames: pci.FriendlyNames{VendorName: "Raspberry Pi Ltd", DeviceName: "RP1 PCIe 2.0 South Bridge"},
		},
	},
})
