import unittest
from unittest.mock import call, patch

import release_notes


PREVIEW = "preview-20261124-171211123"
OLDER_PREVIEW = "preview-20261123-abcdef0"
FAILED_PREVIEW = "preview-20261123-abcdef1"
STABLE = "v1.3.1"


def published_release(tag, time, *, draft=False, prerelease=False):
    return {"tag_name": tag, "published_at": f"2026-11-24T{time}:00Z", "draft": draft, "prerelease": prerelease}


class ReleaseNotesTests(unittest.TestCase):
    @patch.object(release_notes, "api")
    @patch.object(release_notes, "git")
    def test_preview_starts_at_stable_release_published_after_older_preview(self, git, api):
        git.side_effect = ["\n".join([PREVIEW, OLDER_PREVIEW, FAILED_PREVIEW, STABLE]), "- feat: useful change (4a211cb)"]
        api.return_value = [
            published_release(OLDER_PREVIEW, "08:00"),
            published_release("v2.0.0", "14:00"),
            published_release(FAILED_PREVIEW, "13:00", draft=True),
            published_release(PREVIEW, "12:00"),
            published_release(STABLE, "09:00"),
        ]

        title, body = release_notes.release_notes(PREVIEW, preview=True)

        self.assertEqual(git.call_args_list, [
            call("for-each-ref", "--merged=HEAD", "--format=%(refname:strip=2)", "refs/tags"),
            call("log", "--format=- %s (%h)", f"refs/tags/{STABLE}..HEAD"),
        ])
        self.assertEqual(title, "Preview 20261124-171211123")
        self.assertTrue(body.startswith("## 预发布\n"))
        self.assertIn(f"{STABLE} → {PREVIEW}", body)
        self.assertIn("feat: useful change", body)
        self.assertIn(f"zerodeng/sublink-pro:{PREVIEW}", body)
        self.assertIn("- `zerodeng/sublink-pro:preview`（滚动预发布镜像）", body)
        self.assertNotIn("zerodeng/sublink-pro:latest", body)
        self.assertNotIn("Actions", body)
        self.assertNotIn("推送日期", body)

    @patch.object(release_notes, "api")
    @patch.object(release_notes, "git", return_value=f"{OLDER_PREVIEW}\n{STABLE}")
    def test_latest_published_release_can_be_a_preview_on_a_later_api_page(self, git, api):
        first_page = [published_release(f"v1.2.{index}", "07:00") for index in range(99)]
        first_page.append(published_release(STABLE, "09:00"))
        api.side_effect = [first_page, [published_release(OLDER_PREVIEW, "10:00")]]

        self.assertEqual(release_notes.previous_published_release(PREVIEW), OLDER_PREVIEW)
        self.assertEqual(api.call_args_list, [
            call("releases?per_page=100&page=1"),
            call("releases?per_page=100&page=2"),
        ])

    @patch.object(release_notes, "api", return_value=[])
    @patch.object(release_notes, "git", side_effect=[STABLE, "- feat: first public release"])
    def test_preview_with_no_published_ancestor_lists_entire_history(self, git, api):
        release_notes.release_notes(PREVIEW, preview=True)
        git.assert_called_with("log", "--format=- %s (%h)", "HEAD")

    @patch.object(release_notes, "git", return_value="- fix: stable update")
    @patch.object(release_notes, "nearest_tag", return_value="v1.3.0")
    def test_stable_notes_include_versioned_and_latest_images(self, nearest, git):
        title, body = release_notes.release_notes("v1.3.1", preview=False)
        nearest.assert_called_once_with("v*", ["v1.3.1"], "HEAD^")
        git.assert_called_once_with("log", "--format=- %s (%h)", "refs/tags/v1.3.0..HEAD")
        self.assertEqual(title, "Release v1.3.1")
        self.assertIn("zerodeng/sublink-pro:v1.3.1", body)
        self.assertIn("zerodeng/sublink-pro:latest", body)
        self.assertNotIn("zerodeng/sublink-pro:preview", body)

    @patch.object(release_notes, "api")
    def test_rolling_preview_selects_latest_public_preview_across_unordered_pages(self, api):
        first_page = [published_release(f"v1.2.{index}", "15:00") for index in range(99)]
        first_page.append(published_release(OLDER_PREVIEW, "09:00", prerelease=True))
        api.side_effect = [
            first_page,
            [
                published_release(PREVIEW, "10:00", prerelease=True),
                published_release(FAILED_PREVIEW, "14:00", draft=True, prerelease=True),
                published_release("v2.0.0-rc.1", "13:00", prerelease=True),
                published_release("preview-20261124-171211124", "12:00"),
                {**published_release("preview-20261124-171211125", "11:00", prerelease=True), "published_at": None},
            ],
        ]

        self.assertEqual(release_notes.latest_published_preview(), PREVIEW)
        self.assertEqual(api.call_args_list, [
            call("releases?per_page=100&page=1"),
            call("releases?per_page=100&page=2"),
        ])

    @patch.object(release_notes, "api", return_value=[published_release(STABLE, "09:00")])
    def test_rolling_preview_requires_a_public_preview(self, api):
        self.assertIsNone(release_notes.latest_published_preview())
