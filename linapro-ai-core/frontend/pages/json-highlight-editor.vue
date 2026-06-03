<script setup lang="ts">
import { computed, ref } from 'vue';

const props = withDefaults(
  defineProps<{
    invalid?: boolean;
    modelValue?: string;
    placeholder?: string;
    testid?: string;
  }>(),
  {
    invalid: false,
    modelValue: '',
    placeholder: '',
    testid: 'json-highlight-editor',
  },
);

const emit = defineEmits<{
  'update:modelValue': [value: string];
}>();

const highlightRef = ref<HTMLElement>();

function escapeHtml(value: string) {
  return value
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#39;');
}

function tokenClass(token: string) {
  if (/^(&quot;|")/.test(token)) {
    return /:\s*$/.test(token) ? 'json-token-key' : 'json-token-string';
  }
  if (/^-?\d/.test(token)) {
    return 'json-token-number';
  }
  if (token === 'true' || token === 'false') {
    return 'json-token-boolean';
  }
  if (token === 'null') {
    return 'json-token-null';
  }
  return 'json-token-punctuation';
}

function highlightJson(value: string) {
  const escaped = escapeHtml(value || ' ');
  const tokenPattern =
    /(&quot;(?:\\u[\dA-Fa-f]{4}|\\[^u]|[^\\&])*&quot;\s*:|&quot;(?:\\u[\dA-Fa-f]{4}|\\[^u]|[^\\&])*&quot;|\btrue\b|\bfalse\b|\bnull\b|-?\d+(?:\.\d+)?(?:[Ee][+-]?\d+)?|[{}[\],:])/g;

  return escaped.replace(tokenPattern, (token) => {
    return `<span class="${tokenClass(token)}">${token}</span>`;
  });
}

const highlightedValue = computed(() => highlightJson(props.modelValue));

function handleInput(event: Event) {
  emit('update:modelValue', (event.target as HTMLTextAreaElement).value);
}

function handleScroll(event: Event) {
  if (!highlightRef.value) {
    return;
  }
  const target = event.target as HTMLTextAreaElement;
  highlightRef.value.scrollLeft = target.scrollLeft;
  highlightRef.value.scrollTop = target.scrollTop;
}
</script>

<template>
  <div
    :class="{ 'is-invalid': invalid }"
    :data-testid="testid"
    class="json-highlight-editor"
  >
    <pre
      ref="highlightRef"
      aria-hidden="true"
      class="json-highlight-editor__highlight"
      v-html="highlightedValue"
    ></pre>
    <textarea
      :data-testid="`${testid}-input`"
      :placeholder="placeholder"
      :spellcheck="false"
      :value="modelValue"
      class="json-highlight-editor__input"
      wrap="off"
      @input="handleInput"
      @scroll="handleScroll"
    ></textarea>
  </div>
</template>

<style scoped>
.json-highlight-editor {
  position: relative;
  height: 220px;
  overflow: hidden;
  border: 1px solid hsl(var(--border));
  border-radius: 6px;
  background: hsl(var(--background));
  transition:
    border-color 0.16s ease,
    box-shadow 0.16s ease;
}

.json-highlight-editor:focus-within {
  border-color: hsl(var(--primary));
  box-shadow: 0 0 0 2px hsl(var(--primary) / 12%);
}

.json-highlight-editor.is-invalid {
  border-color: hsl(var(--destructive));
}

.json-highlight-editor__highlight,
.json-highlight-editor__input {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  margin: 0;
  padding: 12px 14px;
  border: 0;
  font-family:
    ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono',
    'Courier New', monospace;
  font-size: 12px;
  line-height: 1.6;
  tab-size: 2;
  white-space: pre;
}

.json-highlight-editor__highlight {
  position: absolute;
  inset: 0;
  overflow: hidden;
  color: hsl(var(--foreground));
  pointer-events: none;
}

.json-highlight-editor__input {
  position: relative;
  z-index: 1;
  display: block;
  overflow: auto;
  color: transparent;
  caret-color: hsl(var(--foreground));
  background: transparent;
  outline: none;
  resize: none;
}

.json-highlight-editor__input::placeholder {
  color: hsl(var(--muted-foreground));
  -webkit-text-fill-color: hsl(var(--muted-foreground));
}

.json-highlight-editor__input::selection {
  color: transparent;
  background: hsl(var(--primary) / 22%);
}

.json-highlight-editor__highlight :deep(.json-token-key) {
  color: #2563eb;
}

.json-highlight-editor__highlight :deep(.json-token-string) {
  color: #15803d;
}

.json-highlight-editor__highlight :deep(.json-token-number) {
  color: #b45309;
}

.json-highlight-editor__highlight :deep(.json-token-boolean) {
  color: #7c3aed;
}

.json-highlight-editor__highlight :deep(.json-token-null) {
  color: #64748b;
}

.json-highlight-editor__highlight :deep(.json-token-punctuation) {
  color: hsl(var(--muted-foreground));
}

html[class='dark'] .json-highlight-editor__highlight :deep(.json-token-key) {
  color: #93c5fd;
}

html[class='dark']
  .json-highlight-editor__highlight
  :deep(.json-token-string) {
  color: #86efac;
}

html[class='dark']
  .json-highlight-editor__highlight
  :deep(.json-token-number) {
  color: #fbbf24;
}

html[class='dark']
  .json-highlight-editor__highlight
  :deep(.json-token-boolean) {
  color: #c4b5fd;
}
</style>
