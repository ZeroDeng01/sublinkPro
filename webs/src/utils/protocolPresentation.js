const DEFAULT_PROTOCOL_COLOR = '#757575';
const UNKNOWN_PROTOCOL_COLOR = '#9e9e9e';
const OTHER_PROTOCOL_COLOR = '#616161';
const DEFAULT_PROTOCOL_GRADIENT = 'linear-gradient(135deg, #5046e5 0%, #7c3aed 100%)';

const FALLBACK_PROTOCOL_META = [
  { name: 'vmess', label: 'VMess', color: '#1976d2', icon: 'V', prefixes: ['vmess://'] },
  { name: 'vless', label: 'VLESS', color: '#7b1fa2', icon: 'V', prefixes: ['vless://'] },
  { name: 'trojan', label: 'Trojan', color: '#d32f2f', icon: 'T', prefixes: ['trojan://'] },
  { name: 'ss', label: 'SS', color: '#2e7d32', icon: 'S', prefixes: ['ss://'] },
  { name: 'ssr', label: 'SSR', color: '#e64a19', icon: 'R', prefixes: ['ssr://'] },
  { name: 'hysteria', label: 'Hysteria', color: '#f9a825', icon: 'H', prefixes: ['hysteria://', 'hy://'] },
  { name: 'hysteria2', label: 'Hysteria2', color: '#ef6c00', icon: 'H', prefixes: ['hysteria2://', 'hy2://'] },
  { name: 'tuic', label: 'TUIC', color: '#0277bd', icon: 'T', prefixes: ['tuic://'] },
  { name: 'wireguard', label: 'WireGuard', color: '#88171a', icon: 'W', prefixes: ['wg://', 'wireguard://'] },
  { name: 'naiveproxy', label: 'NaiveProxy', color: '#5d4037', icon: 'N', prefixes: ['naiveproxy://', 'naive://'] },
  { name: 'anytls', label: 'AnyTLS', color: '#20a84c', icon: 'A', prefixes: ['anytls://'] },
  { name: 'socks5', label: 'SOCKS5', color: '#116ea4', icon: 'S', prefixes: ['socks5://'] },
  { name: 'socks', label: 'SOCKS', color: '#dd4984', icon: 'S', prefixes: ['socks://'] },
  { name: 'http', label: 'HTTP', color: '#0288d1', icon: 'H', prefixes: ['http://'] },
  { name: 'https', label: 'HTTPS', color: '#0277bd', icon: 'H', prefixes: ['https://'] },
  { name: 'reality', label: 'Reality', color: '#c2185b', icon: 'R', prefixes: ['reality://'] }
];

const normalizeProtocolName = (protocolName) => {
  if (!protocolName || typeof protocolName !== 'string') {
    return '';
  }

  return protocolName.trim().toLowerCase();
};

const normalizePrefix = (prefix) => {
  if (!prefix || typeof prefix !== 'string') {
    return '';
  }

  return prefix.trim().toLowerCase();
};

const uniqueValues = (values) => [...new Set(values.filter(Boolean))];

const getMetaPrefixes = (meta = {}) => {
  const explicitPrefixes = [];

  if (Array.isArray(meta.prefixes)) {
    explicitPrefixes.push(...meta.prefixes);
  }

  if (Array.isArray(meta.aliases)) {
    explicitPrefixes.push(...meta.aliases.map((alias) => (alias?.includes('://') ? alias : `${alias}://`)));
  }

  const normalizedName = normalizeProtocolName(meta.name);
  if (normalizedName) {
    explicitPrefixes.push(`${normalizedName}://`);
  }

  return uniqueValues(explicitPrefixes.map(normalizePrefix));
};

const buildProtocolEntry = (meta = {}) => {
  const normalizedName = normalizeProtocolName(meta.name);
  const label = meta.label || (normalizedName ? normalizedName.toUpperCase() : '');
  const icon = typeof meta.icon === 'string' && meta.icon ? meta.icon.charAt(0).toUpperCase() : label.charAt(0).toUpperCase();

  return {
    value: normalizedName,
    label,
    color: meta.color || DEFAULT_PROTOCOL_COLOR,
    icon,
    prefixes: getMetaPrefixes(meta)
  };
};

const buildProtocolEntries = (protocolMeta = []) => {
  const mergedMeta = [...FALLBACK_PROTOCOL_META];

  if (Array.isArray(protocolMeta)) {
    protocolMeta.forEach((meta) => {
      const normalizedName = normalizeProtocolName(meta?.name);
      const fallbackIndex = mergedMeta.findIndex((item) => normalizeProtocolName(item.name) === normalizedName);

      if (fallbackIndex >= 0) {
        mergedMeta[fallbackIndex] = { ...mergedMeta[fallbackIndex], ...meta };
      } else if (normalizedName) {
        mergedMeta.push(meta);
      }
    });
  }

  return mergedMeta.map(buildProtocolEntry);
};

const findProtocolEntryByName = (protocolName, protocolMeta = []) => {
  const normalizedName = normalizeProtocolName(protocolName);
  if (!normalizedName) {
    return null;
  }

  return buildProtocolEntries(protocolMeta).find((entry) => entry.value === normalizedName) || null;
};

export const getProtocolPresentation = (protocolName, protocolMeta = []) => {
  const normalizedName = normalizeProtocolName(protocolName);
  const matchedEntry = findProtocolEntryByName(normalizedName, protocolMeta);

  if (matchedEntry) {
    return matchedEntry;
  }

  return {
    value: normalizedName,
    label: normalizedName ? normalizedName.toUpperCase() : '',
    color: DEFAULT_PROTOCOL_COLOR,
    icon: normalizedName ? normalizedName.charAt(0).toUpperCase() : '?',
    prefixes: normalizedName ? [`${normalizedName}://`] : []
  };
};

export const getProtocolOptions = (protocolOptions = [], protocolMeta = []) =>
  (protocolOptions || []).map((protocol) => getProtocolPresentation(protocol, protocolMeta)).filter((option) => option.value);

export const getRegisteredProtocolNames = (protocolMeta = []) =>
  uniqueValues((Array.isArray(protocolMeta) ? protocolMeta : []).map((meta) => normalizeProtocolName(meta?.name)).filter(Boolean));

export const resolveProtocolPresentationFromLink = (link, protocolMeta = []) => {
  if (!link || typeof link !== 'string') {
    return {
      value: '',
      label: '未知',
      color: UNKNOWN_PROTOCOL_COLOR,
      icon: '?',
      prefixes: []
    };
  }

  const linkLower = link.trim().toLowerCase();
  const matchedEntry = buildProtocolEntries(protocolMeta).find((entry) => entry.prefixes.some((prefix) => linkLower.startsWith(prefix)));

  if (matchedEntry) {
    return matchedEntry;
  }

  return {
    value: '',
    label: '其他',
    color: OTHER_PROTOCOL_COLOR,
    icon: '?',
    prefixes: []
  };
};

export const getProtocolGradient = (protocolName, protocolMeta = [], fallbackGradient = DEFAULT_PROTOCOL_GRADIENT) => {
  const presentation = getProtocolPresentation(protocolName, protocolMeta);
  const color = presentation?.color || DEFAULT_PROTOCOL_COLOR;

  if (!color.startsWith('#') || (color.length !== 7 && color.length !== 4)) {
    return fallbackGradient;
  }

  return `linear-gradient(135deg, ${color} 0%, ${color}CC 100%)`;
};

export const buildFieldMetaMap = (fields = []) => {
  const fieldMap = {};
  (fields || []).forEach((field) => {
    if (field?.name) {
      fieldMap[field.name] = field;
    }
  });
  return fieldMap;
};

export const getFieldGroupKey = (fieldName, fieldMeta = null) => {
  if (fieldMeta?.group) {
    return fieldMeta.group;
  }

  const baseName = (fieldName || '').split('.').pop()?.toLowerCase() || '';

  if (['name', 'ps', 'server', 'host', 'add', 'port', 'hostname'].some((keyword) => baseName.includes(keyword.toLowerCase()))) {
    return 'basic';
  }
  if (['password', 'uuid', 'id', 'auth', 'username'].some((keyword) => baseName.includes(keyword.toLowerCase()))) {
    return 'auth';
  }
  if (
    ['net', 'type', 'path', 'encryption', 'cipher', 'method', 'obfs', 'protocol', 'flow', 'mode', 'servicename', 'headertype'].some(
      (keyword) => baseName.includes(keyword.toLowerCase())
    )
  ) {
    return 'transport';
  }
  if (
    ['tls', 'security', 'sni', 'alpn', 'fp', 'pbk', 'sid', 'peer', 'insecure', 'skipcertverify', 'clientfingerprint', 'allowinsecure'].some(
      (keyword) => baseName.includes(keyword.toLowerCase())
    )
  ) {
    return 'tls';
  }

  return fieldMeta?.advanced ? 'advanced' : 'other';
};
