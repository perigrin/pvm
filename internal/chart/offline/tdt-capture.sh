#!/bin/sh
# ABOUTME: Prints one file's own toke.c token stream from the pinned -DDEBUGGING perl's -DT, module tokens dropped.
# ABOUTME: Offline only. Usage: tdt-capture.sh <perl> <file>; PERL5LIB passes through; the form is in README.md.
set -eu

if [ $# -ne 2 ]; then
	echo "usage: tdt-capture.sh <debugging-perl> <file.pl>" >&2
	exit 2
fi
perl=$1
file=$2

[ -r "$file" ] || { echo "tdt-capture.sh: $file: not a readable file" >&2; exit 2; }
# Without DEBUGGING, perl warns about -D and runs on, which would read as an empty stream.
"$perl" -V:ccflags | grep -q -- -DDEBUGGING ||
	{ echo "tdt-capture.sh: $perl is not a -DDEBUGGING build; -DT needs one" >&2; exit 2; }
eval "$("$perl" -V:archlibexp)"
perly_h=$archlibexp/CORE/perly.h
[ -r "$perly_h" ] || { echo "tdt-capture.sh: no $perly_h" >&2; exit 2; }

# -Dp adds the parser's "lookahead token is token GRAMPROG" line, the one mark a
# compilation unit's start leaves (a loaded file or a string eval); -DT's EOF ends
# it. Only the first unit's tokens print. Perl's stdin is /dev/null, its stdout
# is dropped, and its exit status arrives as the stream's last line.
# Known limit: -Dp spins without end once BEGIN puts an :encoding layer on STDERR (README.md).
{
	"$perl" -c -DpT "$file" 2>&1 >/dev/null </dev/null && st=0 || st=$?
	echo "tdt-capture-status $st"
} | awk -v perly_h="$perly_h" '
BEGIN {
	# -DT prints "?? N" for a token missing from toke.c debug_tokens; perly.h names it.
	while ((getline l < perly_h) > 0)
		if (l ~ /^ *[A-Za-z_0-9]+ = [0-9]+/) { split(l, f, /[ =,]+/); name[f[3]] = f[2]; names++ }
	if (!names) { bad = "no token enum in " perly_h; exit }
}
/^tdt-capture-status / { st = $2; next }
/^lookahead token is token GRAMPROG / { unit[++d] = ++n; next }
/^### [0-9]+:LEX_/ { ident = ""; next }
/^### forced token:$/ { forced = 1; next }
/^### Pending identifier / { ident = substr($0, 25, length($0) - 25); next }
/^### <== / {
	if (forced) { forced = 0; next }   # reported again when delivered
	r = substr($0, 9)
	if (r == "EOF") { if (d-- == 0) { bad = "EOF outside any unit"; exit } next }
	if (unit[d] != 1) next
	if (r ~ /^\?\? [0-9]+$/) {
		t = name[substr(r, 4)]
		if (t == "") { bad = "no perly.h name for " r; exit }
		print t
		next
	}
	if (r !~ /^[A-Z]/) { bad = "unexpected token report: " r; exit }
	t = r; sub(/\(.*/, "", t)
	if ((i = index(r, "(opval=op_const) ")) > 0) v = substr(r, i + 17)
	else if (ident != "") v = ident
	else if (match(r, /=op_[a-z0-9_]+\)/)) v = substr(r, RSTART + 1, RLENGTH - 2)
	else v = ""
	print (v == "" ? t : t "\t" v)
}
END {
	if (bad == "" && st == "") bad = "perl exit status missing"
	if (bad == "" && st == 0 && d != 0) {
		# An exit in BEGIN (plan skip_all) stops perl with status 0 before the file ends.
		print "# truncated: perl exited 0 before the file reached EOF"
		bad = d " unit(s) never reached EOF"
	}
	if (bad != "") { print "tdt-capture.sh: " bad > "/dev/stderr"; exit 2 }
	if (st != 0) { print "# refused: perl exited " st; exit 1 }
}'
