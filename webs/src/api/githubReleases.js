const GITHUB_REPO = 'ZeroDeng01/sublinkPro';
export const GITHUB_URL = `https://github.com/${GITHUB_REPO}`;
export const GITHUB_RELEASES_URL = `${GITHUB_URL}/releases`;
export const DEV_IMAGE_URL = 'https://hub.docker.com/r/zerodeng/sublink-pro/tags?name=dev';

const GITHUB_API = `https://api.github.com/repos/${GITHUB_REPO}`;
export const GITHUB_RELEASES_API = `${GITHUB_API}/releases`;
const DEV_REF_API = `${GITHUB_API}/git/ref/tags/dev-image`;
const STABLE_RELEASE_API = `${GITHUB_RELEASES_API}/latest`;

const STABLE_VERSION = /^v?([0-9]+)\.([0-9]+)\.([0-9]+)$/;
const PREVIEW_VERSION = /^preview-([0-9]{8})-[a-f0-9]{7,40}$/;
const DEV_VERSION = /^dev-([a-f0-9]{7,40})$/;
const RELEASES_PER_PAGE = 100;

export const RELEASE_REQUEST_OPTIONS = { revalidateOnFocus: false, shouldRetryOnError: false, dedupingInterval: 60 * 1000 };

export function getVersionChannel(version) {
  if (version === 'dev' || DEV_VERSION.test(version)) return 'dev';
  if (PREVIEW_VERSION.test(version)) return 'preview';
  if (STABLE_VERSION.test(version)) return 'stable';
  return null;
}

export function getVersionSource(version) {
  switch (getVersionChannel(version)) {
    case 'dev':
      return version === 'dev' ? null : DEV_REF_API;
    case 'preview':
      return GITHUB_RELEASES_API;
    case 'stable':
      return STABLE_RELEASE_API;
    default:
      return null;
  }
}

async function fetchJson(url, request) {
  const response = await request(url);
  if (!response.ok) throw new Error(`GitHub API returned ${response.status}`);
  return response.json();
}

export async function fetchReleaseData(url, request = fetch) {
  if (url !== GITHUB_RELEASES_API) return fetchJson(url, request);

  const releases = [];
  for (let page = 1; ; page += 1) {
    const data = await fetchJson(`${url}?per_page=${RELEASES_PER_PAGE}&page=${page}`, request);
    if (!Array.isArray(data)) throw new Error('GitHub did not return a release list');
    releases.push(...data.filter((release) => release?.draft === false && Number.isFinite(Date.parse(release.published_at))));
    if (data.length < RELEASES_PER_PAGE) break;
  }
  return releases.sort((first, second) => {
    const publishedLater = Date.parse(second.published_at) - Date.parse(first.published_at);
    return publishedLater || (first.tag_name < second.tag_name ? 1 : first.tag_name > second.tag_name ? -1 : 0);
  });
}

function isStableRelease(release) {
  return (
    release?.draft === false &&
    release.prerelease === false &&
    typeof release.tag_name === 'string' &&
    !PREVIEW_VERSION.test(release.tag_name)
  );
}

export function getStableChangelog(releases) {
  return releases.filter(isStableRelease).slice(0, 10);
}

export function getVersionUpdate(version, data) {
  if (!data) return null;

  switch (getVersionChannel(version)) {
    case 'stable': {
      if (!isStableRelease(data) || !STABLE_VERSION.test(data.tag_name)) return null;
      const current = STABLE_VERSION.exec(version).slice(1).map(Number);
      const latest = STABLE_VERSION.exec(data.tag_name).slice(1).map(Number);
      const difference = latest.findIndex((part, index) => part !== current[index]);
      return { latestVersion: data.tag_name, hasUpdate: difference !== -1 && latest[difference] > current[difference] };
    }
    case 'preview': {
      const latest = data.find((release) => release.prerelease === true && PREVIEW_VERSION.test(release.tag_name));
      if (!latest) return null;
      if (latest.tag_name === version) return { latestVersion: version, hasUpdate: false };

      const current = data.find((release) => release.prerelease === true && release.tag_name === version);
      const latestDate = PREVIEW_VERSION.exec(latest.tag_name)[1];
      const currentDate = PREVIEW_VERSION.exec(version)[1];
      if (!current && latestDate === currentDate) return null;
      return { latestVersion: latest.tag_name, hasUpdate: Boolean(current) || latestDate > currentDate };
    }
    case 'dev': {
      const current = DEV_VERSION.exec(version)?.[1];
      const sha = data.object?.sha;
      if (!current || data.object?.type !== 'commit' || !/^[a-f0-9]{40}$/.test(sha)) return null;
      return { latestVersion: `dev-${sha.slice(0, 12)}`, hasUpdate: !sha.startsWith(current) };
    }
    default:
      return null;
  }
}
