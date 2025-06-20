// ABOUTME: Core output methods for PVM UI framework
// ABOUTME: Implements Fang-powered CLI output with consistent styling

package ui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
)

// Output provides the main UI output functionality
type Output struct {
	context *UIContext
	styles  Styles
}

// NewOutput creates a new output instance with the given context
func NewOutput(ctx *UIContext) *Output {
	if ctx == nil {
		ctx = &UIContext{
			Writer:      os.Stdout,
			ErrorWriter: os.Stderr,
			ColorMode:   ColorAuto,
			Quiet:       false,
			Verbose:     false,
			Interactive: true,
		}
	}

	styles := GetDefaultStyles()

	return &Output{
		context: ctx,
		styles:  styles,
	}
}

// NewDefaultOutput creates a new output instance with default settings
func NewDefaultOutput() *Output {
	return NewOutput(nil)
}

// Info displays an informational message
func (o *Output) Info(message string, args ...interface{}) {
	if o.context.Quiet {
		return
	}

	formatted := fmt.Sprintf(message, args...)
	styled := o.renderWithColorMode(o.styles.Info, "ℹ ") + formatted
	o.safeWriteln(o.context.Writer, styled)
}

// Success displays a success message
func (o *Output) Success(message string, args ...interface{}) {
	if o.context.Quiet {
		return
	}

	formatted := fmt.Sprintf(message, args...)
	styled := o.renderWithColorMode(o.styles.Success, "✓ ") + formatted
	o.safeWriteln(o.context.Writer, styled)
}

// Warning displays a warning message
func (o *Output) Warning(message string, args ...interface{}) {
	if o.context.Quiet {
		return
	}

	formatted := fmt.Sprintf(message, args...)
	styled := o.renderWithColorMode(o.styles.Warning, "⚠ ") + formatted
	o.safeWriteln(o.context.Writer, styled)
}

// Error displays an error message
func (o *Output) Error(message string, args ...interface{}) {
	formatted := fmt.Sprintf(message, args...)
	styled := o.renderWithColorMode(o.styles.Error, "✗ ") + formatted
	writer := o.context.ErrorWriter
	if writer == nil {
		writer = o.context.Writer
	}
	o.safeWriteln(writer, styled)
}

// Debug displays a debug message (only if verbose mode is enabled)
func (o *Output) Debug(message string, args ...interface{}) {
	if !o.context.Verbose || o.context.Quiet {
		return
	}

	formatted := fmt.Sprintf(message, args...)
	styled := o.renderWithColorMode(o.styles.Debug, "🐛 ") + formatted
	o.safeWriteln(o.context.Writer, styled)
}

// Printf provides formatted output
func (o *Output) Printf(format string, args ...interface{}) {
	if o.context.Quiet {
		return
	}

	if o.context.Writer != nil {
		fmt.Fprintf(o.context.Writer, format, args...)
	}
}

// Println provides line output
func (o *Output) Println(args ...interface{}) {
	if o.context.Quiet {
		return
	}

	if o.context.Writer != nil {
		fmt.Fprintln(o.context.Writer, args...)
	}
}

// Table displays data in a table format
func (o *Output) Table(headers []string, rows [][]string) {
	if o.context.Quiet {
		return
	}

	o.TableWithOptions(TableOptions{
		Headers:     headers,
		Rows:        rows,
		ShowBorders: true,
	})
}

// TableWithOptions displays a table with custom options
func (o *Output) TableWithOptions(opts TableOptions) {
	if o.context.Quiet {
		return
	}

	if opts.Title != "" {
		titleStyled := o.styles.SubHeader.Render(opts.Title)
		o.safeWriteln(o.context.Writer, titleStyled)
		o.safeWriteln(o.context.Writer, "")
	}

	// Calculate column widths
	colWidths := make([]int, len(opts.Headers))
	for i, header := range opts.Headers {
		colWidths[i] = len(header)
	}
	for _, row := range opts.Rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Write headers if present
	if len(opts.Headers) > 0 {
		var headerRow strings.Builder
		for i, header := range opts.Headers {
			styled := o.styles.TableHeader.Render(fmt.Sprintf("%-*s", colWidths[i], header))
			headerRow.WriteString(styled)
			if i < len(opts.Headers)-1 {
				headerRow.WriteString("  ")
			}
		}
		o.safeWriteln(o.context.Writer, headerRow.String())

		// Write separator
		var separator strings.Builder
		for i, width := range colWidths {
			separator.WriteString(strings.Repeat("-", width))
			if i < len(colWidths)-1 {
				separator.WriteString("  ")
			}
		}
		o.safeWriteln(o.context.Writer, separator.String())
	}

	// Write rows
	for _, row := range opts.Rows {
		var rowStr strings.Builder
		for i, cell := range row {
			if i < len(colWidths) {
				styled := o.styles.TableCell.Render(fmt.Sprintf("%-*s", colWidths[i], cell))
				rowStr.WriteString(styled)
				if i < len(row)-1 {
					rowStr.WriteString("  ")
				}
			}
		}
		o.safeWriteln(o.context.Writer, rowStr.String())
	}
	o.safeWriteln(o.context.Writer, "") // Add spacing after table
}

// List displays items in a list format
func (o *Output) List(items []string) {
	if o.context.Quiet {
		return
	}

	o.ListWithOptions(ListOptions{
		Items:      items,
		Numbered:   false,
		BulletChar: "•",
	})
}

// ListWithOptions displays a list with custom options
func (o *Output) ListWithOptions(opts ListOptions) {
	if o.context.Quiet {
		return
	}

	if opts.Title != "" {
		titleStyled := o.styles.SubHeader.Render(opts.Title)
		o.safeWriteln(o.context.Writer, titleStyled)
		o.safeWriteln(o.context.Writer, "")
	}

	for i, item := range opts.Items {
		if opts.Numbered {
			number := o.styles.ListNumber.Render(fmt.Sprintf("%d.", i+1))
			if o.context.Writer != nil {
				fmt.Fprintf(o.context.Writer, "%s %s\n", number, item)
			}
		} else {
			bullet := opts.BulletChar
			if bullet == "" {
				bullet = "•"
			}
			bulletStyled := o.styles.ListBullet.Render(bullet)
			if o.context.Writer != nil {
				fmt.Fprintf(o.context.Writer, "%s %s\n", bulletStyled, item)
			}
		}
	}
	o.safeWriteln(o.context.Writer, "") // Add spacing after list
}

// KeyValue displays key-value pairs
func (o *Output) KeyValue(pairs map[string]string) {
	if o.context.Quiet {
		return
	}

	for key, value := range pairs {
		keyStyled := o.styles.Bold.Render(key + ":")
		if o.context.Writer != nil {
			fmt.Fprintf(o.context.Writer, "  %-20s %s\n", keyStyled, value)
		}
	}
}

// Status displays a status message (typically for ongoing operations)
func (o *Output) Status(message string) {
	if o.context.Quiet {
		return
	}

	styled := o.styles.Info.Render("→ ") + message
	o.safeWriteln(o.context.Writer, styled)
}

// Progress displays progress information
func (o *Output) Progress(current, total int, message string) {
	if o.context.Quiet {
		return
	}

	percentage := int((float64(current) / float64(total)) * 100)
	progress := fmt.Sprintf("[%d/%d] (%d%%)", current, total, percentage)

	styled := o.styles.Info.Render("⚡ ") + message + " " + o.styles.Muted.Render(progress)
	o.safeWriteln(o.context.Writer, styled)
}

// Header displays a formatted header
func (o *Output) Header(title string) {
	if o.context.Quiet {
		return
	}

	styled := o.styles.Header.Render(title)
	o.safeWriteln(o.context.Writer, styled)
}

// SubHeader displays a formatted sub-header
func (o *Output) SubHeader(title string) {
	if o.context.Quiet {
		return
	}

	styled := o.styles.SubHeader.Render(title)
	o.safeWriteln(o.context.Writer, styled)
}

// Section displays a formatted section with content
func (o *Output) Section(title, content string) {
	if o.context.Quiet {
		return
	}

	o.SubHeader(title)
	o.safeWriteln(o.context.Writer, content)
	o.safeWriteln(o.context.Writer, "") // Add spacing
}

// Box displays content in a bordered box
func (o *Output) Box(content string) {
	if o.context.Quiet {
		return
	}

	styled := o.styles.Box.Render(content)
	o.safeWriteln(o.context.Writer, styled)
}

// Markdown renders basic markdown-style content
func (o *Output) Markdown(content string) {
	if o.context.Quiet {
		return
	}

	// Basic markdown parsing for headers and emphasis
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			o.safeWriteln(o.context.Writer, "")
			continue
		}

		// Handle headers
		switch {
		case strings.HasPrefix(line, "# "):
			title := strings.TrimPrefix(line, "# ")
			styled := o.styles.Header.Render(title)
			o.safeWriteln(o.context.Writer, styled)
		case strings.HasPrefix(line, "## "):
			title := strings.TrimPrefix(line, "## ")
			styled := o.styles.SubHeader.Render(title)
			o.safeWriteln(o.context.Writer, styled)
		case strings.HasPrefix(line, "- "):
			item := strings.TrimPrefix(line, "- ")
			bullet := o.styles.ListBullet.Render("•")
			if o.context.Writer != nil {
				fmt.Fprintf(o.context.Writer, "%s %s\n", bullet, item)
			}
		default:
			// Handle basic emphasis (this is simplified)
			if strings.Contains(line, "**") && strings.Count(line, "**") >= 2 {
				// Simple bold handling
				parts := strings.Split(line, "**")
				for i, part := range parts {
					if i%2 == 1 {
						if o.context.Writer != nil {
							fmt.Fprint(o.context.Writer, o.styles.Bold.Render(part))
						}
					} else {
						if o.context.Writer != nil {
							fmt.Fprint(o.context.Writer, part)
						}
					}
				}
				o.safeWriteln(o.context.Writer, "")
			} else {
				o.safeWriteln(o.context.Writer, line)
			}
		}
	}
}

// SetWriter changes the output writer
func (o *Output) SetWriter(w io.Writer) {
	o.context.Writer = w
}

// SetQuiet sets quiet mode
func (o *Output) SetQuiet(quiet bool) {
	o.context.Quiet = quiet
}

// SetVerbose sets verbose mode
func (o *Output) SetVerbose(verbose bool) {
	o.context.Verbose = verbose
}

// SetColorMode sets the color mode
func (o *Output) SetColorMode(mode ColorMode) {
	o.context.ColorMode = mode
}

// Context returns the current UI context
func (o *Output) Context() *UIContext {
	return o.context
}

// safeWrite writes to the writer if it's not nil, otherwise does nothing
func (o *Output) safeWrite(writer io.Writer, message string) {
	if writer != nil {
		fmt.Fprint(writer, message)
	}
}

// safeWriteln writes to the writer with newline if it's not nil, otherwise does nothing
func (o *Output) safeWriteln(writer io.Writer, message string) {
	if writer != nil {
		fmt.Fprintln(writer, message)
	}
}

// renderWithColorMode renders text with a style only if color mode allows it
func (o *Output) renderWithColorMode(style lipgloss.Style, text string) string {
	if o.context.ColorMode == ColorNever {
		return text
	}
	return style.Render(text)
}
