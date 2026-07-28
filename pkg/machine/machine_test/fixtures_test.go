package machine_test

import "github.com/canonical/lscompute/pkg/machine"

// Expected maps each test_data/machines directory name to the machine.Machine
// value that Get is expected to produce for that fake host.
var Expected = map[string]machine.Machine{
	"ampere-one-m-banshee-12":            AmpereOneMBanshee12,
	"ampere-one-siryn":                   AmpereOneSiryn,
	"ampere-one-x-banshee-8":             AmpereOneXBanshee8,
	"asus-ux301l":                        AsusUx301l,
	"dummy-machine":                      DummyMachine,
	"hp-pavilion-15-cs-3037nl":           HpPavilion15Cs3037nl,
	"hp-proliant-rl300-gen11-altra":      HpProliantRl300Gen11Altra,
	"hp-proliant-rl300-gen11-altra-max":  HpProliantRl300Gen11AltraMax,
	"hp-zbook-i712850HX+RadeonPROW6600M": HpZbookI712850HXRadeonPROW6600M,
	"hp-zbook-power-16-inch-g11":         HpZbookPower16InchG11,
	"i5-3570k+arc-a580+gtx1080ti":        I53570kArcA580Gtx1080ti,
	"i7-10510U":                          I710510U,
	"i7-1165G7":                          I71165G7,
	"i7-2600k+arc-a580":                  I72600kArcA580,
	"intel-core-ultra-7-256v":            IntelCoreUltra7256v,
	"lenovo-thinkpad-p16s":               LenovoThinkpadP16s,
	"mustang":                            Mustang,
	"orange-pi-rv2":                      OrangePiRv2,
	"raspberry-pi-5":                     RaspberryPi5,
	"raspberry-pi-5+hailo-8":             RaspberryPi5Hailo8,
	"system76-addw4":                     System76Addw4,
	"xps13-7390":                         Xps137390,
	"xps13-9350":                         Xps139350,
}
