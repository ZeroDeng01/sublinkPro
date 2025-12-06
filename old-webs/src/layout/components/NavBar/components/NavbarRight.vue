<template>
  <div class="flex">
    <template v-if="!isMobile">
      <!--全屏 -->
      <div class="setting-item" @click="toggle">
        <svg-icon
          :icon-class="isFullscreen ? 'fullscreen-exit' : 'fullscreen'"
        />
      </div>

      <!-- 布局大小 -->
      <el-tooltip
        :content="$t('sizeSelect.tooltip')"
        effect="dark"
        placement="bottom"
      >
        <size-select class="setting-item" />
      </el-tooltip>

      <!-- 语言选择 -->
      <lang-select class="setting-item" />
    </template>

    <!-- 通知 -->
    <Notify />

    <!-- 用户头像 -->
    <el-dropdown class="setting-item" trigger="click">
      <div class="flex-center h100% p10px">
        <img
          :src="userStore.user.avatar + '?imageView2/1/w/80/h/80'"
          class="rounded-full mr-10px w24px w24px"
        />
        <span>{{ userStore.user.username }}</span>
      </div>
      <template #dropdown>
        <el-dropdown-menu>
          <router-link to="/apikey/index">
            <el-dropdown-item>{{ $t("apikey.manage") }}</el-dropdown-item>
          </router-link>
          <router-link to="/personal/center">
            <el-dropdown-item>{{
              $t("navbar.personalCenter")
            }}</el-dropdown-item>
          </router-link>
          <el-dropdown-item divided @click="backup">
            {{ $t("navbar.backup") }}
          </el-dropdown-item>
          <el-dropdown-item divided @click="logout">
            {{ $t("navbar.logout") }}
          </el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>

    <!-- 设置 -->
    <template v-if="defaultSettings.showSettings">
      <div class="setting-item" @click="settingStore.settingsVisible = true">
        <svg-icon icon-class="setting" />
      </div>
    </template>
  </div>
</template>
<script setup lang="ts">
import Notify from "./Notify.vue";
import {
  useAppStore,
  useTagsViewStore,
  useUserStore,
  useSettingsStore,
} from "@/store";
import defaultSettings from "@/settings";
import { DeviceEnum } from "@/enums/DeviceEnum";
import request from "@/utils/request";
import { ElLoading } from "element-plus";

const appStore = useAppStore();
const tagsViewStore = useTagsViewStore();
const userStore = useUserStore();
const settingStore = useSettingsStore();

const route = useRoute();
const router = useRouter();

const isMobile = computed(() => appStore.device === DeviceEnum.MOBILE);

const { isFullscreen, toggle } = useFullscreen();

/**
 * 注销
 */
function logout() {
  ElMessageBox.confirm("确定注销并退出系统吗？", "提示", {
    confirmButtonText: "确定",
    cancelButtonText: "取消",
    type: "warning",
    lockScroll: false,
  }).then(() => {
    userStore
      .logout()
      .then(() => {
        tagsViewStore.delAllViews();
      })
      .then(() => {
        router.push(`/login?redirect=${route.fullPath}`);
      });
  });
}

/**
 * 数据备份
 */
function backup() {
  ElMessageBox.confirm(
    "确定备份系统数据吗？数据备份文件存有您的机密信息，请妥善保管。",
    "温馨提示",
    {
      confirmButtonText: "Yes！搞起！",
      cancelButtonText: "No！我是随缘主义！",
      type: "warning",
      lockScroll: false,
    }
  ).then(() => {
    // 1. 显示全屏加载
    const loadingInstance = ElLoading.service({
      lock: true,
      text: "正在生成备份文件，请稍候...",
      background: "rgba(0, 0, 0, 0.7)",
    });

    request({
      url: "/api/v1/backup/download",
      method: "get",
      responseType: "blob",
    })
      .then((response) => {
        const data = response.data;

        // 1. 严格的类型守卫：检查返回的是否真的是 Blob
        if (!(data instanceof Blob)) {
          ElMessage.error("备份失败：服务器未返回有效的备份文件。");
          console.error("Backup failed: response.data is not a Blob", data);
          return;
        }

        // 2. 到这里， TypeScript 知道 'data' 100% 是 Blob
        const blob: Blob = data;

        // 3. 检查 Blob 大小 (现在类型安全了)
        if (!blob || blob.size === 0) {
          ElMessage.error("下载失败，获取到的文件为空");
          return;
        }

        // --- 核心修改结束 ---

        // 提取文件名 (你的逻辑)
        let filename = `sublink-pro-backup.zip`;
        if (response.headers && response.headers["content-disposition"]) {
          const contentDisposition = response.headers["content-disposition"];
          const match = contentDisposition.match(
            /filename="?([^"]+)"?|filename\*=UTF-8''([^"]+)/
          );
          if (match && (match[1] || match[2])) {
            filename = decodeURIComponent(match[2] || match[1]);
          }
        }

        // 创建下载链接 (现在类型安全了)
        const url = window.URL.createObjectURL(blob);
        const link = document.createElement("a");
        link.href = url;
        link.download = filename;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        window.URL.revokeObjectURL(url);

        ElMessage.success("备份文件已开始下载");
      })
      .catch((err) => {
        // 网络层或请求配置错误
        ElMessage.error("备份请求失败，请检查网络或服务器日志");
        console.error("Backup request failed:", err);
      })
      .finally(() => {
        // 3. 无论成功还是失败，都关闭 loading
        loadingInstance.close();
      });
  });
}
</script>
<style lang="scss" scoped>
.setting-item {
  display: inline-block;
  min-width: 40px;
  height: $navbar-height;
  line-height: $navbar-height;
  color: var(--el-text-color);
  text-align: center;
  cursor: pointer;

  &:hover {
    background: rgb(0 0 0 / 10%);
  }
}

.layout-top,
.layout-mix {
  .setting-item,
  .el-icon {
    color: var(--el-color-white);
  }
}

.dark .setting-item:hover {
  background: rgb(255 255 255 / 20%);
}
</style>
