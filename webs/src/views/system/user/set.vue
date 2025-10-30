<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useUserStore } from "@/store";
import { updateUserPassword } from "@/api/user";
import { useI18n } from "vue-i18n";
// 创建 i18n 实例
const { t } = useI18n();
const userinfo = ref();
const userStore = useUserStore();
// 获取用户信息
onMounted(async () => {
  userinfo.value = await userStore.getUserInfo();
});
const username: Ref<string> = ref("");
const password: Ref<string> = ref("");

/** 重置密码 */
function resetPassword(row: { [key: string]: any }) {
  if (!username.value || !password.value) {
    ElMessage.error(t("userset.message.xx1"));
    return;
  }
  if (password.value.length < 6) {
    ElMessage.error(t("userset.message.xx2"));
    return;
  }
  ElMessageBox.confirm(t("userset.message.xx3"), t("userset.message.title"), {
    confirmButtonText: "OK",
    cancelButtonText: "Cancel",
    type: "warning",
  }).then(() => {
    updateUserPassword({
      username: username.value.trim(),
      password: password.value.trim(),
    }).then(() => {
      ElMessage.success(t("userset.message.xx4") + password.value);
      window.location.reload();
    });
  });
}
</script>

<template>
  <el-card style="margin: 10px; text-align: center">
    <el-row :gutter="20">
      <el-col :span="18">
        <h2>{{ $t("userset.title") }}</h2>
      </el-col>
      <el-col :span="18" v-if="userinfo">
        <el-badge :value="userinfo.username" class="item">
          <el-image :src="userinfo.avatar" />
        </el-badge>
      </el-col>

      <el-col :span="18">
        <el-input v-model="username" :placeholder="$t('userset.newUsername')" />
      </el-col>
      <el-col :span="18">
        <el-input
          type="password"
          v-model="password"
          :placeholder="$t('userset.newPassword')"
        />
      </el-col>
      <el-col :span="18">
        <el-button type="primary" @click="resetPassword">修改</el-button>
      </el-col>
    </el-row>
  </el-card>
</template>

<style scoped>
.el-col {
  margin-bottom: 10px;
}
</style>
