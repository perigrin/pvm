# A read between a mutation and its result

Every case here is one destructive quote-like whose count is bound. What
differs is WHERE the subject is read -- in the same statement as the
count, or in a statement between the mutation and the count's use.

**Tier 09 regex.** Introduces `trans`. Depends on 01_literals.

THE TIER'S OTHER CASES PAIR CONSTRUCTS. This one pairs a construct with a
READ POSITION, which is a different axis and the reason it is here rather
than folded into the adjacency case. A destructive `tr///` and a
destructive `s///` each do two things -- change the subject, and yield a
count -- and an implementation that renders the mutation wherever its
value is first read is correct on every body that reads both in one
statement and wrong on every body that does not.

The property is not the operator's. Any op that mutates a container and
yields a value has it; `shift @q` is the same shape on an array. This tier
can only write the quote-like forms, so that is what these cases are, and
the generalisation is the point rather than the spelling. (`shift` is
claimed by tier 11 today, which is tracked as a grading defect in
`01a0dde3`; when it moves to tier 02 the array spelling belongs there.)

Contributed by the B::SoN session, which found the property by reverting a
fix whose three passing assertions all lived in the one-statement shape.
Every output below was verified against perl 5.42.0 on arrival.

## Both reads in one statement

The control. A mutation and its count read together cannot distinguish
"emitted where it happened" from "emitted where its value was wanted" --
there is only one position. A case like this passes under the defect the
next two cases catch, which is what makes it worth writing down: it
establishes that the `tr///` itself is right before the position is asked
about.

```perl
my $t = "a.c";
my $c = ($t =~ tr/./Z/);
print "both $t $c\n";
```

```behavior
parses: yes
```

```output
both aZc 1
```

## A read between the `tr///` and its count

The subject is read in a statement of its own, before the count is used.
That read must see the transliterated string, because the mutation has
already happened -- the binding of `$c` is where it happened, not where
`$c` is wanted.

An implementation that defers the `tr///` to the second print prints
`mid a.c` here and `end 1` after it. Both halves are individually
plausible: the count is right, and `a.c` is what `$t` held a statement
earlier. Only the pairing is wrong.

```perl
my $t = "a.c";
my $c = ($t =~ tr/./Z/);
print "mid $t\n";
print "end $c\n";
```

```behavior
parses: yes
```

```output
mid aZc
end 1
```

## The same read between an `s///` and its count

`s///` is the second two-region quote-like and has the identical shape: a
destructive form that mutates the subject and yields a count. A tier that
caught the `tr///` case and not this one would have fixed a spelling
rather than a property.

```perl
my $s = "aaa";
my $n = ($s =~ s/a/b/g);
print "mid $s\n";
print "end $n\n";
```

```behavior
parses: yes
```

```output
mid bbb
end 3
```
