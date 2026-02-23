package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "foreman-builder",
	Short: "Container environment builder for OrbStack",
	Long:  "A CLI tool to create and manage OrbStack containers.",
}
