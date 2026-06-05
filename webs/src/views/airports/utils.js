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

    const segments = part.split(',');
    for (const segment of segments) {
      if (segment === '*') {
        continue;
      }

      const wildcardStepMatch = segment.match(/^\*\/(\d+)$/);
      if (wildcardStepMatch) {
        if (parseInt(wildcardStepMatch[1], 10) <= 0) {
          return false;
        }
        continue;
      }

      const stepMatch = segment.match(/^(\d+|\d+-\d+)\/(\d+)$/);
      if (stepMatch) {
        const step = parseInt(stepMatch[2], 10);
        if (step <= 0) {
          return false;
        }

        const [startRaw, endRaw] = stepMatch[1].split('-');
        const start = parseInt(startRaw, 10);
        const end = endRaw ? parseInt(endRaw, 10) : start;
        if (start > end) {
          return false;
        }
        if (start < range.min || end > range.max) {
          return false;
        }
        continue;
      }

      const rangeMatch = segment.match(/^(\d+)-(\d+)$/);
      if (rangeMatch) {
        const start = parseInt(rangeMatch[1], 10);
        const end = parseInt(rangeMatch[2], 10);
        if (start > end || start < range.min || end > range.max) {
          return false;
        }
        continue;
      }

      if (/^\d+$/.test(segment)) {
        const value = parseInt(segment, 10);
        if (value < range.min || value > range.max) {
          return false;
        }
        continue;
      }

      return false;
    }
  }
  return true;
};

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

// User-Agent 预设选项
export const USER_AGENT_OPTIONS = [
  { labelKey: 'airports.form.userAgentOptions.none', value: '' },
  { label: 'clash.meta', value: 'clash.meta' },
  { label: 'clash', value: 'clash' },
  { label: 'v2ray', value: 'v2ray' },
  { label: 'clash-verge/v1.5.1', value: 'clash-verge/v1.5.1' }
];

// 格式化字节为人类可读格式
export const formatBytes = (bytes, decimals = 2) => {
  if (!bytes || bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];
  const i = Math.floor(Math.log(Math.abs(bytes)) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(decimals)) + ' ' + sizes[Math.min(i, sizes.length - 1)];
};

// 格式化过期时间 - 直接显示日期
export const formatExpireTime = (timestamp) => {
  if (!timestamp || timestamp === 0) return '';

  const expireDate = new Date(timestamp * 1000);
  const now = new Date();
  const year = expireDate.getFullYear();
  const month = String(expireDate.getMonth() + 1).padStart(2, '0');
  const day = String(expireDate.getDate()).padStart(2, '0');
  const hours = String(expireDate.getHours()).padStart(2, '0');
  const minutes = String(expireDate.getMinutes()).padStart(2, '0');
  const seconds = String(expireDate.getSeconds()).padStart(2, '0');
  const dateStr = `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;

  // 判断是否已过期
  if (expireDate < now) {
    return `${dateStr} (已过期)`;
  }
  return dateStr;
};

// 根据使用率百分比返回颜色
export const getUsageColor = (percent) => {
  if (percent < 60) return '#22c55e';
  if (percent < 85) return '#f59e0b';
  return '#ef4444';
};
