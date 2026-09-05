<!-- ABOUTME: Coordinator's ruling on the statement-level contradiction, for round 2. -->
<!-- ABOUTME: Decided on the merits rather than by the ownership rule. -->

# The statement-level call

Ch5:2973 — "Do not attempt statement-level incremental parsing. The bookkeeping
costs more than the parse."
Ch6:786 — "Start with statement-level and…"

The ownership rule says ch6 owns incremental machinery, so ch6 would win by
default. **On the merits, ch5 is right**, and ch6's own numbers are why.

Measured (00-findings §0.8, ch6 §6.10.2): a single parse of a real perl5/lib
file costs 6-500 ms. `Analyze` costs 0.5-2 ms. Per-statement bookkeeping — a
cache entry, an invalidation edge, and a proof of boundary safety per statement
— buys a fraction of the cheaper number while adding state that must be correct
on every edit.

Ch6 §6.11 already agrees, in the same chapter that contradicts it: step 3 is
"full-parse-on-every-change server… Ship it", step 4 is "Measure. Do not
optimize before this step produces numbers", and step 7 is damage/repair behind
a flag. Statement-level granularity at step 7 contradicts steps 3 and 4 of its
own ladder.

**Ruling:** sub-body is the first incremental boundary. Ch6:786 changes to
sub-body and cites ch5's boundary table; ch5:2973 stays. This is also what
perl-lsp's failure argues for — it built fine-grained reuse machinery that
production never calls (`text_sync.rs:1-9`), which is the same mistake at a
smaller granularity.

The general point: an ownership rule decides *where* a topic is specified. It
does not decide *what is true*. Where the owner is wrong, the owner changes its
mind — it does not win by position.
