# sddviz

A desktop viewer for [SDD](https://github.com/networkteam/sdd) decision graphs. It opens a
repository that carries an `.sdd/` graph and shows what the project decided and why, through a
filterable entry browser with search and a 3D view of the graph itself.

sddviz is **read-only**. It never captures, never decides, and never writes to the repository it
reads. That is a property of the code rather than a promise: the graph reaches sdd through a store
whose write methods refuse.

sdd runs inside sddviz as a library, not as a tool it shells out to, so nothing has to be installed
alongside it and the version it reads is pinned in `go.mod`.

## Status

The repository holds the project scaffold: a Wails v3 application with a React and TypeScript front
end talking to a Go back end. Neither the entry browser, the search, nor the 3D view exists yet.
`docs/index.md` indexes what is documented; anything unbuilt lives in the decision graph under
`.sdd/`.

## Why a viewer

SDD records a project's reasoning as an append-only graph of signals (what was noticed) and decisions
(what was committed to). The graph is read through an agent today, which serves a developer at a
keyboard well and serves everyone else poorly. A viewer puts the same reasoning in front of anyone
who needs it, without a terminal and without an agent session.

## Requirements

macOS, for now. Wails builds Linux and Windows from the same source, and neither is exercised.

The toolchain is pinned in `devbox.json` ([devbox](https://www.jetify.com/devbox) plus
[direnv](https://direnv.net)). Entering the directory loads Go 1.26, Node.js 24 and the linters.
`devbox run tools` installs the Wails CLI into the Go bin directory. Building the desktop
application also needs the Xcode command line tools, since it links the system WebView through cgo.

## Running

```bash
devbox run dev            # the application, front end in watch mode
devbox run build          # a signed `.app` bundle
devbox run build:server   # the same application as an HTTP server
devbox run sign           # sign the bundle for distribution
devbox run test           # both halves
devbox run lint           # both halves, prose, then the repository checks
```

`devbox run bindings` regenerates the typed front end bindings after a service method changes.

## Signing

The bundle identifier is `com.fza.sddviz`, and every build carries the hardened runtime. The signing
identity comes from `SDDVIZ_SIGN_IDENTITY`, which `devbox.json` sets to the development certificate
on this machine.

A development certificate produces a bundle that runs where it was built. Gatekeeper accepts a
bundle elsewhere only under a `Developer ID Application` certificate that has also been notarized,
which needs two things this repository cannot create: that certificate in the keychain, and
notarization credentials stored with
`xcrun notarytool store-credentials sddviz-notary`. With both in place,
`devbox run sign:notarize` signs and submits.

## The web build

The same code serves a browser. `go build -tags server` produces an HTTP server with no native
window, keeping every service method and event, and dropping the system tray, the application menus
and the native dialogs.

What that build does not solve, and what a real deployment would decide: the graph then lives on the
server rather than on the reader's machine, authentication stops being optional, and one reader per
process becomes many.

## Structure

```
source              The Go module: the application, and everything that reads a graph
source/frontend     React and TypeScript, the web application the module embeds
docs                What the implementation is, as it stands
.sdd                The decision graph: why the implementation is the way it is
```

The front end reaches the back end only through generated bindings over Wails services, and touches
no filesystem path of its own.

## License

MIT. See [`LICENSE`](LICENSE).

## Limits

sddviz knows the entry format it reads, currently `graph_schema_version: 1`. A graph declaring a
version it does not know is reported to the reader rather than guessed at.

Reads follow the repository's working tree. Selecting an arbitrary branch or an older revision is not
available: sdd takes a branch only as a session binding validated against a live checkout, and sddviz
checks nothing out.
