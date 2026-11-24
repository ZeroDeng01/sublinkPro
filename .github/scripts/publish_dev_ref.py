"""Record the revision of the rolling dev image after Docker Hub accepts it."""

import os
import re

from prepare_preview import api


DEV_REF = "tags/dev-image"


def record_dev_image(sha):
    if not re.fullmatch(r"[0-9a-f]{40}", sha):
        raise ValueError("Expected the full commit SHA of the published image.")

    path = f"git/ref/{DEV_REF}"
    ref = api(path, allow_missing=True)
    if not ref:
        api("git/refs", data={"ref": f"refs/{DEV_REF}", "sha": sha})
    elif ref["object"]["sha"] != sha:
        api(f"git/refs/{DEV_REF}", method="PATCH", data={"sha": sha, "force": True})


if __name__ == "__main__":
    record_dev_image(os.environ["GITHUB_SHA"])
