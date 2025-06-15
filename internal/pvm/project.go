// ABOUTME: Project management commands for PVM
// ABOUTME: Handles project initialization, status, and configuration

package pvm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
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

// newProjectCommand creates the main project command with subcommands
func newProjectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Project management commands",
		Long:  "Commands for initializing, managing, and working with Perl projects",
	}

	// Add project subcommands
	cmd.AddCommand(
		newProjectInitCommand(),
		newProjectStatusCommand(),
		newProjectTemplatesCommand(),
	)

	return cmd
}

// newProjectInitCommand creates the project init command
func newProjectInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [project-name]",
		Short: "Initialize a new Perl project",
		Long: `Initialize a new Perl project with proper directory structure and configuration.

If project-name is provided, creates a new directory with that name.
If no name is provided, initializes in the current directory.

This command creates:
- .perl-version file with current/default Perl version
- cpanfile for dependency management
- pvm.toml for project configuration
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

			// Load the template
			template, err := loadTemplate(templateName)
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
	cmd.Flags().String("template", "minimal", "Project template to use (run 'pvm project templates' to list available)")
	cmd.Flags().Bool("force", false, "Initialize even if directory is not empty or project files exist")

	return cmd
}

// newProjectStatusCommand creates the project status command
func newProjectStatusCommand() *cobra.Command {
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
			return showProjectStatus(cmd)
		},
	}

	// Add flags
	cmd.Flags().Bool("json", false, "Output status in JSON format")

	return cmd
}

// newProjectTemplatesCommand creates the project templates command
func newProjectTemplatesCommand() *cobra.Command {
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
		cmd.Printf("Created project directory: %s\n", projectDir)
	}

	// Check if we're already in a project (unless force is used)
	if !force {
		existingProject, err := project.DetectProject(absProjectDir)
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

	cmd.Printf("\nProject '%s' initialized successfully!\n", projectName)
	cmd.Printf("Location: %s\n", absProjectDir)

	// Show next steps
	showNextSteps(cmd, createDir, projectDir)

	return nil
}

// createProjectFiles creates the necessary project configuration files
func createProjectFiles(cmd *cobra.Command, projectDir, projectName string, template *ProjectTemplate) error {
	// Create .perl-version file
	perlVersion := getCurrentPerlVersion()
	perlVersionPath := filepath.Join(projectDir, ".perl-version")
	if err := os.WriteFile(perlVersionPath, []byte(perlVersion+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to create .perl-version: %w", err)
	}
	cmd.Printf("Created .perl-version (%s)\n", perlVersion)

	// Create cpanfile
	cpanfileContent := generateCpanfile(projectName, template)
	cpanfilePath := filepath.Join(projectDir, "cpanfile")
	if err := os.WriteFile(cpanfilePath, []byte(cpanfileContent), 0644); err != nil {
		return fmt.Errorf("failed to create cpanfile: %w", err)
	}
	cmd.Printf("Created cpanfile\n")

	// Create pvm.toml
	pvmTomlContent := generatePvmToml(projectName, perlVersion, template)
	pvmTomlPath := filepath.Join(projectDir, "pvm.toml")
	if err := os.WriteFile(pvmTomlPath, []byte(pvmTomlContent), 0644); err != nil {
		return fmt.Errorf("failed to create pvm.toml: %w", err)
	}
	cmd.Printf("Created pvm.toml\n")

	// Create .gitignore
	gitignoreContent := generateGitignore(template)
	gitignorePath := filepath.Join(projectDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		return fmt.Errorf("failed to create .gitignore: %w", err)
	}
	cmd.Printf("Created .gitignore\n")

	return nil
}

// createDirectoryStructure creates the standard project directory structure
func createDirectoryStructure(cmd *cobra.Command, projectDir string, template *ProjectTemplate) error {
	dirs := template.Directories

	for _, dir := range dirs {
		dirPath := filepath.Join(projectDir, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		cmd.Printf("Created directory: %s/\n", dir)
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
local_lib = "lib"

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


// showProjectStatus shows comprehensive project status
func showProjectStatus(cmd *cobra.Command) error {
	// Get project context
	ctx, err := project.GetCurrentProject()
	if err != nil {
		return fmt.Errorf("failed to detect project: %w", err)
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

// outputStatusHuman outputs project status in human-readable format
func outputStatusHuman(cmd *cobra.Command, ctx *project.ProjectContext) error {
	if !ctx.IsProject {
		cmd.Println("No project detected in current directory")
		cmd.Println("Use 'pvm project init' to initialize a new project")
		return nil
	}

	cmd.Printf("Project Status\n")
	cmd.Printf("==============\n\n")

	// Project information
	cmd.Printf("Project Root: %s\n", ctx.RootDir)
	cmd.Printf("Detected via: %s\n", ctx.DetectionInfo)

	// Perl version
	if ctx.PerlVersion != "" {
		cmd.Printf("Perl Version: %s\n", ctx.PerlVersion)
		// TODO: Check if this version is actually installed
	} else {
		cmd.Printf("Perl Version: not specified\n")
	}

	// Dependencies
	if ctx.HasCpanfile {
		cmd.Printf("Dependencies: cpanfile present\n")
		// TODO: Check if dependencies are installed
	} else {
		cmd.Printf("Dependencies: no cpanfile\n")
	}

	// Local lib
	cmd.Printf("Local lib: %s\n", ctx.LocalLibDir)
	if _, err := os.Stat(ctx.LocalLibDir); err == nil {
		cmd.Printf("  Status: exists\n")
		// TODO: Count installed modules
	} else {
		cmd.Printf("  Status: not created\n")
	}

	// Configuration
	if ctx.ConfigFile != "" {
		cmd.Printf("Configuration: %s\n", ctx.ConfigFile)
	} else {
		cmd.Printf("Configuration: no pvm.toml\n")
	}

	// Build status
	buildDir := filepath.Join(ctx.RootDir, "build")
	if _, err := os.Stat(buildDir); err == nil {
		cmd.Printf("Build artifacts: present\n")
	} else {
		cmd.Printf("Build artifacts: none\n")
	}

	// Git status
	gitDir := filepath.Join(ctx.RootDir, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		cmd.Printf("Git repository: yes\n")
	} else {
		cmd.Printf("Git repository: no\n")
	}

	cmd.Printf("\nNext Steps:\n")
	if !ctx.HasCpanfile {
		cmd.Printf("- Add dependencies with 'pvm module add <module>'\n")
	}
	if ctx.PerlVersion == "" {
		cmd.Printf("- Set Perl version in .perl-version file\n")
	}
	cmd.Printf("- Install dependencies with 'pvm module install'\n")
	cmd.Printf("- Run type check with 'pvm build --check-only'\n")

	return nil
}

// outputStatusJSON outputs project status in JSON format
func outputStatusJSON(cmd *cobra.Command, ctx *project.ProjectContext) error {
	// TODO: Implement JSON output
	return fmt.Errorf("JSON output not yet implemented")
}

// showNextSteps shows what the user should do next after project initialization
func showNextSteps(cmd *cobra.Command, createdDir bool, projectDir string) {
	cmd.Printf("\nNext steps:\n")

	if createdDir {
		cmd.Printf("  cd %s\n", projectDir)
	}

	cmd.Printf("  pvm module install          # Install dependencies\n")
	cmd.Printf("  pvm project status          # Check project health\n")
	cmd.Printf("  pvm build --check-only      # Run type checking\n")
	cmd.Printf("\nLearn more:\n")
	cmd.Printf("  pvm help project            # Project management commands\n")
	cmd.Printf("  pvm help module             # Module management commands\n")
	cmd.Printf("  pvm help build              # Build system commands\n")
}

// loadTemplate loads a template by name, checking user templates first, then built-in
func loadTemplate(name string) (ProjectTemplate, error) {
	// First try to load from user templates
	template, err := loadUserTemplate(name)
	if err == nil {
		return template, nil
	}

	// Fall back to built-in templates
	return getBuiltinTemplate(name)
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

// getBuiltinTemplate returns a built-in template
func getBuiltinTemplate(name string) (ProjectTemplate, error) {
	switch name {
	case "minimal":
		return ProjectTemplate{
			Name:        "minimal",
			Description: "Minimal Perl project with basic structure",
			Directories: []string{"lib", "t"},
			Dependencies: map[string]string{
				"strict":   "",
				"warnings": "",
				"utf8":     "",
			},
			DevDeps: map[string]string{
				"Perl::Critic": "",
				"Perl::Tidy":   "",
			},
			TestDeps: map[string]string{
				"Test::More":      ">= 0.98",
				"Test::Exception": "",
			},
			Config:    map[string]any{},
			GitIgnore: []string{},
		}, nil
	default:
		return ProjectTemplate{}, fmt.Errorf("unknown template: %s", name)
	}
}

// listTemplates lists all available templates
func listTemplates(cmd *cobra.Command) error {
	cmd.Println("Available project templates:")
	cmd.Println()

	// List built-in templates
	cmd.Println("Built-in templates:")
	builtinTemplates := []string{"minimal"}
	for _, name := range builtinTemplates {
		template, err := getBuiltinTemplate(name)
		if err == nil {
			cmd.Printf("  %-12s - %s\n", name, template.Description)
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