import { QUALITY_STATUS_OPTIONS } from './fraudScore';

export const UNLOCK_STATUS_OPTIONS = [
  { value: 'available', label: '解锁' },
  { value: 'partial', label: '部分' },
  { value: 'reachable', label: '直连' },
  { value: 'restricted', label: '受限' },
  { value: 'unsupported', label: '不支持' },
  { value: 'unknown', label: '未知' },
  { value: 'error', label: '异常' },
  { value: 'untested', label: '未测' }
];

export const UNLOCK_PROVIDER_OPTIONS = [
  { value: 'netflix', label: 'Netflix' },
  { value: 'disney', label: 'Disney+' },
  { value: 'youtube_premium', label: 'YouTube Premium' },
  { value: 'openai', label: 'OpenAI' },
  { value: 'gemini', label: 'Gemini' },
  { value: 'claude', label: 'Claude' }
];

export const NODE_STATUS_OPTIONS = [
  { value: 'untested', label: '未测速' },
  { value: 'success', label: '成功' },
  { value: 'timeout', label: '超时' },
  { value: 'error', label: '失败' }
];

export const NODE_IP_TYPE_OPTIONS = [
  { value: 'native', label: '原生IP' },
  { value: 'broadcast', label: '广播IP' },
  { value: 'untested', label: '未检测' }
];

export const NODE_RESIDENTIAL_TYPE_OPTIONS = [
  { value: 'residential', label: '住宅IP' },
  { value: 'datacenter', label: '机房IP' },
  { value: 'untested', label: '未检测' }
];

export const NODE_CONDITION_NUMERIC_FIELDS = ['speed', 'delay_time', 'fraud_score'];

export const NODE_CONDITION_VALUE_OPTIONS = {
  speed_status: NODE_STATUS_OPTIONS,
  delay_status: NODE_STATUS_OPTIONS,
  quality_status: QUALITY_STATUS_OPTIONS.filter((option) => option.value !== ''),
  unlock_status: UNLOCK_STATUS_OPTIONS,
  unlock_provider: UNLOCK_PROVIDER_OPTIONS,
  ip_type: NODE_IP_TYPE_OPTIONS,
  residential_type: NODE_RESIDENTIAL_TYPE_OPTIONS
};

export const isNodeConditionNumericField = (field) => NODE_CONDITION_NUMERIC_FIELDS.includes(field);

export const getNodeConditionValueOptions = (field) => NODE_CONDITION_VALUE_OPTIONS[field] || null;

export const isNodeConditionSelectField = (field) => Boolean(getNodeConditionValueOptions(field));
