import request from './request';

// 节点检测策略 API

// 获取所有策略列表
export function getNodeCheckProfiles() {
  return request({
    url: '/v1/node-check/profiles',
    method: 'get'
  });
}

// 获取单个策略
export function getNodeCheckProfile(id) {
  return request({
    url: `/v1/node-check/profiles/${id}`,
    method: 'get'
  });
}

// 创建策略
export function createNodeCheckProfile(data) {
  return request({
    url: '/v1/node-check/profiles',
    method: 'post',
    data
  });
}

// 更新策略
export function updateNodeCheckProfile(id, data) {
  return request({
    url: `/v1/node-check/profiles/${id}`,
    method: 'put',
    data
  });
}

// 删除策略
export function deleteNodeCheckProfile(id) {
  return request({
    url: `/v1/node-check/profiles/${id}`,
    method: 'delete'
  });
}

// 执行节点检测（通用入口）
// profileId: 策略ID（可选）
// nodeIds: 节点ID列表（可选）
export function runNodeCheck(profileId, nodeIds) {
  return request({
    url: '/v1/node-check/run',
    method: 'post',
    data: {
      profileId: profileId || 0,
      nodeIds: nodeIds || []
    }
  });
}

// 使用指定策略执行节点检测
export function runNodeCheckWithProfile(profileId) {
  return request({
    url: `/v1/node-check/profiles/${profileId}/run`,
    method: 'post'
  });
}
