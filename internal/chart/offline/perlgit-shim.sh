#!/bin/sh
# ABOUTME: Builds the shim perl.git t/ compiles in: a copy of $PERL5_CORPUS/t beside the pinned perl's library as ../lib.
# ABOUTME: Run by tdt-align.sh and label.sh; usage: perlgit-shim.sh <perl> <dest>. Refuses a perl5 tree not at 94e50866 or not clean.
#
# t/test.pl clears @INC and unshifts ../lib, so a perl.git test compiles only
# from a t/ whose ../lib holds a built library (parseoracle.BuildShim, the
# oracle's measurement). lib/ is every @INC root but site_perl and
# vendor_perl; an earlier root wins, so the roots are copied last first.
set -eu

perl=${1:?usage: perlgit-shim.sh <perl> <dest>}
dest=${2:?usage: perlgit-shim.sh <perl> <dest>}
tree=${PERL5_CORPUS:-/opt/corpora/perl5}

die() { echo "perlgit-shim.sh: $*" >&2; exit 1; }
[ -d "$tree/t" ] || die "PERL5_CORPUS=$tree has no t/"
head=$(git -C "$tree" rev-parse --short=8 HEAD)
[ "$head" = 94e50866 ] || die "PERL5_CORPUS=$tree is at $head, not 94e50866"
[ -z "$(git -C "$tree" status --porcelain)" ] || die "PERL5_CORPUS=$tree has uncommitted changes; the pinned tree must be clean"

mkdir -p "$dest/lib"
cp -R "$tree/t" "$dest/t"
for r in $("$perl" -e 'print "$_\n" for reverse grep { -d && !m{site_perl|vendor_perl} } @INC'); do
	cp -R "$r/." "$dest/lib/"
done
[ -f "$dest/lib/Config.pm" ] || die "$perl's @INC gave no Config.pm; the shim's lib/ is incomplete"
