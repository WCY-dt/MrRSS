<script setup>
import { store } from './store.js';
import Sidebar from './components/Sidebar.vue';
import ArticleList from './components/ArticleList.vue';
import ArticleDetail from './components/ArticleDetail.vue';
import AddFeedModal from './components/modals/AddFeedModal.vue';
import EditFeedModal from './components/modals/EditFeedModal.vue';
import SettingsModal from './components/modals/SettingsModal.vue';
import DiscoverFeedsModal from './components/modals/DiscoverFeedsModal.vue';
import ContextMenu from './components/ContextMenu.vue';
import ConfirmDialog from './components/modals/ConfirmDialog.vue';
import InputDialog from './components/modals/InputDialog.vue';
import Toast from './components/Toast.vue';
import { BrowserOpenURL } from './wailsjs/wailsjs/runtime/runtime.js';
import { onMounted, onBeforeUnmount, ref, computed } from 'vue';

const showAddFeed = ref(false);
const showEditFeed = ref(false);
const feedToEdit = ref(null);
const showSettings = ref(false);
const showDiscoverBlogs = ref(false);
const feedToDiscover = ref(null);
const isSidebarOpen = ref(false);

// Global notification system
const confirmDialog = ref(null);
const inputDialog = ref(null);
const toasts = ref([]);

// Computed property to check if any modal is open (for keyboard shortcut handling)
const isAnyModalOpen = computed(() => {
    return showSettings.value || showAddFeed.value || showEditFeed.value || 
           showDiscoverBlogs.value || confirmDialog.value || inputDialog.value;
});

// Keyboard shortcuts
const shortcuts = ref({
    nextArticle: 'j',
    previousArticle: 'k',
    openArticle: 'Enter',
    closeArticle: 'Escape',
    toggleReadStatus: 'r',
    toggleFavoriteStatus: 's',
    openInBrowser: 'o',
    toggleContentView: 'v',
    refreshFeeds: 'Shift+r',
    markAllRead: 'Shift+a',
    openSettings: ',',
    addFeed: 'a',
    focusSearch: '/',
    goToAllArticles: '1',
    goToUnread: '2',
    goToFavorites: '3'
});

function showConfirm(options) {
    return new Promise((resolve) => {
        confirmDialog.value = {
            ...options,
            onConfirm: () => {
                confirmDialog.value = null;
                resolve(true);
            },
            onCancel: () => {
                confirmDialog.value = null;
                resolve(false);
            }
        };
    });
}

function showInput(options) {
    return new Promise((resolve) => {
        inputDialog.value = {
            ...options,
            onConfirm: (value) => {
                inputDialog.value = null;
                resolve(value);
            },
            onCancel: () => {
                inputDialog.value = null;
                resolve(null);
            }
        };
    });
}

function showToast(message, type = 'info', duration = 3000) {
    const id = Date.now();
    toasts.value.push({ id, message, type, duration });
}

// Make these available globally
window.showConfirm = showConfirm;
window.showInput = showInput;
window.showToast = showToast;

// Resizable columns state
const sidebarWidth = ref(256);
const articleListWidth = ref(400);
const isResizingSidebar = ref(false);
const isResizingArticleList = ref(false);

function startResizeSidebar(e) {
    isResizingSidebar.value = true;
    document.body.style.cursor = 'col-resize';
    document.body.style.userSelect = 'none';
    window.addEventListener('mousemove', handleResizeSidebar);
    window.addEventListener('mouseup', stopResizeSidebar);
}

function handleResizeSidebar(e) {
    if (!isResizingSidebar.value) return;
    const newWidth = e.clientX;
    if (newWidth >= 180 && newWidth <= 450) {
        sidebarWidth.value = newWidth;
    }
}

function stopResizeSidebar() {
    isResizingSidebar.value = false;
    document.body.style.cursor = '';
    document.body.style.userSelect = '';
    window.removeEventListener('mousemove', handleResizeSidebar);
    window.removeEventListener('mouseup', stopResizeSidebar);
}

function startResizeArticleList(e) {
    isResizingArticleList.value = true;
    document.body.style.cursor = 'col-resize';
    document.body.style.userSelect = 'none';
    window.addEventListener('mousemove', handleResizeArticleList);
    window.addEventListener('mouseup', stopResizeArticleList);
}

function handleResizeArticleList(e) {
    if (!isResizingArticleList.value) return;
    // Assuming sidebar is visible and at the left
    const newWidth = e.clientX - sidebarWidth.value;
    if (newWidth >= 250 && newWidth <= 600) {
        articleListWidth.value = newWidth;
    }
}

function stopResizeArticleList() {
    isResizingArticleList.value = false;
    document.body.style.cursor = '';
    document.body.style.userSelect = '';
    window.removeEventListener('mousemove', handleResizeArticleList);
    window.removeEventListener('mouseup', stopResizeArticleList);
}

// Context Menu State
const contextMenu = ref({
    show: false,
    x: 0,
    y: 0,
    items: [],
    data: null
});

// Keyboard shortcut handler
function buildKeyCombo(e) {
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

function handleKeyboardShortcut(e) {
    // Skip if we're in an input field, textarea, or contenteditable
    const target = e.target;
    const tagName = target.tagName.toLowerCase();
    const isEditable = target.isContentEditable;
    const isInput = tagName === 'input' || tagName === 'textarea' || tagName === 'select';
    
    // Allow certain shortcuts even in input fields
    const key = buildKeyCombo(e);
    
    // Check for escape key to close modals first (always allow)
    if (key === shortcuts.value.closeArticle) {
        // Close modals in order of priority
        if (showSettings.value) {
            showSettings.value = false;
            e.preventDefault();
            return;
        }
        if (showAddFeed.value) {
            showAddFeed.value = false;
            e.preventDefault();
            return;
        }
        if (showEditFeed.value) {
            showEditFeed.value = false;
            e.preventDefault();
            return;
        }
        if (showDiscoverBlogs.value) {
            showDiscoverBlogs.value = false;
            e.preventDefault();
            return;
        }
        if (contextMenu.value.show) {
            contextMenu.value.show = false;
            e.preventDefault();
            return;
        }
        if (store.currentArticleId) {
            store.currentArticleId = null;
            e.preventDefault();
            return;
        }
        return;
    }
    
    // Skip shortcuts if in input field (except escape)
    if (isInput || isEditable) {
        return;
    }
    
    // Skip if a modal is open (except escape which is handled above)
    if (isAnyModalOpen.value) {
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
            markAllAsRead();
            break;
        case 'openSettings':
            showSettings.value = true;
            break;
        case 'addFeed':
            showAddFeed.value = true;
            break;
        case 'focusSearch':
            focusSearchInput();
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
    }
}

function navigateArticle(direction) {
    const articles = store.articles;
    if (!articles || articles.length === 0) return;
    
    const currentIndex = store.currentArticleId 
        ? articles.findIndex(a => a.id === store.currentArticleId)
        : -1;
    
    let newIndex;
    if (currentIndex === -1) {
        // No article selected, select first or last based on direction
        newIndex = direction > 0 ? 0 : articles.length - 1;
    } else {
        newIndex = currentIndex + direction;
        // Clamp to valid range
        if (newIndex < 0) newIndex = 0;
        if (newIndex >= articles.length) newIndex = articles.length - 1;
    }
    
    selectArticleByIndex(newIndex);
}

function selectArticleByIndex(index) {
    const article = store.articles[index];
    if (!article) return;
    
    store.currentArticleId = article.id;
    
    // Mark as read
    if (!article.is_read) {
        article.is_read = true;
        fetch(`/api/articles/read?id=${article.id}&read=true`, { method: 'POST' })
            .then(() => store.fetchUnreadCounts())
            .catch(e => console.error('Error marking as read:', e));
    }
    
    // Scroll the article into view
    setTimeout(() => {
        const articleEl = document.querySelector(`[data-article-id="${article.id}"]`);
        if (articleEl) {
            articleEl.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
        }
    }, 50);
}

function toggleCurrentArticleRead() {
    const article = store.articles.find(a => a.id === store.currentArticleId);
    if (!article) return;
    
    const newState = !article.is_read;
    article.is_read = newState;
    fetch(`/api/articles/read?id=${article.id}&read=${newState}`, { method: 'POST' })
        .then(() => store.fetchUnreadCounts())
        .catch(e => {
            console.error('Error toggling read:', e);
            article.is_read = !newState;
        });
}

function toggleCurrentArticleFavorite() {
    const article = store.articles.find(a => a.id === store.currentArticleId);
    if (!article) return;
    
    const newState = !article.is_favorite;
    article.is_favorite = newState;
    fetch(`/api/articles/favorite?id=${article.id}`, { method: 'POST' })
        .catch(e => {
            console.error('Error toggling favorite:', e);
            article.is_favorite = !newState;
        });
}

function openCurrentArticleInBrowser() {
    const article = store.articles.find(a => a.id === store.currentArticleId);
    if (article && article.url) {
        BrowserOpenURL(article.url);
    }
}

async function markAllAsRead() {
    await store.markAllAsRead();
    window.showToast(store.i18n.t('markedAllAsRead'), 'success');
}

function focusSearchInput() {
    const searchInput = document.querySelector('[data-search-input]');
    if (searchInput) {
        searchInput.focus();
    }
}

// Handle shortcuts changed event
function handleShortcutsChanged(e) {
    if (e.detail && e.detail.shortcuts) {
        shortcuts.value = { ...shortcuts.value, ...e.detail.shortcuts };
    }
}

onMounted(async () => {
    // Initialize theme system immediately (lightweight)
    store.initTheme();
    
    // Add keyboard shortcut listener
    window.addEventListener('keydown', handleKeyboardShortcut);
    
    // Listen for shortcuts changes
    window.addEventListener('shortcuts-changed', handleShortcutsChanged);
    
    // Defer heavy operations to allow UI to render first
    setTimeout(() => {
        // Load feeds and articles in background
        store.fetchFeeds();
        store.fetchArticles();
        
        // Trigger feed refresh after initial load
        setTimeout(() => {
            store.refreshFeeds();
        }, 1000);
    }, 100);
    
    // Initialize settings asynchronously
    setTimeout(async () => {
        try {
            const res = await fetch('/api/settings');
            const data = await res.json();
            if (data.update_interval) {
                store.startAutoRefresh(parseInt(data.update_interval));
            }
            // Apply saved theme preference
            if (data.theme) {
                store.setTheme(data.theme);
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
            console.error(e);
        }
    }, 200);
    
    // Listen for events from Sidebar
    window.addEventListener('show-add-feed', () => showAddFeed.value = true);
    window.addEventListener('show-edit-feed', (e) => {
        feedToEdit.value = e.detail;
        showEditFeed.value = true;
    });
    window.addEventListener('show-settings', () => showSettings.value = true);
    window.addEventListener('show-discover-blogs', (e) => {
        feedToDiscover.value = e.detail;
        showDiscoverBlogs.value = true;
    });
    
    // Global Context Menu Event Listener
    window.addEventListener('open-context-menu', (e) => {
        contextMenu.value = {
            show: true,
            x: e.detail.x,
            y: e.detail.y,
            items: e.detail.items,
            data: e.detail.data,
            callback: e.detail.callback
        };
    });
});

onBeforeUnmount(() => {
    window.removeEventListener('keydown', handleKeyboardShortcut);
    window.removeEventListener('shortcuts-changed', handleShortcutsChanged);
});

function toggleSidebar() {
    isSidebarOpen.value = !isSidebarOpen.value;
}

function onFeedAdded() {
    store.fetchFeeds();
    store.fetchArticles(); // Refresh articles too
}

function onFeedUpdated() {
    store.fetchFeeds();
}

function handleContextMenuAction(action) {
    if (contextMenu.value.callback) {
        contextMenu.value.callback(action, contextMenu.value.data);
    }
}
</script>

<template>
    <div class="app-container flex h-screen w-full bg-bg-primary text-text-primary overflow-hidden"
         :style="{ '--sidebar-width': sidebarWidth + 'px', '--article-list-width': articleListWidth + 'px' }">
        <Sidebar :isOpen="isSidebarOpen" @toggle="toggleSidebar" />
        
        <div class="resizer hidden md:block" @mousedown="startResizeSidebar"></div>
        
        <ArticleList :isSidebarOpen="isSidebarOpen" @toggleSidebar="toggleSidebar" />
        
        <div class="resizer hidden md:block" @mousedown="startResizeArticleList"></div>
        
        <ArticleDetail />
        
        <AddFeedModal v-if="showAddFeed" @close="showAddFeed = false" @added="onFeedAdded" />
        <EditFeedModal v-if="showEditFeed" :feed="feedToEdit" @close="showEditFeed = false" @updated="onFeedUpdated" />
        <SettingsModal v-if="showSettings" @close="showSettings = false" />
        <DiscoverFeedsModal v-if="showDiscoverBlogs && feedToDiscover" 
                            :feed="feedToDiscover" 
                            :show="showDiscoverBlogs"
                            @close="showDiscoverBlogs = false" />
        
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
            :confirmText="confirmDialog.confirmText"
            :cancelText="confirmDialog.cancelText"
            :isDanger="confirmDialog.isDanger"
            @confirm="confirmDialog.onConfirm"
            @cancel="confirmDialog.onCancel"
            @close="confirmDialog = null"
        />
        
        <InputDialog 
            v-if="inputDialog"
            :title="inputDialog.title"
            :message="inputDialog.message"
            :placeholder="inputDialog.placeholder"
            :defaultValue="inputDialog.defaultValue"
            :confirmText="inputDialog.confirmText"
            :cancelText="inputDialog.cancelText"
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
                @close="toasts = toasts.filter(t => t.id !== toast.id)"
            />
        </div>
    </div>
</template>

<style>
.toast-container {
    position: fixed;
    top: 10px;
    right: 10px;
    left: 10px;
    z-index: 60;
    display: flex;
    flex-direction: column;
    gap: 8px;
    pointer-events: none;
}
.toast-container > * {
    pointer-events: auto;
}
@media (min-width: 640px) {
    .toast-container {
        top: 20px;
        right: 20px;
        left: auto;
        gap: 10px;
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
.resizer:hover, .resizer:active {
    background-color: var(--color-accent, #3b82f6);
}
/* Global styles if needed */
</style>
