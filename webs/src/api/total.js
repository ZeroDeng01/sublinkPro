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
