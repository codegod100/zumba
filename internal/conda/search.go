package conda

import (
	"strings"
)

// Search searches for packages matching the query
func Search(repodata *RepoData, channelData *ChannelData, query string) []SearchResult {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return nil
	}
	
	// Collect unique package names from repodata
	packageNames := make(map[string]bool)
	for _, pkg := range repodata.Packages {
		if pkg.Name != "" {
			packageNames[pkg.Name] = true
		}
	}
	for _, pkg := range repodata.PackagesCon {
		if pkg.Name != "" {
			packageNames[pkg.Name] = true
		}
	}
	results := make([]SearchResult, 0, len(packageNames))
	
	for name := range packageNames {
		score := matchScore(strings.ToLower(name), query, channelData)
		if score > 0 {
			result := SearchResult{
				Name:       name,
				MatchScore: score,
			}
			// Enrich with metadata from channeldata
			if pkgInfo, ok := channelData.Packages[name]; ok {
				result.Summary = pkgInfo.Summary
				result.Description = pkgInfo.Description
				if len(pkgInfo.Home) > 0 {
					result.Homepage = pkgInfo.Home[0]
				}
				result.License = pkgInfo.License
				result.Version = pkgInfo.Version
				result.Versions = pkgInfo.Versions
				result.Platforms = pkgInfo.Subdirs.Platforms()
			}
			results = append(results, result)
		}
	}
	
	// Sort by match score (descending), then by name
	sortResults(results)
	
	return results
}

func matchScore(name, query string, channelData *ChannelData) int {
	// Exact match
	if name == query {
		return 100
	}
	
	// Starts with query
	if strings.HasPrefix(name, query) {
		return 80
	}
	// Contains query
	if strings.Contains(name, query) {
		return 60
	}
	
	// Check summary for query
	if pkgInfo, ok := channelData.Packages[name]; ok {
		summary := strings.ToLower(pkgInfo.Summary)
		if strings.Contains(summary, query) {
			return 40
		}
		desc := strings.ToLower(pkgInfo.Description)
		if strings.Contains(desc, query) {
			return 20
		}
	}
	
	return 0
}

func sortResults(results []SearchResult) {
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].MatchScore < results[j].MatchScore ||
				(results[i].MatchScore == results[j].MatchScore && results[i].Name > results[j].Name) {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}