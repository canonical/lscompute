package memory

import (
	"log"
	"testing"

	"github.com/canonical/lscompute/pkg/machine/host"
)

func TestInfoFromData(t *testing.T) {
	info, err := Info(host.Real())
	if err != nil {
		t.Fatalf("error getting host memory info: %v", err)
	}
	log.Printf("%+v", info)
}
