package conda

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	DefaultChannel = "conda-forge"
	DefaultPlatform = "noarch"
	CacheDirEnv     = "PKGSPEC_CACHE_DIR"
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
		c.CacheDir = filepath.Join(homeDir, ".cache", "pkgspec")
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

// FetchRepoData downloads repodata.json for the channel/platform
func (c *Client) FetchRepoData(forceRefresh bool) (*RepoData, error) {
	cacheFile := filepath.Join(c.CacheDir, c.Channel, c.Platform, "repodata.json")
	
	// Check cache
	if !forceRefresh {
		if cached, err := c.loadCachedRepoData(cacheFile); err == nil {
			return cached, nil
		}
	}
	
	// Download
	url := fmt.Sprintf("https://conda.anaconda.org/%s/%s/repodata.json", c.Channel, c.Platform)
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
func (c *Client) FetchChannelData(forceRefresh bool) (*ChannelData, error) {
	cacheFile := filepath.Join(c.CacheDir, c.Channel, "channeldata.json")
	
	// Check cache
	if !forceRefresh {
		if cached, err := c.loadCachedChannelData(cacheFile); err == nil {
			return cached, nil
		}
	}
	
	// Download
	url := fmt.Sprintf("https://conda.anaconda.org/%s/channeldata.json", c.Channel)
	data, err := c.download(url)
	if err != nil {
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

func (c *Client) cacheData(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}