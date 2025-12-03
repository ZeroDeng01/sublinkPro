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
</style>
