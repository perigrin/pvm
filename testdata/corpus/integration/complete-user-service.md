---
category: integration
subcategory: complete-programs
tags:
    - types
    - classes
    - roles
    - generics
    - constraints
    - inheritance
type_check: false
should_error: true
error_count: 8
min_lines: 40
---

# Complete User Service

Complete user service with types, classes, and roles

```perl
use v5.38;
use strict;
use warnings;

# Type definitions
type UserId = Int where { $_ > 0 };
type Email = Str where { $_ =~ /\@/ };
type Result<T, E> = Success<T> | Failure<E>;

# Role definitions
role Serializable {
    method Str serialize();
    method Self deserialize(Str $data);
}

role Cacheable<K> where K: Serializable {
    field Optional[DateTime] $cached_at;
    method K cache_key();
    method Bool is_stale();
}

# User class
class User does Serializable, Cacheable<UserId> {
    field UserId $id;
    field Str $name;
    field Email $email;
    field ArrayRef[Role] $roles = [];

    method User new(UserId $id, Str $name, Email $email) {
        return bless {
            id => $id,
            name => $name,
            email => $email,
            roles => []
        }, __PACKAGE__;
    }

    method Void add_role(Role $role) where $role->is_valid() {
        push @{$roles}, $role;
    }

    method Str serialize() {
        return encode_json({
            id => $id,
            name => $name,
            email => $email,
            roles => [map { $_->serialize() } @{$roles}]
        });
    }

    method UserId cache_key() {
        return $id;
    }
}

# Generic service class
class UserService<T> where T: User&Cacheable<UserId> {
    field HashRef[UserId, T] $cache = {};
    field CodeRef[UserId, Optional[T]] $loader;

    method UserService<T> new(CodeRef[UserId, Optional[T]] $loader) {
        return bless { cache => {}, loader => $loader }, __PACKAGE__;
    }

    method Result<T, Str> get(UserId $id) {
        if (exists $cache->{$id} && !$cache->{$id}->is_stale()) {
            return Success->new($cache->{$id});
        }

        my $user = $loader->($id);
        return Failure->new("User not found") unless defined $user;

        $cache->{$id} = $user;
        return Success->new($user);
    }
}
```

# Expected Compilation Outcomes

## Clean Perl

```perl
use v5.38;
use strict;
use warnings;

# Role definitions
{
    package Serializable;
    sub serialize;
    sub deserialize;
}

{
    package Cacheable;
    sub cache_key;
    sub is_stale;
}

# User class
{
    package User;
    use base qw(Serializable Cacheable);

    sub new {
        my ($class, $id, $name, $email) = @_;
        return bless {
            id => $id,
            name => $name,
            email => $email,
            roles => []
        }, $class;
    }

    sub add_role {
        my ($self, $role) = @_;
        push @{$self->{roles}}, $role;
    }

    sub serialize {
        my ($self) = @_;
        return encode_json({
            id => $self->{id},
            name => $self->{name},
            email => $self->{email},
            roles => [map { $_->serialize() } @{$self->{roles}}]
        });
    }

    sub cache_key {
        my ($self) = @_;
        return $self->{id};
    }
}

# Generic service class
{
    package UserService;

    sub new {
        my ($class, $loader) = @_;
        return bless { cache => {}, loader => $loader }, $class;
    }

    sub get {
        my ($self, $id) = @_;
        if (exists $self->{cache}->{$id} && !$self->{cache}->{$id}->is_stale()) {
            return Success->new($self->{cache}->{$id});
        }

        my $user = $self->{loader}->($id);
        return Failure->new("User not found") unless defined $user;

        $self->{cache}->{$id} = $user;
        return Success->new($user);
    }
}
```

## Typed Perl

```perl
use v5.38;
use strict;
use warnings;

# Type definitions
type UserId = Int where { $_ > 0 };
type Email = Str where { $_ =~ /\@/ };
type Result<T, E> = Success<T> | Failure<E>;

# Role definitions
role Serializable {
    method Str serialize();
    method Self deserialize(Str $data);
}

role Cacheable<K> where K: Serializable {
    field Optional[DateTime] $cached_at;
    method K cache_key();
    method Bool is_stale();
}

# User class
class User does Serializable, Cacheable<UserId> {
    field UserId $id;
    field Str $name;
    field Email $email;
    field ArrayRef[Role] $roles = [];

    method User new(UserId $id, Str $name, Email $email) {
        return bless {
            id => $id,
            name => $name,
            email => $email,
            roles => []
        }, __PACKAGE__;
    }

    method Void add_role(Role $role) where $role->is_valid() {
        push @{$roles}, $role;
    }

    method Str serialize() {
        return encode_json({
            id => $id,
            name => $name,
            email => $email,
            roles => [map { $_->serialize() } @{$roles}]
        });
    }

    method UserId cache_key() {
        return $id;
    }
}

# Generic service class
class UserService<T> where T: User&Cacheable<UserId> {
    field HashRef[UserId, T] $cache = {};
    field CodeRef[UserId, Optional[T]] $loader;

    method UserService<T> new(CodeRef[UserId, Optional[T]] $loader) {
        return bless { cache => {}, loader => $loader }, __PACKAGE__;
    }

    method Result<T, Str> get(UserId $id) {
        if (exists $cache->{$id} && !$cache->{$id}->is_stale()) {
            return Success->new($cache->{$id});
        }

        my $user = $loader->($id);
        return Failure->new("User not found") unless defined $user;

        $cache->{$id} = $user;
        return Success->new($user);
    }
}
```

## Inferred Perl

```perl
# Not implemented - type inference not yet available
```

# Expected Type Errors

```
(Type declarations not implemented in tree-sitter grammar)
```
