import assert from 'node:assert/strict';
import test from 'node:test';
import { GITHUB_RELEASES_API, fetchReleaseData, getStableChangelog, getVersionSource, getVersionUpdate } from './githubReleases.js';

const release = (tag_name, day, extra = {}) => ({
  tag_name,
  published_at: `2026-11-${String(day).padStart(2, '0')}T12:00:00Z`,
  draft: false,
  prerelease: false,
  ...extra
});

test('stable and preview builds select their release sources', () => {
  assert.equal(getVersionSource('v1.3.0'), `${GITHUB_RELEASES_API}/latest`);
  assert.equal(getVersionSource('preview-20261124-4a211cb'), GITHUB_RELEASES_API);
  assert.equal(getVersionSource('preview-20261124-171211123'), GITHUB_RELEASES_API);
  assert.equal(getVersionSource('4a211cb'), null);
});

test('stable compares numeric versions and ignores prereleases', () => {
  const latest = release('v1.10.0', 24);
  assert.deepEqual(getVersionUpdate('1.9.9', latest), { latestVersion: 'v1.10.0', hasUpdate: true });
  assert.equal(getVersionUpdate('v1.10.0', latest).hasUpdate, false);
  assert.equal(getVersionUpdate('v1.11.0', latest).hasUpdate, false);
  assert.equal(getVersionUpdate('v1.9.9', { ...latest, prerelease: true }), null);
});

test('preview finds a timestamped update for an older hash-based build on the same day', () => {
  const current = 'preview-20261124-0000000';
  const latest = 'preview-20261124-171211123';
  const releases = [
    release('v1.3.0', 25),
    release(latest, 24, { published_at: '2026-11-24T18:00:00Z', prerelease: true }),
    release(current, 24, { prerelease: true })
  ];
  assert.deepEqual(getVersionUpdate(current, releases), { latestVersion: latest, hasUpdate: true });
  assert.equal(getVersionUpdate(latest, releases).hasUpdate, false);
  assert.equal(getVersionUpdate('preview-20261125-aaaaaaa', releases).hasUpdate, false);
  assert.equal(getVersionUpdate('preview-20261124-aaaaaaa', releases), null);
});

test('development builds have no automatic update source', () => {
  assert.equal(getVersionSource('dev'), null);
  assert.equal(getVersionSource('dev-4a211cb00000'), null);
  assert.equal(getVersionUpdate('dev-4a211cb00000', release('v1.4.0', 24)), null);
});

test('stable changelog fills ten slots after pagination and sorting by publication', async () => {
  const previews = Array.from({ length: 99 }, (_, index) =>
    release(`preview-20261124-${String(index).padStart(7, '0')}`, 24, { prerelease: true })
  );
  const stable = Array.from({ length: 11 }, (_, index) => release(`v1.3.${index}`, index + 1));
  const draft = release('v2.0.0', 26, { draft: true });
  const pages = [
    [...previews, stable[0]],
    [draft, ...stable.slice(1)]
  ];
  const urls = [];
  const releases = await fetchReleaseData(GITHUB_RELEASES_API, async (url) => {
    urls.push(url);
    return { ok: true, json: async () => pages.shift() };
  });
  assert.deepEqual(urls, [`${GITHUB_RELEASES_API}?per_page=100&page=1`, `${GITHUB_RELEASES_API}?per_page=100&page=2`]);
  assert.deepEqual(
    getStableChangelog(releases).map(({ tag_name }) => tag_name),
    stable
      .slice(1)
      .reverse()
      .map(({ tag_name }) => tag_name)
  );
});
