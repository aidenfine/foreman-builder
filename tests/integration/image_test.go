//go:build integration
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

func TestImageLifecycle(t *testing.T) {
	if !foremanbuilder.IsOrbStackRunning() {
		t.Skip("OrbStack is not running, skipping integration test")
	}

	currentUser, err := user.Current()
	if err != nil {
		t.Fatalf("failed to get current user: %v", err)
	}
	username := currentUser.Username
	containerName := fmt.Sprintf("fb-test-%d", time.Now().Unix())

	// Generate cloud-init config in a temp directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "cloud-init.yml")

	data := foremanbuilder.OrbstackConfigData{
		Username: username,
		Packages: []string{"tmux", "tree"},
	}
	if err := foremanbuilder.GenerateContainerConfig(data, configPath); err != nil {
		t.Fatalf("failed to generate config: %v", err)
	}

	// Create container
	orbArgs := []string{"create", "-a", "amd64", "-c", configPath, "rocky:9", containerName}
	t.Logf("creating container: orb %s", strings.Join(orbArgs, " "))
	createCmd := exec.Command("orb", orbArgs...)
	if out, err := createCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to create container: %v\noutput: %s", err, out)
	}
	t.Logf("container %s created", containerName)

	// Always clean up the container
	t.Cleanup(func() {
		t.Logf("cleaning up container %s", containerName)
		exec.Command("orbctl", "stop", containerName).Run()
		exec.Command("orbctl", "delete", containerName, "-f").Run()
	})

	waitForSSH(t, containerName)
	waitForCloudInit(t, containerName)

	t.Run("ssh_as_user", func(t *testing.T) {
		out, err := sshRun(t, containerName, username, "echo hello")
		if err != nil {
			t.Fatalf("SSH as %s failed: %v, output: %s", username, err, out)
		}
		if out != "hello" {
			t.Fatalf("expected 'hello', got %q", out)
		}
	})

	t.Run("hostname", func(t *testing.T) {
		out, err := sshRun(t, containerName, "root", "hostname")
		if err != nil {
			t.Fatalf("failed: %v, output: %s", err, out)
		}
		if !strings.Contains(out, "foreman-dev") {
			t.Errorf("expected hostname containing 'foreman-dev', got %q", out)
		}
	})

	t.Run("base_packages", func(t *testing.T) {
		for _, pkg := range []string{"git", "curl", "wget", "htop", "gcc"} {
			t.Run(pkg, func(t *testing.T) {
				out, err := sshRun(t, containerName, "root", fmt.Sprintf("rpm -q %s", pkg))
				if err != nil {
					t.Errorf("not installed: %v, output: %s", err, out)
				}
			})
		}
	})
	t.Run("foreman_repos_exist", func(t *testing.T) {
		repoList := []string{"foreman", "katello", "foreman-certs", "foreman_remote_execution"}

		for _, repo := range repoList{
			out, err := sshRun(t, containerName, "root", fmt.Sprintf("test -d %s && echo exists",  repo))
			if err != nil {
				t.Fatalf("folder not found: %v, output: %s",err, out)
			}
			if out != "exists"{
				t.Errorf("expected %s to exist, got %q", repo,out);
			}
		}
	});

	t.Run("config_packages", func(t *testing.T) {
		for _, pkg := range []string{"tmux", "tree"} {
			t.Run(pkg, func(t *testing.T) {
				out, err := sshRun(t, containerName, "root", fmt.Sprintf("rpm -q %s", pkg))
				if err != nil {
					t.Errorf("not installed: %v, output: %s", err, out)
				}
			})
		}
	})

	t.Run("user_exists_in_wheel", func(t *testing.T) {
		out, err := sshRun(t, containerName, "root", fmt.Sprintf("id %s", username))
		if err != nil {
			t.Fatalf("user %s not found: %v, output: %s", username, err, out)
		}
		if !strings.Contains(out, "wheel") {
			t.Errorf("user not in wheel group: %s", out)
		}
	})

	t.Run("passwordless_sudo", func(t *testing.T) {
		out, err := sshRun(t, containerName, username, "sudo -n whoami")
		if err != nil {
			t.Fatalf("passwordless sudo failed: %v, output: %s", err, out)
		}
		if !strings.Contains(out, "root") {
			t.Errorf("expected 'root', got %q", out)
		}
	})

	t.Run("node_installed", func(t *testing.T) {
		out, err := sshRun(t, containerName, "root", "node --version")
		if err != nil {
			t.Fatalf("node not found: %v, output: %s", err, out)
		}
		if !strings.HasPrefix(out, "v") {
			t.Errorf("unexpected node version: %q", out)
		}
		t.Logf("node version: %s", out)
	})

	t.Run("ruby_installed", func(t *testing.T) {
		out, err := sshRun(t, containerName, username, "bash -lc 'ruby --version'")
		if err != nil {
			t.Fatalf("ruby not found: %v, output: %s", err, out)
		}
		// if !strings.Contains(out, "3.1.6") {
		// 	t.Errorf("expected ruby 3.1.6, got %q", out)
		// }
		t.Logf("ruby version: %s", out)
	})

	t.Run("foreman_installer", func(t *testing.T) {
		out, err := sshRun(t, containerName, "root", "rpm -q foreman-installer")
		if err != nil {
			t.Fatalf("foreman-installer not found: %v, output: %s", err, out)
		}
		t.Logf("foreman-installer: %s", out)
	})
}
