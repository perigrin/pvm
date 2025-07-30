# Custom Help Commands with Subcommands

This document describes how to implement custom help commands with subcommands in Cobra CLI applications, particularly when integrating with UI frameworks like charmbracelet/fang.

## Problem Statement

By default, Cobra CLI automatically generates help commands, but this can lead to:
- Duplicate help commands in the command list
- Conflicts with custom UI frameworks (like Fang)
- Inability to implement rich help subcommands
- Limited control over help formatting and behavior

## Solution Overview

The solution uses a hybrid approach combining `SetHelpCommand()` and `AddCommand()` with strategic command hiding to achieve:
- Single help command entry in command lists
- Full subcommand functionality
- Custom help formatting and behavior
- Seamless integration with UI frameworks

## Implementation

### Step 1: Create Enhanced Help Command

First, create your enhanced help command with proper subcommands:

```go
func newEnhancedHelpCommand() *cobra.Command {
    helpCmd := &cobra.Command{
        Use:   "help [command]",
        Short: "Help about any command",
        Long:  "Help provides help for any command in the application.",
        RunE: func(cmd *cobra.Command, args []string) error {
            rootCmd := cmd.Root()
            if len(args) == 0 {
                // Show standard help
                return rootCmd.Help()
            }
            
            // Try to find the command
            topic := args[0]
            if targetCmd, _, err := rootCmd.Find([]string{topic}); err == nil && targetCmd != rootCmd {
                return targetCmd.Help()
            }
            
            // Command not found - show available topics
            ui := cli.GetUI(cmd)
            ui.Error("Unknown help topic: %s", topic)
            ui.Info("Available help topics: workflows, getting-started, troubleshooting")
            return fmt.Errorf("unknown help topic: %s", topic)
        },
    }

    // Add help subcommands
    helpCmd.AddCommand(
        newHelpWorkflowsCommand(),
        newHelpGettingStartedCommand(),
        newHelpTroubleshootingCommand(),
    )

    return helpCmd
}
```

### Step 2: Create Help Subcommands

Implement individual help subcommands:

```go
func newHelpWorkflowsCommand() *cobra.Command {
    return &cobra.Command{
        Use:   "workflows",
        Short: "Common development workflows",
        Long:  "Show common development workflows and usage patterns",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Your custom help implementation
            helpManager := cli.NewHelpManager()
            return showWorkflowHelp(cmd, helpManager)
        },
    }
}

func newHelpGettingStartedCommand() *cobra.Command {
    return &cobra.Command{
        Use:   "getting-started",
        Short: "New user onboarding",
        Long:  "Guide for new users to get started",
        RunE: func(cmd *cobra.Command, args []string) error {
            helpManager := cli.NewHelpManager()
            return showGettingStartedHelp(cmd, helpManager)
        },
    }
}
```

### Step 3: Register Help Commands (Critical)

This is the key part that prevents duplicate help commands:

```go
func NewRootCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "myapp",
        Short: "My Application",
        Long:  "Description of my application",
    }
    
    // Create enhanced help command
    helpCmd := newEnhancedHelpCommand()
    
    // Set as official help command (prevents Cobra auto-generation)
    cmd.SetHelpCommand(helpCmd)
    
    // Also add as hidden regular command (enables subcommands)
    hiddenHelpCmd := newEnhancedHelpCommand()
    hiddenHelpCmd.Hidden = true
    
    // Add commands
    cmd.AddCommand(
        hiddenHelpCmd, // Hidden version for subcommands to work
        newOtherCommand(),
        // ... other commands
    )

    return cmd
}
```

## Why This Works

### The Cobra Help System

Cobra's help system works as follows:
1. `InitDefaultHelpCmd()` only adds help if no help command exists
2. `SetHelpCommand()` prevents auto-generation
3. Subcommands need to be registered via `AddCommand()` to be discoverable
4. `Hidden = true` prevents commands from showing in help lists

### The Hybrid Approach

- **`SetHelpCommand(helpCmd)`**: Prevents duplicate auto-generated help command
- **`AddCommand(hiddenHelpCmd)`**: Enables subcommand discovery and routing
- **`hiddenHelpCmd.Hidden = true`**: Prevents second help entry in command list

## Integration with UI Frameworks

When using UI frameworks like charmbracelet/fang, the help command integrates seamlessly:

```go
// In your help implementation
func showWorkflowHelp(cmd *cobra.Command, helpManager *HelpManager) error {
    ui := cli.GetUI(cmd)
    
    // Use your UI framework for rich formatting
    ui.Header("Common Workflows")
    ui.Section("Development", workflowContent)
    
    // Support paging for long content
    return cli.ShowWithPager(content)
}
```

## Results

This implementation provides:

✅ **Single help command** in `myapp -h` (no duplicates)  
✅ **Standard help behavior** with `myapp help`  
✅ **Working subcommands** like `myapp help workflows`  
✅ **Custom formatting** with your UI framework  
✅ **Smart paging** for long help content  

## Example Usage

```bash
# Shows standard command list
$ myapp help

# Shows rich workflow help with custom formatting
$ myapp help workflows

# Shows getting started guide
$ myapp help getting-started

# Standard command help still works
$ myapp help install
```

## Common Pitfalls

### ❌ Don't: Use only SetHelpCommand
```go
// This breaks subcommands
cmd.SetHelpCommand(helpCmd)
// No AddCommand - subcommands won't work
```

### ❌ Don't: Use only AddCommand
```go
// This creates duplicate help commands
cmd.AddCommand(helpCmd)
// No SetHelpCommand - Cobra adds its own
```

### ✅ Do: Use the hybrid approach
```go
// This is the correct pattern
cmd.SetHelpCommand(helpCmd)
hiddenHelp := newEnhancedHelpCommand()
hiddenHelp.Hidden = true
cmd.AddCommand(hiddenHelp)
```

## Template System Integration

For maintainable help content, consider using embedded templates:

```go
//go:embed help
var helpTemplatesFS embed.FS

func RenderHelpTemplate(name string, data interface{}) (string, error) {
    tmpl, err := template.ParseFS(helpTemplatesFS, fmt.Sprintf("help/%s.md", name))
    if err != nil {
        return "", err
    }
    
    var buf bytes.Buffer
    err = tmpl.Execute(&buf, data)
    return buf.String(), err
}
```

This allows you to maintain help content in markdown files while providing dynamic data injection.

## Conclusion

This pattern enables rich, customizable help systems while maintaining compatibility with Cobra's conventions and preventing duplicate command entries. The key insight is using `SetHelpCommand()` to prevent auto-generation while strategically using hidden commands to enable full subcommand functionality.