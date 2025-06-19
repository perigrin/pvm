// ABOUTME: Test execution command for project-aware Perl test running
// ABOUTME: Integrates with project detection and build system for comprehensive testing

package pvm

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/cli/ui"
	"tamarou.com/pvm/internal/project"
	"tamarou.com/pvm/internal/pvx"
)

// SingleTestResult represents the result of executing a single test
type SingleTestResult struct {
	File      string
	Passed    bool
	Failed    bool
	Skipped   bool
	TestCount int
	Duration  time.Duration
	Output    string
	Error     string
}

// TestSummary represents the overall test execution summary
type TestSummary struct {
	TotalTests   int
	PassedTests  int
	FailedTests  int
	SkippedTests int
	TotalTime    time.Duration
	Results      []*SingleTestResult
}

// TestRunner handles test discovery and execution
type TestRunner struct {
	ProjectCtx  *project.ProjectContext
	PerlVersion string
	Verbose     bool
	Parallel    bool
	Coverage    bool
	TestPattern string
}

// newTestCommand creates the test command
func newTestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test [pattern]",
		Short: "Run project tests",
		Long:  "Run tests with project-aware environment setup and proper module paths",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui := cli.GetUI(cmd)

			// Get flags
			verbose, _ := cmd.Flags().GetBool("verbose")
			parallel, _ := cmd.Flags().GetBool("parallel")
			coverage, _ := cmd.Flags().GetBool("coverage")
			perlVersion, _ := cmd.Flags().GetString("perl")
			testDir, _ := cmd.Flags().GetString("test-dir")

			// Get test pattern from args
			var testPattern string
			if len(args) > 0 {
				testPattern = args[0]
			}

			// Detect project context
			wd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}

			projectCtx, err := project.DetectProject(wd)
			if err != nil {
				return fmt.Errorf("failed to detect project context: %w", err)
			}
			if !projectCtx.IsProject {
				ui.Warning("Not in a project directory. Tests will run without project-specific configuration.")
			}

			// Create test runner
			runner := &TestRunner{
				ProjectCtx:  projectCtx,
				PerlVersion: perlVersion,
				Verbose:     verbose,
				Parallel:    parallel,
				Coverage:    coverage,
				TestPattern: testPattern,
			}

			// Discover tests
			testFiles, err := runner.DiscoverTests(testDir)
			if err != nil {
				return fmt.Errorf("failed to discover tests: %w", err)
			}

			if len(testFiles) == 0 {
				ui.Info("No tests found.")
				if projectCtx.IsProject {
					ui.Info("Looked in: %s", filepath.Join(projectCtx.RootDir, "t"))
				} else {
					ui.Info("Looked in: ./t")
				}
				return nil
			}

			// Execute tests
			ui.Status(fmt.Sprintf("Running %d test(s)...", len(testFiles)))
			summary, err := runner.ExecuteTests(testFiles)
			if err != nil {
				return fmt.Errorf("test execution failed: %w", err)
			}

			// Report results
			runner.ReportResults(ui, summary)

			// Exit with appropriate code
			if summary.FailedTests > 0 {
				os.Exit(1)
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().BoolP("verbose", "v", false, "Verbose test output")
	cmd.Flags().Bool("parallel", false, "Run tests in parallel")
	cmd.Flags().Bool("coverage", false, "Generate test coverage report")
	cmd.Flags().String("perl", "", "Perl version to use (default: project version or system)")
	cmd.Flags().String("test-dir", "", "Test directory (default: t/)")

	return cmd
}

// DiscoverTests finds all test files in the project
func (tr *TestRunner) DiscoverTests(testDir string) ([]string, error) {
	var baseDir string

	switch {
	case testDir != "":
		baseDir = testDir
	case tr.ProjectCtx.IsProject:
		baseDir = filepath.Join(tr.ProjectCtx.RootDir, "t")
	default:
		baseDir = "t"
	}

	// Check if test directory exists
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("test directory does not exist: %s", baseDir)
	}

	var testFiles []string

	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Look for .t files and .pl files that look like tests
		if strings.HasSuffix(path, ".t") ||
			(strings.HasSuffix(path, ".pl") &&
				(strings.Contains(strings.ToLower(path), "test") || strings.Contains(strings.ToLower(path), "spec"))) {

			// Apply pattern filter if specified
			if tr.TestPattern != "" {
				matched, err := regexp.MatchString(tr.TestPattern, filepath.Base(path))
				if err != nil {
					return fmt.Errorf("invalid test pattern: %w", err)
				}
				if !matched {
					return nil
				}
			}

			testFiles = append(testFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk test directory: %w", err)
	}

	return testFiles, nil
}

// ExecuteTests runs all discovered tests
func (tr *TestRunner) ExecuteTests(testFiles []string) (*TestSummary, error) {
	summary := &TestSummary{
		Results: make([]*SingleTestResult, 0, len(testFiles)),
	}

	startTime := time.Now()

	for _, testFile := range testFiles {
		result, err := tr.ExecuteSingleTest(testFile)
		if err != nil {
			// Create a failed result for this test
			result = &SingleTestResult{
				File:   testFile,
				Passed: false,
				Failed: true,
				Error:  err.Error(),
			}
		}

		summary.Results = append(summary.Results, result)
		summary.TotalTests++

		switch {
		case result.Passed:
			summary.PassedTests++
		case result.Failed:
			summary.FailedTests++
		case result.Skipped:
			summary.SkippedTests++
		}
	}

	summary.TotalTime = time.Since(startTime)
	return summary, nil
}

// ExecuteSingleTest runs a single test file
func (tr *TestRunner) ExecuteSingleTest(testFile string) (*SingleTestResult, error) {
	startTime := time.Now()

	// Determine Perl version to use
	perlVersion := tr.PerlVersion
	if perlVersion == "" && tr.ProjectCtx.IsProject && tr.ProjectCtx.PerlVersion != "" {
		perlVersion = tr.ProjectCtx.PerlVersion
	}

	// Set up execution options for project-aware test execution
	options := &pvx.ExecutionOptions{
		PerlVersion:    perlVersion,
		IsolationLevel: pvx.IsolationLow, // Use low isolation for tests
		ScriptPath:     testFile,
		Verbose:        tr.Verbose,
	}

	// If we're in a project, set up the local lib environment
	if tr.ProjectCtx.IsProject {
		options.IsolationDir = tr.ProjectCtx.LocalLibDir

		// Add project lib directory to additional module paths
		if tr.ProjectCtx.LocalLibDir != "" {
			libPath := filepath.Join(tr.ProjectCtx.RootDir, "lib")
			if _, err := os.Stat(libPath); err == nil {
				options.AdditionalModulePaths = []string{libPath}
			}
		}
	}

	// Execute the test using PVX
	output, err := pvx.ExecuteScript(options)
	duration := time.Since(startTime)

	testResult := &SingleTestResult{
		File:     testFile,
		Duration: duration,
		Output:   output,
	}

	if err != nil {
		testResult.Failed = true
		testResult.Error = err.Error()
		return testResult, nil
	}

	// For now, assume success if no error (could enhance with TAP parsing)
	testResult.Passed = true

	// Try to extract test count from output
	testResult.TestCount = tr.parseTestCount(output)

	return testResult, nil
}

// parseTestCount attempts to extract the number of tests from TAP output
func (tr *TestRunner) parseTestCount(output string) int {
	// Look for "1..N" pattern in TAP output
	re := regexp.MustCompile(`^1\.\.(\d+)`)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if matches := re.FindStringSubmatch(strings.TrimSpace(line)); matches != nil {
			if len(matches) > 1 {
				var count int
				fmt.Sscanf(matches[1], "%d", &count)
				return count
			}
		}
	}

	// Fallback: count "ok" and "not ok" lines
	lines = strings.Split(output, "\n")
	count := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "ok ") || strings.HasPrefix(trimmed, "not ok ") {
			count++
		}
	}
	return count
}

// ReportResults prints test execution results
func (tr *TestRunner) ReportResults(output *ui.Output, summary *TestSummary) {
	output.Println("")
	output.Header("TEST RESULTS")

	// Show individual test results if verbose or if there are failures
	if tr.Verbose || summary.FailedTests > 0 {
		for _, result := range summary.Results {
			var status string
			resultLine := fmt.Sprintf("%s (%s)", result.File, result.Duration.Round(time.Millisecond))

			switch {
			case result.Failed:
				status = "FAIL"
			case result.Skipped:
				status = "SKIP"
			default:
				status = "PASS"
			}
			output.Printf("%-8s %s\n", status, resultLine)

			if result.Failed && result.Error != "" {
				output.Printf("         Error: %s\n", result.Error)
			}

			if tr.Verbose && result.Output != "" {
				// Show first few lines of output
				lines := strings.Split(result.Output, "\n")
				maxLines := 5
				if len(lines) > maxLines {
					for i := 0; i < maxLines; i++ {
						output.Printf("         %s\n", lines[i])
					}
					output.Printf("         ... (%d more lines)\n", len(lines)-maxLines)
				} else {
					for _, line := range lines {
						if strings.TrimSpace(line) != "" {
							output.Printf("         %s\n", line)
						}
					}
				}
			}
		}
		output.Println("")
	}

	// Summary statistics
	output.Info("Tests: %d total, %d passed, %d failed, %d skipped",
		summary.TotalTests, summary.PassedTests, summary.FailedTests, summary.SkippedTests)
	output.Info("Time:  %s", summary.TotalTime.Round(time.Millisecond))

	// Show project context information
	if tr.ProjectCtx.IsProject {
		output.Info("Project: %s", tr.ProjectCtx.RootDir)
		if tr.ProjectCtx.PerlVersion != "" {
			output.Info("Perl version: %s", tr.ProjectCtx.PerlVersion)
		}
	}

	// Final result
	output.Println("")
	switch {
	case summary.FailedTests > 0:
		output.Error("Result: FAILED")
	case summary.PassedTests > 0:
		output.Success("Result: PASSED")
	default:
		output.Warning("Result: NO TESTS RUN")
	}
}
