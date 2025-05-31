// ABOUTME: Debug test for AST node types to troubleshoot subroutine extraction
// ABOUTME: Walks parsed AST to show actual node types and structure for debugging

package embeddings

import (
	"fmt"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parser"
)

// TestDebugASTNodeTypes creates a test file and walks through the AST to show
// what node types are actually being produced by the parser
func TestDebugASTNodeTypes(t *testing.T) {
	// Test Perl code with subroutines
	testCode := `sub hello {
    print "Hello!\n";
}

sub goodbye {
    print "Goodbye!\n";
}`

	t.Logf("Parsing test code:\n%s", testCode)

	// Use the parser pool the same way the extractor does
	result, err := parser.PooledParserFunc(func(p parser.Parser) (string, error) {
		ast, err := p.ParseString(testCode)
		if err != nil {
			return "", fmt.Errorf("failed to parse: %v", err)
		}

		if ast == nil {
			return "", fmt.Errorf("AST is nil")
		}

		if ast.Root == nil {
			return "", fmt.Errorf("AST root is nil")
		}

		// Walk the AST and collect debug information
		var debugInfo strings.Builder
		debugInfo.WriteString("AST Analysis:\n")
		debugInfo.WriteString(fmt.Sprintf("Source length: %d\n", len(ast.Source)))
		debugInfo.WriteString(fmt.Sprintf("Source: %q\n", ast.Source))
		debugInfo.WriteString("\nComplete AST Structure (deeper analysis):\n")

		depth := 0
		walkNodeDebugDeep(ast.Root, &debugInfo, &depth, 0)

		// Also try to extract text using source positions
		debugInfo.WriteString("\nExtracting text using source positions:\n")
		extractTextFromPositions(ast, &debugInfo)

		return debugInfo.String(), nil
	})

	if err != nil {
		t.Fatalf("Error during parsing: %v", err)
	}

	t.Logf("AST Debug Information:\n%s", result)

	// Now test what the extractor would see by simulating its AST walking
	t.Logf("\n--- Testing Extractor AST Walking ---")
	extractorResult, err := parser.PooledParserFunc(func(p parser.Parser) (string, error) {
		ast, err := p.ParseString(testCode)
		if err != nil {
			return "", err
		}

		// Simulate what the extractor does - look for specific node types
		var extractorDebug strings.Builder
		extractorDebug.WriteString("Extractor node type search:\n")

		targetTypes := []string{
			"subroutine_declaration_statement",
			"subroutine_declaration",
			"method_declaration_statement",
			"method_declaration",
			"class_declaration",
			"package_declaration",
		}

		for _, targetType := range targetTypes {
			count := 0
			walkNodeForType(ast.Root, targetType, &count, &extractorDebug)
			extractorDebug.WriteString(fmt.Sprintf("Found %d nodes of type '%s'\n", count, targetType))
		}

		return extractorDebug.String(), nil
	})

	if err != nil {
		t.Fatalf("Error during extractor simulation: %v", err)
	}

	t.Logf("%s", extractorResult)
}

// walkNodeDebug recursively walks AST nodes and logs their types and structure
func walkNodeDebug(node parser.Node, debug *strings.Builder, depth *int) {
	if node == nil {
		return
	}

	// Indent based on depth
	indent := strings.Repeat("  ", *depth)

	// Get node information
	nodeType := node.Type()
	nodeText := truncateContent(node.Text(), 30)

	// Write node information
	debug.WriteString(fmt.Sprintf("%s%s: %q", indent, nodeType, nodeText))

	// Add position information if available
	if start := node.Start(); start.Line > 0 {
		debug.WriteString(fmt.Sprintf(" [%d:%d-%d:%d]",
			start.Line, start.Column, node.End().Line, node.End().Column))
	}

	debug.WriteString("\n")

	// Visit children
	children := node.Children()
	if len(children) > 0 {
		*depth++
		for _, child := range children {
			walkNodeDebug(child, debug, depth)
		}
		*depth--
	}
}

// walkNodeDebugDeep recursively walks AST nodes with deeper analysis, including source positions
func walkNodeDebugDeep(node parser.Node, debug *strings.Builder, depth *int, nodeIndex int) {
	if node == nil {
		return
	}

	// Indent based on depth
	indent := strings.Repeat("  ", *depth)

	// Get node information
	nodeType := node.Type()
	nodeText := node.Text() // Get full text, not truncated

	// Write detailed node information
	debug.WriteString(fmt.Sprintf("%s[%d] %s:\n", indent, nodeIndex, nodeType))
	debug.WriteString(fmt.Sprintf("%s    Text: %q\n", indent, nodeText))
	debug.WriteString(fmt.Sprintf("%s    Position: %d:%d-%d:%d\n", indent,
		node.Start().Line, node.Start().Column, node.End().Line, node.End().Column))

	// Visit children with detailed analysis
	children := node.Children()
	debug.WriteString(fmt.Sprintf("%s    Children: %d\n", indent, len(children)))

	if len(children) > 0 {
		*depth++
		for i, child := range children {
			walkNodeDebugDeep(child, debug, depth, i)
		}
		*depth--
	}
}

// walkNodeForType walks the AST looking for nodes of a specific type
func walkNodeForType(node parser.Node, targetType string, count *int, debug *strings.Builder) {
	if node == nil {
		return
	}

	if node.Type() == targetType {
		*count++
		debug.WriteString(fmt.Sprintf("  Found %s: %q\n", targetType, truncateContent(node.Text(), 40)))
	}

	for _, child := range node.Children() {
		walkNodeForType(child, targetType, count, debug)
	}
}

// truncateContent truncates content to a maximum length for display
func truncateContent(content string, maxLen int) string {
	// Remove newlines and excessive whitespace for display
	content = strings.ReplaceAll(content, "\n", "\\n")
	content = strings.ReplaceAll(content, "\t", "\\t")

	if len(content) <= maxLen {
		return content
	}

	return content[:maxLen-3] + "..."
}

// TestDebugSubDeclStructure examines the internal structure of sub_decl nodes
func TestDebugSubDeclStructure(t *testing.T) {
	testCode := `sub hello {
    print "Hello!\n";
}

sub goodbye {
    print "Goodbye!\n";
}`

	// Parse and examine sub_decl nodes in detail
	result, err := parser.PooledParserFunc(func(p parser.Parser) (string, error) {
		ast, err := p.ParseString(testCode)
		if err != nil {
			return "", err
		}

		var debug strings.Builder
		debug.WriteString("Detailed sub_decl node analysis:\n")

		// Find and examine each sub_decl node
		subDeclCount := 0
		walkNodeExamineSubDecl(ast.Root, &debug, &subDeclCount)

		return debug.String(), nil
	})

	if err != nil {
		t.Fatalf("Error during parsing: %v", err)
	}

	t.Logf("%s", result)
}

// walkNodeExamineSubDecl walks the AST and examines sub_decl nodes in detail
func walkNodeExamineSubDecl(node parser.Node, debug *strings.Builder, count *int) {
	if node == nil {
		return
	}

	if node.Type() == "sub_decl" {
		*count++
		debug.WriteString(fmt.Sprintf("\n=== sub_decl #%d ===\n", *count))
		debug.WriteString(fmt.Sprintf("Text: %q\n", node.Text()))
		debug.WriteString(fmt.Sprintf("Position: %d:%d-%d:%d\n",
			node.Start().Line, node.Start().Column, node.End().Line, node.End().Column))

		debug.WriteString("Children:\n")
		for i, child := range node.Children() {
			debug.WriteString(fmt.Sprintf("  [%d] %s: %q\n", i, child.Type(), truncateContent(child.Text(), 30)))

			// If this child might contain the name, examine its children too
			if child.Type() == "sub_name" || child.Type() == "name" || child.Type() == "identifier" {
				debug.WriteString(fmt.Sprintf("    -> This looks like the name!\n"))
			}

			// Examine grandchildren for potential name nodes
			for j, grandchild := range child.Children() {
				debug.WriteString(fmt.Sprintf("    [%d.%d] %s: %q\n", i, j, grandchild.Type(), truncateContent(grandchild.Text(), 20)))
			}
		}
	}

	for _, child := range node.Children() {
		walkNodeExamineSubDecl(child, debug, count)
	}
}

// TestDebugSpecificNodeTypes looks for specific node types that should contain subroutines
func TestDebugSpecificNodeTypes(t *testing.T) {
	testCode := `sub hello {
    print "Hello!\n";
}

sub goodbye {
    print "Goodbye!\n";
}`

	// Parse and look for specific node types
	foundNodeTypes, err := parser.PooledParserFunc(func(p parser.Parser) (map[string]int, error) {
		ast, err := p.ParseString(testCode)
		if err != nil {
			return nil, err
		}

		nodeTypes := make(map[string]int)
		countNodeTypes(ast.Root, nodeTypes)

		return nodeTypes, nil
	})

	if err != nil {
		t.Fatalf("Error during parsing: %v", err)
	}

	t.Logf("Node type counts:")
	for nodeType, count := range foundNodeTypes {
		t.Logf("  %s: %d", nodeType, count)
	}

	// Check for subroutine-related node types
	subroutineTypes := []string{
		"subroutine_declaration_statement",
		"subroutine_declaration",
		"sub_declaration",
		"subroutine",
		"sub",
		"function_declaration",
		"method_declaration_statement",
		"method_declaration",
	}

	t.Logf("\nLooking for subroutine-related node types:")
	foundAny := false
	for _, expectedType := range subroutineTypes {
		if count, exists := foundNodeTypes[expectedType]; exists {
			t.Logf("  Found %s: %d", expectedType, count)
			foundAny = true
		} else {
			t.Logf("  Missing: %s", expectedType)
		}
	}

	if !foundAny {
		t.Logf("\nNo subroutine-related node types found!")
		t.Logf("This explains why the extractor isn't finding subroutines.")
		t.Logf("All found node types: %v", getKeys(foundNodeTypes))
	}
}

// countNodeTypes recursively counts all node types in the AST
func countNodeTypes(node parser.Node, counts map[string]int) {
	if node == nil {
		return
	}

	nodeType := node.Type()
	counts[nodeType]++

	for _, child := range node.Children() {
		countNodeTypes(child, counts)
	}
}

// extractTextFromPositions tries to extract text from nodes using source positions
func extractTextFromPositions(ast *parser.AST, debug *strings.Builder) {
	if ast.Source == "" {
		debug.WriteString("No source text available\n")
		return
	}

	lines := strings.Split(ast.Source, "\n")

	walkNodeExtractText(ast.Root, lines, debug, 0)
}

// walkNodeExtractText walks nodes and extracts their text from source using positions
func walkNodeExtractText(node parser.Node, lines []string, debug *strings.Builder, depth int) {
	if node == nil {
		return
	}

	indent := strings.Repeat("  ", depth)
	start := node.Start()
	end := node.End()

	if start.Line > 0 && start.Line <= len(lines) {
		// Extract text from source using line/column positions
		text := ""
		if start.Line == end.Line {
			// Single line
			line := lines[start.Line-1] // Lines are 1-indexed
			// Try 0-based columns first
			if start.Column < len(line) && end.Column <= len(line) {
				text = line[start.Column:end.Column]
			}
		} else {
			// Multi-line - extract full text range
			var parts []string
			for i := start.Line - 1; i < end.Line && i < len(lines); i++ {
				line := lines[i]
				if i == start.Line-1 {
					// First line - start from column
					if start.Column < len(line) {
						parts = append(parts, line[start.Column:])
					}
				} else if i == end.Line-1 {
					// Last line - end at column
					if end.Column <= len(line) {
						parts = append(parts, line[:end.Column])
					}
				} else {
					// Middle lines - full line
					parts = append(parts, line)
				}
			}
			text = strings.Join(parts, "\n")
		}

		debug.WriteString(fmt.Sprintf("%s%s: %q (extracted from source)\n",
			indent, node.Type(), text))
	}

	for _, child := range node.Children() {
		walkNodeExtractText(child, lines, debug, depth+1)
	}
}

// TestExtractorWithFixedNodeTypes tests that the extractor now works with sub_decl nodes
func TestExtractorWithFixedNodeTypes(t *testing.T) {
	testCode := `sub hello {
    print "Hello!\n";
}

sub goodbye {
    print "Goodbye!\n";
}`

	// Test with a real extractor using our debug pattern
	extractorResult, err := parser.PooledParserFunc(func(p parser.Parser) (string, error) {
		ast, err := p.ParseString(testCode)
		if err != nil {
			return "", err
		}

		// Create extractor and simulate extraction with debug
		extractor := &Extractor{}

		var result strings.Builder
		result.WriteString("Debug: Walking AST for sub_decl nodes:\n")

		// First, let's see what nodes we find when walking
		subDeclCount := 0
		extractor.walkNode(ast.Root, func(node parser.Node) bool {
			if node.Type() == "sub_decl" {
				subDeclCount++
				result.WriteString(fmt.Sprintf("Found sub_decl #%d:\n", subDeclCount))

				// Test name extraction
				name := extractor.extractSubroutineNameWithSource(node, ast.Source)
				result.WriteString(fmt.Sprintf("  Extracted name: %q\n", name))

				// Test content extraction
				content := extractor.extractNodeTextFromSource(node, ast.Source)
				result.WriteString(fmt.Sprintf("  Extracted content: %q\n", truncateContent(content, 50)))
			}
			return true
		})

		result.WriteString(fmt.Sprintf("\nFound %d sub_decl nodes\n\n", subDeclCount))

		// Now run the actual extraction
		blocks := extractor.extractBlocksFromAST(ast, "test-project", "test.pl")

		result.WriteString(fmt.Sprintf("Extracted %d blocks:\n", len(blocks)))
		for i, block := range blocks {
			result.WriteString(fmt.Sprintf("Block %d:\n", i+1))
			result.WriteString(fmt.Sprintf("  ID: %s\n", block.ID))
			result.WriteString(fmt.Sprintf("  Type: %s\n", block.Type))
			result.WriteString(fmt.Sprintf("  Name: %s\n", block.Name))
			result.WriteString(fmt.Sprintf("  Content: %q\n", truncateContent(block.Content, 100)))
			result.WriteString(fmt.Sprintf("  Lines: %d-%d\n", block.StartLine, block.EndLine))
		}

		return result.String(), nil
	})

	if err != nil {
		t.Fatalf("Error during extraction test: %v", err)
	}

	t.Logf("Extraction test results:\n%s", extractorResult)

	// Parse the results to verify we found subroutines
	if !strings.Contains(extractorResult, "hello") {
		t.Error("Expected to find 'hello' subroutine")
	}

	if !strings.Contains(extractorResult, "goodbye") {
		t.Error("Expected to find 'goodbye' subroutine")
	}

	if !strings.Contains(extractorResult, "function") {
		t.Error("Expected blocks to be of type 'function'")
	}
}

// getKeys returns the keys of a map as a slice
func getKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
