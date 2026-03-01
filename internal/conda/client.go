package conda

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	DefaultChannel = "conda-forge"
	DefaultPlatform = "noarch"
	CacheDirEnv     = "ZUMBA_CACHE_DIR"
	CacheTTL        = 1 * time.Hour
)

// Client handles fetching and caching conda channel data
type Client struct {
	HTTPClient *http.Client
	CacheDir   string
	Channel    string
	Platform   string
}

// NewClient creates a new conda client
func NewClient(opts ...Option) *Client {
	c := &Client{
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		Channel:    DefaultChannel,
		Platform:   DefaultPlatform,
	}
	
	// Set cache directory
	if dir := os.Getenv(CacheDirEnv); dir != "" {
		c.CacheDir = dir
	} else {
		homeDir, _ := os.UserHomeDir()
		c.CacheDir = filepath.Join(homeDir, ".cache", "zumba")
	}
	
	for _, opt := range opts {
		opt(c)
	}
	
	return c
}

type Option func(*Client)

func WithChannel(channel string) Option {
	return func(c *Client) {
		c.Channel = channel
	}
}

func WithPlatform(platform string) Option {
	return func(c *Client) {
		c.Platform = platform
	}
}

func WithCacheDir(dir string) Option {
	return func(c *Client) {
		c.CacheDir = dir
	}
}

// isFullURL checks if the channel is a full URL
func isFullURL(channel string) bool {
	return strings.HasPrefix(channel, "http://") || strings.HasPrefix(channel, "https://")
}

// channelBaseURL returns the base URL for a channel
func channelBaseURL(channel string) string {
	if isFullURL(channel) {
		return strings.TrimSuffix(channel, "/")
	}
	return "https://conda.anaconda.org/" + channel
}

// isPrefixDev checks if the channel is a prefix.dev URL
func isPrefixDev(channel string) bool {
	return strings.Contains(channel, "prefix.dev")
}

// IsPrefixDev checks if the channel is a prefix.dev URL (exported)
func IsPrefixDev(channel string) bool {
	return isPrefixDev(channel)
}

// prefixDevChannelName extracts the channel name from a prefix.dev URL
func prefixDevChannelName(channel string) string {
	// https://prefix.dev/nandi-testing -> nandi-testing
	parts := strings.Split(strings.TrimSuffix(channel, "/"), "/")
	return parts[len(parts)-1]
}

// PrefixDevPackage represents package info from prefix.dev GraphQL API
type PrefixDevPackage struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Variants    struct {
		Page []struct {
			Version     string `json:"version"`
			BuildString string `json:"buildString"`
			Platform    string `json:"platform"`
			Filename    string `json:"filename"`
			Size        int64  `json:"size"`
			RawIndex    struct {
				Depends  []string `json:"depends"`
				License  string   `json:"license"`
				Subdir   string   `json:"subdir"`
			} `json:"rawIndex"`
			RawAbout struct {
				Home        string `json:"home"`
				DevURL      string `json:"dev_url"`
				Summary     string `json:"summary"`
				Description string `json:"description"`
				License     string `json:"license"`
			} `json:"rawAbout"`
		} `json:"page"`
	} `json:"variants"`
}

// FetchPrefixDevPackage fetches package info from prefix.dev GraphQL API
func (c *Client) FetchPrefixDevPackage(pkgName string) (*PrefixDevPackage, error) {
	channelName := prefixDevChannelName(c.Channel)
	
	query := map[string]interface{}{
		"query": fmt.Sprintf(`{
			package(channelName: "%s", name: "%s") {
				name
				description
				variants(limit: 10) {
					page {
						version
						buildString
						platform
						filename
						size
						rawIndex
						rawAbout
					}
				}
			}
		}`, channelName, pkgName),
	}
	
	body, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	
	resp, err := c.HTTPClient.Post("https://prefix.dev/api/graphql", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GraphQL request failed: HTTP %d", resp.StatusCode)
	}
	
	var result struct {
		Data struct {
			Package PrefixDevPackage `json:"package"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL error: %s", result.Errors[0].Message)
	}
	
	if result.Data.Package.Name == "" {
		return nil, fmt.Errorf("package not found")
	}
	
	return &result.Data.Package, nil
}

// FetchRepoData downloads repodata.json for the channel/platform
func (c *Client) FetchRepoData(forceRefresh bool) (*RepoData, error) {
	cacheFile := filepath.Join(c.CacheDir, strings.ReplaceAll(c.Channel, "/", "_"), c.Platform, "repodata.json")
	
	// Check cache
	if !forceRefresh {
		if cached, err := c.loadCachedRepoData(cacheFile); err == nil {
			return cached, nil
		}
	}
	
	// Download
	baseURL := channelBaseURL(c.Channel)
	url := fmt.Sprintf("%s/%s/repodata.json", baseURL, c.Platform)
	data, err := c.download(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repodata: %w", err)
	}
	
	// Parse
	var repodata RepoData
	if err := json.Unmarshal(data, &repodata); err != nil {
		return nil, fmt.Errorf("failed to parse repodata: %w", err)
	}
	
	// Cache it
	if err := c.cacheData(cacheFile, data); err != nil {
		// Non-fatal: continue without caching
		fmt.Fprintf(os.Stderr, "Warning: failed to cache repodata: %v\n", err)
	}
	
	return &repodata, nil
}
func (c *Client) loadCachedRepoData(path string) (*RepoData, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	// Check if cache is still valid
	if time.Since(info.ModTime()) > CacheTTL {
		return nil, fmt.Errorf("cache expired")
	}
	
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var repodata RepoData
	if err := json.Unmarshal(data, &repodata); err != nil {
		return nil, err
	}
	return &repodata, nil
}

// FetchChannelData downloads channeldata.json for the channel
// Returns nil if not available (some channels don't have it)
func (c *Client) FetchChannelData(forceRefresh bool) (*ChannelData, error) {
	cacheFile := filepath.Join(c.CacheDir, strings.ReplaceAll(c.Channel, "/", "_"), "channeldata.json")
	
	// Check cache
	if !forceRefresh {
		if cached, err := c.loadCachedChannelData(cacheFile); err == nil {
			return cached, nil
		}
	}
	
	// Download
	baseURL := channelBaseURL(c.Channel)
	url := fmt.Sprintf("%s/channeldata.json", baseURL)
	data, err := c.download(url)
	if err != nil {
		// channeldata.json is optional - return empty data if not found
		if strings.Contains(err.Error(), "HTTP 404") {
			return &ChannelData{Packages: make(map[string]PackageInfo)}, nil
		}
		return nil, fmt.Errorf("failed to fetch channeldata: %w", err)
	}
	
	// Parse
	var channelData ChannelData
	if err := json.Unmarshal(data, &channelData); err != nil {
		return nil, fmt.Errorf("failed to parse channeldata: %w", err)
	}
	
	// Cache it
	if err := c.cacheData(cacheFile, data); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to cache channeldata: %v\n", err)
	}
	return &channelData, nil
}

func (c *Client) loadCachedChannelData(path string) (*ChannelData, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if time.Since(info.ModTime()) > CacheTTL {
		return nil, fmt.Errorf("cache expired")
	}
	
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var channelData ChannelData
	if err := json.Unmarshal(data, &channelData); err != nil {
		return nil, err
	}
	return &channelData, nil
}

func (c *Client) download(url string) ([]byte, error) {
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, url)
	}
	return io.ReadAll(resp.Body)
}

// FetchStreaming returns a reader for streaming the URL content
func (c *Client) FetchStreaming(url string) (io.ReadCloser, error) {
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, url)
	}
	return resp.Body, nil
}

func (c *Client) cacheData(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// StreamPackageFromChannelData fetches and streams channeldata.json to find a package
func (c *Client) StreamPackageFromChannelData(pkgName string) (*PackageInfo, error) {
	baseURL := channelBaseURL(c.Channel)
	url := fmt.Sprintf("%s/channeldata.json", baseURL)
	
	rc, err := c.FetchStreaming(url)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	
	pkgInfo, found := PackageExistsInChannelData(rc, pkgName)
	if !found {
		return nil, fmt.Errorf("package %q not found", pkgName)
	}
	return pkgInfo, nil
}

// StreamPackageFromRepoData fetches and streams repodata.json to find a package
func (c *Client) StreamPackageFromRepoData(pkgName string) (*Package, error) {
	baseURL := channelBaseURL(c.Channel)
	url := fmt.Sprintf("%s/%s/repodata.json", baseURL, c.Platform)
	
	rc, err := c.FetchStreaming(url)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	
	pkg, found := PackageExistsInRepoData(rc, pkgName)
	if !found {
		return nil, fmt.Errorf("package %q not found", pkgName)
	}
	return pkg, nil
}