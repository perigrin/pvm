---
category: typed-perl
subcategory: smartmatch
tags:
    - smartmatch
    - type-constraints
    - union-types
    - operator-behavior
type_check: true
should_error: true
expected_error: "parse error"
---

# Smartmatch with Type-Constrained Values

Tests smartmatch (`~~`) operator behavior with type-constrained values to ensure predictable
and safe matching behavior when types are known at compile time.

```perl
# Basic smartmatch with string literal unions
type Status = 'active' | 'inactive' | 'pending';
my Status $status = 'active';

# Type-constrained vs literal comparison
my $is_active = $status ~~ 'active';           # Should work predictably
my $is_valid = $status ~~ ['active', 'busy'];  # Array membership with mixed types

# HTTP status code matching
type HttpCode = '200' | '404' | '500';
my HttpCode $code = '200';

# String equality with type constraints
my $success = $code ~~ '200';              # Type-constrained vs string literal
my $error = $code ~~ ['404', '500'];       # Array membership testing
my $pattern_ok = $code ~~ /^2\d\d$/;       # Regex matching with type constraint

# Numeric union smartmatch
type Priority = 1 | 2 | 3 | 4 | 5;
my Priority $pri = 3;

my $is_high = $pri ~~ [1, 2];              # Numeric array membership
my $is_medium = $pri ~~ 3;                 # Direct numeric comparison
my $is_range = $pri ~~ qr/[3-5]/;          # Regex pattern with numeric

# Mixed string/numeric union
type MixedId = 42 | '42' | 'none';
my MixedId $id1 = 42;
my MixedId $id2 = '42';

# Critical behavior: numeric vs string with type constraints
my $numeric_match = $id1 ~~ 42;            # Numeric vs numeric
my $string_match = $id2 ~~ '42';           # String vs string
my $cross_match = $id1 ~~ '42';            # Should this work? PSC should be predictable

# Boolean-like values
type BoolLike = 1 | 0 | 'true' | 'false';
my BoolLike $flag = 1;

my $is_truthy = $flag ~~ 1;                # Numeric true
my $is_string_true = $flag ~~ 'true';      # String true - should not match numeric 1
my $truth_array = $flag ~~ [1, 'true'];    # Mixed array membership

# Complex array membership testing
type Color = 'red' | 'green' | 'blue';
my Color $color = 'red';

my $is_warm = $color ~~ ['red', 'orange', 'yellow'];    # Partial match array
my $is_primary = $color ~~ ['red', 'blue', 'green'];    # Complete union match

# Regex pattern matching with unions
type Protocol = 'http' | 'https' | 'ftp' | 'sftp';
my Protocol $proto = 'https';

my $is_secure = $proto ~~ qr/s$/;          # Ends with 's' - https, sftp
my $is_http = $proto ~~ qr/^http/;         # Starts with 'http' - http, https
my $has_p = $proto ~~ qr/p/;               # Contains 'p' - all except ftp... wait, ftp has p

# Case sensitivity with type constraints
type CaseId = 'User' | 'Admin' | 'Guest';
my CaseId $role = 'User';

my $exact_match = $role ~~ 'User';         # Exact case match
my $wrong_case = $role ~~ 'user';          # Wrong case - should not match
my $case_insensitive = $role ~~ qr/user/i; # Regex with case insensitive flag
```

# Expected AST

## Before Type Inference

### Text Format

```
AST {
  Path:
  Source length: 2841 characters
  Type Annotations:
    TypeAnnotation: Status = 'active' | 'inactive' | 'pending' at 1:1
    VarAnnotation: $status :: Status at 2:1
    TypeAnnotation: HttpCode = '200' | '404' | '500' at 8:1
    VarAnnotation: $code :: HttpCode at 9:1
    TypeAnnotation: Priority = 1 | 2 | 3 | 4 | 5 at 16:1
    VarAnnotation: $pri :: Priority at 17:1
    TypeAnnotation: MixedId = 42 | '42' | 'none' at 23:1
    VarAnnotation: $id1 :: MixedId at 24:1
    VarAnnotation: $id2 :: MixedId at 25:1
    TypeAnnotation: BoolLike = 1 | 0 | 'true' | 'false' at 32:1
    VarAnnotation: $flag :: BoolLike at 33:1
    TypeAnnotation: Color = 'red' | 'green' | 'blue' at 39:1
    VarAnnotation: $color :: Color at 40:1
    TypeAnnotation: Protocol = 'http' | 'https' | 'ftp' | 'sftp' at 45:1
    VarAnnotation: $proto :: Protocol at 46:1
    TypeAnnotation: CaseId = 'User' | 'Admin' | 'Guest' at 52:1
    VarAnnotation: $role :: CaseId at 53:1
  Root: source_file
  Tree Structure:
  source_file
    type_declaration
      type_name(Status)
      union_type
        literal('active')
        literal('inactive')
        literal('pending')
    var_decl
      type_expression(Status)
      scalar($status)
    var_decl
      scalar($is_active)
      smartmatch_expression
        scalar($status)
        literal('active')
    var_decl
      scalar($is_valid)
      smartmatch_expression
        scalar($status)
        array_constructor
          literal('active')
          literal('busy')
    type_declaration
      type_name(HttpCode)
      union_type
        literal('200')
        literal('404')
        literal('500')
    var_decl
      type_expression(HttpCode)
      scalar($code)
    var_decl
      scalar($success)
      smartmatch_expression
        scalar($code)
        literal('200')
    var_decl
      scalar($error)
      smartmatch_expression
        scalar($code)
        array_constructor
          literal('404')
          literal('500')
    var_decl
      scalar($pattern_ok)
      smartmatch_expression
        scalar($code)
        regex_pattern(/^2\d\d$/)
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
    var_decl
      scalar($is_high)
      smartmatch_expression
        scalar($pri)
        array_constructor
          literal(1)
          literal(2)
    var_decl
      scalar($is_medium)
      smartmatch_expression
        scalar($pri)
        literal(3)
    var_decl
      scalar($is_range)
      smartmatch_expression
        scalar($pri)
        regex_pattern(qr/[3-5]/)
    type_declaration
      type_name(MixedId)
      union_type
        literal(42)
        literal('42')
        literal('none')
    var_decl
      type_expression(MixedId)
      scalar($id1)
    var_decl
      type_expression(MixedId)
      scalar($id2)
    var_decl
      scalar($numeric_match)
      smartmatch_expression
        scalar($id1)
        literal(42)
    var_decl
      scalar($string_match)
      smartmatch_expression
        scalar($id2)
        literal('42')
    var_decl
      scalar($cross_match)
      smartmatch_expression
        scalar($id1)
        literal('42')
    type_declaration
      type_name(BoolLike)
      union_type
        literal(1)
        literal(0)
        literal('true')
        literal('false')
    var_decl
      type_expression(BoolLike)
      scalar($flag)
    var_decl
      scalar($is_truthy)
      smartmatch_expression
        scalar($flag)
        literal(1)
    var_decl
      scalar($is_string_true)
      smartmatch_expression
        scalar($flag)
        literal('true')
    var_decl
      scalar($truth_array)
      smartmatch_expression
        scalar($flag)
        array_constructor
          literal(1)
          literal('true')
    type_declaration
      type_name(Color)
      union_type
        literal('red')
        literal('green')
        literal('blue')
    var_decl
      type_expression(Color)
      scalar($color)
    var_decl
      scalar($is_warm)
      smartmatch_expression
        scalar($color)
        array_constructor
          literal('red')
          literal('orange')
          literal('yellow')
    var_decl
      scalar($is_primary)
      smartmatch_expression
        scalar($color)
        array_constructor
          literal('red')
          literal('blue')
          literal('green')
    type_declaration
      type_name(Protocol)
      union_type
        literal('http')
        literal('https')
        literal('ftp')
        literal('sftp')
    var_decl
      type_expression(Protocol)
      scalar($proto)
    var_decl
      scalar($is_secure)
      smartmatch_expression
        scalar($proto)
        regex_pattern(qr/s$/)
    var_decl
      scalar($is_http)
      smartmatch_expression
        scalar($proto)
        regex_pattern(qr/^http/)
    var_decl
      scalar($has_p)
      smartmatch_expression
        scalar($proto)
        regex_pattern(qr/p/)
    type_declaration
      type_name(CaseId)
      union_type
        literal('User')
        literal('Admin')
        literal('Guest')
    var_decl
      type_expression(CaseId)
      scalar($role)
    var_decl
      scalar($exact_match)
      smartmatch_expression
        scalar($role)
        literal('User')
    var_decl
      scalar($wrong_case)
      smartmatch_expression
        scalar($role)
        literal('user')
    var_decl
      scalar($case_insensitive)
      smartmatch_expression
        scalar($role)
        regex_pattern(qr/user/i)
}
```

## After Type Inference

### Text Format

```
AST {
  Path:
  Source length: 2841 characters
  Type Annotations:
    TypeAnnotation: Status = 'active' | 'inactive' | 'pending' at 1:1
    VarAnnotation: $status :: Status at 2:1
    TypeAnnotation: HttpCode = '200' | '404' | '500' at 8:1
    VarAnnotation: $code :: HttpCode at 9:1
    TypeAnnotation: Priority = 1 | 2 | 3 | 4 | 5 at 16:1
    VarAnnotation: $pri :: Priority at 17:1
    TypeAnnotation: MixedId = 42 | '42' | 'none' at 23:1
    VarAnnotation: $id1 :: MixedId at 24:1
    VarAnnotation: $id2 :: MixedId at 25:1
    TypeAnnotation: BoolLike = 1 | 0 | 'true' | 'false' at 32:1
    VarAnnotation: $flag :: BoolLike at 33:1
    TypeAnnotation: Color = 'red' | 'green' | 'blue' at 39:1
    VarAnnotation: $color :: Color at 40:1
    TypeAnnotation: Protocol = 'http' | 'https' | 'ftp' | 'sftp' at 45:1
    VarAnnotation: $proto :: Protocol at 46:1
    TypeAnnotation: CaseId = 'User' | 'Admin' | 'Guest' at 52:1
    VarAnnotation: $role :: CaseId at 53:1
  Root: source_file
  Tree Structure:
  source_file
    type_declaration
      type_name(Status)
      union_type
        literal('active')
        literal('inactive')
        literal('pending')
    var_decl
      type_expression(Status)
      scalar($status)
    var_decl
      scalar($is_active)
      smartmatch_expression
        scalar($status)
        literal('active')
    var_decl
      scalar($is_valid)
      smartmatch_expression
        scalar($status)
        array_constructor
          literal('active')
          literal('busy')
    type_declaration
      type_name(HttpCode)
      union_type
        literal('200')
        literal('404')
        literal('500')
    var_decl
      type_expression(HttpCode)
      scalar($code)
    var_decl
      scalar($success)
      smartmatch_expression
        scalar($code)
        literal('200')
    var_decl
      scalar($error)
      smartmatch_expression
        scalar($code)
        array_constructor
          literal('404')
          literal('500')
    var_decl
      scalar($pattern_ok)
      smartmatch_expression
        scalar($code)
        regex_pattern(/^2\d\d$/)
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
    var_decl
      scalar($is_high)
      smartmatch_expression
        scalar($pri)
        array_constructor
          literal(1)
          literal(2)
    var_decl
      scalar($is_medium)
      smartmatch_expression
        scalar($pri)
        literal(3)
    var_decl
      scalar($is_range)
      smartmatch_expression
        scalar($pri)
        regex_pattern(qr/[3-5]/)
    type_declaration
      type_name(MixedId)
      union_type
        literal(42)
        literal('42')
        literal('none')
    var_decl
      type_expression(MixedId)
      scalar($id1)
    var_decl
      type_expression(MixedId)
      scalar($id2)
    var_decl
      scalar($numeric_match)
      smartmatch_expression
        scalar($id1)
        literal(42)
    var_decl
      scalar($string_match)
      smartmatch_expression
        scalar($id2)
        literal('42')
    var_decl
      scalar($cross_match)
      smartmatch_expression
        scalar($id1)
        literal('42')
    type_declaration
      type_name(BoolLike)
      union_type
        literal(1)
        literal(0)
        literal('true')
        literal('false')
    var_decl
      type_expression(BoolLike)
      scalar($flag)
    var_decl
      scalar($is_truthy)
      smartmatch_expression
        scalar($flag)
        literal(1)
    var_decl
      scalar($is_string_true)
      smartmatch_expression
        scalar($flag)
        literal('true')
    var_decl
      scalar($truth_array)
      smartmatch_expression
        scalar($flag)
        array_constructor
          literal(1)
          literal('true')
    type_declaration
      type_name(Color)
      union_type
        literal('red')
        literal('green')
        literal('blue')
    var_decl
      type_expression(Color)
      scalar($color)
    var_decl
      scalar($is_warm)
      smartmatch_expression
        scalar($color)
        array_constructor
          literal('red')
          literal('orange')
          literal('yellow')
    var_decl
      scalar($is_primary)
      smartmatch_expression
        scalar($color)
        array_constructor
          literal('red')
          literal('blue')
          literal('green')
    type_declaration
      type_name(Protocol)
      union_type
        literal('http')
        literal('https')
        literal('ftp')
        literal('sftp')
    var_decl
      type_expression(Protocol)
      scalar($proto)
    var_decl
      scalar($is_secure)
      smartmatch_expression
        scalar($proto)
        regex_pattern(qr/s$/)
    var_decl
      scalar($is_http)
      smartmatch_expression
        scalar($proto)
        regex_pattern(qr/^http/)
    var_decl
      scalar($has_p)
      smartmatch_expression
        scalar($proto)
        regex_pattern(qr/p/)
    type_declaration
      type_name(CaseId)
      union_type
        literal('User')
        literal('Admin')
        literal('Guest')
    var_decl
      type_expression(CaseId)
      scalar($role)
    var_decl
      scalar($exact_match)
      smartmatch_expression
        scalar($role)
        literal('User')
    var_decl
      scalar($wrong_case)
      smartmatch_expression
        scalar($role)
        literal('user')
    var_decl
      scalar($case_insensitive)
      smartmatch_expression
        scalar($role)
        regex_pattern(qr/user/i)
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $status = 'active';

my $is_active = $status ~~ 'active';
my $is_valid = $status ~~ ['active', 'busy'];

my $code = '200';

my $success = $code ~~ '200';
my $error = $code ~~ ['404', '500'];
my $pattern_ok = $code ~~ /^2\d\d$/;

my $pri = 3;

my $is_high = $pri ~~ [1, 2];
my $is_medium = $pri ~~ 3;
my $is_range = $pri ~~ qr/[3-5]/;

my $id1 = 42;
my $id2 = '42';

my $numeric_match = $id1 ~~ 42;
my $string_match = $id2 ~~ '42';
my $cross_match = $id1 ~~ '42';

my $flag = 1;

my $is_truthy = $flag ~~ 1;
my $is_string_true = $flag ~~ 'true';
my $truth_array = $flag ~~ [1, 'true'];

my $color = 'red';

my $is_warm = $color ~~ ['red', 'orange', 'yellow'];
my $is_primary = $color ~~ ['red', 'blue', 'green'];

my $proto = 'https';

my $is_secure = $proto ~~ qr/s$/;
my $is_http = $proto ~~ qr/^http/;
my $has_p = $proto ~~ qr/p/;

my $role = 'User';

my $exact_match = $role ~~ 'User';
my $wrong_case = $role ~~ 'user';
my $case_insensitive = $role ~~ qr/user/i;
```

## Typed Perl Output

```perl
type Status = 'active' | 'inactive' | 'pending';
my Status $status = 'active';

my $is_active = $status ~~ 'active';
my $is_valid = $status ~~ ['active', 'busy'];

type HttpCode = '200' | '404' | '500';
my HttpCode $code = '200';

my $success = $code ~~ '200';
my $error = $code ~~ ['404', '500'];
my $pattern_ok = $code ~~ /^2\d\d$/;

type Priority = 1 | 2 | 3 | 4 | 5;
my Priority $pri = 3;

my $is_high = $pri ~~ [1, 2];
my $is_medium = $pri ~~ 3;
my $is_range = $pri ~~ qr/[3-5]/;

type MixedId = 42 | '42' | 'none';
my MixedId $id1 = 42;
my MixedId $id2 = '42';

my $numeric_match = $id1 ~~ 42;
my $string_match = $id2 ~~ '42';
my $cross_match = $id1 ~~ '42';

type BoolLike = 1 | 0 | 'true' | 'false';
my BoolLike $flag = 1;

my $is_truthy = $flag ~~ 1;
my $is_string_true = $flag ~~ 'true';
my $truth_array = $flag ~~ [1, 'true'];

type Color = 'red' | 'green' | 'blue';
my Color $color = 'red';

my $is_warm = $color ~~ ['red', 'orange', 'yellow'];
my $is_primary = $color ~~ ['red', 'blue', 'green'];

type Protocol = 'http' | 'https' | 'ftp' | 'sftp';
my Protocol $proto = 'https';

my $is_secure = $proto ~~ qr/s$/;
my $is_http = $proto ~~ qr/^http/;
my $has_p = $proto ~~ qr/p/;

type CaseId = 'User' | 'Admin' | 'Guest';
my CaseId $role = 'User';

my $exact_match = $role ~~ 'User';
my $wrong_case = $role ~~ 'user';
my $case_insensitive = $role ~~ qr/user/i;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none - all smartmatch operations should be deterministic with type constraints)
```
