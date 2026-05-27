package amd

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/canonical/lscompute/pkg/machine/host"
)

func gpuProperties(h host.Host, slot string) (map[string]string, error) {
	properties := make(map[string]string)

	vRamVal, err := vRam(h, slot)
	if err != nil {
		return nil, fmt.Errorf("looking up vram: %w", err)
	}
	if vRamVal != nil {
		properties["vram"] = strconv.FormatUint(*vRamVal, 10)
	}
	gfxArch, err := gfxArchitecture(h, slot)
	if err != nil {
		return nil, fmt.Errorf("looking up gfx architecture: %w", err)
	}
	if len(gfxArch) > 0 {
		properties["microarchitecture"] = gfxArch
	}

	return properties, nil
}

func vRam(h host.Host, slot string) (*uint64, error) {
	/*
		AMD vram is listed under /sys/bus/pci/devices/${pci_slot}/mem_info_vram_total
	*/
	path := filepath.Join("sys/bus/pci/devices", slot, "mem_info_vram_total")
	data, err := fs.ReadFile(h.FS(), path)
	if err != nil {
		return nil, err
	}
	dataStr := strings.TrimSpace(string(data))
	vram, err := strconv.ParseUint(dataStr, 10, 64)
	if err != nil {
		return nil, err
	}
	return &vram, nil
}

func gfxArchitecture(h host.Host, slot string) (string, error) {
	nodesDir := "sys/class/kfd/kfd/topology/nodes"
	files, err := fs.ReadDir(h.FS(), nodesDir)
	if err != nil {
		return "", err
	}

	for _, file := range files {
		if file.IsDir() {
			propertiesPath := filepath.Join(nodesDir, file.Name(), "properties")
			data, err := fs.ReadFile(h.FS(), propertiesPath)
			if err != nil {
				continue // skip this node if we can't read its properties
			}

			nodeMatchesDevice := false
			var nodeGfxTargetVersion string
			for _, line := range strings.Split(string(data), "\n") {
				if strings.HasPrefix(line, "drm_render_minor") {
					pciSlot, err := getAmdGpuPciSlot(h, line)
					if err != nil {
						break
					}
					if pciSlot != slot {
						break
					}
					nodeMatchesDevice = true
				} else if strings.HasPrefix(line, "gfx_target_version") {
					nodeGfxTargetVersion, err = parseGfxTargetVersion(line)
					if err != nil {
						break
					}
				}
			}

			if nodeMatchesDevice && len(nodeGfxTargetVersion) > 0 {
				return nodeGfxTargetVersion, nil
			}

		}
	}
	return "", fmt.Errorf("gfx_target_version not found for device with pci slot %s", slot)
}

func getAmdGpuPciSlot(h host.Host, drmRenderMinor string) (string, error) {
	parts := strings.Split(drmRenderMinor, " ")
	if len(parts) == 2 {
		renderMinor := parts[1]
		// EvalSymlinks uses io/fs convention (no leading slash)
		symlinkPath := filepath.Join("sys/class/drm/renderD"+renderMinor, "device")
		resolvedPath, err := h.EvalSymlinks(symlinkPath)
		if err != nil {
			return "", err
		}
		// The resolved path ends in the PCI slot name (e.g. "sys/bus/pci/devices/0000:03:00.0")
		pciSlot := filepath.Base(resolvedPath)
		return pciSlot, nil
	} else {
		return "", fmt.Errorf("unexpected format for drm_render_minor: %s", drmRenderMinor)
	}
}

func parseGfxTargetVersion(gfxTargetVersionLine string) (string, error) {
	parts := strings.Split(gfxTargetVersionLine, " ")
	if len(parts) == 2 {
		if parts[1] == "0" {
			return "", fmt.Errorf("gfx_target_version is invalid for this device")
		}
		gfxTargetVersion := parts[1]
		deviceLower := strings.ToLower(gfxTargetVersion)
		if len(deviceLower) < 6 {
			return "", fmt.Errorf("gfx_target_version has an unexpected format: %s", gfxTargetVersion)
		}

		majorInt, err := strconv.Atoi(deviceLower[0:2])
		if err != nil {
			return "", fmt.Errorf("parsing major version from gfx_target_version: %w", err)
		}
		major := strconv.Itoa(majorInt)

		minorInt, err := strconv.Atoi(deviceLower[2:4])
		if err != nil {
			return "", fmt.Errorf("parsing minor version from gfx_target_version: %w", err)
		}
		minor := strconv.Itoa(minorInt)

		revisionInt, err := strconv.Atoi(deviceLower[4:6])
		if err != nil {
			return "", fmt.Errorf("parsing revision from gfx_target_version: %w", err)
		}
		revision := strconv.Itoa(revisionInt)

		arch := "gfx" + major + minor + revision
		return arch, nil
	}
	return "", fmt.Errorf("unexpected format for gfx_target_version: %s", gfxTargetVersionLine)
}
