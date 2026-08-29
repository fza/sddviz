# Repository checks

Six checks assert that a claim one artifact makes about another still holds. `refs.py` runs all six
and reports every one, rather than stopping at the first failure. `devbox run refs:lint` is the
repository's own invocation, and `devbox run lint` calls it last.

Each check exists because no other tool sees the class it covers. A prose linter reads style and
never follows a pointer. A compiler reads code and never reads a document. So a reference rots for
as long as nobody happens to try it.

| Check | Asserts |
|---|---|
| `paths` | a bare tree path a tracked file names exists |
| `links` | a relative link destination a Markdown file writes resolves |
| `identifiers` | a code identifier a document names exists under `source/` |
| `vocabulary` | the rules document and the Vale taxonomy name the same forms |
| `taxonomy-probes` | every banned-word check reports the word it bans |
| `vocabulary-shape` | no accepted term names code instead of a word |

`identifiers` asserts this repository's own symbols. A documentation line naming the embedded sdd
library names that library's symbols, which this tree does not carry, so such a line is skipped.

`vocabulary` runs in both directions: a taxonomy rule may not tell a writer to use a form the rules
document never explains, and the rules document may not define a term its own checks report.

`taxonomy-probes` is the guard on the guards. A Vale token that silently stops matching passes every
document, so each banned-word rule carries lines it must report and lines it must leave alone.
Repair the token when it fires, never the probe.

Naming one or more as arguments runs those alone, which is how a repair loop stays fast:

```bash
python3 scripts/lint/refs.py identifiers paths
```

A clean check reports one line. A failing one reports its sites under it, then what to do about
them, and the run leaves with a non-zero status.
