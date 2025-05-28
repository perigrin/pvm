// ABOUTME: Comprehensive tests for class and role declaration parsing
// ABOUTME: Validates parsing of typed classes, roles, inheritance, and composition

package parser

import (
	"path/filepath"
	"strings"
	"testing"
	
	"tamarou.com/pvm/internal/ast"
)

func TestClassDeclarationParsing(t *testing.T) {
	// Set up test framework
	testDataDir := filepath.Join("testdata", "typed-perl", "classes-roles")
	framework := NewParserTestFramework(testDataDir)
	framework.Verbose = testing.Verbose()

	// Create parser
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	framework.Parser = parser

	// Test specific fixtures
	testFiles := []string{
		"basic_class_declarations.json",
		"generic_class_declarations.json", 
		"class_inheritance.json",
		"complex_inheritance_constraints.json",
		"access_modifiers_visibility.json",
		"constructor_destructor_methods.json",
	}

	for _, testFile := range testFiles {
		t.Run(testFile, func(t *testing.T) {
			testFixtureFile := filepath.Join(testDataDir, testFile)
			
			testCase, err := framework.LoadTestCase(testFixtureFile)
			if err != nil {
				t.Fatalf("Failed to load test fixture %s: %v", testFile, err)
			}

			astResult, err := parser.ParseString(testCase.Input)
			if testCase.ShouldError {
				if err == nil {
					t.Errorf("Expected parsing to fail, but it succeeded")
				}
				return
			}

			if err != nil {
				t.Fatalf("Failed to parse input: %v", err)
			}

			if astResult == nil {
				t.Fatal("Parser returned nil AST")
			}

			// Run validation based on test name
			switch testCase.Name {
			case "basic_class_declarations":
				validateBasicClassStructure(t, astResult)
			case "generic_class_declarations":
				validateGenericClassStructure(t, astResult)
			case "class_inheritance":
				validateClassInheritance(t, astResult)
			case "complex_inheritance_constraints":
				validateComplexInheritance(t, astResult)
			case "access_modifiers_visibility":
				validateAccessModifiers(t, astResult)
			case "constructor_destructor_methods":
				validateConstructorDestructor(t, astResult)
			default:
				t.Logf("No specific validation for test case: %s", testCase.Name)
			}
		})
	}
}

func TestRoleDeclarationParsing(t *testing.T) {
	// Set up test framework
	testDataDir := filepath.Join("testdata", "typed-perl", "classes-roles")
	framework := NewParserTestFramework(testDataDir)
	framework.Verbose = testing.Verbose()

	// Create parser
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	framework.Parser = parser

	// Test specific fixtures
	testFiles := []string{
		"basic_role_declarations.json",
		"generic_role_declarations.json",
		"role_composition_conflicts.json",
	}

	for _, testFile := range testFiles {
		t.Run(testFile, func(t *testing.T) {
			testFixtureFile := filepath.Join(testDataDir, testFile)
			
			testCase, err := framework.LoadTestCase(testFixtureFile)
			if err != nil {
				t.Fatalf("Failed to load test fixture %s: %v", testFile, err)
			}

			astResult, err := parser.ParseString(testCase.Input)
			if testCase.ShouldError {
				if err == nil {
					t.Errorf("Expected parsing to fail, but it succeeded")
				}
				return
			}

			if err != nil {
				t.Fatalf("Failed to parse input: %v", err)
			}

			if astResult == nil {
				t.Fatal("Parser returned nil AST")
			}

			// Run validation based on test name
			switch testCase.Name {
			case "basic_role_declarations":
				validateBasicRoleStructure(t, astResult)
			case "generic_role_declarations":
				validateGenericRoleStructure(t, astResult)
			case "role_composition_conflicts":
				validateRoleComposition(t, astResult)
			default:
				t.Logf("No specific validation for test case: %s", testCase.Name)
			}
		})
	}
}

func TestComprehensiveClassRoleIntegration(t *testing.T) {
	// Set up test framework
	testDataDir := filepath.Join("testdata", "typed-perl", "classes-roles")
	framework := NewParserTestFramework(testDataDir)
	framework.Verbose = testing.Verbose()

	testFixtureFile := filepath.Join(testDataDir, "all_features_combined.json")
	
	testCase, err := framework.LoadTestCase(testFixtureFile)
	if err != nil {
		t.Fatalf("Failed to load comprehensive test fixture: %v", err)
	}

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	astResult, err := parser.ParseString(testCase.Input)
	if err != nil {
		t.Fatalf("Failed to parse comprehensive input: %v", err)
	}

	if astResult == nil {
		t.Fatal("Parser returned nil AST")
	}

	// Validate all features work together
	validateComprehensiveIntegration(t, astResult)
}

func validateBasicClassStructure(t *testing.T, astResult *ast.AST) {
	// Find class declaration in AST
	classDecl := findClassDeclaration(astResult, "User")
	if classDecl == nil {
		t.Fatal("Could not find User class declaration in AST")
	}

	// Validate class has expected fields
	if len(classDecl.Fields) < 3 {
		t.Errorf("Expected at least 3 fields, got %d", len(classDecl.Fields))
	}

	// Validate class has expected methods
	if len(classDecl.Methods) < 2 {
		t.Errorf("Expected at least 2 methods, got %d", len(classDecl.Methods))
	}

	// Validate field types
	for _, field := range classDecl.Fields {
		if field.TypeExpr == nil {
			t.Errorf("Field %s missing type annotation", field.Name)
		}
	}

	// Validate method signatures
	for _, method := range classDecl.Methods {
		if method.Name == "new" {
			if method.ReturnType == nil {
				t.Error("Constructor missing return type")
			}
		}
	}
}

func validateGenericClassStructure(t *testing.T, astResult *ast.AST) {
	classDecl := findClassDeclaration(astResult, "Container")
	if classDecl == nil {
		t.Fatal("Could not find Container class declaration in AST")
	}

	// Validate generic parameters
	if len(classDecl.TypeParameters) == 0 {
		t.Error("Expected generic class to have type parameters")
	}

	// Validate constraints
	if len(classDecl.Constraints) == 0 {
		t.Error("Expected generic class to have type constraints")
	}

	// Validate generic methods use type parameters
	for _, method := range classDecl.Methods {
		if method.Name == "add" && len(method.Parameters) > 0 {
			param := method.Parameters[0]
			if param.TypeExpr == nil {
				t.Error("Expected generic method parameter to have type")
			}
		}
	}
}

func validateClassInheritance(t *testing.T, astResult *ast.AST) {
	classDecl := findClassDeclaration(astResult, "Document")
	if classDecl == nil {
		t.Fatal("Could not find Document class declaration in AST")
	}

	// Validate superclass
	if classDecl.Superclass == nil {
		t.Error("Expected class to have superclass")
	}

	// Validate roles
	if len(classDecl.Roles) < 2 {
		t.Errorf("Expected at least 2 roles, got %d", len(classDecl.Roles))
	}

	// Validate implemented methods
	hasSerialize := false
	hasDeserialize := false
	for _, method := range classDecl.Methods {
		if method.Name == "serialize" {
			hasSerialize = true
		}
		if method.Name == "deserialize" {
			hasDeserialize = true
		}
	}

	if !hasSerialize {
		t.Error("Expected serialize method implementation")
	}
	if !hasDeserialize {
		t.Error("Expected deserialize method implementation")
	}
}

func validateComplexInheritance(t *testing.T, astResult *ast.AST) {
	classDecl := findClassDeclaration(astResult, "ProcessingQueue")
	if classDecl == nil {
		t.Fatal("Could not find ProcessingQueue class declaration in AST")
	}

	// Validate complex inheritance pattern
	if classDecl.Superclass == nil {
		t.Error("Expected complex class to have superclass")
	}

	// Validate intersection type constraints
	if len(classDecl.Constraints) == 0 {
		t.Error("Expected complex constraints")
	}

	// Validate method with constraints
	for _, method := range classDecl.Methods {
		if method.Name == "enqueue" {
			if len(method.Constraints) == 0 {
				t.Error("Expected method-level constraints")
			}
		}
	}
}

func validateAccessModifiers(t *testing.T, astResult *ast.AST) {
	classDecl := findClassDeclaration(astResult, "BankAccount")
	if classDecl == nil {
		t.Fatal("Could not find BankAccount class declaration in AST")
	}

	// Validate field access modifiers
	hasPrivateField := false
	hasProtectedField := false
	hasPublicField := false

	for _, field := range classDecl.Fields {
		switch field.AccessLevel {
		case "private":
			hasPrivateField = true
		case "protected":
			hasProtectedField = true
		case "public":
			hasPublicField = true
		}
	}

	if !hasPrivateField {
		t.Error("Expected private field")
	}
	if !hasProtectedField {
		t.Error("Expected protected field")
	}
	if !hasPublicField {
		t.Error("Expected public field")
	}

	// Validate method access modifiers
	hasPrivateMethod := false
	hasPublicMethod := false
	hasProtectedMethod := false

	for _, method := range classDecl.Methods {
		switch method.AccessLevel {
		case "private":
			hasPrivateMethod = true
		case "public":
			hasPublicMethod = true
		case "protected":
			hasProtectedMethod = true
		}
	}

	if !hasPrivateMethod {
		t.Error("Expected private method")
	}
	if !hasPublicMethod {
		t.Error("Expected public method")
	}
	if !hasProtectedMethod {
		t.Error("Expected protected method")
	}
}

func validateConstructorDestructor(t *testing.T, astResult *ast.AST) {
	classDecl := findClassDeclaration(astResult, "Resource")
	if classDecl == nil {
		t.Fatal("Could not find Resource class declaration in AST")
	}

	// Validate constructor methods
	hasBUILD := false
	hasNew := false
	hasDESTROY := false

	for _, method := range classDecl.Methods {
		switch method.Name {
		case "BUILD":
			hasBUILD = true
		case "new":
			hasNew = true
		case "DESTROY":
			hasDESTROY = true
		}
	}

	if !hasBUILD {
		t.Error("Expected BUILD method")
	}
	if !hasNew {
		t.Error("Expected new method")
	}
	if !hasDESTROY {
		t.Error("Expected DESTROY method")
	}
}

func validateBasicRoleStructure(t *testing.T, astResult *ast.AST) {
	serializableRole := findRoleDeclaration(astResult, "Serializable")
	cacheableRole := findRoleDeclaration(astResult, "Cacheable")

	if serializableRole == nil {
		t.Fatal("Could not find Serializable role declaration in AST")
	}
	if cacheableRole == nil {
		t.Fatal("Could not find Cacheable role declaration in AST")
	}

	// Validate required methods
	if len(serializableRole.RequiredMethods) < 2 {
		t.Errorf("Expected at least 2 required methods, got %d", len(serializableRole.RequiredMethods))
	}

	// Validate provided methods
	if len(cacheableRole.ProvidedMethods) < 2 {
		t.Errorf("Expected at least 2 provided methods, got %d", len(cacheableRole.ProvidedMethods))
	}

	// Validate role fields
	if len(cacheableRole.Fields) == 0 {
		t.Error("Expected Cacheable role to have fields")
	}
}

func validateGenericRoleStructure(t *testing.T, astResult *ast.AST) {
	processableRole := findRoleDeclaration(astResult, "Processable")
	eventHandlerRole := findRoleDeclaration(astResult, "EventHandler")

	if processableRole == nil {
		t.Fatal("Could not find Processable role declaration in AST")
	}
	if eventHandlerRole == nil {
		t.Fatal("Could not find EventHandler role declaration in AST")
	}

	// Validate generic parameters
	if len(processableRole.TypeParameters) == 0 {
		t.Error("Expected Processable role to have type parameters")
	}
	if len(eventHandlerRole.TypeParameters) == 0 {
		t.Error("Expected EventHandler role to have type parameters")
	}

	// Validate constraints
	if len(processableRole.Constraints) == 0 {
		t.Error("Expected Processable role to have constraints")
	}
	if len(eventHandlerRole.Constraints) == 0 {
		t.Error("Expected EventHandler role to have constraints")
	}
}

func validateRoleComposition(t *testing.T, astResult *ast.AST) {
	widgetClass := findClassDeclaration(astResult, "Widget")
	if widgetClass == nil {
		t.Fatal("Could not find Widget class declaration in AST")
	}

	// Validate multiple role composition
	if len(widgetClass.Roles) < 3 {
		t.Errorf("Expected at least 3 roles, got %d", len(widgetClass.Roles))
	}

	// Validate conflict resolution
	hasBoundsMethod := false
	for _, method := range widgetClass.Methods {
		if method.Name == "get_bounds" {
			hasBoundsMethod = true
			break
		}
	}

	if !hasBoundsMethod {
		t.Error("Expected conflict resolution method get_bounds")
	}
}

func validateComprehensiveIntegration(t *testing.T, astResult *ast.AST) {
	// Validate type aliases
	typeInfo := ast.ExtractTypeInformation(astResult)
	if len(typeInfo.TypeAliases) == 0 {
		t.Error("Expected type aliases in comprehensive example")
	}

	// Validate roles
	if len(typeInfo.Roles) == 0 {
		t.Error("Expected roles in comprehensive example")
	}

	// Validate classes
	if len(typeInfo.Classes) == 0 {
		t.Error("Expected classes in comprehensive example")
	}

	// Validate comprehensive UserRepository class
	userRepoClass := findClassDeclaration(astResult, "UserRepository")
	if userRepoClass == nil {
		t.Fatal("Could not find UserRepository class")
	}

	// Validate it has all features
	if len(userRepoClass.TypeParameters) == 0 {
		t.Error("Expected generic class")
	}
	if userRepoClass.Superclass == nil {
		t.Error("Expected inheritance")
	}
	if len(userRepoClass.Roles) < 3 {
		t.Error("Expected multiple role composition")
	}
	if len(userRepoClass.Constraints) == 0 {
		t.Error("Expected type constraints")
	}
}

// Helper functions to find AST nodes
func findClassDeclaration(astResult *ast.AST, name string) *ast.ClassDecl {
	var result *ast.ClassDecl
	
	walker := func(node ast.Node) bool {
		if classDecl, ok := node.(*ast.ClassDecl); ok {
			if classDecl.Name == name {
				result = classDecl
				return false // Stop walking
			}
		}
		return true // Continue walking
	}
	
	walkAST(astResult.Root, walker)
	return result
}

func findRoleDeclaration(astResult *ast.AST, name string) *ast.RoleDecl {
	var result *ast.RoleDecl
	
	walker := func(node ast.Node) bool {
		if roleDecl, ok := node.(*ast.RoleDecl); ok {
			if roleDecl.Name == name {
				result = roleDecl
				return false // Stop walking
			}
		}
		return true // Continue walking
	}
	
	walkAST(astResult.Root, walker)
	return result
}

func walkAST(node ast.Node, visitor func(ast.Node) bool) {
	if node == nil {
		return
	}
	
	if !visitor(node) {
		return
	}
	
	for _, child := range node.Children() {
		walkAST(child, visitor)
	}
}

// debugPrintASTNodes prints AST node structure for debugging
func debugPrintASTNodes(t *testing.T, node ast.Node, depth int) {
	if node == nil {
		return
	}
	
	indent := strings.Repeat("  ", depth)
	t.Logf("%s%s: %q", indent, node.Type(), node.Text())
	
	for _, child := range node.Children() {
		debugPrintASTNodes(t, child, depth+1)
	}
}