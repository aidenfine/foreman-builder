package cmd

import (
	"fmt"
	"strings"

	foremanbuilder "github.com/aidenfine/foreman-builder/foreman-builder"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all containers created with foreman-builder.",
	Long:  "List all containers created with foreman-builder.",
	Run: func(cmd *cobra.Command, args []string) {
		foremanUser.runList()
	},
}

func (u User) runList() {
	fmt.Println("Foreman containers:")
	containersPath := u.containersPath
	lines, err := foremanbuilder.GetAllLines(containersPath, "")
	if err != nil {
		fmt.Printf("Error getting all containers %v\n", err)
		return
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "::")
		name := parts[0]

		if len(parts) >= 4 && parts[2] == "capsule" {
			fmt.Printf("- %s (capsule -> %s)\n", name, parts[3])
		} else {
			fmt.Printf("- %s (server)\n", name)
		}
	}
}
