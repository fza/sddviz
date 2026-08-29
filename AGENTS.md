# AGENTS.md

Canonical guidance for `sddviz`, agent-agnostic. Claude Code loads it through the `@AGENTS.md` import
in [`CLAUDE.md`](CLAUDE.md), where anything Claude-Code-specific lives. **Work under `source/` loads
[`source/AGENTS.md`](source/AGENTS.md) too**, which carries the TypeScript and Rust rules, and only
when a session touches that tree.

- **Never build to check that code compiles.** `cargo check` inside `source/backend` and
  `npm run lint` inside `source/ui` are the quick checks. Run tests to learn whether something works.
- **Never push.**

---

## Project overview

sddviz is a read-only desktop viewer for [SDD](https://github.com/networkteam/sdd) decision graphs,
built on Tauri v2: React and TypeScript in the webview, Rust in the back end. It opens a repository
carrying an `.sdd/` graph and renders it as a filterable entry browser and a 3D graph view. Entry
data comes from the graph files; search and linting come from the `sdd` binary.

**Read [`docs/index.md`](docs/index.md) before implementing anything.** It indexes the documentation,
the repository structure and the module boundaries, and it stays in step with every structural change.

## Setup and commands

Devbox and direnv. `devbox.json` pins Node.js 24 and a `rustup` toolchain for both modules.

```bash
devbox run dev            # the application, front end in watch mode
devbox run build          # a bundled application
devbox run test           # both modules
devbox run lint           # both modules, prose, then the repository checks
devbox run prose:lint     # prose alone - Vale over the repository
devbox run refs:lint      # the repository checks - see scripts/lint/README.md
devbox run ui:test        # one module; ui:lint, ui:install, backend:test, backend:lint likewise
```

`scripts/git-hooks/commit-msg` refuses a commit message the rule refuses. `devbox`'s init hook
points `core.hooksPath` at it on first entry into the directory.

## Module boundaries

- **`source/backend` owns every read.** Opening a repository, reading its graph files, running `sdd`
  and turning either into records all live there. The front end never touches a filesystem path and
  never runs a subprocess.
- **The Tauri command interface is the seam**, and it carries data rather than rendered output.
- **`source/ui` owns every pixel.** No presentation decision belongs in Rust.
- **A new module under `source/` needs explicit approval. BAIL, FLAG, PAUSE.**

## Read-only is a hard boundary

- **sddviz never writes to a repository it reads.** Not the working directory, not the graph, not the
  index, not a lock file. The constraint is enforced where a repository is opened, never left to a
  caller to respect.
- **Never `git checkout`, and never run a git command against a user's repository at all.** Reads
  follow the working tree as they find it.
- **Only sdd's read surface is invoked.** `search`, `view`, `show`, `lint` and `info` are the whole
  vocabulary. Nothing that opens a session, starts a procedure, stages an attachment or captures an
  entry is ever run, and no invocation may write to the graph, the index, or a lock file.
- **A capability the viewer does not need is not granted.** `source/backend/capabilities/` stays at
  what the window actually uses.

## What is read from a file, and what is asked of the tool

Entry data is structured on disk and is read straight from `.sdd/graph/**/*.md`. Everything the files
do not answer is asked of `sdd`. The split is not a preference, it is which half fails how.

- **Never rebuild what the tool owns.** Derived status, search across all three modes, the vector
  index under `.sdd/index/`, and linting exist in sdd and are invoked, never reimplemented. Reaching
  for a local version means capturing a signal, not writing one.
- **Status is never computed here.** It comes from one `sdd view --layout "source(graph):as-list"`
  call. sdd derives eight states from the whole graph and stores none of them, so a local rule is a
  guess wearing an answer's clothes.
- **Never parse a rendered entry.** sdd renders `show` for a reader and carries no machine-readable
  format for it. Where the tool's output is consumed at all, consume the ID lists rather than the
  prose around them, and isolate that in one module pinned to the version it reads.
- **Schema knowledge is admitted, and it is bounded.** The reader knows the entry format for a stated
  `graph_schema_version` and reports an unknown one rather than guessing.
- **A view is cut before its data is invented.** Nothing is shown that neither the files nor the tool
  can support.

## Authoring

### 1. Code comments, the minimum that earns its place

- **A file with no comments is correct** where the code is self-explanatory. Density is never a
  quality signal.
- **Write one only for what the code cannot say**: a rule imposed from outside the file, a mechanism
  a reader would otherwise reconstruct, a rejected alternative, an invariant a future edit would
  break silently.
- **Never restate the code.** Delete such a comment rather than improving it.
- **HOW and WHY, never HOW IT CAME TO BE.** Reference no discussion, session, or review.

### 2. Every document states the what and the why

**Never the HOW IT CAME TO BE.** Write the current state of the world and its reasoning as a fact
standing on its own. Rejected alternatives and the reasoning that discarded them belong in the
record; the order in which the work happened does not.

### 3. `docs/` describes the present implementation, and nothing else

- **Every sentence holds true of the code as it stands right now.** Nothing here is a design, a
  proposal, a plan, or a promise.
- **Unbuilt, undecided and uncovered belong to the decision graph, and nowhere else.** A document
  naming a thing that does not exist is wrong, not incomplete.
- **A status claim names what makes it true**, the symbol, file, or command carrying it.
- **Documentation and implementation move together, in one pass.** Work that builds, moves, renames
  or removes anything a document describes stays unfinished until that document matches the tree.
- **Section numbers are permanent.** Append or letter, never renumber.
- **Mark an obsolete section, never remove it.** The heading stays and says it is obsolete, why, and
  where the rule went.

### 4. Vocabulary

**[`docs/rules.md`](docs/rules.md) fixes the vocabulary.** It binds all prose - documents, commit
messages, code comments, sentences written to a reader - and every developer-facing value. It does
not bind internal identifiers, and its §1 carries the scoping table.

- **Its term, exact form, every single time.** `.vale/styles/Taxonomy/` carries the words, never a
  list here. Banned and reported as **errors**: the competitors to `entry`. Explicit and reported as
  **suggestions**: `node`, `tag`, `label`.
- **SDD's words are used as SDD spells them**, deliberately, for everything the graph holds. That
  licence stops at the graph; everything sddviz builds of its own takes a name the rules document
  decides.
- **Banned words stay banned. Never coin a competing term.** Needing a name the document lacks means
  changing the document.
- **In doubt: STOP AND ASK.**
- **Correct on sight**, outranking consistency with surrounding text.

### 5. Prose is linted before it lands

**Vale decides whether prose ships.** `.vale.ini` covers every Markdown file and the **comments** in
every language it parses, carving out `.sdd/` and `.claude/`, which belong to their tools, and the
generated trees. `devbox run prose:lint` is the repository's invocation and runs as part of
`devbox run lint`. Act on `vale --output=JSON <path>` for a single file.

`vale` comes from `PATH` rather than from `devbox.json`, because the build in use is ahead of the one
nixpkgs carries.

- **New prose clears every `error`, `warning` and `suggestion`**: a fresh document, a line added to an
  old one, a graph entry body, a commit message, a code comment.
- **Leave the alerts prose already in the repository carries.** Clearing those is its own task and
  needs its own ask.
- **A hook holds the bar.** `.claude/hooks/vale_new_prose.py` runs after each write, reports only
  alerts on lines that write added, and skips a path git ignores.
- **`git ls-files` drives the lint, so an untracked new file is invisible to it.** Stage before
  linting.
- **Never silence a rule** beyond a wrapped single-line exemption naming one rule and switching it
  back on immediately: a quoted defect, a quoted claim in someone else's voice. Never a whole-file
  exemption.
- **Backticks come first in a code comment**, exactly as in Markdown: an identifier, a flag, a path, a
  wire value. A worked example goes in a fenced block.
- **Wrap a passage quoted verbatim in a directive**, the one thing backticks cannot cover. Cover the
  passage alone and switch the linter back on directly after it. `<vale off>` opens a region and
  `</vale off>` closes it; one rule alone is `<vale House.Passive = NO>`. Write it on its own line
  inside the comment it acts on, never as a comment of its own.
- **Never change meaning to satisfy style.** Leave the alerts and say so.
- **Replace a rule that misreads this repository, never edit it in place.** Write the replacement
  under `.vale/styles/House/`, name the rule it derives from and what it changed, and delete the one
  it replaces.
- **The Project vocabulary takes terms, not rules**: a real term into `accept.txt`, a retired one into
  `reject.txt`, a typo fixed in the prose. It starts empty and earns each entry.

## Working agreements

- **Plan and execute are two separate gates.** Approving a plan is not authorization to implement it.
  Wait for "go ahead", "implement", "do it". Ambiguous: ask.
- **Breaking-change gate. BAIL, FLAG, PAUSE.** Before changing a document under `docs/`, or changing
  a Tauri command's shape, stop, flag it, wait for confirmation.
- **Commit messages are one line, Conventional Commits, at most 100 characters**, with no body and no
  trailers: `feat:`, `fix:`, `docs:`, `chore:`, `refactor:`, `test:`, `perf:`, `build:`, `ci:`.
- **Never push.**
- **The `sdd:` prefix belongs to that tool alone.** Never write it by hand, never amend or reword a
  commit carrying it, never read it as a Conventional Commits type.
- **Working-tree safety.** Before any bulk or destructive git operation, run `git status` and
  `git diff` scoped to the files this turn touched. Revert surgically with `git checkout -- <files>`,
  never a sweeping `-A` or `.`.
- **Verify delegated work.** A subagent's summary is what it intended, not what it did. Spot-check
  the diff and re-derive its claims. Coverage figures and `file:line` references go wrong most often.
- **Raise a gap or a defect the moment you see it, whatever you were doing**, and capture it as its
  own signal before the turn ends.
- **Name the surface at decision time. An unnamed surface is an undecided one.** A Tauri command, a
  configuration key, a cache path: "we'll add a command" is not settled, `graph_entries` is. The
  entry carrying the decision carries the name, and it is the only place that name lives until the
  surface exists.

## The decision graph

`.sdd/` answers *why* the system is the way it is, where `docs/` answers *what* it is.

- **An entry's body is never edited, and no entry is ever deleted.** Supersede a decision, close a
  signal. Status is derived, never written into a file.
- **Query the graph before every capture, every direction taken, and every gathering of ground
  truth.** Search, read what looks close with `show`, then compile ground truth and decide.
- **Find entries with `sdd search`, never grep.** Reserve `grep` for source code. `--term` is plain
  text and always works; `--query` is semantic and needs a configured embedding endpoint.
- **Every entry belongs to exactly one area**, a topic label opening with `area-`, naming the body of
  work the entry is part of. Facet labels are bare and an entry may carry several.
- **`sdd --version` before relying on any of this.** These rules measure 0.16.2.

## Memory

**Do not keep a parallel memory store for this project.** A rule about how the work happens here
belongs in this file. Anything one developer's own belongs in their gitignored `CLAUDE.local.md`.
The graph captures substance about sddviz itself. `.memory/` is gitignored scratch that carries a
decision only until the graph does.
