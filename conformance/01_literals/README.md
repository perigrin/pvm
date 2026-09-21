# 01_literals

Numbers, strings, quoting, `qw`.

## Why this tier sits here

Nothing precedes it. A literal is the smallest thing a program can
contain, and every later tier needs one: tier 02 must put a value in a
variable before it can say anything about variables, tier 04 must have
operands before it can have operators.

Check 2 asks whether this tier could move earlier. It could not — there is
no earlier.

## DEPENDS ON

    nothing

This is the only tier that can say that.

`my` and `print` appear in these files as FIXTURES rather than as subjects:
a literal has to be bound to something and observed somehow. `my` is tier
05's subject and `print` is tier 10's. That is the partial order working as
documented — using a construct is not the same as introducing it.

## INTRODUCES

    const enter leave multiconcat nextstate padrange padsv padsv_store print pushmark

## Why those ops, and not the ones the source implies

This list is what `perl -MO=Concise,-exec` EMITS for the files in this
tier, measured, not what reading them suggests. Two places where those
differ:

- **`multiconcat`, not `concat`.** `print "$x\n"` compiles to one
  multiconcat op; the peephole optimiser fuses the interpolation.
- **No `add`, though arithmetic appears nowhere here anyway.** Worth
  stating because `my $x = 1+2` would also emit no `add`: it arrives as
  `const[IV 3] s/FOLD`. The optimiser can erase the very construct a tier
  is about, which is why the ops LINT a declared tier and cannot derive
  one.

- **`padrange`, which no single-statement file emits.** It is the
  optimiser fusing consecutive `my` declarations into one op, so it appears
  only once a file declares several in a row -- which the adjacency file is
  the first here to do. Statement machinery, same class as `enter` and
  `leave`, and claimed for the same reason.

`enter`, `leave`, `nextstate` and `pushmark` are statement machinery that
every file in every tier emits. They are claimed here because this is the
first tier, and a tier claims an op once.
