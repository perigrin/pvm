---
category: untyped-perl
subcategory: variables
tags:
    - edge-cases
    - special-vars
    - variables
should_error: true
expected_error: error[TSP001]
---

# Variable Edge Cases

Test edge cases like empty declarations and special variable names

```perl
my $;
my @;
my %;
my $$;
my $0;
my $var = undef;
my @empty = ();
my %empty = ();
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
# Compilation failed: Error: failed to parse file /tmp/tmp5g8pfaaa.pl: SYS-007: error[TSP001]: parse error (2 ERROR nodes detected)
  --> :26:2
   |
26 | my %;
   |  ^^^^ unexpected token: ''

  --> :27:6
   |
27 | my $$;
   |      ^ unexpected token: ''

note: This indicates Perl syntax that is not yet supported by the tree-sitter grammar.
      Please add test cases for this syntax to improve parser coverage.
 (System Error)
2025-06-25T05:33:01Z [ERROR] [psc] Error: failed to parse file /tmp/tmp5g8pfaaa.pl: SYS-007: error[TSP001]: parse error (2 ERROR nodes detected)
  --> :26:2
   |
26 | my %;
   |  ^^^^ unexpected token: ''

  --> :27:6
   |
27 | my $$;
   |      ^ unexpected token: ''

note: This indicates Perl syntax that is not yet supported by the tree-sitter grammar.
      Please add test cases for this syntax to improve parser coverage.
 (System Error)
```

## Typed Perl Output

```perl
# Compilation failed: Error: failed to parse file /tmp/tmp5g8pfaaa.pl: SYS-007: error[TSP001]: parse error (2 ERROR nodes detected)
  --> :26:2
   |
26 | my %;
   |  ^^^^ unexpected token: ''

  --> :27:6
   |
27 | my $$;
   |      ^ unexpected token: ''

note: This indicates Perl syntax that is not yet supported by the tree-sitter grammar.
      Please add test cases for this syntax to improve parser coverage.
 (System Error)
2025-06-25T05:33:01Z [ERROR] [psc] Error: failed to parse file /tmp/tmp5g8pfaaa.pl: SYS-007: error[TSP001]: parse error (2 ERROR nodes detected)
  --> :26:2
   |
26 | my %;
   |  ^^^^ unexpected token: ''

  --> :27:6
   |
27 | my $$;
   |      ^ unexpected token: ''

note: This indicates Perl syntax that is not yet supported by the tree-sitter grammar.
      Please add test cases for this syntax to improve parser coverage.
 (System Error)
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
