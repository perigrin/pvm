#!/usr/bin/perl
use v5.36;

# ABOUTME: Sample typed Perl file for PSC e2e testing
# ABOUTME: Demonstrates various type annotations and language features

# Variable type annotations
my Int $count = 42;
my Str $name = "Sample Module";
my Bool $is_active = 1;
my ArrayRef[Int] $numbers = [1, 2, 3, 4, 5];
my HashRef[Str, Int] $scores = {
    alice => 95,
    bob => 87,
    charlie => 92
};

# Maybe types for nullable values
my Maybe[Str] $optional_message = undef;
my Maybe[Int] $optional_count = 10;

# Union types for flexible values
my Union[Int, Str] $flexible_value = "hello";

# Function with type annotations
sub Int add_numbers(Int $a, Int $b) {
    return $a + $b;
}

# Function with complex parameter types
sub Num calculate_average(ArrayRef[Int] $numbers) {
    my Int $sum = 0;
    my Int $count = 0;

    for my Int $num (@$numbers) {
        $sum += $num;
        $count++;
    }

    return $count > 0 ? $sum / $count : 0;
}

# Function with Maybe parameter
sub Str greet_user(Maybe[Str] $name) {
    if (defined($name)) {
        return "Hello, $name!";
    } else {
        return "Hello, anonymous user!";
    }
}

# Method with type annotations (simplified class syntax)
sub Object new(Str $class, Str $name, Int $age) {
    my HashRef[Str, Any] $self = {
        name => $name,
        age => $age,
        scores => {},
    };
    return bless $self, $class;
}

# Method with typed parameters
sub Void add_score(Object $self, Str $subject, Int $score) {
    $self->{scores}->{$subject} = $score;
}

# Method with typed return value
sub Num get_average_score(Object $self) {
    my HashRef[Str, Int] $scores = $self->{scores};
    my @values = values %$scores;

    return @values ? (sum(@values) / @values) : 0;
}

# Flow-sensitive type checking examples
sub Str process_input(Union[Int, Str] $input) {
    if ($input =~ /^\d+$/) {
        # Here $input is refined to be numeric
        my Int $num = int($input);
        return "Number: " . ($num * 2);
    } else {
        # Here $input is treated as string
        return "Text: " . uc($input);
    }
}

# Pattern matching with type refinement
sub Union[Int, Str] validate_and_process(Str $input) {
    if ($input =~ /^(\d+)$/) {
        # $1 is captured as string but we know it's numeric
        return int($1);
    } else {
        return "Invalid format: $input";
    }
}

# Reference type checking
sub Str process_reference(Scalar $ref) {
    if (ref($ref) eq 'ARRAY') {
        # $ref refined to ArrayRef
        return "Array with " . scalar(@$ref) . " elements";
    } elsif (ref($ref) eq 'HASH') {
        # $ref refined to HashRef
        my @keys = keys %$ref;
        return "Hash with keys: " . join(", ", @keys);
    } else {
        return "Simple scalar: $ref";
    }
}

# Main execution
my Int $sum = add_numbers($count, 58);
my Num $average = calculate_average($numbers);
my Str $greeting = greet_user($optional_message);

say "Sum: $sum";
say "Average: $average";
say $greeting;

# Test flexible value processing
my Str $processed1 = process_input(42);
my Str $processed2 = process_input("hello world");
say $processed1;
say $processed2;

# Test validation
my Union[Int, Str] $validated1 = validate_and_process("123");
my Union[Int, Str] $validated2 = validate_and_process("abc");
say "Validated1: $validated1";
say "Validated2: $validated2";

# Test reference processing
my Str $ref_result1 = process_reference($numbers);
my Str $ref_result2 = process_reference($scores);
my Str $ref_result3 = process_reference("simple");
say $ref_result1;
say $ref_result2;
say $ref_result3;

# Create and use an object
my Object $user = new("User", "Alice", 25);
add_score($user, "Math", 95);
add_score($user, "Science", 87);
my Num $user_average = get_average_score($user);
say "User average: $user_average";

1;
