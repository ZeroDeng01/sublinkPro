import request from './request';

// 获取订阅调度器列表
export function getSubSchedulers() {
  return request({
    url: '/v1/sub_scheduler/get',
    method: 'get'
  });
}

// 添加订阅调度器
export function addSubScheduler(data) {
  return request({
    url: '/v1/sub_scheduler/add',
    method: 'post',
    data
  });
}

// 更新订阅调度器
export function updateSubScheduler(data) {
  return request({
    url: '/v1/sub_scheduler/update',
    method: 'put',
    data
  });
}

// 立即拉取订阅
export function pullSubScheduler(data) {
  return request({
    url: '/v1/sub_scheduler/pull',
    method: 'post',
    data
  });
}

// 删除订阅调度器
export function deleteSubScheduler(id) {
  return request({
    url: `/v1/sub_scheduler/delete/${id}`,
    method: 'delete'
  });
}
