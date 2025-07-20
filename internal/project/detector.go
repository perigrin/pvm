// ABOUTME: Project context detection for automatic project-aware behavior
// ABOUTME: Detects project boundaries using markers like .perl-version, cpanfile, pvm.toml

package project

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// ProjectContext holds information about a detected project
type ProjectContext struct {
	IsProject     bool
	RootDir       string
	PerlVersion   string
	HasCpanfile   bool
	LocalLibDir   string
	ConfigFile    string
	DetectionInfo string // How the project was detected
}

// detectionCache stores cached project detection results
var detectionCache = make(map[string]*ProjectContext)
var cacheMutex sync.RWMutex

// projectMarkers defines files that indicate a project root (in priority order)
var projectMarkers = []string{
	".perl-version",
	"cpanfile",
	"pvm.toml",
	".git",
}

// DetectProject detects project context starting from workingDir and walking up the directory tree
func DetectProject(workingDir string) (*ProjectContext, error) {
	// Clean the working directory path
	cleanDir, err := filepath.Abs(workingDir)
	if err != nil {
		return nil, err
	}

	// Check if the directory exists and is accessible
	if _, err := os.Stat(cleanDir); err != nil {
		return nil, err
	}

	// Check cache first
	cacheMutex.RLock()
	if cached, exists := detectionCache[cleanDir]; exists {
		cacheMutex.RUnlock()
		return cached, nil
	}
	cacheMutex.RUnlock()

	// Walk up directory tree looking for project markers
	result := &ProjectContext{
		IsProject:   false,
		LocalLibDir: filepath.Join(cleanDir, "local"), // Changed default from "lib" to "local"
	}

	currentDir := cleanDir
	for {
		// Check for project markers in current directory
		for _, marker := range projectMarkers {
			markerPath := filepath.Join(currentDir, marker)
			if fileExists(markerPath) {
				result.IsProject = true
				result.RootDir = currentDir
				result.DetectionInfo = marker

				// Set specific information based on marker type
				switch marker {
				case ".perl-version":
					if version, err := readPerlVersion(markerPath); err == nil {
						result.PerlVersion = version
					}
				case "cpanfile":
					result.HasCpanfile = true
				case "pvm.toml":
					result.ConfigFile = markerPath
				}

				// Update local lib directory to be relative to project root
				result.LocalLibDir = filepath.Join(result.RootDir, "local") // Changed default from "lib" to "local"

				// Check for additional project information now that we found the root
				enrichProjectContext(result)

				// Cache the result
				cacheMutex.Lock()
				detectionCache[cleanDir] = result
				cacheMutex.Unlock()

				return result, nil
			}
		}

		// Move up one directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// Reached filesystem root without finding project markers
			break
		}
		currentDir = parentDir
	}

	// Cache negative result (not a project)
	cacheMutex.Lock()
	detectionCache[cleanDir] = result
	cacheMutex.Unlock()

	return result, nil
}

// enrichProjectContext adds additional information once project root is found
func enrichProjectContext(ctx *ProjectContext) {
	if ctx.RootDir == "" {
		return
	}

	// Check for .perl-version if not already found
	if ctx.PerlVersion == "" {
		perlVersionPath := filepath.Join(ctx.RootDir, ".perl-version")
		if version, err := readPerlVersion(perlVersionPath); err == nil {
			ctx.PerlVersion = version
		}
	}

	// Check for cpanfile if not already found
	if !ctx.HasCpanfile {
		cpanfilePath := filepath.Join(ctx.RootDir, "cpanfile")
		ctx.HasCpanfile = fileExists(cpanfilePath)
	}

	// Check for pvm.toml if not already set
	if ctx.ConfigFile == "" {
		pvmConfigPath := filepath.Join(ctx.RootDir, "pvm.toml")
		if fileExists(pvmConfigPath) {
			ctx.ConfigFile = pvmConfigPath
		}
	}
}

// readPerlVersion reads the Perl version from .perl-version file
func readPerlVersion(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	version := strings.TrimSpace(string(content))
	return version, nil
}

// fileExists checks if a file or directory exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ClearDetectionCache clears the project detection cache (useful for testing)
func ClearDetectionCache() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	detectionCache = make(map[string]*ProjectContext)
}

// GetCurrentProject returns project context for the current working directory
func GetCurrentProject() (*ProjectContext, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return DetectProject(wd)
}
