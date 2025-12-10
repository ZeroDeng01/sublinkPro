import request from './request';

// 获取订阅列表（支持分页参数）
// params: { page, pageSize }
// 带page/pageSize时返回 { items, total, page, pageSize, totalPages }
export function getSubscriptions(params = {}) {
  return request({
    url: '/v1/subcription/get',
    method: 'get',
    params
  }).then((response) => {
    // 处理分页响应
    const data = response.data?.items || response.data;
    // 确保每个订阅都有 Nodes 数组
    if (data && Array.isArray(data)) {
      data.forEach((sub) => {
        if (!sub.Nodes || sub.Nodes.length === 0) {
          sub.Nodes = [];
        }
      });
    }
    // 如果是分页响应，保持原结构
    if (response.data?.items !== undefined) {
      response.data.items = data;
    } else {
      response.data = data;
    }
    return response;
  });
}

// 添加订阅
export function addSubscription(data) {
  const formData = new FormData();
  Object.keys(data).forEach((key) => {
    if (data[key] !== undefined && data[key] !== null) {
      formData.append(key, data[key]);
    }
  });
  return request({
    url: '/v1/subcription/add',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  });
}

// 更新订阅
export function updateSubscription(data) {
  const formData = new FormData();
  Object.keys(data).forEach((key) => {
    if (data[key] !== undefined && data[key] !== null) {
      formData.append(key, data[key]);
    }
  });
  return request({
    url: '/v1/subcription/update',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  });
}

// 删除订阅
export function deleteSubscription(data) {
  return request({
    url: '/v1/subcription/delete',
    method: 'delete',
    params: data
  });
}

// 订阅排序
export function sortSubscription(data) {
  return request({
    url: '/v1/subcription/sort',
    method: 'post',
    data,
    headers: {
      'Content-Type': 'application/json'
    }
  });
}
