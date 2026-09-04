# ABOUTME: Corpus for measuring PSC inference precision against observed runtime types.
# ABOUTME: Each __observe call records what a variable ACTUALLY held at that line.

use v5.36;
use strict;
use warnings;

require './internal/infer/testdata/observe_types.pl';

# --- literals: the cases inference should get exactly right ---
my $int    = 42;                    ::__observe(__LINE__, '$int',    $int);
my $num    = 3.14;                  ::__observe(__LINE__, '$num',    $num);
my $str    = "hello";               ::__observe(__LINE__, '$str',    $str);
my $href   = {};                    ::__observe(__LINE__, '$href',   $href);
my $aref   = [];                    ::__observe(__LINE__, '$aref',   $aref);
my $cref   = sub { 1 };             ::__observe(__LINE__, '$cref',   $cref);
my $rx     = qr/a/;                 ::__observe(__LINE__, '$rx',     $rx);
my $sref   = \$int;                 ::__observe(__LINE__, '$sref',   $sref);
my $undef  = undef;                 ::__observe(__LINE__, '$undef',  $undef);

# --- arithmetic and string results ---
my $sum    = $int + 1;              ::__observe(__LINE__, '$sum',    $sum);
my $quot   = $int / 8;              ::__observe(__LINE__, '$quot',   $quot);
my $concat = $str . "!";            ::__observe(__LINE__, '$concat', $concat);
my $len    = length($str);          ::__observe(__LINE__, '$len',    $len);
my $upper  = uc($str);              ::__observe(__LINE__, '$upper',  $upper);
my $idx    = index($str, "e");      ::__observe(__LINE__, '$idx',    $idx);

# --- aggregates in scalar context ---
my @arr    = (1, 2, 3);
my $count  = @arr;                  ::__observe(__LINE__, '$count',  $count);
my %h      = (a => 1);
my $keys   = keys %h;               ::__observe(__LINE__, '$keys',   $keys);

# --- comparison results ---
my $lt     = ($int < 100);          ::__observe(__LINE__, '$lt',     $lt);
my $eq     = ($str eq "hello");     ::__observe(__LINE__, '$eq',     $eq);

# --- ternary joining two arms ---
my $tern   = 1 ? 42 : 43;           ::__observe(__LINE__, '$tern',   $tern);

# --- builtins returning specific types ---
my $joined = join ",", @arr;        ::__observe(__LINE__, '$joined', $joined);
my $sub    = substr($str, 0, 2);    ::__observe(__LINE__, '$sub',    $sub);
my $ordv   = ord("A");              ::__observe(__LINE__, '$ordv',   $ordv);
my $chrv   = chr(65);               ::__observe(__LINE__, '$chrv',   $chrv);

# --- element access ---
my @nums   = (10, 20, 30);
my $elem   = $nums[1];              ::__observe(__LINE__, '$elem',   $elem);
my %conf   = (name => "psc", n => 7);
my $hval   = $conf{name};           ::__observe(__LINE__, '$hval',   $hval);
my $hnum   = $conf{n};              ::__observe(__LINE__, '$hnum',   $hnum);
my $aref2  = \@nums;
my $deref  = $aref2->[0];           ::__observe(__LINE__, '$deref',  $deref);
my $href2  = \%conf;
my $hderef = $href2->{name};        ::__observe(__LINE__, '$hderef', $hderef);

# --- subroutine returns ---
sub ret_int  { return 7 }
sub ret_str  { return "seven" }
sub ret_ref  { return {} }
sub ret_list { return (1, 2, 3) }
my $ri     = ret_int();             ::__observe(__LINE__, '$ri',     $ri);
my $rs     = ret_str();             ::__observe(__LINE__, '$rs',     $rs);
my $rr     = ret_ref();             ::__observe(__LINE__, '$rr',     $rr);
my $rl     = ret_list();            ::__observe(__LINE__, '$rl',     $rl);

# --- loop variables ---
for my $n (@nums) {
    ::__observe(__LINE__, '$n', $n);
    last;
}
foreach my $k (sort keys %conf) {
    ::__observe(__LINE__, '$k', $k);
    last;
}

# --- ternary with unlike arms, both directions ---
my $t_int  = (1 ? 1 : "s");         ::__observe(__LINE__, '$t_int',  $t_int);
my $t_str  = (0 ? 1 : "s");         ::__observe(__LINE__, '$t_str',  $t_str);

# --- more builtins ---
my $fmt    = sprintf("%d", 42);     ::__observe(__LINE__, '$fmt',    $fmt);
my $rev    = reverse("abc");        ::__observe(__LINE__, '$rev',    $rev);
my @sorted = sort { $a <=> $b } @nums;
my $first  = $sorted[0];            ::__observe(__LINE__, '$first',  $first);
my $lc     = lc("ABC");             ::__observe(__LINE__, '$lc',     $lc);
my $abs    = abs(-5);               ::__observe(__LINE__, '$abs',    $abs);
my $int_of = int(3.9);              ::__observe(__LINE__, '$int_of', $int_of);

# --- string ops that must NOT come back numeric ---
my $rep    = "ab" x 3;              ::__observe(__LINE__, '$rep',    $rep);
my $joined2 = join("-", 1, 2);      ::__observe(__LINE__, '$joined2', $joined2);

# --- defined-or and boolean results ---
my $dor    = undef // "fallback";   ::__observe(__LINE__, '$dor',    $dor);
my $neg    = !1;                    ::__observe(__LINE__, '$neg',    $neg);

# --- mutation, then read: the family bson found silent miscompiles in ---
# PSC is a checker with no memory model, so these measure whether a read
# AFTER a mutation still reports the right type.
my @mut = (1, 2, 3);
shift @mut;
my $after_shift = scalar(@mut);     ::__observe(__LINE__, '$after_shift', $after_shift);
push @mut, 9;
my $after_push  = scalar(@mut);     ::__observe(__LINE__, '$after_push',  $after_push);
splice(@mut, 0, 1);
my $after_splice = scalar(@mut);    ::__observe(__LINE__, '$after_splice', $after_splice);

# Return values, which are easy to get wrong by symmetry.
my @rv = (1, 2);
my $push_rv  = push @rv, 3;         ::__observe(__LINE__, '$push_rv',  $push_rv);
my $unsh_rv  = unshift @rv, 0;      ::__observe(__LINE__, '$unsh_rv',  $unsh_rv);
my @spl = (1, 2, 3);
my $spl_rv   = splice(@spl, 1, 1);  ::__observe(__LINE__, '$spl_rv',   $spl_rv);
my $pop_rv   = pop @spl;            ::__observe(__LINE__, '$pop_rv',   $pop_rv);
my $shift_rv = shift @spl;          ::__observe(__LINE__, '$shift_rv', $shift_rv);

# A foreach variable ALIASES the array, so a body write mutates the source.
my @alias = (1, 2, 3);
for my $x (@alias) { $x = $x * 10 }
my $aliased = $alias[0];            ::__observe(__LINE__, '$aliased', $aliased);
my $alias_str = "@alias";           ::__observe(__LINE__, '$alias_str', $alias_str);
