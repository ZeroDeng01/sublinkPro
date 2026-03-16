export const FRAUD_SCORE_LEVELS = [
  { max: 10, category: '极佳', icon: '⚪' },
  { max: 30, category: '优秀', icon: '🟢' },
  { max: 50, category: '良好', icon: '🟡' },
  { max: 70, category: '中等', icon: '🟠' },
  { max: 89, category: '差', icon: '🔴' },
  { max: Infinity, category: '极差', icon: '⚫' }
];

export const getFraudScoreLevel = (fraudScore) => {
  if (fraudScore === undefined || fraudScore === null || fraudScore < 0) {
    return null;
  }
  return FRAUD_SCORE_LEVELS.find((level) => fraudScore <= level.max) || FRAUD_SCORE_LEVELS[FRAUD_SCORE_LEVELS.length - 1];
};

export const getFraudScoreIcon = (fraudScore) => {
  const level = getFraudScoreLevel(fraudScore);
  return level?.icon || '⛔️';
};
