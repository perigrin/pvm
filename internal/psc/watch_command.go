// ABOUTME: PSC watch command implementation
// ABOUTME: Provides file watching for continuous type checking

package psc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/parser"
)

// File watch constants
const (
	watchDebounceTime = 500 * time.Millisecond // Time to wait for multiple changes
	watchShowErrors   = 10                     // Maximum number of errors to show
)

// newWatchCommand creates a command to watch files for type errors
func newWatchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watch [file|dir]",
		Short: "Watch files and report type errors on change",
		Long:  "Continuously monitor files for changes and perform type checking",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get flags
			exclude, _ := cmd.Flags().GetStringArray("exclude")
			recursive, _ := cmd.Flags().GetBool("recursive")
			verbose, _ := cmd.Flags().GetBool("verbose")
			noFlowSensitive, _ := cmd.Flags().GetBool("no-flow-sensitive")
			skipFlowChecks, _ := cmd.Flags().GetBool("skip-flow-checks")

			target := args[0]

			// Validate that the path exists
			info, err := os.Stat(target)
			if os.IsNotExist(err) {
				return errors.NewTypeError(
					"PSC-701",
					fmt.Sprintf("Path does not exist: %s", target),
					err).WithLocation(target)
			}

			// Initialize the file watcher
			watcher, err := fsnotify.NewWatcher()
			if err != nil {
				return errors.NewTypeError(
					"PSC-702",
					"Failed to create file watcher",
					err)
			}
			defer watcher.Close()

			// Store all paths to watch
			watchPaths := make(map[string]bool)

			if info.IsDir() {
				if verbose {
					fmt.Printf("Watching directory: %s\n", target)
				}

				// Add the directory
				watchPaths[target] = true

				// If recursive, add all subdirectories
				if recursive {
					if verbose {
						fmt.Println("Recursively watching subdirectories")
					}

					err := filepath.Walk(target, func(path string, info os.FileInfo, err error) error {
						if err != nil {
							return err
						}

						// Skip excluded paths
						for _, pattern := range exclude {
							matched, _ := filepath.Match(pattern, filepath.Base(path))
							if matched {
								return nil
							}
						}

						// Add directories to watch
						if info.IsDir() {
							watchPaths[path] = true
						}

						return nil
					})

					if err != nil {
						return errors.NewTypeError(
							"PSC-703",
							fmt.Sprintf("Failed to walk directory: %s", target),
							err).WithLocation(target)
					}
				}
			} else {
				// Single file
				if verbose {
					fmt.Printf("Watching file: %s\n", target)
				}
				watchPaths[filepath.Dir(target)] = true
			}

			// Add all paths to the watcher
			for path := range watchPaths {
				if err := watcher.Add(path); err != nil {
					return errors.NewTypeError(
						"PSC-704",
						fmt.Sprintf("Failed to watch path: %s", path),
						err).WithLocation(path)
				}
			}

			// Check initially
			if verbose {
				fmt.Println("Performing initial type check...")
			}

			err = checkPath(target, recursive, exclude, !noFlowSensitive, skipFlowChecks, verbose)
			if err != nil {
				fmt.Printf("Initial check failed: %v\n", err)
			}

			// Set up debouncing for file events
			var (
				currentFiles  = make(map[string]bool)
				shouldRecheck bool
				timer         *time.Timer
			)

			// Start the timer to handle debouncing
			timer = time.NewTimer(watchDebounceTime)
			timer.Stop() // Stop initially until we get an event

			// Wait for events
			fmt.Printf("Watching for changes (press Ctrl+C to stop)...\n")

			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return nil
					}

					// Skip some temp/hidden files
					if shouldSkipFile(event.Name, exclude) {
						continue
					}

					// Only care about Perl files
					if !isPerlFile(event.Name) && !info.IsDir() {
						continue
					}

					currentFiles[event.Name] = true
					shouldRecheck = true

					// Reset the timer to wait for a pause in events
					if !timer.Stop() {
						select {
						case <-timer.C:
						default:
						}
					}
					timer.Reset(watchDebounceTime)

				case <-timer.C:
					if shouldRecheck {
						// If it's a directory, check the whole directory
						if info.IsDir() {
							if verbose {
								fmt.Printf("\nChanges detected. Rechecking...\n")
							} else {
								fmt.Printf("\n=== %s ===\n", time.Now().Format("15:04:05"))
							}

							err := checkPath(target, recursive, exclude, !noFlowSensitive, skipFlowChecks, verbose)
							if err != nil {
								fmt.Printf("Check failed: %v\n", err)
							}
						} else {
							// For single files, we can just check the files that changed
							for file := range currentFiles {
								if verbose {
									fmt.Printf("\nChecking %s...\n", file)
								} else {
									fmt.Printf("\n=== %s: %s ===\n", time.Now().Format("15:04:05"), filepath.Base(file))
								}

								err := checkPath(file, false, exclude, !noFlowSensitive, skipFlowChecks, verbose)
								if err != nil {
									fmt.Printf("Check failed: %v\n", err)
								}
							}
						}

						// Reset for next round
						currentFiles = make(map[string]bool)
						shouldRecheck = false
					}

				case err, ok := <-watcher.Errors:
					if !ok {
						return nil
					}
					fmt.Printf("Error: %v\n", err)
				}
			}
		},
	}

	// Add flags for the watch command
	cmd.Flags().StringArrayP("exclude", "e", []string{}, "Patterns to exclude from watching")
	cmd.Flags().BoolP("recursive", "r", true, "Watch directories recursively")
	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
	cmd.Flags().Bool("no-flow-sensitive", false, "Disable flow-sensitive type analysis")
	cmd.Flags().Bool("skip-flow-checks", false, "Skip flow checks but still perform refinements")

	return cmd
}

// checkPath performs type checking on a file or directory
func checkPath(path string, recursive bool, exclude []string, flowSensitive bool, skipFlowChecks bool, verbose bool) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	// Create type checker
	tc, err := parser.NewTypeCheck()
	if err != nil {
		return err
	}

	// Configure type checker
	tc.EnableFlowSensitiveAnalysis = flowSensitive
	tc.SkipFlowChecks = skipFlowChecks

	if info.IsDir() {
		// Check all Perl files in the directory
		var files []string

		walkFn := func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip excluded files
			for _, pattern := range exclude {
				matched, _ := filepath.Match(pattern, filepath.Base(path))
				if matched {
					if info.IsDir() && recursive {
						return filepath.SkipDir
					}
					return nil
				}
			}

			// Only process files
			if !info.IsDir() && isPerlFile(path) {
				files = append(files, path)
			}

			return nil
		}

		if recursive {
			err = filepath.Walk(path, walkFn)
		} else {
			// Just scan the top-level directory
			items, err := os.ReadDir(path)
			if err != nil {
				return err
			}

			for _, item := range items {
				itemPath := filepath.Join(path, item.Name())
				info, err := item.Info()
				if err != nil {
					return err
				}
				_ = walkFn(itemPath, info, nil)
			}
		}

		if err != nil {
			return err
		}

		// Check each file
		errorCount := 0
		fileCount := 0

		for _, file := range files {
			fileCount++
			result, err := tc.CheckFile(file)
			if err != nil {
				fmt.Printf("Error checking %s: %v\n", file, err)
				errorCount++
				continue
			}

			if len(result.Errors) > 0 {
				// Print errors
				relPath, err := filepath.Rel(path, file)
				if err != nil {
					relPath = file
				}

				fmt.Printf("%s: %d errors\n", relPath, len(result.Errors))
				for i, tErr := range result.Errors {
					if i < watchShowErrors {
						fmt.Printf("  %s:%d:%d: %s\n", filepath.Base(tErr.Path), tErr.Line, tErr.Column, tErr.Message)
					} else if i == watchShowErrors {
						fmt.Printf("  ... and %d more errors\n", len(result.Errors)-watchShowErrors)
						break
					}
				}
				errorCount += len(result.Errors)
			} else if verbose {
				relPath, err := filepath.Rel(path, file)
				if err != nil {
					relPath = file
				}
				fmt.Printf("%s: No errors\n", relPath)
			}
		}

		// Summary
		if fileCount > 0 {
			if errorCount > 0 {
				fmt.Printf("\nFound %d errors in %d files\n", errorCount, fileCount)
			} else {
				fmt.Printf("\nNo errors found in %d files\n", fileCount)
			}
		} else {
			fmt.Println("No Perl files found to check")
		}
	} else {
		// Check a single file
		result, err := tc.CheckFile(path)
		if err != nil {
			return err
		}

		if len(result.Errors) > 0 {
			// Print errors
			fmt.Printf("%s: %d errors\n", filepath.Base(path), len(result.Errors))
			for i, tErr := range result.Errors {
				if i < watchShowErrors {
					fmt.Printf("  %s:%d:%d: %s\n", filepath.Base(tErr.Path), tErr.Line, tErr.Column, tErr.Message)
				} else if i == watchShowErrors {
					fmt.Printf("  ... and %d more errors\n", len(result.Errors)-watchShowErrors)
					break
				}
			}
		} else {
			fmt.Println("No type errors found")
		}
	}

	return nil
}

// isPerlFile checks if a file is a Perl source file
func isPerlFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".pl" || ext == ".pm" || ext == ".t"
}

// shouldSkipFile determines if a file should be skipped
func shouldSkipFile(path string, exclude []string) bool {
	// Skip hidden files and temporary files
	base := filepath.Base(path)
	if strings.HasPrefix(base, ".") || strings.HasSuffix(base, "~") || strings.HasPrefix(base, "#") {
		return true
	}

	// Skip files matching exclude patterns
	for _, pattern := range exclude {
		matched, _ := filepath.Match(pattern, base)
		if matched {
			return true
		}
	}

	return false
}
