#!/usr/bin/env python3
"""Block a write whose newly added lines carry Vale alerts.

Alerts on lines the write did not touch are ignored, so a document that
already carries alerts stays writable while new prose is held to zero.
"""

import json
import os
import re
import shutil
import subprocess
import sys

HUNK = re.compile(r'^@@ -\d+(?:,\d+)? \+(\d+)(?:,(\d+))? @@')


def ignored(path):
    """True when git ignores the path.

    An ignored file never reaches the repository, so the prose bar that
    governs what lands there does not govern it. Graph entry drafts under
    `.tmp/` are the case this exempts: `agentic-capture` lints their bodies
    itself, because a captured body cannot be corrected afterwards.
    """
    run = subprocess.run(
        ['git', 'check-ignore', '-q', '--', path],
        capture_output=True,
    )

    return run.returncode == 0


def added_lines(path):
    """Line numbers this working tree adds on top of HEAD.

    An untracked file counts whole: every line of it is new.
    """
    tracked = subprocess.run(
        ['git', 'ls-files', '--error-unmatch', '--', path],
        capture_output=True,
    )

    if tracked.returncode != 0:
        with open(path, encoding='utf-8', errors='replace') as handle:
            return set(range(1, sum(1 for _ in handle) + 1))

    diff = subprocess.run(
        ['git', 'diff', '-U0', 'HEAD', '--', path],
        capture_output=True,
        text=True,
    )

    lines = set()

    for line in diff.stdout.splitlines():
        match = HUNK.match(line)

        if match:
            start = int(match.group(1))
            count = int(match.group(2) or 1)
            lines.update(range(start, start + count))

    return lines


def alerts(path):
    """Vale's findings on the path, read from the repository root.

    `.vale.ini` matches its sections against paths relative to the root, and
    an absolute path falls outside every one of them, carve-outs included.
    """
    root = os.environ.get('CLAUDE_PROJECT_DIR') or os.getcwd()

    run = subprocess.run(
        ['vale', '--output=JSON', os.path.relpath(os.path.abspath(path), root)],
        capture_output=True,
        text=True,
        cwd=root,
    )

    try:
        report = json.loads(run.stdout or '{}')
    except json.JSONDecodeError:
        return []

    found = []

    for file_alerts in report.values():
        found.extend(file_alerts)

    return found


def main():
    try:
        event = json.load(sys.stdin)
    except json.JSONDecodeError:
        return 0

    path = (event.get('tool_input') or {}).get('file_path')

    if not path or not os.path.isfile(path):
        return 0

    if not shutil.which('vale'):
        return 0

    if ignored(path):
        return 0

    touched = added_lines(path)

    if not touched:
        return 0

    hits = [a for a in alerts(path) if a.get('Line') in touched]

    if not hits:
        return 0

    hits.sort(key=lambda a: (a.get('Line', 0), a.get('Span', [0])[0]))

    report = [
        'Vale alerts on lines this write added. Fix every one of them, or,',
        'where a fix would change what the sentence says, keep the wording',
        'and say so in your reply. Never silence a rule to clear this.',
        '',
        path,
    ]

    for alert in hits:
        report.append(
            '  {}:{}  {}  {}  {}'.format(
                alert.get('Line'),
                alert.get('Span', [0])[0],
                alert.get('Severity'),
                alert.get('Check'),
                alert.get('Message'),
            )
        )

    print('\n'.join(report), file=sys.stderr)

    return 2


if __name__ == '__main__':
    sys.exit(main())
