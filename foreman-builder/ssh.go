package foremanbuilder

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func SSHRun(container, user, cmd string) (string, error) {
	target := fmt.Sprintf("%s@%s@orb", user, container)
	out, err := exec.Command("ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "ConnectTimeout=5",
		"-o", "LogLevel=ERROR",
		target, cmd,
	).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func WaitForSSH(container string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_, err := SSHRun(container, "root", "echo ok")
		if err == nil {
			Logger.Info("SSH is available")
			return nil
		}
		Logger.Info("waiting for SSH... retrying in 10s")
		time.Sleep(10 * time.Second)
	}
	return fmt.Errorf("SSH did not become available within %v", timeout)
}

func WaitForCloudInit(container string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		out, err := SSHRun(container, "root", "cloud-init status 2>/dev/null || echo unknown")
		if err == nil {
			if strings.Contains(out, "done") {
				Logger.Info("cloud-init completed")
				return nil
			}
			if strings.Contains(out, "error") || strings.Contains(out, "recoverable error") {
				Logger.Infof("cloud-init finished with status: %s", out)
				return nil
			}
		}
		Logger.Info("cloud-init still running, checking again in 30s...")
		time.Sleep(30 * time.Second)
	}
	return fmt.Errorf("cloud-init did not complete within %v", timeout)
}

func SCPFrom(container, user, remotePath, localPath string) error {
	remote := fmt.Sprintf("%s@%s@orb:%s", user, container, remotePath)
	cmd := exec.Command("scp",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=ERROR",
		remote, localPath,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("scp from %s failed: %w\noutput: %s", container, err, string(out))
	}
	return nil
}

func SCPTo(container, user, localPath, remotePath string) error {
	remote := fmt.Sprintf("%s@%s@orb:%s", user, container, remotePath)
	cmd := exec.Command("scp",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=ERROR",
		localPath, remote,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("scp to %s failed: %w\noutput: %s", container, err, string(out))
	}
	return nil
}

func ParseCertGenOutput(output string) (string, error) {
	lines := strings.Split(output, "\n")
	var capturing bool
	var cmdLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !capturing && (strings.HasPrefix(trimmed, "foreman-installer ") || trimmed == "foreman-installer\\" || trimmed == "foreman-installer" ||
			strings.HasPrefix(trimmed, "satellite-installer ") || trimmed == "satellite-installer\\" || trimmed == "satellite-installer") {
			capturing = true
		}
		if capturing {
			cmdLines = append(cmdLines, trimmed)
			if !strings.HasSuffix(trimmed, "\\") {
				break
			}
		}
	}

	if len(cmdLines) == 0 {
		return "", fmt.Errorf("could not find foreman-installer command in cert generation output")
	}

	cmd := strings.Join(cmdLines, " ")
	cmd = strings.ReplaceAll(cmd, "\\", "")
	// collapse multiple spaces
	for strings.Contains(cmd, "  ") {
		cmd = strings.ReplaceAll(cmd, "  ", " ")
	}
	return strings.TrimSpace(cmd), nil
}
