import request from './request';

// 获取脚本列表（支持分页参数）
// params: { page, pageSize }
// 带page/pageSize时返回 { items, total, page, pageSize, totalPages }
export function getScripts(params = {}) {
  return request({
    url: '/v1/script/list',
    method: 'get',
    params
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
