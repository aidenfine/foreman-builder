//go:build integration

package foremanbuilder_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"os/user"

	foremanbuilder "github.com/aidenfine/foreman-builder/foreman-builder"
)

// TODO: reduce code count many of these tests are pulled from, image_test.go. If we can find a way
// to use cleanly merge these tests together and share same logic that would be nice.


var tarFile = "foreman-base.tar.zst"

func prebuiltBuild(t *testing.T, containerName string){
	// pull image
	tmpDir := t.TempDir()
	imgDir := filepath.Join(tmpDir, tarFile)

	imageURL := fmt.Sprintf("https://checkpoint-distributed-production.up.railway.app/images/%s", tarFile)
	curlCmd := exec.Command("curl", "-o", imgDir, imageURL)
	curlCmd.Stdout = os.Stdout
	curlCmd.Stderr = os.Stderr

	if err := curlCmd.Run(); err != nil {
		t.Fatalf("Failed to download image %v", err)
	}

	orbArgs := []string{"import", "-n", containerName, imgDir}
	cmd := exec.Command("orb", orbArgs...)
	curlCmd.Stdout = os.Stdout
	curlCmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("Error importing container: %v", err)
	}


	startCmd := exec.Command("orb", "start", containerName)
	startCmd.Stdout = os.Stdout
	startCmd.Stderr = os.Stderr
	if err := startCmd.Run(); err != nil {
		t.Fatalf("Error starting container: %v", err)
	}

}


func TestPrebuiltImageLifecycle(t *testing.T) {
	if !foremanbuilder.IsOrbStackRunning() {
		t.Skip("OrbStack is not running, skipping integration test")
	}
	currUser, err := user.Current()
	if err != nil {
		t.Fatalf("failed to get current user: %v", err)
	}
	username := currUser.Username
	containerName := fmt.Sprintf("fb-test-%d", time.Now().Unix())
	// Create container
	prebuiltBuild(t, containerName)
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
			t.Fatalf("SSH as %s failed: %v, output: %s", "root", err, out)
		}
		if out != "hello" {
			t.Fatalf("expected 'hello', got %q", out)
		}
	})

	// hostname is reset by OrbStack to the container name on import,
	// so we skip this check for prebuilt images

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

		for _, repo := range repoList {
			out, err := sshRun(t, containerName, "root", fmt.Sprintf("test -d /home/%s/%s && echo exists", username, repo))
			if err != nil {
				t.Fatalf("folder not found: %v, output: %s", err, out)
			}
			if out != "exists" {
				t.Errorf("expected %s to exist, got %q", repo, out)
			}
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
