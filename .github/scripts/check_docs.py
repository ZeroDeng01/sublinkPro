"""Validate repository documentation without fetching external URLs."""

from collections import defaultdict
from pathlib import Path, PurePosixPath
import posixpath
import re
import subprocess
import unicodedata
from urllib.parse import unquote, urlsplit


ROOT = Path(__file__).resolve().parents[2]
SKILL = "skills/sublinkpro/"
REPOSITORY = "zerodeng01/sublinkpro"
INLINE_LINK = re.compile(r"\[[^\]\n]+\]\(\s*(?:<([^>\n]+)>|([^\s)]+))")
REFERENCE = re.compile(r"^\s{0,3}\[([^]]+)\]:\s*(?:<([^>]+)>|(\S+))")
REFERENCE_USE = re.compile(r"\[([^]\n]+)\]\[([^]\n]*)\]")
HTML_LINK = re.compile(r"""<(?:a|img)\b[^>]*?\b(?:href|src)=["']([^"']+)["']""", re.I)
HTML_ANCHOR = re.compile(r"""<(?:a|h[1-6])\b[^>]*?\b(?:id|name)=["']([^"']+)["']""", re.I)
HEADING = re.compile(r"^\s{0,3}#{1,6}\s+(.+?)(?:\s+#+)?\s*$")
FENCE = re.compile(r"^\s*(" + chr(96) + r"{3,}|~{3,})(.*)$")
LEGACY = re.compile(
    r"(?<![\w])(?:docs/(?:features/|(?:installation|configuration|security-guidelines|"
    r"development(?:-workflow)?|build-and-deployment|practical-recipes|frontend-theme-guidelines|"
    r"internationalization|script_support)(?:\.zh-CN)?\.md)|skill-sublinkpro/)"
)
ROOT_PAIRS = {"README", "CONTRIBUTING", "CODE_OF_CONDUCT", "SECURITY", SKILL + "README"}


def paths_in_checkout():
    names = subprocess.check_output(
        ["git", "ls-files", "-z", "--cached", "--others", "--exclude-standard"], cwd=ROOT
    ).decode().split("\0")
    return {name for name in names if name and (ROOT / name).is_file()}


def document_lines(contents):
    """Yield content lines outside fenced blocks, including quoted fences."""
    fence = ""
    for number, line in enumerate(contents.splitlines(), 1):
        line = re.sub(r"^\s*(?:>\s*)+", "", line)
        marker = FENCE.match(line)
        if fence:
            if marker and marker[1].startswith(fence) and not marker[2].strip():
                fence = ""
            continue
        if marker:
            fence = marker[1]
        else:
            yield number, line


def anchors_in(lines):
    anchors = set()
    counts = defaultdict(int)
    previous = ""
    for _, line in lines:
        heading = HEADING.match(line)
        title = heading[1] if heading else previous if re.fullmatch(r"\s{0,3}(?:=+|-+)\s*", line) else ""
        if title:
            title = re.sub(r"\[([^]]+)\]\([^)]+\)", r"\1", title)
            title = re.sub(r"<[^>]+>", "", title).lower()
            slug = "".join(c for c in title if c.isalnum() or c in "_- " or is_combining(c))
            slug = slug.replace(" ", "-")
            index = counts[slug]
            anchor = f"{slug}-{index}" if index else slug
            while anchor in anchors:
                index += 1
                anchor = f"{slug}-{index}"
            anchors.add(anchor)
            counts[slug] = index + 1
        anchors.update(HTML_ANCHOR.findall(line))
        previous = line if line.strip() and not heading else ""
    return anchors


def is_combining(char):
    return unicodedata.category(char).startswith("M")


def links_in(lines):
    references = {}
    uses = []
    for number, line in lines:
        line = re.sub(r"(`+).*?\1", "code", line)
        if definition := REFERENCE.match(line):
            references[definition[1].casefold()] = (number, definition[2] or definition[3])
        for match in INLINE_LINK.finditer(line):
            yield number, match[1] or match[2]
        for href in HTML_LINK.findall(line):
            yield number, href
        for text, label in REFERENCE_USE.findall(line):
            uses.append((label or text).casefold())
    for label in uses:
        if label in references:
            yield references[label]


def link_target(source, url):
    link = urlsplit(url)
    if not link.scheme and not link.netloc:
        path = unquote(link.path)
        result = posixpath.normpath(posixpath.join(posixpath.dirname(source), path)) if path else source
        return result, unquote(link.fragment), False
    path = unquote(link.path).lstrip("/")
    parts = path.split("/")
    if "/".join(parts[:2]).lower() != REPOSITORY:
        return None, "", False
    if link.hostname == "github.com" and len(parts) >= 4 and parts[2] in {"blob", "tree", "raw"}:
        if parts[3] in {"main", "dev", "master"}:
            return "/".join(parts[4:]), unquote(link.fragment), True
    if link.hostname == "raw.githubusercontent.com":
        branch = parts[4] if parts[2:4] == ["refs", "heads"] and len(parts) > 4 else parts[2] if len(parts) > 2 else ""
        if branch in {"main", "dev", "master"}:
            start = 5 if parts[2:4] == ["refs", "heads"] else 3
            return "/".join(parts[start:]), unquote(link.fragment), True
    return None, "", False


def language_peer(filename):
    specific = filename.removesuffix(".md").removesuffix(".zh-CN")
    if specific in ROOT_PAIRS or filename.startswith("docs/"):
        return specific + (".md" if filename.endswith(".zh-CN.md") else ".zh-CN.md")
    return None


def check():
    files = paths_in_checkout()
    docs = sorted(file for file in files if file.endswith(".md"))
    directories = {str(parent) for file in files for parent in PurePosixPath(file).parents}
    errors = set()
    texts = {file: (ROOT / file).read_text(encoding="utf-8") for file in docs}
    lines = {file: list(document_lines(text)) for file, text in texts.items()}
    anchors = {file: anchors_in(doc) for file, doc in lines.items()}

    for file in docs:
        checked_local_links = set()
        for number, url in links_in(lines[file]):
            dest, fragment, remote = link_target(file, url)
            if dest is None or "<" in dest:
                continue
            dest = dest or "."
            if not remote:
                checked_local_links.add(dest)
            if file.startswith(SKILL) and not remote and not dest.startswith(SKILL):
                errors.add((file, number, f"Relative link escapes the skill package: {url}"))
            if dest not in files and dest not in directories:
                errors.add((file, number, f"Link target is missing: {url}"))
            elif fragment and dest in anchors and fragment not in anchors[dest]:
                errors.add((file, number, f"Heading or anchor is missing: {url}"))

        if peer := language_peer(file):
            if peer not in texts:
                errors.add((file, 1, f"Missing language version: {peer}"))
            elif peer not in checked_local_links:
                errors.add((file, 1, f"Missing link to language version: {peer}"))

    for skill in sorted(file for file in docs if file.endswith("/SKILL.md")):
        header = texts[skill].split("---", 2)
        folder = PurePosixPath(skill).parent.name
        if (
            len(header) < 3
            or not re.search(rf"(?m)^name:\s*{re.escape(folder)}\s*$", header[1])
            or not re.search(r"(?m)^description:\s*\S", header[1])
        ):
            errors.add((skill, 1, f"SKILL.md must declare name: {folder} and a description"))
        if skill.startswith(SKILL):
            # Packaged documentation references are also used as literal paths by agents.
            base = PurePosixPath(SKILL)
            for number, line in lines[skill]:
                for ref in re.findall(r"(?:reference|scripts)/[a-z][\w.-]*\.(?:md|py)", line):
                    if str(base / ref) not in files:
                        errors.add((skill, number, f"Missing bundled reference: {ref}"))

    sources = files & {
        "main.go", "api/skill.go", ".github/workflows/build-release.yml",
    }
    sources.update(
        file for file in files
        if file.startswith("webs/src/") and Path(file).suffix in {".js", ".jsx", ".ts", ".tsx"}
    )
    for file in [*docs, *sorted(sources)]:
        text = texts[file] if file in texts else (ROOT / file).read_text(encoding="utf-8")
        for number, line in enumerate(text.splitlines(), 1):
            if match := LEGACY.search(line):
                errors.add((file, number, f"Retired document path: {match[0]}"))
            if "github.com/ZeroDeng01/sublinkPro/blob/master/" in line:
                errors.add((file, number, "Use the main branch in project links"))
            if file in sources:
                for url in re.findall(r"""https://[^"' )\t]+""", line):
                    dest, fragment, _ = link_target(file, url)
                    if dest and (dest not in files and dest not in directories):
                        errors.add((file, number, f"Link target is missing: {url}"))
                    elif fragment and dest in anchors and fragment not in anchors[dest]:
                        errors.add((file, number, f"Heading or anchor is missing: {url}"))

    for file, number, message in sorted(errors):
        print(f"{file}:{number}: {message}")
    if errors:
        print(f"{len(errors)} documentation errors")
    else:
        print(f"Checked {len(docs)} Markdown files")
    return bool(errors)


if __name__ == "__main__":
    raise SystemExit(check())
