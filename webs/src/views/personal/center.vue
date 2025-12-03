<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import { useUserStore } from "@/store";
import request from "@/utils/request";
import { useI18n } from "vue-i18n";
import type { FormInstance, FormRules } from "element-plus";

const { t } = useI18n();
const userStore = useUserStore();

// User info
const userinfo = ref<any>();

// Password form
const passwordFormRef = ref<FormInstance>();
const passwordForm = reactive({
  oldPassword: "",
  newPassword: "",
  confirmPassword: "",
});

// Form validation rules
const passwordRules = reactive<FormRules>({
  oldPassword: [
    {
      required: true,
      message: t("personalCenter.message.oldPasswordRequired"),
      trigger: "blur",
    },
  ],
  newPassword: [
    {
      required: true,
      message: t("personalCenter.message.newPasswordRequired"),
      trigger: "blur",
    },
    {
      min: 6,
      message: t("personalCenter.message.passwordTooShort"),
      trigger: "blur",
    },
  ],
  confirmPassword: [
    {
      required: true,
      message: t("personalCenter.message.confirmPasswordRequired"),
      trigger: "blur",
    },
    {
      validator: (rule: any, value: any, callback: any) => {
        if (value !== passwordForm.newPassword) {
          callback(new Error(t("personalCenter.message.passwordMismatch")));
        } else {
          callback();
        }
      },
      trigger: "blur",
    },
  ],
});

// Profile form
const profileFormRef = ref<FormInstance>();
const profileForm = reactive({
  username: "",
  nickname: "",
});

const loading = ref(false);

// Get user info
onMounted(async () => {
  userinfo.value = await userStore.getUserInfo();
  profileForm.username = userinfo.value.username;
  profileForm.nickname = userinfo.value.nickname || "";
  profileForm.nickname = userinfo.value.nickname || "";

  getWebhookConfig();
});

/** Change password */
async function handleChangePassword() {
  if (!passwordFormRef.value) return;

  await passwordFormRef.value.validate(async (valid) => {
    if (!valid) return;

    loading.value = true;
    try {
      await request({
        url: "/api/v1/users/change-password",
        method: "post",
        data: {
          oldPassword: passwordForm.oldPassword,
          newPassword: passwordForm.newPassword,
          confirmPassword: passwordForm.confirmPassword,
        },
      });

      ElMessage.success(t("personalCenter.message.changeSuccess"));

      // Reset form
      passwordFormRef.value?.resetFields();
      passwordForm.oldPassword = "";
      passwordForm.newPassword = "";
      passwordForm.confirmPassword = "";

      // Optional: Logout user after password change
      setTimeout(() => {
        ElMessageBox.confirm("密码修改成功，需要重新登录。", "提示", {
          confirmButtonText: "确定",
          showCancelButton: false,
          type: "success",
        }).then(() => {
          userStore.logout().then(() => {
            window.location.href = "/#/login";
          });
        });
      }, 500);
    } catch (error: any) {
      const errorMsg = error.response?.data?.message || error.message;
      if (errorMsg.includes("password") || errorMsg.includes("密码")) {
        ElMessage.error(t("personalCenter.message.oldPasswordIncorrect"));
      } else {
        ElMessage.error(t("personalCenter.message.changeFailed"));
      }
    } finally {
      loading.value = false;
    }
  });
}

/** Update profile */
async function handleUpdateProfile() {
  if (!profileFormRef.value) return;

  // Check if username changed
  const usernameChanged = userinfo.value.username !== profileForm.username;

  loading.value = true;
  try {
    await request({
      url: "/api/v1/users/update-profile",
      method: "post",
      data: {
        username: profileForm.username,
        nickname: profileForm.nickname,
      },
    });

    ElMessage.success(t("personalCenter.message.updateSuccess"));

    // If username changed, force logout
    if (usernameChanged) {
      setTimeout(() => {
        ElMessageBox.confirm(
          t("personalCenter.message.usernameChangedRelogin"),
          t("userset.message.title"),
          {
            confirmButtonText: t("confirm"),
            showCancelButton: false,
            type: "warning",
          }
        ).then(() => {
          userStore.logout().then(() => {
            window.location.href = "/#/login";
          });
        });
      }, 500);
    } else {
      // Refresh user info if only nickname changed
      await userStore.getUserInfo();
    }
  } catch (error: any) {
    ElMessage.error(
      t("personalCenter.message.updateFailed") +
        "：" +
        (error.response?.data?.message || error.message)
    );
  } finally {
    loading.value = false;
  }
}

// Webhook form
const webhookFormRef = ref<FormInstance>();
const webhookForm = reactive({
  webhookUrl: "",
  webhookMethod: "POST",
  webhookContentType: "application/json",
  webhookHeaders: "",
  webhookBody: "",
  webhookEnabled: false,
});

const webhookRules = {
  webhookUrl: [
    { required: true, message: "请输入 Webhook URL", trigger: "blur" },
  ],
};

async function getWebhookConfig() {
  try {
    const { data } = await request({
      url: "/api/v1/settings/webhook",
      method: "get",
    });
    webhookForm.webhookUrl = data.webhookUrl || "";
    webhookForm.webhookMethod = data.webhookMethod || "POST";
    webhookForm.webhookContentType =
      data.webhookContentType || "application/json";
    webhookForm.webhookHeaders = data.webhookHeaders || "";
    webhookForm.webhookBody = data.webhookBody || "";
    webhookForm.webhookEnabled = data.webhookEnabled || false;
  } catch (error: any) {
    // Silent error or log
    console.error("Failed to fetch webhook config", error);
  }
}

async function handleTestWebhook() {
  if (!webhookFormRef.value) return;

  await webhookFormRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true;
      try {
        await request({
          url: "/api/v1/settings/webhook/test",
          method: "post",
          data: webhookForm,
        });
        ElMessage.success("Webhook 测试发送成功");
      } catch (error: any) {
        ElMessage.error(
          "测试失败：" + (error.response?.data?.message || error.message)
        );
      } finally {
        loading.value = false;
      }
    }
  });
}

async function handleUpdateWebhook() {
  if (!webhookFormRef.value) return;

  await webhookFormRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true;
      try {
        await request({
          url: "/api/v1/settings/webhook",
          method: "post",
          data: webhookForm,
        });
        ElMessage.success("Webhook 设置保存成功");
      } catch (error: any) {
        ElMessage.error(
          "保存失败：" + (error.response?.data?.message || error.message)
        );
      } finally {
        loading.value = false;
      }
    }
  });
}
</script>

<template>
  <div class="personal-center">
    <el-row :gutter="20">
      <!-- Left Column: User Profile -->
      <el-col :xs="24" :sm="24" :md="8" :lg="6">
        <el-card class="profile-card">
          <template #header>
            <div class="card-header">
              <span>{{ $t("personalCenter.profileSection") }}</span>
            </div>
          </template>

          <div class="profile-content">
            <div class="avatar-wrapper">
              <el-avatar
                :size="120"
                :src="userinfo?.avatar + '?imageView2/1/w/200/h/200'"
              />
            </div>

            <div class="user-info">
              <h2>{{ userinfo?.username }}</h2>
              <p v-if="userinfo?.nickname" class="nickname">
                {{ userinfo.nickname }}
              </p>
            </div>

            <el-divider />

            <el-form
              ref="profileFormRef"
              :model="profileForm"
              label-position="top"
            >
              <el-form-item :label="$t('personalCenter.username')">
                <el-input
                  v-model="profileForm.username"
                  :placeholder="$t('personalCenter.username')"
                />
              </el-form-item>

              <el-form-item :label="$t('personalCenter.nickname')">
                <el-input
                  v-model="profileForm.nickname"
                  :placeholder="$t('personalCenter.nickname')"
                />
              </el-form-item>

              <el-form-item>
                <el-button
                  type="primary"
                  @click="handleUpdateProfile"
                  :loading="loading"
                  style="width: 100%"
                >
                  更新资料
                </el-button>
              </el-form-item>
            </el-form>
          </div>
        </el-card>
      </el-col>

      <!-- Right Column: Password Management -->
      <el-col :xs="24" :sm="24" :md="16" :lg="18">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>{{ $t("personalCenter.passwordSection") }}</span>
            </div>
          </template>

          <el-form
            ref="passwordFormRef"
            :model="passwordForm"
            :rules="passwordRules"
            label-width="140px"
            class="password-form"
          >
            <el-form-item
              :label="$t('personalCenter.oldPassword')"
              prop="oldPassword"
            >
              <el-input
                v-model="passwordForm.oldPassword"
                type="password"
                :placeholder="$t('personalCenter.oldPassword')"
                show-password
                autocomplete="off"
              />
            </el-form-item>

            <el-form-item
              :label="$t('personalCenter.newPassword')"
              prop="newPassword"
            >
              <el-input
                v-model="passwordForm.newPassword"
                type="password"
                :placeholder="$t('personalCenter.newPassword')"
                show-password
                autocomplete="off"
              />
            </el-form-item>

            <el-form-item
              :label="$t('personalCenter.confirmPassword')"
              prop="confirmPassword"
            >
              <el-input
                v-model="passwordForm.confirmPassword"
                type="password"
                :placeholder="$t('personalCenter.confirmPassword')"
                show-password
                autocomplete="off"
              />
            </el-form-item>

            <el-form-item>
              <el-button
                type="primary"
                @click="handleChangePassword"
                :loading="loading"
              >
                {{ $t("personalCenter.changePassword") }}
              </el-button>
              <el-button @click="passwordFormRef?.resetFields()">
                {{ $t("cancel") }}
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <!-- Webhook Settings -->
        <el-card style="margin-top: 20px">
          <template #header>
            <div class="card-header">
              <span>Webhook 设置</span>
              <el-switch
                v-model="webhookForm.webhookEnabled"
                active-text="启用"
                inactive-text="禁用"
              />
            </div>
          </template>

          <el-form
            ref="webhookFormRef"
            :model="webhookForm"
            :rules="webhookRules"
            label-width="140px"
            class="webhook-form"
          >
            <el-form-item label="Webhook URL" prop="webhookUrl">
              <el-input
                v-model="webhookForm.webhookUrl"
                placeholder="https://example.com/webhook"
              />
            </el-form-item>

            <el-form-item label="请求方法" prop="webhookMethod">
              <el-select
                v-model="webhookForm.webhookMethod"
                placeholder="Select"
              >
                <el-option label="POST" value="POST" />
                <el-option label="GET" value="GET" />
              </el-select>
            </el-form-item>

            <el-form-item label="Content-Type" prop="webhookContentType">
              <el-select
                v-model="webhookForm.webhookContentType"
                placeholder="Select"
              >
                <el-option label="application/json" value="application/json" />
                <el-option
                  label="application/x-www-form-urlencoded"
                  value="application/x-www-form-urlencoded"
                />
              </el-select>
            </el-form-item>

            <el-form-item label="Headers (JSON)" prop="webhookHeaders">
              <MonacoEditor
                v-model="webhookForm.webhookHeaders"
                language="json"
                height="400px"
                style="margin-bottom: 10px"
                placeholder='{
  "Authorization": "Bearer your-token",
  "Custom-Header": "value"
}'
              />
            </el-form-item>

            <el-form-item label="Body Template（JSON）" prop="webhookBody">
              <MonacoEditor
                v-model="webhookForm.webhookBody"
                language="json"
                height="400px"
                style="margin-bottom: 10px"
                placeholder='{
  "title": "{{title}}",
  "body": "{{message}}",
  "url": "https://example.com",
  "event": "{{event}}",
  "time": "{{time}}"
}'
              />
              <div class="form-tip" v-pre>
                <p>
                  支持在 <strong>URL</strong> 和
                  <strong>Body</strong> 中使用以下变量:
                </p>
                <ul>
                  <li>
                    <code>{{ title }}</code
                    >: 消息标题
                  </li>
                  <li>
                    <code>{{ message }}</code
                    >: 消息内容
                  </li>
                  <li>
                    <code>{{ event }}</code
                    >: 事件类型 (e.g., sub_update, speed_test)
                  </li>
                  <li>
                    <code>{{ time }}</code
                    >: 事件时间 (yyyy-MM-dd HH:mm:ss)
                  </li>
                  <li>
                    <code>{{json .}}</code
                    >: 完整 JSON 数据 (仅支持 Body)
                  </li>
                </ul>
                <p>
                  例如 Bark URL:
                  <code>https://api.day.app/key/{{ title }}/{{ message }}</code>
                </p>
              </div>
            </el-form-item>

            <el-form-item>
              <el-button
                type="success"
                @click="handleTestWebhook"
                :loading="loading"
              >
                测试 Webhook
              </el-button>
              <el-button
                type="primary"
                @click="handleUpdateWebhook"
                :loading="loading"
              >
                保存 Webhook 设置
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped lang="scss">
.personal-center {
  padding: 20px;
}

.profile-card {
  .profile-content {
    text-align: center;

    .avatar-wrapper {
      margin-bottom: 20px;
    }

    .user-info {
      h2 {
        margin: 10px 0 5px;
        font-size: 24px;
        color: var(--el-text-color-primary);
      }

      .nickname {
        margin: 0;
        color: var(--el-text-color-secondary);
        font-size: 14px;
      }
    }
  }
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
  font-size: 16px;
}

.password-form {
  max-width: 600px;
  margin: 0 auto;
}

@media (max-width: 768px) {
  .personal-center {
    padding: 10px;
  }

  .password-form {
    :deep(.el-form-item__label) {
      width: 100% !important;
      text-align: left;
    }

    :deep(.el-form-item__content) {
      margin-left: 0 !important;
    }
  }
}

.form-tip {
  margin-top: 5px;
  color: #909399;
  font-size: 12px;
  line-height: 1.5;

  ul {
    padding-left: 20px;
    margin: 5px 0;
  }

  code {
    background-color: #f4f4f5;
    padding: 2px 4px;
    border-radius: 4px;
    color: #c0392b;
  }
}
</style>
