package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/canonical/inference-snaps-cli/pkg/hardware_info"
)

func main() {
	// Do not print timestamps
	log.SetFlags(0)

	output, warnings, err := hardware_info.Get(true)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	for _, warning := range warnings {
		log.Printf("Warning: %s", warning)
	}

	jsonOutput, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		log.Fatalf("Error marshalling output to JSON: %s", err)
	}

	fmt.Println(string(jsonOutput))
}
