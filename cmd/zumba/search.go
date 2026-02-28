package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/codegod100/zumba/internal/conda"
	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	var (
		channel      string
		platform     string
		forceRefresh bool
		outputJSON   bool
		wide         bool
	)
	
	cmd := &cobra.Command{
		Use:   "search TERM",
		Short: "Search for packages",
		Long:  `Search for conda packages matching the given term.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]
			
			client := conda.NewClient(
				conda.WithChannel(channel),
				conda.WithPlatform(platform),
			)
			
			// Fetch data
			repodata, err := client.FetchRepoData(forceRefresh)
			if err != nil {
				return fmt.Errorf("failed to fetch repodata: %w", err)
			}
			
			channelData, err := client.FetchChannelData(forceRefresh)
			if err != nil {
				return fmt.Errorf("failed to fetch channeldata: %w", err)
			}
			
			// Search
			results := conda.Search(repodata, channelData, query)
			
			if len(results) == 0 {
				fmt.Fprintf(os.Stderr, "No packages found matching %q\n", query)
				return nil
			}
			
			// Output
			if outputJSON {
				return outputJSONResults(results)
			}
			
			if wide {
				return outputWideResults(results)
			}
			
			return outputTableResults(results)
		},
	}
	
	cmd.Flags().StringVarP(&channel, "channel", "c", conda.DefaultChannel, "conda channel to search")
	cmd.Flags().StringVarP(&platform, "platform", "p", conda.DefaultPlatform, "platform (e.g., noarch, linux-64)")
	cmd.Flags().BoolVarP(&forceRefresh, "refresh", "r", false, "force refresh cached data")
	cmd.Flags().BoolVarP(&outputJSON, "json", "j", false, "output as JSON")
	cmd.Flags().BoolVarP(&wide, "wide", "w", false, "show more columns")
	
	return cmd
}

func outputTableResults(results []conda.SearchResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tVERSION\tLICENSE\tSUMMARY")
	
	for _, r := range results {
		summary := r.Summary
		if len(summary) > 60 {
			summary = summary[:57] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.Name, r.Version, r.License, summary)
	}
	
	return w.Flush()
}

func outputWideResults(results []conda.SearchResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tVERSION\tLICENSE\tHOMEPAGE\tSUMMARY")
	
	for _, r := range results {
		summary := r.Summary
		if len(summary) > 50 {
			summary = summary[:47] + "..."
		}
		homepage := r.Homepage
		if len(homepage) > 40 {
			homepage = homepage[:37] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", r.Name, r.Version, r.License, homepage, summary)
	}
	
	return w.Flush()
}

func outputJSONResults(results []conda.SearchResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}