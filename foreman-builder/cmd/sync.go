package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"

	foremanbuilder "github.com/aidenfine/foreman-builder/foreman-builder"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync foreman-builder containers with orbstack.",
	Run: func(cmd *cobra.Command, args []string) {
		runSync()
	},
}

func runSync() {
	SyncContainers()

}

func SyncContainers() {
	orbContainers, err := foremanbuilder.GetOrbstackContainers()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		log.Panicf("Failed to get home directory: %v\n", err)
	}
	dotFolderPath := filepath.Join(home, ".foreman-builder")
	containersPath := filepath.Join(dotFolderPath, "containers")
	foremanContainers, err := foremanbuilder.GetAllLines(containersPath)
	if err != nil {
		fmt.Printf("Error getting all containers %v\n", err)
	}

	for i := 0; i < len(foremanContainers)-1; i++ {
		if !slices.Contains(orbContainers, foremanContainers[i]) {
			// mark for deletion
			err = foremanbuilder.DeleteLineInFile(containersPath, foremanContainers[i])
			if err != nil {
				fmt.Printf("Error deleting container from dotfile %v \n", err)
			}
		}
	}

}
