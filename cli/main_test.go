package main

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/fatih/color"
)

// mockTransport implements http.RoundTripper for testing
type mockTransport struct {
	responses map[string]mockResponse
}

type mockResponse struct {
	body string
	err  error
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, ok := m.responses[req.URL.String()]
	if !ok {
		return nil, errors.New("no mock response for " + req.URL.String())
	}

	if resp.err != nil {
		return nil, resp.err
	}

	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(resp.body)),
	}, nil
}

func TestGetStatusColor_RedStatuses(t *testing.T) {
	tests := []string{"testfail", "error", "TESTFAIL", "Error"}
	expected := color.New(color.FgRed)

	for _, status := range tests {
		result := getStatusColor(status)
		if result.Sprint("x") != expected.Sprint("x") {
			t.Errorf("getStatusColor(%q) should return red", status)
		}
	}
}

func TestGetStatusColor_GreenStatuses(t *testing.T) {
	tests := []string{"testok", "testing", "merging", "building", "deploy", "TESTING"}
	expected := color.New(color.FgGreen)

	for _, status := range tests {
		result := getStatusColor(status)
		if result.Sprint("x") != expected.Sprint("x") {
			t.Errorf("getStatusColor(%q) should return green", status)
		}
	}
}

func TestGetStatusColor_BlueStatuses(t *testing.T) {
	tests := []string{"pr", "PR", "Pr"}
	expected := color.New(color.FgBlue)

	for _, status := range tests {
		result := getStatusColor(status)
		if result.Sprint("x") != expected.Sprint("x") {
			t.Errorf("getStatusColor(%q) should return blue", status)
		}
	}
}

func TestGetStatusColor_WhiteStatuses(t *testing.T) {
	tests := []string{"complete", "unknown", "random"}
	expected := color.New(color.FgWhite)

	for _, status := range tests {
		result := getStatusColor(status)
		if result.Sprint("x") != expected.Sprint("x") {
			t.Errorf("getStatusColor(%q) should return white", status)
		}
	}
}

func TestGetStatusColor_EmptyStringReturnsRed(t *testing.T) {
	result := getStatusColor("")
	expected := color.New(color.FgRed)

	if result.Sprint("x") != expected.Sprint("x") {
		t.Error("getStatusColor(\"\") should return red")
	}
}

func TestFetchStatus_Success(t *testing.T) {
	originalClient := httpClient
	defer func() { httpClient = originalClient }()

	httpClient = &http.Client{
		Transport: &mockTransport{
			responses: map[string]mockResponse{
				statusURLs["overall"]: {body: "complete\n"},
			},
		},
	}

	result := fetchStatus("overall")

	if result.err != nil {
		t.Errorf("unexpected error: %v", result.err)
	}
	if result.status != "complete" {
		t.Errorf("expected 'complete', got %q", result.status)
	}
	if result.region != "overall" {
		t.Errorf("expected region 'overall', got %q", result.region)
	}
}

func TestFetchStatus_Error(t *testing.T) {
	originalClient := httpClient
	defer func() { httpClient = originalClient }()

	httpClient = &http.Client{
		Transport: &mockTransport{
			responses: map[string]mockResponse{
				statusURLs["au"]: {err: errors.New("connection refused")},
			},
		},
	}

	result := fetchStatus("au")

	if result.err == nil {
		t.Error("expected error, got nil")
	}
	if result.status != "" {
		t.Errorf("expected empty status on error, got %q", result.status)
	}
	if result.region != "au" {
		t.Errorf("expected region 'au', got %q", result.region)
	}
}

func TestFetchAllStatuses_ReturnsAllRegions(t *testing.T) {
	originalClient := httpClient
	defer func() { httpClient = originalClient }()

	responses := make(map[string]mockResponse)
	for _, region := range regions {
		responses[statusURLs[region]] = mockResponse{body: "complete"}
	}

	httpClient = &http.Client{
		Transport: &mockTransport{responses: responses},
	}

	tmpFile := filepath.Join(t.TempDir(), "statuses.json")
	cache := NewStatusCacheWithPath(tmpFile)
	fetchAllStatuses(cache)
	results := cache.GetAll()

	for _, region := range regions {
		result, ok := results[region]
		if !ok {
			t.Errorf("missing region %q in results", region)
			continue
		}
		if result.err != nil {
			t.Errorf("region %q had error: %v", region, result.err)
		}
		if result.status != "complete" {
			t.Errorf("region %q expected 'complete', got %q", region, result.status)
		}
	}
}

func TestFetchAllStatuses_MixedResults(t *testing.T) {
	originalClient := httpClient
	defer func() { httpClient = originalClient }()

	httpClient = &http.Client{
		Transport: &mockTransport{
			responses: map[string]mockResponse{
				statusURLs["overall"]: {body: "testing"},
				statusURLs["au"]:      {body: "complete"},
				statusURLs["ca"]:      {err: errors.New("timeout")},
				statusURLs["or"]:      {body: "pr"},
				statusURLs["us"]:      {body: "building"},
			},
		},
	}

	tmpFile := filepath.Join(t.TempDir(), "statuses.json")
	cache := NewStatusCacheWithPath(tmpFile)
	fetchAllStatuses(cache)
	results := cache.GetAll()

	if results["overall"].status != "testing" {
		t.Errorf("overall: expected 'testing', got %q", results["overall"].status)
	}
	if results["au"].status != "complete" {
		t.Errorf("au: expected 'complete', got %q", results["au"].status)
	}
	if results["ca"].err == nil {
		t.Error("ca: expected error, got nil")
	}
	if results["or"].status != "pr" {
		t.Errorf("or: expected 'pr', got %q", results["or"].status)
	}
	if results["us"].status != "building" {
		t.Errorf("us: expected 'building', got %q", results["us"].status)
	}
}

func TestIntegration_FetchRealStatuses(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpFile := filepath.Join(t.TempDir(), "statuses.json")
	cache := NewStatusCacheWithPath(tmpFile)
	fetchAllStatuses(cache)
	results := cache.GetAll()

	for region, result := range results {
		if result.err != nil {
			t.Errorf("region %q had error: %v", region, result.err)
		}
		if result.status == "" {
			t.Errorf("region %q had empty status", region)
		}
	}
}

func TestStatusCache_PersistsToDisk(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "statuses.json")

	// Create cache and update it
	cache := NewStatusCacheWithPath(tmpFile)
	cache.Update(statusResult{region: "overall", status: "testing"})
	cache.Update(statusResult{region: "au", status: "complete"})

	// Create new cache from same file - should load persisted data
	cache2 := NewStatusCacheWithPath(tmpFile)
	results := cache2.GetAll()

	if results["overall"].status != "testing" {
		t.Errorf("overall: expected 'testing', got %q", results["overall"].status)
	}
	if results["au"].status != "complete" {
		t.Errorf("au: expected 'complete', got %q", results["au"].status)
	}
}

func TestStatusCache_PersistsErrors(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "statuses.json")

	cache := NewStatusCacheWithPath(tmpFile)
	cache.Update(statusResult{region: "ca", err: errors.New("connection refused")})

	cache2 := NewStatusCacheWithPath(tmpFile)
	result, ok := cache2.Get("ca")

	if !ok {
		t.Fatal("expected to find 'ca' in cache")
	}
	if result.err == nil {
		t.Error("expected error to be persisted")
	}
	if result.err.Error() != "connection refused" {
		t.Errorf("expected 'connection refused', got %q", result.err.Error())
	}
}

func TestStatusCache_UpdateAll(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "statuses.json")
	cache := NewStatusCacheWithPath(tmpFile)

	results := map[string]statusResult{
		"overall": {region: "overall", status: "testing"},
		"au":      {region: "au", status: "complete"},
		"ca":      {region: "ca", status: "pr"},
	}
	cache.UpdateAll(results)

	// Verify all were saved
	cache2 := NewStatusCacheWithPath(tmpFile)
	loaded := cache2.GetAll()

	if len(loaded) != 3 {
		t.Errorf("expected 3 statuses, got %d", len(loaded))
	}
	if loaded["overall"].status != "testing" {
		t.Errorf("overall: expected 'testing', got %q", loaded["overall"].status)
	}
}

func TestStatusCache_GetUpdatedAt(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "statuses.json")
	cache := NewStatusCacheWithPath(tmpFile)

	cache.Update(statusResult{region: "overall", status: "testing"})

	updatedAt, ok := cache.GetUpdatedAt("overall")
	if !ok {
		t.Fatal("expected to find 'overall' timestamp")
	}
	if updatedAt.IsZero() {
		t.Error("expected non-zero timestamp")
	}

	_, ok = cache.GetUpdatedAt("nonexistent")
	if ok {
		t.Error("expected false for nonexistent region")
	}
}

func TestStatusCache_EmptyFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "statuses.json")

	// Create empty file
	os.WriteFile(tmpFile, []byte(""), 0644)

	cache := NewStatusCacheWithPath(tmpFile)
	results := cache.GetAll()

	if len(results) != 0 {
		t.Errorf("expected empty cache, got %d entries", len(results))
	}
}

func TestStatusCache_InvalidJSON(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "statuses.json")

	// Create file with invalid JSON
	os.WriteFile(tmpFile, []byte("not valid json"), 0644)

	cache := NewStatusCacheWithPath(tmpFile)
	results := cache.GetAll()

	if len(results) != 0 {
		t.Errorf("expected empty cache for invalid JSON, got %d entries", len(results))
	}
}
