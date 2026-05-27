package fastrpc

import "github.com/canonical/lscompute/pkg/machine/types"

/*
Devices returns a slice of FastRPC devices that are detected on the current system.
*/
func Devices(friendlyNames bool) ([]types.FastRpc, []string, error) {
	// Not implemented
	var devices []types.FastRpc

	return devices, []string{}, nil
}
