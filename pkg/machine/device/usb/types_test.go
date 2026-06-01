package usb

import (
	"testing"
)

// TestDeviceBusName verifies that Device.BusName returns the USB bus constant.
func TestDeviceBusName(t *testing.T) {
	d := &Device{}
	if got := d.BusName(); got != BusName {
		t.Errorf("Device.BusName() = %q, want %q", got, BusName)
	}
}
