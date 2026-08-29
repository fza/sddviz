---
name: sddviz-sdd-capture
description: Capture an entry into this repository's decision graph. Use whenever this repository gains a signal or a decision, alongside any sdd skill, and before reaching for sdd new by hand.
allowed-tools: Read Write Edit Bash(python3 scripts/graph-capture.py *) Bash(git worktree *) Bash(devbox run -- sdd *)
---

# Capturing a graph entry in sddviz

One command captures an entry: `python3 scripts/graph-capture.py <draft.md>`. Never run `sdd new`
by hand here, and never write a wrapper script for one capture.

## The draft

One Markdown file. A leading YAML block carries the fields, the body sits below it.

```markdown
---
type: decision              # signal | decision
layer: tactical             # strategic | conceptual | tactical | operational | process
kind: plan                  # signal: gap fact question insight done actor annotation
                            # decision: directive activity plan contract aspiration role focus procedure
intent: guiding             # directives only: pending | guiding | settled
confidence: high
participants: [Felix, Claude]
topics: [area-driver-contract, drivers, defect]
closes: [20260824-204029-s-tac-nxn]
supersedes: [20260821-041258-d-tac-cif]
refs:
  - id: 20260828-233250-d-tac-tyi
    kind: refines           # grounded-in builds-on refines addresses surfaces
                            # surfaced-by depends-on required-by related
    desc: what this narrows in the target
---
The body, opening with one sentence that stands alone as the entry's claim.
```

Write the draft under `.tmp/`. The script writes and removes its own scratch file. The draft is
yours to delete once the entry lands.

## What the script does

Three gates, stopping at the first that reports.

1. The draft shape, and this repository's rule that `topics` carry exactly one `area-` label.
2. Vale over the body, read from a real Markdown file. New prose clears every error, warning and
   suggestion. A Vale finding stops the run, and the pre-flight never runs on a body Vale rejects.
3. `sdd new --dry-run`, the tool's own pre-flight. Anything above low severity blocks.

The script captures a draft clearing all three in the same run, and `sdd` commits the entry. The
capture passes `--preflight-verified`, which records that the pre-flight ran and reported nothing
blocking, where `--skip-preflight` would mark the entry as one nothing ever checked.

A build listing no such flag still captures. The script reads `sdd new --help` to tell. Where the
flag is missing, the capture carries no pre-flight flag at all, so the tool runs its own pass a
second time and the entry lands the way any plain capture does. That second pass costs the run the
flag exists to save. Whatever it reports reaches you under the same rule: medium and high, never
low.

**A low-severity pre-flight finding never reaches you.** Nobody acts on one, and the script strips
every one before it reports.

`--force` is the only flag. It captures despite pre-flight findings above low severity, and it is
for a finding you judge wrong, never for one you would rather not fix.

## The repair loop

A run reports findings and captures nothing. Edit the same draft file, run the script again. Repeat
until it captures. Nothing carries over between runs, so the draft file is the whole state.

```bash
python3 scripts/graph-capture.py .tmp/entry.md   # reports two findings above low severity
# edit .tmp/entry.md
python3 scripts/graph-capture.py .tmp/entry.md   # clean, captured, committed
```

Exit codes: `0` captured, `1` refused by the draft shape, by Vale, or by the tool, `2` blocked on
pre-flight findings above low severity.

## What it does not do

Every value the tool judges belongs to the tool: types, layers, kinds, confidences, reference
kinds, and the fields a kind requires. Do not restate them in a draft check or argue with a refusal
the tool issues.

## Before writing a draft

- **Search the graph.** One `search` changes the framing of a topic more often than it finds a duplicate. Read
  what looks close with `show` before asserting what the graph holds.
- **Pick the sharp reference kind.** `refines` narrows a live decision. `required-by` records a
  prerequisite from the prerequisite's side, where `grounded-in` says this entry reasons from the
  target. `addresses` takes up an open gap without closing it. `related` is the floor and understates most links.
- **Name every surface the entry implies** - a command, flag, RPC verb, config key, label, storage
  path, or environment variable. The entry is the only place that name lives until the surface
  exists.
- **Retire what the entry replaces.** `supersedes` for a decision it retires, `closes` for a signal
  it answers. Where an entry overturns part of an active decision, name which criteria retire in the
  body.

## While another session holds the checkout

A capture commits, so it goes in a worktree rather than on the shared tree:

```bash
git worktree add -b corpus/<subject> .claude/worktrees/<name> HEAD
```

Capture there, leave the branch for merging once the other session finishes, and remove the
worktree with `git worktree remove .claude/worktrees/<name>`.

## After capture

An entry is immutable. A body, its topics, and its references never change. A correction is a new
entry, and only the summary can be rewritten, through `sdd summarize`.
