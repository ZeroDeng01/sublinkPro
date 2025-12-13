import request from './request';

// 获取任务列表
export function getTasks(params) {
  return request({
    url: '/v1/tasks',
    method: 'get',
    params
  });
}

// 获取任务详情
export function getTask(id) {
  return request({
    url: `/v1/tasks/${id}`,
    method: 'get'
  });
}

// 停止任务
export function stopTask(id) {
  return request({
    url: `/v1/tasks/${id}/stop`,
    method: 'post'
  });
}

// 获取任务统计
export function getTaskStats() {
  return request({
    url: '/v1/tasks/stats',
    method: 'get'
  });
}

// 获取运行中的任务
export function getRunningTasks() {
  return request({
    url: '/v1/tasks/running',
    method: 'get'
  });
}

// 清理任务历史
export function clearTaskHistory(params) {
  return request({
    url: '/v1/tasks',
    method: 'delete',
    data: params
  });
}

// 获取任务流量明细（支持分组/来源过滤、搜索、分页）
export function getTaskTrafficDetails(taskId, params) {
  return request({
    url: `/v1/tasks/${taskId}/traffic`,
    method: 'get',
    params
  });
}
