package machine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SplitPathIntoDirectories takes a file path and returns a slice of strings containing the individual directory names that makes up the path
func SplitPathIntoDirectories(p string) []string {
	var parts []string
	for {
		dir, file := filepath.Split(p)
		if file != "" {
			parts = append([]string{file}, parts...)
		}
		if dir == "" || dir == "/" || dir == "\\" { // Handle root directory and empty path
			break
		}
		p = strings.TrimSuffix(dir, string(filepath.Separator)) // Remove trailing separator
	}
	return parts
}

func SubDirectories(dirPath string) ([]string, error) {
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
