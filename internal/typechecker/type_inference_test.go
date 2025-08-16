// ABOUTME: Enhanced type inference capability tests
// ABOUTME: Tests sophisticated type inference from untyped Perl code patterns

package typechecker

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/typedef"
)

// TestBuiltinFunctionInference tests type inference from Perl built-in functions
func TestBuiltinFunctionInference(t *testing.T) {
	// Re-enabled: Flow analysis should now work with TypeChecker symbol information
	testCases := []struct {
		name          string
		code          string
		expectedTypes map[string]string
		description   string
	}{
		{
			name: "ref_function_inference",
			code: `
sub check_types($data) {
    my $type = ref($data);
    my $is_hash = ref($data) eq 'HASH';
    my $is_array = ref($data) eq 'ARRAY';
    return ($type, $is_hash, $is_array);
}`,
			expectedTypes: map[string]string{
				"type":     "Str",
				"is_hash":  "Bool",
				"is_array": "Bool",
			},
			description: "Should infer string type from ref() and boolean from comparisons",
		},
		{
			name: "defined_function_inference",
			code: `
sub check_defined($value) {
    my $is_def = defined($value);
    my $count_def = defined(@array);
    my $hash_def = defined(%hash);
    return ($is_def, $count_def, $hash_def);
}`,
			expectedTypes: map[string]string{
				"is_def":    "Bool",
				"count_def": "Bool",
				"hash_def":  "Bool",
			},
			description: "Should infer boolean type from defined() calls",
		},
		{
			name: "keys_values_inference",
			code: `
sub hash_operations(%input) {
    my @keys = keys(%input);
    my @values = values(%input);
    my $count = keys(%input);
    return (\@keys, \@values, $count);
}`,
			expectedTypes: map[string]string{
				"keys":   "Array[Str]",
				"values": "ArrayRef[Any]",
				"count":  "Int",
			},
			description: "Should infer array types from keys/values and scalar context",
		},
		{
			name: "numeric_functions_inference",
			code: `
sub math_operations($x, $y) {
    my $sum = $x + $y;
    my $length = length($x);
    my $int_val = int($x);
    my $abs_val = abs($y);
    return ($sum, $length, $int_val, $abs_val);
}`,
			expectedTypes: map[string]string{
				"sum":     "Num",
				"length":  "Int",
				"int_val": "Int",
				"abs_val": "Num",
			},
			description: "Should infer numeric types from mathematical operations",
		},
		{
			name: "string_functions_inference",
			code: `
sub string_operations($text) {
    my $upper = uc($text);
    my $lower = lc($text);
    my $chomp_result = chomp($text);
    my @split_result = split(/,/, $text);
    return ($upper, $lower, $chomp_result, \@split_result);
}`,
			expectedTypes: map[string]string{
				"upper":        "Str",
				"lower":        "Str",
				"chomp_result": "Int",
				"split_result": "ArrayRef[Str]",
			},
			description: "Should infer string types from string manipulation functions",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			analyzer := setupTypeInferenceTest(t)
			cfg := buildInferenceCFG(t, analyzer, tc.code)

			_ = analyzer.analyzeDataFlow(cfg)

			for varName, expectedType := range tc.expectedTypes {
				found := verifyTypeInferred(cfg, varName, expectedType)
				if !found {
					t.Errorf("%s: Expected to infer type '%s' for variable $%s",
						tc.description, expectedType, varName)
				}
			}
		})
	}
}

// TestLibraryFunctionInference tests inference from common library functions
func TestLibraryFunctionInference(t *testing.T) {
	// Re-enabled: Flow analysis should now work with TypeChecker symbol information
	testCases := []struct {
		name          string
		code          string
		expectedTypes map[string]string
		description   string
	}{
		{
			name: "json_functions_inference",
			code: `
sub json_operations($json_str, $data) {
    my $decoded = decode_json($json_str);
    my $encoded = encode_json($data);
    my $json_obj = JSON->new();
    my $pretty = $json_obj->pretty->encode($data);
    return ($decoded, $encoded, $json_obj, $pretty);
}`,
			expectedTypes: map[string]string{
				"decoded":  "HashRef[Str, Any]",
				"encoded":  "Str",
				"json_obj": "JSON",
				"pretty":   "Str",
			},
			description: "Should infer types from JSON manipulation functions",
		},
		{
			name: "database_functions_inference",
			code: `
sub database_operations($dbh, $sql) {
    my $row = $dbh->selectrow_hashref($sql);
    my @rows = $dbh->selectall_array($sql);
    my $arrayref = $dbh->selectall_arrayref($sql);
    my $count = $dbh->do($sql);
    return ($row, \@rows, $arrayref, $count);
}`,
			expectedTypes: map[string]string{
				"row":      "Maybe[HashRef[Str, Str]]",
				"rows":     "ArrayRef[ArrayRef[Str]]",
				"arrayref": "ArrayRef[ArrayRef[Str]]",
				"count":    "Maybe[Int]",
			},
			description: "Should infer types from database operations",
		},
		{
			name: "file_operations_inference",
			code: `
sub file_operations($filename) {
    my $content = slurp($filename);
    my $fh = open_file($filename);
    my $lines = read_lines($filename);
    my $size = file_size($filename);
    return ($content, $fh, $lines, $size);
}`,
			expectedTypes: map[string]string{
				"content": "Str",
				"fh":      "Maybe[FileHandle]",
				"lines":   "ArrayRef[Str]",
				"size":    "Maybe[Int]",
			},
			description: "Should infer types from file operation functions",
		},
		{
			name: "utility_functions_inference",
			code: `
sub utility_operations($data, $pattern) {
    my $dumped = Dumper($data);
    my $cloned = clone($data);
    my $matched = $data =~ /$pattern/;
    my @captures = $data =~ /($pattern)/g;
    return ($dumped, $cloned, $matched, \@captures);
}`,
			expectedTypes: map[string]string{
				"dumped":   "Str",
				"cloned":   "Any",
				"matched":  "Bool",
				"captures": "ArrayRef[Str]",
			},
			description: "Should infer types from utility functions",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			analyzer := setupTypeInferenceTest(t)
			cfg := buildInferenceCFG(t, analyzer, tc.code)

			_ = analyzer.analyzeDataFlow(cfg)

			for varName, expectedType := range tc.expectedTypes {
				found := verifyTypeInferred(cfg, varName, expectedType)
				if !found {
					t.Errorf("%s: Expected to infer type '%s' for variable $%s",
						tc.description, expectedType, varName)
				}
			}
		})
	}
}

// TestConstructorCallInference tests type inference from constructor patterns
func TestConstructorCallInference(t *testing.T) {
	// Re-enabled: Flow analysis should now work with TypeChecker symbol information
	testCases := []struct {
		name          string
		code          string
		expectedTypes map[string]string
		description   string
	}{
		{
			name: "class_constructor_inference",
			code: `
sub create_objects() {
    my $user = User->new(name => 'John');
    my $db = Database->connect($dsn);
    my $logger = Logger->get_logger('app');
    my $config = Config::Simple->new();
    return ($user, $db, $logger, $config);
}`,
			expectedTypes: map[string]string{
				"user":   "User",
				"db":     "Database",
				"logger": "Logger",
				"config": "Config::Simple",
			},
			description: "Should infer class types from constructor calls",
		},
		{
			name: "builtin_constructor_inference",
			code: `
sub create_builtins() {
    my $regex = qr/pattern/;
    my $arrayref = [];
    my $hashref = {};
    my $coderef = sub { return 42; };
    return ($regex, $arrayref, $hashref, $coderef);
}`,
			expectedTypes: map[string]string{
				"regex":    "Regexp",
				"arrayref": "ArrayRef",
				"hashref":  "HashRef",
				"coderef":  "CodeRef",
			},
			description: "Should infer types from built-in constructors",
		},
		{
			name: "factory_method_inference",
			code: `
sub factory_patterns() {
    my $parser = XML::Parser->create_parser();
    my $writer = File::Writer->for_file($filename);
    my $handler = Event::Handler->register_handler($event);
    return ($parser, $writer, $handler);
}`,
			expectedTypes: map[string]string{
				"parser":  "XML::Parser",
				"writer":  "File::Writer",
				"handler": "Event::Handler",
			},
			description: "Should infer types from factory method patterns",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			analyzer := setupTypeInferenceTest(t)
			cfg := buildInferenceCFG(t, analyzer, tc.code)

			_ = analyzer.analyzeDataFlow(cfg)

			for varName, expectedType := range tc.expectedTypes {
				found := verifyTypeInferred(cfg, varName, expectedType)
				if !found {
					t.Errorf("%s: Expected to infer type '%s' for variable $%s",
						tc.description, expectedType, varName)
				}
			}
		})
	}
}

// TestMethodChainInference tests type inference through method chains
func TestMethodChainInference(t *testing.T) {
	// Re-enabled: Flow analysis should now work with TypeChecker symbol information
	testCases := []struct {
		name          string
		code          string
		expectedTypes map[string]string
		description   string
	}{
		{
			name: "fluent_interface_chains",
			code: `
sub fluent_operations($data) {
    my $result = JSON->new()->pretty()->canonical()->encode($data);
    my $query = SQLBuilder->select('*')->from('users')->where('active = 1')->build();
    my $response = HTTP::Client->new()->get($url)->decode_content();
    return ($result, $query, $response);
}`,
			expectedTypes: map[string]string{
				"result":   "Str",
				"query":    "Str",
				"response": "Str",
			},
			description: "Should infer final types through fluent interface chains",
		},
		{
			name: "data_transformation_chains",
			code: `
sub transformation_chains(@data) {
    my @processed = map { uc($_) } grep { defined($_) } @data;
    my $joined = join(', ', @processed);
    my @split_again = split(/, /, $joined);
    return (\@processed, $joined, \@split_again);
}`,
			expectedTypes: map[string]string{
				"processed":   "ArrayRef[Str]",
				"joined":      "Str",
				"split_again": "ArrayRef[Str]",
			},
			description: "Should infer types through data transformation chains",
		},
		{
			name: "object_method_chains",
			code: `
sub object_chains($user) {
    my $name = $user->get_profile()->get_name()->to_string();
    my $age = $user->get_profile()->get_age()->as_number();
    my $active = $user->is_active()->as_boolean();
    return ($name, $age, $active);
}`,
			expectedTypes: map[string]string{
				"name":   "Str",
				"age":    "Num",
				"active": "Bool",
			},
			description: "Should infer types through object method chains",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			analyzer := setupTypeInferenceTest(t)
			cfg := buildInferenceCFG(t, analyzer, tc.code)

			_ = analyzer.analyzeDataFlow(cfg)

			for varName, expectedType := range tc.expectedTypes {
				found := verifyTypeInferred(cfg, varName, expectedType)
				if !found {
					t.Errorf("%s: Expected to infer type '%s' for variable $%s",
						tc.description, expectedType, varName)
				}
			}
		})
	}
}

// TestBinaryExpressionInference tests type inference from binary expressions
func TestBinaryExpressionInference(t *testing.T) {
	// Re-enabled: Flow analysis should now work with TypeChecker symbol information
	testCases := []struct {
		name          string
		code          string
		expectedTypes map[string]string
		description   string
	}{
		{
			name: "arithmetic_expressions",
			code: `
sub arithmetic_ops($x, $y) {
    my $sum = $x + $y;
    my $product = $x * $y;
    my $quotient = $x / $y;
    my $remainder = $x % $y;
    my $power = $x ** $y;
    return ($sum, $product, $quotient, $remainder, $power);
}`,
			expectedTypes: map[string]string{
				"sum":       "Num",
				"product":   "Num",
				"quotient":  "Num",
				"remainder": "Int",
				"power":     "Num",
			},
			description: "Should infer numeric types from arithmetic operations",
		},
		{
			name: "string_expressions",
			code: `
sub string_ops($a, $b) {
    my $concat = $a . $b;
    my $repeat = $a x 3;
    my $match = $a =~ /$b/;
    my $substitute = $a =~ s/$b/replacement/r;
    return ($concat, $repeat, $match, $substitute);
}`,
			expectedTypes: map[string]string{
				"concat":     "Str",
				"repeat":     "Str",
				"match":      "Bool",
				"substitute": "Str",
			},
			description: "Should infer string types from string operations",
		},
		{
			name: "comparison_expressions",
			code: `
sub comparisons($x, $y, $a, $b) {
    my $num_eq = $x == $y;
    my $num_lt = $x < $y;
    my $str_eq = $a eq $b;
    my $str_lt = $a lt $b;
    my $defined_or = $x // $y;
    my $logical_and = $a && $b;
    return ($num_eq, $num_lt, $str_eq, $str_lt, $defined_or, $logical_and);
}`,
			expectedTypes: map[string]string{
				"num_eq":      "Bool",
				"num_lt":      "Bool",
				"str_eq":      "Bool",
				"str_lt":      "Bool",
				"defined_or":  "Any", // Could be either operand type
				"logical_and": "Any", // Could be either operand type
			},
			description: "Should infer boolean types from comparison operations",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			analyzer := setupTypeInferenceTest(t)
			cfg := buildInferenceCFG(t, analyzer, tc.code)

			_ = analyzer.analyzeDataFlow(cfg)

			for varName, expectedType := range tc.expectedTypes {
				found := verifyTypeInferred(cfg, varName, expectedType)
				if !found {
					t.Errorf("%s: Expected to infer type '%s' for variable $%s",
						tc.description, expectedType, varName)
				}
			}
		})
	}
}

// TestConditionalExpressionInference tests type inference from conditional expressions
func TestConditionalExpressionInference(t *testing.T) {
	// Re-enabled: Flow analysis should now work with TypeChecker symbol information
	testCases := []struct {
		name          string
		code          string
		expectedTypes map[string]string
		description   string
	}{
		{
			name: "ternary_expressions",
			code: `
sub ternary_ops($condition, $x, $y) {
    my $result = $condition ? $x : $y;
    my $number = $condition ? 42 : 0;
    my $string = $condition ? "yes" : "no";
    my $mixed = $condition ? $x : "default";
    return ($result, $number, $string, $mixed);
}`,
			expectedTypes: map[string]string{
				"result": "Any", // Union of $x and $y types
				"number": "Int", // Both branches are Int
				"string": "Str", // Both branches are Str
				"mixed":  "Any", // Mixed types
			},
			description: "Should infer types from ternary conditional expressions",
		},
		{
			name: "logical_expressions",
			code: `
sub logical_ops($a, $b, $c) {
    my $and_result = $a && $b;
    my $or_result = $a || $b;
    my $not_result = !$a;
    my $chain = $a && $b && $c;
    return ($and_result, $or_result, $not_result, $chain);
}`,
			expectedTypes: map[string]string{
				"and_result": "Any",  // Could be $a (if false) or $b
				"or_result":  "Any",  // Could be $a (if true) or $b
				"not_result": "Bool", // Logical negation always boolean
				"chain":      "Any",  // Chain of logical operations
			},
			description: "Should infer types from logical expressions",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			analyzer := setupTypeInferenceTest(t)
			cfg := buildInferenceCFG(t, analyzer, tc.code)

			_ = analyzer.analyzeDataFlow(cfg)

			for varName, expectedType := range tc.expectedTypes {
				found := verifyTypeInferred(cfg, varName, expectedType)
				if !found {
					t.Errorf("%s: Expected to infer type '%s' for variable $%s",
						tc.description, expectedType, varName)
				}
			}
		})
	}
}

// Helper functions for type inference tests

func setupTypeInferenceTest(t *testing.T) *FlowAnalyzer {
	typeStore, err := typedef.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create type store: %v", err)
	}

	hierarchy := typedef.NewTypeHierarchy(typeStore)
	symbolTable := binder.NewSymbolTable()
	symbolTable.Package = "inference_test"

	tc := NewTypeChecker(hierarchy, symbolTable, "inference_test")
	tc.SafetyAnalysisEnabled = true

	analyzer := NewFlowAnalyzer(tc)
	return analyzer
}

func buildInferenceCFG(t *testing.T, analyzer *FlowAnalyzer, code string) *ControlFlowGraph {
	astResult := parseInferenceCode(t, code)
	cfg, err := analyzer.buildControlFlowGraph(astResult)
	if err != nil {
		t.Fatalf("Failed to build CFG: %v", err)
	}
	return cfg
}

func parseInferenceCode(t *testing.T, code string) *ast.AST {
	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	result, err := p.ParseString(code)
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	return result
}

func verifyTypeInferred(cfg *ControlFlowGraph, varName, expectedType string) bool {
	for _, block := range cfg.Nodes {
		if block.TypeState != nil && block.TypeState.VariableTypes != nil {
			if actualType, exists := block.TypeState.VariableTypes[varName]; exists {
				if strings.Contains(actualType, expectedType) || actualType == expectedType {
					return true
				}
			}
		}
	}
	return false
}
