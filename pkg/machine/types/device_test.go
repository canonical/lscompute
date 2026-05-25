package types

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestDeviceMarshalJSON verifies that a USB device does not leak PCI fields
// and that a PCI device does not leak USB fields when marshalled to JSON.
func TestDeviceMarshalJSON(t *testing.T) {
	usbDevice := DeviceInfo{
		Bus: "usb",
		UsbDevice: UsbDevice{
			BusNumber:    1,
			DeviceNumber: 16,
			VendorId:     0x0bda,
			ProductId:    0x5487,
			ProductName:  new("Realtek Semiconductor Corp. Dell dock"),
		},
	}

	data, err := json.Marshal(usbDevice)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}
	jsonStr := string(data)
	t.Logf("USB device JSON: %s", jsonStr)

	pciOnlyFields := []string{`"slot"`, `"device-class"`, `"device-id"`}
	for _, field := range pciOnlyFields {
		if strings.Contains(jsonStr, field) {
			t.Errorf("USB device JSON contains PCI-only field %s: %s", field, jsonStr)
		}
	}

	usbExpectedFields := []string{`"bus":"usb"`, `"product-id"`, `"product-name"`}
	for _, field := range usbExpectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("USB device JSON is missing expected field %s: %s", field, jsonStr)
		}
	}

	// Round-trip: unmarshal back and check
	var decoded DeviceInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}
	if decoded.Bus != "usb" {
		t.Errorf("expected bus=usb, got %q", decoded.Bus)
	}
	if decoded.UsbDevice.ProductId != 0x5487 {
		t.Errorf("expected product-id=0x5487, got %v", decoded.UsbDevice.ProductId)
	}

	// PCI device: USB fields must not appear
	pciDevice := DeviceInfo{
		Bus: "pci",
		PciDevice: PciDevice{
			Slot:        "0000:00:02.0",
			VendorId:    0x8086,
			DeviceId:    0x1234,
			DeviceClass: 0x0300,
		},
	}
	pciData, err := json.Marshal(pciDevice)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}
	pciJsonStr := string(pciData)
	t.Logf("PCI device JSON: %s", pciJsonStr)

	usbOnlyFields := []string{`"product-id"`, `"product-name"`, `"device-number"`}
	for _, field := range usbOnlyFields {
		if strings.Contains(pciJsonStr, field) {
			t.Errorf("PCI device JSON contains USB-only field %s: %s", field, pciJsonStr)
		}
	}
}
