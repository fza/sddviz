#!/usr/bin/env bash
#
# Everything a machine needs before the toolchain builds.
#
# Safe to rerun: it checks each tool before installing anything, so a developer
# whose machine drifted runs what a newcomer runs.
#
# `--check` names what is absent and installs nothing, which is what shell
# entry runs. Each list below is the only place its tools are named.
#
# Homebrew carries what the Devbox environment cannot: the shell hook that
# loads Devbox in the first place, and a decision-graph tool this repository
# drives by bare name.

set -euo pipefail

cd "$(dirname "$0")/.."

# direnv loads the Devbox environment on entering the directory, so it cannot
# come from that environment. Homebrew's Git build is here because Xcode's lags,
# and Go builds the two programs below before Devbox is in play.
FORMULAE=(direnv git golang)

# The decision-graph tool. sddviz embeds sdd as a library and needs no binary at
# runtime; this repository's own graph work drives the command line, and
# `.mcp.json` launches its server.
CASKS=(networkteam/tap/sdd)

# The Wails command line, reached by bare name, so `$(go env GOPATH)/bin`
# belongs on PATH. `devbox run tools` installs the same program inside the
# environment.
GO_PROGRAMS=(github.com/wailsapp/wails/v3/cmd/wails3@latest)

# The prose linter is a fork carrying rules this repository depends on, chief
# among them reading a source file as code so only its comments are prose.
# Homebrew's own `vale` is the upstream build and lints those files as English,
# which turns a clean tree into thousands of findings. Its module path still
# names upstream, so `go install` cannot reach the fork and a clone is what
# builds it.
VALE_FORK=git@github.com:fza/vale.git
VALE_TAG=v3.19.0-fdbox.1

have() { command -v "$1" >/dev/null 2>&1; }

# The command a tool installs under, derived from what installs it. Homebrew's
# Go formula is the one name that differs from its command.
command_of() {
    name=${1##*/}
    name=${name%@*}

    case "$name" in
        golang) echo go ;;
        *) echo "$name" ;;
    esac
}

absent() {
    for source in "${FORMULAE[@]}" "${CASKS[@]}" "${GO_PROGRAMS[@]}"; do
        have "$(command_of "$source")" || printf '%s ' "$(command_of "$source")"
    done

    vale --version 2>/dev/null | grep -qF "$VALE_TAG" || printf 'vale '
}

# The fork builds from a clone this script keeps at one tag, so every machine
# carries the same linter and moving to newer fork work is an edit here.
install_vale() {
    clone=${XDG_CACHE_HOME:-$HOME/.cache}/sddviz/vale

    if [ -d "$clone/.git" ]; then
        git -C "$clone" fetch --depth 1 origin "refs/tags/$VALE_TAG:refs/tags/$VALE_TAG"
    else
        mkdir -p "$(dirname "$clone")"
        git clone --branch "$VALE_TAG" --depth 1 "$VALE_FORK" "$clone"
    fi

    git -C "$clone" checkout --quiet "$VALE_TAG"
    (cd "$clone" && go build -ldflags "-s -w -X main.version=$VALE_TAG" -o bin/vale ./cmd/vale)

    # Asking the toolchain rather than assuming ~/go, since a developer may have
    # moved GOPATH and this is the same directory every other Go program lands in.
    bin=$(go env GOPATH)/bin
    mkdir -p "$bin"
    install "$clone/bin/vale" "$bin/vale"
}

if [ "${1:-}" = "--check" ]; then
    gone=$(absent)
    [ -z "$gone" ] || printf 'missing: %s\n  ./scripts/dev-setup.sh installs them\n' "${gone% }"
    exit 0
fi

if ! have brew; then
    echo "Homebrew is missing: https://brew.sh" >&2
    exit 1
fi

printf '\n== tools\n'
for source in "${FORMULAE[@]}"; do
    have "$(command_of "$source")" && echo "have $(command_of "$source")" || brew install "$source"
done

for source in "${CASKS[@]}"; do
    have "$(command_of "$source")" && echo "have $(command_of "$source")" || brew install --cask "$source"
done

for source in "${GO_PROGRAMS[@]}"; do
    have "$(command_of "$source")" && echo "have $(command_of "$source")" || go install "$source"
done

printf '\n== prose linter\n'
if brew list --versions vale >/dev/null 2>&1; then
    echo "Homebrew carries the upstream vale, which shadows the fork." >&2
    echo "Remove it first: brew uninstall vale" >&2
    exit 1
fi

if vale --version 2>/dev/null | grep -qF "$VALE_TAG"; then
    echo "have $(vale --version)"
else
    install_vale
fi

printf '\n== next\n'
cat <<'MESSAGE'
direnv allow    # builds the toolchain and loads it on every cd here

direnv allow needs direnv's own shell hook in ~/.zshrc first:
    eval "$(direnv hook zsh)"
MESSAGE
