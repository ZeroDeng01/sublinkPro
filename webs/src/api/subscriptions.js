import request from './request';

// 获取订阅列表（支持分页参数）
// params: { page, pageSize }
// 带page/pageSize时返回 { items, total, page, pageSize, totalPages }
export function getSubscriptions(params = {}) {
  return request({
    url: '/v1/subcription/get',
    method: 'get',
    params
  }).then((response) => {
    // 处理分页响应
    const data = response.data?.items || response.data;
    // 确保每个订阅都有 Nodes 数组
    if (data && Array.isArray(data)) {
      data.forEach((sub) => {
        if (!sub.Nodes || sub.Nodes.length === 0) {
          sub.Nodes = [];
        }
      });
    }
    // 如果是分页响应，保持原结构
    if (response.data?.items !== undefined) {
      response.data.items = data;
    } else {
      response.data = data;
    }
    return response;
  });
}

// 添加订阅
export function addSubscription(data) {
  const formData = new FormData();
  Object.keys(data).forEach((key) => {
    if (data[key] !== undefined && data[key] !== null) {
      formData.append(key, data[key]);
    }
  });
  return request({
    url: '/v1/subcription/add',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  });
}

// 更新订阅
export function updateSubscription(data) {
  const formData = new FormData();
  Object.keys(data).forEach((key) => {
    if (data[key] !== undefined && data[key] !== null) {
      formData.append(key, data[key]);
    }
  });
  return request({
    url: '/v1/subcription/update',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  });
}

// 删除订阅
export function deleteSubscription(data) {
  return request({
    url: '/v1/subcription/delete',
    method: 'delete',
    params: data
  });
}

// 订阅排序
export function sortSubscription(data) {
  return request({
    url: '/v1/subcription/sort',
    method: 'post',
    data,
    headers: {
      'Content-Type': 'application/json'
    }
  });
}

// 批量排序订阅节点
export function batchSortSubscription(data) {
  return request({
    url: '/v1/subcription/batch-sort',
    method: 'post',
    data,
    headers: {
      'Content-Type': 'application/json'
    }
  });
}

// 复制订阅
export function copySubscription(id) {
  return request({
    url: '/v1/subcription/copy',
    method: 'post',
    params: { id }
  });
}

// 预览订阅节点
// data: { Nodes, Groups, DelayTime, MinSpeed, CountryWhitelist, ... }
export function previewSubscriptionNodes(data) {
  return request({
    url: '/v1/subcription/preview',
    method: 'post',
    data,
    headers: {
      'Content-Type': 'application/json'
    }
  });
}

// 获取协议元数据（协议列表及其可用字段）
export function getProtocolMeta() {
  return request({
    url: '/v1/subcription/protocol-meta',
    method: 'get'
  });
}

// 获取节点通用字段元数据
export function getNodeFieldsMeta() {
  return request({
    url: '/v1/subcription/node-fields-meta',
    method: 'get'
  });
}

// ========== 链式代理规则 API ==========

// 获取链式代理规则列表
export function getChainRules(subId) {
  return request({
    url: `/v1/subcription/${subId}/chain-rules`,
    method: 'get'
  });
}

// 创建链式代理规则
export function createChainRule(subId, data) {
  return request({
    url: `/v1/subcription/${subId}/chain-rules`,
    method: 'post',
    data,
    headers: {
      'Content-Type': 'application/json'
    }
  });
}

// 更新链式代理规则
export function updateChainRule(subId, ruleId, data) {
  return request({
    url: `/v1/subcription/${subId}/chain-rules/${ruleId}`,
    method: 'put',
    data,
    headers: {
      'Content-Type': 'application/json'
    }
  });
}

// 删除链式代理规则
export function deleteChainRule(subId, ruleId) {
  return request({
    url: `/v1/subcription/${subId}/chain-rules/${ruleId}`,
    method: 'delete'
  });
}

// 切换链式代理规则启用状态
export function toggleChainRule(subId, ruleId) {
  return request({
    url: `/v1/subcription/${subId}/chain-rules/${ruleId}/toggle`,
    method: 'put'
  });
}

// 批量排序链式代理规则
export function sortChainRules(subId, ruleIds) {
  return request({
    url: `/v1/subcription/${subId}/chain-rules/sort`,
    method: 'put',
    data: { ruleIds },
    headers: {
      'Content-Type': 'application/json'
    }
  });
}

// 获取链式代理可用选项
export function getChainOptions(subId) {
  return request({
    url: `/v1/subcription/${subId}/chain-options`,
    method: 'get'
  });
}

// 预览订阅的整体链式代理配置
export function previewChainLinks(subId) {
  return request({
    url: `/v1/subcription/${subId}/chain-rules/preview`,
    method: 'get'
  });
}
