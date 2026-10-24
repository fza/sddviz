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
# loads Devbox in the first place, and the two forks this repository drives by
# bare name.

set -euo pipefail

cd "$(dirname "$0")/.."

# direnv loads the Devbox environment on entering the directory, so it cannot
# come from that environment. Homebrew's Git build is here because Xcode's lags,
# and Go builds the program below before Devbox is in play.
FORMULAE=(direnv git golang)

# The decision-graph tool and the prose linter, both forks. sddviz embeds the sdd
# fork as a library and needs no binary at runtime; this repository's own graph
# work drives the command line, `agentic-capture` refuses any other build, and
# `.mcp.json` launches its server. The Vale fork carries rules this repository
# depends on, chief among them reading a source file as code so only its
# comments are prose. Homebrew's own `vale` formula is the upstream build and
# lints those files as English, which turns a clean tree into thousands of
# findings.
CASKS=(fza/tap/sdd fza/tap/vale)

# The mark every fork release carries in its version.
FORK_MARK=fza

# The Wails command line, reached by bare name, so `$(go env GOPATH)/bin`
# belongs on PATH. `devbox run tools` installs the same program inside the
# environment.
GO_PROGRAMS=(github.com/wailsapp/wails/v3/cmd/wails3@latest)

have() { command -v "$1" >/dev/null 2>&1; }

# A command on PATH that is not the fork counts as absent, since an upstream
# build answers to the same name.
have_fork() { "$1" --version 2>/dev/null | grep -qF -- "$FORK_MARK"; }

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
    for source in "${FORMULAE[@]}" "${GO_PROGRAMS[@]}"; do
        have "$(command_of "$source")" || printf '%s ' "$(command_of "$source")"
    done

    for source in "${CASKS[@]}"; do
        have_fork "$(command_of "$source")" || printf '%s ' "$(command_of "$source")"
    done
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

for source in "${GO_PROGRAMS[@]}"; do
    have "$(command_of "$source")" && echo "have $(command_of "$source")" || go install "$source"
done

printf '\n== forks\n'
for source in "${CASKS[@]}"; do
    name=$(command_of "$source")

    # An upstream formula links the same command into Homebrew's bin, so the
    # fork's cask cannot install beside it.
    if brew list --formula --versions "$name" >/dev/null 2>&1; then
        echo "Homebrew carries the upstream $name, which collides with the fork." >&2
        echo "Remove it first: brew uninstall --formula $name" >&2
        exit 1
    fi

    if have_fork "$name"; then
        echo "have $("$name" --version)"
    else
        brew install --cask "$source"
    fi

    if ! have_fork "$name"; then
        echo "$(command -v "$name") is not the fork and comes first on PATH." >&2
        exit 1
    fi
done

printf '\n== next\n'
cat <<'MESSAGE'
direnv allow    # builds the toolchain and loads it on every cd here

direnv allow needs direnv's own shell hook in ~/.zshrc first:
    eval "$(direnv hook zsh)"
MESSAGE
