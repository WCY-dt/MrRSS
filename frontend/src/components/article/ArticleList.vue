<script setup lang="ts">
import { useAppStore } from '@/stores/app';
import { useI18n } from 'vue-i18n';
import { ref, computed, onMounted, onBeforeUnmount, watch, type Ref } from 'vue';
import {
  PhCheckCircle,
  PhArrowClockwise,
  PhList,
  PhSpinner,
  PhFunnel,
  PhTrash,
} from '@phosphor-icons/vue';
import ArticleFilterModal from '../modals/filter/ArticleFilterModal.vue';
import ArticleItem from './ArticleItem.vue';
import { useArticleTranslation } from '@/composables/article/useArticleTranslation';
import { useArticleLabels } from '@/composables/article/useArticleLabels';
import { useArticleFilter } from '@/composables/article/useArticleFilter';
import { useArticleActions } from '@/composables/article/useArticleActions';
import type { Article } from '@/types/models';

const store = useAppStore();
const { t } = useI18n();

const listRef: Ref<HTMLDivElement | null> = ref(null);
const defaultViewMode = ref<'original' | 'rendered'>('original');
const showFilterModal = ref(false);
const isRefreshing = ref(false);
const savedScrollTop = ref(0);

interface Props {
  isSidebarOpen?: boolean;
}

defineProps<Props>();

const emit = defineEmits<{
  toggleSidebar: [];
}>();

// Use composables
const {
  translationSettings,
  loadTranslationSettings,
  setupIntersectionObserver,
  observeArticle,
  handleTranslationSettingsChange,
  cleanup: cleanupTranslation,
} = useArticleTranslation();

const {
  labelSettings,
  loadLabelSettings,
  setupIntersectionObserver: setupLabelObserver,
  observeArticle: observeLabelArticle,
  handleLabelSettingsChange,
  cleanup: cleanupLabels,
} = useArticleLabels();

const {
  activeFilters,
  filteredArticlesFromServer,
  isFilterLoading,
  resetFilterState,
  fetchFilteredArticles,
  loadMoreFilteredArticles,
} = useArticleFilter();

const { showArticleContextMenu } = useArticleActions(t, defaultViewMode, () =>
  store.fetchUnreadCounts()
);

// Computed filtered articles
const filteredArticles = computed(() => {
  // If filters are active, use server-filtered articles
  return activeFilters.value.length > 0 ? filteredArticlesFromServer.value : store.articles;
});

// Load settings and setup
onMounted(async () => {
  await loadTranslationSettings();
  await loadLabelSettings();

  try {
    const res = await fetch('/api/settings');
    const data = await res.json();
    defaultViewMode.value = data.default_view_mode || 'original';

    // Set up intersection observer for auto-translation
    if (translationSettings.value.enabled && listRef.value) {
      setupIntersectionObserver(listRef.value, store.articles);
    }

    // Set up intersection observer for auto-labeling
    if (labelSettings.value.enabled && listRef.value) {
      setupLabelObserver(listRef.value, store.articles);
    }
  } catch (e) {
    console.error('Error loading settings:', e);
  }

  // Listen for translation settings changes
  window.addEventListener(
    'translation-settings-changed',
    onTranslationSettingsChanged as EventListener
  );
  // Listen for label settings changes
  window.addEventListener('label-settings-changed', onLabelSettingsChanged as EventListener);
  // Listen for default view mode changes
  window.addEventListener('default-view-mode-changed', onDefaultViewModeChanged as EventListener);
  // Listen for refresh articles events
  window.addEventListener('refresh-articles', onRefreshArticles);
  // Listen for toggle filter events (from keyboard shortcut)
  window.addEventListener('toggle-filter', onToggleFilter);
});

// Watch for articles changes during refresh to maintain scroll position
watch(
  () => store.articles,
  () => {
    if (isRefreshing.value && listRef.value) {
      // Restore scroll position during refresh
      listRef.value.scrollTop = savedScrollTop.value;
    }
  },
  { deep: true }
);

// Watch for refresh completion to scroll to top
watch(
  () => store.refreshProgress.isRunning,
  (isRunning) => {
    if (!isRunning && isRefreshing.value) {
      // Refresh completed, scroll to top and reset state
      isRefreshing.value = false;
      if (listRef.value) {
        listRef.value.scrollTop = 0;
      }
    }
  }
);

onBeforeUnmount(() => {
  cleanupTranslation();
  cleanupLabels();
  window.removeEventListener(
    'translation-settings-changed',
    onTranslationSettingsChanged as EventListener
  );
  window.removeEventListener('label-settings-changed', onLabelSettingsChanged as EventListener);
  window.removeEventListener(
    'default-view-mode-changed',
    onDefaultViewModeChanged as EventListener
  );
  window.removeEventListener('refresh-articles', onRefreshArticles);
  window.removeEventListener('toggle-filter', onToggleFilter);
});

interface CustomEventDetail {
  mode?: string;
  enabled?: boolean;
  targetLang?: string;
  provider?: string;
  maxCount?: number;
}

// Event handlers
function onDefaultViewModeChanged(e: Event): void {
  const customEvent = e as CustomEvent<CustomEventDetail>;
  if (customEvent.detail.mode) {
    defaultViewMode.value = customEvent.detail.mode as 'original' | 'rendered';
  }
}

function onTranslationSettingsChanged(e: Event): void {
  const customEvent = e as CustomEvent<CustomEventDetail>;
  const { enabled, targetLang } = customEvent.detail;
  if (enabled !== undefined && targetLang) {
    handleTranslationSettingsChange(enabled, targetLang);

    // Re-setup observer if needed
    if (enabled && listRef.value) {
      setupIntersectionObserver(listRef.value, store.articles);
    }
  }
}

function onLabelSettingsChanged(e: Event): void {
  const customEvent = e as CustomEvent<CustomEventDetail>;
  const { enabled, provider, maxCount } = customEvent.detail;
  if (enabled !== undefined && provider && maxCount !== undefined) {
    handleLabelSettingsChange(enabled, provider, maxCount);

    // Re-setup observer if needed
    if (enabled && listRef.value) {
      setupLabelObserver(listRef.value, store.articles);
    }
  }
}

function onRefreshArticles(): void {
  store.fetchArticles();
}

function onToggleFilter(): void {
  showFilterModal.value = !showFilterModal.value;
}

// Article selection and interaction
function selectArticle(article: Article): void {
  // If clicking the same article, close the detail view
  if (store.currentArticleId === article.id) {
    store.currentArticleId = null;
    return;
  }

  // Reset user preference when selecting article via normal click
  window.dispatchEvent(new CustomEvent('reset-user-view-preference'));

  store.currentArticleId = article.id;
  if (!article.is_read) {
    article.is_read = true;
    fetch(`/api/articles/read?id=${article.id}&read=true`, { method: 'POST' })
      .then(() => {
        store.fetchUnreadCounts();
      })
      .catch((e) => {
        console.error('Error marking as read:', e);
      });
  }
}

// Observe article for both translation and labels
function observeArticleElement(el: Element | null): void {
  observeArticle(el);
  observeLabelArticle(el);
}

// Scrolling handler
function handleScroll(e: Event): void {
  const target = e.target as HTMLElement;
  const { scrollTop, clientHeight, scrollHeight } = target;
  if (scrollTop + clientHeight >= scrollHeight - 200) {
    if (activeFilters.value.length > 0) {
      loadMoreFilteredArticles();
    } else {
      store.loadMore();
    }
  }
}

// Filter handlers
async function handleApplyFilters(filters: typeof activeFilters.value): Promise<void> {
  activeFilters.value = filters;
  if (filters.length === 0) {
    resetFilterState();
    store.page = 1;
    await store.fetchArticles(false);
  } else {
    await fetchFilteredArticles(filters, false);
  }
}

// Actions
async function refreshArticles(): Promise<void> {
  // Save current scroll position and set refreshing state
  if (listRef.value) {
    savedScrollTop.value = listRef.value.scrollTop;
  }
  isRefreshing.value = true;

  await store.refreshFeeds();
  // Note: Scrolling to top is now handled by the watch on refreshProgress.isRunning
}

async function markAllAsRead(): Promise<void> {
  await store.markAllAsRead();
  window.showToast(t('markedAllAsRead'), 'success');
}

async function clearReadLater(): Promise<void> {
  try {
    const res = await fetch('/api/articles/clear-read-later', { method: 'POST' });
    if (res.ok) {
      await store.fetchArticles();
      window.showToast(t('clearedReadLater'), 'success');
    }
  } catch (e) {
    console.error('Error clearing read later:', e);
  }
}
</script>

<template>
  <section
    class="article-list flex flex-col w-full border-r border-border bg-bg-primary shrink-0 h-full"
  >
    <div class="p-2 sm:p-4 border-b border-border bg-bg-primary">
      <div class="flex items-center justify-between sm:mb-2">
        <h3 class="m-0 text-base sm:text-lg font-semibold">{{ t('articles') }}</h3>
        <div class="flex items-center gap-1 sm:gap-2">
          <!-- Clear Read Later button - only shown when viewing Read Later list -->
          <button
            v-if="store.currentFilter === 'readLater'"
            @click="clearReadLater"
            class="text-text-secondary hover:text-red-500 hover:bg-bg-tertiary p-1 sm:p-1.5 rounded transition-colors"
            :title="t('clearReadLater')"
          >
            <PhTrash :size="20" class="sm:w-6 sm:h-6" />
          </button>
          <button
            @click="markAllAsRead"
            class="text-text-secondary hover:text-text-primary hover:bg-bg-tertiary p-1 sm:p-1.5 rounded transition-colors"
            :title="t('markAllRead')"
          >
            <PhCheckCircle :size="20" class="sm:w-6 sm:h-6" />
          </button>
          <div class="relative">
            <button
              @click="showFilterModal = true"
              class="text-text-secondary hover:text-text-primary hover:bg-bg-tertiary p-1 sm:p-1.5 rounded transition-colors"
              :class="activeFilters.length > 0 ? 'filter-active' : ''"
              :title="t('filter')"
            >
              <PhFunnel :size="18" class="sm:w-5 sm:h-5" />
            </button>
            <div
              v-if="activeFilters.length > 0"
              class="absolute -top-1 -right-1 bg-accent text-white text-[9px] sm:text-[10px] font-bold rounded-full min-w-[14px] sm:min-w-[16px] h-3.5 sm:h-4 px-0.5 sm:px-1 flex items-center justify-center"
            >
              {{ activeFilters.length }}
            </div>
          </div>
          <div class="relative">
            <button
              @click="refreshArticles"
              class="text-text-secondary hover:text-text-primary hover:bg-bg-tertiary p-1 sm:p-1.5 rounded transition-colors"
              :title="t('refresh')"
            >
              <PhArrowClockwise
                :size="20"
                class="sm:w-6 sm:h-6"
                :class="store.refreshProgress.isRunning ? 'animate-spin' : ''"
              />
            </button>
            <div
              v-if="
                store.refreshProgress.isRunning &&
                store.refreshProgress.total > store.refreshProgress.current
              "
              class="absolute -top-1 -right-1 bg-accent text-white text-[9px] sm:text-[10px] font-bold rounded-full min-w-[14px] sm:min-w-[16px] h-3.5 sm:h-4 px-0.5 sm:px-1 flex items-center justify-center"
            >
              {{ store.refreshProgress.total - store.refreshProgress.current }}
            </div>
          </div>
          <button @click="emit('toggleSidebar')" class="md:hidden text-xl sm:text-2xl p-1">
            <PhList :size="20" class="sm:w-6 sm:h-6" />
          </button>
        </div>
      </div>
    </div>

    <div class="flex-1 overflow-y-auto" @scroll="handleScroll" ref="listRef">
      <div
        v-if="filteredArticles.length === 0 && !store.isLoading && !isFilterLoading"
        class="p-4 sm:p-5 text-center text-text-secondary text-sm sm:text-base"
      >
        {{ t('noArticles') }}
      </div>

      <ArticleItem
        v-for="article in filteredArticles"
        :key="article.id"
        :article="article"
        :isActive="store.currentArticleId === article.id"
        @click="selectArticle(article)"
        @contextmenu="(e) => showArticleContextMenu(e, article)"
        @observeElement="observeArticleElement"
      />

      <div
        v-if="store.isLoading || isFilterLoading"
        class="p-3 sm:p-4 text-center text-text-secondary"
      >
        <PhSpinner :size="20" class="animate-spin sm:w-6 sm:h-6" />
      </div>
    </div>

    <!-- Filter Modal -->
    <ArticleFilterModal
      :show="showFilterModal"
      :currentFilters="activeFilters"
      @close="showFilterModal = false"
      @apply="handleApplyFilters"
    />
  </section>
</template>

<style scoped>
@media (min-width: 768px) {
  .article-list {
    width: var(--article-list-width, 400px);
  }
}

.filter-active {
  @apply text-accent border-accent;
  background-color: rgba(59, 130, 246, 0.1);
}

.animate-spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
