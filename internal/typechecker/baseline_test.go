// ABOUTME: Baseline testing for type checker component to prevent regressions
// ABOUTME: Tests type checker output against known good baselines for type inference and error detection

package typechecker

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/parser"
	basetesting "tamarou.com/pvm/internal/testing"
	"tamarou.com/pvm/internal/typedef"
)

func TestTypeChecker_Baselines(t *testing.T) {
	basetesting.SampleTypeCheckerTest(t)
	// Create processor function that type checks input and returns results
	processor := func(input []byte) ([]byte, error) {
		// Parse the input
		p, err := parser.NewParser()
		if err != nil {
			return nil, err
		}

		ast, err := p.ParseString(string(input))
		if err != nil {
			return nil, err
		}

		// Bind symbols
		b := binder.NewBinder()
		symbolTable, err := b.BindAST(ast)
		if err != nil {
			return nil, err
		}

		// Type check
		store, _ := typedef.NewStorage()
		hierarchy := typedef.NewTypeHierarchy(store)
		checker := NewTypeChecker(hierarchy, symbolTable, "test_module")
		errors := checker.CheckAST(ast)

		// Format results as baseline output
		var result strings.Builder

		// Include symbol information
		result.WriteString("=== SYMBOLS ===\n")
		symbols := symbolTable.GetVisibleSymbols()
		for _, symbol := range symbols {
			result.WriteString(symbol.String() + "\n")
		}

		// Include type checking results
		result.WriteString("=== TYPE ERRORS ===\n")
		if len(errors) == 0 {
			result.WriteString("No type errors\n")
		} else {
			for _, err := range errors {
				result.WriteString(err.Error() + "\n")
			}
		}

		return []byte(result.String()), nil
	}

	// Create test suite
	suite := basetesting.NewBaselineTestSuite("typechecker", "../../testdata/typechecker", processor)

	// Run all baseline tests
	suite.RunAllTests(t)
}

func TestTypeChecker_SpecificBaselines(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{
			name: "simple_types",
			input: `my Int $count = 42;
my Str $name = "Alice";
my Bool $active = true;`,
		},
		{
			name: "type_errors",
			input: `my Int $count = "not a number";
my Str $name = 42;
my Bool $flag = "maybe";`,
		},
		{
			name: "union_types",
			input: `my Int|Str $value = 42;
$value = "hello";
my Bool|Undef $maybe = undef;`,
		},
		{
			name: "intersection_types",
			input: `my Object&Serializable $data = create_object();
my ArrayRef&Iterable $list = [];`,
		},
		{
			name: "complex_inference",
			input: `my $count = 42;  # Should infer Int
my $name = "Alice";  # Should infer Str
my $list = [1, 2, 3];  # Should infer ArrayRef[Int]
my $hash = {a => 1, b => 2};  # Should infer HashRef[Int]`,
		},
		{
			name: "function_types",
			input: `sub add(Int $a, Int $b) -> Int {
    return $a + $b;
}

sub greet(Str $name) -> Str {
    return "Hello, $name!";
}

my $result = add(1, 2);  # Should be Int
my $message = greet("Alice");  # Should be Str`,
		},
		{
			name: "method_chaining",
			input: `my $object = MyClass->new();
my $result = $object->method1()->method2()->value();`,
		},
		{
			name: "conditional_types",
			input: `my Int|Undef $maybe_count;
if (defined $maybe_count) {
    my $safe_count = $maybe_count;  # Should narrow to Int
}`,
		},
	}

	processor := func(input []byte) ([]byte, error) {
		// Parse the input
		p, err := parser.NewParser()
		if err != nil {
			return nil, err
		}

		ast, err := p.ParseString(string(input))
		if err != nil {
			return nil, err
		}

		// Bind symbols
		b := binder.NewBinder()
		symbolTable, err := b.BindAST(ast)
		if err != nil {
			return nil, err
		}

		// Type check
		store, _ := typedef.NewStorage()
		hierarchy := typedef.NewTypeHierarchy(store)
		checker := NewTypeChecker(hierarchy, symbolTable, "test_module")
		errors := checker.CheckAST(ast)

		// Format results
		var result strings.Builder
		result.WriteString("=== SYMBOLS ===\n")
		symbols := symbolTable.GetVisibleSymbols()
		for _, symbol := range symbols {
			result.WriteString(symbol.String() + "\n")
		}

		result.WriteString("=== TYPE ERRORS ===\n")
		if len(errors) == 0 {
			result.WriteString("No type errors\n")
		} else {
			for _, err := range errors {
				result.WriteString(err.Error() + "\n")
			}
		}

		return []byte(result.String()), nil
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			basetesting.BaselineTestFunc(t, tc.name, processor, []byte(tc.input))
		})
	}
}

func BenchmarkTypeChecker_Performance(b *testing.B) {
	monitor := basetesting.NewPerformanceMonitor("testdata/performance/typechecker")
	helper := basetesting.NewBenchmarkHelper(monitor)

	// Simple type checking
	helper.BenchmarkTypeChecker(b, "simple_types", func() error {
		p, err := parser.NewParser()
		if err != nil {
			return err
		}

		script := `my Int $x = 42; my Str $y = "hello";`
		ast, err := p.ParseString(script)
		if err != nil {
			return err
		}

		binder := binder.NewBinder()
		symbolTable, err := binder.BindAST(ast)
		if err != nil {
			return err
		}

		store, _ := typedef.NewStorage()
		hierarchy := typedef.NewTypeHierarchy(store)
		checker := NewTypeChecker(hierarchy, symbolTable, "test_module")
		_ = checker.CheckAST(ast)
		return nil
	})

	// Complex type checking
	helper.BenchmarkTypeChecker(b, "complex_types", func() error {
		p, err := parser.NewParser()
		if err != nil {
			return err
		}

		script := `
my ArrayRef[HashRef[Int|Str]] $data = [
    {name => "Alice", age => 30},
    {name => "Bob", age => 25},
];

sub process(ArrayRef[HashRef[Any]] $input) -> HashRef[Int] {
    my %result;
    for my $item (@$input) {
        $result{$item->{name}} = $item->{age};
    }
    return \%result;
}

my $processed = process($data);`

		ast, err := p.ParseString(script)
		if err != nil {
			return err
		}

		binder := binder.NewBinder()
		symbolTable, err := binder.BindAST(ast)
		if err != nil {
			return err
		}

		store, _ := typedef.NewStorage()
		hierarchy := typedef.NewTypeHierarchy(store)
		checker := NewTypeChecker(hierarchy, symbolTable, "test_module")
		_ = checker.CheckAST(ast)
		return nil
	})

	// Type inference benchmarking
	helper.BenchmarkTypeChecker(b, "inference", func() error {
		p, err := parser.NewParser()
		if err != nil {
			return err
		}

		script := `
my $count = 0;
my $list = [];
my $hash = {};

for my $i (1..100) {
    $count += $i;
    push @$list, $i;
    $hash->{$i} = $i * 2;
}

my $result = $count + @$list + keys %$hash;`

		ast, err := p.ParseString(script)
		if err != nil {
			return err
		}

		binder := binder.NewBinder()
		symbolTable, err := binder.BindAST(ast)
		if err != nil {
			return err
		}

		store, _ := typedef.NewStorage()
		hierarchy := typedef.NewTypeHierarchy(store)
		checker := NewTypeChecker(hierarchy, symbolTable, "test_module")
		_ = checker.CheckAST(ast)
		return nil
	})

	// Run benchmark suite
	benchmarks := map[string]func(*testing.B){
		"simple_types": func(b *testing.B) {
			helper.BenchmarkTypeChecker(b, "simple_types", func() error {
				// Benchmark implementation here
				return nil
			})
		},
		"complex_types": func(b *testing.B) {
			helper.BenchmarkTypeChecker(b, "complex_types", func() error {
				// Benchmark implementation here
				return nil
			})
		},
		"inference": func(b *testing.B) {
			helper.BenchmarkTypeChecker(b, "inference", func() error {
				// Benchmark implementation here
				return nil
			})
		},
	}

	t := &testing.T{} // Create a testing.T for the monitor
	monitor.RunBenchmarkSuite(t, "typechecker", benchmarks)
}
