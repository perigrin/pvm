// ABOUTME: Tests for BuildProgressCalculator functionality
// ABOUTME: Ensures proper progress tracking across build stages

package perl

import (
	"testing"
	"time"
)

func TestNewBuildProgressCalculator(t *testing.T) {
	calc := NewBuildProgressCalculator()

	// Test initial state
	if calc.currentStage != StageExtract {
		t.Errorf("Expected initial stage to be StageExtract, got %v", calc.currentStage)
	}

	if calc.stageProgress != 0.0 {
		t.Errorf("Expected initial stage progress to be 0.0, got %f", calc.stageProgress)
	}

	// Test stage weights are properly initialized
	expectedWeights := map[BuildProgressStage]float64{
		StageExtract:   0.05,
		StageConfigure: 0.15,
		StageCompile:   0.60,
		StageTest:      0.15,
		StageInstall:   0.05,
		StageCleanup:   0.0,
	}

	for stage, expectedWeight := range expectedWeights {
		if weight, exists := calc.stageWeights[stage]; !exists || weight != expectedWeight {
			t.Errorf("Expected stage %v to have weight %f, got %f", stage, expectedWeight, weight)
		}
	}
}

func TestSetStage(t *testing.T) {
	calc := NewBuildProgressCalculator()

	// Test setting to configure stage
	calc.SetStage(StageConfigure)

	if calc.currentStage != StageConfigure {
		t.Errorf("Expected stage to be StageConfigure, got %v", calc.currentStage)
	}

	if calc.stageProgress != 0.0 {
		t.Errorf("Expected stage progress to be reset to 0.0, got %f", calc.stageProgress)
	}

	if calc.totalLines != 0 {
		t.Errorf("Expected totalLines to be reset to 0, got %d", calc.totalLines)
	}

	if calc.currentLines != 0 {
		t.Errorf("Expected currentLines to be reset to 0, got %d", calc.currentLines)
	}

	// Test estimated duration is set
	if calc.estimatedStageDuration != 2*time.Minute {
		t.Errorf("Expected configure stage duration to be 2 minutes, got %v", calc.estimatedStageDuration)
	}
}

func TestUpdateStageProgress(t *testing.T) {
	calc := NewBuildProgressCalculator()

	// Test normal progress update
	calc.UpdateStageProgress(0.5)
	if calc.stageProgress != 0.5 {
		t.Errorf("Expected stage progress to be 0.5, got %f", calc.stageProgress)
	}

	// Test bounds checking - negative value
	calc.UpdateStageProgress(-0.1)
	if calc.stageProgress != 0.0 {
		t.Errorf("Expected stage progress to be clamped to 0.0, got %f", calc.stageProgress)
	}

	// Test bounds checking - value over 1.0
	calc.UpdateStageProgress(1.5)
	if calc.stageProgress != 1.0 {
		t.Errorf("Expected stage progress to be clamped to 1.0, got %f", calc.stageProgress)
	}
}

func TestUpdateTimeBasedProgress(t *testing.T) {
	calc := NewBuildProgressCalculator()
	calc.SetStage(StageConfigure)

	// Simulate time passing
	calc.stageStartTime = time.Now().Add(-1 * time.Minute)
	calc.UpdateTimeBasedProgress()

	// Should be about 50% complete (1 minute out of 2 minute estimate)
	expectedProgress := 0.5
	tolerance := 0.1

	if calc.stageProgress < expectedProgress-tolerance || calc.stageProgress > expectedProgress+tolerance {
		t.Errorf("Expected time-based progress to be around %f, got %f", expectedProgress, calc.stageProgress)
	}

	// Test capping at 95%
	calc.stageStartTime = time.Now().Add(-10 * time.Minute)
	calc.UpdateTimeBasedProgress()

	if calc.stageProgress > 0.95 {
		t.Errorf("Expected time-based progress to be capped at 95%%, got %f", calc.stageProgress)
	}
}

func TestUpdateLineBasedProgress(t *testing.T) {
	calc := NewBuildProgressCalculator()
	calc.totalLines = 100

	// Test line-based progress
	calc.UpdateLineBasedProgress()
	if calc.currentLines != 1 {
		t.Errorf("Expected currentLines to be 1, got %d", calc.currentLines)
	}

	if calc.stageProgress != 0.01 {
		t.Errorf("Expected stage progress to be 0.01, got %f", calc.stageProgress)
	}

	// Test with more lines
	for i := 0; i < 49; i++ {
		calc.UpdateLineBasedProgress()
	}

	if calc.currentLines != 50 {
		t.Errorf("Expected currentLines to be 50, got %d", calc.currentLines)
	}

	if calc.stageProgress != 0.5 {
		t.Errorf("Expected stage progress to be 0.5, got %f", calc.stageProgress)
	}
}

func TestGetOverallProgress(t *testing.T) {
	calc := NewBuildProgressCalculator()

	// Test progress at start of extraction
	progress := calc.GetOverallProgress()
	if progress != 0.0 {
		t.Errorf("Expected overall progress to be 0.0 at start, got %f", progress)
	}

	// Test progress at 50% of extraction stage
	calc.UpdateStageProgress(0.5)
	progress = calc.GetOverallProgress()
	expected := 0.05 * 0.5 // 5% stage weight * 50% stage progress
	if progress != expected {
		t.Errorf("Expected overall progress to be %f, got %f", expected, progress)
	}

	// Test progress at start of configure stage
	calc.SetStage(StageConfigure)
	progress = calc.GetOverallProgress()
	expected = 0.05 // Complete extraction stage
	if progress != expected {
		t.Errorf("Expected overall progress to be %f, got %f", expected, progress)
	}

	// Test progress at 50% of configure stage
	calc.UpdateStageProgress(0.5)
	progress = calc.GetOverallProgress()
	expected = 0.05 + (0.15 * 0.5) // Extraction + 50% of configure
	if progress != expected {
		t.Errorf("Expected overall progress to be %f, got %f", expected, progress)
	}

	// Test progress at start of compile stage
	calc.SetStage(StageCompile)
	progress = calc.GetOverallProgress()
	expected = 0.05 + 0.15 // Extraction + Configure
	if progress != expected {
		t.Errorf("Expected overall progress to be %f, got %f", expected, progress)
	}

	// Test progress at 50% of compile stage
	calc.UpdateStageProgress(0.5)
	progress = calc.GetOverallProgress()
	expected = 0.05 + 0.15 + (0.60 * 0.5) // Previous stages + 50% of compile
	if progress != expected {
		t.Errorf("Expected overall progress to be %f, got %f", expected, progress)
	}
}

func TestParseMakeProgress(t *testing.T) {
	calc := NewBuildProgressCalculator()
	calc.SetStage(StageCompile)

	// Test parsing 100% completion
	result := calc.ParseMakeProgress("something [100%] something")
	if !result {
		t.Error("Expected ParseMakeProgress to return true for [100%] pattern")
	}
	if calc.stageProgress != 1.0 {
		t.Errorf("Expected stage progress to be 1.0, got %f", calc.stageProgress)
	}

	// Test parsing percentage
	calc.stageProgress = 0.0
	result = calc.ParseMakeProgress("building [45%] complete")
	if !result {
		t.Error("Expected ParseMakeProgress to return true for [45%] pattern")
	}
	if calc.stageProgress != 0.45 {
		t.Errorf("Expected stage progress to be 0.45, got %f", calc.stageProgress)
	}

	// Test parsing compilation indicators
	calc.stageProgress = 0.0
	calc.totalLines = 0
	calc.currentLines = 0
	result = calc.ParseMakeProgress("CC some_file.c")
	if !result {
		t.Error("Expected ParseMakeProgress to return true for CC pattern")
	}
	if calc.currentLines != 1 {
		t.Errorf("Expected currentLines to be 1, got %d", calc.currentLines)
	}

	// Test parsing linking indicators
	calc.stageProgress = 0.0
	calc.totalLines = 0
	calc.currentLines = 0
	result = calc.ParseMakeProgress("LD perl")
	if !result {
		t.Error("Expected ParseMakeProgress to return true for LD pattern")
	}
	if calc.currentLines != 1 {
		t.Errorf("Expected currentLines to be 1, got %d", calc.currentLines)
	}

	// Test non-matching line
	result = calc.ParseMakeProgress("some random output")
	if result {
		t.Error("Expected ParseMakeProgress to return false for non-matching pattern")
	}
}

func TestProgressCalculatorIntegration(t *testing.T) {
	calc := NewBuildProgressCalculator()

	// Simulate a full build process
	stages := []BuildProgressStage{
		StageExtract,
		StageConfigure,
		StageCompile,
		StageTest,
		StageInstall,
	}

	previousProgress := 0.0

	for _, stage := range stages {
		calc.SetStage(stage)

		// Test progress at stage start
		progress := calc.GetOverallProgress()
		if progress < previousProgress {
			t.Errorf("Overall progress should not decrease. Previous: %f, Current: %f", previousProgress, progress)
		}

		// Test progress at stage middle
		calc.UpdateStageProgress(0.5)
		middleProgress := calc.GetOverallProgress()
		if middleProgress <= progress {
			t.Errorf("Progress should increase within stage. Start: %f, Middle: %f", progress, middleProgress)
		}

		// Test progress at stage end
		calc.UpdateStageProgress(1.0)
		endProgress := calc.GetOverallProgress()
		if endProgress <= middleProgress {
			t.Errorf("Progress should increase to stage end. Middle: %f, End: %f", middleProgress, endProgress)
		}

		previousProgress = endProgress
	}

	// Final progress should be close to 1.0 (100%)
	finalProgress := calc.GetOverallProgress()
	if finalProgress < 0.95 {
		t.Errorf("Final progress should be at least 95%%, got %f", finalProgress)
	}
}

func TestProgressCalculatorStageWeights(t *testing.T) {
	calc := NewBuildProgressCalculator()

	// Test that stage weights sum to 1.0 (excluding cleanup)
	totalWeight := 0.0
	for stage, weight := range calc.stageWeights {
		if stage != StageCleanup {
			totalWeight += weight
		}
	}

	if totalWeight != 1.0 {
		t.Errorf("Stage weights should sum to 1.0, got %f", totalWeight)
	}

	// Test that compile stage has the highest weight
	compileWeight := calc.stageWeights[StageCompile]
	for stage, weight := range calc.stageWeights {
		if stage != StageCompile && weight > compileWeight {
			t.Errorf("Compile stage should have highest weight, but %v has weight %f > %f", stage, weight, compileWeight)
		}
	}
}
