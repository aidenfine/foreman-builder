package main

import (
	"log"
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
		log.Fatalf("Orbstack must be running")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		log.Panicf("Failed to get home directory: %v\n", err)
	}

	dotFolderPath := filepath.Join(home, ".foreman-builder")
	containersPath := filepath.Join(dotFolderPath, "containers")

	dotfolderExists, err := foremanbuilder.DoesFileOrDirectoryExist(dotFolderPath)
	if err != nil {
		log.Printf("Error checking dotfolder: %v\n", err)
	}

	containersExists, err := foremanbuilder.DoesFileOrDirectoryExist(containersPath)
	if err != nil {
		log.Printf("Error checking containers file: %v\n", err)
	}

	if !dotfolderExists {
		log.Printf("Creating dotfolder...\n")
		err := os.MkdirAll(dotFolderPath, 0755)
		if err != nil {
			log.Panicf("Failed to create dotfolder: %v\n", err)
		}
	}

	if !containersExists {
		log.Printf("Creating containers file...\n")
		err := os.WriteFile(containersPath, []byte(""), 0644)
		if err != nil {
			log.Panicf("Failed to create container file: %v\n", err)
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
		log.Fatal("foreman-builder does not currently support non apple-silicon devices!")
		return false
	}
	return true
}
