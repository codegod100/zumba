package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/codegod100/zumba/internal/conda"
	"github.com/spf13/cobra"
)

func newInfoCmd() *cobra.Command {
	var (
		channel      string
		platform     string
		forceRefresh bool
		outputJSON   bool
	)
	
	cmd := &cobra.Command{
		Use:   "info PACKAGE",
		Short: "Show detailed package information",
		Long:  `Show detailed information about a specific package.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pkgName := args[0]
			
			client := conda.NewClient(
				conda.WithChannel(channel),
				conda.WithPlatform(platform),
			)
			
			// Try channel data first
			channelData, err := client.FetchChannelData(forceRefresh)
			if err != nil {
				return fmt.Errorf("failed to fetch channeldata: %w", err)
			}
			
			// Find package in channeldata
			if pkgInfo, ok := channelData.Packages[pkgName]; ok && pkgInfo.Version != "" {
				if outputJSON {
					enc := json.NewEncoder(os.Stdout)
					enc.SetIndent("", "  ")
					return enc.Encode(pkgInfo)
				}
				return printPackageInfo(pkgName, pkgInfo)
			}
			
			// Fall back to repodata
			repodata, err := client.FetchRepoData(forceRefresh)
			if err != nil {
				return fmt.Errorf("failed to fetch repodata: %w", err)
			}
			
			// Find package in repodata
			pkg := findPackageInRepoData(repodata, pkgName)
			if pkg == nil {
				return fmt.Errorf("package %q not found in channel %s", pkgName, channel)
			}
			
			if outputJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(pkg)
			}
			return printPackageFromRepoData(pkgName, pkg)
		},
	}
	
	cmd.Flags().StringVarP(&channel, "channel", "c", conda.DefaultChannel, "conda channel")
	cmd.Flags().StringVarP(&platform, "platform", "p", conda.DefaultPlatform, "platform")
	cmd.Flags().BoolVarP(&forceRefresh, "refresh", "r", false, "force refresh cached data")
	cmd.Flags().BoolVarP(&outputJSON, "json", "j", false, "output as JSON")
	
	return cmd
}

func printPackageInfo(name string, pkg conda.PackageInfo) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	
	fmt.Fprintf(w, "Name:\t%s\n", name)
	fmt.Fprintf(w, "Version:\t%s\n", pkg.Version)
	fmt.Fprintf(w, "License:\t%s\n", pkg.License)
	
	if len(pkg.Home) > 0 {
		fmt.Fprintf(w, "Homepage:\t%s\n", pkg.Home[0])
	}
	if len(pkg.DevURL) > 0 {
		fmt.Fprintf(w, "Dev URL:\t%s\n", pkg.DevURL[0])
	}
	if pkg.DocURL != nil && len(pkg.DocURL) > 0 {
		fmt.Fprintf(w, "Doc URL:\t%s\n", pkg.DocURL[0])
	}
	if pkg.SourceGitURL != "" {
		fmt.Fprintf(w, "Source:\t%s\n", pkg.SourceGitURL)
	}
	
	fmt.Fprintf(w, "Platforms:\t%v\n", pkg.Subdirs.Platforms())
	
	if len(pkg.Versions) > 0 {
		fmt.Fprintf(w, "Versions:\t%v\n", pkg.Versions)
	}
	
	fmt.Fprintf(w, "\nSummary:\t%s\n", pkg.Summary)
	
	if pkg.Description != "" {
		w.Flush()
		fmt.Printf("\nDescription:\n%s\n", pkg.Description)
		return nil
	}
	
	return w.Flush()
}

func findPackageInRepoData(repodata *conda.RepoData, name string) *conda.Package {
	// Search in both packages and packages.conda
	for _, pkg := range repodata.Packages {
		if pkg.Name == name {
			return &pkg
		}
	}
	for _, pkg := range repodata.PackagesCon {
		if pkg.Name == name {
			return &pkg
		}
	}
	return nil
}

func printPackageFromRepoData(name string, pkg *conda.Package) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	
	fmt.Fprintf(w, "Name:\t%s\n", name)
	fmt.Fprintf(w, "Version:\t%s\n", pkg.Version)
	fmt.Fprintf(w, "Build:\t%s\n", pkg.Build)
	fmt.Fprintf(w, "License:\t%s\n", pkg.License)
	fmt.Fprintf(w, "Platform:\t%s\n", pkg.Subdir)
	fmt.Fprintf(w, "Size:\t%d bytes\n", pkg.Size)
	
	if len(pkg.Depends) > 0 {
		fmt.Fprintf(w, "Depends:\t%v\n", pkg.Depends)
	}
	
	return w.Flush()
}