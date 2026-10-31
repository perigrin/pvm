#!/bin/sh
# ABOUTME: Labels every waypoint-1 corpus file with the pinned -DDEBUGGING perl's perl -c verdict (spec §9.1).
# ABOUTME: Writes testdata/labels/{in-tree,t1}.txt with a provenance header; each refusal gets a proposed category.
#
# usage: T1_CORPUS=<PerlOnJava>/src/test/resources/unit sh internal/chart/offline/label.sh <perl>
#
# Each file is checked with `perl -c` under `env -i` (PATH and PERL5LIB only)
# and a per-file timeout. A refusal's category is proposed from perl's first
# error, the first stderr line matching a pattern below; a refusal matching
# none stops the script, so every refusal is classified. Precedence is never
# proposed: it is shown by referee evidence, which LABELS-review adds.
set -eu

perl=${1:?usage: label.sh <perl>}
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

top=$(git rev-parse --show-toplevel)
out=$top/internal/chart/testdata/labels
mkdir -p "$out"

# category<TAB>ERE, matched against each stderr line in order.
patterns='environment	^Can.t locate .* in @INC
environment	^This Perl not built to support threads
environment	^"-T" is on the #! line
environment	too modern--this is v
semantic	^Can.t modify
semantic	not allowed while "strict subs" in use
semantic	^Invalid CODE attributes?:
semantic	is no longer supported as of Perl
grammar	^syntax error
grammar	found where operator expected
grammar	^Can.t find string terminator
grammar	^Use of bare << to mean <<"" is forbidden
grammar	^No digits found for
grammar	^Illegal (binary|octal|hexadecimal) digit
grammar	^Invalid version format
grammar	^Illegal declaration of anonymous subroutine'

# label <cwd> <root> <PERL5LIB> <corpus> <file>...: the header, then one line
# per file named relative to <root>, itself relative to <cwd>.
label() {
	cwd=$1 root=$2 lib=$3 corpus=$4
	shift 4
	printf '# perl: %s\n# version: %s\n# git_commit_id: %s\n# ccflags: %s\n' \
		"$perl" "$(config version)" "$commit" "$ccflags"
	printf '# corpus: %s\n# cwd: %s\n# root: %s\n# PERL5LIB: %s\n# environment: env -i, PATH and PERL5LIB only\n' \
		"$corpus" "$cwd" "${root:-.}" "${lib:-(empty)}"
	printf '# timeout: %ss per file; a timeout is labelled environment\n' "$timeout"
	printf '# limit: a label speaks for perl'"'"'s first error only; a later error in the same file is not seen\n'
	printf '# status: every category is proposed by label.sh until LABELS-review reviews it\n'
	for f; do
		s=0
		err=$(cd "$cwd" && env -i PATH="$PATH" PERL5LIB="$lib" \
			timeout "$timeout" "$perl" -c "${root:+$root/}$f" 2>&1 >/dev/null </dev/null) || s=$?
		if [ "$s" -eq 0 ]; then
			printf '%s\taccept\n' "$f"
			continue
		fi
		if [ "$s" -eq 124 ]; then
			printf '%s\trefuse\tenvironment\tproposed\ttimeout after %ss\n' "$f" "$timeout"
			continue
		fi
		hit=$(printf '%s\n' "$err" | tr '\t' ' ' | sed 's/ (@INC entries checked: [^)]*)//' |
			awk -v pats="$patterns" 'BEGIN { n = split(pats, p, "\n") }
				{ for (i = 1; i <= n; i++) { split(p[i], c, "\t"); if ($0 ~ c[2]) { print c[1] "\t" $0; exit } } }')
		[ -n "$hit" ] || die "$f: no pattern classifies perl's refusal (exit $s): $err"
		printf '%s\trefuse\t%s\tproposed\t%s\n' "$f" "${hit%%	*}" "${hit#*	}"
	done
}

cd "$top"
label "." "" internal/chart/testdata/guacamole/stubs "pvm, paths from the repository root" \
	internal/chart/testdata/guacamole/parses/*.pl internal/chart/testdata/guacamole/parsent/*.pl \
	internal/chart/testdata/std5k.pl internal/psc/testdata/latency/edit5k.pl >"$out/in-tree.txt.tmp"
mv "$out/in-tree.txt.tmp" "$out/in-tree.txt"

t1top=$(git -C "$t1" rev-parse --show-toplevel)
t1root=$(git -C "$t1" rev-parse --show-prefix)
label "$t1top" "${t1root%/}" "" "PerlOnJava $t1commit" \
	$(cd "$t1" && ls -- *.t) >"$out/t1.txt.tmp"
mv "$out/t1.txt.tmp" "$out/t1.txt"
