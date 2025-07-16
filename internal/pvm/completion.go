// ABOUTME: Shell completion system for PVM commands
// ABOUTME: Generates context-aware completions for various shell types

package pvm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/perl"
)

// newCompletionCommand creates a new completion command
func newCompletionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [shell]",
		Short: "Generate shell completion",
		Long: `Generate shell completion for the specified shell.

Supported shells:
  bash        Generate completions for bash
  zsh         Generate completions for zsh
  fish        Generate completions for fish
  powershell  Generate completions for PowerShell

Example:
  pvm completion bash   # Generate bash completions
  pvm completion zsh    # Generate zsh completions
  pvm completion fish   # Generate fish completions
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			shell := args[0]

			// Convert shell string to ShellType
			var shellType perl.ShellType
			switch shell {
			case "bash":
				shellType = perl.ShellBash
			case "zsh":
				shellType = perl.ShellZsh
			case "fish":
				shellType = perl.ShellFish
			case "powershell":
				shellType = perl.ShellPowerShell
			default:
				return fmt.Errorf("unsupported shell: %s", shell)
			}

			// Generate completion script
			script, err := generateCompletionScript(shellType)
			if err != nil {
				return fmt.Errorf("failed to generate completion script: %w", err)
			}

			// Output the script
			fmt.Print(script)
			return nil
		},
	}

	return cmd
}

// generateCompletionScript generates shell completion script for the given shell
func generateCompletionScript(shellType perl.ShellType) (string, error) {
	switch shellType {
	case perl.ShellBash:
		return generateBashCompletion()
	case perl.ShellZsh:
		return generateZshCompletion()
	case perl.ShellFish:
		return generateFishCompletion()
	case perl.ShellPowerShell:
		return generatePowerShellCompletion()
	default:
		return "", fmt.Errorf("unsupported shell type: %s", shellType)
	}
}

// generateBashCompletion generates bash completion script
func generateBashCompletion() (string, error) {
	return `# PVM Bash Completion
_pvm_completion() {
    local cur prev words cword
    _init_completion || return

    # Get current word being completed
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    # Main commands
    local commands="install use versions list available current local global init shell completion uninstall update build run module project dev test exec import rehash resolve mcp"

    # Subcommands for specific commands
    case "$prev" in
        pvm)
            COMPREPLY=($(compgen -W "$commands" -- "$cur"))
            return 0
            ;;
        use|install|uninstall)
            # Complete with available/installed versions
            local versions=$(pvm versions --short 2>/dev/null)
            COMPREPLY=($(compgen -W "$versions" -- "$cur"))
            return 0
            ;;
        shell)
            COMPREPLY=($(compgen -W "init setup" -- "$cur"))
            return 0
            ;;
        completion)
            COMPREPLY=($(compgen -W "bash zsh fish powershell" -- "$cur"))
            return 0
            ;;
        module)
            COMPREPLY=($(compgen -W "install remove list search update outdated" -- "$cur"))
            return 0
            ;;
        project)
            COMPREPLY=($(compgen -W "init build test run" -- "$cur"))
            return 0
            ;;
        *)
            # Default completion with files
            COMPREPLY=($(compgen -f -- "$cur"))
            return 0
            ;;
    esac
}

# Register completion function
complete -F _pvm_completion pvm
`, nil
}

// generateZshCompletion generates zsh completion script
func generateZshCompletion() (string, error) {
	return `#compdef pvm

# PVM Zsh Completion
_pvm() {
    local -a commands
    local state line

    commands=(
        'install:Install a Perl version'
        'use:Use a specific Perl version'
        'versions:List installed Perl versions'
        'list:List installed Perl versions (alias for versions)'
        'available:List available Perl versions'
        'current:Show current Perl version'
        'local:Set local Perl version'
        'global:Set global Perl version'
        'init:Initialize PVM'
        'shell:Shell integration commands'
        'completion:Generate shell completion'
        'uninstall:Uninstall a Perl version'
        'update:Update PVM'
        'build:Build commands'
        'run:Run scripts/modules'
        'module:Module management'
        'project:Project management'
        'dev:Development environment'
        'test:Test execution'
        'exec:Execute commands'
        'import:Import existing Perl installations'
        'rehash:Rehash shims'
        'resolve:Resolve version conflicts'
        'mcp:MCP server functionality'
    )

    _arguments \
        '1: :->commands' \
        '*: :->args' \
        && return 0

    case $state in
        commands)
            _describe -t commands 'pvm commands' commands
            ;;
        args)
            case $words[2] in
                use|install|uninstall)
                    # Complete with available/installed versions
                    local versions=($(pvm versions --short 2>/dev/null))
                    _describe -t versions 'perl versions' versions
                    ;;
                shell)
                    _arguments \
                        '1: :(init setup)'
                    ;;
                completion)
                    _arguments \
                        '1: :(bash zsh fish powershell)'
                    ;;
                module)
                    _arguments \
                        '1: :(install remove list search update outdated)'
                    ;;
                project)
                    _arguments \
                        '1: :(init build test run)'
                    ;;
                *)
                    _files
                    ;;
            esac
            ;;
    esac
}

compdef _pvm pvm
`, nil
}

// generateFishCompletion generates fish completion script
func generateFishCompletion() (string, error) {
	return `# PVM Fish Completion

# Main commands
complete -c pvm -f -a 'install' -d 'Install a Perl version'
complete -c pvm -f -a 'use' -d 'Use a specific Perl version'
complete -c pvm -f -a 'versions' -d 'List installed Perl versions'
complete -c pvm -f -a 'list' -d 'List installed Perl versions (alias for versions)'
complete -c pvm -f -a 'available' -d 'List available Perl versions'
complete -c pvm -f -a 'current' -d 'Show current Perl version'
complete -c pvm -f -a 'local' -d 'Set local Perl version'
complete -c pvm -f -a 'global' -d 'Set global Perl version'
complete -c pvm -f -a 'init' -d 'Initialize PVM'
complete -c pvm -f -a 'shell' -d 'Shell integration commands'
complete -c pvm -f -a 'completion' -d 'Generate shell completion'
complete -c pvm -f -a 'uninstall' -d 'Uninstall a Perl version'
complete -c pvm -f -a 'update' -d 'Update PVM'
complete -c pvm -f -a 'build' -d 'Build commands'
complete -c pvm -f -a 'run' -d 'Run scripts/modules'
complete -c pvm -f -a 'module' -d 'Module management'
complete -c pvm -f -a 'project' -d 'Project management'
complete -c pvm -f -a 'dev' -d 'Development environment'
complete -c pvm -f -a 'test' -d 'Test execution'
complete -c pvm -f -a 'exec' -d 'Execute commands'
complete -c pvm -f -a 'import' -d 'Import existing Perl installations'
complete -c pvm -f -a 'rehash' -d 'Rehash shims'
complete -c pvm -f -a 'resolve' -d 'Resolve version conflicts'
complete -c pvm -f -a 'mcp' -d 'MCP server functionality'

# Version completions for install/use/uninstall
function __pvm_versions
    pvm versions --short 2>/dev/null
end

complete -c pvm -f -n '__fish_seen_subcommand_from use install uninstall' -a '(__pvm_versions)'

# Shell subcommands
complete -c pvm -f -n '__fish_seen_subcommand_from shell' -a 'init setup'

# Completion subcommands
complete -c pvm -f -n '__fish_seen_subcommand_from completion' -a 'bash zsh fish powershell'

# Module subcommands
complete -c pvm -f -n '__fish_seen_subcommand_from module' -a 'install remove list search update outdated'

# Project subcommands
complete -c pvm -f -n '__fish_seen_subcommand_from project' -a 'init build test run'
`, nil
}

// generatePowerShellCompletion generates PowerShell completion script
func generatePowerShellCompletion() (string, error) {
	return `# PVM PowerShell Completion

Register-ArgumentCompleter -Native -CommandName pvm -ScriptBlock {
    param($commandName, $wordToComplete, $cursorPosition)

    # Get the current command line
    $line = $wordToComplete
    $words = $line -split '\s+'

    # Main commands
    $commands = @(
        'install', 'use', 'versions', 'list', 'available', 'current', 'local', 'global',
        'init', 'shell', 'completion', 'uninstall', 'update', 'build', 'run', 'module',
        'project', 'dev', 'test', 'exec', 'import', 'rehash', 'resolve', 'mcp'
    )

    # If we're completing the first argument (command)
    if ($words.Count -le 2) {
        $commands | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
            [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
        }
        return
    }

    # Get the command (second word)
    $command = $words[1]

    switch ($command) {
        'use' {
            # Complete with installed versions
            $versions = & pvm versions --short 2>$null
            $versions | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
            }
        }
        'install' {
            # Complete with available versions
            $versions = & pvm available --short 2>$null
            $versions | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
            }
        }
        'uninstall' {
            # Complete with installed versions
            $versions = & pvm versions --short 2>$null
            $versions | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
            }
        }
        'shell' {
            $subcommands = @('init', 'setup')
            $subcommands | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
            }
        }
        'completion' {
            $shells = @('bash', 'zsh', 'fish', 'powershell')
            $shells | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
            }
        }
        'module' {
            $subcommands = @('install', 'remove', 'list', 'search', 'update', 'outdated')
            $subcommands | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
            }
        }
        'project' {
            $subcommands = @('init', 'build', 'test', 'run')
            $subcommands | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
            }
        }
        default {
            # Default file completion
            # PowerShell will handle this automatically
        }
    }
}
`, nil
}

// getInstalledVersions returns a list of installed Perl versions for completion
func getInstalledVersions() ([]string, error) {
	versions, err := perl.GetInstalledVersions()
	if err != nil {
		return nil, err
	}

	var versionStrings []string
	for _, v := range versions {
		versionStrings = append(versionStrings, v.Version)
	}

	return versionStrings, nil
}

// getAvailableVersions returns a list of available Perl versions for completion
func getAvailableVersions() ([]string, error) {
	// This would need to be implemented in the perl package
	// For now, return empty list
	return []string{}, nil
}

// completeFiles provides file completion for paths
func completeFiles(prefix string) ([]string, error) {
	dir := filepath.Dir(prefix)
	if dir == "." {
		dir = ""
	}

	base := filepath.Base(prefix)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var matches []string
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), base) {
			fullPath := filepath.Join(dir, entry.Name())
			if entry.IsDir() {
				fullPath += "/"
			}
			matches = append(matches, fullPath)
		}
	}

	return matches, nil
}
