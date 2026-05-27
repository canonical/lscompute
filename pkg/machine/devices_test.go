package machine

import (
	"encoding/json"
	"testing"

	"github.com/canonical/lscompute/pkg/machine/types"
)

func TestMarshalUnmarshalDevices(t *testing.T) {
	machineDevices, warnings, err := Devices(true)
	if err != nil {
		t.Fatalf("Failed to marshal devices: %v", err)
	}
	t.Log(warnings)

	jsonStr, err := json.Marshal(machineDevices)
	if err != nil {
		t.Fatalf("Failed to marshal devices: %v", err)
	}

	var newDevices []types.DeviceInfo
	if err := json.Unmarshal(jsonStr, &newDevices); err != nil {
		t.Fatalf("Failed to unmarshal devices: %v", err)
	}
}
