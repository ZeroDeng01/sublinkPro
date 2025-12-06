import request from './request';

// 获取脚本列表
export function getScripts() {
  return request({
    url: '/v1/script/list',
    method: 'get'
  });
}

// 添加脚本
export function addScript(data) {
  return request({
    url: '/v1/script/add',
    method: 'post',
    data
  });
}

// 更新脚本
export function updateScript(data) {
  return request({
    url: '/v1/script/update',
    method: 'post',
    data
  });
}

// 删除脚本
export function deleteScript(data) {
  return request({
    url: '/v1/script/delete',
    method: 'delete',
    data
  });
}
