package fastrpc

// Device represents a single FastRPC device detected on the system.
type Device struct {
	// TODO: add FastRPC-specific fields (e.g. domain, instance, subsystem)

	// Vendor specific device key-value pairs
	AdditionalProperties map[string]string `json:"additional-properties,omitempty"`
}

