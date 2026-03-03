package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"

	foremanbuilder "github.com/aidenfine/foreman-builder/foreman-builder"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete container via name.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 || len(args) > 1 {
			log.Fatalf("Arg Error: Got %d args, when 1 was expected \n", len(args))
			os.Exit(1)
		}
		containerName := args[0]
		runDelete(containerName)

	},
}

func runDelete(containerName string) {

	// check if item actually exists first AND foreman-builder created it
	home, err := os.UserHomeDir()
	if err != nil {
		log.Panicf("Failed to get home directory: %v\n", err)
	}
	dotFolderPath := filepath.Join(home, ".foreman-builder")
	containersPath := filepath.Join(dotFolderPath, "containers")
	containers, err := foremanbuilder.GetAllLines(containersPath)

	if !slices.Contains(containers, containerName) {
		log.Fatalf("%s was not created with foreman-builder, foreman-builder will not delete it", containerName)
		os.Exit(1)

	}
	containerInfo, err := foremanbuilder.ContainerInfo(containerName)
	if err != nil {
		log.Fatalf("Failed to get container info %v \n", err)
		os.Exit(1)
	}
	if containerInfo.State == "running" {
		var input string
		for input != "y" && input != "n" {
			fmt.Println("Container is currently running! \n Do you want to stop the container and delete? \n [y/n]")
			fmt.Scanln(&input)
		}
		if input == "n" {
			os.Exit(1)
		}
		// stop container
		exec.Command("orbctl", "stop", containerName)

	}
	exec.Command("orbctl", "delete", containerName, "-f")

	// delete container in dotfile
	foremanbuilder.DeleteLineInFile(containersPath, containerName)
}
