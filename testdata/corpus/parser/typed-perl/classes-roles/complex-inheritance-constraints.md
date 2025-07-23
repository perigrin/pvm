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

Complex inheritance with multiple type constraints and method constraints

```perl
class ProcessingQueue<T> : BaseQueue<T>
    where T: Serializable&Processable {

    field CodeRef[T, ProcessResult] $processor;
    field ArrayRef[T] $pending = [];
    field HashRef[Str, T] $processing = {};
    field Int $max_concurrent = 5;

    method ArrayRef[ProcessResult] process_all() {
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

    method Void enqueue(T $item) where $item->can('get_id') {
        push @{$pending}, $item;
    }

    method QueueStatus get_queue_status() {
        return QueueStatus->new(
            pending => scalar @{$pending},
            processing => scalar keys %{$processing},
            max_concurrent => $max_concurrent
        );
    }
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
class ProcessingQueue<T> : BaseQueue<T>
    where T:  {
    field $processor;
    field $pending = [];
    field $processing = {};
    field $max_concurrent = 5;
    method process_all() {
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
    method enqueue($item) where $item->can('get_id') {
        push @{$pending}, $item;
    }
    method QueueStatus () {
        return QueueStatus->new(
            pending => scalar @{$pending},
            processing => scalar keys %{$processing},
            max_concurrent => $max_concurrent
        );
    }
}
```

## Typed Perl Output

```perl
class ProcessingQueue<T> : BaseQueue<T>
    where T: Serializable&Processable {

    field CodeRef[T, ProcessResult] $processor;
    field ArrayRef[T] $pending = [];
    field HashRef[Str, T] $processing = {};
    field Int $max_concurrent = 5;

    method ArrayRef[ProcessResult] process_all() {
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

    method Void enqueue(T $item) where $item->can('get_id') {
        push @{$pending}, $item;
    }

    method QueueStatus get_queue_status() {
        return QueueStatus->new(
            pending => scalar @{$pending},
            processing => scalar keys %{$processing},
            max_concurrent => $max_concurrent
        );
    }
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
