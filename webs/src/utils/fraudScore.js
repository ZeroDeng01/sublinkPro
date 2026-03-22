export const FRAUD_SCORE_LEVELS = [
  { max: 10, category: '极佳', icon: '⚪' },
  { max: 30, category: '优秀', icon: '🟢' },
  { max: 50, category: '良好', icon: '🟡' },
  { max: 70, category: '中等', icon: '🟠' },
  { max: 89, category: '差', icon: '🔴' },
  { max: Infinity, category: '极差', icon: '⚫' }
];

export const QUALITY_STATUS = {
  UNTESTED: 'untested',
  SUCCESS: 'success',
  PARTIAL: 'partial',
  FAILED: 'failed',
  DISABLED: 'disabled'
};

export const QUALITY_STATUS_OPTIONS = [
  { value: '', label: '全部' },
  { value: QUALITY_STATUS.SUCCESS, label: '完整结果' },
  { value: QUALITY_STATUS.PARTIAL, label: '信息不全' },
  { value: QUALITY_STATUS.FAILED, label: '检测失败' },
  { value: QUALITY_STATUS.DISABLED, label: '未启用' },
  { value: QUALITY_STATUS.UNTESTED, label: '未检测' }
];

export const getFraudScoreLevel = (fraudScore) => {
  if (fraudScore === undefined || fraudScore === null || fraudScore < 0) {
    return null;
  }
  return FRAUD_SCORE_LEVELS.find((level) => fraudScore <= level.max) || FRAUD_SCORE_LEVELS[FRAUD_SCORE_LEVELS.length - 1];
};

export const getFraudScoreIcon = (fraudScore, qualityStatus = QUALITY_STATUS.SUCCESS) => {
  if (qualityStatus === QUALITY_STATUS.PARTIAL) return 'ℹ️';
  if (qualityStatus && qualityStatus !== QUALITY_STATUS.SUCCESS) return '⛔️';
  const level = getFraudScoreLevel(fraudScore);
  return level?.icon || '⛔️';
};

export const getQualityStatusMeta = (qualityStatus, qualityFamily) => {
  switch (qualityStatus) {
    case QUALITY_STATUS.SUCCESS:
      return {
        label: qualityFamily === 'ipv6' ? 'IPv6完整结果' : '完整结果',
        shortLabel: '完整结果',
        color: 'success',
        variant: 'outlined'
      };
    case QUALITY_STATUS.PARTIAL:
      return {
        label: '信息不全',
        shortLabel: '信息不全',
        color: 'info',
        variant: 'outlined',
        tooltip:
          qualityFamily === 'ipv6'
            ? 'IPv6 环境下质量接口未返回完整的欺诈评分、住宅属性或 IP 类型信息'
            : '质量接口未返回完整的欺诈评分、住宅属性或 IP 类型信息'
      };
    case QUALITY_STATUS.FAILED:
      return { label: '检测失败', shortLabel: '失败', color: 'error', variant: 'outlined' };
    case QUALITY_STATUS.DISABLED:
      return { label: '未启用', shortLabel: '未启用', color: 'default', variant: 'outlined' };
    case QUALITY_STATUS.UNTESTED:
    default:
      return { label: '未检测', shortLabel: '未检测', color: 'default', variant: 'outlined' };
  }
};
