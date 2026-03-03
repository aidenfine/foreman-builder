package main

import (
	"os"
	"path/filepath"
	"runtime"

	foremanbuilder "github.com/aidenfine/foreman-builder/foreman-builder"
	"github.com/aidenfine/foreman-builder/foreman-builder/cmd"
)

func main() {

	if !isSystemSupported() {
		os.Exit(1)
	}

	if !foremanbuilder.IsOrbStackRunning() {
		foremanbuilder.Logger.Fatal("Orbstack must be running")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		foremanbuilder.Logger.Fatalf("Failed to get home directory: %v", err)
	}

	dotFolderPath := filepath.Join(home, ".foreman-builder")
	containersPath := filepath.Join(dotFolderPath, "containers")

	dotfolderExists, err := foremanbuilder.DoesFileOrDirectoryExist(dotFolderPath)
	if err != nil {
		foremanbuilder.Logger.Errorf("Error checking dotfolder: %v", err)
	}

	containersExists, err := foremanbuilder.DoesFileOrDirectoryExist(containersPath)
	if err != nil {
		foremanbuilder.Logger.Errorf("Error checking containers file: %v", err)
	}

	if !dotfolderExists {
		foremanbuilder.Logger.Info("Creating dotfolder...")
		err := os.MkdirAll(dotFolderPath, 0755)
		if err != nil {
			foremanbuilder.Logger.Fatalf("Failed to create dotfolder: %v", err)
		}
	}

	if !containersExists {
		foremanbuilder.Logger.Info("Creating containers file...")
		err := os.WriteFile(containersPath, []byte(""), 0644)
		if err != nil {
			foremanbuilder.Logger.Fatalf("Failed to create container file: %v", err)
		}
	}
	cmd.SyncContainers()

	cmd.Execute()
}

func isSystemSupported() bool {
	// arm64 = silicon chip
	// amd64 = non silicon chip
	userCPU := runtime.GOARCH

	if userCPU != "arm64" {
		foremanbuilder.Logger.Fatal("foreman-builder does not currently support non apple-silicon devices!")
		return false
	}
	return true
}
