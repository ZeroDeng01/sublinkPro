import unittest

import check_docs


class DocumentationChecks(unittest.TestCase):
    def test_fences_and_inline_examples_do_not_produce_links(self):
        text = "`[example](filename.md)`\n> ```bash\n> [example](filename.md)\n> ```\n[actual](guide.md)"
        self.assertEqual(list(check_docs.links_in(check_docs.document_lines(text))), [(5, "guide.md")])

    def test_markdown_anchors_and_duplicates(self):
        text = "# 📚 代理网关\n# 📚 代理网关\n## A + B\n<a id=\"manual\"></a>"
        self.assertEqual(check_docs.anchors_in(check_docs.document_lines(text)), {"-代理网关", "-代理网关-1", "a--b", "manual"})

    def test_github_project_links_and_checkout_links(self):
        self.assertEqual(check_docs.link_target("docs/user/features/foo.md", "../../images/foo.png"),
                         ("docs/images/foo.png", "", False))
        url = "https://raw.githubusercontent.com/ZeroDeng01/sublinkPro/refs/heads/main/docs/README.md"
        self.assertEqual(check_docs.link_target("README.md", url), ("docs/README.md", "", True))
        self.assertEqual(check_docs.link_target("README.md", "https://example.com/docs/README.md"), (None, "", False))

    def test_language_pair_is_required_for_public_docs(self):
        self.assertEqual(check_docs.language_peer("docs/user/README.zh-CN.md"), "docs/user/README.md")
        self.assertIsNone(check_docs.language_peer("skills/sublinkpro/reference/api.md"))
