from datetime import datetime, timezone
from io import BytesIO
import unittest
from unittest.mock import call, patch
from urllib.error import HTTPError

import prepare_preview


COMMIT = "274ad3e415391cb3f3705fd5efcd65630bbb7f26"
COLLISION = "274ad3e965391cb3f3705fd5efcd65630bbb7f26"
TAG = "pre-261124003051"
NEXT_TAG = "pre-261124003052"
NOW = datetime(2026, 11, 23, 16, 30, 51, 123456, tzinfo=timezone.utc)


def tag_ref(tag, commit=COMMIT):
    return {"ref": f"refs/tags/{tag}", "object": {"type": "commit", "sha": commit}}


class PreviewTagTests(unittest.TestCase):
    def setUp(self):
        self.clock = self.enterContext(patch.object(prepare_preview, "datetime"))
        self.clock.now.return_value = NOW

    @patch.object(prepare_preview, "api")
    def test_reserves_a_compact_shanghai_timestamp_for_the_commit(self, api):
        api.side_effect = [[], tag_ref(TAG)]

        self.assertEqual(prepare_preview.reserve_preview_tag(COMMIT), TAG)
        self.clock.now.assert_called_once_with(timezone.utc)
        api.assert_has_calls([
            call("git/matching-refs/tags/pre-"),
            call("git/refs", data={"ref": f"refs/tags/{TAG}", "sha": COMMIT}),
        ])

    @patch.object(prepare_preview, "api")
    def test_retry_reuses_the_existing_tag_for_the_verified_commit(self, api):
        api.return_value = [tag_ref(TAG)]
        self.assertEqual(prepare_preview.reserve_preview_tag(COMMIT), TAG)
        api.assert_called_once_with("git/matching-refs/tags/pre-")
        self.clock.now.assert_not_called()

    @patch.object(prepare_preview, "api")
    def test_uses_the_next_second_if_timestamp_belongs_to_another_commit(self, api):
        api.side_effect = [[tag_ref(TAG, COLLISION)], tag_ref(NEXT_TAG)]

        self.assertEqual(prepare_preview.reserve_preview_tag(COMMIT), NEXT_TAG)
        api.assert_called_with("git/refs", data={"ref": f"refs/tags/{NEXT_TAG}", "sha": COMMIT})

    @patch.object(prepare_preview, "api")
    def test_timestamp_collision_advances_to_the_next_day_after_midnight(self, api):
        self.clock.now.return_value = datetime(2026, 11, 24, 15, 59, 59, 999000, tzinfo=timezone.utc)
        next_day = "pre-261125000000"
        api.side_effect = [[tag_ref("pre-261124235959", COLLISION)], tag_ref(next_day)]

        self.assertEqual(prepare_preview.reserve_preview_tag(COMMIT), next_day)

    @patch.object(prepare_preview, "api")
    def test_verifies_the_winner_of_a_concurrent_tag_reservation(self, api):
        collision_error = HTTPError("https://api.github.com", 422, "Already exists", {}, BytesIO())
        api.side_effect = [[], collision_error, tag_ref(TAG, COLLISION), tag_ref(NEXT_TAG)]

        self.assertEqual(prepare_preview.reserve_preview_tag(COMMIT), NEXT_TAG)
        self.assertIn(call(f"git/ref/tags/{TAG}", allow_missing=True), api.call_args_list)

    @patch.object(prepare_preview, "api")
    def test_validation_failure_without_an_existing_tag_is_not_treated_as_collision(self, api):
        validation_error = HTTPError("https://api.github.com", 422, "Validation failed", {}, BytesIO())
        api.side_effect = [[], validation_error, None]

        with self.assertRaises(HTTPError):
            prepare_preview.reserve_preview_tag(COMMIT)

    @patch.object(prepare_preview, "api")
    def test_dereferences_an_annotated_tag_to_check_the_commit(self, api):
        annotation = {"ref": f"refs/tags/{TAG}", "object": {"type": "tag", "sha": "annotation"}}
        api.side_effect = [[annotation], {"object": {"type": "commit", "sha": COMMIT}}]

        self.assertEqual(prepare_preview.reserve_preview_tag(COMMIT), TAG)

    @patch.object(prepare_preview, "api")
    def test_partial_and_incompatible_releases_do_not_count_as_published(self, api):
        api.side_effect = [None, {"draft": True}, {"draft": False, "prerelease": False}]
        self.assertFalse(prepare_preview.preview_is_published(TAG))
        self.assertFalse(prepare_preview.preview_is_published(TAG))
        with self.assertRaises(RuntimeError):
            prepare_preview.preview_is_published(TAG)

    @patch.object(prepare_preview, "api")
    def test_published_preview_is_reused(self, api):
        api.return_value = {"name": TAG, "draft": False, "prerelease": True}
        self.assertTrue(prepare_preview.preview_is_published(TAG))


if __name__ == "__main__":
    unittest.main()
