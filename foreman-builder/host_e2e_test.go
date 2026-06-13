//go:build e2e

package foremanbuilder_test

import (
	"fmt"
	"os/user"
	"strings"
	"testing"
)

func TestContainerLifecycle(t *testing.T) {
	currentUser, err := user.Current()
	if err != nil {
		t.Fatalf("failed to get current user: %v", err)
	}
	username := currentUser.Username

	containerName := createHostContainer(t)

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
