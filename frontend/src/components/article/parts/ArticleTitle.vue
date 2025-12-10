<script setup lang="ts">
import { computed, ref, onMounted } from 'vue';
import { PhSpinnerGap, PhTranslate, PhTag } from '@phosphor-icons/vue';
import type { Article } from '@/types/models';
import { formatDate } from '@/utils/date';
import { useI18n } from 'vue-i18n';
import { useArticleLabels } from '@/composables/article/useArticleLabels';
import ArticleLabels from './ArticleLabels.vue';

interface Props {
  article: Article;
  translatedTitle: string;
  isTranslatingTitle: boolean;
  translationEnabled: boolean;
}

const props = defineProps<Props>();

const { t } = useI18n();
const { isGeneratingLabels, generateLabels, parseLabels } = useArticleLabels();

const labelEnabled = ref(false);
const currentLabels = ref<string[]>([]);

// Computed: check if we should show bilingual title
const showBilingualTitle = computed(() => {
  return (
    props.translationEnabled &&
    props.translatedTitle &&
    props.translatedTitle !== props.article?.title
  );
});

onMounted(async () => {
  // Load label settings
  try {
    const res = await fetch('/api/settings');
    const settings = await res.json();
    labelEnabled.value = settings.label_enabled === 'true';
    currentLabels.value = parseLabels(props.article.labels);
  } catch (e) {
    console.error('Failed to load label settings:', e);
  }
});

async function handleGenerateLabels() {
  try {
    const labels = await generateLabels(props.article.id);
    currentLabels.value = labels;
    // Emit event to parent to update article data instead of mutating prop directly
    // Note: For now, we just update the local state. Parent component refresh will sync data.
    window.showToast(t('generateLabels') + ' - ' + t('success'), 'success');
  } catch (error: unknown) {
    const message = error instanceof Error ? error.message : String(error);
    console.error('Failed to generate labels:', error);
    window.showToast(message || t('generateLabels') + ' - ' + t('failed'), 'error');
  }
}
</script>

<template>
  <!-- Title Section - Bilingual when translation enabled -->
  <div class="mb-3 sm:mb-4">
    <!-- Original Title -->
    <h1 class="text-xl sm:text-3xl font-bold leading-tight text-text-primary">
      {{ article.title }}
    </h1>
    <!-- Translated Title (shown below if different from original) -->
    <h2
      v-if="showBilingualTitle"
      class="text-base sm:text-xl font-medium leading-tight mt-2 text-text-secondary"
    >
      {{ translatedTitle }}
    </h2>
    <!-- Translation loading indicator for title -->
    <div v-if="isTranslatingTitle" class="flex items-center gap-1 mt-1 text-text-secondary">
      <PhSpinnerGap :size="12" class="animate-spin" />
      <span class="text-xs">Translating...</span>
    </div>
  </div>

  <div
    class="text-xs sm:text-sm text-text-secondary mb-4 sm:mb-6 flex flex-wrap items-center gap-2 sm:gap-4"
  >
    <span>{{ article.feed_title }}</span>
    <span class="hidden sm:inline">•</span>
    <span>{{ formatDate(article.published_at, $i18n.locale.value) }}</span>
    <span v-if="translationEnabled" class="flex items-center gap-1 text-accent">
      <PhTranslate :size="14" />
      <span class="text-xs">{{ t('autoTranslateEnabled') }}</span>
    </span>
  </div>

  <!-- Labels Section -->
  <div v-if="labelEnabled" class="mb-4 flex flex-wrap items-center gap-2">
    <ArticleLabels
      v-if="currentLabels.length > 0"
      :labelsJson="JSON.stringify(currentLabels)"
      :maxDisplay="10"
      size="md"
    />
    <button
      @click="handleGenerateLabels"
      :disabled="isGeneratingLabels"
      class="label-generate-btn"
      :class="{ loading: isGeneratingLabels }"
    >
      <PhSpinnerGap v-if="isGeneratingLabels" :size="14" class="animate-spin" />
      <PhTag v-else :size="14" />
      <span class="text-xs">{{
        isGeneratingLabels ? t('generatingLabels') : t('generateLabels')
      }}</span>
    </button>
  </div>
</template>

<style scoped>
.label-generate-btn {
  @apply flex items-center gap-1.5 px-2.5 py-1.5 rounded-md border border-border bg-bg-secondary text-text-secondary hover:bg-bg-tertiary hover:border-accent transition-all;
}

.label-generate-btn:disabled {
  @apply opacity-70 cursor-not-allowed;
}

.label-generate-btn.loading {
  @apply pointer-events-none;
}
</style>
