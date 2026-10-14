package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	foremanbuilder "github.com/aidenfine/foreman-builder/foreman-builder"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new container environment",
	Run: func(cmd *cobra.Command, args []string) {
		foremanUser.runCreate()
	},
}

var containerType = "orb"

func (u User) runCreate() {
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

	containerNameExists, err := foremanbuilder.GetLineInFile(u.containersPath, containerName, "")
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

	fmt.Println("\nWhat would you like to create?")
	fmt.Println("1) Foreman/Katello server (default)")
	fmt.Println("2) Capsule (Smart Proxy)")
	fmt.Print("Choice [1]: ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)
	if choice == "" {
		choice = "1"
	}

	if choice == "2" {
		fmt.Print("Parent Foreman container name: ")
		parentName, _ := reader.ReadString('\n')
		parentName = strings.TrimSpace(parentName)
		if parentName == "" {
			fmt.Println("Parent container name is required")
			os.Exit(1)
		}

		parentLine, err := foremanbuilder.GetLineInFile(u.containersPath, parentName, "")
		if err != nil || parentLine == "" {
			fmt.Printf("Parent container '%s' not found. Create a Foreman server first.\n", parentName)
			os.Exit(1)
		}

		if strings.Contains(parentLine, "::capsule::") {
			fmt.Printf("'%s' is a capsule, not a Foreman server. Choose a Foreman server as the parent.\n", parentName)
			os.Exit(1)
		}

		parentInfo, err := foremanbuilder.ContainerInfo(parentName)
		if err != nil {
			fmt.Printf("Could not get info for parent container '%s': %v\n", parentName, err)
			os.Exit(1)
		}
		if parentInfo.Record.State != "running" {
			fmt.Printf("Parent container '%s' is not running (state: %s). Start it first.\n", parentName, parentInfo.Record.State)
			os.Exit(1)
		}

		err = foremanbuilder.AppendToFile(u.containersPath, fmt.Sprintf("%s::%s::capsule::%s", containerName, containerType, parentName))
		if err != nil {
			foremanbuilder.Logger.Error("Failed to write container to container file")
		}

		foremanbuilder.Logger.Info("Starting capsule environment creation")

		opts := foremanbuilder.OrbOptions{
			Username:      username,
			ContainerName: containerName,
			ParentName:    parentName,
			CapsuleFQDN:   containerName + ".orb.local",
			ParentFQDN:    parentName + ".orb.local",
		}
		err = u.createCapsuleContainer(opts)
		if err != nil {
			fmt.Printf("An error has occurred during capsule creation: %v\n", err)
			os.Exit(1)
		}
		return
	}

	err = foremanbuilder.AppendToFile(u.containersPath, fmt.Sprintf("%s::%s", containerName, containerType))
	if err != nil {
		foremanbuilder.Logger.Error("Failed to write container to container file")
	}

	foremanbuilder.Logger.Info("Starting environment creation")

	if containerType == "orb" {
		orbOpts := foremanbuilder.OrbOptions{
			Username:      username,
			ContainerName: containerName,
		}
		err := foremanUser.createOrbstackContainer(orbOpts)
		if err != nil {
			fmt.Println("An error has occured during container creation")
			os.Exit(1)
		}
	}

}

func (u User) createOrbstackContainer(opts foremanbuilder.OrbOptions) error {
	config, err := foremanbuilder.GetYmlValues("./config.yml")
	if err != nil {
		foremanbuilder.Logger.Info("No config file found, skipping")
	}

	data := foremanbuilder.OrbstackConfigData{
		Username: opts.Username,
		Packages: config.Packages,
	}

	confsDir := filepath.Join(u.dotFilePath, "confs")
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

func (u User) createCapsuleContainer(opts foremanbuilder.OrbOptions) error {
	config, err := foremanbuilder.GetYmlValues("./config.yml")
	if err != nil {
		foremanbuilder.Logger.Info("No config file found, skipping")
	}

	data := foremanbuilder.CapsuleConfigData{
		Username:      opts.Username,
		Packages:      config.Packages,
		CapsuleFQDN:   opts.CapsuleFQDN,
		ParentFQDN:    opts.ParentFQDN,
		ContainerName: opts.ContainerName,
	}

	confsDir := filepath.Join(u.dotFilePath, "confs")
	if err := os.MkdirAll(confsDir, 0755); err != nil {
		foremanbuilder.Logger.Errorf("Failed to create confs directory: %v", err)
		return err
	}

	pathName := filepath.Join(confsDir, fmt.Sprintf("orbstack-capsule-%s.yml", data.Username))
	foremanbuilder.Logger.Info("using", pathName)
	err = foremanbuilder.GenerateCapsuleConfig(data, pathName)
	if err != nil {
		foremanbuilder.Logger.Errorf("failed to generate capsule config, err: %v\n", err)
		return err
	}

	// Create the OrbStack container
	orbArgs := []string{"create", "-a", "amd64", "-c", pathName, "rocky:9", opts.ContainerName}
	foremanbuilder.Logger.Info("running: orb", strings.Join(orbArgs, " "))
	cmd := exec.Command("orb", orbArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		foremanbuilder.Logger.Errorf("Error creating container: %s", err)
		return err
	}
	fmt.Println("Capsule container created. Waiting for setup to complete...")

	// Wait for SSH and cloud-init
	fmt.Println("Waiting for SSH...")
	if err := foremanbuilder.WaitForSSH(opts.ContainerName, 5*time.Minute); err != nil {
		return fmt.Errorf("SSH failed: %w", err)
	}

	fmt.Println("Waiting for cloud-init to complete (this may take 15-30 minutes)...")
	if err := foremanbuilder.WaitForCloudInit(opts.ContainerName, 45*time.Minute); err != nil {
		return fmt.Errorf("cloud-init failed: %w", err)
	}

	// Orchestrate certificate generation and installer
	if err := u.orchestrateCapsuleCerts(opts); err != nil {
		return fmt.Errorf("certificate orchestration failed: %w", err)
	}

	fmt.Println("Capsule setup complete!")
	fmt.Printf("  SSH: ssh %s@%s@orb\n", opts.Username, opts.ContainerName)
	fmt.Printf("  Parent: %s\n", opts.ParentName)
	return nil
}

func (u User) orchestrateCapsuleCerts(opts foremanbuilder.OrbOptions) error {
	certsFile := fmt.Sprintf("/root/%s-certs.tar", opts.ContainerName)

	// Generate certs on the parent Foreman server
	fmt.Println("Generating certificates on parent server...")
	certGenCmd := fmt.Sprintf(
		"sudo foreman-proxy-certs-generate --foreman-proxy-fqdn %s --certs-tar %s",
		opts.CapsuleFQDN, certsFile,
	)
	certOutput, err := foremanbuilder.SSHRun(opts.ParentName, "root", certGenCmd)
	if err != nil {
		return fmt.Errorf("cert generation failed on '%s': %w\noutput: %s", opts.ParentName, err, certOutput)
	}
	foremanbuilder.Logger.Debugf("cert gen output: %s", certOutput)

	// Transfer certs: parent -> macOS host -> capsule
	fmt.Println("Transferring certificates...")
	tmpDir := filepath.Join(u.dotFilePath, "tmp")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return fmt.Errorf("failed to create tmp directory: %w", err)
	}
	localCertsPath := filepath.Join(tmpDir, fmt.Sprintf("%s-certs.tar", opts.ContainerName))
	defer os.Remove(localCertsPath)

	if err := foremanbuilder.SCPFrom(opts.ParentName, "root", certsFile, localCertsPath); err != nil {
		return fmt.Errorf("failed to copy certs from parent: %w", err)
	}

	if err := foremanbuilder.SCPTo(opts.ContainerName, "root", localCertsPath, certsFile); err != nil {
		return fmt.Errorf("failed to copy certs to capsule: %w", err)
	}

	// Add /etc/hosts entries so both VMs can resolve each other
	fmt.Println("Configuring cross-resolution between capsule and parent...")
	parentIP, err := foremanbuilder.SSHRun(opts.ParentName, "root", "hostname -I | awk '{print $1}'")
	if err != nil {
		return fmt.Errorf("failed to get parent IP: %w", err)
	}
	parentFQDN, err := foremanbuilder.SSHRun(opts.ParentName, "root", "hostname -f")
	if err != nil {
		return fmt.Errorf("failed to get parent FQDN: %w", err)
	}
	capsuleIP, err := foremanbuilder.SSHRun(opts.ContainerName, "root", "hostname -I | awk '{print $1}'")
	if err != nil {
		return fmt.Errorf("failed to get capsule IP: %w", err)
	}

	hostsEntry := fmt.Sprintf("echo '%s %s' >> /etc/hosts", parentIP, parentFQDN)
	if _, err := foremanbuilder.SSHRun(opts.ContainerName, "root", hostsEntry); err != nil {
		return fmt.Errorf("failed to add parent to capsule /etc/hosts: %w", err)
	}
	hostsEntry = fmt.Sprintf("echo '%s %s' >> /etc/hosts", capsuleIP, opts.CapsuleFQDN)
	if _, err := foremanbuilder.SSHRun(opts.ParentName, "root", hostsEntry); err != nil {
		return fmt.Errorf("failed to add capsule to parent /etc/hosts: %w", err)
	}

	// Parse the cert-gen output for the foreman-installer command
	installerCmd, err := foremanbuilder.ParseCertGenOutput(certOutput)
	if err != nil {
		return fmt.Errorf("failed to parse cert generation output: %w", err)
	}
	foremanbuilder.Logger.Debugf("installer command: %s", installerCmd)

	// Run foreman-installer on the capsule
	fmt.Println("Running foreman-installer on capsule (this may take 10-15 minutes)...")
	installOutput, err := foremanbuilder.SSHRun(opts.ContainerName, "root", installerCmd)
	if err != nil {
		return fmt.Errorf("foreman-installer failed on capsule: %w\noutput: %s", err, installOutput)
	}
	foremanbuilder.Logger.Debugf("installer output: %s", installOutput)

	return nil
}
