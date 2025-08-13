// ABOUTME: Unit tests for plenv compatibility interface
// ABOUTME: Tests command mapping and argument transformation for plenv

package compat

import (
	"strings"
	"testing"
)

func TestPlenvMapper_MapCommand(t *testing.T) {
	mapper := NewPlenvMapper()

	tests := []struct {
		name     string
		args     []string
		wantCmd  string
		wantArgs []string
		wantErr  bool
	}{
		{
			name:     "Init command",
			args:     []string{"init"},
			wantCmd:  "init",
			wantArgs: nil,
			wantErr:  false,
		},
		{
			name:     "Install version",
			args:     []string{"install", "5.38.0"},
			wantCmd:  "install",
			wantArgs: []string{"5.38.0"},
			wantErr:  false,
		},
		{
			name:     "Install with list flag",
			args:     []string{"install", "--list"},
			wantCmd:  "available",
			wantArgs: nil,
			wantErr:  false,
		},
		{
			name:     "Install with short list flag",
			args:     []string{"install", "-l"},
			wantCmd:  "available",
			wantArgs: nil,
			wantErr:  false,
		},
		{
			name:     "Install with skip existing",
			args:     []string{"install", "--skip-existing", "5.38.0"},
			wantCmd:  "install",
			wantArgs: []string{"5.38.0", "--skip-existing"},
			wantErr:  false,
		},
		{
			name:     "Install with verbose",
			args:     []string{"install", "--verbose", "5.38.0"},
			wantCmd:  "install",
			wantArgs: []string{"5.38.0", "--verbose"},
			wantErr:  false,
		},
		{
			name:     "Install with keep build",
			args:     []string{"install", "--keep", "5.38.0"},
			wantCmd:  "install",
			wantArgs: []string{"5.38.0", "--keep-build"},
			wantErr:  false,
		},
		{
			name:     "Install with debug",
			args:     []string{"install", "--debug", "5.38.0"},
			wantCmd:  "install",
			wantArgs: []string{"5.38.0", "--debug"},
			wantErr:  false,
		},
		{
			name:     "Uninstall version",
			args:     []string{"uninstall", "5.36.0"},
			wantCmd:  "uninstall",
			wantArgs: []string{"5.36.0"},
			wantErr:  false,
		},
		{
			name:     "Uninstall with force",
			args:     []string{"uninstall", "5.36.0", "--force"},
			wantCmd:  "uninstall",
			wantArgs: []string{"5.36.0", "--force"},
			wantErr:  false,
		},
		{
			name:     "List versions",
			args:     []string{"versions"},
			wantCmd:  "list",
			wantArgs: nil,
			wantErr:  false,
		},
		{
			name:     "Show current version",
			args:     []string{"version"},
			wantCmd:  "current",
			wantArgs: nil,
			wantErr:  false,
		},
		{
			name:     "Set global version",
			args:     []string{"global", "5.38.0"},
			wantCmd:  "use",
			wantArgs: []string{"5.38.0", "--global"},
			wantErr:  false,
		},
		{
			name:     "Show global version",
			args:     []string{"global"},
			wantCmd:  "current",
			wantArgs: []string{"--global"},
			wantErr:  false,
		},
		{
			name:     "Set local version",
			args:     []string{"local", "5.38.0"},
			wantCmd:  "use",
			wantArgs: []string{"5.38.0", "--local"},
			wantErr:  false,
		},
		{
			name:     "Show local version",
			args:     []string{"local"},
			wantCmd:  "current",
			wantArgs: []string{"--local"},
			wantErr:  false,
		},
		{
			name:     "Unset local version",
			args:     []string{"local", "--unset"},
			wantCmd:  "use",
			wantArgs: []string{"--unset-local"},
			wantErr:  false,
		},
		{
			name:     "Set shell version",
			args:     []string{"shell", "5.38.0"},
			wantCmd:  "shell",
			wantArgs: []string{"5.38.0"},
			wantErr:  false,
		},
		{
			name:     "Show shell version",
			args:     []string{"shell"},
			wantCmd:  "current",
			wantArgs: []string{"--shell"},
			wantErr:  false,
		},
		{
			name:     "Unset shell version",
			args:     []string{"shell", "--unset"},
			wantCmd:  "shell",
			wantArgs: []string{"--unset"},
			wantErr:  false,
		},
		{
			name:     "Rehash command",
			args:     []string{"rehash"},
			wantCmd:  "",
			wantArgs: nil,
			wantErr:  false,
		},
		{
			name:     "Which command",
			args:     []string{"which", "cpanm"},
			wantCmd:  "which",
			wantArgs: []string{"cpanm"},
			wantErr:  false,
		},
		{
			name:     "Whence command",
			args:     []string{"whence", "perl"},
			wantCmd:  "whence",
			wantArgs: []string{"perl"},
			wantErr:  false,
		},
		{
			name:     "Exec command",
			args:     []string{"exec", "perl", "-v"},
			wantCmd:  "exec",
			wantArgs: []string{"perl", "-v"},
			wantErr:  false,
		},
		{
			name:     "Shims command",
			args:     []string{"shims"},
			wantCmd:  "shims",
			wantArgs: nil,
			wantErr:  false,
		},
		{
			name:     "Prefix with version",
			args:     []string{"prefix", "5.38.0"},
			wantCmd:  "prefix",
			wantArgs: []string{"5.38.0"},
			wantErr:  false,
		},
		{
			name:     "Prefix current version",
			args:     []string{"prefix"},
			wantCmd:  "prefix",
			wantArgs: nil,
			wantErr:  false,
		},
		{
			name:     "Root command",
			args:     []string{"root"},
			wantCmd:  "root",
			wantArgs: nil,
			wantErr:  false,
		},
		{
			name:     "Version name command",
			args:     []string{"version-name"},
			wantCmd:  "current",
			wantArgs: []string{"--name-only"},
			wantErr:  false,
		},
		{
			name:     "Version origin command",
			args:     []string{"version-origin"},
			wantCmd:  "current",
			wantArgs: []string{"--origin"},
			wantErr:  false,
		},
		{
			name:     "Commands command",
			args:     []string{"commands"},
			wantCmd:  "help",
			wantArgs: nil,
			wantErr:  false,
		},
		{
			name:     "Help command",
			args:     []string{"help"},
			wantCmd:  "",
			wantArgs: nil,
			wantErr:  false,
		},
		{
			name:    "No command provided",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "Install without version",
			args:    []string{"install"},
			wantErr: true,
		},
		{
			name:    "Install skip-existing without version",
			args:    []string{"install", "--skip-existing"},
			wantErr: true,
		},
		{
			name:    "Uninstall without version",
			args:    []string{"uninstall"},
			wantErr: true,
		},
		{
			name:    "Which without command",
			args:    []string{"which"},
			wantErr: true,
		},
		{
			name:    "Whence without command",
			args:    []string{"whence"},
			wantErr: true,
		},
		{
			name:    "Exec without command",
			args:    []string{"exec"},
			wantErr: true,
		},
		{
			name:    "Unknown command",
			args:    []string{"unknown"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCmd, gotArgs, err := mapper.MapCommand(tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("MapCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return // Expected error, test passed
			}

			if gotCmd != tt.wantCmd {
				t.Errorf("MapCommand() gotCmd = %v, want %v", gotCmd, tt.wantCmd)
			}

			if len(gotArgs) != len(tt.wantArgs) {
				t.Errorf("MapCommand() gotArgs length = %v, want %v", len(gotArgs), len(tt.wantArgs))
				return
			}

			for i, arg := range gotArgs {
				if arg != tt.wantArgs[i] {
					t.Errorf("MapCommand() gotArgs[%d] = %v, want %v", i, arg, tt.wantArgs[i])
				}
			}
		})
	}
}

func TestPlenvMapper_GetHelp(t *testing.T) {
	mapper := NewPlenvMapper()
	help := mapper.GetHelp()

	expectedStrings := []string{
		"plenv",
		"Usage:",
		"Commands:",
		"Examples:",
		"install",
		"global",
		"local",
		"shell",
		"versions",
		"exec",
		"--list",
		"--verbose",
		"PVM compatibility interface",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(help, expected) {
			t.Errorf("GetHelp() missing expected string: %s", expected)
		}
	}
}

func TestPlenvMapper_GetToolName(t *testing.T) {
	mapper := NewPlenvMapper()
	if mapper.GetToolName() != "plenv" {
		t.Errorf("GetToolName() = %v, want %v", mapper.GetToolName(), "plenv")
	}
}
