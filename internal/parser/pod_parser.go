// ABOUTME: POD documentation parser for extracting type hints
// ABOUTME: Analyzes Perl POD for method signatures and type information

package parser

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// PODParser extracts type information from POD documentation
type PODParser struct {
	// TypePatterns contains patterns for recognizing type hints in POD
	TypePatterns []*TypePattern

	// SignaturePatterns contains patterns for method signatures
	SignaturePatterns []*SignaturePattern

	// Current parsing context
	context *PODContext
}

// TypePattern represents a pattern for extracting type information
type TypePattern struct {
	// Name is the pattern name
	Name string

	// Pattern is the regex pattern
	Pattern *regexp.Regexp

	// ExtractType extracts type information from matches
	ExtractType func(matches []string) *TypeHint
}

// SignaturePattern represents a pattern for extracting method signatures
type SignaturePattern struct {
	// Name is the pattern name
	Name string

	// Pattern is the regex pattern
	Pattern *regexp.Regexp

	// ExtractSignature extracts signature information from matches
	ExtractSignature func(matches []string) *PODMethodSignature
}

// TypeHint represents a type hint extracted from POD
type TypeHint struct {
	// Variable or parameter name
	Name string

	// Type specification
	Type string

	// IsOptional indicates if the parameter is optional
	IsOptional bool

	// DefaultValue if specified
	DefaultValue string

	// Description from documentation
	Description string
}

// PODMethodSignature represents a method signature from POD
type PODMethodSignature struct {
	// MethodName is the method name
	MethodName string

	// Parameters is the list of parameters
	Parameters []*PODParameter

	// ReturnType is the documented return type
	ReturnType string

	// Description is the method description
	Description string

	// Examples contains usage examples
	Examples []string
}

// PODParameter represents a parameter documented in POD
type PODParameter struct {
	// Name is the parameter name
	Name string

	// Type is the parameter type
	Type string

	// IsOptional indicates if optional
	IsOptional bool

	// DefaultValue if any
	DefaultValue string

	// Description of the parameter
	Description string
}

// PODContext tracks the current parsing context
type PODContext struct {
	// CurrentSection is the current POD section
	CurrentSection string

	// CurrentMethod is the method being documented
	CurrentMethod string

	// CurrentItem is the current item in a list
	CurrentItem string

	// InCodeBlock indicates if we're in a code block
	InCodeBlock bool

	// Depth tracks nesting depth
	Depth int
}

// PODDocument represents a parsed POD document
type PODDocument struct {
	// Methods contains all documented methods
	Methods map[string]*PODMethodSignature

	// Types contains type definitions found in POD
	Types map[string]*TypeDefinition

	// Attributes contains documented attributes
	Attributes map[string]*PODAttribute

	// Synopsis contains code examples from SYNOPSIS
	Synopsis []string

	// Description contains the module description
	Description string
}

// TypeDefinition represents a type defined in POD
type TypeDefinition struct {
	// Name is the type name
	Name string

	// Definition is the type definition
	Definition string

	// Examples of usage
	Examples []string
}

// PODAttribute represents an attribute documented in POD
type PODAttribute struct {
	// Name is the attribute name
	Name string

	// Type is the attribute type
	Type string

	// Description of the attribute
	Description string

	// IsRequired indicates if required
	IsRequired bool

	// DefaultValue if any
	DefaultValue string
}

// NewPODParser creates a new POD parser
func NewPODParser() *PODParser {
	parser := &PODParser{
		TypePatterns:      []*TypePattern{},
		SignaturePatterns: []*SignaturePattern{},
		context: &PODContext{
			CurrentSection: "",
			CurrentMethod:  "",
		},
	}

	// Initialize patterns
	parser.initializePatterns()

	return parser
}

// initializePatterns sets up recognition patterns
func (p *PODParser) initializePatterns() {
	// Common type hint patterns in POD

	// Pattern: C<$param> - I<Type>
	p.TypePatterns = append(p.TypePatterns, &TypePattern{
		Name:    "code_type",
		Pattern: regexp.MustCompile(`C<(\$\w+)>\s*-\s*I<([^>]+)>`),
		ExtractType: func(matches []string) *TypeHint {
			if len(matches) >= 3 {
				return &TypeHint{
					Name: matches[1],
					Type: matches[2],
				}
			}
			return nil
		},
	})

	// Pattern: $param (Type) - description
	p.TypePatterns = append(p.TypePatterns, &TypePattern{
		Name:    "param_parens",
		Pattern: regexp.MustCompile(`(\$\w+)\s*\(([^)]+)\)\s*-\s*(.+)`),
		ExtractType: func(matches []string) *TypeHint {
			if len(matches) >= 4 {
				return &TypeHint{
					Name:        matches[1],
					Type:        matches[2],
					Description: matches[3],
				}
			}
			return nil
		},
	})

	// Pattern: Type $param - description
	p.TypePatterns = append(p.TypePatterns, &TypePattern{
		Name:    "type_param",
		Pattern: regexp.MustCompile(`([A-Z]\w+(?:::\w+)*)\s+(\$\w+)\s*-\s*(.+)`),
		ExtractType: func(matches []string) *TypeHint {
			if len(matches) >= 4 {
				return &TypeHint{
					Name:        matches[2],
					Type:        matches[1],
					Description: matches[3],
				}
			}
			return nil
		},
	})

	// Method signature patterns

	// Pattern: method_name($param1, $param2) -> ReturnType
	p.SignaturePatterns = append(p.SignaturePatterns, &SignaturePattern{
		Name:    "arrow_return",
		Pattern: regexp.MustCompile(`(\w+)\s*\(([^)]*)\)\s*->\s*([A-Z]\w+(?:::\w+)*)`),
		ExtractSignature: func(matches []string) *PODMethodSignature {
			if len(matches) >= 4 {
				sig := &PODMethodSignature{
					MethodName: matches[1],
					ReturnType: matches[3],
					Parameters: []*PODParameter{},
				}

				// Parse parameters
				if matches[2] != "" {
					params := strings.Split(matches[2], ",")
					for _, param := range params {
						param = strings.TrimSpace(param)
						if param != "" {
							sig.Parameters = append(sig.Parameters, &PODParameter{
								Name: param,
								Type: "Any", // Default type
							})
						}
					}
				}

				return sig
			}
			return nil
		},
	})

	// Pattern: =method method_name
	p.SignaturePatterns = append(p.SignaturePatterns, &SignaturePattern{
		Name:    "pod_weaver_method",
		Pattern: regexp.MustCompile(`^=method\s+(\w+)`),
		ExtractSignature: func(matches []string) *PODMethodSignature {
			if len(matches) >= 2 {
				return &PODMethodSignature{
					MethodName: matches[1],
					Parameters: []*PODParameter{},
				}
			}
			return nil
		},
	})
}

// ParsePOD parses POD documentation from a reader
func (p *PODParser) ParsePOD(reader io.Reader) (*PODDocument, error) {
	doc := &PODDocument{
		Methods:    make(map[string]*PODMethodSignature),
		Types:      make(map[string]*TypeDefinition),
		Attributes: make(map[string]*PODAttribute),
		Synopsis:   []string{},
	}

	scanner := bufio.NewScanner(reader)
	var currentBlock []string
	var inPOD bool

	for scanner.Scan() {
		line := scanner.Text()

		// Check for POD markers
		if strings.HasPrefix(line, "=pod") {
			inPOD = true
			continue
		}
		if strings.HasPrefix(line, "=cut") {
			inPOD = false
			// Process any accumulated block
			if len(currentBlock) > 0 {
				p.processBlock(currentBlock, doc)
				currentBlock = nil
			}
			continue
		}

		// Skip if not in POD
		if !inPOD && !strings.HasPrefix(line, "=") {
			continue
		}

		// Handle POD commands
		if strings.HasPrefix(line, "=") {
			// Process previous block if any
			if len(currentBlock) > 0 {
				p.processBlock(currentBlock, doc)
				currentBlock = nil
			}

			// Update context based on POD command
			p.updateContext(line)
			currentBlock = append(currentBlock, line)
		} else if inPOD || p.context.CurrentSection != "" {
			// Accumulate content
			currentBlock = append(currentBlock, line)
		}
	}

	// Process final block
	if len(currentBlock) > 0 {
		p.processBlock(currentBlock, doc)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return doc, nil
}

// updateContext updates parsing context based on POD commands
func (p *PODParser) updateContext(line string) {
	if strings.HasPrefix(line, "=head1") {
		p.context.CurrentSection = strings.TrimSpace(strings.TrimPrefix(line, "=head1"))
		p.context.CurrentMethod = ""
	} else if strings.HasPrefix(line, "=head2") {
		subsection := strings.TrimSpace(strings.TrimPrefix(line, "=head2"))
		// Check if this is a method documentation
		if p.context.CurrentSection == "METHODS" {
			p.context.CurrentMethod = subsection
		}
	} else if strings.HasPrefix(line, "=item") {
		p.context.CurrentItem = strings.TrimSpace(strings.TrimPrefix(line, "=item"))
	} else if strings.HasPrefix(line, "=method") {
		p.context.CurrentMethod = strings.TrimSpace(strings.TrimPrefix(line, "=method"))
		p.context.CurrentSection = "METHODS"
	}
}

// processBlock processes a block of POD content
func (p *PODParser) processBlock(block []string, doc *PODDocument) {
	if len(block) == 0 {
		return
	}

	switch p.context.CurrentSection {
	case "SYNOPSIS":
		p.processSynopsis(block, doc)
	case "METHODS":
		p.processMethod(block, doc)
	case "ATTRIBUTES":
		p.processAttributes(block, doc)
	case "DESCRIPTION":
		p.processDescription(block, doc)
	case "TYPES":
		p.processTypes(block, doc)
	}
}

// processSynopsis extracts code examples from SYNOPSIS
func (p *PODParser) processSynopsis(block []string, doc *PODDocument) {
	var codeLines []string
	inCode := false

	for _, line := range block {
		if strings.TrimSpace(line) == "" {
			if inCode && len(codeLines) > 0 {
				doc.Synopsis = append(doc.Synopsis, strings.Join(codeLines, "\n"))
				codeLines = nil
				inCode = false
			}
			continue
		}

		// Code blocks in SYNOPSIS usually start with whitespace
		if strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "\t") {
			inCode = true
			codeLines = append(codeLines, strings.TrimSpace(line))
		}
	}

	// Add final code block
	if len(codeLines) > 0 {
		doc.Synopsis = append(doc.Synopsis, strings.Join(codeLines, "\n"))
	}
}

// processMethod extracts method documentation
func (p *PODParser) processMethod(block []string, doc *PODDocument) {
	if p.context.CurrentMethod == "" {
		return
	}

	sig := &PODMethodSignature{
		MethodName:  p.context.CurrentMethod,
		Parameters:  []*PODParameter{},
		Description: "",
		Examples:    []string{},
	}

	var description []string
	var examples []string
	inExample := false

	for _, line := range block {
		// Skip the header line
		if strings.Contains(line, "=head2") || strings.Contains(line, "=method") {
			continue
		}

		// Check for method signature patterns
		for _, pattern := range p.SignaturePatterns {
			if matches := pattern.Pattern.FindStringSubmatch(line); matches != nil {
				if extracted := pattern.ExtractSignature(matches); extracted != nil {
					// Merge extracted info with current signature
					if extracted.ReturnType != "" {
						sig.ReturnType = extracted.ReturnType
					}
					if len(extracted.Parameters) > 0 {
						sig.Parameters = extracted.Parameters
					}
				}
			}
		}

		// Check for parameter type hints
		for _, pattern := range p.TypePatterns {
			if matches := pattern.Pattern.FindStringSubmatch(line); matches != nil {
				if hint := pattern.ExtractType(matches); hint != nil {
					// Add or update parameter info
					p.updateParameterFromHint(sig, hint)
				}
			}
		}

		// Detect example blocks
		if strings.Contains(strings.ToLower(line), "example:") {
			inExample = true
			continue
		}

		// Collect content
		if inExample {
			if strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "\t") {
				examples = append(examples, strings.TrimSpace(line))
			} else if strings.TrimSpace(line) == "" && len(examples) > 0 {
				sig.Examples = append(sig.Examples, strings.Join(examples, "\n"))
				examples = nil
				inExample = false
			}
		} else {
			description = append(description, line)
		}
	}

	// Add final example if any
	if len(examples) > 0 {
		sig.Examples = append(sig.Examples, strings.Join(examples, "\n"))
	}

	sig.Description = strings.TrimSpace(strings.Join(description, " "))
	doc.Methods[sig.MethodName] = sig
}

// updateParameterFromHint updates parameter info from a type hint
func (p *PODParser) updateParameterFromHint(sig *PODMethodSignature, hint *TypeHint) {
	// Find existing parameter or create new one
	var param *PODParameter
	for _, p := range sig.Parameters {
		if p.Name == hint.Name {
			param = p
			break
		}
	}

	if param == nil {
		param = &PODParameter{
			Name: hint.Name,
		}
		sig.Parameters = append(sig.Parameters, param)
	}

	// Update parameter info
	if hint.Type != "" {
		param.Type = hint.Type
	}
	if hint.Description != "" {
		param.Description = hint.Description
	}
	param.IsOptional = hint.IsOptional
	param.DefaultValue = hint.DefaultValue
}

// processAttributes extracts attribute documentation
func (p *PODParser) processAttributes(block []string, doc *PODDocument) {
	var currentAttr *PODAttribute
	var description []string

	for _, line := range block {
		// Check for attribute patterns
		if strings.HasPrefix(line, "=item") {
			// Save previous attribute
			if currentAttr != nil && currentAttr.Name != "" {
				currentAttr.Description = strings.TrimSpace(strings.Join(description, " "))
				doc.Attributes[currentAttr.Name] = currentAttr
			}

			// Start new attribute
			attrName := strings.TrimSpace(strings.TrimPrefix(line, "=item"))
			currentAttr = &PODAttribute{
				Name: attrName,
				Type: "Any", // Default
			}
			description = nil
			continue
		}

		// Look for type information in description
		if currentAttr != nil {
			for _, pattern := range p.TypePatterns {
				if matches := pattern.Pattern.FindStringSubmatch(line); matches != nil {
					if hint := pattern.ExtractType(matches); hint != nil && hint.Type != "" {
						currentAttr.Type = hint.Type
					}
				}
			}

			// Check for required/optional indicators
			if strings.Contains(strings.ToLower(line), "required") {
				currentAttr.IsRequired = true
			}
			if strings.Contains(strings.ToLower(line), "optional") {
				currentAttr.IsRequired = false
			}

			// Check for default values
			if idx := strings.Index(line, "default:"); idx >= 0 {
				defaultStr := strings.TrimSpace(line[idx+8:])
				currentAttr.DefaultValue = defaultStr
			}

			description = append(description, line)
		}
	}

	// Save final attribute
	if currentAttr != nil && currentAttr.Name != "" {
		currentAttr.Description = strings.TrimSpace(strings.Join(description, " "))
		doc.Attributes[currentAttr.Name] = currentAttr
	}
}

// processDescription extracts module description
func (p *PODParser) processDescription(block []string, doc *PODDocument) {
	var description []string
	for _, line := range block {
		if !strings.HasPrefix(line, "=") {
			description = append(description, line)
		}
	}
	doc.Description = strings.TrimSpace(strings.Join(description, " "))
}

// processTypes extracts type definitions
func (p *PODParser) processTypes(block []string, doc *PODDocument) {
	var currentType *TypeDefinition
	var content []string

	for _, line := range block {
		// Check for type definition patterns
		if strings.HasPrefix(line, "=item") {
			// Save previous type
			if currentType != nil && currentType.Name != "" {
				currentType.Definition = strings.TrimSpace(strings.Join(content, " "))
				doc.Types[currentType.Name] = currentType
			}

			// Start new type
			typeName := strings.TrimSpace(strings.TrimPrefix(line, "=item"))
			currentType = &TypeDefinition{
				Name:     typeName,
				Examples: []string{},
			}
			content = nil
			continue
		}

		if currentType != nil {
			content = append(content, line)
		}
	}

	// Save final type
	if currentType != nil && currentType.Name != "" {
		currentType.Definition = strings.TrimSpace(strings.Join(content, " "))
		doc.Types[currentType.Name] = currentType
	}
}

// ExtractTypeHints extracts all type hints from POD content
func (p *PODParser) ExtractTypeHints(content string) ([]*TypeHint, error) {
	var hints []*TypeHint

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		for _, pattern := range p.TypePatterns {
			if matches := pattern.Pattern.FindAllStringSubmatch(line, -1); matches != nil {
				for _, match := range matches {
					if hint := pattern.ExtractType(match); hint != nil {
						hints = append(hints, hint)
					}
				}
			}
		}
	}

	return hints, nil
}

// MergeWithIntrospection merges POD documentation with introspection results
func MergeWithIntrospection(podDoc *PODDocument, introspection *ModuleIntrospectionResult) {
	// Merge method signatures
	for methodName, podSig := range podDoc.Methods {
		// Find corresponding method in introspection
		for pkgName, pkg := range introspection.Packages {
			if methodSig, exists := pkg.Methods[methodName]; exists {
				// Update with POD documentation
				if podSig.ReturnType != "" {
					methodSig.ReturnType = podSig.ReturnType
				}
				methodSig.Documentation = podSig.Description

				// Update parameters
				for i, podParam := range podSig.Parameters {
					if i < len(methodSig.Parameters) {
						param := methodSig.Parameters[i]
						if podParam.Type != "" && podParam.Type != "Any" {
							param.Type = podParam.Type
						}
						param.Documentation = podParam.Description
						param.IsOptional = podParam.IsOptional
						param.DefaultValue = podParam.DefaultValue
					}
				}

				// Add any missing parameters from POD
				if len(podSig.Parameters) > len(methodSig.Parameters) {
					for i := len(methodSig.Parameters); i < len(podSig.Parameters); i++ {
						podParam := podSig.Parameters[i]
						methodSig.Parameters = append(methodSig.Parameters, &ParameterInfo{
							Name:          podParam.Name,
							Type:          podParam.Type,
							IsOptional:    podParam.IsOptional,
							DefaultValue:  podParam.DefaultValue,
							Documentation: podParam.Description,
						})
					}
				}

				// Update in introspection result
				introspection.Packages[pkgName].Methods[methodName] = methodSig
			}
		}
	}

	// Merge attributes
	for attrName, podAttr := range podDoc.Attributes {
		for _, pkg := range introspection.Packages {
			if attr, exists := pkg.Attributes[attrName]; exists {
				// Update with POD documentation
				if podAttr.Type != "" && podAttr.Type != "Any" {
					attr.Type = podAttr.Type
				}
				attr.Documentation = podAttr.Description
				attr.IsRequired = podAttr.IsRequired
				attr.DefaultValue = podAttr.DefaultValue
			} else {
				// Add attribute from POD if not found in introspection
				pkg.Attributes[attrName] = &AttributeInfo{
					Name:          attrName,
					Type:          podAttr.Type,
					IsRequired:    podAttr.IsRequired,
					DefaultValue:  podAttr.DefaultValue,
					Documentation: podAttr.Description,
				}
			}
		}
	}
}

// ParsePODFromFile parses POD from a file path
func (p *PODParser) ParsePODFromFile(path string) (*PODDocument, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	reader := strings.NewReader(string(content))
	return p.ParsePOD(reader)
}
