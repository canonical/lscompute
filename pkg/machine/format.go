package machine

import "errors"

// Marshal is intentionally disabled. The core MachineInfo type must not be
// serialized directly: presentation concerns (custom, human-readable
// formatting) live in the visualization package. Convert the value with
// visualization.New and serialize it with visualization.Marshal instead.
func Marshal(*MachineInfo) ([]byte, error) {
	return nil, errors.New("marshalling machine.MachineInfo directly is not supported: convert it with visualization.New and use visualization.Marshal")
}
