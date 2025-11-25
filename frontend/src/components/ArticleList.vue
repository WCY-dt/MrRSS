<script setup>
import { store } from '../store.js';
import { ref, computed, onMounted, onBeforeUnmount, watch, nextTick } from 'vue';
import { BrowserOpenURL } from '../wailsjs/wailsjs/runtime/runtime.js';
import { 
    PhCheckCircle, PhArrowClockwise, PhList, PhMagnifyingGlass, 
    PhEyeSlash, PhStar, PhSpinner, PhFunnel 
} from "@phosphor-icons/vue";
import ArticleFilterModal from './modals/ArticleFilterModal.vue';

const listRef = ref(null);
const translationSettings = ref({
    enabled: false,
    targetLang: 'en'
});
const translatingArticles = ref(new Set());
const defaultViewMode = ref('original'); // Track default view mode for context menu

// Filter state
const showFilterModal = ref(false);
const activeFilters = ref([]);
const filteredArticlesFromServer = ref([]);
const isFilterLoading = ref(false);
const filterPage = ref(1);
const filterHasMore = ref(true);
const filterTotal = ref(0);

const props = defineProps(['isSidebarOpen']);
const emit = defineEmits(['toggleSidebar']);

// Virtual scrolling configuration
const ITEM_HEIGHT_MOBILE = 70; // Estimated height of each article card on mobile
const ITEM_HEIGHT_DESKTOP = 80; // Estimated height on desktop (sm: breakpoint and above)
const BUFFER_SIZE = 10; // Number of extra items to render above/below viewport

// Virtual scrolling state
const scrollTop = ref(0);
const containerHeight = ref(600); // Will be updated on mount and resize
const isMobile = ref(window.innerWidth < 640); // 640px is Tailwind's sm: breakpoint

// Computed item height based on screen size
const itemHeight = computed(() => isMobile.value ? ITEM_HEIGHT_MOBILE : ITEM_HEIGHT_DESKTOP);

// Update container height and mobile detection for virtual scrolling
function updateContainerHeight() {
    if (listRef.value) {
        containerHeight.value = listRef.value.clientHeight;
    }
    isMobile.value = window.innerWidth < 640;
}

// Load translation settings
onMounted(async () => {
    // Initialize container height for virtual scrolling
    nextTick(() => {
        updateContainerHeight();
    });
    window.addEventListener('resize', updateContainerHeight);
    
    try {
        const res = await fetch('/api/settings');
        const data = await res.json();
        translationSettings.value = {
            enabled: data.translation_enabled === 'true',
            targetLang: data.target_language || 'en'
        };
        defaultViewMode.value = data.default_view_mode || 'original';
        
        // Set up intersection observer for auto-translation
        if (translationSettings.value.enabled) {
            setupIntersectionObserver();
        }
    } catch (e) {
        console.error('Error loading translation settings:', e);
    }
    
    // Listen for translation settings changes
    window.addEventListener('translation-settings-changed', handleTranslationSettingsChange);
    // Listen for default view mode changes
    window.addEventListener('default-view-mode-changed', handleDefaultViewModeChange);
});

onBeforeUnmount(() => {
    if (observer) {
        observer.disconnect();
        observer = null;
    }
    window.removeEventListener('resize', updateContainerHeight);
    window.removeEventListener('translation-settings-changed', handleTranslationSettingsChange);
    window.removeEventListener('default-view-mode-changed', handleDefaultViewModeChange);
});

// Handle default view mode changes
function handleDefaultViewModeChange(e) {
    defaultViewMode.value = e.detail.mode;
}

// Handle translation settings changes
function handleTranslationSettingsChange(e) {
    const { enabled, targetLang } = e.detail;
    translationSettings.value = { enabled, targetLang };
    
    // Disconnect observer if translation is disabled
    if (!enabled && observer) {
        observer.disconnect();
        observer = null;
    }
    // Set up observer if translation is enabled
    else if (enabled && !observer) {
        setupIntersectionObserver();
        // Observe all current article cards
        setTimeout(() => {
            const cards = document.querySelectorAll('[data-article-id]');
            cards.forEach(card => observer.observe(card));
        }, 100);
    }
}

// Intersection Observer for auto-translation
let observer = null;

function setupIntersectionObserver() {
    // Use listRef if available, otherwise fall back to viewport (null)
    observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                const articleId = parseInt(entry.target.dataset.articleId);
                const article = store.articles.find(a => a.id === articleId);
                
                // Only translate if article exists, has no translation, and is not already being translated
                if (article && !article.translated_title && !translatingArticles.value.has(articleId)) {
                    translateArticle(article);
                }
            }
        });
    }, {
        root: listRef.value || null,
        rootMargin: '100px',
        threshold: 0.1
    });
}

async function translateArticle(article) {
    if (translatingArticles.value.has(article.id)) return;
    
    translatingArticles.value.add(article.id);
    
    try {
        const res = await fetch('/api/articles/translate', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                article_id: article.id,
                title: article.title,
                target_language: translationSettings.value.targetLang
            })
        });
        
        if (res.ok) {
            const data = await res.json();
            // Update the article in the store
            article.translated_title = data.translated_title;
        }
    } catch (e) {
        console.error('Error translating article:', e);
    } finally {
        translatingArticles.value.delete(article.id);
    }
}

function observeArticle(el) {
    if (el && observer && translationSettings.value.enabled) {
        observer.observe(el);
    }
}

function handleScroll(e) {
    const target = e.target;
    // Update scroll position for virtual scrolling
    scrollTop.value = target.scrollTop;
    
    // Load more when near bottom
    if (target.scrollTop + target.clientHeight >= target.scrollHeight - 200) {
        // If filters are active, load more filtered results, otherwise use store's loadMore
        if (activeFilters.value.length > 0) {
            loadMoreFilteredArticles();
        } else {
            store.loadMore();
        }
    }
}

function selectArticle(article) {
    // If clicking the same article, close the detail view
    if (store.currentArticleId === article.id) {
        store.currentArticleId = null;
        return;
    }
    
    store.currentArticleId = article.id;
    if (!article.is_read) {
        article.is_read = true;
        fetch(`/api/articles/read?id=${article.id}&read=true`, { method: 'POST' })
            .then(() => {
                // Update unread counts after marking as read
                store.fetchUnreadCounts();
            })
            .catch(e => {
                console.error('Error marking as read:', e);
                // Continue anyway, the visual change is made
            });
    }
}

function formatDate(dateStr) {
    return new Date(dateStr).toLocaleDateString();
}

// Search filtering
const searchQuery = ref('');
const filteredArticles = computed(() => {
    // If filters are active, use server-filtered articles
    let articles = activeFilters.value.length > 0 ? filteredArticlesFromServer.value : store.articles;
    
    // Apply search query filter (client-side, on top of server filter)
    if (searchQuery.value) {
        const lower = searchQuery.value.toLowerCase();
        articles = articles.filter(a => 
            (a.title && a.title.toLowerCase().includes(lower)) || 
            (a.feed_title && a.feed_title.toLowerCase().includes(lower))
        );
    }
    
    return articles;
});

// Virtual scrolling computed properties
const visibleRange = computed(() => {
    const height = itemHeight.value;
    const startIndex = Math.max(0, Math.floor(scrollTop.value / height) - BUFFER_SIZE);
    const visibleCount = Math.ceil(containerHeight.value / height) + 2 * BUFFER_SIZE;
    const endIndex = Math.min(filteredArticles.value.length, startIndex + visibleCount);
    return { startIndex, endIndex };
});

const visibleArticles = computed(() => {
    const { startIndex, endIndex } = visibleRange.value;
    return filteredArticles.value.slice(startIndex, endIndex);
});

const paddingTop = computed(() => visibleRange.value.startIndex * itemHeight.value);
const paddingBottom = computed(() => 
    Math.max(0, (filteredArticles.value.length - visibleRange.value.endIndex) * itemHeight.value)
);

// Reset scroll position when filter/feed changes
watch(() => [store.currentFilter, store.currentFeedId, store.currentCategory], () => {
    scrollTop.value = 0;
    if (listRef.value) {
        listRef.value.scrollTop = 0;
    }
});

// Reset filter state
function resetFilterState() {
    filteredArticlesFromServer.value = [];
    filterPage.value = 1;
    filterHasMore.value = true;
    filterTotal.value = 0;
}

// Fetch filtered articles from server with pagination
async function fetchFilteredArticles(filters, append = false) {
    if (filters.length === 0) {
        resetFilterState();
        return;
    }
    
    isFilterLoading.value = true;
    try {
        const page = append ? filterPage.value : 1;
        const res = await fetch('/api/articles/filter', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ 
                conditions: filters,
                page: page,
                limit: 50
            })
        });
        
        if (res.ok) {
            const data = await res.json();
            const articles = data.articles || [];
            
            if (append) {
                filteredArticlesFromServer.value = [...filteredArticlesFromServer.value, ...articles];
            } else {
                filteredArticlesFromServer.value = articles;
                filterPage.value = 1;
            }
            
            filterHasMore.value = data.has_more;
            filterTotal.value = data.total;
        } else {
            console.error('Error fetching filtered articles');
            if (!append) {
                filteredArticlesFromServer.value = [];
            }
        }
    } catch (e) {
        console.error('Error fetching filtered articles:', e);
        if (!append) {
            filteredArticlesFromServer.value = [];
        }
    } finally {
        isFilterLoading.value = false;
    }
}

// Load more filtered articles
async function loadMoreFilteredArticles() {
    if (isFilterLoading.value || !filterHasMore.value) return;
    
    filterPage.value++;
    await fetchFilteredArticles(activeFilters.value, true);
}

// Filter handlers
async function handleApplyFilters(filters) {
    activeFilters.value = filters;
    // Reset scroll position when filters change
    scrollTop.value = 0;
    if (listRef.value) {
        listRef.value.scrollTop = 0;
    }
    
    if (filters.length === 0) {
        // When clearing filters, reset to normal article list by refreshing store
        resetFilterState();
        store.page = 1;
        await store.fetchArticles(false);
    } else {
        await fetchFilteredArticles(filters, false);
    }
}

function clearAllFilters() {
    activeFilters.value = [];
    resetFilterState();
}

function onArticleContextMenu(e, article) {
    e.preventDefault();
    e.stopPropagation();
    
    // Determine context menu text based on default view mode
    const contentActionLabel = defaultViewMode.value === 'rendered' 
        ? store.i18n.t('showOriginal')
        : store.i18n.t('renderContent');
    const contentActionIcon = defaultViewMode.value === 'rendered' 
        ? 'ph-globe'
        : 'ph-article';
    
    window.dispatchEvent(new CustomEvent('open-context-menu', {
        detail: {
            x: e.clientX,
            y: e.clientY,
            items: [
                { label: article.is_read ? store.i18n.t('markAsUnread') : store.i18n.t('markAsRead'), action: 'toggleRead', icon: article.is_read ? 'ph-envelope' : 'ph-envelope-open' },
                { label: article.is_favorite ? store.i18n.t('removeFromFavorites') : store.i18n.t('addToFavorites'), action: 'toggleFavorite', icon: 'ph-star', iconWeight: article.is_favorite ? 'fill' : 'regular', iconColor: article.is_favorite ? 'text-yellow-500' : '' },
                { separator: true },
                { label: contentActionLabel, action: 'renderContent', icon: contentActionIcon },
                { label: article.is_hidden ? store.i18n.t('unhideArticle') : store.i18n.t('hideArticle'), action: 'toggleHide', icon: article.is_hidden ? 'ph-eye' : 'ph-eye-slash', danger: !article.is_hidden },
                { separator: true },
                { label: store.i18n.t('openInBrowser'), action: 'openBrowser', icon: 'ph-arrow-square-out' }
            ],
            data: article,
            callback: handleArticleAction
        }
    }));
}

async function handleArticleAction(action, article) {
    if (action === 'toggleRead') {
        const newState = !article.is_read;
        article.is_read = newState;
        try {
            await fetch(`/api/articles/read?id=${article.id}&read=${newState}`, { method: 'POST' });
            // Update unread counts after toggling read status
            store.fetchUnreadCounts();
        } catch (e) {
            console.error('Error toggling read status:', e);
            // Revert the state change on error
            article.is_read = !newState;
            window.showToast(store.i18n.t('errorSavingSettings'), 'error');
        }
    } else if (action === 'toggleFavorite') {
        const newState = !article.is_favorite;
        article.is_favorite = newState;
        try {
            await fetch(`/api/articles/favorite?id=${article.id}`, { method: 'POST' });
        } catch (e) {
            console.error('Error toggling favorite:', e);
            // Revert the state change on error
            article.is_favorite = !newState;
            window.showToast(store.i18n.t('errorSavingSettings'), 'error');
        }
    } else if (action === 'toggleHide') {
        try {
            await fetch(`/api/articles/toggle-hide?id=${article.id}`, { method: 'POST' });
            // Refresh article list to remove/show the hidden article
            store.fetchArticles();
        } catch (e) {
            console.error('Error toggling hide:', e);
            window.showToast(store.i18n.t('errorSavingSettings'), 'error');
        }
    } else if (action === 'renderContent') {
        // Determine the action based on default view mode
        const renderAction = defaultViewMode.value === 'rendered' ? 'showOriginal' : 'showContent';
        
        // Dispatch explicit action event before selecting article
        window.dispatchEvent(new CustomEvent('explicit-render-action', {
            detail: { action: renderAction }
        }));
        
        // Select the article
        store.currentArticleId = article.id;
        
        // Mark as read
        if (!article.is_read) {
            article.is_read = true;
            try {
                await fetch(`/api/articles/read?id=${article.id}&read=true`, { method: 'POST' });
                // Update unread counts after marking as read
                store.fetchUnreadCounts();
            } catch (e) {
                console.error('Error marking as read:', e);
                // Continue anyway, the visual change is made
            }
        }
        
        // Trigger the render action
        window.dispatchEvent(new CustomEvent('render-article-content', {
            detail: { action: renderAction }
        }));
    } else if (action === 'openBrowser') {
        BrowserOpenURL(article.url);
    }
}

async function refreshArticles() {
    await store.refreshFeeds();
    if (listRef.value) {
        listRef.value.scrollTop = 0;
    }
}

async function markAllAsRead() {
    await store.markAllAsRead();
    window.showToast(store.i18n.t('markedAllAsRead'), 'success');
}

</script>

<template>
    <section class="article-list flex flex-col w-full border-r border-border bg-bg-primary shrink-0 h-full">
        <div class="p-2 sm:p-4 border-b border-border bg-bg-primary">
            <div class="flex items-center justify-between mb-2 sm:mb-3">
                <h3 class="m-0 text-base sm:text-lg font-semibold">{{ store.i18n.t('articles') }}</h3>
                <div class="flex items-center gap-1 sm:gap-2">
                    <button @click="markAllAsRead" class="text-text-secondary hover:text-text-primary hover:bg-bg-tertiary p-1 sm:p-1.5 rounded transition-colors" :title="store.i18n.t('markAllRead')">
                        <PhCheckCircle :size="20" class="sm:w-6 sm:h-6" />
                    </button>
                    <div class="relative">
                        <button @click="refreshArticles" class="text-text-secondary hover:text-text-primary hover:bg-bg-tertiary p-1 sm:p-1.5 rounded transition-colors" :title="store.i18n.t('refresh')">
                            <PhArrowClockwise :size="20" class="sm:w-6 sm:h-6" :class="store.refreshProgress.isRunning ? 'animate-spin' : ''" />
                        </button>
                        <div v-if="store.refreshProgress.isRunning && store.refreshProgress.total > store.refreshProgress.current" class="absolute -top-1 -right-1 bg-accent text-white text-[9px] sm:text-[10px] font-bold rounded-full min-w-[14px] sm:min-w-[16px] h-3.5 sm:h-4 px-0.5 sm:px-1 flex items-center justify-center">
                            {{ store.refreshProgress.total - store.refreshProgress.current }}
                        </div>
                    </div>
                    <button @click="emit('toggleSidebar')" class="md:hidden text-xl sm:text-2xl p-1">
                        <PhList :size="20" class="sm:w-6 sm:h-6" />
                    </button>
                </div>
            </div>
            <div class="flex items-center gap-2">
                <div class="flex-1 flex items-center bg-bg-secondary border border-border rounded-lg px-2 sm:px-3 py-1.5 sm:py-2 focus-within:border-accent transition-colors">
                    <PhMagnifyingGlass :size="18" class="text-text-secondary sm:w-5 sm:h-5" />
                    <input type="text" v-model="searchQuery" :placeholder="store.i18n.t('search')" class="bg-transparent border-none outline-none w-full ml-1.5 sm:ml-2 text-text-primary text-xs sm:text-sm">
                </div>
                <div class="relative">
                    <button @click="showFilterModal = true" class="filter-btn p-1.5 sm:p-2 rounded-lg transition-colors" :class="activeFilters.length > 0 ? 'filter-active' : ''" :title="store.i18n.t('filter')">
                        <PhFunnel :size="18" class="sm:w-5 sm:h-5" />
                    </button>
                    <div v-if="activeFilters.length > 0" class="absolute -top-1 -right-1 bg-accent text-white text-[9px] sm:text-[10px] font-bold rounded-full min-w-[14px] sm:min-w-[16px] h-3.5 sm:h-4 px-0.5 sm:px-1 flex items-center justify-center">
                        {{ activeFilters.length }}
                    </div>
                </div>
            </div>
        </div>
        
        <div class="flex-1 overflow-y-auto" @scroll="handleScroll" ref="listRef">
            <div v-if="filteredArticles.length === 0 && !store.isLoading && !isFilterLoading" class="p-4 sm:p-5 text-center text-text-secondary text-sm sm:text-base">
                {{ store.i18n.t('noArticles') }}
            </div>
            
            <!-- Virtual scroll container -->
            <div v-else :style="{ paddingTop: paddingTop + 'px', paddingBottom: paddingBottom + 'px' }">
                <div v-for="article in visibleArticles" :key="article.id" 
                     :data-article-id="article.id"
                     :ref="el => observeArticle(el)"
                     @click="selectArticle(article)"
                     @contextmenu="onArticleContextMenu($event, article)"
                     :class="['article-card', article.is_read ? 'read' : '', article.is_favorite ? 'favorite' : '', article.is_hidden ? 'hidden' : '', store.currentArticleId === article.id ? 'active' : '']">
                    
                    <img v-if="article.image_url" :src="article.image_url" class="w-16 h-12 sm:w-20 sm:h-[60px] object-cover rounded bg-bg-tertiary shrink-0 border border-border" @error="$event.target.style.display='none'">
                    
                    <div class="flex-1 min-w-0">
                        <div class="flex items-start gap-1.5 sm:gap-2">
                            <h4 v-if="!article.translated_title || article.translated_title === article.title" class="flex-1 m-0 mb-1 sm:mb-1.5 text-sm sm:text-base font-semibold leading-snug text-text-primary">{{ article.title }}</h4>
                            <div v-else class="flex-1">
                                <h4 class="m-0 mb-0.5 sm:mb-1 text-sm sm:text-base font-semibold leading-snug text-text-primary">{{ article.translated_title }}</h4>
                                <div class="text-[10px] sm:text-xs text-text-secondary italic mb-0.5 sm:mb-1">{{ article.title }}</div>
                            </div>
                            <PhEyeSlash v-if="article.is_hidden" :size="18" class="text-text-secondary flex-shrink-0 sm:w-5 sm:h-5" :title="store.i18n.t('hideArticle')" />
                        </div>

                        <div class="flex justify-between items-center text-[10px] sm:text-xs text-text-secondary mt-1.5 sm:mt-2">
                            <span class="font-medium text-accent truncate flex-1 min-w-0 mr-2">{{ article.feed_title }}</span>
                            <div class="flex items-center gap-1 sm:gap-2 shrink-0">
                                <PhStar v-if="article.is_favorite" :size="14" class="text-yellow-500 sm:w-[18px] sm:h-[18px]" weight="fill" />
                                <span class="whitespace-nowrap">{{ formatDate(article.published_at) }}</span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
            
            <div v-if="store.isLoading || isFilterLoading" class="p-3 sm:p-4 text-center text-text-secondary">
                <PhSpinner :size="20" class="animate-spin sm:w-6 sm:h-6" />
            </div>
        </div>
        
        <!-- Filter Modal -->
        <ArticleFilterModal 
            :show="showFilterModal" 
            :currentFilters="activeFilters"
            @close="showFilterModal = false"
            @apply="handleApplyFilters"
        />
    </section>
</template>

<style scoped>
@media (min-width: 768px) {
    .article-list {
        width: var(--article-list-width, 400px);
    }
}
.article-card {
    @apply p-2 sm:p-3 border-b border-border cursor-pointer transition-colors flex gap-2 sm:gap-3 relative hover:bg-bg-tertiary;
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
.article-card.hidden {
    @apply opacity-60 bg-gray-100 dark:bg-gray-800;
}
.article-card.hidden:hover {
    @apply opacity-80;
}
.filter-btn {
    @apply text-text-secondary hover:text-text-primary hover:bg-bg-tertiary border border-border bg-bg-secondary;
}
.filter-btn.filter-active {
    @apply text-accent border-accent;
    background-color: rgba(59, 130, 246, 0.1);
}
.animate-spin {
    animation: spin 1s linear infinite;
}
@keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
}
</style>
