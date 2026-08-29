# Documentation index

What sddviz is, as it stands. Anything unbuilt, undecided or uncovered lives in the decision graph
under `.sdd/`, never here.

## Repository structure

| Path | Holds |
|---|---|
| `source` | The Go module: the Wails application and every read of a graph. |
| `source/frontend` | The web application: React 19 and TypeScript, built by Vite, embedded by the module. |
| `docs` | This set. |
| `.sdd` | The decision graph. |

The web application sits inside the Go module because `go:embed` cannot reach outside its own package
directory. Its build output lands in `source/frontend/dist`, which the module embeds.

## Module boundaries

- The front end reaches the back end only through generated bindings over Wails services, and reaches
  no filesystem path and no subprocess directly.
- Services carry data. Rendering happens in the front end.
- cgo is allowed. A desktop build needs it for the WebView binding; a server build does not.

## The macOS build

`source/Taskfile.yml` and `source/build/` carry the build system. `source/build/config.yml` is the
single place the bundle's identity is written, and `wails3 task common:update:build-assets`
regenerates `Info.plist` from it. The identifier is `com.fza.sddviz`.

A desktop build links the system WebView through cgo, which pulls in `-lresolv`. Built inside a
package manager's environment, that resolves against the manager's own copy of the library, and dyld
refuses to load it under the hardened runtime because the process and the mapped file carry
different Team IDs. The bundle step repoints such a library at its system copy and then fails the
build if anything outside `/usr/lib` or `/System` remains, so the failure lands at build time rather
than at launch.

## Reading a graph

sdd is embedded as a library. The module implements sdd's `GraphStore` port, which package
`application` documents as an infrastructure port for an embedding host: one walk of
`.sdd/graph/**/*.md` builds an `application.SnapshotData`, whose `EntryDocument` values carry the
frontmatter as structured data, and the same snapshot is handed to `application.Application` for
search and views.

Read-only is enforced there. The store's write methods, `Apply` and `Reconcile` in sdd's port,
refuse, so no call path exists that could write to a reader's repository.

Two properties of sdd's read surface bound what the viewer offers:

- **Derived status is not reachable structurally.** sdd's `Snapshot` exposes `Project()`,
  `Revision()` and `Health()` and nothing further, so status is read from the line its view renders,
  whose shape sdd's own unit tests hold.
- **Reads follow the working tree.** The read surface takes no branch or revision argument. A branch
  is a session binding validated against a live registered checkout, and sddviz checks nothing out,
  so browsing an arbitrary ref is not offered.

## Documentation sets

No implementation set exists yet. The application is a shell: `source/internal/graph` registers one
service method, `BackendVersion`, and `source/frontend/src/App.tsx` calls it through its generated
binding to show that the seam between the two halves works.
