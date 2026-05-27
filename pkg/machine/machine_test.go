package machine

import (
	"encoding/json"
	"testing"

	"github.com/canonical/lscompute/pkg/machine/types"
)

var devices = []string{
	"raspberry-pi-5",
	"raspberry-pi-5+hailo-8",
	"xps13-7390",
}

//func TestGetFromFiles(t *testing.T) {
//	for _, device := range devices {
//		t.Run(device, func(t *testing.T) {
//			hwInfo, err := GetFromRawData(device, true, "../../test_data")
//			if err != nil {
//				t.Error(err)
//			}
//
//			var hardwareInfo types.Machine
//			devicePath := "../../test_data/machines/" + device + "/"
//			hardwareInfoData, err := os.ReadFile(devicePath + "hardware-info.json")
//			if err != nil {
//				t.Fatal(err)
//			}
//			err = json.Unmarshal(hardwareInfoData, &hardwareInfo)
//			if err != nil {
//				t.Fatal(err)
//			}
//
//			// Ignore friendly names during deep equal, as it depends on the version of the pci-id database
//			for i := range hwInfo.PciDevices {
//				hwInfo.PciDevices[i].VendorName = nil
//				hwInfo.PciDevices[i].DeviceName = nil
//				hwInfo.PciDevices[i].SubvendorName = nil
//				hwInfo.PciDevices[i].SubdeviceName = nil
//			}
//			for i := range hardwareInfo.PciDevices {
//				hardwareInfo.PciDevices[i].VendorName = nil
//				hardwareInfo.PciDevices[i].DeviceName = nil
//				hardwareInfo.PciDevices[i].SubvendorName = nil
//				hardwareInfo.PciDevices[i].SubdeviceName = nil
//			}
//
//			if diff := deep.Equal(*hwInfo, hardwareInfo); diff != nil {
//				t.Error(diff)
//			}
//		})
//	}
//}

func TestGetCurrentHostMarshalUnmarshal(t *testing.T) {
	machineInfo, _, err := Get(false)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	jsonData, err := json.Marshal(machineInfo)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	var unmarshalled types.MachineInfo
	err = json.Unmarshal(jsonData, &unmarshalled)
	if err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	t.Logf("Successfully marshalled and unmarshalled machine info: %s", string(jsonData))
}
