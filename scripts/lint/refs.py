#!/usr/bin/env python3
"""Every repository reference check, over one read of the tree.

The checks share a corpus and differ only in what they look for, so they run in
one process against one set of file contents held in memory.

Each check exists because no other tool sees the class it covers. A prose linter
reads style and never follows a pointer. A compiler reads code and never reads a
document. So a reference rots for as long as nobody happens to try it.

Each check collects its findings and returns them, and every check runs whatever
the ones before it found, because a reader fixing one wants to see the rest in
the same pass.
"""

import logging
import os
import re
import subprocess
import sys

CORPUS = ('*.md', '*.rs', '*.ts', '*.tsx', '*.js', '*.css', '*.html', '*.py',
          '*.sh', '*.yml', '*.yaml', '*.json')

LINK_DEST = re.compile(r'\[[^\]]*\]\(([^)\s]+)')
TREE_PATH = re.compile(r'`?((?:docs|source|scripts)/[A-Za-z0-9_./@-]+)`?')
CODE_SPAN = re.compile(r'`([^`\n]+)`')
# A bare word in backticks is a directory, a command or a field as often as it
# is a symbol. Requiring snake_case bounds the check to what Rust and the Tauri
# command surface actually name, and keeps a word like `docs` out of it.
IDENTIFIER = re.compile(r'^[a-z][a-z0-9]*(?:_[a-z0-9]+)+$')
PLACEHOLDER = set('<>{}*$')

log = logging.getLogger('refs')


def tracked(*patterns):
    listing = subprocess.run(
        ['git', 'ls-files', '-z', '-co', '--exclude-standard', '--', *patterns],
        capture_output=True, text=True, check=True).stdout
    return [path for path in listing.split('\0') if path]


def read(path):
    try:
        with open(path, 'rb') as handle:
            return handle.read().decode('utf-8', 'replace')
    except OSError:
        return ''


def is_documentation(path):
    return path.startswith('docs/') and path.endswith('.md')


class Report:
    """What one check found, and what a reader does about it."""

    def __init__(self, asserts, repair):
        self.asserts = asserts
        self.repair = repair
        self.findings = []

    def note(self, site, problem):
        self.findings.append((site, problem))

    def emit(self):
        if not self.findings:
            log.info('%s', self.asserts)
            return True
        log.error('%s', self.asserts)
        for site, problem in self.findings:
            log.error('    %s %s', site, problem)
        log.error('    %s', self.repair)
        return False


def check_links(corpus):
    """Every relative link destination, in whichever tree it lives."""
    report = Report('every relative link this repository writes resolves',
                    'Repoint the link, or restore the target.')

    for path, text in sorted(corpus.items()):
        if not path.endswith('.md') or path.startswith(('.sdd/', '.claude/')):
            continue

        here = os.path.dirname(path)
        seen = set()
        for number, line in enumerate(text.split('\n'), 1):
            for dest in LINK_DEST.findall(line):
                dest = dest.split('#', 1)[0]
                if not dest or dest.startswith(('/', '#')):
                    continue
                if '://' in dest or dest.startswith('mailto:'):
                    continue
                if PLACEHOLDER & set(dest):
                    continue

                target = os.path.normpath(os.path.join(here, dest))
                if os.path.exists(target) or (number, dest) in seen:
                    continue
                seen.add((number, dest))
                report.note(f'{path}:{number}',
                            f'links a destination that does not exist: {dest}')

    return report


def check_paths(corpus):
    """Every bare tree path a tracked file names."""
    report = Report('every bare tree path this repository names exists',
                    'Remove the pointer, or restore the target.')

    for path, text in sorted(corpus.items()):
        if path.startswith(('.sdd/', '.claude/', 'scripts/lint/')):
            continue

        seen = set()
        for number, line in enumerate(text.split('\n'), 1):
            for candidate in TREE_PATH.findall(line):
                candidate = candidate.rstrip('.,;:)')
                if PLACEHOLDER & set(candidate):
                    continue
                if os.path.exists(candidate) or (number, candidate) in seen:
                    continue
                seen.add((number, candidate))
                report.note(f'{path}:{number}',
                            f'names a path that does not exist: {candidate}')

    return report


def check_identifiers(corpus):
    """Every code identifier a document claims the tree carries."""
    report = Report('every identifier the documentation names exists under source/',
                    'Build the thing, drop the claim, or name it as absent.')

    bodies = [text for path, text in corpus.items() if path.startswith('source/')]

    for path, text in sorted(corpus.items()):
        if not is_documentation(path):
            continue

        seen = set()
        for number, line in enumerate(text.split('\n'), 1):
            for span in CODE_SPAN.findall(line):
                if not IDENTIFIER.match(span) or span in seen:
                    continue
                if any(span in body for body in bodies):
                    continue
                seen.add(span)
                report.note(f'{path}:{number}',
                            f'names `{span}`, which nothing under source/ carries')

    return report


CHECKS = (check_paths, check_links, check_identifiers)


def main(argv):
    logging.basicConfig(format='%(message)s', level=logging.INFO)

    wanted = {name.replace('-', '_') for name in argv}
    selected = [check for check in CHECKS
                if not wanted or check.__name__.removeprefix('check_') in wanted]
    if not selected:
        log.error('no such check: %s', ', '.join(sorted(argv)))
        return 2

    corpus = {path: read(path) for path in tracked(*CORPUS)}
    return 0 if all([check(corpus).emit() for check in selected]) else 1


if __name__ == '__main__':
    sys.exit(main(sys.argv[1:]))
