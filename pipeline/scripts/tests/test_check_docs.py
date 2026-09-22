import importlib.util
import tempfile
import unittest
from pathlib import Path


spec = importlib.util.spec_from_file_location("check_docs", Path(__file__).parents[1] / "check_docs.py")
checker = importlib.util.module_from_spec(spec)
spec.loader.exec_module(checker)


class DocumentationChecks(unittest.TestCase):
    def test_paths_and_examples(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory).resolve()
            (root / "file with space.md").write_text("# File\n", encoding="utf-8")
            path = root / "README.md"
            path.write_text('# Title\n[real](file%20with%20space.md#section)\n'
                            '[angle](<file with space.md>)\n[remote](https://example.invalid/file)\n'
                            '[anchor](#here)\n`[example](missing.md)`\n'
                            '[real-ref]: file%20with%20space.md\n'
                            '[^footnote]: [remote source](https://example.invalid)\n'
                            '```md\n[example](missing.md)\n```\n', encoding="utf-8")
            self.assertEqual(checker.check_document(root, path), [])
            path.write_text('# Title\n[missing](absent.md)\n[outside](../file.md)\n', encoding="utf-8")
            self.assertEqual(len(checker.check_document(root, path)), 2)

    def test_missing_reference(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory).resolve()
            path = root / "README.md"
            path.write_text("# Title\n[reference]: absent.md\n", encoding="utf-8")
            self.assertEqual(len(checker.check_document(root, path)), 1)

    def test_missing_and_encoding(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory).resolve()
            path = root / "README.md"
            self.assertEqual(len(checker.check_document(root, path)), 1)
            path.write_bytes(b"# Title\n\xff")
            self.assertEqual(len(checker.check_document(root, path)), 1)
            path.write_text("No heading\n\ufffd", encoding="utf-8")
            self.assertEqual(len(checker.check_document(root, path)), 2)


if __name__ == "__main__":
    unittest.main()
