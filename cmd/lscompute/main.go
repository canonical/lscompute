package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/canonical/lscompute/pkg/machine"
	"github.com/canonical/lscompute/pkg/machine/host"
)

func main() {
	log.SetFlags(0) // no timestamps

	format := flag.String("format", string(machine.FormatPlain), "output serialization format: plain or json")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n\nOptions:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	output, warnings, err := machine.Get(host.Real(), true)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	for _, warning := range warnings {
		log.Printf("Warning: %s", warning)
	}

	rendered, err := machine.Marshal(output, machine.Format(*format))
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	fmt.Println(string(rendered))
}
