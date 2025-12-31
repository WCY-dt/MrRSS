<script setup lang="ts">
import { useAppStore } from '@/stores/app';
import { useI18n } from 'vue-i18n';
import { ref, computed, type Ref } from 'vue';
import {
  PhRss,
  PhPlus,
  PhTrash,
  PhFolder,
  PhPencil,
  PhSortAscending,
  PhCode,
  PhEyeSlash,
  PhFileCode,
  PhEnvelope,
} from '@phosphor-icons/vue';
import type { Feed } from '@/types/models';

const store = useAppStore();
const { t } = useI18n();

const emit = defineEmits<{
  'add-feed': [];
  'edit-feed': [feed: Feed];
  'delete-feed': [id: number];
  'batch-delete': [ids: number[]];
  'batch-move': [ids: number[]];
}>();

const selectedFeeds: Ref<number[]> = ref([]);

// Sorting state
type SortField = 'name' | 'date' | 'category';
type SortDirection = 'asc' | 'desc';
const sortField = ref<SortField>('name');
const sortDirection = ref<SortDirection>('asc');

const sortedFeeds = computed(() => {
  if (!store.feeds) return [];
  const feeds = [...store.feeds];

  feeds.sort((a, b) => {
    let comparison = 0;

    if (sortField.value === 'name') {
      comparison = a.title.localeCompare(b.title, undefined, { sensitivity: 'base' });
    } else if (sortField.value === 'date') {
      // Use feed ID as proxy for add time (higher ID = newer)
      comparison = a.id - b.id;
    } else if (sortField.value === 'category') {
      const catA = a.category || '';
      const catB = b.category || '';
      comparison = catA.localeCompare(catB, undefined, { sensitivity: 'base' });
    }

    return sortDirection.value === 'asc' ? comparison : -comparison;
  });

  return feeds;
});

const isAllSelected = computed(() => {
  if (!store.feeds || store.feeds.length === 0) return false;
  // Get non-FreshRSS feeds
  const nonFreshRSSFeeds = store.feeds.filter((f) => !f.is_freshrss_source);
  if (nonFreshRSSFeeds.length === 0) return false;
  // Check if all non-FreshRSS feeds are selected
  return nonFreshRSSFeeds.every((f) => selectedFeeds.value.includes(f.id));
});

function toggleSort(field: SortField) {
  if (sortField.value === field) {
    sortDirection.value = sortDirection.value === 'asc' ? 'desc' : 'asc';
  } else {
    sortField.value = field;
    sortDirection.value = 'asc';
  }
}

function toggleSelectAll(e: Event) {
  const target = e.target as HTMLInputElement;
  if (!store.feeds) return;
  if (target.checked) {
    // Select only non-FreshRSS feeds
    selectedFeeds.value = store.feeds.filter((f) => !f.is_freshrss_source).map((f) => f.id);
  } else {
    selectedFeeds.value = [];
  }
}

function handleAddFeed() {
  emit('add-feed');
}

function handleEditFeed(feed: Feed) {
  emit('edit-feed', feed);
}

function handleDeleteFeed(id: number) {
  emit('delete-feed', id);
}

function handleBatchDelete() {
  if (selectedFeeds.value.length === 0) return;
  emit('batch-delete', selectedFeeds.value);
  selectedFeeds.value = [];
}

function handleBatchMove() {
  if (selectedFeeds.value.length === 0) return;
  emit('batch-move', selectedFeeds.value);
  selectedFeeds.value = [];
}

function getFavicon(url: string): string {
  try {
    return `https://www.google.com/s2/favicons?domain=${new URL(url).hostname}`;
  } catch {
    return '';
  }
}

function isScriptFeed(feed: Feed): boolean {
  return !!feed.script_path;
}

function isXPathFeed(feed: Feed): boolean {
  return feed.type === 'HTML+XPath' || feed.type === 'XML+XPath';
}

function isEmailFeed(feed: Feed): boolean {
  return feed.type === 'email';
}

function isFreshRSSFeed(feed: Feed): boolean {
  return !!feed.is_freshrss_source;
}

async function openScriptsFolder() {
  try {
    await fetch('/api/scripts/open', { method: 'POST' });
    window.showToast(t('scriptsFolderOpened'), 'success');
  } catch (e) {
    console.error('Failed to open scripts folder:', e);
  }
}
</script>

<template>
  <div class="setting-group">
    <label
      class="font-semibold mb-2 sm:mb-3 text-text-secondary uppercase text-xs tracking-wider flex items-center gap-2"
    >
      <PhRss :size="14" class="sm:w-4 sm:h-4" />
      {{ t('manageFeeds') }}
    </label>

    <div class="flex flex-wrap gap-1.5 sm:gap-2 mb-2 text-xs sm:text-sm">
      <button class="btn-secondary py-1.5 px-2.5 sm:px-3" @click="handleAddFeed">
        <PhPlus :size="14" class="sm:w-4 sm:h-4" />
        <span class="hidden sm:inline">{{ t('addFeed') }}</span
        ><span class="sm:hidden">{{ t('addFeed').split(' ')[0] }}</span>
      </button>
      <button
        class="btn-danger py-1.5 px-2.5 sm:px-3"
        :disabled="selectedFeeds.length === 0"
        @click="handleBatchDelete"
      >
        <PhTrash :size="14" class="sm:w-4 sm:h-4" />
        <span class="hidden sm:inline">{{ t('deleteSelected') }}</span
        ><span class="sm:hidden">{{ t('delete') }}</span>
      </button>
      <button
        class="btn-secondary py-1.5 px-2.5 sm:px-3"
        :disabled="selectedFeeds.length === 0"
        @click="handleBatchMove"
      >
        <PhFolder :size="14" class="sm:w-4 sm:h-4" />
        <span class="hidden sm:inline">{{ t('moveSelected') }}</span
        ><span class="sm:hidden">{{ t('move') }}</span>
      </button>
    </div>

    <div class="border border-border rounded-lg bg-bg-secondary">
      <!-- Table Header -->
      <div
        class="flex items-center justify-between p-1.5 sm:p-2 border-b border-border bg-bg-tertiary"
      >
        <label class="flex items-center gap-1.5 sm:gap-2 cursor-pointer select-none">
          <input
            type="checkbox"
            :checked="isAllSelected"
            class="w-3.5 h-3.5 sm:w-4 sm:h-4 rounded border-border text-accent focus:ring-2 focus:ring-accent cursor-pointer"
            @change="toggleSelectAll"
          />
          <span class="hidden sm:inline text-xs sm:text-sm">{{ t('selectAll') }}</span>
        </label>
        <div class="flex items-center gap-1">
          <PhSortAscending :size="12" class="text-text-secondary" />
          <button
            :class="[
              'px-1.5 py-0.5 text-xs rounded transition-colors',
              sortField === 'name'
                ? 'bg-accent text-white'
                : 'bg-bg-secondary text-text-primary hover:bg-bg-primary',
            ]"
            @click="toggleSort('name')"
          >
            {{ t('sortByName') }}
            <span v-if="sortField === 'name'">{{ sortDirection === 'asc' ? '↑' : '↓' }}</span>
          </button>
          <button
            :class="[
              'px-1.5 py-0.5 text-xs rounded transition-colors',
              sortField === 'date'
                ? 'bg-accent text-white'
                : 'bg-bg-secondary text-text-primary hover:bg-bg-primary',
            ]"
            @click="toggleSort('date')"
          >
            {{ t('sortByDate') }}
            <span v-if="sortField === 'date'">{{ sortDirection === 'asc' ? '↑' : '↓' }}</span>
          </button>
          <button
            :class="[
              'px-1.5 py-0.5 text-xs rounded transition-colors',
              sortField === 'category'
                ? 'bg-accent text-white'
                : 'bg-bg-secondary text-text-primary hover:bg-bg-primary',
            ]"
            @click="toggleSort('category')"
          >
            {{ t('sortByCategory') }}
            <span v-if="sortField === 'category'">{{ sortDirection === 'asc' ? '↑' : '↓' }}</span>
          </button>
        </div>
      </div>

      <!-- Scrollable Content -->
      <div class="overflow-y-auto max-h-48 sm:max-h-80">
        <!-- Feed Rows -->
        <div
          v-for="feed in sortedFeeds"
          :key="feed.id"
          :class="[
            'flex items-center p-1.5 sm:p-2 border-b border-border last:border-0 gap-1.5 sm:gap-2',
            feed.is_freshrss_source ? 'bg-info/10' : 'bg-bg-primary hover:bg-bg-secondary',
          ]"
        >
          <input
            v-model="selectedFeeds"
            type="checkbox"
            :value="feed.id"
            :disabled="feed.is_freshrss_source"
            class="w-3.5 h-3.5 sm:w-4 sm:h-4 shrink-0 rounded border-border text-accent focus:ring-2 focus:ring-accent cursor-pointer"
            :class="{ 'cursor-not-allowed opacity-50': feed.is_freshrss_source }"
          />
          <div class="w-4 h-4 flex items-center justify-center shrink-0">
            <img
              :src="getFavicon(feed.url)"
              class="w-full h-full object-contain"
              @error="
                ($event: Event) => {
                  const target = $event.target as HTMLImageElement;
                  if (target) target.style.display = 'none';
                }
              "
            />
          </div>
          <div class="truncate flex-1 min-w-0">
            <div class="font-medium truncate text-xs sm:text-sm flex items-center gap-1">
              {{ feed.title }}
              <!-- FreshRSS icon indicator -->
              <img
                v-if="feed.is_freshrss_source"
                src="/assets/plugin_icons/freshrss.svg"
                class="w-3.5 h-3.5 shrink-0 inline"
                :title="t('freshRSSSyncedFeed')"
                alt="FreshRSS"
              />
              <PhEyeSlash
                v-if="feed.hide_from_timeline"
                :size="12"
                class="text-text-secondary shrink-0"
                :title="t('hideFromTimeline')"
              />
            </div>
            <div class="text-xs text-text-secondary truncate hidden sm:block">
              <span v-if="feed.category" class="inline-flex items-center gap-1">
                <PhFolder :size="10" class="inline" />
                {{ feed.category }}
                <span class="mx-1">•</span>
              </span>
              <span
                v-if="isFreshRSSFeed(feed)"
                class="inline-flex items-center gap-1 text-info"
                :title="t('freshRSSSyncedFeed')"
              >
                {{ feed.url }}
              </span>
              <span
                v-else-if="isScriptFeed(feed)"
                class="inline-flex items-center gap-1"
                :title="t('customScript')"
              >
                <PhCode :size="10" class="inline text-accent" />
                <button
                  class="text-accent hover:underline"
                  :title="t('openScriptsFolder')"
                  @click.stop="openScriptsFolder"
                >
                  {{ feed.script_path }}
                </button>
              </span>
              <span
                v-else-if="isXPathFeed(feed)"
                class="inline-flex items-center gap-1"
                :title="feed.type"
              >
                <PhFileCode :size="10" class="inline text-accent" />
                <span class="text-accent">{{ feed.type }}</span>
                <span class="mx-1">•</span>
                {{ feed.url }}
              </span>
              <span
                v-else-if="isEmailFeed(feed)"
                class="inline-flex items-center gap-1"
                :title="t('emailNewsletter')"
              >
                <PhEnvelope :size="10" class="inline text-accent" />
                <span class="text-accent">{{ t('emailNewsletter') }}</span>
                <span v-if="feed.email_address" class="mx-1">•</span>
                <span v-if="feed.email_address">{{ feed.email_address }}</span>
              </span>
              <span v-else>{{ feed.url }}</span>
            </div>
          </div>
          <div class="flex gap-0.5 sm:gap-1 shrink-0">
            <button
              class="text-accent hover:bg-bg-tertiary p-1 rounded text-sm"
              :title="feed.is_freshrss_source ? t('freshRSSFeedLocked') : t('edit')"
              :disabled="feed.is_freshrss_source"
              :class="{ 'cursor-not-allowed opacity-50': feed.is_freshrss_source }"
              @click="!feed.is_freshrss_source && handleEditFeed(feed)"
            >
              <PhPencil :size="14" class="sm:w-4 sm:h-4" />
            </button>
            <button
              class="text-red-500 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 p-1 rounded text-sm"
              :title="feed.is_freshrss_source ? t('freshRSSFeedLocked') : t('delete')"
              :disabled="feed.is_freshrss_source"
              :class="{ 'cursor-not-allowed opacity-50': feed.is_freshrss_source }"
              @click="!feed.is_freshrss_source && handleDeleteFeed(feed.id)"
            >
              <PhTrash :size="14" class="sm:w-4 sm:h-4" />
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "../../../../style.css";

.btn-primary {
  @apply bg-accent text-white px-3 sm:px-4 py-1.5 sm:py-2 rounded-md cursor-pointer flex items-center gap-1.5 sm:gap-2 font-semibold hover:bg-accent-hover transition-colors shadow-sm;
}
.btn-primary:disabled {
  @apply opacity-50 cursor-not-allowed;
}
.btn-secondary {
  @apply bg-transparent border border-border text-text-primary px-3 sm:px-4 py-1.5 sm:py-2 rounded-md cursor-pointer flex items-center gap-1.5 sm:gap-2 font-medium hover:bg-bg-tertiary transition-colors;
}
.btn-secondary:disabled {
  @apply opacity-50 cursor-not-allowed;
}
.btn-danger {
  @apply bg-transparent border border-red-300 text-red-600 px-3 sm:px-4 py-1.5 sm:py-2 rounded-md cursor-pointer flex items-center gap-1.5 sm:gap-2 font-semibold hover:bg-red-50 dark:hover:bg-red-900/20 dark:border-red-400 dark:text-red-400 transition-colors;
}
.btn-danger:disabled {
  @apply opacity-50 cursor-not-allowed;
}
</style>
