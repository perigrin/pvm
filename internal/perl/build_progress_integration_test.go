// ABOUTME: Integration tests for BuildProgressCalculator with build stages
// ABOUTME: Tests progress calculation behavior in realistic build scenarios

package perl

import (
	"context"
	"testing"
	"time"
)

func TestBuildProgressIntegration(t *testing.T) {
	// Create a mock build options with progress callback
	var progressUpdates []struct {
		stage    BuildProgressStage
		details  string
		progress float64
	}

	options := &BuildOptions{
		Version:   "5.38.0",
		RunTests:  true,
		BuildOnly: false,
		Context:   context.Background(),
		ProgressCallback: func(stage BuildProgressStage, details string, progress float64) {
			progressUpdates = append(progressUpdates, struct {
				stage    BuildProgressStage
				details  string
				progress float64
			}{stage, details, progress})
		},
	}

	// Create progress calculator
	calc := NewBuildProgressCalculator()

	// Simulate the build process stages
	stages := []struct {
		stage    BuildProgressStage
		details  string
		progress float64
	}{
		{StageExtract, "Extracting source archive", 0.0},
		{StageExtract, "Extraction complete", 1.0},
		{StageConfigure, "Running Configure script", 0.0},
		{StageConfigure, "Configure complete", 1.0},
		{StageCompile, "Compiling Perl", 0.0},
		{StageCompile, "CC main.c", 0.2},
		{StageCompile, "CC perl.c", 0.4},
		{StageCompile, "CC utils.c", 0.6},
		{StageCompile, "LD perl", 0.9},
		{StageCompile, "Compilation complete", 1.0},
		{StageTest, "Running tests", 0.0},
		{StageTest, "Test complete", 1.0},
		{StageInstall, "Installing Perl", 0.0},
		{StageInstall, "Installation complete", 1.0},
	}

	// Process each stage
	var previousOverallProgress float64
	for _, stage := range stages {
		calc.SetStage(stage.stage)
		calc.UpdateStageProgress(stage.progress)

		overallProgress := calc.GetOverallProgress()

		// Call the progress callback
		if options.ProgressCallback != nil {
			options.ProgressCallback(stage.stage, stage.details, overallProgress)
		}

		// Verify progress doesn't go backwards
		if overallProgress < previousOverallProgress {
			t.Errorf("Overall progress went backwards: %.3f -> %.3f", previousOverallProgress, overallProgress)
		}

		// Verify progress is within bounds
		if overallProgress < 0.0 || overallProgress > 1.0 {
			t.Errorf("Overall progress %.3f is out of bounds [0.0, 1.0]", overallProgress)
		}

		previousOverallProgress = overallProgress
	}

	// Verify final progress is close to 100%
	finalProgress := calc.GetOverallProgress()
	if finalProgress < 0.95 {
		t.Errorf("Final progress %.3f should be at least 95%%", finalProgress)
	}

	// Verify we got progress callbacks
	if len(progressUpdates) == 0 {
		t.Error("Expected progress callbacks to be called")
	}

	// Verify progress updates are non-decreasing
	for i := 1; i < len(progressUpdates); i++ {
		if progressUpdates[i].progress < progressUpdates[i-1].progress {
			t.Errorf("Progress update %d went backwards: %.3f -> %.3f",
				i, progressUpdates[i-1].progress, progressUpdates[i].progress)
		}
	}
}

func TestBuildProgressWithStuckDetection(t *testing.T) {
	calc := NewBuildProgressCalculator()

	// Test stuck progress at configure stage
	calc.SetStage(StageConfigure)

	// Simulate stuck progress
	for i := 0; i < 5; i++ {
		calc.UpdateTimeBasedProgress()
		time.Sleep(1 * time.Millisecond) // Small delay to simulate time passing
	}

	progress := calc.GetOverallProgress()

	// Should have some progress due to time-based advancement
	if progress <= 0.0 {
		t.Errorf("Expected some progress due to time-based advancement, got %.3f", progress)
	}

	// Should not exceed the stage's maximum contribution
	expectedMax := calc.stageWeights[StageConfigure] // 15% for configure
	if progress > expectedMax {
		t.Errorf("Progress %.3f should not exceed stage maximum %.3f", progress, expectedMax)
	}
}

func TestBuildProgressMakeOutputParsing(t *testing.T) {
	calc := NewBuildProgressCalculator()
	calc.SetStage(StageCompile)

	// Test various make output patterns
	testCases := []struct {
		line     string
		expected bool
		desc     string
	}{
		{"CC main.c", true, "C compilation"},
		{"LD perl", true, "Linking"},
		{"[50%] Building target", true, "Percentage output"},
		{"[100%] Complete", true, "100% completion"},
		{"make: some error", false, "Error message"},
		{"random output", false, "Random text"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			initialProgress := calc.stageProgress
			result := calc.ParseMakeProgress(tc.line)

			if result != tc.expected {
				t.Errorf("ParseMakeProgress(%q) = %v, expected %v", tc.line, result, tc.expected)
			}

			if tc.expected && calc.stageProgress == initialProgress {
				t.Errorf("Expected progress to change for line %q", tc.line)
			}
		})
	}
}

func TestBuildProgressStageTransitions(t *testing.T) {
	calc := NewBuildProgressCalculator()

	// Test stage transitions
	stages := []BuildProgressStage{
		StageExtract,
		StageConfigure,
		StageCompile,
		StageTest,
		StageInstall,
	}

	totalWeight := 0.0
	for _, stage := range stages {
		calc.SetStage(stage)

		// Test progress at start of stage
		startProgress := calc.GetOverallProgress()

		// Complete the stage
		calc.UpdateStageProgress(1.0)
		endProgress := calc.GetOverallProgress()

		// Verify progress increased
		if endProgress <= startProgress {
			t.Errorf("Stage %v: progress should increase from %.3f to %.3f", stage, startProgress, endProgress)
		}

		// Add to total weight
		if weight, exists := calc.stageWeights[stage]; exists {
			totalWeight += weight
		}
	}

	// Verify total weight is reasonable (should be 1.0 excluding cleanup)
	if totalWeight < 0.95 || totalWeight > 1.05 {
		t.Errorf("Total stage weights %.3f should be close to 1.0", totalWeight)
	}
}

func TestBuildProgressBoundsChecking(t *testing.T) {
	calc := NewBuildProgressCalculator()

	// Test negative progress
	calc.UpdateStageProgress(-0.5)
	if calc.stageProgress != 0.0 {
		t.Errorf("Negative progress should be clamped to 0.0, got %.3f", calc.stageProgress)
	}

	// Test progress over 1.0
	calc.UpdateStageProgress(1.5)
	if calc.stageProgress != 1.0 {
		t.Errorf("Progress over 1.0 should be clamped to 1.0, got %.3f", calc.stageProgress)
	}

	// Test overall progress bounds
	calc.SetStage(StageCompile)
	calc.UpdateStageProgress(1.0)
	overallProgress := calc.GetOverallProgress()

	if overallProgress < 0.0 || overallProgress > 1.0 {
		t.Errorf("Overall progress %.3f should be within [0.0, 1.0]", overallProgress)
	}
}

func TestBuildProgressRealisticScenario(t *testing.T) {
	calc := NewBuildProgressCalculator()

	// Simulate a realistic build scenario
	calc.SetStage(StageExtract)
	calc.UpdateStageProgress(1.0) // Quick extraction

	calc.SetStage(StageConfigure)
	calc.UpdateTimeBasedProgress() // Time-based configure

	calc.SetStage(StageCompile)
	// Simulate parsing make output
	makeOutputLines := []string{
		"CC main.c",
		"CC perl.c",
		"CC utils.c",
		"[75%] Building target",
		"LD perl",
		"[100%] Complete",
	}

	for _, line := range makeOutputLines {
		calc.ParseMakeProgress(line)
	}

	calc.SetStage(StageTest)
	calc.UpdateStageProgress(1.0) // Complete tests

	calc.SetStage(StageInstall)
	calc.UpdateStageProgress(1.0) // Complete install

	finalProgress := calc.GetOverallProgress()

	// Should be close to 100%
	if finalProgress < 0.95 {
		t.Errorf("Final progress %.3f should be at least 95%%", finalProgress)
	}

	// Should not exceed 100%
	if finalProgress > 1.0 {
		t.Errorf("Final progress %.3f should not exceed 100%%", finalProgress)
	}
}
