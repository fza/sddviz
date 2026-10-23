# AGENTS.md - `source/`

Code rules, loaded when a session touches anything under `source/`. The
[root `AGENTS.md`](../AGENTS.md) binds here unchanged: module boundaries, read-only, authoring,
working agreements, commands. This file restates none of it.

**Run the tests to check that code works. Never build for it.** `go vet ./...` from `source/` is the
quick compile check, and `npm run lint` inside `source/frontend` is its front end counterpart.

## Layout

`source/` is the Go module. `source/frontend` is the web application inside it, which is what
`go:embed` requires: it cannot reach outside its own package directory, so the built assets live in
`source/frontend/dist` and the module embeds them from there.

- **No `/pkg`.** Every non-`main` package lives under `internal/`.
- **cgo is allowed**, and a desktop build needs it for the WebView binding. A server build
  (`go build -tags server`) does not.
- **`source/build/appicon.png` is the icon source.** `wails3 task common:generate:icons` derives the
  platform icon sets from it.
- **The generated bindings are committed.** `wails3 generate bindings` writes
  `source/frontend/bindings`, and committing them lets the front end build without the Go toolchain.
  Never edit one by hand.
- **`source/build/` and `source/Taskfile.yml` are the framework's template**, vendored rather than
  written here. The linters skip them. An edit to one carries a comment saying why, as the darwin
  build and icon tasks do.

## Comments

**[Root `AGENTS.md` → Authoring §1](../AGENTS.md#1-code-comments-the-minimum-that-earns-its-place)
binds and is not repeated here.** In brief: the bare minimum that earns its place, HOW and WHY and
never how it came to be. Go-specific on top:

- **Write no comment per function, variable, or constant.** A well-named symbol needs none, and a
  file with no comments at all is correct. Delete a comment restating a signature, a name or an
  obvious branch on sight.
- **Put reasoning that needs paragraphs in `docs/` and cite it from the comment.** Prefer updating
  that document over commenting at all.
- **Put a package comment above the `package` clause of a real source file**, never in a file holding
  nothing else.
- **Write no doc comment per exported identifier under `internal/`.** `.golangci.yml` enables
  `revive`'s `var-naming` alone, never `exported`.
- **A package reading sdd's output, its file layout, or implementing one of its ports names the sdd
  version it was written against.** None of the three is a stable contract, and a reader cannot
  otherwise tell what it was checked on.

## Error handling

**Assign, then check. Never inline.**

```go
err := foo()
if err != nil {
    return err
}
```

Not:

```go
if err := foo(); err != nil {
    return err
}
```

**Never log an error and then return it.** Handle it, or return it wrapped.

Wrong:

```go
if err != nil {
    logger.ErrorContext(ctx, "Something failed", "error", err)
    return err
}
```

Return with context:

```go
if err != nil {
    return fmt.Errorf("something failed: %w", err)
}
```

Handle locally, return nothing:

```go
if err != nil {
    logger.ErrorContext(ctx, "Non-critical operation failed, continuing", "error", err)
}
```

- **`log/slog`, never `log`.** `.golangci.yml`'s `depguard` denies the latter.
- **A service method returning an error hands the front end something a reader can act on.** Name
  what failed and on what; never surface a raw path or a stack.

## TypeScript

- **Single quotes.** Strings are interpolated rather than concatenated.
- **`strict` stays on**, along with `noUnusedLocals`, `noUnusedParameters` and
  `verbatimModuleSyntax`. Never widen the compiler options to make code pass; change the code.
- **Never `any`.** `unknown` at a boundary, narrowed immediately.
- **The front end holds no knowledge of the graph format.** It renders what the service hands it and
  knows nothing of frontmatter, file paths, or how a status was arrived at.
- **Call a service through its generated binding**, never by composing a call by name.

## Writing tests

**Coverage is a quality mechanism here, never a nice-to-have.** All code written here is covered, and
the tests are run before the work is reported done.

- **`-race` is never optional.** Every test run carries it.
- **Aim for real-logic coverage, and document an accepted ceiling instead of forcing a fragile
  test.** A ceiling is a known gap, not a failure. An operating-system boundary and a trivial
  one-liner in an otherwise covered file each have a real one.
- **Exclude last.** Prefer a real test through an existing seam, or a white-box in-package test. Two
  things are excludable by default and nothing else is: the `main` entrypoint, and generated code.
- **Never instrument the thing under test to make a number go green.** Production surface existing
  only for a test is a signal to exclude or defer, and so is fragile process-level manipulation in
  the harness. Write an accepted-gap record, never the workaround. The tell is a test fix reaching
  for global process state rather than the package under test.
- **Prefer table-driven tests, structured with `t.Run` subtests.**
- **Name a test file inside the package under test `*_internal_test.go`. Every other is `*_test.go`,
  declaring the external `<pkg>_test` package.** `testpackage` in `.golangci.yml` enforces it, so
  reaching for unexported symbols is deliberate and the filename declares it.
- **`sddtest.RunGraphStoreTests` does not apply to this store.** The suite publishes an entry and
  asserts it landed, so a store with no write method fails its first assertion. Three tests prove
  this store instead. One builds a snapshot from a fixture graph and drives sdd's own `Search` and
  `View` over it, the same path the application uses. One checks that the store implements no write
  port. One checks that the access resolver refuses every write and every session.
  `sddtest.RunEmbedderTests` does apply, and the embedder runs it.
- **NEVER inline a fixture graph as a Go string literal.** Entry files, `config.yaml` and
  `meta.json` are real files under `testdata/`, loaded by a helper. A graph written as Go strings
  drifts from the format sdd actually writes, and it passes while the thing it stands for has
  changed. A one-off override on top of a loaded fixture is fine: the rule is against whole fixture
  trees as Go strings.
- **Test against real captured material.** A fixture graph is real entry files, and a test over sdd's
  rendered output is over output a real call produced.
- **Put test support code in one place per package tree**, never duplicated inline across
  `*_test.go` files. A helper genuinely local to one package's tests may stay in its `_test.go`: the
  rule is against duplication, and against test infrastructure hiding in test files.
- **Break every guard, watch it fail, restore the tree byte-identically, and say so.** A guard that
  cannot fail is worthless. Any test asserting an invariant has to show it fails when the invariant
  breaks.
- **Never repeat expected values in assertion messages.** Describe the expected behaviour ("values
  should be changed", not "value should be 200").
- **A test never writes to a repository under test.** Read-only is the product's contract, and a
  suite that violates it proves nothing about the shipped path.

## Dependencies

- **A new dependency is a decision, and it gets captured.** Name what it does, what it replaces, and
  what it costs.
- **Nothing is added that reimplements what sdd already does.** Search across all three modes, the
  vector index, and linting are the tool's, reached through the embedded `application`.
- **Reach for sdd's ports before writing around them.** Package `pkg/application` documents the seams
  an embedding host implements, `pkg/local` holds the reference adapters, and `pkg/sddtest` proves an
  implementation conforms.
