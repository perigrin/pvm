---
category: typed-perl
subcategory: classes-roles
tags:
    - BUILD-method
    - DESTROY-method
    - access-modifiers
    - class-declaration
    - complex-inheritance
    - comprehensive
    - conflict-resolution
    - constructor
    - destructor
    - field-visibility
    - generic-class
    - generic-classes
    - generic-role
    - inheritance
    - interface-implementation
    - intersection-types
    - lifecycle-management
    - method-conflicts
    - method-constraints
    - multiple-constraints
    - multiple-inheritance
    - multiple-roles
    - parameterized-methods
    - parameterized-role-methods
    - private-methods
    - protected-methods
    - provided-methods
    - readonly-fields
    - required-methods
    - role-composition
    - role-declaration
    - role-fields
    - type-constraints
    - type-parameters
    - typed-fields
    - typed-methods
---

# Access Modifiers Visibility

Class with access modifiers and field visibility

```perl
class BankAccount {
    field private Num $balance = 0.0;
    field protected Str $account_number;
    field public Str $account_holder;
    field readonly DateTime $created_at;

    method new(Str $holder, Str $number) -> BankAccount {
        return bless {
            account_holder => $holder,
            account_number => $number,
            balance => 0.0,
            created_at => DateTime->now()
        }, __PACKAGE__;
    }

    method private validate_amount(Num $amount) -> Bool {
        return $amount > 0;
    }

    method public deposit(Num $amount) -> Bool {
        return 0 unless $self->validate_amount($amount);
        $balance += $amount;
        return 1;
    }

    method public get_balance() -> Num {
        return $balance;
    }

    method protected get_account_number() -> Str {
        return $account_number;
    }
}
```

## All Features Combined
<!-- should_error: true -->
<!-- expected_error: UnknownTypeError -->

Complex example combining all class and role features

```perl
# Type definitions
type UserId = Int where { $_ > 0 };
type Result<T, E> = Success<T> | Failure<E>;

# Generic role with constraints
role Repository<T, K> where T: Serializable, K: Hashable {
    method find(K $key) -> Optional[T];
    method save(T $entity) -> Result<K, SaveError>;
    method delete(K $key) -> Result<Bool, DeleteError>;
}

# Role with provided implementations
role Auditable {
    field Optional[DateTime] $created_at;
    field Optional[DateTime] $updated_at;

    method touch() -> Void {
        $updated_at = DateTime->now();
    }

    method mark_created() -> Void {
        $created_at = DateTime->now();
        $updated_at = $created_at;
    }
}

# Complex class with all features
class UserRepository<T> : BaseRepository<T, UserId>
    does Repository<T, UserId>, Auditable, Cacheable<UserId>
    where T: User&Serializable {

    field private HashRef[UserId, T] $cache = {};
    field protected CodeRef[UserId, Optional[T]] $loader;
    field public Int $cache_size = 1000;
    field readonly Str $table_name;

    method BUILD(
        CodeRef[UserId, Optional[T]] $loader,
        :$table_name as Str = 'users',
        :$cache_size as Int = 1000
    ) -> Void where $cache_size > 0 {
        $self->{loader} = $loader;
        $self->{table_name} = $table_name;
        $self->{cache_size} = $cache_size;
        $self->mark_created();
    }

    method find(UserId $id) -> Optional[T] {
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

    method save(T $user) -> Result<UserId, SaveError> {
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

    method delete(UserId $id) -> Result<Bool, DeleteError> {
        # Remove from cache
        delete $cache->{$id};

        # Delete from storage
        # ... actual delete logic ...

        $self->touch();
        return Success->new(1);
    }

    method cache_key() -> UserId {
        return UserId->new($table_name . '_cache');
    }

    method clear_cache() -> Void {
        %{$cache} = ();
    }

    method get_cache_stats() -> HashRef[Str, Int] {
        return {
            size => scalar keys %{$cache},
            max_size => $cache_size,
            hit_rate => $self->calculate_hit_rate()
        };
    }
}
```

## Basic Class Declarations

Basic class with typed fields and methods

```perl
class User {
    field Str $name;
    field Int $age;
    field Optional[Email] $email;

    method new(Str $name, Int $age) -> User {
        return bless {
            name => $name,
            age => $age
        }, __PACKAGE__;
    }

    method get_name() -> Str {
        return $name;
    }
}
```

## Basic Role Declarations
<!-- should_error: true -->
<!-- expected_error: UnknownTypeError -->

Basic role declarations with required and provided methods

```perl
role Serializable {
    method serialize() -> Str;
    method deserialize(Str $data) -> Self;
}

role Cacheable {
    field Optional[DateTime] $cached_at;

    method cache_key() -> Str;

    method is_stale() -> Bool {
        return 0 unless defined $cached_at;
        return time() - $cached_at->epoch > 3600;
    }

    method invalidate() -> Void {
        $cached_at = undef;
    }
}
```

## Class Inheritance

Class with inheritance and role composition

```perl
class Document : BaseDocument does Serializable, Cacheable {
    field Str $content;
    field DateTime $created;
    field Optional[UserRef] $author;

    method serialize() -> Str {
        return encode_json({
            content => $content,
            created => $created->iso8601,
            author => $author ? $author->id : undef
        });
    }

    method deserialize(Str $data) -> Self {
        my $decoded = decode_json($data);
        return __PACKAGE__->new(
            content => $decoded->{content},
            created => DateTime->from_epoch(epoch => $decoded->{created}),
            author => $decoded->{author} ? UserRef->new(id => $decoded->{author}) : undef
        );
    }
}
```

## Complex Inheritance Constraints

Complex inheritance with multiple type constraints and method constraints

```perl
class ProcessingQueue<T> : BaseQueue<T>
    where T: Serializable&Processable {

    field CodeRef[T, ProcessResult] $processor;
    field ArrayRef[T] $pending = [];
    field HashRef[Str, T] $processing = {};
    field Int $max_concurrent = 5;

    method process_all() -> ArrayRef[ProcessResult] {
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

    method enqueue(T $item) -> Void where $item->can('get_id') {
        push @{$pending}, $item;
    }

    method get_queue_status() -> QueueStatus {
        return QueueStatus->new(
            pending => scalar @{$pending},
            processing => scalar keys %{$processing},
            max_concurrent => $max_concurrent
        );
    }
}
```

## Constructor Destructor Methods

Class with constructor, destructor, and lifecycle methods

```perl
class Resource {
    field Str $name;
    field FileHandle $handle;
    field Bool $is_open = 0;

    method BUILD(Str $name, Optional[Str] $mode = 'r') -> Void {
        $self->{name} = $name;
        $self->{handle} = IO::File->new($name, $mode);
        $self->{is_open} = defined $self->{handle};
    }

    method new(Str $name, Optional[Str] $mode = 'r') -> Resource {
        my $self = bless {}, __PACKAGE__;
        $self->BUILD($name, $mode);
        return $self;
    }

    method DESTROY() -> Void {
        $self->close() if $is_open;
    }

    method close() -> Bool {
        return 0 unless $is_open;
        my $result = $handle->close();
        $is_open = 0;
        return $result;
    }

    method read(Int $bytes) -> Optional[Str] {
        return undef unless $is_open;
        my $data;
        my $read_bytes = $handle->read($data, $bytes);
        return defined $read_bytes ? $data : undef;
    }
}
```

## Generic Class Declarations

Generic class with type parameters and constraints

```perl
class Container<T> where T: Serializable {
    field ArrayRef[T] $items = [];

    method add(T $item) -> Void {
        push @{$items}, $item;
    }

    method get_all() -> ArrayRef[T] {
        return $items;
    }

    method find(CodeRef[T, Bool] $predicate) -> Optional[T] {
        for my $item (@{$items}) {
            return $item if $predicate->($item);
        }
        return undef;
    }
}
```

## Generic Role Declarations
<!-- should_error: true -->
<!-- expected_error: UnknownTypeError -->

Generic roles with type parameters and constraints

```perl
role Processable<T> where T: Defined {
    method process(T $input) -> ProcessResult;
    method validate(T $input) -> Bool;
}

role EventHandler<T> where T: Event {
    field ArrayRef[CodeRef[T, Void]] $handlers = [];

    method add_handler(CodeRef[T, Void] $handler) -> Void {
        push @{$handlers}, $handler;
    }

    method handle_event(T $event) -> Void {
        for my $handler (@{$handlers}) {
            $handler->($event);
        }
    }

    method handler_count() -> Int {
        return scalar @{$handlers};
    }
}
```

## Role Composition Conflicts
<!-- should_error: true -->
<!-- expected_error: UnknownTypeError -->

Role composition with method conflicts and resolution

```perl
role Drawable {
    method draw() -> Void;
    method get_bounds() -> Rectangle;
}

role Clickable {
    method on_click(Event $event) -> Void;
    method get_bounds() -> Rectangle;  # Conflict with Drawable
}

role Resizable {
    method resize(Int $width, Int $height) -> Void;
    method get_size() -> Size;
}

class Widget does Drawable, Clickable, Resizable {
    field Int $x = 0;
    field Int $y = 0;
    field Int $width = 100;
    field Int $height = 50;

    # Resolve conflict by implementing the conflicting method
    method get_bounds() -> Rectangle {
        return Rectangle->new($x, $y, $width, $height);
    }

    method draw() -> Void {
        # Implementation for drawing
    }

    method on_click(Event $event) -> Void {
        # Handle click event
    }

    method resize(Int $new_width, Int $new_height) -> Void {
        $width = $new_width;
        $height = $new_height;
    }

    method get_size() -> Size {
        return Size->new($width, $height);
    }
}
```
