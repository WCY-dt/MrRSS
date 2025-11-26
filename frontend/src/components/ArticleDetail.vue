<script setup>
import { store } from '../store.js';
import { computed, ref, onMounted, onBeforeUnmount, watch, nextTick } from 'vue';
import { BrowserOpenURL } from '../wailsjs/wailsjs/runtime/runtime.js';
import { 
    PhNewspaper, PhArrowLeft, PhGlobe, PhArticle, PhEnvelopeOpen, 
    PhEnvelope, PhStar, PhArrowSquareOut 
} from "@phosphor-icons/vue";
import ImageViewer from './ImageViewer.vue';

const article = computed(() => store.articles.find(a => a.id === store.currentArticleId));
const showContent = ref(false); // Toggle between original webpage and RSS content
const articleContent = ref(''); // Dynamically fetched content
const isLoadingContent = ref(false); // Loading state
const currentArticleId = ref(null); // Track which article content we've loaded
const defaultViewMode = ref('original'); // Default view mode from settings
const pendingRenderAction = ref(null); // Track if there's a pending render action from context menu
const imageViewerSrc = ref(null); // Image source for viewer
const imageViewerAlt = ref(''); // Image alt text for viewer

// Watch for article changes and apply default view mode
watch(() => store.currentArticleId, async (newId, oldId) => {
    if (newId && newId !== oldId) {
        // Reset content when switching articles
        articleContent.value = '';
        currentArticleId.value = null;
        
        // Check if there's a pending render action from context menu
        if (pendingRenderAction.value) {
            // Apply the explicit action instead of default
            if (pendingRenderAction.value === 'showContent') {
                showContent.value = true;
                await fetchArticleContent();
            } else if (pendingRenderAction.value === 'showOriginal') {
                showContent.value = false;
            }
            pendingRenderAction.value = null; // Clear the pending action
        } else {
            // Apply default view mode
            if (defaultViewMode.value === 'rendered') {
                showContent.value = true;
                await fetchArticleContent();
            } else {
                showContent.value = false;
            }
        }
    }
});

// Listen for default view mode changes from settings
window.addEventListener('default-view-mode-changed', (e) => {
    defaultViewMode.value = e.detail.mode;
});

function close() {
    store.currentArticleId = null;
    showContent.value = false;
    articleContent.value = '';
    currentArticleId.value = null;
}

function toggleRead() {
    if (!article.value) return;
    const newState = !article.value.is_read;
    article.value.is_read = newState;
    fetch(`/api/articles/read?id=${article.value.id}&read=${newState}`, { method: 'POST' });
}

function toggleFavorite() {
    if (!article.value) return;
    const newState = !article.value.is_favorite;
    article.value.is_favorite = newState;
    fetch(`/api/articles/favorite?id=${article.value.id}`, { method: 'POST' });
}

function openOriginal() {
    if (article.value) BrowserOpenURL(article.value.url);
}

async function toggleContentView() {
    if (!showContent.value) {
        // Switching to content view - fetch content if needed
        if (!article.value) return;
        // Check if we need to fetch content (different article or no content yet)
        if (currentArticleId.value !== article.value.id) {
            await fetchArticleContent();
        }
    }
    showContent.value = !showContent.value;
}

async function fetchArticleContent() {
    if (!article.value) return;
    
    isLoadingContent.value = true;
    currentArticleId.value = article.value.id; // Track which article we're loading
    try {
        const res = await fetch(`/api/articles/content?id=${article.value.id}`);
        if (res.ok) {
            const data = await res.json();
            articleContent.value = data.content || '';
            // Wait for DOM to update, then attach event listeners
            await nextTick();
            attachContentEventListeners();
        } else {
            console.error('Failed to fetch article content');
            articleContent.value = '';
        }
    } catch (e) {
        console.error('Error fetching article content:', e);
        articleContent.value = '';
    } finally {
        isLoadingContent.value = false;
    }
}

// Attach event listeners to links and images in rendered content
function attachContentEventListeners() {
    // Handle all links - open in default browser
    const links = document.querySelectorAll('.prose a');
    links.forEach(link => {
        link.addEventListener('click', (e) => {
            e.preventDefault();
            const href = link.getAttribute('href');
            if (href) {
                BrowserOpenURL(href);
            }
        });
    });
    
    // Handle all images - make them clickable for zoom/pan
    const images = document.querySelectorAll('.prose img');
    images.forEach(img => {
        img.style.cursor = 'pointer';
        img.addEventListener('click', (e) => {
            e.preventDefault();
            imageViewerSrc.value = img.src;
            imageViewerAlt.value = img.alt || '';
        });
    });
}

function closeImageViewer() {
    imageViewerSrc.value = null;
    imageViewerAlt.value = '';
}

// Listen for render content event from context menu
async function handleRenderContent(e) {
    if (!article.value) return;
    
    const action = e.detail?.action || 'showContent';
    
    // Mark as read when rendering content
    if (!article.value.is_read) {
        article.value.is_read = true;
        fetch(`/api/articles/read?id=${article.value.id}&read=true`, { method: 'POST' });
    }
    
    if (action === 'showContent') {
        // Check if we need to fetch content for this article
        if (currentArticleId.value !== article.value.id) {
            await fetchArticleContent();
        }
        showContent.value = true;
    } else if (action === 'showOriginal') {
        showContent.value = false;
    }
}

// Listen for explicit render action from context menu (before article selection)
function handleExplicitRenderAction(e) {
    pendingRenderAction.value = e.detail?.action;
}

// Handle toggle content view from keyboard shortcut
function handleToggleContentView() {
    if (article.value) {
        toggleContentView();
    }
}

onMounted(async () => {
    window.addEventListener('render-article-content', handleRenderContent);
    window.addEventListener('explicit-render-action', handleExplicitRenderAction);
    window.addEventListener('toggle-content-view', handleToggleContentView);
    
    // Load default view mode from settings
    try {
        const res = await fetch('/api/settings');
        const data = await res.json();
        defaultViewMode.value = data.default_view_mode || 'original';
    } catch (e) {
        console.error('Error loading settings:', e);
    }
});

onBeforeUnmount(() => {
    window.removeEventListener('render-article-content', handleRenderContent);
    window.removeEventListener('explicit-render-action', handleExplicitRenderAction);
    window.removeEventListener('toggle-content-view', handleToggleContentView);
});
</script>

<template>
    <main :class="['flex-1 bg-bg-primary flex flex-col h-full absolute w-full md:static md:w-auto z-30 transition-transform duration-300', article ? 'translate-x-0' : 'translate-x-full md:translate-x-0']">
        <div v-if="!article" class="hidden md:flex flex-col items-center justify-center h-full text-text-secondary text-center px-4">
            <PhNewspaper :size="48" class="mb-4 sm:mb-5 opacity-50 sm:w-16 sm:h-16" />
            <p class="text-sm sm:text-base">{{ store.i18n.t('selectArticle') }}</p>
        </div>

        <div v-else class="flex flex-col h-full bg-bg-primary">
            <div class="h-[44px] sm:h-[50px] px-3 sm:px-5 border-b border-border flex justify-between items-center bg-bg-primary shrink-0">
                <button @click="close" class="md:hidden flex items-center gap-1.5 sm:gap-2 text-text-secondary hover:text-text-primary text-sm sm:text-base">
                    <PhArrowLeft :size="18" class="sm:w-5 sm:h-5" /> <span class="hidden xs:inline">{{ store.i18n.t('back') }}</span>
                </button>
                <div class="flex gap-1 sm:gap-2 ml-auto">
                    <button @click="toggleContentView" class="action-btn" :title="showContent ? store.i18n.t('viewOriginal') : store.i18n.t('viewContent')">
                        <PhGlobe v-if="showContent" :size="18" class="sm:w-5 sm:h-5" />
                        <PhArticle v-else :size="18" class="sm:w-5 sm:h-5" />
                    </button>
                    <button @click="toggleRead" class="action-btn" :title="article.is_read ? store.i18n.t('markAsUnread') : store.i18n.t('markAsRead')">
                        <PhEnvelopeOpen v-if="article.is_read" :size="18" class="sm:w-5 sm:h-5" />
                        <PhEnvelope v-else :size="18" class="sm:w-5 sm:h-5" />
                    </button>
                    <button @click="toggleFavorite" :class="['action-btn', article.is_favorite ? 'text-yellow-500 hover:text-yellow-600' : 'hover:text-yellow-500']" :title="article.is_favorite ? store.i18n.t('removeFromFavorite') : store.i18n.t('addToFavorite')">
                        <PhStar :size="18" class="sm:w-5 sm:h-5" :weight="article.is_favorite ? 'fill' : 'regular'" />
                    </button>
                    <button @click="openOriginal" class="action-btn" :title="store.i18n.t('openInBrowser')">
                        <PhArrowSquareOut :size="18" class="sm:w-5 sm:h-5" />
                    </button>
                </div>
            </div>
            
            <!-- Original webpage view -->
            <div v-if="!showContent" class="flex-1 bg-white w-full">
                <iframe :src="article.url" class="w-full h-full border-none" sandbox="allow-scripts allow-same-origin allow-popups"></iframe>
            </div>
            
            <!-- RSS content view -->
            <div v-else class="flex-1 overflow-y-auto bg-bg-primary p-3 sm:p-6">
                <div class="max-w-3xl mx-auto bg-bg-primary">
                    <h1 class="text-xl sm:text-3xl font-bold mb-3 sm:mb-4 text-text-primary leading-tight">{{ article.title }}</h1>
                    <div class="text-xs sm:text-sm text-text-secondary mb-4 sm:mb-6 flex flex-wrap items-center gap-2 sm:gap-4">
                        <span>{{ article.feed_title }}</span>
                        <span class="hidden sm:inline">•</span>
                        <span>{{ new Date(article.published_at).toLocaleDateString() }}</span>
                    </div>
                    
                    <!-- Loading state with proper background -->
                    <div v-if="isLoadingContent" class="flex flex-col items-center justify-center py-12 sm:py-16 bg-bg-primary">
                        <div class="relative mb-4 sm:mb-6">
                            <!-- Outer pulsing ring -->
                            <div class="absolute inset-0 rounded-full border-2 sm:border-4 border-accent animate-ping opacity-20"></div>
                            <!-- Middle spinning ring -->
                            <div class="absolute inset-0 rounded-full border-2 sm:border-4 border-t-accent border-r-transparent border-b-transparent border-l-transparent animate-spin"></div>
                            <!-- Inner icon -->
                            <div class="relative bg-bg-secondary rounded-full p-4 sm:p-6">
                                <PhArticle :size="48" class="text-accent sm:w-16 sm:h-16" />
                            </div>
                        </div>
                        <p class="text-base sm:text-lg font-medium text-text-primary mb-1 sm:mb-2">
                            {{ store.i18n.t('loadingContent') }}
                        </p>
                        <p class="text-xs sm:text-sm text-text-secondary px-4 text-center">
                            {{ store.i18n.t('fetchingArticleContent') }}
                        </p>
                    </div>
                    
                    <!-- Content display -->
                    <div v-else-if="articleContent" class="prose prose-sm sm:prose-lg max-w-none text-text-primary" v-html="articleContent"></div>
                    
                    <!-- No content available -->
                    <div v-else class="text-center text-text-secondary py-6 sm:py-8">
                        <PhArticle :size="48" class="mb-2 sm:mb-3 opacity-50 mx-auto sm:w-16 sm:h-16" />
                        <p class="text-sm sm:text-base">{{ store.i18n.t('noContent') }}</p>
                    </div>
                </div>
            </div>
        </div>
        
        <!-- Image Viewer Modal -->
        <ImageViewer v-if="imageViewerSrc" :src="imageViewerSrc" :alt="imageViewerAlt" @close="closeImageViewer" />
    </main>
</template>

<style scoped>
.action-btn {
    @apply text-lg sm:text-xl cursor-pointer text-text-secondary p-1 sm:p-1.5 rounded-md transition-colors hover:bg-bg-tertiary hover:text-text-primary;
}

/* Prose styling for article content */
.prose {
    color: var(--text-primary);
}
.prose :deep(h1), .prose :deep(h2), .prose :deep(h3), .prose :deep(h4), .prose :deep(h5), .prose :deep(h6) {
    color: var(--text-primary);
    font-weight: 600;
    margin-top: 1.5em;
    margin-bottom: 0.75em;
}
.prose :deep(p) {
    margin-bottom: 1em;
    line-height: 1.7;
}
.prose :deep(a) {
    color: var(--accent-color);
    text-decoration: underline;
    cursor: pointer;
}
.prose :deep(img) {
    max-width: 100%;
    height: auto;
    border-radius: 0.5rem;
    margin: 1.5em 0;
    cursor: pointer;
    transition: opacity 0.2s;
}
.prose :deep(img:hover) {
    opacity: 0.9;
}
.prose :deep(pre) {
    background-color: var(--bg-secondary);
    padding: 1em;
    border-radius: 0.5rem;
    overflow-x: auto;
    margin: 1em 0;
}
.prose :deep(code) {
    background-color: var(--bg-secondary);
    padding: 0.2em 0.4em;
    border-radius: 0.25rem;
    font-size: 0.9em;
}
.prose :deep(blockquote) {
    border-left: 4px solid var(--accent-color);
    padding-left: 1em;
    margin: 1em 0;
    font-style: italic;
    color: var(--text-secondary);
}
.prose :deep(ul), .prose :deep(ol) {
    margin: 1em 0;
    padding-left: 2em;
}
.prose :deep(li) {
    margin-bottom: 0.5em;
}
</style>
