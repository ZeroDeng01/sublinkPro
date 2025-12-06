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
  const reconnectTimeout = ref<NodeJS.Timeout | null>(null);
  const heartbeatTimeout = ref<NodeJS.Timeout | null>(null);

  function resetHeartbeat() {
    if (heartbeatTimeout.value) clearTimeout(heartbeatTimeout.value);
    heartbeatTimeout.value = setTimeout(() => {
      console.warn("SSE Heartbeat timeout, reconnecting...");
      eventSource.value?.close();
      eventSource.value = null;
      connectSSE();
    }, 15000); // 15s timeout (backend sends heartbeat every 10s)
  }

  function handleReconnect() {
    if (eventSource.value) {
      eventSource.value.close();
      eventSource.value = null;
    }
    if (reconnectTimeout.value) clearTimeout(reconnectTimeout.value);

    console.log("Attempting to reconnect SSE in 5s...");
    reconnectTimeout.value = setTimeout(() => {
      connectSSE();
    }, 5000);
  }

  function connectSSE() {
    if (eventSource.value?.readyState === 1) return; // Already connected

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

    // Close existing connection if any
    if (eventSource.value) {
      eventSource.value.close();
    }

    eventSource.value = new EventSource(url);

    eventSource.value.onopen = () => {
      console.log("SSE Connected");
      resetHeartbeat();
    };

    eventSource.value.addEventListener("heartbeat", () => {
      console.log("SSE Heartbeat received");
      resetHeartbeat();
    });

    eventSource.value.addEventListener("task_update", (event) => {
      resetHeartbeat();
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
      handleReconnect();
    };
  }

  function disconnectSSE() {
    if (eventSource.value) {
      eventSource.value.close();
      eventSource.value = null;
    }
    if (reconnectTimeout.value) clearTimeout(reconnectTimeout.value);
    if (heartbeatTimeout.value) clearTimeout(heartbeatTimeout.value);
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
