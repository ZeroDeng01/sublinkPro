import { loginApi, logoutApi } from "@/api/auth";
import { getUserInfoApi } from "@/api/user";
import { resetRouter } from "@/router";
import { store, useNoticeStore } from "@/store";

import { LoginData } from "@/api/auth/types";
import { UserInfo } from "@/api/user/types";

export const useUserStore = defineStore("user", () => {
  const user = ref<UserInfo>({
    roles: [],
    perms: [],
  });

  const eventSource = ref<EventSource | null>(null);

  function connectSSE() {
    if (eventSource.value) return;

    const token = localStorage.getItem("accessToken");
    if (!token) return;

    // Extract the actual token string if it has "Bearer " prefix
    const tokenStr = token.replace("Bearer ", "");

    let url = "/api/sse?token=" + tokenStr;
    if (
      import.meta.env.VITE_APP_BASE_API &&
      import.meta.env.VITE_APP_BASE_API !== undefined
    ) {
      url = import.meta.env.VITE_APP_BASE_API + url;
    }
    eventSource.value = new EventSource(url);

    eventSource.value.onopen = () => {
      console.log("SSE Connected");
    };

    eventSource.value.addEventListener("task_update", (event) => {
      try {
        const data = JSON.parse(event.data);
        const type = data.status === "success" ? "success" : "error";
        const message = data.message;

        // Add to notice store
        const noticeStore = useNoticeStore();
        noticeStore.addNotification({
          title: data.status === "success" ? "成功" : "失败",
          message: message,
          type: type,
        });
      } catch (e) {
        console.error("Failed to parse SSE message", e);
      }
    });

    eventSource.value.onerror = (err) => {
      console.error("SSE Error:", err);
      eventSource.value?.close();
      eventSource.value = null;
    };
  }

  function disconnectSSE() {
    if (eventSource.value) {
      eventSource.value.close();
      eventSource.value = null;
    }
  }

  /**
   * 登录
   *
   * @param {LoginData}
   * @returns
   */
  function login(loginData: LoginData) {
    return new Promise<void>((resolve, reject) => {
      loginApi(loginData)
        .then((response) => {
          const { tokenType, accessToken } = response.data;
          localStorage.setItem("accessToken", tokenType + " " + accessToken); // Bearer eyJhbGciOiJIUzI1NiJ9.xxx.xxx
          connectSSE();
          resolve();
        })
        .catch((error) => {
          reject(error);
        });
    });
  }

  // 获取信息(用户昵称、头像、角色集合、权限集合)
  function getUserInfo() {
    return new Promise<UserInfo>((resolve, reject) => {
      getUserInfoApi()
        .then(({ data }) => {
          if (!data) {
            reject("Verification failed, please Login again.");
            return;
          }
          if (!data.roles || data.roles.length <= 0) {
            reject("getUserInfo: roles must be a non-null array!");
            return;
          }
          Object.assign(user.value, { ...data });
          resolve(data);
        })
        .catch((error) => {
          reject(error);
        });
    });
  }

  // user logout
  function logout() {
    return new Promise<void>((resolve, reject) => {
      logoutApi()
        .then(() => {
          localStorage.setItem("accessToken", "");
          disconnectSSE();
          location.reload(); // 清空路由
          resolve();
        })
        .catch((error) => {
          reject(error);
        });
    });
  }

  // remove token
  function resetToken() {
    console.log("resetToken");
    return new Promise<void>((resolve) => {
      localStorage.setItem("accessToken", "");
      resetRouter();
      resolve();
    });
  }

  return {
    user,
    login,
    getUserInfo,
    logout,
    resetToken,
    connectSSE,
    disconnectSSE,
  };
});

// 非setup
export function useUserStoreHook() {
  return useUserStore(store);
}
