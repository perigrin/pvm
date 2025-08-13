// ABOUTME: Enhanced help system with context-aware suggestions and workflow guidance
// ABOUTME: Provides intelligent help content based on project state and user context

package cli

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/cli/docs"
	"tamarou.com/pvm/internal/cli/ui"
	"tamarou.com/pvm/internal/project"
)

// HelpManager provides enhanced help functionality
type HelpManager struct {
	projectContext *project.ProjectContext
	docManager     docs.DocumentManager
}

// NewHelpManager creates a new help manager with current project context
func NewHelpManager() *HelpManager {
	projectCtx, _ := project.GetCurrentProject()
	docManager, _ := docs.NewDocumentManager()
	return &HelpManager{
		projectContext: projectCtx,
		docManager:     docManager,
	}
}

// GetProjectContext returns the current project context
func (h *HelpManager) GetProjectContext() *project.ProjectContext {
	return h.projectContext
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
		categories = append(categories, h.getWorkspaceWorkflowCategory())
		categories = append(categories, h.getBuildAndTestCategory())
		categories = append(categories, h.getModuleManagementCategory())
	} else {
		categories = append(categories, h.getWorkspaceSetupCategory())
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
			Command:     "pvm workspace status",
			Description: "Show current workspace status",
			Example:     "pvm workspace status",
			Relevance:   "Check workspace health and configuration",
		})
	} else {
		commands = append(commands, CommandSuggestion{
			Command:     "pvm workspace init",
			Description: "Initialize a new Perl workspace",
			Example:     "pvm workspace init my-app",
			Relevance:   "Start a new workspace with PVM",
		})
	}

	return HelpCategory{
		Name:        "Getting Started",
		Description: "Essential commands to get you started with PVM",
		Commands:    commands,
	}
}

// getWorkspaceWorkflowCategory provides workspace-specific workflow commands
func (h *HelpManager) getWorkspaceWorkflowCategory() HelpCategory {
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
		Description: "Commands for working within your Perl workspace",
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

// getWorkspaceSetupCategory provides workspace setup commands for non-workspace directories
func (h *HelpManager) getWorkspaceSetupCategory() HelpCategory {
	commands := []CommandSuggestion{
		{
			Command:     "pvm workspace init",
			Description: "Initialize a new Perl workspace in current directory",
			Example:     "pvm workspace init",
			Relevance:   "You're not in a workspace directory",
		},
		{
			Command:     "pvm workspace init my-app",
			Description: "Create a new workspace directory",
			Example:     "pvm workspace init my-app",
			Relevance:   "Start a new workspace from scratch",
		},
	}

	return HelpCategory{
		Name:        "Project Setup",
		Description: "Commands for setting up new Perl workspaces",
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
			Command:     "pvm workspace doctor",
			Description: "Check workspace health and suggest fixes",
			Example:     "pvm workspace doctor",
			Relevance:   "Diagnose common issues",
		},
		{
			Command:     "pvm workspace doctor --fix",
			Description: "Automatically fix common workspace issues",
			Example:     "pvm workspace doctor --fix",
			Relevance:   "Resolve problems automatically",
		},
	}

	if h.projectContext != nil && h.projectContext.IsProject {
		commands = append(commands, CommandSuggestion{
			Command:     "pvm workspace status --json",
			Description: "Get detailed workspace status in JSON format",
			Example:     "pvm workspace status --json",
			Relevance:   "Programmatic access to workspace information",
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
   pvm workspace init my-app
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

1. Check workspace status:
   pvm workspace status

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
   pvm workspace init --template=module My::Module

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
			"Initialize a new workspace: pvm workspace init",
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
		// Has cpanfile - check if dependencies are already installed
		if hasLocalLib() {
			suggestions = append(suggestions,
				"Sync dependencies to latest versions: pvm module sync",
				"Alternative sync command: pm sync",
				"Add a new dependency: pvm module add Module::Name",
			)
		} else {
			suggestions = append(suggestions,
				"Install dependencies: pvm module install",
				"Add a new dependency: pvm module add Module::Name",
			)
		}
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
		"Check workspace status: pvm workspace status",
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

// hasLocalLib checks if the project has a local-lib installation
func hasLocalLib() bool {
	// Check for local/ directory with lib/ subdirectory (typical local-lib structure)
	if _, err := os.Stat("local/lib"); err == nil {
		return true
	}
	// Also check for .carton/local which is another common structure
	if _, err := os.Stat(".carton/local/lib"); err == nil {
		return true
	}
	return false
}

// GetDocumentationHelp returns help content for embedded documentation
func (h *HelpManager) GetDocumentationHelp() []HelpCategory {
	if !h.docManager.IsAvailable() {
		return []HelpCategory{
			{
				Name:        "Documentation (Offline)",
				Description: "Documentation is not available offline in development builds",
				Commands: []CommandSuggestion{
					{
						Command:     "Online documentation",
						Description: "View documentation online at GitHub",
						Example:     "https://github.com/perigrin/pvm/blob/pu/docs/",
						Relevance:   "Full documentation available online",
					},
				},
			},
		}
	}

	categories := h.docManager.GetCategories()
	helpCategories := make([]HelpCategory, 0, len(categories))

	for _, category := range categories {
		docs, err := h.docManager.GetDocumentsByCategory(category)
		if err != nil {
			continue
		}

		commands := make([]CommandSuggestion, 0, len(docs))
		for _, doc := range docs {
			commands = append(commands, CommandSuggestion{
				Command:     fmt.Sprintf("pvm help docs %s", doc.Name),
				Description: doc.Description,
				Example:     fmt.Sprintf("pvm help docs %s", doc.Name),
				Relevance:   fmt.Sprintf("View %s documentation", doc.Name),
			})
		}

		helpCategories = append(helpCategories, HelpCategory{
			Name:        fmt.Sprintf("Documentation - %s", category),
			Description: fmt.Sprintf("%s documentation topics", category),
			Commands:    commands,
		})
	}

	return helpCategories
}

// ShowDocument displays a specific document
func (h *HelpManager) ShowDocument(name string) error {
	if !h.docManager.IsAvailable() {
		fmt.Printf("Documentation is not available offline.\n")
		fmt.Printf("View online: https://github.com/perigrin/pvm/blob/pu/docs/%s.md\n", name)
		return nil
	}

	content, err := h.docManager.GetDocument(name)
	if err != nil {
		// Try to find similar documents
		docs, listErr := h.docManager.ListDocuments()
		if listErr == nil {
			var suggestions []string
			for _, doc := range docs {
				if strings.Contains(strings.ToLower(doc.Name), strings.ToLower(name)) ||
					strings.Contains(strings.ToLower(name), strings.ToLower(doc.Name)) {
					suggestions = append(suggestions, doc.Name)
				}
			}
			if len(suggestions) > 0 {
				fmt.Printf("Document '%s' not found. Did you mean:\n", name)
				for _, suggestion := range suggestions {
					fmt.Printf("  pvm help docs %s\n", suggestion)
				}
				return nil
			}
		}
		return fmt.Errorf("document not found: %s", name)
	}

	// Display the document content with glow formatting
	displayName := strings.ReplaceAll(name, "-", " ")
	if len(displayName) > 0 {
		displayName = strings.ToUpper(displayName[:1]) + strings.ToLower(displayName[1:])
	}

	// Create UI instance for enhanced markdown display
	uiOutput := ui.NewDefaultOutput()

	// Create properly formatted markdown content
	markdownContent := fmt.Sprintf("# %s\n\n%s", displayName, string(content))

	// Use glow to render the documentation
	uiOutput.GlowMarkdown(markdownContent)

	return nil
}

// SearchDocuments searches for documents containing the query
func (h *HelpManager) SearchDocuments(query string) error {
	if !h.docManager.IsAvailable() {
		fmt.Printf("Documentation search is not available offline.\n")
		fmt.Printf("Search online: https://github.com/perigrin/pvm/search?q=%s&type=wiki\n", query)
		return nil
	}

	results, err := h.docManager.SearchDocuments(query)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(results) == 0 {
		fmt.Printf("No documents found matching '%s'\n", query)
		return nil
	}

	fmt.Printf("Found %d document(s) matching '%s':\n\n", len(results), query)
	for i, result := range results {
		if i >= 10 { // Limit results
			fmt.Printf("... and %d more results\n", len(results)-i)
			break
		}

		fmt.Printf("## %s (%s)\n", result.Document.Name, result.Document.Category)
		fmt.Printf("%s\n", result.Document.Description)
		fmt.Printf("Command: pvm help docs %s\n", result.Document.Name)

		if len(result.Matches) > 0 {
			fmt.Printf("Matches:\n")
			for j, match := range result.Matches {
				if j >= 3 { // Limit matches per document
					break
				}
				fmt.Printf("  • %s\n", match)
			}
		}
		fmt.Println()
	}

	return nil
}

// ListDocuments lists all available documents
func (h *HelpManager) ListDocuments() error {
	if !h.docManager.IsAvailable() {
		fmt.Printf("Documentation is not available offline.\n")
		fmt.Printf("View online: https://github.com/perigrin/pvm/blob/pu/docs/\n")
		return nil
	}

	docs, err := h.docManager.ListDocuments()
	if err != nil {
		return fmt.Errorf("failed to list documents: %w", err)
	}

	if len(docs) == 0 {
		fmt.Printf("No documentation available.\n")
		return nil
	}

	fmt.Printf("Available Documentation (%d documents):\n\n", len(docs))

	currentCategory := ""
	for _, doc := range docs {
		if doc.Category != currentCategory {
			currentCategory = doc.Category
			fmt.Printf("## %s\n", currentCategory)
		}
		fmt.Printf("  %-25s %s\n", doc.Name, doc.Description)
		fmt.Printf("  %-25s Command: pvm help docs %s\n", "", doc.Name)
		fmt.Println()
	}

	fmt.Printf("Use 'pvm help docs <name>' to view specific documentation.\n")
	fmt.Printf("Use 'pvm help docs search <query>' to search documentation.\n")

	return nil
}

// CreateHelpCommand creates the enhanced help command
func CreateHelpCommand() *cobra.Command {
	helpCmd := &cobra.Command{
		Use:   "guide [topic]",
		Short: "Development workflows and guidance",
		Long: `Enhanced help system that provides context-aware suggestions and workflow guidance.

Available topics:
  workflows     - Common development workflows
  getting-started - New user onboarding
  troubleshooting - Diagnostic and problem-solving commands
  next          - Suggested next steps based on current context
  docs          - List embedded documentation (if available)
  docs <name>   - View specific documentation
  docs search <query> - Search documentation

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
			case "docs":
				return showDocumentationHelp(cmd, helpManager, args[1:])
			default:
				// Fall back to default cobra help behavior for specific commands
				if rootCmd := cmd.Root(); rootCmd != nil {
					if targetCmd, _, err := rootCmd.Find(args); err == nil {
						return targetCmd.Help()
					}
				}
				ui := GetUI(cmd)
				ui.Error("unknown help topic: %s", topic)
				ui.Println("")
				ui.Info("Available topics: workflows, getting-started, troubleshooting, next, docs")
				return fmt.Errorf("unknown help topic: %s", topic)
			}
		},
	}

	return helpCmd
}

// ShowContextualHelpWithPager captures help output and uses smart paging
func ShowContextualHelpWithPager(cmd *cobra.Command, helpManager *HelpManager) error {
	// Check if we can use pager before capturing output
	canUsePager := isTerminal(os.Stdout) && os.Getenv("PAGER") != "cat" && os.Getenv("NO_PAGER") == ""

	if !canUsePager {
		// Just use regular help output if pager is not available
		return showContextualHelp(cmd, helpManager)
	}

	// Capture the output in a buffer
	var buf bytes.Buffer

	// Create a UI that writes to our buffer
	ctx := &ui.UIContext{
		Writer:      &buf,
		ErrorWriter: os.Stderr, // Keep errors going to stderr
		ColorMode:   ui.ColorAuto,
		Quiet:       false, // Don't suppress output when capturing
		Verbose:     false,
		Interactive: true,
		RawMarkdown: RawMarkdown,
	}
	bufferUI := ui.NewOutput(ctx)

	// Generate help content to buffer
	err := generateContextualHelp(cmd, bufferUI, helpManager)
	if err != nil {
		return err
	}

	// Count lines and decide on pager
	content := buf.String()
	lines := strings.Split(content, "\n")
	lineCount := len(lines)
	terminalHeight := getTerminalHeight()

	if terminalHeight > 0 && lineCount > (terminalHeight-3) {
		// Use pager for long content
		PrintWithPager(content)
		return nil
	}

	// Content fits in terminal, output directly
	fmt.Print(content)
	return nil
}

// showContextualHelp displays help content based on current context
func showContextualHelp(cmd *cobra.Command, helpManager *HelpManager) error {
	ui := GetUI(cmd)
	return generateContextualHelp(cmd, ui, helpManager)
}

// generateContextualHelp generates the help content using the provided UI
func generateContextualHelp(cmd *cobra.Command, ui *ui.Output, helpManager *HelpManager) error {
	categories := helpManager.GetContextualHelp()

	// Show project context
	if helpManager.projectContext != nil && helpManager.projectContext.IsProject {
		ui.Info("📁 Project Context: %s (detected via %s)",
			helpManager.projectContext.RootDir,
			helpManager.projectContext.DetectionInfo)
		ui.Println("")
	} else {
		ui.Info("📁 Context: Not in a Perl workspace directory")
		ui.Println("")
	}

	// Show each category
	for _, category := range categories {
		ui.SubHeader(category.Name)
		ui.Println(category.Description)
		ui.Println("")

		// Format commands as a structured list
		commandItems := make([]string, 0, len(category.Commands))
		for _, suggestion := range category.Commands {
			commandLine := fmt.Sprintf("%-25s %s", suggestion.Command, suggestion.Description)
			if suggestion.Relevance != "" {
				commandLine += "\n" + fmt.Sprintf("%25s └─ %s", "", suggestion.Relevance)
			}
			commandItems = append(commandItems, commandLine)
		}
		ui.List(commandItems)
	}

	// Show next steps
	ui.SubHeader("💡 Suggested next steps")
	suggestions := helpManager.SuggestNextSteps()
	ui.List(suggestions)

	ui.Info("For more detailed help: pvm help workflows")
	ui.Info("For command-specific help: pvm [command] --help")

	return nil
}

// showWorkflowHelp displays workflow-specific help
func showWorkflowHelp(cmd *cobra.Command, helpManager *HelpManager) error {
	ui := GetUI(cmd)
	workflows := helpManager.GetWorkflowHelp()

	ui.Header("Common PVM Workflows")
	ui.Println("")

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
			title := strings.Title(strings.ReplaceAll(workflowKey, "-", " "))
			ui.Section(title, strings.TrimSpace(content))
		}
	}

	return nil
}

// showGettingStartedHelp displays new user onboarding help
func showGettingStartedHelp(cmd *cobra.Command, helpManager *HelpManager) error {
	ui := GetUI(cmd)

	ui.Header("Getting Started with PVM")
	ui.Println("")
	ui.Println("PVM (Perl Version Manager) provides modern tooling for Perl development with TypeScript-quality developer experience.")
	ui.Println("")

	ui.SubHeader("First Time Setup")
	setupSteps := []string{
		"Verify installation: pvm version",
		"Install a Perl version:\n   pvm available          # See available versions\n   pvm install 5.38.0     # Install modern Perl",
		"Set up shell integration: pvm init               # Follow the instructions",
	}
	for i, step := range setupSteps {
		ui.Printf("%d. %s\n", i+1, step)
	}

	ui.SubHeader("Create Your First Project")
	projectSteps := []string{
		"Initialize a new workspace:\n   pvm workspace init my-app\n   cd my-app",
		"Add dependencies:\n   pvm module add DBI\n   pvm module add Test::More --dev",
		"Install dependencies: pvm module install",
		"Start development: pvm dev",
	}
	for i, step := range projectSteps {
		ui.Printf("%d. %s\n", i+1, step)
	}

	ui.SubHeader("Key Concepts")
	concepts := []string{
		"**Project Context**: PVM automatically detects projects via .perl-version, cpanfile, or pvm.toml",
		"**Module Management**: Use 'pvm module' commands for dependency management",
		"**Build System**: 'pvm build' provides type checking and distribution creation",
		"**Development Mode**: 'pvm dev' watches files and provides instant feedback",
	}
	ui.List(concepts)

	ui.SubHeader("Need Help?")
	helpCommands := []string{
		"Check workspace status: pvm workspace status",
		"Get contextual help: pvm help",
		"See workflows: pvm help workflows",
		"Diagnose issues: pvm workspace doctor",
	}
	ui.List(helpCommands)

	return nil
}

// showTroubleshootingHelp displays troubleshooting help
func showTroubleshootingHelp(cmd *cobra.Command, helpManager *HelpManager) error {
	ui := GetUI(cmd)

	ui.Header("Troubleshooting PVM Issues")
	ui.Println("")

	ui.SubHeader("Common Issues and Solutions")
	ui.Println("")

	// Project Detection Issues
	ui.Warning("Project Detection Issues")
	ui.Info("Problem: PVM doesn't recognize your project")
	ui.Success("Solution: Add a .perl-version file or cpanfile to your project root")
	ui.Printf("Command: echo \"5.38.0\" > .perl-version\n")
	ui.Println("")

	// Module Installation Issues
	ui.Warning("Module Installation Issues")
	ui.Info("Problem: Modules fail to install")
	ui.Success("Solution: Check if you're in a project and have proper permissions")
	ui.Printf("Commands: pvm workspace status, pvm workspace doctor\n")
	ui.Println("")

	// Build Issues
	ui.Warning("Build Issues")
	ui.Info("Problem: Build fails with type errors")
	ui.Success("Solution: Check PSC configuration and fix type annotations")
	ui.Printf("Commands: pvm build --check-only, pvm workspace doctor\n")
	ui.Println("")

	// Environment Issues
	ui.Warning("Environment Issues")
	ui.Info("Problem: Wrong Perl version being used")
	ui.Success("Solution: Check version resolution and shell integration")
	ui.Printf("Commands: pvm resolve, pvm shell setup\n")
	ui.Println("")

	ui.SubHeader("Diagnostic Commands")
	diagnosticCommands := []string{
		"Check overall workspace health: pvm workspace doctor",
		"View detailed workspace status: pvm workspace status --json",
		"Check dependency status: pvm module status",
		"Verify Perl version resolution: pvm resolve",
	}
	ui.List(diagnosticCommands)

	ui.SubHeader("Getting Help")
	ui.Println("If you're still having issues:")
	ui.Println("")
	helpSteps := []string{
		"Check the workspace status: pvm workspace status",
		"Run the doctor: pvm workspace doctor --fix",
		"Check verbose output: pvm --verbose [command]",
		"Report issues at: https://github.com/anthropics/claude-code/issues",
	}
	for i, step := range helpSteps {
		ui.Printf("%d. %s\n", i+1, step)
	}

	return nil
}

// showNextStepsHelp displays suggested next steps
func showNextStepsHelp(cmd *cobra.Command, helpManager *HelpManager) error {
	ui := GetUI(cmd)
	suggestions := helpManager.SuggestNextSteps()

	ui.Header("💡 Suggested next steps based on your current context")
	ui.Println("")

	for i, suggestion := range suggestions {
		ui.Printf("%d. %s\n", i+1, suggestion)
	}

	ui.SubHeader("For more guidance")
	moreHelp := []string{
		"pvm help workflows     # See common development workflows",
		"pvm help getting-started # New user guide",
		"pvm workspace status     # Check current project state",
	}
	ui.List(moreHelp)

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

// showDocumentationHelp handles documentation-related help commands
func showDocumentationHelp(cmd *cobra.Command, helpManager *HelpManager, args []string) error {
	if len(args) == 0 {
		// List all documentation
		return helpManager.ListDocuments()
	}

	subCommand := args[0]
	switch subCommand {
	case "search":
		if len(args) < 2 {
			fmt.Printf("Usage: pvm help docs search <query>\n")
			return fmt.Errorf("search query required")
		}
		query := strings.Join(args[1:], " ")
		return helpManager.SearchDocuments(query)
	default:
		// Treat as document name
		docName := subCommand
		return helpManager.ShowDocument(docName)
	}
}

// ShowHybridHelp displays a concise command listing with pointers to detailed help topics
// This provides the best of both worlds: brevity of cobra's default help with access to rich help content
func ShowHybridHelp(cmd *cobra.Command, args []string) {
	ui := GetUI(cmd)

	// Help should always be shown, even in quiet mode, since user explicitly requested it
	originalQuiet := ui.Context().Quiet
	if originalQuiet {
		ui.SetQuiet(false)
		defer ui.SetQuiet(true)
	}

	// Show basic command information with Fang UI styling
	ui.Header(fmt.Sprintf("%s - %s", cmd.Name(), cmd.Short))
	ui.Println("")

	if cmd.Long != "" && cmd.Long != cmd.Short {
		ui.Println(cmd.Long)
		ui.Println("")
	}

	// Show usage
	ui.SubHeader("USAGE")
	ui.Printf("  %s\n", cmd.UseLine())
	ui.Println("")

	// Show available commands in a concise format
	commands := cmd.Commands()
	if len(commands) > 0 {
		ui.SubHeader("COMMANDS")

		// Group commands for better organization
		coreCommands := []string{}
		managementCommands := []string{}
		utilityCommands := []string{}

		for _, subCmd := range commands {
			if subCmd.Hidden {
				continue
			}

			cmdLine := fmt.Sprintf("  %-30s %s", subCmd.Use, subCmd.Short)

			// Categorize commands based on their names/functions
			cmdName := subCmd.Name()
			switch {
			case cmdName == "install" || cmdName == "uninstall" || cmdName == "versions" ||
				cmdName == "available" || cmdName == "current":
				coreCommands = append(coreCommands, cmdLine)
			case cmdName == "build" || cmdName == "test" || cmdName == "dev" ||
				cmdName == "module" || cmdName == "workspace":
				managementCommands = append(managementCommands, cmdLine)
			default:
				utilityCommands = append(utilityCommands, cmdLine)
			}
		}

		// Print core commands first
		for _, cmdLine := range coreCommands {
			ui.Printf("%s\n", cmdLine)
		}
		for _, cmdLine := range managementCommands {
			ui.Printf("%s\n", cmdLine)
		}
		for _, cmdLine := range utilityCommands {
			ui.Printf("%s\n", cmdLine)
		}
		ui.Println("")
	}

	// Show global flags
	if cmd.HasAvailableFlags() {
		ui.SubHeader("FLAGS")
		flagUsages := cmd.Flags().FlagUsages()
		if flagUsages != "" {
			ui.Printf("%s", flagUsages)
		}
		ui.Println("")
	}

	// Add context-dependent pointers to detailed help
	showContextualHelpPointers(ui)
}

// showContextualHelpPointers displays context-aware help pointers
func showContextualHelpPointers(ui *ui.Output) {
	// Help should always be shown, even in quiet mode, since user explicitly requested it
	originalQuiet := ui.Context().Quiet
	if originalQuiet {
		ui.SetQuiet(false)
		defer ui.SetQuiet(true)
	}

	// Detect project context
	projectCtx, _ := project.GetCurrentProject()
	isInProject := projectCtx != nil && projectCtx.IsProject

	ui.SubHeader("DETAILED HELP")

	if isInProject {
		// Inside a project - show project-focused help
		ui.Info("For project workflows:")
		ui.Printf("  pvm help next             Your next steps for this project\n")
		ui.Printf("  pvm dev --help            Development environment guide\n")
		ui.Printf("  pvm help workflows        Common development workflows\n")
		ui.Printf("  pvm help troubleshooting  Fix project issues\n")
		ui.Println("")

		ui.Info("For general guidance:")
		ui.Printf("  pvm help                  Show contextual help based on your project\n")
		ui.Printf("  pvm help getting-started  New user onboarding guide\n")
	} else {
		// Outside a project - show setup-focused help
		ui.Info("For getting started:")
		ui.Printf("  pvm help getting-started  New user onboarding guide\n")
		ui.Printf("  pvm workspace init --help Create a new workspace\n")
		ui.Printf("  pvm help next             Suggested next steps\n")
		ui.Println("")

		ui.Info("For advanced usage:")
		ui.Printf("  pvm help workflows        Development workflows\n")
		ui.Printf("  pvm help troubleshooting  Diagnostic commands\n")
	}

	ui.Printf("\n")
	ui.Info("For command-specific help:")
	ui.Printf("  pvm [command] --help      Show detailed help for any command\n")
}
