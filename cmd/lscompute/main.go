package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/canonical/lscompute/pkg/machine"
)

func main() {
	log.SetFlags(0) // no timestamps

	output, warnings, err := machine.Get(true)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	for _, warning := range warnings {
		log.Printf("Warning: %s", warning)
	}

	jsonOutput, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		log.Fatalf("Error: marshalling to JSON: %s", err)
	}

	fmt.Println(string(jsonOutput))
}
