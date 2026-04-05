// ABOUTME: Fang styling definitions for PVM CLI output
// ABOUTME: Provides consistent color schemes and formatting patterns

package ui

import (
	"image/color"

	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/lipgloss/v2"
)

// Theme defines the color and style theme for PVM
type Theme struct {
	// Base colors
	Primary   string
	Secondary string
	Accent    string

	// Status colors
	Success string
	Warning string
	Error   string
	Info    string
	Debug   string

	// UI elements
	Border    string
	Highlight string
	Muted     string
}

// DefaultTheme provides the default PVM color scheme
var DefaultTheme = Theme{
	Primary:   "#7C3AED", // Purple
	Secondary: "#3B82F6", // Blue
	Accent:    "#10B981", // Green

	Success: "#10B981", // Green
	Warning: "#F59E0B", // Amber
	Error:   "#EF4444", // Red
	Info:    "#3B82F6", // Blue
	Debug:   "#6B7280", // Gray

	Border:    "#E5E7EB", // Light gray
	Highlight: "#F3F4F6", // Very light gray
	Muted:     "#9CA3AF", // Medium gray
}

// Styles contains all the lipgloss styles used throughout PVM
type Styles struct {
	// Text styles
	Success lipgloss.Style
	Error   lipgloss.Style
	Warning lipgloss.Style
	Info    lipgloss.Style
	Debug   lipgloss.Style

	// Emphasis styles
	Bold      lipgloss.Style
	Italic    lipgloss.Style
	Underline lipgloss.Style
	Code      lipgloss.Style

	// Layout styles
	Header    lipgloss.Style
	SubHeader lipgloss.Style
	Section   lipgloss.Style
	Box       lipgloss.Style

	// Interactive styles
	Button       lipgloss.Style
	ButtonActive lipgloss.Style
	Link         lipgloss.Style

	// Table styles
	TableHeader lipgloss.Style
	TableCell   lipgloss.Style
	TableBorder lipgloss.Style

	// List styles
	ListItem   lipgloss.Style
	ListBullet lipgloss.Style
	ListNumber lipgloss.Style

	// Additional utility styles
	Muted lipgloss.Style
}

// NewStyles creates a new set of styles with the given theme
func NewStyles(theme Theme) Styles {
	return Styles{
		// Text styles
		Success: lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Success)),
		Error:   lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Error)),
		Warning: lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Warning)),
		Info:    lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Info)),
		Debug:   lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Debug)),

		// Emphasis styles
		Bold:      lipgloss.NewStyle().Bold(true),
		Italic:    lipgloss.NewStyle().Italic(true),
		Underline: lipgloss.NewStyle().Underline(true),
		Code:      lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Accent)).Background(lipgloss.Color(theme.Highlight)).Padding(0, 1),

		// Layout styles
		Header: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Primary)).
			Bold(true).
			Padding(1, 0),
		SubHeader: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Secondary)).
			Bold(true),
		Section: lipgloss.NewStyle().
			Padding(1, 0),
		Box: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(theme.Border)).
			Padding(1, 2),

		// Interactive styles
		Button: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color(theme.Primary)).
			Padding(0, 2).
			Border(lipgloss.RoundedBorder()),
		ButtonActive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color(theme.Accent)).
			Padding(0, 2).
			Border(lipgloss.RoundedBorder()).
			Bold(true),
		Link: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Primary)).
			Underline(true),

		// Table styles
		TableHeader: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Primary)).
			Bold(true).
			Padding(0, 1),
		TableCell: lipgloss.NewStyle().
			Padding(0, 1),
		TableBorder: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Border)),

		// List styles
		ListItem: lipgloss.NewStyle().
			Padding(0, 1),
		ListBullet: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Primary)),
		ListNumber: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Secondary)),

		// Additional utility styles
		Muted: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Muted)),
	}
}

// FangColorScheme creates a Fang-compatible color scheme adapted for the
// terminal's background brightness. Pass isDark=true for dark terminals,
// isDark=false for light terminals.
//
// Fang uses Codeblock as a BACKGROUND color for code example blocks, so it
// must contrast with the terminal background. All foreground colors rendered
// inside the codeblock must contrast against the codeblock background.
//
// Colors are chosen to be distinguishable under common forms of color
// blindness (protanopia, deuteranopia): we avoid pairing red and green,
// and rely on blue/purple/amber hue differences instead.
func (s Styles) FangColorScheme(isDark bool) fang.ColorScheme {
	c := lipgloss.LightDark(isDark)

	return fang.ColorScheme{
		Base:           c(lipgloss.Color("#3A3943"), lipgloss.Color("#DFDBDD")),                                                                          // Body text
		Title:          s.Header.GetForeground(),                                                                                                         // Purple — section headers
		Description:    c(lipgloss.Color("#605F6B"), lipgloss.Color("#BFBCC8")),                                                                          // Flag/command descriptions
		Codeblock:      c(lipgloss.Color("#F0EDE8"), lipgloss.Color("#2D2B36")),                                                                          // Code block background
		Program:        c(lipgloss.Color("#4A30D9"), lipgloss.Color("#9B8AFF")),                                                                          // Purple — program name
		DimmedArgument: c(lipgloss.Color("#6B6A78"), s.Muted.GetForeground()),                                                                            // Gray — optional args
		Comment:        c(lipgloss.Color("#6B6A78"), lipgloss.Color("#9896A6")),                                                                          // Gray — code comments
		Flag:           c(lipgloss.Color("#785A0F"), lipgloss.Color("#F5C542")),                                                                          // Amber — flags
		FlagDefault:    c(lipgloss.Color("#6B6A78"), s.Muted.GetForeground()),                                                                            // Gray — default values
		Command:        c(lipgloss.Color("#1A5276"), lipgloss.Color("#5EB5FF")),                                                                          // Blue — subcommands
		QuotedString:   c(lipgloss.Color("#3A3943"), lipgloss.Color("#DFDBDD")),                                                                          // Neutral — quoted strings
		Argument:       c(lipgloss.Color("#3A3943"), lipgloss.Color("#DFDBDD")),                                                                          // Neutral — argument values
		Help:           c(lipgloss.Color("#3A3943"), lipgloss.Color("#DFDBDD")),                                                                          // Neutral — help text
		Dash:           c(lipgloss.Color("#6B6A78"), s.Muted.GetForeground()),                                                                            // Gray — flag dashes
		ErrorHeader:    [2]color.Color{c(lipgloss.Color("#FFFAF1"), lipgloss.Color("#FFFAF1")), c(lipgloss.Color("#B0304A"), lipgloss.Color("#C43A5A"))}, // Error badge
		ErrorDetails:   c(lipgloss.Color("#B0304A"), lipgloss.Color("#F08090")),                                                                          // Error text
	}
}

// GetDefaultStyles returns the default PVM styles
func GetDefaultStyles() Styles {
	return NewStyles(DefaultTheme)
}
