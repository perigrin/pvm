// ABOUTME: Unit tests for carton compatibility interface
// ABOUTME: Tests command mapping and argument transformation for carton

package compat

import (
	"strings"
	"testing"
)

func TestCartonMapper_MapCommand(t *testing.T) {
	mapper := NewCartonMapper()

	tests := []struct {
		name     string
		args     []string
		wantCmd  string
		wantArgs []string
		wantErr  bool
	}{
		{
			name:     "Default to install with no args",
			args:     []string{},
			wantCmd:  "module",
			wantArgs: []string{"install", "--deps-only"},
			wantErr:  false,
		},
		{
			name:     "Install command",
			args:     []string{"install"},
			wantCmd:  "module",
			wantArgs: []string{"install", "--deps-only"},
			wantErr:  false,
		},
		{
			name:     "Install with deployment flag",
			args:     []string{"install", "--deployment"},
			wantCmd:  "module",
			wantArgs: []string{"install", "--deps-only", "--production"},
			wantErr:  false,
		},
		{
			name:     "Install with path",
			args:     []string{"install", "--path", "/tmp/modules"},
			wantCmd:  "module",
			wantArgs: []string{"install", "--deps-only", "--install-dir", "/tmp/modules"},
			wantErr:  false,
		},
		{
			name:     "Install with path using equals",
			args:     []string{"install", "--path=/tmp/modules"},
			wantCmd:  "module",
			wantArgs: []string{"install", "--deps-only", "--install-dir", "/tmp/modules"},
			wantErr:  false,
		},
		{
			name:     "Install with without flag",
			args:     []string{"install", "--without", "test"},
			wantCmd:  "module",
			wantArgs: []string{"install", "--deps-only", "--without", "test"},
			wantErr:  false,
		},
		{
			name:     "Exec command",
			args:     []string{"exec", "--", "perl", "script.pl"},
			wantCmd:  "run",
			wantArgs: []string{"perl", "script.pl"},
			wantErr:  false,
		},
		{
			name:     "Exec without double dash",
			args:     []string{"exec", "prove", "t/"},
			wantCmd:  "run",
			wantArgs: []string{"prove", "t/"},
			wantErr:  false,
		},
		{
			name:     "List command",
			args:     []string{"list"},
			wantCmd:  "module",
			wantArgs: []string{"list"},
			wantErr:  false,
		},
		{
			name:     "Show command",
			args:     []string{"show", "Moose"},
			wantCmd:  "module",
			wantArgs: []string{"info", "Moose"},
			wantErr:  false,
		},
		{
			name:     "Update command with modules",
			args:     []string{"update", "DBI", "JSON"},
			wantCmd:  "module",
			wantArgs: []string{"update", "DBI", "JSON"},
			wantErr:  false,
		},
		{
			name:     "Update all modules",
			args:     []string{"update"},
			wantCmd:  "module",
			wantArgs: []string{"update"},
			wantErr:  false,
		},
		{
			name:     "Tree command",
			args:     []string{"tree"},
			wantCmd:  "module",
			wantArgs: []string{"tree"},
			wantErr:  false,
		},
		{
			name:     "Check command",
			args:     []string{"check"},
			wantCmd:  "module",
			wantArgs: []string{"outdated"},
			wantErr:  false,
		},
		{
			name:     "Bundle command",
			args:     []string{"bundle"},
			wantCmd:  "module",
			wantArgs: []string{"bundle"},
			wantErr:  false,
		},
		{
			name:     "Version command",
			args:     []string{"version"},
			wantCmd:  "version",
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
			name:    "Show without module name",
			args:    []string{"show"},
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

func TestCartonMapper_GetHelp(t *testing.T) {
	mapper := NewCartonMapper()
	help := mapper.GetHelp()

	expectedStrings := []string{
		"carton",
		"Usage:",
		"Commands:",
		"Examples:",
		"install",
		"exec",
		"--deployment",
		"PVM compatibility interface",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(help, expected) {
			t.Errorf("GetHelp() missing expected string: %s", expected)
		}
	}
}

func TestCartonMapper_GetToolName(t *testing.T) {
	mapper := NewCartonMapper()
	if mapper.GetToolName() != "carton" {
		t.Errorf("GetToolName() = %v, want %v", mapper.GetToolName(), "carton")
	}
}
