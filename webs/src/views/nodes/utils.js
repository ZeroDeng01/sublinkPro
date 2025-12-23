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
