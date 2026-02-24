package foremanbuilder

import "os/exec"

func IsOrbStackRunning() bool {
	cmd := exec.Command("orbctl", "status")
	return cmd.Run() == nil
}
