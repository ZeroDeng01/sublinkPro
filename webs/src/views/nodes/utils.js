// Cron 表达式预设 - 包含友好的说明
export const CRON_OPTIONS = [
  { label: "每30分钟", value: "*/30 * * * *" },
  { label: "每1小时", value: "0 * * * *" },
  { label: "每6小时", value: "0 */6 * * *" },
  { label: "每12小时", value: "0 */12 * * *" },
  { label: "每天", value: "0 0 * * *" },
  { label: "每周一", value: "0 0 * * 1" }
];

// 测速URL选项 - TCP模式 (204轻量)
export const SPEED_TEST_TCP_OPTIONS = [
  { label: "Cloudflare (cp.cloudflare.com)", value: "http://cp.cloudflare.com/generate_204" },
  { label: "Apple (captive.apple.com)", value: "http://captive.apple.com/generate_204" },
  { label: "Gstatic (www.gstatic.com)", value: "http://www.gstatic.com/generate_204" }
];

// 测速URL选项 - Mihomo模式 (真速度测试用下载)
export const SPEED_TEST_MIHOMO_OPTIONS = [
  { label: "10MB (Cloudflare)", value: "https://speed.cloudflare.com/__down?bytes=10000000" },
  { label: "50MB (Cloudflare)", value: "https://speed.cloudflare.com/__down?bytes=50000000" },
  { label: "100MB (Cloudflare)", value: "https://speed.cloudflare.com/__down?bytes=100000000" }
];

// User-Agent 预设选项
export const USER_AGENT_OPTIONS = [
  { label: "Clash (默认)", value: "Clash" },
  { label: "clash.meta", value: "clash.meta" },
  { label: "clash-verge/v1.5.1", value: "clash-verge/v1.5.1" }
];

// 格式化日期时间
export const formatDateTime = (dateTimeString) => {
  if (!dateTimeString || dateTimeString === "0001-01-01T00:00:00Z") {
    return "-";
  }
  try {
    const date = new Date(dateTimeString);
    if (isNaN(date.getTime())) {
      return "-";
    }
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, "0");
    const day = String(date.getDate()).padStart(2, "0");
    const hours = String(date.getHours()).padStart(2, "0");
    const minutes = String(date.getMinutes()).padStart(2, "0");
    const seconds = String(date.getSeconds()).padStart(2, "0");
    return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
  } catch {
    return "-";
  }
};

// ISO国家代码转换为国旗emoji
export const isoToFlag = (isoCode) => {
  if (!isoCode || isoCode.length !== 2) return "";
  const codePoints = isoCode
    .toUpperCase()
    .split("")
    .map((char) => 127397 + char.charCodeAt(0));
  return String.fromCodePoint(...codePoints);
};

// 格式化国家显示 (国旗emoji + 代码)
export const formatCountry = (linkCountry) => {
  if (!linkCountry) return "";
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
  if (delay <= 0) return "default";
  if (delay < 100) return "success";
  if (delay < 500) return "warning";
  return "error";
};
