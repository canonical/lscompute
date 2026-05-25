package types

type PciDevice struct {
	Score                int     `json:"score,omitempty"`
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
