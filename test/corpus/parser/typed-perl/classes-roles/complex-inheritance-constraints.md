---
category: typed-perl
subcategory: classes-roles
tags:
    - complex-inheritance
    - multiple-constraints
    - method-constraints
    - intersection-types
type_check: true
---

# Complex Inheritance Constraints
<!-- should_error: true -->
<!-- expected_error: error[TSP001] -->

Complex inheritance with multiple type constraints and method constraints

```perl
class ProcessingQueue<T> : BaseQueue<T>
    where T: Serializable&Processable {

    field CodeRef[T, ProcessResult] $processor;
    field ArrayRef[T] $pending = [];
    field HashRef[Str, T] $processing = {};
    field Int $max_concurrent = 5;

    method process_all() returns ArrayRef[ProcessResult] {
        my @results;
        while (@{$pending} && keys %{$processing} < $max_concurrent) {
            my $item = shift @{$pending};
            my $id = $item->get_id();
            $processing->{$id} = $item;

            my $result = $processor->($item);
            delete $processing->{$id};
            push @results, $result;
        }
        return \@results;
    }

    method enqueue(T $item) returns Void where $item->can('get_id') {
        push @{$pending}, $item;
    }

    method get_queue_status() returns QueueStatus {
        return QueueStatus->new(
            pending => scalar @{$pending},
            processing => scalar keys %{$processing},
            max_concurrent => $max_concurrent
        );
    }
}
```

# Expected Parse Error

This test case is expected to fail parsing due to unsupported syntax:
- Intersection types: `Serializable&Processable`
- Method constraints: `where $item->can('get_id')`

The parser correctly rejects this syntax with parse errors.


# Expected Compilation Outcomes

## Clean Perl Output

```perl
# Compilation failed: Error: failed to parse file /tmp/tmp6ca0doyi.pl: SYS-007: error[TSP001]: parse error (3 ERROR nodes detected)
  --> :2:2
   |
 2 |     where T: Serializable&Processable {
   |  ^ unexpected token: ''

  --> :2:38
   |
 2 |     where T: Serializable&Processable {
   |                                      ^ unexpected token: ''

  --> :3:28
   |
 3 |
   |                            ^^^^^^^^^^^ unexpected token: ''

note: This indicates Perl syntax that is not yet supported by the tree-sitter grammar.
      Please add test cases for this syntax to improve parser coverage.
 (System Error)
2025-06-25T05:32:39Z [ERROR] [psc] Error: failed to parse file /tmp/tmp6ca0doyi.pl: SYS-007: error[TSP001]: parse error (3 ERROR nodes detected)
  --> :2:2
   |
 2 |     where T: Serializable&Processable {
   |  ^ unexpected token: ''

  --> :2:38
   |
 2 |     where T: Serializable&Processable {
   |                                      ^ unexpected token: ''

  --> :3:28
   |
 3 |
   |                            ^^^^^^^^^^^ unexpected token: ''

note: This indicates Perl syntax that is not yet supported by the tree-sitter grammar.
      Please add test cases for this syntax to improve parser coverage.
 (System Error)
```

## Typed Perl Output

```perl
# Compilation failed: Error: failed to parse file /tmp/tmp6ca0doyi.pl: SYS-007: error[TSP001]: parse error (3 ERROR nodes detected)
  --> :2:2
   |
 2 |     where T: Serializable&Processable {
   |  ^ unexpected token: ''

  --> :2:38
   |
 2 |     where T: Serializable&Processable {
   |                                      ^ unexpected token: ''

  --> :3:28
   |
 3 |
   |                            ^^^^^^^^^^^ unexpected token: ''

note: This indicates Perl syntax that is not yet supported by the tree-sitter grammar.
      Please add test cases for this syntax to improve parser coverage.
 (System Error)
2025-06-25T05:32:39Z [ERROR] [psc] Error: failed to parse file /tmp/tmp6ca0doyi.pl: SYS-007: error[TSP001]: parse error (3 ERROR nodes detected)
  --> :2:2
   |
 2 |     where T: Serializable&Processable {
   |  ^ unexpected token: ''

  --> :2:38
   |
 2 |     where T: Serializable&Processable {
   |                                      ^ unexpected token: ''

  --> :3:28
   |
 3 |
   |                            ^^^^^^^^^^^ unexpected token: ''

note: This indicates Perl syntax that is not yet supported by the tree-sitter grammar.
      Please add test cases for this syntax to improve parser coverage.
 (System Error)
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
Parse error: unsupported syntax for intersection types and method constraints
```
