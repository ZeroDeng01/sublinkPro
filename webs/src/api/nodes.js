import request from './request';

// 获取节点列表（支持过滤参数）
export function getNodes(params = {}) {
  return request({
    url: '/v1/nodes/get',
    method: 'get',
    params
  });
}

// 添加节点
export function addNodes(data) {
  const formData = new FormData();
  Object.keys(data).forEach((key) => {
    if (data[key] !== undefined && data[key] !== null) {
      formData.append(key, data[key]);
    }
  });
  return request({
    url: '/v1/nodes/add',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  });
}

// 更新节点
export function updateNode(data) {
  const formData = new FormData();
  Object.keys(data).forEach((key) => {
    if (data[key] !== undefined && data[key] !== null) {
      formData.append(key, data[key]);
    }
  });
  return request({
    url: '/v1/nodes/update',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  });
}

// 删除节点
export function deleteNode(data) {
  return request({
    url: '/v1/nodes/delete',
    method: 'delete',
    params: data
  });
}

// 批量删除节点
export function deleteNodesBatch(ids) {
  return request({
    url: '/v1/nodes/batch-delete',
    method: 'delete',
    data: { ids }
  });
}

// 获取测速配置
export function getSpeedTestConfig() {
  return request({
    url: '/v1/nodes/speed-test/config',
    method: 'get'
  });
}

// 更新测速配置
export function updateSpeedTestConfig(data) {
  return request({
    url: '/v1/nodes/speed-test/config',
    method: 'post',
    data
  });
}

// 运行测速
export function runSpeedTest(ids) {
  return request({
    url: '/v1/nodes/speed-test/run',
    method: 'post',
    data: { ids }
  });
}

// 获取节点国家代码列表
export function getNodeCountries() {
  return request({
    url: '/v1/nodes/countries',
    method: 'get'
  });
}

// 获取节点分组列表
export function getNodeGroups() {
  return request({
    url: '/v1/nodes/groups',
    method: 'get'
  });
}

// 获取节点来源列表
export function getNodeSources() {
  return request({
    url: '/v1/nodes/sources',
    method: 'get'
  });
}

// 批量更新节点分组
export function batchUpdateNodeGroup(ids, group) {
  return request({
    url: "/v1/nodes/batch-update-group",
    method: "post",
    data: { ids, group }
  });
}

// 批量更新节点前置代理
export function batchUpdateNodeDialerProxy(ids, dialerProxyName) {
  return request({
    url: "/v1/nodes/batch-update-dialer-proxy",
    method: "post",
    data: { ids, dialerProxyName }
  });
}
