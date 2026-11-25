import json
import os
from pathlib import Path
import subprocess
import unittest


INSTALLER = Path(__file__).resolve().parents[2] / "install.sh"
INSTALL_FUNCTIONS = INSTALLER.read_text(encoding="utf-8").rsplit('\nmain "$@"', 1)[0]
ASSET = "sublinkPro-linux-arm64"
PREVIEW_STUBS = r'''
check_root() { :; }
detect_alpine() { is_alpine=true; }
uname() { printf '%s\n' aarch64; }
show_not_installed_menu() { printf '%s\n' "$latest_release"; }
jq() { return 127; }
python3() { return 127; }
apk() { return 127; }
curl() {
    for last_argument do :; done
    case "$last_argument" in
        "$RELEASES_API/latest") printf '{\n  "tag_name": "v1.3.0"\n}\n' ;;
        "$RELEASES_API?per_page=100&page=1") printf '%s\n' "$TEST_PAGE_ONE" ;;
        "$RELEASES_API?per_page=100&page=2") printf '%s\n' "$TEST_PAGE_TWO" ;;
        *) return 1 ;;
    esac
}
'''


def release(tag, day, *, prerelease=True, draft=False, name=ASSET, state="uploaded"):
    return {
        "tag_name": tag,
        "published_at": f"2026-11-{day:02d}T12:00:00Z",
        "prerelease": prerelease,
        "draft": draft,
        "assets": [{"name": name, "state": state}],
    }


def run_installer(script, *args, **env):
    return subprocess.run(
        ["sh", "-c", INSTALL_FUNCTIONS + "\nBINARY_PATH=/dev/null\n" + script, "install.sh", *args],
        env=os.environ | env,
        text=True,
        capture_output=True,
        timeout=10,
    )


def select_release(*args, first_page=(), second_page=(), indent=None):
    return run_installer(
        PREVIEW_STUBS + '\nmain "$@"',
        *args,
        TEST_PAGE_ONE=json.dumps(first_page, indent=indent),
        TEST_PAGE_TWO=json.dumps(second_page, indent=indent),
    )


class InstallPreviewTests(unittest.TestCase):
    def test_default_selects_stable_release(self):
        result = select_release(first_page=[release("pre-261124171211", 24)])
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(result.stdout.strip(), "v1.3.0")

    def test_selects_newest_published_preview_with_uploaded_asset_across_pages(self):
        stable = release("v1.3.0", 24, prerelease=False)
        expected = "pre-261123171211"
        first_page = [release("pre-261122171211", 22)] + [stable] * 99
        second_page = [
            release("pre-261128171211", 28, draft=True),
            release("v1.4.0-rc.1", 27),
            release("pre-261126171211", 26, name="sublinkPro-linux-amd64"),
            release("pre-261125171211", 25, state="new"),
            release("pre-261124171211", 24, prerelease=False),
            release(expected, 23),
        ]
        for indent in [None, 2]:
            with self.subTest(indent=indent):
                result = select_release("--preview", first_page=first_page, second_page=second_page, indent=indent)
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertEqual(result.stdout.strip(), expected)

    def test_ignores_release_notes_and_nested_metadata_when_selecting_assets(self):
        expected = "pre-261122171211"
        fake_asset = {"name": ASSET, "state": "uploaded"}
        metadata = release("pre-261123171211", 23, name="sublinkPro-linux-amd64")
        metadata["body"] = '预发布\\notes\n"assets": [' + json.dumps(fake_asset) + ']'
        metadata["assets"][0]["uploader"] = fake_asset
        result = select_release("--preview", first_page=[metadata, release(expected, 22)])
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(result.stdout.strip(), expected)

    def test_invalid_release_response_stops_before_selection(self):
        result = run_installer(
            PREVIEW_STUBS + '\nmain "$@"',
            "--preview",
            TEST_PAGE_ONE='[{"tag_name": "pre-261124171211",}]',
        )
        self.assertEqual(result.returncode, 1)
        self.assertIn("无法解析 GitHub 发布列表", result.stdout)

    def test_missing_compatible_preview_exits_without_installing_stable(self):
        result = select_release("--preview", first_page=[release("v1.3.0", 24, prerelease=False)])
        self.assertEqual(result.returncode, 1)
        self.assertIn(ASSET, result.stdout)
        self.assertNotIn("v1.3.0", result.stdout)

    def test_unknown_argument_exits_with_usage_hint(self):
        result = select_release("--unknown")
        self.assertEqual(result.returncode, 1)
        self.assertIn("--help", result.stdout)

    def test_failed_download_preserves_running_service(self):
        result = run_installer(r'''
latest_release=pre-261124171211
file_name=sublinkPro-linux-arm64
mkdir() { :; }
mktemp() { printf '%s\n' /dev/null; }
rm() { :; }
curl() { return 22; }
stop_service() { printf '%s\n' STOPPED; }
update_program
''')
        self.assertEqual(result.returncode, 1)
        self.assertIn("无法准备", result.stdout)
        self.assertNotIn("STOPPED", result.stdout)
