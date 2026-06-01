package bus

import (
	"encoding/json"
	"fmt"
)

// DeviceInfo wraps a bus-specific device payload with its bus name.
type DeviceInfo struct {
	Bus     string `json:"bus"`
	Payload any    `json:"-"`
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
