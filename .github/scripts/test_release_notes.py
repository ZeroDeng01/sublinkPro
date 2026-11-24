import unittest
from unittest.mock import call, patch

import publish_dev_ref
import release_notes


PREVIEW = "preview-20261124-4a211cb"
OLDER_PREVIEW = "preview-20261123-abcdef0"
FAILED_PREVIEW = "preview-20261123-abcdef1"
SHA = "a" * 40


class ReleaseNotesTests(unittest.TestCase):
    @patch.object(release_notes, "git", return_value="- feat: useful change (4a211cb)")
    @patch.object(release_notes, "api")
    @patch.object(release_notes, "nearest_tag", side_effect=[FAILED_PREVIEW, OLDER_PREVIEW])
    def test_preview_uses_published_baseline_and_dated_image(self, nearest, api, git):
        api.side_effect = [{"draft": True, "prerelease": True}, {"draft": False, "prerelease": True}]
        title, body = release_notes.release_notes(PREVIEW, preview=True)
        nearest.assert_has_calls([call("preview-*", (PREVIEW,)), call("preview-*", (PREVIEW, FAILED_PREVIEW))])
        git.assert_called_once_with("log", "--format=- %s (%h)", f"{OLDER_PREVIEW}..HEAD")
        self.assertEqual(title, "Preview 20261124-4a211cb")
        self.assertTrue(body.startswith("## 预发布\n"))
        self.assertIn("feat: useful change", body)
        self.assertIn(f"zerodeng/sublink-pro:{PREVIEW}", body)
        self.assertNotIn("zerodeng/sublink-pro:latest", body)
        self.assertNotIn("Actions", body)
        self.assertNotIn("推送日期", body)

    @patch.object(release_notes, "git", return_value="- fix: preview-only update")
    @patch.object(release_notes, "nearest_tag", side_effect=[None, "v1.3.0"])
    def test_first_preview_starts_at_the_latest_stable_tag(self, nearest, git):
        release_notes.release_notes(PREVIEW, preview=True)
        nearest.assert_called_with("v*", [PREVIEW], "HEAD")
        git.assert_called_once_with("log", "--format=- %s (%h)", "v1.3.0..HEAD")

    @patch.object(release_notes, "git", return_value="- fix: stable update")
    @patch.object(release_notes, "nearest_tag", return_value="v1.3.0")
    def test_stable_notes_include_versioned_and_latest_images(self, nearest, git):
        title, body = release_notes.release_notes("v1.3.1", preview=False)
        nearest.assert_called_once_with("v*", ["v1.3.1"], "HEAD^")
        git.assert_called_once_with("log", "--format=- %s (%h)", "v1.3.0..HEAD")
        self.assertEqual(title, "Release v1.3.1")
        self.assertIn("zerodeng/sublink-pro:v1.3.1", body)
        self.assertIn("zerodeng/sublink-pro:latest", body)


class DevelopmentPublicationTests(unittest.TestCase):
    @patch.object(publish_dev_ref, "api", return_value=None)
    def test_creates_marker_for_the_first_successful_dev_publish(self, api):
        publish_dev_ref.record_dev_image(SHA)
        self.assertEqual(api.call_args_list, [
            call("git/ref/tags/dev-image", allow_missing=True),
            call("git/refs", data={"ref": "refs/tags/dev-image", "sha": SHA}),
        ])

    @patch.object(publish_dev_ref, "api", return_value={"object": {"sha": "b" * 40}})
    def test_advances_marker_for_a_new_dev_image(self, api):
        publish_dev_ref.record_dev_image(SHA)
        api.assert_called_with("git/refs/tags/dev-image", method="PATCH", data={"sha": SHA, "force": True})
