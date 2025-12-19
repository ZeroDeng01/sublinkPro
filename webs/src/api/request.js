import axios from 'axios';

// 创建 axios 实例
const request = axios.create({
  baseURL: '/api',
  timeout: 30000, // 请求超时时间
  headers: {
    'Content-Type': 'application/json'
  }
});

// 请求拦截器 - 添加 token
request.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('accessToken');
    if (token) {
      config.headers.Authorization = token;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器 - 处理错误
request.interceptors.response.use(
  (response) => {
    const data = response.data;
    // 检查业务逻辑错误码（后端返回 code 非 200 表示业务错误）
    if (data && typeof data.code === 'number' && data.code !== 200) {
      // 构造业务错误对象并 reject，让调用方的 catch 能够捕获
      const error = new Error(data.msg || '操作失败');
      error.response = response;
      error.code = data.code;
      error.data = data;
      error.isBusinessError = true; // 标记为业务错误，便于区分
      return Promise.reject(error);
    }
    return data;
  },
  (error) => {
    if (error.response) {
      const { status } = error.response;

      // 401 未授权 - 清除 token 并跳转登录
      if (status === 401) {
        localStorage.removeItem('accessToken');
        window.location.href = '/login';
      }

      // 403 禁止访问
      if (status === 403) {
        console.error('没有权限访问该资源');
      }

      // 500 服务器错误
      if (status >= 500) {
        console.error('服务器错误');
      }
    }

    return Promise.reject(error);
  }
);

export default request;
