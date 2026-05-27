package amd

import (
	"fmt"

	"github.com/canonical/lscompute/pkg/machine/host"
)

// AdditionalProperties returns device specific properties as a map[string]string.
// No error is returned as a failure to look up properties is considered non-fatal, and likely due to missing drivers.
// Any errors are logged to STDERR.
func AdditionalProperties(h host.Host, slot string, isGpu bool) (map[string]string, error) {
	var properties map[string]string
	var err error

	if isGpu {
		properties, err = gpuProperties(h, slot)
		if err != nil {
			return nil, fmt.Errorf("getting gpu properties: %w", err)
		}
	}

	// Future: handle other AMD device classes

	return properties, nil
}
