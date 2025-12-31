<script setup lang="ts">
import { ref, onMounted, watch } from 'vue';
import { useAppStore } from '@/stores/app';
import { useI18n } from 'vue-i18n';
import { PhPlus, PhGear, PhMagnifyingGlass, PhX, PhPencil, PhCheck } from '@phosphor-icons/vue';
import { useSidebar } from '@/composables/core/useSidebar';
import { useDragDrop } from '@/composables/ui/useDragDrop';
import SidebarNavItem from './SidebarNavItem.vue';
import SidebarCategory from './SidebarCategory.vue';

const store = useAppStore();
const { t } = useI18n();

// Edit mode for drag reordering
const isEditMode = ref(false);

// Local state to track if user is actively dragging (independent of drop operation)
const isDragging = ref(false);
// Track if drop was handled to clear isDragging correctly
let dropHandled = false;

function toggleEditMode() {
  isEditMode.value = !isEditMode.value;
}

// Check if image gallery feature is enabled
const imageGalleryEnabled = ref(false);

async function loadImageGallerySetting() {
  try {
    const res = await fetch('/api/settings');
    if (res.ok) {
      const data = await res.json();
      imageGalleryEnabled.value = data.image_gallery_enabled === 'true';
    }
  } catch (e) {
    console.error('Failed to load settings:', e);
  }
}

onMounted(async () => {
  await loadImageGallerySetting();

  // Listen for settings changes
  window.addEventListener('image-gallery-setting-changed', (e: Event) => {
    const customEvent = e as CustomEvent;
    imageGalleryEnabled.value = customEvent.detail.enabled;
  });
});

interface Props {
  isOpen?: boolean;
}

defineProps<Props>();

const emit = defineEmits<{
  toggle: [];
}>();

const {
  tree,
  categoryUnreadCounts,
  toggleCategory,
  isCategoryOpen: checkIsCategoryOpen,
  searchQuery,
  onFeedContextMenu,
  onCategoryContextMenu,
} = useSidebar();

// Drag and drop functionality
const {
  draggingFeedId,
  dragOverCategory,
  dropPreview,
  onDragStart,
  onDragEnd,
  onDragOver: onDragOverComposable,
  onDragLeave: onDragLeaveComposable,
  onDrop,
} = useDragDrop();

// Handle drag events from categories
function handleDragStart(feedId: number, event: Event) {
  console.log('[handleDragStart] Starting drag for feed:', feedId);

  // Prevent dragging FreshRSS feeds
  const feed = store.feeds?.find((f) => f.id === feedId);
  if (feed?.is_freshrss_source) {
    console.log('[handleDragStart] Blocked drag for FreshRSS feed:', feedId);
    event.preventDefault();
    window.showToast(t('freshRSSFeedLocked'), 'info');
    return;
  }

  isDragging.value = true;
  dropHandled = false;
  onDragStart(feedId, event);
}

function handleDragEnd() {
  // Don't clear isDragging here - let handleDrop or a timeout handle it
  // This prevents premature clearing if drop is still being processed
  onDragEnd();
}

function handleDragOver(categoryName: string, feedId: number | null, event: Event) {
  console.log('[Sidebar] handleDragOver called with:', { categoryName, feedId, event });
  onDragOverComposable(categoryName, feedId, event);
}

function handleDragLeave(categoryName: string, event: Event) {
  onDragLeaveComposable(categoryName, event);
}

async function handleDrop(categoryName: string, feeds: any[]) {
  if (dropHandled) {
    console.log('[handleDrop] Drop already handled, skipping');
    return;
  }

  dropHandled = true;

  console.log('[handleDrop] Starting drop operation:', {
    categoryName,
    feedsCount: feeds.length,
    feeds: feeds.map((f) => ({ id: f.id, title: f.title, category: f.category })),
    draggingFeedId: draggingFeedId.value,
  });

  // Prevent dropping into FreshRSS categories
  // Check if any feed in the target category is a FreshRSS feed
  const targetCategoryFeeds = feeds.filter(
    (f) => f.category === categoryName || (categoryName === 'uncategorized' && !f.category)
  );
  const hasFreshRSSFeedInTarget = targetCategoryFeeds.some((f) => f.is_freshrss_source);

  if (hasFreshRSSFeedInTarget) {
    console.log('[handleDrop] Blocked drop into FreshRSS category:', categoryName);
    window.showToast(t('freshRSSFeedLocked'), 'info');
    isDragging.value = false;
    return;
  }

  // Also check if the dragged feed is being dropped into a category that contains FreshRSS feeds
  const draggedFeed = store.feeds?.find((f) => f.id === draggingFeedId.value);
  if (draggedFeed && !draggedFeed.is_freshrss_source) {
    // Check all feeds to see if target category has any FreshRSS feeds
    const allFeedsInCategory =
      store.feeds?.filter(
        (f) => f.category === categoryName || (categoryName === 'uncategorized' && !f.category)
      ) || [];
    if (allFeedsInCategory.some((f) => f.is_freshrss_source)) {
      console.log('[handleDrop] Cannot drop non-FreshRSS feed into FreshRSS category');
      window.showToast(t('freshRSSFeedLocked'), 'info');
      isDragging.value = false;
      return;
    }
  }

  try {
    // Keep isDragging true until after the data refreshes
    const result = await onDrop(categoryName, feeds);

    console.log('[handleDrop] Drop result:', result);

    if (result.success) {
      // Refresh feeds to show updated order
      await store.fetchFeeds();
      console.log(
        '[handleDrop] Feeds refreshed, tree.uncategorized.length:',
        tree.value.uncategorized.length
      );
      window.showToast(t('feedReordered'), 'success');
    } else {
      window.showToast(t('errorReorderingFeed') + ': ' + result.error, 'error');
    }
  } finally {
    // Always clear dragging state, even if there's an error
    console.log('[handleDrop] Clearing isDragging');
    isDragging.value = false;
  }
}

// Watch for drag end to clear isDragging if no drop occurred
// This handles the case where user drags outside sidebar and cancels
watch(draggingFeedId, (newValue, oldValue) => {
  // If draggingFeedId changes from non-null to null, but handleDrop wasn't called
  // (which would clear isDragging), we need to clear isDragging after a short delay
  if (oldValue !== null && newValue === null && !dropHandled) {
    setTimeout(() => {
      isDragging.value = false;
    }, 100);
  }
});

const emitShowAddFeed = () => window.dispatchEvent(new CustomEvent('show-add-feed'));
const emitShowSettings = () => window.dispatchEvent(new CustomEvent('show-settings'));
</script>

<template>
  <aside
    :class="[
      'sidebar flex flex-col bg-bg-secondary border-r border-border h-full transition-transform duration-300',
      'absolute z-20',
      isOpen ? 'translate-x-0' : '-translate-x-full',
      'md:relative md:translate-x-0',
    ]"
  >
    <div class="p-3 sm:p-5 border-b border-border flex justify-between items-center">
      <h2 class="m-0 text-base sm:text-lg font-bold flex items-center gap-1.5 sm:gap-2 text-accent">
        <img src="/assets/logo.svg" alt="Logo" class="h-6 sm:h-7 w-auto" />
        <span class="xs:inline">{{ t('appName') }}</span>
      </h2>
    </div>

    <nav class="p-2 sm:p-3 space-y-1">
      <SidebarNavItem
        :label="t('allArticles')"
        :is-active="store.currentFilter === 'all'"
        icon="all"
        :unread-count="store.unreadCounts.total"
        @click="store.setFilter('all')"
      />
      <SidebarNavItem
        :label="t('unread')"
        :is-active="store.currentFilter === 'unread'"
        icon="unread"
        @click="store.setFilter('unread')"
      />
      <SidebarNavItem
        :label="t('favorites')"
        :is-active="store.currentFilter === 'favorites'"
        icon="favorites"
        @click="store.setFilter('favorites')"
      />
      <SidebarNavItem
        :label="t('readLater')"
        :is-active="store.currentFilter === 'readLater'"
        icon="readLater"
        @click="store.setFilter('readLater')"
      />
      <SidebarNavItem
        v-if="imageGalleryEnabled"
        :label="t('imageGallery')"
        :is-active="store.currentFilter === 'imageGallery'"
        icon="imageGallery"
        @click="store.setFilter('imageGallery')"
      />
    </nav>

    <!-- Search Box (kept outside scrollable list so it doesn't scroll) -->
    <div class="px-2 sm:px-3 pt-2 border-t border-border bg-bg-secondary z-10">
      <div class="mb-3">
        <div
          class="flex items-center bg-bg-secondary border border-border rounded-lg px-3 py-2 focus-within:border-accent transition-colors"
        >
          <PhMagnifyingGlass :size="18" class="text-text-secondary mr-2 flex-shrink-0" />
          <input
            v-model="searchQuery"
            type="text"
            :placeholder="t('searchFeeds')"
            class="w-full bg-transparent border-none outline-none text-text-primary text-sm placeholder-text-secondary"
          />
          <button
            v-if="searchQuery"
            class="ml-2 p-0.5 text-text-secondary hover:text-text-primary hover:bg-bg-tertiary rounded transition-colors flex-shrink-0"
            :title="t('clear')"
            @click="searchQuery = ''"
          >
            <PhX :size="16" />
          </button>
        </div>
      </div>
    </div>

    <div class="flex-1 overflow-y-auto p-1.5 sm:p-2">
      <!-- Categories -->
      <SidebarCategory
        v-for="(data, name) in tree.tree"
        :key="name"
        :name="name"
        :feeds="data._feeds"
        :children="data._children"
        :level="0"
        :is-open="checkIsCategoryOpen(name)"
        :is-active="store.currentCategory === name"
        :unread-count="categoryUnreadCounts[name] || 0"
        :current-feed-id="store.currentFeedId"
        :feed-unread-counts="store.unreadCounts.feedCounts"
        :is-drag-over="dragOverCategory === name"
        :is-edit-mode="isEditMode"
        :drop-preview="dropPreview"
        :dragging-feed-id="draggingFeedId"
        :is-category-open="checkIsCategoryOpen"
        @toggle="toggleCategory(name)"
        @select-category="store.setCategory(name)"
        @select-feed="store.setFeed"
        @category-context-menu="(e) => onCategoryContextMenu(e, name)"
        @child-toggle="toggleCategory"
        @child-select-category="store.setCategory"
        @child-context-menu="(e, path) => onCategoryContextMenu(e, path)"
        @feed-context-menu="onFeedContextMenu"
        @dragstart="(feedId, e) => handleDragStart(feedId, e)"
        @dragend="handleDragEnd"
        @feed-drag-over="(feedId, e) => handleDragOver(name, feedId, e)"
        @dragleave="(categoryName, e) => handleDragLeave(categoryName, e)"
        @drop="() => handleDrop(name, data._feeds)"
      />

      <!-- Uncategorized -->
      <SidebarCategory
        v-if="tree.uncategorized.length > 0 || isDragging"
        :name="t('uncategorized')"
        :feeds="tree.uncategorized"
        :is-open="
          checkIsCategoryOpen('uncategorized') || (tree.uncategorized.length === 0 && isDragging)
        "
        :is-active="false"
        :is-uncategorized="true"
        :unread-count="categoryUnreadCounts['uncategorized'] || 0"
        :current-feed-id="store.currentFeedId"
        :feed-unread-counts="store.unreadCounts.feedCounts"
        :is-drag-over="dragOverCategory === 'uncategorized'"
        :is-edit-mode="isEditMode"
        :drop-preview="dropPreview"
        :dragging-feed-id="draggingFeedId"
        :is-category-open="checkIsCategoryOpen"
        @toggle="toggleCategory('uncategorized')"
        @select-feed="store.setFeed"
        @category-context-menu="(e) => onCategoryContextMenu(e, 'uncategorized')"
        @feed-context-menu="onFeedContextMenu"
        @dragstart="(feedId, e) => handleDragStart(feedId, e)"
        @dragend="handleDragEnd"
        @feed-drag-over="(feedId, e) => handleDragOver('uncategorized', feedId, e)"
        @dragleave="(categoryName, e) => handleDragLeave(categoryName, e)"
        @drop="() => handleDrop('uncategorized', tree.uncategorized)"
      />
    </div>

    <div class="p-2 sm:p-4 border-t border-border flex gap-1.5 sm:gap-2">
      <button class="footer-btn" :title="t('addFeed')" @click="emitShowAddFeed">
        <PhPlus :size="18" class="sm:w-5 sm:h-5" />
      </button>
      <button
        class="footer-btn"
        :class="{ 'text-accent': isEditMode }"
        :title="isEditMode ? t('done') : t('edit')"
        @click="toggleEditMode"
      >
        <PhPencil v-if="!isEditMode" :size="18" class="sm:w-5 sm:h-5" />
        <PhCheck v-else :size="18" class="sm:w-5 sm:h-5" />
      </button>
      <button class="footer-btn" :title="t('settings')" @click="emitShowSettings">
        <PhGear :size="18" class="sm:w-5 sm:h-5" />
      </button>
    </div>
  </aside>
  <!-- Overlay for mobile -->
  <div v-if="isOpen" class="fixed inset-0 bg-black/50 z-10 md:hidden" @click="emit('toggle')"></div>
</template>

<style scoped>
@reference "../../style.css";

.sidebar {
  width: 16rem;
}
@media (min-width: 768px) {
  .sidebar {
    width: var(--sidebar-width, 16rem);
  }
}
.footer-btn {
  @apply flex-1 flex items-center justify-center gap-2 p-2 sm:p-2.5 text-text-secondary rounded-lg text-lg sm:text-xl hover:bg-bg-tertiary hover:text-text-primary transition-colors;
}
</style>
