package main

import (
	"os"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "zumba",
		Short: "Search conda packages with enhanced metadata",
		Long:  `A CLI tool to search conda packages and display metadata that mamba doesn't provide.`,
	}
	
	rootCmd.AddCommand(newSearchCmd())
	rootCmd.AddCommand(newInfoCmd())
	
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}