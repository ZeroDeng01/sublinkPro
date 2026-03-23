import request from './request';

// 获取用户信息
export function getUserInfo() {
  return request({
    url: '/v1/users/me',
    method: 'get'
  });
}

// 修改密码 (需要旧密码验证)
export function changePassword(data) {
  return request({
    url: '/v1/users/change-password',
    method: 'post',
    data
  });
}

// 更新用户资料 (用户名/昵称)
export function updateProfile(data) {
  return request({
    url: '/v1/users/update-profile',
    method: 'post',
    data
  });
}

export function getAISettings() {
  return request({
    url: '/v1/users/ai-settings',
    method: 'get'
  });
}

export function listAIModels(data) {
  return request({
    url: '/v1/users/ai-settings/models',
    method: 'post',
    data
  });
}

export function updateAISettings(data) {
  return request({
    url: '/v1/users/ai-settings',
    method: 'post',
    data
  });
}

export function testAISettings(data) {
  return request({
    url: '/v1/users/ai-settings/test',
    method: 'post',
    data
  });
}
