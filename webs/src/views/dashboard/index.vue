<template>
  <div class="dashboard-container">
    <!-- github角标 -->
    <github-corner class="github-corner" />

    <el-card shadow="never">
      <el-row justify="space-between">
        <el-col :span="18" :xs="24">
          <div class="flex h-full items-center">
            <img
              class="w-20 h-20 mr-5 rounded-full"
              :src="userStore.user.avatar + '?imageView2/1/w/80/h/80'"
            />
            <div>
              <p>{{ greetings }}</p>
              <p class="text-sm text-gray">Keep Coding, Keep Growing!</p>
            </div>
          </div>
        </el-col>

        <el-col :span="6" :xs="24">
          <div class="flex h-full items-center justify-around">
            <el-statistic
              v-for="item in statisticData"
              :key="item.key"
              :value="item.value"
            >
              <template #title>
                <div class="flex items-center">
                  <svg-icon :icon-class="item.iconClass" size="20px" />
                  <span class="text-[16px] ml-1">{{ item.title }}</span>
                </div>
              </template>
            </el-statistic>
          </div>
        </el-col>
      </el-row>
    </el-card>

    <el-card shadow="never" class="mt-5">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="text-lg font-bold">📝 更新日志</span>
          <el-badge :is-dot="hasNewVersion" type="danger" class="version-badge">
            <el-tag
              @click="checkForUpdate"
              type="success"
              size="default"
              class="ml-3"
              effect="dark"
            >
              当前版本: {{ versionStore.version }}
            </el-tag>
          </el-badge>
          <el-button link type="primary" @click="openGithubReleases">
            查看全部
          </el-button>
        </div>
      </template>

      <div v-loading="releaseLoading">
        <div v-if="releases.length > 0">
          <div
            v-for="(release, index) in releases"
            :key="release.id"
            class="release-item"
            :class="{ 'border-t': index > 0 }"
          >
            <div class="flex items-center justify-between mb-2">
              <div class="flex items-center">
                <span class="text-lg font-semibold ml-3 mr-2">
                  {{ release.name }}
                </span>
                <el-tag :type="index === 0 ? 'success' : 'info'" size="small">
                  {{ release.tag_name }}
                </el-tag>
                <el-tag
                  v-if="release.prerelease"
                  type="warning"
                  size="small"
                  class="ml-2"
                >
                  Pre-release
                </el-tag>
                <el-tag
                  v-if="index === 0 && hasNewVersion"
                  type="danger"
                  size="small"
                  class="ml-2"
                >
                  新版本
                </el-tag>
              </div>
              <span class="text-sm text-gray">
                {{ formatDate(release.published_at) }}
              </span>
            </div>
            <div
              class="release-body"
              v-html="formatMarkdown(release.body)"
            ></div>
          </div>
        </div>

        <el-empty v-else-if="!releaseLoading" description="暂无更新日志" />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
defineOptions({
  name: "Dashboard",
  inheritAttrs: false,
});

import { useVersionStore } from "@/store/modules/version";
import { useUserStore } from "@/store/modules/user";
import { getNodeTotal, getSubTotal } from "@/api/total";
import MarkdownIt from "markdown-it";
import axios from "axios";

const versionStore = useVersionStore();
const userStore = useUserStore();
const date: Date = new Date();
const subTotal = ref(0);
const nodeTotal = ref(0);
const markdown = new MarkdownIt();

// 更新日志相关
const releases = ref<any[]>([]);
const releaseLoading = ref(false);
const checkingUpdate = ref(false);
const hasNewVersion = ref(false);

// 右上角数量
const statisticData = ref([
  {
    value: 0,
    iconClass: "message",
    title: "订阅",
    key: "message",
  },
  {
    value: 0,
    iconClass: "link",
    title: "节点",
    key: "upcoming",
  },
]);

const getsubtotal = async () => {
  const { data } = await getSubTotal();
  subTotal.value = data;
  statisticData.value[0].value = data;
};

const getnodetotal = async () => {
  const { data } = await getNodeTotal();
  nodeTotal.value = data;
  statisticData.value[1].value = data;
};

// 获取 GitHub Releases
const fetchGithubReleases = async () => {
  releaseLoading.value = true;
  try {
    const response = await axios.get(
      "https://api.github.com/repos/ZeroDeng01/sublinkPro/releases",
      {
        params: {
          per_page: 5, // 只显示最近3个版本
        },
      }
    );
    releases.value = response.data;
    // 检查是否有新版本
    checkVersionDifference();
  } catch (error) {
    console.error("获取更新日志失败:", error);
    ElMessage.error("获取更新日志失败");
  } finally {
    releaseLoading.value = false;
  }
};
// 检查版本差异
const checkVersionDifference = () => {
  if (releases.value.length > 0) {
    const latestVersion = releases.value[0].tag_name;
    const currentVersion = versionStore.version;
    hasNewVersion.value = latestVersion !== currentVersion;
  }
};

// 检查更新
const checkForUpdate = async () => {
  checkingUpdate.value = true;
  try {
    await fetchGithubReleases();
    if (hasNewVersion.value) {
      ElMessageBox.confirm(
        `新版本 ${releases.value[0].tag_name}，建议您及时更新以获得更好的体验！`,
        "发现新版本",
        {
          confirmButtonText: "查看详情",
          cancelButtonText: "取消",
          type: "success",
        }
      )
        .then(() => {
          openGithubReleases();
        })
        .catch(() => {
          // 用户点击取消或关闭，不做任何操作
        });
    } else {
      await ElMessageBox.alert(
        "您当前使用的已是最新版本，无需更新。",
        "版本检查",
        {
          confirmButtonText: "确定",
          type: "info",
        }
      ).catch(() => {
        // 用户点击取消或关闭，不做任何操作
      });
    }
  } catch (error) {
    console.error("检查更新失败:", error);
    await ElMessageBox.alert("检查更新失败,请检查网络连接后重试。", "错误", {
      confirmButtonText: "确定",
      type: "error",
    }).catch(() => {
      // 用户点击取消或关闭，不做任何操作
    });
  } finally {
    checkingUpdate.value = false;
  }
};

// 打开 GitHub Releases 页面
const openGithubReleases = () => {
  window.open("https://github.com/ZeroDeng01/sublinkPro/releases", "_blank");
};

// 格式化日期
const formatDate = (dateString: string) => {
  const date = new Date(dateString);
  const now = new Date();
  const diff = now.getTime() - date.getTime();
  const days = Math.floor(diff / (1000 * 60 * 60 * 24));

  if (days === 0) {
    return "今天";
  } else if (days === 1) {
    return "昨天";
  } else if (days < 7) {
    return `${days} 天前`;
  } else if (days < 30) {
    return `${Math.floor(days / 7)} 周前`;
  } else {
    return date.toLocaleDateString("zh-CN", {
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
    });
  }
};

// 简单的 Markdown 格式化（基础版本）
const formatMarkdown = (md: string) => {
  if (!md) return "";
  return markdown.render(md);
};

onMounted(() => {
  getsubtotal();
  getnodetotal();
  fetchGithubReleases();
});

const greetings = computed(() => {
  const hours = date.getHours();
  if (hours >= 6 && hours < 8) {
    return "晨起披衣出草堂,轩窗已自喜微凉🌅！";
  } else if (hours >= 8 && hours < 12) {
    return "上午好，" + userStore.user.nickname + "！";
  } else if (hours >= 12 && hours < 18) {
    return "下午好，" + userStore.user.nickname + "！";
  } else if (hours >= 18 && hours < 24) {
    return "晚上好，" + userStore.user.nickname + "！";
  } else {
    return "偷偷向银河要了一把碎星，只等你闭上眼睛撒入你的梦中，晚安🌛！";
  }
});
</script>

<style lang="scss" scoped>
.dashboard-container {
  position: relative;
  padding: 24px;

  .user-avatar {
    width: 40px;
    height: 40px;
    border-radius: 50%;
  }

  .github-corner {
    position: absolute;
    top: 0;
    right: 0;
    z-index: 1;
    border: 0;
  }

  .data-box {
    display: flex;
    justify-content: space-between;
    padding: 20px;
    font-weight: bold;
    color: var(--el-text-color-regular);
    background: var(--el-bg-color-overlay);
    border-color: var(--el-border-color);
    box-shadow: var(--el-box-shadow-dark);
  }

  .svg-icon {
    fill: currentcolor !important;
  }

  .release-item {
    padding: 5px 0;

    &:first-child {
      padding-top: 0;
    }

    &:last-child {
      padding-bottom: 0;
    }

    .release-body {
      margin-top: 2px;
      line-height: 1.2;
      color: var(--el-text-color-regular);

      :deep(h1),
      :deep(h2),
      :deep(h3) {
        margin: 12px 0 6px;
        font-weight: 600;
      }

      :deep(h1) {
        font-size: 1.5em;
      }

      :deep(h2) {
        font-size: 1.3em;
      }

      :deep(h3) {
        font-size: 1.1em;
      }

      :deep(ul) {
        padding-left: 20px;
        margin: 6px 0;
      }

      :deep(li) {
        margin: 2px 0;
        list-style-type: disc;
      }

      :deep(code) {
        padding: 2px 6px;
        font-family: "Courier New", monospace;
        font-size: 0.9em;
        background-color: var(--el-fill-color-light);
        border-radius: 4px;
      }

      :deep(a) {
        color: var(--el-color-primary);
        text-decoration: none;

        &:hover {
          text-decoration: underline;
        }
      }

      :deep(strong) {
        font-weight: 600;
      }
    }
  }
}
</style>
