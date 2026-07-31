package machine_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/canonical/lscompute/pkg/machine"
	"github.com/canonical/lscompute/pkg/machine/host"
)

var expected = map[string]machine.Machine{}

func register(dir string, m machine.Machine) {
	expected[dir] = m
}

func TestGet_AllFakeHosts(t *testing.T) {
	machinesDir := filepath.Join("..", "..", "..", "test_data", "machines")
	entries, err := os.ReadDir(machinesDir)
	if err != nil {
		t.Fatalf("reading %s: %v", machinesDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()

		t.Run(name, func(t *testing.T) {
			machineRoot := filepath.Join(machinesDir, name, "machine-root")
			if _, err := os.Stat(machineRoot); err != nil {
				t.Fatalf("machine-root missing for %s: %v", name, err)
			}

			// Enable friendly names only when a curated pci.ids is present,
			// matching how the expected fixtures were produced.
			pciIDs := filepath.Join(machineRoot, "usr", "share", "misc", "pci.ids")
			_, pciErr := os.Stat(pciIDs)
			friendlyNames := pciErr == nil

			got, _, err := machine.Get(host.Fake(machineRoot), friendlyNames)
			if err != nil {
				t.Fatalf("Get() failed: %v", err)
			}

			want, ok := expected[name]
			if !ok {
				// No expected fixture defined for this machine: the Get call
				// above still exercises the full pipeline, so a parse failure
				// or panic would already have failed the test.
				t.Skipf("no expected fixture defined for %q; ran parse-only check", name)
			}

			if !reflect.DeepEqual(*got, want) {
				gotJSON, _ := json.MarshalIndent(got, "", "  ")
				wantJSON, _ := json.MarshalIndent(want, "", "  ")
				t.Errorf("machine %s does not match expected fixture\n\nGOT:\n%s\n\nWANT:\n%s", name, gotJSON, wantJSON)
			}
		})
	}
}
