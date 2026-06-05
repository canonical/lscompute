package apusys

import (
	"encoding/json"
	"errors"
	"io/fs"
	"strings"

	"github.com/canonical/lscompute/pkg/machine/device/bus"
	"github.com/canonical/lscompute/pkg/machine/host"
)

const BusName = "apusys"

const (
	mdlaDevfreqDir = "sys/devices/platform/soc/soc:mdla_devfreq/devfreq/soc:mdla_devfreq"
	socIDPath      = "sys/bus/soc/devices/soc0/soc_id"
)

// Device represents a MediaTek APUSYS MDLA NPU detected on the system.
type Device struct {
	Bus  string `json:"bus"`
	Type string `json:"type"`

	SocID   string  `json:"soc-id,omitempty"`
	SocInfo SocInfo `json:"soc-info,omitempty"`

	// Vendor specific device key-value pairs.
	AdditionalProperties map[string]string `json:"additional-properties,omitempty"`
}

type SocInfo struct {
	ChipModel       string
	ProductFamily   string
	NPUArchitecture string
}

var socIDInfo = map[string]SocInfo{
	"8195": {
		ChipModel:       "MT8195 / MT8395",
		ProductFamily:   "Kompanio 1200 / Genio 1200",
		NPUArchitecture: "4.0 TOPS (Dual-core APU)",
	},
	"8188": {
		ChipModel:       "MT8188 / MT8390",
		ProductFamily:   "Kompanio 520 / Genio 700",
		NPUArchitecture: "2.0 TOPS (Single-core APU)",
	},
	"8370": {
		ChipModel:       "MT8370",
		ProductFamily:   "Genio 510",
		NPUArchitecture: "1.0 TOPS (Single-core APU)",
	},
	"8365": {
		ChipModel:       "MT8365",
		ProductFamily:   "Genio 350",
		NPUArchitecture: "0.5 TOPS (Single-core APU)",
	},
	"8192": {
		ChipModel:       "MT8192",
		ProductFamily:   "Kompanio 828",
		NPUArchitecture: "Similar to MT8195 (NPU present)",
	},
	"8186": {
		ChipModel:       "MT8186",
		ProductFamily:   "Kompanio 528",
		NPUArchitecture: "Entry-level NPU",
	},
}

// apusys implements bus.Bus for MediaTek APUSYS NPU discovery.
type apusys struct {
	host host.Host
	opts Options
}

// Options holds APUSYS-specific scanner configuration.
type Options struct {
	// e.g. FriendlyName
}

// NewBus returns an APUSYS bus configured with the given options.
func NewBus(host host.Host, opts Options) bus.Bus {
	return &apusys{host: host, opts: opts}
}

// Devices discovers MediaTek APUSYS NPUs exposed through the MDLA devfreq node.
func (bus *apusys) Devices() ([]any, []string, error) {
	present, err := hasMDLADevfreq(bus.host.FS())
	if err != nil {
		return nil, nil, err
	}
	if !present {
		return nil, nil, nil
	}

	socID, err := readSoCID(bus.host.FS())
	if err != nil {
		return nil, nil, err
	}

	device := Device{
		Bus:   BusName,
		Type:  "mdla",
		SocID: socID,
	}
	addSoCInfo(&device)

	return []any{
		device,
	}, nil, nil
}

func Decode(bytes []byte) (*Device, error) {
	var device Device
	if err := json.Unmarshal(bytes, &device); err != nil {
		return nil, err
	}
	return &device, nil
}

func hasMDLADevfreq(fsys fs.FS) (bool, error) {
	info, err := fs.Stat(fsys, mdlaDevfreqDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	}

	return info.IsDir(), nil
}

func readSoCID(fsys fs.FS) (string, error) {
	data, err := fs.ReadFile(fsys, socIDPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", nil
		}
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

func addSoCInfo(device *Device) {
	info, ok := socIDInfo[socIDProductID(device.SocID)]
	if !ok {
		return
	}

	device.SocInfo = info
}

func socIDProductID(socID string) string {
	socID = strings.ToLower(strings.TrimSpace(socID))
	if socID == "" {
		return ""
	}

	if parts := strings.Split(socID, ":"); len(parts) > 0 {
		socID = parts[len(parts)-1]
	}
	socID = strings.TrimPrefix(socID, "0x")
	socID = strings.TrimPrefix(socID, "mt")

	return socID
}
