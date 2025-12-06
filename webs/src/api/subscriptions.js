import request from './request';

// 获取订阅列表
export function getSubscriptions() {
  return request({
    url: '/v1/subcription/get',
    method: 'get'
  }).then((response) => {
    // 确保每个订阅都有 Nodes 数组
    if (response.data && Array.isArray(response.data)) {
      response.data.forEach((sub) => {
        if (!sub.Nodes || sub.Nodes.length === 0) {
          sub.Nodes = [];
        }
      });
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
