// ABOUTME: Advanced LSP features implementation for PSC
// ABOUTME: Provides go-to-definition, find references, formatting, and code actions

package lsp

import (
	"fmt"
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
	// For now, we'll implement basic formatting
	// TODO: Integrate with perltidy or implement more sophisticated formatting

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
	actions := []CodeAction{}

	// TODO: Implement type mismatch fixes
	// This could include:
	// - Adding type conversions
	// - Changing variable types
	// - Adding type annotations

	return actions
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
