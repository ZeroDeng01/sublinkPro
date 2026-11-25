"""Reserve the dated Git tag shared by a main-branch image and prerelease."""

from datetime import datetime, timedelta, timezone
import json
import os
from pathlib import Path
import re
from urllib.error import HTTPError
from urllib.request import Request, urlopen
from zoneinfo import ZoneInfo


PREVIEW_TAG = re.compile(r"pre-[0-9]{12}")


def api(path, *, data=None, allow_missing=False):
    repository = f"{os.environ['GITHUB_API_URL']}/repos/{os.environ['GITHUB_REPOSITORY']}"
    request = Request(
        f"{repository}/{path}",
        headers={
            "Accept": "application/vnd.github+json",
            "Authorization": f"Bearer {os.environ['GH_TOKEN']}",
            "Content-Type": "application/json",
            "User-Agent": "sublinkpro-preview-release",
        },
        data=json.dumps(data).encode() if data is not None else None,
    )
    try:
        with urlopen(request, timeout=30) as response:
            return json.load(response)
    except HTTPError as error:
        if allow_missing and error.code == 404:
            error.close()
            return None
        raise


def tag_commit(ref):
    target = ref["object"]
    while target["type"] == "tag":
        target = api(f"git/tags/{target['sha']}")["object"]
    return target["sha"] if target["type"] == "commit" else None


def reserve_preview_tag(commit):
    if not re.fullmatch(r"[0-9a-f]{40}", commit):
        raise ValueError("The preview release requires a full commit SHA.")

    refs = api("git/matching-refs/tags/pre-")
    tags = {ref["ref"].removeprefix("refs/tags/"): ref for ref in refs}
    for tag, ref in sorted(tags.items()):
        if PREVIEW_TAG.fullmatch(tag) and tag_commit(ref) == commit:
            return tag

    timestamp = datetime.now(timezone.utc).astimezone(ZoneInfo("Asia/Shanghai"))
    while True:
        tag = f"pre-{timestamp:%y%m%d%H%M%S}"
        timestamp += timedelta(seconds=1)
        if tag in tags:
            continue
        try:
            api("git/refs", data={"ref": f"refs/tags/{tag}", "sha": commit})
        except HTTPError as error:
            if error.code not in (409, 422):
                raise
            error.close()
            # A concurrent run may have reserved the short tag after the list call.
            occupied = api(f"git/ref/tags/{tag}", allow_missing=True)
            if occupied is None:
                raise
            if tag_commit(occupied) != commit:
                continue
        return tag


def preview_is_published(tag):
    release = api(f"releases/tags/{tag}", allow_missing=True)
    if not release or release["draft"]:
        return False
    if not release["prerelease"] or release["name"] != tag:
        raise RuntimeError(f"An incompatible release already exists for {tag}.")
    return True


if __name__ == "__main__":
    tag = reserve_preview_tag(os.environ["GITHUB_SHA"])
    published = str(preview_is_published(tag)).lower()
    with Path(os.environ["GITHUB_OUTPUT"]).open("a") as output:
        print(f"tag={tag}", file=output)
        print(f"published={published}", file=output)
