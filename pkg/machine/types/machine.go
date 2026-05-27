package types

type MachineInfo struct {
	Cpus    []CpuInfo           `json:"cpus,omitempty" yaml:"cpus,omitempty"`
	Memory  MemoryInfo          `json:"memory,omitempty" yaml:"memory,omitempty"`
	Disk    map[string]DirStats `json:"disk,omitempty" yaml:"disk,omitempty"`
	Devices []DeviceInfo        `json:"devices,omitempty" yaml:"devices,omitempty"`
}
