package device

import (
	"testing"

	"github.com/canonical/lscompute/pkg/machine/device/apusys"
	"github.com/canonical/lscompute/pkg/machine/device/pci"
	"github.com/canonical/lscompute/pkg/machine/device/usb"
)

func TestDecodeDeviceInfo(t *testing.T) {
	t.Run("decodes pci device", func(t *testing.T) {
		data := []byte(`{"bus":"pci","slot":"0000:00:02.0","bus-number":"0x00","device-class":"0x0300","vendor-id":"0x8086","device-id":"0x5916"}`)
		dev, err := Decode(data)
		if err != nil {
			t.Fatalf("Decode() error: %v", err)
		}
		pciDev, ok := dev.(*pci.Device)
		if !ok {
			t.Fatalf("expected *pci.Device, got %T", dev)
		}
		if pciDev.Bus != pci.BusName {
			t.Fatalf("Bus = %q, want %q", pciDev.Bus, pci.BusName)
		}
	})

	t.Run("decodes usb device", func(t *testing.T) {
		data := []byte(`{"bus":"usb","bus-number":1,"device-number":2,"vendor-id":"0x0bda","product-id":"0x5487"}`)
		dev, err := Decode(data)
		if err != nil {
			t.Fatalf("Decode() error: %v", err)
		}
		usbDev, ok := dev.(*usb.Device)
		if !ok {
			t.Fatalf("expected *usb.Device, got %T", dev)
		}
		if usbDev.Bus != usb.BusName {
			t.Fatalf("Bus = %q, want %q", usbDev.Bus, usb.BusName)
		}
	})

	t.Run("decodes apusys device", func(t *testing.T) {
		data := []byte(`{"bus":"apusys","type":"mdla","soc-id":"jep106:0426:8195","chip-model":"MT8195 / MT8395","product-family":"Kompanio 1200 / Genio 1200","npu-architecture":"4.0 TOPS (Dual-core APU)"}`)
		dev, err := Decode(data)
		if err != nil {
			t.Fatalf("Decode() error: %v", err)
		}
		apusysDev, ok := dev.(*apusys.Device)
		if !ok {
			t.Fatalf("expected *apusys.Device, got %T", dev)
		}
		if apusysDev.Bus != apusys.BusName {
			t.Fatalf("Bus = %q, want %q", apusysDev.Bus, apusys.BusName)
		}
		if apusysDev.SocID != "jep106:0426:8195" {
			t.Fatalf("SocID = %q, want %q", apusysDev.SocID, "jep106:0426:8195")
		}
		if apusysDev.SocInfo.ChipModel != "MT8195 / MT8395" {
			t.Fatalf("ChipModel = %q, want %q", apusysDev.SocInfo.ChipModel, "MT8195 / MT8395")
		}
	})

	t.Run("unknown bus returns error", func(t *testing.T) {
		_, err := Decode([]byte(`{"bus":"unknown","vendor-id":1}`))
		if err == nil {
			t.Fatal("expected error for unknown bus, got nil")
		}
	})
}
