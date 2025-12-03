<template>
  <el-dropdown trigger="click" class="setting-item">
    <div class="h100% p10px relative flex items-center">
      <el-icon><Bell /></el-icon>
      <el-badge
        :value="unreadCount"
        :hidden="unreadCount === 0"
        :max="99"
        class="absolute top-2px right-2px"
      />
    </div>

    <template #dropdown>
      <el-dropdown-menu class="notify-dropdown">
        <div class="notify-header">
          <span>系统通知</span>
          <el-button
            link
            type="primary"
            @click="handleMarkAllRead"
            :disabled="unreadCount === 0"
            >全部已读</el-button
          >
        </div>
        <el-scrollbar max-height="300px">
          <div v-if="notifications.length === 0" class="empty-notify">
            暂无通知
          </div>
          <div v-else>
            <div
              v-for="(item, index) in notifications"
              :key="index"
              class="notify-item"
              :class="{ 'is-read': item.read }"
              @click="handleRead(index)"
            >
              <div class="notify-title">
                <span>{{ item.message }}</span>
                <el-tag
                  size="small"
                  :type="item.type === 'error' ? 'danger' : item.type"
                  effect="plain"
                  >{{ item.read ? "已读" : "未读" }}</el-tag
                >
              </div>
              <div class="notify-time">{{ item.time }}</div>
            </div>
          </div>
        </el-scrollbar>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<script setup lang="ts">
import { useNoticeStore } from "@/store";

const noticeStore = useNoticeStore();
const notifications = computed(() => noticeStore.notifications);
const unreadCount = computed(() => noticeStore.unreadCount);

function handleRead(index: number) {
  noticeStore.markAsRead(index);
}

function handleMarkAllRead() {
  noticeStore.markAllAsRead();
}
</script>

<style lang="scss" scoped>
.setting-item {
  display: inline-block;
  min-width: 40px;
  height: 50px;
  line-height: 50px;
  color: var(--el-text-color);
  text-align: center;
  cursor: pointer;

  &:hover {
    background: rgb(0 0 0 / 10%);
  }
}

.notify-dropdown {
  width: 300px;
}

.notify-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 15px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  font-weight: bold;
}

.empty-notify {
  padding: 20px;
  text-align: center;
  color: var(--el-text-color-secondary);
}

.notify-item {
  padding: 10px 15px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  cursor: pointer;
  transition: background-color 0.3s;

  &:hover {
    background-color: var(--el-fill-color-light);
  }

  &.is-read {
    opacity: 0.6;
  }
}

.notify-title {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 5px;
  font-size: 14px;
  line-height: 1.4;
}

.notify-time {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>
