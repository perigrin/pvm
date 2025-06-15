// ABOUTME: Enhanced help system with context-aware suggestions and workflow guidance
// ABOUTME: Provides intelligent help content based on project state and user context

package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/project"
)

// HelpManager provides enhanced help functionality
type HelpManager struct {
	projectContext *project.ProjectContext
}

// NewHelpManager creates a new help manager with current project context
func NewHelpManager() *HelpManager {
	projectCtx, _ := project.GetCurrentProject()
	return &HelpManager{
		projectContext: projectCtx,
	}
}

// HelpCategory represents different types of help content
type HelpCategory struct {
	Name        string
	Description string
	Commands    []CommandSuggestion
}

// CommandSuggestion represents a suggested command with context
type CommandSuggestion struct {
	Command     string
	Description string
	Example     string
	Relevance   string // Why this command is relevant
}

// GetContextualHelp returns help content tailored to current project context
func (h *HelpManager) GetContextualHelp() []HelpCategory {
	categories := []HelpCategory{}

	// Getting Started category - always present
	categories = append(categories, h.getGettingStartedCategory())

	// Project-specific categories
	if h.projectContext != nil && h.projectContext.IsProject {
		categories = append(categories, h.getProjectWorkflowCategory())
		categories = append(categories, h.getBuildAndTestCategory())
		categories = append(categories, h.getModuleManagementCategory())
	} else {
		categories = append(categories, h.getProjectSetupCategory())
		categories = append(categories, h.getPerlVersionManagementCategory())
	}

	// Universal categories
	categories = append(categories, h.getTroubleshootingCategory())

	return categories
}

// getGettingStartedCategory provides basic getting started commands
func (h *HelpManager) getGettingStartedCategory() HelpCategory {
	commands := []CommandSuggestion{
		{
			Command:     "pvm version",
			Description: "Show PVM version information",
			Example:     "pvm version",
			Relevance:   "Verify PVM installation",
		},
		{
			Command:     "pvm help workflows",
			Description: "Show common development workflows",
			Example:     "pvm help workflows",
			Relevance:   "Learn common usage patterns",
		},
	}

	// Add project-specific getting started
	if h.projectContext != nil && h.projectContext.IsProject {
		commands = append(commands, CommandSuggestion{
			Command:     "pvm project status",
			Description: "Show current project status",
			Example:     "pvm project status",
			Relevance:   "Check project health and configuration",
		})
	} else {
		commands = append(commands, CommandSuggestion{
			Command:     "pvm project init",
			Description: "Initialize a new Perl project",
			Example:     "pvm project init my-app",
			Relevance:   "Start a new project with PVM",
		})
	}

	return HelpCategory{
		Name:        "Getting Started",
		Description: "Essential commands to get you started with PVM",
		Commands:    commands,
	}
}

// getProjectWorkflowCategory provides project-specific workflow commands
func (h *HelpManager) getProjectWorkflowCategory() HelpCategory {
	commands := []CommandSuggestion{
		{
			Command:     "pvm dev",
			Description: "Start development environment with file watching",
			Example:     "pvm dev",
			Relevance:   "Get instant feedback while coding",
		},
		{
			Command:     "pvm test",
			Description: "Run project tests",
			Example:     "pvm test",
			Relevance:   "Verify code functionality",
		},
		{
			Command:     "pvm build",
			Description: "Build project for distribution",
			Example:     "pvm build",
			Relevance:   "Create deployable artifacts",
		},
	}

	if h.projectContext.HasCpanfile {
		commands = append(commands, CommandSuggestion{
			Command:     "pvm module install",
			Description: "Install dependencies from cpanfile",
			Example:     "pvm module install",
			Relevance:   "Project has cpanfile with dependencies",
		})
	}

	return HelpCategory{
		Name:        "Project Workflow",
		Description: "Commands for working within your Perl project",
		Commands:    commands,
	}
}

// getBuildAndTestCategory provides build and test related commands
func (h *HelpManager) getBuildAndTestCategory() HelpCategory {
	commands := []CommandSuggestion{
		{
			Command:     "pvm build --inline",
			Description: "Build for development (.pmc files)",
			Example:     "pvm build --inline",
			Relevance:   "Fast development builds with type checking",
		},
		{
			Command:     "pvm build --watch",
			Description: "Continuous building with file monitoring",
			Example:     "pvm build --watch",
			Relevance:   "Automatic rebuilds on file changes",
		},
		{
			Command:     "pvm test --verbose",
			Description: "Run tests with detailed output",
			Example:     "pvm test --verbose",
			Relevance:   "Debug test failures",
		},
	}

	return HelpCategory{
		Name:        "Build & Test",
		Description: "Commands for building and testing your project",
		Commands:    commands,
	}
}

// getModuleManagementCategory provides module management commands
func (h *HelpManager) getModuleManagementCategory() HelpCategory {
	commands := []CommandSuggestion{
		{
			Command:     "pvm module add",
			Description: "Add a new dependency to the project",
			Example:     "pvm module add DBI",
			Relevance:   "Add dependencies that get automatically installed",
		},
		{
			Command:     "pvm module sync",
			Description: "Generate/update dependency lockfile",
			Example:     "pvm module sync",
			Relevance:   "Ensure reproducible builds",
		},
	}

	if !h.projectContext.HasCpanfile {
		commands = append(commands, CommandSuggestion{
			Command:     "touch cpanfile",
			Description: "Create cpanfile for dependency management",
			Example:     "touch cpanfile",
			Relevance:   "Project lacks dependency management",
		})
	}

	return HelpCategory{
		Name:        "Module Management",
		Description: "Commands for managing CPAN modules and dependencies",
		Commands:    commands,
	}
}

// getProjectSetupCategory provides project setup commands for non-project directories
func (h *HelpManager) getProjectSetupCategory() HelpCategory {
	commands := []CommandSuggestion{
		{
			Command:     "pvm project init",
			Description: "Initialize a new Perl project in current directory",
			Example:     "pvm project init",
			Relevance:   "You're not in a project directory",
		},
		{
			Command:     "pvm project init my-app",
			Description: "Create a new project directory",
			Example:     "pvm project init my-app",
			Relevance:   "Start a new project from scratch",
		},
	}

	return HelpCategory{
		Name:        "Project Setup",
		Description: "Commands for setting up new Perl projects",
		Commands:    commands,
	}
}

// getPerlVersionManagementCategory provides Perl version management commands
func (h *HelpManager) getPerlVersionManagementCategory() HelpCategory {
	commands := []CommandSuggestion{
		{
			Command:     "pvm available",
			Description: "List available Perl versions for installation",
			Example:     "pvm available",
			Relevance:   "See what Perl versions you can install",
		},
		{
			Command:     "pvm install 5.38.0",
			Description: "Install a specific Perl version",
			Example:     "pvm install 5.38.0",
			Relevance:   "Get a modern Perl version",
		},
		{
			Command:     "pvm versions",
			Description: "List installed Perl versions",
			Example:     "pvm versions",
			Relevance:   "See what's available locally",
		},
	}

	return HelpCategory{
		Name:        "Perl Version Management",
		Description: "Commands for managing Perl installations",
		Commands:    commands,
	}
}

// getTroubleshootingCategory provides troubleshooting commands
func (h *HelpManager) getTroubleshootingCategory() HelpCategory {
	commands := []CommandSuggestion{
		{
			Command:     "pvm project doctor",
			Description: "Check project health and suggest fixes",
			Example:     "pvm project doctor",
			Relevance:   "Diagnose common issues",
		},
		{
			Command:     "pvm project doctor --fix",
			Description: "Automatically fix common project issues",
			Example:     "pvm project doctor --fix",
			Relevance:   "Resolve problems automatically",
		},
	}

	if h.projectContext != nil && h.projectContext.IsProject {
		commands = append(commands, CommandSuggestion{
			Command:     "pvm project status --json",
			Description: "Get detailed project status in JSON format",
			Example:     "pvm project status --json",
			Relevance:   "Programmatic access to project information",
		})
	}

	return HelpCategory{
		Name:        "Troubleshooting",
		Description: "Commands for diagnosing and fixing issues",
		Commands:    commands,
	}
}

// GetWorkflowHelp returns workflow-specific help content
func (h *HelpManager) GetWorkflowHelp() map[string]string {
	workflows := map[string]string{
		"new-project": `
Creating a New Perl Project:

1. Initialize the project:
   pvm project init my-app
   cd my-app

2. Add dependencies:
   pvm module add DBI
   pvm module add Test::More --dev

3. Install dependencies:
   pvm module install

4. Start development:
   pvm dev

This sets up a complete project with dependency management and development tools.
`,

		"existing-project": `
Working with an Existing Project:

1. Check project status:
   pvm project status

2. Install dependencies:
   pvm module install

3. Run tests:
   pvm test

4. Start development mode:
   pvm dev

The dev command provides file watching, automatic builds, and test running.
`,

		"module-development": `
Developing a CPAN Module:

1. Initialize with module template:
   pvm project init --template=module My::Module

2. Set up distribution metadata:
   # Edit pvm.toml to configure author, license, etc.

3. Develop with type checking:
   pvm dev

4. Build for distribution:
   pvm build

5. Test the distribution:
   cd build && perl Makefile.PL && make test

The build command creates a CPAN-ready distribution in the build/ directory.
`,

		"testing": `
Testing Workflows:

1. Run all tests:
   pvm test

2. Run specific test file:
   pvm test t/basic.t

3. Run tests with coverage:
   pvm test --coverage

4. Continuous testing:
   pvm dev  # Includes automatic test running

5. Debug test failures:
   pvm test --verbose

Tests are automatically discovered in the t/ directory.
`,

		"building": `
Build Workflows:

1. Development build (fast, with .pmc files):
   pvm build --inline

2. Distribution build (for CPAN):
   pvm build

3. Continuous building:
   pvm build --watch

4. Type-check only:
   pvm build --check-only

5. Clean build:
   pvm build --clean

Build outputs go to the build/ directory and include all necessary CPAN metadata.
`,
	}

	return workflows
}

// SuggestNextSteps provides contextual suggestions for what to do next
func (h *HelpManager) SuggestNextSteps() []string {
	suggestions := []string{}

	if h.projectContext == nil || !h.projectContext.IsProject {
		suggestions = append(suggestions,
			"Initialize a new project: pvm project init",
			"Check available Perl versions: pvm available",
			"Install a Perl version: pvm install 5.38.0",
		)
		return suggestions
	}

	// We're in a project - suggest project-specific next steps
	if !h.projectContext.HasCpanfile {
		suggestions = append(suggestions,
			"Create a cpanfile for dependency management: touch cpanfile",
			"Add a dependency: pvm module add DBI",
		)
	} else {
		suggestions = append(suggestions,
			"Install dependencies: pvm module install",
			"Add a new dependency: pvm module add Module::Name",
		)
	}

	// Check if project has tests
	if hasTestDirectory() {
		suggestions = append(suggestions,
			"Run tests: pvm test",
			"Start development mode: pvm dev",
		)
	} else {
		suggestions = append(suggestions,
			"Create test directory: mkdir -p t",
			"Start development mode: pvm dev",
		)
	}

	suggestions = append(suggestions,
		"Check project status: pvm project status",
		"Build project: pvm build",
	)

	return suggestions
}

// hasTestDirectory checks if the project has a test directory
func hasTestDirectory() bool {
	if _, err := os.Stat("t"); err == nil {
		return true
	}
	if _, err := os.Stat("test"); err == nil {
		return true
	}
	return false
}

// CreateHelpCommand creates the enhanced help command
func CreateHelpCommand() *cobra.Command {
	helpCmd := &cobra.Command{
		Use:   "help [topic]",
		Short: "Context-aware help and command suggestions",
		Long: `Enhanced help system that provides context-aware suggestions and workflow guidance.

Available topics:
  workflows     - Common development workflows
  getting-started - New user onboarding
  troubleshooting - Diagnostic and problem-solving commands
  next          - Suggested next steps based on current context

Without a topic, shows contextual help based on your current project state.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			helpManager := NewHelpManager()

			if len(args) == 0 {
				return showContextualHelp(cmd, helpManager)
			}

			topic := args[0]
			switch topic {
			case "workflows":
				return showWorkflowHelp(cmd, helpManager)
			case "getting-started":
				return showGettingStartedHelp(cmd, helpManager)
			case "troubleshooting":
				return showTroubleshootingHelp(cmd, helpManager)
			case "next":
				return showNextStepsHelp(cmd, helpManager)
			default:
				// Fall back to default cobra help behavior for specific commands
				if rootCmd := cmd.Root(); rootCmd != nil {
					if targetCmd, _, err := rootCmd.Find(args); err == nil {
						return targetCmd.Help()
					}
				}
				return fmt.Errorf("unknown help topic: %s\n\nAvailable topics: workflows, getting-started, troubleshooting, next", topic)
			}
		},
	}

	return helpCmd
}

// showContextualHelp displays help content based on current context
func showContextualHelp(cmd *cobra.Command, helpManager *HelpManager) error {
	categories := helpManager.GetContextualHelp()

	// Show project context
	if helpManager.projectContext != nil && helpManager.projectContext.IsProject {
		cmd.Printf("📁 Project Context: %s (detected via %s)\n\n",
			helpManager.projectContext.RootDir,
			helpManager.projectContext.DetectionInfo)
	} else {
		cmd.Println("📁 Context: Not in a Perl project directory")
		cmd.Println()
	}

	// Show each category
	for _, category := range categories {
		cmd.Printf("## %s\n", category.Name)
		cmd.Printf("%s\n\n", category.Description)

		for _, suggestion := range category.Commands {
			cmd.Printf("  %-25s %s\n", suggestion.Command, suggestion.Description)
			if suggestion.Relevance != "" {
				cmd.Printf("  %-25s └─ %s\n", "", suggestion.Relevance)
			}
		}
		cmd.Println()
	}

	// Show next steps
	cmd.Println("💡 Suggested next steps:")
	suggestions := helpManager.SuggestNextSteps()
	for _, suggestion := range suggestions {
		cmd.Printf("   • %s\n", suggestion)
	}

	cmd.Printf("\nFor more detailed help: pvm help workflows\n")
	cmd.Printf("For command-specific help: pvm [command] --help\n")

	return nil
}

// showWorkflowHelp displays workflow-specific help
func showWorkflowHelp(cmd *cobra.Command, helpManager *HelpManager) error {
	workflows := helpManager.GetWorkflowHelp()

	cmd.Println("# Common PVM Workflows")
	cmd.Println()

	// Order workflows logically
	workflowOrder := []string{
		"new-project",
		"existing-project",
		"module-development",
		"testing",
		"building",
	}

	for _, workflowKey := range workflowOrder {
		if content, exists := workflows[workflowKey]; exists {
			cmd.Printf("## %s\n", strings.Title(strings.ReplaceAll(workflowKey, "-", " ")))
			cmd.Print(content)
		}
	}

	return nil
}

// showGettingStartedHelp displays new user onboarding help
func showGettingStartedHelp(cmd *cobra.Command, helpManager *HelpManager) error {
	cmd.Print(`# Getting Started with PVM

PVM (Perl Version Manager) provides modern tooling for Perl development with TypeScript-quality developer experience.

## First Time Setup

1. Verify installation:
   pvm version

2. Install a Perl version:
   pvm available          # See available versions
   pvm install 5.38.0     # Install modern Perl

3. Set up shell integration:
   pvm init               # Follow the instructions

## Create Your First Project

1. Initialize a new project:
   pvm project init my-app
   cd my-app

2. Add dependencies:
   pvm module add DBI
   pvm module add Test::More --dev

3. Install dependencies:
   pvm module install

4. Start development:
   pvm dev

## Key Concepts

- **Project Context**: PVM automatically detects projects via .perl-version, cpanfile, or pvm.toml
- **Module Management**: Use 'pvm module' commands for dependency management
- **Build System**: 'pvm build' provides type checking and distribution creation
- **Development Mode**: 'pvm dev' watches files and provides instant feedback

## Need Help?

- Check project status: pvm project status
- Get contextual help: pvm help
- See workflows: pvm help workflows
- Diagnose issues: pvm project doctor

`)

	return nil
}

// showTroubleshootingHelp displays troubleshooting help
func showTroubleshootingHelp(cmd *cobra.Command, helpManager *HelpManager) error {
	cmd.Print(`# Troubleshooting PVM Issues

## Common Issues and Solutions

### Project Detection Issues
- Problem: PVM doesn't recognize your project
- Solution: Add a .perl-version file or cpanfile to your project root
- Command: echo "5.38.0" > .perl-version

### Module Installation Issues
- Problem: Modules fail to install
- Solution: Check if you're in a project and have proper permissions
- Commands: pvm project status, pvm project doctor

### Build Issues
- Problem: Build fails with type errors
- Solution: Check PSC configuration and fix type annotations
- Commands: pvm build --check-only, pvm project doctor

### Environment Issues
- Problem: Wrong Perl version being used
- Solution: Check version resolution and shell integration
- Commands: pvm resolve, pvm shell setup

## Diagnostic Commands

Check overall project health:
  pvm project doctor

View detailed project status:
  pvm project status --json

Check dependency status:
  pvm module status

Verify Perl version resolution:
  pvm resolve

## Getting Help

If you're still having issues:

1. Check the project status: pvm project status
2. Run the doctor: pvm project doctor --fix
3. Check verbose output: pvm --verbose [command]
4. Report issues at: https://github.com/anthropics/claude-code/issues
`)

	return nil
}

// showNextStepsHelp displays suggested next steps
func showNextStepsHelp(cmd *cobra.Command, helpManager *HelpManager) error {
	suggestions := helpManager.SuggestNextSteps()

	cmd.Println("💡 Suggested next steps based on your current context:")
	cmd.Println()

	for i, suggestion := range suggestions {
		cmd.Printf("%d. %s\n", i+1, suggestion)
	}

	cmd.Println("\nFor more guidance:")
	cmd.Println("  pvm help workflows     # See common development workflows")
	cmd.Println("  pvm help getting-started # New user guide")
	cmd.Println("  pvm project status     # Check current project state")

	return nil
}

// SuggestCommand provides "did you mean?" functionality for typos
func SuggestCommand(invalidCommand string, availableCommands []string) []string {
	suggestions := []string{}

	// Simple fuzzy matching based on string similarity
	for _, cmd := range availableCommands {
		if similarity := calculateSimilarity(invalidCommand, cmd); similarity > 0.6 {
			suggestions = append(suggestions, cmd)
		}
	}

	// Sort by similarity (most similar first)
	sort.Slice(suggestions, func(i, j int) bool {
		return calculateSimilarity(invalidCommand, suggestions[i]) >
			calculateSimilarity(invalidCommand, suggestions[j])
	})

	// Limit to top 3 suggestions
	if len(suggestions) > 3 {
		suggestions = suggestions[:3]
	}

	return suggestions
}

// calculateSimilarity calculates string similarity using simple heuristics
func calculateSimilarity(s1, s2 string) float64 {
	// Handle empty strings first
	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	if s1 == s2 {
		return 1.0
	}

	// Simple edit distance approximation
	longer := s1
	shorter := s2
	if len(s2) > len(s1) {
		longer = s2
		shorter = s1
	}

	// Count common characters at same positions
	common := 0
	for i := 0; i < len(shorter); i++ {
		if i < len(longer) && shorter[i] == longer[i] {
			common++
		}
	}

	// Also count common characters regardless of position
	charCount1 := make(map[rune]int)
	charCount2 := make(map[rune]int)

	for _, r := range s1 {
		charCount1[r]++
	}
	for _, r := range s2 {
		charCount2[r]++
	}

	commonChars := 0
	for char, count1 := range charCount1 {
		if count2, exists := charCount2[char]; exists {
			if count1 < count2 {
				commonChars += count1
			} else {
				commonChars += count2
			}
		}
	}

	// Combine positional and character similarity
	positionalSimilarity := float64(common) / float64(len(longer))
	characterSimilarity := float64(commonChars) / float64(len(longer))

	// Weight positional matches more heavily
	return (positionalSimilarity * 0.7) + (characterSimilarity * 0.3)
}
