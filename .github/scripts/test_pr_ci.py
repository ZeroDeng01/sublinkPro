import os
import unittest
from unittest.mock import call, mock_open, patch

import pr_ci


INITIAL_SHA = "a" * 40
FIXED_SHA = "b" * 40
BASE = {"id": 1, "full_name": "example/project"}
FORK = {"id": 2, "full_name": "contributor/project", "owner": {"login": "contributor"}}
AUTHOR = {"id": 3, "login": "contributor"}
ADMIN = {"id": 4, "login": "maintainer"}
FIRST_CHECK = {
    "id": 12,
    "run_attempt": 1,
    "event": "pull_request",
    "repository": BASE,
    "path": ".github/workflows/pr-checks.yml",
    "head_repository": FORK,
    "head_branch": "fix/feature",
    "head_sha": INITIAL_SHA,
    "created_at": "2026-11-24T01:01:00Z",
    "pull_requests": [],
    "conclusion": "failure",
}


def pull_request(**overrides):
    return {
        "number": 42,
        "state": "open",
        "base": {"ref": "dev"},
        "head": {"repo": FORK, "ref": "fix/feature", "sha": INITIAL_SHA},
        "user": AUTHOR,
        "created_at": "2026-11-24T01:00:00Z",
        "closed_at": None,
        **overrides,
    }


def recheck_event(author=AUTHOR, body="/recheck"):
    return {"issue": {"number": 42}, "comment": {"user": author, "body": body, "id": 21}, "repository": BASE}


class RecheckTests(unittest.TestCase):
    def setUp(self):
        output = patch.dict(os.environ, GITHUB_OUTPUT="output", GITHUB_STEP_SUMMARY="summary")
        output.start()
        self.addCleanup(output.stop)

    @patch.object(pr_ci, "api")
    def test_regular_comments_and_inexact_commands_do_not_call_api(self, api):
        for body in ("Thanks", "/recheck please", "text\n/recheck", "/RECHECK"):
            pr_ci.prepare_recheck(recheck_event(body=body))
        api.assert_not_called()

    @patch.object(pr_ci, "write_check_target")
    @patch.object(pr_ci, "api")
    def test_wrong_target_and_closed_prs_do_not_start_checks(self, api, write_target):
        for pr in (pull_request(state="closed"), pull_request(base={"ref": "main"})):
            api.return_value = pr
            pr_ci.prepare_recheck(recheck_event())
        write_target.assert_not_called()

    @patch.object(pr_ci, "write_check_target")
    @patch.object(pr_ci, "api")
    def test_only_repository_admins_can_run_another_authors_pr(self, api, write_target):
        pr = pull_request()
        api.side_effect = [pr, {"permission": "write"}, pr, {"permission": "admin"}]
        pr_ci.prepare_recheck(recheck_event(ADMIN))
        write_target.assert_not_called()
        pr_ci.prepare_recheck(recheck_event(ADMIN))
        write_target.assert_called_once_with(pr)

    @patch.object(pr_ci, "write_check_target")
    @patch.object(pr_ci, "api")
    def test_fork_without_approved_checks_is_held(self, api, write_target):
        pr = pull_request()
        untrusted_comment = {"user": AUTHOR, "body": pr_ci.fork_approval_marker(pr)}
        api.side_effect = [pr, [[untrusted_comment]], [{"workflow_runs": [FIRST_CHECK]}],
                           [{"jobs": []}], {"permission": "none"}]
        output = mock_open()
        with patch.object(pr_ci.Path, "open", output):
            pr_ci.prepare_recheck(recheck_event())
        output().write.assert_any_call("awaiting_approval=true")
        write_target.assert_not_called()

    @patch.object(pr_ci, "write_check_target")
    @patch.object(pr_ci, "api")
    def test_failed_first_check_is_enough_to_recheck_a_new_commit(self, api, write_target):
        pr = pull_request(head={**pull_request()["head"], "sha": FIXED_SHA})
        started = {"name": pr_ci.PREPARE_JOB, "conclusion": "success"}
        api.side_effect = [pr, [[]], [{"workflow_runs": [FIRST_CHECK]}],
                           [[{"sha": INITIAL_SHA}]], [{"jobs": [started]}]]
        pr_ci.prepare_recheck(recheck_event(body="/recheck\r\nFix ready"))
        write_target.assert_called_once_with(pr)

    @patch.object(pr_ci, "write_check_target")
    @patch.object(pr_ci, "api")
    def test_approved_report_allows_recheck_after_force_push(self, api, write_target):
        pr = pull_request(head={**pull_request()["head"], "sha": FIXED_SHA})
        report = {"user": {"login": pr_ci.BOT}, "body": pr_ci.fork_approval_marker(pr)}
        api.side_effect = [pr, [[report]]]
        pr_ci.prepare_recheck(recheck_event())
        write_target.assert_called_once_with(pr)


class RunAssociationTests(unittest.TestCase):
    @patch.object(pr_ci, "api")
    def test_fork_run_uses_repository_id_and_sha_when_association_is_missing(self, api):
        pr = pull_request()
        other_fork = pull_request(number=43, head={**pr["head"], "repo": {"id": 9}})
        api.return_value = [[other_fork, pr]]
        self.assertEqual(pr_ci.find_run_pr(FIRST_CHECK), pr)
        api.assert_called_once_with("pulls", paginate=True, head="contributor:fix/feature", base="dev", state="all", per_page=100)

    @patch.object(pr_ci, "api")
    def test_missing_or_ambiguous_match_is_not_reported(self, api):
        api.return_value = [[pull_request(number=42), pull_request(number=43)]]
        self.assertIsNone(pr_ci.find_run_pr(FIRST_CHECK))
        api.return_value = [[pull_request(base={"ref": "main"})]]
        self.assertIsNone(pr_ci.find_run_pr(FIRST_CHECK))

    @patch.object(pr_ci, "api")
    def test_only_current_results_replace_earlier_attempts(self, api):
        new_backend = {"name": "Backend Checks", "conclusion": "success"}
        old_backend = {"name": "Backend Checks", "conclusion": "failure"}
        frontend = {"name": "Frontend Checks", "conclusion": "success"}
        api.side_effect = [[{"jobs": [new_backend]}], [{"jobs": [old_backend, frontend]}]]
        jobs = pr_ci.run_jobs({**FIRST_CHECK, "run_attempt": 2}, set(pr_ci.CHECKS))
        self.assertEqual(jobs, {"Backend Checks": new_backend, "Frontend Checks": frontend})


class ReportTests(unittest.TestCase):
    @patch.dict(os.environ, GITHUB_REPOSITORY="example/project", GITHUB_SERVER_URL="https://github.com")
    def test_report_has_failed_and_unrun_checks_and_both_commits(self):
        pr = pull_request(head={**pull_request()["head"], "sha": FIXED_SHA})
        report = pr_ci.render_report("Check", {"GO_BUILD": "failure", "FRONTEND_BUILD": "success"},
                                    "failure", INITIAL_SHA, pr, 12, 2, comment_id=21)
        self.assertIn("`go build ./...` | ❌", report)
        self.assertIn("`go test ./...` | ⚪", report)
        self.assertIn("`yarn run build` | ✅", report)
        self.assertIn(f"/commits/{INITIAL_SHA}", report)
        self.assertIn(f"/commits/{FIXED_SHA}", report)
        self.assertIn("/actions/runs/12/attempts/2", report)
        self.assertIn("#issuecomment-21", report)

    @patch.object(pr_ci, "api")
    def test_retry_updates_the_bot_comment_and_preserves_newer_attempts(self, api):
        previous = {"id": 777, "user": {"login": "github-actions[bot]"}, "body": "<!-- pr-check-report:12:1 -->\nOld"}
        imitation = {**previous, "id": 123, "user": AUTHOR}
        api.side_effect = [[[imitation, previous]], {}]
        pr_ci.publish_comment(42, 12, 2, "New")
        self.assertEqual(api.call_args_list[1], call("issues/comments/777", method="PATCH", body="<!-- pr-check-report:12:2 -->\nNew"))

        api.reset_mock(side_effect=True)
        api.return_value = [[{**previous, "body": "<!-- pr-check-report:12:2 -->\nNew"}]]
        pr_ci.publish_comment(42, 12, 1, "Old")
        api.assert_called_once_with("issues/42/comments", paginate=True, per_page=100)

    @patch.object(pr_ci, "publish_comment")
    @patch.object(pr_ci, "api")
    def test_publisher_rejects_a_run_from_another_workflow(self, api, publish):
        api.return_value = {**FIRST_CHECK, "path": ".github/workflows/unrelated.yml"}
        pr_ci.report_automatic({"workflow_run": FIRST_CHECK, "repository": BASE})
        publish.assert_not_called()

    @patch.dict(os.environ, GITHUB_REPOSITORY="example/project", GITHUB_SERVER_URL="https://github.com")
    @patch.object(pr_ci, "publish_comment")
    @patch.object(pr_ci, "api")
    def test_automatic_report_preserves_step_failure_for_fork(self, api, publish):
        pr = pull_request()
        backend = {"name": "Backend Checks", "steps": [{"name": "Run Go tests", "conclusion": "failure"}]}
        prepare = {"name": pr_ci.PREPARE_JOB, "conclusion": "success"}
        api.side_effect = [FIRST_CHECK, [{"jobs": [prepare, backend]}], [[pr]]]
        pr_ci.report_automatic({"workflow_run": FIRST_CHECK, "repository": BASE})
        number, run_id, attempt, body = publish.call_args.args
        self.assertEqual((number, run_id, attempt), (42, 12, 1))
        self.assertIn("`go test ./...` | ❌", body)
        self.assertIn(pr_ci.fork_approval_marker(pr), body)


if __name__ == "__main__":
    unittest.main()
