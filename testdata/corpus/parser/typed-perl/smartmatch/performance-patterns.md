---
category: typed-perl
subcategory: smartmatch
tags:
    - performance
    - optimization
    - efficiency
    - benchmarks
    - scalability
type_check: true
performance_sensitive: true
---

# Performance Patterns for Type-Constrained Pattern Matching

Tests performance characteristics and optimization patterns to ensure that PSC's
type-constrained pattern matching performs efficiently and provides better
optimization opportunities than untyped Perl pattern matching.

```perl
# Efficient literal matching with small unions
type SmallEnum = 'red' | 'green' | 'blue';
my SmallEnum $color = 'red';

# This should compile to efficient jump table or switch statement
given ($color) {
    when ('red')   { say 'warm' }
    when ('green') { say 'nature' }
    when ('blue')  { say 'sky' }
}

# Comparison: Manual if/elsif chain (baseline performance)
my $manual_color = 'red';
if    ($manual_color eq 'red')   { say 'warm' }
elsif ($manual_color eq 'green') { say 'nature' }
elsif ($manual_color eq 'blue')  { say 'sky' }

# Numeric union optimization
type Priority = 1 | 2 | 3 | 4 | 5;
my Priority $pri = 3;

# Should optimize to numeric comparison rather than string comparison
given ($pri) {
    when (1) { say 'critical' }
    when (2) { say 'high' }
    when (3) { say 'medium' }
    when (4) { say 'low' }
    when (5) { say 'minimal' }
}

# Array membership optimization with type constraints
type HttpCode = '200' | '201' | '404' | '500';
my HttpCode $code = '200';

# Type constraints should enable optimized membership testing
my $is_success = $code ~~ ['200', '201'];  # Only 2 values to check
my $is_error = $code ~~ ['404', '500'];    # Only 2 values to check

# vs untyped version requiring all possibilities
my $untyped_code = '200';
my $untyped_success = $untyped_code ~~ ['200', '201', '202', '203', '204'];

# Range-based optimization for sequential numeric types
type SmallRange = 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10;
my SmallRange $num = 5;

# Should optimize to range check: 1 <= $num <= 10
my $in_range = $num ~~ [1..10];

# Complex union with optimization potential
type FileType = '.txt' | '.md' | '.rst' | '.pod' | '.html';
my FileType $ext = '.txt';

# Should compile to optimized string matching (perfect hash, trie, etc.)
given ($ext) {
    when ('.txt')  { say 'text' }
    when ('.md')   { say 'markdown' }
    when ('.rst')  { say 'restructured' }
    when ('.pod')  { say 'perl_doc' }
    when ('.html') { say 'hypertext' }
}

# Regex optimization with type constraints
type Protocol = 'http' | 'https' | 'ftp' | 'sftp';
my Protocol $proto = 'https';

# Type constraints should enable regex optimization
my $is_secure = $proto ~~ qr/s$/;  # Only need to check 'https' and 'sftp'

# vs untyped version
my $untyped_proto = 'https';
my $untyped_secure = $untyped_proto ~~ qr/s$/;  # Must check full string

# Large union performance test
type LargeEnum = 'opt1' | 'opt2' | 'opt3' | 'opt4' | 'opt5' |
                'opt6' | 'opt7' | 'opt8' | 'opt9' | 'opt10' |
                'opt11' | 'opt12' | 'opt13' | 'opt14' | 'opt15' |
                'opt16' | 'opt17' | 'opt18' | 'opt19' | 'opt20';
my LargeEnum $large = 'opt10';

# Should scale efficiently even with large unions
given ($large) {
    when ('opt1')  { say '1' }
    when ('opt2')  { say '2' }
    when ('opt3')  { say '3' }
    when ('opt4')  { say '4' }
    when ('opt5')  { say '5' }
    when ('opt6')  { say '6' }
    when ('opt7')  { say '7' }
    when ('opt8')  { say '8' }
    when ('opt9')  { say '9' }
    when ('opt10') { say '10' }
    when ('opt11') { say '11' }
    when ('opt12') { say '12' }
    when ('opt13') { say '13' }
    when ('opt14') { say '14' }
    when ('opt15') { say '15' }
    when ('opt16') { say '16' }
    when ('opt17') { say '17' }
    when ('opt18') { say '18' }
    when ('opt19') { say '19' }
    when ('opt20') { say '20' }
}

# Nested structure optimization
type NestedPerf = [Int, Str] | [Str, Int] | [Int, Int] | [Str, Str];
my NestedPerf $nested = [42, 'test'];

# Should optimize nested structure matching
given ($nested) {
    when ([Int, Str]) { say 'int_str' }
    when ([Str, Int]) { say 'str_int' }
    when ([Int, Int]) { say 'int_int' }
    when ([Str, Str]) { say 'str_str' }
}

# Hash structure optimization
type ConfigType = {mode => 'dev'} | {mode => 'prod'} | {mode => 'test'};
my ConfigType $config = {mode => 'dev'};

# Should optimize hash structure matching
given ($config) {
    when ({mode => 'dev'})  { say 'development' }
    when ({mode => 'prod'}) { say 'production' }
    when ({mode => 'test'}) { say 'testing' }
}

# Smartmatch optimization with type information
type OptimizedMatch = 'fast' | 'medium' | 'slow';
my OptimizedMatch $speed = 'fast';

# Type information should enable compile-time optimizations
my @speed_array = ('fast', 'medium', 'slow');
my $array_match = $speed ~~ @speed_array;  # Known to always be true

# String length optimization
type FixedLength = 'ABC' | 'DEF' | 'GHI';  # All length 3
my FixedLength $fixed = 'ABC';

# Should optimize by checking length first
my $length_opt = $fixed ~~ qr/^.{3}$/;  # Always true for this type

# Character class optimization
type VowelEnum = 'a' | 'e' | 'i' | 'o' | 'u';
my VowelEnum $vowel = 'a';

# Should optimize to character class
my $is_vowel = $vowel ~~ qr/[aeiou]/;  # Always true

# Mixed type optimization potential
type MixedOptim = 1 | 'one' | 2 | 'two' | 3 | 'three';
my MixedOptim $mixed = 1;

# Should separate numeric vs string optimizations
given ($mixed) {
    when (1)       { say 'numeric one' }
    when (2)       { say 'numeric two' }
    when (3)       { say 'numeric three' }
    when ('one')   { say 'string one' }
    when ('two')   { say 'string two' }
    when ('three') { say 'string three' }
}

# Compile-time constant folding
type Constant = 42;
my Constant $const = 42;

# Should be optimized away at compile time
my $always_true = $const ~~ 42;      # Always true
my $always_false = $const ~~ 43;     # Always false

# Partial evaluation opportunities
type PartialEnum = 'alpha' | 'beta' | 'gamma';
my PartialEnum $partial = 'alpha';

# With constant literal, should partially evaluate
my $partial_true = $partial ~~ 'alpha';   # Can optimize based on assignment
my $partial_false = $partial ~~ 'delta';  # Can detect impossibility

# Loop optimization with type constraints
type LoopType = 'continue' | 'break' | 'return';
my @loop_commands = ('continue', 'break', 'continue', 'return');

# Type constraints should enable loop optimizations
for my LoopType $cmd (@loop_commands) {
    given ($cmd) {
        when ('continue') { next }
        when ('break')    { last }
        when ('return')   { return }
    }
}

# Memory allocation optimization
type ShortString = 'a' | 'bb' | 'ccc';  # Different lengths but all short
my ShortString $short = 'a';

# Should not allocate temporary strings for comparison
my $mem_opt = $short ~~ qr/^[abc]+$/;

# Branch prediction friendly patterns
type FrequentEnum = 'common' | 'rare' | 'very_rare';
my FrequentEnum $freq = 'common';

# Most frequent case should be first for branch prediction
given ($freq) {
    when ('common')    { say 'frequent case' }     # Most common - predict taken
    when ('rare')      { say 'uncommon case' }     # Less common
    when ('very_rare') { say 'rare case' }         # Least common
}

# Cache-friendly data structure access
type CacheType = 'hot' | 'warm' | 'cold';
my CacheType $cache = 'hot';

# Enum values should be designed for cache efficiency
my %cache_data = (
    'hot'  => 'frequently_accessed',
    'warm' => 'sometimes_accessed',
    'cold' => 'rarely_accessed'
);

my $cache_lookup = $cache_data{$cache};  # Should be cache-friendly

# SIMD optimization potential (if applicable)
type SIMDable = 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8;
my @simd_data = (1, 2, 3, 4, 5, 6, 7, 8);

# Could potentially use SIMD for bulk comparisons
my @matches = grep { $_ ~~ SIMDable } @simd_data;  # All should match

# Stack frame optimization
sub optimized_match(SmallEnum $input) {
    # Small union types should minimize stack frame
    given ($input) {
        when ('red')   { return 1 }
        when ('green') { return 2 }
        when ('blue')  { return 3 }
    }
}

# Tail call optimization potential
sub tail_recursive_match(SmallEnum $input, $depth) {
    return $depth if $depth > 100;

    given ($input) {
        when ('red')   { return tail_recursive_match('green', $depth + 1) }
        when ('green') { return tail_recursive_match('blue', $depth + 1) }
        when ('blue')  { return tail_recursive_match('red', $depth + 1) }
    }
}

# Inlining optimization for small patterns
sub inline_candidate(SmallEnum $color) {
    # Small function with type constraints - good inline candidate
    return $color ~~ 'red' ? 'warm' : 'cool';
}

# Vectorization potential for bulk operations
sub bulk_match(@colors) {
    my @results;
    # Could potentially vectorize this loop
    for my SmallEnum $color (@colors) {
        push @results, ($color ~~ 'red' ? 1 : 0);
    }
    return @results;
}
```

# Expected AST

## Before Type Inference

### Text Format

```
AST {
  Path:
  Source length: 8934 characters
  Type Annotations:
    TypeAnnotation: SmallEnum = 'red' | 'green' | 'blue' at 1:1
    VarAnnotation: $color :: SmallEnum at 2:1
    TypeAnnotation: Priority = 1 | 2 | 3 | 4 | 5 at 17:1
    VarAnnotation: $pri :: Priority at 18:1
    TypeAnnotation: HttpCode = '200' | '201' | '404' | '500' at 27:1
    VarAnnotation: $code :: HttpCode at 28:1
    TypeAnnotation: SmallRange = 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 at 37:1
    VarAnnotation: $num :: SmallRange at 38:1
    TypeAnnotation: FileType = '.txt' | '.md' | '.rst' | '.pod' | '.html' at 43:1
    VarAnnotation: $ext :: FileType at 44:1
    TypeAnnotation: Protocol = 'http' | 'https' | 'ftp' | 'sftp' at 55:1
    VarAnnotation: $proto :: Protocol at 56:1
    TypeAnnotation: LargeEnum = 'opt1' | ... | 'opt20' at 63:1
    VarAnnotation: $large :: LargeEnum at 67:1
    TypeAnnotation: NestedPerf = [Int, Str] | [Str, Int] | [Int, Int] | [Str, Str] at 92:1
    VarAnnotation: $nested :: NestedPerf at 93:1
    TypeAnnotation: ConfigType = {mode => 'dev'} | {mode => 'prod'} | {mode => 'test'} at 102:1
    VarAnnotation: $config :: ConfigType at 103:1
    TypeAnnotation: OptimizedMatch = 'fast' | 'medium' | 'slow' at 111:1
    VarAnnotation: $speed :: OptimizedMatch at 112:1
    TypeAnnotation: FixedLength = 'ABC' | 'DEF' | 'GHI' at 118:1
    VarAnnotation: $fixed :: FixedLength at 119:1
    TypeAnnotation: VowelEnum = 'a' | 'e' | 'i' | 'o' | 'u' at 124:1
    VarAnnotation: $vowel :: VowelEnum at 125:1
    TypeAnnotation: MixedOptim = 1 | 'one' | 2 | 'two' | 3 | 'three' at 130:1
    VarAnnotation: $mixed :: MixedOptim at 131:1
    TypeAnnotation: Constant = 42 at 143:1
    VarAnnotation: $const :: Constant at 144:1
    TypeAnnotation: PartialEnum = 'alpha' | 'beta' | 'gamma' at 150:1
    VarAnnotation: $partial :: PartialEnum at 151:1
    TypeAnnotation: LoopType = 'continue' | 'break' | 'return' at 157:1
    TypeAnnotation: ShortString = 'a' | 'bb' | 'ccc' at 168:1
    VarAnnotation: $short :: ShortString at 169:1
    TypeAnnotation: FrequentEnum = 'common' | 'rare' | 'very_rare' at 174:1
    VarAnnotation: $freq :: FrequentEnum at 175:1
    TypeAnnotation: CacheType = 'hot' | 'warm' | 'cold' at 183:1
    VarAnnotation: $cache :: CacheType at 184:1
    TypeAnnotation: SIMDable = 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 at 194:1
    SubroutineAnnotation: optimized_match(SmallEnum $input) at 200:1
    SubroutineAnnotation: tail_recursive_match(SmallEnum $input, $depth) at 209:1
    SubroutineAnnotation: inline_candidate(SmallEnum $color) at 220:1
    SubroutineAnnotation: bulk_match(@colors) at 226:1
  Root: source_file
  Tree Structure:
  source_file
    comment("# Efficient literal matching with small unions")
    type_declaration
      type_name(SmallEnum)
      union_type
        literal('red')
        literal('green')
        literal('blue')
    var_decl
      type_expression(SmallEnum)
      scalar($color)
    comment("# This should compile to efficient jump table or switch statement")
    given_statement
      condition(scalar($color))
      given_block
        when_clause
          condition(literal('red'))
          block
        when_clause
          condition(literal('green'))
          block
        when_clause
          condition(literal('blue'))
          block
    comment("# Comparison: Manual if/elsif chain (baseline performance)")
    var_decl
      scalar($manual_color)
      literal('red')
    conditional_statement
      if_clause
        condition(equality_expression($manual_color eq 'red'))
        block
      elsif_clause
        condition(equality_expression($manual_color eq 'green'))
        block
      elsif_clause
        condition(equality_expression($manual_color eq 'blue'))
        block
    comment("# Numeric union optimization")
    type_declaration
      type_name(Priority)
      union_type
        literal(1)
        literal(2)
        literal(3)
        literal(4)
        literal(5)
    var_decl
      type_expression(Priority)
      scalar($pri)
    comment("# Should optimize to numeric comparison rather than string comparison")
    given_statement
      condition(scalar($pri))
      given_block
        when_clause
          condition(literal(1))
          block
        when_clause
          condition(literal(2))
          block
        when_clause
          condition(literal(3))
          block
        when_clause
          condition(literal(4))
          block
        when_clause
          condition(literal(5))
          block
}
```

## After Type Inference

### Text Format

```
# Same as before type inference - performance optimizations applied at compile time
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $color = 'red';

given ($color) {
    when ('red')   { say 'warm' }
    when ('green') { say 'nature' }
    when ('blue')  { say 'sky' }
}

my $manual_color = 'red';
if    ($manual_color eq 'red')   { say 'warm' }
elsif ($manual_color eq 'green') { say 'nature' }
elsif ($manual_color eq 'blue')  { say 'sky' }

my $pri = 3;

given ($pri) {
    when (1) { say 'critical' }
    when (2) { say 'high' }
    when (3) { say 'medium' }
    when (4) { say 'low' }
    when (5) { say 'minimal' }
}

my $code = '200';

my $is_success = $code ~~ ['200', '201'];
my $is_error = $code ~~ ['404', '500'];

my $untyped_code = '200';
my $untyped_success = $untyped_code ~~ ['200', '201', '202', '203', '204'];

my $num = 5;

my $in_range = $num ~~ [1..10];

my $ext = '.txt';

given ($ext) {
    when ('.txt')  { say 'text' }
    when ('.md')   { say 'markdown' }
    when ('.rst')  { say 'restructured' }
    when ('.pod')  { say 'perl_doc' }
    when ('.html') { say 'hypertext' }
}

my $proto = 'https';

my $is_secure = $proto ~~ qr/s$/;

my $untyped_proto = 'https';
my $untyped_secure = $untyped_proto ~~ qr/s$/;

my $large = 'opt10';

given ($large) {
    when ('opt1')  { say '1' }
    when ('opt2')  { say '2' }
    when ('opt3')  { say '3' }
    when ('opt4')  { say '4' }
    when ('opt5')  { say '5' }
    when ('opt6')  { say '6' }
    when ('opt7')  { say '7' }
    when ('opt8')  { say '8' }
    when ('opt9')  { say '9' }
    when ('opt10') { say '10' }
    when ('opt11') { say '11' }
    when ('opt12') { say '12' }
    when ('opt13') { say '13' }
    when ('opt14') { say '14' }
    when ('opt15') { say '15' }
    when ('opt16') { say '16' }
    when ('opt17') { say '17' }
    when ('opt18') { say '18' }
    when ('opt19') { say '19' }
    when ('opt20') { say '20' }
}

my $nested = [42, 'test'];

given ($nested) {
    when ([Int, Str]) { say 'int_str' }
    when ([Str, Int]) { say 'str_int' }
    when ([Int, Int]) { say 'int_int' }
    when ([Str, Str]) { say 'str_str' }
}

my $config = {mode => 'dev'};

given ($config) {
    when ({mode => 'dev'})  { say 'development' }
    when ({mode => 'prod'}) { say 'production' }
    when ({mode => 'test'}) { say 'testing' }
}

my $speed = 'fast';

my @speed_array = ('fast', 'medium', 'slow');
my $array_match = $speed ~~ @speed_array;

my $fixed = 'ABC';

my $length_opt = $fixed ~~ qr/^.{3}$/;

my $vowel = 'a';

my $is_vowel = $vowel ~~ qr/[aeiou]/;

my $mixed = 1;

given ($mixed) {
    when (1)       { say 'numeric one' }
    when (2)       { say 'numeric two' }
    when (3)       { say 'numeric three' }
    when ('one')   { say 'string one' }
    when ('two')   { say 'string two' }
    when ('three') { say 'string three' }
}

my $const = 42;

my $always_true = $const ~~ 42;
my $always_false = $const ~~ 43;

my $partial = 'alpha';

my $partial_true = $partial ~~ 'alpha';
my $partial_false = $partial ~~ 'delta';

my @loop_commands = ('continue', 'break', 'continue', 'return');

for my $cmd (@loop_commands) {
    given ($cmd) {
        when ('continue') { next }
        when ('break')    { last }
        when ('return')   { return }
    }
}

my $short = 'a';

my $mem_opt = $short ~~ qr/^[abc]+$/;

my $freq = 'common';

given ($freq) {
    when ('common')    { say 'frequent case' }
    when ('rare')      { say 'uncommon case' }
    when ('very_rare') { say 'rare case' }
}

my $cache = 'hot';

my %cache_data = (
    'hot'  => 'frequently_accessed',
    'warm' => 'sometimes_accessed',
    'cold' => 'rarely_accessed'
);

my $cache_lookup = $cache_data{$cache};

my @simd_data = (1, 2, 3, 4, 5, 6, 7, 8);

my @matches = grep { $_ ~~ SIMDable } @simd_data;

sub optimized_match($input) {
    given ($input) {
        when ('red')   { return 1 }
        when ('green') { return 2 }
        when ('blue')  { return 3 }
    }
}

sub tail_recursive_match($input, $depth) {
    return $depth if $depth > 100;

    given ($input) {
        when ('red')   { return tail_recursive_match('green', $depth + 1) }
        when ('green') { return tail_recursive_match('blue', $depth + 1) }
        when ('blue')  { return tail_recursive_match('red', $depth + 1) }
    }
}

sub inline_candidate($color) {
    return $color ~~ 'red' ? 'warm' : 'cool';
}

sub bulk_match(@colors) {
    my @results;
    for my $color (@colors) {
        push @results, ($color ~~ 'red' ? 1 : 0);
    }
    return @results;
}
```

## Typed Perl Output

```perl
type SmallEnum = 'red' | 'green' | 'blue';
my SmallEnum $color = 'red';

given ($color) {
    when ('red')   { say 'warm' }
    when ('green') { say 'nature' }
    when ('blue')  { say 'sky' }
}

my $manual_color = 'red';
if    ($manual_color eq 'red')   { say 'warm' }
elsif ($manual_color eq 'green') { say 'nature' }
elsif ($manual_color eq 'blue')  { say 'sky' }

type Priority = 1 | 2 | 3 | 4 | 5;
my Priority $pri = 3;

given ($pri) {
    when (1) { say 'critical' }
    when (2) { say 'high' }
    when (3) { say 'medium' }
    when (4) { say 'low' }
    when (5) { say 'minimal' }
}

type HttpCode = '200' | '201' | '404' | '500';
my HttpCode $code = '200';

my $is_success = $code ~~ ['200', '201'];
my $is_error = $code ~~ ['404', '500'];

my $untyped_code = '200';
my $untyped_success = $untyped_code ~~ ['200', '201', '202', '203', '204'];

type SmallRange = 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10;
my SmallRange $num = 5;

my $in_range = $num ~~ [1..10];

type FileType = '.txt' | '.md' | '.rst' | '.pod' | '.html';
my FileType $ext = '.txt';

given ($ext) {
    when ('.txt')  { say 'text' }
    when ('.md')   { say 'markdown' }
    when ('.rst')  { say 'restructured' }
    when ('.pod')  { say 'perl_doc' }
    when ('.html') { say 'hypertext' }
}

type Protocol = 'http' | 'https' | 'ftp' | 'sftp';
my Protocol $proto = 'https';

my $is_secure = $proto ~~ qr/s$/;

my $untyped_proto = 'https';
my $untyped_secure = $untyped_proto ~~ qr/s$/;

type LargeEnum = 'opt1' | 'opt2' | 'opt3' | 'opt4' | 'opt5' |
                'opt6' | 'opt7' | 'opt8' | 'opt9' | 'opt10' |
                'opt11' | 'opt12' | 'opt13' | 'opt14' | 'opt15' |
                'opt16' | 'opt17' | 'opt18' | 'opt19' | 'opt20';
my LargeEnum $large = 'opt10';

given ($large) {
    when ('opt1')  { say '1' }
    when ('opt2')  { say '2' }
    when ('opt3')  { say '3' }
    when ('opt4')  { say '4' }
    when ('opt5')  { say '5' }
    when ('opt6')  { say '6' }
    when ('opt7')  { say '7' }
    when ('opt8')  { say '8' }
    when ('opt9')  { say '9' }
    when ('opt10') { say '10' }
    when ('opt11') { say '11' }
    when ('opt12') { say '12' }
    when ('opt13') { say '13' }
    when ('opt14') { say '14' }
    when ('opt15') { say '15' }
    when ('opt16') { say '16' }
    when ('opt17') { say '17' }
    when ('opt18') { say '18' }
    when ('opt19') { say '19' }
    when ('opt20') { say '20' }
}

type NestedPerf = [Int, Str] | [Str, Int] | [Int, Int] | [Str, Str];
my NestedPerf $nested = [42, 'test'];

given ($nested) {
    when ([Int, Str]) { say 'int_str' }
    when ([Str, Int]) { say 'str_int' }
    when ([Int, Int]) { say 'int_int' }
    when ([Str, Str]) { say 'str_str' }
}

type ConfigType = {mode => 'dev'} | {mode => 'prod'} | {mode => 'test'};
my ConfigType $config = {mode => 'dev'};

given ($config) {
    when ({mode => 'dev'})  { say 'development' }
    when ({mode => 'prod'}) { say 'production' }
    when ({mode => 'test'}) { say 'testing' }
}

type OptimizedMatch = 'fast' | 'medium' | 'slow';
my OptimizedMatch $speed = 'fast';

my @speed_array = ('fast', 'medium', 'slow');
my $array_match = $speed ~~ @speed_array;

type FixedLength = 'ABC' | 'DEF' | 'GHI';
my FixedLength $fixed = 'ABC';

my $length_opt = $fixed ~~ qr/^.{3}$/;

type VowelEnum = 'a' | 'e' | 'i' | 'o' | 'u';
my VowelEnum $vowel = 'a';

my $is_vowel = $vowel ~~ qr/[aeiou]/;

type MixedOptim = 1 | 'one' | 2 | 'two' | 3 | 'three';
my MixedOptim $mixed = 1;

given ($mixed) {
    when (1)       { say 'numeric one' }
    when (2)       { say 'numeric two' }
    when (3)       { say 'numeric three' }
    when ('one')   { say 'string one' }
    when ('two')   { say 'string two' }
    when ('three') { say 'string three' }
}

type Constant = 42;
my Constant $const = 42;

my $always_true = $const ~~ 42;
my $always_false = $const ~~ 43;

type PartialEnum = 'alpha' | 'beta' | 'gamma';
my PartialEnum $partial = 'alpha';

my $partial_true = $partial ~~ 'alpha';
my $partial_false = $partial ~~ 'delta';

type LoopType = 'continue' | 'break' | 'return';
my @loop_commands = ('continue', 'break', 'continue', 'return');

for my LoopType $cmd (@loop_commands) {
    given ($cmd) {
        when ('continue') { next }
        when ('break')    { last }
        when ('return')   { return }
    }
}

type ShortString = 'a' | 'bb' | 'ccc';
my ShortString $short = 'a';

my $mem_opt = $short ~~ qr/^[abc]+$/;

type FrequentEnum = 'common' | 'rare' | 'very_rare';
my FrequentEnum $freq = 'common';

given ($freq) {
    when ('common')    { say 'frequent case' }
    when ('rare')      { say 'uncommon case' }
    when ('very_rare') { say 'rare case' }
}

type CacheType = 'hot' | 'warm' | 'cold';
my CacheType $cache = 'hot';

my %cache_data = (
    'hot'  => 'frequently_accessed',
    'warm' => 'sometimes_accessed',
    'cold' => 'rarely_accessed'
);

my $cache_lookup = $cache_data{$cache};

type SIMDable = 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8;
my @simd_data = (1, 2, 3, 4, 5, 6, 7, 8);

my @matches = grep { $_ ~~ SIMDable } @simd_data;

sub optimized_match(SmallEnum $input) {
    given ($input) {
        when ('red')   { return 1 }
        when ('green') { return 2 }
        when ('blue')  { return 3 }
    }
}

sub tail_recursive_match(SmallEnum $input, $depth) {
    return $depth if $depth > 100;

    given ($input) {
        when ('red')   { return tail_recursive_match('green', $depth + 1) }
        when ('green') { return tail_recursive_match('blue', $depth + 1) }
        when ('blue')  { return tail_recursive_match('red', $depth + 1) }
    }
}

sub inline_candidate(SmallEnum $color) {
    return $color ~~ 'red' ? 'warm' : 'cool';
}

sub bulk_match(@colors) {
    my @results;
    for my SmallEnum $color (@colors) {
        push @results, ($color ~~ 'red' ? 1 : 0);
    }
    return @results;
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none - performance tests should not have type errors)
```

# Performance Expectations

```
Pattern matching with type constraints should achieve:
- 20-50% better performance than untyped pattern matching
- O(1) lookup for small unions (3-10 values) via jump tables
- O(log n) lookup for medium unions (11-100 values) via binary search
- Compile-time optimization of impossible patterns
- Memory allocation optimization for string comparisons
- Branch prediction optimization for frequency-ordered patterns
- Potential SIMD optimization for bulk operations
```
