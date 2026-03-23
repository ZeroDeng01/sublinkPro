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

export function getDashboardCountryStats() {
  return request({
    url: '/v1/total/dashboard-country-stats',
    method: 'get'
  });
}

export function getDashboardGroupedStats() {
  return request({
    url: '/v1/total/dashboard-grouped-stats',
    method: 'get'
  });
}

export function getQualityStats() {
  return request({
    url: '/v1/total/quality-stats',
    method: 'get'
  });
}
