<script setup lang="ts">
import { ref, onBeforeUnmount } from "vue";
import Editor from "monaco-editor-vue3";
import * as monaco from "monaco-editor";
import { getSuggestions } from "./suggestions";
// Import configuration to auto-configure Monaco Editor
import "./config";

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

const editorRef = ref();
let completionProvider: monaco.IDisposable | null = null;

const handleMount = (editor: any) => {
  editorRef.value = editor;

  // Register a custom completion provider for additional suggestions
  // This needs to be done per editor instance
  completionProvider = monaco.languages.registerCompletionItemProvider(
    "javascript",
    {
      triggerCharacters: ["."], // Trigger on dot notation
      provideCompletionItems: (model, position) => {
        const word = model.getWordUntilPosition(position);
        const range = {
          startLineNumber: position.lineNumber,
          endLineNumber: position.lineNumber,
          startColumn: word.startColumn,
          endColumn: word.endColumn,
        };

        // Get custom suggestions
        const customSuggestions = getSuggestions(range);

        return {
          suggestions: customSuggestions,
        };
      },
    }
  );
};

onBeforeUnmount(() => {
  if (completionProvider) {
    completionProvider.dispose();
  }
});

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
        // Validation and error display
        renderValidationDecorations: 'on', // Enable error squiggles
        // IntelliSense settings
        quickSuggestions: {
          other: true,
          comments: false,
          strings: true,
        },
        suggestOnTriggerCharacters: true,
        acceptSuggestionOnCommitCharacter: true,
        acceptSuggestionOnEnter: 'on',
        tabCompletion: 'on',
        wordBasedSuggestions: true,
        // Parameter hints
        parameterHints: {
          enabled: true,
          cycle: true,
        },
        // Hover information
        hover: {
          enabled: true,
          delay: 300,
        },
        // Code lens
        codeLens: true,
        // Minimap
        minimap: {
          enabled: true,
        },
        // Scrollbar
        scrollbar: {
          verticalScrollbarSize: 10,
          horizontalScrollbarSize: 10,
        },
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
