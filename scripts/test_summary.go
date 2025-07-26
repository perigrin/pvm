// ABOUTME: Test summary reporter that parses gotestsum JSON output
// ABOUTME: Provides structured failure analysis for make test command

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// TestEvent represents a test event from gotestsum JSON output
type TestEvent struct {
	Time    string  `json:"Time"`
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test"`
	Elapsed float64 `json:"Elapsed"`
	Output  string  `json:"Output"`
}

// PackageStats holds statistics for a package
type PackageStats struct {
	Name     string
	Passed   int
	Failed   int
	Skipped  int
	Tests    []string
	Failures []string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <test-results.json>\n", os.Args[0])
		os.Exit(1)
	}

	filename := os.Args[1]
	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening %s: %v\n", filename, err)
		os.Exit(1)
	}
	defer file.Close()

	packageStats := make(map[string]*PackageStats)
	var totalPassed, totalFailed, totalSkipped int

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var event TestEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue // Skip malformed lines
		}

		pkg := event.Package
		if pkg == "" {
			continue
		}

		// Skip non-test and non-package events (only process pass/fail/skip actions)
		if event.Action != "pass" && event.Action != "fail" && event.Action != "skip" {
			continue
		}

		// Initialize package stats if needed
		if packageStats[pkg] == nil {
			packageStats[pkg] = &PackageStats{
				Name:     pkg,
				Tests:    []string{},
				Failures: []string{},
			}
		}

		stats := packageStats[pkg]

		// Determine test name (use "Package-level failure" for package failures without test name)
		testName := event.Test
		if testName == "" && event.Action == "fail" {
			testName = "Package-level failure"
		} else if testName == "" {
			// Skip package-level pass/skip events (these are summary events, not actual tests)
			continue
		}

		switch event.Action {
		case "pass":
			stats.Passed++
			totalPassed++
			stats.Tests = append(stats.Tests, testName)
		case "fail":
			stats.Failed++
			totalFailed++
			stats.Tests = append(stats.Tests, testName)
			stats.Failures = append(stats.Failures, testName)
		case "skip":
			stats.Skipped++
			totalSkipped++
			stats.Tests = append(stats.Tests, testName)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		return
	}

	// Generate summary report
	generateSummaryReport(packageStats, totalPassed, totalFailed, totalSkipped)
}

func generateSummaryReport(packageStats map[string]*PackageStats, totalPassed, totalFailed, totalSkipped int) {
	totalTests := totalPassed + totalFailed + totalSkipped

	// Overall statistics
	fmt.Printf("📈 OVERALL STATISTICS\n")
	fmt.Printf("Total Tests:   %d\n", totalTests)
	fmt.Printf("✅ Passed:     %d (%.1f%%)\n", totalPassed, float64(totalPassed)/float64(totalTests)*100)
	fmt.Printf("❌ Failed:     %d (%.1f%%)\n", totalFailed, float64(totalFailed)/float64(totalTests)*100)
	fmt.Printf("⏭️  Skipped:    %d (%.1f%%)\n", totalSkipped, float64(totalSkipped)/float64(totalTests)*100)
	fmt.Printf("\n")

	if totalFailed == 0 {
		fmt.Printf("🎉 ALL TESTS PASSING! Great job!\n")
		return
	}

	// Package-level breakdown for failures
	fmt.Printf("📦 FAILURE BREAKDOWN BY PACKAGE\n")

	// Sort packages by failure count (highest first)
	var packages []*PackageStats
	for _, stats := range packageStats {
		if stats.Failed > 0 {
			packages = append(packages, stats)
		}
	}

	sort.Slice(packages, func(i, j int) bool {
		return packages[i].Failed > packages[j].Failed
	})

	for _, stats := range packages {
		fmt.Printf("\n🔸 %s\n", stats.Name)
		fmt.Printf("   Failed: %d/%d tests\n", stats.Failed, len(stats.Tests))

		if len(stats.Failures) <= 5 {
			// Show all failures if 5 or fewer
			for _, failure := range stats.Failures {
				fmt.Printf("   • %s\n", failure)
			}
		} else {
			// Show first 5 and count
			for i, failure := range stats.Failures[:5] {
				fmt.Printf("   • %s\n", failure)
				if i == 4 {
					fmt.Printf("   ... and %d more failures\n", len(stats.Failures)-5)
				}
			}
		}
	}

	// Focus areas
	fmt.Printf("\n🎯 PRIORITY FOCUS AREAS\n")
	if len(packages) > 0 {
		fmt.Printf("1. %s (%d failures) - Start here\n", packages[0].Name, packages[0].Failed)
		if len(packages) > 1 {
			fmt.Printf("2. %s (%d failures)\n", packages[1].Name, packages[1].Failed)
		}
		if len(packages) > 2 {
			fmt.Printf("3. %s (%d failures)\n", packages[2].Name, packages[2].Failed)
		}
	}

	// Progress tracking (compare to previous runs if we had historical data)
	fmt.Printf("\n📊 PROGRESS TRACKING\n")
	failureRate := float64(totalFailed) / float64(totalTests) * 100
	switch {
	case failureRate <= 5:
		fmt.Printf("🟢 Excellent: %.1f%% failure rate (Very close to 100%% pass rate!)\n", failureRate)
	case failureRate <= 15:
		fmt.Printf("🟡 Good: %.1f%% failure rate (Making solid progress)\n", failureRate)
	case failureRate <= 30:
		fmt.Printf("🟠 Moderate: %.1f%% failure rate (Room for improvement)\n", failureRate)
	default:
		fmt.Printf("🔴 High: %.1f%% failure rate (Needs attention)\n", failureRate)
	}
}
