<template>
  <el-config-provider :locale="locale" :size="size">
    <!-- 开启水印 -->
    <el-watermark
      v-if="watermarkEnabled"
      :font="{ color: fontColor }"
      :content="defaultSettings.watermarkContent"
      class="wh-full"
    >
      <router-view />
    </el-watermark>
    <!-- 关闭水印 -->
    <router-view v-else />
  </el-config-provider>
</template>

<script setup lang="ts">
import { useAppStore, useSettingsStore, useUserStore } from "@/store";
import defaultSettings from "@/settings";
import { ThemeEnum } from "@/enums/ThemeEnum";
import { SizeEnum } from "@/enums/SizeEnum";
import { useVersionStore } from "@/store/modules/version";

const appStore = useAppStore();
const settingsStore = useSettingsStore();
const userStore = useUserStore();

const locale = computed(() => appStore.locale);
const size = computed(() => appStore.size as SizeEnum);
const watermarkEnabled = computed(() => settingsStore.watermarkEnabled);

const versionStore = useVersionStore();

async function fetchVersion() {
  try {
    await versionStore.clearVersion();
    await versionStore.getVersion();
    console.log("版本获取成功:", versionStore.version);
  } catch (error) {
    console.error("版本获取失败");
  }
}
onMounted(async () => {
  await fetchVersion();
  userStore.connectSSE();
});
// 明亮/暗黑主题水印字体颜色适配
const fontColor = computed(() => {
  return settingsStore.theme === ThemeEnum.DARK
    ? "rgba(255, 255, 255, .15)"
    : "rgba(0, 0, 0, .15)";
});
</script>
