package types

type PciDevice struct {
	Slot                 string  `json:"slot"`
	BusNumber            HexInt  `json:"bus-number"`
	DeviceClass          HexInt  `json:"device-class"`
	ProgrammingInterface *uint8  `json:"programming-interface,omitempty"`
	VendorId             HexInt  `json:"vendor-id"`
	DeviceId             HexInt  `json:"device-id"`
	SubvendorId          *HexInt `json:"subvendor-id,omitempty"`
	SubdeviceId          *HexInt `json:"subdevice-id,omitempty"`
	PciFriendlyNames     `json:",inline"`

	// Vendor specific device key-value pairs
	AdditionalProperties map[string]string `json:"additional-properties,omitempty"`
}

// IsGpu reports whether the device is a GPU or display controller by PCI class.
// Covers legacy VGA (0x0001) and the full display-controller class (0x03xx).
func (d PciDevice) IsGpu() bool {
	return d.DeviceClass == 0x0001 || d.DeviceClass&0xFF00 == 0x0300
}

type PciFriendlyNames struct {
	VendorName    *string `json:"vendor-name,omitempty"`
	DeviceName    *string `json:"device-name,omitempty"`
	SubvendorName *string `json:"subvendor-name,omitempty"`
	SubdeviceName *string `json:"subdevice-name,omitempty"`
}

type UsbDevice struct {
	BusNumber        int    `json:"bus-number"`
	DeviceNumber     int    `json:"device-number"`
	VendorId         HexInt `json:"vendor-id"`
	ProductId        HexInt `json:"product-id"`
	UsbFriendlyNames `json:",inline"`

	// Vendor specific device key-value pairs
	AdditionalProperties map[string]string `json:"additional-properties,omitempty"`
}

type UsbFriendlyNames struct {
	VendorName  *string `json:"vendor-name,omitempty"`
	ProductName *string `json:"product-name,omitempty"`
}

type FastRpc struct {
	// TODO

	// Optional: Vendor specific device key-value pairs
	//AdditionalProperties map[string]string `json:"additional-properties,omitempty"`
}
