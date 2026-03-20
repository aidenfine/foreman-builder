package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	foremanbuilder "github.com/aidenfine/foreman-builder/foreman-builder"
)

// var syncCmd = &cobra.Command{
// 	Use:   "sync",
// 	Short: "Sync foreman-builder containers with orbstack.",
// 	Run: func(cmd *cobra.Command, args []string) {
// 		runSync()
// 	},
// }

// func runSync() {
// 	SyncContainers()

// }

func SyncContainers() {
	orbContainers, err := foremanbuilder.GetOrbStackContainers()
	if err != nil {
		foremanbuilder.Logger.Fatalf("Error: %v\n", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		foremanbuilder.Logger.Fatalf("Failed to get home directory: %v", err)
	}
	dotFolderPath := filepath.Join(home, ".foreman-builder")
	containersPath := filepath.Join(dotFolderPath, "containers")
	foremanContainers, err := foremanbuilder.GetAllLines(containersPath, "::")
	if err != nil {
		fmt.Printf("Error getting all containers %v\n", err)
		foremanbuilder.Logger.Errorf("Error has occurred getting all containers %v\n", err)
	}

	for i := 0; i < len(foremanContainers)-1; i++ {
		if !slices.Contains(orbContainers, foremanContainers[i]) {
			// mark for deletion
			foremanbuilder.Logger.Debugf("During sync... deleting %s\n", foremanContainers[i])
			err = foremanbuilder.DeleteLineInFile(containersPath, foremanContainers[i])
			if err != nil {
				fmt.Printf("Error deleting container from dotfile %v \n", err)
			}
		}
	}

}
