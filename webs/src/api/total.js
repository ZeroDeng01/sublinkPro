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
