<script setup lang="ts">
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhPlus, PhTrash } from '@phosphor-icons/vue';

interface Props {
  modelValue: string;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  'update:modelValue': [value: string];
}>();

const { t } = useI18n();
const newPrompt = ref('');

const prompts = computed<string[]>(() => {
  try {
    const parsed = JSON.parse(props.modelValue || '[]');
    return Array.isArray(parsed) ? parsed.filter((item): item is string => typeof item === 'string') : [];
  } catch {
    return [];
  }
});

function addPrompt() {
  const prompt = newPrompt.value.trim();
  if (!prompt || prompts.value.includes(prompt)) return;
  emit('update:modelValue', JSON.stringify([...prompts.value, prompt]));
  newPrompt.value = '';
}

function removePrompt(index: number) {
  emit('update:modelValue', JSON.stringify(prompts.value.filter((_, itemIndex) => itemIndex !== index)));
}
</script>

<template>
  <div class="w-full max-w-xl space-y-2">
    <div v-if="prompts.length > 0" class="space-y-1.5">
      <div
        v-for="(prompt, index) in prompts"
        :key="`${prompt}-${index}`"
        class="flex items-center gap-2 rounded-lg border border-border bg-bg-primary px-3 py-2"
      >
        <span class="min-w-0 flex-1 truncate text-sm text-text-primary" :title="prompt">
          {{ prompt }}
        </span>
        <button
          type="button"
          class="shrink-0 cursor-pointer rounded p-1 text-text-secondary hover:bg-bg-tertiary hover:text-red-500"
          :title="t('setting.ai.removeQuickPrompt')"
          @click="removePrompt(index)"
        >
          <PhTrash :size="16" />
        </button>
      </div>
    </div>

    <div class="flex gap-2">
      <input
        v-model="newPrompt"
        type="text"
        maxlength="500"
        class="min-w-0 flex-1 rounded-lg border border-border bg-bg-primary px-3 py-2 text-sm text-text-primary outline-none transition-colors focus:border-accent"
        :placeholder="t('setting.ai.quickPromptPlaceholder')"
        @keydown.enter.prevent="addPrompt"
      />
      <button
        type="button"
        class="btn-secondary shrink-0"
        :disabled="!newPrompt.trim()"
        @click="addPrompt"
      >
        <PhPlus :size="16" />
        {{ t('common.action.add') }}
      </button>
    </div>
    <p class="text-xs text-text-tertiary">{{ t('setting.ai.quickPromptsHint') }}</p>
  </div>
</template>
