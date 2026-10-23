# Op counts mined from t/perf/opcount.t

## L76: basic aelemfast

```perl
sub { our @a; $a[0] = 1 }
```

| op | count |
| --- | --- |
| aelem | 0 |
| aelemfast | 1 |
| ex-aelem | 1 |

## L90: bench.pl empty loop

```perl
sub { for my $x (1..$ARGV[0]) { 1; } }
```

| op | count |
| --- | --- |
| aelemfast | 1 |
| and | 1 |
| const | 1 |
| enteriter | 1 |
| iter | 1 |
| leaveloop | 1 |
| leavesub | 1 |
| lineseq | 2 |
| nextstate | 2 |
| null | 1 |
| pushmark | 1 |
| unstack | 1 |

## L109: bench.pl active loop

```perl prelude
no warnings 'void';
```

```perl
sub { for my $x (1..$ARGV[0]) { $x; } }
```

| op | count |
| --- | --- |
| aelemfast | 1 |
| and | 1 |
| const | 1 |
| enteriter | 1 |
| iter | 1 |
| leaveloop | 1 |
| leavesub | 1 |
| lineseq | 2 |
| nextstate | 2 |
| null | 1 |
| padsv | 1 |
| pushmark | 1 |
| unstack | 1 |

## L196: multideref expressions

```perl
sub { ($_[0] // $_)->[0]{2*$_[0]} }
```

| op | count |
| --- | --- |
| aelemfast | 2 |
| helem | 1 |
| multideref | 1 |

## L208: multideref const index

```perl
sub { $_->{1}{1.1} }
```

| op | count |
| --- | --- |
| helem | 0 |
| multideref | 1 |

## L217: multideref undef const index

```perl prelude
use constant my_undef => undef;
```

```perl
sub { $_->{+my_undef} }
```

| op | count |
| --- | --- |
| helem | 1 |
| multideref | 0 |

## L227: multideref op_other etc

```perl
sub { $_{foo} = $_ ? $_{bar} : $_{baz} }
```

| op | count |
| --- | --- |
| helem | 0 |
| multideref | 3 |

## L241: multideref no hints

```perl prelude
no strict;
no warnings;
```

```perl
sub { $_{foo}[0] }
```

| op | count |
| --- | --- |
| aelem | 0 |
| helem | 0 |
| multideref | 1 |

## L253: multideref exists

```perl
sub { exists $_[0] }
```

| op | count |
| --- | --- |
| aelem | 0 |
| aelemfast | 0 |
| multideref | 1 |

## L272: signature default expressions get optimised

```perl prelude
use feature 'signatures';

my @a;
```

```perl
sub ($s = $a[0]) {}
```

| op | count |
| --- | --- |
| aelem | 0 |
| aelemfast_lex | 1 |

## L286: in-place sort of global

```perl prelude
local our @global = (3,2,1);
my @lex = qw(a b c);
```

```perl
sub { @global = sort @global; 1 }
```

| op | count |
| --- | --- |
| aassign | 0 |
| rv2av | 1 |

## L293: in-place sort of lexical

```perl prelude
local our @global = (3,2,1);
my @lex = qw(a b c);
```

```perl
sub { @lex = sort @lex; 1 }
```

| op | count |
| --- | --- |
| aassign | 0 |
| padav | 1 |

## L300: in-place reversed sort of global

```perl prelude
local our @global = (3,2,1);
my @lex = qw(a b c);
```

```perl
sub { @global = sort { $b <=> $a } @global; 1 }
```

| op | count |
| --- | --- |
| aassign | 0 |
| rv2av | 1 |

## L308: in-place custom sort of global

```perl prelude
local our @global = (3,2,1);
my @lex = qw(a b c);
```

```perl
sub { @global = sort {  $a<$b?1:$a>$b?-1:0 } @global; 1 }
```

| op | count |
| --- | --- |
| aassign | 0 |
| rv2av | 1 |

## L316: in-place sort with function of lexical

```perl prelude
local our @global = (3,2,1);
my @lex = qw(a b c);

sub mysort { $b cmp $a };
```

```perl
sub { @lex = sort mysort @lex; 1 }
```

| op | count |
| --- | --- |
| aassign | 0 |
| padav | 1 |

## L407: sprintf pessimised

```perl prelude
my ($i,$j,$s);
```

```perl
sub { $s .= sprintf "%d%d",$i, $j }
```

| op | count |
| --- | --- |
| concat | 0 |
| const | 1 |
| multiconcat | 1 |
| padsv | 2 |
| sprintf | 1 |

## L421: sprintf constant args

```perl
sub { sprintf "%s%s", "abc", "def" }
```

| op | count |
| --- | --- |
| const | 1 |
| multiconcat | 0 |
| sprintf | 0 |

## L567: $lex1 = $lex2 . $lex1

```perl
sub { my ($x,$y); $x = $y . $x }
```

| op | count |
| --- | --- |
| concat | 0 |
| multiconcat | 1 |
| padsv | 4 |
| sassign | 0 |
| stringify | 0 |

## L575: $lex1 = "$lex2$lex1"

```perl
sub { my ($x,$y); $x = "$y$x" }
```

| op | count |
| --- | --- |
| concat | 0 |
| multiconcat | 1 |
| padsv | 4 |
| sassign | 0 |
| stringify | 0 |

## L589: my $a .= $b.$c.$d

```perl
sub { our ($b,$c,$d); my $a .= $b.$c.$d }
```

| op | count |
| --- | --- |
| padsv | 1 |

## L596: rcatline

```perl
sub { my ($x,$y); open FOO, "xxx"; $x .= <FOO> }
```

| op | count |
| --- | --- |
| concat | 0 |
| multiconcat | 0 |
| rcatline | 1 |
| readline | 0 |

## L625: state works with multiconcat

```perl
sub { use feature 'state'; our ($a, $b, $c); state $s = $a . $b . $c }
```

| op | count |
| --- | --- |
| concat | 0 |
| multiconcat | 1 |
| once | 1 |
| padsv | 2 |
| sassign | 0 |

## L639: multiconcat: 2 adjacent consts

```perl
sub { my ($a, $b); $a = $b . "c" . "d" }
```

| op | count |
| --- | --- |
| concat | 0 |
| const | 1 |
| multiconcat | 1 |
| sassign | 0 |

## L647: multiconcat: 3 adjacent consts

```perl
sub { my ($a, $b); $a = $b . "c" . "d" . "e" }
```

| op | count |
| --- | --- |
| concat | 0 |
| const | 1 |
| multiconcat | 1 |
| sassign | 0 |

## L655: multiconcat: 4 adjacent consts

```perl
sub { my ($a, $b); $a = $b . "c" . "d" . "e" ."f" }
```

| op | count |
| --- | --- |
| concat | 0 |
| const | 2 |
| multiconcat | 1 |
| sassign | 0 |

## L666: multiconcat: local assign

```perl
sub { our $global; local $global = "$global-X" }
```

| op | count |
| --- | --- |
| concat | 0 |
| const | 0 |
| gvsv | 2 |
| multiconcat | 1 |
| sassign | 1 |

## L679: try/catch: catch block is optimized

```perl prelude
use feature 'try';
```

```perl
sub { my @a; try {} catch($e) { $a[0] } }
```

| op | count |
| --- | --- |
| aelem | 0 |
| aelemfast_lex | 1 |

## L691: pushdefer: block is optimized

```perl prelude
use feature 'defer';
no warnings 'experimental::defer';
```

```perl
sub { my @a; defer { $a[0] } }
```

| op | count |
| --- | --- |
| aelem | 0 |
| aelemfast_lex | 1 |

## L717: builtin::true/false are replaced with constants

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { my $x = builtin::true(); my $y = builtin::false() }
```

| op | count |
| --- | --- |
| const | 2 |
| entersub | 0 |

## L724: imported true/false are replaced with constants

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { my $x = true(); my $y = false() }
```

| op | count |
| --- | --- |
| const | 2 |
| entersub | 0 |

## L731: builtin::is_bool is replaced with direct opcode

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { my $x; my $y; $y = builtin::is_bool($x); 1; }
```

| op | count |
| --- | --- |
| entersub | 0 |
| is_bool | 1 |
| padsv | 3 |
| padsv_store | 1 |

## L740: imported is_bool is replaced with direct opcode

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { my $x; my $y; $y = is_bool($x); 1; }
```

| op | count |
| --- | --- |
| entersub | 0 |
| is_bool | 1 |
| padsv | 3 |
| padsv_store | 1 |

## L749: builtin::is_bool gets constant-folded

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { builtin::is_bool(123); }
```

| op | count |
| --- | --- |
| const | 1 |
| entersub | 0 |
| is_bool | 0 |

## L757: imported is_bool gets constant-folded

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { is_bool(123); }
```

| op | count |
| --- | --- |
| const | 1 |
| entersub | 0 |
| is_bool | 0 |

## L765: builtin::weaken is replaced with direct opcode

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { my $x = []; builtin::weaken($x); }
```

| op | count |
| --- | --- |
| entersub | 0 |
| weaken | 1 |

## L772: imported weaken is replaced with direct opcode

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { my $x = []; weaken($x); }
```

| op | count |
| --- | --- |
| entersub | 0 |
| weaken | 1 |

## L779: builtin::unweaken is replaced with direct opcode

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { my $x = []; builtin::unweaken($x); }
```

| op | count |
| --- | --- |
| entersub | 0 |
| unweaken | 1 |

## L786: imported unweaken is replaced with direct opcode

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { my $x = []; unweaken($x); }
```

| op | count |
| --- | --- |
| entersub | 0 |
| unweaken | 1 |

## L793: builtin::is_weak is replaced with direct opcode

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { builtin::is_weak([]); }
```

| op | count |
| --- | --- |
| entersub | 0 |
| is_weak | 1 |

## L800: imported is_weak is replaced with direct opcode

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { is_weak([]); }
```

| op | count |
| --- | --- |
| entersub | 0 |
| is_weak | 1 |

## L807: builtin::blessed is replaced with direct opcode

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { builtin::blessed([]); }
```

| op | count |
| --- | --- |
| blessed | 1 |
| entersub | 0 |

## L814: imported blessed is replaced with direct opcode

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { blessed([]); }
```

| op | count |
| --- | --- |
| blessed | 1 |
| entersub | 0 |

## L821: builtin::refaddr is replaced with direct opcode

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { builtin::refaddr([]); }
```

| op | count |
| --- | --- |
| entersub | 0 |
| refaddr | 1 |

## L828: imported refaddr is replaced with direct opcode

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { refaddr([]); }
```

| op | count |
| --- | --- |
| entersub | 0 |
| refaddr | 1 |

## L835: builtin::reftype is replaced with direct opcode

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { builtin::reftype([]); }
```

| op | count |
| --- | --- |
| entersub | 0 |
| reftype | 1 |

## L842: imported reftype is replaced with direct opcode

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { reftype([]); }
```

| op | count |
| --- | --- |
| entersub | 0 |
| reftype | 1 |

## L850: builtin::ceil is replaced with direct opcode

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);

my $one_point_five = 1.5;   # Prevent const-folding.
```

```perl
sub { builtin::ceil($one_point_five); }
```

| op | count |
| --- | --- |
| ceil | 1 |
| entersub | 0 |

## L857: imported ceil is replaced with direct opcode

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);

my $one_point_five = 1.5;   # Prevent const-folding.
```

```perl
sub { ceil($one_point_five); }
```

| op | count |
| --- | --- |
| ceil | 1 |
| entersub | 0 |

## L864: builtin::floor is replaced with direct opcode

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);

my $one_point_five = 1.5;   # Prevent const-folding.
```

```perl
sub { builtin::floor($one_point_five); }
```

| op | count |
| --- | --- |
| entersub | 0 |
| floor | 1 |

## L871: imported floor is replaced with direct opcode

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);

my $one_point_five = 1.5;   # Prevent const-folding.
```

```perl
sub { floor($one_point_five); }
```

| op | count |
| --- | --- |
| entersub | 0 |
| floor | 1 |

## L878: builtin::is_tainted is replaced with direct opcode

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { builtin::is_tainted($0); }
```

| op | count |
| --- | --- |
| entersub | 0 |
| is_tainted | 1 |

## L885: imported is_tainted is replaced with direct opcode

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { is_tainted($0); }
```

| op | count |
| --- | --- |
| entersub | 0 |
| is_tainted | 1 |

## L893: sassign + padsv replaced by padsv_store

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { my $y; my $z = $y = 3; 1; }
```

| op | count |
| --- | --- |
| padsv | 2 |
| padsv_store | 1 |

## L901: undef + padsv (undef my \$x) is reduced to undef

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { undef my $x }
```

| op | count |
| --- | --- |
| padsv | 0 |
| padsv_store | 0 |
| sassign | 0 |
| undef | 1 |

## L909: undef + padsv + sassign (my \$x = undef) is reduced to undef

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { my $x = undef }
```

| op | count |
| --- | --- |
| padsv | 0 |
| padsv_store | 0 |
| sassign | 0 |
| undef | 1 |

## L917: undef + padsv (undef \$x) is reduced to undef

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { my $x; undef $x }
```

| op | count |
| --- | --- |
| padsv | 1 |
| padsv_store | 0 |
| sassign | 0 |
| undef | 1 |

## L925: undef + padsv + sassign (\$x = undef) is reduced to undef

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { my $x; $x = undef }
```

| op | count |
| --- | --- |
| padsv | 1 |
| padsv_store | 0 |
| sassign | 0 |
| undef | 1 |

## L934: my $y= 1; my @x= ($y= undef);

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { my $y= 1; my @x= ($y= undef); }
```

| op | count |
| --- | --- |
| aassign | 1 |
| padav | 1 |
| padsv | 0 |
| padsv_store | 1 |
| sassign | 0 |
| undef | 1 |

## L945: my $x= 1; sub f{} f($x=undef);

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { my $x= 1; sub f{} f($x=undef); }
```

| op | count |
| --- | --- |
| gv | 1 |
| padsv | 0 |
| padsv_store | 1 |
| sassign | 0 |
| undef | 1 |

## L955: my ($x,$p)=(1,2); sub g{} g(($x=undef),$p);

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { my ($x,$p)=(1,2); sub g{} g(($x=undef),$p); }
```

| op | count |
| --- | --- |
| aassign | 1 |
| gv | 1 |
| padrange | 1 |
| padsv | 3 |
| padsv_store | 0 |
| sassign | 0 |
| undef | 1 |

## L967: my $h= {}; my @k= keys %{($h=undef)||{}};

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { my $h= {}; my @k= keys %{($h=undef)||{}}; }
```

| op | count |
| --- | --- |
| aassign | 1 |
| emptyavhv | 2 |
| padav | 1 |
| padsv | 0 |
| padsv_store | 0 |
| sassign | 0 |
| undef | 1 |

## L979: my $y= 1; my @x= \($y= undef);

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { my $y= 1; my @x= \($y= undef); }
```

| op | count |
| --- | --- |
| aassign | 1 |
| padav | 1 |
| padsv | 0 |
| padsv_store | 1 |
| sassign | 0 |
| srefgen | 1 |
| undef | 1 |

## L992: simple aelemfast_lex + sassign replacement

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { my @x; $x[0] = "foo"; 1 }
```

| op | count |
| --- | --- |
| aelemfast_lex | 0 |
| aelemfastlex_store | 1 |
| padav | 1 |
| sassign | 0 |

## L1003: aelemfast_lex + sassign replacement with neg key

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { my @x = (1,2); $x[-1] = 7; 1 }
```

| op | count |
| --- | --- |
| aelemfast_lex | 0 |
| aelemfastlex_store | 1 |
| padav | 1 |
| sassign | 0 |

## L1013: no aelemfast_lex + sassign replacement with multideref

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { my @x = ([1,2]); $x[0][1] = 1; }
```

| op | count |
| --- | --- |
| aelemfast_lex | 0 |
| aelemfastlex_store | 0 |
| multideref | 1 |
| padav | 1 |
| sassign | 1 |

## L1025: Empty anonlist

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { [] }
```

| op | count |
| --- | --- |
| anonlist | 0 |
| emptyavhv | 1 |
| sassign | 0 |

## L1032: Empty anonlist with global assignment

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { our $x; $x = [] }
```

| op | count |
| --- | --- |
| anonlist | 0 |
| emptyavhv | 1 |
| gvsv | 1 |
| pushmark | 0 |
| sassign | 1 |

## L1041: Empty anonlist and lexical assignment

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { my $x; $x = [] }
```

| op | count |
| --- | --- |
| anonlist | 0 |
| emptyavhv | 1 |
| padsv | 1 |
| pushmark | 0 |
| sassign | 0 |

## L1050: Empty anonlist and direct lexical assignment

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { my $x = [] }
```

| op | count |
| --- | --- |
| anonlist | 0 |
| emptyavhv | 1 |
| padsv | 0 |
| pushmark | 0 |
| sassign | 0 |

## L1059: Empty anonlist ref and direct lexical assignment

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { my $x = \[]; 1; }
```

| op | count |
| --- | --- |
| anonlist | 0 |
| emptyavhv | 1 |
| padsv | 0 |
| padsv_store | 1 |
| pushmark | 0 |
| sassign | 0 |
| srefgen | 1 |

## L1070: Empty anonhash

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { {} }
```

| op | count |
| --- | --- |
| anonhash | 0 |
| emptyavhv | 1 |
| sassign | 0 |

## L1077: Empty anonhash with global assignment

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { our $x; $x = {} }
```

| op | count |
| --- | --- |
| anonhash | 0 |
| emptyavhv | 1 |
| gvsv | 1 |
| pushmark | 0 |
| sassign | 1 |

## L1086: Empty anonhash and lexical assignment

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { my $x; $x = {} }
```

| op | count |
| --- | --- |
| anonhash | 0 |
| emptyavhv | 1 |
| padsv | 1 |
| pushmark | 0 |
| sassign | 0 |

## L1095: Empty anonhash and direct lexical assignment

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { my $x = {} }
```

| op | count |
| --- | --- |
| anonhash | 0 |
| emptyavhv | 1 |
| padsv | 0 |
| pushmark | 0 |
| sassign | 0 |

## L1104: Empty anonhash ref and direct lexical assignment

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { my $x = \{}; 1 }
```

| op | count |
| --- | --- |
| anonhash | 0 |
| emptyavhv | 1 |
| padsv | 0 |
| padsv_store | 1 |
| pushmark | 0 |
| sassign | 0 |
| srefgen | 1 |

## L1116: foreach 2 lexicals on builtin::indexed ARRAY

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { my @input = (); foreach my ($i, $x) (builtin::indexed @input) { } }
```

| op | count |
| --- | --- |
| enteriter | 1 |
| entersub | 0 |
| iter | 1 |
| padav | 2 |

## L1125: foreach 2 lexicals on imported indexed ARRAY

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { my @input = (); foreach my ($i, $x) (indexed @input) { } }
```

| op | count |
| --- | --- |
| enteriter | 1 |
| entersub | 0 |
| iter | 1 |
| padav | 2 |

## L1134: foreach 2 lexicals on builtin::indexed LIST

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { foreach my ($i, $x) (builtin::indexed qw( x y z )) { } }
```

| op | count |
| --- | --- |
| enteriter | 1 |
| entersub | 0 |
| iter | 1 |

## L1142: foreach 2 lexicals on imported indexed LIST

```perl prelude
no warnings 'experimental::builtin';
use builtin qw(
    blessed
    ceil
    false
    floor
    indexed
    is_bool
    is_tainted
    is_weak
    refaddr
    reftype
    true
    unweaken
    weaken
);
```

```perl
sub { foreach my ($i, $x) (indexed qw( x y z )) { } }
```

| op | count |
| --- | --- |
| enteriter | 1 |
| entersub | 0 |
| iter | 1 |

## L1152: Zero-arg empty subroutine uses OP_MULTIPARAM

```perl prelude
use feature qw( signatures );
```

```perl
sub () {}
```

| op | count |
| --- | --- |
| argcheck | 0 |
| multiparam | 1 |

## L1159: Zero-arg subroutine uses OP_MULTIPARAM

```perl prelude
use feature qw( signatures );
```

```perl
sub () { return; }
```

| op | count |
| --- | --- |
| argcheck | 0 |
| multiparam | 1 |

## L1166: Two-arg subroutine uses OP_MULTIPARAM

```perl prelude
use feature qw( signatures );
```

```perl
sub ($x, $y) { return; }
```

| op | count |
| --- | --- |
| argcheck | 0 |
| argelem | 0 |
| multiparam | 1 |

## L1174: Two-arg one-optional subroutine uses OP_MULTIPARAM

```perl prelude
use feature qw( signatures );
```

```perl
sub ($x, $y = "default") { return; }
```

| op | count |
| --- | --- |
| argcheck | 0 |
| argelem | 0 |
| multiparam | 1 |

## L1182: Two-arg one-anon subroutine uses OP_MULTIPARAM

```perl prelude
use feature qw( signatures );
```

```perl
sub ($, $y) { return; }
```

| op | count |
| --- | --- |
| argcheck | 0 |
| argelem | 0 |
| multiparam | 1 |

## L1190: One-arg plus slurpy array subroutine uses OP_MULTIPARAM

```perl prelude
use feature qw( signatures );
```

```perl
sub ($x, @rest) { return; }
```

| op | count |
| --- | --- |
| argcheck | 0 |
| argelem | 0 |
| multiparam | 1 |

## L1198: One-arg plus slurpy hash subroutine uses OP_MULTIPARAM

```perl prelude
use feature qw( signatures );
```

```perl
sub ($x, %rest) { return; }
```

| op | count |
| --- | --- |
| argcheck | 0 |
| argelem | 0 |
| multiparam | 1 |

## L1208: substr with const zero offset and '' repl (void)

```perl
sub { my $z; substr($z, 0, 2, "") }
```

| op | count |
| --- | --- |
| const | 1 |
| substr | 0 |
| substr_left | 1 |

## L1216: substr with const zero offset and '' repl (lexical)

```perl
sub { my $z; my $x = substr($z, 0, 2, "") }
```

| op | count |
| --- | --- |
| const | 1 |
| padsv | 3 |
| sassign | 1 |
| substr | 0 |
| substr_left | 1 |

## L1226: substr with const zero offset and '' repl (lexical TARGMY)

```perl
sub { my ($z, $x); $x = substr($z, 0, 2, "") }
```

| op | count |
| --- | --- |
| const | 1 |
| padsv | 3 |
| padsv_store | 0 |
| sassign | 0 |
| substr | 0 |
| substr_left | 1 |

## L1237: substr with const zero offset and '' repl (gv)

```perl
sub { my $z; our $x = substr($z, 0, 2, "") }
```

| op | count |
| --- | --- |
| const | 1 |
| gvsv | 1 |
| sassign | 1 |
| substr | 0 |
| substr_left | 1 |

## L1247: substr with const zero offset (void)

```perl
sub { my $z; substr($z, 0, 2) }
```

| op | count |
| --- | --- |
| const | 1 |
| substr | 0 |
| substr_left | 1 |

## L1255: substr with const zero offset (lexical)

```perl
sub { my $z; my $x = substr($z, 0, 2) }
```

| op | count |
| --- | --- |
| const | 1 |
| padsv | 3 |
| sassign | 1 |
| substr | 0 |
| substr_left | 1 |

## L1265: substr with const zero offset (lexical TARGMY)

```perl
sub { my ($z, $x); $x = substr($z, 0, 2) }
```

| op | count |
| --- | --- |
| const | 1 |
| padsv | 3 |
| sassign | 0 |
| substr | 0 |
| substr_left | 1 |

## L1275: substr with const zero offset  (gv)

```perl
sub { my $z; our $x = substr($z, 0, 2) }
```

| op | count |
| --- | --- |
| const | 1 |
| gvsv | 1 |
| sassign | 1 |
| substr | 0 |
| substr_left | 1 |

## L1320: basic any/all

```perl
sub {
                    use feature      'keyword_any', 'keyword_all';
                    use experimental 'keyword_any', 'keyword_all';
                    my (@a, @x, @b);
                    @a = any { $x[0] } @b;
                    @a = all { $x[0] } @b;
                }
```

| op | count |
| --- | --- |
| aelem | 0 |
| aelemfast_lex | 2 |
| ex-aelem | 2 |

## L1336: Don't fold string repetition above a multiplier threshold

```perl
sub { my $x = "A" x (2**22) }
```

| op | count |
| --- | --- |
| repeat | 1 |
