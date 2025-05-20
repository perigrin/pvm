// ABOUTME: Integration with PSC type checking
// ABOUTME: Connects the parser with the PSC component

package parser

import (
	"fmt"
	"os"
	"strings"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/typedef"
)

// TypeCheckResult represents the result of type checking a file
type TypeCheckResult struct {
	// Path is the path to the checked file
	Path string

	// Errors is a list of type errors found during checking
	Errors []TypeCheckError

	// TypeAnnotations is a list of type annotations found in the code
	TypeAnnotations []*TypeAnnotation
}

// TypeCheckError represents a type checking error
type TypeCheckError struct {
	// Message is the error message
	Message string

	// Line is the line number where the error occurred
	Line int

	// Column is the column number where the error occurred
	Column int

	// Path is the path to the file where the error occurred
	Path string
}

// Error implements the error interface
func (e TypeCheckError) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s", e.Path, e.Line, e.Column, e.Message)
}

// TypeChecker provides type checking functionality
type TypeChecker struct {
	// Parser is the parser used for parsing Perl code
	Parser Parser

	// TypeStore is the store for type definitions
	TypeStore *typedef.Storage
}

// NewTypeChecker creates a new TypeChecker
func NewTypeChecker() (*TypeChecker, error) {
	// Create a parser
	parser, err := NewParser()
	if err != nil {
		return nil, err
	}

	// Create a type store
	typeStore, err := typedef.NewStorage()
	if err != nil {
		return nil, err
	}

	return &TypeChecker{
		Parser:    parser,
		TypeStore: typeStore,
	}, nil
}

// CheckFile performs type checking on a Perl file
func (tc *TypeChecker) CheckFile(path string) (*TypeCheckResult, error) {
	// Parse the file using our enhanced parser
	// The TreeSitterParser now provides better support for type annotations
	ast, err := tc.Parser.ParseFile(path)
	if err != nil {
		return nil, err
	}

	// Check for parser errors
	if len(ast.Errors) > 0 {
		result := &TypeCheckResult{
			Path:            path,
			Errors:          []TypeCheckError{},
			TypeAnnotations: ast.TypeAnnotations,
		}

		// Convert parser errors to type check errors
		for _, parseErr := range ast.Errors {
			var typErr TypeCheckError

			if perr, ok := parseErr.(*ParseError); ok {
				typErr = TypeCheckError{
					Message: perr.Message,
					Line:    perr.Line,
					Column:  perr.Column,
					Path:    path,
				}
			} else {
				typErr = TypeCheckError{
					Message: parseErr.Error(),
					Line:    0,
					Column:  0,
					Path:    path,
				}
			}

			result.Errors = append(result.Errors, typErr)
		}

		return result, nil
	}

	// With our enhanced parser, we should have more accurate type annotations
	// from advanced parsing capabilities like field declarations, method parameters,
	// type declarations, etc.

	// Type check annotations
	return tc.checkTypeAnnotations(ast)
}

// checkTypeAnnotations checks type annotations found in the AST
func (tc *TypeChecker) checkTypeAnnotations(ast *AST) (*TypeCheckResult, error) {
	result := &TypeCheckResult{
		Path:            ast.Path,
		Errors:          []TypeCheckError{},
		TypeAnnotations: ast.TypeAnnotations,
	}

	// Build module import map
	imports := make(map[string]bool)

	// This would involve scanning the AST for use statements
	// For the simplified implementation, we'll assume no imports

	// Check each type annotation
	for _, annotation := range ast.TypeAnnotations {
		// Verify the type exists
		err := tc.verifyType(annotation.TypeExpression, imports)
		if err != nil {
			result.Errors = append(result.Errors, TypeCheckError{
				Message: err.Error(),
				Line:    annotation.Pos.Line,
				Column:  annotation.Pos.Column,
				Path:    ast.Path,
			})
		}
	}

	return result, nil
}

// verifyType checks if a type exists and is valid
func (tc *TypeChecker) verifyType(typeExpr *TypeExpression, imports map[string]bool) error {
	// Handle negation types
	if typeExpr.Negation {
		// A negation type is valid if the underlying type is valid
		return tc.verifyType(typeExpr, imports)
	}

	// Check if it's a built-in type
	if isBuiltinType(typeExpr.Name) {
		return nil
	}

	// For parameterized types, check the parameters
	for _, param := range typeExpr.Params {
		if err := tc.verifyType(param, imports); err != nil {
			return err
		}
	}

	// For union and intersection types, the type itself is valid if the components are valid
	if typeExpr.Union || typeExpr.Intersection {
		// For explicit union and intersection types where we have the components,
		// we don't need to verify the name itself
		return nil
	}

	// Check if it's an imported type
	parts := strings.Split(typeExpr.Name, "::")
	if len(parts) > 1 {
		moduleName := strings.Join(parts[:len(parts)-1], "::")
		if !imports[moduleName] {
			return fmt.Errorf("type %s comes from module %s which is not imported", typeExpr.Name, moduleName)
		}

		// Check if the module has a type definition
		_, err := tc.TypeStore.Load(moduleName)
		if err != nil {
			return fmt.Errorf("no type definition found for module %s", moduleName)
		}

		// In a real implementation, we would check if the specific type exists in the module
		return nil
	}

	// Check if it's a type defined in the current module
	// This is a simplified check - in a real implementation, we would look up
	// the type in a registry of types defined in the current module
	// For now, we assume it's valid if it starts with an uppercase letter
	// (a common convention for type names)
	if len(typeExpr.Name) > 0 && typeExpr.Name[0] >= 'A' && typeExpr.Name[0] <= 'Z' {
		return nil
	}

	// If we can't verify the type, log a warning but don't fail
	// In a real implementation, we would have a more thorough check
	return fmt.Errorf("unrecognized type %s", typeExpr.Name)
}

// isBuiltinType returns true if the type is a built-in type
func isBuiltinType(typeName string) bool {
	builtinTypes := map[string]bool{
		// Basic types
		"Any":    true,
		"Scalar": true,
		"Str":    true,
		"Num":    true,
		"Int":    true,
		"Float":  true,
		"Double": true,
		"Bool":   true,
		"Undef":  true,

		// Reference types
		"Ref":        true,
		"ScalarRef":  true,
		"ArrayRef":   true,
		"HashRef":    true,
		"CodeRef":    true,
		"RegexpRef":  true,
		"GlobRef":    true,
		"FileHandle": true,

		// Container types
		"List":  true,
		"Array": true,
		"Hash":  true,
		"Code":  true,
		"Glob":  true,

		// Type modifiers
		"Maybe":    true,
		"Optional": true,

		// Role/trait types
		"Callable":    true,
		"Iterable":    true,
		"Positional":  true,
		"Associative": true,

		// IO and system types
		"IO":   true,
		"Path": true,
		"File": true,
		"Dir":  true,

		// Additional scalar types
		"ClassName":  true,
		"RoleName":   true,
		"MethodName": true,
		"Byte":       true,
		"Char":       true,
		"VarName":    true,
	}

	return builtinTypes[typeName]
}

// StripAnnotations removes type annotations from Perl code
func StripAnnotations(path string) (string, error) {
	// Create a parser
	p, err := NewParser()
	if err != nil {
		return "", err
	}

	// Parse the file with our enhanced parser
	ast, err := p.ParseFile(path)
	if err != nil {
		return "", err
	}

	// Read the file content
	content, err := os.ReadFile(path)
	if err != nil {
		return "", errors.NewSystemError("001",
			"Failed to read file", err).
			WithLocation(path)
	}

	// Convert content to string
	originalCode := string(content)

	// Now that we have a more advanced parser, we can implement a more sophisticated
	// annotation stripping process. The approach is to rewrite the code line by line,
	// removing the type annotations where they are found.

	// We need to be careful to avoid changing line numbers, so we'll replace annotations
	// with spaces rather than removing them entirely.

	if len(ast.TypeAnnotations) == 0 {
		// No annotations to strip
		return originalCode, nil
	}

	// Split the code into lines for easier processing
	lines := strings.Split(originalCode, "\n")

	// Map annotations by line for efficient lookup
	annotationsByLine := make(map[int][]*TypeAnnotation)
	for _, ann := range ast.TypeAnnotations {
		lineNum := ann.Pos.Line
		annotationsByLine[lineNum] = append(annotationsByLine[lineNum], ann)
	}

	// Process each line that has annotations
	for lineNum, annotations := range annotationsByLine {
		if lineNum <= 0 || lineNum > len(lines) {
			continue
		}

		line := lines[lineNum-1]

		// Sort annotations in reverse order of column position
		// This ensures we process them from right to left
		// to avoid affecting the positions of subsequent annotations
		for i := 0; i < len(annotations); i++ {
			for j := i + 1; j < len(annotations); j++ {
				if annotations[i].Pos.Column < annotations[j].Pos.Column {
					annotations[i], annotations[j] = annotations[j], annotations[i]
				}
			}
		}

		// Process each annotation on this line
		for _, ann := range annotations {
			switch ann.Kind {
			case VarAnnotation:
				// Handle variable annotations - e.g., "my Type $var" -> "my $var"
				if typePos := strings.Index(line, ann.TypeExpression.String()); typePos >= 0 {
					typeLen := len(ann.TypeExpression.String())
					// Replace the type with spaces to maintain line structure
					spaces := strings.Repeat(" ", typeLen)
					line = line[:typePos] + spaces + line[typePos+typeLen:]
				}

			case SubParamAnnotation, MethodParamAnnotation:
				// Handle parameter annotations - e.g., "Type $param" -> "$param"
				if ann.AnnotatedItem != "" {
					paramPos := strings.Index(line, ann.AnnotatedItem)
					if paramPos > 0 {
						// Look for the type that appears before the parameter
						beforeParam := line[:paramPos]
						if typePos := strings.LastIndex(beforeParam, ann.TypeExpression.String()); typePos >= 0 {
							typeLen := len(ann.TypeExpression.String())
							// Replace the type with spaces to maintain line structure
							spaces := strings.Repeat(" ", typeLen)
							line = line[:typePos] + spaces + line[typePos+typeLen:]
						}
					}
				}

			case SubReturnAnnotation, MethodReturnAnnotation:
				// Handle return type annotations - e.g., "-> Type" -> "-> "
				arrowPos := strings.Index(line, "->")
				if arrowPos >= 0 {
					returnTypePos := arrowPos + 2 // Skip the "->"
					// Find the type after the arrow
					afterArrow := strings.TrimSpace(line[returnTypePos:])
					if strings.HasPrefix(afterArrow, ann.TypeExpression.String()) {
						typeLen := len(ann.TypeExpression.String())
						// Calculate the actual position of the type
						actualTypePos := returnTypePos + strings.Index(line[returnTypePos:], ann.TypeExpression.String())
						// Replace the type with spaces to maintain line structure
						spaces := strings.Repeat(" ", typeLen)
						line = line[:actualTypePos] + spaces + line[actualTypePos+typeLen:]
					}
				}

			case AttrAnnotation:
				// Handle field/attribute annotations - e.g., "field Type $attr" -> "field $attr"
				if typePos := strings.Index(line, ann.TypeExpression.String()); typePos >= 0 {
					typeLen := len(ann.TypeExpression.String())
					// Replace the type with spaces to maintain line structure
					spaces := strings.Repeat(" ", typeLen)
					line = line[:typePos] + spaces + line[typePos+typeLen:]
				}

			case TypeDeclAnnotation:
				// For type declarations, we might want to keep them as-is or
				// replace the entire line with a comment
				// For now, we'll keep them as they don't affect runtime behavior
			}
		}

		// Update the line in the array
		lines[lineNum-1] = line
	}

	// Combine the lines back into a single string
	strippedCode := strings.Join(lines, "\n")

	return strippedCode, nil
}
