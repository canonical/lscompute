package pci

import (
	"testing"

	"github.com/canonical/lscompute/pkg/machine/types"
)

func TestDeviceBusName(t *testing.T) {
	d := &Device{}
	if got := d.BusName(); got != BusName {
		t.Errorf("Device.BusName() = %q, want %q", got, BusName)
	}
}

func TestIsGpu(t *testing.T) {
	cases := []struct {
		name        string
		deviceClass uint64
		want        bool
	}{
		{"VGA legacy (0x0001)", 0x0001, true},
		{"display controller (0x0300)", 0x0300, true},
		{"3D controller (0x0302)", 0x0302, true},
		{"display — any 0x03xx subclass", 0x0301, true},
		{"network controller (0x0200)", 0x0200, false},
		{"storage controller (0x0100)", 0x0100, false},
		{"audio device (0x0403)", 0x0403, false},
		{"USB host controller (0x0c03)", 0x0c03, false},
		{"zero / unclassified (0x0000)", 0x0000, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := Device{DeviceClass: types.HexInt(tc.deviceClass)}
			if got := d.IsGpu(); got != tc.want {
				t.Errorf("Device{DeviceClass: 0x%04x}.IsGpu() = %v, want %v",
					tc.deviceClass, got, tc.want)
			}
		})
	}
}
