import request from './request';

/**
 * 国家规则 API
 */

// 获取国家规则列表
export const getCountryRules = () => {
  return request({
    url: '/v1/country-rules',
    method: 'get'
  });
};

// 创建国家规则
export const createCountryRule = (data) => {
  return request({
    url: '/v1/country-rules',
    method: 'post',
    data
  });
};

// 更新国家规则
export const updateCountryRule = (id, data) => {
  return request({
    url: `/v1/country-rules/${id}`,
    method: 'put',
    data
  });
};

// 删除国家规则
export const deleteCountryRule = (id) => {
  return request({
    url: `/v1/country-rules/${id}`,
    method: 'delete'
  });
};

// 测试国家规则
export const testCountryRule = (data) => {
  return request({
    url: '/v1/country-rules/test',
    method: 'post',
    data
  });
};

// 批量测试节点名称
export const batchTestCountryRules = (nodeNames) => {
  return request({
    url: '/v1/country-rules/batch-test',
    method: 'post',
    data: { nodeNames }
  });
};

// 批量操作规则（导入/覆盖）
export const batchCountryRules = (data) => {
  return request({
    url: '/v1/country-rules/batch',
    method: 'post',
    data
  });
};

// 导出国家规则为文本格式
export const exportCountryRules = () => {
  return request({
    url: '/v1/country-rules/export',
    method: 'get'
  });
};

// 从文本同步国家规则
export const syncCountryRules = (text) => {
  return request({
    url: '/v1/country-rules/sync',
    method: 'post',
    data: { text }
  });
};
