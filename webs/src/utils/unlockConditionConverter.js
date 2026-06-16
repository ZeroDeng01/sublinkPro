/**
 * 解锁条件转换工具
 *
 * 用于在UI层的"解锁情况"虚拟字段和后端存储的独立字段之间进行转换
 */

/**
 * 解锁条件的子字段类型
 */
const UNLOCK_FIELDS = {
  PROVIDER: 'unlock_provider',
  STATUS: 'unlock_status',
  KEYWORD: 'unlock_keyword',
  RESULT: 'unlock_result'
};

/**
 * 解锁情况虚拟字段名
 */
export const UNLOCK_CONDITION_FIELD = 'unlock_condition';

/**
 * 检测给定索引位置是否是解锁条件组的开始
 *
 * @param {Array} conditions - 条件数组
 * @param {number} index - 检查的索引
 * @returns {boolean} 是否是解锁条件组
 */
export function isUnlockConditionGroup(conditions, index) {
  if (!conditions || index >= conditions.length) {
    return false;
  }

  const cond = conditions[index];
  const unlockFields = [UNLOCK_FIELDS.PROVIDER, UNLOCK_FIELDS.STATUS, UNLOCK_FIELDS.KEYWORD, UNLOCK_FIELDS.RESULT];

  return unlockFields.includes(cond.field);
}

/**
 * 从后端格式转换为UI格式（编辑时使用）
 * 将连续的解锁相关字段合并为一个"解锁情况"虚拟条件
 *
 * @param {Array} backendConditions - 后端条件数组
 * @returns {Array} UI条件数组
 */
export function convertBackendToUI(backendConditions) {
  if (!Array.isArray(backendConditions) || backendConditions.length === 0) {
    return backendConditions;
  }

  const uiConditions = [];
  let i = 0;

  while (i < backendConditions.length) {
    const cond = backendConditions[i];

    // 检测是否是解锁相关字段
    if (isUnlockConditionGroup(backendConditions, i)) {
      // 尝试收集连续的解锁字段
      const unlockGroup = collectUnlockGroup(backendConditions, i);

      if (unlockGroup.count > 0) {
        // 转换为虚拟的"解锁情况"条件
        uiConditions.push({
          field: UNLOCK_CONDITION_FIELD,
          operator: 'equals', // 虚拟操作符，实际不使用
          value: unlockGroup.value
        });
        i += unlockGroup.count;
        continue;
      }
    }

    // 非解锁字段，直接保留
    uiConditions.push(cond);
    i++;
  }

  return uiConditions;
}

/**
 * 收集从指定位置开始的连续解锁字段
 *
 * @param {Array} conditions - 条件数组
 * @param {number} startIndex - 开始索引
 * @returns {Object} { value: 组合值对象, count: 消费的条件数量 }
 */
function collectUnlockGroup(conditions, startIndex) {
  const result = {
    provider: '',
    status: '',
    keyword: '',
    providerOperator: 'equals',
    statusOperator: 'equals',
    keywordOperator: 'contains'
  };
  let count = 0;
  const seenFields = new Set();

  // 最多向前看3个条件，遇到重复字段或非解锁字段时停止
  for (let i = startIndex; i < Math.min(startIndex + 3, conditions.length); i++) {
    const cond = conditions[i];

    if (cond.field === UNLOCK_FIELDS.PROVIDER) {
      if (seenFields.has(UNLOCK_FIELDS.PROVIDER)) break; // 遇到重复provider，说明是下一组
      result.provider = cond.value || '';
      result.providerOperator = cond.operator || 'equals';
      seenFields.add(UNLOCK_FIELDS.PROVIDER);
      count++;
    } else if (cond.field === UNLOCK_FIELDS.STATUS) {
      if (seenFields.has(UNLOCK_FIELDS.STATUS)) break; // 遇到重复status，说明是下一组
      result.status = cond.value || '';
      result.statusOperator = cond.operator || 'equals';
      seenFields.add(UNLOCK_FIELDS.STATUS);
      count++;
    } else if (cond.field === UNLOCK_FIELDS.KEYWORD) {
      if (seenFields.has(UNLOCK_FIELDS.KEYWORD)) break; // 遇到重复keyword，说明是下一组
      result.keyword = cond.value || '';
      result.keywordOperator = cond.operator || 'contains';
      seenFields.add(UNLOCK_FIELDS.KEYWORD);
      count++;
    } else if (cond.field === UNLOCK_FIELDS.RESULT) {
      // unlock_result 字段暂不处理，保持独立
      break;
    } else {
      // 遇到非解锁字段，停止收集
      break;
    }
  }

  return { value: result, count };
}

/**
 * 从UI格式转换为后端格式（保存时使用）
 * 将"解锁情况"虚拟条件展开为独立的后端字段
 *
 * @param {Array} uiConditions - UI条件数组
 * @returns {Array} 后端条件数组
 */
export function convertUIToBackend(uiConditions) {
  if (!Array.isArray(uiConditions) || uiConditions.length === 0) {
    return uiConditions;
  }

  const backendConditions = [];

  for (const cond of uiConditions) {
    if (cond.field === UNLOCK_CONDITION_FIELD) {
      // 展开解锁情况条件
      const expanded = expandUnlockCondition(cond);
      backendConditions.push(...expanded);
    } else {
      // 非解锁情况，直接保留
      backendConditions.push(cond);
    }
  }

  return backendConditions;
}

/**
 * 展开一个"解锁情况"虚拟条件为多个后端条件
 *
 * @param {Object} unlockCondition - 解锁情况条件
 * @returns {Array} 展开后的条件数组
 */
function expandUnlockCondition(unlockCondition) {
  const expanded = [];
  const value = unlockCondition.value || {};

  // Provider 条件
  if (value.provider && value.provider.trim() !== '') {
    expanded.push({
      field: UNLOCK_FIELDS.PROVIDER,
      operator: value.providerOperator || 'equals',
      value: value.provider
    });
  }

  // Status 条件
  if (value.status && value.status.trim() !== '') {
    expanded.push({
      field: UNLOCK_FIELDS.STATUS,
      operator: value.statusOperator || 'equals',
      value: value.status
    });
  }

  // Keyword 条件（可选）
  if (value.keyword && value.keyword.trim() !== '') {
    expanded.push({
      field: UNLOCK_FIELDS.KEYWORD,
      operator: value.keywordOperator || 'contains',
      value: value.keyword
    });
  }

  return expanded;
}

/**
 * 验证解锁情况值是否有效（至少配置了一项）
 *
 * @param {Object} unlockValue - 解锁情况值对象
 * @returns {boolean} 是否有效
 */
export function isValidUnlockValue(unlockValue) {
  if (!unlockValue) return false;

  return Boolean(
    (unlockValue.provider && unlockValue.provider.trim()) ||
    (unlockValue.status && unlockValue.status.trim()) ||
    (unlockValue.keyword && unlockValue.keyword.trim())
  );
}
