<script setup lang="ts">
import { PhNewspaper } from '@phosphor-icons/vue';
import { useArticleDetail } from '@/composables/article/useArticleDetail';
import ArticleToolbar from './ArticleToolbar.vue';
import ArticleContent from './ArticleContent.vue';
import ImageViewer from '../common/ImageViewer.vue';
import FindInPage from '../common/FindInPage.vue';

import { ref, onMounted, onBeforeUnmount } from 'vue';

const {
  article,
  showContent,
  articleContent,
  isLoadingContent,
  imageViewerSrc,
  imageViewerAlt,
  close,
  toggleRead,
  toggleFavorite,
  toggleReadLater,
  openOriginal,
  toggleContentView,
  closeImageViewer,
  attachImageEventListeners,
  exportToObsidian,
  handleRetryLoadContent,
  t,
} = useArticleDetail();

const showTranslations = ref(true);
const showFindInPage = ref(false);

function toggleTranslations() {
  showTranslations.value = !showTranslations.value;
}

function openFindInPage() {
  showFindInPage.value = true;
}

function closeFindInPage() {
  showFindInPage.value = false;
}

function handleKeydown(e: KeyboardEvent) {
  // Open find in page with Ctrl+F or Cmd+F
  if ((e.ctrlKey || e.metaKey) && e.key === 'f') {
    // Only if we're showing an article in content mode (not webpage view)
    if (article.value && showContent.value) {
      e.preventDefault();
      openFindInPage();
    }
  }
  // Close with Escape
  if (e.key === 'Escape' && showFindInPage.value) {
    closeFindInPage();
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown);
});

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleKeydown);
});
</script>

<template>
  <main
    :class="[
      'flex-1 bg-bg-primary flex flex-col h-full absolute w-full md:static md:w-auto z-30 transition-transform duration-300',
      article ? 'translate-x-0' : 'translate-x-full md:translate-x-0',
    ]"
  >
    <div
      v-if="!article"
      class="hidden md:flex flex-col items-center justify-center h-full text-text-secondary text-center px-4"
    >
      <PhNewspaper :size="48" class="mb-4 sm:mb-5 opacity-50 sm:w-16 sm:h-16" />
      <p class="text-sm sm:text-base">{{ t('selectArticle') }}</p>
    </div>

    <div v-else class="flex flex-col h-full bg-bg-primary">
      <ArticleToolbar
        :article="article"
        :show-content="showContent"
        :show-translations="showTranslations"
        @close="close"
        @toggle-content-view="toggleContentView"
        @toggle-read="toggleRead"
        @toggle-favorite="toggleFavorite"
        @toggle-read-later="toggleReadLater"
        @open-original="openOriginal"
        @toggle-translations="toggleTranslations"
        @export-to-obsidian="exportToObsidian"
      />

      <!-- Original webpage view -->
      <div v-if="!showContent" class="flex-1 bg-bg-primary w-full">
        <iframe
          :key="article.id"
          :src="`/api/webpage/proxy?url=${encodeURIComponent(article.url)}`"
          class="w-full h-full border-none"
          sandbox="allow-scripts allow-same-origin allow-popups"
        ></iframe>
      </div>

      <!-- RSS content view -->
      <ArticleContent
        v-else
        :article="article"
        :article-content="articleContent"
        :is-loading-content="isLoadingContent"
        :attach-image-event-listeners="attachImageEventListeners"
        :show-translations="showTranslations"
        :show-content="showContent"
        @retry-load-content="handleRetryLoadContent"
      />
    </div>

    <!-- Find in Page (only shown in content mode) -->
    <FindInPage
      v-if="showFindInPage && showContent"
      container-selector=".prose-content"
      @close="closeFindInPage"
    />

    <!-- Image Viewer Modal -->
    <ImageViewer
      v-if="imageViewerSrc"
      :src="imageViewerSrc"
      :alt="imageViewerAlt"
      @close="closeImageViewer"
    />
  </main>
</template>
