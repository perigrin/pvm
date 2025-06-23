// ABOUTME: Mode detection and argument parsing for tool execution
// ABOUTME: Determines when PVX should operate in tool vs script mode

package tool

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Detector handles execution mode detection and argument parsing
type Detector struct {
	knownTools       map[string]bool
	scriptExtensions map[string]bool
	toolNamePattern  *regexp.Regexp
	allowAmbiguous   bool
	preferScriptMode bool
}

// NewDetector creates a new execution mode detector
func NewDetector() *Detector {
	// Known Perl tools that should always be interpreted as tools
	knownTools := map[string]bool{
		"perl":       true,
		"cpan":       true,
		"cpanm":      true,
		"prove":      true,
		"perldoc":    true,
		"h2ph":       true,
		"h2xs":       true,
		"enc2xs":     true,
		"xsubpp":     true,
		"corelist":   true,
		"plackup":    true,
		"carton":     true,
		"dzil":       true,
		"perlcritic": true,
		"perltidy":   true,
		"ack":        true,
		"fatpack":    true,
		"minicpan":   true,
		"perlbrew":   true,
		"cpanfile":   true,
	}

	// Common script file extensions
	scriptExtensions := map[string]bool{
		".pl":   true,
		".pm":   true,
		".t":    true,
		".pod":  true,
		".perl": true,
	}

	// Pattern for valid tool names (alphanumeric, hyphens, underscores)
	toolNamePattern := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)

	return &Detector{
		knownTools:       knownTools,
		scriptExtensions: scriptExtensions,
		toolNamePattern:  toolNamePattern,
		allowAmbiguous:   false,
		preferScriptMode: false,
	}
}

// SetOptions configures detector behavior
func (d *Detector) SetOptions(allowAmbiguous bool, preferScriptMode bool) {
	d.allowAmbiguous = allowAmbiguous
	d.preferScriptMode = preferScriptMode
}

// AddKnownTool adds a tool to the known tools list
func (d *Detector) AddKnownTool(toolName string) {
	d.knownTools[toolName] = true
}

// DetectExecutionMode analyzes command-line arguments to determine execution mode
func (d *Detector) DetectExecutionMode(args []string) (*DetectionResult, error) {
	if len(args) == 0 {
		return &DetectionResult{
			Mode:       ModeAmbiguous,
			Confidence: 0.0,
			Reason:     "no arguments provided",
		}, nil
	}

	firstArg := args[0]
	remainingArgs := []string{}
	if len(args) > 1 {
		remainingArgs = args[1:]
	}

	// Check for inline code execution (handled elsewhere)
	if firstArg == "-e" || firstArg == "--execute" {
		return &DetectionResult{
			Mode:       ModeInline,
			InlineCode: strings.Join(remainingArgs, " "),
			Confidence: 1.0,
			Reason:     "inline code execution flag detected",
		}, nil
	}

	// Apply detection logic
	result := d.analyzeArgument(firstArg, remainingArgs)

	// Handle ambiguous cases
	if result.Mode == ModeAmbiguous {
		if !d.allowAmbiguous {
			alternatives := d.generateAlternatives(firstArg)
			return result, NewAmbiguousModeError(firstArg, result.Reason, alternatives)
		}

		// Apply preference for ambiguous cases
		if d.preferScriptMode {
			result.Mode = ModeScript
			result.ScriptPath = firstArg
			result.Reason += " (defaulting to script mode)"
		} else {
			result.Mode = ModeTool
			result.ToolName = firstArg
			result.Reason += " (defaulting to tool mode)"
		}
		result.Confidence = 0.5
	}

	return result, nil
}

// analyzeArgument performs the core analysis of the first argument
func (d *Detector) analyzeArgument(arg string, remainingArgs []string) *DetectionResult {
	result := &DetectionResult{
		Arguments: remainingArgs,
	}

	// Check if it's a known tool
	if d.knownTools[arg] {
		result.Mode = ModeTool
		result.ToolName = arg
		result.Confidence = 1.0
		result.Reason = "matches known tool name"
		return result
	}

	// Check if file exists first (highest confidence)
	if _, err := os.Stat(arg); err == nil {
		result.Mode = ModeScript
		result.ScriptPath = arg
		result.Confidence = 0.9
		result.Reason = "existing file detected"
		return result
	}

	// Check if it has a script extension (high confidence)
	if d.hasScriptExtension(arg) {
		result.Mode = ModeScript
		result.ScriptPath = arg
		result.Confidence = 0.8
		result.Reason = "has script file extension"
		return result
	}

	// Check if it looks like a file path but doesn't exist
	if d.looksLikeFilePath(arg) {
		result.Mode = ModeScript
		result.ScriptPath = arg
		result.Confidence = 0.7
		result.Reason = "looks like file path (file not found)"
		return result
	}

	// Check if it matches tool name pattern
	if d.toolNamePattern.MatchString(arg) {
		// Could be a tool name
		result.Mode = ModeTool
		result.ToolName = arg
		result.Confidence = 0.6
		result.Reason = "matches tool name pattern"

		// But also could be a script without extension
		if !strings.Contains(arg, "/") && !strings.Contains(arg, "\\") {
			// Increase ambiguity for simple names
			result.Mode = ModeAmbiguous
			result.Confidence = 0.5
			result.Reason = "could be tool name or script without extension"
			result.Alternatives = []string{
				"tool: " + arg,
				"script: " + arg,
			}
		}
		return result
	}

	// Default to ambiguous
	result.Mode = ModeAmbiguous
	result.Confidence = 0.0
	result.Reason = "cannot determine mode from argument format"
	result.Alternatives = d.generateAlternatives(arg)

	return result
}

// looksLikeFilePath determines if an argument looks like a file path
func (d *Detector) looksLikeFilePath(arg string) bool {
	// Contains path separators
	if strings.Contains(arg, "/") || strings.Contains(arg, "\\") {
		return true
	}

	// Starts with current directory indicators
	if strings.HasPrefix(arg, "./") || strings.HasPrefix(arg, ".\\") {
		return true
	}

	// Has a file extension
	if strings.Contains(arg, ".") {
		ext := filepath.Ext(arg)
		return ext != ""
	}

	return false
}

// hasScriptExtension checks if the argument has a script file extension
func (d *Detector) hasScriptExtension(arg string) bool {
	ext := strings.ToLower(filepath.Ext(arg))
	return d.scriptExtensions[ext]
}

// generateAlternatives generates possible interpretations for ambiguous input
func (d *Detector) generateAlternatives(arg string) []string {
	alternatives := []string{}

	// Always suggest tool interpretation
	alternatives = append(alternatives, "tool: "+arg)

	// Suggest script interpretation
	alternatives = append(alternatives, "script: "+arg)

	// Suggest common variations
	if !strings.Contains(arg, ".") {
		alternatives = append(alternatives, "script: "+arg+".pl")
	}

	// Suggest path variations
	if !strings.Contains(arg, "/") && !strings.Contains(arg, "\\") {
		alternatives = append(alternatives, "script: ./"+arg)
	}

	return alternatives
}

// ValidateToolName validates a tool name according to naming conventions
func (d *Detector) ValidateToolName(toolName string) error {
	if toolName == "" {
		return NewInvalidToolNameError(toolName, "tool name cannot be empty")
	}

	if len(toolName) > 64 {
		return NewInvalidToolNameError(toolName, "tool name too long (max 64 characters)")
	}

	if !d.toolNamePattern.MatchString(toolName) {
		return NewInvalidToolNameError(toolName, "tool name must start with a letter and contain only letters, numbers, hyphens, and underscores")
	}

	// Check for reserved names
	reserved := []string{"help", "version", "config", "install", "update", "remove", "list"}
	for _, r := range reserved {
		if toolName == r {
			return NewInvalidToolNameError(toolName, "tool name is reserved")
		}
	}

	return nil
}

// IsKnownTool checks if a tool name is in the known tools list
func (d *Detector) IsKnownTool(toolName string) bool {
	return d.knownTools[toolName]
}

// GetKnownTools returns a list of all known tool names
func (d *Detector) GetKnownTools() []string {
	tools := make([]string, 0, len(d.knownTools))
	for tool := range d.knownTools {
		tools = append(tools, tool)
	}
	return tools
}
