import request from './request';

// 获取机场列表
export function getAirports(params = {}) {
  return request({
    url: '/v1/airports',
    method: 'get',
    params
  });
}

// 获取单个机场
export function getAirport(id) {
  return request({
    url: `/v1/airports/${id}`,
    method: 'get'
  });
}

// 添加机场
export function addAirport(data) {
  return request({
    url: '/v1/airports',
    method: 'post',
    data
  });
}

// 更新机场
export function updateAirport(id, data) {
  return request({
    url: `/v1/airports/${id}`,
    method: 'put',
    data
  });
}

// 批量更新机场
export function batchUpdateAirports(data) {
  return request({
    url: '/v1/airports/batch-update',
    method: 'post',
    data
  });
}

// 删除机场
export function deleteAirport(id, deleteNodes = false) {
  return request({
    url: `/v1/airports/${id}`,
    method: 'delete',
    params: { deleteNodes }
  });
}

// 拉取机场订阅
export function pullAirport(id) {
  return request({
    url: `/v1/airports/${id}/pull`,
    method: 'post'
  });
}

// 批量拉取机场订阅
// @param {number[]} ids - 可选，要拉取的机场ID列表。如果不提供或为空，则拉取所有已启用的机场
export function pullAllAirports(ids) {
  return request({
    url: '/v1/airports/pull-all',
    method: 'post',
    data: ids && ids.length > 0 ? { ids } : undefined
  });
}

// 仅刷新机场用量信息
export function refreshAirportUsage(id) {
  return request({
    url: `/v1/airports/${id}/refresh-usage`,
    method: 'post'
  });
}
