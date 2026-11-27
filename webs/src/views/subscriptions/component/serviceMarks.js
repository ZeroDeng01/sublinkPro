const start = '$ServiceMarks(';
export const SERVICE_MARKS_SEPARATOR = ' · ';
const invalidText = /[,=\p{Cc}\u2028\u2029]/u;
const defaultLabels = { youtube_premium: 'YT', openai: 'GPT', netflix: 'NF' };

export const createServiceMark = (provider) => ({
  provider: provider.value,
  label: defaultLabels[provider.value] || provider.label || provider.value,
  acceptedStatuses: ['available'],
  region: false
});

export const createServiceMarks = () => ({ prefix: SERVICE_MARKS_SEPARATOR, separator: SERVICE_MARKS_SEPARATOR, suffix: '', services: [] });

export const editServiceMarks = (saved) => ({
  prefix: typeof saved.prefix === 'string' ? saved.prefix : '',
  separator: typeof saved.separator === 'string' ? saved.separator : SERVICE_MARKS_SEPARATOR,
  suffix: typeof saved.suffix === 'string' ? saved.suffix : '',
  services: Array.isArray(saved.services)
    ? saved.services
        .filter((mark) => typeof mark?.provider === 'string')
        .map((mark) => ({
          provider: normalizeProvider(mark.provider),
          label: typeof mark.label === 'string' ? mark.label : '',
          acceptedStatuses: Array.isArray(mark.acceptedStatuses)
            ? mark.acceptedStatuses.filter((status) => typeof status === 'string').map((status) => status.trim().toLowerCase())
            : [],
          region: mark.region === true
        }))
    : []
});

export const formatServiceMarks = (marks) => `${start}${JSON.stringify(marks)})`;

const readServiceMarks = (rule, from) => {
  let begin = from + start.length;
  while (/[ \t\r\n]/.test(rule[begin] || '')) begin++;
  if (rule[begin] !== '{') return null;

  let depth = 0;
  let quoted = false;
  for (let index = begin; index < rule.length; index++) {
    const character = rule[index];
    if (quoted && character === '\\') {
      index++;
      continue;
    }
    if (character === '"') quoted = !quoted;
    if (quoted) continue;
    if (character === '{') depth++;
    if (character !== '}') continue;
    depth--;
    if (depth !== 0) continue;

    let end = index + 1;
    while (/[ \t\r\n]/.test(rule[end] || '')) end++;
    if (rule[end] !== ')') return null;
    try {
      return { end: end + 1, marks: JSON.parse(rule.slice(begin, index + 1)) };
    } catch {
      return null;
    }
  }
  return null;
};

export const splitServiceMarks = (rule) => {
  const parts = [];
  // Consume legacy calls whole so a provider or tag group name stays intact.
  const callStart = /\$(?:TagGroup|Unlock)\([^)]+\)|\$ServiceMarks\([ \t\r\n]*\{/g;
  let offset = 0;
  let match;
  while ((match = callStart.exec(rule)) !== null) {
    if (!match[0].startsWith(start)) continue;
    const index = match.index;
    const block = readServiceMarks(rule, index);
    if (!block) continue;
    if (offset < index) parts.push({ value: rule.slice(offset, index) });
    parts.push({ value: rule.slice(index, block.end), marks: block.marks });
    offset = block.end;
    callStart.lastIndex = offset;
  }
  if (offset < rule.length) parts.push({ value: rule.slice(offset) });
  return parts;
};

const normalizeProvider = (provider) =>
  provider
    .trim()
    .toLowerCase()
    .replaceAll(' ', '_')
    .replaceAll('-', '_')
    .replace(/^_+|_+$/g, '');

export const serviceMarksError = (marks, statuses) => {
  const text = [marks.prefix, marks.separator, marks.suffix];
  if (text.some((value) => value != null && (typeof value !== 'string' || invalidText.test(value)))) return { reason: 'text' };
  if (!Array.isArray(marks.services) || marks.services.length === 0) return { reason: 'services' };

  const seen = new Set();
  for (const mark of marks.services) {
    const provider = typeof mark?.provider === 'string' ? normalizeProvider(mark.provider) : '';
    if (!provider) return { reason: 'provider' };
    if (seen.has(provider)) return { reason: 'duplicate', provider };
    seen.add(provider);
    if (typeof mark.label !== 'string') return { reason: 'label', provider };
    if (invalidText.test(mark.label)) return { reason: 'text', provider };
    if (!mark.label.trim()) return { reason: 'label', provider };
    if (
      !Array.isArray(mark.acceptedStatuses) ||
      mark.acceptedStatuses.length === 0 ||
      mark.acceptedStatuses.some((status) => typeof status !== 'string' || !statuses.includes(status.trim().toLowerCase()))
    ) {
      return { reason: 'statuses', provider };
    }
  }
  return null;
};

export const renderServiceMarks = (marks, results) => {
  if (!Array.isArray(marks.services)) return '';
  const labels = [];
  for (const mark of marks.services) {
    if (typeof mark?.provider !== 'string' || typeof mark.label !== 'string' || !Array.isArray(mark.acceptedStatuses)) continue;
    const result = results.get(normalizeProvider(mark.provider));
    const label = mark.label.trim();
    if (
      !label ||
      !result ||
      !mark.acceptedStatuses.some((status) => typeof status === 'string' && status.trim().toLowerCase() === result.status)
    )
      continue;
    const region = result.region?.trim();
    labels.push(mark.region && region && !invalidText.test(region) ? `${label}-${region}` : label);
  }
  return labels.length ? `${marks.prefix ?? ''}${labels.join(marks.separator ?? SERVICE_MARKS_SEPARATOR)}${marks.suffix ?? ''}` : '';
};

export const previewServiceMarks = (marks) => {
  const results = new Map();
  if (Array.isArray(marks.services)) {
    for (const mark of marks.services) {
      const status = mark?.acceptedStatuses?.[0];
      if (typeof mark?.provider === 'string' && typeof status === 'string') {
        results.set(normalizeProvider(mark.provider), { status: status.trim().toLowerCase(), region: 'US' });
      }
    }
  }
  return renderServiceMarks(marks, results);
};
