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

	format := flag.String("format", FormatPlain, "output serialization format: plain or json")
	all := flag.Bool("all", false, "List all devices, even those that are not useful for computational tasks.")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n\nOptions:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	output, warnings, err := machine.Get(host.Real(), true, *all)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	for _, warning := range warnings {
		log.Printf("Warning: %s", warning)
	}

	machineDetails := NewMachineDetails(output)
	b, err := machineDetails.Marshal(*format)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	fmt.Print(string(b))
}
