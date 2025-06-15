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
