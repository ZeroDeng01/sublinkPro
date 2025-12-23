import request from './request';

// 获取节点列表（支持过滤和分页参数）
// params: { search, group, source, maxDelay, minSpeed, countries[], sortBy, sortOrder, page, pageSize }
// 带page/pageSize时返回 { items, total, page, pageSize, totalPages }
// 不带分页参数时返回节点数组（向后兼容）
export function getNodes(params = {}) {
  return request({
    url: '/v1/nodes/get',
    method: 'get',
    params
  });
}

// 获取符合过滤条件的所有节点ID（用于全选操作）
export function getNodeIds(params = {}) {
  return request({
    url: '/v1/nodes/ids',
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

// 获取节点检测策略列表（替代原测速配置）
export function getSpeedTestConfig() {
  return request({
    url: '/v1/node-check/profiles',
    method: 'get'
  });
}

// 运行测速（向后兼容，使用新的node-check API）
export function runSpeedTest(ids, profileId) {
  return request({
    url: '/v1/node-check/run',
    method: 'post',
    data: {
      nodeIds: ids || [],
      profileId: profileId || 0
    }
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
    url: '/v1/nodes/batch-update-group',
    method: 'post',
    data: { ids, group }
  });
}

// 批量更新节点前置代理
export function batchUpdateNodeDialerProxy(ids, dialerProxyName) {
  return request({
    url: '/v1/nodes/batch-update-dialer-proxy',
    method: 'post',
    data: { ids, dialerProxyName }
  });
}

// 批量更新节点来源
export function batchUpdateNodeSource(ids, source) {
  return request({
    url: '/v1/nodes/batch-update-source',
    method: 'post',
    data: { ids, source }
  });
}

// 获取IP详细信息
export function getIPDetails(ip) {
  return request({
    url: '/v1/nodes/ip-info',
    method: 'get',
    params: { ip }
  });
}

// 获取IP缓存统计
export function getIPCacheStats() {
  return request({
    url: '/v1/nodes/ip-cache/stats',
    method: 'get'
  });
}

// 清除所有IP缓存
export function clearIPCache() {
  return request({
    url: '/v1/nodes/ip-cache',
    method: 'delete'
  });
}

// 获取协议 UI 元数据（包含颜色、图标、字段信息）
export function getProtocolUIMeta() {
  return request({
    url: '/v1/nodes/protocol-ui-meta',
    method: 'get'
  });
}

// 解析节点链接
export function parseNodeLink(link) {
  return request({
    url: '/v1/nodes/parse-link',
    method: 'get',
    params: { link }
  });
}

// 获取节点原始信息
export function getNodeRawInfo(id) {
  return request({
    url: '/v1/nodes/raw-info',
    method: 'get',
    params: { id }
  });
}

// 更新节点原始信息
export function updateNodeRawInfo(nodeId, fields) {
  return request({
    url: '/v1/nodes/update-raw',
    method: 'post',
    data: { nodeId, fields }
  });
}

// 获取所有使用中的协议类型列表（用于过滤器选项）
export function getNodeProtocols() {
  return request({
    url: '/v1/nodes/protocols',
    method: 'get'
  });
}
