#!/bin/sh
# ABOUTME: Labels every corpus file, waypoint 1's and perl.git t/, with the pinned -DDEBUGGING perl's perl -c verdict (spec §9.1).
# ABOUTME: Writes testdata/labels/{in-tree,t1,perlgit}.txt with a provenance header; each refusal gets a proposed category and its evidence.
#
# usage: T1_CORPUS=<PerlOnJava>/src/test/resources/unit [TDT_FIXTURES=<dir>] [PERL5_CORPUS=<dir>] \
#        sh internal/chart/offline/label.sh <perl> <referee-dir>
#
# <referee-dir> holds referee-prec and referee-noprec (referee/build.sh).
# TDT_FIXTURES is tdt-align.sh's output (default /var/tmp/pvm-chart-tdt).
# PERL5_CORPUS is the perl5 tree at 94e50866 (default /opt/corpora/perl5);
# its t/ runs from a shim t/ (perlgit-shim.sh), with the paths under t/.
#
# Each file is checked with `perl -c` under `env -i` (PATH and PERL5LIB only)
# and a per-file timeout. A refusal's category is proposed from perl's first
# error, its first stderr line, by the first pattern below it matches; a
# first line matching none stops the script, so no refusal is classified by
# a later line. A grammar refusal's -DT stream goes to both referees, and is
# precedence when perl's refusing token is a viable prefix without precedence
# (spec §9.1). A file perl -c accepts whose -DT capture is truncated stopped
# in BEGIN without compiling the rest, and is refused as environment
# (perigrin, "Label environment").
set -eu

perl=${1:?usage: label.sh <perl> <referee-dir>}
referee=${2:?usage: label.sh <perl> <referee-dir>}
fixtures=${TDT_FIXTURES:-/var/tmp/pvm-chart-tdt}
t1=${T1_CORPUS:?set T1_CORPUS to PerlOnJava 73be162b src/test/resources/unit}
timeout=30

die() { echo "label.sh: $*" >&2; exit 1; }
config() { "$perl" -MConfig -e "print \$Config{$1}"; }

commit=$(config git_commit_id)
case $commit in 94e50866*) ;; *) die "$perl is at commit '$commit', not 94e50866" ;; esac
ccflags=$(config ccflags)
case " $ccflags " in *" -DDEBUGGING "*) ;; *) die "$perl is not a -DDEBUGGING build" ;; esac
t1commit=$(git -C "$t1" rev-parse --short=8 HEAD)
[ "$t1commit" = 73be162b ] || die "T1_CORPUS is at PerlOnJava $t1commit, not 73be162b"
[ -z "$(git -C "$t1" status --porcelain)" ] || die "T1_CORPUS=$t1 has uncommitted changes; the pinned corpus must be clean"

for v in prec noprec; do [ -x "$referee/referee-$v" ] || die "no $referee/referee-$v; run referee/build.sh"; done
[ -d "$fixtures" ] || die "TDT_FIXTURES=$fixtures is not a directory; run tdt-align.sh"
fixtures=$(cd "$fixtures" && pwd)

top=$(git rev-parse --show-toplevel)
here=$top/internal/chart/offline
out=$top/internal/chart/testdata/labels
nocap=$top/internal/chart/testdata/tdt/perlgit-uncaptured.txt
mkdir -p "$out"
shim=$(mktemp -d)
trap 'rm -rf "$shim"' EXIT
sh "$here/perlgit-shim.sh" "$perl" "$shim"

# category<TAB>ERE, matched against perl's first error, in order.
# "$# / $* is no longer supported" is grammar, not semantic: -DpT shows gv.c
# croaking from toke.c's S_pending_ident before the name token reaches the
# parser, so it is a fatal lexing error (LABELS-review overturned semantic).
# A module a use loads by eval, and a ./ or ../ require, which searches no
# @INC, are perl.git t/ files needing what the shim's lib/ or t/ lacks.
patterns='environment	^(use [A-Za-z_:]+: )?Can.t locate .* in @INC
environment	^Can.t locate [.][.]?/[^ ]* at
environment	^This Perl not built to support threads
environment	^"-[Tt]" is on the #! line
environment	too modern--this is v
semantic	^Can.t modify
semantic	not allowed while "strict subs" in use
semantic	^Invalid CODE attributes?:
grammar	^syntax error
grammar	is no longer supported as of Perl
grammar	found where operator expected
grammar	^Can.t find string terminator
grammar	^Use of bare << to mean <<"" is forbidden
grammar	^No digits found for
grammar	^Illegal (binary|octal|hexadecimal) digit
grammar	^Invalid version format
grammar	^Illegal declaration of anonymous subroutine'

# referees <cwd> <PERL5LIB> <file>: both referees' verdicts on the file's -DT
# stream, cut after the token referee-prec (perl's grammar) refuses.
referees() {
	toks=$(cd "$1" && PERL5LIB=$2 sh "$here/tdt-capture.sh" "$perl" "$3" 2>/dev/null | grep -v '^#' | cut -f 1) || true
	[ -n "$toks" ] || die "$3: no -DT tokens"
	toks=$(printf 'GRAMPROG\n%s' "$toks")
	prec=$(printf '%s\n' "$toks" | "$referee/referee-prec" -p)
	n=$(printf '%s\n' "$toks" | wc -l)
	case $prec in refuse@*) n=${prec#refuse@} ;; esac
	noprec=$(printf '%s\n' "$toks" | head -n "$n" | "$referee/referee-noprec" -p)
	printf 'referee-noprec %s; referee-prec %s' "$noprec" "$prec"
}

# label <cwd> <root> <PERL5LIB> <corpus> <fixtures> <file>...: the header, then
# one line per file named relative to <root>, itself relative to <cwd>;
# <fixtures> holds each file's aligned -DT capture as <file>.tdt.
label() {
	cwd=$1 root=$2 lib=$3 corpus=$4 fx=$5
	shift 5
	printf '# perl: %s\n# version: %s\n# git_commit_id: %s\n# ccflags: %s\n' \
		"$perl" "$(config version)" "$commit" "$ccflags"
	printf '# corpus: %s\n# cwd: %s\n# root: %s\n# PERL5LIB: %s\n# environment: env -i, PATH and PERL5LIB only\n' \
		"$corpus" "$cwd" "${root:-.}" "${lib:-(empty)}"
	printf '# timeout: %ss per file; a timeout is labelled environment\n' "$timeout"
	printf '# limit: a label speaks for perl'"'"'s first error only; a later error in the same file is not seen\n'
	printf '# status: every category is proposed by label.sh until it is reviewed against its evidence\n'
	printf '# review: referee verdicts on the -DT stream up to the token perl refused; precedence when only referee-prec refuses there\n'
	for f; do
		s=0
		err=$(cd "$cwd" && env -i PATH="$PATH" PERL5LIB="$lib" \
			timeout "$timeout" "$perl" -c "${root:+$root/}$f" 2>&1 >/dev/null </dev/null) || s=$?
		if [ "$s" -eq 0 ]; then
			# A perl.git file tdt-align.sh does not capture keeps its perl -c verdict unchecked for a skip in BEGIN.
			if [ "$fx" = "$fixtures/perlgit" ] && grep -qxF "$f" "$nocap"; then
				printf '%s\taccept\n' "$f"
				continue
			fi
			c=$(sed -n 2p "$fx/$f.tdt") || die "$f: no -DT capture in $fx"
			case $c in
			'# capture: truncated'*)
				printf '%s\trefuse\tenvironment\tproposed\tperl -c stops at a skip (skip_all or exit) in BEGIN; the file is not compiled\t-DT %s\n' "$f" "${c#\# }" ;;
			*) printf '%s\taccept\n' "$f" ;;
			esac
			continue
		fi
		# shortcut: 124 is read as a timeout by status alone (tdt-capture.sh reads it with -Dp's open units),
		# so a file exiting 124 from BEGIN would be labelled a timeout; none does at 73be162b or in-tree
		# (no label carries timeout evidence). Upgrade if one ever does.
		if [ "$s" -eq 124 ]; then
			printf '%s\trefuse\tenvironment\tproposed\ttimeout after %ss\n' "$f" "$timeout"
			continue
		fi
		hit=$(printf '%s\n' "$err" | head -n 1 | tr '\t' ' ' | sed 's/ (@INC entries checked: [^)]*)//' |
			awk -v pats="$patterns" 'BEGIN { n = split(pats, p, "\n") }
				{ for (i = 1; i <= n; i++) { split(p[i], c, "\t"); if ($0 ~ c[2]) { print c[1] "\t" $0; exit } } }')
		[ -n "$hit" ] || die "$f: no pattern classifies perl's first error (exit $s): $err"
		cat=${hit%%	*}
		if [ "$cat" = grammar ]; then
			ev=$(referees "$cwd" "$lib" "${root:+$root/}$f")
			case $ev in "referee-noprec viable; referee-prec refuse@"*) cat=precedence ;; esac
			printf '%s\trefuse\t%s\tproposed\t%s\t%s\n' "$f" "$cat" "${hit#*	}" "$ev"
		else
			printf '%s\trefuse\t%s\tproposed\t%s\n' "$f" "$cat" "${hit#*	}"
		fi
	done
}

cd "$top"
label "." "" internal/chart/testdata/guacamole/stubs "pvm, paths from the repository root" "$fixtures" \
	internal/chart/testdata/guacamole/parses/*.pl internal/chart/testdata/guacamole/parsent/*.pl \
	internal/chart/testdata/std5k.pl internal/psc/testdata/latency/edit5k.pl >"$out/in-tree.txt.tmp"
mv "$out/in-tree.txt.tmp" "$out/in-tree.txt"

t1top=$(git -C "$t1" rev-parse --show-toplevel)
t1root=$(git -C "$t1" rev-parse --show-prefix)
label "$t1top" "${t1root%/}" "" "PerlOnJava $t1commit" "$fixtures/T1" \
	$(cd "$t1" && ls -- *.t) >"$out/t1.txt.tmp"
mv "$out/t1.txt.tmp" "$out/t1.txt"

(cd "$shim/t" && label "." "" "" "perl.git t/ at $commit, run from a shim t/ (perlgit-shim.sh): perl5 t/ beside the pinned perl's library as ../lib" \
	"$fixtures/perlgit" $(find . -name '*.t' | sed 's|^\./||' | LC_ALL=C sort)) >"$out/perlgit.txt.tmp"
mv "$out/perlgit.txt.tmp" "$out/perlgit.txt"
