---
category: typed-perl
subcategory: methods-fields
tags: [deprecated-syntax, returns-syntax, method-definitions]
type_check: true
---

# Deprecated Returns Syntax

## Method with Returns Syntax

<!-- should_warn: true -->
<!-- expected_warning: warning[TSP010] -->
<!-- expected_suggestion: Use 'method Type name()' instead of 'method name() returns Type' -->

```perl
method calculate() returns Int {
    return 42;
}
```

### Expected AST

```
AST {
  Path:
  Source length: 48 characters
  Type Annotations:
    MethodReturnAnnotation: calculate :: Int at 1:8
  Warnings:
    DeprecationWarning: 'returns' syntax is deprecated, use 'method Type name()' instead at 1:20
  Root: source_file
  Tree Structure:
  source_file
    method_decl
      block_stmt
        token
        expression_stmt
          literal
        token
        token
}
```

### Expected Compilation Outcomes

#### Clean Perl Output

```perl
use v5.36;
method calculate() {
    return 42;
}
```

#### Typed Perl Output

```perl
method Int calculate() {
    return 42;
}
```

#### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Method with Parameters and Returns Syntax

<!-- should_warn: true -->
<!-- expected_warning: warning[TSP010] -->
<!-- expected_suggestion: Use 'method Type name(ParamType $param)' instead of 'method name(ParamType $param) returns Type' -->

```perl
method greet(Str $name) returns Str {
    return "Hello, $name!";
}
```

### Expected AST

```
AST {
  Path:
  Source length: 70 characters
  Type Annotations:
    MethodReturnAnnotation: greet :: Str at 1:8
    MethodParamAnnotation: $name :: Str at 1:14
  Warnings:
    DeprecationWarning: 'returns' syntax is deprecated, use 'method Type name()' instead at 1:25
  Root: source_file
  Tree Structure:
  source_file
    method_decl
      block_stmt
        token
        expression_stmt
          literal
        token
        token
}
```

### Expected Compilation Outcomes

#### Clean Perl Output

```perl
use v5.36;
method greet($name) {
    return "Hello, $name!";
}
```

#### Typed Perl Output

```perl
method Str greet(Str $name) {
    return "Hello, $name!";
}
```

#### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Method with Complex Return Type and Returns Syntax

<!-- should_warn: true -->
<!-- expected_warning: warning[TSP010] -->
<!-- expected_suggestion: Use 'method Type name()' instead of 'method name() returns Type' -->

```perl
method get_data() returns ArrayRef[HashRef[Str, Int]] {
    return [];
}
```

### Expected AST

```
AST {
  Path:
  Source length: 75 characters
  Type Annotations:
    MethodReturnAnnotation: get_data :: ArrayRef[HashRef[Str, Int]] at 1:8
  Warnings:
    DeprecationWarning: 'returns' syntax is deprecated, use 'method Type name()' instead at 1:19
  Root: source_file
  Tree Structure:
  source_file
    method_decl
      block_stmt
        token
        expression_stmt
          literal
        token
        token
}
```

### Expected Compilation Outcomes

#### Clean Perl Output

```perl
use v5.36;
method get_data() {
    return [];
}
```

#### Typed Perl Output

```perl
method ArrayRef[HashRef[Str, Int]] get_data() {
    return [];
}
```

#### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
