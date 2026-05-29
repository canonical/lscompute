package usb

import (
	"testing"

	"github.com/canonical/lscompute/pkg/machine/device/bus"
)

// TestDeviceBusName verifies that Device.BusName returns the USB bus constant.
func TestDeviceBusName(t *testing.T) {
	d := &Device{}
	if got := d.BusName(); got != bus.BusUsb {
		t.Errorf("Device.BusName() = %q, want %q", got, bus.BusUsb)
	}
}

