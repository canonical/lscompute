package machine_test

import (
	"github.com/canonical/lscompute/pkg/machine"
	"github.com/canonical/lscompute/pkg/machine/cpu"
	"github.com/canonical/lscompute/pkg/machine/device/pci"
	"github.com/canonical/lscompute/pkg/machine/disk"
	"github.com/canonical/lscompute/pkg/machine/memory"
)

func init() {
	register("dummy-machine", machine.Machine{
		CPUs: []cpu.CPU{
			{
				Architecture:   "amd64",
				ManufacturerId: "GenuineIntel",
				Flags:          []string{"fpu", "vme", "de"},
			},
		},
		Memory: memory.Memory{TotalRam: 67012501504, TotalSwap: 0},
		Disk: []disk.Disk{
			{MountPoint: new("/"), Path: "/fakehost/var/lib/snapd/snaps", Total: 219902325555200, Available: 54975581388800},
		},
		PCIDevices: []pci.PCIDevice{
			{
				Bus:         "pci",
				Slot:        "0000:00:00.0",
				BusNumber:   0x0,
				DeviceClass: 0x600,
				VendorId:    0x8086,
				DeviceId:    0x4637,
				SubvendorId: new(uint64(0x103C)),
				SubdeviceId: new(uint64(0x89C6)),
			},
		},
	})
}
