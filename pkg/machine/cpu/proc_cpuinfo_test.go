package cpu

import (
	"log"
	"os"
	"testing"

	"github.com/canonical/lscompute/pkg/machine/constants"
)

var procCpuInfoTestFiles = map[string]string{
	"../../../test_data/machines/ampere-one-m-banshee-12/machine-root/proc/cpuinfo":           constants.Arm64,
	"../../../test_data/machines/ampere-one-siryn/machine-root/proc/cpuinfo":                  constants.Arm64,
	"../../../test_data/machines/ampere-one-x-banshee-8/machine-root/proc/cpuinfo":            constants.Arm64,
	"../../../test_data/machines/hp-proliant-rl300-gen11-altra/machine-root/proc/cpuinfo":     constants.Arm64,
	"../../../test_data/machines/hp-proliant-rl300-gen11-altra-max/machine-root/proc/cpuinfo": constants.Arm64,
	"../../../test_data/machines/i7-2600k+arc-a580/machine-root/proc/cpuinfo":                 constants.Amd64,
	"../../../test_data/machines/i7-10510U/machine-root/proc/cpuinfo":                         constants.Amd64,
	"../../../test_data/machines/mustang/machine-root/proc/cpuinfo":                           constants.Amd64,
	"../../../test_data/machines/raspberry-pi-5/machine-root/proc/cpuinfo":                    constants.Arm64,
	"../../../test_data/machines/raspberry-pi-5+hailo-8/machine-root/proc/cpuinfo":            constants.Arm64,
	"../../../test_data/machines/xps13-7390/machine-root/proc/cpuinfo":                        constants.Amd64,
	"../../../test_data/machines/xps13-9350/machine-root/proc/cpuinfo":                        constants.Amd64,
}

func TestParseProcCpuInfo(t *testing.T) {
	for procCpuInfoFile, arch := range procCpuInfoTestFiles {
		t.Run(procCpuInfoFile, func(t *testing.T) {
			procCpuInfoBytes, err := os.ReadFile(procCpuInfoFile)
			if err != nil {
				if os.IsNotExist(err) {
					t.Skipf("fixture not present yet: %s", procCpuInfoFile)
				}
				t.Fatal(err)
			}

			parsed, err := parseProcCpuInfo(string(procCpuInfoBytes), arch)
			if err != nil {
				t.Fatal(err)
			}

			for _, cpuInfo := range parsed {
				log.Printf("%+v", cpuInfo)
			}
		})
	}
}

func TestParseProcCpuInfoAmd64(t *testing.T) {
	cpuInfoData, err := os.ReadFile("../../../test_data/machines/xps13-7390/machine-root/proc/cpuinfo")
	if err != nil {
		if os.IsNotExist(err) {
			t.Skip("fixture not present yet: xps13-7390 machine-root/proc/cpuinfo")
		}
		t.Fatal(err)
	}

	cpuInfos, err := parseProcCpuInfoAmd64(string(cpuInfoData))
	if err != nil {
		t.Fatal(err)
	}

	for _, cpuInfo := range cpuInfos {
		log.Printf("%+v", cpuInfo)
	}
}

func TestParseProcCpuInfoArm64(t *testing.T) {
	cpuInfoData, err := os.ReadFile("../../../test_data/machines/raspberry-pi-5/machine-root/proc/cpuinfo")
	if err != nil {
		if os.IsNotExist(err) {
			t.Skip("fixture not present yet: raspberry-pi-5 machine-root/proc/cpuinfo")
		}
		t.Fatal(err)
	}

	cpuInfos, err := parseProcCpuInfoArm64(string(cpuInfoData))
	if err != nil {
		t.Fatal(err)
	}

	for _, cpuInfo := range cpuInfos {
		log.Printf("%+v", cpuInfo)
	}
}
