---
category: typed-perl
subcategory: smartmatch
tags:
    - edge-cases
    - controversial-behavior
    - reference-matching
    - object-dispatch
    - complex-structures
type_check: true
should_error: true
expected_error: "parse error"
---

# Controversial Smartmatch Edge Cases with Type Constraints

Tests controversial and complex smartmatch behaviors to validate that PSC's type system
makes these edge cases predictable and safe, unlike in standard Perl where they can be
unpredictable or dangerous.

```perl
# Reference smartmatch - PSC should make this predictable
type RefValue = [] | {} | qr/pattern/ | sub { };
my RefValue $ref1 = [];
my RefValue $ref2 = {};
my RefValue $ref3 = qr/test/;

# Array reference comparisons with type constraints
my @array1 = (1, 2, 3);
my @array2 = (1, 2, 3);
my $same_content = $ref1 ~~ [];              # Empty vs empty
my $different_refs = \@array1 ~~ \@array2;   # Same content, different refs
my $ref_vs_literal = $ref1 ~~ [1, 2, 3];     # Array ref vs array literal

# Hash reference comparisons
my %hash1 = (key => 'value');
my %hash2 = (key => 'value');
my $hash_same = $ref2 ~~ {};                 # Empty vs empty
my $hash_different = \%hash1 ~~ \%hash2;     # Same content, different refs
my $hash_vs_literal = $ref2 ~~ {key => 'value'}; # Hash ref vs hash literal

# Regex object behavior with type constraints
type RegexValue = qr/\d+/ | qr/\w+/ | qr/\s+/;
my RegexValue $pattern = qr/\d+/;

my $regex_vs_string = $pattern ~~ '\d+';     # Regex object vs string pattern
my $regex_vs_regex = $pattern ~~ qr/\d+/;    # Regex object vs regex object
my $string_vs_regex = '123' ~~ $pattern;     # String vs regex object (reverse)

# Code reference matching - extremely controversial
type CodeValue = sub { 'hello' } | sub { 'world' };
my CodeValue $code1 = sub { 'hello' };
my CodeValue $code2 = sub { 'world' };

my $code_vs_code = $code1 ~~ $code2;         # Code ref vs code ref
my $code_vs_literal = $code1 ~~ sub { 'hello' }; # Same code, different refs
my $code_identity = $code1 ~~ $code1;        # Same reference

# Object method dispatch - controversial smartmatch behavior
class TestObject {
    method matches($other) {
        return $other eq 'test';
    }
}

type ObjectValue = TestObject | 'string' | 42;
my ObjectValue $obj = TestObject->new();

# Object vs object smartmatch (should call overloaded ~~ or cmp?)
my $obj_vs_string = $obj ~~ 'test';          # Should this call obj->matches?
my $obj_vs_obj = $obj ~~ TestObject->new(); # Object vs object comparison

# Nested data structure matching - deep comparison issues
type NestedData = [[1,2], [3,4]] | {'a' => [1,2], 'b' => [3,4]};
my NestedData $nested1 = [[1,2], [3,4]];
my NestedData $nested2 = {'a' => [1,2], 'b' => [3,4]};

# Deep structure comparison
my $nested_same = $nested1 ~~ [[1,2], [3,4]];       # Same structure
my $nested_different = $nested1 ~~ [[1,2], [3,5]];  # Different values
my $array_vs_hash = $nested1 ~~ {'a' => [1,2]};     # Array vs hash structure

# Mixed reference types in complex structures
type ComplexRef = [[], {}] | {array => [], hash => {}};
my ComplexRef $complex = [[], {}];

my $complex_match = $complex ~~ [[], {}];           # Array of refs vs literal
my $partial_complex = $complex ~~ [[]];             # Partial structure match

# Scalar reference edge cases
type ScalarRef = \42 | \'string' | \undef;
my ScalarRef $scalar_ref = \42;

my $scalar_deref = $scalar_ref ~~ 42;              # Scalar ref vs value
my $scalar_ref_match = $scalar_ref ~~ \42;         # Scalar ref vs scalar ref
my $different_scalar_ref = $scalar_ref ~~ \43;     # Different scalar refs

# Typeglob and special reference handling
type SpecialRef = \*STDIN | \*STDOUT | \&print;
my SpecialRef $special = \*STDIN;

my $glob_match = $special ~~ \*STDIN;              # Typeglob comparison
my $glob_vs_different = $special ~~ \*STDOUT;      # Different typeglobs

# IO handle and file descriptor edge cases
type IOValue = *STDIN{IO} | *STDOUT{IO} | 42;
my IOValue $io = *STDIN{IO};

my $io_match = $io ~~ *STDIN{IO};                  # IO handle comparison
my $io_vs_fd = $io ~~ 0;                           # IO handle vs file descriptor

# Overloaded object behavior
class OverloadedObj {
    method new() { bless {}, __CLASS__ }

    # Overload smartmatch operator
    use overload '~~' => sub {
        my ($self, $other, $reversed) = @_;
        return ref($other) eq ref($self);
    };
}

type OverloadedValue = OverloadedObj | 'string';
my OverloadedValue $overloaded1 = OverloadedObj->new();
my OverloadedValue $overloaded2 = OverloadedObj->new();

my $overloaded_match = $overloaded1 ~~ $overloaded2;    # Should use overload
my $overloaded_vs_string = $overloaded1 ~~ 'string';    # Overloaded vs string

# Circular reference handling
type CircularRef = {} | [];
my CircularRef $circular = {};
$circular->{self} = $circular;  # Create circular reference

my $circular_match = $circular ~~ {};                   # Circular vs empty
my $circular_vs_circular = $circular ~~ $circular;      # Same circular ref

# Weak reference behavior with smartmatch
use Scalar::Util 'weaken';

type WeakRef = [] | {};
my WeakRef $weak_target = [];
my $weak_ref = $weak_target;
weaken($weak_ref);

my $weak_vs_strong = $weak_ref ~~ $weak_target;         # Weak vs strong ref
my $weak_vs_literal = $weak_ref ~~ [];                  # Weak ref vs literal

# Thread-shared variable edge cases (if threading enabled)
# use threads::shared;

# my shared %shared_hash;
# type SharedValue = %shared_hash | {};
# my SharedValue $shared = %shared_hash;

# my $shared_match = $shared ~~ {};                       # Shared vs literal
# my $shared_vs_shared = $shared ~~ %shared_hash;         # Shared vs shared

# Tied variable smartmatch behavior
package TiedHash {
    sub TIEHASH { bless {}, __PACKAGE__ }
    sub FETCH { $_[0]->{$_[1]} }
    sub STORE { $_[0]->{$_[1]} = $_[2] }
    sub DELETE { delete $_[0]->{$_[1]} }
    sub CLEAR { %{$_[0]} = () }
    sub EXISTS { exists $_[0]->{$_[1]} }
    sub FIRSTKEY { each %{$_[0]} }
    sub NEXTKEY { each %{$_[0]} }
}

tie my %tied_hash, 'TiedHash';
%tied_hash = (key => 'value');

type TiedValue = %tied_hash | {};
my TiedValue $tied = %tied_hash;

my $tied_match = $tied ~~ {key => 'value'};            # Tied vs literal
my $tied_vs_tied = $tied ~~ %tied_hash;                # Tied vs tied

# Unicode and encoding edge cases
type UnicodeValue = 'café' | 'naïve' | '北京';
my UnicodeValue $unicode = 'café';

my $unicode_match = $unicode ~~ 'café';                # Unicode exact match
my $unicode_normalize = $unicode ~~ 'cafe';            # Without diacritics
my $unicode_encoding = $unicode ~~ "caf\x{e9}";       # Different encoding

# Numeric precision and floating point edge cases
type FloatValue = 3.14159 | 2.71828 | 1.41421;
my FloatValue $pi = 3.14159;

my $float_exact = $pi ~~ 3.14159;                      # Exact match
my $float_imprecise = $pi ~~ 3.14160;                  # Close but not exact
my $float_string = $pi ~~ '3.14159';                   # Float vs string

# BigInt and arbitrary precision edge cases
use Math::BigInt;

type BigValue = Math::BigInt->new('123456789012345678901234567890');
my BigValue $big = Math::BigInt->new('123456789012345678901234567890');

my $big_match = $big ~~ Math::BigInt->new('123456789012345678901234567890');
my $big_vs_string = $big ~~ '123456789012345678901234567890';
my $big_vs_regular = $big ~~ 123456789012345678901234567890;

# Version object comparison
use version;

type VersionValue = version->parse('v1.2.3') | version->parse('v2.0.0');
my VersionValue $ver = version->parse('v1.2.3');

my $version_match = $ver ~~ version->parse('v1.2.3');  # Version vs version
my $version_string = $ver ~~ 'v1.2.3';                 # Version vs string
my $version_number = $ver ~~ 1.2.3;                    # Version vs number

# DateTime and temporal edge cases
use DateTime;

type DateValue = DateTime->new(year => 2023, month => 1, day => 1);
my DateValue $date = DateTime->new(year => 2023, month => 1, day => 1);

my $date_match = $date ~~ DateTime->new(year => 2023, month => 1, day => 1);
my $date_string = $date ~~ '2023-01-01T00:00:00';      # DateTime vs string
my $date_epoch = $date ~~ 1672531200;                  # DateTime vs epoch

# Path and filesystem object edge cases
use Path::Tiny;

type PathValue = Path::Tiny->new('/tmp') | Path::Tiny->new('/var');
my PathValue $path = Path::Tiny->new('/tmp');

my $path_match = $path ~~ Path::Tiny->new('/tmp');     # Path vs path
my $path_string = $path ~~ '/tmp';                     # Path vs string
my $path_exists = $path ~~ -d;                         # Path vs file test
```

# Expected AST

## Before Type Inference

### Text Format

```
AST {
  Path:
  Source length: 8923 characters
  Type Annotations:
    TypeAnnotation: RefValue = [] | {} | qr/pattern/ | sub { } at 1:1
    VarAnnotation: $ref1 :: RefValue at 2:1
    VarAnnotation: $ref2 :: RefValue at 3:1
    VarAnnotation: $ref3 :: RefValue at 4:1
    TypeAnnotation: RegexValue = qr/\d+/ | qr/\w+/ | qr/\s+/ at 25:1
    VarAnnotation: $pattern :: RegexValue at 26:1
    TypeAnnotation: CodeValue = sub { 'hello' } | sub { 'world' } at 32:1
    VarAnnotation: $code1 :: CodeValue at 33:1
    VarAnnotation: $code2 :: CodeValue at 34:1
    ClassAnnotation: TestObject at 40:1
    TypeAnnotation: ObjectValue = TestObject | 'string' | 42 at 46:1
    VarAnnotation: $obj :: ObjectValue at 47:1
    TypeAnnotation: NestedData = [[1,2], [3,4]] | {'a' => [1,2], 'b' => [3,4]} at 53:1
    VarAnnotation: $nested1 :: NestedData at 54:1
    VarAnnotation: $nested2 :: NestedData at 55:1
    TypeAnnotation: ComplexRef = [[], {}] | {array => [], hash => {}} at 63:1
    VarAnnotation: $complex :: ComplexRef at 64:1
    TypeAnnotation: ScalarRef = \42 | \'string' | \undef at 69:1
    VarAnnotation: $scalar_ref :: ScalarRef at 70:1
    TypeAnnotation: SpecialRef = \*STDIN | \*STDOUT | \&print at 76:1
    VarAnnotation: $special :: SpecialRef at 77:1
    TypeAnnotation: IOValue = *STDIN{IO} | *STDOUT{IO} | 42 at 82:1
    VarAnnotation: $io :: IOValue at 83:1
    ClassAnnotation: OverloadedObj at 88:1
    TypeAnnotation: OverloadedValue = OverloadedObj | 'string' at 99:1
    VarAnnotation: $overloaded1 :: OverloadedValue at 100:1
    VarAnnotation: $overloaded2 :: OverloadedValue at 101:1
    TypeAnnotation: CircularRef = {} | [] at 106:1
    VarAnnotation: $circular :: CircularRef at 107:1
    TypeAnnotation: WeakRef = [] | {} at 115:1
    VarAnnotation: $weak_target :: WeakRef at 116:1
    PackageAnnotation: TiedHash at 129:1
    TypeAnnotation: TiedValue = %tied_hash | {} at 142:1
    VarAnnotation: $tied :: TiedValue at 143:1
    TypeAnnotation: UnicodeValue = 'café' | 'naïve' | '北京' at 149:1
    VarAnnotation: $unicode :: UnicodeValue at 150:1
    TypeAnnotation: FloatValue = 3.14159 | 2.71828 | 1.41421 at 156:1
    VarAnnotation: $pi :: FloatValue at 157:1
    TypeAnnotation: BigValue = Math::BigInt->new('123456789012345678901234567890') at 165:1
    VarAnnotation: $big :: BigValue at 166:1
    TypeAnnotation: VersionValue = version->parse('v1.2.3') | version->parse('v2.0.0') at 174:1
    VarAnnotation: $ver :: VersionValue at 175:1
    TypeAnnotation: DateValue = DateTime->new(year => 2023, month => 1, day => 1) at 183:1
    VarAnnotation: $date :: DateValue at 184:1
    TypeAnnotation: PathValue = Path::Tiny->new('/tmp') | Path::Tiny->new('/var') at 192:1
    VarAnnotation: $path :: PathValue at 193:1
  Root: source_file
  Tree Structure:
  source_file
    type_declaration
      type_name(RefValue)
      union_type
        array_constructor
        hash_constructor
        regex_pattern(qr/pattern/)
        subroutine_declaration
    var_decl
      type_expression(RefValue)
      scalar($ref1)
    var_decl
      type_expression(RefValue)
      scalar($ref2)
    var_decl
      type_expression(RefValue)
      scalar($ref3)
    var_decl
      array(@array1)
      array_constructor(literal(1), literal(2), literal(3))
    var_decl
      array(@array2)
      array_constructor(literal(1), literal(2), literal(3))
    var_decl
      scalar($same_content)
      smartmatch_expression
        scalar($ref1)
        array_constructor
    var_decl
      scalar($different_refs)
      smartmatch_expression
        reference_expression(\@array1)
        reference_expression(\@array2)
    var_decl
      scalar($ref_vs_literal)
      smartmatch_expression
        scalar($ref1)
        array_constructor(literal(1), literal(2), literal(3))
    var_decl
      hash(%hash1)
      hash_constructor(key => 'value')
    var_decl
      hash(%hash2)
      hash_constructor(key => 'value')
    var_decl
      scalar($hash_same)
      smartmatch_expression
        scalar($ref2)
        hash_constructor
    var_decl
      scalar($hash_different)
      smartmatch_expression
        reference_expression(\%hash1)
        reference_expression(\%hash2)
    var_decl
      scalar($hash_vs_literal)
      smartmatch_expression
        scalar($ref2)
        hash_constructor(key => 'value')
    type_declaration
      type_name(RegexValue)
      union_type
        regex_pattern(qr/\d+/)
        regex_pattern(qr/\w+/)
        regex_pattern(qr/\s+/)
    var_decl
      type_expression(RegexValue)
      scalar($pattern)
    var_decl
      scalar($regex_vs_string)
      smartmatch_expression
        scalar($pattern)
        literal('\d+')
    var_decl
      scalar($regex_vs_regex)
      smartmatch_expression
        scalar($pattern)
        regex_pattern(qr/\d+/)
    var_decl
      scalar($string_vs_regex)
      smartmatch_expression
        literal('123')
        scalar($pattern)
    type_declaration
      type_name(CodeValue)
      union_type
        subroutine_declaration(sub { 'hello' })
        subroutine_declaration(sub { 'world' })
    var_decl
      type_expression(CodeValue)
      scalar($code1)
    var_decl
      type_expression(CodeValue)
      scalar($code2)
    var_decl
      scalar($code_vs_code)
      smartmatch_expression
        scalar($code1)
        scalar($code2)
    var_decl
      scalar($code_vs_literal)
      smartmatch_expression
        scalar($code1)
        subroutine_declaration(sub { 'hello' })
    var_decl
      scalar($code_identity)
      smartmatch_expression
        scalar($code1)
        scalar($code1)
    class_declaration
      class_name(TestObject)
      method_declaration
        method_name(matches)
        parameter_list($other)
        block
    type_declaration
      type_name(ObjectValue)
      union_type
        class_type(TestObject)
        literal('string')
        literal(42)
    var_decl
      type_expression(ObjectValue)
      scalar($obj)
    var_decl
      scalar($obj_vs_string)
      smartmatch_expression
        scalar($obj)
        literal('test')
    var_decl
      scalar($obj_vs_obj)
      smartmatch_expression
        scalar($obj)
        method_call(TestObject->new())
    type_declaration
      type_name(NestedData)
      union_type
        array_constructor(
          array_constructor(literal(1), literal(2)),
          array_constructor(literal(3), literal(4))
        )
        hash_constructor(
          'a' => array_constructor(literal(1), literal(2)),
          'b' => array_constructor(literal(3), literal(4))
        )
    var_decl
      type_expression(NestedData)
      scalar($nested1)
    var_decl
      type_expression(NestedData)
      scalar($nested2)
    var_decl
      scalar($nested_same)
      smartmatch_expression
        scalar($nested1)
        array_constructor(
          array_constructor(literal(1), literal(2)),
          array_constructor(literal(3), literal(4))
        )
    var_decl
      scalar($nested_different)
      smartmatch_expression
        scalar($nested1)
        array_constructor(
          array_constructor(literal(1), literal(2)),
          array_constructor(literal(3), literal(5))
        )
    var_decl
      scalar($array_vs_hash)
      smartmatch_expression
        scalar($nested1)
        hash_constructor('a' => array_constructor(literal(1), literal(2)))
}
```

## After Type Inference

### Text Format

```
# Same as before type inference - edge case handling not yet implemented
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $ref1 = [];
my $ref2 = {};
my $ref3 = qr/test/;

my @array1 = (1, 2, 3);
my @array2 = (1, 2, 3);
my $same_content = $ref1 ~~ [];
my $different_refs = \@array1 ~~ \@array2;
my $ref_vs_literal = $ref1 ~~ [1, 2, 3];

my %hash1 = (key => 'value');
my %hash2 = (key => 'value');
my $hash_same = $ref2 ~~ {};
my $hash_different = \%hash1 ~~ \%hash2;
my $hash_vs_literal = $ref2 ~~ {key => 'value'};

my $pattern = qr/\d+/;

my $regex_vs_string = $pattern ~~ '\d+';
my $regex_vs_regex = $pattern ~~ qr/\d+/;
my $string_vs_regex = '123' ~~ $pattern;

my $code1 = sub { 'hello' };
my $code2 = sub { 'world' };

my $code_vs_code = $code1 ~~ $code2;
my $code_vs_literal = $code1 ~~ sub { 'hello' };
my $code_identity = $code1 ~~ $code1;

class TestObject {
    method matches($other) {
        return $other eq 'test';
    }
}

my $obj = TestObject->new();

my $obj_vs_string = $obj ~~ 'test';
my $obj_vs_obj = $obj ~~ TestObject->new();

my $nested1 = [[1,2], [3,4]];
my $nested2 = {'a' => [1,2], 'b' => [3,4]};

my $nested_same = $nested1 ~~ [[1,2], [3,4]];
my $nested_different = $nested1 ~~ [[1,2], [3,5]];
my $array_vs_hash = $nested1 ~~ {'a' => [1,2]};

my $complex = [[], {}];

my $complex_match = $complex ~~ [[], {}];
my $partial_complex = $complex ~~ [[]];

my $scalar_ref = \42;

my $scalar_deref = $scalar_ref ~~ 42;
my $scalar_ref_match = $scalar_ref ~~ \42;
my $different_scalar_ref = $scalar_ref ~~ \43;

my $special = \*STDIN;

my $glob_match = $special ~~ \*STDIN;
my $glob_vs_different = $special ~~ \*STDOUT;

my $io = *STDIN{IO};

my $io_match = $io ~~ *STDIN{IO};
my $io_vs_fd = $io ~~ 0;

class OverloadedObj {
    method new() { bless {}, __CLASS__ }

    use overload '~~' => sub {
        my ($self, $other, $reversed) = @_;
        return ref($other) eq ref($self);
    };
}

my $overloaded1 = OverloadedObj->new();
my $overloaded2 = OverloadedObj->new();

my $overloaded_match = $overloaded1 ~~ $overloaded2;
my $overloaded_vs_string = $overloaded1 ~~ 'string';

my $circular = {};
$circular->{self} = $circular;

my $circular_match = $circular ~~ {};
my $circular_vs_circular = $circular ~~ $circular;

use Scalar::Util 'weaken';

my $weak_target = [];
my $weak_ref = $weak_target;
weaken($weak_ref);

my $weak_vs_strong = $weak_ref ~~ $weak_target;
my $weak_vs_literal = $weak_ref ~~ [];

package TiedHash {
    sub TIEHASH { bless {}, __PACKAGE__ }
    sub FETCH { $_[0]->{$_[1]} }
    sub STORE { $_[0]->{$_[1]} = $_[2] }
    sub DELETE { delete $_[0]->{$_[1]} }
    sub CLEAR { %{$_[0]} = () }
    sub EXISTS { exists $_[0]->{$_[1]} }
    sub FIRSTKEY { each %{$_[0]} }
    sub NEXTKEY { each %{$_[0]} }
}

tie my %tied_hash, 'TiedHash';
%tied_hash = (key => 'value');

my $tied = %tied_hash;

my $tied_match = $tied ~~ {key => 'value'};
my $tied_vs_tied = $tied ~~ %tied_hash;

my $unicode = 'café';

my $unicode_match = $unicode ~~ 'café';
my $unicode_normalize = $unicode ~~ 'cafe';
my $unicode_encoding = $unicode ~~ "caf\x{e9}";

my $pi = 3.14159;

my $float_exact = $pi ~~ 3.14159;
my $float_imprecise = $pi ~~ 3.14160;
my $float_string = $pi ~~ '3.14159';

use Math::BigInt;

my $big = Math::BigInt->new('123456789012345678901234567890');

my $big_match = $big ~~ Math::BigInt->new('123456789012345678901234567890');
my $big_vs_string = $big ~~ '123456789012345678901234567890';
my $big_vs_regular = $big ~~ 123456789012345678901234567890;

use version;

my $ver = version->parse('v1.2.3');

my $version_match = $ver ~~ version->parse('v1.2.3');
my $version_string = $ver ~~ 'v1.2.3';
my $version_number = $ver ~~ 1.2.3;

use DateTime;

my $date = DateTime->new(year => 2023, month => 1, day => 1);

my $date_match = $date ~~ DateTime->new(year => 2023, month => 1, day => 1);
my $date_string = $date ~~ '2023-01-01T00:00:00';
my $date_epoch = $date ~~ 1672531200;

use Path::Tiny;

my $path = Path::Tiny->new('/tmp');

my $path_match = $path ~~ Path::Tiny->new('/tmp');
my $path_string = $path ~~ '/tmp';
my $path_exists = $path ~~ -d;
```

## Typed Perl Output

```perl
type RefValue = [] | {} | qr/pattern/ | sub { };
my RefValue $ref1 = [];
my RefValue $ref2 = {};
my RefValue $ref3 = qr/test/;

my @array1 = (1, 2, 3);
my @array2 = (1, 2, 3);
my $same_content = $ref1 ~~ [];
my $different_refs = \@array1 ~~ \@array2;
my $ref_vs_literal = $ref1 ~~ [1, 2, 3];

my %hash1 = (key => 'value');
my %hash2 = (key => 'value');
my $hash_same = $ref2 ~~ {};
my $hash_different = \%hash1 ~~ \%hash2;
my $hash_vs_literal = $ref2 ~~ {key => 'value'};

type RegexValue = qr/\d+/ | qr/\w+/ | qr/\s+/;
my RegexValue $pattern = qr/\d+/;

my $regex_vs_string = $pattern ~~ '\d+';
my $regex_vs_regex = $pattern ~~ qr/\d+/;
my $string_vs_regex = '123' ~~ $pattern;

type CodeValue = sub { 'hello' } | sub { 'world' };
my CodeValue $code1 = sub { 'hello' };
my CodeValue $code2 = sub { 'world' };

my $code_vs_code = $code1 ~~ $code2;
my $code_vs_literal = $code1 ~~ sub { 'hello' };
my $code_identity = $code1 ~~ $code1;

class TestObject {
    method matches($other) {
        return $other eq 'test';
    }
}

type ObjectValue = TestObject | 'string' | 42;
my ObjectValue $obj = TestObject->new();

my $obj_vs_string = $obj ~~ 'test';
my $obj_vs_obj = $obj ~~ TestObject->new();

type NestedData = [[1,2], [3,4]] | {'a' => [1,2], 'b' => [3,4]};
my NestedData $nested1 = [[1,2], [3,4]];
my NestedData $nested2 = {'a' => [1,2], 'b' => [3,4]};

my $nested_same = $nested1 ~~ [[1,2], [3,4]];
my $nested_different = $nested1 ~~ [[1,2], [3,5]];
my $array_vs_hash = $nested1 ~~ {'a' => [1,2]};

type ComplexRef = [[], {}] | {array => [], hash => {}};
my ComplexRef $complex = [[], {}];

my $complex_match = $complex ~~ [[], {}];
my $partial_complex = $complex ~~ [[]];

type ScalarRef = \42 | \'string' | \undef;
my ScalarRef $scalar_ref = \42;

my $scalar_deref = $scalar_ref ~~ 42;
my $scalar_ref_match = $scalar_ref ~~ \42;
my $different_scalar_ref = $scalar_ref ~~ \43;

type SpecialRef = \*STDIN | \*STDOUT | \&print;
my SpecialRef $special = \*STDIN;

my $glob_match = $special ~~ \*STDIN;
my $glob_vs_different = $special ~~ \*STDOUT;

type IOValue = *STDIN{IO} | *STDOUT{IO} | 42;
my IOValue $io = *STDIN{IO};

my $io_match = $io ~~ *STDIN{IO};
my $io_vs_fd = $io ~~ 0;

class OverloadedObj {
    method new() { bless {}, __CLASS__ }

    use overload '~~' => sub {
        my ($self, $other, $reversed) = @_;
        return ref($other) eq ref($self);
    };
}

type OverloadedValue = OverloadedObj | 'string';
my OverloadedValue $overloaded1 = OverloadedObj->new();
my OverloadedValue $overloaded2 = OverloadedObj->new();

my $overloaded_match = $overloaded1 ~~ $overloaded2;
my $overloaded_vs_string = $overloaded1 ~~ 'string';

type CircularRef = {} | [];
my CircularRef $circular = {};
$circular->{self} = $circular;

my $circular_match = $circular ~~ {};
my $circular_vs_circular = $circular ~~ $circular;

use Scalar::Util 'weaken';

type WeakRef = [] | {};
my WeakRef $weak_target = [];
my $weak_ref = $weak_target;
weaken($weak_ref);

my $weak_vs_strong = $weak_ref ~~ $weak_target;
my $weak_vs_literal = $weak_ref ~~ [];

package TiedHash {
    sub TIEHASH { bless {}, __PACKAGE__ }
    sub FETCH { $_[0]->{$_[1]} }
    sub STORE { $_[0]->{$_[1]} = $_[2] }
    sub DELETE { delete $_[0]->{$_[1]} }
    sub CLEAR { %{$_[0]} = () }
    sub EXISTS { exists $_[0]->{$_[1]} }
    sub FIRSTKEY { each %{$_[0]} }
    sub NEXTKEY { each %{$_[0]} }
}

tie my %tied_hash, 'TiedHash';
%tied_hash = (key => 'value');

type TiedValue = %tied_hash | {};
my TiedValue $tied = %tied_hash;

my $tied_match = $tied ~~ {key => 'value'};
my $tied_vs_tied = $tied ~~ %tied_hash;

type UnicodeValue = 'café' | 'naïve' | '北京';
my UnicodeValue $unicode = 'café';

my $unicode_match = $unicode ~~ 'café';
my $unicode_normalize = $unicode ~~ 'cafe';
my $unicode_encoding = $unicode ~~ "caf\x{e9}";

type FloatValue = 3.14159 | 2.71828 | 1.41421;
my FloatValue $pi = 3.14159;

my $float_exact = $pi ~~ 3.14159;
my $float_imprecise = $pi ~~ 3.14160;
my $float_string = $pi ~~ '3.14159';

use Math::BigInt;

type BigValue = Math::BigInt->new('123456789012345678901234567890');
my BigValue $big = Math::BigInt->new('123456789012345678901234567890');

my $big_match = $big ~~ Math::BigInt->new('123456789012345678901234567890');
my $big_vs_string = $big ~~ '123456789012345678901234567890';
my $big_vs_regular = $big ~~ 123456789012345678901234567890;

use version;

type VersionValue = version->parse('v1.2.3') | version->parse('v2.0.0');
my VersionValue $ver = version->parse('v1.2.3');

my $version_match = $ver ~~ version->parse('v1.2.3');
my $version_string = $ver ~~ 'v1.2.3';
my $version_number = $ver ~~ 1.2.3;

use DateTime;

type DateValue = DateTime->new(year => 2023, month => 1, day => 1);
my DateValue $date = DateTime->new(year => 2023, month => 1, day => 1);

my $date_match = $date ~~ DateTime->new(year => 2023, month => 1, day => 1);
my $date_string = $date ~~ '2023-01-01T00:00:00';
my $date_epoch = $date ~~ 1672531200;

use Path::Tiny;

type PathValue = Path::Tiny->new('/tmp') | Path::Tiny->new('/var');
my PathValue $path = Path::Tiny->new('/tmp');

my $path_match = $path ~~ Path::Tiny->new('/tmp');
my $path_string = $path ~~ '/tmp';
my $path_exists = $path ~~ -d;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none - but PSC should document expected behavior for all edge cases)
```
