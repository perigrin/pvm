# Typed-Perl Coding Patterns Workflow

This workflow provides comprehensive coding patterns, best practices, and idioms for writing effective typed-Perl code using the PVM ecosystem, enabling developers to leverage the full power of the gradual type system.

## Executive Summary

This document compiles proven coding patterns for typed-Perl development, covering type annotation strategies, object-oriented patterns, functional programming approaches, error handling techniques, and performance optimizations. It serves as a practical reference for writing maintainable, type-safe Perl code that takes full advantage of PSC's type checking capabilities.

## Prerequisites

- Understanding of [typed-perl-specification.md](typed-perl-specification.md)
- Familiarity with [workflow-new-development.md](workflow-new-development.md)
- Basic knowledge of Perl programming
- PSC (Perl Script Compiler) installed and configured

## Core Type Annotation Patterns

### Variable Declaration Patterns

#### Basic Type Annotations

```perl
# Scalar variables with explicit types
my Str $name = "Alice";
my Int $age = 30;
my Num $price = 29.99;
my Bool $is_active = 1;

# Array variables with element types
my ArrayRef[Str] $names = ["Alice", "Bob", "Carol"];
my ArrayRef[Int] $scores = [85, 92, 78];
my Array[HashRef] @records;  # Array of hash references

# Hash variables with value types
my HashRef[Str] $config = {
    database_url => "localhost:5432",
    api_key      => "secret_key"
};
my HashRef[Int] $counters = {
    visitors => 42,
    downloads => 156
};
```

#### Optional and Maybe Types

```perl
# Using Maybe types for optional values
my Maybe[Str] $middle_name = undef;  # Can be Str or undef
my Maybe[Int] $error_code = get_status();

# Optional hash keys
my HashRef[Str|Int] $user_data = {
    name => "Alice",      # Str
    age  => 30,          # Int
    bio  => undef,       # Optional Str
};

# Flow-sensitive refinement pattern
my Maybe[Str] $input = get_user_input();
if (defined($input)) {
    # Here $input is refined from Maybe[Str] to Str
    my Int $length = length($input);  # Safe to use string operations
    say "Input length: $length";
}
```

#### Complex Type Compositions

```perl
# Union types for multiple acceptable types
my Str|Int $id = get_identifier();
my ArrayRef[Str]|HashRef[Str] $config_data = load_config();

# Intersection types for multiple constraints
my Object&Serializable $api_response = get_api_data();

# Negation types to exclude specific types
my !Undef $required_value = get_required_data();

# Parameterized types with constraints
my ArrayRef[HashRef[Str, Int|Str]] $table_data = [
    { name => "Alice", age => 30, status => "active" },
    { name => "Bob",   age => 25, status => "pending" }
];
```

### Function and Method Annotation Patterns

#### Function Type Signatures

```perl
# Simple function with typed parameters and return
sub Num calculate_discount(Num $price, Num $rate) {
    return $price * ($rate / 100);
}

# Function with optional parameters using Maybe types
sub Str format_name(Str $first, Maybe[Str] $middle, Str $last) {
    my $full_name = $first;
    if (defined($middle)) {
        $full_name .= " $middle";
    }
    $full_name .= " $last";
    return $full_name;
}

# Function with complex parameter types
sub ArrayRef[HashRef] process_orders(ArrayRef[HashRef[Str, Str|Int|Num]] $orders,
                  HashRef[Str] $config) {
    my @processed;
    for my $order (@$orders) {
        my $processed_order = process_single_order($order, $config);
        push @processed, $processed_order;
    }
    return \@processed;
}

# Higher-order function patterns
sub ArrayRef[Any] map_transform(ArrayRef[Any] $array, CodeRef $transformer) {
    my @result;
    for my $item (@$array) {
        push @result, $transformer->($item);
    }
    return \@result;
}
```

#### Method Type Patterns

```perl
# Object method with typed parameters and return
method Bool validate_email(Str $email) {
    return $email =~ /^[\w._%+-]+@[\w.-]+\.[A-Za-z]{2,}$/;
}

# Method with complex object interactions
method Self merge_data(HashRef[Any] $new_data) {
    for my $key (keys %$new_data) {
        $self->{data}->{$key} = $new_data->{$key};
    }
    return $self;
}

# Method returning Maybe type for safe operations
method Maybe[User] find_user(Int $user_id) {
    my $user_data = $self->get_user_data($user_id);
    return defined($user_data) ? User->new($user_data) : undef;
}

# Method with callback parameter
method Promise process_async(CodeRef $callback, Maybe[HashRef] $options = undef) {
    my $opts = $options // {};
    return $self->async_processor->process($callback, $opts);
}
```

## Object-Oriented Patterns

### Class Definition Patterns

#### Basic Class with Typed Fields

```perl
class User {
    field Str $name ;
    field Int $age ;
    field Maybe[Str] $email = undef;
    field ArrayRef[Str] $roles = [];
    field Bool $is_active = 1;

    # Constructor pattern with validation
    method BUILD($args) {
        $self->validate_age($age) if defined $age;
        $self->validate_email($email) if defined $email;
    }

    # Getter with return type annotation
    method Str get_name() {
        return $name;
    }

    # Setter with validation
    method Self set_email(Maybe[Str] $new_email) {
        if (defined($new_email)) {
            die "Invalid email format" unless $self->validate_email($new_email);
        }
        $email = $new_email;
        return $self;
    }

    # Method with complex logic and flow-sensitive analysis
    method Str get_display_name() {
        my Maybe[Str] $display = $email;
        if (defined($display) && $display =~ /^([^@]+)/) {
            # Flow analysis knows $display is Str here
            return ucfirst($1);
        }
        return $name;
    }
}
```

#### Inheritance and Interface Patterns

```perl
# Base class with abstract methods
class Shape {
    field Num $x ;
    field Num $y ;

    # Abstract method declaration
    method Num calculate_area() {
        die "Abstract method must be implemented";
    }

    # Concrete method available to subclasses
    method Self move_to(Num $new_x, Num $new_y) {
        $x = $new_x;
        $y = $new_y;
        return $self;
    }
}

# Concrete implementation
class Circle isa Shape {
    field Num $radius ;

    method Num calculate_area() {
        return 3.14159 * $radius * $radius;
    }

    method Num get_circumference() {
        return 2 * 3.14159 * $radius;
    }
}

# Interface-like role pattern
role Serializable {
    method Str to_json();
    method Self from_json(Str $json);
}

class User does Serializable {
    field Str $name ;
    field Int $age ;

    method Str to_json() {
        my HashRef $data = { name => $name, age => $age };
        return encode_json($data);
    }

    method Self from_json(Str $json) {
        my HashRef $data = decode_json($json);
        return $class->new(
            name => $data->{name},
            age  => $data->{age}
        );
    }
}
```

#### Builder and Factory Patterns

```perl
# Builder pattern with fluent interface
class UserBuilder {
    field Maybe[Str] $name = undef;
    field Maybe[Int] $age = undef;
    field ArrayRef[Str] $roles = [];
    field Maybe[Str] $email = undef;

    method Self with_name(Str $user_name) {
        $name = $user_name;
        return $self;
    }

    method Self with_age(Int $user_age) {
        die "Age must be positive" if $user_age < 0;
        $age = $user_age;
        return $self;
    }

    method Self with_email(Str $user_email) {
        $email = $user_email;
        return $self;
    }

    method Self add_role(Str $role) {
        push @$roles, $role;
        return $self;
    }

    method User build() {
        die "Name is required" unless defined($name);
        die "Age is required" unless defined($age);

        return User->new(
            name  => $name,
            age   => $age,
            roles => $roles,
            email => $email
        );
    }
}

# Factory pattern with type validation
class UserFactory {
    method User create_from_data(HashRef[Any] $data) {
        my UserBuilder $builder = UserBuilder->new();

        # Type-safe data extraction with validation
        if (exists $data->{name} && defined $data->{name}) {
            my Str $name = $data->{name};
            $builder = $builder->with_name($name);
        }

        if (exists $data->{age} && defined $data->{age}) {
            my Int $age = int($data->{age});
            $builder = $builder->with_age($age);
        }

        if (exists $data->{email} && defined $data->{email}) {
            my Str $email = $data->{email};
            $builder = $builder->with_email($email);
        }

        return $builder->build();
    }

    method User create_admin(Str $name, Str $email) {
        return UserBuilder->new()
            ->with_name($name)
            ->with_email($email)
            ->add_role("admin")
            ->add_role("user")
            ->with_age(25)  # Default admin age
            ->build();
    }
}
```

## Functional Programming Patterns

### Higher-Order Function Patterns

```perl
# Function composition with type safety
sub CodeRef compose(CodeRef $f, CodeRef $g) {
    return sub {
        my @args = @_;
        return $f->($g->(@args));
    };
}

# Map operation with type preservation
sub ArrayRef[Any] typed_map(ArrayRef[Any] $array, CodeRef $transform) {
    my @result;
    for my $item (@$array) {
        push @result, $transform->($item);
    }
    return \@result;
}

# Filter operation with type-safe predicates
sub ArrayRef[Any] typed_filter(ArrayRef[Any] $array, CodeRef $predicate) {
    my @result;
    for my $item (@$array) {
        push @result, $item if $predicate->($item);
    }
    return \@result;
}

# Reduce operation with accumulator typing
sub Any typed_reduce(ArrayRef[Any] $array, Any $initial, CodeRef $reducer) {
    my $accumulator = $initial;
    for my $item (@$array) {
        $accumulator = $reducer->($accumulator, $item);
    }
    return $accumulator;
}

# Partial application with type preservation
sub CodeRef partial(CodeRef $func, @fixed_args) {
    return sub {
        my @remaining_args = @_;
        return $func->(@fixed_args, @remaining_args);
    };
}

# Example usage of functional patterns
my ArrayRef[Int] $numbers = [1, 2, 3, 4, 5];

# Compose operations
my CodeRef $double = sub { my Int $x = shift; return $x * 2; };
my CodeRef $add_one = sub { my Int $x = shift; return $x + 1; };
my CodeRef $double_and_add = compose($add_one, $double);

# Chain operations with type safety
my ArrayRef[Int] $doubled = typed_map($numbers, $double);
my ArrayRef[Int] $evens = typed_filter($doubled, sub { shift() % 2 == 0 });
my Int $sum = typed_reduce($evens, 0, sub { my ($acc, $val) = @_; return $acc + $val; });
```

### Monad-Like Patterns for Error Handling

```perl
# Result type for error handling
class Result[T] {
    field Bool $is_success ;
    field Maybe[T] $value = undef;
    field Maybe[Str] $error = undef;

    method Result[T] success(T $val) {
        return $class->new(
            is_success => 1,
            value      => $val,
            error      => undef
        );
    }

    method Result[T] failure(Str $err) {
        return $class->new(
            is_success => 0,
            value      => undef,
            error      => $err
        );
    }

    method Result[Any] map(CodeRef $transform) {
        return $self unless $is_success;

        my $transformed_value;
        eval {
            $transformed_value = $transform->($value);
        };
        if ($@) {
            return Result->failure("Transform failed: $@");
        }

        return Result->success($transformed_value);
    }

    method Result[Any] flat_map(CodeRef $transform) {
        return $self unless $is_success;

        my Result $result;
        eval {
            $result = $transform->($value);
        };
        if ($@) {
            return Result->failure("FlatMap failed: $@");
        }

        return $result;
    }

    method T get_or_else(T $default) {
        return $is_success ? $value : $default;
    }
}

# Usage example with chained operations
sub Result[Num] divide(Num $a, Num $b) {
    return $b == 0
        ? Result->failure("Division by zero")
        : Result->success($a / $b);
}

sub Result[Num] square_root(Num $x) {
    return $x < 0
        ? Result->failure("Cannot take square root of negative number")
        : Result->success(sqrt($x));
}

# Chain operations safely
my Result[Num] $result = divide(16, 4)
    ->flat_map(sub { my Num $x = shift; return square_root($x); })
    ->map(sub { my Num $x = shift; return $x * 2; });

if ($result->is_success) {
    say "Result: " . $result->value;
} else {
    say "Error: " . $result->error;
}
```

## Error Handling Patterns

### Exception Handling with Types

```perl
# Custom exception classes with type information
class ValidationException isa Exception {
    field Str $field_name ;
    field Any $invalid_value ;
    field Str $constraint ;

    method BUILD($args) {
        my Str $message = "Validation failed for field '$field_name': " .
                         "value '$invalid_value' violates constraint '$constraint'";
        $self->SUPER::BUILD({ message => $message });
    }
}

class BusinessLogicException isa Exception {
    field Str $operation ;
    field HashRef[Any] $context ;

    method BUILD($args) {
        my Str $message = "Business logic error in operation '$operation'";
        $self->SUPER::BUILD({ message => $message });
    }
}

# Service class with comprehensive error handling
class UserService {
    field DatabaseConnection $db ;
    field Logger $logger ;

    method Result[User] create_user(HashRef[Any] $user_data) {
        eval {
            # Validate input data
            $self->validate_user_data($user_data);

            # Check business rules
            $self->check_business_rules($user_data);

            # Create user
            my User $user = User->new($user_data);
            my Int $user_id = $db->insert_user($user);
            $user->set_id($user_id);

            $logger->info("User created successfully: ID $user_id");
            return Result->success($user);
        };

        # Handle specific exception types
        if (my ValidationException $e = $@) {
            $logger->warn("Validation error: " . $e->message);
            return Result->failure("Invalid user data: " . $e->constraint);
        } elsif (my BusinessLogicException $e = $@) {
            $logger->error("Business logic error: " . $e->message);
            return Result->failure("Business rule violation in " . $e->operation);
        } elsif ($@) {
            $logger->error("Unexpected error: $@");
            return Result->failure("Internal server error");
        }
    }

    method Void validate_user_data(HashRef[Any] $data) {
        # Type-safe validation with detailed errors
        unless (exists $data->{name} && defined $data->{name}) {
            die ValidationException->new(
                field_name    => "name",
                invalid_value => $data->{name} // "<undefined>",
                constraint    => "required"
            );
        }

        my $name = $data->{name};
        unless (length($name) >= 2) {
            die ValidationException->new(
                field_name    => "name",
                invalid_value => $name,
                constraint    => "minimum length 2"
            );
        }

        if (exists $data->{email} && defined $data->{email}) {
            my Str $email = $data->{email};
            unless ($email =~ /^[\w._%+-]+@[\w.-]+\.[A-Za-z]{2,}$/) {
                die ValidationException->new(
                    field_name    => "email",
                    invalid_value => $email,
                    constraint    => "valid email format"
                );
            }
        }
    }
}
```

### Safe Data Access Patterns

```perl
# Safe hash access with Maybe types
sub Maybe[Any] safe_get(HashRef[Any] $hash, Str $key) {
    return exists $hash->{$key} ? $hash->{$key} : undef;
}

# Safe array access with bounds checking
sub Maybe[Any] safe_array_get(ArrayRef[Any] $array, Int $index) {
    return ($index >= 0 && $index < @$array) ? $array->[$index] : undef;
}

# Chained safe access
sub Maybe[Any] safe_get_nested(HashRef[Any] $data, ArrayRef[Str] $path) {
    my $current = $data;

    for my $key (@$path) {
        # Flow-sensitive analysis helps here
        return undef unless defined($current) && ref($current) eq 'HASH';

        my Maybe[Any] $next = safe_get($current, $key);
        return undef unless defined($next);

        $current = $next;
    }

    return $current;
}

# Usage examples with flow-sensitive refinement
my HashRef[Any] $config = load_config();

# Safe access with explicit checking
my Maybe[Any] $db_config = safe_get($config, "database");
if (defined($db_config) && ref($db_config) eq 'HASH') {
    # Flow analysis knows $db_config is HashRef here
    my Maybe[Str] $host = safe_get($db_config, "host");
    my Maybe[Int] $port = safe_get($db_config, "port");

    if (defined($host) && defined($port)) {
        # Both $host and $port are refined to their concrete types
        say "Connecting to $host:$port";
    }
}

# Chained access example
my Maybe[Any] $user_email = safe_get_nested($config, ["users", "admin", "email"]);
if (defined($user_email)) {
    say "Admin email: $user_email";
}
```

## Performance Optimization Patterns

### Lazy Evaluation Patterns

```perl
# Lazy initialization with type safety
class ExpensiveResource {
    field Maybe[DatabaseConnection] $_connection = undef;
    field Str $connection_string ;

    method DatabaseConnection get_connection() {
        # Lazy initialization with flow-sensitive refinement
        unless (defined($_connection)) {
            $_connection = DatabaseConnection->new($connection_string);
        }

        # Flow analysis knows $_connection is defined here
        return $_connection;
    }
}

# Memoization pattern with typed cache
class MemoizedCalculator {
    field HashRef[Num] $_cache = {};

    method Num fibonacci(Int $n) {
        # Type-safe cache key
        my Str $cache_key = "fib_$n";

        # Check cache first
        my Maybe[Num] $cached = safe_get($_cache, $cache_key);
        if (defined($cached)) {
            return $cached;
        }

        # Calculate and cache result
        my Num $result;
        if ($n <= 1) {
            $result = $n;
        } else {
            $result = $self->fibonacci($n - 1) + $self->fibonacci($n - 2);
        }

        $_cache->{$cache_key} = $result;
        return $result;
    }
}

# Iterator pattern for memory efficiency
class NumberIterator {
    field Int $start ;
    field Int $end ;
    field Int $_current ;

    method BUILD($args) {
        $_current = $start;
    }

    method Bool has_next() {
        return $_current <= $end;
    }

    method Maybe[Int] next() {
        return undef unless $self->has_next();

        my Int $value = $_current;
        $_current++;
        return $value;
    }

    method ArrayRef[Int] to_array() {
        my @result;
        while ($self->has_next()) {
            push @result, $self->next();
        }
        return \@result;
    }
}
```

### Type-Safe Caching Patterns

```perl
# Generic cache with type parameters
class TypedCache[K, V] {
    field HashRef[V] $_cache = {};
    field CodeRef $_serializer ;
    field Int $max_size = 1000;
    field Int $_access_count = 0;

    method Maybe[V] get(K $key) {
        my Str $cache_key = $self->serialize_key($key);
        $_access_count++;

        return safe_get($_cache, $cache_key);
    }

    method Void set(K $key, V $value) {
        my Str $cache_key = $self->serialize_key($key);

        # Evict old entries if cache is full
        if (keys(%$_cache) >= $max_size) {
            $self->evict_oldest();
        }

        $_cache->{$cache_key} = $value;
    }

    method Str serialize_key(K $key) {
        return $_serializer->($key);
    }

    method Void evict_oldest() {
        # Simple FIFO eviction (could be enhanced with LRU)
        my @keys = keys %$_cache;
        my Str $oldest_key = shift @keys;
        delete $_cache->{$oldest_key};
    }
}

# Usage with specific types
my TypedCache[Int, Str] $string_cache = TypedCache->new(
    serializer => sub { my Int $key = shift; return "key_$key"; }
);

$string_cache->set(1, "First value");
$string_cache->set(2, "Second value");

my Maybe[Str] $value = $string_cache->get(1);
if (defined($value)) {
    say "Cached value: $value";
}
```

## Testing Patterns

### Type-Safe Test Utilities

```perl
# Test fixture with typed data
class UserTestFixture {
    method HashRef[Str|Int] create_valid_user_data() {
        return {
            name  => "Test User",
            age   => 25,
            email => "test@example.com"
        };
    }

    method HashRef[Str|Int] create_invalid_user_data() {
        return {
            name  => "",  # Invalid: empty name
            age   => -5,  # Invalid: negative age
            email => "invalid-email"  # Invalid: bad format
        };
    }

    method ArrayRef[User] create_test_users(Int $count) {
        my @users;
        for my $i (1..$count) {
            push @users, User->new(
                name  => "User $i",
                age   => 20 + $i,
                email => "user$i\@example.com"
            );
        }
        return \@users;
    }
}

# Type-safe assertion helpers
sub Void assert_type(Any $value, Str $expected_type, Str $message = "") {
    my Str $actual_type = ref($value) || "SCALAR";

    if ($actual_type ne $expected_type) {
        my Str $error = $message || "Type assertion failed";
        die "$error: expected $expected_type, got $actual_type";
    }
}

sub Void assert_result_success(Result[Any] $result, Str $message = "") {
    unless ($result->is_success) {
        my Str $error = $message || "Expected successful result";
        die "$error: " . ($result->error // "unknown error");
    }
}

sub Void assert_result_failure(Result[Any] $result, Str $message = "") {
    if ($result->is_success) {
        my Str $error = $message || "Expected failed result";
        die "$error: got successful result instead";
    }
}

# Example test with type-safe patterns
use Test2::V0;

subtest "User creation and validation" => sub {
    my UserTestFixture $fixture = UserTestFixture->new();
    my UserService $service = UserService->new(
        db     => MockDatabase->new(),
        logger => MockLogger->new()
    );

    # Test successful user creation
    my HashRef[Str|Int] $valid_data = $fixture->create_valid_user_data();
    my Result[User] $result = $service->create_user($valid_data);

    assert_result_success($result, "User creation should succeed");

    my User $user = $result->value;
    assert_type($user, "User", "Result should contain User object");
    is($user->get_name(), "Test User", "User name should match");

    # Test validation failure
    my HashRef[Str|Int] $invalid_data = $fixture->create_invalid_user_data();
    my Result[User] $invalid_result = $service->create_user($invalid_data);

    assert_result_failure($invalid_result, "Invalid user data should fail");
    like($invalid_result->error, qr/validation/i, "Error should mention validation");
};
```

## Integration Patterns

### Database Integration with Type Safety

```perl
# Type-safe database row representation
class DatabaseRow {
    field HashRef[Any] $_data ;
    field ArrayRef[Str] $_columns ;

    method Maybe[Str] get_string(Str $column) {
        my Maybe[Any] $value = safe_get($_data, $column);
        return defined($value) ? "$value" : undef;
    }

    method Maybe[Int] get_int(Str $column) {
        my Maybe[Any] $value = safe_get($_data, $column);
        return defined($value) ? int($value) : undef;
    }

    method Maybe[Bool] get_bool(Str $column) {
        my Maybe[Any] $value = safe_get($_data, $column);
        return undef unless defined($value);
        return $value ? 1 : 0;
    }
}

# Repository pattern with type safety
class UserRepository {
    field DatabaseConnection $db ;

    method Maybe[User] find_by_id(Int $user_id) {
        my Maybe[DatabaseRow] $row = $self->query_single(
            "SELECT * FROM users WHERE id = ?",
            [$user_id]
        );

        return defined($row) ? $self->row_to_user($row) : undef;
    }

    method Maybe[User] find_by_email(Str $email) {
        my Maybe[DatabaseRow] $row = $self->query_single(
            "SELECT * FROM users WHERE email = ?",
            [$email]
        );

        return defined($row) ? $self->row_to_user($row) : undef;
    }

    method Result[Int] save(User $user) {
        eval {
            if (defined($user->get_id())) {
                return $self->update_user($user);
            } else {
                return $self->insert_user($user);
            }
        };

        if ($@) {
            return Result->failure("Database error: $@");
        }
    }

    method User row_to_user(DatabaseRow $row) {
        return User->new(
            id    => $row->get_int("id"),
            name  => $row->get_string("name"),
            email => $row->get_string("email"),
            age   => $row->get_int("age")
        );
    }
}
```

### API Integration Patterns

```perl
# Type-safe HTTP client
class ApiClient {
    field Str $base_url ;
    field HashRef[Str] $_headers = {};

    method Result[HashRef[Any]] get(Str $endpoint) {
        return $self->request("GET", $endpoint);
    }

    method Result[HashRef[Any]] post(Str $endpoint, HashRef[Any] $data) {
        return $self->request("POST", $endpoint, $data);
    }

    method Result[HashRef[Any]] request(Str $method, Str $endpoint, Maybe[HashRef[Any]] $data = undef) {
        eval {
            my Str $url = $base_url . $endpoint;
            my HashRef[Any] $response = $self->http_request($method, $url, $data);

            # Validate response structure
            unless (exists $response->{status}) {
                die "Invalid response: missing status";
            }

            my Int $status = int($response->{status});
            if ($status >= 400) {
                my Str $error = $response->{error} // "HTTP error $status";
                return Result->failure($error);
            }

            return Result->success($response->{data} // {});
        };

        if ($@) {
            return Result->failure("Request failed: $@");
        }
    }
}

# Typed API response models
class ApiResponse[T] {
    field Bool $success ;
    field Maybe[T] $data = undef;
    field Maybe[Str] $error = undef;
    field HashRef[Any] $metadata = {};

    method ApiResponse[T] from_hash(HashRef[Any] $response_data, CodeRef $parser) {
        my Bool $success = $response_data->{success} // 0;

        if ($success) {
            my T $parsed_data = $parser->($response_data->{data});
            return $class->new(
                success  => 1,
                data     => $parsed_data,
                metadata => $response_data->{metadata} // {}
            );
        } else {
            return $class->new(
                success => 0,
                error   => $response_data->{error} // "Unknown error"
            );
        }
    }
}
```

## Best Practices and Anti-Patterns

### Type Annotation Best Practices

```perl
# ✅ GOOD: Use specific types when possible
my ArrayRef[User] $users = get_users();
my HashRef[Str, Int] $counters = {};

# ❌ AVOID: Overly generic types
my ArrayRef[Any] $data = get_users();  # Too generic
my HashRef $config = {};               # Missing value type

# ✅ GOOD: Use Maybe types for optional values
my Maybe[Str] $optional_field = get_optional_value();
if (defined($optional_field)) {
    # Safe to use $optional_field as Str here
}

# ❌ AVOID: Assuming values are always defined
my Str $field = get_optional_value();  # Might be undef

# ✅ GOOD: Leverage flow-sensitive analysis
my Maybe[User] $user = find_user($id);
if (defined($user)) {
    # $user is refined to User type here
    say $user->get_name();
}

# ❌ AVOID: Unnecessary type assertions
my Maybe[User] $user = find_user($id);
if (defined($user)) {
    my User $confirmed_user = $user;  # Unnecessary
    say $confirmed_user->get_name();
}
```

### Error Handling Best Practices

```perl
# ✅ GOOD: Use Result types for operations that can fail
method Result[Num] divide(Num $a, Num $b) {
    return $b == 0
        ? Result->failure("Division by zero")
        : Result->success($a / $b);
}

# ❌ AVOID: Throwing exceptions for expected failures
method Num divide(Num $a, Num $b) {
    die "Division by zero" if $b == 0;  # Expected failure
    return $a / $b;
}

# ✅ GOOD: Chain operations with proper error propagation
my Result[Str] $result = get_user_input()
    ->flat_map(sub { validate_input(shift) })
    ->map(sub { process_input(shift) });

# ❌ AVOID: Nested try-catch blocks
my Str $final_result;
eval {
    my $input = get_user_input();
    eval {
        my $validated = validate_input($input);
        $final_result = process_input($validated);
    };
    die "Validation failed: $@" if $@;
};
die "Input failed: $@" if $@;
```

### Performance Best Practices

```perl
# ✅ GOOD: Use lazy initialization for expensive resources
field Maybe[DatabaseConnection] $_connection = undef;

method DatabaseConnection get_connection() {
    $_connection //= DatabaseConnection->new($self->connection_string);
    return $_connection;
}

# ❌ AVOID: Eager initialization of expensive resources
field DatabaseConnection $connection = DatabaseConnection->new($connection_string);

# ✅ GOOD: Use iterators for large datasets
method Void process_large_dataset() {
    my Iterator[DataRow] $iterator = $self->get_data_iterator();
    while ($iterator->has_next()) {
        my DataRow $row = $iterator->next();
        $self->process_row($row);
    }
}

# ❌ AVOID: Loading entire datasets into memory
method Void process_large_dataset() {
    my ArrayRef[DataRow] $all_data = $self->get_all_data();  # Memory intensive
    for my $row (@$all_data) {
        $self->process_row($row);
    }
}
```

## Advanced Type Patterns

### Generic Programming Patterns

```perl
# Generic container with type constraints
class Container[T] {
    field ArrayRef[T] $_items = [];
    field Int $max_size = 100;

    method Result[Void] add(T $item) {
        if (@$_items >= $max_size) {
            return Result->failure("Container is full");
        }

        push @$_items, $item;
        return Result->success(undef);
    }

    method Maybe[T] get(Int $index) {
        return safe_array_get($_items, $index);
    }

    method Container[Any] map(CodeRef $transform) {
        my @transformed = map { $transform->($_) } @$_items;
        return Container->new(items => \@transformed);
    }

    method Container[T] filter(CodeRef $predicate) {
        my @filtered = grep { $predicate->($_) } @$_items;
        return Container->new(items => \@filtered);
    }
}

# Usage with specific types
my Container[User] $user_container = Container->new();
$user_container->add(User->new(name => "Alice"));
$user_container->add(User->new(name => "Bob"));

my Container[Str] $name_container = $user_container->map(
    sub { my User $user = shift; return $user->get_name(); }
);
```

### State Machine Patterns

```perl
# Type-safe state machine
class OrderState {
    # Abstract base class for states
}

class PendingState isa OrderState {
    method Bool can_transition_to(Str $state) {
        return $state eq "confirmed" || $state eq "cancelled";
    }
}

class ConfirmedState isa OrderState {
    method Bool can_transition_to(Str $state) {
        return $state eq "shipped" || $state eq "cancelled";
    }
}

class ShippedState isa OrderState {
    method Bool can_transition_to(Str $state) {
        return $state eq "delivered";
    }
}

class Order {
    field Int $id ;
    field OrderState $state ;
    field ArrayRef[Str] $history = [];

    method Result[Void] transition_to(Str $new_state) {
        unless ($state->can_transition_to($new_state)) {
            return Result->failure(
                "Cannot transition from " . ref($state) . " to $new_state"
            );
        }

        my OrderState $new_state_obj = $self->create_state($new_state);
        push @$history, ref($state) . " returns " . ref($new_state_obj);
        $state = $new_state_obj;

        return Result->success(undef);
    }

    method OrderState create_state(Str $state_name) {
        given ($state_name) {
            when ("pending")   { return PendingState->new() }
            when ("confirmed") { return ConfirmedState->new() }
            when ("shipped")   { return ShippedState->new() }
            default { die "Unknown state: $state_name" }
        }
    }
}
```

## Related Documentation

- [typed-perl-specification.md](typed-perl-specification.md) - Complete type system reference
- [workflow-new-development.md](workflow-new-development.md) - Development environment setup
- [workflow-ci-cd-integration.md](workflow-ci-cd-integration.md) - CI/CD integration patterns

## Advanced Topics

For advanced type system features, performance optimization techniques, and integration with external systems, see the [Development Log](development-log.md) for detailed implementation examples and lessons learned from real-world typed-Perl projects.
