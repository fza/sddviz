# Documentation index

What sddviz is, as it stands. Anything unbuilt, undecided or uncovered lives in the decision graph
under `.sdd/`, never here.

## Repository structure

| Path | Holds |
|---|---|
| `source/ui` | The front end: React 19 and TypeScript, built by Vite. Everything the user sees. |
| `source/backend` | The back end: a Rust crate carrying `tauri.conf.json`. Every read, from the graph files and from `sdd`. |
| `docs` | This set. |
| `.sdd` | The decision graph. |

The two modules under `source/` are peers rather than a package and its subdirectory, because the
back end owns every read and the records it answers with in its own right.

Tauri resolves the front end directory to the first `package.json` at or above the working
directory. No such file sits at or above `source/backend`, so `TAURI_FRONTEND_PATH` and
`TAURI_APP_PATH` are set by the `tauri` script in `source/ui/package.json` rather than left to the
operator.

## Module boundaries

- The front end reaches the back end only through Tauri commands, and reaches no filesystem path and
  no subprocess directly.
- Commands carry data. Rendering happens in the front end.
- `source/backend/capabilities/default.json` grants `core:default` and nothing further. The viewer
  needs no filesystem or shell permission of its own.

## Reading a graph

Entry data is read from `.sdd/graph/**/*.md`, where an entry is YAML frontmatter above a markdown
body. Search and linting are run through the `sdd` binary on `PATH`, addressed at a repository with
the global `--graph-dir` flag. Browsing works without the binary, with those features absent.

Two properties of the tool's read surface bound what the viewer offers:

- **Answers are rendered for a reader.** No graph read carries a machine-readable format;
  `--format json` exists on `config` and `stats` alone. Only the ID lists in that output are consumed.
- **Reads follow the working tree.** The read surface takes no branch or revision argument. A branch
  is a session binding validated against a live registered checkout, and sddviz checks nothing out,
  so browsing an arbitrary ref is not offered.

## Documentation sets

No implementation set exists yet. The application is a shell: `source/backend/src/lib.rs` registers
one command, `backend_version`, which returns the crate version, and `source/ui/src/App.tsx` calls it
to show that the seam between the modules works.
