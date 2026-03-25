import { getFraudScoreLevel, getQualityStatusMeta, QUALITY_STATUS, QUALITY_STATUS_OPTIONS } from 'utils/fraudScore';

// Cron 表达式预设 - 包含友好的说明
export const CRON_OPTIONS = [
  { label: '每30分钟', value: '*/30 * * * *' },
  { label: '每1小时', value: '0 * * * *' },
  { label: '每6小时', value: '0 */6 * * *' },
  { label: '每12小时', value: '0 */12 * * *' },
  { label: '每天', value: '0 0 * * *' },
  { label: '每周一', value: '0 0 * * 1' }
];

// 测速URL选项 - TCP模式 (204轻量)
export const SPEED_TEST_TCP_OPTIONS = [
  { label: 'Cloudflare (cp.cloudflare.com)', value: 'https://cp.cloudflare.com/generate_204' },
  { label: 'Apple (captive.apple.com)', value: 'https://captive.apple.com/generate_204' },
  { label: 'Gstatic (www.gstatic.com)', value: 'https://www.gstatic.com/generate_204' }
];

// 测速URL选项 - Mihomo模式 (真速度测试用下载)
export const SPEED_TEST_MIHOMO_OPTIONS = [
  { label: '1MB (Cloudflare)', value: 'https://speed.cloudflare.com/__down?bytes=1000000' },
  { label: '3MB (Cloudflare)', value: 'https://speed.cloudflare.com/__down?bytes=3000000' },
  { label: '5MB (Cloudflare)', value: 'https://speed.cloudflare.com/__down?bytes=5000000' },
  { label: '10MB (Cloudflare)', value: 'https://speed.cloudflare.com/__down?bytes=10000000' },
  { label: '50MB (Cloudflare)', value: 'https://speed.cloudflare.com/__down?bytes=50000000' },
  { label: '100MB (Cloudflare)', value: 'https://speed.cloudflare.com/__down?bytes=100000000' }
];

// 延迟测试URL选项 (用于Mihomo模式的阶段一)
export const LATENCY_TEST_URL_OPTIONS = [
  { label: 'Cloudflare 204 (推荐)', value: 'https://cp.cloudflare.com/generate_204' },
  { label: 'Apple 204', value: 'https://captive.apple.com/generate_204' },
  { label: 'Gstatic 204', value: 'https://www.gstatic.com/generate_204' }
];

// 落地IP查询接口选项
export const LANDING_IP_URL_OPTIONS = [
  { label: 'ipify.org (推荐)', value: 'https://api.ipify.org' },
  { label: 'ip.sb', value: 'https://api.ip.sb/ip' },
  { label: 'ifconfig.me', value: 'https://ifconfig.me/ip' },
  { label: 'icanhazip.com', value: 'https://icanhazip.com' },
  { label: 'ipinfo.io', value: 'https://ipinfo.io/ip' }
];

// IP质量检测接口选项
export const QUALITY_CHECK_URL_OPTIONS = [
  { label: 'IPPure (推荐)', value: 'https://my.123169.xyz/v1/info' },
  { label: 'IPPure', value: 'https://my.ippure.com/v1/info' }
];

const UNLOCK_PROVIDER_LABEL_MAP = {
  netflix: 'Netflix',
  disney: 'Disney+',
  disneyplus: 'Disney+',
  youtube: 'YouTube',
  youtube_premium: 'YouTube Premium',
  youtubepremium: 'YouTube Premium',
  tiktok: 'TikTok',
  openai: 'OpenAI',
  chatgpt: 'ChatGPT',
  claude: 'Claude',
  copilot: 'GitHub Copilot',
  gemini: 'Gemini',
  spotify: 'Spotify',
  bbc_iplayer: 'BBC iPlayer',
  prime_video: 'Prime Video',
  amazon_prime: 'Prime Video',
  bilibili_hk_mo_tw: 'Bilibili 港澳台',
  bahamut: 'Bahamut',
  tvbanywhere: 'TVB Anywhere'
};

const DEFAULT_UNLOCK_STATUS_METAS = [
  { value: 'available', label: '解锁', shortLabel: '解锁', color: 'success', severity: 'success', description: '可正常使用' },
  {
    value: 'partial',
    label: '部分',
    shortLabel: '部分',
    color: 'warning',
    severity: 'warning',
    description: '仅部分能力可用'
  },
  {
    value: 'reachable',
    label: '直连',
    shortLabel: '直连',
    color: 'info',
    severity: 'info',
    description: '可以访问，但不代表完整能力可用'
  },
  { value: 'restricted', label: '受限', shortLabel: '受限', color: 'error', severity: 'error', description: '当前地区或出口被限制' },
  {
    value: 'unsupported',
    label: '不支持',
    shortLabel: '不支持',
    color: 'warning',
    severity: 'warning',
    description: '当前地区不在官方支持范围内'
  },
  { value: 'unknown', label: '未知', shortLabel: '未知', color: 'default', severity: 'default', description: '当前无法稳定判断结果' },
  { value: 'error', label: '异常', shortLabel: '异常', color: 'error', severity: 'error', description: '本轮检测出现错误' },
  { value: 'untested', label: '未测', shortLabel: '未测', color: 'default', severity: 'default', description: '尚未执行检测' }
];

let unlockMetaState = {
  providers: [],
  statuses: DEFAULT_UNLOCK_STATUS_METAS,
  renameVariables: [],
  conditionFields: []
};

const getCleanString = (value) => {
  if (value === undefined || value === null) return '';
  return String(value).trim();
};

export const setUnlockMeta = (meta = {}) => {
  unlockMetaState = {
    providers: Array.isArray(meta.unlockProviders) ? meta.unlockProviders : Array.isArray(meta.providers) ? meta.providers : [],
    statuses: Array.isArray(meta.unlockStatuses) && meta.unlockStatuses.length > 0 ? meta.unlockStatuses : DEFAULT_UNLOCK_STATUS_METAS,
    renameVariables: Array.isArray(meta.unlockRenameVariables) ? meta.unlockRenameVariables : [],
    conditionFields: Array.isArray(meta.conditionFields) ? meta.conditionFields : []
  };
};

export const getUnlockProviderOptions = () => unlockMetaState.providers || [];
export const getUnlockRenameVariables = () => unlockMetaState.renameVariables || [];
export const getNodeConditionFieldMetas = () => unlockMetaState.conditionFields || [];

export const getUnlockStatusOptions = (includeAll = true) => {
  const statusOptions = (unlockMetaState.statuses || DEFAULT_UNLOCK_STATUS_METAS).map((item) => ({
    value: item.value,
    label: item.label || item.shortLabel || item.value,
    shortLabel: item.shortLabel || item.label || item.value,
    color: item.color || item.severity || 'default',
    description: item.description || ''
  }));

  return includeAll ? [{ value: '', label: '全部' }, ...statusOptions] : statusOptions;
};

export const resolveNodeConditionOptionSource = (optionSource, includeAll = false) => {
  switch (optionSource) {
    case 'unlockProviders':
      return getUnlockProviderOptions();
    case 'unlockStatuses':
      return getUnlockStatusOptions(includeAll);
    default:
      return [];
  }
};

export const createEmptyUnlockRule = () => ({ provider: '', status: '', keyword: '' });

export const getUnlockRuleModeOptions = () => [
  { value: 'or', label: '满足任意一条', description: '多条规则之间按 OR 匹配' },
  { value: 'and', label: '同时满足所有规则', description: '多条规则之间按 AND 匹配' }
];

export const normalizeUnlockRules = (rules) => {
  if (!Array.isArray(rules)) return [];
  return rules
    .map((rule) => {
      if (!rule || typeof rule !== 'object') return null;
      return {
        provider: getCleanString(rule.provider || rule.Provider || rule.value),
        status: getCleanString(rule.status || rule.Status),
        keyword: getCleanString(rule.keyword || rule.Keyword)
      };
    })
    .filter((rule) => rule && (rule.provider || rule.status || rule.keyword));
};

export const buildUnlockRulesPayload = (rules) => JSON.stringify(normalizeUnlockRules(rules));

export const parseNodeCheckProfileList = (value) => {
  if (Array.isArray(value)) {
    return value
      .map((item) => {
        if (typeof item === 'string') return item.trim();
        if (item && typeof item === 'object') return getCleanString(item.name || item.value || item.label);
        return '';
      })
      .filter(Boolean);
  }

  if (typeof value === 'string') {
    return value
      .split(',')
      .map((item) => item.trim())
      .filter(Boolean);
  }

  return [];
};

export const normalizeUnlockProviders = (providers) => {
  if (Array.isArray(providers)) {
    return [
      ...new Set(
        providers
          .map((provider) => {
            if (typeof provider === 'string') return provider.trim();
            if (provider && typeof provider === 'object') {
              return getCleanString(provider.value || provider.provider || provider.name || provider.label);
            }
            return '';
          })
          .filter(Boolean)
      )
    ];
  }

  if (typeof providers === 'string') {
    const trimmed = providers.trim();
    if (!trimmed) return [];

    const parsed = safeParseJson(trimmed);
    if (Array.isArray(parsed)) {
      return normalizeUnlockProviders(parsed);
    }

    return trimmed
      .split(',')
      .map((provider) => provider.trim())
      .filter(Boolean);
  }

  return [];
};

export const formatUnlockProviderLabel = (provider) => {
  const raw = getCleanString(provider);
  if (!raw) return '';

  const normalized = raw.toLowerCase().replace(/[\s-]+/g, '_');
  const dynamicOption = getUnlockProviderOptions().find((item) => getCleanString(item?.value).toLowerCase() === normalized);
  if (dynamicOption?.label) {
    return dynamicOption.label;
  }
  if (UNLOCK_PROVIDER_LABEL_MAP[normalized]) {
    return UNLOCK_PROVIDER_LABEL_MAP[normalized];
  }

  return raw
    .split(/[_\s-]+/)
    .filter(Boolean)
    .map((segment) => segment.charAt(0).toUpperCase() + segment.slice(1))
    .join(' ');
};

export const formatUnlockProvidersSummary = (providers, limit = 2) => {
  const normalizedProviders = normalizeUnlockProviders(providers);
  if (normalizedProviders.length === 0) return '未选择';

  const labels = normalizedProviders.slice(0, limit).map((provider) => formatUnlockProviderLabel(provider));
  const extraCount = normalizedProviders.length - labels.length;
  return extraCount > 0 ? `${labels.join('、')} +${extraCount}` : labels.join('、');
};

export const createNodeCheckProfileFormState = (profile = null) => {
  if (profile) {
    return {
      name: profile.name || '',
      enabled: profile.enabled || false,
      cronExpr: profile.cronExpr || '',
      mode: profile.mode || 'tcp',
      testUrl: profile.testUrl || '',
      latencyUrl: profile.latencyUrl || '',
      timeout: profile.timeout || 5,
      groups: parseNodeCheckProfileList(profile.groups),
      tags: parseNodeCheckProfileList(profile.tags),
      latencyConcurrency: profile.latencyConcurrency || 0,
      speedConcurrency: profile.speedConcurrency ?? 0,
      detectCountry: profile.detectCountry || false,
      landingIpUrl: profile.landingIpUrl || '',
      includeHandshake: profile.includeHandshake !== false,
      speedRecordMode: profile.speedRecordMode || 'average',
      peakSampleInterval: profile.peakSampleInterval || 100,
      trafficByGroup: profile.trafficByGroup !== false,
      trafficBySource: profile.trafficBySource !== false,
      trafficByNode: profile.trafficByNode || false,
      preserveSpeedResult: profile.preserveSpeedResult || false,
      detectQuality: profile.detectQuality || false,
      qualityCheckUrl: profile.qualityCheckUrl || '',
      detectUnlock: profile.detectUnlock || false,
      unlockProviders: normalizeUnlockProviders(profile.unlockProviders)
    };
  }

  return {
    name: '',
    enabled: false,
    cronExpr: '',
    mode: 'tcp',
    testUrl: SPEED_TEST_TCP_OPTIONS[0]?.value || '',
    latencyUrl: '',
    timeout: 5,
    groups: [],
    tags: [],
    latencyConcurrency: 0,
    speedConcurrency: 0,
    detectCountry: false,
    landingIpUrl: '',
    includeHandshake: true,
    speedRecordMode: 'average',
    peakSampleInterval: 100,
    trafficByGroup: true,
    trafficBySource: true,
    trafficByNode: false,
    preserveSpeedResult: false,
    detectQuality: false,
    qualityCheckUrl: '',
    detectUnlock: false,
    unlockProviders: []
  };
};

export const buildNodeCheckProfilePayload = (profile, overrides = {}) => ({
  name: getCleanString(profile.name),
  enabled: Boolean(profile.enabled),
  cronExpr: profile.cronExpr || '',
  mode: profile.mode || 'tcp',
  testUrl: profile.testUrl || '',
  latencyUrl: profile.latencyUrl || '',
  timeout: profile.timeout,
  groups: parseNodeCheckProfileList(profile.groups),
  tags: parseNodeCheckProfileList(profile.tags),
  latencyConcurrency: profile.latencyConcurrency ?? 0,
  speedConcurrency: profile.speedConcurrency ?? 0,
  detectCountry: Boolean(profile.detectCountry),
  landingIpUrl: profile.landingIpUrl || '',
  includeHandshake: profile.includeHandshake !== false,
  speedRecordMode: profile.speedRecordMode || 'average',
  peakSampleInterval: profile.peakSampleInterval || 100,
  trafficByGroup: profile.trafficByGroup !== false,
  trafficBySource: profile.trafficBySource !== false,
  trafficByNode: Boolean(profile.trafficByNode),
  preserveSpeedResult: Boolean(profile.preserveSpeedResult),
  detectQuality: Boolean(profile.detectQuality),
  qualityCheckUrl: profile.qualityCheckUrl || '',
  detectUnlock: Boolean(profile.detectUnlock),
  unlockProviders: normalizeUnlockProviders(profile.unlockProviders),
  ...overrides
});

export const getUnlockStatusMeta = (status) => {
  const normalized = getCleanString(status).toLowerCase();

  const dynamicMeta = (unlockMetaState.statuses || DEFAULT_UNLOCK_STATUS_METAS).find(
    (item) => getCleanString(item?.value).toLowerCase() === normalized
  );
  if (dynamicMeta) {
    return {
      key: normalized || 'default',
      label: dynamicMeta.label || dynamicMeta.shortLabel || status,
      shortLabel: dynamicMeta.shortLabel || dynamicMeta.label || status,
      color: dynamicMeta.color || dynamicMeta.severity || 'default',
      variant: 'outlined',
      description: dynamicMeta.description || ''
    };
  }

  if (!normalized) {
    return { key: 'default', label: '未知', shortLabel: '未知', color: 'default', variant: 'outlined' };
  }

  if (['success', 'supported', 'unlock', 'unlocked', 'available', 'ok', 'passed', 'pass', 'yes'].includes(normalized)) {
    return { key: 'success', label: '可用', shortLabel: '可用', color: 'success', variant: 'outlined' };
  }

  if (['partial', 'limited', 'warning', 'region_restricted'].includes(normalized)) {
    return { key: 'warning', label: '受限', shortLabel: '受限', color: 'warning', variant: 'outlined' };
  }

  if (['pending', 'running', 'checking', 'processing'].includes(normalized)) {
    return { key: 'info', label: '检测中', shortLabel: '检测中', color: 'info', variant: 'outlined' };
  }

  if (['timeout', 'timed_out'].includes(normalized)) {
    return { key: 'timeout', label: '超时', shortLabel: '超时', color: 'warning', variant: 'outlined' };
  }

  if (['blocked', 'locked', 'unsupported', 'unavailable', 'denied', 'forbidden', 'error', 'failed', 'fail', 'no'].includes(normalized)) {
    return { key: 'error', label: '不可用', shortLabel: '不可用', color: 'error', variant: 'outlined' };
  }

  return {
    key: 'custom',
    label: getCleanString(status),
    shortLabel: getCleanString(status),
    color: 'default',
    variant: 'outlined'
  };
};

const safeParseJson = (value) => {
  if (typeof value !== 'string') return value;
  try {
    return JSON.parse(value);
  } catch {
    return null;
  }
};

export const parseUnlockSummary = (unlockSummary) => {
  if (!unlockSummary) return null;

  const parsed = typeof unlockSummary === 'string' ? safeParseJson(unlockSummary) : unlockSummary;
  if (!parsed || typeof parsed !== 'object') return null;

  const providers = Array.isArray(parsed.providers)
    ? parsed.providers
        .map((item) => ({
          provider: getCleanString(item?.provider || item?.name),
          status: getCleanString(item?.status),
          region: getCleanString(item?.region),
          reason: getCleanString(item?.reason),
          detail: getCleanString(item?.detail)
        }))
        .filter((item) => item.provider)
    : [];

  if (providers.length === 0 && !parsed.updatedAt) {
    return null;
  }

  return {
    providers,
    updatedAt: parsed.updatedAt || '',
    counts: parsed.counts && typeof parsed.counts === 'object' ? parsed.counts : null
  };
};

export const extractUnlockSummaryFromTaskResult = (taskResult) => {
  const parsedResult = typeof taskResult === 'string' ? safeParseJson(taskResult) : taskResult;
  if (!parsedResult || typeof parsedResult !== 'object') return null;

  if (parsedResult.data && typeof parsedResult.data === 'object') {
    return extractUnlockSummaryFromTaskResult(parsedResult.data);
  }

  if (parsedResult.unlockSummary) {
    return parseUnlockSummary(parsedResult.unlockSummary);
  }

  if (parsedResult.unlock && typeof parsedResult.unlock === 'object') {
    return parseUnlockSummary(parsedResult.unlock);
  }

  if (parsedResult.unlockEnabled && parsedResult.unlock && typeof parsedResult.unlock === 'object') {
    const aggregateProviders = Array.isArray(parsedResult.unlock.providers) ? parsedResult.unlock.providers : [];
    const aggregateCounts = parsedResult.unlock.counts && typeof parsedResult.unlock.counts === 'object' ? parsedResult.unlock.counts : {};
    const aggregateSummary = aggregateProviders
      .map((provider) => {
        const counts = aggregateCounts[provider] || {};
        const bestStatus = ['available', 'reachable', 'partial', 'restricted', 'unsupported', 'error', 'unknown'].find(
          (status) => counts[status] > 0
        );
        if (!bestStatus) return null;
        return {
          provider,
          status: bestStatus,
          detail: Object.entries(counts)
            .map(([status, count]) => `${status}:${count}`)
            .join(' · ')
        };
      })
      .filter(Boolean);

    if (aggregateSummary.length > 0) {
      return parseUnlockSummary({ providers: aggregateSummary });
    }
  }

  if (Array.isArray(parsedResult.providers)) {
    return parseUnlockSummary(parsedResult);
  }

  if (Array.isArray(parsedResult.unlockProviders) && parsedResult.unlockProviders.some((item) => item && typeof item === 'object')) {
    return parseUnlockSummary({ providers: parsedResult.unlockProviders, updatedAt: parsedResult.updatedAt || parsedResult.unlockCheckAt });
  }

  return null;
};

export const getUnlockProviderDisplay = (providerResult) => {
  const providerLabel = formatUnlockProviderLabel(providerResult?.provider);
  const statusMeta = getUnlockStatusMeta(providerResult?.status);
  const region = getCleanString(providerResult?.region);
  const reason = getCleanString(providerResult?.reason);
  const detail = getCleanString(providerResult?.detail);

  const compactStatus = statusMeta.key === 'available' && region ? region : statusMeta.shortLabel;
  const compactLabel = compactStatus ? `${providerLabel} ${compactStatus}` : providerLabel;
  const tooltip = [statusMeta.label, statusMeta.description || '', region ? `地区: ${region}` : '', reason ? `原因: ${reason}` : '', detail]
    .filter(Boolean)
    .join(' · ');

  return {
    provider: providerResult?.provider,
    providerLabel,
    compactLabel,
    statusLabel: statusMeta.label,
    statusShortLabel: statusMeta.shortLabel,
    color: statusMeta.color,
    variant: statusMeta.variant,
    region,
    reason,
    detail,
    tooltip,
    description: statusMeta.description || ''
  };
};

export const getUnlockSummaryDisplay = (unlockSummary, options = {}) => {
  const parsedSummary = parseUnlockSummary(unlockSummary);
  if (!parsedSummary) return null;

  const limit = options.limit ?? 2;
  const items = parsedSummary.providers.map((provider) => getUnlockProviderDisplay(provider));
  const compactItems = items.slice(0, limit);
  const extraCount = Math.max(items.length - compactItems.length, 0);

  return {
    items,
    compactItems,
    extraCount,
    updatedAt: parsedSummary.updatedAt,
    summaryText: compactItems.map((item) => item.compactLabel).join(' · ')
  };
};

export const getNodeUnlockSummaryDisplay = (node, options = {}) => {
  if (!node) return null;

  const display = getUnlockSummaryDisplay(node.unlockSummary || node.UnlockSummary, options);
  if (!display) return null;

  return {
    ...display,
    checkedAt: node.unlockCheckAt || node.UnlockCheckAt || display.updatedAt || ''
  };
};

export const getUnlockTaskResultText = (taskResult, limit = 2) => {
  const summary = getUnlockSummaryDisplay(extractUnlockSummaryFromTaskResult(taskResult), { limit });
  if (!summary || summary.compactItems.length === 0) return null;

  return `解锁 ${summary.summaryText}${summary.extraCount > 0 ? ` +${summary.extraCount}` : ''}`;
};

// User-Agent 预设选项
export const USER_AGENT_OPTIONS = [
  { label: '无 (空)', value: '' },
  { label: 'clash.meta', value: 'clash.meta' },
  { label: 'clash', value: 'clash' },
  { label: 'v2ray', value: 'v2ray' },
  { label: 'clash-verge/v1.5.1', value: 'clash-verge/v1.5.1' }
];

// 格式化日期时间
export const formatDateTime = (dateTimeString) => {
  if (!dateTimeString || dateTimeString === '0001-01-01T00:00:00Z') {
    return '-';
  }
  try {
    const date = new Date(dateTimeString);
    if (isNaN(date.getTime())) {
      return '-';
    }
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');
    const seconds = String(date.getSeconds()).padStart(2, '0');
    return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
  } catch {
    return '-';
  }
};

// ISO国家代码转换为国旗emoji
export const isoToFlag = (isoCode) => {
  if (!isoCode || isoCode.length !== 2) return '';
  isoCode = isoCode.toUpperCase() === 'TW' ? 'CN' : isoCode;
  const codePoints = isoCode
    .toUpperCase()
    .split('')
    .map((char) => 127397 + char.charCodeAt(0));
  return String.fromCodePoint(...codePoints);
};

// 格式化国家显示 (国旗emoji + 代码)
export const formatCountry = (linkCountry) => {
  if (!linkCountry) return '';
  const flag = isoToFlag(linkCountry);
  return flag ? `${flag} ${linkCountry}` : linkCountry;
};

// Cron 表达式验证
export const validateCronExpression = (cron) => {
  if (!cron) return false;
  const parts = cron.trim().split(/\s+/);
  if (parts.length !== 5) return false;

  const ranges = [
    { min: 0, max: 59 }, // 分钟
    { min: 0, max: 23 }, // 小时
    { min: 1, max: 31 }, // 日
    { min: 1, max: 12 }, // 月
    { min: 0, max: 7 } // 星期 (0和7都表示周日)
  ];

  for (let i = 0; i < 5; i++) {
    const part = parts[i];
    const range = ranges[i];

    // 支持的模式: *, */n, n, n-m, n,m,o
    const patterns = [
      /^\*$/, // *
      /^\*\/\d+$/, // */n
      /^\d+$/, // n
      /^\d+-\d+$/, // n-m
      /^[\d,]+$/ // n,m,o
    ];

    if (!patterns.some((p) => p.test(part))) {
      return false;
    }

    // 验证数字范围
    const numbers = part.match(/\d+/g);
    if (numbers) {
      for (const num of numbers) {
        const n = parseInt(num, 10);
        if (n < range.min || n > range.max) {
          return false;
        }
      }
    }
  }
  return true;
};

// 延迟颜色
export const getDelayColor = (delay) => {
  if (delay <= 0) return 'default';
  if (delay < 200) return 'success';
  if (delay < 500) return 'warning';
  return 'error';
};

// ========== 节点测试状态常量 (与后端 models/status_constants.go 保持同步) ==========
export const NODE_STATUS = {
  UNTESTED: 'untested', // 未测试
  SUCCESS: 'success', // 成功
  TIMEOUT: 'timeout', // 超时
  ERROR: 'error' // 错误
};

export { QUALITY_STATUS_OPTIONS };

// 状态选择器选项 (用于过滤器下拉框)
export const STATUS_OPTIONS = [
  { value: '', label: '全部' },
  { value: NODE_STATUS.UNTESTED, label: '未测速', color: 'default' },
  { value: NODE_STATUS.SUCCESS, label: '成功', color: 'success' },
  { value: NODE_STATUS.TIMEOUT, label: '超时', color: 'warning' },
  { value: NODE_STATUS.ERROR, label: '失败', color: 'error' }
];

// 速度颜色 (基于数值)
export const getSpeedColor = (speed) => {
  if (speed === -1) return 'error';
  if (speed <= 0) return 'default';
  if (speed >= 5) return 'success';
  if (speed >= 1) return 'warning';
  return 'error';
};

// 速度状态显示 - 统一处理所有速度显示逻辑
export const getSpeedDisplay = (speed, speedStatus) => {
  // 优先根据状态判断
  if (speedStatus === NODE_STATUS.TIMEOUT) {
    return { label: '超时', color: 'warning', variant: 'outlined' };
  }
  if (speedStatus === NODE_STATUS.ERROR || speed === -1) {
    return { label: '失败', color: 'error', variant: 'outlined' };
  }
  if (speedStatus === NODE_STATUS.UNTESTED || (!speedStatus && speed <= 0)) {
    return { label: '未测速', color: 'default', variant: 'outlined' };
  }
  // 成功状态，显示具体速度值
  return { label: `${speed.toFixed(2)}MB/s`, color: getSpeedColor(speed), variant: 'outlined' };
};

// 延迟状态显示 - 统一处理所有延迟显示逻辑
export const getDelayDisplay = (delay, delayStatus) => {
  // 优先根据状态判断
  if (delayStatus === NODE_STATUS.TIMEOUT || delay === -1) {
    return { label: '超时', color: 'error', variant: 'outlined' };
  }
  if (delayStatus === NODE_STATUS.ERROR) {
    return { label: '失败', color: 'error', variant: 'outlined' };
  }
  if (delayStatus === NODE_STATUS.UNTESTED || (!delayStatus && delay <= 0)) {
    return { label: '未测速', color: 'default', variant: 'outlined' };
  }
  // 成功状态，显示具体延迟值
  return { label: `${delay}ms`, color: getDelayColor(delay), variant: 'outlined' };
};

export const getFraudScoreDisplay = (fraudScore, qualityStatus = QUALITY_STATUS.UNTESTED, qualityFamily = '') => {
  if (qualityStatus !== QUALITY_STATUS.SUCCESS) {
    const statusMeta = getQualityStatusMeta(qualityStatus, qualityFamily);
    return {
      label: statusMeta.shortLabel,
      detailLabel: statusMeta.label,
      color: statusMeta.color,
      variant: statusMeta.variant,
      tooltip: statusMeta.tooltip
    };
  }

  if (fraudScore === undefined || fraudScore === null || fraudScore < 0) {
    return { label: '未检测', color: 'default', variant: 'outlined' };
  }

  const matchedLevel = getFraudScoreLevel(fraudScore);
  const levelStyles = {
    极佳: {
      sx: (theme) => ({
        color: theme.palette.mode === 'dark' ? '#e5e7eb' : '#6b7280',
        borderColor: theme.palette.mode === 'dark' ? '#94a3b8' : '#9ca3af',
        backgroundColor: theme.palette.mode === 'dark' ? 'rgba(148,163,184,0.12)' : 'rgba(148,163,184,0.08)'
      })
    },
    优秀: {
      sx: {
        color: '#15803d',
        borderColor: '#22c55e',
        backgroundColor: 'rgba(34,197,94,0.08)'
      }
    },
    良好: {
      sx: {
        color: '#a16207',
        borderColor: '#eab308',
        backgroundColor: 'rgba(234,179,8,0.10)'
      }
    },
    中等: {
      sx: {
        color: '#c2410c',
        borderColor: '#f97316',
        backgroundColor: 'rgba(249,115,22,0.10)'
      }
    },
    差: {
      sx: {
        color: '#b91c1c',
        borderColor: '#ef4444',
        backgroundColor: 'rgba(239,68,68,0.10)'
      }
    },
    极差: {
      sx: (theme) => ({
        color: theme.palette.mode === 'dark' ? '#cbd5e1' : '#111827',
        borderColor: theme.palette.mode === 'dark' ? '#64748b' : '#1f2937',
        backgroundColor: theme.palette.mode === 'dark' ? 'rgba(15,23,42,0.35)' : 'rgba(15,23,42,0.08)'
      })
    }
  };
  const matchedStyle = levelStyles[matchedLevel.category] || levelStyles.极差;
  return {
    label: `${fraudScore}`,
    detailLabel: `${fraudScore} (${matchedLevel.category})`,
    color: 'default',
    variant: 'outlined',
    sx: matchedStyle.sx
  };
};

export const getIpTypeDisplay = (isBroadcast, qualityStatus = QUALITY_STATUS.UNTESTED, qualityFamily = '') => {
  if (qualityStatus !== QUALITY_STATUS.SUCCESS) {
    const statusMeta = getQualityStatusMeta(qualityStatus, qualityFamily);
    return { label: statusMeta.shortLabel, color: statusMeta.color, variant: statusMeta.variant, tooltip: statusMeta.tooltip };
  }
  return isBroadcast
    ? { label: '广播IP', color: 'warning', variant: 'outlined' }
    : { label: '原生IP', color: 'success', variant: 'outlined' };
};

export const getResidentialDisplay = (isResidential, qualityStatus = QUALITY_STATUS.UNTESTED, qualityFamily = '') => {
  if (qualityStatus !== QUALITY_STATUS.SUCCESS) {
    const statusMeta = getQualityStatusMeta(qualityStatus, qualityFamily);
    return { label: statusMeta.shortLabel, color: statusMeta.color, variant: statusMeta.variant, tooltip: statusMeta.tooltip };
  }
  return isResidential
    ? { label: '住宅IP', color: 'success', variant: 'outlined' }
    : { label: '机房IP', color: 'default', variant: 'outlined' };
};

export const getQualityStatusDisplay = (qualityStatus, qualityFamily) => getQualityStatusMeta(qualityStatus, qualityFamily);
