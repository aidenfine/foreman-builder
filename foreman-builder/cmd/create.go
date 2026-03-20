package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	foremanbuilder "github.com/aidenfine/foreman-builder/foreman-builder"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new container environment",
	Run: func(cmd *cobra.Command, args []string) {
		runCreate()
	},
}

var containerType = "orb"

func runCreate() {
	currentUser, err := user.Current()
	if err != nil {
		foremanbuilder.Logger.Fatalf("Failed to get current user: %s", err)
	}
	username := currentUser.Username
	fmt.Println("Username:", username)

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Container name (default foreman): ")
	containerName, _ := reader.ReadString('\n')
	containerName = strings.TrimSpace(containerName)
	if containerName == "" {
		containerName = "foreman"
	}

	home, err := os.UserHomeDir()
	if err != nil {
		foremanbuilder.Logger.Fatalf("Failed to get home directory: %v", err)
	}

	containerNameExists, err := foremanbuilder.GetLineInFile(filepath.Join(home, ".foreman-builder/containers"), containerName, "")
	if err != nil {
		if err.Error() != "not_found" {
			fmt.Println("An Error has occured searching dotfile")
			os.Exit(1)
		}
	}
	foremanbuilder.Logger.Debug("container name: %s \n", containerName)
	foremanbuilder.Logger.Debug("container name EXISTS: %s \n", containerNameExists)
	if containerNameExists != "" {
		fmt.Println("Container name must be unique")
		os.Exit(1)
	}

	err = foremanbuilder.AppendToFile(filepath.Join(home, ".foreman-builder/containers"), fmt.Sprintf("%s::%s", containerName, containerType))
	if err != nil {
		foremanbuilder.Logger.Error("Failed to write container to container file")
	}

	foremanbuilder.Logger.Info("Starting environment creation")

	if containerType == "orb" {
		orbOpts := foremanbuilder.OrbOptions{
			Username:      username,
			ContainerName: containerName,
		}
		err := createOrbstackContainer(orbOpts)
		if err != nil {
			// better error message to show?
			fmt.Println("An error has occured during container creation")
			os.Exit(1)
		}
	}

}

func createOrbstackContainer(opts foremanbuilder.OrbOptions) error {
	config, err := foremanbuilder.GetYmlValues("./config.yml")
	if err != nil {
		foremanbuilder.Logger.Info("No config file found, skipping")
	}

	data := foremanbuilder.OrbstackConfigData{
		Username: opts.Username,
		Packages: config.Packages,
	}

	// check for errors by doing ssh <container-name>@orb cat /var/log/cloud-init-output.log
	home, err := os.UserHomeDir()
	if err != nil {
		foremanbuilder.Logger.Errorf("Failed to get home directory: %v", err)
		return err
	}
	confsDir := filepath.Join(home, ".foreman-builder", "confs")
	if err := os.MkdirAll(confsDir, 0755); err != nil {
		foremanbuilder.Logger.Errorf("Failed to create confs directory: %v", err)
		return err
	}

	pathName := filepath.Join(confsDir, fmt.Sprintf("orbstack-foreman-%s.yml", data.Username))
	foremanbuilder.Logger.Info("using", pathName)
	err = foremanbuilder.GenerateContainerConfig(data, pathName)
	if err != nil {
		foremanbuilder.Logger.Errorf("failed to generate container config, err : %v\n", err)
		return err
	}

	// run command to create container
	orbArgs := []string{"create", "-a", "amd64", "-c", pathName, "rocky:9", opts.ContainerName}
	foremanbuilder.Logger.Info("running: orb", strings.Join(orbArgs, " "))
	cmd := exec.Command("orb", orbArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		foremanbuilder.Logger.Errorf("Error creating container: %s", err)
		return err
	}
	fmt.Println("Container has been created!")

	return nil

}
