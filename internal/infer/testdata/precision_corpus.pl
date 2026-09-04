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
