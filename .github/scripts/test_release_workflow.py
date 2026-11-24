from pathlib import Path
import re
import unittest


WORKFLOW = Path(__file__).resolve().parents[1] / "workflows" / "build-release.yml"


class ReleaseWorkflowTests(unittest.TestCase):
    def test_binary_merge_checks_the_matrix_when_preview_preparation_is_skipped(self):
        merge_job = re.search(r"^  merge-binaries:\n((?:    [^\n]*\n|\n)*)", WORKFLOW.read_text(), re.MULTILINE)
        self.assertIsNotNone(merge_job)
        self.assertRegex(merge_job[1], r"(?m)^    needs: build-go-binaries$")
        self.assertRegex(
            merge_job[1],
            r"(?m)^    if:\s+(?:>-\s+)?(?:\$\{\{\s+)?!cancelled\(\)\s+&&\s+"
            r"needs\.build-go-binaries\.result\s+==\s+'success'(?:\s+\}\})?$",
        )

    def test_rolling_preview_promotion_is_serialized_after_publication(self):
        alias_job = re.search(r"^  publish-preview-alias:\n((?:    [^\n]*\n|\n)*)", WORKFLOW.read_text(), re.MULTILINE)
        self.assertIsNotNone(alias_job)
        config = alias_job[1]
        self.assertIn("needs: [prepare-preview, release]", config)
        self.assertIn("github.ref == 'refs/heads/main'", config)
        self.assertIn("needs.release.result == 'success' || needs.prepare-preview.outputs.published == 'true'", config)
        self.assertIn("group: docker-preview-alias\n      cancel-in-progress: false", config)
