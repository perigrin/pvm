#!/bin/sh
# ABOUTME: Reads internal/parseoracle/testdata/corpus.pin and prints its halves as key=value lines.
# ABOUTME: The fidelity workflow needs the pin before Go exists on the runner, so it is read here.
#
# Usage: parseoracle-pin.sh [PIN_FILE] >> "$GITHUB_OUTPUT"
#
# Prints four lines:
#
#   interpreter=5.042000    perl's $], as the pin and the baseline header record it
#   perl_version=5.42.0     the same version dotted, which is what actions-setup-perl takes
#   revision=94e5086...     the perl5 commit the corpus was measured against
#   threads=define          perl's useithreads, which $] does not carry
#
# Every field is mandatory. An absent revision would check out blead at HEAD
# and an absent interpreter would install whatever the runner ships; either
# measures a different world than the baseline describes, and does so quietly.
# threads is the field the first real CI run proved necessary: a runner perl
# at 5.042000 passed the version check while built useithreads=undef, and
# three corpus files that skip_all inside BEGIN moved exact -> no-answer.
# Failing loudly here is the whole reason this is a script with a test
# (internal/parseoracle/ci_test.go) rather than three lines of inline YAML.

set -eu

PIN="${1:-internal/parseoracle/testdata/corpus.pin}"

if [ ! -f "$PIN" ]; then
	echo "parseoracle-pin.sh: no pin at $PIN" >&2
	exit 2
fi

# Strip comments, then take the last assignment for each key. `cut -d= -f2-`
# rather than -f2 so a value containing `=` survives; none does today, but a
# pin reader that silently truncates is worse than one that does not.
field() {
	sed 's/#.*//' "$PIN" |
		grep "^[[:space:]]*$1[[:space:]]*=" |
		tail -n 1 |
		cut -d= -f2- |
		tr -d '[:space:]'
}

interpreter=$(field interpreter)
revision=$(field revision)
threads=$(field threads)

if [ -z "$interpreter" ] || [ -z "$revision" ] || [ -z "$threads" ]; then
	echo "parseoracle-pin.sh: $PIN must record interpreter, revision and threads" >&2
	echo "  interpreter=${interpreter:-<missing>} revision=${revision:-<missing>} threads=${threads:-<missing>}" >&2
	exit 3
fi

case "$threads" in
define | undef) ;;
*)
	echo "parseoracle-pin.sh: threads is $threads; perl spells it define or undef" >&2
	exit 5
	;;
esac

# 5.042000 -> 5.42.0. perl's $] packs the minor and patch versions as
# three zero-padded digits each; actions-setup-perl wants them dotted and
# unpadded. Arithmetic rather than string surgery, so 5.008009 gives 5.8.9
# and not 5.08.09.
case "$interpreter" in
5.??????) ;;
*)
	echo "parseoracle-pin.sh: interpreter $interpreter is not a 5.NNNNNN version" >&2
	exit 4
	;;
esac

# The leading zeros make `${interpreter#5.}` look octal to $(( )), so strip
# them before the arithmetic sees the number: 5.008009 -> 008009 -> 8009.
# Without this, 5.008009 is a syntax error rather than a wrong answer, which
# is at least loud — but 5.042000 would quietly become an invalid octal too.
digits=$(printf '%s' "${interpreter#5.}" | sed 's/^0*//')
digits=${digits:-0}
minor=$((digits / 1000))
patch=$((digits % 1000))

printf 'interpreter=%s\n' "$interpreter"
printf 'perl_version=5.%d.%d\n' "$minor" "$patch"
printf 'revision=%s\n' "$revision"
printf 'threads=%s\n' "$threads"
