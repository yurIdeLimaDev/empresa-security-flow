from pathlib import Path
import sys
import unittest

SCRIPTS = Path(__file__).resolve().parents[1] / "scripts"
sys.path.insert(0, str(SCRIPTS))

from build_corpus import build_corpus, corpus_fingerprint, project_root


class BuildCorpusTests(unittest.TestCase):
    def test_current_project_has_a_stable_safe_corpus(self):
        documents = build_corpus(project_root())
        self.assertGreater(len(documents), 10)
        self.assertEqual(len(documents), len({document.source for document in documents}))
        self.assertTrue(all("correcao/casos/" not in document.source for document in documents))
        self.assertTrue(all("tool-artifacts/" not in document.source for document in documents))
        self.assertIn("AGENTS.md", {document.source for document in documents})
        self.assertIn("pipeline/knowledge/hipporag/README.md", {document.source for document in documents})
        self.assertEqual(len(corpus_fingerprint(documents)), 64)


if __name__ == "__main__":
    unittest.main()
