# sddviz

A desktop viewer for [SDD](https://github.com/networkteam/sdd) decision graphs. It opens a
repository that carries an `.sdd/` graph and shows what the project decided and why, through a
filterable entry browser with search and a 3D view of the graph itself.

sddviz is **read-only**. It never captures, never decides, and never writes to the repository it
reads.

Entry data is read from the graph files themselves, which store it as structured YAML. Search,
linting and anything else the files do not answer are run through the `sdd` binary, so nothing that
belongs to the tool is rebuilt here.

## Status

The repository holds the project scaffold: a Tauri application shell with a React and TypeScript
front end talking to a Rust back end. Neither the entry browser, the search, nor the 3D view exists
yet. `docs/index.md` indexes what is documented; anything unbuilt lives in the decision graph under
`.sdd/`.

## Why a viewer

SDD records a project's reasoning as an append-only graph of signals (what was noticed) and decisions
(what was committed to). The graph is read through an agent today, which serves a developer at a
keyboard well and serves everyone else poorly. A viewer puts the same reasoning in front of anyone
who needs it, without a terminal and without an agent session.

## Requirements

macOS, for now. Tauri builds Linux and Windows from the same source, and neither is exercised.

`sdd` on `PATH`, at version 0.16.2 or later, for search and linting. Browsing a graph works without
it, with those features absent.

The build toolchain is pinned in `devbox.json` ([devbox](https://www.jetify.com/devbox) plus
[direnv](https://direnv.net)). Entering the directory loads Node.js 24 and a Rust toolchain through
`rustup`. Building without devbox needs both installed by hand, plus the Xcode command line tools for
the macOS webview.

## Running

```bash
devbox run dev      # the application, with the front end in watch mode
devbox run build    # a bundled application
devbox run test     # front end and back end
devbox run lint     # type checking and clippy
```

Each script delegates into one of the two modules, so `cd source/ui && npm run tauri dev` does the
same thing. Bundling needs application icons, which are generated once with
`npm run tauri icon <path-to-png>` from `source/ui`.

## Structure

```
source/ui         React and TypeScript, the front end and everything the user sees
source/backend    Rust, the Tauri back end: it reads the graph, runs sdd, and answers the front end
docs              What the implementation is, as it stands
.sdd              The decision graph: why the implementation is the way it is
```

The two modules are peers. The back end owns every read, whether it comes from a file or from the
tool, and the front end never reaches a filesystem path or a subprocess directly; the seam between
them is the Tauri command interface.

## Limits

sddviz knows the entry format it reads, currently `graph_schema_version: 1`. A graph declaring a
version it does not know is reported to the user rather than guessed at.

Reads follow the repository's working tree. Selecting an arbitrary branch or an older revision is not
available: the tool takes a branch only as a session binding validated against a live checkout, and
sddviz will not check anything out.
