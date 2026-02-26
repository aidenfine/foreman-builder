package foremanbuilder

import (
	"encoding/json"
	"os/exec"
)

type OrbInstance struct {
	Name    string `json:"name"`
	State   string `json:"state"`
	Distro  string `json:"distro"`
	Version string `json:"version"`
	Arch    string `json:"arch"`
	Size    string `json:"size"`
	IP      string `json:"ip"`
}

func IsOrbStackRunning() bool {
	cmd := exec.Command("orbctl", "status")
	return cmd.Run() == nil
}

func GetOrbstackContainers() ([]string, error) {
	cmd := exec.Command("orbctl", "list", "--format", "json")

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var instances []OrbInstance
	if err := json.Unmarshal(output, &instances); err != nil {
		return nil, err
	}

	var names []string
	for _, inst := range instances {
		names = append(names, inst.Name)
	}

	return names, nil
}
