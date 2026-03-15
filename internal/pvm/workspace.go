// ABOUTME: Project management commands for PVM
// ABOUTME: Handles project initialization, status, and configuration

package pvm

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/perl"
	"tamarou.com/pvm/internal/project"
	"tamarou.com/pvm/internal/xdg"
)

// ProjectTemplate represents a project template configuration
type ProjectTemplate struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Directories  []string          `json:"directories"`
	Dependencies map[string]string `json:"dependencies"`
	DevDeps      map[string]string `json:"dev_dependencies"`
	TestDeps     map[string]string `json:"test_dependencies"`
	Config       map[string]any    `json:"config"`
	GitIgnore    []string          `json:"gitignore_additions"`
}

// HealthStatus represents the health status of a project component
type HealthStatus string

const (
	HealthStatusHealthy  HealthStatus = "healthy"
	HealthStatusWarning  HealthStatus = "warning"
	HealthStatusCritical HealthStatus = "critical"
	HealthStatusUnknown  HealthStatus = "unknown"
)

// HealthCheck represents a single health check result
type HealthCheck struct {
	Name       string       `json:"name"`
	Status     HealthStatus `json:"status"`
	Message    string       `json:"message"`
	Details    string       `json:"details,omitempty"`
	Suggestion string       `json:"suggestion,omitempty"`
	CheckedAt  time.Time    `json:"checked_at"`
}

// WorkspaceHealth represents the overall project health
type WorkspaceHealth struct {
	OverallStatus HealthStatus  `json:"overall_status"`
	Checks        []HealthCheck `json:"checks"`
	Summary       string        `json:"summary"`
	NextSteps     []string      `json:"next_steps"`
	CheckedAt     time.Time     `json:"checked_at"`
}

// newWorkspaceCommand creates the main workspace command with subcommands
func newWorkspaceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "workspace",
		Short:   "Workspace management commands",
		Long:    "Commands for initializing, managing, and working with Perl workspaces",
		Aliases: []string{"ws"},
	}

	// Add workspace subcommands
	cmd.AddCommand(
		newWorkspaceInitCommand(),
		newWorkspaceStatusCommand(),
		newWorkspaceDoctorCommand(),
		newWorkspaceTemplatesCommand(),
		newWorkspaceValidateCommand(),
	)

	return cmd
}

// newWorkspaceInitCommand creates the project init command
func newWorkspaceInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [project-name]",
		Short: "Initialize a new Perl project",
		Long: `Initialize a new Perl project with proper directory structure and configuration.

If project-name is provided, creates a new directory with that name.
If no name is provided, initializes in the current directory.

This command creates:
- .perl-version file with current/default Perl version
- cpanfile for dependency management
- pvm.toml for workspace configuration
- Standard directory structure (lib/, t/, script/)
- .gitignore with PVM-specific ignores`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var projectName string
			var projectDir string

			if len(args) > 0 {
				projectName = args[0]
				// Validate project name
				if err := validateProjectName(projectName); err != nil {
					return err
				}
				projectDir = projectName
			} else {
				// Use current directory
				wd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get current directory: %w", err)
				}
				projectDir = "."
				projectName = filepath.Base(wd)
			}

			// Get template from flags
			templateName, err := cmd.Flags().GetString("template")
			if err != nil {
				return err
			}

			// Load the template with project variables
			template, err := loadTemplateWithVariables(templateName, projectName)
			if err != nil {
				return fmt.Errorf("failed to load template '%s': %w", templateName, err)
			}

			// Get force flag
			force, err := cmd.Flags().GetBool("force")
			if err != nil {
				return err
			}

			return initializeProject(cmd, projectName, projectDir, &template, force)
		},
	}

	// Add flags
	cmd.Flags().String("template", "minimal", "Workspace template to use (run 'pvm workspace templates' to list available)")
	cmd.Flags().Bool("force", false, "Initialize even if directory is not empty or workspace files exist")

	return cmd
}

// newWorkspaceStatusCommand creates the project status command
func newWorkspaceStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show project status and health",
		Long: `Show comprehensive project status including:
- Project detection results
- Perl version consistency
- Dependencies status
- Build artifacts status
- Configuration health`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return showWorkspaceStatus(cmd)
		},
	}

	// Add flags
	cmd.Flags().Bool("json", false, "Output status in JSON format")

	return cmd
}

// newWorkspaceDoctorCommand creates the project doctor command
func newWorkspaceDoctorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Run comprehensive project health checks",
		Long: `Run comprehensive health checks for the project including:
- Perl version compatibility
- Dependency installation status
- Build system health
- Configuration validation
- Development environment setup
- Common issues detection`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWorkspaceDoctor(cmd)
		},
	}

	// Add flags
	cmd.Flags().Bool("json", false, "Output results in JSON format")
	cmd.Flags().Bool("fix", false, "Attempt to automatically fix issues where possible")
	cmd.Flags().Bool("verbose", false, "Show detailed information for all checks")

	return cmd
}

// newWorkspaceTemplatesCommand creates the project templates command
func newWorkspaceTemplatesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "templates",
		Short: "List available project templates",
		Long:  "List all available project templates, both built-in and user-defined",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listTemplates(cmd)
		},
	}

	return cmd
}

// validateProjectName validates that a project name is suitable
func validateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	// Check for invalid characters
	if strings.ContainsAny(name, " \t\n\r/\\:*?\"<>|") {
		return fmt.Errorf("project name contains invalid characters")
	}

	// Check if name looks like a valid Perl module name (optional validation)
	if strings.Contains(name, "::") {
		// If it contains ::, validate as module name
		parts := strings.Split(name, "::")
		for _, part := range parts {
			if part == "" || !isValidIdentifier(part) {
				return fmt.Errorf("invalid module name format: %s", name)
			}
		}
	}

	return nil
}

// isValidIdentifier checks if a string is a valid Perl identifier
func isValidIdentifier(s string) bool {
	if s == "" {
		return false
	}

	// Must start with letter or underscore
	first := s[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}

	// Rest can be letters, numbers, or underscores
	for i := 1; i < len(s); i++ {
		c := s[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}

	return true
}

// initializeProject handles the actual project initialization
func initializeProject(cmd *cobra.Command, projectName, projectDir string, template *ProjectTemplate, force bool) error {
	// Convert relative path to absolute
	absProjectDir, err := filepath.Abs(projectDir)
	if err != nil {
		return fmt.Errorf("failed to resolve project directory: %w", err)
	}

	// Check if we need to create the directory
	createDir := projectDir != "."

	if createDir {
		// Check if directory already exists
		if _, err := os.Stat(absProjectDir); err == nil {
			if !force {
				return fmt.Errorf("directory '%s' already exists (use --force to initialize anyway)", projectDir)
			}
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("failed to check directory: %w", err)
		}

		// Create the project directory
		if err := os.MkdirAll(absProjectDir, 0755); err != nil {
			return fmt.Errorf("failed to create project directory: %w", err)
		}
		ui := cli.GetUI(cmd)
		ui.Success("Created project directory: %s", projectDir)
	}

	// Check if we're already in a project (unless force is used)
	// Use DetectProjectInDirectory to check only the target directory, not parent directories
	if !force {
		existingProject, err := project.DetectProjectInDirectory(absProjectDir)
		if err != nil {
			return fmt.Errorf("failed to detect existing project: %w", err)
		}
		if existingProject.IsProject {
			return fmt.Errorf("directory is already a project (detected via %s, use --force to reinitialize)", existingProject.DetectionInfo)
		}
	}

	// Initialize project files
	if err := createProjectFiles(cmd, absProjectDir, projectName, template); err != nil {
		return fmt.Errorf("failed to create project files: %w", err)
	}

	// Create directory structure
	if err := createDirectoryStructure(cmd, absProjectDir, template); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}

	ui := cli.GetUI(cmd)
	ui.Success("Project '%s' initialized successfully!", projectName)
	ui.Info("Location: %s", absProjectDir)

	// Show next steps
	showNextSteps(cmd, createDir, projectDir)

	return nil
}

// createProjectFiles creates the necessary workspace configuration files
func createProjectFiles(cmd *cobra.Command, projectDir, projectName string, template *ProjectTemplate) error {
	ui := cli.GetUI(cmd)

	// Create .perl-version file
	perlVersion := getCurrentPerlVersion()
	perlVersionPath := filepath.Join(projectDir, ".perl-version")
	if err := os.WriteFile(perlVersionPath, []byte(perlVersion+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to create .perl-version: %w", err)
	}
	ui.Success("Created .perl-version (%s)", perlVersion)

	// Create cpanfile
	cpanfileContent := generateCpanfile(projectName, template)
	cpanfilePath := filepath.Join(projectDir, "cpanfile")
	if err := os.WriteFile(cpanfilePath, []byte(cpanfileContent), 0644); err != nil {
		return fmt.Errorf("failed to create cpanfile: %w", err)
	}
	ui.Success("Created cpanfile")

	// Create pvm.toml
	pvmTomlContent := generatePvmToml(projectName, perlVersion, template)
	pvmTomlPath := filepath.Join(projectDir, "pvm.toml")
	if err := os.WriteFile(pvmTomlPath, []byte(pvmTomlContent), 0644); err != nil {
		return fmt.Errorf("failed to create pvm.toml: %w", err)
	}
	ui.Success("Created pvm.toml")

	// Create .gitignore
	gitignoreContent := generateGitignore(template)
	gitignorePath := filepath.Join(projectDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		return fmt.Errorf("failed to create .gitignore: %w", err)
	}
	ui.Success("Created .gitignore")

	return nil
}

// createDirectoryStructure creates the standard project directory structure
func createDirectoryStructure(cmd *cobra.Command, projectDir string, template *ProjectTemplate) error {
	ui := cli.GetUI(cmd)
	dirs := template.Directories

	for _, dir := range dirs {
		dirPath := filepath.Join(projectDir, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		ui.Success("Created directory: %s/", dir)
	}

	return nil
}

// getCurrentPerlVersion gets the current Perl version to use as default
func getCurrentPerlVersion() string {
	// Try to resolve the current Perl version
	resolved, err := perl.ResolveVersion(&perl.ResolutionOptions{})
	if err == nil && resolved != nil {
		return resolved.Version
	}

	// If that fails, try to detect system Perl
	systemPerl, err := perl.DetectSystemPerl()
	if err == nil && systemPerl != nil {
		return systemPerl.Version
	}

	// Fall back to a reasonable default
	return "5.38.0"
}

// generateCpanfile generates cpanfile content based on template
func generateCpanfile(projectName string, template *ProjectTemplate) string {
	content := fmt.Sprintf(`# cpanfile for %s
requires 'perl', '5.038000';

`, projectName)

	// Add runtime dependencies
	if len(template.Dependencies) > 0 {
		content += "# Runtime dependencies\n"
		for module, version := range template.Dependencies {
			if version == "" {
				content += fmt.Sprintf("requires '%s';\n", module)
			} else {
				content += fmt.Sprintf("requires '%s', '%s';\n", module, version)
			}
		}
		content += "\n"
	}

	// Add test dependencies
	if len(template.TestDeps) > 0 {
		content += "on 'test' => sub {\n"
		for module, version := range template.TestDeps {
			if version == "" {
				content += fmt.Sprintf("    requires '%s';\n", module)
			} else {
				content += fmt.Sprintf("    requires '%s', '%s';\n", module, version)
			}
		}
		content += "};\n\n"
	}

	// Add development dependencies
	if len(template.DevDeps) > 0 {
		content += "on 'develop' => sub {\n"
		for module, version := range template.DevDeps {
			if version == "" {
				content += fmt.Sprintf("    requires '%s';\n", module)
			} else {
				content += fmt.Sprintf("    requires '%s', '%s';\n", module, version)
			}
		}
		content += "};\n"
	}

	return content
}

// generatePvmToml generates pvm.toml configuration content
func generatePvmToml(projectName, perlVersion string, template *ProjectTemplate) string {
	content := fmt.Sprintf(`[project]
name = "%s"
version = "0.01"
perl_version = "%s"

[dependencies]
cpanfile = "cpanfile"
local_lib = "local"

[build]
output_dir = "build"
`, projectName, perlVersion)

	// Add template-specific configuration
	if len(template.Config) > 0 {
		content += "\n# Template-specific configuration\n"
		// This is a simplified approach - in a real implementation
		// you'd want proper TOML generation
		for key, value := range template.Config {
			content += fmt.Sprintf("%s = %v\n", key, value)
		}
	}

	return content
}

// generateGitignore generates .gitignore content
func generateGitignore(template *ProjectTemplate) string {
	content := `# PVM specific
/build/
*.pmc
local/
.pvm/

# Perl specific
/blib/
/_build/
cover_db/
inc/
Build
!Build/
Build.bat
.last_cover_stats
/Makefile
/Makefile.old
/MANIFEST.bak
/META.yml
/META.json
/MYMETA.*
nytprof.out
/pm_to_blib
*.o
*.bs
/_eumm/

# Common files
*~
*.bak
*.orig
*.rej
.DS_Store
Thumbs.db

# IDEs
.vscode/
.idea/
*.swp
*.swo

# Testing
/cover_db/
proof/
`

	// Add template-specific gitignore entries
	if len(template.GitIgnore) > 0 {
		content += "\n# Template-specific\n"
		for _, entry := range template.GitIgnore {
			content += entry + "\n"
		}
	}

	return content
}

// showWorkspaceStatus shows comprehensive workspace status
func showWorkspaceStatus(cmd *cobra.Command) error {
	// Get workspace context
	ctx, err := project.GetCurrentProject()
	if err != nil {
		return fmt.Errorf("failed to detect workspace: %w", err)
	}

	// Check JSON output flag
	jsonOutput, err := cmd.Flags().GetBool("json")
	if err != nil {
		return err
	}

	if jsonOutput {
		return outputStatusJSON(cmd, ctx)
	}

	return outputStatusHuman(cmd, ctx)
}

// runWorkspaceDoctor runs comprehensive workspace health checks
func runWorkspaceDoctor(cmd *cobra.Command) error {
	// Get workspace context
	ctx, err := project.GetCurrentProject()
	if err != nil {
		return fmt.Errorf("failed to detect workspace: %w", err)
	}

	// Get flags
	jsonOutput, _ := cmd.Flags().GetBool("json")
	autofixEnabled, _ := cmd.Flags().GetBool("fix")
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Run health checks
	health := performHealthChecks(ctx, autofixEnabled)

	if jsonOutput {
		return outputHealthJSON(cmd, health)
	}

	return outputHealthHuman(cmd, health, verbose)
}

// outputStatusHuman outputs project status in human-readable format
func outputStatusHuman(cmd *cobra.Command, ctx *project.ProjectContext) error {
	ui := cli.GetUI(cmd)

	if !ctx.IsProject {
		ui.Info("No workspace detected in current directory")
		ui.Info("Use 'pvm workspace init' to initialize a new workspace")
		return nil
	}

	ui.Header("Workspace Status")
	ui.Println()

	// Workspace information
	ui.Info("Workspace Root: %s", ctx.RootDir)
	ui.Info("Detected via: %s", ctx.DetectionInfo)

	// Perl version
	if ctx.PerlVersion != "" {
		ui.Info("Perl Version: %s", ctx.PerlVersion)
		// Check if this version is actually installed
		if installed, err := checkPerlVersionInstalled(ctx.PerlVersion); err == nil {
			if installed {
				ui.Success("  Status: installed")
			} else {
				ui.Warning("  Status: not installed (run 'pvm install %s')", ctx.PerlVersion)
			}
		} else {
			ui.Warning("  Status: unable to check (%v)", err)
		}
	} else {
		ui.Warning("Perl Version: not specified")
	}

	// Dependencies
	if ctx.HasCpanfile {
		ui.Info("Dependencies: cpanfile present")
		// Check if dependencies are installed
		if installed, missing := checkDependenciesInstalled(ctx); installed {
			ui.Success("  Status: all dependencies installed")
		} else if missing > 0 {
			ui.Warning("  Status: %d dependencies missing (run 'pvm module install')", missing)
		} else {
			ui.Info("  Status: unable to check")
		}
	} else {
		ui.Warning("Dependencies: no cpanfile")
	}

	// Local lib
	ui.Info("Local lib: %s", ctx.LocalLibDir)
	if _, err := os.Stat(ctx.LocalLibDir); err == nil {
		ui.Info("  Status: exists")
		// Count installed modules
		moduleCount := countInstalledModules(ctx.LocalLibDir)
		if moduleCount > 0 {
			ui.Success("  Modules: %d installed", moduleCount)
		} else {
			ui.Info("  Modules: none installed")
		}
	} else {
		ui.Warning("  Status: not created")
	}

	// Configuration
	if ctx.ConfigFile != "" {
		ui.Info("Configuration: %s", ctx.ConfigFile)
	} else {
		ui.Warning("Configuration: no pvm.toml")
	}

	// Build status
	buildDir := filepath.Join(ctx.RootDir, "build")
	if _, err := os.Stat(buildDir); err == nil {
		ui.Info("Build artifacts: present")
	} else {
		ui.Info("Build artifacts: none")
	}

	// Git status
	gitDir := filepath.Join(ctx.RootDir, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		ui.Info("Git repository: yes")
	} else {
		ui.Info("Git repository: no")
	}

	ui.Println()
	ui.SubHeader("Next Steps:")
	if !ctx.HasCpanfile {
		ui.Info("- Add dependencies with 'pvm module add <module>'")
	}
	if ctx.PerlVersion == "" {
		ui.Info("- Set Perl version in .perl-version file")
	}
	ui.Info("- Install dependencies with 'pvm module install'")
	ui.Info("- Run build with 'pvm build'")

	return nil
}

// outputStatusJSON outputs project status in JSON format
func outputStatusJSON(cmd *cobra.Command, ctx *project.ProjectContext) error {
	status := map[string]interface{}{
		"is_project":     ctx.IsProject,
		"root_dir":       ctx.RootDir,
		"detection_info": ctx.DetectionInfo,
		"perl_version":   ctx.PerlVersion,
		"has_cpanfile":   ctx.HasCpanfile,
		"local_lib_dir":  ctx.LocalLibDir,
		"config_file":    ctx.ConfigFile,
		"timestamp":      time.Now(),
	}

	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	cmd.Println(string(data))
	return nil
}

// outputHealthJSON outputs health check results in JSON format
func outputHealthJSON(cmd *cobra.Command, health *WorkspaceHealth) error {
	data, err := json.MarshalIndent(health, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal health data: %w", err)
	}

	cmd.Println(string(data))
	return nil
}

// outputHealthHuman outputs health check results in human-readable format
func outputHealthHuman(cmd *cobra.Command, health *WorkspaceHealth, verbose bool) error {
	cmd.Printf("Workspace Health Check\n")
	cmd.Printf("===================\n\n")

	// Overall status
	statusColor := getStatusColor(health.OverallStatus)
	cmd.Printf("Overall Status: %s%s%s\n", statusColor, health.OverallStatus, "\033[0m")
	cmd.Printf("Summary: %s\n\n", health.Summary)

	// Individual checks
	cmd.Printf("Health Checks:\n")
	for _, check := range health.Checks {
		color := getStatusColor(check.Status)
		status := fmt.Sprintf("%s%s%s", color, check.Status, "\033[0m")
		cmd.Printf("  %-20s %s - %s\n", check.Name+":", status, check.Message)

		if verbose && check.Details != "" {
			cmd.Printf("    Details: %s\n", check.Details)
		}

		if check.Suggestion != "" {
			cmd.Printf("    Suggestion: %s\n", check.Suggestion)
		}
	}

	// Next steps
	if len(health.NextSteps) > 0 {
		cmd.Printf("\nRecommended Actions:\n")
		for i, step := range health.NextSteps {
			cmd.Printf("  %d. %s\n", i+1, step)
		}
	}

	return nil
}

// getStatusColor returns ANSI color code for status
func getStatusColor(status HealthStatus) string {
	switch status {
	case HealthStatusHealthy:
		return "\033[32m" // Green
	case HealthStatusWarning:
		return "\033[33m" // Yellow
	case HealthStatusCritical:
		return "\033[31m" // Red
	default:
		return "\033[37m" // White
	}
}

// showNextSteps shows what the user should do next after project initialization
func showNextSteps(cmd *cobra.Command, createdDir bool, projectDir string) {
	cmd.Printf("\nNext steps:\n")

	if createdDir {
		cmd.Printf("  cd %s\n", projectDir)
	}

	cmd.Printf("  pvm module install          # Install dependencies\n")
	cmd.Printf("  pvm workspace status        # Check workspace health\n")
	cmd.Printf("  pvm build                   # Build for distribution\n")
	cmd.Printf("\nLearn more:\n")
	cmd.Printf("  pvm help workspace          # Workspace management commands\n")
	cmd.Printf("  pvm help module             # Module management commands\n")
	cmd.Printf("  pvm help build              # Build system commands\n")
}

// loadTemplate loads a template by name, checking user templates first, then embedded templates
func loadTemplate(name string) (ProjectTemplate, error) {
	return loadTemplateWithVariables(name, "example")
}

// loadTemplateWithVariables loads a template with project-specific variables
func loadTemplateWithVariables(name, projectName string) (ProjectTemplate, error) {
	// First try to load from user templates
	template, err := loadUserTemplate(name)
	if err == nil {
		return template, nil
	}

	// Fall back to embedded templates with variables
	return loadEmbeddedTemplateWithVariables(name, projectName)
}

// loadEmbeddedTemplateWithVariables loads an embedded template with project-specific variables
func loadEmbeddedTemplateWithVariables(name, projectName string) (ProjectTemplate, error) {
	manager := NewEmbeddedTemplateManager()
	variables := NewTemplateVariables(projectName)

	template, err := manager.LoadEmbeddedTemplate(name, variables)
	if err != nil {
		return ProjectTemplate{}, fmt.Errorf("unknown template: %s", name)
	}

	return template, nil
}

// loadUserTemplate loads a template from the user's XDG config directory
func loadUserTemplate(name string) (ProjectTemplate, error) {
	dirs, err := xdg.GetDirs()
	if err != nil {
		return ProjectTemplate{}, err
	}

	templatePath := filepath.Join(dirs.ConfigDir, "templates", name+".json")
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return ProjectTemplate{}, err
	}

	var template ProjectTemplate
	if err := json.Unmarshal(data, &template); err != nil {
		return ProjectTemplate{}, fmt.Errorf("invalid template format: %w", err)
	}

	return template, nil
}

// getEmbeddedTemplate returns an embedded template
func getEmbeddedTemplate(name string) (ProjectTemplate, error) {
	return loadEmbeddedTemplateWithVariables(name, "example")
}

// listTemplates lists all available templates
func listTemplates(cmd *cobra.Command) error {
	cmd.Println("Available project templates:")
	cmd.Println()

	// List embedded templates
	cmd.Println("Built-in templates:")
	manager := NewEmbeddedTemplateManager()
	embeddedTemplates, err := manager.ListEmbeddedTemplates()
	if err != nil {
		return fmt.Errorf("failed to list embedded templates: %w", err)
	}

	for _, name := range embeddedTemplates {
		description, err := manager.GetEmbeddedTemplateDescription(name)
		if err == nil {
			cmd.Printf("  %-12s - %s\n", name, description)
		}
	}

	// List user templates
	userTemplates, err := getUserTemplates()
	if err == nil && len(userTemplates) > 0 {
		cmd.Println()
		cmd.Println("User templates:")
		for _, template := range userTemplates {
			cmd.Printf("  %-12s - %s\n", template.Name, template.Description)
		}
	}

	cmd.Println()
	cmd.Printf("To create a custom template, add a JSON file to:\n")

	dirs, err := xdg.GetDirs()
	if err == nil {
		cmd.Printf("%s\n", filepath.Join(dirs.ConfigDir, "templates"))
	}

	return nil
}

// getUserTemplates returns all user-defined templates
func getUserTemplates() ([]ProjectTemplate, error) {
	dirs, err := xdg.GetDirs()
	if err != nil {
		return nil, err
	}

	templatesDir := filepath.Join(dirs.ConfigDir, "templates")
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		return nil, err
	}

	var templates []ProjectTemplate
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".json")
		template, err := loadUserTemplate(name)
		if err != nil {
			continue // Skip invalid templates
		}
		templates = append(templates, template)
	}

	return templates, nil
}

// performHealthChecks runs all health checks for the project
func performHealthChecks(ctx *project.ProjectContext, autofix bool) *WorkspaceHealth {
	checks := []HealthCheck{}
	nextSteps := []string{}

	// Project detection check
	checks = append(checks, checkProjectDetection(ctx))

	// Perl version checks
	checks = append(checks, checkPerlVersion(ctx, autofix)...)

	// Dependencies checks
	checks = append(checks, checkDependencies(ctx, autofix)...)

	// Build system checks
	checks = append(checks, checkBuildSystem(ctx)...)

	// Configuration checks
	checks = append(checks, checkConfiguration(ctx, autofix)...)

	// Development environment checks
	checks = append(checks, checkDevelopmentEnvironment(ctx)...)

	// Determine overall status
	overallStatus := HealthStatusHealthy
	healthyCount := 0
	warningCount := 0
	criticalCount := 0

	for _, check := range checks {
		switch check.Status {
		case HealthStatusHealthy:
			healthyCount++
		case HealthStatusWarning:
			warningCount++
			if overallStatus == HealthStatusHealthy {
				overallStatus = HealthStatusWarning
			}
		case HealthStatusCritical:
			criticalCount++
			overallStatus = HealthStatusCritical
		}

		// Collect suggestions as next steps
		if check.Suggestion != "" {
			nextSteps = append(nextSteps, check.Suggestion)
		}
	}

	// Generate summary
	summary := fmt.Sprintf("%d checks passed, %d warnings, %d critical issues",
		healthyCount, warningCount, criticalCount)

	return &WorkspaceHealth{
		OverallStatus: overallStatus,
		Checks:        checks,
		Summary:       summary,
		NextSteps:     nextSteps,
		CheckedAt:     time.Now(),
	}
}

// checkProjectDetection verifies project detection is working
func checkProjectDetection(ctx *project.ProjectContext) HealthCheck {
	if !ctx.IsProject {
		return HealthCheck{
			Name:       "Workspace Detection",
			Status:     HealthStatusCritical,
			Message:    "No workspace detected in current directory",
			Suggestion: "Run 'pvm workspace init' to initialize a new workspace",
			CheckedAt:  time.Now(),
		}
	}

	return HealthCheck{
		Name:      "Workspace Detection",
		Status:    HealthStatusHealthy,
		Message:   fmt.Sprintf("Project detected via %s", ctx.DetectionInfo),
		Details:   fmt.Sprintf("Root directory: %s", ctx.RootDir),
		CheckedAt: time.Now(),
	}
}

// checkPerlVersion checks Perl version consistency
func checkPerlVersion(ctx *project.ProjectContext, autofix bool) []HealthCheck {
	checks := []HealthCheck{}

	// Check if .perl-version exists
	if ctx.PerlVersion == "" {
		check := HealthCheck{
			Name:       "Perl Version File",
			Status:     HealthStatusWarning,
			Message:    "No .perl-version file found",
			Suggestion: "Create .perl-version file to specify Perl version",
			CheckedAt:  time.Now(),
		}

		if autofix {
			if err := createPerlVersionFile(ctx.RootDir); err == nil {
				check.Status = HealthStatusHealthy
				check.Message = "Created .perl-version file with current Perl version"
			}
		}

		checks = append(checks, check)
		return checks
	}

	// Check if specified Perl version is installed
	installedVersion, err := getInstalledPerlVersion()
	if err != nil {
		checks = append(checks, HealthCheck{
			Name:      "Perl Installation",
			Status:    HealthStatusCritical,
			Message:   "Cannot detect installed Perl version",
			Details:   err.Error(),
			CheckedAt: time.Now(),
		})
		return checks
	}

	if installedVersion != ctx.PerlVersion {
		checks = append(checks, HealthCheck{
			Name:       "Perl Version Consistency",
			Status:     HealthStatusWarning,
			Message:    fmt.Sprintf("Project expects %s, but %s is installed", ctx.PerlVersion, installedVersion),
			Suggestion: "Install the correct Perl version or update .perl-version",
			CheckedAt:  time.Now(),
		})
	} else {
		checks = append(checks, HealthCheck{
			Name:      "Perl Version Consistency",
			Status:    HealthStatusHealthy,
			Message:   fmt.Sprintf("Perl version %s matches project requirement", installedVersion),
			CheckedAt: time.Now(),
		})
	}

	return checks
}

// checkDependencies checks dependency status
func checkDependencies(ctx *project.ProjectContext, autofix bool) []HealthCheck {
	checks := []HealthCheck{}

	// Check if cpanfile exists
	if !ctx.HasCpanfile {
		checks = append(checks, HealthCheck{
			Name:       "Dependencies File",
			Status:     HealthStatusWarning,
			Message:    "No cpanfile found",
			Suggestion: "Create cpanfile to manage dependencies",
			CheckedAt:  time.Now(),
		})
		return checks
	}

	// Check if local lib directory exists
	if _, err := os.Stat(ctx.LocalLibDir); os.IsNotExist(err) {
		check := HealthCheck{
			Name:       "Local Library",
			Status:     HealthStatusWarning,
			Message:    "Local lib directory does not exist",
			Suggestion: "Run 'pvm module install' to install dependencies",
			CheckedAt:  time.Now(),
		}

		if autofix {
			if err := os.MkdirAll(ctx.LocalLibDir, 0755); err == nil {
				check.Status = HealthStatusHealthy
				check.Message = "Created local lib directory"
			}
		}

		checks = append(checks, check)
	} else {
		// Count installed modules
		moduleCount := countInstalledModules(ctx.LocalLibDir)
		checks = append(checks, HealthCheck{
			Name:      "Local Library",
			Status:    HealthStatusHealthy,
			Message:   fmt.Sprintf("Local lib exists with %d modules", moduleCount),
			Details:   fmt.Sprintf("Path: %s", ctx.LocalLibDir),
			CheckedAt: time.Now(),
		})
	}

	return checks
}

// checkBuildSystem checks build system health
func checkBuildSystem(ctx *project.ProjectContext) []HealthCheck {
	checks := []HealthCheck{}

	// Check if build directory exists
	buildDir := filepath.Join(ctx.RootDir, "build")
	if _, err := os.Stat(buildDir); os.IsNotExist(err) {
		checks = append(checks, HealthCheck{
			Name:       "Build Artifacts",
			Status:     HealthStatusWarning,
			Message:    "No build artifacts found",
			Suggestion: "Run 'pvm build' to create build artifacts",
			CheckedAt:  time.Now(),
		})
	} else {
		checks = append(checks, HealthCheck{
			Name:      "Build Artifacts",
			Status:    HealthStatusHealthy,
			Message:   "Build directory exists",
			Details:   fmt.Sprintf("Path: %s", buildDir),
			CheckedAt: time.Now(),
		})
	}

	// Check for PSC availability
	if _, err := exec.LookPath("psc"); err != nil {
		checks = append(checks, HealthCheck{
			Name:       "PSC Type Checker",
			Status:     HealthStatusWarning,
			Message:    "PSC command not found in PATH",
			Suggestion: "Ensure PSC is built and available in PATH",
			CheckedAt:  time.Now(),
		})
	} else {
		checks = append(checks, HealthCheck{
			Name:      "PSC Type Checker",
			Status:    HealthStatusHealthy,
			Message:   "PSC command available",
			CheckedAt: time.Now(),
		})
	}

	return checks
}

// checkConfiguration checks configuration validity
func checkConfiguration(ctx *project.ProjectContext, autofix bool) []HealthCheck {
	checks := []HealthCheck{}

	// Check if pvm.toml exists
	if ctx.ConfigFile == "" {
		check := HealthCheck{
			Name:       "Workspace Configuration",
			Status:     HealthStatusWarning,
			Message:    "No pvm.toml configuration file",
			Suggestion: "Create pvm.toml for workspace configuration",
			CheckedAt:  time.Now(),
		}

		if autofix {
			if err := createDefaultConfig(ctx.RootDir); err == nil {
				check.Status = HealthStatusHealthy
				check.Message = "Created default pvm.toml configuration"
			}
		}

		checks = append(checks, check)
	} else {
		checks = append(checks, HealthCheck{
			Name:      "Workspace Configuration",
			Status:    HealthStatusHealthy,
			Message:   "Configuration file exists",
			Details:   fmt.Sprintf("Path: %s", ctx.ConfigFile),
			CheckedAt: time.Now(),
		})
	}

	return checks
}

// checkDevelopmentEnvironment checks development setup
func checkDevelopmentEnvironment(ctx *project.ProjectContext) []HealthCheck {
	checks := []HealthCheck{}

	// Check if it's a git repository
	gitDir := filepath.Join(ctx.RootDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		checks = append(checks, HealthCheck{
			Name:       "Version Control",
			Status:     HealthStatusWarning,
			Message:    "Not a Git repository",
			Suggestion: "Initialize Git repository for version control",
			CheckedAt:  time.Now(),
		})
	} else {
		checks = append(checks, HealthCheck{
			Name:      "Version Control",
			Status:    HealthStatusHealthy,
			Message:   "Git repository initialized",
			CheckedAt: time.Now(),
		})
	}

	// Check for .gitignore
	gitignorePath := filepath.Join(ctx.RootDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		checks = append(checks, HealthCheck{
			Name:       "Git Ignore",
			Status:     HealthStatusWarning,
			Message:    "No .gitignore file",
			Suggestion: "Create .gitignore to exclude build artifacts",
			CheckedAt:  time.Now(),
		})
	} else {
		checks = append(checks, HealthCheck{
			Name:      "Git Ignore",
			Status:    HealthStatusHealthy,
			Message:   ".gitignore file exists",
			CheckedAt: time.Now(),
		})
	}

	return checks
}

// Helper functions for health checks

func createPerlVersionFile(rootDir string) error {
	version := getCurrentPerlVersion()
	path := filepath.Join(rootDir, ".perl-version")
	return os.WriteFile(path, []byte(version+"\n"), 0644)
}

func getInstalledPerlVersion() (string, error) {
	cmd := exec.Command("perl", "-E", "say $^V")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Parse version from output like "v5.38.0"
	versionStr := strings.TrimSpace(string(output))
	if strings.HasPrefix(versionStr, "v") {
		return versionStr[1:], nil
	}
	return versionStr, nil
}

func countInstalledModules(localLibDir string) int {
	count := 0
	perl5Dir := filepath.Join(localLibDir, "perl5")

	filepath.Walk(perl5Dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if strings.HasSuffix(path, ".pm") {
			count++
		}
		return nil
	})

	return count
}

func createDefaultConfig(rootDir string) error {
	projectName := filepath.Base(rootDir)
	config := fmt.Sprintf(`[project]
name = "%s"
version = "0.01"

[build]
output_dir = "build"
`, projectName)
	path := filepath.Join(rootDir, "pvm.toml")
	return os.WriteFile(path, []byte(config), 0644)
}

// newWorkspaceValidateCommand creates a new validate command for workspace validation
func newWorkspaceValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate [script]",
		Short: "Validate complete workspace setup",
		Long: `Validates the complete PVM workspace setup including configuration,
dependencies, type checking, and execution.

This command provides comprehensive workspace validation by:
- Checking project configuration and structure
- Verifying Perl version availability
- Resolving and testing all dependencies
- Executing validation script if provided
- Generating detailed validation report`,
		RunE: runWorkspaceValidate,
	}

	cmd.Flags().Bool("verbose", false, "Enable verbose validation output")

	return cmd
}

// runWorkspaceValidate executes the workspace validation workflow
func runWorkspaceValidate(cmd *cobra.Command, args []string) error {
	ui := cli.GetUI(cmd)
	ui.Warning("Workspace validation requires type-system components that are not yet available in this build.")
	return fmt.Errorf("workspace validation is not yet available in this build")
}

// checkPerlVersionInstalled checks if a specific Perl version is installed
func checkPerlVersionInstalled(version string) (bool, error) {
	return perl.IsVersionInstalled(version)
}

// checkDependenciesInstalled checks if dependencies from cpanfile are installed
// Returns (all_installed bool, missing_count int)
func checkDependenciesInstalled(ctx *project.ProjectContext) (bool, int) {
	// Parse cpanfile to get required dependencies
	cpanfilePath := filepath.Join(ctx.RootDir, "cpanfile")
	if _, err := os.Stat(cpanfilePath); err != nil {
		return false, 0
	}

	// Read cpanfile to get dependency list
	data, err := os.ReadFile(cpanfilePath)
	if err != nil {
		return false, 0
	}

	// Simple parse for required modules (basic implementation)
	// In production, we'd use a proper cpanfile parser
	lines := strings.Split(string(data), "\n")
	var requiredModules []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "requires '") {
			// Extract module name from "requires 'Module::Name'"
			parts := strings.Split(line, "'")
			if len(parts) >= 2 {
				moduleName := parts[1]
				if moduleName != "perl" { // Skip perl version requirement
					requiredModules = append(requiredModules, moduleName)
				}
			}
		}
	}

	if len(requiredModules) == 0 {
		return true, 0 // No dependencies to check
	}

	// Check each module in local lib
	missingCount := 0
	for _, module := range requiredModules {
		if !isModuleInstalled(ctx.LocalLibDir, module) {
			missingCount++
		}
	}

	return missingCount == 0, missingCount
}

// isModuleInstalled checks if a specific module is installed in local lib
func isModuleInstalled(localLibDir, moduleName string) bool {
	// Convert Module::Name to Module/Name.pm
	modulePath := strings.ReplaceAll(moduleName, "::", "/") + ".pm"

	// Check in local lib perl5 directory
	perl5Dir := filepath.Join(localLibDir, "perl5")

	// Search for the module file
	found := false
	filepath.Walk(perl5Dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if strings.HasSuffix(path, modulePath) {
			found = true
			return filepath.SkipDir // Stop walking once found
		}
		return nil
	})

	return found
}
