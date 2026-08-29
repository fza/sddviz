# Repository checks

Three checks assert that a claim one artifact makes about another still holds. `refs.py` runs all
three and reports every one, rather than stopping at the first failure. `devbox run refs:lint` is
the repository's own invocation, and `devbox run lint` calls it last.

Each check exists because no other tool sees the class it covers. A prose linter reads style and
never follows a pointer. A compiler reads code and never reads a document. So a reference rots for
as long as nobody happens to try it.

| Check | Asserts |
|---|---|
| `paths` | a bare tree path a tracked file names exists |
| `links` | a relative link destination a Markdown file writes resolves |
| `identifiers` | a code identifier a document names exists under `source/` |

Naming one or more as arguments runs those alone, which is how a repair loop stays fast:

```bash
python3 scripts/lint/refs.py identifiers paths
```

A clean check reports one line. A failing one reports its sites under it, then what to do about
them, and the run leaves with a non-zero status.
