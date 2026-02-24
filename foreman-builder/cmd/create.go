package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	foremanbuilder "github.com/aidenfine/foreman-builder/foreman-builder"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

type opts struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
}

func initialOpts() opts {
	return opts{
		choices:  []string{"bash", "zsh"},
		selected: make(map[int]struct{}),
	}
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new container environment",
	Run: func(cmd *cobra.Command, args []string) {
		runCreate()
	},
}

func runCreate() {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf("Failed to get current user: %s", err)
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

	p := tea.NewProgram(initialOpts())
	finalOpts, err := p.Run()
	if err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
	m := finalOpts.(opts)

	// currently will not do anything
	selectedShell := m.choices[m.cursor]

	fmt.Println("Selected Shell", selectedShell)

	log.Printf("Starting enviornment creation ")

	config, err := foremanbuilder.GetYmlValues("./config.yml")
	if err != nil {
		log.Printf("No config file found skipping...")
	}

	data := foremanbuilder.OrbstackConfigData{
		Username: username,
		Packages: config.Packages,
	}

	// check for errors by doing ssh <container-name>@orb cat /var/log/cloud-init-output.log

	pathName := fmt.Sprintf("./confs/orbstack-foreman-%s.yml", data.Username)
	fmt.Println("using", pathName)
	err = foremanbuilder.GenerateContainerConfig(data, pathName)
	if err != nil {
		log.Fatal(err)
	}

	// orb create -a amd64 -c ~/orbstack-foreman.yml rocky:9 foreman-dev
	// run command to create container
	cmd := exec.Command("orb", "create", "-a", "amd64", "-c", pathName, "rocky:9", containerName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Error creating container: %s", err)
	}
	fmt.Println("Container has been created!")

	home, err := os.UserHomeDir()
	if err != nil {
		log.Panicf("Failed to get home directory: %v\n", err)
	}

	err = foremanbuilder.AppendToFile(filepath.Join(home, ".foreman-builder/containers"), containerName)
	if err != nil {
		log.Println("Failed to write container to container file")
	}

}
