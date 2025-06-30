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

    method BUILD(Str $name, Optional[Str] $mode = 'r') returns Void {
        $self->{name} = $name;
        $self->{handle} = IO::File->new($name, $mode);
        $self->{is_open} = defined $self->{handle};
    }

    method new(Str $name, Optional[Str] $mode = 'r') returns Resource {
        my $self = bless {}, __PACKAGE__;
        $self->BUILD($name, $mode);
        return $self;
    }

    method DESTROY() returns Void {
        $self->close() if $is_open;
    }

    method close() returns Bool {
        return 0 unless $is_open;
        my $result = $handle->close();
        $is_open = 0;
        return $result;
    }

    method read(Int $bytes) returns Optional[Str] {
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
  Source length: 953 characters
  Type Annotations:
    MethodReturnAnnotation: BUILD :: Void at 6:64
    MethodReturnAnnotation: new :: Resource at 12:62
    MethodReturnAnnotation: DESTROY :: Void at 18:30
    MethodReturnAnnotation: close :: Bool at 22:28
    MethodReturnAnnotation: read :: Optional[Str] at 29:37
    VarAnnotation: Resource :: class at 1:1
    VarAnnotation: $name :: Str at 2:5
    VarAnnotation: $handle :: FileHandle at 3:5
    VarAnnotation: $is_open :: Bool at 4:5
    MethodReturnAnnotation: BUILD :: Void at 6:64
    MethodReturnAnnotation: new :: Resource at 12:62
    MethodReturnAnnotation: DESTROY :: Void at 18:30
    MethodReturnAnnotation: close :: Bool at 22:28
    MethodReturnAnnotation: read :: Optional[Str] at 29:37
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
  Source length: 953 characters
  Type Annotations:
    MethodReturnAnnotation: BUILD :: Void at 6:64
    MethodReturnAnnotation: new :: Resource at 12:62
    MethodReturnAnnotation: DESTROY :: Void at 18:30
    MethodReturnAnnotation: close :: Bool at 22:28
    MethodReturnAnnotation: read :: Optional[Str] at 29:37
    VarAnnotation: Resource :: class at 1:1
    VarAnnotation: $name :: Str at 2:5
    VarAnnotation: $handle :: FileHandle at 3:5
    VarAnnotation: $is_open :: Bool at 4:5
    MethodReturnAnnotation: BUILD :: Void at 6:64
    MethodReturnAnnotation: new :: Resource at 12:62
    MethodReturnAnnotation: DESTROY :: Void at 18:30
    MethodReturnAnnotation: close :: Bool at 22:28
    MethodReturnAnnotation: read :: Optional[Str] at 29:37
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
{ $self->{name} = $name; $self->{handle} = $mode); $self->{is_open} = defined $self->{handle} }{ my $self = bless {}, __PACKAGE__; $self->BUILD($name, $mode); return $self; }{ $self->close() if $is_open; }{ return 0 unless $is_open; my $result = $handle->close(); $is_open = 0; return $result; }{ return undef unless $is_open; my $data; my $read_bytes = $handle->read($data, $bytes); return defined $read_bytes ? $data : undef; }
```

## Typed Perl Output

```perl
class Resource {
    field Str $name;
    field FileHandle $handle;
    field Bool $is_open = 0;

    method BUILD(Str $name, Optional[Str] $mode = 'r') returns Void {
        $self->{name} = $name;
        $self->{handle} = IO::File->new($name, $mode);
        $self->{is_open} = defined $self->{handle};
    }

    method new(Str $name, Optional[Str] $mode = 'r') returns Resource {
        my $self = bless {}, __PACKAGE__;
        $self->BUILD($name, $mode);
        return $self;
    }

    method DESTROY() returns Void {
        $self->close() if $is_open;
    }

    method close() returns Bool {
        return 0 unless $is_open;
        my $result = $handle->close();
        $is_open = 0;
        return $result;
    }

    method read(Int $bytes) returns Optional[Str] {
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
