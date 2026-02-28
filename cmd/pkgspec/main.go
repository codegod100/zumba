package main

import (
	"os"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "pkgspec",
		Short: "Search conda packages with enhanced metadata",
		Long:  `A CLI tool to search conda packages and display metadata that mamba doesn't provide.`,
	}
	
	rootCmd.AddCommand(newSearchCmd())
	
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}