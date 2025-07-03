---
category: typed-perl
subcategory: methods-fields
tags:
    - complex-methods
    - method-signatures
    - optional-parameters
    - parameterized-types
    - complex-return-types
type_check: true
---

# Complex Method Signatures

Complex method signatures with parameterized types, optional parameters, and multiple parameter types

```perl
method process(ArrayRef[Str] $data, Bool $validate = 1) returns ArrayRef[Str] {
    my @result = @{$data};
    return \@result;
}

method transform(HashRef[Int] $input, CodeRef $callback) returns HashRef[Int] {
    my %result;
    for my $key (keys %{$input}) {
        $result{$key} = $callback->($input->{$key});
    }
    return \%result;
}

method complex_method(
    ArrayRef[HashRef[Int]] $data,
    Optional[CodeRef] $processor,
    Slurpy[Str] @extra_args
) returns Bool {
    return 1;
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 496 characters
  Type Annotations:
    MethodReturnAnnotation: process :: ArrayRef[Str] at 1:65
    MethodReturnAnnotation: transform :: HashRef[Int] at 6:66
    MethodReturnAnnotation: complex_method :: Bool at 18:11
    MethodParamAnnotation: $data :: ArrayRef[Str] at 1:1
    MethodParamAnnotation: 1 :: Bool at 1:1
    MethodParamAnnotation: $input :: HashRef[Int] at 6:1
    MethodParamAnnotation: $callback :: CodeRef at 6:1
  Root: source_file
  Tree Structure:
  source_file
    method_decl
      block_stmt
        token
        expression_stmt
          literal
        token
        expression_stmt
          literal
        token
        token
    method_decl
      block_stmt
        token
        expression_stmt
          literal
        token
        expression_stmt
          literal
        expression_stmt
          literal
        token
        token
    method_decl
      block_stmt
        token
        expression_stmt
          literal
        token
        token
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 496 characters
  Type Annotations:
    MethodReturnAnnotation: process :: ArrayRef[Str] at 1:65
    MethodReturnAnnotation: transform :: HashRef[Int] at 6:66
    MethodReturnAnnotation: complex_method :: Bool at 18:11
    MethodParamAnnotation: $data :: ArrayRef[Str] at 1:1
    MethodParamAnnotation: 1 :: Bool at 1:1
    MethodParamAnnotation: $input :: HashRef[Int] at 6:1
    MethodParamAnnotation: $callback :: CodeRef at 6:1
  Root: source_file
  Tree Structure:
  source_file
    method_decl
      block_stmt
        token
        expression_stmt
          literal
        token
        expression_stmt
          literal
        token
        token
    method_decl
      block_stmt
        token
        expression_stmt
          literal
        token
        expression_stmt
          literal
        expression_stmt
          literal
        token
        token
    method_decl
      block_stmt
        token
        expression_stmt
          literal
        token
        token
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
method process($data, $validate = 1) {
    my @result = @{$data};
    return \@result;
}

method transform($input, $callback) {
    my %result;
    for my $key (keys %{$input}) {
        $result{$key} = $callback->($input->{$key});
    }
    return \%result;
}

method complex_method(
    $data,
    $processor,
    @extra_args
) {
    return 1;
}
```

## Typed Perl Output

```perl
method process(ArrayRef[Str] $data, Bool $validate = 1) returns ArrayRef[Str] {
    my @result = @{$data};
    return \@result;
}

method transform(HashRef[Int] $input, CodeRef $callback) returns HashRef[Int] {
    my %result;
    for my $key (keys %{$input}) {
        $result{$key} = $callback->($input->{$key});
    }
    return \%result;
}

method complex_method(
    ArrayRef[HashRef[Int]] $data,
    Optional[CodeRef] $processor,
    Slurpy[Str] @extra_args
) returns Bool {
    return 1;
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
