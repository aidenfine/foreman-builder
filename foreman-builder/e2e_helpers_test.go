//go:build e2e

package foremanbuilder_test

import (
	"fmt"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"testing"
	"time"

	foremanbuilder "github.com/aidenfine/foreman-builder/foreman-builder"
)

func sshRun(t *testing.T, container, sshUser, cmd string) (string, error) {
	t.Helper()
	target := fmt.Sprintf("%s@%s@orb", sshUser, container)
	out, err := exec.Command("ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "ConnectTimeout=5",
		"-o", "LogLevel=ERROR",
		target, cmd,
	).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func waitForSSH(t *testing.T, container string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Minute)
	for time.Now().Before(deadline) {
		_, err := sshRun(t, container, "root", "echo ok")
		if err == nil {
			t.Log("SSH is available")
			return
		}
		t.Log("waiting for SSH... retrying in 10s")
		time.Sleep(10 * time.Second)
	}
	t.Fatal("SSH did not become available within 5 minutes")
}

func waitForCloudInit(t *testing.T, container string) {
	t.Helper()
	deadline := time.Now().Add(45 * time.Minute)
	for time.Now().Before(deadline) {
		out, err := sshRun(t, container, "root", "cloud-init status 2>/dev/null || echo unknown")
		if err == nil {
			if strings.Contains(out, "done") {
				t.Log("cloud-init completed")
				return
			}
			if strings.Contains(out, "error") || strings.Contains(out, "recoverable error") {
				t.Logf("cloud-init finished with status: %s", out)
				return
			}
		}
		t.Log("cloud-init still running, checking again in 30s...")
		time.Sleep(30 * time.Second)
	}
	t.Fatal("cloud-init did not complete within 45 minutes")
}

func createHostContainer(t *testing.T) string {
	t.Helper()
	if !foremanbuilder.IsOrbStackRunning() {
		t.Skip("OrbStack is not running, skipping integration test")
	}

	currentUser, err := user.Current()
	if err != nil {
		t.Fatalf("failed to get current user: %v", err)
	}
	username := currentUser.Username
	containerName := fmt.Sprintf("fb-test-%d", time.Now().Unix())

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "cloud-init.yml")

	data := foremanbuilder.OrbstackConfigData{
		Username: username,
		Packages: []string{"tmux", "tree"},
	}
	if err := foremanbuilder.GenerateContainerConfig(data, configPath); err != nil {
		t.Fatalf("failed to generate config: %v", err)
	}

	orbArgs := []string{"create", "-a", "amd64", "-c", configPath, "rocky:9", containerName}
	t.Logf("creating container: orb %s", strings.Join(orbArgs, " "))
	createCmd := exec.Command("orb", orbArgs...)
	if out, err := createCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to create container: %v\noutput: %s", err, out)
	}
	t.Logf("container %s created", containerName)

	t.Cleanup(func() {
		t.Logf("cleaning up container %s", containerName)
		exec.Command("orbctl", "stop", containerName).Run()
		exec.Command("orbctl", "delete", containerName, "-f").Run()
	})

	waitForSSH(t, containerName)
	waitForCloudInit(t, containerName)
	return containerName
}

func createCapsuleContainer(t *testing.T, mainServerName string) string {
	t.Helper()
	if !foremanbuilder.IsOrbStackRunning() {
		t.Skip("OrbStack is not running, skipping integration test")
	}

	currentUser, err := user.Current()
	if err != nil {
		t.Fatalf("failed to get current user: %v", err)
	}
	username := currentUser.Username
	containerName := fmt.Sprintf("fb-capsule-test-%d", time.Now().Unix())

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "cloud-init.yml")

	data := foremanbuilder.OrbstackConfigData{
		Username: username,
		Packages: []string{"tmux", "tree"},
	}
	if err := foremanbuilder.GenerateContainerConfig(data, configPath); err != nil {
		t.Fatalf("failed to generate config: %v", err)
	}

	orbArgs := []string{"create", "-a", "amd64", "-c", configPath, "rocky:9", containerName}
	t.Logf("creating container: orb %s", strings.Join(orbArgs, " "))
	createCmd := exec.Command("orb", orbArgs...)
	if out, err := createCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to create container: %v\noutput: %s", err, out)
	}
	t.Logf("container %s created", containerName)

	t.Cleanup(func() {
		t.Logf("cleaning up container %s", containerName)
		exec.Command("orbctl", "stop", containerName).Run()
		exec.Command("orbctl", "delete", containerName, "-f").Run()
	})

	waitForSSH(t, containerName)
	waitForCloudInit(t, containerName)
	return containerName
}
