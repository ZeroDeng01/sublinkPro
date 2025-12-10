import request from './request';

// 获取 Access Key 列表（支持分页参数）
// params: { page, pageSize }
// 带page/pageSize时返回 { items, total, page, pageSize, totalPages }
export function getAccessKeys(userId, params = {}) {
  return request({
    url: `/v1/accesskey/get/${userId}`,
    method: 'get',
    params
  });
}

// 创建 Access Key
export function createAccessKey(data) {
  return request({
    url: '/v1/accesskey/add',
    method: 'post',
    data
  });
}

// 删除 Access Key
export function deleteAccessKey(id) {
  return request({
    url: `/v1/accesskey/delete/${id}`,
    method: 'delete'
  });
}
