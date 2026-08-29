#!/usr/bin/env python3
"""One draft file to a captured graph entry: lint, pre-flight, capture.

A draft carries its own metadata in a leading YAML block and its body below, so the
fields that decide an entry travel with the prose that fills it and a capture
repeats no arguments from memory.

The body reaches `sdd new` as one argument of an argument vector, never through
a shell word. `sdd new "$(cat body.md)"` re-evaluates that string in a second
shell, so backticks in a body run as command substitution and every code
identifier vanishes from the entry.

Vale reads a real Markdown file and reports nothing for a body piped to it, so
the body is written out under `.tmp/` for the pass and removed again whatever
the outcome.

Three gates run in order and each stops the run: the leading block answers what an
entry of that kind needs, Vale reports nothing on the body, and the tool's
pre-flight reports nothing above low severity.
"""

import argparse
import json
import logging
import os
import re
import subprocess
import sys
import tempfile

import yaml

REPO_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SCRATCH_DIR = os.path.join(REPO_ROOT, ".tmp")

# Draft keys carrying a comma-joined list on the command line.
LIST_FLAGS = {"topics": "--topics", "closes": "--closes", "supersedes": "--supersedes",
              "participants": "--participants", "aliases": "--aliases", "actors": "--actors"}

# Draft keys carrying a single value on the command line.
VALUE_FLAGS = {"kind": "--kind", "confidence": "--confidence", "intent": "--intent",
               "canonical": "--canonical", "actor": "--actor", "class": "--class",
               "summary": "--summary"}

# Draft keys carrying one JSON object per list item, the flag repeated.
JSON_FLAGS = {"refs": "--refs", "involvement": "--involvement", "topic": "--topic"}

SEVERITY = re.compile(r"^\s*\[(low|medium|high)\]")

log = logging.getLogger("graph-capture")


class DraftError(Exception):
    """A draft the tool refuses before any pass runs."""


def parse_draft(path):
    """Split a draft into its leading mapping, its body, and the line the body opens on."""
    with open(path, encoding="utf-8") as handle:
        text = handle.read()

    if not text.startswith("---\n"):
        raise DraftError(f"{path} opens with no --- line, so it carries no leading block")

    end = text.find("\n---\n", 3)
    if end < 0:
        raise DraftError(f"{path} carries no closing --- line for its leading block")

    front = yaml.safe_load(text[4:end]) or {}
    raw_body = text[end + 5:]
    body = raw_body.strip()

    # A Vale finding names a line of the body, and a reader looks for it in the draft.
    offset = text[:end + 5].count("\n") + raw_body[:len(raw_body) - len(raw_body.lstrip("\n"))].count("\n")

    if not isinstance(front, dict):
        raise DraftError(f"{path} opens with something other than a mapping of keys to values")
    if not body:
        raise DraftError(f"{path} carries no body below its leading block")

    return front, body, offset


def validate(front):
    """Refuse what `sdd` cannot judge: the shape this call needs, and the area rule.

    Every value check - type, layer, kind, confidence, ref kinds, the fields a
    kind requires - belongs to `sdd new`, which refuses them itself.
    """
    for key in ("type", "layer"):
        if key not in front:
            raise DraftError(f"the leading block needs a {key}")

    topics = front.get("topics", [])
    if isinstance(topics, str):
        topics = [part.strip() for part in topics.split(",")]
    areas = [topic for topic in topics if topic.startswith("area-")]
    if len(areas) != 1:
        raise DraftError("topics carry exactly one label opening with area-, "
                         f"and these carry {len(areas)}")


def command(front, body, dry_run):
    """Build the argument vector `sdd new` runs, body included as one argument."""
    args = ["sdd", "new", front["type"], front["layer"], body]

    for key, flag in VALUE_FLAGS.items():
        if key in front:
            args += [flag, str(front[key])]

    for key, flag in LIST_FLAGS.items():
        if key not in front:
            continue
        value = front[key]
        joined = value if isinstance(value, str) else ",".join(str(item) for item in value)
        args += [flag, joined]

    for key, flag in JSON_FLAGS.items():
        for item in front.get(key) or []:
            args += [flag, json.dumps(item, ensure_ascii=False)]

    for item in front.get("attach") or []:
        args += ["--attach", str(item)]

    if dry_run:
        args.append("--dry-run")

    return args


def lint_body(body, offset):
    """Report every Vale finding on the body, at the line the draft carries it on."""
    os.makedirs(SCRATCH_DIR, exist_ok=True)
    handle, path = tempfile.mkstemp(prefix="graph-capture-", suffix=".md", dir=SCRATCH_DIR)

    try:
        with os.fdopen(handle, "w", encoding="utf-8") as draft:
            draft.write(body + "\n")

        result = subprocess.run(["vale", "--output=line", path], cwd=REPO_ROOT,
                                capture_output=True, text=True, check=False)
        return [shift(line, path, offset) for line in result.stdout.splitlines() if line.strip()]
    finally:
        os.unlink(path)


def shift(alert, path, offset):
    """One Vale line, its body line number moved to the line the draft carries."""
    parts = alert.replace(path, "draft", 1).split(":", 3)
    if len(parts) < 4 or not parts[1].isdigit():
        return alert

    parts[1] = str(int(parts[1]) + offset)

    return ":".join(parts)


def preflight_findings(output):
    """Every severity the pre-flight reported, in the order it reported them."""
    return [match.group(1) for match in
            (SEVERITY.match(line) for line in output.splitlines()) if match]


def run(args):
    front, body, offset = parse_draft(args.draft)
    validate(front)

    alerts = lint_body(body, offset)
    if alerts:
        log.error("Vale reports %s on the body:",
                  "one finding" if len(alerts) == 1 else f"{len(alerts)} findings")
        for alert in alerts:
            log.error("  %s", alert)
        return 1

    log.info("body clears Vale")

    result = subprocess.run(command(front, body, dry_run=True), cwd=REPO_ROOT,
                            capture_output=True, text=True, check=False)
    output = result.stdout + result.stderr
    if result.returncode != 0:
        log.error("pre-flight refused the draft:\n%s", output.strip())
        return 1

    findings = preflight_findings(output)
    blocking = [severity for severity in findings if severity != "low"]
    if output.strip():
        log.info("pre-flight:\n%s", output.strip())

    if blocking and not args.force:
        log.error("pre-flight reports %s above low severity - fix the draft, or pass --force "
                  "to capture anyway",
                  "one finding" if len(blocking) == 1 else f"{len(blocking)} findings")
        return 2

    if args.dry_run:
        log.info("draft passes; nothing captured")
        return 0

    result = subprocess.run(command(front, body, dry_run=False), cwd=REPO_ROOT, check=False)
    return result.returncode


def main():
    parser = argparse.ArgumentParser(
        description="Lint, pre-flight and capture one graph entry from a draft file.",
        epilog="The draft opens with a YAML block (type, layer, kind, confidence, topics, and "
               "whatever else the kind needs: intent, refs, closes, supersedes, participants, "
               "summary) above the entry body.")
    parser.add_argument("draft", help="path to the draft Markdown file")
    parser.add_argument("--dry-run", action="store_true",
                        help="stop after the pre-flight, capturing nothing")
    parser.add_argument("--force", action="store_true",
                        help="capture despite pre-flight findings above low severity")
    parsed = parser.parse_args()

    logging.basicConfig(level=logging.INFO, format="%(message)s")

    try:
        return run(parsed)
    except DraftError as error:
        log.error("%s", error)
        return 1


if __name__ == "__main__":
    sys.exit(main())
