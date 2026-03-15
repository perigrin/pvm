// ABOUTME: Unit tests for cpanm compatibility interface
// ABOUTME: Tests command mapping and argument transformation

package compat

import (
	"strings"
	"testing"
)

func TestCpanmMapper_MapCommand(t *testing.T) {
	mapper := NewCpanmMapper()

	tests := []struct {
		name     string
		args     []string
		wantCmd  string
		wantArgs []string
		wantErr  bool
	}{
		{
			name:     "Install simple module",
			args:     []string{"Moose"},
			wantCmd:  "install",
			wantArgs: []string{"Moose"},
			wantErr:  false,
		},
		{
			name:     "Install with verbose flag",
			args:     []string{"-v", "DBI"},
			wantCmd:  "install",
			wantArgs: []string{"--verbose", "DBI"},
			wantErr:  false,
		},
		{
			name:     "Install with force and notest",
			args:     []string{"-f", "-n", "Test::More"},
			wantCmd:  "install",
			wantArgs: []string{"--force", "--no-test", "Test::More"},
			wantErr:  false,
		},
		{
			name:     "Install to local lib",
			args:     []string{"-L", "/tmp/perl5", "JSON"},
			wantCmd:  "install",
			wantArgs: []string{"--install-dir", "/tmp/perl5", "JSON"},
			wantErr:  false,
		},
		{
			name:     "Install deps only",
			args:     []string{"--installdeps", "."},
			wantCmd:  "install",
			wantArgs: []string{"--deps-only", "."},
			wantErr:  false,
		},
		{
			name:     "Show dependencies",
			args:     []string{"--showdeps", "Catalyst"},
			wantCmd:  "module",
			wantArgs: []string{"info", "--dependencies", "Catalyst"},
			wantErr:  false,
		},
		{
			name:     "Self upgrade",
			args:     []string{"--self-upgrade"},
			wantCmd:  "self-update",
			wantArgs: nil,
			wantErr:  false,
		},
		{
			name:     "Multiple modules",
			args:     []string{"Moose", "DBI", "JSON"},
			wantCmd:  "install",
			wantArgs: []string{"Moose", "DBI", "JSON"},
			wantErr:  false,
		},
		{
			name:     "Help flag",
			args:     []string{"--help"},
			wantCmd:  "",
			wantArgs: nil,
			wantErr:  false,
		},
		{
			name:    "No arguments",
			args:    []string{},
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

func TestCpanmMapper_GetHelp(t *testing.T) {
	mapper := NewCpanmMapper()
	help := mapper.GetHelp()

	// Check that help contains key cpanm information
	expectedStrings := []string{
		"cpanm",
		"Usage:",
		"Options:",
		"Examples:",
		"-v, --verbose",
		"-f, --force",
		"--installdeps",
		"PVM compatibility interface",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(help, expected) {
			t.Errorf("GetHelp() missing expected string: %s", expected)
		}
	}
}

func TestCpanmMapper_GetToolName(t *testing.T) {
	mapper := NewCpanmMapper()
	if mapper.GetToolName() != "cpanm" {
		t.Errorf("GetToolName() = %v, want %v", mapper.GetToolName(), "cpanm")
	}
}
