// ABOUTME: Continuous build system with file watching capabilities
// ABOUTME: Monitors file changes and triggers appropriate builds with debouncing

package build

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"tamarou.com/pvm/internal/project"
)

// BuildType represents different types of builds that can be triggered
type BuildType int

const (
	BuildTypeTypeCheck BuildType = iota
	BuildTypeInline
	BuildTypeDistribution
	BuildTypeFull
)

func (bt BuildType) String() string {
	switch bt {
	case BuildTypeTypeCheck:
		return "typecheck"
	case BuildTypeInline:
		return "inline"
	case BuildTypeDistribution:
		return "distribution"
	case BuildTypeFull:
		return "full"
	default:
		return "unknown"
	}
}

// FileChangeEvent represents a file system change event
type FileChangeEvent struct {
	Path      string
	Operation string // "create", "modify", "delete", "rename"
	Timestamp time.Time
}

// BuildEvent represents a build trigger event
type BuildEvent struct {
	Type      BuildType
	Files     []string
	Reason    string
	Timestamp time.Time
}

// BuildResult represents the result of a build operation
type BuildResult struct {
	Type      BuildType
	Success   bool
	Duration  time.Duration
	Error     error
	Files     []string
	Timestamp time.Time
}

// BuildWatcher monitors file system for changes and triggers appropriate builds
type BuildWatcher struct {
	mu              sync.RWMutex
	projectContext  *project.ProjectContext
	watchDirs       []string
	patterns        []string
	excludePatterns []string

	// Build components
	pscBuilder    *PSCBuilder
	inlineBuilder *InlineBuilder
	distBuilder   *DistributionBuilder

	// Event handling
	debounceDelay time.Duration
	eventQueue    chan FileChangeEvent
	buildQueue    chan BuildEvent
	resultChannel chan BuildResult

	// State tracking
	fileModTimes  map[string]time.Time
	lastBuildTime time.Time
	isWatching    bool

	// Cancellation
	ctx    context.Context
	cancel context.CancelFunc
}

// WatcherConfig configures the build watcher
type WatcherConfig struct {
	WatchDirs       []string
	DebounceDelay   time.Duration
	FilePatterns    []string
	ExcludePatterns []string
	EnableTypeCheck bool
	EnableInline    bool
	EnableDist      bool
}

// DefaultWatcherConfig returns sensible defaults for the watcher
func DefaultWatcherConfig() *WatcherConfig {
	return &WatcherConfig{
		WatchDirs:     []string{"lib", "script", "t"},
		DebounceDelay: 500 * time.Millisecond,
		FilePatterns:  []string{"*.pm", "*.pl", "*.t"},
		ExcludePatterns: []string{
			"*.pmc",
			"build/*",
			"local/*",
			".git/*",
			"*.tmp",
			"*.bak",
		},
		EnableTypeCheck: true,
		EnableInline:    true,
		EnableDist:      false, // Distribution builds are typically slower
	}
}

// NewBuildWatcher creates a new build watcher instance
func NewBuildWatcher(projectCtx *project.ProjectContext, config *WatcherConfig) (*BuildWatcher, error) {
	if config == nil {
		config = DefaultWatcherConfig()
	}

	// Create build components
	pscBuilder, err := NewPSCBuilder(projectCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to create PSC builder: %w", err)
	}

	inlineBuilder, err := NewInlineBuilder(projectCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to create inline builder: %w", err)
	}

	distBuilder, err := NewDistributionBuilder(projectCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to create distribution builder: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &BuildWatcher{
		projectContext:  projectCtx,
		watchDirs:       config.WatchDirs,
		patterns:        config.FilePatterns,
		excludePatterns: config.ExcludePatterns,
		pscBuilder:      pscBuilder,
		inlineBuilder:   inlineBuilder,
		distBuilder:     distBuilder,
		debounceDelay:   config.DebounceDelay,
		eventQueue:      make(chan FileChangeEvent, 100),
		buildQueue:      make(chan BuildEvent, 10),
		resultChannel:   make(chan BuildResult, 10),
		fileModTimes:    make(map[string]time.Time),
		ctx:             ctx,
		cancel:          cancel,
	}, nil
}

// Start begins watching for file changes and processing build events
func (w *BuildWatcher) Start() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.isWatching {
		return fmt.Errorf("watcher is already running")
	}

	w.isWatching = true

	// Initialize file modification times
	if err := w.initializeFileStates(); err != nil {
		return fmt.Errorf("failed to initialize file states: %w", err)
	}

	// Start goroutines
	go w.fileWatcher()
	go w.eventProcessor()
	go w.buildProcessor()

	return nil
}

// Stop stops the file watcher and cleans up resources
func (w *BuildWatcher) Stop() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.isWatching {
		return nil
	}

	w.isWatching = false
	w.cancel()

	return nil
}

// Results returns the channel for build results
func (w *BuildWatcher) Results() <-chan BuildResult {
	return w.resultChannel
}

// TriggerBuild manually triggers a specific type of build
func (w *BuildWatcher) TriggerBuild(buildType BuildType, reason string) {
	if !w.isWatching {
		return
	}

	select {
	case w.buildQueue <- BuildEvent{
		Type:      buildType,
		Reason:    reason,
		Timestamp: time.Now(),
	}:
	default:
		// Build queue is full, skip this trigger
	}
}

// initializeFileStates scans watch directories and records initial file states
func (w *BuildWatcher) initializeFileStates() error {
	for _, dir := range w.watchDirs {
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(w.projectContext.RootDir, dir)
		}

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue // Skip non-existent directories
		}

		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && w.shouldWatchFile(path) {
				w.fileModTimes[path] = info.ModTime()
			}

			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to walk directory %s: %w", dir, err)
		}
	}

	return nil
}

// fileWatcher monitors directories for file changes using polling
func (w *BuildWatcher) fileWatcher() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			w.checkForChanges()
		}
	}
}

// checkForChanges scans directories for file modifications
func (w *BuildWatcher) checkForChanges() {
	for _, dir := range w.watchDirs {
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(w.projectContext.RootDir, dir)
		}

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || !w.shouldWatchFile(path) {
				return nil
			}

			w.mu.Lock()
			lastModTime, exists := w.fileModTimes[path]
			w.mu.Unlock()

			if !exists {
				// New file
				w.mu.Lock()
				w.fileModTimes[path] = info.ModTime()
				w.mu.Unlock()

				w.sendFileEvent(path, "create")
			} else if info.ModTime().After(lastModTime) {
				// Modified file
				w.mu.Lock()
				w.fileModTimes[path] = info.ModTime()
				w.mu.Unlock()

				w.sendFileEvent(path, "modify")
			}

			return nil
		})
	}

	// Check for deleted files
	w.mu.Lock()
	var deletedFiles []string
	for path := range w.fileModTimes {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			deletedFiles = append(deletedFiles, path)
		}
	}
	for _, path := range deletedFiles {
		delete(w.fileModTimes, path)
	}
	w.mu.Unlock()

	for _, path := range deletedFiles {
		w.sendFileEvent(path, "delete")
	}
}

// sendFileEvent sends a file change event to the event queue
func (w *BuildWatcher) sendFileEvent(path, operation string) {
	select {
	case w.eventQueue <- FileChangeEvent{
		Path:      path,
		Operation: operation,
		Timestamp: time.Now(),
	}:
	default:
		// Event queue is full, skip this event
	}
}

// eventProcessor processes file change events and determines build actions
func (w *BuildWatcher) eventProcessor() {
	var pendingEvents []FileChangeEvent
	debounceTimer := time.NewTimer(w.debounceDelay)
	debounceTimer.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case event, ok := <-w.eventQueue:
			if !ok {
				return // Channel closed
			}
			pendingEvents = append(pendingEvents, event)

			// Reset debounce timer
			debounceTimer.Stop()
			debounceTimer = time.NewTimer(w.debounceDelay)

		case <-debounceTimer.C:
			if len(pendingEvents) > 0 {
				w.processPendingEvents(pendingEvents)
				pendingEvents = nil
			}
		}
	}
}

// processPendingEvents analyzes file changes and triggers appropriate builds
func (w *BuildWatcher) processPendingEvents(events []FileChangeEvent) {
	// Group events by type
	var hasSourceChanges, hasTestChanges, hasConfigChanges bool
	changedFiles := make(map[string]struct{})

	for _, event := range events {
		changedFiles[event.Path] = struct{}{}

		switch {
		case w.isSourceFile(event.Path):
			hasSourceChanges = true
		case w.isTestFile(event.Path):
			hasTestChanges = true
		case w.isConfigFile(event.Path):
			hasConfigChanges = true
		}
	}

	// Determine build strategy
	var buildType BuildType
	var reason string

	switch {
	case hasConfigChanges:
		buildType = BuildTypeFull
		reason = "configuration changes detected"
	case hasSourceChanges:
		buildType = BuildTypeInline
		reason = "source file changes detected"
	case hasTestChanges:
		buildType = BuildTypeTypeCheck
		reason = "test file changes detected"
	default:
		return // No relevant changes
	}

	// Extract file list
	var files []string
	for file := range changedFiles {
		files = append(files, file)
	}

	// Queue build event
	select {
	case w.buildQueue <- BuildEvent{
		Type:      buildType,
		Files:     files,
		Reason:    reason,
		Timestamp: time.Now(),
	}:
	case <-w.ctx.Done():
		return // Context cancelled
	default:
		// Build queue is full, skip this build
	}
}

// buildProcessor executes build operations from the build queue
func (w *BuildWatcher) buildProcessor() {
	for {
		select {
		case <-w.ctx.Done():
			return
		case buildEvent, ok := <-w.buildQueue:
			if !ok {
				return // Channel closed
			}
			result := w.executeBuild(buildEvent)

			select {
			case w.resultChannel <- result:
			case <-w.ctx.Done():
				return
			default:
				// Result channel is full, skip this result
			}
		}
	}
}

// executeBuild performs the actual build operation
func (w *BuildWatcher) executeBuild(event BuildEvent) BuildResult {
	start := time.Now()

	result := BuildResult{
		Type:      event.Type,
		Files:     event.Files,
		Timestamp: start,
	}

	switch event.Type {
	case BuildTypeTypeCheck:
		pscResult, err := w.pscBuilder.TypeCheck(w.ctx, []string{w.projectContext.RootDir})
		result.Success = err == nil && pscResult != nil && pscResult.Success
		result.Error = err

	case BuildTypeInline:
		// First type check, then build inline if successful
		pscResult, err := w.pscBuilder.TypeCheck(w.ctx, []string{w.projectContext.RootDir})
		if err != nil || pscResult == nil || !pscResult.Success {
			result.Success = false
			result.Error = err
			break
		}

		targetDirs := []string{filepath.Join(w.projectContext.RootDir, "lib")}
		inlineResult, err := w.inlineBuilder.Build(w.ctx, targetDirs)
		result.Success = err == nil && inlineResult != nil && inlineResult.Success
		result.Error = err

	case BuildTypeDistribution:
		// Full distribution build
		distResult, err := w.distBuilder.Build(w.ctx, nil) // Pass nil for default options
		result.Success = err == nil && distResult != nil && distResult.Success
		result.Error = err

	case BuildTypeFull:
		// Type check + inline + distribution
		pscResult, err := w.pscBuilder.TypeCheck(w.ctx, []string{w.projectContext.RootDir})
		if err != nil || pscResult == nil || !pscResult.Success {
			result.Success = false
			result.Error = err
			break
		}

		targetDirs := []string{filepath.Join(w.projectContext.RootDir, "lib")}
		inlineResult, err := w.inlineBuilder.Build(w.ctx, targetDirs)
		if err != nil || inlineResult == nil || !inlineResult.Success {
			result.Success = false
			result.Error = err
			break
		}

		distResult, err := w.distBuilder.Build(w.ctx, nil) // Pass nil for default options
		result.Success = err == nil && distResult != nil && distResult.Success
		result.Error = err
	}

	result.Duration = time.Since(start)
	w.lastBuildTime = time.Now()

	return result
}

// shouldWatchFile determines if a file should be monitored for changes
func (w *BuildWatcher) shouldWatchFile(path string) bool {
	// Check exclude patterns first (handle both directory paths and wildcards)
	for _, pattern := range w.excludePatterns {
		// Remove leading/trailing wildcards for directory matching
		cleanPattern := strings.Trim(pattern, "*")
		if strings.Contains(path, cleanPattern) {
			return false
		}
		// Also check wildcard patterns
		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			return false
		}
	}

	// Check include patterns
	for _, pattern := range w.patterns {
		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			return true
		}
	}

	return false
}

// isSourceFile checks if a file is a source file (.pm, .pl)
func (w *BuildWatcher) isSourceFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".pm" || ext == ".pl"
}

// isTestFile checks if a file is a test file (.t)
func (w *BuildWatcher) isTestFile(path string) bool {
	return strings.ToLower(filepath.Ext(path)) == ".t"
}

// isConfigFile checks if a file is a configuration file
func (w *BuildWatcher) isConfigFile(path string) bool {
	base := filepath.Base(path)
	return base == "pvm.toml" || base == "cpanfile" || base == ".perl-version"
}
