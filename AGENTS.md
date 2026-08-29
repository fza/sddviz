# AGENTS.md

Canonical guidance for `sddviz`, agent-agnostic. Claude Code loads it through the `@AGENTS.md` import
in [`CLAUDE.md`](CLAUDE.md), where anything Claude-Code-specific lives. **Work under `source/` loads
[`source/AGENTS.md`](source/AGENTS.md) too**, which carries the Go and TypeScript rules, and only
when a session touches that tree.

- **Never build to check that code compiles.** `go vet ./...` from `source/` and `npm run lint`
  inside `source/frontend` are the quick checks. Run tests to learn whether something works.
- **Never push.**

---

## Project overview

sddviz is a read-only desktop viewer for [SDD](https://github.com/networkteam/sdd) decision graphs,
built on Wails v3: React and TypeScript in the WebView, Go in the back end. It opens a repository
carrying an `.sdd/` graph and renders it as a filterable entry browser and a 3D graph view. sdd is
embedded as a library rather than called as a tool.

**Read [`docs/index.md`](docs/index.md) before implementing anything.** It indexes the documentation,
the repository structure and the module boundaries, and it stays in step with every structural change.

## Setup and commands

Devbox and direnv. `devbox.json` pins Go 1.26, Node.js 24, `golangci-lint` and `go-task`.
`scripts/dev-setup.sh` installs what Devbox cannot pin, and `--check` names what is absent without
installing anything.

```bash
devbox run dev            # the application, front end in watch mode
devbox run build          # a bundled application
devbox run build:server   # the same application as an HTTP server
devbox run bindings       # regenerate the typed front end bindings
devbox run test           # both halves
devbox run lint           # both halves, prose, then the repository checks
devbox run prose:lint     # prose alone - Vale over the repository
devbox run refs:lint      # the repository checks - see scripts/lint/README.md
devbox run tools          # install the Wails CLI into the Go bin directory
```

`scripts/git-hooks/commit-msg` refuses a commit message the rule refuses. `devbox`'s init hook
points `core.hooksPath` at it on first entry into the directory.

## Module boundaries

- **The Go module owns every read.** Opening a repository, reading its graph files, driving the
  embedded sdd and turning either into records all live there. The front end never touches a
  filesystem path and never runs a subprocess.
- **A Wails service is the seam**, and it carries data rather than rendered output.
- **`source/frontend` owns every pixel.** No presentation decision belongs in Go.
- **CGO is allowed.** A desktop build needs it for the WebView binding; a server build does not.
- **A second Go module anywhere in the tree needs explicit approval. BAIL, FLAG, PAUSE.**
- **Nothing carries a tag, so everything else is advisory.** Redesign a mechanic found structurally
  wrong, never freelance past it silently.

## Read-only is a hard boundary

- **sddviz never writes to a repository it reads.** Not the working directory, not the graph, not the
  index, not a lock file. The constraint is enforced where a repository is opened, never left to a
  caller to respect.
- **Never `git checkout`, and never run a git command against a user's repository at all.** Reads
  follow the working tree as they find it.
- **Only sdd's read surface is called.** `Search`, `View`, `Show` and `Info` on the embedded
  application are the whole vocabulary. Nothing that opens a session, starts a procedure, stages an
  attachment or captures an entry is ever called, and no call may write to the graph, the index, or
  a lock file.
- **Read-only is a property of the type, not a convention.** The `GraphStore` implementation refuses
  every write method, so no call path exists that could reach a reader's repository.

## What sddviz reads, and what sdd answers

sddviz embeds sdd and implements its `GraphStore` port. One walk of `.sdd/graph/**/*.md` builds the
snapshot; sdd answers everything derived from it.

- **Never rebuild what the tool owns.** Derived status, search across all three modes, the vector
  index under `.sdd/index/`, and linting live in sdd and are reached through the embedded
  `application`, never reimplemented. Reaching for a local version means capturing a signal, not
  writing one.
- **Status is never computed here.** No embedding reaches the derivation, since `Snapshot` exposes
  `Project()`, `Revision()` and `Health()` alone. It comes from the line sdd's view renders. sdd
  derives eight states from the whole graph and stores none of them, so a local rule is a guess
  wearing an answer's clothes.
- **Never parse a rendered entry.** sdd renders entries for a reader and carries no machine-readable
  format for them. Where rendered output is consumed at all, consume the entry line and the ID lists
  rather than the prose around them, and isolate that in one package naming the sdd version it reads.
- **Reach for sdd's ports before writing around them.** Package `application` documents the seams an
  embedding host implements, `local` holds the reference adapters, and `sddtest` proves an
  implementation conforms.
- **Schema knowledge is admitted, and it is bounded.** The store knows the entry format for a stated
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

- **Present tense, and never a claim sourced from a tracker, a ticket, or an id.** Where it would
  cite a decision, state the decision.
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
- **A new rule kind is a new section in the document that already owns that rule**, never a second
  document.
- **A retired document keeps its row in the index**, carrying the heading it had, saying it retired,
  saying why, and naming where its subject lives now.

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

### 5. Citation

- **Cite a document by its path, and a section by its heading quoted exactly.** A quoted title quotes
  the current wording, opening words onward. It may stop at a clause boundary and may never
  paraphrase. Rewording a heading breaks every citation quoting it, which `devbox run refs:lint`
  catches.
- **Never cite anything outside `docs/`.** Not a plan, a logbook, a tracker, a scratch file, or an
  id. Where the rule matters, state the rule.
- **No exception, anywhere: nothing in the tree carries an entry id.** Not a comment, a doc comment,
  a document under `docs/`, or a struct field. A refusal states what is missing in words, and the
  graph is where a reader finds the entry tracking it.

### 6. Prose is linted before it lands

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
  it replaces. `.vale/styles/Taxonomy/` belongs to this repository, so a change goes straight into it.
- **A banned-word rule carries a probe.** `devbox run refs:lint` writes lines the rule must report
  and lines it must leave alone, because a token that silently stops matching passes every document.
  Repair the token when a probe fires, never the probe.
- **The Project vocabulary takes terms, not rules**: a real term into `accept.txt`, a retired one into
  `reject.txt`, a typo fixed in the prose. It starts empty and earns each entry.

## Working agreements

- **Plan and execute are two separate gates.** Approving a plan is not authorization to implement it.
  Wait for "go ahead", "implement", "do it". Ambiguous: ask.
- **Breaking-change gate. BAIL, FLAG, PAUSE.** Before changing a document under `docs/`, or changing
  a service method's shape, stop, flag it, wait for confirmation.
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
  own signal before the turn ends. A comment asserting wiring nothing performs. A pointer to a path
  that no longer exists. A guard that cannot fail. A branch no test reaches.
- **Fix a small one on the spot, and capture that the fix landed.** Small means all four: about fifty
  lines of code or documentation, surfacing no signal of its own, testable where it is code, and
  changing neither how the product works nor how it should work. Query the graph for ground truth
  before starting one. Anything past those bounds stays captured and unfixed until the fix is the
  work.
- **Name what a reader needs to act**: what is wrong, where, and what it costs. Verify a measurement
  against the code and cite what proves it. Where a repair has more than one shape, state each and
  what each claims rather than picking silently.
- **Name the surface at decision time. An unnamed surface is an undecided one.** A service method, a
  configuration key, a cache path: "we'll add a method" is not settled, `BackendVersion` is. The
  entry carrying the decision carries the name, and it is the only place that name lives until the
  surface exists.

## The decision graph

`.sdd/` answers *why* the system is the way it is, where `docs/` answers *what* it is.

- **An entry's body is never edited, and no entry is ever deleted.** Supersede a decision, close a
  signal. Status is derived, never written into a file. Two stored fields are the tool's to change:
  `sdd rewrite` corrects an entry's type and kind and repoints inbound references, and
  `sdd summarize` replaces a summary, which is normal practice where one misrepresents its body.
- **Write every entry through `scripts/graph-capture.py`.** One draft file carries an entry's fields
  above its body, and the script lints, pre-flights and captures it. The `sddviz-sdd-capture` skill
  carries the draft shape. Never run `sdd new` by hand: `sdd new "$(cat body.md)"` re-evaluates the
  body in a second shell, so backticks in it run as command substitution and every identifier
  vanishes from the entry.
- **Fix every pre-flight finding above low severity. Low-only findings pass, and are never chased.**
  A finding that names a dropped criterion or an unstated transition is repaired in the draft, not
  overridden. `--force` captures anyway and is a judged override that says so in the body.
- **Lint a graph entry before capture. Nothing can lint it after.** The capture script runs Vale over
  the body as a real Markdown file, because piping a body into `vale -` reports nothing and a fenced
  draft reports nothing either.
- **An entry's generated summary is exempt from every rule, and only the summary.** Correct one that
  misleads a reader through `sdd summarize`. The body clears every finding first.
- **Query the graph before every capture, every direction taken, and every gathering of ground
  truth.** Search, read what looks close with `show`, then compile ground truth and decide.
- **A summary is a pointer.** Read the entries that matter in full with `show` before asserting what
  the graph holds.
- **Find entries with `sdd search`, never grep.** Reserve `grep` for source code. `--term` is plain
  text and always works; `--query` is semantic and needs a configured embedding endpoint. Prefer the
  semantic one wherever a plain-text search could miss an entry, and reach for `--term` where the
  target is a literal token already known.
- **Ask which search modes are available once, before the session's first search.** `sdd info`
  answers it: its search field reads `vector,text` wherever an embedding endpoint answers, `text`
  otherwise. One call settles it for the whole session.
- **Pick the sharp reference kind, and check its direction.** `refines` narrows a live decision.
  `required-by` records a prerequisite from the prerequisite's side, where `grounded-in` says this
  entry reasons from the target. `addresses` takes up an open gap without closing it. `related` is
  the floor and understates most links. Refs are immutable, so the moment to get this right is
  before the capture.

### Two kinds of topic label

| Kind | Form | Per entry | Answers |
|---|---|---|---|
| Facet | bare - `dependencies`, `stack`, `defect` | several | which concerns it touches |
| Area | prefixed - `area-scaffold` | exactly one | which body of work it belongs to |

Both share one namespace. The prefix tells them apart, so a facet label never takes it. **Area names
spell their words in full**, and **the separator is a hyphen, never a slash.**

- **Every entry raised carries its area at capture.** A label missed at capture costs a second entry
  to repair.
- **Work raising an entry that fits no area means the area is missing.** Name it, and say so in the
  same turn. Never park the entry outside every area to decide later.
- **An entry already written joins an area through an `annotation` signal**, whose topic clusters
  bind a label to a member list. Labels add up, so a second annotation extends and supersedes
  nothing.
- **The guard is `untagged`.** `sdd view --layout "untagged:as-list"` holds nothing but the procedure
  definitions and participant records the tool seeds. Anything else appearing there is an area nobody
  assigned, and it gets one.
- **Read the area before touching its code.** `sdd view --layout "topic(area-<name>):as-list"` is the
  whole membership, ranked.

**`sdd --version` before relying on any of this.** These rules measure 0.16.2.

## Memory

**Do not keep a parallel memory store for this project.** A rule about how the work happens here
belongs in this file. Anything one developer's own belongs in their gitignored `CLAUDE.local.md`.
The graph captures substance about sddviz itself. `.memory/` is gitignored scratch that carries a
decision only until the graph does.
