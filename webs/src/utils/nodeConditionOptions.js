import {
  getNodeConditionFieldMetas,
  getUnlockProviderOptions,
  getUnlockStatusOptions,
  resolveNodeConditionOptionSource
} from 'views/nodes/utils';
import { QUALITY_STATUS_OPTIONS } from './fraudScore';

export const NODE_CONDITION_FIELDS = [
  { value: 'name', label: '备注' },
  { value: 'link_name', label: '原始名称' },
  { value: 'link_country', label: '国家代码' },
  { value: 'protocol', label: '协议类型' },
  { value: 'source', label: '来源' },
  { value: 'group', label: '分组' },
  { value: 'speed', label: '速度 (MB/s)' },
  { value: 'delay_time', label: '延迟 (ms)' },
  { value: 'fraud_score', label: '欺诈评分' },
  { value: 'quality_status', label: '质量状态' },
  { value: 'unlock_provider', label: '解锁 Provider' },
  { value: 'unlock_status', label: '解锁状态' },
  { value: 'unlock_keyword', label: '解锁关键词' },
  { value: 'unlock_result', label: '解锁摘要' },
  { value: 'ip_type', label: 'IP类型' },
  { value: 'residential_type', label: '住宅属性' },
  { value: 'speed_status', label: '速度状态' },
  { value: 'delay_status', label: '延迟状态' },
  { value: 'link_address', label: '地址' },
  { value: 'link_host', label: 'Host' },
  { value: 'link_port', label: '端口' },
  { value: 'dialer_proxy_name', label: '前置代理' },
  { value: 'link', label: '节点链接' }
];

const STATIC_FIELD_META_MAP = NODE_CONDITION_FIELDS.reduce((acc, field) => {
  acc[field.value] = field;
  return acc;
}, {});

export const UNLOCK_STATUS_OPTIONS = () => getUnlockStatusOptions(false);

export const UNLOCK_PROVIDER_OPTIONS = () => getUnlockProviderOptions();

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
  ip_type: NODE_IP_TYPE_OPTIONS,
  residential_type: NODE_RESIDENTIAL_TYPE_OPTIONS
};

const resolveFallbackFieldMeta = (field) => {
  if (!field) return null;

  if (field === 'unlock_status') {
    return { ...STATIC_FIELD_META_MAP[field], dataType: 'enum', inputType: 'select', optionSource: 'unlockStatuses' };
  }

  if (field === 'unlock_provider') {
    return { ...STATIC_FIELD_META_MAP[field], dataType: 'enum', inputType: 'select', optionSource: 'unlockProviders' };
  }

  if (NODE_CONDITION_NUMERIC_FIELDS.includes(field)) {
    return { ...STATIC_FIELD_META_MAP[field], dataType: 'number', inputType: 'text' };
  }

  if (NODE_CONDITION_VALUE_OPTIONS[field]) {
    return {
      ...STATIC_FIELD_META_MAP[field],
      dataType: 'enum',
      inputType: 'select',
      options: NODE_CONDITION_VALUE_OPTIONS[field]
    };
  }

  return STATIC_FIELD_META_MAP[field] ? { ...STATIC_FIELD_META_MAP[field], dataType: 'string', inputType: 'text' } : null;
};

export const getNodeConditionFieldMeta = (field) => {
  const dynamicMeta = getNodeConditionFieldMetas().find((item) => item?.value === field);
  return dynamicMeta || resolveFallbackFieldMeta(field);
};

export const getNodeConditionFields = () => {
  const dynamicFields = getNodeConditionFieldMetas();
  return dynamicFields.length > 0 ? dynamicFields : NODE_CONDITION_FIELDS;
};

export const getNodeConditionValueOptions = (field) => {
  const fieldMeta = getNodeConditionFieldMeta(field);
  if (!fieldMeta) {
    return null;
  }

  if (fieldMeta.optionSource) {
    return resolveNodeConditionOptionSource(fieldMeta.optionSource, false);
  }

  if (Array.isArray(fieldMeta.options) && fieldMeta.options.length > 0) {
    return fieldMeta.options;
  }

  if (field === 'unlock_status') {
    return UNLOCK_STATUS_OPTIONS();
  }

  if (field === 'unlock_provider') {
    return UNLOCK_PROVIDER_OPTIONS();
  }

  return NODE_CONDITION_VALUE_OPTIONS[field] || null;
};

export const isNodeConditionNumericFieldDynamic = (field) => getNodeConditionFieldMeta(field)?.dataType === 'number';

export const isNodeConditionNumericField = (field) =>
  isNodeConditionNumericFieldDynamic(field) || NODE_CONDITION_NUMERIC_FIELDS.includes(field);

export const isNodeConditionSelectField = (field) => {
  const fieldMeta = getNodeConditionFieldMeta(field);
  if (fieldMeta?.inputType) {
    return fieldMeta.inputType === 'select';
  }
  return Boolean(getNodeConditionValueOptions(field));
};
