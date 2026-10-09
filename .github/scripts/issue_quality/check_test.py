import unittest
from pathlib import Path

from rules import check_issue

FIXTURES = Path(__file__).parent / "fixtures"


def _read(name: str) -> str:
    return (FIXTURES / name).read_text(encoding="utf-8")


class IssueQualityCheckTest(unittest.TestCase):
    def test_latest_version_fails(self):
        body = _read("bug_latest_version.md")
        result = check_issue(body, ["bug"])
        self.assertFalse(result.ok)
        self.assertFalse(result.skipped)
        ids = {d.id for d in result.deficiencies}
        self.assertIn("invalid_version", ids)

    def test_valid_bug_passes(self):
        body = _read("bug_valid.md")
        result = check_issue(body, ["bug"])
        self.assertTrue(result.ok)
        self.assertFalse(result.skipped)

    def test_short_description_fails(self):
        body = _read("bug_short_description.md")
        result = check_issue(body, ["bug"])
        self.assertFalse(result.ok)
        self.assertIn("short_description", {d.id for d in result.deficiencies})

    def test_placeholder_steps_fails(self):
        body = _read("bug_placeholder_steps.md")
        result = check_issue(body, ["bug"])
        self.assertFalse(result.ok)
        self.assertIn("incomplete_steps", {d.id for d in result.deficiencies})

    def test_short_motivation_fails(self):
        body = _read("feature_short_motivation.md")
        result = check_issue(body, ["enhancement"])
        self.assertFalse(result.ok)
        self.assertIn("short_motivation", {d.id for d in result.deficiencies})

    def test_non_template_skipped(self):
        body = "Free-form issue without template headers."
        result = check_issue(body, [])
        self.assertTrue(result.ok)
        self.assertTrue(result.skipped)


if __name__ == "__main__":
    unittest.main()
