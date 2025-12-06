import { defineStore } from "pinia";
import { store } from "@/store";
import { ElNotification } from "element-plus";

export interface NotificationItem {
  title: string;
  message: string;
  type: "success" | "warning" | "info" | "error";
  time: string;
  read: boolean;
}

export const useNoticeStore = defineStore("notice", () => {
  const notifications = ref<NotificationItem[]>([]);

  const unreadCount = computed(() => {
    return notifications.value.filter((item) => !item.read).length;
  });

  function addNotification(
    notification: Omit<NotificationItem, "read" | "time">
  ) {
    notifications.value.unshift({
      ...notification,
      read: false,
      time: new Date().toLocaleString(),
    });

    ElMessage({
      message: notification.message,
      type: notification.type,
    });
  }

  function markAsRead(index: number) {
    if (notifications.value[index]) {
      notifications.value[index].read = true;
    }
  }

  function markAllAsRead() {
    notifications.value.forEach((item) => {
      item.read = true;
    });
  }

  function clearAll() {
    notifications.value = [];
  }

  return {
    notifications,
    unreadCount,
    addNotification,
    markAsRead,
    markAllAsRead,
    clearAll,
  };
});

export function useNoticeStoreHook() {
  return useNoticeStore(store);
}
