import axios, { InternalAxiosRequestConfig, AxiosResponse } from "axios";
import { useUserStoreHook } from "@/store/modules/user";

// 创建 axios 实例
const service = axios.create({
  baseURL: import.meta.env.VITE_APP_BASE_API,
  timeout: 50000,
  headers: { "Content-Type": "application/json;charset=utf-8" },
});

// 请求拦截器
service.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const accessToken = localStorage.getItem("accessToken");
    if (accessToken) {
      config.headers.Authorization = accessToken;
    }
    return config;
  },
  (error: any) => {
    return Promise.reject(error);
  }
);

// 响应拦截器
service.interceptors.response.use(
  (response: AxiosResponse) => {
    // 1. 优先处理二进制数据（Blob 或 ArrayBuffer）
    // 这些响应类型通常用于文件下载，没有 { code, msg } 结构
    if (response.data instanceof Blob || response.data instanceof ArrayBuffer) {
      // 直接返回完整的 response 对象，
      // 这样 .then() 中才能访问 response.headers
      return response;
    }

    // 2. 如果不是二进制，再当作 JSON 处理
    const { code, msg } = response.data;

    // 3. 处理业务成功
    if (code === 200) {
      return response.data;
    }

    // 4. 处理 token 过期
    if (code === 401) {
      ElMessageBox.confirm("当前页面已失效，请重新登录", "提示", {
        confirmButtonText: "确定",
        cancelButtonText: "取消",
        type: "warning",
        lockScroll: false,
      }).then(() => {
        const userStore = useUserStoreHook();
        userStore.resetToken().then(() => {
          location.reload();
        });
      });
      return Promise.reject(new Error(msg || "Token Expired"));
    }

    // 5. 处理业务失败
    ElMessage.error(msg || "系统出错");
    return Promise.reject(new Error(msg || "Error"));
  },
  (error: any) => {
    // 6. 处理 HTTP 错误
    if (error.response && error.response.data) {
      const { code, msg } = error.response.data;
      // token 过期，重新登录 (保留以防万一)
      if (code === 401 || code === "A0230") {
        ElMessageBox.confirm("当前页面已失效，请重新登录", "提示", {
          confirmButtonText: "确定",
          cancelButtonText: "取消",
          type: "warning",
          lockScroll: false,
        }).then(() => {
          const userStore = useUserStoreHook();
          userStore.resetToken().then(() => {
            location.reload();
          });
        });
      } else {
        ElMessage.error(msg || "系统接口异常");
      }
    } else {
      ElMessage.error("网络连接异常");
    }
    return Promise.reject(error);
  }
);

// 导出 axios 实例
export default service;
