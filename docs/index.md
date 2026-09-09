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
refuse, so no call path exists that could write to a reader's repository. That also puts sdd's own
graph-store conformance suite out of reach, since it asserts a mutation lands.

Two properties of sdd's read surface bound what the viewer offers:

- **Derived status is not reachable structurally.** sdd's `Snapshot` exposes `Project()`,
  `Revision()` and `Health()` and nothing further, so status is read from the line its view renders,
  whose shape sdd's own unit tests hold.
- **Reads follow the working tree.** The read surface takes no branch or revision argument. A branch
  is a session binding validated against a live registered checkout, and sddviz checks nothing out,
  so browsing an arbitrary ref is not offered.

## What the module reads today

`source/internal/graph` carries the store. `Open` takes a project identity and a graph directory;
`Current` walks it once, builds the documents sdd needs, and hands back the snapshot sdd builds from
them. `Documents` returns those same documents, which is what the views will draw, and `Unreadable`
names any document that would not decode, so one malformed file leaves the rest of a graph readable
rather than shortening the list silently.

`Apply` and `Reconcile` return `ErrReadOnly`. A document whose path is not an entry path is skipped
before it reaches sdd, which rejects one it cannot turn into an entry ID.

The tests run against a fixture of real entry documents. One drives sdd's own `View` and `Search`
over the store, which is what shows the port is satisfied rather than merely compiled against.

## Opening a repository

A repository is a checkout carrying `.sdd/config.yaml`, whose `graph_dir` names the graph inside it
and defaults to `.sdd/graph`.

The first command-line argument names a repository to open at once: `sddviz <path>`. `Startup()`
reports what that opened, or why it did not open, and reports that nothing was requested where the
argument was absent. It is the only way in where no window system offers a directory picker.

Two further service methods open one on demand:

- `OpenRepository(path)` reads that configuration, opens the store over the graph directory, and
  reports the name, the path, the graph directory, the repository identity, the revision, and how
  many entries and undecodable documents it found. Every build answers it. The name is the directory
  the graph lives in, split off here because a path is the front end's to display and never to take
  apart.
- `ChooseRepositoryDirectory()` opens the native folder picker and returns what the reader chose,
  without opening anything. Desktop builds alone answer it; elsewhere it reports that choosing a
  directory needs a window system, and the path arrives another way.

A repository declaring a `repo_id` is identified to sdd by it, so every clone shares one identity.
One declaring none is local, and its path is the only identity available.

## What the front end shows

A bar across the top names the open repository and the revision its graph reads at. It keeps its
top-left corner clear only where the window controls sit over it, which the front end settles once at
startup from what the framework's runtime reports. Below it a sidebar
names the two views and the main area draws the selected one. Opening a repository sits in the
sidebar's footer, beside the path, the entry count and the link count of the one that is open. A
panel on the right shows whichever entry is selected, and both views share that selection.

- **Entries** is a table: identifier, layer, kind, derived status, summary, and a control drawing
  that entry in the graph. It scrolls in its own region, a row selects, and the selected row follows
  a selection made elsewhere. It carries no filters.
- **Graph** draws the whole graph through `3d-force-graph` over three.js. An entry is one shape,
  coloured by type and greyed where the graph has retired it, sized by how many links reach it. A
  link is one line, coloured by its kind and pointing at its target. Repulsion reaches a bounded
  distance, so an entry group linking to nothing else stays beside the body it belongs to rather than
  drifting past everything.
- **The selection is drawn as a halo**: red on the selected entry, purple and smaller on what links
  to it, purple and smaller again on what links to those. The camera flies to the selection, which a
  view of the whole graph draws too small to find by eye. With nothing selected, the settled layout
  is framed whole.
- **The panel** carries the entry's type, layer, kind, intent and derived status, its summary, its
  topics with the area marked, its participants and confidence, the links it declares and the links
  that reach it, the identifiers it names that the graph does not hold, and its body. Every link is a
  control selecting the entry at its other end.

`Entries()` reports one record per entry, newest first, since an entry's identifier opens with the
moment it was written. Each record carries the frontmatter fields, the links the file declares, and
the derived status, with the live head named where an entry is closed or superseded. Status comes
from one `View` call over `source(graph):as-list`, matched to entries by identifier, and the line it
is read from is the one sdd's own unit tests hold.

The front end builds the drawn network from those records. A link naming an entry the graph does not
hold, and a link an entry declares to itself, are drawn nowhere and reported in the panel instead.

An entry body is Markdown, hard wrapped by the tool that wrote it. A block of running prose is
rejoined into one paragraph, and a block carrying a heading, a list, a quotation, a table or a fence
keeps its lines as written.

Opening a repository mounts sdd over the store. Its session and staged-blob stores sit in scratch the
viewer owns, because sdd requires both and a read writes to neither.

## Documentation sets

No implementation set exists yet.
