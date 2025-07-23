---
category: typed-perl
subcategory: classes-roles
tags:
    - comprehensive
    - multiple-inheritance
    - multiple-roles
    - generic-classes
    - BUILD-method
    - DESTROY-method
    - complex-inheritance
    - interface-implementation
    - lifecycle-management
type_check: true
should_error: true
---

# All Features Combined

Complex example combining all class and role features

```perl
# Type definitions
type UserId = Int where { $_ > 0 };
type Result<T, E> = Success<T> | Failure<E>;

# Generic role with constraints
role Repository<T, K> where T: Serializable, K: Hashable {
    method Optional[T] find(K $key);
    method Result<K, SaveError> save(T $entity);
    method Result<Bool, DeleteError> delete(K $key);
}

# Role with provided implementations
role Auditable {
    field Optional[DateTime] $created_at;
    field Optional[DateTime] $updated_at;

    method Void touch() {
        $updated_at = DateTime->now();
    }

    method Void mark_created() {
        $created_at = DateTime->now();
        $updated_at = $created_at;
    }
}

# Complex class with all features
class UserRepository<T> : BaseRepository<T, UserId>
    does Repository<T, UserId>, Auditable, Cacheable<UserId>
    where T: User&Serializable {

    field HashRef[UserId, T] $cache = {};
    field CodeRef[UserId, Optional[T]] $loader;
    field Int $cache_size = 1000;
    field Str $table_name;

    method Void BUILD(CodeRef[UserId, Optional[T]] $loader, :$table_name as Str = 'users', :$cache_size as Int = 1000) where $cache_size > 0 {
        $self->{loader} = $loader;
        $self->{table_name} = $table_name;
        $self->{cache_size} = $cache_size;
        $self->mark_created();
    }

    method Optional[T] find(UserId $id) {
        # Check cache first
        return $cache->{$id} if exists $cache->{$id};

        # Load from source
        my $user = $loader->($id);
        return undef unless defined $user;

        # Cache if room
        if (keys %{$cache} < $cache_size) {
            $cache->{$id} = $user;
        }

        return $user;
    }

    method Result<UserId, SaveError> save(T $user) {
        # Validate user
        return Failure->new(SaveError->new('Invalid user'))
            unless $user->is_valid();

        # Save to storage
        my $id = $user->get_id();
        # ... actual save logic ...

        # Update cache
        $cache->{$id} = $user;
        $self->touch();

        return Success->new($id);
    }

    method Result<Bool, DeleteError> delete(UserId $id) {
        # Remove from cache
        delete $cache->{$id};

        # Delete from storage
        # ... actual delete logic ...

        $self->touch();
        return Success->new(1);
    }

    method UserId cache_key() {
        return UserId->new($table_name . '_cache');
    }

    method Void clear_cache() {
        %{$cache} = ();
    }

    method HashRef[Str, Int] get_cache_stats() {
        return {
            size => scalar keys %{$cache},
            max_size => $cache_size,
            hit_rate => $self->calculate_hit_rate()
        };
    }
}
```

# Expected Parse Error

This comprehensive test case is expected to fail parsing due to multiple unsupported syntax features:
- Type definitions with constraints: `type UserId = Int where { $_ > 0 }`
- Union types: `Success<T> | Failure<E>`
- Multiple type constraints: `T: Serializable, K: Hashable`
- Intersection types: `T: User&Serializable`
- Named parameters: `:$table_name as Str`
- Method constraints: `where $cache_size > 0`

The parser correctly rejects this advanced syntax.


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
# Compilation failed: Error: failed to parse file /tmp/tmpgabtjb_y.pl: SYS-007: error[TSP001]: parse error (4 ERROR nodes detected)
  --> :2:2
   |
 2 | type UserId = Int where { $_ > 0 };
   |  ^ unexpected token: ''

  --> :7:17
   |
 7 |     method Optional[T] find(K $key);
   |                 ^^^^^^^^^^^^^^ unexpected token: ''

  --> :7:45
   |
 7 |     method Optional[T] find(K $key);
   |                                             ^ unexpected token: ''

  --> :65:59
   |
65 |         # Validate user
   |                                                           ^ unexpected token: ''

note: This indicates Perl syntax that is not yet supported by the tree-sitter grammar.
      Please add test cases for this syntax to improve parser coverage.
 (System Error)
2025-06-25T05:32:39Z [ERROR] [psc] Error: failed to parse file /tmp/tmpgabtjb_y.pl: SYS-007: error[TSP001]: parse error (4 ERROR nodes detected)
  --> :2:2
   |
 2 | type UserId = Int where { $_ > 0 };
   |  ^ unexpected token: ''

  --> :7:17
   |
 7 |     method Optional[T] find(K $key);
   |                 ^^^^^^^^^^^^^^ unexpected token: ''

  --> :7:45
   |
 7 |     method Optional[T] find(K $key);
   |                                             ^ unexpected token: ''

  --> :65:59
   |
65 |         # Validate user
   |                                                           ^ unexpected token: ''

note: This indicates Perl syntax that is not yet supported by the tree-sitter grammar.
      Please add test cases for this syntax to improve parser coverage.
 (System Error)
```

## Typed Perl Output

```perl
# Compilation failed: Error: failed to parse file /tmp/tmpgabtjb_y.pl: SYS-007: error[TSP001]: parse error (4 ERROR nodes detected)
  --> :2:2
   |
 2 | type UserId = Int where { $_ > 0 };
   |  ^ unexpected token: ''

  --> :7:17
   |
 7 |     method Optional[T] find(K $key);
   |                 ^^^^^^^^^^^^^^ unexpected token: ''

  --> :7:45
   |
 7 |     method Optional[T] find(K $key);
   |                                             ^ unexpected token: ''

  --> :65:59
   |
65 |         # Validate user
   |                                                           ^ unexpected token: ''

note: This indicates Perl syntax that is not yet supported by the tree-sitter grammar.
      Please add test cases for this syntax to improve parser coverage.
 (System Error)
2025-06-25T05:32:39Z [ERROR] [psc] Error: failed to parse file /tmp/tmpgabtjb_y.pl: SYS-007: error[TSP001]: parse error (4 ERROR nodes detected)
  --> :2:2
   |
 2 | type UserId = Int where { $_ > 0 };
   |  ^ unexpected token: ''

  --> :7:17
   |
 7 |     method Optional[T] find(K $key);
   |                 ^^^^^^^^^^^^^^ unexpected token: ''

  --> :7:45
   |
 7 |     method Optional[T] find(K $key);
   |                                             ^ unexpected token: ''

  --> :65:59
   |
65 |         # Validate user
   |                                                           ^ unexpected token: ''

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
Parse error: unsupported syntax for advanced type system features
```
