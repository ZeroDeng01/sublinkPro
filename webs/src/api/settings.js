import request from './request';

export function getWebhooks() {
  return request({
    url: '/v1/settings/webhooks',
    method: 'get'
  });
}

export function createWebhook(data) {
  return request({
    url: '/v1/settings/webhooks',
    method: 'post',
    data
  });
}

export function updateWebhook(id, data) {
  return request({
    url: `/v1/settings/webhooks/${id}`,
    method: 'put',
    data
  });
}

export function deleteWebhook(id) {
  return request({
    url: `/v1/settings/webhooks/${id}`,
    method: 'delete'
  });
}

export function testWebhookById(id) {
  return request({
    url: `/v1/settings/webhooks/${id}/test`,
    method: 'post'
  });
}

// 获取基础模板配置
export function getBaseTemplates() {
  return request({
    url: '/v1/settings/base-templates',
    method: 'get'
  });
}

// 更新基础模板配置
export function updateBaseTemplate(category, content) {
  return request({
    url: '/v1/settings/base-templates',
    method: 'post',
    data: { category, content }
  });
}

// 获取系统域名配置
export function getSystemDomain() {
  return request({
    url: '/v1/settings/system-domain',
    method: 'get'
  });
}

// 保存系统域名配置
export function updateSystemDomain(data) {
  return request({
    url: '/v1/settings/system-domain',
    method: 'post',
    data
  });
}

// 获取节点去重配置
export function getNodeDedupConfig() {
  return request({
    url: '/v1/settings/node-dedup',
    method: 'get'
  });
}

// 保存节点去重配置
export function updateNodeDedupConfig(data) {
  return request({
    url: '/v1/settings/node-dedup',
    method: 'post',
    data
  });
}

// 获取全局节点处理配置
export function getGlobalNodeProcessingConfig() {
  return request({
    url: '/v1/settings/global-node-processing',
    method: 'get'
  });
}

// 保存全局节点处理配置
export function updateGlobalNodeProcessingConfig(data) {
  return request({
    url: '/v1/settings/global-node-processing',
    method: 'post',
    data
  });
}

export function getAISettings() {
  return request({
    url: '/v1/settings/ai-assistant',
    method: 'get'
  });
}

export function listAIModels(data) {
  return request({
    url: '/v1/settings/ai-assistant/models',
    method: 'post',
    data
  });
}

export function updateAISettings(data) {
  return request({
    url: '/v1/settings/ai-assistant',
    method: 'post',
    data
  });
}

export function testAISettings(data) {
  return request({
    url: '/v1/settings/ai-assistant/test',
    method: 'post',
    data
  });
}

export function getCloudflaredStatus() {
  return request({
    url: '/v1/settings/cloudflared',
    method: 'get'
  });
}

export function updateCloudflaredConfig(data) {
  return request({
    url: '/v1/settings/cloudflared',
    method: 'post',
    data
  });
}

export function startCloudflared(data) {
  return request({
    url: '/v1/settings/cloudflared/start',
    method: 'post',
    data
  });
}

export function stopCloudflared() {
  return request({
    url: '/v1/settings/cloudflared/stop',
    method: 'post'
  });
}

export function removeCloudflaredToken() {
  return request({
    url: '/v1/settings/cloudflared/token',
    method: 'delete'
  });
}

export function getSubStoreSettings() {
  return request({
    url: '/v1/settings/substore',
    method: 'get'
  });
}

export function updateSubStoreSettings(data) {
  return request({
    url: '/v1/settings/substore',
    method: 'post',
    data
  });
}

export function testSubStoreSettings(data) {
  return request({
    url: '/v1/settings/substore/test',
    method: 'post',
    data
  });
}

// 导入 SQLite 备份/数据库
export function importDatabaseMigration(formData) {
  return request({
    url: '/v1/settings/database-migration/import',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  });
}
