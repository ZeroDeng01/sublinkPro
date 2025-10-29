<template>
  <div class="app-container">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <div class="left">
            <span class="title">{{ $t("apikey.title") }}</span>
            <el-input
              v-model="searchQuery"
              :placeholder="$t('apikey.search')"
              style="width: 200px; margin-left: 15px"
              clearable
              @input="handleSearch"
            >
              <template #prefix>
                <el-icon><search /></el-icon>
              </template>
            </el-input>
          </div>
          <el-button type="primary" @click="openCreateDialog">{{
            $t("apikey.createNew")
          }}</el-button>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="filteredApiKeys"
        style="width: 100%"
        :default-sort="{ prop: 'created_at', order: 'descending' }"
      >
        <el-table-column prop="id" label="ID" width="80" sortable />
        <el-table-column prop="description" :label="$t('apikey.description')" />
        <el-table-column
          prop="created_at"
          :label="$t('apikey.createdAt')"
          width="180"
          sortable
        >
          <template #default="scope">
            {{ formatDateTime(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="expiredAt"
          :label="$t('apikey.expiredAt')"
          width="180"
          sortable
        >
          <template #default="scope">
            <el-tag
              :type="getExpirationTagType(scope.row)"
              :effect="scope.row.expiredAt ? 'light' : 'plain'"
            >
              {{
                scope.row.expiredAt
                  ? formatDateTime(scope.row.expiredAt)
                  : $t("apikey.neverExpire")
              }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          :label="$t('apikey.actions')"
          width="120"
          fixed="right"
        >
          <template #default="scope">
            <el-button
              type="danger"
              size="small"
              @click="handleDelete(scope.row)"
            >
              {{ $t("apikey.delete") }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty
        v-if="filteredApiKeys.length === 0 && !loading"
        :description="$t('apikey.noData')"
      />
    </el-card>
    <!-- 创建API Key对话框 -->
    <el-dialog
      v-model="createDialogVisible"
      :title="$t('apikey.createNew')"
      width="500px"
      :close-on-click-modal="false"
      :close-on-press-escape="!creating"
    >
      <el-form :model="newApiKey" label-width="120px" ref="apiKeyFormRef">
        <el-form-item :label="$t('apikey.description')" required>
          <el-input
            v-model="newApiKey.description"
            :placeholder="$t('apikey.descriptionPlaceholder')"
            maxlength="10"
            show-word-limit
            clearable
          />
        </el-form-item>
        <el-form-item :label="$t('apikey.expiration')">
          <el-radio-group v-model="expirationOption">
            <el-radio value="never">{{ $t("apikey.neverExpire") }}</el-radio>
            <el-radio value="custom">{{
              $t("apikey.customExpire")
            }}</el-radio> </el-radio-group
          ><el-date-picker
            v-if="expirationOption === 'custom'"
            v-model="newApiKey.expiredAt"
            type="datetime"
            :placeholder="$t('apikey.selectExpiration')"
            :disabled-date="disabledDate"
            style="width: 100%; margin-top: 10px"
            value-format="YYYY-MM-DDTHH:mm:ss"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="createDialogVisible = false" :disabled="creating">
            {{ $t("cancel") || "取消" }}
          </el-button>
          <el-button
            type="primary"
            @click="handleCreateAPIKey"
            :loading="creating"
          >
            {{ $t("confirm") || "确定" }}
          </el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 显示新创建的API Key对话框 -->
    <el-dialog
      v-model="keyDialogVisible"
      :title="$t('apikey.newKeyCreated')"
      width="600px"
      :close-on-click-modal="false"
      :show-close="false"
      center
    >
      <div class="api-key-display">
        <el-alert
          :title="$t('apikey.saveKeyWarning')"
          type="warning"
          show-icon
          :closable="false"
        />

        <el-input v-model="createdApiKey" readonly type="text" class="mt-4">
          <template #append>
            <el-button @click="copyApiKey" type="primary">
              <el-icon class="mr-1"><CopyDocument /></el-icon>
              {{ $t("apikey.copy") }}
            </el-button>
          </template>
        </el-input>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button type="primary" @click="keyDialogVisible = false">
            {{ $t("apikey.iSavedIt") }}
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
import { ElMessage, ElMessageBox, ElLoading } from "element-plus";
import { useI18n } from "vue-i18n";
import {
  getAPIKeys,
  createAPIKey,
  deleteAPIKey,
  type APIKey,
  type CreateAPIKeyParams,
} from "@/api/user/apikey";
import { useUserStore } from "@/store/modules/user";

// 使用统一的接口定义，但内部映射字段名
interface LocalAPIKey {
  id: number;
  user_id: number;
  username: string;
  created_at: string;
  expiredAt: string | null;
  description: string;
}

const { t } = useI18n();

const apiKeys = ref<LocalAPIKey[]>([]);
const filteredApiKeys = computed(() => {
  if (!searchQuery.value) return apiKeys.value;
  const query = searchQuery.value.toLowerCase();
  return apiKeys.value.filter((key) => key.description.includes(query));
});

const searchQuery = ref("");
const createDialogVisible = ref(false);
const keyDialogVisible = ref(false);
const creating = ref(false);
const createdApiKey = ref("");
const expirationOption = ref("never");

const newApiKey = ref<CreateAPIKeyParams>({
  description: "",
  expiredAt: undefined,
});

// 获取API密钥列表
const loading = ref(false);
const fetchAPIKeys = async () => {
  loading.value = true;
  try {
    const userStore = useUserStore();
    let userId = userStore.user.userId;
    if (!userId) {
      ElMessage.error("用户ID不存在");
      return;
    }
    const res = await getAPIKeys(userId);
    if (res.data && Array.isArray(res.data)) {
      apiKeys.value = res.data.map((item: APIKey) => ({
        id: item.id,
        user_id: item.userID,
        username: item.username,
        created_at: item.createdAt,
        expiredAt: item.expiredAt,
        description: item.description,
      }));
    } else {
      apiKeys.value = [];
    }
  } catch (error) {
    console.error("获取API密钥列表失败:", error);
    ElMessage.error(t("apikey.fetchFailed"));
  } finally {
    loading.value = false;
  }
};

// 打开创建对话框
const openCreateDialog = () => {
  newApiKey.value = {
    description: "",
    expiredAt: undefined,
  };
  expirationOption.value = "never";
  createDialogVisible.value = true;
};

// 创建API密钥
const handleCreateAPIKey = async () => {
  if (!newApiKey.value.description) {
    ElMessage.warning(t("apikey.descriptionRequired"));
    return;
  }

  try {
    creating.value = true;
    const params: CreateAPIKeyParams = {
      description: newApiKey.value.description,
    };

    if (expirationOption.value === "custom" && newApiKey.value.expiredAt) {
      params.expiredAt = toBeijingISO8601(newApiKey.value.expiredAt);
    }
    const userStore = useUserStore();
    let username = userStore.user.username;
    params.username = username;
    const res = await createAPIKey(params);
    creating.value = false;
    createDialogVisible.value = false;
    createdApiKey.value = res.data.accessKey;

    keyDialogVisible.value = true;

    // 刷新列表
    await fetchAPIKeys();
  } catch (error) {
    creating.value = false;
    console.error("创建API密钥失败:", error);
    ElMessage.error(t("apikey.createFailed"));
  }
};

// 删除API密钥
const handleDelete = (row: LocalAPIKey) => {
  ElMessageBox.confirm(
    t("apikey.deleteConfirmMessage"),
    t("apikey.deleteConfirmTitle"),
    {
      confirmButtonText: t("confirm") || "确定",
      cancelButtonText: t("cancel") || "取消",
      type: "warning",
    }
  )
    .then(async () => {
      const loadingInstance = ElLoading.service({
        lock: true,
        text: t("apikey.deleting"),
        background: "rgba(0, 0, 0, 0.7)",
      });
      try {
        await deleteAPIKey(row.id);
        ElMessage.success(t("apikey.deleteSuccess"));
        await fetchAPIKeys();
      } catch (error) {
        console.error("删除API密钥失败:", error);
        ElMessage.error(t("apikey.deleteFailed"));
      } finally {
        loadingInstance.close();
      }
    })
    .catch(() => {
      // 用户取消删除
    });
};

// 处理搜索
const handleSearch = () => {
  // 搜索功能通过computed实现
};

// 获取过期标签类型
const getExpirationTagType = (row: LocalAPIKey) => {
  if (!row.expiredAt) return "success"; // 永不过期

  const expireDate = new Date(row.expiredAt);
  const now = new Date();

  if (expireDate < now) {
    return "danger"; // 已过期
  }

  // 计算剩余天数
  const diffTime = expireDate.getTime() - now.getTime();
  const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));

  if (diffDays <= 7) {
    return "warning"; // 即将过期
  }

  return "info"; // 正常
};

// 通用复制文本到剪贴板函数
const copyTextToClipboard = (text: string) => {
  const textarea = document.createElement("textarea");
  textarea.value = text;
  textarea.style.position = "fixed";
  textarea.style.opacity = "0";
  document.body.appendChild(textarea);
  textarea.focus();
  textarea.select();

  const successful = document.execCommand("copy");
  document.body.removeChild(textarea);

  return successful;
};

// 复制API密钥
const copyApiKey = () => {
  const successful = copyTextToClipboard(createdApiKey.value);

  if (successful) {
    ElMessage({
      message: t("apikey.copySuccess"),
      type: "success",
      duration: 1500,
    });
  } else {
    ElMessage({
      message: t("apikey.copyFailed"),
      type: "error",
      duration: 3000,
    });
  }
};

// 格式化日期时间
const formatDateTime = (dateStr: string) => {
  if (!dateStr) return "";
  const date = new Date(dateStr);
  return date.toLocaleString();
};

// 禁用今天之前的日期
const disabledDate = (time: Date) => {
  return time.getTime() < Date.now() - 8.64e7; // 禁用今天之前的日期
};

function toBeijingISO8601(input: string | Date): string {
  const date = new Date(input);
  const pad = (n: number) => n.toString().padStart(2, "0");

  const year = date.getFullYear();
  const month = pad(date.getMonth() + 1);
  const day = pad(date.getDate());
  const hours = pad(date.getHours());
  const minutes = pad(date.getMinutes());
  const seconds = pad(date.getSeconds());

  return `${year}-${month}-${day}T${hours}:${minutes}:${seconds}+08:00`;
}

onMounted(() => {
  fetchAPIKeys();
});
</script>

<style scoped>
.breadcrumb-container {
  padding: 8px 16px;
  margin-bottom: 15px;
  background-color: #fff;
  border-radius: 4px;
  box-shadow: 0 1px 4px rgb(0 21 41 / 8%);
}

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
  font-size: 16px;
  font-weight: 600;
}

.api-key-display {
  margin: 20px 0;
  text-align: center;
}

.api-key-display p {
  margin-bottom: 15px;
  font-weight: bold;
  color: #f56c6c;
}

.mt-4 {
  margin-top: 16px;
}

.mr-1 {
  margin-right: 4px;
}
</style>
