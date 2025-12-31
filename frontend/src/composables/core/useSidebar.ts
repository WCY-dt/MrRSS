import { computed, ref, watch, type Ref } from 'vue';
import { useAppStore } from '@/stores/app';
import { useI18n } from 'vue-i18n';
import { openInBrowser } from '@/utils/browser';
import type { Feed } from '@/types/models';

interface TreeNode {
  _feeds: Feed[];
  _children: Record<string, TreeNode>;
  isOpen: boolean;
}

interface TreeData {
  tree: Record<string, TreeNode>;
  uncategorized: Feed[];
  categories: Set<string>;
}

export function useSidebar() {
  const store = useAppStore();
  const { t } = useI18n();

  // Load saved category state from localStorage
  const savedCategories = localStorage.getItem('openCategories');
  const openCategories: Ref<Set<string>> = ref(
    savedCategories ? new Set(JSON.parse(savedCategories)) : new Set()
  );
  const searchQuery: Ref<string> = ref('');

  // Build category tree with search filtering
  const tree = computed<TreeData>(() => {
    const t: Record<string, TreeNode> = {};
    const uncategorized: Feed[] = [];
    const categories = new Set<string>();

    if (!store.feeds) return { tree: {}, uncategorized: [], categories };

    const query = searchQuery.value.toLowerCase().trim();

    store.feeds.forEach((feed: Feed) => {
      const matchesSearch =
        query === '' ||
        feed.title.toLowerCase().includes(query) ||
        feed.url.toLowerCase().includes(query);

      if (!matchesSearch) return;

      if (feed.category) {
        const parts = feed.category.split('/');
        let currentLevel = t;
        parts.forEach((part, index) => {
          if (!currentLevel[part]) {
            currentLevel[part] = { _feeds: [], _children: {}, isOpen: false };
          }
          if (index === parts.length - 1) {
            currentLevel[part]._feeds.push(feed);
            categories.add(feed.category);
          } else {
            currentLevel = currentLevel[part]._children;
          }
        });
      } else {
        uncategorized.push(feed);
      }
    });
    if (uncategorized.length > 0) {
      categories.add('uncategorized');
    }
    return { tree: t, uncategorized, categories };
  });

  // Compute unread counts for categories
  const categoryUnreadCounts = computed<Record<string, number>>(() => {
    const counts: Record<string, number> = {};
    if (!store.feeds || !store.unreadCounts.feedCounts) return counts;

    store.feeds.forEach((feed: Feed) => {
      if (feed.category) {
        const unreadCount = store.unreadCounts.feedCounts[feed.id] || 0;
        if (unreadCount > 0) {
          counts[feed.category] = (counts[feed.category] || 0) + unreadCount;
        }
      }
    });

    // Calculate uncategorized count
    const uncategorizedFeeds = store.feeds.filter((f) => !f.category);
    counts['uncategorized'] = uncategorizedFeeds.reduce((sum, feed) => {
      return sum + (store.unreadCounts.feedCounts[feed.id] || 0);
    }, 0);

    return counts;
  });

  // Auto-expand new categories only if no saved state exists
  watch(
    () => tree.value.categories,
    (newCategories) => {
      if (newCategories) {
        const hasSavedState = localStorage.getItem('openCategories') !== null;
        newCategories.forEach((cat) => {
          // Always auto-expand 'uncategorized' category
          if (cat === 'uncategorized' && !openCategories.value.has(cat)) {
            openCategories.value.add(cat);
            return;
          }
          // Only auto-expand if this is a new category and no saved state exists
          if (!openCategories.value.has(cat) && !hasSavedState) {
            openCategories.value.add(cat);
          }
        });

        // Also auto-expand parent categories for multi-level
        // For example, if "Tech/Blogs" exists, also expand "Tech"
        const parentCategories = new Set<string>();
        newCategories.forEach((cat) => {
          const parts = cat.split('/');
          for (let i = 1; i < parts.length; i++) {
            const parentPath = parts.slice(0, i).join('/');
            parentCategories.add(parentPath);
          }
        });

        parentCategories.forEach((parentCat) => {
          if (!openCategories.value.has(parentCat) && !hasSavedState) {
            openCategories.value.add(parentCat);
          }
        });
      }
    },
    { immediate: true }
  );

  function toggleCategory(path: string): void {
    if (openCategories.value.has(path)) {
      openCategories.value.delete(path);
    } else {
      openCategories.value.add(path);
    }
    // Save to localStorage
    localStorage.setItem('openCategories', JSON.stringify([...openCategories.value]));
  }

  function isCategoryOpen(path: string): boolean {
    return openCategories.value.has(path);
  }

  // Feed actions
  async function handleFeedAction(action: string, feed: Feed): Promise<void> {
    if (action === 'markAllRead') {
      await store.markAllAsRead(feed.id);
      window.showToast(t('markedAllAsRead'), 'success');
    } else if (action === 'refreshFeed') {
      await fetch(`/api/feeds/refresh?id=${feed.id}`, { method: 'POST' });
      window.showToast(t('feedRefreshStarted'), 'success');
      // Start polling for progress as the backend is now fetching articles for this feed
      store.pollProgress();
    } else if (action === 'syncFeed') {
      // Sync individual FreshRSS feed
      await fetch(`/api/freshrss/sync-feed?stream_id=${feed.freshrss_stream_id}`, {
        method: 'POST',
      });
      window.showToast(t('syncFeedStarted'), 'success');
      // Start polling for progress
      store.pollProgress();
    } else if (action === 'delete') {
      const confirmed = await window.showConfirm({
        title: t('unsubscribeTitle'),
        message: t('unsubscribeMessage', { name: feed.title }),
        confirmText: t('unsubscribe'),
        cancelText: t('cancel'),
        isDanger: true,
      });
      if (confirmed) {
        await fetch(`/api/feeds/delete?id=${feed.id}`, { method: 'POST' });
        store.fetchFeeds();
        window.showToast(t('unsubscribedSuccess'), 'success');
      }
    } else if (action === 'edit') {
      window.dispatchEvent(new CustomEvent('show-edit-feed', { detail: feed }));
    } else if (action === 'openWebsite') {
      const urlToOpen = feed.website_url || feed.url;
      openInBrowser(urlToOpen);
    } else if (action === 'discover') {
      window.dispatchEvent(new CustomEvent('show-discover-blogs', { detail: feed }));
    }
  }

  function onFeedContextMenu(e: MouseEvent, feed: Feed): void {
    e.preventDefault();
    e.stopPropagation();

    // Build menu items dynamically based on whether this is a FreshRSS feed
    const items: Array<{
      label: string;
      action: string;
      icon: string;
      separator?: boolean;
      danger?: boolean;
    }> = [];

    // For FreshRSS feeds, show "Sync Feed" instead of "Refresh Feed"
    if (feed.is_freshrss_source) {
      items.push({ label: t('syncFeed'), action: 'syncFeed', icon: 'PhArrowsClockwise' });
    } else {
      items.push({ label: t('refreshFeed'), action: 'refreshFeed', icon: 'PhArrowsClockwise' });
    }

    items.push({ label: t('markAllAsReadFeed'), action: 'markAllRead', icon: 'PhCheckCircle' });
    items.push({ separator: true });
    items.push({ label: t('openWebsite'), action: 'openWebsite', icon: 'PhGlobe' });

    // Only add discover for non-FreshRSS feeds
    if (!feed.is_freshrss_source) {
      items.push({ label: t('discoverFeeds'), action: 'discover', icon: 'PhBinoculars' });
    }

    // Only add edit and delete options for non-FreshRSS feeds
    if (!feed.is_freshrss_source) {
      items.push({ separator: true });
      items.push({ label: t('editSubscription'), action: 'edit', icon: 'PhPencil' });
      items.push({ label: t('unsubscribe'), action: 'delete', icon: 'PhTrash', danger: true });
    }

    window.dispatchEvent(
      new CustomEvent('open-context-menu', {
        detail: {
          x: e.clientX,
          y: e.clientY,
          items,
          data: feed,
          callback: handleFeedAction,
        },
      })
    );
  }

  // Category actions
  async function handleCategoryAction(action: string, categoryName: string): Promise<void> {
    if (action === 'markAllRead') {
      // Use the category parameter for the API call
      const category = categoryName === 'uncategorized' ? '' : categoryName;
      await fetch(`/api/articles/mark-all-read?category=${encodeURIComponent(category)}`, {
        method: 'POST',
      });
      store.fetchUnreadCounts();
      window.showToast(t('markedAllAsRead'), 'success');
    } else if (action === 'rename') {
      const newName = await window.showInput({
        title: t('renameCategory'),
        message: t('enterCategoryName'),
        defaultValue: categoryName,
        confirmText: t('confirm'),
        cancelText: t('cancel'),
      });
      if (newName && newName !== categoryName) {
        const feedsToUpdate = store.feeds.filter(
          (f) => f.category === categoryName || f.category.startsWith(categoryName + '/')
        );

        const promises = feedsToUpdate.map((feed) => {
          let newCategory = feed.category;
          if (feed.category === categoryName) {
            newCategory = newName;
          } else if (feed.category.startsWith(categoryName + '/')) {
            newCategory = newName + feed.category.substring(categoryName.length);
          }

          return fetch('/api/feeds/update', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              id: feed.id,
              title: feed.title,
              url: feed.url,
              category: newCategory,
            }),
          });
        });

        await Promise.all(promises);
        store.fetchFeeds();
      }
    }
  }

  function onCategoryContextMenu(e: MouseEvent, categoryName: string): void {
    e.preventDefault();
    e.stopPropagation();

    const items: Array<{ label?: string; action?: string; icon?: string; separator?: boolean }> = [
      { label: t('markAllAsReadFeed'), action: 'markAllRead', icon: 'ph-check-circle' },
    ];

    if (categoryName !== 'uncategorized') {
      items.push({ separator: true });
      items.push({ label: t('renameCategory'), action: 'rename', icon: 'ph-pencil' });
    }

    window.dispatchEvent(
      new CustomEvent('open-context-menu', {
        detail: {
          x: e.clientX,
          y: e.clientY,
          items: items,
          data: categoryName,
          callback: handleCategoryAction,
        },
      })
    );
  }

  return {
    tree,
    categoryUnreadCounts,
    openCategories,
    searchQuery,
    toggleCategory,
    isCategoryOpen,
    onFeedContextMenu,
    onCategoryContextMenu,
  };
}
