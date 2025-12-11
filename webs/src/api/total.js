import request from './request';

// 获取订阅总数
export function getSubTotal() {
  return request({
    url: '/v1/total/sub',
    method: 'get'
  });
}

// 获取节点总数
export function getNodeTotal() {
  return request({
    url: '/v1/total/node',
    method: 'get'
  });
}

// 获取最快速度节点
export function getFastestSpeedNode() {
  return request({
    url: '/v1/total/fastest-speed',
    method: 'get'
  });
}

// 获取最低延迟节点
export function getLowestDelayNode() {
  return request({
    url: '/v1/total/lowest-delay',
    method: 'get'
  });
}

// 获取国家统计
export function getCountryStats() {
  return request({
    url: '/v1/total/country-stats',
    method: 'get'
  });
}

// 获取协议统计
export function getProtocolStats() {
  return request({
    url: '/v1/total/protocol-stats',
    method: 'get'
  });
}

// 获取标签统计
export function getTagStats() {
  return request({
    url: '/v1/total/tag-stats',
    method: 'get'
  });
}

// 获取分组统计
export function getGroupStats() {
  return request({
    url: '/v1/total/group-stats',
    method: 'get'
  });
}

// 获取来源统计
export function getSourceStats() {
  return request({
    url: '/v1/total/source-stats',
    method: 'get'
  });
}
