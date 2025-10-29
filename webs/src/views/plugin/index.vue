<template>
  <div class="app-container">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <div class="left">
            <el-input
              v-model="searchQuery"
              :placeholder="$t('plugin.search')"
              style="width: 200px; margin-left: 15px"
              clearable
              @input="handleSearch"
            >
              <template #prefix>
                <el-icon><search /></el-icon>
              </template>
            </el-input>
          </div>
          <div>
            <el-button
              type="primary"
              @click="reloadPlugins"
              :loading="reloading"
            >
              <el-icon><refresh /></el-icon>
              {{ $t("plugin.reload") }}
            </el-button>
          </div>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="filteredPlugins"
        style="width: 100%"
        row-key="name"
      >
        <el-table-column
          prop="name"
          :label="$t('plugin.name')"
          min-width="120"
        />
        <el-table-column
          prop="version"
          :label="$t('plugin.version')"
          width="100"
        />
        <el-table-column
          prop="description"
          :label="$t('plugin.description')"
          min-width="250"
          show-overflow-tooltip
        />
        <el-table-column
          prop="filePath"
          :label="$t('plugin.filePath')"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column
          prop="enabled"
          :label="$t('plugin.status')"
          width="100"
        >
          <template #default="scope">
            <el-tag :type="scope.row.enabled ? 'success' : 'info'">
              {{
                scope.row.enabled ? $t("plugin.enabled") : $t("plugin.disabled")
              }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          :label="$t('plugin.actions')"
          width="220"
          fixed="right"
        >
          <template #default="scope">
            <el-button
              v-if="!scope.row.enabled"
              type="success"
              size="small"
              @click="handleEnable(scope.row)"
              :loading="actionLoading === scope.row.name + '_enable'"
            >
              {{ $t("plugin.enable") }}
            </el-button>
            <el-button
              v-else
              type="warning"
              size="small"
              @click="handleDisable(scope.row)"
              :loading="actionLoading === scope.row.name + '_disable'"
            >
              {{ $t("plugin.disable") }}
            </el-button>
            <el-button
              type="primary"
              size="small"
              @click="openConfigDialog(scope.row)"
              :loading="actionLoading === scope.row.name + '_config'"
            >
              {{ $t("plugin.config") }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty
        v-if="!loading && (!filteredPlugins || filteredPlugins.length === 0)"
        :description="$t('plugin.noData')"
      >
        <el-button type="primary" @click="fetchPlugins">
          <el-icon><refresh /></el-icon>
          {{ $t("plugin.retry") }}
        </el-button>
      </el-empty>
    </el-card>

    <!-- 插件配置对话框 -->
    <el-dialog
      v-model="configDialogVisible"
      :title="
        currentPlugin
          ? `${$t('plugin.configTitle')}: ${currentPlugin.name}`
          : $t('plugin.configTitle')
      "
      width="600px"
      :close-on-click-modal="false"
      :close-on-press-escape="!saving"
    >
      <div v-if="currentPlugin" class="plugin-config">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="插件名称">{{
            currentPlugin.name
          }}</el-descriptions-item>
          <el-descriptions-item label="版本">{{
            currentPlugin.version
          }}</el-descriptions-item>
          <el-descriptions-item label="描述">{{
            currentPlugin.description
          }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="currentPlugin.enabled ? 'success' : 'info'">
              {{ currentPlugin.enabled ? "已启用" : "已禁用" }}
            </el-tag>
          </el-descriptions-item>
        </el-descriptions>

        <div class="config-form mt-4">
          <h3 class="mb-2">{{ $t("plugin.configTitle") }}</h3>
          <el-empty v-if="!configLoaded" description="Loading..." />
          <el-empty
            v-else-if="!configData"
            :description="$t('plugin.configEmpty')"
          />
          <el-form v-else label-position="top">
            <el-form-item label="配置 (JSON格式)">
              <el-input
                type="textarea"
                v-model="configDataJson"
                :rows="10"
                placeholder="请输入JSON格式的配置"
                :spellcheck="false"
                style="font-family: monospace"
              />
            </el-form-item>
            <div v-if="jsonError" class="json-error">
              <el-alert
                :title="jsonError"
                type="error"
                show-icon
                :closable="false"
              />
            </div>
          </el-form>
        </div>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="configDialogVisible = false" :disabled="saving">
            {{ $t("cancel") }}
          </el-button>
          <el-button type="primary" @click="handleSaveConfig" :loading="saving">
            {{ $t("confirm") }}
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
defineOptions({
  name: "PluginList",
  inheritAttrs: false,
});

import { ref, onMounted, computed } from "vue";
import { ElMessage } from "element-plus";
import { useI18n } from "vue-i18n";
import {
  getPluginsApi,
  enablePluginApi,
  disablePluginApi,
  getPluginConfigApi,
  updatePluginConfigApi,
  reloadPluginsApi,
} from "@/api/plugin";
import type { PluginInfo } from "@/api/plugin/types";

const { t } = useI18n();

// 状态
const loading = ref(false);
const reloading = ref(false);
const plugins = ref<PluginInfo[]>([]);
const searchQuery = ref("");
const actionLoading = ref(""); // 记录当前执行操作的插件: name_operation

// 配置对话框
const configDialogVisible = ref(false);
const currentPlugin = ref<PluginInfo | null>(null);
const configData = ref<Record<string, any>>({});
const configDataJson = ref("");
const jsonError = ref("");
const configLoaded = ref(false);
const saving = ref(false);

// 过滤后的插件列表
const filteredPlugins = computed(() => {
  if (!searchQuery.value) {
    return plugins.value;
  }

  const query = searchQuery.value.toLowerCase();
  return plugins.value.filter(
    (plugin) =>
      plugin.name.toLowerCase().includes(query) ||
      plugin.description.toLowerCase().includes(query)
  );
});

// 初始化加载插件列表
onMounted(async () => {
  try {
    await fetchPlugins();
  } catch (error) {
    console.error("初始加载失败，尝试重试", error);
    // 延迟1秒后自动重试一次
    setTimeout(() => {
      fetchPlugins();
    }, 1000);
  }
});

// 获取插件列表
const fetchPlugins = async () => {
  loading.value = true;
  try {
    console.log("开始获取插件列表...");
    const response = await getPluginsApi();
    console.log("获取插件列表响应:", response);

    // 检查响应是否有效
    if (response && response.data) {
      plugins.value = (response.data as unknown as PluginInfo[]) || [];
      console.log("获取到的插件数量:", plugins.value.length);
    } else {
      console.error("插件数据无效:", response);
      ElMessage.warning("获取到的插件数据格式不正确");
      plugins.value = [];
    }
  } catch (error: any) {
    console.error("获取插件列表出错:", error);
    // 显示更详细的错误信息
    const errorMsg = error.message || "未知错误";
    const statusCode = error.response?.status || "未知状态码";
    ElMessage.error(`获取插件列表失败: ${errorMsg} (${statusCode})`);
    plugins.value = []; // 确保在错误情况下也设置为空数组
  } finally {
    loading.value = false;
  }
};

// 搜索处理
const handleSearch = () => {
  // filteredPlugins是计算属性，会自动更新
};

// 重新加载插件
const reloadPlugins = async () => {
  reloading.value = true;
  try {
    await reloadPluginsApi();
    await fetchPlugins();
    ElMessage.success(t("plugin.message.reloadSuccess"));
  } catch (error) {
    ElMessage.error("重新加载插件失败，请检查网络连接");
  } finally {
    reloading.value = false;
  }
};

// 启用插件
const handleEnable = async (plugin: PluginInfo) => {
  actionLoading.value = `${plugin.name}_enable`;
  try {
    await enablePluginApi(plugin.name);
    ElMessage.success(t("plugin.message.enableSuccess"));
    // 更新插件状态
    const index = plugins.value.findIndex((p) => p.name === plugin.name);
    if (index !== -1) {
      plugins.value[index].enabled = true;
    }
  } catch (error) {
    console.error("启用插件出错:", error);
    ElMessage.error(t("plugin.message.enableFailed"));
  } finally {
    actionLoading.value = "";
  }
};

// 禁用插件
const handleDisable = async (plugin: PluginInfo) => {
  actionLoading.value = `${plugin.name}_disable`;
  try {
    await disablePluginApi(plugin.name);
    ElMessage.success(t("plugin.message.disableSuccess"));
    // 更新插件状态
    const index = plugins.value.findIndex((p) => p.name === plugin.name);
    if (index !== -1) {
      plugins.value[index].enabled = false;
    }
  } catch (error) {
    console.error("禁用插件出错:", error);
    ElMessage.error(t("plugin.message.disableFailed"));
  } finally {
    actionLoading.value = "";
  }
};

// 打开配置对话框
const openConfigDialog = async (plugin: PluginInfo) => {
  currentPlugin.value = { ...plugin };
  configDialogVisible.value = true;
  configLoaded.value = false;
  configData.value = {};
  configDataJson.value = "";
  jsonError.value = "";

  // 获取插件配置
  actionLoading.value = `${plugin.name}_config`;
  try {
    const { data } = await getPluginConfigApi(plugin.name);
    configData.value = data || {};
    // 将配置对象转换为格式化的 JSON 字符串
    configDataJson.value = JSON.stringify(configData.value, null, 2);
  } catch (error) {
    console.error("获取插件配置出错:", error);
    ElMessage.warning(t("plugin.configEmpty"));
  } finally {
    actionLoading.value = "";
    configLoaded.value = true;
  }
};

// 保存插件配置
const handleSaveConfig = async () => {
  if (!currentPlugin.value) return;

  // 检查JSON格式是否有效
  jsonError.value = "";
  try {
    // 尝试解析JSON字符串
    const parsedConfig = JSON.parse(configDataJson.value);
    configData.value = parsedConfig;
  } catch (e) {
    jsonError.value = `JSON格式错误: ${(e as Error).message}`;
    return;
  }

  saving.value = true;
  try {
    await updatePluginConfigApi({
      name: currentPlugin.value.name,
      config: configData.value,
    });
    ElMessage.success(t("plugin.message.configSuccess"));
    configDialogVisible.value = false;

    // 更新插件配置
    const index = plugins.value.findIndex(
      (p) => p.name === currentPlugin.value?.name
    );
    if (index !== -1) {
      plugins.value[index].config = { ...configData.value };
    }
  } catch (error) {
    console.error("保存插件配置出错:", error);
    ElMessage.error(t("plugin.message.configFailed"));
  } finally {
    saving.value = false;
  }
};
</script>

<style scoped>
.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.left {
  display: flex;
  align-items: center;
}

.title {
  font-size: 18px;
  font-weight: bold;
}

.mt-4 {
  margin-top: 16px;
}

.mb-2 {
  margin-bottom: 8px;
}
</style>
