package cmd

import (
	"fmt"
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

func(u User) runList() {
	fmt.Println("Foreman containers:")
	containersPath := u.containersPath
	containers, err := foremanbuilder.GetAllLines(containersPath, "::")
	if err != nil {
		fmt.Printf("Error getting all containers %v\n", err)
	}

	// i - 1 due to the empty line present at the end
	for i := 0; i < len(containers)-1; i++ {
		fmt.Printf("- %s\n", containers[i])
	}
}
