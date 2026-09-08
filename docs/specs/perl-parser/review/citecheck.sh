#!/bin/bash
# For each toke.c:NNNN citation in the spec, print the enclosing C function.
# A citation whose function is unrelated to the claim is drift.
cd /home/perigrin/dev/pvm/docs/specs/perl-parser
grep -ohE 'toke\.c:[0-9]+' *.md ../../plans/2026-09-05-*.md 2>/dev/null | sort -u -t: -k2 -n | while IFS=: read _ line; do
  fn=$(awk -v L="$line" 'NR<=L && /^[A-Za-z_].*\(/ && !/^ / && !/;$/ {f=$0} END{print f}' /home/perigrin/dev/perl5/toke.c | grep -oE '\b(S_|Perl_)?[a-z_0-9]+\(' | tail -1 | tr -d '(')
  printf "%-6s %s\n" "$line" "${fn:-?}"
done
