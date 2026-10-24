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

import glob
import logging
import os
import re
import subprocess
import sys

CORPUS = ('*.md', '*.go', '*.ts', '*.tsx', '*.js', '*.css', '*.html', '*.py',
          '*.sh', '*.yml', '*.yaml', '*.json')

LINK_DEST = re.compile(r'\[[^\]]*\]\(([^)\s]+)')
TREE_PATH = re.compile(r'`?((?:docs|source|scripts)/[A-Za-z0-9_./@-]+)`?')
CODE_SPAN = re.compile(r'`([^`\n]+)`')
# A bare word in backticks is a directory, a command or a field as often as it
# is a symbol. Requiring an exported Go name bounds the check to what the
# service surface actually carries, and keeps a word like `docs` out of it. The
# lowercase letter is what separates a Go name from an environment variable such
# as PATH.
IDENTIFIER = re.compile(r'^[A-Z][A-Za-z0-9]*[a-z][A-Za-z0-9]*$')
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
        if not path.endswith('.md') or path.startswith(('.sdd/', '.claude/', 'source/build/')):
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
        if path.startswith(('.sdd/', '.claude/', 'scripts/lint/', 'source/build/')):
            continue

        # A graph entry is immutable and names paths as they stood at capture,
        # so a fixture graph copied from one is exempt wherever it sits.
        if '/.sdd/' in path:
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


# A line about the embedded library names that library's symbols, which this
# tree is not expected to carry. The check asserts our own identifiers, so such
# a line is outside it.
UPSTREAM = re.compile(r'\bsdd\b|\bapplication\b|\bsddtest\b|\bGraphStore\b')


def check_identifiers(corpus):
    """Every code identifier a document claims this tree carries."""
    report = Report('every identifier the documentation names exists under source/',
                    'Build the thing, drop the claim, or name it as absent.')

    bodies = [text for path, text in corpus.items() if path.startswith('source/')]

    for path, text in sorted(corpus.items()):
        if not is_documentation(path):
            continue

        seen = set()
        for number, line in enumerate(text.split('\n'), 1):
            if UPSTREAM.search(line):
                continue
            for span in CODE_SPAN.findall(line):
                if not IDENTIFIER.match(span) or span in seen:
                    continue
                if any(span in body for body in bodies):
                    continue
                seen.add(span)
                report.note(f'{path}:{number}',
                            f'names `{span}`, which nothing under source/ carries')

    return report


# --- vocabulary --------------------------------------------------------------

RULES = 'docs/rules.md'
TAXONOMY = '.vale/styles/Taxonomy'
ACCEPT = '.vale/styles/config/vocabularies/Project/accept.txt'
SCRATCH = '.tmp/vocabulary-headings.md'
QUOTED_FORM = re.compile(r"'([^']+)'")


def check_vocabulary(corpus):
    """The checks and the rules document name the same forms, both directions."""
    report = Report(f'the vocabulary checks and {RULES} name the same forms',
                    f'Teach {RULES} the form, or retire it from the check.')

    rules = corpus.get(RULES) or read(RULES)
    folded = rules.lower()

    # Direction one: a check tells a writer to use a form the rules document
    # never explains.
    for style in sorted(glob.glob(f'{TAXONOMY}/*.yml')):
        for line in read(style).split('\n'):
            if not line.startswith('message:'):
                continue
            for form in QUOTED_FORM.findall(line):
                # Vale substitutes the matched token for %s, so it names no form.
                if form == '%s' or form.lower() in folded:
                    continue
                report.note(style, f'tells a writer to use "{form}", which '
                                   f'{RULES} never explains')
            break

    # Direction two: the rules document defines a form its own checks reject.
    # Vale needs a real file inside the repository, or no style applies to it.
    headings = [re.sub(r'^###[ \t]*', '', line)
                for line in rules.split('\n') if line.startswith('###')]
    if not headings:
        return report

    os.makedirs('.tmp', exist_ok=True)
    with open(SCRATCH, 'w', encoding='utf-8') as handle:
        handle.write('\n'.join(headings) + '\n')
    try:
        alerts = subprocess.run(['vale', '--output=line', '--minAlertLevel=error', SCRATCH],
                                capture_output=True, text=True, check=False).stdout
    except OSError:
        alerts = ''
    finally:
        os.remove(SCRATCH)

    for alert in alerts.split('\n'):
        if 'Taxonomy' not in alert:
            continue
        term = headings[int(alert.split(':')[1]) - 1]
        report.note(RULES, f'defines "{term}", which its own checks report as a defect')

    return report


# --- taxonomy probes ---------------------------------------------------------

PROBE = '.tmp/taxonomy-probe.md'

# One banned word, the lines it must report and the lines it must leave alone.
# A token that stops matching passes every document silently, so the rule needs
# a case that fails on purpose to prove it fires.
PROBES = {
    'Taxonomy.Entry': {
        'fires': [
            'The card is superseded.',
            'Two cards are superseded.',
            'Vertices carry no status.',
        ],
        'passes': [
            'The entry is superseded.',
            'The signal is closed.',
            'A discard pile is somebody else\'s word.',
        ],
    },
}


def check_taxonomy_probes(_corpus):
    """Every banned-word check reports the word it bans."""
    report = Report('every banned-word check fires on the word it bans',
                    'Repair the token in `.vale/styles/Taxonomy/`, never the probe.')

    for rule, cases in sorted(PROBES.items()):
        lines = cases['fires'] + cases['passes']
        os.makedirs('.tmp', exist_ok=True)
        with open(PROBE, 'w', encoding='utf-8') as handle:
            handle.write('\n'.join(lines) + '\n')

        try:
            alerts = subprocess.run(['vale', '--output=line', '--minAlertLevel=error', PROBE],
                                    capture_output=True, text=True, check=False).stdout
        except OSError:
            report.note(rule, 'vale is absent, so nothing proved this rule fires')
            continue
        finally:
            os.remove(PROBE)

        fired = {int(alert.split(':')[1]) for alert in alerts.split('\n') if rule in alert}

        for at, line in enumerate(cases['fires'], start=1):
            if at not in fired:
                report.note(rule, f'stays silent on "{line}", which it bans')

        for at, line in enumerate(cases['passes'], start=len(cases['fires']) + 1):
            if at in fired:
                report.note(rule, f'reports "{line}", which is correct')

    return report


# --- vocabulary shape --------------------------------------------------------

# A term the vocabulary accepts is a word a reader writes. A name the code binds
# is not: it reaches prose in backticks, which Vale reads as code and never
# reports. Three shapes tell them apart, and each of them names code.
CODE_SHAPED = re.compile(r'_|[0-9]|[a-z][A-Z]')

# Real terms this project writes that carry an inner capital anyway.
INNER_CAPITAL = frozenset({
    'macOS', 'GitLab', 'JavaScript', 'TypeScript', 'WebView', 'snake_case', 'camelCase',
})


def check_vocabulary_shape(corpus):
    """No accepted term names code instead of a word."""
    report = Report('no accepted term names code instead of a word',
                    'Wrap the name in backticks where it is written, and drop the entry.')

    text = corpus.get(ACCEPT) or read(ACCEPT)
    for number, line in enumerate(text.split('\n'), start=1):
        term = line.strip()
        if not term:
            continue
        bare = re.sub(r"\('s\)\?|\(s\)\?|\?$", '', term).strip()
        if bare in INNER_CAPITAL or not CODE_SHAPED.search(bare):
            continue
        report.note(f'{ACCEPT}:{number}', f'names code rather than a word: {bare}')

    return report


CHECKS = (check_paths, check_links, check_identifiers, check_vocabulary,
          check_taxonomy_probes, check_vocabulary_shape)


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
