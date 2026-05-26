package machine

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/lscompute/pkg/machine/host"
	"github.com/canonical/lscompute/pkg/machine/types"
)

var update = flag.Bool("update", false, "rewrite golden files instead of asserting")

// TestGetFromMachineDirs is the full-pipeline golden test. It iterates every
// directory under test_data/machines/, runs the entire pipeline against the
// machine-root sub-directory using host.Fake(), and compares against a golden
// lscompute.json if one exists.
//
// Machines without a golden file are still exercised end-to-end — if any
// machine's raw data trips a parser, the test fails even without a golden file.
//
// To add a golden file for a machine, run:
//
//	go test ./pkg/machine -run TestGetFromMachineDirs/my-machine-name -update
//
// and verify the output is correct before committing.
func TestGetFromMachineDirs(t *testing.T) {
	entries, err := os.ReadDir("../../test_data/machines")
	if err != nil {
		t.Fatalf("reading test_data/machines: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		machineName := entry.Name()
		t.Run(machineName, func(t *testing.T) {
			dir := filepath.Join("../../test_data/machines", machineName)
			machineRoot := filepath.Join(dir, "machine-root")

			// machine-root must exist; if not, skip (shouldn't happen after migration).
			if _, err := os.Stat(machineRoot); os.IsNotExist(err) {
				t.Skipf("no machine-root directory found, skipping %s", machineName)
			}

			// During fixture migration some machine-root directories may exist but
			// still miss core proc files required by Get(). Skip those until data
			// is fully captured in the dedicated fixture PR.
			required := []string{
				filepath.Join(machineRoot, "proc", "cpuinfo"),
				filepath.Join(machineRoot, "proc", "meminfo"),
			}
			for _, req := range required {
				if _, err := os.Stat(req); os.IsNotExist(err) {
					t.Skipf("incomplete machine-root fixture, missing %s", req)
				}
			}

			h := host.Fake(machineRoot)

			// Run the full pipeline with friendly names on. Machines without a
			// curated machine-root/usr/share/misc/pci.ids will log warnings; that
			// is intentional and not an assertion.
			got, warnings, err := Get(h, true)
			if err != nil {
				t.Fatalf("Get() failed: %v", err)
			}
			for _, w := range warnings {
				t.Logf("warning: %s", w)
			}

			goldenPath := filepath.Join(dir, "lscompute.json")
			if _, err := os.Stat(goldenPath); os.IsNotExist(err) {
				// No golden file: parse-only check. Confirms no panics/fatal errors.
				return
			}

			if *update {
				writeGolden(t, goldenPath, got)
				return
			}
			assertEqualToGolden(t, goldenPath, got)
		})
	}
}

func writeGolden(t *testing.T, path string, info *types.MachineInfo) {
	t.Helper()
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		t.Fatalf("marshalling golden: %v", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0644); err != nil {
		t.Fatalf("writing golden %s: %v", path, err)
	}
	t.Logf("updated golden: %s", path)
}

func assertEqualToGolden(t *testing.T, path string, got *types.MachineInfo) {
	t.Helper()
	goldenData, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading golden %s: %v", path, err)
	}

	want, err := DecodeMachineInfo(goldenData)
	if err != nil {
		t.Fatalf("parsing golden %s: %v", path, err)
	}

	// Compare via re-marshalling to normalize field ordering.
	gotData, err := json.MarshalIndent(got, "", "  ")
	if err != nil {
		t.Fatalf("marshalling result: %v", err)
	}
	wantData, err := json.MarshalIndent(want, "", "  ")
	if err != nil {
		t.Fatalf("re-marshalling golden: %v", err)
	}

	if string(gotData) != string(wantData) {
		t.Errorf("result does not match golden %s\n\nGOT:\n%s\n\nWANT:\n%s", path, gotData, wantData)
	}
}
