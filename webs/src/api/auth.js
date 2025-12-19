import request from './request';

// 获取验证码
export function getCaptcha() {
  return request({
    url: '/v1/auth/captcha',
    method: 'get'
  });
}

// 登录
export function login(data) {
  const formData = new FormData();
  formData.append('username', data.username);
  formData.append('password', data.password);
  formData.append('captchaKey', data.captchaKey || '');
  formData.append('captchaCode', data.captchaCode || '');
  formData.append('rememberMe', data.rememberMe ? 'true' : 'false');
  // Turnstile token（可选）
  if (data.turnstileToken) {
    formData.append('turnstileToken', data.turnstileToken);
  }

  return request({
    url: '/v1/auth/login',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  });
}

// 使用令牌自动登录
export function rememberLogin(rememberToken) {
  const formData = new FormData();
  formData.append('rememberToken', rememberToken);

  return request({
    url: '/v1/auth/remember-login',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  });
}

// 登出
export function logout(rememberToken) {
  const params = rememberToken ? `?rememberToken=${encodeURIComponent(rememberToken)}` : '';
  return request({
    url: `/v1/auth/logout${params}`,
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
