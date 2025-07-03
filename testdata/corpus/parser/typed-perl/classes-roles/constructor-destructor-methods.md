---
category: typed-perl
subcategory: classes-roles
tags:
    - constructor
    - destructor
    - BUILD-method
    - DESTROY-method
    - lifecycle-management
type_check: true
---

# Constructor Destructor Methods

Class with constructor, destructor, and lifecycle methods

```perl
class Resource {
    field Str $name;
    field FileHandle $handle;
    field Bool $is_open = 0;

    method Void BUILD(Str $name, Optional[Str] $mode = 'r') {
        $self->{name} = $name;
        $self->{handle} = IO::File->new($name, $mode);
        $self->{is_open} = defined $self->{handle};
    }

    method Resource new(Str $name, Optional[Str] $mode = 'r') {
        my $self = bless {}, __PACKAGE__;
        $self->BUILD($name, $mode);
        return $self;
    }

    method Void DESTROY() {
        $self->close() if $is_open;
    }

    method Bool close() {
        return 0 unless $is_open;
        my $result = $handle->close();
        $is_open = 0;
        return $result;
    }

    method Optional[Str] read(Int $bytes) {
        return undef unless $is_open;
        my $data;
        my $read_bytes = $handle->read($data, $bytes);
        return defined $read_bytes ? $data : undef;
    }
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 913 characters
  Type Annotations:
    MethodReturnAnnotation: BUILD :: Void at 6:12
    MethodReturnAnnotation: new :: Resource at 12:12
    MethodReturnAnnotation: DESTROY :: Void at 18:12
    MethodReturnAnnotation: close :: Bool at 22:12
    MethodReturnAnnotation: read :: Optional[Str] at 29:12
    VarAnnotation: Resource :: class at 1:1
    VarAnnotation: $name :: Str at 2:5
    VarAnnotation: $handle :: FileHandle at 3:5
    VarAnnotation: $is_open :: Bool at 4:5
    MethodReturnAnnotation: BUILD :: Void at 6:12
    MethodReturnAnnotation: new :: Resource at 12:12
    MethodReturnAnnotation: DESTROY :: Void at 18:12
    MethodReturnAnnotation: close :: Bool at 22:12
    MethodReturnAnnotation: read :: Optional[Str] at 29:12
    MethodParamAnnotation: $name :: Str at 6:1
    MethodParamAnnotation: 'r' :: Optional[Str] at 6:1
    MethodParamAnnotation: $name :: Str at 12:1
    MethodParamAnnotation: 'r' :: Optional[Str] at 12:1
    MethodParamAnnotation: $bytes :: Int at 29:1
  Root: source_file
  Tree Structure:
  source_file
    class_decl
      field_decl
        variable
      field_decl
        variable
      method_decl
        block_stmt
          token
          expression_stmt
            literal
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
          token
      method_decl
        block_stmt
          token
          expression_stmt
            literal
          token
          expression_stmt
            literal
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
          token
          expression_stmt
            literal
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
  Source length: 913 characters
  Type Annotations:
    MethodReturnAnnotation: BUILD :: Void at 6:12
    MethodReturnAnnotation: new :: Resource at 12:12
    MethodReturnAnnotation: DESTROY :: Void at 18:12
    MethodReturnAnnotation: close :: Bool at 22:12
    MethodReturnAnnotation: read :: Optional[Str] at 29:12
    VarAnnotation: Resource :: class at 1:1
    VarAnnotation: $name :: Str at 2:5
    VarAnnotation: $handle :: FileHandle at 3:5
    VarAnnotation: $is_open :: Bool at 4:5
    MethodReturnAnnotation: BUILD :: Void at 6:12
    MethodReturnAnnotation: new :: Resource at 12:12
    MethodReturnAnnotation: DESTROY :: Void at 18:12
    MethodReturnAnnotation: close :: Bool at 22:12
    MethodReturnAnnotation: read :: Optional[Str] at 29:12
    MethodParamAnnotation: $name :: Str at 6:1
    MethodParamAnnotation: 'r' :: Optional[Str] at 6:1
    MethodParamAnnotation: $name :: Str at 12:1
    MethodParamAnnotation: 'r' :: Optional[Str] at 12:1
    MethodParamAnnotation: $bytes :: Int at 29:1
  Root: source_file
  Tree Structure:
  source_file
    class_decl
      field_decl
        variable
      field_decl
        variable
      method_decl
        block_stmt
          token
          expression_stmt
            literal
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
          token
      method_decl
        block_stmt
          token
          expression_stmt
            literal
          token
          expression_stmt
            literal
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
          token
          expression_stmt
            literal
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
class Resource {
    field $name;
    field $handle;
    field $is_open = 0;

    method BUILD($name, $mode = 'r') {
        $self->{name} = $name;
        $self->{handle} = IO::File->new($name, $mode);
        $self->{is_open} = defined $self->{handle};
    }

    method new($name, $mode = 'r') {
        my $self = bless {}, __PACKAGE__;
        $self->BUILD($name, $mode);
        return $self;
    }

    method DESTROY() {
        $self->close() if $is_open;
    }

    method close() {
        return 0 unless $is_open;
        my $result = $handle->close();
        $is_open = 0;
        return $result;
    }

    method read($bytes) {
        return undef unless $is_open;
        my $data;
        my $read_bytes = $handle->read($data, $bytes);
        return defined $read_bytes ? $data : undef;
    }
}
```

## Typed Perl Output

```perl
class Resource {
    field Str $name;
    field FileHandle $handle;
    field Bool $is_open = 0;

    method Void BUILD(Str $name, Optional[Str] $mode = 'r') {
        $self->{name} = $name;
        $self->{handle} = IO::File->new($name, $mode);
        $self->{is_open} = defined $self->{handle};
    }

    method Resource new(Str $name, Optional[Str] $mode = 'r') {
        my $self = bless {}, __PACKAGE__;
        $self->BUILD($name, $mode);
        return $self;
    }

    method Void DESTROY() {
        $self->close() if $is_open;
    }

    method Bool close() {
        return 0 unless $is_open;
        my $result = $handle->close();
        $is_open = 0;
        return $result;
    }

    method Optional[Str] read(Int $bytes) {
        return undef unless $is_open;
        my $data;
        my $read_bytes = $handle->read($data, $bytes);
        return defined $read_bytes ? $data : undef;
    }
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
