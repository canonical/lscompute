package machine_test

import (
	"github.com/canonical/lscompute/pkg/machine"
	"github.com/canonical/lscompute/pkg/machine/cpu"
	"github.com/canonical/lscompute/pkg/machine/device/pci"
	"github.com/canonical/lscompute/pkg/machine/disk"
	"github.com/canonical/lscompute/pkg/machine/memory"
)

// RaspberryPi5 is the expected machine info for the "raspberry-pi-5" test fixture.
var RaspberryPi5 = machine.Machine{
	CPUs: []cpu.CPU{
		{
			Architecture:  "arm64",
			ImplementerId: 0x41,
			PartNumber:    0xD0B,
			Features:      []string{"fp", "asimd", "evtstrm", "aes", "pmull", "sha1", "sha2", "crc32", "atomics", "fphp", "asimdhp", "cpuid", "asimdrdm", "lrcpc", "dcpop", "asimddp"},
		},
	},
	Memory: memory.Memory{TotalRam: 8323276800, TotalSwap: 0},
	Disk: []disk.Disk{
		{MountPoint: "/var/lib/snapd/snaps", Total: 62270910464, Available: 40666140672},
	},
	PCIDevices: []pci.PCIDevice{
		{
			Bus:           "pci",
			Slot:          "0000:00:00.0",
			BusNumber:     0x0,
			DeviceClass:   0x604,
			VendorId:      0x14E4,
			DeviceId:      0x2712,
			FriendlyNames: pci.FriendlyNames{VendorName: "Broadcom Inc. and subsidiaries", DeviceName: "BCM2712 PCIe Bridge"},
		},
		{
			Bus:           "pci",
			Slot:          "0000:01:00.0",
			BusNumber:     0x1,
			DeviceClass:   0x200,
			VendorId:      0x1DE4,
			DeviceId:      0x1,
			FriendlyNames: pci.FriendlyNames{VendorName: "Raspberry Pi Ltd", DeviceName: "RP1 PCIe 2.0 South Bridge"},
		},
	},
}
