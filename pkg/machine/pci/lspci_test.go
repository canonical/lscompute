package pci

import (
	"fmt"
	"os"
	"testing"
)

func TestParseLsCpu(t *testing.T) {
	machines, err := subDirectories("../../../test_data/machines")
	if err != nil {
		t.Fatal(err)
	}

	for _, machine := range machines {
		lsPciFile := "../../../test_data/machines/" + machine + "/lspci.txt"
		t.Run(machine, func(t *testing.T) {
			_, err := os.Stat(lsPciFile)
			if err != nil {
				if os.IsNotExist(err) {
					// Device does not have lspci test data, skipping
					return
				} else {
					t.Fatal(err)
				}
			}

			lsPci, err := os.ReadFile(lsPciFile)
			if err != nil {
				t.Fatal(err)
			}

			_, warnings, err := ParseLsPci(string(lsPci), true)
			if len(warnings) > 0 {
				t.Logf("Warnings: %v", warnings)
			}
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
