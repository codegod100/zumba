package conda

import "encoding/json"

// RepoData represents the repodata.json structure from conda channels
type RepoData struct {
	Info        RepoInfo            `json:"info,omitempty"`
	Packages    map[string]Package  `json:"packages,omitempty"`
	PackagesCon map[string]Package  `json:"packages.conda,omitempty"`
	Removed     []string            `json:"removed,omitempty"`
	RepodataVersion int             `json:"repodata_version,omitempty"`
}

// RepoInfo contains channel metadata
type RepoInfo struct {
	Subdir              string `json:"subdir,omitempty"`
	BaseURL             string `json:"base_url,omitempty"`
	Platform            string `json:"platform,omitempty"`
	Arch                string `json:"arch,omitempty"`
	DefaultURL          string `json:"default_url,omitempty"`
	Version             string `json:"version,omitempty"`
}

// Package represents a single package entry in repodata
type Package struct {
	Name          string          `json:"name,omitempty"`
	Version       string          `json:"version,omitempty"`
	Build         string          `json:"build,omitempty"`
	BuildNumber   int             `json:"build_number,omitempty"`
	Depends       []string        `json:"depends,omitempty"`
	License       string          `json:"license,omitempty"`
	LicenseFamily string          `json:"license_family,omitempty"`
	MD5           string          `json:"md5,omitempty"`
	SHA256        string          `json:"sha256,omitempty"`
	Size          int64           `json:"size,omitempty"`
	Subdir        string          `json:"subdir,omitempty"`
	Timestamp     int64           `json:"timestamp,omitempty"`
	TrackFeatures StringOrArray   `json:"track_features,omitempty"`
	Features      string          `json:"features,omitempty"`
	Noarch        StringOrBool    `json:"noarch,omitempty"`
	Platform      string          `json:"platform,omitempty"`
	DependsName   []string        `json:"depends_name,omitempty"`
	Constrains    []string        `json:"constrains,omitempty"`
}

// StringOrArray handles JSON fields that can be either string or []string
type StringOrArray []string

func (s *StringOrArray) UnmarshalJSON(data []byte) error {
	// Try as string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		if str != "" {
			*s = []string{str}
		}
		return nil
	}
	// Try as array
	var arr []string
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	*s = arr
	return nil
}

// StringOrBool handles JSON fields that can be either string or bool
type StringOrBool string

func (s *StringOrBool) UnmarshalJSON(data []byte) error {
	// Try as bool first
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		if b {
			*s = "true"
		}
		return nil
	}
	// Try as string
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*s = StringOrBool(str)
	return nil
}

// Subdirs handles fields that can be either map or array
type Subdirs struct {
	AsMap   map[string]SubdirInfo
	AsArray []string
	IsMap   bool
}

func (s *Subdirs) UnmarshalJSON(data []byte) error {
	// Try as array first
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		s.AsArray = arr
		s.IsMap = false
		return nil
	}
	// Try as map
	var m map[string]SubdirInfo
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	s.AsMap = m
	s.IsMap = true
	return nil
}

func (s *Subdirs) Platforms() []string {
	if s.IsMap {
		var platforms []string
		for k := range s.AsMap {
			platforms = append(platforms, k)
		}
		return platforms
	}
	return s.AsArray
}

// ChannelData represents channeldata.json from conda channels
type ChannelData struct {
	ChannelDataInfo ChannelDataInfo          `json:"channeldata_info,omitempty"`
	Packages        map[string]PackageInfo   `json:"packages,omitempty"`
}

// ChannelDataInfo contains channel metadata
type ChannelDataInfo struct {
	Subdirs          []string `json:"subdirs,omitempty"`
	BaseURL          string   `json:"base_url,omitempty"`
	Platform         string   `json:"platform,omitempty"`
	Arch             string   `json:"arch,omitempty"`
	DefaultURL       string   `json:"default_url,omitempty"`
	Modified         int64    `json:"modified,omitempty"`
	SubdirMapversion map[string]int `json:"subdir_mapversion,omitempty"`
}

// PackageInfo contains detailed package metadata from channeldata
type PackageInfo struct {
	Description   string        `json:"description,omitempty"`
	DevURL        StringOrArray `json:"dev_url,omitempty"`
	DocURL        StringOrArray `json:"doc_url,omitempty"`
	DocSourceURL  string        `json:"doc_source_url,omitempty"`
	Home          StringOrArray `json:"home,omitempty"`
	License       string        `json:"license,omitempty"`
	SourceGitURL  string        `json:"source_git_url,omitempty"`
	SourceGitTag  string        `json:"source_git_tag,omitempty"`
	SourceURL     StringOrArray `json:"source_url,omitempty"`
	Subdirs       Subdirs       `json:"subdirs,omitempty"`
	Summary       string        `json:"summary,omitempty"`
	Timestamp     int64         `json:"timestamp,omitempty"`
	Version       string        `json:"version,omitempty"`
	ActivateD     bool          `json:"activate_d,omitempty"`
	LicenseFamily string        `json:"license_family,omitempty"`
	Identifiers   []interface{} `json:"identifiers,omitempty"`
	Keywords      StringOrArray `json:"keywords,omitempty"`
	Name          string        `json:"name,omitempty"`
	Owner         string        `json:"owner,omitempty"`
	Recipe        string        `json:"recipe,omitempty"`
	SourceGitRev  string        `json:"source_git_rev,omitempty"`
	Upstream      string        `json:"upstream,omitempty"`
	Versions      StringOrArray `json:"versions,omitempty"`
}

// SubdirInfo contains platform-specific package info
type SubdirInfo struct {
	Depends      []string `json:"depends,omitempty"`
	Sha256      string   `json:"sha256,omitempty"`
	Size        int64    `json:"size,omitempty"`
	Subdir      string   `json:"subdir,omitempty"`
	Timestamp   int64    `json:"timestamp,omitempty"`
	Version     string   `json:"version,omitempty"`
	Build       string   `json:"build,omitempty"`
	BuildNumber int      `json:"build_number,omitempty"`
	License     string   `json:"license,omitempty"`
	MD5         string   `json:"md5,omitempty"`
	Noarch      string   `json:"noarch,omitempty"`
}

// SearchResult combines package data with metadata
type SearchResult struct {
	Name        string
	Version     string
	Summary     string
	Description string
	Homepage    string
	License     string
	Platforms   []string
	Versions    []string
	MatchScore  int// Higher = better match
}