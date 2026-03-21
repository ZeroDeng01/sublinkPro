import request from './request';

function appendIfPresent(formData, key, value) {
  if (value !== undefined && value !== null && value !== '') {
    formData.append(key, value);
  }
}

function buildLoginFormData(data) {
  const formData = new FormData();
  formData.append('username', data.username);
  formData.append('password', data.password);
  formData.append('captchaKey', data.captchaKey || '');
  formData.append('captchaCode', data.captchaCode || '');
  formData.append('rememberMe', data.rememberMe ? 'true' : 'false');
  appendIfPresent(formData, 'turnstileToken', data.turnstileToken);

  return formData;
}

// 获取验证码
export function getCaptcha() {
  return request({
    url: '/v1/auth/captcha',
    method: 'get'
  });
}

// 登录
export function login(data) {
  return request({
    url: '/v1/auth/login',
    method: 'post',
    data: buildLoginFormData(data),
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  });
}

export function verifyMfaLogin(data) {
  const isRecoveryCode = data.type === 'recovery_code';
  return request({
    url: isRecoveryCode ? '/v1/auth/mfa/verify-recovery-code' : '/v1/auth/mfa/verify-totp',
    method: 'post',
    data: isRecoveryCode
      ? {
          challengeToken: data.challengeToken,
          recoveryCode: data.recoveryCode
        }
      : {
          challengeToken: data.challengeToken,
          code: data.code
        }
  });
}

// 登出
export function logout() {
  return request({
    url: '/v1/auth/logout',
    method: 'delete'
  });
}

// 获取用户信息
export function getUserInfo() {
  return request({
    url: '/v1/users/me',
    method: 'get'
  });
}

// 更新用户名和密码
export function updateUserPassword(data) {
  const params = new URLSearchParams();
  params.append('username', data.username);
  params.append('password', data.password);

  return request({
    url: '/v1/users/update',
    method: 'post',
    data: params,
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded'
    }
  });
}

export function getTotpStatus() {
  return request({
    url: '/v1/users/mfa',
    method: 'get'
  });
}

export function setupTotp(data = {}) {
  return request({
    url: '/v1/users/mfa/totp/begin',
    method: 'post',
    data
  });
}

export function confirmTotpSetup(data) {
  return request({
    url: '/v1/users/mfa/totp/confirm',
    method: 'post',
    data
  });
}

export function disableTotp(data = {}) {
  return request({
    url: '/v1/users/mfa/totp/disable',
    method: 'post',
    data
  });
}

export function regenerateRecoveryCodes(data = {}) {
  return request({
    url: '/v1/users/mfa/recovery-codes/regenerate',
    method: 'post',
    data
  });
}
