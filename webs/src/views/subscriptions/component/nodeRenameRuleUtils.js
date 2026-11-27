import { formatServiceMarks, previewServiceMarks, SERVICE_MARKS_SEPARATOR, splitServiceMarks } from './serviceMarks.js';

const QUICK_SEPARATOR_CHARS = new Set(['-', '_', '|', ' ', '[', ']', '(', ')']);

const sampleValues = {
  $Protocol: 'VMess',
  $LinkCountry: 'HK',
  $LinkCountryName: '香港',
  $Flag: '🇭🇰',
  $Name: 'HK-Node-Remark',
  $LinkName: 'HK-01',
  $SpeedIcon: '🟢',
  $DelayIcon: '🟡',
  $IpType: 'Native',
  $Residential: 'Residential',
  $FraudScoreIcon: '🟢',
  $FraudScore: '15',
  $Unlock: 'Netflix-US',
  $UnlockStatus: 'available',
  $UnlockLabel: '解锁',
  $UnlockRegion: 'US',
  $Speed: '12.5MB/s',
  $Delay: '125ms',
  $Group: 'VIP',
  $Source: 'ProviderA',
  $Index: '01',
  $DuplicateIndex: '1',
  $Tags: 'Fast|HK'
};

const splitSeparatorText = (text) => {
  const tokens = [];
  let buffer = '';

  for (const character of text) {
    if (QUICK_SEPARATOR_CHARS.has(character)) {
      if (buffer) {
        tokens.push(buffer);
        buffer = '';
      }
      tokens.push(character);
    } else {
      buffer += character;
    }
  }

  if (buffer) tokens.push(buffer);
  return tokens;
};

export const parseNodeRenameRule = (rule) => {
  if (!rule) return [];

  const items = [];
  let id = 0;
  const pushSeparators = (text) => {
    for (const token of splitSeparatorText(text)) {
      items.push({ id: `sep-${id++}`, type: 'separator', value: token });
    }
  };
  const variablePattern =
    /\$(Name|LinkName|LinkCountryName|LinkCountry|Flag|SpeedIcon|Speed|DelayIcon|Delay|IpType|Residential|FraudScoreIcon|FraudScore|Unlock\([^)]+\)|UnlockStatus|UnlockLabel|UnlockRegion|Unlock|Group|Source|DuplicateIndex|Index|Protocol|Tags|TagGroup\([^)]+\))/g;
  for (const part of splitServiceMarks(rule)) {
    if (part.marks) {
      items.push({ id: `var-${id++}`, type: 'serviceMarks', ...part });
      continue;
    }
    let lastIndex = 0;
    for (const match of part.value.matchAll(variablePattern)) {
      pushSeparators(part.value.slice(lastIndex, match.index));
      items.push({ id: `var-${id++}`, type: 'variable', value: match[0] });
      lastIndex = match.index + match[0].length;
    }
    pushSeparators(part.value.slice(lastIndex));
  }

  return items;
};

const defaultMarkAffixes = (items, index) => ({
  prefix: items[index - 1] && items[index - 1].type !== 'separator' ? SERVICE_MARKS_SEPARATOR : '',
  suffix: items[index + 1]?.type === 'variable' ? SERVICE_MARKS_SEPARATOR : ''
});

export const updateServiceMarkAffixes = (previous, items) => {
  const previousPositions = new Map(previous.map((item, index) => [item.id, index]));
  return items.map((item, index) => {
    if (item.type !== 'serviceMarks') return item;
    const oldIndex = previousPositions.get(item.id);
    const old = oldIndex === undefined ? { prefix: SERVICE_MARKS_SEPARATOR, suffix: '' } : defaultMarkAffixes(previous, oldIndex);
    if ((item.marks.prefix ?? '') !== old.prefix || (item.marks.suffix ?? '') !== old.suffix) return item;
    const next = defaultMarkAffixes(items, index);
    if (next.prefix === old.prefix && next.suffix === old.suffix) return item;
    const marks = { ...item.marks, ...next };
    return { ...item, marks, value: formatServiceMarks(marks) };
  });
};

export const previewNodeRenameItems = (items) => {
  const preview = items
    .map((item) => {
      if (item.type === 'serviceMarks') return previewServiceMarks(item.marks);
      if (item.type !== 'variable') return item.value;
      if (item.value.startsWith('$TagGroup(')) return 'Fast';
      if (item.value.startsWith('$Unlock(')) return 'Unlock-US';
      return sampleValues[item.value] ?? item.value;
    })
    .join('');
  return items.some((item) => item.type === 'serviceMarks') ? preview.trim() || sampleValues.$Name : preview;
};

export const previewSubscriptionNodeName = (rule) => previewNodeRenameItems(parseNodeRenameRule(rule));
