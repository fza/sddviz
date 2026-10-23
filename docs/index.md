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

sdd is embedded as a library. The module embeds the sdd fork at `github.com/fza/sdd`, release
`v0.18.0-alpha.2.fza.24`, through a `replace` in `source/go.mod`. That build writes the title that
names each entry. The module implements sdd's `GraphStore` port, which package
`pkg/application` documents as an infrastructure port for an embedding host: one walk of
`.sdd/graph/**/*.md` builds an `application.SnapshotData`, whose `EntryDocument` values carry the
frontmatter as structured data, and the same snapshot is handed to `application.Application` for
search and views.

Read-only is enforced there. sdd's `GraphStore` port holds `Current` and `ReadAttachmentPage` alone,
and a write travels through a separate mutation port the store does not implement. The runtime names
no default branch and no target acquirer. The access resolver refuses write access and every
session. No call path exists that could write to a reader's repository. sdd's own graph-store
conformance suite asserts a mutation lands, so it does not apply here. Its embedder suite does, and
the viewer's embedder runs it.

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

The store has no write method. A document whose path is not an entry path is skipped before it
reaches sdd, which rejects one it cannot turn into an entry ID.

The tests run against a fixture of real entry documents. One drives sdd's own `View` and `Search`
over the store, which is what shows the port is satisfied rather than merely compiled against.

## Opening a repository

A repository is a checkout carrying `.sdd/config.yaml`, whose `graph_dir` names the graph inside it
and defaults to `.sdd/graph`.

The first command-line argument names a repository to open at once: `sddviz <path>`. `Startup()`
reports what that opened, or why it did not open, and reports that nothing was requested where the
argument was absent. It is the only way in where no window system offers a directory picker.

Two further service methods open one on demand:

- `OpenRepository(path)` reads that configuration and the local one beside it, opens the store over the graph directory, and
  reports the name, the path, the graph directory, the repository identity, the revision, and how
  many entries and undecodable documents it found. Every build answers it. The name is the directory
  the graph lives in, split off here because a path is the front end's to display and never to take
  apart.
- `ChooseRepositoryDirectory()` opens the native folder picker and returns what the reader chose,
  without opening anything. Desktop builds alone answer it; elsewhere it reports that choosing a
  directory needs a window system, and the path arrives another way.

A repository declaring a `repo_id` is identified to sdd by it, so every clone shares one identity.
One declaring none is local, and its path is the only identity available.

`.sdd/config.local.yaml` overlays `.sdd/config.yaml` field by field, which is where a machine keeps
the embedding endpoint and key that no checkout shares. Both overlay sdd's machine-wide file,
`$XDG_CONFIG_HOME/sdd/config.yaml` (or `~/.config/sdd/config.yaml`), in the order sdd itself reads
them. One embedding setting there reaches every repository on that machine. `UserConfigPath()` names
that file, and the application hands it to the service when it starts.

**A committed file never chooses where a key goes.** The viewer ignores an embedding `endpoint` and
`api_keys` in `.sdd/config.yaml` and takes both from the two files a machine owns. Whoever commits
to a repository writes that file, and a key from a reader's machine would otherwise travel to
whatever server it named. A team sharing a self-hosted endpoint names it in each machine's own file.

## Searching a graph

`Search(phrase)` reports the entries sdd matched, in the order sdd ranked them, alongside the mode
that produced them. Comparing entries against a phrase happens in sdd, never here.

The mode comes from sdd's own `Info`, which answers `vector,text` where the mount carries both an
embedding provider and a stored index, and `text` alone otherwise. A phrase reaches sdd's vector
search in the first and its term search in the second. A request names an explicit limit, since sdd
answers with ten where none is named.

**The word count decides the mode.** One word is what a reader types to find a token they already
know, and matching it literally beats the nearest neighbours of its meaning, so a single word always
searches over terms. Several words describe a thing rather than name it, which is what a vector space
answers well, so they reach the vector search wherever one is available.

**A score ranks one match against the strongest of the same search.** sdd renders it on the citation
line beneath each entry it ranked, normalised so the top of every search reads 100. It orders matches
within one search and says nothing about how good any of them is on its own. A request asks for one
citation per entry, the strongest part of it that matched, because a request asking for none renders
no citation line and no score with it.

**Filtering by topic happens in the front end**, over the records already in hand. sdd's search
filters by type, layer and kind, and carries no topic filter to reach for.

**The topics on offer are those still worth choosing.** Each one listed is carried by an entry that
survives the current choice, so choosing one never empties the table, and a topic leading nowhere is
absent rather than shown and refused. A chosen topic stays listed so a reader can undo it.

**Done signals are listed until asked away.** A control beside the search field drops the entries
recording that work happened, which is what a reader following what is still open reads without. The
topics on offer follow that choice like every other.

**A search starts from what it found.** It clears the chosen topics and the sort, because a topic
chosen against the whole graph rarely fits what a search returns, and a sort left in place would
replace the ranking before a reader ever saw it.

**The viewer speaks one embedding transport**, the `POST {endpoint}/v1/embeddings` contract that
OpenAI and its compatible self-hosted services answer. A repository configuring `openai` with a
model and a key gets vector search. One configuring any other provider searches over terms, which is
an answer rather than a failure.

**The stored index sits in the machine's cache, never in a repository.** It lives under
`$XDG_CACHE_HOME/sdd` (or `~/.cache/sdd`), keyed by the repository identity and by a fingerprint of
the embedder, which is the same layout sdd's own tool fills. Reproducing that fingerprint exactly is
what lets the viewer read the index that tool already built rather than embedding every entry again
to fill a second one, and a test holds the fingerprint to the string sdd derives.

A search asks sdd to reconcile the open graph's index first (`SearchSyncLocal`), embedding whatever
is missing and writing it to that cache. An index already current is left alone. Read-only is untouched either way, because the
cache is not the repository the viewer reads.

## What the front end shows

A bar across the top names the open repository and the revision its graph reads at, and carries the
control letting go of it. It keeps its top-left corner clear only where the window controls sit over
it, which the front end settles once at startup from what the framework's runtime reports. The bar
names nothing while nothing is open.

Below it a segmented control chooses the view, and the main area draws the chosen one. A repository
opens through a control the main area centres, which is the only thing offered while nothing is open.
Once a repository opens, Entries and Graph replace it, and Entries is where a reader lands, because a
list of what a graph holds reads straight away. A panel on the right shows whichever entry is
selected, and both views share that selection.

- **Entries** is a table: a counter, the moment the graph received the entry, the rest of its
  identifier, layer, kind, derived status, title, and a control drawing that entry in the graph.
  An identifier opens with that moment and closes with what the entry is, so the two sit in columns
  of their own. Selecting a heading sorts, and selecting the sorted one again turns it around and then
  clears it. Sorting is offered over every entry alone: a search answers in the order sdd ranked it,
  and reordering that throws the ranking away. Every column sorts except the counter and the
  title, which orders by nothing a reader wants. A search adds a score column, dropped again when the
  search clears.

  A width fixes each column, and pulling the edge of a heading changes it. A column stops at the
  width its own heading needs rather than closing to nothing, and the widths outlive a sort, a
  search, and a switch between the views. A search leaves every one of them where it was and adds the
  score beside them. Every column a reader reads carries a width, the title among them, and the
  control drawing an entry in the graph keeps its own size and its place at the right edge. An
  unnamed column between them takes up whatever the rest leave. A floor under the table's width makes
  a narrow view scroll sideways rather than crush the columns together. Rows are drawn only where the
  reader is looking, a window either side of the visible ones standing in as plain height, which is
  what answers a sort over a thousand entries inside a frame.

  Pulling the panel's own left edge widens it, between the width its headings need and the width that
  would crowd out the view beside it.

  It scrolls in its own region, a row selects, and the selected row follows
  a selection made elsewhere. A search field sits above it, and a search replaces what the table
  lists with what sdd matched, in the order sdd ranked it. A turning ring beside the field says a
  search is running, and a control on the field returns the table to every entry. A search outlives a
  switch between the views. A control at the right of that row filters by topic: choosing several
narrows to the entries carrying every one of them, and the areas sort ahead of the facets.
The search field, the topic filter and the done control sit above both views and narrow what each
draws, so the graph shows the entries a search and its filters left standing rather than always the
whole graph. The graph takes the whole area under that row, to the window's own edges.

- **Graph** draws what is listed, through `3d-force-graph` over three.js. An entry is one shape,
  coloured by type and greyed where the graph has retired it, sized by how many links reach it, and
  showing its identifier and title where the pointer rests on it. A
  link is one line, coloured by its kind and pointing at its target. Repulsion reaches a bounded
  distance, so an entry group linking to nothing else stays beside the body it belongs to rather than
  drifting past everything.
- **The selection is drawn as a halo**: red on the selected entry, purple and smaller on what links
  to it, purple and smaller again on what links to those. The camera flies to the selection, which a
  view of the whole graph draws too small to find by eye. With nothing selected, the settled layout
  is framed whole.
- **The panel** carries the entry's type, layer, kind, intent and derived status, its title, its
  summary, its topics with the area marked, its participants and confidence, the links it declares
  and the links that reach it, the identifiers it names that the graph does not hold, and its body.
  Every link is a control selecting the entry at its other end. It names that entry by its title,
  above the description the link carries.

**The title names an entry.** sdd writes a short title into each entry from graph schema 2 on. The
table, the graph's hover text, the panel's heading and every link in the panel use it. A graph
written before schema 2 carries none, and the summary stands in. The table shows the summary when the
pointer rests on a title.

`Entries()` reports one record per entry, newest first, since an entry's identifier opens with the
moment it was written. Each record carries the frontmatter fields, the links the file declares, and
the derived status, with the live head named where an entry is closed or superseded. The title is one
of those fields. Status comes
from one `View` call over `source(graph):as-list`, matched to entries by identifier, and the line it
is read from is the one sdd's own unit tests hold.

The front end builds the drawn network from those records. A link naming an entry the graph does not
hold, and a link an entry declares to itself, are drawn nowhere and reported in the panel instead.

An entry body is Markdown, and the panel renders it: headings, lists, emphasis, code spans, fenced
blocks, quotations, tables and links. The renderer builds elements rather than an HTML string, so
markup written inside an entry is text a reader sees and never markup the page runs, and no sanitiser
stands between the two. A link opens away from the viewer.

Opening a repository mounts sdd over the store. Its session and staged-blob stores sit in scratch the
viewer owns, because sdd requires both and a read writes to neither.

## Documentation sets

No implementation set exists yet.
