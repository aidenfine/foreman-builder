package foremanbuilder

import (
	"os"
)

func GetHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		Logger.Fatalf("Failed to get home directory: %v", err)
	}
	return home
}
