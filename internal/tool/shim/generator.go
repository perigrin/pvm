// ABOUTME: Generates platform-specific shim scripts for global tool execution
// ABOUTME: Handles Unix shell scripts and Windows batch files with proper error handling

package shim

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"text/template"

	"tamarou.com/pvm/internal/tool"
)

const (
	unixShimTemplate = `#!/bin/bash
# PVX Global Tool Shim: {{.ToolName}}
# Generated automatically - do not edit
# Tool: {{.ToolName}}
# Module: {{.Module}}
# Version: {{.Version}}
# Created: {{.CreatedAt}}

set -e

# Find PVX executable
PVX_EXEC=""
if command -v pvx >/dev/null 2>&1; then
    PVX_EXEC="pvx"
elif [ -x "$(dirname "$0")/../../../bin/pvx" ]; then
    PVX_EXEC="$(dirname "$0")/../../../bin/pvx"
elif [ -x "/usr/local/bin/pvx" ]; then
    PVX_EXEC="/usr/local/bin/pvx"
else
    echo "Error: pvx executable not found" >&2
    exit 1
fi

# Execute tool through PVX
exec "$PVX_EXEC" "{{.ToolName}}" "$@"
`

	windowsShimTemplate = `@echo off
REM PVX Global Tool Shim: {{.ToolName}}
REM Generated automatically - do not edit
REM Tool: {{.ToolName}}
REM Module: {{.Module}}
REM Version: {{.Version}}
REM Created: {{.CreatedAt}}

setlocal EnableDelayedExpansion

REM Find PVX executable
set "PVX_EXEC="
where pvx >nul 2>&1
if !errorlevel! equ 0 (
    set "PVX_EXEC=pvx"
) else (
    if exist "%~dp0..\..\..\bin\pvx.exe" (
        set "PVX_EXEC=%~dp0..\..\..\bin\pvx.exe"
    ) else if exist "C:\Program Files\PVM\bin\pvx.exe" (
        set "PVX_EXEC=C:\Program Files\PVM\bin\pvx.exe"
    ) else (
        echo Error: pvx executable not found >&2
        exit /b 1
    )
)

REM Execute tool through PVX
"%PVX_EXEC%" "{{.ToolName}}" %*
`
)

// ShimData holds the data needed to generate a shim
type ShimData struct {
	ToolName  string
	Module    string
	Version   string
	CreatedAt string
}

// generateShimContent generates the content for a shim script
func (m *Manager) generateShimContent(toolName string, info *tool.ToolInfo) (string, error) {
	data := ShimData{
		ToolName:  toolName,
		Module:    info.Module,
		Version:   info.Version,
		CreatedAt: info.InstallDate.Format("2006-01-02 15:04:05"),
	}

	var templateStr string
	switch m.platform {
	case "windows":
		templateStr = windowsShimTemplate
	default:
		templateStr = unixShimTemplate
	}

	tmpl, err := template.New("shim").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse shim template: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute shim template: %w", err)
	}

	return buf.String(), nil
}

// ValidateShimName validates that a tool name is suitable for use as a shim
func ValidateShimName(toolName string) error {
	if toolName == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	if len(toolName) > 255 {
		return fmt.Errorf("tool name too long (max 255 characters)")
	}

	// Check for invalid characters
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalid {
		if strings.Contains(toolName, char) {
			return fmt.Errorf("tool name contains invalid character: %s", char)
		}
	}

	// Check for reserved names on Windows
	if runtime.GOOS == "windows" {
		reserved := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
		upperTool := strings.ToUpper(toolName)
		for _, name := range reserved {
			if upperTool == name {
				return fmt.Errorf("tool name conflicts with Windows reserved name: %s", name)
			}
		}
	}

	return nil
}

// GetShimInfo extracts information from an existing shim file
func GetShimInfo(shimPath string) (*ShimData, error) {
	content, err := os.ReadFile(shimPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read shim file: %w", err)
	}

	contentStr := string(content)
	data := &ShimData{}

	// Extract tool name
	if toolName := extractFromComment(contentStr, "Tool:"); toolName != "" {
		data.ToolName = toolName
	}

	// Extract module
	if module := extractFromComment(contentStr, "Module:"); module != "" {
		data.Module = module
	}

	// Extract version
	if version := extractFromComment(contentStr, "Version:"); version != "" {
		data.Version = version
	}

	// Extract created at
	if createdAt := extractFromComment(contentStr, "Created:"); createdAt != "" {
		data.CreatedAt = createdAt
	}

	return data, nil
}

// extractFromComment extracts a value from a comment line
func extractFromComment(content, key string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Handle both Unix (#) and Windows (REM) comment styles
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "REM") {
			if strings.Contains(line, key) {
				parts := strings.SplitN(line, key, 2)
				if len(parts) == 2 {
					return strings.TrimSpace(parts[1])
				}
			}
		}
	}
	return ""
}

// IsShimFile checks if a file is a PVX-generated shim
func IsShimFile(filePath string) bool {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}

	contentStr := string(content)
	return strings.Contains(contentStr, "PVX Global Tool Shim")
}

// GenerateShimName generates a safe shim name from a tool name
func GenerateShimName(toolName string) string {
	// Replace invalid characters with underscores
	name := toolName
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", " "}
	for _, char := range invalid {
		name = strings.ReplaceAll(name, char, "_")
	}

	// Remove leading/trailing underscores
	name = strings.Trim(name, "_")

	// Ensure it's not empty
	if name == "" {
		name = "tool"
	}

	return name
}
