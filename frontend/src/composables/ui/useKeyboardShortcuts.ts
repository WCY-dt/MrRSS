import { ref, onMounted, onBeforeUnmount } from 'vue';
import { useAppStore } from '@/stores/app';
import { openInBrowser } from '@/utils/browser';

export interface KeyboardShortcuts {
  nextArticle: string;
  previousArticle: string;
  openArticle: string;
  closeArticle: string;
  toggleReadStatus: string;
  toggleFavoriteStatus: string;
  toggleReadLaterStatus: string;
  openInBrowser: string;
  toggleContentView: string;
  refreshFeeds: string;
  markAllRead: string;
  openSettings: string;
  addFeed: string;
  focusSearch: string;
  toggleFilter: string;
  goToAllArticles: string;
  goToUnread: string;
  goToFavorites: string;
  goToReadLater: string;
}

export interface KeyboardShortcutCallbacks {
  onOpenSettings: () => void;
  onAddFeed: () => void;
  onMarkAllRead: () => Promise<void>;
}

export function useKeyboardShortcuts(callbacks: KeyboardShortcutCallbacks) {
  const store = useAppStore();

  const shortcutsEnabled = ref(true);
  const shortcuts = ref<KeyboardShortcuts>({
    nextArticle: 'j',
    previousArticle: 'k',
    openArticle: 'Enter',
    closeArticle: 'Escape',
    toggleReadStatus: 'r',
    toggleFavoriteStatus: 's',
    toggleReadLaterStatus: 'l',
    openInBrowser: 'o',
    toggleContentView: 'v',
    refreshFeeds: 'Shift+r',
    markAllRead: 'Shift+a',
    openSettings: ',',
    addFeed: 'a',
    focusSearch: '/',
    toggleFilter: 'f',
    goToAllArticles: '1',
    goToUnread: '2',
    goToFavorites: '3',
    goToReadLater: '4',
  });

  // Helper functions
  function buildKeyCombo(e: KeyboardEvent): string {
    let key = '';
    if (e.ctrlKey) key += 'Ctrl+';
    if (e.altKey) key += 'Alt+';
    if (e.shiftKey) key += 'Shift+';
    if (e.metaKey) key += 'Meta+';

    let actualKey = e.key;
    if (actualKey === ' ') actualKey = 'Space';
    else if (actualKey.length === 1) actualKey = actualKey.toLowerCase();

    key += actualKey;
    return key;
  }

  function navigateArticle(direction: number): void {
    const articles = store.articles;
    if (!articles || articles.length === 0) return;

    const currentIndex = store.currentArticleId
      ? articles.findIndex((a) => a.id === store.currentArticleId)
      : -1;

    let newIndex: number;
    if (currentIndex === -1) {
      newIndex = direction > 0 ? 0 : articles.length - 1;
    } else {
      newIndex = currentIndex + direction;
      if (newIndex < 0) newIndex = 0;
      if (newIndex >= articles.length) newIndex = articles.length - 1;
    }

    selectArticleByIndex(newIndex);
  }

  function selectArticleByIndex(index: number): void {
    const article = store.articles[index];
    if (!article) return;

    store.currentArticleId = article.id;

    // Mark as read
    if (!article.is_read) {
      article.is_read = true;
      fetch(`/api/articles/read?id=${article.id}&read=true`, { method: 'POST' })
        .then(() => store.fetchUnreadCounts())
        .catch((e) => console.error('Error marking as read:', e));
    }

    // Scroll the article into view
    setTimeout(() => {
      const articleEl = document.querySelector(`[data-article-id="${article.id}"]`);
      if (articleEl) {
        articleEl.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
      }
    }, 50);
  }

  function toggleCurrentArticleRead(): void {
    const article = store.articles.find((a) => a.id === store.currentArticleId);
    if (!article) return;

    const newState = !article.is_read;
    article.is_read = newState;
    fetch(`/api/articles/read?id=${article.id}&read=${newState}`, { method: 'POST' })
      .then(() => store.fetchUnreadCounts())
      .catch((e) => {
        console.error('Error toggling read:', e);
        article.is_read = !newState;
      });
  }

  function toggleCurrentArticleFavorite(): void {
    const article = store.articles.find((a) => a.id === store.currentArticleId);
    if (!article) return;

    const newState = !article.is_favorite;
    article.is_favorite = newState;
    fetch(`/api/articles/favorite?id=${article.id}`, { method: 'POST' }).catch((e) => {
      console.error('Error toggling favorite:', e);
      article.is_favorite = !newState;
    });
  }

  function toggleCurrentArticleReadLater(): void {
    const article = store.articles.find((a) => a.id === store.currentArticleId);
    if (!article) return;

    const newState = !article.is_read_later;
    article.is_read_later = newState;
    // When adding to read later, also mark as unread
    if (newState) {
      article.is_read = false;
    }
    fetch(`/api/articles/toggle-read-later?id=${article.id}`, { method: 'POST' })
      .then(() => store.fetchUnreadCounts())
      .catch((e) => {
        console.error('Error toggling read later:', e);
        article.is_read_later = !newState;
      });
  }

  function openCurrentArticleInBrowser(): void {
    const article = store.articles.find((a) => a.id === store.currentArticleId);
    if (article && article.url) {
      openInBrowser(article.url);
    }
  }

  function focusSearchInput(): void {
    const searchInput = document.querySelector('[data-search-input]') as HTMLInputElement;
    if (searchInput) {
      searchInput.focus();
    }
  }

  // Keyboard event handler
  function handleKeyboardShortcut(e: KeyboardEvent): void {
    // Skip if shortcuts are disabled
    if (!shortcutsEnabled.value) {
      return;
    }

    // Skip if we're in an input field, textarea, or contenteditable
    const target = e.target as HTMLElement;
    const tagName = target.tagName.toLowerCase();
    const isEditable = target.isContentEditable;
    const isInput = tagName === 'input' || tagName === 'textarea' || tagName === 'select';

    const key = buildKeyCombo(e);

    // Check for escape key to close modals first (always allow, even when shortcuts disabled)
    if (key === shortcuts.value.closeArticle) {
      // Check if there are any open modals
      const hasOpenModal = document.querySelector('[data-modal-open="true"]') !== null;

      if (!hasOpenModal) {
        // No modals open, handle article close
        if (store.currentArticleId) {
          store.currentArticleId = null;
          e.preventDefault();
        }
      }
      // If modals are open, let them handle ESC themselves
      return;
    }

    // Skip shortcuts if in input field (except escape)
    if (isInput || isEditable) {
      return;
    }

    // Match the key combination to a shortcut action
    const action = Object.entries(shortcuts.value).find(([, shortcut]) => shortcut === key)?.[0];

    if (!action) return;

    e.preventDefault();

    // Execute the action
    switch (action) {
      case 'nextArticle':
        navigateArticle(1);
        break;
      case 'previousArticle':
        navigateArticle(-1);
        break;
      case 'openArticle':
        if (store.articles.length > 0 && !store.currentArticleId) {
          selectArticleByIndex(0);
        }
        break;
      case 'toggleReadStatus':
        toggleCurrentArticleRead();
        break;
      case 'toggleFavoriteStatus':
        toggleCurrentArticleFavorite();
        break;
      case 'toggleReadLaterStatus':
        toggleCurrentArticleReadLater();
        break;
      case 'openInBrowser':
        openCurrentArticleInBrowser();
        break;
      case 'toggleContentView':
        window.dispatchEvent(new CustomEvent('toggle-content-view'));
        break;
      case 'refreshFeeds':
        store.refreshFeeds();
        break;
      case 'markAllRead':
        callbacks.onMarkAllRead();
        break;
      case 'openSettings':
        callbacks.onOpenSettings();
        break;
      case 'addFeed':
        callbacks.onAddFeed();
        break;
      case 'focusSearch':
        focusSearchInput();
        break;
      case 'toggleFilter':
        window.dispatchEvent(new CustomEvent('toggle-filter'));
        break;
      case 'goToAllArticles':
        store.setFilter('all');
        break;
      case 'goToUnread':
        store.setFilter('unread');
        break;
      case 'goToFavorites':
        store.setFilter('favorites');
        break;
      case 'goToReadLater':
        store.setFilter('readLater');
        break;
    }
  }

  // Handle shortcuts changed event
  function handleShortcutsChanged(e: Event): void {
    const customEvent = e as CustomEvent;
    if (customEvent.detail && customEvent.detail.shortcuts) {
      shortcuts.value = { ...shortcuts.value, ...customEvent.detail.shortcuts };
    }
  }

  // Handle shortcuts enabled changed event
  function handleShortcutsEnabledChanged(e: Event): void {
    const customEvent = e as CustomEvent;
    if (customEvent.detail && typeof customEvent.detail.enabled === 'boolean') {
      shortcutsEnabled.value = customEvent.detail.enabled;
    }
  }

  // Initialize shortcuts enabled state from settings
  function initializeShortcutsEnabled(): void {
    // Note: store.settings is not available in the store
    // The shortcuts_enabled state is initialized via the shortcuts-enabled-changed event
    // Default is true (enabled)
  }

  // Lifecycle
  onMounted(() => {
    initializeShortcutsEnabled();
    window.addEventListener('keydown', handleKeyboardShortcut);
    window.addEventListener('shortcuts-changed', handleShortcutsChanged);
    window.addEventListener('shortcuts-enabled-changed', handleShortcutsEnabledChanged);
  });

  onBeforeUnmount(() => {
    window.removeEventListener('keydown', handleKeyboardShortcut);
    window.removeEventListener('shortcuts-changed', handleShortcutsChanged);
    window.removeEventListener('shortcuts-enabled-changed', handleShortcutsEnabledChanged);
  });

  return {
    shortcuts,
  };
}
