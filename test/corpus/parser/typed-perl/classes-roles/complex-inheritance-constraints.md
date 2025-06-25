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
<!-- expected_error: UnknownTypeError -->

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

# Expected Type Errors

```
Parse error: unsupported syntax for intersection types and method constraints
```
