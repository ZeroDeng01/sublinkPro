import request from './request';

/**
 * 获取订阅的所有分享列表
 * @param {number} subId 订阅ID
 * @param {number} page 页码（可选，用于分页）
 * @param {number} pageSize 每页数量（可选，用于分页）
 * @param {string} keyword 搜索关键词（可选）
 * @param {string} ipFilter IP地址筛选（可选）
 * @param {string} sortBy 排序字段（可选，如: access_count）
 * @param {string} sortOrder 排序方向（可选，asc/desc）
 */
export function getShares(subId, page, pageSize, keyword, ipFilter, sortBy, sortOrder) {
  const params = { subId };
  if (page !== undefined && pageSize !== undefined) {
    params.page = page;
    params.pageSize = pageSize;
  }
  if (keyword && keyword.trim()) {
    params.keyword = keyword.trim();
  }
  if (ipFilter && ipFilter.trim()) {
    params.ipFilter = ipFilter.trim();
  }
  if (sortBy) {
    params.sortBy = sortBy;
  }
  if (sortOrder) {
    params.sortOrder = sortOrder;
  }
  return request({
    url: '/v1/shares/get',
    method: 'get',
    params
  });
}

/**
 * 创建新分享
 * @param {object} data 分享数据 { subscription_id, name, token?, expire_type, expire_days?, expire_at? }
 */
export function createShare(data) {
  return request({
    url: '/v1/shares/add',
    method: 'post',
    data
  });
}

/**
 * 更新分享设置
 * @param {object} data 分享数据 { id, name, token, expire_type, expire_days?, expire_at?, enabled }
 */
export function updateShare(data) {
  return request({
    url: '/v1/shares/update',
    method: 'post',
    data
  });
}

/**
 * 删除分享
 * @param {number} id 分享ID
 */
export function deleteShare(id) {
  return request({
    url: '/v1/shares/delete',
    method: 'delete',
    params: { id }
  });
}

/**
 * 获取分享的访问日志
 * @param {number} shareId 分享ID
 */
export function getShareLogs(shareId) {
  return request({
    url: '/v1/shares/logs',
    method: 'get',
    params: { shareId }
  });
}

/**
 * 刷新分享Token
 * @param {number} id 分享ID
 */
export function refreshShareToken(id) {
  return request({
    url: '/v1/shares/refresh',
    method: 'post',
    params: { id }
  });
}

/**
 * 批量创建分享
 * @param {number} subscriptionId 订阅ID
 * @param {object} data { baseName, count, expireType, expireDays?, expireAt?, enabled }
 */
export function batchCreateShares(subscriptionId, data) {
  return request({
    url: '/v1/shares/batch-add',
    method: 'post',
    data: {
      subscription_id: subscriptionId,
      base_name: data.baseName,
      count: data.count,
      expire_type: data.expireType,
      expire_days: data.expireDays,
      expire_at: data.expireAt,
      enabled: data.enabled
    }
  });
}

/**
 * 批量删除分享
 * @param {number[]} ids 分享ID数组
 */
export function batchDeleteShares(ids) {
  return request({
    url: '/v1/shares/batch-delete',
    method: 'post',
    data: { ids }
  });
}

/**
 * 批量更新分享
 * @param {number[]} ids 分享ID数组
 * @param {object} updates 更新字段 { enabled?, expire_type?, expire_days?, expire_at? }
 */
export function batchUpdateShares(ids, updates) {
  return request({
    url: '/v1/shares/batch-update',
    method: 'post',
    data: {
      ids,
      ...updates
    }
  });
}
