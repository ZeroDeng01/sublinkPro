import i18n from 'i18n';

export const FRAUD_SCORE_LEVELS = [
  { max: 10, category: 'excellentPlus', icon: '⚪' },
  { max: 30, category: 'excellent', icon: '🟢' },
  { max: 50, category: 'good', icon: '🟡' },
  { max: 70, category: 'medium', icon: '🟠' },
  { max: 89, category: 'poor', icon: '🔴' },
  { max: Infinity, category: 'veryPoor', icon: '⚫' }
];

const translate = (key, defaultValue) => i18n.t(key, { defaultValue });

export const QUALITY_STATUS = {
  UNTESTED: 'untested',
  SUCCESS: 'success',
  PARTIAL: 'partial',
  FAILED: 'failed',
  DISABLED: 'disabled'
};

export const QUALITY_STATUS_OPTIONS = [
  { value: '', label: 'All', labelKey: 'nodeConditions.qualityStatus.all' },
  { value: QUALITY_STATUS.SUCCESS, label: 'Complete result', labelKey: 'nodeConditions.qualityStatus.success' },
  { value: QUALITY_STATUS.PARTIAL, label: 'Incomplete info', labelKey: 'nodeConditions.qualityStatus.partial' },
  { value: QUALITY_STATUS.FAILED, label: 'Check failed', labelKey: 'nodeConditions.qualityStatus.failed' },
  { value: QUALITY_STATUS.DISABLED, label: 'Disabled', labelKey: 'nodeConditions.qualityStatus.disabled' },
  { value: QUALITY_STATUS.UNTESTED, label: 'Untested', labelKey: 'nodeConditions.qualityStatus.untested' }
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
        label:
          qualityFamily === 'ipv6'
            ? translate('nodeConditions.qualityStatus.ipv6Success', 'IPv6 complete result')
            : translate('nodeConditions.qualityStatus.success', 'Complete result'),
        shortLabel: translate('nodeConditions.qualityStatus.success', 'Complete result'),
        color: 'success',
        variant: 'outlined'
      };
    case QUALITY_STATUS.PARTIAL:
      return {
        label: translate('nodeConditions.qualityStatus.partial', 'Incomplete info'),
        shortLabel: translate('nodeConditions.qualityStatus.partial', 'Incomplete info'),
        color: 'info',
        variant: 'outlined',
        tooltip:
          qualityFamily === 'ipv6'
            ? translate(
                'nodeConditions.qualityStatus.tooltip.ipv6Partial',
                'The quality API did not return complete fraud score, residential, or IP type details in IPv6 environments'
              )
            : translate(
                'nodeConditions.qualityStatus.tooltip.partial',
                'The quality API did not return complete fraud score, residential, or IP type details'
              )
      };
    case QUALITY_STATUS.FAILED:
      return {
        label: translate('nodeConditions.qualityStatus.failed', 'Check failed'),
        shortLabel: translate('nodeConditions.qualityStatus.shortFailed', 'Failed'),
        color: 'error',
        variant: 'outlined'
      };
    case QUALITY_STATUS.DISABLED:
      return {
        label: translate('nodeConditions.qualityStatus.disabled', 'Disabled'),
        shortLabel: translate('nodeConditions.qualityStatus.disabled', 'Disabled'),
        color: 'default',
        variant: 'outlined'
      };
    case QUALITY_STATUS.UNTESTED:
    default:
      return {
        label: translate('nodeConditions.qualityStatus.untested', 'Untested'),
        shortLabel: translate('nodeConditions.qualityStatus.untested', 'Untested'),
        color: 'default',
        variant: 'outlined',
        state: QUALITY_STATUS.UNTESTED
      };
  }
};
