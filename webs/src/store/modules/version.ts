import { defineStore } from "pinia";
import { ref } from "vue";
import type { Ref } from "vue";
import { store } from "@/store";
import { GetVersion } from "@/api/auth";

/**
 * 版本管理 Store
 */
export const useVersionStore = defineStore("version", () => {
  // 状态
  const version: Ref<string> = ref(localStorage.getItem("version") || "");
  const loading: Ref<boolean> = ref(false);
  const error: Ref<string | null> = ref(null);

  /**
   * 获取版本信息
   * @returns Promise<string> 返回版本号
   */
  async function getVersion(): Promise<string> {
    loading.value = true;
    error.value = null;

    try {
      const res = await GetVersion();
      console.log("Version fetched:", res.data);

      version.value = res.data;
      localStorage.setItem("version", version.value);

      return version.value;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "获取版本失败";
      console.error("Error fetching version:", err);

      error.value = errorMessage;
      version.value = "unknown";
      localStorage.setItem("version", "unknown");

      throw err;
    } finally {
      loading.value = false;
    }
  }

  /**
   * 清除版本信息
   */
  function clearVersion(): void {
    version.value = "";
    error.value = null;
    localStorage.removeItem("version");
  }

  /**
   * 从本地存储恢复版本信息
   */
  function restoreVersion(): string {
    const storedVersion = localStorage.getItem("version") || "";
    version.value = storedVersion;
    return storedVersion;
  }

  return {
    // 状态
    version,
    loading,
    error,
    // 方法
    getVersion,
    clearVersion,
    restoreVersion,
  };
});

/**
 * 在非 setup 环境中使用
 */
export function useVersionStoreHook() {
  return useVersionStore(store);
}
