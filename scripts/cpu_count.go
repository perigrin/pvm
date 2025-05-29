// ABOUTME: Cross-platform CPU count detection utility for build parallelization
// ABOUTME: Returns optimal parallel job count (CPU cores + 1) for make and test execution

//go:build ignore

package main

import (
	"fmt"
	"runtime"
)

func main() {
	// Get number of logical CPUs
	cpus := runtime.NumCPU()
	
	// Add 1 for optimal parallelization (common practice)
	parallelJobs := cpus + 1
	
	fmt.Print(parallelJobs)
}