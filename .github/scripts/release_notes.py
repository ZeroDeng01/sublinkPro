"""Generate a Preview or stable release's title and body."""

import os
from pathlib import Path
import subprocess

from prepare_preview import PREVIEW_TAG, api, preview_title


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


def previous_preview(current_tag):
    excluded = (current_tag,)
    while tag := nearest_tag("preview-*", excluded):
        if PREVIEW_TAG.fullmatch(tag):
            release = api(f"releases/tags/{tag}", allow_missing=True)
            if release and not release["draft"] and release["prerelease"]:
                return tag
        excluded = (*excluded, tag)
    return None


def release_notes(tag, preview):
    if preview and not PREVIEW_TAG.fullmatch(tag):
        raise ValueError("Expected the reserved preview tag.")

    previous = previous_preview(tag) if preview else None
    previous = previous or nearest_tag("v*", [tag], "HEAD" if preview else "HEAD^")
    history = f"{previous}..HEAD" if previous else "HEAD"
    changes = git("log", "--format=- %s (%h)", history)
    title = preview_title(tag) if preview else f"Release {tag}"
    heading = "## 🚀 更新内容" if preview or previous else "## 🎉 首次发布"
    if previous:
        heading += f" ({previous} → {tag})"
    lines = ["## 预发布", ""] if preview else []
    lines.extend([heading, "", changes or "本次发布没有新增提交。", "", "### Docker 镜像", "", f"- `zerodeng/sublink-pro:{tag}`"])
    if not preview:
        lines.append("- `zerodeng/sublink-pro:latest`")
    lines.extend(["", "### 📦 发布文件", "", *(f"- `{asset}` - {platform}" for asset, platform in ASSETS)])
    return title, "\n".join(lines) + "\n"


if __name__ == "__main__":
    preview = os.environ["GITHUB_REF"] == "refs/heads/main"
    tag = os.environ["PREVIEW_TAG"] if preview else os.environ["GITHUB_REF_NAME"]
    title, body = release_notes(tag, preview)
    Path("changelog.md").write_text(body, encoding="utf-8")
    with Path(os.environ["GITHUB_OUTPUT"]).open("a", encoding="utf-8") as output:
        print(f"tag={tag}", file=output)
        print(f"title={title}", file=output)
