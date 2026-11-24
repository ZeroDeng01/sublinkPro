"""Resolve PR check requests and publish reports using GitHub's API."""

import json
import os
from pathlib import Path
import subprocess
import sys


WORKFLOW = "pr-checks.yml"
PREPARE_JOB = "Prepare PR Check"
BOT = "github-actions[bot]"
CHECKS = {
    "Backend Checks": {
        "Compile Go packages": ("Go 编译 / `go build ./...`", "GO_BUILD"),
        "Run golangci-lint": ("Go lint / `golangci-lint`", "GO_LINT"),
        "Run Go tests": ("Go 测试 / `go test ./...`", "GO_TEST"),
    },
    "Frontend Checks": {
        "Run frontend lint": ("前端 lint / `yarn run lint`", "FRONTEND_LINT"),
        "Build frontend": ("前端构建 / `yarn run build`", "FRONTEND_BUILD"),
    },
}
STATUS_LABELS = {
    "success": "✅ 通过 / Passed",
    "failure": "❌ 失败 / Failed",
    "cancelled": "⚪ 已取消 / Cancelled",
}


def api(path, *, method="GET", paginate=False, **fields):
    endpoint = f"repos/{os.environ['GITHUB_REPOSITORY']}/{path}"
    command = ["gh", "api", "--method", method]
    if paginate:
        command.extend(["--paginate", "--slurp"])
    for name, value in fields.items():
        command.extend(["--raw-field", f"{name}={value}"])
    output = subprocess.check_output([*command, endpoint], text=True)
    return json.loads(output)


def is_pr_check_run(run, repository_id):
    return (
        run["event"] == "pull_request"
        and run["repository"]["id"] == repository_id
        and run["path"].split("@", 1)[0] == f".github/workflows/{WORKFLOW}"
    )


def has_run_head(run, pr):
    return (
        pr["head"]["repo"] is not None
        and run["head_repository"] is not None
        and pr["head"]["repo"]["id"] == run["head_repository"]["id"]
        and pr["head"]["ref"] == run["head_branch"]
        and pr["created_at"] <= run["created_at"]
        and (pr["closed_at"] is None or run["created_at"] <= pr["closed_at"])
    )


def run_matches_pr(run, pr):
    if pr["base"]["ref"] != "dev" or not has_run_head(run, pr):
        return False
    if run["pull_requests"]:
        return any(link["number"] == pr["number"] for link in run["pull_requests"])
    if run["head_sha"] == pr["head"]["sha"]:
        return True
    pages = api(f"pulls/{pr['number']}/commits", paginate=True, per_page=100)
    return any(commit["sha"] == run["head_sha"] for page in pages for commit in page)


def find_run_pr(run):
    if not run["head_repository"] or not run["head_branch"]:
        return None
    if run["pull_requests"]:
        candidates = [api(f"pulls/{link['number']}") for link in run["pull_requests"]]
    else:
        owner = run["head_repository"]["owner"]["login"]
        # GitHub can omit the PR association for forks. Verify the repository ID
        # and commit separately because different forks can share a branch name.
        pages = api("pulls", paginate=True, head=f"{owner}:{run['head_branch']}",
                    base="dev", state="all", per_page=100)
        candidates = [pr for page in pages for pr in page]
    matched = [pr for pr in candidates if run_matches_pr(run, pr)]
    return matched[0] if len(matched) == 1 else None


def run_jobs(run, names):
    jobs = {}
    # Re-running only failed jobs can leave successful jobs in older attempts.
    for attempt in range(run["run_attempt"], 0, -1):
        pages = api(f"actions/runs/{run['id']}/attempts/{attempt}/jobs", paginate=True, per_page=100)
        for page in pages:
            for job in page["jobs"]:
                if job["name"] in names:
                    jobs.setdefault(job["name"], job)
        if names <= jobs.keys():
            break
    return jobs


def fork_approval_marker(pr):
    return f"<!-- pr-check-approved-fork:{pr['head']['repo']['id']} -->"


def fork_check_has_run(pr, repository_id):
    # The bot report preserves proof of an approved run when a force-push
    # removes the old commit from the pull request's commit list.
    marker = fork_approval_marker(pr)
    for page in api(f"issues/{pr['number']}/comments", paginate=True, per_page=100):
        for comment in page:
            if (comment["user"] or {}).get("login") == BOT and marker in (comment["body"] or "").splitlines():
                return True

    pages = api(f"actions/workflows/{WORKFLOW}/runs", paginate=True, event="pull_request",
                branch=pr["head"]["ref"], status="completed", created=f">={pr['created_at']}", per_page=100)
    for page in pages:
        for run in page["workflow_runs"]:
            if not is_pr_check_run(run, repository_id) or not run_matches_pr(run, pr):
                continue
            # Expired or unapproved runs have no successful preparation job.
            job = run_jobs(run, {PREPARE_JOB}).get(PREPARE_JOB)
            if job and job["conclusion"] == "success":
                return True
    return False


def write_check_target(pr):
    with Path(os.environ["GITHUB_OUTPUT"]).open("a") as output:
        print("should_run=true", file=output)
        print(f"checkout_repository={pr['head']['repo']['full_name']}", file=output)
        print(f"checkout_ref={pr['head']['sha']}", file=output)
    with Path(os.environ["GITHUB_STEP_SUMMARY"]).open("a") as summary:
        print(f"Checking PR #{pr['number']} at `{pr['head']['sha']}`.", file=summary)


def is_repository_admin(user):
    permission = api(f"collaborators/{user['login']}/permission")
    return permission["permission"] == "admin"


def prepare_recheck(event):
    comment = event["comment"]
    if comment["body"].partition("\n")[0].removesuffix("\r") != "/recheck":
        return
    pr = api(f"pulls/{event['issue']['number']}")
    if pr["state"] != "open" or pr["base"]["ref"] != "dev" or not pr["head"]["repo"]:
        return

    author = comment["user"]
    if author["id"] != pr["user"]["id"]:
        if not is_repository_admin(author):
            return
    elif pr["head"]["repo"]["id"] != event["repository"]["id"]:
        if not fork_check_has_run(pr, event["repository"]["id"]) and not is_repository_admin(author):
            # An administrator can explicitly trigger the check while the fork
            # is waiting for a maintainer to approve the automatic run.
            with Path(os.environ["GITHUB_OUTPUT"]).open("a") as output:
                print("awaiting_approval=true", file=output)
            return

    write_check_target(pr)


def publish_comment(pr_number, run_id, attempt, body):
    marker = f"<!-- pr-check-report:{run_id}:"
    body = f"{marker}{attempt} -->\n{body}"
    endpoint = f"issues/{pr_number}/comments"
    for page in api(endpoint, paginate=True, per_page=100):
        for comment in page:
            if (comment["user"] or {}).get("login") == BOT and (comment["body"] or "").startswith(marker):
                previous = comment["body"].splitlines()[0].removeprefix(marker).removesuffix(" -->")
                if previous.isdecimal() and int(previous) <= attempt:
                    api(f"issues/comments/{comment['id']}", method="PATCH", body=body)
                return
    api(endpoint, method="POST", body=body)


def render_report(title, results, overall, checked_sha, pr, run_id, attempt, comment_id=None):
    repo_url = f"{os.environ['GITHUB_SERVER_URL']}/{os.environ['GITHUB_REPOSITORY']}"
    pr_url = f"{repo_url}/pull/{pr['number']}"
    run_url = f"{repo_url}/actions/runs/{run_id}/attempts/{attempt}"
    if overall == "failure" or "failure" in results.values():
        result = "❌ 未通过 / Failed"
    elif overall == "success" and all(status == "success" for status in results.values()):
        result = "✅ 全部通过 / All checks passed"
    else:
        result = "⚪ 未完成 / Incomplete"

    lines = [f"### {title} · {result}", "",
             f"检查提交 / Commit [`{checked_sha[:7]}`]({pr_url}/commits/{checked_sha})"]
    current_sha = pr["head"]["sha"]
    if current_sha != checked_sha:
        lines.extend(["", f"当前提交 / Current commit [`{current_sha[:7]}`]({pr_url}/commits/{current_sha})"])
    lines.extend(["", "| 检查项 / Check | 结果 / Result |", "| --- | --- |"])
    for steps in CHECKS.values():
        for label, key in steps.values():
            lines.append(f"| {label} | {STATUS_LABELS.get(results.get(key), '⚪ 未运行 / Not run')} |")
    links = [f"[Actions 日志 / Logs]({run_url})"]
    if comment_id is not None:
        links.insert(0, f"[触发评论 / Command]({pr_url}#issuecomment-{comment_id})")
    lines.extend(["", " · ".join(links)])
    return "\n".join(lines)


def report_automatic(event):
    source = event["workflow_run"]
    run = api(f"actions/runs/{source['id']}/attempts/{source['run_attempt']}")
    if not is_pr_check_run(run, event["repository"]["id"]):
        return
    jobs = run_jobs(run, {PREPARE_JOB, *CHECKS})
    prepare = jobs.get(PREPARE_JOB)
    if not prepare or prepare["conclusion"] == "skipped":
        return
    pr = find_run_pr(run)
    if pr is None:
        print(f"::warning::Could not verify a single dev pull request for check run {run['id']}.")
        return

    results = {}
    for name, checks in CHECKS.items():
        job = jobs.get(name, {})
        for step in job.get("steps", []):
            if step["name"] in checks:
                results[checks[step["name"]][1]] = step["conclusion"]
    body = render_report("PR 自动检查 / Automatic PR checks", results, run["conclusion"],
                         run["head_sha"], pr, run["id"], run["run_attempt"])
    if prepare["conclusion"] == "success" and pr["head"]["repo"]["id"] != event["repository"]["id"]:
        body = f"{fork_approval_marker(pr)}\n{body}"
    publish_comment(pr["number"], run["id"], run["run_attempt"], body)


def report_recheck(event):
    pr = api(f"pulls/{event['issue']['number']}")
    if pr["base"]["ref"] != "dev":
        return
    run_id = int(os.environ["GITHUB_RUN_ID"])
    attempt = int(os.environ["GITHUB_RUN_ATTEMPT"])
    if os.environ["AWAITING_APPROVAL"] == "true":
        if pr["state"] != "open" or not pr["head"]["repo"]:
            return
        body = ("⏳ `/recheck` 等待首轮检查完成。维护者可批准 fork 的首轮检查，"
                "或由仓库管理员评论 `/recheck` 明确触发检查。\n\n"
                "`/recheck` is waiting for the first check run to finish. A maintainer can approve the initial "
                "fork run, or a repository administrator can request a recheck.")
    else:
        results = {key: os.environ[key] for steps in CHECKS.values() for _, key in steps.values()}
        job_results = {os.environ["BACKEND_RESULT"], os.environ["FRONTEND_RESULT"]}
        if "failure" in job_results:
            overall = "failure"
        elif job_results == {"success"}:
            overall = "success"
        else:
            overall = "incomplete"
        body = render_report("`/recheck` 复查 / PR recheck", results, overall,
                             os.environ["CHECKED_SHA"], pr, run_id, attempt, event["comment"]["id"])
    publish_comment(pr["number"], run_id, attempt, body)


if __name__ == "__main__":
    event = json.loads(Path(os.environ["GITHUB_EVENT_PATH"]).read_text())
    commands = {
        "prepare-recheck": prepare_recheck,
        "report-automatic": report_automatic,
        "report-recheck": report_recheck,
    }
    commands[sys.argv[1]](event)
