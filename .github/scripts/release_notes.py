"""Generate release notes and look up published release tags."""

import os
from pathlib import Path
import subprocess
import sys

from prepare_preview import PREVIEW_TAG, api


ASSETS = (
    ("sublinkPro-linux-amd64", "Linux x64"),
    ("sublinkPro-linux-arm64", "Linux ARM64"),
    ("sublinkPro-linux-armv7", "Linux ARMv7 (32-bit)"),
    ("sublinkPro-linux-x86", "Linux x86 (32-bit)"),
    ("sublinkPro-windows-amd64.exe", "Windows x64"),
    ("sublinkPro-windows-x86.exe", "Windows x86 (32-bit)"),
    ("sublinkPro-darwin-amd64", "macOS x64"),
    ("sublinkPro-darwin-arm64", "macOS ARM64 (Apple Silicon)"),
)


def git(*args):
    return subprocess.check_output(["git", *args], text=True).strip()


def nearest_tag(pattern, excluded, ref="HEAD"):
    exclusions = [f"--exclude={tag}" for tag in excluded]
    result = subprocess.run(
        ["git", "describe", "--tags", "--abbrev=0", f"--match={pattern}", *exclusions, ref],
        text=True, capture_output=True, check=False,
    )
    return result.stdout.strip() if result.returncode == 0 else None


def latest_published_tag(matches):
    latest = ("", "")
    page = 1
    while True:
        releases = api(f"releases?per_page=100&page={page}")
        for release in releases:
            if not release["draft"] and release["published_at"] and matches(release):
                latest = max(latest, (release["published_at"], release["tag_name"]))
        if len(releases) < 100:
            return latest[1] or None
        page += 1


def previous_published_release(current_tag):
    reachable_tags = set(git("for-each-ref", "--merged=HEAD", "--format=%(refname:strip=2)", "refs/tags").splitlines())
    reachable_tags.discard(current_tag)
    if not reachable_tags:
        return None
    return latest_published_tag(lambda release: release["tag_name"] in reachable_tags)


def latest_published_preview():
    return latest_published_tag(lambda release: release["prerelease"] and PREVIEW_TAG.fullmatch(release["tag_name"]))


def release_notes(tag, preview):
    if preview and not PREVIEW_TAG.fullmatch(tag):
        raise ValueError("Expected the reserved preview tag.")

    previous = previous_published_release(tag) if preview else nearest_tag("v*", [tag], "HEAD^")
    history = f"refs/tags/{previous}..HEAD" if previous else "HEAD"
    changes = git("log", "--format=- %s (%h)", history)
    title = tag if preview else f"Release {tag}"
    heading = "## 🚀 更新内容" if preview or previous else "## 🎉 首次发布"
    if previous:
        heading += f" ({previous} → {tag})"
    lines = ["## 预发布", ""] if preview else []
    lines.extend([heading, "", changes or "本次发布没有新增提交。", "", "### Docker 镜像", "", f"- `zerodeng/sublink-pro:{tag}`"])
    if preview:
        lines.append("- `zerodeng/sublink-pro:preview`（滚动预发布镜像）")
    else:
        lines.append("- `zerodeng/sublink-pro:latest`")
    lines.extend(["", "### 📦 发布文件", "", *(f"- `{asset}` - {platform}" for asset, platform in ASSETS)])
    return title, "\n".join(lines) + "\n"


if __name__ == "__main__":
    if sys.argv[1:] == ["--latest-preview-tag"]:
        latest = latest_published_preview()
        if not latest:
            raise SystemExit("No published Preview release found.")
        print(latest)
        sys.exit(0)

    preview = os.environ["GITHUB_REF"] == "refs/heads/main"
    tag = os.environ["PREVIEW_TAG"] if preview else os.environ["GITHUB_REF_NAME"]
    title, body = release_notes(tag, preview)
    Path("changelog.md").write_text(body, encoding="utf-8")
    with Path(os.environ["GITHUB_OUTPUT"]).open("a", encoding="utf-8") as output:
        print(f"tag={tag}", file=output)
        print(f"title={title}", file=output)
