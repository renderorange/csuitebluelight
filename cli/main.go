package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

var statusURLs = map[string]string{
	"overall": "https://content.fcsuite.com/deploy/deploy",
	"au":      "https://content.fcsuite.com/deploy/deploy-au",
	"ca":      "https://content.fcsuite.com/deploy/deploy-ca",
	"or":      "https://content.fcsuite.com/deploy/deploy-or",
	"us":      "https://content.fcsuite.com/deploy/deploy-us",
}

var regions = []string{"overall", "au", "ca", "or", "us"}

// Status colors
var (
	redStatuses   = []string{"testfail", "error"}
	greenStatuses = []string{"testok", "testing", "merging", "building", "deploy"}
	blueStatuses  = []string{"pr"}
)

type statusResult struct {
	region string
	status string
	err    error
}

// cachedStatus represents a status entry stored on disk
type cachedStatus struct {
	Status    string    `json:"status"`
	Error     string    `json:"error,omitempty"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// StatusCache stores deployment statuses in memory and persists to disk
type StatusCache struct {
	mu       sync.RWMutex
	statuses map[string]cachedStatus
	filePath string
}

// getCacheDir returns the cache directory path using OS-appropriate location
func getCacheDir() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "csuitebluelight"), nil
}

// NewStatusCache creates a new StatusCache and loads existing data from disk
func NewStatusCache() (*StatusCache, error) {
	cacheDir, err := getCacheDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache directory: %w", err)
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	cache := &StatusCache{
		statuses: make(map[string]cachedStatus),
		filePath: filepath.Join(cacheDir, "statuses.json"),
	}

	cache.load()
	return cache, nil
}

// NewStatusCacheWithPath creates a StatusCache with a custom file path (for testing)
func NewStatusCacheWithPath(filePath string) *StatusCache {
	cache := &StatusCache{
		statuses: make(map[string]cachedStatus),
		filePath: filePath,
	}
	cache.load()
	return cache
}

// load reads the cache from disk
func (c *StatusCache) load() {
	data, err := os.ReadFile(c.filePath)
	if err != nil {
		return // File doesn't exist or can't be read, start with empty cache
	}

	var statuses map[string]cachedStatus
	if err := json.Unmarshal(data, &statuses); err != nil {
		return // Invalid JSON, start with empty cache
	}

	c.mu.Lock()
	c.statuses = statuses
	c.mu.Unlock()
}

// save writes the cache to disk
func (c *StatusCache) save() error {
	c.mu.RLock()
	data, err := json.MarshalIndent(c.statuses, "", "  ")
	c.mu.RUnlock()

	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	if err := os.WriteFile(c.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache: %w", err)
	}

	return nil
}

// Update stores a status result in the cache and saves to disk
func (c *StatusCache) Update(result statusResult) error {
	c.mu.Lock()
	cached := cachedStatus{
		Status:    result.status,
		UpdatedAt: time.Now(),
	}
	if result.err != nil {
		cached.Error = result.err.Error()
	}
	c.statuses[result.region] = cached
	c.mu.Unlock()

	return c.save()
}

// UpdateAll stores multiple status results and saves once
func (c *StatusCache) UpdateAll(results map[string]statusResult) error {
	c.mu.Lock()
	now := time.Now()
	for region, result := range results {
		cached := cachedStatus{
			Status:    result.status,
			UpdatedAt: now,
		}
		if result.err != nil {
			cached.Error = result.err.Error()
		}
		c.statuses[region] = cached
	}
	c.mu.Unlock()

	return c.save()
}

// Get retrieves a status result from the cache
func (c *StatusCache) Get(region string) (statusResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, ok := c.statuses[region]
	if !ok {
		return statusResult{}, false
	}

	result := statusResult{
		region: region,
		status: cached.Status,
	}
	if cached.Error != "" {
		result.err = fmt.Errorf("%s", cached.Error)
	}
	return result, true
}

// GetAll returns all statuses from the cache
func (c *StatusCache) GetAll() map[string]statusResult {
	c.mu.RLock()
	defer c.mu.RUnlock()

	results := make(map[string]statusResult, len(c.statuses))
	for region, cached := range c.statuses {
		result := statusResult{
			region: region,
			status: cached.Status,
		}
		if cached.Error != "" {
			result.err = fmt.Errorf("%s", cached.Error)
		}
		results[region] = result
	}
	return results
}

// GetUpdatedAt returns the last update time for a region
func (c *StatusCache) GetUpdatedAt(region string) (time.Time, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, ok := c.statuses[region]
	if !ok {
		return time.Time{}, false
	}
	return cached.UpdatedAt, true
}

// httpClient is the HTTP client used for fetching statuses (overridable for testing)
var httpClient = &http.Client{Timeout: 10 * time.Second}

func getStatusColor(status string) *color.Color {
	if status == "" {
		return color.New(color.FgRed)
	}

	lower := strings.ToLower(status)

	for _, s := range redStatuses {
		if lower == s {
			return color.New(color.FgRed)
		}
	}
	for _, s := range greenStatuses {
		if lower == s {
			return color.New(color.FgGreen)
		}
	}
	for _, s := range blueStatuses {
		if lower == s {
			return color.New(color.FgBlue)
		}
	}

	return color.New(color.FgWhite)
}

func fetchStatus(region string) statusResult {
	url := statusURLs[region]

	resp, err := httpClient.Get(url)
	if err != nil {
		return statusResult{region: region, err: err}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return statusResult{region: region, err: err}
	}

	return statusResult{region: region, status: strings.TrimSpace(string(body))}
}

func fetchAllStatuses(cache *StatusCache) {
	results := make(map[string]statusResult)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, region := range regions {
		wg.Add(1)
		go func(r string) {
			defer wg.Done()
			result := fetchStatus(r)
			mu.Lock()
			results[r] = result
			mu.Unlock()
		}(region)
	}

	wg.Wait()
	cache.UpdateAll(results)
}

func clearScreen() {
	fmt.Print("\033[2J\033[H")
}

func printStatus(cache *StatusCache, showTimestamp bool) {
	bold := color.New(color.Bold)
	gray := color.New(color.FgHiBlack)

	statuses := cache.GetAll()

	bold.Println("CSuite Deploy Status")
	fmt.Println()

	// Overall status
	result := statuses["overall"]
	value := result.status
	if result.err != nil {
		value = result.err.Error()
	}
	statusColor := getStatusColor(result.status)
	if result.err != nil {
		statusColor = color.New(color.FgRed)
	}

	fmt.Printf("%-10s ", "Status")
	statusColor.Println(value)

	// Regional statuses
	for _, region := range []string{"au", "ca", "or", "us"} {
		result := statuses[region]
		value := result.status
		if result.err != nil {
			value = result.err.Error()
		}
		statusColor := getStatusColor(result.status)
		if result.err != nil {
			statusColor = color.New(color.FgRed)
		}

		fmt.Printf("%-10s ", strings.ToUpper(region))
		statusColor.Println(value)
	}

	if showTimestamp {
		fmt.Println()
		if updatedAt, ok := cache.GetUpdatedAt("overall"); ok {
			gray.Printf("Last updated: %s\n", updatedAt.Format("15:04:05"))
		}
		gray.Println("Press Ctrl+C to exit")
	}
}

func main() {
	watch := flag.Bool("watch", false, "Continuously refresh status")
	flag.Parse()

	cache, err := NewStatusCache()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing cache: %v\n", err)
		os.Exit(1)
	}

	if *watch {
		for {
			clearScreen()
			fetchAllStatuses(cache)
			printStatus(cache, true)
			time.Sleep(60 * time.Second)
		}
	} else {
		fetchAllStatuses(cache)
		printStatus(cache, false)
	}
}
