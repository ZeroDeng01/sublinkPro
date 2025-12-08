import request from './request';

// 获取节点列表
export function getNodes() {
  return request({
    url: '/v1/nodes/get',
    method: 'get'
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
