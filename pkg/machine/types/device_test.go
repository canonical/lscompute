package types

import (
	"encoding/json"
	"strings"
	"testing"
)

// --- stub bus device types for testing ---

type stubUsbDevice struct {
	BusNumber    int    `json:"bus-number"`
	DeviceNumber int    `json:"device-number"`
	VendorId     int    `json:"vendor-id"`
	ProductId    int    `json:"product-id"`
	ProductName  string `json:"product-name,omitempty"`
}

func (d *stubUsbDevice) BusName() string { return "usb" }

type stubPciDevice struct {
	Slot        string `json:"slot"`
	VendorId    int    `json:"vendor-id"`
	DeviceId    int    `json:"device-id"`
	DeviceClass int    `json:"device-class"`
}

func (d *stubPciDevice) BusName() string { return "pci" }

func init() {
	RegisterBusDecoder("usb", func(data []byte) (BusDevice, error) {
		var d stubUsbDevice
		return &d, json.Unmarshal(data, &d)
	})
	RegisterBusDecoder("pci", func(data []byte) (BusDevice, error) {
		var d stubPciDevice
		return &d, json.Unmarshal(data, &d)
	})
}

// TestDeviceMarshalJSON verifies that a USB device does not leak PCI fields
// and that a PCI device does not leak USB fields when marshalled to JSON.
func TestDeviceMarshalJSON(t *testing.T) {
	productName := "Realtek Semiconductor Corp. Dell dock"
	usbDevice := DeviceInfo{
		Bus: "usb",
		Payload: &stubUsbDevice{
			BusNumber:    1,
			DeviceNumber: 16,
			VendorId:     0x0bda,
			ProductId:    0x5487,
			ProductName:  productName,
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

	// PCI device: USB fields must not appear
	pciDevice := DeviceInfo{
		Bus: "pci",
		Payload: &stubPciDevice{
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

// TestDeviceMarshalJSON_NilPayload verifies that MarshalJSON returns an error
// when the Payload field is nil.
func TestDeviceMarshalJSON_NilPayload(t *testing.T) {
	di := DeviceInfo{Bus: "usb", Payload: nil}
	_, err := di.MarshalJSON()
	if err == nil {
		t.Fatal("expected error for nil Payload, got nil")
	}
}

// TestDeviceUnmarshalJSON_UnknownBus verifies that UnmarshalJSON returns an
// error when no decoder is registered for the bus type.
func TestDeviceUnmarshalJSON_UnknownBus(t *testing.T) {
	data := []byte(`{"bus":"unknown-bus-xyz","vendor-id":1}`)
	var di DeviceInfo
	if err := di.UnmarshalJSON(data); err == nil {
		t.Fatal("expected error for unknown bus, got nil")
	}
}

// TestDeviceUnmarshalJSON_MalformedJSON verifies that UnmarshalJSON returns an
// error when the JSON is not valid.
func TestDeviceUnmarshalJSON_MalformedJSON(t *testing.T) {
	data := []byte(`not-valid-json`)
	var di DeviceInfo
	if err := di.UnmarshalJSON(data); err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}
