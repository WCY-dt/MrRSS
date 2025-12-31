<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhEyeSlash, PhStar, PhClockCountdown } from '@phosphor-icons/vue';
import type { Article } from '@/types/models';
import { formatDate as formatDateUtil } from '@/utils/date';
import { getProxiedMediaUrl, isMediaCacheEnabled } from '@/utils/mediaProxy';
import { useShowPreviewImages } from '@/composables/ui/useShowPreviewImages';
import { useAppStore } from '@/stores/app';

interface Props {
  article: Article;
  isActive: boolean;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  click: [];
  contextmenu: [event: MouseEvent];
  observeElement: [element: Element | null];
  hoverMarkAsRead: [articleId: number];
}>();

const { t, locale } = useI18n();
const { showPreviewImages } = useShowPreviewImages();
const store = useAppStore();

const mediaCacheEnabled = ref(false);
const hoverMarkAsRead = ref(false);
let hoverTimeout: ReturnType<typeof setTimeout> | null = null;

const imageUrl = computed(() => {
  if (!props.article.image_url) return '';
  if (mediaCacheEnabled.value) {
    return getProxiedMediaUrl(props.article.image_url, props.article.url);
  }
  return props.article.image_url;
});

const shouldShowImage = computed(() => {
  return showPreviewImages.value && props.article.image_url;
});

function formatDate(dateStr: string): string {
  return formatDateUtil(dateStr, locale.value === 'zh-CN' ? 'zh-CN' : 'en-US');
}

function handleImageError(event: Event) {
  const target = event.target as HTMLImageElement;
  target.style.display = 'none';
}

// Hover mark as read functionality
function handleMouseEnter() {
  // Don't mark as read if:
  // - Setting is disabled
  // - Article is already read
  // - Article is in "Read Later" list (user explicitly wants to read it later)
  if (!hoverMarkAsRead.value || props.article.is_read || props.article.is_read_later) {
    return;
  }

  // Use a small delay to avoid marking as read when quickly scrolling through the list
  hoverTimeout = setTimeout(() => {
    markAsRead();
  }, 300);
}

function handleMouseLeave() {
  if (hoverTimeout) {
    clearTimeout(hoverTimeout);
    hoverTimeout = null;
  }
}

async function markAsRead() {
  if (props.article.is_read) return;

  try {
    await fetch(`/api/articles/read?id=${props.article.id}&read=true`, {
      method: 'POST',
    });
    // Emit event to parent to update article state
    emit('hoverMarkAsRead', props.article.id);
    store.fetchUnreadCounts();
  } catch (e) {
    console.error('Error marking as read on hover:', e);
  }
}

async function loadSettings() {
  try {
    const res = await fetch('/api/settings');
    const data = await res.json();
    hoverMarkAsRead.value = data.hover_mark_as_read === 'true';
  } catch (e) {
    console.error('Error loading hover mark as read setting:', e);
  }
}

onMounted(async () => {
  mediaCacheEnabled.value = await isMediaCacheEnabled();
  await loadSettings();
});

onUnmounted(() => {
  if (hoverTimeout) {
    clearTimeout(hoverTimeout);
  }
});
</script>

<template>
  <div
    :ref="(el) => emit('observeElement', el as Element | null)"
    :data-article-id="article.id"
    :class="[
      'article-card',
      article.is_read ? 'read' : '',
      article.is_favorite ? 'favorite' : '',
      article.is_hidden ? 'hidden' : '',
      article.is_read_later ? 'read-later' : '',
      isActive ? 'active' : '',
    ]"
    @click="emit('click')"
    @contextmenu="emit('contextmenu', $event)"
    @mouseenter="handleMouseEnter"
    @mouseleave="handleMouseLeave"
  >
    <img
      v-if="shouldShowImage"
      :src="imageUrl"
      class="w-16 h-12 sm:w-20 sm:h-[60px] object-cover rounded bg-bg-tertiary shrink-0 border border-border"
      @error="handleImageError"
    />

    <div class="flex-1 min-w-0">
      <div class="flex items-start gap-1.5 sm:gap-2">
        <h4
          v-if="!article.translated_title || article.translated_title === article.title"
          class="flex-1 m-0 mb-1 sm:mb-1.5 text-sm sm:text-base font-semibold leading-snug text-text-primary article-title"
        >
          {{ article.title }}
        </h4>
        <div v-else class="flex-1">
          <h4
            class="m-0 mb-0.5 sm:mb-1 text-sm sm:text-base font-semibold leading-snug text-text-primary article-title"
          >
            {{ article.translated_title }}
          </h4>
          <div
            class="text-[10px] sm:text-xs text-text-secondary italic mb-0.5 sm:mb-1 article-title"
          >
            {{ article.title }}
          </div>
        </div>
        <PhEyeSlash
          v-if="article.is_hidden"
          :size="18"
          class="text-text-secondary flex-shrink-0 sm:w-5 sm:h-5"
          :title="t('hideArticle')"
        />
      </div>

      <div
        class="flex justify-between items-center text-[10px] sm:text-xs text-text-secondary mt-1.5 sm:mt-2"
      >
        <span class="font-medium text-accent truncate flex-1 min-w-0 mr-2">
          {{ article.feed_title }}
        </span>
        <div class="flex items-center gap-1 sm:gap-2 shrink-0">
          <PhClockCountdown
            v-if="article.is_read_later"
            :size="14"
            class="text-blue-500 sm:w-[18px] sm:h-[18px]"
            weight="fill"
          />
          <PhStar
            v-if="article.is_favorite"
            :size="14"
            class="text-yellow-500 sm:w-[18px] sm:h-[18px]"
            weight="fill"
          />
          <!-- FreshRSS indicator -->
          <img
            v-if="article.freshrss_item_id"
            src="/assets/plugin_icons/freshrss.svg"
            class="w-3.5 h-3.5 shrink-0 sm:w-4 sm:h-4"
            :title="t('freshRSSSyncedFeed')"
            alt="FreshRSS"
          />
          <span class="whitespace-nowrap">{{ formatDate(article.published_at) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "../../style.css";

.article-card {
  @apply p-2 sm:p-3 border-b border-border cursor-pointer transition-colors flex gap-2 sm:gap-3 relative;
}

.article-card:hover {
  @apply bg-bg-tertiary;
}

.article-card.active {
  @apply bg-bg-tertiary border-l-2 sm:border-l-[3px] border-l-accent;
}

.article-card.read h4 {
  @apply text-text-secondary font-normal;
}

.article-card.read .text-sm {
  @apply text-text-secondary opacity-80;
}

.article-card.favorite {
  background-color: rgba(255, 215, 0, 0.05);
}

.article-card.read-later {
  background-color: rgba(59, 130, 246, 0.05);
}

.article-card.hidden {
  @apply opacity-60 bg-gray-100 dark:bg-gray-800;
}

.article-card.hidden:hover {
  @apply opacity-80;
}

.article-title {
  word-break: break-word;
  overflow-wrap: anywhere;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  display: -webkit-box;
  overflow: hidden;
}
</style>
