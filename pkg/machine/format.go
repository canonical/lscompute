package machine

import (
	"bytes"
	"encoding/json"
	"fmt"

	"go.yaml.in/yaml/v4"
)

// Format is an output serialization format for machine information.
type Format string

const (
	// FormatPlain is a human-readable YAML output.
	FormatPlain Format = "plain"
	// FormatJSON is a machine-readable, indented JSON output with kebab-cased keys.
	FormatJSON Format = "json"
)

// Marshal serializes the given machine information using the requested format.
// It returns an error if the format is not recognized or if info is nil.
func Marshal(info *MachineInfo, f Format) ([]byte, error) {
	if info == nil {
		return nil, fmt.Errorf("cannot marshal nil machine info")
	}
	switch f {
	case FormatJSON:
		return json.MarshalIndent(info, "", "  ")
	case FormatPlain:
		return marshalPlain(info)
	default:
		return nil, fmt.Errorf("unknown format %q (choices: plain, json)", f)
	}
}

// marshalPlain renders machine information as human-readable YAML using a
// two-space indent.
func marshalPlain(info *MachineInfo) ([]byte, error) {
	var b bytes.Buffer
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(2)
	if err := enc.Encode(info); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
