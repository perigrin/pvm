# Assess round 4: pushback lens, raiser dispositions (assess-pushback)

- **Subject:** revision 4, commit 6dc1d9558, `docs/plans/2026-11-23-chart-stateless-precedence.md` ("Chart parser: precedence without parser state"), blob 123012c1a5765883ded36de157c8d0e8ac764f0a (`git rev-parse` at the commit).
- **Role:** I act only as raiser. I am also the minute's drafter, so I am biased toward finding my own bar met.

## Dispositions

| id | disposition | against what text, or why |
|---|---|---|
| P16 | **Released** | §3.2: "No hash table, built-in or hand-written, is read or written per derivation. Per-node lists may grow. … The maps `NewPrecedence` builds once per grammar stay." That settles both readings P16 named, and construction-time maps too. |
| P1 | **Moot: subject out of scope** | §6: "A value on items (revision 1's Option C)" is out of scope. P1 concerned how that value attaches to items. |
| P2 | **Moot: subject out of scope** | §6, as for P1. P2 concerned re-propagating a value that widens after its item completes. |
| P10 | **Moot: subject out of scope** | §6, as for P1. P10 concerned the value's integer width. |
| P3 | **Moot: subject out of scope** | §6, as for P1. P3's consequence needed a `%prec` unit rule to *set* an item value. Revision 4 does have an analogue: its shift side keys on the "spine-top item" (§3.1), and whether that key sees through `termbinop: termrelop %prec PREC_LOW` is the same question in a new place. But the differential oracle over the aggregate, the referee and the toys (§3, "Identical output"), together with the first stop, now detects that defect if it happens. So I do not carry P3 forward. |

## New in my lens

### P17 (medium-low; §3 "Identical output" and "Recognition's contract")

The differential test does not check two of the four outputs it promises. In one case the two requirements also pull against each other.

**What is promised and what is tested.** "Identical output" requires that "the parse verdict, the refusal reason, the kept set and the readings count equal today's". But the named check "compares verdict and exact kept set per file". The refusal reason and the readings count are left to the ratchet count keys and the sample digest.

**Where the two requirements meet.** "Recognition's contract" lets the recognizer, with lookahead on, accept a strict subset of what the reference accepts. Wherever the reduce-side filter (§3.3) leaves recognition no derivation, `forest.parseStream` refuses with reason `grammar` and a byte, where today it refuses with `precedence` and Byte 0. Two cases:

- On the aggregate, the prototype saw recognition acceptance unchanged, so no such input appeared.
- The README's `print $fh []`, where "the one bison builds does not exist", is the shape of input that could flip. Whether it flips depends on whether §3.3's "at that point" applies a state-dependent pair (indirob) in recognition. §3.1 keys indirob on the parent item for the forest only.

**If it ships:** a refusal reason can change with every named check green. If such a file is ever found, the first stop fires on output that the recognition contract allows.

**Modify:** add the refusal reason and byte, and the readings count, to the differential test. Then do one of two things:

- state that the reduce-side filter in recognition applies no pair whose decision differs between states; or
- state which refusal reason is expected when recognition refuses first.

**What would change my mind:** a statement that every pair §3.3 applies in recognition is global, and that recognition therefore can only refuse where today's `Constrain` refuses with no surviving derivation **and** the refusal reason is redefined. Without both, the gap stands.
