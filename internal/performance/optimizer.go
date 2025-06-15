// ABOUTME: Performance optimization system for PVM operations
// ABOUTME: Provides automated performance tuning and bottleneck detection

package performance

import (
	"fmt"
	"runtime"
	"sort"
	"time"
)

// Optimizer provides automated performance optimization
type Optimizer struct {
	analyzer    *Analyzer
	enabled     bool
	thresholds  *Thresholds
	suggestions []Suggestion
}

// Thresholds define performance targets
type Thresholds struct {
	MaxParseTime    time.Duration // Maximum acceptable parse time
	MaxMemoryUsage  int64         // Maximum memory usage in bytes
	MaxCacheSize    int           // Maximum cache entries
	MinCacheHitRate float64       // Minimum cache hit rate
}

// Suggestion represents an optimization recommendation
type Suggestion struct {
	Category    string
	Description string
	Impact      ImpactLevel
	Action      string
	Metric      string
	Current     interface{}
	Target      interface{}
}

// ImpactLevel represents the potential impact of an optimization
type ImpactLevel int

const (
	LowImpact ImpactLevel = iota
	MediumImpact
	HighImpact
	CriticalImpact
)

func (i ImpactLevel) String() string {
	switch i {
	case LowImpact:
		return "Low"
	case MediumImpact:
		return "Medium"
	case HighImpact:
		return "High"
	case CriticalImpact:
		return "Critical"
	default:
		return "Unknown"
	}
}

// NewOptimizer creates a new performance optimizer
func NewOptimizer() *Optimizer {
	return &Optimizer{
		analyzer: NewAnalyzer(),
		enabled:  true,
		thresholds: &Thresholds{
			MaxParseTime:    500 * time.Millisecond,
			MaxMemoryUsage:  100 * 1024 * 1024, // 100MB
			MaxCacheSize:    1000,
			MinCacheHitRate: 0.8, // 80%
		},
		suggestions: make([]Suggestion, 0),
	}
}

// AnalyzePerformance analyzes current performance and generates suggestions
func (o *Optimizer) AnalyzePerformance() []Suggestion {
	if !o.enabled {
		return nil
	}

	o.suggestions = o.suggestions[:0] // Reset suggestions

	// Analyze timing metrics
	o.analyzeTimingMetrics()

	// Analyze memory usage
	o.analyzeMemoryUsage()

	// Analyze cache performance
	o.analyzeCachePerformance()

	// Sort suggestions by impact
	sort.Slice(o.suggestions, func(i, j int) bool {
		return o.suggestions[i].Impact > o.suggestions[j].Impact
	})

	return o.suggestions
}

// analyzeTimingMetrics checks for slow operations
func (o *Optimizer) analyzeTimingMetrics() {
	metrics := o.analyzer.GetMetrics()
	
	for name, metric := range metrics {
		// Check for slow average times
		if metric.AvgTime > o.thresholds.MaxParseTime {
			o.addSuggestion(Suggestion{
				Category:    "Performance",
				Description: fmt.Sprintf("Operation '%s' is slower than threshold", name),
				Impact:      o.calculateImpact(metric.AvgTime, o.thresholds.MaxParseTime),
				Action:      "Consider optimizing or caching this operation",
				Metric:      "Average Time",
				Current:     metric.AvgTime,
				Target:      o.thresholds.MaxParseTime,
			})
		}

		// Check for high variance (max >> avg)
		if metric.MaxTime > metric.AvgTime*3 {
			o.addSuggestion(Suggestion{
				Category:    "Performance",
				Description: fmt.Sprintf("Operation '%s' has high time variance", name),
				Impact:      MediumImpact,
				Action:      "Investigate outlier cases causing slow performance",
				Metric:      "Time Variance",
				Current:     fmt.Sprintf("Max: %v, Avg: %v", metric.MaxTime, metric.AvgTime),
				Target:      "More consistent timing",
			})
		}
	}
}

// analyzeMemoryUsage checks memory consumption
func (o *Optimizer) analyzeMemoryUsage() {
	memStats := o.analyzer.GetMemoryUsage()
	
	if int64(memStats.Alloc) > o.thresholds.MaxMemoryUsage {
		o.addSuggestion(Suggestion{
			Category:    "Memory",
			Description: "High memory usage detected",
			Impact:      HighImpact,
			Action:      "Consider reducing memory allocation or running garbage collection",
			Metric:      "Memory Allocation",
			Current:     fmt.Sprintf("%d bytes", memStats.Alloc),
			Target:      fmt.Sprintf("<%d bytes", o.thresholds.MaxMemoryUsage),
		})
	}

	// Check for potential memory leaks (high heap objects)
	if memStats.HeapObjects > 100000 {
		o.addSuggestion(Suggestion{
			Category:    "Memory",
			Description: "High number of heap objects",
			Impact:      MediumImpact,
			Action:      "Check for memory leaks or object pooling opportunities",
			Metric:      "Heap Objects",
			Current:     memStats.HeapObjects,
			Target:      "Fewer objects",
		})
	}
}

// analyzeCachePerformance checks cache efficiency
func (o *Optimizer) analyzeCachePerformance() {
	// Check global caches
	caches := []*Cache{ParserCache, TypeCache, FileCache}
	cacheNames := []string{"Parser", "Type", "File"}

	for i, cache := range caches {
		stats := cache.Stats()
		
		// Check hit rate
		if stats.HitRatio < o.thresholds.MinCacheHitRate && stats.HitCount+stats.MissCount > 10 {
			o.addSuggestion(Suggestion{
				Category:    "Caching",
				Description: fmt.Sprintf("%s cache has low hit rate", cacheNames[i]),
				Impact:      HighImpact,
				Action:      "Increase cache size or improve cache key strategy",
				Metric:      "Hit Rate",
				Current:     fmt.Sprintf("%.1f%%", stats.HitRatio*100),
				Target:      fmt.Sprintf("%.1f%%", o.thresholds.MinCacheHitRate*100),
			})
		}

		// Check cache size utilization
		if stats.Items > o.thresholds.MaxCacheSize {
			o.addSuggestion(Suggestion{
				Category:    "Caching",
				Description: fmt.Sprintf("%s cache is near capacity", cacheNames[i]),
				Impact:      MediumImpact,
				Action:      "Consider increasing cache size or improving eviction policy",
				Metric:      "Cache Size",
				Current:     stats.Items,
				Target:      fmt.Sprintf("<%d items", o.thresholds.MaxCacheSize),
			})
		}
	}
}

// calculateImpact determines the impact level based on how much a value exceeds threshold
func (o *Optimizer) calculateImpact(current, threshold time.Duration) ImpactLevel {
	ratio := float64(current) / float64(threshold)
	
	switch {
	case ratio > 5:
		return CriticalImpact
	case ratio > 3:
		return HighImpact
	case ratio > 2:
		return MediumImpact
	default:
		return LowImpact
	}
}

// addSuggestion adds a new optimization suggestion
func (o *Optimizer) addSuggestion(s Suggestion) {
	o.suggestions = append(o.suggestions, s)
}

// GetAnalyzer returns the underlying performance analyzer
func (o *Optimizer) GetAnalyzer() *Analyzer {
	return o.analyzer
}

// SetThresholds updates performance thresholds
func (o *Optimizer) SetThresholds(t *Thresholds) {
	o.thresholds = t
}

// GetThresholds returns current performance thresholds
func (o *Optimizer) GetThresholds() *Thresholds {
	return o.thresholds
}

// Enable/Disable the optimizer
func (o *Optimizer) SetEnabled(enabled bool) {
	o.enabled = enabled
	o.analyzer.SetEnabled(enabled)
}

// IsEnabled returns whether the optimizer is enabled
func (o *Optimizer) IsEnabled() bool {
	return o.enabled
}

// GetPerformanceReport generates a comprehensive performance report
func (o *Optimizer) GetPerformanceReport() *PerformanceReport {
	suggestions := o.AnalyzePerformance()
	metrics := o.analyzer.GetMetrics()
	memStats := o.analyzer.GetMemoryUsage()

	return &PerformanceReport{
		Timestamp:   time.Now(),
		Metrics:     metrics,
		MemoryStats: memStats,
		Suggestions: suggestions,
		Summary:     o.generateSummary(suggestions, metrics),
	}
}

// PerformanceReport contains comprehensive performance analysis
type PerformanceReport struct {
	Timestamp   time.Time
	Metrics     map[string]*Metric
	MemoryStats runtime.MemStats
	Suggestions []Suggestion
	Summary     ReportSummary
}

// ReportSummary provides a high-level performance overview
type ReportSummary struct {
	TotalOperations    int64
	TotalTime          time.Duration
	AverageTime        time.Duration
	SlowestOperation   string
	MemoryUsage        uint64
	GoroutineCount     int
	CriticalIssues     int
	HighPriorityIssues int
}

// generateSummary creates a performance summary
func (o *Optimizer) generateSummary(suggestions []Suggestion, metrics map[string]*Metric) ReportSummary {
	var totalOps int64
	var totalTime time.Duration
	var slowestOp string
	var slowestTime time.Duration

	for name, metric := range metrics {
		totalOps += metric.Count
		totalTime += metric.TotalTime
		
		if metric.AvgTime > slowestTime {
			slowestTime = metric.AvgTime
			slowestOp = name
		}
	}

	var avgTime time.Duration
	if totalOps > 0 {
		avgTime = totalTime / time.Duration(totalOps)
	}

	// Count issues by severity
	var critical, high int
	for _, s := range suggestions {
		switch s.Impact {
		case CriticalImpact:
			critical++
		case HighImpact:
			high++
		}
	}

	memStats := o.analyzer.GetMemoryUsage()

	return ReportSummary{
		TotalOperations:    totalOps,
		TotalTime:          totalTime,
		AverageTime:        avgTime,
		SlowestOperation:   slowestOp,
		MemoryUsage:        memStats.Alloc,
		GoroutineCount:     runtime.NumGoroutine(),
		CriticalIssues:     critical,
		HighPriorityIssues: high,
	}
}

// OptimizeAutomatically applies automatic optimizations
func (o *Optimizer) OptimizeAutomatically() []string {
	var applied []string
	
	suggestions := o.AnalyzePerformance()
	
	for _, suggestion := range suggestions {
		switch suggestion.Category {
		case "Memory":
			if suggestion.Description == "High memory usage detected" {
				// Force garbage collection
				runtime.GC()
				applied = append(applied, "Performed garbage collection")
			}
		case "Caching":
			if suggestion.Metric == "Cache Size" {
				// Could automatically resize caches here
				applied = append(applied, "Cache optimization analyzed")
			}
		}
	}
	
	return applied
}

// Global optimizer instance
var globalOptimizer = NewOptimizer()

// AnalyzeGlobalPerformance analyzes performance using global optimizer
func AnalyzeGlobalPerformance() []Suggestion {
	return globalOptimizer.AnalyzePerformance()
}

// GetGlobalPerformanceReport gets comprehensive performance report
func GetGlobalPerformanceReport() *PerformanceReport {
	return globalOptimizer.GetPerformanceReport()
}

// OptimizeGlobalPerformance applies automatic optimizations
func OptimizeGlobalPerformance() []string {
	return globalOptimizer.OptimizeAutomatically()
}