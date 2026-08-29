# AGENTS.md - `source/`

Code rules, loaded when a session touches anything under `source/`. The
[root `AGENTS.md`](../AGENTS.md) binds here unchanged: module boundaries, read-only, authoring,
working agreements, commands. This file restates none of it.

**Run the tests to check that code works. Never build for it.** `cargo check --all-targets` inside
`source/backend` and `npm run lint` inside `source/ui` are the quick compile checks.

## Layout

- **`source/ui` and `source/backend` are peers**, and neither imports from the other's tree. The Tauri
  command interface is the only path between them.
- **`TAURI_FRONTEND_PATH` and `TAURI_APP_PATH` in `source/ui`'s `tauri` script decide where Tauri
  looks**, and removing either breaks the build. Tauri resolves the front end to the first
  `package.json` at or above the working directory, and none sits at or above `source/backend`.
- **`source/backend/icons/` is required to compile.** `tauri::generate_context!` reads
  `icons/icon.png` at macro expansion time. Regenerate the set from `icons/app-icon.png` with
  `npm run tauri icon ../backend/icons/app-icon.png`.

## Comments

**[Root `AGENTS.md` → Authoring §1](../AGENTS.md#1-code-comments-the-minimum-that-earns-its-place)
binds and is not repeated here.** In brief: the bare minimum that earns its place, HOW and WHY and
never how it came to be. On top of it:

- **Write no comment per function, type, or constant.** A well-named symbol needs none, and a file
  with no comments at all is correct. Delete a comment restating a signature or an obvious branch on
  sight.
- **A module reading sdd's output or its file format names the sdd version it was written against**,
  because neither is a stable contract and a reader cannot otherwise tell what it was checked on.

## TypeScript

- **Single quotes.** Strings are interpolated rather than concatenated.
- **`strict` stays on**, along with `noUnusedLocals`, `noUnusedParameters` and
  `verbatimModuleSyntax`. Never widen the compiler options to make code pass; change the code.
- **Never `any`.** `unknown` at a boundary, narrowed immediately.
- **The front end holds no knowledge of the graph format.** It renders records the back end hands it
  and knows nothing of frontmatter, file paths, or how a status was arrived at.
- **Descriptive names.** No single-letter identifiers outside a bounded generic.

## Rust

- **Assign, then check.** A binding and a match on it read better than a chained expression doing
  both.
- **`?` over `unwrap`.** `unwrap` and `expect` belong in tests and in `main`'s startup path, nowhere
  else. Where `expect` is unavoidable, its message names what failed, not that something failed.
- **Errors carry context.** An error crossing a module boundary says what operation failed and on
  what, and wraps rather than replaces its cause.
- **Never log an error and then return it.** Handle it, or return it with context added.
- **A Tauri command returns a typed result and never a rendered string.** Presentation belongs to the
  front end.
- **`cargo clippy --all-targets -- -D warnings` is the bar**, and it runs in `devbox run lint`.

## Writing tests

**All code written here is covered by tests, and the tests are run before the work is reported done.**

- **Prefer table-driven tests.** In Rust a slice of cases with a loop and a named case; in TypeScript
  a `describe` with a case array.
- **Test against real captured material, never a hand-written imitation of it.** A fixture graph is
  real entry files, and a test over sdd's output is over output a real invocation produced. An
  imitation passes while the thing it stands for has changed.
- **Break every guard, watch it fail, restore the tree byte-identically, and say so.** A guard that
  cannot fail is worthless.
- **Never repeat expected values in assertion messages.** Describe the expected behaviour.
- **A test never writes to a repository under test.** Read-only is the product's contract, and a
  suite that violates it proves nothing about the shipped path.

## Dependencies

- **A new dependency in either module is a decision, and it gets captured.** Name what it does, what
  it replaces, and what it costs.
- **Nothing is added that reimplements what `sdd` already does.** Search, linting and the vector index
  are the tool's.
