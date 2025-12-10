<script setup lang="ts">
import { computed, ref, onMounted, watch } from 'vue';
import { PhSpinnerGap, PhTranslate } from '@phosphor-icons/vue';
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
  labelEnabled: boolean;
  isLabeling: boolean;
}

const props = defineProps<Props>();

const { t } = useI18n();
const { parseLabels } = useArticleLabels();

const labelShowInList = ref(false);
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
    labelShowInList.value = settings.label_show_in_list === 'true';
    currentLabels.value = parseLabels(props.article.labels);
  } catch (e) {
    console.error('Failed to load label settings:', e);
  }
});

// Watch for label changes from auto-generation
watch(
  () => props.article.labels,
  (newLabels) => {
    currentLabels.value = parseLabels(newLabels);
  }
);
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
    <span>{{ formatDate(article.published_at, $i18n.locale) }}</span>
    <span v-if="translationEnabled" class="flex items-center gap-1 text-accent">
      <PhTranslate :size="14" />
      <span class="text-xs">{{ t('autoTranslateEnabled') }}</span>
    </span>
    <!-- Label generation indicator -->
    <div v-if="isLabeling" class="flex items-center gap-1 text-text-secondary">
      <PhSpinnerGap :size="12" class="animate-spin" />
      <span class="text-xs">{{ t('generatingLabels') }}</span>
    </div>
  </div>

  <!-- Labels Section -->
  <div
    v-if="labelEnabled && currentLabels.length > 0"
    class="mb-4 flex flex-wrap items-center gap-2"
  >
    <ArticleLabels :labelsJson="JSON.stringify(currentLabels)" :maxDisplay="10" size="md" />
  </div>
</template>
