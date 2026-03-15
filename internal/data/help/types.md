# PVM Type Annotations

PVM supports Perl code with type annotations as defined by the typed Perl grammar.

## Basic Type Syntax

Type annotations follow the pattern: `my Type $variable = value;`

```
my Int $count = 42;
my Str $name = "example";
```

## Parsing Type-Annotated Perl

Use `psc parse` to parse and inspect Perl files with type annotations:

```
psc parse script.pl
```

Use `psc analyze` to analyze dependencies in Perl files:

```
psc analyze script.pl
```
