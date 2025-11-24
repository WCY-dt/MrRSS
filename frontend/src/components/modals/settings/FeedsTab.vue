<script setup>
import { store } from '../../../store.js';
import { ref, computed } from 'vue';
import { 
    PhHardDrives, PhUpload, PhDownload, PhBroom, PhRss, PhPlus, 
    PhTrash, PhFolder, PhPencil, PhMagnifyingGlass, PhCircleNotch
} from "@phosphor-icons/vue";

const emit = defineEmits(['import-opml', 'export-opml', 'cleanup-database', 'add-feed', 'edit-feed', 'delete-feed', 'batch-delete', 'batch-move']);

const selectedFeeds = ref([]);
const isDiscoveringAll = ref(false);
const discoveryProgress = ref({ message: '', detail: '' });

const isAllSelected = computed(() => {
    return store.feeds && store.feeds.length > 0 && selectedFeeds.value.length === store.feeds.length;
});

function toggleSelectAll(e) {
    if (!store.feeds) return;
    if (e.target.checked) {
        selectedFeeds.value = store.feeds.map(f => f.id);
    } else {
        selectedFeeds.value = [];
    }
}

function handleImportOPML(event) {
    emit('import-opml', event);
}

function handleExportOPML() {
    emit('export-opml');
}

function handleCleanupDatabase() {
    emit('cleanup-database');
}

async function handleDiscoverAll() {
    isDiscoveringAll.value = true;
    discoveryProgress.value = { message: store.i18n.t('preparingDiscovery'), detail: '' };
    
    try {
        console.log('Opening SSE connection for batch discovery');
        
        // Use EventSource for Server-Sent Events
        const eventSource = new EventSource('/api/feeds/discover-all');
        
        eventSource.onmessage = async (event) => {
            console.log('SSE message received:', event.data);
            const data = JSON.parse(event.data);
            
            if (data.type === 'progress') {
                // Parse main message and detail if the message starts with spaces (detail)
                if (data.message.startsWith('  ')) {
                    // This is a detail message
                    discoveryProgress.value.detail = data.message.trim();
                } else {
                    // This is a main message
                    discoveryProgress.value.message = data.message;
                    discoveryProgress.value.detail = '';
                }
            } else if (data.type === 'complete') {
                console.log('Batch discovery complete:', data.feeds);
                const feedCount = Array.isArray(data.feeds) ? data.feeds.length : 0;
                
                // Refresh feeds to show updated discovery status
                await store.fetchFeeds();
                
                if (feedCount > 0) {
                    window.showToast(
                        store.i18n.t('discoveryComplete') + ': ' + 
                        store.i18n.t('foundFeeds', { count: feedCount }),
                        'success'
                    );
                } else {
                    window.showToast(store.i18n.t('noFriendLinksFound'), 'info');
                }
                
                eventSource.close();
                isDiscoveringAll.value = false;
                discoveryProgress.value = { message: '', detail: '' };
            } else if (data.type === 'error') {
                console.error('Batch discovery error from SSE:', data.message);
                window.showToast(store.i18n.t('discoveryFailed') + ': ' + data.message, 'error');
                eventSource.close();
                isDiscoveringAll.value = false;
                discoveryProgress.value = { message: '', detail: '' };
            }
        };
        
        eventSource.onerror = (error) => {
            console.error('SSE connection error:', error);
            window.showToast(store.i18n.t('discoveryFailed'), 'error');
            eventSource.close();
            isDiscoveringAll.value = false;
            discoveryProgress.value = { message: '', detail: '' };
        };
        
    } catch (error) {
        console.error('Discovery error:', error);
        window.showToast(store.i18n.t('discoveryFailed'), 'error');
        isDiscoveringAll.value = false;
        discoveryProgress.value = { message: '', detail: '' };
    }
}

function handleAddFeed() {
    emit('add-feed');
}

function handleEditFeed(feed) {
    emit('edit-feed', feed);
}

function handleDeleteFeed(id) {
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

function getFavicon(url) {
    try {
        return `https://www.google.com/s2/favicons?domain=${new URL(url).hostname}`;
    } catch {
        return '';
    }
}
</script>

<template>
    <div class="space-y-4 sm:space-y-6">
        <div class="setting-group">
            <label class="font-semibold mb-2 sm:mb-3 text-text-secondary uppercase text-xs tracking-wider flex items-center gap-2">
                <PhHardDrives :size="14" class="sm:w-4 sm:h-4" />
                {{ store.i18n.t('dataManagement') }}
            </label>
            <div class="flex flex-col sm:flex-row gap-2 sm:gap-3 mb-2 sm:mb-3">
                <button @click="$refs.opmlInput.click()" class="btn-secondary flex-1 justify-center text-sm sm:text-base">
                    <PhUpload :size="18" class="sm:w-5 sm:h-5" /> {{ store.i18n.t('importOPML') }}
                </button>
                <input type="file" ref="opmlInput" class="hidden" @change="handleImportOPML">
                <button @click="handleExportOPML" class="btn-secondary flex-1 justify-center text-sm sm:text-base">
                    <PhDownload :size="18" class="sm:w-5 sm:h-5" /> {{ store.i18n.t('exportOPML') }}
                </button>
            </div>
            <div class="flex flex-col sm:flex-row gap-2 sm:gap-3 mb-2 sm:mb-3">
                <button @click="handleDiscoverAll" :disabled="isDiscoveringAll" 
                        :class="['btn-primary flex-1 justify-center text-sm sm:text-base', isDiscoveringAll && 'opacity-50 cursor-not-allowed']">
                    <PhCircleNotch v-if="isDiscoveringAll" :size="18" class="sm:w-5 sm:h-5 animate-spin" />
                    <PhMagnifyingGlass v-else :size="18" class="sm:w-5 sm:h-5" />
                    {{ isDiscoveringAll ? store.i18n.t('discoveringAllFeeds') : store.i18n.t('discoverAllFeeds') }}
                </button>
            </div>
            <p v-if="!isDiscoveringAll" class="text-xs text-text-secondary mb-2">
                {{ store.i18n.t('discoverAllFeedsDesc') }}
            </p>
            <div v-if="isDiscoveringAll" class="bg-gradient-to-r from-accent/10 to-accent/5 border-l-4 border-accent rounded-lg p-4 mb-2 space-y-2">
                <div class="flex items-center gap-2">
                    <PhCircleNotch :size="20" class="animate-spin text-accent shrink-0" />
                    <div class="flex-1 min-w-0">
                        <p class="text-sm font-semibold text-accent">
                            {{ discoveryProgress.message }}
                        </p>
                        <p v-if="discoveryProgress.detail" class="text-xs text-text-secondary mt-1">
                            {{ discoveryProgress.detail }}
                        </p>
                    </div>
                </div>
                <div class="flex items-center gap-2 text-xs text-text-tertiary pt-2 border-t border-accent/20">
                    <PhRss :size="12" />
                    <span>{{ store.i18n.t('pleaseWait') }}</span>
                </div>
            </div>
            <div class="flex">
                <button @click="handleCleanupDatabase" class="btn-danger flex-1 justify-center text-sm sm:text-base">
                    <PhBroom :size="18" class="sm:w-5 sm:h-5" /> {{ store.i18n.t('cleanDatabase') }}
                </button>
            </div>
            <p class="text-xs text-text-secondary mt-2">
                {{ store.i18n.t('cleanDatabaseDesc') }}
            </p>
        </div>
        
        <div class="setting-group">
            <label class="font-semibold mb-2 sm:mb-3 text-text-secondary uppercase text-xs tracking-wider flex items-center gap-2">
                <PhRss :size="14" class="sm:w-4 sm:h-4" />
                {{ store.i18n.t('manageFeeds') }}
            </label>
            
            <div class="flex flex-wrap gap-1.5 sm:gap-2 mb-2 text-xs sm:text-sm">
                <button @click="handleAddFeed" class="btn-secondary py-1.5 px-2.5 sm:px-3">
                    <PhPlus :size="14" class="sm:w-4 sm:h-4" /> <span class="hidden sm:inline">{{ store.i18n.t('addFeed') }}</span><span class="sm:hidden">{{ store.i18n.t('addFeed').split(' ')[0] }}</span>
                </button>
                <button @click="handleBatchDelete" class="btn-danger py-1.5 px-2.5 sm:px-3" :disabled="selectedFeeds.length === 0">
                    <PhTrash :size="14" class="sm:w-4 sm:h-4" /> <span class="hidden sm:inline">{{ store.i18n.t('deleteSelected') }}</span><span class="sm:hidden">{{ store.i18n.t('delete') }}</span>
                </button>
                <button @click="handleBatchMove" class="btn-secondary py-1.5 px-2.5 sm:px-3" :disabled="selectedFeeds.length === 0">
                    <PhFolder :size="14" class="sm:w-4 sm:h-4" /> <span class="hidden sm:inline">{{ store.i18n.t('moveSelected') }}</span><span class="sm:hidden">{{ store.i18n.t('move') }}</span>
                </button>
                <div class="flex-1 min-w-0"></div>
                <label class="flex items-center gap-1.5 sm:gap-2 cursor-pointer select-none whitespace-nowrap">
                    <input type="checkbox" :checked="isAllSelected" @change="toggleSelectAll" class="w-3.5 h-3.5 sm:w-4 sm:h-4 rounded border-border text-accent focus:ring-2 focus:ring-accent cursor-pointer">
                    <span class="hidden sm:inline">{{ store.i18n.t('selectAll') }}</span>
                </label>
            </div>

            <div class="border border-border rounded-lg bg-bg-secondary overflow-y-auto max-h-60 sm:max-h-96">
                <div v-for="feed in store.feeds" :key="feed.id" class="flex items-center p-1.5 sm:p-2 border-b border-border last:border-0 bg-bg-primary hover:bg-bg-secondary gap-1.5 sm:gap-2">
                    <input type="checkbox" :value="feed.id" v-model="selectedFeeds" class="w-3.5 h-3.5 sm:w-4 sm:h-4 shrink-0 rounded border-border text-accent focus:ring-2 focus:ring-accent cursor-pointer">
                    <div class="w-4 h-4 flex items-center justify-center shrink-0">
                        <img :src="feed.image_url || getFavicon(feed.url)" class="w-full h-full object-contain" @error="$event.target.style.display='none'">
                    </div>
                    <div class="truncate flex-1 min-w-0">
                        <div class="font-medium truncate text-xs sm:text-sm">{{ feed.title }}</div>
                        <div class="text-xs text-text-secondary truncate hidden sm:block">
                            <span v-if="feed.category" class="inline-flex items-center gap-1">
                                <PhFolder :size="10" class="inline" />
                                {{ feed.category }}
                                <span class="mx-1">•</span>
                            </span>
                            <span>{{ feed.url }}</span>
                        </div>
                    </div>
                    <div class="flex gap-0.5 sm:gap-1 shrink-0">
                        <button @click="handleEditFeed(feed)" class="text-accent hover:bg-bg-tertiary p-1 rounded text-sm" :title="store.i18n.t('edit')"><PhPencil :size="14" class="sm:w-4 sm:h-4" /></button>
                        <button @click="handleDeleteFeed(feed.id)" class="text-red-500 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 p-1 rounded text-sm" :title="store.i18n.t('delete')"><PhTrash :size="14" class="sm:w-4 sm:h-4" /></button>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
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
