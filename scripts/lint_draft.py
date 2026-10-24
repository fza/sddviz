#!/usr/bin/env python3
"""Lint a graph entry draft's body with Vale, holding it to zero alerts.

`agentic-capture` runs this as the `draft_lint` of `.sdd/grounding.yaml`,
passing the body it cut out of a draft into a Markdown file of its own. A
captured body can never be corrected, so every alert counts, suggestions
included, where Vale's own exit code counts errors alone.

Each alert prints as `<body>:<line>:<column>  <severity>  <check>  <message>`,
the shape `agentic-capture` moves onto the draft's own lines.
"""

import json
import os
import subprocess
import sys

REPO_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))


def alerts(path):
    """Every finding Vale raises on the file, ordered by position.

    The path is relative to the repository root, because `.vale.ini` matches its
    sections against relative paths and an absolute one falls outside them.
    """
    try:
        run = subprocess.run(['vale', '--output=JSON', path], capture_output=True, text=True,
                             cwd=REPO_ROOT)
    except OSError as missing:
        raise RuntimeError(f'vale did not run: {missing}')

    # Vale answers findings with 1 and a refusal with anything above it. A
    # refusal prints its error as JSON in place of a report.
    if run.returncode > 1:
        reason = (run.stdout.strip() or run.stderr.strip())[:200]
        raise RuntimeError(f'vale exited {run.returncode}: {reason}')

    try:
        report = json.loads(run.stdout or '{}')
    except json.JSONDecodeError:
        raise RuntimeError(f'vale answered nothing readable: {run.stderr.strip()[:200]}')

    if not isinstance(report, dict):
        raise RuntimeError(f'vale answered a {type(report).__name__} rather than a report')

    found = [alert for file_alerts in report.values() for alert in file_alerts]

    return sorted(found, key=lambda alert: (alert.get('Line', 0), alert.get('Span', [0])[0]))


def main(argv):
    if len(argv) != 2:
        print('usage: lint_draft.py <body.md>', file=sys.stderr)

        return 2

    body = argv[1]
    try:
        found = alerts(os.path.relpath(os.path.abspath(body), REPO_ROOT))
    except RuntimeError as silent:
        print(f'{body} went unread: {silent}', file=sys.stderr)

        return 2

    for alert in found:
        print(f"{body}:{alert.get('Line')}:{alert.get('Span', [0])[0]}  {alert.get('Severity')}  "
              f"{alert.get('Check')}  {alert.get('Message')}")

    return 1 if found else 0


if __name__ == '__main__':
    sys.exit(main(sys.argv))
