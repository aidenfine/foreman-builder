package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	foremanbuilder "github.com/aidenfine/foreman-builder/foreman-builder"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "foreman-builder",
	Short: "Container environment builder for foreman",
	Long:  "A CLI tool to create and manage foreman containers.",
}

type User struct {
	homeDir string
	dotFilePath string
	containersPath string
}

var foremanUser User

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	foremanUser = User{
		homeDir: foremanbuilder.GetHomeDir(),
		dotFilePath: "",
		containersPath: "",
	}
	foremanUser.dotFilePath = filepath.Join(foremanUser.homeDir, ".foreman-builder")
	if foremanUser.dotFilePath != "" {
		foremanUser.containersPath = filepath.Join(foremanUser.dotFilePath, "containers")
	}

	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(deleteCmd)

	// rootCmd.AddCommand(syncCmd)
}
