package main

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/canonical/lscompute/pkg/machine"
	"github.com/canonical/lscompute/pkg/machine/cpu"
	"github.com/canonical/lscompute/pkg/machine/device/apusys"
	"github.com/canonical/lscompute/pkg/machine/device/fastrpc"
	"github.com/canonical/lscompute/pkg/machine/device/pci"
	"github.com/canonical/lscompute/pkg/machine/device/usb"
	"github.com/canonical/lscompute/pkg/machine/disk"
	"github.com/canonical/lscompute/pkg/machine/memory"
	"go.yaml.in/yaml/v4"
)

func machineInfoForExamples() *machine.Machine {
	return &machine.Machine{
		CPUs: []cpu.CPU{{
			Architecture:   "amd64",
			ManufacturerId: "GenuineIntel",
			Flags:          []string{"fpu", "vme", "de"},
		}},
		Memory: memory.Memory{
			TotalRam:  67012501504,
			TotalSwap: 0,
		},
		Disk: []disk.Disk{{
			MountPoint: new("/"),
			Path:       "/var/lib/snapd/snaps",
			Total:      1006451294208,
			Available:  943543738368,
		},
			{
				Path:      "/home",
				Total:     1000000000000000,
				Available: 40000000,
			}},
		PCIDevices: []pci.Device{
			{
				Bus:         pci.BusName,
				Slot:        "0000:00:00.0",
				BusNumber:   0x0,
				DeviceClass: 0x600,
				VendorId:    0x8086,
				DeviceId:    0x4637,
				SubvendorId: new(uint16(0x103C)),
				SubdeviceId: new(uint16(0x89C6)),
			},
			{
				Bus:         pci.BusName,
				Slot:        "0000:00:02.0",
				BusNumber:   0x0,
				DeviceClass: 0x300,
				VendorId:    0x8086,
				DeviceId:    0x9B41,
				SubvendorId: new(uint16(0x1028)),
				SubdeviceId: new(uint16(0x962)),
				AdditionalProperties: map[string]string{
					"vram": "14477950976",
				},
			},
			{
				Bus:         pci.BusName,
				Slot:        "0000:01:00.0",
				BusNumber:   0x1,
				DeviceClass: 0x300,
				VendorId:    0x10DE,
				DeviceId:    0x1B06,
				SubvendorId: new(uint16(0x10DE)),
				SubdeviceId: new(uint16(0x1B06)),
				AdditionalProperties: map[string]string{
					"vram":               "11811160064",
					"compute-capability": "6.1",
				},
			},
			{
				Bus:         pci.BusName,
				Slot:        "0000:03:00.0",
				BusNumber:   0x3,
				DeviceClass: 0x300,
				VendorId:    0x1002,
				DeviceId:    0x73E1,
				SubvendorId: new(uint16(0x103C)),
				SubdeviceId: new(uint16(0x89C6)),
				AdditionalProperties: map[string]string{
					"microarchitecture": "gfx1032",
					"vram":              "8573157376",
				},
			},
		},
		USBDevices: []usb.Device{
			{
				Bus:          usb.BusName,
				BusNumber:    7,
				DeviceNumber: 1,
				VendorId:     0x1D6B,
				ProductId:    0x2,
				FriendlyNames: usb.FriendlyNames{
					VendorName:  "Linux Foundation",
					ProductName: "2.0 root hub",
				},
			},
		},
		FastRPCDevices: []fastrpc.Device{
			{
				Bus:    fastrpc.BusName,
				Domain: fastrpc.ADSPDomain,
				Index:  0,
				Secure: false,
				AdditionalProperties: map[string]string{
					"test":  "test",
					"test1": "1073741824",
				},
			},
		},
	}
}

func TestNewMachineDetailsIncludesAPUSYS(t *testing.T) {
	info := &machine.Machine{
		APUSYSDevices: []apusys.Device{
			{
				Bus:        apusys.BusName,
				Type:       "mdla",
				SocID:      "mt8395",
				VendorName: "mediatek",
			},
		},
	}

	data, err := NewMachineDetails(info).Marshal(FormatJSON)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	var got struct {
		Devices []struct {
			Bus        string `json:"bus"`
			Type       string `json:"type"`
			SocID      string `json:"soc-id"`
			VendorName string `json:"vendor-name"`
		} `json:"devices"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}
	if len(got.Devices) != 1 {
		t.Fatalf("len(devices) = %d, want 1", len(got.Devices))
	}

	want := struct {
		Bus        string
		Type       string
		SocID      string
		VendorName string
	}{
		Bus:        "apusys",
		Type:       "mdla",
		SocID:      "mt8395",
		VendorName: "mediatek",
	}
	device := got.Devices[0]
	if device.Bus != want.Bus || device.Type != want.Type || device.SocID != want.SocID || device.VendorName != want.VendorName {
		t.Fatalf("APUSYS device = %#v, want %#v", device, want)
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
	//   "disks": [
	//     {
	//       "mount-point": "/",
	//       "path": "/var/lib/snapd/snaps",
	//       "total": 1006451294208,
	//       "avail": 943543738368
	//     },
	//     {
	//       "path": "/home",
	//       "total": 1000000000000000,
	//       "avail": 40000000
	//     }
	//   ],
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
	//     },
	//     {
	//       "bus": "usb",
	//       "bus-number": "0x7",
	//       "device-number": 1,
	//       "vendor-id": "0x1D6B",
	//       "product-id": "0x2",
	//       "vendor-name": "Linux Foundation",
	//       "product-name": "2.0 root hub"
	//     },
	//     {
	//       "bus": "fastrpc",
	//       "domain": "adsp",
	//       "additional-properties": {
	//         "test": "test",
	//         "test1": "1073741824"
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
	// disks:
	//   - mount-point: /
	//     path: /var/lib/snapd/snaps
	//     total: 937.3G
	//     avail: 878.7G
	//   - path: /home
	//     total: 909.5T
	//     avail: 38.1M
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
	//   - bus: usb
	//     bus-number: "0x7"
	//     device-number: 1
	//     vendor-id: "0x1D6B"
	//     product-id: "0x2"
	//     vendor-name: Linux Foundation
	//     product-name: 2.0 root hub
	//   - bus: fastrpc
	//     domain: adsp
	//     additional-properties:
	//       test: test
	//       test1: "1073741824"
}
func TestHexIntMarshalJSON(t *testing.T) {
	tests := []struct {
		val  HexInt
		want string
	}{
		{0x8086, `"0x8086"`},
		{0x0, `"0x0"`},
		{0x10DE, `"0x10DE"`},
		{0x1234, `"0x1234"`},
	}
	for _, tc := range tests {
		data, err := json.Marshal(tc.val)
		if err != nil {
			t.Errorf("Marshal(%#x) error: %v", int(tc.val), err)
			continue
		}
		if string(data) != tc.want {
			t.Errorf("Marshal(%#x) = %s, want %s", int(tc.val), data, tc.want)
		}
	}
}

func TestHexIntUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    HexInt
		wantErr bool
	}{
		{"0x prefix lowercase", `"0x8086"`, 0x8086, false},
		{"no prefix", `"8086"`, 0x8086, false},
		{"uppercase hex digits", `"10DE"`, 0x10DE, false},
		{"zero", `"0x0"`, 0, false},
		{"empty string → zero", `""`, 0, false},
		{"invalid hex", `"ZZZZ"`, 0, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var h HexInt
			err := json.Unmarshal([]byte(tc.input), &h)
			if tc.wantErr {
				if err == nil {
					t.Errorf("UnmarshalJSON(%s): expected error, got nil", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("UnmarshalJSON(%s): unexpected error: %v", tc.input, err)
			}
			if h != tc.want {
				t.Errorf("UnmarshalJSON(%s) = 0x%x, want 0x%x", tc.input, int(h), int(tc.want))
			}
		})
	}
}

// TestHexIntJSONRoundTrip verifies that Marshal → Unmarshal is identity.
func TestHexIntJSONRoundTrip(t *testing.T) {
	vals := []HexInt{0, 0x1, 0x8086, 0x10DE, 0xFFFF}
	for _, v := range vals {
		data, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("Marshal(0x%x): %v", int(v), err)
		}
		var got HexInt
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("Unmarshal(%s): %v", data, err)
		}
		if got != v {
			t.Errorf("round-trip 0x%x: got 0x%x", int(v), int(got))
		}
	}
}

func TestHexIntMarshalYAML(t *testing.T) {
	tests := []struct {
		val  HexInt
		want string
	}{
		{0x8086, "0x8086"},
		{0x0, "0x0"},
		{0x10DE, "0x10DE"},
	}
	for _, tc := range tests {
		val, err := tc.val.MarshalYAML()
		if err != nil {
			t.Errorf("MarshalYAML(%#x) error: %v", int(tc.val), err)
			continue
		}
		got, ok := val.(string)
		if !ok {
			t.Errorf("MarshalYAML(%#x) returned %T, want string", int(tc.val), val)
			continue
		}
		if got != tc.want {
			t.Errorf("MarshalYAML(%#x) = %q, want %q", int(tc.val), got, tc.want)
		}
	}
}

func TestHexIntUnmarshalYAML(t *testing.T) {
	type wrapper struct {
		Val HexInt `yaml:"val"`
	}

	tests := []struct {
		name    string
		yaml    string
		want    HexInt
		wantErr bool
	}{
		{"0x prefix", "val: \"0x8086\"\n", 0x8086, false},
		{"no prefix", "val: \"8086\"\n", 0x8086, false},
		{"uppercase", "val: \"10DE\"\n", 0x10DE, false},
		{"empty string → zero", "val: \"\"\n", 0, false},
		{"invalid hex", "val: \"ZZZZ\"\n", 0, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var w wrapper
			err := yaml.Unmarshal([]byte(tc.yaml), &w)
			if tc.wantErr {
				if err == nil {
					t.Errorf("UnmarshalYAML(%q): expected error, got nil", tc.yaml)
				}
				return
			}
			if err != nil {
				t.Fatalf("UnmarshalYAML(%q): unexpected error: %v", tc.yaml, err)
			}
			if w.Val != tc.want {
				t.Errorf("UnmarshalYAML(%q) = 0x%x, want 0x%x", tc.yaml, int(w.Val), int(tc.want))
			}
		})
	}
}

// TestHexIntYAMLRoundTrip verifies that MarshalYAML → UnmarshalYAML is identity.
func TestHexIntYAMLRoundTrip(t *testing.T) {
	type wrapper struct {
		Val HexInt `yaml:"val"`
	}

	vals := []HexInt{0, 0x1, 0x8086, 0x10DE, 0xFFFF}
	for _, v := range vals {
		w := wrapper{Val: v}
		data, err := yaml.Marshal(w)
		if err != nil {
			t.Fatalf("Marshal(0x%x): %v", int(v), err)
		}
		var got wrapper
		if err := yaml.Unmarshal(data, &got); err != nil {
			t.Fatalf("Unmarshal(%s): %v", data, err)
		}
		if got.Val != v {
			t.Errorf("round-trip 0x%x: got 0x%x", int(v), int(got.Val))
		}
	}
}
