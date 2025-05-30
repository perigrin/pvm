// ABOUTME: Fast parser implementation optimized for common Perl patterns
// ABOUTME: Reduces tree-sitter overhead by using heuristic-based parsing for simple cases

package performance

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/parser"
)

// SimpleNode implements ast.Node for fast parsing
type SimpleNode struct {
	NodeType    string
	ChildNodes  []ast.Node
	Attributes  map[string]string
	StartPos    ast.Position
	EndPos      ast.Position
	ParentNode  ast.Node
	TextContent string
}

// Type returns the node type
func (sn *SimpleNode) Type() string {
	return sn.NodeType
}

// Start returns the start position
func (sn *SimpleNode) Start() ast.Position {
	return sn.StartPos
}

// End returns the end position
func (sn *SimpleNode) End() ast.Position {
	return sn.EndPos
}

// Children returns child nodes
func (sn *SimpleNode) Children() []ast.Node {
	return sn.ChildNodes
}

// Text returns the source text
func (sn *SimpleNode) Text() string {
	return sn.TextContent
}

// Parent returns the parent node
func (sn *SimpleNode) Parent() ast.Node {
	return sn.ParentNode
}

// SetParent sets the parent node
func (sn *SimpleNode) SetParent(parent ast.Node) {
	sn.ParentNode = parent
}

// String returns a string representation
func (sn *SimpleNode) String() string {
	return fmt.Sprintf("SimpleNode{Type: %s, Children: %d}", sn.NodeType, len(sn.ChildNodes))
}

// FastParser implements a hybrid parsing strategy that avoids tree-sitter for simple cases
type FastParser struct {
	simplePatterns *SimplePatternMatcher
	fallbackCount  int64
	fastParseCount int64
	// Note: Using parser pool instead of shared instance for thread safety
}

// SimplePatternMatcher handles basic Perl constructs without full parsing
type SimplePatternMatcher struct {
	varDecl    *regexp.Regexp
	subDecl    *regexp.Regexp
	typeAnnot  *regexp.Regexp
	simpleStmt *regexp.Regexp
	useStmt    *regexp.Regexp
}

// NewFastParser creates a new fast parser
func NewFastParser() *FastParser {
	return &FastParser{
		simplePatterns: newSimplePatternMatcher(),
	}
}

// newSimplePatternMatcher creates pattern matchers for common constructs
func newSimplePatternMatcher() *SimplePatternMatcher {
	return &SimplePatternMatcher{
		// Match variable declarations: my $var = value; or my Type $var = value;
		varDecl: regexp.MustCompile(`^\s*my\s+(?:(\w+)\s+)?(\$\w+)\s*=\s*([^;]+);\s*$`),

		// Match subroutine declarations: sub name { } or sub name(Type $param) -> RetType { }
		subDecl: regexp.MustCompile(`^\s*sub\s+(\w+)(?:\([^)]*\))?(?:\s*->\s*\w+)?\s*\{.*\}\s*$`),

		// Match type annotations: Type or Type|OtherType
		typeAnnot: regexp.MustCompile(`^[A-Z]\w*(?:\|[A-Z]\w*)*$`),

		// Match simple statements: print, return, etc.
		simpleStmt: regexp.MustCompile(`^\s*(?:print|say|return|last|next|die)\b`),

		// Match use statements: use Module;
		useStmt: regexp.MustCompile(`^\s*use\s+[\w:]+(?:\s+qw\([^)]*\))?\s*;\s*$`),
	}
}

// ParseString attempts fast parsing first, falls back to tree-sitter if needed
func (fp *FastParser) ParseString(content string) (*ast.AST, error) {
	// Quick heuristic: if content is simple enough, try fast parsing
	if fp.isSimpleContent(content) {
		if result, ok := fp.tryFastParse(content); ok {
			fp.fastParseCount++
			return result, nil
		}
	}

	// Fall back to tree-sitter for complex content using parser pool
	fp.fallbackCount++
	return parser.PooledParserFunc(func(p parser.Parser) (*parser.AST, error) {
		return p.ParseString(content)
	})
}

// isSimpleContent determines if content might be parseable with fast parser
func (fp *FastParser) isSimpleContent(content string) bool {
	lines := strings.Split(content, "\n")

	// Don't attempt fast parsing for files with too many lines
	if len(lines) > 50 {
		return false
	}

	// Check if all lines match simple patterns
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check if line matches any simple pattern
		if !fp.simplePatterns.matchesSimplePattern(line) {
			return false
		}
	}

	return true
}

// matchesSimplePattern checks if a line matches any of the simple patterns
func (spm *SimplePatternMatcher) matchesSimplePattern(line string) bool {
	return spm.varDecl.MatchString(line) ||
		spm.subDecl.MatchString(line) ||
		spm.simpleStmt.MatchString(line) ||
		spm.useStmt.MatchString(line) ||
		strings.HasPrefix(line, "}") ||
		strings.HasPrefix(line, "{")
}

// tryFastParse attempts to parse simple content without tree-sitter
func (fp *FastParser) tryFastParse(content string) (*ast.AST, bool) {
	lines := strings.Split(content, "\n")

	// Create a simple AST structure using the AST struct
	root := &SimpleNode{
		NodeType:   "root",
		ChildNodes: make([]ast.Node, 0, len(lines)),
	}

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		node := fp.parseSimpleLine(line, lineNum+1)
		if node != nil {
			root.ChildNodes = append(root.ChildNodes, node)
		}
	}

	result := &ast.AST{
		Root:   root,
		Source: content,
		Path:   "", // Will be set by caller if needed
	}

	return result, true
}

// parseSimpleLine parses a single line into an AST node
func (fp *FastParser) parseSimpleLine(line string, lineNum int) ast.Node {
	// Try variable declaration
	if matches := fp.simplePatterns.varDecl.FindStringSubmatch(line); matches != nil {
		return &SimpleNode{
			NodeType: "variable_declaration",
			Attributes: map[string]string{
				"type":  matches[1], // Could be empty for untyped
				"name":  matches[2],
				"value": matches[3],
				"line":  fmt.Sprintf("%d", lineNum),
			},
		}
	}

	// Try subroutine declaration
	if matches := fp.simplePatterns.subDecl.FindStringSubmatch(line); matches != nil {
		return &SimpleNode{
			NodeType: "subroutine_declaration",
			Attributes: map[string]string{
				"name": matches[1],
				"line": fmt.Sprintf("%d", lineNum),
			},
		}
	}

	// Try simple statement
	if fp.simplePatterns.simpleStmt.MatchString(line) {
		return &SimpleNode{
			NodeType: "statement",
			Attributes: map[string]string{
				"content": line,
				"line":    fmt.Sprintf("%d", lineNum),
			},
		}
	}

	// Try use statement
	if fp.simplePatterns.useStmt.MatchString(line) {
		return &SimpleNode{
			NodeType: "use_statement",
			Attributes: map[string]string{
				"content": line,
				"line":    fmt.Sprintf("%d", lineNum),
			},
		}
	}

	// Default: create a generic statement node
	return &SimpleNode{
		NodeType: "statement",
		Attributes: map[string]string{
			"content": line,
			"line":    fmt.Sprintf("%d", lineNum),
		},
	}
}

// GetStats returns parsing statistics
func (fp *FastParser) GetStats() (fastCount, fallbackCount int64, fastPercentage float64) {
	total := fp.fastParseCount + fp.fallbackCount
	if total == 0 {
		return 0, 0, 0
	}

	percentage := float64(fp.fastParseCount) / float64(total) * 100
	return fp.fastParseCount, fp.fallbackCount, percentage
}

// MemoryOptimizedBinder reduces memory allocation during symbol binding
type MemoryOptimizedBinder struct {
	objectPool  *ObjectPool
	stringCache map[string]string // Intern strings to reduce memory
}

// NewMemoryOptimizedBinder creates a memory-optimized binder
func NewMemoryOptimizedBinder() *MemoryOptimizedBinder {
	return &MemoryOptimizedBinder{
		objectPool:  NewObjectPool(),
		stringCache: make(map[string]string, 1000),
	}
}

// AlgorithmicOptimizer provides algorithmic improvements for common operations
type AlgorithmicOptimizer struct {
	// Type lookup cache for faster type resolution
	typeLookupCache map[string]*CachedTypeInfo

	// Symbol lookup acceleration
	symbolIndex map[string][]*CachedSymbolInfo
}

// CachedTypeInfo represents cached type information
type CachedTypeInfo struct {
	TypeName   string
	IsUnion    bool
	Components []string
	LastAccess int64
}

// CachedSymbolInfo represents cached symbol information
type CachedSymbolInfo struct {
	Name       string
	Type       string
	Scope      string
	LineNumber int
}

// NewAlgorithmicOptimizer creates a new algorithmic optimizer
func NewAlgorithmicOptimizer() *AlgorithmicOptimizer {
	return &AlgorithmicOptimizer{
		typeLookupCache: make(map[string]*CachedTypeInfo),
		symbolIndex:     make(map[string][]*CachedSymbolInfo),
	}
}

// OptimizedTypeResolution provides faster type resolution using caching
func (ao *AlgorithmicOptimizer) OptimizedTypeResolution(typeName string) (*CachedTypeInfo, bool) {
	if info, exists := ao.typeLookupCache[typeName]; exists {
		info.LastAccess++
		return info, true
	}

	return nil, false
}

// BuildSymbolIndex creates an optimized index for symbol lookups
func (ao *AlgorithmicOptimizer) BuildSymbolIndex(symbols map[string]interface{}) {
	// Clear existing index
	ao.symbolIndex = make(map[string][]*CachedSymbolInfo)

	// Build new index with optimized data structures
	for name, symbol := range symbols {
		// Extract relevant information (this would be adapted to actual symbol structure)
		_ = symbol // Avoid unused variable warning
		info := &CachedSymbolInfo{
			Name: name,
			// Type, Scope, LineNumber would be extracted from symbol
		}

		// Add to index with multiple keys for faster lookup
		ao.symbolIndex[name] = append(ao.symbolIndex[name], info)

		// Also index by prefix for prefix matching
		if len(name) > 1 {
			prefix := name[:len(name)/2]
			ao.symbolIndex[prefix] = append(ao.symbolIndex[prefix], info)
		}
	}
}

// StreamingParser reduces memory usage for large files by processing incrementally
type StreamingParser struct {
	bufferSize    int
	buffer        []byte
	parseCallback func(chunk ast.Node) error
}

// NewStreamingParser creates a parser that processes large files in chunks
func NewStreamingParser(bufferSize int, callback func(chunk ast.Node) error) *StreamingParser {
	return &StreamingParser{
		bufferSize:    bufferSize,
		buffer:        make([]byte, 0, bufferSize),
		parseCallback: callback,
	}
}

// ParseLargeFile processes a large file incrementally to reduce memory usage
func (sp *StreamingParser) ParseLargeFile(content []byte) error {
	// Split content into logical chunks (e.g., by subroutines or logical blocks)
	chunks := sp.splitIntoChunks(content)

	for _, chunk := range chunks {
		// Parse each chunk separately
		node := sp.parseChunk(chunk)
		if node != nil {
			if err := sp.parseCallback(node); err != nil {
				return err
			}
		}
	}

	return nil
}

// splitIntoChunks divides content into logical parsing units
func (sp *StreamingParser) splitIntoChunks(content []byte) [][]byte {
	var chunks [][]byte
	var currentChunk []byte

	lines := bytes.Split(content, []byte("\n"))

	for _, line := range lines {
		currentChunk = append(currentChunk, line...)
		currentChunk = append(currentChunk, '\n')

		// Split on subroutine boundaries or when buffer is full
		if bytes.HasPrefix(line, []byte("sub ")) || len(currentChunk) > sp.bufferSize {
			if len(currentChunk) > 0 {
				chunks = append(chunks, currentChunk)
				currentChunk = nil
			}
		}
	}

	// Add final chunk
	if len(currentChunk) > 0 {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}

// parseChunk parses a single chunk of content
func (sp *StreamingParser) parseChunk(chunk []byte) ast.Node {
	// Simplified parsing for demonstration
	// In practice, this would use the optimized parser
	return &SimpleNode{
		NodeType: "chunk",
		Attributes: map[string]string{
			"size": fmt.Sprintf("%d", len(chunk)),
		},
	}
}
