import os
import subprocess
import sys
import tempfile
import unittest

SCRIPT = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'lint_draft.py')


def lint(*arguments, path=None, config=None):
    environment = dict(os.environ)
    if path is not None:
        environment['PATH'] = path
    if config is not None:
        environment['VALE_CONFIG_PATH'] = config

    return subprocess.run([sys.executable, SCRIPT, *arguments], capture_output=True, text=True,
                          env=environment)


class LintDraftTest(unittest.TestCase):
    def setUp(self):
        scratch = tempfile.TemporaryDirectory()
        self.addCleanup(scratch.cleanup)
        self.directory = scratch.name

    def body(self, text):
        path = os.path.join(self.directory, '.draft.body.md')
        with open(path, 'w', encoding='utf-8') as handle:
            handle.write(text)

        return path

    def test_clean_body_passes_silently(self):
        run = lint(self.body('The viewer reads the graph.\n'))

        self.assertEqual(0, run.returncode, run.stderr)
        self.assertEqual('', run.stdout, 'a clean body should print nothing')

    def test_suggestion_alone_refuses_the_body(self):
        body = self.body('The viewer reads the graph.\n\nEach node opens a panel.\n')

        run = lint(body)

        self.assertEqual(1, run.returncode, 'a suggestion should refuse the body')
        self.assertIn(f'{body}:3:6  suggestion  Taxonomy.Node  ', run.stdout,
                      'the alert should name the body, its position, severity and check')

    def test_missing_vale_is_never_read_as_clean(self):
        run = lint(self.body('The viewer reads the graph.\n'), path=self.directory)

        self.assertEqual(2, run.returncode, 'a missing linter should fail the lint')
        self.assertIn('went unread', run.stderr)

    def test_refused_run_is_never_read_as_clean(self):
        config = os.path.join(self.directory, 'broken.ini')
        with open(config, 'w', encoding='utf-8') as handle:
            handle.write(f"StylesPath = {os.path.join(self.directory, 'absent')}\n")

        run = lint(self.body('The viewer reads the graph.\n'), config=config)

        self.assertEqual(2, run.returncode, 'a refused run should fail the lint')
        self.assertIn('went unread', run.stderr)

    def test_wrong_argument_count_prints_usage(self):
        run = lint()

        self.assertEqual(2, run.returncode)
        self.assertIn('usage', run.stderr)


if __name__ == '__main__':
    unittest.main()
