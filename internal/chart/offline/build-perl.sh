#!/bin/sh
# ABOUTME: Builds and installs the pinned -DDEBUGGING perl (blead 94e50866) outside the repository.
# ABOUTME: Offline only: make test and the ratchet never run it. Usage: build-perl.sh <perl5-tree> <prefix>.
set -eu

PIN=94e50866

if [ $# -ne 2 ]; then
	echo "usage: build-perl.sh <perl5-tree> <install-prefix>" >&2
	exit 2
fi
tree=$1
prefix=$2

head=$(git -C "$tree" rev-parse HEAD)
case $prefix in
/*) ;;
*) echo "usage: build-perl.sh <perl5-tree> <absolute-install-prefix>" >&2; exit 2 ;;
esac

case $head in
"$PIN"*) ;;
*)
	echo "build-perl.sh: $tree is at $head; the pinned revision is $PIN" >&2
	exit 1
	;;
esac

# -Uversiononly installs bin/perl, which usedevel otherwise suffixes with the version.
# A shared clone keeps $tree and its git config unwritten while make_patchnum still sees the commit.
work=$(mktemp -d "${TMPDIR:-/tmp}/perl-build.XXXXXX")
trap 'rm -rf "$work"' EXIT
git clone -q --shared "$tree" "$work/src"
git -C "$work/src" checkout -q --detach "$head"

cd "$work/src"
sh Configure -des -DDEBUGGING -Dusedevel -Uversiononly -Dprefix="$prefix" >"$work/configure.log" 2>&1 || { tail -20 "$work/configure.log" >&2; exit 1; }
make -j"$(nproc)" >"$work/make.log" 2>&1 || { tail -20 "$work/make.log" >&2; exit 1; }
make install >"$work/install.log" 2>&1 || { tail -20 "$work/install.log" >&2; exit 1; }
echo "$prefix/bin/perl"
