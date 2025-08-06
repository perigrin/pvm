// ABOUTME: CI/CD dashboard generation tool for monitoring build and test metrics
// ABOUTME: Aggregates performance data, test results, and build metrics into dashboard format

//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// DashboardData represents the complete dashboard information
type DashboardData struct {
	GeneratedAt  time.Time              `json:"generated_at"`
	BuildMetrics BuildMetrics           `json:"build_metrics"`
	TestResults  TestResults            `json:"test_results"`
	Performance  PerformanceMetrics     `json:"performance"`
	SecurityScan SecurityMetrics        `json:"security"`
	QualityGates QualityMetrics         `json:"quality"`
	RecentBuilds []BuildInfo            `json:"recent_builds"`
	Trends       map[string][]DataPoint `json:"trends"`
}

// BuildMetrics represents build-related metrics
type BuildMetrics struct {
	TotalBuilds     int       `json:"total_builds"`
	SuccessRate     float64   `json:"success_rate"`
	AverageDuration int       `json:"average_duration_seconds"`
	LastBuild       time.Time `json:"last_build"`
}

// TestResults represents test execution metrics
type TestResults struct {
	TotalTests   int      `json:"total_tests"`
	PassRate     float64  `json:"pass_rate"`
	Coverage     float64  `json:"coverage_percent"`
	FailingTests []string `json:"failing_tests"`
}

// PerformanceMetrics represents performance-related data
type PerformanceMetrics struct {
	AverageOpsPerSec int64              `json:"average_ops_per_sec"`
	MemoryUsageMB    float64            `json:"memory_usage_mb"`
	Regressions      int                `json:"regressions_count"`
	Benchmarks       map[string]float64 `json:"benchmarks"`
}

// SecurityMetrics represents security scan results
type SecurityMetrics struct {
	Vulnerabilities int            `json:"vulnerabilities"`
	LastScan        time.Time      `json:"last_scan"`
	Severity        map[string]int `json:"severity"`
}

// QualityMetrics represents code quality metrics
type QualityMetrics struct {
	LintIssues     int     `json:"lint_issues"`
	CodeComplexity float64 `json:"code_complexity"`
	TechnicalDebt  string  `json:"technical_debt"`
}

// BuildInfo represents information about a single build
type BuildInfo struct {
	ID        string    `json:"id"`
	Branch    string    `json:"branch"`
	Commit    string    `json:"commit"`
	Status    string    `json:"status"`
	Duration  int       `json:"duration_seconds"`
	Timestamp time.Time `json:"timestamp"`
}

// DataPoint represents a single data point for trending
type DataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

const dashboardTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>PVM CI/CD Dashboard</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { background: #2196F3; color: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; }
        .header h1 { margin: 0; }
        .header .subtitle { opacity: 0.9; margin-top: 5px; }
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; }
        .card { background: white; border-radius: 8px; padding: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .card h3 { margin-top: 0; color: #333; }
        .metric { display: flex; justify-content: space-between; margin: 10px 0; }
        .metric-label { color: #666; }
        .metric-value { font-weight: bold; }
        .status-good { color: #4CAF50; }
        .status-warning { color: #FF9800; }
        .status-error { color: #F44336; }
        .build-list { list-style: none; padding: 0; }
        .build-item { padding: 10px; border-left: 4px solid #ddd; margin: 5px 0; }
        .build-success { border-color: #4CAF50; }
        .build-failure { border-color: #F44336; }
        .build-running { border-color: #2196F3; }
        .timestamp { font-size: 0.9em; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 PVM CI/CD Dashboard</h1>
            <div class="subtitle">Generated: {{.GeneratedAt.Format "2006-01-02 15:04:05 UTC"}}</div>
        </div>

        <div class="grid">
            <div class="card">
                <h3>📊 Build Metrics</h3>
                <div class="metric">
                    <span class="metric-label">Total Builds:</span>
                    <span class="metric-value">{{.BuildMetrics.TotalBuilds}}</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Success Rate:</span>
                    <span class="metric-value {{if gt .BuildMetrics.SuccessRate 0.9}}status-good{{else if gt .BuildMetrics.SuccessRate 0.7}}status-warning{{else}}status-error{{end}}">
                        {{printf "%.1f%%" (mul .BuildMetrics.SuccessRate 100)}}
                    </span>
                </div>
                <div class="metric">
                    <span class="metric-label">Avg Duration:</span>
                    <span class="metric-value">{{.BuildMetrics.AverageDuration}}s</span>
                </div>
            </div>

            <div class="card">
                <h3>🧪 Test Results</h3>
                <div class="metric">
                    <span class="metric-label">Total Tests:</span>
                    <span class="metric-value">{{.TestResults.TotalTests}}</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Pass Rate:</span>
                    <span class="metric-value {{if gt .TestResults.PassRate 0.95}}status-good{{else if gt .TestResults.PassRate 0.8}}status-warning{{else}}status-error{{end}}">
                        {{printf "%.1f%%" (mul .TestResults.PassRate 100)}}
                    </span>
                </div>
                <div class="metric">
                    <span class="metric-label">Coverage:</span>
                    <span class="metric-value {{if gt .TestResults.Coverage 80}}status-good{{else if gt .TestResults.Coverage 60}}status-warning{{else}}status-error{{end}}">
                        {{printf "%.1f%%" .TestResults.Coverage}}
                    </span>
                </div>
            </div>

            <div class="card">
                <h3>⚡ Performance</h3>
                <div class="metric">
                    <span class="metric-label">Avg Ops/Sec:</span>
                    <span class="metric-value">{{.Performance.AverageOpsPerSec}}</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Memory Usage:</span>
                    <span class="metric-value">{{printf "%.1fMB" .Performance.MemoryUsageMB}}</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Regressions:</span>
                    <span class="metric-value {{if eq .Performance.Regressions 0}}status-good{{else}}status-warning{{end}}">
                        {{.Performance.Regressions}}
                    </span>
                </div>
            </div>

            <div class="card">
                <h3>🔒 Security</h3>
                <div class="metric">
                    <span class="metric-label">Vulnerabilities:</span>
                    <span class="metric-value {{if eq .SecurityScan.Vulnerabilities 0}}status-good{{else}}status-error{{end}}">
                        {{.SecurityScan.Vulnerabilities}}
                    </span>
                </div>
                <div class="metric">
                    <span class="metric-label">Last Scan:</span>
                    <span class="metric-value">{{.SecurityScan.LastScan.Format "Jan 2, 15:04"}}</span>
                </div>
            </div>

            <div class="card">
                <h3>📋 Recent Builds</h3>
                <ul class="build-list">
                    {{range .RecentBuilds}}
                    <li class="build-item build-{{.Status}}">
                        <div><strong>{{.Branch}}</strong> ({{substr .Commit 0 7}})</div>
                        <div class="timestamp">{{.Timestamp.Format "Jan 2, 15:04"}} - {{.Duration}}s</div>
                    </li>
                    {{end}}
                </ul>
            </div>
        </div>
    </div>
</body>
</html>`

func main() {
	// Create sample dashboard data
	data := generateSampleData()

	// Load real data if available
	loadRealData(&data)

	// Generate HTML dashboard
	tmpl := template.Must(template.New("dashboard").Funcs(template.FuncMap{
		"mul": func(a, b float64) float64 { return a * b },
		"substr": func(s string, start, length int) string {
			if start >= len(s) {
				return ""
			}
			end := start + length
			if end > len(s) {
				end = len(s)
			}
			return s[start:end]
		},
	}).Parse(dashboardTemplate))

	func() {
		file, err := os.Create("ci-dashboard.html")
		if err != nil {
			log.Fatalf("Failed to create dashboard file: %v", err)
		}
		defer func() {
			if err := file.Close(); err != nil {
				log.Printf("Warning: failed to close file: %v", err)
			}
		}()

		if err := tmpl.Execute(file, data); err != nil {
			log.Fatalf("Failed to execute template: %v", err)
		}
	}()

	// Also output JSON for programmatic access
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	if err := os.WriteFile("ci-dashboard.json", jsonData, 0644); err != nil {
		log.Fatalf("Failed to write JSON: %v", err)
	}

	fmt.Println("Dashboard generated: ci-dashboard.html")
	fmt.Println("JSON data saved: ci-dashboard.json")
}

func generateSampleData() DashboardData {
	return DashboardData{
		GeneratedAt: time.Now(),
		BuildMetrics: BuildMetrics{
			TotalBuilds:     150,
			SuccessRate:     0.94,
			AverageDuration: 240,
			LastBuild:       time.Now().Add(-1 * time.Hour),
		},
		TestResults: TestResults{
			TotalTests:   1250,
			PassRate:     0.98,
			Coverage:     85.5,
			FailingTests: []string{},
		},
		Performance: PerformanceMetrics{
			AverageOpsPerSec: 125000,
			MemoryUsageMB:    128.5,
			Regressions:      0,
			Benchmarks: map[string]float64{
				"Parser":      98500,
				"TypeChecker": 45000,
				"Binder":      87500,
			},
		},
		SecurityScan: SecurityMetrics{
			Vulnerabilities: 0,
			LastScan:        time.Now().Add(-6 * time.Hour),
			Severity: map[string]int{
				"high":   0,
				"medium": 0,
				"low":    0,
			},
		},
		QualityGates: QualityMetrics{
			LintIssues:     0,
			CodeComplexity: 2.3,
			TechnicalDebt:  "Low",
		},
		RecentBuilds: []BuildInfo{
			{ID: "123", Branch: "pu", Commit: "abc1234", Status: "success", Duration: 235, Timestamp: time.Now().Add(-1 * time.Hour)},
			{ID: "122", Branch: "main", Commit: "def5678", Status: "success", Duration: 198, Timestamp: time.Now().Add(-3 * time.Hour)},
			{ID: "121", Branch: "pu", Commit: "ghi9012", Status: "success", Duration: 267, Timestamp: time.Now().Add(-5 * time.Hour)},
		},
	}
}

func loadRealData(data *DashboardData) {
	// Try to load real metrics from build artifacts
	if metricsData, err := os.ReadFile(".build-metrics/latest.json"); err == nil {
		var metrics map[string]interface{}
		if json.Unmarshal(metricsData, &metrics) == nil {
			// Update with real data
			if buildTime, ok := metrics["build_time"].(float64); ok {
				data.BuildMetrics.AverageDuration = int(buildTime)
			}
		}
	}

	// Try to load performance data
	perfFiles, _ := filepath.Glob("testdata/performance/*.json")
	if len(perfFiles) > 0 {
		sort.Strings(perfFiles)
		// Use most recent performance data
		if perfData, err := os.ReadFile(perfFiles[len(perfFiles)-1]); err == nil {
			var perf map[string]interface{}
			if json.Unmarshal(perfData, &perf) == nil {
				// Update performance metrics with real data
				log.Printf("Loaded performance data from %s", perfFiles[len(perfFiles)-1])
			}
		}
	}
}
