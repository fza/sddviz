@AGENTS.md

# CLAUDE.md

The rules for this repository live in [`AGENTS.md`](AGENTS.md), imported above so they load on first
read here. They are not restated: one copy, or the two drift and a reader obeys whichever they opened.
Work under `source/` additionally loads [`source/AGENTS.md`](source/AGENTS.md).

What follows is Claude-Code-specific and applies on top.

## Communication

- Describe actions concretely. Say what actually happens ("writing the parser tests to a file"), not
  metaphor shorthand.
- Use simple, clear, meaningful language in dialogue and in written deliverables. No jargon, no
  decorative metaphors. SDD's own defined terms are fine.
- Never use em dashes in outward-facing prose (the README, `docs/`, anything a user reads). Many
  readers take them as the first sign of AI-generated text. Use commas, colons, periods, or
  parentheses instead.

## The decision graph

**The `/sdd` skill drives this repository's graph.** It runs the CLI: checking in on project state,
capturing what you noticed, recording what you decided, evaluating what landed. Invoke it rather than
composing `sdd` commands by hand for anything conversational, since it carries the
playback-before-capture discipline and the entry-shape rules that a bare command does not.

`.mcp.json` starts `sdd serve`. Use its read tools (`search`, `view`, `show`, `read_attachment`) for
graph reads.

## Prose linting

Two skills carry the [prose-linting rule](AGENTS.md#5-prose-is-linted-before-it-lands): `/vale-fix`
runs Vale and repairs alerts, `/vale-vocab` teaches the Project vocabulary a term it should accept.
`/vale-setup` installs the binary where it is missing.

Take their fixing guidance, and override them on two points. **Severity**: `/vale-fix` fixes errors
only; here, fix warnings too, and leave suggestions for judgement. **Workflow**: ignore the
branch-and-pull-request steps. Commit discipline is the one in [`AGENTS.md`](AGENTS.md), which commits
only when told and never pushes.
