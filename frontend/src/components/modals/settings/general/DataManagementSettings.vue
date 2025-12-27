<script setup lang="ts">
import { ref, onMounted, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  PhDatabase,
  PhBroom,
  PhHardDrive,
  PhCalendarX,
  PhImage,
  PhTrash,
} from '@phosphor-icons/vue';
import type { SettingsData } from '@/types/settings';

const { t } = useI18n();

interface Props {
  settings: SettingsData;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:settings': [settings: SettingsData];
}>();

const mediaCacheSize = ref<number>(0);
const articleCacheCount = ref<number>(0);
const isCleaningCache = ref(false);
const isCleaningArticleCache = ref(false);

// Fetch current media cache size
async function fetchMediaCacheSize() {
  try {
    const response = await fetch('/api/media/info');
    if (response.ok) {
      const data = await response.json();
      mediaCacheSize.value = data.cache_size_mb || 0;
    }
  } catch (error) {
    console.error('Failed to fetch media cache size:', error);
  }
}

// Fetch article content cache count
async function fetchArticleCacheCount() {
  try {
    const response = await fetch('/api/articles/content-cache-info');
    if (response.ok) {
      const data = await response.json();
      articleCacheCount.value = data.cached_articles || 0;
    }
  } catch (error) {
    console.error('Failed to fetch article cache count:', error);
  }
}

// Clean media cache
async function cleanMediaCache() {
  isCleaningCache.value = true;
  try {
    const response = await fetch('/api/media/cleanup?all=true', { method: 'POST' });
    if (response.ok) {
      const data = await response.json();
      window.showToast(
        t('mediaCacheCleanup') + ': ' + data.files_cleaned + ' files removed',
        'success'
      );
      // Immediately update cache size
      await fetchMediaCacheSize();
    } else {
      window.showToast(t('errorCleaningDatabase'), 'error');
    }
  } catch (error) {
    console.error('Failed to clean media cache:', error);
    window.showToast(t('errorCleaningDatabase'), 'error');
  } finally {
    isCleaningCache.value = false;
  }
}

// Clean article content cache
async function cleanArticleContentCache() {
  isCleaningArticleCache.value = true;
  try {
    const response = await fetch('/api/articles/cleanup-content', { method: 'POST' });
    if (response.ok) {
      const data = await response.json();
      window.showToast(
        t('articleContentCacheCleanup') + ': ' + data.entries_cleaned + ' entries removed',
        'success'
      );
      // Immediately update cache count
      await fetchArticleCacheCount();
    } else {
      window.showToast(t('errorCleaningDatabase'), 'error');
    }
  } catch (error) {
    console.error('Failed to clean article content cache:', error);
    window.showToast(t('errorCleaningDatabase'), 'error');
  } finally {
    isCleaningArticleCache.value = false;
  }
}

// Fetch all cache data
async function fetchAllCacheData() {
  if (props.settings.media_cache_enabled) {
    await fetchMediaCacheSize();
  }
  await fetchArticleCacheCount();
}

onMounted(() => {
  // Fetch cache sizes when component mounts
  fetchAllCacheData();
});

// Watch for settings changes to refetch media cache info
watch(
  () => props.settings.media_cache_enabled,
  () => {
    fetchAllCacheData();
  }
);
</script>

<template>
  <div class="setting-group">
    <label
      class="font-semibold mb-2 sm:mb-3 text-text-secondary uppercase text-xs tracking-wider flex items-center gap-2"
    >
      <PhDatabase :size="14" class="sm:w-4 sm:h-4" />
      {{ t('dataManagement') }}
    </label>

    <!-- Article Cleanup -->
    <div class="setting-item">
      <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
        <PhBroom :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">{{ t('autoCleanup') }}</div>
          <div class="text-xs text-text-secondary hidden sm:block">
            {{ t('autoCleanupDesc') }}
          </div>
        </div>
      </div>
      <input
        :checked="props.settings.auto_cleanup_enabled"
        type="checkbox"
        class="toggle"
        @change="
          (e) =>
            emit('update:settings', {
              ...props.settings,
              auto_cleanup_enabled: (e.target as HTMLInputElement).checked,
            })
        "
      />
    </div>

    <div
      v-if="props.settings.auto_cleanup_enabled"
      class="ml-2 sm:ml-4 mt-2 sm:mt-3 space-y-2 sm:space-y-3 border-l-2 border-border pl-2 sm:pl-4"
    >
      <div class="sub-setting-item">
        <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
          <PhHardDrive :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
          <div class="flex-1 min-w-0">
            <div class="font-medium mb-0 sm:mb-1 text-sm">{{ t('maxCacheSize') }}</div>
            <div class="text-xs text-text-secondary hidden sm:block">
              {{ t('maxCacheSizeDesc') }}
            </div>
          </div>
        </div>
        <div class="flex items-center gap-1 sm:gap-2 shrink-0">
          <input
            :value="props.settings.max_cache_size_mb"
            type="number"
            min="1"
            max="1000"
            class="input-field w-14 sm:w-20 text-center text-xs sm:text-sm"
            @input="
              (e) =>
                emit('update:settings', {
                  ...props.settings,
                  max_cache_size_mb: parseInt((e.target as HTMLInputElement).value) || 100,
                })
            "
          />
          <span class="text-xs sm:text-sm text-text-secondary">MB</span>
        </div>
      </div>

      <div class="sub-setting-item">
        <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
          <PhCalendarX :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
          <div class="flex-1 min-w-0">
            <div class="font-medium mb-0 sm:mb-1 text-sm">{{ t('maxArticleAge') }}</div>
            <div class="text-xs text-text-secondary hidden sm:block">
              {{ t('maxArticleAgeDesc') }}
            </div>
          </div>
        </div>
        <div class="flex items-center gap-1 sm:gap-2 shrink-0">
          <input
            :value="props.settings.max_article_age_days"
            type="number"
            min="1"
            max="365"
            class="input-field w-14 sm:w-20 text-center text-xs sm:text-sm"
            @input="
              (e) =>
                emit('update:settings', {
                  ...props.settings,
                  max_article_age_days: parseInt((e.target as HTMLInputElement).value) || 30,
                })
            "
          />
          <span class="text-xs sm:text-sm text-text-secondary">{{ t('days') }}</span>
        </div>
      </div>

      <!-- Article Content Cache Cleanup -->
      <div class="sub-setting-item">
        <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
          <PhTrash :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
          <div class="flex-1 min-w-0">
            <div class="font-medium mb-0 sm:mb-1 text-sm">
              {{ t('articleContentCacheCleanup') }}
            </div>
            <div class="text-xs text-text-secondary hidden sm:block">
              {{ t('articleContentCacheCleanupDesc') }}
            </div>
            <div class="text-xs text-text-secondary mt-1">
              {{ t('currentCachedArticles') }}:
              <span class="theme-number">{{ articleCacheCount }}</span>
            </div>
          </div>
        </div>
        <button
          :disabled="isCleaningArticleCache"
          class="btn-secondary"
          @click="cleanArticleContentCache"
        >
          <PhBroom :size="16" class="sm:w-5 sm:h-5" />
          {{ isCleaningArticleCache ? t('cleaning') : t('cleanupArticleContentCache') }}
        </button>
      </div>
    </div>

    <!-- Media Cache -->
    <div class="setting-item mt-2 sm:mt-3">
      <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
        <PhImage :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
            {{ t('mediaCacheEnabled') }}
          </div>
          <div class="text-xs text-text-secondary hidden sm:block">
            {{ t('mediaCacheEnabledDesc') }}
          </div>
        </div>
      </div>
      <input
        :checked="props.settings.media_cache_enabled"
        type="checkbox"
        class="toggle"
        @change="
          (e) =>
            emit('update:settings', {
              ...props.settings,
              media_cache_enabled: (e.target as HTMLInputElement).checked,
            })
        "
      />
    </div>

    <div
      v-if="props.settings.media_cache_enabled"
      class="ml-2 sm:ml-4 mt-2 sm:mt-3 space-y-2 sm:space-y-3 border-l-2 border-border pl-2 sm:pl-4"
    >
      <div class="sub-setting-item">
        <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
          <PhHardDrive :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
          <div class="flex-1 min-w-0">
            <div class="font-medium mb-0 sm:mb-1 text-sm">{{ t('mediaCacheMaxSize') }}</div>
            <div class="text-xs text-text-secondary hidden sm:block">
              {{ t('mediaCacheMaxSizeDesc') }}
            </div>
          </div>
        </div>
        <div class="flex items-center gap-1 sm:gap-2 shrink-0">
          <input
            :value="props.settings.media_cache_max_size_mb"
            type="number"
            min="10"
            max="1000"
            class="input-field w-14 sm:w-20 text-center text-xs sm:text-sm"
            @input="
              (e) =>
                emit('update:settings', {
                  ...props.settings,
                  media_cache_max_size_mb: parseInt((e.target as HTMLInputElement).value) || 100,
                })
            "
          />
          <span class="text-xs sm:text-sm text-text-secondary">MB</span>
        </div>
      </div>

      <div class="sub-setting-item">
        <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
          <PhCalendarX :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
          <div class="flex-1 min-w-0">
            <div class="font-medium mb-0 sm:mb-1 text-sm">{{ t('mediaCacheMaxAge') }}</div>
            <div class="text-xs text-text-secondary hidden sm:block">
              {{ t('mediaCacheMaxAgeDesc') }}
            </div>
          </div>
        </div>
        <div class="flex items-center gap-1 sm:gap-2 shrink-0">
          <input
            :value="props.settings.media_cache_max_age_days"
            type="number"
            min="1"
            max="90"
            class="input-field w-14 sm:w-20 text-center text-xs sm:text-sm"
            @input="
              (e) =>
                emit('update:settings', {
                  ...props.settings,
                  media_cache_max_age_days: parseInt((e.target as HTMLInputElement).value) || 7,
                })
            "
          />
          <span class="text-xs sm:text-sm text-text-secondary">{{ t('days') }}</span>
        </div>
      </div>

      <div class="sub-setting-item">
        <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
          <PhTrash :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
          <div class="flex-1 min-w-0">
            <div class="font-medium mb-0 sm:mb-1 text-sm">{{ t('mediaCacheCleanup') }}</div>
            <div class="text-xs text-text-secondary hidden sm:block">
              {{ t('mediaCacheCleanupDesc') }}
            </div>
            <div class="text-xs text-text-secondary mt-1">
              {{ t('currentCacheSize') }}:
              <span class="theme-number">{{ mediaCacheSize.toFixed(2) }} MB</span>
            </div>
          </div>
        </div>
        <button :disabled="isCleaningCache" class="btn-secondary" @click="cleanMediaCache">
          <PhBroom :size="16" class="sm:w-5 sm:h-5" />
          {{ isCleaningCache ? t('cleaning') : t('cleanupMediaCache') }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "../../../../style.css";

.input-field {
  @apply p-1.5 sm:p-2.5 border border-border rounded-md bg-bg-secondary text-text-primary focus:border-accent focus:outline-none transition-colors;
}
.toggle {
  @apply w-10 h-5 appearance-none bg-bg-tertiary rounded-full relative cursor-pointer border border-border transition-colors checked:bg-accent checked:border-accent shrink-0;
}
.toggle::after {
  content: '';
  @apply absolute top-0.5 left-0.5 w-3.5 h-3.5 bg-white rounded-full shadow-sm transition-transform;
}
.toggle:checked::after {
  transform: translateX(20px);
}
.setting-item {
  @apply flex items-center sm:items-start justify-between gap-2 sm:gap-4 p-2 sm:p-3 rounded-lg bg-bg-secondary border border-border;
}
.sub-setting-item {
  @apply flex items-center sm:items-start justify-between gap-2 sm:gap-4 p-2 sm:p-2.5 rounded-md bg-bg-tertiary;
}
.btn-secondary {
  @apply bg-bg-tertiary border border-border text-text-primary px-3 sm:px-4 py-1.5 sm:py-2 rounded-md cursor-pointer flex items-center gap-1.5 sm:gap-2 font-medium hover:bg-bg-secondary transition-colors;
}
.setting-group {
  @apply space-y-2 sm:space-y-3;
}
.theme-number {
  @apply text-accent font-semibold;
}
</style>
