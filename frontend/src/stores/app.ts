import { defineStore } from 'pinia';
import { ref, type Ref } from 'vue';
import { useI18n } from 'vue-i18n';
import type { Article, Feed, UnreadCounts, RefreshProgress } from '@/types/models';

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
  showOnlyUnread: Ref<boolean>;
}

export interface AppActions {
  setFilter: (filter: Filter) => void;
  setFeed: (feedId: number) => void;
  setCategory: (category: string) => void;
  fetchArticles: (append?: boolean) => Promise<void>;
  loadMore: () => Promise<void>;
  fetchFeeds: () => Promise<void>;
  fetchUnreadCounts: () => Promise<void>;
  markAllAsRead: (feedId?: number) => Promise<void>;
  updateArticleSummary: (articleId: number, summary: string) => void;
  toggleTheme: () => void;
  setTheme: (preference: ThemePreference) => void;
  applyTheme: () => void;
  initTheme: () => void;
  refreshFeeds: () => Promise<void>;
  pollProgress: () => void;
  checkForAppUpdates: () => Promise<void>;
  startAutoRefresh: (minutes: number) => void;
  toggleShowOnlyUnread: () => void;
}

export const useAppStore = defineStore('app', () => {
  const { t } = useI18n();

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
  const showOnlyUnread = ref<boolean>(localStorage.getItem('showOnlyUnread') === 'true');

  // Refresh progress
  const refreshProgress = ref<RefreshProgress>({ isRunning: false });
  let refreshInterval: ReturnType<typeof setInterval> | null = null;

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

  async function markAllAsRead(feedId?: number, category?: string): Promise<void> {
    try {
      const params = new URLSearchParams();
      if (feedId) params.append('feed_id', String(feedId));
      if (category) params.append('category', category);

      const url = params.toString()
        ? `/api/articles/mark-all-read?${params.toString()}`
        : '/api/articles/mark-all-read';
      await fetch(url, { method: 'POST' });
      // Refresh articles and unread counts
      await fetchArticles();
      await fetchUnreadCounts();
    } catch {
      // Error handled silently
    }
  }

  // Update article summary in store
  function updateArticleSummary(articleId: number, summary: string): void {
    const articleIndex = articles.value.findIndex((a) => a.id === articleId);
    if (articleIndex !== -1) {
      articles.value[articleIndex] = {
        ...articles.value[articleIndex],
        summary,
      };
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
      // Immediately fetch progress once before starting polling
      await fetchProgressOnce();
      pollProgress();
    } catch {
      refreshProgress.value.isRunning = false;
    }
  }

  async function fetchProgressOnce(): Promise<void> {
    try {
      // Wait a bit for the backend to start processing
      await new Promise((resolve) => setTimeout(resolve, 100));

      const res = await fetch('/api/progress');
      const data = await res.json();
      console.log('Initial progress update:', data);
      refreshProgress.value = {
        ...refreshProgress.value,
        isRunning: data.is_running,
        errors: data.errors,
        pool_task_count: data.pool_task_count,
        article_click_count: data.article_click_count,
        queue_task_count: data.queue_task_count,
      };
      console.log('Initial refreshProgress:', refreshProgress.value);
    } catch (e) {
      console.error('Error fetching initial progress:', e);
    }
  }

  function pollProgress(): void {
    // Track previous pool/queue counts to detect task completion
    let previousPoolCount = 0;
    let previousQueueCount = 0;

    const interval = setInterval(async () => {
      try {
        const res = await fetch('/api/progress');
        const data = await res.json();
        refreshProgress.value = {
          ...refreshProgress.value, // Preserve existing pool_tasks and queue_tasks
          isRunning: data.is_running,
          errors: data.errors,
          pool_task_count: data.pool_task_count ?? 0,
          article_click_count: data.article_click_count ?? 0,
          queue_task_count: data.queue_task_count ?? 0,
        };

        // Fetch task details if refresh is running
        if (data.is_running && (data.pool_task_count > 0 || data.queue_task_count > 0)) {
          await fetchTaskDetails();
        }

        // Detect task completion and update unread counts immediately
        const currentPoolCount = data.pool_task_count ?? 0;
        const currentQueueCount = data.queue_task_count ?? 0;
        const totalTasks = currentPoolCount + currentQueueCount;
        const previousTotal = previousPoolCount + previousQueueCount;

        // If task count decreased, tasks completed - update unread counts
        if (totalTasks < previousTotal && previousTotal > 0) {
          fetchUnreadCounts();
          fetchFeeds(); // Also update feeds to refresh error marks
        }

        // Update previous counts
        previousPoolCount = currentPoolCount;
        previousQueueCount = currentQueueCount;

        if (!data.is_running) {
          clearInterval(interval);
          fetchFeeds();
          fetchArticles();
          fetchUnreadCounts();

          // Notify components that settings have been updated (e.g., last_article_update)
          // This triggers components using useSettings() to refresh their settings
          window.dispatchEvent(new CustomEvent('settings-updated'));

          // Note: We no longer show error toasts for failed feeds
          // Users can see error status in the feed list sidebar

          // Check for app updates after initial refresh completes
          checkForAppUpdates();
        }
      } catch {
        clearInterval(interval);
        refreshProgress.value.isRunning = false;
      }
    }, 500);
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

  function toggleShowOnlyUnread(): void {
    showOnlyUnread.value = !showOnlyUnread.value;
    localStorage.setItem('showOnlyUnread', String(showOnlyUnread.value));
  }

  async function fetchTaskDetails(): Promise<void> {
    try {
      const res = await fetch('/api/progress/task-details');
      if (res.ok) {
        const data = await res.json();
        refreshProgress.value = {
          ...refreshProgress.value,
          pool_tasks: data.pool_tasks,
          queue_tasks: data.queue_tasks,
        };
      }
    } catch (e) {
      console.error('Error fetching task details:', e);
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
    showOnlyUnread,

    // Actions
    setFilter,
    setFeed,
    setCategory,
    fetchArticles,
    loadMore,
    fetchFeeds,
    fetchUnreadCounts,
    markAllAsRead,
    updateArticleSummary,
    toggleTheme,
    setTheme,
    applyTheme,
    initTheme,
    refreshFeeds,
    pollProgress,
    checkForAppUpdates,
    startAutoRefresh,
    toggleShowOnlyUnread,
    fetchTaskDetails,
  };
});
