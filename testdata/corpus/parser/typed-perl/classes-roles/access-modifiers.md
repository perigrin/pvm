---
category: typed-perl
subcategory: classes-roles
tags:
    - access-modifiers
    - field-visibility
    - private-methods
    - protected-methods
    - public-methods
    - readonly-fields
type_check: true
---

# Access Modifiers Visibility

Class with access modifiers and field visibility

```perl
class BankAccount {
    field private Num $balance = 0.0;
    field protected Str $account_number;
    field public Str $account_holder;
    field readonly DateTime $created_at;

    method new(Str $holder, Str $number) returns BankAccount {
        return bless {
            account_holder => $holder,
            account_number => $number,
            balance => 0.0,
            created_at => DateTime->now()
        }, __PACKAGE__;
    }

    method private validate_amount(Num $amount) returns Bool {
        return $amount > 0;
    }

    method public deposit(Num $amount) returns Bool {
        return 0 unless $self->validate_amount($amount);
        $balance += $amount;
        return 1;
    }

    method public get_balance() returns Num {
        return $balance;
    }

    method protected get_account_number() returns Str {
        return $account_number;
    }
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 880 characters
  Type Annotations:
    MethodReturnAnnotation: new :: BankAccount at 7:50
    MethodReturnAnnotation: validate_amount :: Bool at 16:57
    MethodReturnAnnotation: deposit :: Bool at 20:48
    MethodReturnAnnotation: get_balance :: Num at 26:41
    MethodReturnAnnotation: get_account_number :: Str at 30:51
    VarAnnotation: BankAccount :: class at 1:1
    VarAnnotation: $balance :: private at 2:5
    VarAnnotation: $account_number :: protected at 3:5
    VarAnnotation: $account_holder :: public at 4:5
    VarAnnotation: $created_at :: readonly at 5:5
    MethodReturnAnnotation: new :: BankAccount at 7:50
    MethodReturnAnnotation: validate_amount :: Bool at 16:57
    MethodReturnAnnotation: deposit :: Bool at 20:48
    MethodReturnAnnotation: get_balance :: Num at 26:41
    MethodReturnAnnotation: get_account_number :: Str at 30:51
    MethodParamAnnotation: $holder :: Str at 7:1
    MethodParamAnnotation: $number :: Str at 7:1
    MethodParamAnnotation: $amount :: Num at 16:1
    MethodParamAnnotation: $amount :: Num at 20:1
  Root: source_file
  Tree Structure:
  source_file
    class_decl
      field_decl
        variable
      field_decl
        variable
      field_decl
        variable
      method_decl
        block_stmt
          token
          expression_stmt
            literal
          token
          token
      method_decl
        block_stmt
          token
          expression_stmt
            literal
          token
          token
      method_decl
        block_stmt
          token
          expression_stmt
            literal
          token
          expression_stmt
            literal
          token
          expression_stmt
            literal
          token
          token
      method_decl
        block_stmt
          token
          expression_stmt
            literal
          token
          token
      method_decl
        block_stmt
          token
          expression_stmt
            literal
          token
          token
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 880 characters
  Type Annotations:
    MethodReturnAnnotation: new :: BankAccount at 7:50
    MethodReturnAnnotation: validate_amount :: Bool at 16:57
    MethodReturnAnnotation: deposit :: Bool at 20:48
    MethodReturnAnnotation: get_balance :: Num at 26:41
    MethodReturnAnnotation: get_account_number :: Str at 30:51
    VarAnnotation: BankAccount :: class at 1:1
    VarAnnotation: $balance :: private at 2:5
    VarAnnotation: $account_number :: protected at 3:5
    VarAnnotation: $account_holder :: public at 4:5
    VarAnnotation: $created_at :: readonly at 5:5
    MethodReturnAnnotation: new :: BankAccount at 7:50
    MethodReturnAnnotation: validate_amount :: Bool at 16:57
    MethodReturnAnnotation: deposit :: Bool at 20:48
    MethodReturnAnnotation: get_balance :: Num at 26:41
    MethodReturnAnnotation: get_account_number :: Str at 30:51
    MethodParamAnnotation: $holder :: Str at 7:1
    MethodParamAnnotation: $number :: Str at 7:1
    MethodParamAnnotation: $amount :: Num at 16:1
    MethodParamAnnotation: $amount :: Num at 20:1
  Root: source_file
  Tree Structure:
  source_file
    class_decl
      field_decl
        variable
      field_decl
        variable
      field_decl
        variable
      method_decl
        block_stmt
          token
          expression_stmt
            literal
          token
          token
      method_decl
        block_stmt
          token
          expression_stmt
            literal
          token
          token
      method_decl
        block_stmt
          token
          expression_stmt
            literal
          token
          expression_stmt
            literal
          token
          expression_stmt
            literal
          token
          token
      method_decl
        block_stmt
          token
          expression_stmt
            literal
          token
          token
      method_decl
        block_stmt
          token
          expression_stmt
            literal
          token
          token
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
{ return bless {
            account_holder => $holder,
            account_number => $number,
            balance => 0.0,
            created_at => DateTime->now()
        }, __PACKAGE__; }{ return $amount > 0; }{ return 0 unless $self->validate_amount($amount); $balance += $amount; return 1; }{ return $balance; }{ return $account_number; }
```

## Typed Perl Output

```perl
class BankAccount {
    field private Num $balance = 0.0;
    field protected Str $account_number;
    field public Str $account_holder;
    field readonly DateTime $created_at;

    method new(Str $holder, Str $number) returns BankAccount {
        return bless {
            account_holder => $holder,
            account_number => $number,
            balance => 0.0,
            created_at => DateTime->now()
        }, __PACKAGE__;
    }

    method private validate_amount(Num $amount) returns Bool {
        return $amount > 0;
    }

    method public deposit(Num $amount) returns Bool {
        return 0 unless $self->validate_amount($amount);
        $balance += $amount;
        return 1;
    }

    method public get_balance() returns Num {
        return $balance;
    }

    method protected get_account_number() returns Str {
        return $account_number;
    }
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
