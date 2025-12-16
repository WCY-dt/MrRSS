import { defineStore } from 'pinia';
import { ref, type Ref } from 'vue';
import type { Article, Feed, UnreadCounts, RefreshProgress } from '@/types/models';

// TEMPORARY WORKAROUND: Helper to get platform info for macOS Sequoia throttling
// This is a synchronous helper that reads from cached platform detection
let platformInfoCache: { needsUIThrottling: boolean } | null = null;
async function getPlatformInfo() {
  if (platformInfoCache) {
    return platformInfoCache;
  }
  try {
    const response = await fetch('/api/platform/info');
    if (response.ok) {
      const data = await response.json();
      platformInfoCache = { needsUIThrottling: data.needs_ui_throttling || false };
      if (platformInfoCache.needsUIThrottling) {
        console.log(
          '[macOS Sequoia Workaround] UI throttling enabled to prevent WKWebView crashes'
        );
      }
      return platformInfoCache;
    }
  } catch (error) {
    console.warn('Failed to fetch platform info:', error);
  }
  platformInfoCache = { needsUIThrottling: false };
  return platformInfoCache;
}

// TEMPORARY WORKAROUND: Get optimized interval for platform
function getOptimizedInterval(baseInterval: number): number {
  // If platform info is available and throttling is needed, increase interval
  if (platformInfoCache?.needsUIThrottling) {
    return Math.max(baseInterval * 2, 1000); // 2x interval, minimum 1 second
  }
  return baseInterval;
}

export type Filter = 'all' | 'unread' | 'favorites' | 'readLater' | 'imageGallery' | '';
export type ThemePreference = 'light' | 'dark' | 'auto';
export type Theme = 'light' | 'dark';

export interface AppState {
  articles: Ref<Article[]>;
  feeds: Ref<Feed[]>;
  unreadCounts: Ref<UnreadCounts>;
  currentFilter: Ref<Filter>;
  currentFeedId: Ref<number | null>;
  currentCategory: Ref<string | null>;
  currentArticleId: Ref<number | null>;
  isLoading: Ref<boolean>;
  page: Ref<number>;
  hasMore: Ref<boolean>;
  searchQuery: Ref<string>;
  themePreference: Ref<ThemePreference>;
  theme: Ref<Theme>;
  refreshProgress: Ref<RefreshProgress>;
}

export interface AppActions {
  setFilter: (Filter) => void;
  setFeed: (number) => void;
  setCategory: (string) => void;
  fetchArticles: (boolean?) => Promise<void>;
  loadMore: () => Promise<void>;
  fetchFeeds: () => Promise<void>;
  fetchUnreadCounts: () => Promise<void>;
  markAllAsRead: (number?) => Promise<void>;
  toggleTheme: () => void;
  setTheme: (ThemePreference) => void;
  applyTheme: () => void;
  initTheme: () => void;
  refreshFeeds: () => Promise<void>;
  pollProgress: () => void;
  checkForAppUpdates: () => Promise<void>;
  startAutoRefresh: (number) => void;
}

export const useAppStore = defineStore('app', () => {
  // State
  const articles = ref<Article[]>([]);
  const feeds = ref<Feed[]>([]);
  const unreadCounts = ref<UnreadCounts>({
    total: 0,
    feedCounts: {},
  });
  const currentFilter = ref<Filter>('all');
  const currentFeedId = ref<number | null>(null);
  const currentCategory = ref<string | null>(null);
  const currentArticleId = ref<number | null>(null);
  const isLoading = ref<boolean>(false);
  const page = ref<number>(1);
  const hasMore = ref<boolean>(true);
  const searchQuery = ref<string>('');
  const themePreference = ref<ThemePreference>(
    (localStorage.getItem('themePreference') as ThemePreference) || 'auto'
  );
  const theme = ref<Theme>('light');

  // Refresh progress
  const refreshProgress = ref<RefreshProgress>({ current: 0, total: 0, isRunning: false });
  let refreshInterval: ReturnType<typeof setInterval> | null = null;

  // TEMPORARY WORKAROUND: Initialize platform info for macOS Sequoia throttling
  // This runs once when the store is created
  getPlatformInfo().catch((error) => {
    console.warn('Failed to initialize platform info:', error);
  });

  // Actions - Article Management
  function setFilter(filter: Filter): void {
    currentFilter.value = filter;
    currentFeedId.value = null;
    currentCategory.value = null;
    page.value = 1;
    articles.value = [];
    hasMore.value = true;
    fetchArticles();
  }

  function setFeed(feedId: number): void {
    // Check if this feed is an image mode feed
    const feed = feeds.value.find((f) => f.id === feedId);
    if (feed?.is_image_mode) {
      // Switch to image gallery mode for image mode feeds
      currentFilter.value = 'imageGallery';
      currentFeedId.value = feedId;
      currentCategory.value = null;
      page.value = 1;
      articles.value = [];
      hasMore.value = true;
    } else {
      currentFilter.value = '';
      currentFeedId.value = feedId;
      currentCategory.value = null;
      page.value = 1;
      articles.value = [];
      hasMore.value = true;
      fetchArticles();
    }
  }

  function setCategory(category: string): void {
    currentFilter.value = '';
    currentFeedId.value = null;
    currentCategory.value = category;
    page.value = 1;
    articles.value = [];
    hasMore.value = true;
    fetchArticles();
  }

  async function fetchArticles(append: boolean = false): Promise<void> {
    if (isLoading.value) return;
    if (!append && !hasMore.value) hasMore.value = true;

    isLoading.value = true;
    const limit = 50;

    let url = `/api/articles?page=${page.value}&limit=${limit}`;
    if (currentFilter.value) url += `&filter=${currentFilter.value}`;
    if (currentFeedId.value) url += `&feed_id=${currentFeedId.value}`;
    if (currentCategory.value) url += `&category=${encodeURIComponent(currentCategory.value)}`;

    try {
      const res = await fetch(url);
      const data: Article[] = (await res.json()) || [];

      if (data.length < limit) {
        hasMore.value = false;
      }

      if (append) {
        articles.value = [...articles.value, ...data];
      } else {
        articles.value = data;
      }
    } catch {
      // Error handled silently
    } finally {
      isLoading.value = false;
    }
  }

  async function loadMore(): Promise<void> {
    if (hasMore.value && !isLoading.value) {
      page.value++;
      await fetchArticles(true);
    }
  }

  async function fetchFeeds(): Promise<void> {
    try {
      const res = await fetch('/api/feeds');
      const data: Feed[] = (await res.json()) || [];
      feeds.value = data;
      // Fetch unread counts after fetching feeds
      await fetchUnreadCounts();
    } catch {
      feeds.value = [];
    }
  }

  async function fetchUnreadCounts(): Promise<void> {
    try {
      const res = await fetch('/api/articles/unread-counts');
      const data = await res.json();
      unreadCounts.value = {
        total: data.total || 0,
        feedCounts: data.feed_counts || {},
      };
    } catch {
      unreadCounts.value = { total: 0, feedCounts: {} };
    }
  }

  async function markAllAsRead(feedId?: number): Promise<void> {
    try {
      const url = feedId
        ? `/api/articles/mark-all-read?feed_id=${feedId}`
        : '/api/articles/mark-all-read';
      await fetch(url, { method: 'POST' });
      // Refresh articles and unread counts
      await fetchArticles();
      await fetchUnreadCounts();
    } catch {
      // Error handled silently
    }
  }

  // Theme Management
  function toggleTheme(): void {
    // Cycle through: light -> dark -> auto -> light
    if (themePreference.value === 'light') {
      themePreference.value = 'dark';
    } else if (themePreference.value === 'dark') {
      themePreference.value = 'auto';
    } else {
      themePreference.value = 'light';
    }
    localStorage.setItem('themePreference', themePreference.value);
    applyTheme();
  }

  function setTheme(preference: ThemePreference): void {
    themePreference.value = preference;
    localStorage.setItem('themePreference', preference);
    applyTheme();
  }

  function applyTheme(): void {
    let actualTheme: Theme = themePreference.value as Theme;

    // If auto, detect system preference
    if (themePreference.value === 'auto') {
      actualTheme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    }

    theme.value = actualTheme;

    if (actualTheme === 'dark') {
      document.body.classList.add('dark-mode');
    } else {
      document.body.classList.remove('dark-mode');
    }
  }

  function initTheme(): void {
    // Listen for system theme changes
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    mediaQuery.addEventListener('change', () => {
      if (themePreference.value === 'auto') {
        applyTheme();
      }
    });

    // Apply initial theme
    applyTheme();
  }

  // Auto Refresh
  async function refreshFeeds(): Promise<void> {
    refreshProgress.value.isRunning = true;
    try {
      await fetch('/api/refresh', { method: 'POST' });
      pollProgress();
    } catch {
      refreshProgress.value.isRunning = false;
    }
  }

  function pollProgress(): void {
    let lastCurrent = 0;
    // TEMPORARY WORKAROUND: Use optimized interval for macOS Sequoia
    const pollInterval = getOptimizedInterval(500);
    const interval = setInterval(async () => {
      try {
        const res = await fetch('/api/progress');
        const data = await res.json();
        refreshProgress.value = {
          current: data.current,
          total: data.total,
          isRunning: data.is_running,
        };

        // Update unread counts whenever progress advances (but don't refresh articles to avoid disrupting scroll position)
        if (data.current > lastCurrent) {
          lastCurrent = data.current;
          fetchUnreadCounts();
        }

        if (!data.is_running) {
          clearInterval(interval);
          fetchFeeds();
          fetchArticles();
          fetchUnreadCounts();

          // Check for app updates after initial refresh completes
          checkForAppUpdates();
        }
      } catch {
        clearInterval(interval);
        refreshProgress.value.isRunning = false;
      }
    }, pollInterval);
  }

  async function checkForAppUpdates(): Promise<void> {
    try {
      const res = await fetch('/api/check-updates');
      if (res.ok) {
        const data = await res.json();

        // Only proceed if there's an update available and a download URL
        if (data.has_update && data.download_url) {
          // Show notification to user
          if (window.showToast) {
            window.showToast(`Update available: v${data.latest_version}`, 'info', 5000);
          }

          // Auto download and install in background
          autoDownloadAndInstall(data.download_url, data.asset_name);
        }
      }
    } catch {
      console.error('Auto-update check failed');
      // Silently fail - don't disrupt user experience
    }
  }

  async function autoDownloadAndInstall(downloadUrl: string, assetName?: string): Promise<void> {
    try {
      // Download the update in background
      const downloadRes = await fetch('/api/download-update', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          download_url: downloadUrl,
          asset_name: assetName,
        }),
      });

      if (!downloadRes.ok) {
        console.error('Auto-download failed');
        return;
      }

      const downloadData = await downloadRes.json();
      if (!downloadData.success || !downloadData.file_path) {
        console.error('Auto-download failed: Invalid response');
        return;
      }

      // Wait a moment to ensure file is fully written
      await new Promise((resolve) => setTimeout(resolve, 500));

      // Install the update
      const installRes = await fetch('/api/install-update', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          file_path: downloadData.file_path,
        }),
      });

      if (!installRes.ok) {
        console.error('Auto-install failed');
        return;
      }

      const installData = await installRes.json();
      if (installData.success && window.showToast) {
        window.showToast('Update installed. Restart to apply.', 'success');
      }
    } catch {
      console.error('Auto-update failed');
      // Silently fail - don't disrupt user experience
    }
  }

  function startAutoRefresh(minutes: number): void {
    if (refreshInterval) clearInterval(refreshInterval);
    if (minutes > 0) {
      refreshInterval = setInterval(
        () => {
          refreshFeeds();
        },
        minutes * 60 * 1000
      );
    }
  }

  return {
    // State
    articles,
    feeds,
    unreadCounts,
    currentFilter,
    currentFeedId,
    currentCategory,
    currentArticleId,
    isLoading,
    page,
    hasMore,
    searchQuery,
    themePreference,
    theme,
    refreshProgress,

    // Actions
    setFilter,
    setFeed,
    setCategory,
    fetchArticles,
    loadMore,
    fetchFeeds,
    fetchUnreadCounts,
    markAllAsRead,
    toggleTheme,
    setTheme,
    applyTheme,
    initTheme,
    refreshFeeds,
    pollProgress,
    checkForAppUpdates,
    startAutoRefresh,
  };
});
