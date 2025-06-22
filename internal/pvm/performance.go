// ABOUTME: Performance analysis and optimization commands for PVM
// ABOUTME: Provides CLI interface for performance monitoring and tuning

package pvm

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/performance"
)

// createPerformanceCommand creates the performance analysis command
func createPerformanceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "perf",
		Short: "Performance analysis and optimization",
		Long:  "Analyze PVM performance, identify bottlenecks, and apply optimizations",
	}

	cmd.AddCommand(createPerformanceAnalyzeCommand())
	cmd.AddCommand(createPerformanceReportCommand())
	cmd.AddCommand(createPerformanceOptimizeCommand())
	cmd.AddCommand(createPerformanceResetCommand())

	return cmd
}

// createPerformanceAnalyzeCommand creates the analyze subcommand
func createPerformanceAnalyzeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "analyze",
		Short: "Analyze current performance metrics",
		Long:  "Analyze performance metrics and provide optimization suggestions",
		RunE: func(cmd *cobra.Command, args []string) error {
			suggestions := performance.AnalyzeGlobalPerformance()

			ui := cli.GetUI(cmd)
			if len(suggestions) == 0 {
				ui.Success("No performance issues detected")
				return nil
			}

			ui.Info("Found %d performance suggestions:", len(suggestions))
			ui.Println()

			for i, suggestion := range suggestions {
				ui.Printf("%d. [%s] %s\n", i+1, suggestion.Impact, suggestion.Description)
				ui.Printf("   Category: %s\n", suggestion.Category)
				ui.Printf("   Action: %s\n", suggestion.Action)
				if suggestion.Current != nil {
					ui.Printf("   Current: %v\n", suggestion.Current)
				}
				if suggestion.Target != nil {
					ui.Printf("   Target: %v\n", suggestion.Target)
				}
				ui.Println()
			}

			return nil
		},
	}
}

// createPerformanceReportCommand creates the report subcommand
func createPerformanceReportCommand() *cobra.Command {
	var outputFormat string
	var outputFile string

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate comprehensive performance report",
		Long:  "Generate detailed performance report with metrics and analysis",
		RunE: func(cmd *cobra.Command, args []string) error {
			report := performance.GetGlobalPerformanceReport()

			switch outputFormat {
			case "json":
				return outputJSONReport(report, outputFile)
			default:
				return outputTextReport(report, outputFile)
			}
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "format", "f", "text", "Output format (text|json)")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")

	return cmd
}

// createPerformanceOptimizeCommand creates the optimize subcommand
func createPerformanceOptimizeCommand() *cobra.Command {
	var autoApply bool

	cmd := &cobra.Command{
		Use:   "optimize",
		Short: "Apply performance optimizations",
		Long:  "Apply automatic performance optimizations",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui := cli.GetUI(cmd)
			if !autoApply {
				ui.Warning("Use --auto to apply automatic optimizations")
				ui.Info("Run 'pvm perf analyze' to see suggestions first")
				return nil
			}

			applied := performance.OptimizeGlobalPerformance()

			if len(applied) == 0 {
				ui.Success("No automatic optimizations available")
				return nil
			}

			ui.Success("Applied %d optimizations:", len(applied))
			for _, optimization := range applied {
				ui.Printf("  • %s\n", optimization)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&autoApply, "auto", false, "Automatically apply optimizations")

	return cmd
}

// createPerformanceResetCommand creates the reset subcommand
func createPerformanceResetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Reset performance metrics",
		Long:  "Clear all collected performance metrics",
		RunE: func(cmd *cobra.Command, args []string) error {
			performance.ResetGlobalMetrics()
			ui := cli.GetUI(cmd)
			ui.Success("Performance metrics reset")
			return nil
		},
	}
}

// outputTextReport outputs a human-readable text report
func outputTextReport(report *performance.PerformanceReport, outputFile string) error {
	var output *os.File
	var err error

	if outputFile != "" {
		output, err = os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer output.Close()
	} else {
		output = os.Stdout
	}

	fmt.Fprintf(output, "PVM Performance Report\n")
	fmt.Fprintf(output, "Generated: %s\n\n", report.Timestamp.Format(time.RFC3339))

	// Summary
	fmt.Fprintf(output, "📊 SUMMARY\n")
	fmt.Fprintf(output, "─────────────────────────────────────────\n")
	fmt.Fprintf(output, "Total Operations: %d\n", report.Summary.TotalOperations)
	fmt.Fprintf(output, "Total Time: %v\n", report.Summary.TotalTime)
	fmt.Fprintf(output, "Average Time: %v\n", report.Summary.AverageTime)
	fmt.Fprintf(output, "Slowest Operation: %s\n", report.Summary.SlowestOperation)
	fmt.Fprintf(output, "Memory Usage: %d bytes\n", report.Summary.MemoryUsage)
	fmt.Fprintf(output, "Goroutines: %d\n", report.Summary.GoroutineCount)
	fmt.Fprintf(output, "Critical Issues: %d\n", report.Summary.CriticalIssues)
	fmt.Fprintf(output, "High Priority Issues: %d\n\n", report.Summary.HighPriorityIssues)

	// Metrics
	if len(report.Metrics) > 0 {
		fmt.Fprintf(output, "📈 METRICS\n")
		fmt.Fprintf(output, "─────────────────────────────────────────\n")
		for name, metric := range report.Metrics {
			fmt.Fprintf(output, "%s:\n", name)
			fmt.Fprintf(output, "  Count: %d\n", metric.Count)
			fmt.Fprintf(output, "  Total Time: %v\n", metric.TotalTime)
			fmt.Fprintf(output, "  Average Time: %v\n", metric.AvgTime)
			fmt.Fprintf(output, "  Min Time: %v\n", metric.MinTime)
			fmt.Fprintf(output, "  Max Time: %v\n", metric.MaxTime)
			fmt.Fprintf(output, "\n")
		}
	}

	// Suggestions
	if len(report.Suggestions) > 0 {
		fmt.Fprintf(output, "💡 OPTIMIZATION SUGGESTIONS\n")
		fmt.Fprintf(output, "─────────────────────────────────────────\n")
		for i, suggestion := range report.Suggestions {
			fmt.Fprintf(output, "%d. [%s] %s\n", i+1, suggestion.Impact, suggestion.Description)
			fmt.Fprintf(output, "   Category: %s\n", suggestion.Category)
			fmt.Fprintf(output, "   Action: %s\n", suggestion.Action)
			if suggestion.Current != nil {
				fmt.Fprintf(output, "   Current: %v\n", suggestion.Current)
			}
			if suggestion.Target != nil {
				fmt.Fprintf(output, "   Target: %v\n", suggestion.Target)
			}
			fmt.Fprintf(output, "\n")
		}
	}

	return nil
}

// outputJSONReport outputs a JSON report
func outputJSONReport(report *performance.PerformanceReport, outputFile string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	if outputFile != "" {
		err = os.WriteFile(outputFile, data, 0644)
		if err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
	} else {
		fmt.Println(string(data))
	}

	return nil
}
