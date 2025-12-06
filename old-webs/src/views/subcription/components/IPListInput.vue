<template>
  <div class="ip-list-input">
    <p>{{ title }}</p>
    <textarea
      v-model="inputValue"
      rows="5"
      class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
      placeholder="例如：&#10;192.168.1.1&#10;10.0.0.0/8&#10;2001:db8::/32"
      @input="handleInput"
    ></textarea>
    <div class="mt-2 text-sm text-gray-500">
      每行输入一个IP地址或CIDR格式，支持IPv4和IPv6
    </div>
    <div v-if="errors.length > 0" class="mt-2">
      <div
        v-for="(error, index) in errors"
        :key="index"
        class="text-sm text-red-600"
      >
        {{ error }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";

interface Props {
  modelValue?: string; // 数据库存储格式：逗号分隔的字符串
  title?: string;
}

interface Emits {
  (e: "update:modelValue", value: string): void; // 输出逗号分隔的字符串用于数据库存储
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: "",
  title: "",
});

const emit = defineEmits<Emits>();

// 将逗号分隔的字符串格式化为文本框中的多行文本
const formatDbValueToInput = (value: string): string => {
  if (!value) return "";
  return value.split(",").join("\n");
};

// 将文本框中的多行文本解析为逗号分隔的字符串
const formatInputToDbValue = (input: string): string => {
  return input
    .split("\n")
    .map((line) => line.trim())
    .filter((line) => line.length > 0)
    .join(",");
};

// 验证IP地址或CIDR格式
const validateIP = (ip: string): boolean => {
  // IPv4 正则表达式 (支持CIDR)
  const ipv4Regex = /^(\d{1,3}\.){3}\d{1,3}(\/([0-9]|[1-2][0-9]|3[0-2]))?$/;

  // IPv6 正则表达式 (支持CIDR)
  const ipv6Regex =
    /^([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}(\/([0-9]|[1-9][0-9]|1[0-1][0-9]|12[0-8]))?$|^::1(\/([0-9]|[1-9][0-9]|1[0-1][0-9]|12[0-8]))?$|^::$/;

  // 简单的IPv4 CIDR验证
  if (ipv4Regex.test(ip)) {
    if (ip.includes("/")) {
      const [address] = ip.split("/");
      return validateIPv4Address(address);
    }
    return validateIPv4Address(ip);
  }

  // 简单的IPv6 CIDR验证
  return ipv6Regex.test(ip);
};

// 验证IPv4地址格式
const validateIPv4Address = (address: string): boolean => {
  const parts = address.split(".");
  if (parts.length !== 4) return false;
  for (const part of parts) {
    const num = parseInt(part, 10);
    if (isNaN(num) || num < 0 || num > 255) return false;
  }
  return true;
};

const inputValue = ref(formatDbValueToInput(props.modelValue));
const errors = ref<string[]>([]);

const handleInput = () => {
  const ips = inputValue.value
    .split("\n")
    .map((line) => line.trim())
    .filter((line) => line.length > 0);

  errors.value = [];

  // 验证每个IP
  const invalidIps: string[] = [];
  for (const ip of ips) {
    if (!validateIP(ip)) {
      invalidIps.push(ip);
    }
  }

  if (invalidIps.length > 0) {
    errors.value.push(`以下IP格式不正确: ${invalidIps.join(", ")}`);
  }

  // 如果没有错误则更新modelValue（逗号分隔的格式）
  if (invalidIps.length === 0) {
    const dbValue = formatInputToDbValue(inputValue.value);
    emit("update:modelValue", dbValue);
  }
};

// 监听外部modelValue变化
watch(
  () => props.modelValue,
  (newVal) => {
    inputValue.value = formatDbValueToInput(newVal || "");
  }
);
</script>

<style scoped>
.ip-list-input {
  margin-bottom: 1rem;
}
</style>
