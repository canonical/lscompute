package device

import (
	"testing"

	"github.com/canonical/lscompute/pkg/machine/constants"
	"github.com/canonical/lscompute/pkg/machine/device/pci"
	"github.com/canonical/lscompute/pkg/machine/device/usb"
)

func TestDecodeDeviceInfo(t *testing.T) {
	t.Run("decodes pci device", func(t *testing.T) {
		data := []byte(`{"bus":"pci","slot":"0000:00:02.0","bus-number":"0x00","device-class":"0x0300","vendor-id":"0x8086","device-id":"0x5916"}`)
		dev, err := DecodeDeviceInfo(data)
		if err != nil {
			t.Fatalf("DecodeDeviceInfo() error: %v", err)
		}
		if dev.Bus != constants.BusPci {
			t.Fatalf("Bus = %q, want %q", dev.Bus, constants.BusPci)
		}
		if _, ok := dev.Payload.(*pci.Device); !ok {
			t.Fatalf("Payload type = %T, want *pci.Device", dev.Payload)
		}
	})

	t.Run("decodes usb device", func(t *testing.T) {
		data := []byte(`{"bus":"usb","bus-number":1,"device-number":2,"vendor-id":"0x0bda","product-id":"0x5487"}`)
		dev, err := DecodeDeviceInfo(data)
		if err != nil {
			t.Fatalf("DecodeDeviceInfo() error: %v", err)
		}
		if dev.Bus != constants.BusUsb {
			t.Fatalf("Bus = %q, want %q", dev.Bus, constants.BusUsb)
		}
		if _, ok := dev.Payload.(*usb.Device); !ok {
			t.Fatalf("Payload type = %T, want *usb.Device", dev.Payload)
		}
	})

	t.Run("unknown bus returns error", func(t *testing.T) {
		_, err := DecodeDeviceInfo([]byte(`{"bus":"unknown","vendor-id":1}`))
		if err == nil {
			t.Fatal("expected error for unknown bus, got nil")
		}
	})
}
