// ABOUTME: Advanced LSP features implementation for PSC
// ABOUTME: Provides go-to-definition, find references, formatting, and code actions

package lsp

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// findReferences finds all references to a symbol at the given position
func (s *Server) findReferences(doc *Document, pos Position, includeDeclaration bool) []Location {
	// Extract the symbol at the position
	symbol := s.extractSymbolAtPosition(doc.Text, pos)
	if symbol == "" {
		return []Location{}
	}

	s.logger.Printf("Finding references for symbol: %s (includeDecl: %v)", symbol, includeDeclaration)

	locations := []Location{}
	lines := strings.Split(doc.Text, "\n")

	// Find all occurrences of the symbol
	for lineNum, line := range lines {
		index := 0
		for {
			idx := strings.Index(line[index:], symbol)
			if idx == -1 {
				break
			}

			// Check if it's a whole word match
			startIdx := index + idx
			endIdx := startIdx + len(symbol)

			// Check word boundaries
			if (startIdx == 0 || !isWordChar(line[startIdx-1])) &&
				(endIdx == len(line) || !isWordChar(line[endIdx])) {

				// Check if this is a declaration
				isDeclaration := s.isDeclaration(line, symbol)

				if includeDeclaration || !isDeclaration {
					locations = append(locations, Location{
						URI: doc.URI,
						Range: Range{
							Start: Position{Line: lineNum, Character: startIdx},
							End:   Position{Line: lineNum, Character: endIdx},
						},
					})
				}
			}

			index = startIdx + 1
		}
	}

	return locations
}

// formatDocument formats the entire document
func (s *Server) formatDocument(doc *Document, options FormattingOptions) []TextEdit {
	// Try perltidy formatting first, fallback to basic formatting
	if formattedText, err := s.formatWithPerltidy(doc.Text, options); err == nil {
		// Perltidy succeeded, return the formatted result as a single edit
		if formattedText != doc.Text {
			return []TextEdit{
				{
					Range: Range{
						Start: Position{Line: 0, Character: 0},
						End:   s.getDocumentEnd(doc.Text),
					},
					NewText: formattedText,
				},
			}
		}
		return []TextEdit{} // No changes needed
	}

	// Fallback to basic line-by-line formatting
	edits := []TextEdit{}
	lines := strings.Split(doc.Text, "\n")

	for i, line := range lines {
		formatted := s.formatLine(line, options)
		if formatted != line {
			edits = append(edits, TextEdit{
				Range: Range{
					Start: Position{Line: i, Character: 0},
					End:   Position{Line: i, Character: len(line)},
				},
				NewText: formatted,
			})
		}
	}

	return edits
}

// formatWithPerltidy formats Perl code using perltidy
func (s *Server) formatWithPerltidy(text string, options FormattingOptions) (string, error) {
	// Check if perltidy is available
	if !s.isPerltidyAvailable() {
		return "", fmt.Errorf("perltidy not available")
	}

	// Create perltidy command with appropriate options
	args := s.buildPerltidyArgs(options)
	cmd := exec.Command("perltidy", args...)

	// Set up input and output
	var stdout, stderr bytes.Buffer
	cmd.Stdin = strings.NewReader(text)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run perltidy
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("perltidy failed: %v, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// isPerltidyAvailable checks if perltidy command is available
func (s *Server) isPerltidyAvailable() bool {
	_, err := exec.LookPath("perltidy")
	return err == nil
}

// buildPerltidyArgs builds command line arguments for perltidy based on formatting options
func (s *Server) buildPerltidyArgs(options FormattingOptions) []string {
	args := []string{
		"--standard-output",       // Output to stdout
		"--standard-error-output", // Error to stderr
		"--no-backup-files",       // Don't create backup files
	}

	// Configure indentation
	if options.TabSize > 0 {
		args = append(args, fmt.Sprintf("--indent-columns=%d", options.TabSize))
	}

	// Configure whether to use tabs or spaces
	if options.InsertSpaces {
		args = append(args, "--tabs") // Use spaces (perltidy default with --tabs is actually spaces)
	} else {
		args = append(args, "--entab-leading-whitespace=4") // Use tabs for leading whitespace
	}

	// Add other formatting preferences
	args = append(args,
		"--maximum-line-length=120",     // Reasonable line length
		"--paren-tightness=1",           // Moderate paren tightness
		"--brace-tightness=1",           // Moderate brace tightness
		"--square-bracket-tightness=1",  // Moderate bracket tightness
		"--continuation-indentation=2",  // Continuation indentation
		"--closing-token-indentation=0", // Standard closing token indentation
		"--comma-arrow-breakpoints=3",   // Smart comma-arrow breakpoints
	)

	return args
}

// getDocumentEnd returns the end position of the document
func (s *Server) getDocumentEnd(text string) Position {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return Position{Line: 0, Character: 0}
	}

	lastLineIndex := len(lines) - 1
	lastLineLength := len(lines[lastLineIndex])

	return Position{
		Line:      lastLineIndex,
		Character: lastLineLength,
	}
}

// generateCodeActions generates code actions for the given range and context
func (s *Server) generateCodeActions(doc *Document, rng Range, context CodeActionContext) []CodeAction {
	actions := []CodeAction{}

	// Generate quick fixes for diagnostics
	for _, diag := range context.Diagnostics {
		// Check if the diagnostic is within our range
		if s.rangeOverlaps(diag.Range, rng) {
			// Generate fix actions based on the diagnostic
			if strings.Contains(diag.Message, "undefined") {
				actions = append(actions, s.generateUndefinedVariableFix(doc, diag)...)
			} else if strings.Contains(diag.Message, "type mismatch") {
				actions = append(actions, s.generateTypeMismatchFix(doc, diag)...)
			}
		}
	}

	// Add refactoring actions
	// Extract variable
	if rng.Start.Line == rng.End.Line {
		text := s.getTextInRange(doc.Text, rng)
		if text != "" && !strings.Contains(text, "\n") {
			actions = append(actions, CodeAction{
				Title: "Extract variable",
				Kind:  "refactor.extract",
				Edit: &WorkspaceEdit{
					Changes: map[string][]TextEdit{
						doc.URI: s.generateExtractVariableEdits(doc, rng, text),
					},
				},
			})
		}
	}

	return actions
}

// Helper functions

func (s *Server) extractSymbolAtPosition(text string, pos Position) string {
	lines := strings.Split(text, "\n")
	if pos.Line >= len(lines) {
		return ""
	}

	line := lines[pos.Line]
	if pos.Character >= len(line) {
		return ""
	}

	// Find word boundaries
	start := pos.Character
	end := pos.Character

	// Include sigil for variables
	if start > 0 && (line[start-1] == '$' || line[start-1] == '@' || line[start-1] == '%') {
		start--
	}

	// Move start backwards to beginning of word
	for start > 0 && isWordChar(line[start-1]) {
		start--
	}

	// Include sigil if at the start
	if start > 0 && (line[start-1] == '$' || line[start-1] == '@' || line[start-1] == '%') {
		start--
	}

	// Move end forwards to end of word
	for end < len(line) && isWordChar(line[end]) {
		end++
	}

	if start == end {
		return ""
	}

	return line[start:end]
}

func (s *Server) isDeclaration(line, symbol string) bool {
	// Check if this line contains a declaration of the symbol
	declarationKeywords := []string{"my", "our", "state", "sub"}

	for _, keyword := range declarationKeywords {
		if strings.Contains(line, keyword) && strings.Contains(line, symbol) {
			// Simple heuristic - could be improved
			keywordIdx := strings.Index(line, keyword)
			symbolIdx := strings.Index(line, symbol)
			if keywordIdx != -1 && symbolIdx != -1 && keywordIdx < symbolIdx {
				return true
			}
		}
	}

	return false
}

func (s *Server) formatLine(line string, options FormattingOptions) string {
	// Basic formatting: trim trailing whitespace
	formatted := strings.TrimRight(line, " \t")

	// Convert tabs to spaces if requested
	if options.InsertSpaces {
		spaces := strings.Repeat(" ", options.TabSize)
		formatted = strings.ReplaceAll(formatted, "\t", spaces)
	}

	return formatted
}

func (s *Server) rangeOverlaps(r1, r2 Range) bool {
	// Check if two ranges overlap
	if r1.End.Line < r2.Start.Line || r2.End.Line < r1.Start.Line {
		return false
	}

	if r1.End.Line == r2.Start.Line && r1.End.Character < r2.Start.Character {
		return false
	}

	if r2.End.Line == r1.Start.Line && r2.End.Character < r1.Start.Character {
		return false
	}

	return true
}

func (s *Server) getTextInRange(text string, rng Range) string {
	lines := strings.Split(text, "\n")

	if rng.Start.Line >= len(lines) || rng.End.Line >= len(lines) {
		return ""
	}

	if rng.Start.Line == rng.End.Line {
		line := lines[rng.Start.Line]
		if rng.Start.Character >= len(line) || rng.End.Character > len(line) {
			return ""
		}
		return line[rng.Start.Character:rng.End.Character]
	}

	// Multi-line range - not supported for now
	return ""
}

func (s *Server) generateUndefinedVariableFix(doc *Document, diag Diagnostic) []CodeAction {
	actions := []CodeAction{}

	// Extract variable name from diagnostic message
	varName := s.extractVariableFromDiagnostic(diag.Message)
	if varName == "" {
		return actions
	}

	// Generate "Declare variable" action
	line := diag.Range.Start.Line
	actions = append(actions, CodeAction{
		Title:       fmt.Sprintf("Declare variable '%s'", varName),
		Kind:        "quickfix",
		Diagnostics: []Diagnostic{diag},
		Edit: &WorkspaceEdit{
			Changes: map[string][]TextEdit{
				doc.URI: {
					TextEdit{
						Range: Range{
							Start: Position{Line: line, Character: 0},
							End:   Position{Line: line, Character: 0},
						},
						NewText: fmt.Sprintf("my %s;\n", varName),
					},
				},
			},
		},
	})

	return actions
}

func (s *Server) generateTypeMismatchFix(doc *Document, diag Diagnostic) []CodeAction {
	var actions []CodeAction

	// Extract type mismatch information from diagnostic message
	// Common patterns: "Expected Int, got Str", "Type mismatch: cannot assign Str to Int variable"
	message := diag.Message

	// Parse expected and actual types from error message
	expectedType, actualType, ok := s.parseTypeMismatchMessage(message)
	if !ok {
		return actions
	}

	// Get the range where the error occurred
	errorRange := diag.Range
	errorText := s.getTextInRange(doc.Text, errorRange)

	// Generate different types of fixes

	// Fix 1: Add explicit type conversion
	if conversionFix := s.generateTypeConversionFix(expectedType, actualType, errorRange, errorText); conversionFix != nil {
		actions = append(actions, *conversionFix)
	}

	// Fix 2: Change variable declaration type
	if varTypeFix := s.generateVariableTypeChangeFix(doc, expectedType, actualType, errorRange); varTypeFix != nil {
		actions = append(actions, *varTypeFix)
	}

	// Fix 3: Add type annotation
	if annotationFix := s.generateAddTypeAnnotationFix(expectedType, errorRange, errorText); annotationFix != nil {
		actions = append(actions, *annotationFix)
	}

	// Fix 4: Cast using 'as' operator
	if castFix := s.generateTypeAssertionFix(expectedType, actualType, errorRange, errorText); castFix != nil {
		actions = append(actions, *castFix)
	}

	return actions
}

// Helper methods for type mismatch fixes

func (s *Server) parseTypeMismatchMessage(message string) (expectedType, actualType string, ok bool) {
	// Parse common error message patterns

	// Pattern 1: "Expected Int, got Str"
	if strings.Contains(message, "Expected") && strings.Contains(message, ", got ") {
		parts := strings.Split(message, "Expected ")
		if len(parts) > 1 {
			remainder := parts[1]
			gotIndex := strings.Index(remainder, ", got ")
			if gotIndex > 0 {
				expectedType = strings.TrimSpace(remainder[:gotIndex])
				actualType = strings.TrimSpace(remainder[gotIndex+6:])
				// Clean up any trailing punctuation
				actualType = strings.TrimSuffix(actualType, ".")
				actualType = strings.TrimSuffix(actualType, ",")
				return expectedType, actualType, true
			}
		}
	}

	// Pattern 2: "Type mismatch: cannot assign Str to Int variable"
	if strings.Contains(message, "cannot assign") && strings.Contains(message, " to ") {
		assignIndex := strings.Index(message, "cannot assign ")
		toIndex := strings.Index(message, " to ")
		if assignIndex >= 0 && toIndex > assignIndex {
			actualType = strings.TrimSpace(message[assignIndex+14 : toIndex])
			remainder := message[toIndex+4:]
			varIndex := strings.Index(remainder, " variable")
			if varIndex >= 0 {
				expectedType = strings.TrimSpace(remainder[:varIndex])
			} else {
				expectedType = strings.TrimSpace(remainder)
			}
			return expectedType, actualType, true
		}
	}

	// Pattern 3: "Type error: Int required but Str provided"
	if strings.Contains(message, " required but ") && strings.Contains(message, " provided") {
		requiredIndex := strings.Index(message, " required but ")
		providedIndex := strings.Index(message, " provided")
		if requiredIndex > 0 && providedIndex > requiredIndex {
			// Find the start of the expected type
			prefix := message[:requiredIndex]
			words := strings.Fields(prefix)
			if len(words) > 0 {
				expectedType = words[len(words)-1]
			}
			actualType = strings.TrimSpace(message[requiredIndex+14 : providedIndex])
			return expectedType, actualType, true
		}
	}

	return "", "", false
}

func (s *Server) generateTypeConversionFix(expectedType, actualType string, errorRange Range, errorText string) *CodeAction {
	// Generate conversion functions based on common type pairs
	var conversionExpr string

	switch {
	case expectedType == "Int" && actualType == "Str":
		conversionExpr = fmt.Sprintf("int(%s)", errorText)
	case expectedType == "Str" && actualType == "Int":
		conversionExpr = fmt.Sprintf("\"$%s\"", errorText)
	case expectedType == "Num" && actualType == "Str":
		conversionExpr = fmt.Sprintf("0 + %s", errorText)
	case expectedType == "Bool" && (actualType == "Int" || actualType == "Str"):
		conversionExpr = fmt.Sprintf("!!%s", errorText)
	case strings.HasPrefix(expectedType, "ArrayRef") && actualType == "Array":
		conversionExpr = fmt.Sprintf("\\@{%s}", errorText)
	case strings.HasPrefix(expectedType, "HashRef") && actualType == "Hash":
		conversionExpr = fmt.Sprintf("\\%%{%s}", errorText)
	default:
		// Generic conversion attempt
		conversionExpr = fmt.Sprintf("%s(%s)", expectedType, errorText)
	}

	if conversionExpr == "" {
		return nil
	}

	return &CodeAction{
		Title: fmt.Sprintf("Convert %s to %s", actualType, expectedType),
		Kind:  "quickfix",
		Edit: &WorkspaceEdit{
			Changes: map[string][]TextEdit{
				"": {
					{
						Range:   errorRange,
						NewText: conversionExpr,
					},
				},
			},
		},
	}
}

func (s *Server) generateVariableTypeChangeFix(doc *Document, expectedType, actualType string, errorRange Range) *CodeAction {
	// Find the variable declaration that needs type change
	// This is a simplified implementation - in practice, you'd need AST analysis
	lines := strings.Split(doc.Text, "\n")

	// Look backwards for variable declarations
	for i := errorRange.Start.Line; i >= 0 && i >= errorRange.Start.Line-10; i-- {
		if i >= len(lines) {
			continue
		}
		line := lines[i]

		// Look for patterns like "my $var" or "my Type $var"
		if strings.Contains(line, "my ") {
			// Simple pattern matching - could be improved with AST
			if strings.Contains(line, "$") {
				// Check if this line contains a type annotation
				if strings.Contains(line, fmt.Sprintf(" %s ", actualType)) {
					// Replace the type
					newLine := strings.Replace(line, fmt.Sprintf(" %s ", actualType), fmt.Sprintf(" %s ", expectedType), 1)
					return &CodeAction{
						Title: fmt.Sprintf("Change variable type from %s to %s", actualType, expectedType),
						Kind:  "quickfix",
						Edit: &WorkspaceEdit{
							Changes: map[string][]TextEdit{
								"": {
									{
										Range: Range{
											Start: Position{Line: i, Character: 0},
											End:   Position{Line: i, Character: len(line)},
										},
										NewText: newLine,
									},
								},
							},
						},
					}
				}
			}
		}
	}

	return nil
}

func (s *Server) generateAddTypeAnnotationFix(expectedType string, errorRange Range, errorText string) *CodeAction {
	// Add type annotation to variable declaration
	// Look for variable patterns and suggest adding type annotation

	if strings.HasPrefix(errorText, "my $") {
		// Pattern: "my $var" -> "my Type $var"
		newText := strings.Replace(errorText, "my $", fmt.Sprintf("my %s $", expectedType), 1)
		return &CodeAction{
			Title: fmt.Sprintf("Add %s type annotation", expectedType),
			Kind:  "quickfix",
			Edit: &WorkspaceEdit{
				Changes: map[string][]TextEdit{
					"": {
						{
							Range:   errorRange,
							NewText: newText,
						},
					},
				},
			},
		}
	}

	return nil
}

func (s *Server) generateTypeAssertionFix(expectedType, actualType string, errorRange Range, errorText string) *CodeAction {
	// Generate type assertion using 'as' operator
	assertionExpr := fmt.Sprintf("(%s as %s)", errorText, expectedType)

	return &CodeAction{
		Title: fmt.Sprintf("Assert type as %s", expectedType),
		Kind:  "quickfix",
		Edit: &WorkspaceEdit{
			Changes: map[string][]TextEdit{
				"": {
					{
						Range:   errorRange,
						NewText: assertionExpr,
					},
				},
			},
		},
	}
}

func (s *Server) generateExtractVariableEdits(doc *Document, rng Range, text string) []TextEdit {
	edits := []TextEdit{}

	// Generate a variable name
	varName := "$extracted_var"

	// Insert variable declaration before the current line
	edits = append(edits, TextEdit{
		Range: Range{
			Start: Position{Line: rng.Start.Line, Character: 0},
			End:   Position{Line: rng.Start.Line, Character: 0},
		},
		NewText: fmt.Sprintf("my %s = %s;\n", varName, text),
	})

	// Replace the selected text with the variable
	edits = append(edits, TextEdit{
		Range:   rng,
		NewText: varName,
	})

	return edits
}

func (s *Server) extractVariableFromDiagnostic(message string) string {
	// Simple extraction - look for variable patterns in the message
	// This is a heuristic and should be improved based on actual diagnostic messages

	if idx := strings.Index(message, "$"); idx != -1 {
		end := idx + 1
		for end < len(message) && isWordChar(message[end]) {
			end++
		}
		return message[idx:end]
	}

	return ""
}
