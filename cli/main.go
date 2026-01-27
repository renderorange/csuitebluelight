package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
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

func fetchAllStatuses() map[string]statusResult {
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
	return results
}

func clearScreen() {
	fmt.Print("\033[2J\033[H")
}

func printStatus(statuses map[string]statusResult, showTimestamp bool) {
	bold := color.New(color.Bold)
	gray := color.New(color.FgHiBlack)

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
		gray.Printf("Last updated: %s\n", time.Now().Format("15:04:05"))
		gray.Println("Press Ctrl+C to exit")
	}
}

func main() {
	watch := flag.Bool("watch", false, "Continuously refresh status")
	interval := flag.Int("interval", 60, "Refresh interval in seconds")
	flag.Parse()

	if *watch {
		for {
			clearScreen()
			statuses := fetchAllStatuses()
			printStatus(statuses, true)
			time.Sleep(time.Duration(*interval) * time.Second)
		}
	} else {
		statuses := fetchAllStatuses()
		printStatus(statuses, false)
	}
}
