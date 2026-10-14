//go:build e2e

package foremanbuilder_test

import (
	"os/user"
	"testing"
)

func TestCapsuleLifecycle(t *testing.T) {
	currentUser, err := user.Current()
	if err != nil {
		t.Fatalf("failed to get current user: %v", err)
	}
	username := currentUser.Username

	hostContainer := createHostContainer(t)
	capsuleContainer := createCapsuleContainer(t, hostContainer)

	t.Run("ssh_as_user", func(t *testing.T) {
		out, err := sshRun(t, capsuleContainer, username, "echo hello")
		if err != nil {
			t.Fatalf("SSH as %s failed: %v, output: %s", username, err, out)
		}
		if out != "hello" {
			t.Fatalf("expected 'hello', got %q", out)
		}
	})
}
