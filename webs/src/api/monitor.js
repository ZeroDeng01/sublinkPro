import request from './request';

// 获取系统监控统计
export function getSystemStats() {
  return request({
    url: '/v1/total/system-stats',
    method: 'get'
  });
}
