# PVM Help Templates

This directory contains Markdown templates for PVM help content. These templates are embedded at compile time and can be easily updated without modifying Go source code.

## Available Templates

- `getting-started.md` - Getting started guide for new users
- `troubleshooting.md` - Common issues and solutions  
- `workflows.md` - Common development workflows

## Template System

The templates are processed by `internal/cli/help_templates.go` which:

1. Embeds all `.md` files from this directory using Go's `embed` directive
2. Provides template rendering with variable substitution
3. Converts Markdown to styled CLI output using PVM's UI framework

## Adding New Templates

1. Create a new `.md` file in this directory
2. Use standard Markdown syntax with these special patterns:
   - `# Header` - Main headers (styled with `ui.Header()`)
   - `## Sub Header` - Sub headers (styled with `ui.SubHeader()`)
   - `### Section` - Warning-style sections (styled with `ui.Warning()`)
   - `💡 Info` - Info tips (styled with `ui.Info()`)
   - `**Problem:**` - Problem descriptions (styled with `ui.Info()`)
   - `**Solution:**` - Solutions (styled with `ui.Success()`)
   - `**Command:**` - Command examples (styled as regular text)
   - Code blocks with backticks or indentation are preserved

3. Register the template in the help command system

## Template Variables

Templates can use Go template syntax for variable substitution:

```go
templateData := cli.HelpTemplateData{
    Version:     "0.1.0",
    ProjectPath: "/path/to/project",
}
```

Variables can be referenced in templates as `{{.Version}}`, `{{.ProjectPath}}`, etc.

## Benefits

- **Easy Updates**: Help content can be modified without recompiling
- **Version Control**: Help content changes are tracked in git
- **Collaboration**: Non-developers can contribute help improvements
- **Consistency**: Uniform styling across all help content
- **Maintainability**: Separation of content from code logic

## Usage in Code

```go
// Render a help template
content, err := cli.RenderHelpTemplate("troubleshooting", templateData)
if err != nil {
    return err
}

// Display as formatted help
cli.RenderMarkdownAsHelp(content, output)
```