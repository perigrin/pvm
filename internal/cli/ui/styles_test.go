// ABOUTME: Tests for PVM UI styling framework
// ABOUTME: Verifies theme creation, style application, and Fang integration

package ui

import (
	"testing"
)

func TestDefaultTheme(t *testing.T) {
	theme := DefaultTheme

	// Test that all colors are defined
	if theme.Primary == "" {
		t.Error("DefaultTheme.Primary should not be empty")
	}
	if theme.Success == "" {
		t.Error("DefaultTheme.Success should not be empty")
	}
	if theme.Error == "" {
		t.Error("DefaultTheme.Error should not be empty")
	}
	if theme.Warning == "" {
		t.Error("DefaultTheme.Warning should not be empty")
	}
	if theme.Info == "" {
		t.Error("DefaultTheme.Info should not be empty")
	}
}

func TestNewStyles(t *testing.T) {
	theme := DefaultTheme
	styles := NewStyles(theme)

	// Test that styles are created and have colors assigned
	if styles.Success.GetForeground() == nil {
		t.Error("Success style should have a foreground color")
	}
	if styles.Error.GetForeground() == nil {
		t.Error("Error style should have a foreground color")
	}
	if styles.Warning.GetForeground() == nil {
		t.Error("Warning style should have a foreground color")
	}
	if styles.Info.GetForeground() == nil {
		t.Error("Info style should have a foreground color")
	}
}

func TestStylesCreation(t *testing.T) {
	styles := GetDefaultStyles()

	// Test text styles
	if styles.Success.GetForeground() == nil {
		t.Error("Success style should have foreground color")
	}
	if styles.Error.GetForeground() == nil {
		t.Error("Error style should have foreground color")
	}
	if styles.Warning.GetForeground() == nil {
		t.Error("Warning style should have foreground color")
	}
	if styles.Info.GetForeground() == nil {
		t.Error("Info style should have foreground color")
	}

	// Test emphasis styles
	if !styles.Bold.GetBold() {
		t.Error("Bold style should have bold enabled")
	}
	if !styles.Italic.GetItalic() {
		t.Error("Italic style should have italic enabled")
	}
	if !styles.Underline.GetUnderline() {
		t.Error("Underline style should have underline enabled")
	}

	// Test header styles
	if !styles.Header.GetBold() {
		t.Error("Header style should be bold")
	}
	if styles.Header.GetForeground() == nil {
		t.Error("Header style should have foreground color")
	}
}

func TestFangColorSchemeCreation(t *testing.T) {
	styles := GetDefaultStyles()
	colorScheme := styles.FangColorScheme(true)

	// Test that ColorScheme is created properly
	if colorScheme.Base == nil {
		t.Error("Base color should be set in color scheme")
	}
	if colorScheme.Title == nil {
		t.Error("Title color should be set in color scheme")
	}
	if colorScheme.Command == nil {
		t.Error("Command color should be set in color scheme")
	}
	if colorScheme.ErrorDetails == nil {
		t.Error("ErrorDetails color should be set in color scheme")
	}
}

func TestCustomTheme(t *testing.T) {
	customTheme := Theme{
		Primary:   "#FF0000", // Red
		Secondary: "#00FF00", // Green
		Accent:    "#0000FF", // Blue
		Success:   "#00FF00", // Green
		Warning:   "#FFFF00", // Yellow
		Error:     "#FF0000", // Red
		Info:      "#00FFFF", // Cyan
		Debug:     "#808080", // Gray
		Border:    "#C0C0C0", // Silver
		Highlight: "#F0F0F0", // Light gray
		Muted:     "#808080", // Gray
	}

	styles := NewStyles(customTheme)

	// Test that custom styles are created with colors
	if styles.Success.GetForeground() == nil {
		t.Error("Custom theme success style should have a color")
	}
	if styles.Error.GetForeground() == nil {
		t.Error("Custom theme error style should have a color")
	}
	if styles.Header.GetForeground() == nil {
		t.Error("Custom theme header style should have a color")
	}
}

func TestOutputLevel(t *testing.T) {
	tests := []struct {
		level    OutputLevel
		expected string
	}{
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelSuccess, "SUCCESS"},
		{LevelWarning, "WARNING"},
		{LevelError, "ERROR"},
		{OutputLevel(999), "UNKNOWN"}, // Invalid level
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.level.String()
			if result != tt.expected {
				t.Errorf("OutputLevel.String() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestColorMode(t *testing.T) {
	// Test that color modes are distinct
	if ColorAuto == ColorAlways {
		t.Error("ColorAuto and ColorAlways should be different")
	}
	if ColorAuto == ColorNever {
		t.Error("ColorAuto and ColorNever should be different")
	}
	if ColorAlways == ColorNever {
		t.Error("ColorAlways and ColorNever should be different")
	}
}

func TestThemeColorDefinitions(t *testing.T) {
	theme := DefaultTheme

	// Test that colors are valid hex colors or color names
	colors := map[string]string{
		"Primary":   theme.Primary,
		"Secondary": theme.Secondary,
		"Accent":    theme.Accent,
		"Success":   theme.Success,
		"Warning":   theme.Warning,
		"Error":     theme.Error,
		"Info":      theme.Info,
		"Debug":     theme.Debug,
		"Border":    theme.Border,
		"Highlight": theme.Highlight,
		"Muted":     theme.Muted,
	}

	for name, color := range colors {
		if color == "" {
			t.Errorf("Theme color %s should not be empty", name)
		}
		// Basic validation that it looks like a color (starts with # for hex)
		if len(color) > 0 && color[0] == '#' && len(color) != 7 {
			t.Errorf("Theme color %s appears to be invalid hex color: %s", name, color)
		}
	}
}

func TestStylesConsistency(t *testing.T) {
	styles := GetDefaultStyles()

	// Test that table styles are consistent
	if styles.TableHeader.GetForeground() == nil {
		t.Error("TableHeader should have a foreground color")
	}
	if !styles.TableHeader.GetBold() {
		t.Error("TableHeader should be bold")
	}

	// Test that list styles exist
	if styles.ListItem.String() == "" {
		// ListItem might have minimal styling, that's okay
	}
	if styles.ListBullet.GetForeground() == nil {
		t.Error("ListBullet should have a foreground color")
	}

	// Test that button styles have background
	if styles.Button.GetBackground() == nil {
		t.Error("Button should have a background color")
	}
	if styles.ButtonActive.GetBackground() == nil {
		t.Error("ButtonActive should have a background color")
	}

	// Test that code style has distinctive formatting
	if styles.Code.GetForeground() == nil {
		t.Error("Code should have a foreground color")
	}
}

func TestFangColorSchemeMapping(t *testing.T) {
	styles := GetDefaultStyles()

	t.Run("dark mode", func(t *testing.T) {
		colorScheme := styles.FangColorScheme(true)

		if colorScheme.Base == nil {
			t.Error("ColorScheme Base should not be nil")
		}
		if colorScheme.Title == nil {
			t.Error("ColorScheme Title should not be nil")
		}
		if colorScheme.Command == nil {
			t.Error("ColorScheme Command should not be nil")
		}
		if colorScheme.ErrorDetails == nil {
			t.Error("ColorScheme ErrorDetails should not be nil")
		}
		if colorScheme.Title != styles.Header.GetForeground() {
			t.Error("ColorScheme Title should match Header foreground")
		}
		if colorScheme.Codeblock == styles.Code.GetForeground() {
			t.Error("ColorScheme Codeblock must not use Code foreground as background")
		}
	})

	t.Run("light mode", func(t *testing.T) {
		colorScheme := styles.FangColorScheme(false)

		if colorScheme.Base == nil {
			t.Error("ColorScheme Base should not be nil")
		}
		if colorScheme.Title == nil {
			t.Error("ColorScheme Title should not be nil")
		}
		if colorScheme.Command == nil {
			t.Error("ColorScheme Command should not be nil")
		}
		if colorScheme.ErrorDetails == nil {
			t.Error("ColorScheme ErrorDetails should not be nil")
		}
	})

	t.Run("light mode Title matches Header foreground", func(t *testing.T) {
		colorScheme := styles.FangColorScheme(false)
		if colorScheme.Title != styles.Header.GetForeground() {
			t.Error("Light mode Title should match Header foreground")
		}
	})

	t.Run("dark and light produce different colors", func(t *testing.T) {
		dark := styles.FangColorScheme(true)
		light := styles.FangColorScheme(false)

		if dark.Codeblock == light.Codeblock {
			t.Error("Dark and light codeblock backgrounds should differ")
		}
		if dark.Base == light.Base {
			t.Error("Dark and light base text colors should differ")
		}
		if dark.Program == light.Program {
			t.Error("Dark and light program colors should differ")
		}
		if dark.Command == light.Command {
			t.Error("Dark and light command colors should differ")
		}
		if dark.Flag == light.Flag {
			t.Error("Dark and light flag colors should differ")
		}
		if dark.Description == light.Description {
			t.Error("Dark and light description colors should differ")
		}
		if dark.ErrorDetails == light.ErrorDetails {
			t.Error("Dark and light error details colors should differ")
		}
	})
}
