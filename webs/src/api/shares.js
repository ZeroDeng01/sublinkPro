import request from './request';

/**
 * 获取订阅的所有分享列表
 * @param {number} subId 订阅ID
 */
export function getShares(subId) {
  return request({
    url: '/v1/shares/get',
    method: 'get',
    params: { subId }
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
