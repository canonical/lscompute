package types

import (
	"encoding/json"
	"fmt"
	"sync"
)

// BusDevice is implemented by every bus-specific device struct.
// Each bus package defines its own concrete type in its own directory.
type BusDevice interface {
	BusName() string // returns the canonical bus constant, e.g. "pci"
}

// DeviceInfo wraps a bus-specific device payload with its bus name.
type DeviceInfo struct {
	Bus     string    `json:"bus"`
	Payload BusDevice `json:"-"`
}

// MarshalJSON serialises DeviceInfo as a single flat JSON object merging
// {"bus":"<name>"} with the payload's own fields. No switch needed.
func (d DeviceInfo) MarshalJSON() ([]byte, error) {
	if d.Payload == nil {
		return nil, fmt.Errorf("DeviceInfo has nil Payload for bus %q", d.Bus)
	}
	payloadBytes, err := json.Marshal(d.Payload)
	if err != nil {
		return nil, err
	}
	// Build {"bus":"<name>"} and merge with payload object.
	busBytes, err := json.Marshal(map[string]string{"bus": d.Bus})
	if err != nil {
		return nil, err
	}
	// busBytes:     {"bus":"pci"}
	// payloadBytes: {...fields...}
	// merged:       {"bus":"pci",...fields...}
	if string(payloadBytes) == "{}" {
		return busBytes, nil
	}
	// Strip trailing } from busBytes, strip leading { from payloadBytes, join with comma.
	merged := make([]byte, 0, len(busBytes)+len(payloadBytes))
	merged = append(merged, busBytes[:len(busBytes)-1]...)
	merged = append(merged, ',')
	merged = append(merged, payloadBytes[1:]...)
	return merged, nil
}

// UnmarshalJSON peeks at the "bus" field and delegates to the registered decoder.
func (d *DeviceInfo) UnmarshalJSON(data []byte) error {
	var peek struct {
		Bus string `json:"bus"`
	}
	if err := json.Unmarshal(data, &peek); err != nil {
		return err
	}
	d.Bus = peek.Bus

	decodersMu.RLock()
	decode, ok := decoders[peek.Bus]
	decodersMu.RUnlock()
	if !ok {
		return fmt.Errorf("unknown device bus: %q", peek.Bus)
	}
	payload, err := decode(data)
	if err != nil {
		return err
	}
	d.Payload = payload
	return nil
}

// Decoder registry — populated explicitly in devices.go via RegisterBusDecoder.
var (
	decodersMu sync.RWMutex
	decoders   = map[string]func([]byte) (BusDevice, error){}
)

// RegisterBusDecoder teaches UnmarshalJSON how to decode a given bus type.
// Call this explicitly for every registered bus — see pkg/machine/devices.go.
func RegisterBusDecoder(busName string, decode func([]byte) (BusDevice, error)) {
	decodersMu.Lock()
	defer decodersMu.Unlock()
	decoders[busName] = decode
}
