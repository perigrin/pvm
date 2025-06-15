// ABOUTME: Baseline testing for parser component to prevent regressions
// ABOUTME: Tests parser output against known good baselines for various Perl constructs

package parser

import (
	"testing"

	basetesting "tamarou.com/pvm/internal/testing"
)

func TestParser_Baselines(t *testing.T) {
	// Create processor function that parses input and returns AST as string
	processor := func(input []byte) ([]byte, error) {
		parser, err := NewParser()
		if err != nil {
			return nil, err
		}

		ast, err := parser.ParseString(string(input))
		if err != nil {
			return nil, err
		}

		// Convert AST to string representation for baseline comparison
		result := ast.String()
		return []byte(result), nil
	}

	// Create test suite
	suite := basetesting.NewBaselineTestSuite("parser", "../../test/corpus/parser", processor)

	// Run all baseline tests
	suite.RunAllTests(t)
}

func TestParser_SpecificBaselines(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{
			name: "simple_variables",
			input: `my $name = "hello";
my @array = (1, 2, 3);
my %hash = (key => "value");`,
		},
		{
			name: "type_annotations",
			input: `my Int $count = 42;
my Str $message = "hello";
my ArrayRef[Int] $numbers = [1, 2, 3];`,
		},
		{
			name: "subroutines",
			input: `sub greet {
    my ($name) = @_;
    return "Hello, $name!";
}

sub add(Int $a, Int $b) returns Int {
    return $a + $b;
}`,
		},
		{
			name: "control_structures",
			input: `if ($condition) {
    say "true";
} elsif ($other) {
    say "maybe";
} else {
    say "false";
}

for my $i (1..10) {
    say $i;
}

while ($running) {
    do_work();
}`,
		},
		{
			name: "complex_expressions",
			input: `my $result = $a + $b * $c;
my $comparison = $x > $y && $z < $w;
my $hash_access = $data->{key}[0];
my $method_call = $object->method($arg1, $arg2);`,
		},
	}

	processor := func(input []byte) ([]byte, error) {
		parser, err := NewParser()
		if err != nil {
			return nil, err
		}

		ast, err := parser.ParseString(string(input))
		if err != nil {
			return nil, err
		}

		result := ast.String()
		return []byte(result), nil
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			basetesting.BaselineTestFunc(t, tc.name, processor, []byte(tc.input))
		})
	}
}

func BenchmarkParser_Performance(b *testing.B) {
	monitor := basetesting.NewPerformanceMonitor("../../test/corpus/parser/performance/parser")
	helper := basetesting.NewBenchmarkHelper(monitor)

	// Simple script parsing
	helper.BenchmarkParser(b, "simple_script", func() error {
		parser, err := NewParser()
		if err != nil {
			return err
		}

		script := `my $x = 42; say $x;`
		_, err = parser.ParseString(script)
		return err
	})

	// Complex script parsing
	helper.BenchmarkParser(b, "complex_script", func() error {
		parser, err := NewParser()
		if err != nil {
			return err
		}

		script := `
package MyClass;
use strict;
use warnings;

sub new {
    my ($class, %args) = @_;
    return bless \%args, $class;
}

sub process {
    my ($self, $data) = @_;
    my @results;

    for my $item (@$data) {
        if ($item->{type} eq 'important') {
            push @results, $self->transform($item);
        }
    }

    return \@results;
}

1;`
		_, err = parser.ParseString(script)
		return err
	})

	// Type annotation parsing
	helper.BenchmarkParser(b, "typed_script", func() error {
		parser, err := NewParser()
		if err != nil {
			return err
		}

		script := `
my Int $count = 0;
my ArrayRef[Str] $names = ["Alice", "Bob"];
my HashRef[Int] $scores = {alice => 95, bob => 87};

sub calculate(Int $a, Int $b) returns Int {
    return $a + $b;
}

sub process_data(ArrayRef[HashRef[Any]] $data) returns ArrayRef[Int] {
    my @results;
    for my $item (@$data) {
        push @results, $item->{score} // 0;
    }
    return \@results;
}`
		_, err = parser.ParseString(script)
		return err
	})

	// Run benchmark suite
	benchmarks := map[string]func(*testing.B){
		"simple_script": func(b *testing.B) {
			helper.BenchmarkParser(b, "simple_script", func() error {
				parser, err := NewParser()
				if err != nil {
					return err
				}
				script := `my $x = 42; say $x;`
				_, err = parser.ParseString(script)
				return err
			})
		},
		"complex_script": func(b *testing.B) {
			helper.BenchmarkParser(b, "complex_script", func() error {
				parser, err := NewParser()
				if err != nil {
					return err
				}
				script := `package MyClass; sub new { my ($class) = @_; return bless {}, $class; } 1;`
				_, err = parser.ParseString(script)
				return err
			})
		},
	}

	t := &testing.T{} // Create a testing.T for the monitor
	monitor.RunBenchmarkSuite(t, "parser", benchmarks)
}
