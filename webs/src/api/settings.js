import request from './request';

// 获取 Webhook 配置
export function getWebhookConfig() {
  return request({
    url: '/v1/settings/webhook',
    method: 'get'
  });
}

// 保存 Webhook 配置
export function updateWebhookConfig(data) {
  return request({
    url: '/v1/settings/webhook',
    method: 'post',
    data
  });
}

// 测试 Webhook
export function testWebhook(data) {
  return request({
    url: '/v1/settings/webhook/test',
    method: 'post',
    data
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
