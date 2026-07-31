package machine_test

import (
	"github.com/canonical/lscompute/pkg/machine"
	"github.com/canonical/lscompute/pkg/machine/cpu"
	"github.com/canonical/lscompute/pkg/machine/device/pci"
	"github.com/canonical/lscompute/pkg/machine/disk"
	"github.com/canonical/lscompute/pkg/machine/memory"
)

func init() {
	register("asus-ux301l", machine.Machine{
		CPUs: []cpu.CPU{
			{
				Architecture:   "amd64",
				ManufacturerId: "GenuineIntel",
				Flags:          []string{"fpu", "vme", "de", "pse", "tsc", "msr", "pae", "mce", "cx8", "apic", "sep", "mtrr", "pge", "mca", "cmov", "pat", "pse36", "clflush", "dts", "acpi", "mmx", "fxsr", "sse", "sse2", "ss", "ht", "tm", "pbe", "syscall", "nx", "pdpe1gb", "rdtscp", "lm", "constant_tsc", "arch_perfmon", "pebs", "bts", "rep_good", "nopl", "xtopology", "nonstop_tsc", "cpuid", "aperfmperf", "pni", "pclmulqdq", "dtes64", "monitor", "ds_cpl", "vmx", "est", "tm2", "ssse3", "sdbg", "fma", "cx16", "xtpr", "pdcm", "pcid", "sse4_1", "sse4_2", "x2apic", "movbe", "popcnt", "tsc_deadline_timer", "aes", "xsave", "avx", "f16c", "rdrand", "lahf_lm", "abm", "cpuid_fault", "epb", "pti", "ssbd", "ibrs", "ibpb", "stibp", "tpr_shadow", "flexpriority", "ept", "vpid", "ept_ad", "fsgsbase", "tsc_adjust", "bmi1", "avx2", "smep", "bmi2", "erms", "invpcid", "xsaveopt", "dtherm", "ida", "arat", "pln", "pts", "vnmi", "md_clear", "flush_l1d"},
			},
		},
		Memory: memory.Memory{TotalRam: 7663951872, TotalSwap: 4294963200},
		Disk: []disk.Disk{
			{MountPoint: new("/"), Path: "/fakehost/var/lib/snapd/snaps", Total: 214748364800, Available: 53687091200},
		},
		PCIDevices: []pci.Device{
			{
				Bus:         "pci",
				Slot:        "0000:00:00.0",
				BusNumber:   0x0,
				DeviceClass: 0x600,
				VendorId:    0x8086,
				DeviceId:    0xA04,
				SubvendorId: new(uint16(0x1043)),
				SubdeviceId: new(uint16(0x13BD)),
			},
			{
				Bus:         "pci",
				Slot:        "0000:00:02.0",
				BusNumber:   0x0,
				DeviceClass: 0x300,
				VendorId:    0x8086,
				DeviceId:    0xA2E,
				SubvendorId: new(uint16(0x1043)),
				SubdeviceId: new(uint16(0x13BD)),
			},
			{
				Bus:         "pci",
				Slot:        "0000:00:03.0",
				BusNumber:   0x0,
				DeviceClass: 0x403,
				VendorId:    0x8086,
				DeviceId:    0xA0C,
				SubvendorId: new(uint16(0x8086)),
				SubdeviceId: new(uint16(0x2010)),
			},
			{
				Bus:         "pci",
				Slot:        "0000:00:04.0",
				BusNumber:   0x0,
				DeviceClass: 0x1180,
				VendorId:    0x8086,
				DeviceId:    0xA03,
				SubvendorId: new(uint16(0x8086)),
				SubdeviceId: new(uint16(0x2010)),
			},
			{
				Bus:                  "pci",
				Slot:                 "0000:00:14.0",
				BusNumber:            0x0,
				DeviceClass:          0xC03,
				ProgrammingInterface: new(uint8(48)),
				VendorId:             0x8086,
				DeviceId:             0x9C31,
				SubvendorId:          new(uint16(0x1043)),
				SubdeviceId:          new(uint16(0x201F)),
			},
			{
				Bus:         "pci",
				Slot:        "0000:00:16.0",
				BusNumber:   0x0,
				DeviceClass: 0x780,
				VendorId:    0x8086,
				DeviceId:    0x9C3A,
				SubvendorId: new(uint16(0x1043)),
				SubdeviceId: new(uint16(0x13BD)),
			},
			{
				Bus:         "pci",
				Slot:        "0000:00:1b.0",
				BusNumber:   0x0,
				DeviceClass: 0x403,
				VendorId:    0x8086,
				DeviceId:    0x9C20,
				SubvendorId: new(uint16(0x1043)),
				SubdeviceId: new(uint16(0x13BD)),
			},
			{
				Bus:         "pci",
				Slot:        "0000:00:1c.0",
				BusNumber:   0x0,
				DeviceClass: 0x604,
				VendorId:    0x8086,
				DeviceId:    0x9C10,
				SubvendorId: new(uint16(0x1043)),
				SubdeviceId: new(uint16(0x13BD)),
			},
			{
				Bus:         "pci",
				Slot:        "0000:00:1c.3",
				BusNumber:   0x0,
				DeviceClass: 0x604,
				VendorId:    0x8086,
				DeviceId:    0x9C16,
				SubvendorId: new(uint16(0x1043)),
				SubdeviceId: new(uint16(0x13BD)),
			},
			{
				Bus:         "pci",
				Slot:        "0000:00:1f.0",
				BusNumber:   0x0,
				DeviceClass: 0x601,
				VendorId:    0x8086,
				DeviceId:    0x9C43,
				SubvendorId: new(uint16(0x1043)),
				SubdeviceId: new(uint16(0x13BD)),
			},
			{
				Bus:                  "pci",
				Slot:                 "0000:00:1f.2",
				BusNumber:            0x0,
				DeviceClass:          0x106,
				ProgrammingInterface: new(uint8(1)),
				VendorId:             0x8086,
				DeviceId:             0x9C03,
				SubvendorId:          new(uint16(0x1043)),
				SubdeviceId:          new(uint16(0x13BD)),
			},
			{
				Bus:         "pci",
				Slot:        "0000:00:1f.3",
				BusNumber:   0x0,
				DeviceClass: 0xC05,
				VendorId:    0x8086,
				DeviceId:    0x9C22,
				SubvendorId: new(uint16(0x1043)),
				SubdeviceId: new(uint16(0x13BD)),
			},
			{
				Bus:         "pci",
				Slot:        "0000:02:00.0",
				BusNumber:   0x2,
				DeviceClass: 0x280,
				VendorId:    0x8086,
				DeviceId:    0x8B1,
				SubvendorId: new(uint16(0x8086)),
				SubdeviceId: new(uint16(0xC070)),
			},
		},
	})
}
