// ABOUTME: Benchmarks demonstrating tree-sitter shim architecture performance improvements
// ABOUTME: Verifies single-parse efficiency and type annotation preservation

package compiler

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parser"
)

// BenchmarkTreeSitterShimCompilation benchmarks the complete pipeline with tree-sitter shim
func BenchmarkTreeSitterShimCompilation(b *testing.B) {
	// Complex typed Perl code to test
	testCode := `
use v5.38;

class Person {
    field Str $name;
    field Int $age;
    field ArrayRef[Str] $emails;

    method new(Str $name, Int $age) : Person {
        $self->{name} = $name;
        $self->{age} = $age;
        $self->{emails} = [];
        return $self;
    }

    method add_email(Str $email) : Void {
        push @{$self->{emails}}, $email;
    }

    method get_info() : HashRef[Any] {
        return {
            name => $self->{name},
            age => $self->{age},
            email_count => scalar(@{$self->{emails}})
        };
    }
}

sub calculate_average(ArrayRef[Num] $numbers) : Num {
    my Num $sum = 0;
    for my Num $n (@$numbers) {
        $sum += $n;
    }
    return $sum / @$numbers;
}

my Person $john = Person->new("John Doe", 30);
$john->add_email("john@example.com");
$john->add_email("johndoe@work.com");

my ArrayRef[Num] $scores = [95.5, 87.3, 92.1, 88.7];
my Num $avg = calculate_average($scores);
`

	b.Run("WithTreeSitterShim", func(b *testing.B) {
		shimParser, err := parser.NewShimParser()
		if err != nil {
			b.Fatalf("Failed to create shim parser: %v", err)
		}
		compiler := NewInferredTypedPerlCompiler()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Parse once with tree-sitter shim
			shimAST, err := shimParser.ParseStringShim(testCode)
			if err != nil {
				b.Fatalf("Parsing failed: %v", err)
			}

			// Compile directly using CST from shim AST (no re-parsing!)
			output, err := compiler.Compile(shimAST)
			if err != nil {
				b.Fatalf("Compilation failed: %v", err)
			}

			// Verify type annotations are preserved
			if !strings.Contains(output, "field Str $name") {
				b.Error("Type annotations not preserved")
			}
		}
	})

	b.Run("WithTraditionalAST", func(b *testing.B) {
		traditionalParser, err := parser.NewParser()
		if err != nil {
			b.Fatalf("Failed to create parser: %v", err)
		}
		compiler := NewInferredTypedPerlCompiler()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Parse with traditional parser
			ast, err := traditionalParser.ParseString(testCode)
			if err != nil {
				b.Fatalf("Parsing failed: %v", err)
			}

			// Compile (will re-parse internally since no CST)
			output, err := compiler.Compile(ast)
			if err != nil {
				b.Fatalf("Compilation failed: %v", err)
			}

			// Verify type annotations are preserved
			if !strings.Contains(output, "field Str $name") {
				b.Error("Type annotations not preserved")
			}
		}
	})
}

// BenchmarkPipelineCompilerWithShim benchmarks pipeline compiler with tree-sitter shim
func BenchmarkPipelineCompilerWithShim(b *testing.B) {
	testCode := `
my Int $count = 42;
my Str $name = "test";
my ArrayRef[Int] $numbers = [1, 2, 3, 4, 5];

sub add(Int $a, Int $b) : Int {
    return $a + $b;
}

for my Int $i (0..10) {
    $count = add($count, $i);
}
`

	b.Run("TypedPerlPipeline_WithShim", func(b *testing.B) {
		shimParser, err := parser.NewShimParser()
		if err != nil {
			b.Fatalf("Failed to create shim parser: %v", err)
		}
		compiler := NewTypedPerlPipelineCompiler()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Parse once
			shimAST, err := shimParser.ParseStringShim(testCode)
			if err != nil {
				b.Fatalf("Parsing failed: %v", err)
			}

			// Compile with direct CST access
			_, err = compiler.Compile(shimAST)
			if err != nil {
				b.Fatalf("Compilation failed: %v", err)
			}
		}
	})

	b.Run("CleanPerlPipeline_WithShim", func(b *testing.B) {
		shimParser, err := parser.NewShimParser()
		if err != nil {
			b.Fatalf("Failed to create shim parser: %v", err)
		}
		compiler := NewCleanPerlPipelineCompiler()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Parse once
			shimAST, err := shimParser.ParseStringShim(testCode)
			if err != nil {
				b.Fatalf("Parsing failed: %v", err)
			}

			// Compile with direct CST access
			_, err = compiler.Compile(shimAST)
			if err != nil {
				b.Fatalf("Compilation failed: %v", err)
			}
		}
	})
}

// BenchmarkTreeSitterShimMemoryUsage compares memory usage between approaches
func BenchmarkTreeSitterShimMemoryUsage(b *testing.B) {
	// Large file simulation - repeat code to make it bigger
	baseCode := `
my Int $var = 42;
sub process(Str $input) : Str { return uc($input); }
`
	var builder strings.Builder
	for i := 0; i < 100; i++ {
		builder.WriteString(baseCode)
	}
	largeCode := builder.String()

	b.Run("MemoryWithShim", func(b *testing.B) {
		shimParser, err := parser.NewShimParser()
		if err != nil {
			b.Fatalf("Failed to create shim parser: %v", err)
		}
		compiler := NewInferredTypedPerlCompiler()

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			shimAST, _ := shimParser.ParseStringShim(largeCode)
			compiler.Compile(shimAST)
		}
	})

	b.Run("MemoryWithoutShim", func(b *testing.B) {
		traditionalParser, err := parser.NewParser()
		if err != nil {
			b.Fatalf("Failed to create parser: %v", err)
		}
		compiler := NewInferredTypedPerlCompiler()

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ast, _ := traditionalParser.ParseString(largeCode)
			compiler.Compile(ast)
		}
	})
}
