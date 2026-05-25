package machine

import (
	"encoding/json"
	"testing"

	"github.com/canonical/lscompute/pkg/machine/types"
)

func TestGetCurrentHostMarshalUnmarshal(t *testing.T) {
	machineInfo, _, err := Get(false)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	jsonData, err := json.Marshal(machineInfo)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	var unmarshalled types.MachineInfo
	err = json.Unmarshal(jsonData, &unmarshalled)
	if err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	t.Logf("Successfully marshalled and unmarshalled machine info: %s", string(jsonData))
}
