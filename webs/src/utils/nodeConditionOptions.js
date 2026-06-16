import {
  getNodeConditionFieldMetas,
  getUnlockProviderOptions,
  getUnlockStatusOptions,
  resolveNodeConditionOptionSource
} from 'views/nodes/utils';
import { QUALITY_STATUS_OPTIONS } from './fraudScore';

export const NODE_CONDITION_FIELDS = [
  { value: 'name', label: 'Remark', labelKey: 'nodeConditions.fields.name' },
  { value: 'link_name', label: 'Original name', labelKey: 'nodeConditions.fields.linkName' },
  { value: 'link_country', label: 'Country code', labelKey: 'nodeConditions.fields.linkCountry' },
  { value: 'protocol', label: 'Protocol type', labelKey: 'nodeConditions.fields.protocol' },
  { value: 'source', label: 'Source', labelKey: 'nodeConditions.fields.source' },
  { value: 'group', label: 'Group', labelKey: 'nodeConditions.fields.group' },
  { value: 'speed', label: 'Speed (MB/s)', labelKey: 'nodeConditions.fields.speed' },
  { value: 'delay_time', label: 'Delay (ms)', labelKey: 'nodeConditions.fields.delayTime' },
  { value: 'fraud_score', label: 'Fraud score', labelKey: 'nodeConditions.fields.fraudScore' },
  { value: 'quality_status', label: 'Quality status', labelKey: 'nodeConditions.fields.qualityStatus' },
  { value: 'unlock_condition', label: 'Unlock condition', labelKey: 'nodeConditions.fields.unlockCondition', isVirtual: true },
  { value: 'unlock_provider', label: 'Unlock provider', labelKey: 'nodeConditions.fields.unlockProvider', hidden: true },
  { value: 'unlock_status', label: 'Unlock status', labelKey: 'nodeConditions.fields.unlockStatus', hidden: true },
  { value: 'unlock_keyword', label: 'Unlock keyword', labelKey: 'nodeConditions.fields.unlockKeyword', hidden: true },
  { value: 'unlock_result', label: 'Unlock summary', labelKey: 'nodeConditions.fields.unlockResult' },
  { value: 'ip_type', label: 'IP type', labelKey: 'nodeConditions.fields.ipType' },
  { value: 'residential_type', label: 'Residential attribute', labelKey: 'nodeConditions.fields.residentialType' },
  { value: 'speed_status', label: 'Speed status', labelKey: 'nodeConditions.fields.speedStatus' },
  { value: 'delay_status', label: 'Delay status', labelKey: 'nodeConditions.fields.delayStatus' },
  { value: 'link_address', label: 'Address', labelKey: 'nodeConditions.fields.linkAddress' },
  { value: 'link_host', label: 'Host' },
  { value: 'link_port', label: 'Port', labelKey: 'nodeConditions.fields.linkPort' },
  { value: 'dialer_proxy_name', label: 'Dialer proxy', labelKey: 'nodeConditions.fields.dialerProxyName' },
  { value: 'link', label: 'Node link', labelKey: 'nodeConditions.fields.link' }
];

const STATIC_FIELD_META_MAP = NODE_CONDITION_FIELDS.reduce((acc, field) => {
  acc[field.value] = field;
  return acc;
}, {});

export const UNLOCK_STATUS_OPTIONS = () => getUnlockStatusOptions(false);

export const UNLOCK_PROVIDER_OPTIONS = () => getUnlockProviderOptions();

export const NODE_STATUS_OPTIONS = [
  { value: 'untested', label: 'Untested', labelKey: 'nodeConditions.status.untested' },
  { value: 'success', label: 'Success', labelKey: 'nodeConditions.status.success' },
  { value: 'timeout', label: 'Timeout', labelKey: 'nodeConditions.status.timeout' },
  { value: 'error', label: 'Failed', labelKey: 'nodeConditions.status.error' }
];

export const NODE_IP_TYPE_OPTIONS = [
  { value: 'native', label: 'Native IP', labelKey: 'nodeConditions.ipType.native' },
  { value: 'broadcast', label: 'Broadcast IP', labelKey: 'nodeConditions.ipType.broadcast' },
  { value: 'untested', label: 'Untested', labelKey: 'nodeConditions.ipType.untested' }
];

export const NODE_RESIDENTIAL_TYPE_OPTIONS = [
  { value: 'residential', label: 'Residential IP', labelKey: 'nodeConditions.residential.residential' },
  { value: 'datacenter', label: 'Datacenter IP', labelKey: 'nodeConditions.residential.datacenter' },
  { value: 'untested', label: 'Untested', labelKey: 'nodeConditions.residential.untested' }
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

  if (field === 'unlock_condition') {
    return {
      ...STATIC_FIELD_META_MAP[field],
      dataType: 'composite',
      inputType: 'unlock_composite',
      isVirtual: true,
      operators: ['equals']
    };
  }

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
  const baseFields = dynamicFields.length > 0 ? dynamicFields : NODE_CONDITION_FIELDS;

  // 确保包含 unlock_condition 虚拟字段，并过滤掉被标记为 hidden 的独立解锁字段
  const hasUnlockCondition = baseFields.some((f) => f.value === 'unlock_condition');

  if (!hasUnlockCondition) {
    // 找到 quality_status 的位置，在其后插入 unlock_condition
    const qualityStatusIndex = baseFields.findIndex((f) => f.value === 'quality_status');
    const unlockConditionField = NODE_CONDITION_FIELDS.find((f) => f.value === 'unlock_condition');

    if (qualityStatusIndex >= 0 && unlockConditionField) {
      const fieldsWithVirtual = [
        ...baseFields.slice(0, qualityStatusIndex + 1),
        unlockConditionField,
        ...baseFields.slice(qualityStatusIndex + 1)
      ];
      return fieldsWithVirtual;
    }
  }

  return baseFields;
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
