---
category: typed-perl
subcategory: classes-roles
tags:
    - inheritance
    - role-composition
    - class-declaration
type_check: true
---

# Class Inheritance

Class with inheritance and role composition

```perl
class Document : BaseDocument does Serializable, Cacheable {
    field Str $content;
    field DateTime $created;
    field Optional[UserRef] $author;

    method serialize() returns Str {
        return encode_json({
            content => $content,
            created => $created->iso8601,
            author => $author ? $author->id : undef
        });
    }

    method deserialize(Str $data) returns Self {
        my $decoded = decode_json($data);
        return __PACKAGE__->new(
            content => $decoded->{content},
            created => DateTime->from_epoch(epoch => $decoded->{created}),
            author => $decoded->{author} ? UserRef->new(id => $decoded->{author}) : undef
        );
    }
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 715 characters
  Type Annotations:
    MethodReturnAnnotation: serialize :: Str at 6:32
    MethodReturnAnnotation: deserialize :: Self at 14:43
    VarAnnotation: Document :: class at 1:1
    VarAnnotation: $content :: Str at 2:5
    VarAnnotation: $created :: DateTime at 3:5
    VarAnnotation: $author :: Optional[UserRef] at 4:5
    MethodReturnAnnotation: serialize :: Str at 6:32
    MethodReturnAnnotation: deserialize :: Self at 14:43
    MethodParamAnnotation: $data :: Str at 14:1
  Root: source_file
  Tree Structure:
  source_file
    class_decl
      field_decl
        variable
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
          token
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 715 characters
  Type Annotations:
    MethodReturnAnnotation: serialize :: Str at 6:32
    MethodReturnAnnotation: deserialize :: Self at 14:43
    VarAnnotation: Document :: class at 1:1
    VarAnnotation: $content :: Str at 2:5
    VarAnnotation: $created :: DateTime at 3:5
    VarAnnotation: $author :: Optional[UserRef] at 4:5
    MethodReturnAnnotation: serialize :: Str at 6:32
    MethodReturnAnnotation: deserialize :: Self at 14:43
    MethodParamAnnotation: $data :: Str at 14:1
  Root: source_file
  Tree Structure:
  source_file
    class_decl
      field_decl
        variable
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
          token
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
{ return encode_json({
            content => $content,
            created => $created->iso8601,
            author => $author ? $author->id : undef
        }); }{ my $decoded = decode_json($data); return __PACKAGE__->new(
            content => $decoded->{content},
            created => DateTime->from_epoch(epoch => $decoded->{created}),
            author => $decoded->{author} ? UserRef->new(id => $decoded->{author}) : undef
        ); }
```

## Typed Perl Output

```perl
class Document : BaseDocument does Serializable, Cacheable {
    field Str $content;
    field DateTime $created;
    field Optional[UserRef] $author;

    method serialize() returns Str {
        return encode_json({
            content => $content,
            created => $created->iso8601,
            author => $author ? $author->id : undef
        });
    }

    method deserialize(Str $data) returns Self {
        my $decoded = decode_json($data);
        return __PACKAGE__->new(
            content => $decoded->{content},
            created => DateTime->from_epoch(epoch => $decoded->{created}),
            author => $decoded->{author} ? UserRef->new(id => $decoded->{author}) : undef
        );
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
