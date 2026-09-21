# ABOUTME: A named hash slice is rv2hv for a package hash and padhv for a lexical, from identical source text.
# ABOUTME: The wider bucket's fixture: we hedge rather than commit, because the declaration is undecidable from the tree alone.
our %h;
@h{"a","b"} = (1, 2);
