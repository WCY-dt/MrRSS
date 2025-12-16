<script setup lang="ts">
import { useAppStore } from './stores/app';
import { useI18n } from 'vue-i18n';
import Sidebar from './components/sidebar/Sidebar.vue';
import ArticleList from './components/article/ArticleList.vue';
import ArticleDetail from './components/article/ArticleDetail.vue';
import ImageGalleryView from './components/article/ImageGalleryView.vue';
import AddFeedModal from './components/modals/feed/AddFeedModal.vue';
import EditFeedModal from './components/modals/feed/EditFeedModal.vue';
import SettingsModal from './components/modals/SettingsModal.vue';
import DiscoverFeedsModal from './components/modals/discovery/DiscoverFeedsModal.vue';
import ContextMenu from './components/common/ContextMenu.vue';
import ConfirmDialog from './components/modals/common/ConfirmDialog.vue';
import InputDialog from './components/modals/common/InputDialog.vue';
import Toast from './components/common/Toast.vue';
import { onMounted, ref, computed } from 'vue';
import { useNotifications } from './composables/ui/useNotifications';
import { useKeyboardShortcuts } from './composables/ui/useKeyboardShortcuts';
import { useContextMenu } from './composables/ui/useContextMenu';
import { useResizablePanels } from './composables/ui/useResizablePanels';
import { useWindowState } from './composables/core/useWindowState';
import { usePlatform } from './composables/core/usePlatform';
import type { Feed } from './types/models';

const store = useAppStore();
const { t } = useI18n();

const showAddFeed = ref(false);
const showEditFeed = ref(false);
const feedToEdit = ref<Feed | null>(null);
const showSettings = ref(false);
const showDiscoverBlogs = ref(false);
const feedToDiscover = ref<Feed | null>(null);
const isSidebarOpen = ref(false);

// Check if we're in image gallery mode
const isImageGalleryMode = computed(() => store.currentFilter === 'imageGallery');

// Use composables
const { confirmDialog, inputDialog, toasts, removeToast, installGlobalHandlers } =
  useNotifications();

const { contextMenu, openContextMenu, handleContextMenuAction } = useContextMenu();

const { sidebarWidth, articleListWidth, startResizeSidebar, startResizeArticleList } =
  useResizablePanels();

// Initialize window state management
const windowState = useWindowState();
windowState.init();

// Detect platform for MacOS-specific styles
const { isMacOS } = usePlatform();

// Track fullscreen state for macOS
const isFullscreen = ref(false);

// Initialize keyboard shortcuts
const { shortcuts } = useKeyboardShortcuts({
  onOpenSettings: () => {
    showSettings.value = true;
  },
  onAddFeed: () => {
    showAddFeed.value = true;
  },
  onMarkAllRead: async () => {
    await store.markAllAsRead();
    window.showToast(t('markedAllAsRead'), 'success');
  },
});

onMounted(async () => {
  // Install global notification handlers
  installGlobalHandlers();

  // Track fullscreen state for macOS (to hide title bar padding)
  if (isMacOS.value) {
    const checkFullscreen = () => {
      isFullscreen.value =
        document.fullscreenElement !== null || (document as any).webkitFullscreenElement !== null;
    };

    // Check on load
    checkFullscreen();

    // Listen for fullscreen changes
    document.addEventListener('fullscreenchange', checkFullscreen);
    document.addEventListener('webkitfullscreenchange', checkFullscreen);
  }

  // Initialize theme system immediately (lightweight)
  store.initTheme();

  // Load remaining settings (theme and other settings are already loaded in main.ts)
  let updateInterval = 10;
  let lastArticleUpdate = '';

  try {
    const res = await fetch('/api/settings');
    const data = await res.json();

    // Apply saved theme preference (already applied in main.ts, but ensure it's set)
    if (data.theme) {
      store.setTheme(data.theme);
    }

    // Apply other settings
    if (data.update_interval) {
      updateInterval = parseInt(data.update_interval);
      store.startAutoRefresh(updateInterval);
    }

    if (data.last_article_update) {
      lastArticleUpdate = data.last_article_update;
    }

    // Load saved shortcuts
    if (data.shortcuts) {
      try {
        const parsed = JSON.parse(data.shortcuts);
        shortcuts.value = { ...shortcuts.value, ...parsed };
      } catch (e) {
        console.error('Error parsing shortcuts:', e);
      }
    }
  } catch (e) {
    console.error('Error loading initial settings:', e);
  }

  // Defer heavy operations to allow UI to render first
  setTimeout(() => {
    // Load feeds and articles in background
    store.fetchFeeds();
    store.fetchArticles();

    // Only trigger feed refresh if enough time has passed since last update
    setTimeout(() => {
      const shouldRefresh = shouldTriggerRefresh(lastArticleUpdate, updateInterval);
      if (shouldRefresh) {
        store.refreshFeeds();
      }
    }, 1000);
  }, 100);

  // Listen for events from Sidebar
  window.addEventListener('show-add-feed', () => (showAddFeed.value = true));
  window.addEventListener('show-edit-feed', (e: Event) => {
    const customEvent = e as CustomEvent;
    feedToEdit.value = customEvent.detail;
    showEditFeed.value = true;
  });
  window.addEventListener('show-settings', () => (showSettings.value = true));
  window.addEventListener('show-discover-blogs', (e: Event) => {
    const customEvent = e as CustomEvent;
    feedToDiscover.value = customEvent.detail;
    showDiscoverBlogs.value = true;
  });

  // Global Context Menu Event Listener
  window.addEventListener('open-context-menu', (e: Event) => {
    openContextMenu(e as CustomEvent);
  });
});

// Check if we should trigger refresh based on last update time and interval
function shouldTriggerRefresh(lastUpdate: string, intervalMinutes: number): boolean {
  if (!lastUpdate) {
    return true; // Never updated, should refresh
  }

  try {
    const lastUpdateTime = new Date(lastUpdate).getTime();
    const now = Date.now();
    const intervalMs = intervalMinutes * 60 * 1000;

    // Refresh if more than interval time has passed since last update
    return now - lastUpdateTime >= intervalMs;
  } catch {
    return true; // Invalid date, should refresh
  }
}

function toggleSidebar(): void {
  isSidebarOpen.value = !isSidebarOpen.value;
}

function onFeedAdded(): void {
  store.fetchFeeds();
  // Start polling for progress as the backend is now fetching articles for the new feed
  store.pollProgress();
}

function onFeedUpdated(): void {
  store.fetchFeeds();
  // Refresh articles to immediately apply hide_from_timeline changes
  store.fetchArticles();
}
</script>

<template>
  <div
    class="app-container flex h-screen w-full bg-bg-primary text-text-primary overflow-hidden"
    :class="{ 'macos-padding': isMacOS && !isFullscreen, 'macos-fullscreen': isMacOS && isFullscreen }"
    :style="{
      '--sidebar-width': sidebarWidth + 'px',
      '--article-list-width': articleListWidth + 'px',
    }"
  >
    <Sidebar :is-open="isSidebarOpen" @toggle="toggleSidebar" />

    <div class="resizer hidden md:block" @mousedown="startResizeSidebar"></div>

    <!-- Show ImageGalleryView when in image gallery mode -->
    <template v-if="isImageGalleryMode">
      <ImageGalleryView :is-sidebar-open="isSidebarOpen" @toggle-sidebar="toggleSidebar" />
    </template>

    <!-- Show ArticleList and ArticleDetail when not in image gallery mode -->
    <template v-else>
      <ArticleList :is-sidebar-open="isSidebarOpen" @toggle-sidebar="toggleSidebar" />

      <div class="resizer hidden md:block" @mousedown="startResizeArticleList"></div>

      <ArticleDetail />
    </template>

    <AddFeedModal v-if="showAddFeed" @close="showAddFeed = false" @added="onFeedAdded" />
    <EditFeedModal
      v-if="showEditFeed && feedToEdit"
      :feed="feedToEdit"
      @close="showEditFeed = false"
      @updated="onFeedUpdated"
    />
    <SettingsModal v-if="showSettings" @close="showSettings = false" />
    <DiscoverFeedsModal
      v-if="showDiscoverBlogs && feedToDiscover"
      :feed="feedToDiscover"
      :show="showDiscoverBlogs"
      @close="showDiscoverBlogs = false"
    />

    <ContextMenu
      v-if="contextMenu.show"
      :x="contextMenu.x"
      :y="contextMenu.y"
      :items="contextMenu.items"
      @close="contextMenu.show = false"
      @action="handleContextMenuAction"
    />

    <!-- Global Notification System -->
    <ConfirmDialog
      v-if="confirmDialog"
      :title="confirmDialog.title"
      :message="confirmDialog.message"
      :confirm-text="confirmDialog.confirmText"
      :cancel-text="confirmDialog.cancelText"
      :is-danger="confirmDialog.isDanger"
      @confirm="confirmDialog.onConfirm"
      @cancel="confirmDialog.onCancel"
      @close="confirmDialog = null"
    />

    <InputDialog
      v-if="inputDialog"
      :title="inputDialog.title"
      :message="inputDialog.message"
      :placeholder="inputDialog.placeholder"
      :default-value="inputDialog.defaultValue"
      :confirm-text="inputDialog.confirmText"
      :cancel-text="inputDialog.cancelText"
      @confirm="inputDialog.onConfirm"
      @cancel="inputDialog.onCancel"
      @close="inputDialog = null"
    />

    <div class="toast-container">
      <Toast
        v-for="toast in toasts"
        :key="toast.id"
        :message="toast.message"
        :type="toast.type"
        :duration="toast.duration"
        @close="removeToast(toast.id)"
      />
    </div>
  </div>
</template>

<style>
/* MacOS-specific styles */
.app-container.macos-padding {
  padding-top: 28px; /* Space for MacOS window controls */
}

/* MacOS fullscreen mode - hide the title bar padding */
.app-container.macos-fullscreen {
  padding-top: 0;
}

/* MacOS window dragging support - enable drag on top sections only */
.app-container.macos-padding .sidebar > div:first-child,
.app-container.macos-padding .article-list > div:first-child,
.app-container.macos-padding main > div > div:first-child {
  -webkit-app-region: drag;
}

/* Prevent drag on interactive elements within draggable areas on MacOS */
.app-container.macos-padding button,
.app-container.macos-padding input,
.app-container.macos-padding textarea,
.app-container.macos-padding select,
.app-container.macos-padding a,
.app-container.macos-padding .resizer,
.app-container.macos-padding [role='button'],
.app-container.macos-padding img[role='button'],
.app-container.macos-padding .clickable,
.app-container.macos-padding h2,
.app-container.macos-padding h3 {
  -webkit-app-region: no-drag;
}

.toast-container {
  position: fixed;
  top: 10px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 60;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  pointer-events: none;
}

/* Adjust toast position for MacOS */
.app-container.macos-padding .toast-container {
  top: 38px; /* Account for MacOS top padding */
}

.toast-container > * {
  pointer-events: auto;
}
@media (min-width: 640px) {
  .toast-container {
    top: 20px;
    gap: 10px;
  }
  .app-container.macos-padding .toast-container {
    top: 48px;
  }
}
.resizer {
  width: 4px;
  cursor: col-resize;
  background-color: transparent;
  flex-shrink: 0;
  transition: background-color 0.2s;
  z-index: 10;
  margin-left: -2px;
  margin-right: -2px;
}
.resizer:hover,
.resizer:active {
  background-color: var(--color-accent, #3b82f6);
}
/* Global styles if needed */
</style>
