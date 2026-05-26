package types

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"
)

func TestParseMachineInfo(t *testing.T) {
	machines, err := subDirectories("../../../test_data/machines")
	if err != nil {
		t.Fatal(err)
	}

	for _, machine := range machines {
		machineInfoFile := "../../test_data/machines/" + machine + "/lscompute.json"
		t.Run(machine, func(t *testing.T) {
			_, err := os.Stat(machineInfoFile)
			if err != nil {
				if os.IsNotExist(err) {
					// Device does not have lscompute.json test data, skipping
					return
				} else {
					t.Fatal(err)
				}
			}

			file, err := os.Open(machineInfoFile)
			if err != nil {
				t.Fatal(err)
			}

			data, err := io.ReadAll(file)
			if err != nil {
				t.Fatal(err)
			}

			var hardwareInfo MachineInfo
			err = json.Unmarshal(data, &hardwareInfo)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func subDirectories(dirPath string) ([]string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var directories []string
	for _, entry := range entries {
		if entry.IsDir() {
			directories = append(directories, entry.Name())
		}
	}
	return directories, nil
}
