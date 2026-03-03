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

type ContainerInfoStruct struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	State string `json:"state"`
}

var execCommand = exec.Command

func IsOrbStackRunning() bool {
	cmd := execCommand("orbctl", "status")
	return cmd.Run() == nil
}

func GetOrbStackContainers() ([]string, error) {
	cmd := execCommand("orbctl", "list", "--format", "json")

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

func ContainerInfo(containerName string) (ContainerInfoStruct, error) {
	cmd := execCommand("orbctl", "info", containerName, "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		return ContainerInfoStruct{}, err
	}

	var containerInfo ContainerInfoStruct

	if err := json.Unmarshal(output, &containerInfo); err != nil {
		return ContainerInfoStruct{}, err
	}
	return containerInfo, nil
}
