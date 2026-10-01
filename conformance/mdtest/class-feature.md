# The `class` feature

`class`, `field`, `method` and `ADJUST`: four keywords that a reader
would call the subject of this tier and that produce, between them,
tier 05's `enterloop` and `leaveloop` plus `stub` and `methstart`.

**Tier 11 oo.** Introduces `anonhash`, `bless`, `emptyavhv`, `isa`,
`method`, `method_named`, `method_super`, `methstart`, `stub`,
`tie`, `tied`. Depends on 08_references.

This is the sharpest case in the corpus of a construct the optimiser
erases. Measured, neither `use feature 'class'` nor `no warnings
'experimental::class'` emits a runtime op, so the pragmas the syntax
requires cost this tier nothing and do not borrow tier 12's `use`. `use
v5.42;` is not enough on its own: it does not enable `feature 'class'`
in 5.42.0 and `class Foo` under it is a syntax error, so each case
names the feature. The constructor perl generates is XS code with no
Perl optree, so nothing this tier introduces can be measured through
it; `ref` on the result is what shows the class exists.

## An empty class body

`class Empty {}` is a complete, working class -- `Empty->new` returns
an object -- and it compiles to a bare block containing `stub`. There
is no `class` op:

    1  <0> enter v
    2  <;> nextstate(main 3 d.pl:3) v:%,{,fea=15
    3  <{> enterloop(next->5 last->5 redo->4) v
    4  <0> stub v
    5  <2> leaveloop vK/2

```perl
use feature 'class';
no warnings 'experimental::class';
class Empty {}
print ref(Empty->new), "\n";
```

```behavior
parses: yes
```

```output
Empty
```

## `field` and `method` in a class body

`field` and `method` emit no op of their own. A class body holding both
compiles to a bare block of `nextstate`s and nothing else:

    3  <{> enterloop(next->5 last->5 redo->4) v
    4  <;> nextstate(Foo 7 a.pl:4) v:%,us,*,&,{,$,fea=15,0x10
    5  <2> leaveloop vK/2

`field $x = 5;` contributes one `nextstate` and NO store: the
initialiser is compiled into a field-init tree that is not reachable
from the main optree. The method body is likewise its own CV, so the
`methstart` that opens it is not in the main stream -- naming it takes
`-exec,Foo::m`:

    Foo::m:
    1  <+> methstart() v
    2  <;> nextstate(Foo 10 a.pl:4) v:%,us,*,&,$,fea=15,0x10
    3  <0> padsv[$x:FAKE:] s
    4  <1> leavesub[1 ref] K/REFC,1

That is the finding, not an omission: the op stream of a file using
`class` says almost nothing about the class. `methstart` has no classic
counterpart -- a `sub` doing the same job opens with `shift`. Two
spellings of one object system, two disjoint prefixes.

```perl
use feature 'class';
no warnings 'experimental::class';
class Foo {
    field $x = 5;
    method m { $x }
}
my $o = Foo->new;
print $o->m, "\n";
```

```behavior
parses: yes
```

```output
5
```

## `ADJUST` alone in a class body

`ADJUST` runs after construction and emits no op of its own. It becomes
an anonymous sub in the class stash with no CODE slot, so B::Concise
cannot name it and its body is unreachable from any dump; what the main
optree shows is a `nextstate` where the block was, and the whole class
body is three ops:

    3  <{> enterloop(next->5 last->5 redo->4) v
    4  <;> nextstate(Foo 5 f.pl:4) v:%,{,fea=15
    5  <2> leaveloop vK/2

ADJUST sits here BY ITSELF on purpose, and it parses under our parser.
The adjacency body is the same construct with a `method` after it, and
the pair is the whole demonstration: one construct per file goes green
over an adjacency bug by construction, because every construct in such
a corpus is measured alone.

```perl
use feature 'class';
no warnings 'experimental::class';
class Foo {
    field $x = 1;
    ADJUST { $x = 2 }
}
print ref(Foo->new), "\n";
```

```behavior
parses: yes
```

```output
Foo
```

## A `:param` field defaulting to a folded boolean

`!!0` and `!!1` fold at compile time to perl's shared immortals `sv_no`
and `sv_yes`, which B reports as `B::SPECIAL` rather than an IV, NV or
PV. A reader that knows only those three drops the default, and the
`:param` becomes required: `Flags->new` with no arguments then dies.
Found by B::SoN translating chalk's lib/ and running chalk's own suite
against the emitted Perl, which compiled and gave a wrong answer.

```perl
use feature 'class';
no warnings 'experimental::class';
class Flags { field $quiet :param :reader = !!0; field $loud :param :reader = !!1; }
my $f = Flags->new;
print $f->quiet ? "q" : "-", $f->loud ? "l" : "-", Flags->new(quiet => 1)->quiet ? "Q" : "-", "\n";
```

```behavior
parses: yes
```

```output
-lQ
```

## A field post-increment whose value is used

`$count++` on a field must store the incremented value AND yield the
old one, here after a guarded early return that leaves the field
untouched. An implementation that only stores, or only yields, numbers
the keys wrongly. Found the same way as the case above.

```perl
use feature 'class';
no warnings 'experimental::class';
class Index {
    field %id_for;
    field $count = 0;
    method register ($key) {
        return $id_for{$key} if exists $id_for{$key};
        my $id = $count++;
        $id_for{$key} = $id;
        return $id;
    }
}
my $ix = Index->new;
print join(",", $ix->register("a"), $ix->register("b"), $ix->register("a"), $ix->register("c")), "\n";
```

```behavior
parses: yes
```

```output
0,1,0,2
```
