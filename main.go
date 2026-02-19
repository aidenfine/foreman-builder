package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"text/template"

	tea "github.com/charmbracelet/bubbletea"
)

type opts struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
}

type OrbstackConfigData struct {
	Username string
}

func initialOpts() opts {
	return opts{
		choices:  []string{"bash", "zsh"},
		selected: make(map[int]struct{}),
	}
}

func main() {

	// arm64 = silicon chip
	// amd64 = non silicon chip
	userCPU := runtime.GOARCH

	if userCPU != "arm64" {
		log.Fatal("foreman-builder does not currently support non apple-silicon devices!")
	}

	if !isOrbStackRunning() {
		log.Fatalf("Orbstack must be running")

	}

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

	data := OrbstackConfigData{
		Username: username,
	}

	// check for errors by doing ssh <container-name>@orb cat /var/log/cloud-init-output.log

	pathName := fmt.Sprintf("./confs/orbstack-foreman-%s.yml", data.Username)
	fmt.Println("using", pathName)
	err = generateYAML(data, pathName)
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
}

func isOrbStackRunning() bool {
	cmd := exec.Command("orbctl", "status")
	return cmd.Run() == nil
}

func generateYAML(data OrbstackConfigData, pathName string) error {
	// tmpl, err := template.ParseFiles("./confs/orbstack-foreman.yml.tmpl")
	tmpl, err := template.ParseFiles("./confs/orbstack-foreman.yml.tmpl")
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}

	return os.WriteFile(pathName, buf.Bytes(), 0644)
}
