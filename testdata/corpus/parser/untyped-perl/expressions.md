---
category: untyped-perl
subcategory: expressions
tags:
    - addition
    - and
    - arithmetic
    - array
    - assignment
    - basic
    - binary_operator
    - bit_manipulation
    - bitwise
    - chained
    - comma
    - comparison
    - complement
    - complex
    - concatenation
    - conditional
    - decrement
    - defined_or
    - dereferencing
    - division
    - equality
    - evaluation
    - exponentiation
    - flag_clearing
    - flag_setting
    - flag_toggling
    - function_call
    - function_calls
    - greater_equal
    - greater_than
    - hash
    - hexadecimal
    - identity
    - increment
    - indexing
    - inequality
    - interpolation
    - key_access
    - left_shift
    - less_equal
    - less_than
    - list
    - literals
    - logical
    - low_precedence
    - mask
    - matching
    - math
    - method_calls
    - minus
    - mixed
    - mixed_operators
    - modulo
    - multiple
    - multiple_operators
    - multiplication
    - nested
    - not
    - numeric
    - objects
    - or
    - ordering
    - parentheses
    - plus
    - postfix
    - precedence
    - prefix
    - quoting
    - range
    - references
    - regex
    - repetition
    - right_shift
    - sequence
    - sequential
    - short_circuit
    - side_effects
    - slice
    - spaceship
    - string
    - subtraction
    - ternary
    - three_way
    - unary_operator
    - undef_handling
    - variables
    - word_form
    - xor
---

# Addition Assignment

Addition assignment operator

```perl
$total += $increment;
```

## Array Element Expression

Array element access in arithmetic expression

```perl
$value = $array[$index + 1] + $matrix[$row][$col];
```

## Basic Addition

Basic addition operator

```perl
$result = $a + $b;
```

## Basic Assignment

Basic assignment operator

```perl
$var = $value;
```

## Basic Division

Basic division operator

```perl
$result = $a / $b;
```

## Basic Multiplication

Basic multiplication operator

```perl
$result = $a * $b;
```

## Basic Subtraction

Basic subtraction operator

```perl
$result = $a - $b;
```

## Bit Clearing

Clearing a specific bit using NOT and AND

```perl
$bit_cleared = $flags & ~(1 << $bit_number);
```

## Bit Manipulation

Setting a specific bit using shift and OR

```perl
$bit_set = $flags | (1 << $bit_number);
```

## Bit Toggling

Toggling a specific bit using XOR

```perl
$bit_toggled = $flags ^ (1 << $bit_number);
```

## Bitwise And

Bitwise AND operation

```perl
$result = $a & $b;
```

## Bitwise And Assignment

Bitwise AND assignment operator

```perl
$flags &= $mask;
```

## Bitwise Not

Bitwise NOT (complement) operation

```perl
$result = ~$value;
```

## Bitwise Or

Bitwise OR operation

```perl
$result = $a | $b;
```

## Bitwise Or Assignment

Bitwise OR assignment operator

```perl
$flags |= $bits;
```

## Bitwise Precedence

Bitwise operations with operator precedence

```perl
$result = $a | $b & $c ^ $d;
```

## Bitwise With Hex

Bitwise operation with hexadecimal literal

```perl
$masked = $value & 0xFF00;
```

## Bitwise Xor

Bitwise XOR operation

```perl
$result = $a ^ $b;
```

## Bitwise Xor Assignment

Bitwise XOR assignment operator

```perl
$flags ^= $toggle;
```

## Chained Assignment

Chained assignment operations

```perl
$a = $b = $c = $value;
```

## Chained Comparisons

Chained comparisons for ordering check

```perl
$ordered = $a < $b && $b < $c && $c < $d;
```

## Chained Logical

Chained logical AND operations

```perl
$valid = $a && $b && $c && $d;
```

## Comma Operator

Comma operator for sequential evaluation

```perl
$result = ($operation1, $operation2, $final_value);
```

## Comparison With Literals

Comparisons with literal values

```perl
$check = $value > 0 && $count <= 100 && $name ne '';
```

## Complex Arithmetic

Complex arithmetic expression with operator precedence

```perl
$result = $a + $b * $c / $d - $e % $f;
```

## Complex Bitwise

Complex bitwise expression with multiple operations

```perl
$flags = ($a & $mask) | ($b << 4) | (~$c & 0xFF);
```

## Complex Logical

Complex logical expression with precedence

```perl
$result = ($a && $b) || ($c && !$d) || $e;
```

## Complex String Expression

Complex string expression with multiple operators

```perl
$result = ($prefix . $name) x 3 . $suffix;
```

## Conditional Assignment

Using logical OR for conditional assignment

```perl
$value = $input || $default_value;
```

## Deeply Nested Parentheses

Deeply nested parenthesized expression

```perl
$result = ((($a + $b) * ($c - $d)) / (($e + $f) || 1));
```

## Defined Or

Defined-or operator for handling undef values

```perl
$result = $value // $default;
```

## Defined Or Assignment

Defined-or assignment operator

```perl
$value //= $default;
```

## Division Assignment

Division assignment operator

```perl
$total /= $divisor;
```

## Exponentiation

Exponentiation operator

```perl
$power = $base ** $exponent;
```

## Exponentiation Assignment

Exponentiation assignment operator

```perl
$total **= $exponent;
```

## Function Call Expression

Function calls in arithmetic expression

```perl
$result = sqrt($value) + sin($angle * 3.14159 / 180);
```

## Hash Element Expression

Hash element access in arithmetic expression

```perl
$total = $hash{$key} * $config{'multiplier'} + $defaults{rate};
```

## Interpolated Strings

String interpolation in double quotes

```perl
$message = "Hello $name, your score is $score";
```

## Left Shift

Left shift operation

```perl
$result = $value << $positions;
```

## Left Shift Assignment

Left shift assignment operator

```perl
$value <<= $positions;
```

## List Assignment Expression

List assignment with function call

```perl
($first, $second, @rest) = split /,/, $csv_line;
```

## Logical And

Logical AND operator with short-circuit evaluation

```perl
$and_result = $a && $b;
```

## Logical And Assignment

Logical AND assignment operator

```perl
$flag &&= $condition;
```

## Logical Not

Logical NOT operator

```perl
$not_result = !$condition;
```

## Logical Or

Logical OR operator with short-circuit evaluation

```perl
$or_result = $x || $y;
```

## Logical Or Assignment

Logical OR assignment operator

```perl
$value ||= $default;
```

## Logical With Comparison

Logical operators combined with comparison operators

```perl
$result = ($a > 0) && ($b < 100) || ($c == 0);
```

## Method Call Expression

Method calls in arithmetic expression

```perl
$result = $object->method($arg) + $other->calculate($x, $y);
```

## Mixed Numeric String

Mixed numeric and string comparisons

```perl
$result = ($num == 42) && ($str eq 'hello');
```

## Mixed Operator Types

Expression mixing string, arithmetic, and comparison operators

```perl
$result = ($str . '_' . $num) eq ($prefix . ($count + 1));
```

## Mixed Precedence

Mixed logical operators with different precedence levels

```perl
$result = $a and $b || $c && $d or $e;
```

## Modulo Assignment

Modulo assignment operator

```perl
$total %= $modulus;
```

## Modulo Operation

Modulo operator for remainder calculation

```perl
$remainder = $dividend % $divisor;
```

## Multiple Assignment

Multiple assignment with list context

```perl
($a, $b, $c) = ($x, $y, $z);
```

## Multiplication Assignment

Multiplication assignment operator

```perl
$total *= $multiplier;
```

## Nested Ternary

Nested ternary operators for multiple conditions

```perl
$grade = $score >= 90 ? 'A' : $score >= 80 ? 'B' : $score >= 70 ? 'C' : 'F';
```

## Numeric Equality

Numeric equality comparison

```perl
$equal = $a == $b;
```

## Numeric Greater Equal

Numeric greater than or equal comparison

```perl
$greater_eq = $a >= $b;
```

## Numeric Greater Than

Numeric greater than comparison

```perl
$greater = $a > $b;
```

## Numeric Inequality

Numeric inequality comparison

```perl
$not_equal = $a != $b;
```

## Numeric Less Equal

Numeric less than or equal comparison

```perl
$less_eq = $a <= $b;
```

## Numeric Less Than

Numeric less than comparison

```perl
$less = $a < $b;
```

## Numeric Literals

Arithmetic with various numeric literal formats

```perl
$result = 42 + 3.14 + 1e5 + 0xFF;
```

## Numeric Spaceship

Numeric three-way comparison (spaceship operator)

```perl
$cmp_result = $a <=> $b;
```

## Parenthesized Arithmetic

Arithmetic with parentheses for precedence control

```perl
$result = ($a + $b) * ($c - $d) / ($e || 1);
```

## Postfix Increment

Postfix increment in expression

```perl
$result = $array[$counter++] + $hash{$key++};
```

## Prefix Decrement

Prefix increment and decrement in expression

```perl
$result = --$counter * ++$multiplier;
```

## Range Comparison

Range check using multiple comparisons

```perl
$in_range = ($min <= $value) && ($value <= $max);
```

## Range Operator

Range operator for generating sequences

```perl
@numbers = ($start .. $end);
```

## Reference Comparison

Reference equality comparison

```perl
$same_ref = $ref1 == $ref2;
```

## Reference Dereferencing

Reference dereferencing in expressions

```perl
$sum = ${$scalar_ref} + @{$array_ref} + keys %{$hash_ref};
```

## Regex Match Expression

Regular expression matching in logical expression

```perl
$valid = ($input =~ /^\d+$/) && ($email =~ /@\w+\.\w+$/);
```

## Right Shift

Right shift operation

```perl
$result = $value >> $positions;
```

## Right Shift Assignment

Right shift assignment operator

```perl
$value >>= $positions;
```

## Slice Expression

Array slice with range operator

```perl
@subset = @array[$start .. $end];
```

## String Compare

String three-way comparison operator

```perl
$cmp_result = $left cmp $right;
```

## String Concatenation

Basic string concatenation operator

```perl
$combined = $first . $second;
```

## String Concatenation Assignment

String concatenation assignment operator

```perl
$message .= $suffix;
```

## String Equality

String equality comparison

```perl
$equal = $left eq $right;
```

## String Greater Equal

String greater than or equal comparison

```perl
$greater_eq = $left ge $right;
```

## String Greater Than

String greater than comparison

```perl
$greater = $left gt $right;
```

## String Inequality

String inequality comparison

```perl
$not_equal = $left ne $right;
```

## String Less Equal

String less than or equal comparison

```perl
$less_eq = $left le $right;
```

## String Less Than

String less than comparison

```perl
$less = $left lt $right;
```

## String Literals

String operations with different literal types

```perl
$result = 'single' . "double" . `backtick` . qq{quoted};
```

## String Repetition

String repetition operator

```perl
$repeated = $string x $count;
```

## String Repetition Assignment

String repetition assignment operator

```perl
$pattern x= $count;
```

## Subtraction Assignment

Subtraction assignment operator

```perl
$total -= $decrement;
```

## Ternary Operator

Ternary conditional operator

```perl
$result = $condition ? $true_value : $false_value;
```

## Unary Minus

Unary minus operator

```perl
$result = -$value;
```

## Unary Plus

Unary plus operator

```perl
$result = +$value;
```

## Word And

Word form of logical AND with lower precedence

```perl
$and_result = $a and $b;
```

## Word Not

Word form of logical NOT

```perl
$not_result = not $condition;
```

## Word Or

Word form of logical OR with lower precedence

```perl
$or_result = $x or $y;
```

# Expected Compilation Outcomes

## Addition Assignment

### Clean Perl Output

```perl
$total += $increment;
```

### Typed Perl Output

```perl
$total += $increment;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Array Element Expression

### Clean Perl Output

```perl
$value = $array[$index + 1] + $matrix[$row][$col];
```

### Typed Perl Output

```perl
$value = $array[$index + 1] + $matrix[$row][$col];
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Basic Addition

### Clean Perl Output

```perl
$result = $a + $b;
```

### Typed Perl Output

```perl
$result = $a + $b;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Basic Assignment

### Clean Perl Output

```perl
$var = $value;
```

### Typed Perl Output

```perl
$var = $value;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Basic Division

### Clean Perl Output

```perl
$result = $a / $b;
```

### Typed Perl Output

```perl
$result = $a / $b;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Basic Multiplication

### Clean Perl Output

```perl
$result = $a * $b;
```

### Typed Perl Output

```perl
$result = $a * $b;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Basic Subtraction

### Clean Perl Output

```perl
$result = $a - $b;
```

### Typed Perl Output

```perl
$result = $a - $b;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Bit Clearing

### Clean Perl Output

```perl
$bit_cleared = $flags & ~(1 << $bit_number);
```

### Typed Perl Output

```perl
$bit_cleared = $flags & ~(1 << $bit_number);
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Bit Manipulation

### Clean Perl Output

```perl
$bit_set = $flags | (1 << $bit_number);
```

### Typed Perl Output

```perl
$bit_set = $flags | (1 << $bit_number);
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Bit Toggling

### Clean Perl Output

```perl
$bit_toggled = $flags ^ (1 << $bit_number);
```

### Typed Perl Output

```perl
$bit_toggled = $flags ^ (1 << $bit_number);
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Bitwise And

### Clean Perl Output

```perl
$result = $a & $b;
```

### Typed Perl Output

```perl
$result = $a & $b;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Bitwise And Assignment

### Clean Perl Output

```perl
$flags &= $mask;
```

### Typed Perl Output

```perl
$flags &= $mask;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Bitwise Not

### Clean Perl Output

```perl
$result = ~$value;
```

### Typed Perl Output

```perl
$result = ~$value;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Bitwise Or

### Clean Perl Output

```perl
$result = $a | $b;
```

### Typed Perl Output

```perl
$result = $a | $b;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Bitwise Or Assignment

### Clean Perl Output

```perl
$flags |= $bits;
```

### Typed Perl Output

```perl
$flags |= $bits;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Bitwise Precedence

### Clean Perl Output

```perl
$result = $a | $b & $c ^ $d;
```

### Typed Perl Output

```perl
$result = $a | $b & $c ^ $d;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Bitwise With Hex

### Clean Perl Output

```perl
$masked = $value & 0xFF00;
```

### Typed Perl Output

```perl
$masked = $value & 0xFF00;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Bitwise Xor

### Clean Perl Output

```perl
$result = $a ^ $b;
```

### Typed Perl Output

```perl
$result = $a ^ $b;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Bitwise Xor Assignment

### Clean Perl Output

```perl
$flags ^= $toggle;
```

### Typed Perl Output

```perl
$flags ^= $toggle;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Chained Assignment

### Clean Perl Output

```perl
$a = $b = $c = $value;
```

### Typed Perl Output

```perl
$a = $b = $c = $value;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Chained Comparisons

### Clean Perl Output

```perl
$ordered = $a < $b && $b < $c && $c < $d;
```

### Typed Perl Output

```perl
$ordered = $a < $b && $b < $c && $c < $d;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Chained Logical

### Clean Perl Output

```perl
$valid = $a && $b && $c && $d;
```

### Typed Perl Output

```perl
$valid = $a && $b && $c && $d;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Comma Operator

### Clean Perl Output

```perl
$result = ($operation1, $operation2, $final_value);
```

### Typed Perl Output

```perl
$result = ($operation1, $operation2, $final_value);
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Comparison With Literals

### Clean Perl Output

```perl
$check = $value > 0 && $count <= 100 && $name ne '';
```

### Typed Perl Output

```perl
$check = $value > 0 && $count <= 100 && $name ne '';
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Complex Arithmetic

### Clean Perl Output

```perl
$result = $a + $b * $c / $d - $e % $f;
```

### Typed Perl Output

```perl
$result = $a + $b * $c / $d - $e % $f;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Complex Bitwise

### Clean Perl Output

```perl
$flags = ($a & $mask) | ($b << 4) | (~$c & 0xFF);
```

### Typed Perl Output

```perl
$flags = ($a & $mask) | ($b << 4) | (~$c & 0xFF);
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Complex Logical

### Clean Perl Output

```perl
$result = ($a && $b) || ($c && !$d) || $e;
```

### Typed Perl Output

```perl
$result = ($a && $b) || ($c && !$d) || $e;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Complex String Expression

### Clean Perl Output

```perl
$result = ($prefix . $name) x 3 . $suffix;
```

### Typed Perl Output

```perl
$result = ($prefix . $name) x 3 . $suffix;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Conditional Assignment

### Clean Perl Output

```perl
$value = $input || $default_value;
```

### Typed Perl Output

```perl
$value = $input || $default_value;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Deeply Nested Parentheses

### Clean Perl Output

```perl
$result = ((($a + $b) * ($c - $d)) / (($e + $f) || 1));
```

### Typed Perl Output

```perl
$result = ((($a + $b) * ($c - $d)) / (($e + $f) || 1));
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Defined Or

### Clean Perl Output

```perl
$result = $value // $default;
```

### Typed Perl Output

```perl
$result = $value // $default;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Defined Or Assignment

### Clean Perl Output

```perl
$value //= $default;
```

### Typed Perl Output

```perl
$value //= $default;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Division Assignment

### Clean Perl Output

```perl
$total /= $divisor;
```

### Typed Perl Output

```perl
$total /= $divisor;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Exponentiation

### Clean Perl Output

```perl
$power = $base ** $exponent;
```

### Typed Perl Output

```perl
$power = $base ** $exponent;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Exponentiation Assignment

### Clean Perl Output

```perl
$total **= $exponent;
```

### Typed Perl Output

```perl
$total **= $exponent;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Function Call Expression

### Clean Perl Output

```perl
$result = sqrt($value) + sin($angle * 3.14159 / 180);
```

### Typed Perl Output

```perl
$result = sqrt($value) + sin($angle * 3.14159 / 180);
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Hash Element Expression

### Clean Perl Output

```perl
$total = $hash{$key} * $config{'multiplier'} + $defaults{rate};
```

### Typed Perl Output

```perl
$total = $hash{$key} * $config{'multiplier'} + $defaults{rate};
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Interpolated Strings

### Clean Perl Output

```perl
$message = "Hello $name, your score is $score";
```

### Typed Perl Output

```perl
$message = "Hello $name, your score is $score";
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Left Shift

### Clean Perl Output

```perl
$result = $value << $positions;
```

### Typed Perl Output

```perl
$result = $value << $positions;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Left Shift Assignment

### Clean Perl Output

```perl
$value <<= $positions;
```

### Typed Perl Output

```perl
$value <<= $positions;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## List Assignment Expression

### Clean Perl Output

```perl
($first, $second, @rest) = split /,/, $csv_line;
```

### Typed Perl Output

```perl
($first, $second, @rest) = split /,/, $csv_line;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Logical And

### Clean Perl Output

```perl
$and_result = $a && $b;
```

### Typed Perl Output

```perl
$and_result = $a && $b;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Logical And Assignment

### Clean Perl Output

```perl
$flag &&= $condition;
```

### Typed Perl Output

```perl
$flag &&= $condition;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Logical Not

### Clean Perl Output

```perl
$not_result = !$condition;
```

### Typed Perl Output

```perl
$not_result = !$condition;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Logical Or

### Clean Perl Output

```perl
$or_result = $x || $y;
```

### Typed Perl Output

```perl
$or_result = $x || $y;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Logical Or Assignment

### Clean Perl Output

```perl
$value ||= $default;
```

### Typed Perl Output

```perl
$value ||= $default;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Logical With Comparison

### Clean Perl Output

```perl
$result = ($a > 0) && ($b < 100) || ($c == 0);
```

### Typed Perl Output

```perl
$result = ($a > 0) && ($b < 100) || ($c == 0);
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Method Call Expression

### Clean Perl Output

```perl
$result = $object->method($arg) + $other->calculate($x, $y);
```

### Typed Perl Output

```perl
$result = $object->method($arg) + $other->calculate($x, $y);
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Mixed Numeric String

### Clean Perl Output

```perl
$result = ($num == 42) && ($str eq 'hello');
```

### Typed Perl Output

```perl
$result = ($num == 42) && ($str eq 'hello');
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Mixed Operator Types

### Clean Perl Output

```perl
$result = ($str . '_' . $num) eq ($prefix . ($count + 1));
```

### Typed Perl Output

```perl
$result = ($str . '_' . $num) eq ($prefix . ($count + 1));
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Mixed Precedence

### Clean Perl Output

```perl
$result = $a and $b || $c && $d or $e;
```

### Typed Perl Output

```perl
$result = $a and $b || $c && $d or $e;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Modulo Assignment

### Clean Perl Output

```perl
$total %= $modulus;
```

### Typed Perl Output

```perl
$total %= $modulus;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Modulo Operation

### Clean Perl Output

```perl
$remainder = $dividend % $divisor;
```

### Typed Perl Output

```perl
$remainder = $dividend % $divisor;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Multiple Assignment

### Clean Perl Output

```perl
($a, $b, $c) = ($x, $y, $z);
```

### Typed Perl Output

```perl
($a, $b, $c) = ($x, $y, $z);
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Multiplication Assignment

### Clean Perl Output

```perl
$total *= $multiplier;
```

### Typed Perl Output

```perl
$total *= $multiplier;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Nested Ternary

### Clean Perl Output

```perl
$grade = $score >= 90 ? 'A' : $score >= 80 ? 'B' : $score >= 70 ? 'C' : 'F';
```

### Typed Perl Output

```perl
$grade = $score >= 90 ? 'A' : $score >= 80 ? 'B' : $score >= 70 ? 'C' : 'F';
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Numeric Equality

### Clean Perl Output

```perl
$equal = $a == $b;
```

### Typed Perl Output

```perl
$equal = $a == $b;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Numeric Greater Equal

### Clean Perl Output

```perl
$greater_eq = $a >= $b;
```

### Typed Perl Output

```perl
$greater_eq = $a >= $b;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Numeric Greater Than

### Clean Perl Output

```perl
$greater = $a > $b;
```

### Typed Perl Output

```perl
$greater = $a > $b;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Numeric Inequality

### Clean Perl Output

```perl
$not_equal = $a != $b;
```

### Typed Perl Output

```perl
$not_equal = $a != $b;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Numeric Less Equal

### Clean Perl Output

```perl
$less_eq = $a <= $b;
```

### Typed Perl Output

```perl
$less_eq = $a <= $b;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Numeric Less Than

### Clean Perl Output

```perl
$less = $a < $b;
```

### Typed Perl Output

```perl
$less = $a < $b;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Numeric Literals

### Clean Perl Output

```perl
$result = 42 + 3.14 + 1e5 + 0xFF;
```

### Typed Perl Output

```perl
$result = 42 + 3.14 + 1e5 + 0xFF;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Numeric Spaceship

### Clean Perl Output

```perl
$cmp_result = $a <=> $b;
```

### Typed Perl Output

```perl
$cmp_result = $a <=> $b;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Parenthesized Arithmetic

### Clean Perl Output

```perl
$result = ($a + $b) * ($c - $d) / ($e || 1);
```

### Typed Perl Output

```perl
$result = ($a + $b) * ($c - $d) / ($e || 1);
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Postfix Increment

### Clean Perl Output

```perl
$result = $array[$counter++] + $hash{$key++};
```

### Typed Perl Output

```perl
$result = $array[$counter++] + $hash{$key++};
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Prefix Decrement

### Clean Perl Output

```perl
$result = --$counter * ++$multiplier;
```

### Typed Perl Output

```perl
$result = --$counter * ++$multiplier;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Range Comparison

### Clean Perl Output

```perl
$in_range = ($min <= $value) && ($value <= $max);
```

### Typed Perl Output

```perl
$in_range = ($min <= $value) && ($value <= $max);
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Range Operator

### Clean Perl Output

```perl
@numbers = ($start .. $end);
```

### Typed Perl Output

```perl
@numbers = ($start .. $end);
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Reference Comparison

### Clean Perl Output

```perl
$same_ref = $ref1 == $ref2;
```

### Typed Perl Output

```perl
$same_ref = $ref1 == $ref2;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Reference Dereferencing

### Clean Perl Output

```perl
$sum = ${$scalar_ref} + @{$array_ref} + keys %{$hash_ref};
```

### Typed Perl Output

```perl
$sum = ${$scalar_ref} + @{$array_ref} + keys %{$hash_ref};
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Regex Match Expression

### Clean Perl Output

```perl
$valid = ($input =~ /^\d+$/) && ($email =~ /@\w+\.\w+$/);
```

### Typed Perl Output

```perl
$valid = ($input =~ /^\d+$/) && ($email =~ /@\w+\.\w+$/);
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Right Shift

### Clean Perl Output

```perl
$result = $value >> $positions;
```

### Typed Perl Output

```perl
$result = $value >> $positions;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Right Shift Assignment

### Clean Perl Output

```perl
$value >>= $positions;
```

### Typed Perl Output

```perl
$value >>= $positions;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Slice Expression

### Clean Perl Output

```perl
@subset = @array[$start .. $end];
```

### Typed Perl Output

```perl
@subset = @array[$start .. $end];
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## String Compare

### Clean Perl Output

```perl
$cmp_result = $left cmp $right;
```

### Typed Perl Output

```perl
$cmp_result = $left cmp $right;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## String Concatenation

### Clean Perl Output

```perl
$combined = $first . $second;
```

### Typed Perl Output

```perl
$combined = $first . $second;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## String Concatenation Assignment

### Clean Perl Output

```perl
$message .= $suffix;
```

### Typed Perl Output

```perl
$message .= $suffix;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## String Equality

### Clean Perl Output

```perl
$equal = $left eq $right;
```

### Typed Perl Output

```perl
$equal = $left eq $right;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## String Greater Equal

### Clean Perl Output

```perl
$greater_eq = $left ge $right;
```

### Typed Perl Output

```perl
$greater_eq = $left ge $right;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## String Greater Than

### Clean Perl Output

```perl
$greater = $left gt $right;
```

### Typed Perl Output

```perl
$greater = $left gt $right;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## String Inequality

### Clean Perl Output

```perl
$not_equal = $left ne $right;
```

### Typed Perl Output

```perl
$not_equal = $left ne $right;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## String Less Equal

### Clean Perl Output

```perl
$less_eq = $left le $right;
```

### Typed Perl Output

```perl
$less_eq = $left le $right;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## String Less Than

### Clean Perl Output

```perl
$less = $left lt $right;
```

### Typed Perl Output

```perl
$less = $left lt $right;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## String Literals

### Clean Perl Output

```perl
$result = 'single' . "double" . `backtick` . qq{quoted};
```

### Typed Perl Output

```perl
$result = 'single' . "double" . `backtick` . qq{quoted};
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## String Repetition

### Clean Perl Output

```perl
$repeated = $string x $count;
```

### Typed Perl Output

```perl
$repeated = $string x $count;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## String Repetition Assignment

### Clean Perl Output

```perl
$pattern x= $count;
```

### Typed Perl Output

```perl
$pattern x= $count;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Subtraction Assignment

### Clean Perl Output

```perl
$total -= $decrement;
```

### Typed Perl Output

```perl
$total -= $decrement;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Ternary Operator

### Clean Perl Output

```perl
$result = $condition ? $true_value : $false_value;
```

### Typed Perl Output

```perl
$result = $condition ? $true_value : $false_value;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Unary Minus

### Clean Perl Output

```perl
$result = -$value;
```

### Typed Perl Output

```perl
$result = -$value;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Unary Plus

### Clean Perl Output

```perl
$result = +$value;
```

### Typed Perl Output

```perl
$result = +$value;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Word And

### Clean Perl Output

```perl
$and_result = $a and $b;
```

### Typed Perl Output

```perl
$and_result = $a and $b;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Word Not

### Clean Perl Output

```perl
$not_result = not $condition;
```

### Typed Perl Output

```perl
$not_result = not $condition;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Word Or

### Clean Perl Output

```perl
$or_result = $x or $y;
```

### Typed Perl Output

```perl
$or_result = $x or $y;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
