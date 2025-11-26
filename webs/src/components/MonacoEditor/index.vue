<script setup lang="ts">
import { ref, watch, shallowRef } from "vue";
import Editor from "monaco-editor-vue3";

const props = defineProps({
  modelValue: {
    type: String,
    default: "",
  },
  language: {
    type: String,
    default: "javascript",
  },
  theme: {
    type: String,
    default: "vs-dark",
  },
  options: {
    type: Object,
    default: () => ({}),
  },
});

const emit = defineEmits(["update:modelValue", "change"]);

const editorRef = shallowRef();
const handleMount = (editor: any) => {
  editorRef.value = editor;
};

const onChange = (val: string) => {
  emit("update:modelValue", val);
  emit("change", val);
};
</script>

<template>
  <div class="monaco-editor-container">
    <Editor
      :value="modelValue"
      :language="language"
      :theme="theme"
      :options="{
        automaticLayout: true,
        formatOnType: true,
        formatOnPaste: true,
        ...options,
      }"
      @mount="handleMount"
      @change="onChange"
      class="editor"
    />
  </div>
</template>

<style scoped>
.monaco-editor-container {
  width: 100%;
  height: 400px; /* Default height */
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  overflow: hidden;
}

.editor {
  width: 100%;
  height: 100%;
}
</style>
