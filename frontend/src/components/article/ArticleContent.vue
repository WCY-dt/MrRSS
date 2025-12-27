<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount, computed, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhSpinnerGap, PhArticleNyTimes } from '@phosphor-icons/vue';
import type { Article } from '@/types/models';
import ArticleTitle from './parts/ArticleTitle.vue';
import ArticleSummary from './parts/ArticleSummary.vue';
import ArticleLoading from './parts/ArticleLoading.vue';
import ArticleBody from './parts/ArticleBody.vue';
import AudioPlayer from './parts/AudioPlayer.vue';
import VideoPlayer from './parts/VideoPlayer.vue';
import ArticleChatButton from './ArticleChatButton.vue';
import ArticleChatPanel from './ArticleChatPanel.vue';
import { useArticleSummary } from '@/composables/article/useArticleSummary';
import { useArticleTranslation } from '@/composables/article/useArticleTranslation';
import { useArticleRendering } from '@/composables/article/useArticleRendering';
import {
  extractTextWithPlaceholders,
  restorePreservedElements,
  hasOnlyPreservedContent,
} from '@/composables/article/useContentTranslation';
import { useSettings } from '@/composables/core/useSettings';
import { useAppStore } from '@/stores/app';
import { proxyImagesInHtml, isMediaCacheEnabled } from '@/utils/mediaProxy';
import './ArticleContent.css';

interface SummaryResult {
  summary: string;
  sentence_count: number;
  is_too_short: boolean;
  error?: string;
}

interface Props {
  article: Article;
  articleContent: string;
  isLoadingContent: boolean;
  attachImageEventListeners?: () => void;
  showTranslations?: boolean;
  showContent?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  showTranslations: true,
  attachImageEventListeners: undefined,
  showContent: true,
});

const emit = defineEmits<{
  retryLoadContent: [];
}>();

const { t } = useI18n();

// Handle retry load content
function handleRetryLoad() {
  emit('retryLoadContent');
}

// Chat state
const { settings: appSettings, fetchSettings } = useSettings();
const store = useAppStore();
const isChatPanelOpen = ref(false);

// Full-text fetching state
const isFetchingFullArticle = ref(false);
const fullArticleContent = ref('');
const autoShowAllContent = ref(false);

// Computed property to determine if auto-expand should be enabled for this feed
const shouldAutoExpandContent = computed(() => {
  // First check if feed has auto_expand_content setting
  const feed = store.feeds.find((f) => f.id === props.article.feed_id);
  if (feed?.auto_expand_content) {
    if (feed.auto_expand_content === 'enabled') return true;
    if (feed.auto_expand_content === 'disabled') return false;
    // If 'global', fall through to global setting
  }

  // Fall back to global setting
  return autoShowAllContent.value;
});

// Fetch settings on mount to get actual values
onMounted(async () => {
  try {
    const data = await fetchSettings();
    autoShowAllContent.value = data.auto_show_all_content === true;
  } catch (e) {
    console.error('Error fetching settings for chat:', e);
  }

  // Listen for auto show all content setting changes
  window.addEventListener(
    'auto-show-all-content-changed',
    onAutoShowAllContentChanged as EventListener
  );
});

// Computed to check if chat should be shown
const showChatButton = computed(() => {
  return (
    appSettings.value.ai_chat_enabled &&
    !props.isLoadingContent &&
    props.articleContent &&
    props.showContent
  );
});

// Computed to check if full-text fetching should be shown
const showFullTextButton = computed(() => {
  return (
    appSettings.value.full_text_fetch_enabled &&
    !props.isLoadingContent &&
    props.articleContent &&
    props.article?.url &&
    props.showContent &&
    !fullArticleContent.value // Don't show if we already have full content
  );
});

// Computed for the content to display (full article if available, otherwise RSS content)
const displayContent = computed(() => {
  return fullArticleContent.value || props.articleContent;
});

// Use composables for summary and translation
const {
  summarySettings,
  loadSummarySettings,
  generateSummary: generateSummaryComposable,
  isSummaryLoading,
} = useArticleSummary();

const { translationSettings, loadTranslationSettings } = useArticleTranslation();

// Use composable for enhanced rendering (math formulas, etc.)
const { enhanceRendering, renderMathFormulas, highlightCodeBlocks } = useArticleRendering();

// Computed properties for easier access
const summaryEnabled = computed(() => summarySettings.value.enabled);
const summaryProvider = computed(() => summarySettings.value.provider);
const summaryTriggerMode = computed(() => summarySettings.value.triggerMode);
const translationEnabled = computed(() => translationSettings.value.enabled);
const targetLanguage = computed(() => translationSettings.value.targetLang);

// Current article summary
const summaryResult = ref<SummaryResult | null>(null);
const isLoadingSummary = computed(() =>
  props.article ? isSummaryLoading(props.article.id) : false
);

// Additional state for summary translation
const translatedSummary = ref('');
const isTranslatingSummary = ref(false);

// Additional state for translation
const translatedTitle = ref('');
const isTranslatingTitle = ref(false);
const isTranslatingContent = ref(false);
const lastTranslatedArticleId = ref<number | null>(null);

// Load settings using composables
async function loadSettings() {
  await loadSummarySettings();
  await loadTranslationSettings();
}

// Translate text using the API
async function translateText(text: string): Promise<string> {
  if (!text || !translationEnabled.value) return '';

  try {
    const res = await fetch('/api/articles/translate-text', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        text: text,
        target_language: targetLanguage.value,
      }),
    });

    if (res.ok) {
      const data = await res.json();
      return data.translated_text || '';
    } else {
      console.error('Error translating text:', res.status);
      window.showToast(t('errorTranslatingContent'), 'error');
    }
  } catch (e) {
    console.error('Error translating text:', e);
    window.showToast(t('errorTranslating'), 'error');
  }
  return '';
}

// Fetch full article content from the original URL
async function fetchFullArticle() {
  if (!props.article?.id) return;

  isFetchingFullArticle.value = true;
  try {
    const res = await fetch(`/api/articles/fetch-full?id=${props.article.id}`, {
      method: 'POST',
    });

    if (res.ok) {
      const data = await res.json();
      let content = data.content || '';

      // Proxy images if media cache is enabled
      const cacheEnabled = await isMediaCacheEnabled();
      if (cacheEnabled && content) {
        // Use feed URL as referer for anti-hotlinking (more reliable than article URL)
        const feedUrl = data.feed_url || props.article.url;
        content = proxyImagesInHtml(content, feedUrl);
      }

      fullArticleContent.value = content;
      window.showToast(t('fullArticleFetched'), 'success');

      // After fetching full content, regenerate summary and trigger translation
      if (props.article) {
        if (shouldAutoGenerateSummary()) {
          setTimeout(() => generateSummary(props.article), 100);
        }
        if (translationEnabled.value) {
          // Reset translation tracking to allow re-translation with full content
          lastTranslatedArticleId.value = null;
          translateTitle(props.article);
          // Wait for DOM to update with new content before translating
          await nextTick();
          translateContentParagraphs(fullArticleContent.value);
        }
      }
    } else {
      console.error('Error fetching full article:', res.status);
      window.showToast(t('errorFetchingFullArticle'), 'error');
    }
  } catch (e) {
    console.error('Error fetching full article:', e);
    window.showToast(t('errorFetchingFullArticle'), 'error');
  } finally {
    isFetchingFullArticle.value = false;
  }
}

// Generate summary for the current article
async function generateSummary(article: Article, force: boolean = false) {
  if (!summaryEnabled.value || !article) {
    return;
  }

  // Only clear state if forcing regeneration
  if (force) {
    summaryResult.value = null;
    translatedSummary.value = '';
  }

  const result = await generateSummaryComposable(article, displayContent.value, force);
  summaryResult.value = result;

  // Update the article summary in store for caching
  if (result?.summary) {
    store.updateArticleSummary(article.id, result.summary);
  }

  // Auto-translate summary if translation is enabled
  // Only translate if we got a new summary (result is from API, not cached)
  if (translationEnabled.value && result?.summary && !result.is_too_short) {
    isTranslatingSummary.value = true;
    translatedSummary.value = await translateText(result.summary);
    isTranslatingSummary.value = false;
  }
}

// Check if should auto-generate summary
function shouldAutoGenerateSummary(): boolean {
  if (!summaryEnabled.value) return false;

  // For local provider, always auto-generate
  if (summaryProvider.value === 'local') return true;

  // For AI provider, check trigger mode
  if (summaryProvider.value === 'ai') {
    return summaryTriggerMode.value === 'auto';
  }

  return false;
}

// Translate title
async function translateTitle(article: Article) {
  if (!translationEnabled.value || !article?.title) return;

  isTranslatingTitle.value = true;
  translatedTitle.value = await translateText(article.title);
  isTranslatingTitle.value = false;
}

// Translate content paragraphs while preserving inline elements (formulas, code, images)
async function translateContentParagraphs(content: string) {
  if (!translationEnabled.value || !content) return;

  // Prevent duplicate translations for the same article
  if (lastTranslatedArticleId.value === props.article?.id) {
    return;
  }

  isTranslatingContent.value = true;
  lastTranslatedArticleId.value = props.article?.id || null;

  // Wait for content to render
  await nextTick();

  // Find all text elements in the prose content
  const proseContainer = document.querySelector('.prose-content');
  if (!proseContainer) {
    isTranslatingContent.value = false;
    return;
  }

  // Remove any existing translations first
  const existingTranslations = proseContainer.querySelectorAll('.translation-text');
  existingTranslations.forEach((el) => el.remove());

  // Find all translatable elements
  // For lists: translate individual li items, translation stays inside the same li
  // For tables: translate td/th cells, translation stays inside the same cell
  // For blockquotes: translate inner paragraphs, not the blockquote itself
  const textTags = [
    'P',
    'H1',
    'H2',
    'H3',
    'H4',
    'H5',
    'H6',
    'LI',
    'TD',
    'TH',
    'FIGCAPTION',
    'DT',
    'DD',
  ];
  const elements = proseContainer.querySelectorAll(textTags.join(','));

  // Track which elements we've already translated to avoid duplicates
  const translatedElements = new Set<HTMLElement>();

  for (const el of elements) {
    const htmlEl = el as HTMLElement;

    // Skip if inside a translation element
    if (htmlEl.closest('.translation-text')) continue;

    // Skip if already has translation inside
    if (htmlEl.querySelector('.translation-text')) continue;

    // Skip if we've already translated this element
    if (translatedElements.has(htmlEl)) continue;

    // Skip if this element's parent was already translated
    // This prevents duplicate translations of nested elements
    let hasTranslatedAncestor = false;
    let ancestor = htmlEl.parentElement;
    while (ancestor && ancestor !== proseContainer) {
      if (translatedElements.has(ancestor)) {
        hasTranslatedAncestor = true;
        break;
      }
      ancestor = ancestor.parentElement;
    }
    if (hasTranslatedAncestor) continue;

    // Skip elements that are entirely technical content (no translatable text)
    if (
      htmlEl.closest('pre') ||
      htmlEl.tagName === 'CODE' ||
      htmlEl.closest('kbd') ||
      htmlEl.classList.contains('katex') ||
      htmlEl.classList.contains('katex-display') ||
      htmlEl.classList.contains('katex-inline')
    ) {
      continue;
    }

    // Skip elements that only contain preserved content (no translatable text)
    if (hasOnlyPreservedContent(htmlEl)) {
      continue;
    }

    // Extract text with placeholders for inline elements (formulas, code, images) and hyperlinks
    const {
      text: textWithPlaceholders,
      preservedElements,
      hyperlinks,
    } = extractTextWithPlaceholders(htmlEl);

    if (!textWithPlaceholders || textWithPlaceholders.length < 2) continue;

    // Translate the text (with placeholders and link markers)
    const translatedText = await translateText(textWithPlaceholders);

    // Skip if translation is same as original or empty
    if (!translatedText || translatedText === textWithPlaceholders) continue;

    // Restore preserved elements and hyperlinks in the translated text
    const translatedHTML = restorePreservedElements(translatedText, preservedElements, hyperlinks);

    // Determine how to insert translation based on element type
    const tagName = htmlEl.tagName;

    if (
      tagName === 'LI' ||
      tagName === 'TD' ||
      tagName === 'TH' ||
      tagName === 'DD' ||
      tagName === 'DT'
    ) {
      // For list items, table cells, definition list items: append translation inside the same element
      const translationEl = document.createElement('div');
      translationEl.className = 'translation-text translation-inline';
      translationEl.innerHTML = translatedHTML;
      htmlEl.appendChild(translationEl);
    } else if (htmlEl.closest('blockquote')) {
      // For elements inside blockquote: append translation inside, styled differently
      const translationEl = document.createElement('div');
      translationEl.className = 'translation-text translation-blockquote';
      translationEl.innerHTML = translatedHTML;
      htmlEl.appendChild(translationEl);
    } else {
      // For standalone paragraphs, headings, figcaption: insert after as sibling
      const translationEl = document.createElement('div');
      translationEl.className = 'translation-text';
      translationEl.innerHTML = translatedHTML;
      htmlEl.parentNode?.insertBefore(translationEl, htmlEl.nextSibling);
    }

    // Mark this element as translated
    translatedElements.add(htmlEl);
  }

  // Re-apply rendering enhancements to translation elements (for math formulas)
  await nextTick();
  proseContainer.querySelectorAll('.translation-text').forEach((el) => {
    renderMathFormulas(el as HTMLElement);
    highlightCodeBlocks(el as HTMLElement);
  });

  // Re-attach ALL event listeners after translation modifies the DOM
  // This includes unwrapping images from links, attaching image handlers, and link handlers
  await reattachImageInteractions();

  isTranslatingContent.value = false;
}

async function reattachImageInteractions() {
  if (!props.attachImageEventListeners || !props.articleContent) return;
  await nextTick();
  props.attachImageEventListeners();
}

// Handle auto show all content setting change
function onAutoShowAllContentChanged(e: Event): void {
  const customEvent = e as CustomEvent<{ value: boolean }>;
  autoShowAllContent.value = customEvent.detail.value;
}

// Watch for article changes and regenerate summary + translations
watch(
  () => props.article?.id,
  async (newId, oldId) => {
    if (newId !== oldId) {
      summaryResult.value = null;
      translatedSummary.value = '';
      translatedTitle.value = '';
      lastTranslatedArticleId.value = null; // Reset translation tracking
      fullArticleContent.value = ''; // Reset full article content when switching articles

      if (props.article) {
        // Check if article has a cached summary first
        if (props.article.summary && props.article.summary.trim() !== '') {
          // Load the cached summary immediately
          summaryResult.value = {
            summary: props.article.summary,
            sentence_count: 0,
            is_too_short: false,
          };

          // Translate the cached summary if translation is enabled
          if (translationEnabled.value) {
            isTranslatingSummary.value = true;
            translatedSummary.value = await translateText(props.article.summary);
            isTranslatingSummary.value = false;
          }
        } else if (shouldAutoGenerateSummary()) {
          // Only auto-generate if no cached summary exists
          setTimeout(() => generateSummary(props.article), 100);
        }

        // Translate title
        if (translationEnabled.value) {
          translateTitle(props.article);
        }
      }
    }
  }
);

// Watch for content loading completion only
watch(
  () => props.isLoadingContent,
  async (isLoading, wasLoading) => {
    if (wasLoading && !isLoading && props.article) {
      // Enhance rendering first (math formulas, etc.)
      await nextTick();
      enhanceRendering('.prose-content');

      // Re-attach image event listeners after rendering enhancements
      await reattachImageInteractions();

      // Auto-fetch full article if setting is enabled
      if (shouldAutoExpandContent.value && !fullArticleContent.value) {
        setTimeout(() => fetchFullArticle(), 200);
      }

      // Delay summary generation to prioritize content display
      if (shouldAutoGenerateSummary()) {
        setTimeout(() => generateSummary(props.article), 100);
      }
      if (translationEnabled.value && lastTranslatedArticleId.value !== props.article.id) {
        await nextTick();
        translateContentParagraphs(props.articleContent);
      }
    }
  }
);

onMounted(async () => {
  await loadSettings();
  if (props.article) {
    // Enhance rendering if content is already loaded
    if (props.articleContent && !props.isLoadingContent) {
      await nextTick();
      enhanceRendering('.prose-content');
      // Re-attach image event listeners after rendering
      await reattachImageInteractions();

      // Auto-fetch full article if setting is enabled and content is already loaded
      if (
        shouldAutoExpandContent.value &&
        !fullArticleContent.value &&
        !isFetchingFullArticle.value
      ) {
        setTimeout(() => fetchFullArticle(), 200);
      }
    }

    // Check for cached summary first
    if (props.article.summary && props.article.summary.trim() !== '') {
      // Load the cached summary immediately
      summaryResult.value = {
        summary: props.article.summary,
        sentence_count: 0,
        is_too_short: false,
      };

      // Translate the cached summary if translation is enabled
      if (translationEnabled.value) {
        isTranslatingSummary.value = true;
        translatedSummary.value = await translateText(props.article.summary);
        isTranslatingSummary.value = false;
      }
    } else if (shouldAutoGenerateSummary() && props.articleContent) {
      // Only auto-generate if no cached summary exists
      setTimeout(() => generateSummary(props.article), 100);
    }

    if (translationEnabled.value) {
      translateTitle(props.article);
      if (props.articleContent && !props.isLoadingContent) {
        await nextTick();
        translateContentParagraphs(props.articleContent);
      }
    }
  }
});

// Ensure image interactions stay attached when content is (re)rendered
watch(
  () => props.articleContent,
  async (content) => {
    if (content) {
      // Wait for v-html to update the DOM before attaching event listeners
      await nextTick();
      await reattachImageInteractions();
    }
  },
  { immediate: true }
);

// Clean up event listeners
onBeforeUnmount(() => {
  window.removeEventListener(
    'auto-show-all-content-changed',
    onAutoShowAllContentChanged as EventListener
  );
});
</script>

<template>
  <div class="flex-1 overflow-y-auto bg-bg-primary p-3 sm:p-6">
    <div
      class="max-w-3xl mx-auto bg-bg-primary"
      :class="{ 'hide-translations': !showTranslations }"
    >
      <ArticleTitle
        :article="article"
        :translated-title="translatedTitle"
        :is-translating-title="isTranslatingTitle"
        :translation-enabled="translationEnabled"
      />

      <!-- Audio Player (if article has audio) -->
      <AudioPlayer
        v-if="article.audio_url"
        :audio-url="article.audio_url"
        :article-title="article.title"
      />

      <!-- Video Player (if article has video) -->
      <VideoPlayer
        v-if="article.video_url"
        :video-url="article.video_url"
        :article-title="article.title"
      />

      <ArticleSummary
        :summary-result="summaryResult"
        :is-loading-summary="isLoadingSummary"
        :translated-summary="translatedSummary"
        :is-translating-summary="isTranslatingSummary"
        :translation-enabled="translationEnabled"
        :summary-provider="summaryProvider"
        :summary-trigger-mode="summaryTriggerMode"
        :is-loading-content="props.isLoadingContent"
        @generate-summary="generateSummary(props.article, true)"
      />

      <ArticleLoading v-if="isLoadingContent" />

      <ArticleBody
        v-else
        :article-content="displayContent"
        :is-translating-content="isTranslatingContent"
        :has-media-content="!!(article.audio_url || article.video_url)"
        :is-loading-content="isLoadingContent"
        @retry-load="handleRetryLoad"
      />

      <!-- Full-text fetch button -->
      <div v-if="showFullTextButton" class="flex justify-center mt-4 mb-4">
        <button
          :disabled="isFetchingFullArticle"
          class="btn-secondary-compact flex items-center gap-1.5"
          @click="fetchFullArticle"
        >
          <PhSpinnerGap v-if="isFetchingFullArticle" :size="12" class="animate-spin" />
          <PhArticleNyTimes v-else :size="12" />
          <span class="text-xs">{{
            isFetchingFullArticle ? t('fetchingFullArticle') : t('fetchFullArticle')
          }}</span>
        </button>
      </div>
    </div>

    <!-- Chat Button (shown when content is loaded and chat is enabled) -->
    <ArticleChatButton v-if="showChatButton && !isChatPanelOpen" @click="isChatPanelOpen = true" />

    <!-- Chat Panel -->
    <ArticleChatPanel
      v-if="isChatPanelOpen"
      :article="article"
      :article-content="articleContent"
      :settings="{ ai_chat_enabled: appSettings.ai_chat_enabled }"
      @close="isChatPanelOpen = false"
    />
  </div>
</template>

<style scoped>
@reference "../../../../style.css";

.btn-secondary {
  @apply bg-bg-tertiary border border-border text-text-primary px-3 sm:px-4 py-1.5 sm:py-2 rounded-md cursor-pointer flex items-center gap-1.5 sm:gap-2 font-medium hover:bg-bg-secondary transition-colors;
}

.btn-secondary-compact {
  @apply bg-bg-tertiary border border-border text-text-primary px-2 py-1 rounded cursor-pointer flex items-center gap-1.5 text-xs hover:bg-bg-secondary transition-colors opacity-70 hover:opacity-100;
}
</style>
