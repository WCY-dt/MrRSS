<script setup>
import { store } from '../../store.js';
import { ref, onMounted } from 'vue';
import GeneralTab from './settings/GeneralTab.vue';
import FeedsTab from './settings/FeedsTab.vue';
import ShortcutsTab from './settings/ShortcutsTab.vue';
import AboutTab from './settings/AboutTab.vue';
import DiscoverAllFeedsModal from './DiscoverAllFeedsModal.vue';
import { PhGear } from "@phosphor-icons/vue";

const emit = defineEmits(['close']);
const activeTab = ref('general');
const showDiscoverAllModal = ref(false);

const settings = ref({
    update_interval: 10,
    translation_enabled: false,
    target_language: 'zh',
    translation_provider: 'google',
    deepl_api_key: '',
    auto_cleanup_enabled: false,
    max_cache_size_mb: 20,
    max_article_age_days: 30,
    language: store.i18n.locale.value,
    theme: 'auto',
    last_article_update: '',
    show_hidden_articles: false,
    default_view_mode: 'original',
    startup_on_boot: false,
    shortcuts: ''
});

const updateInfo = ref(null);
const checkingUpdates = ref(false);
const downloadingUpdate = ref(false);
const installingUpdate = ref(false);
const downloadProgress = ref(0);

onMounted(async () => {
    // Fetch settings
    try {
        const res = await fetch('/api/settings');
        const data = await res.json();
        settings.value = {
            update_interval: data.update_interval || 10,
            translation_enabled: data.translation_enabled === 'true',
            target_language: data.target_language || 'zh',
            translation_provider: data.translation_provider || 'google',
            deepl_api_key: data.deepl_api_key || '',
            auto_cleanup_enabled: data.auto_cleanup_enabled === 'true',
            max_cache_size_mb: parseInt(data.max_cache_size_mb) || 20,
            max_article_age_days: parseInt(data.max_article_age_days) || 30,
            language: data.language || store.i18n.locale.value,
            theme: data.theme || 'auto',
            last_article_update: data.last_article_update || '',
            show_hidden_articles: data.show_hidden_articles === 'true',
            default_view_mode: data.default_view_mode || 'original',
            startup_on_boot: data.startup_on_boot === 'true',
            shortcuts: data.shortcuts || ''
        };
        // Apply the saved language
        if (data.language) {
            store.i18n.setLocale(data.language);
        }
        // Apply the saved theme
        if (data.theme) {
            store.setTheme(data.theme);
        }
        // Initialize shortcuts in store
        if (data.shortcuts) {
            try {
                const parsed = JSON.parse(data.shortcuts);
                window.dispatchEvent(new CustomEvent('shortcuts-changed', {
                    detail: { shortcuts: parsed }
                }));
            } catch (e) {
                console.error('Error parsing shortcuts:', e);
            }
        }
    } catch (e) {
        console.error(e);
    }
});

// Feeds tab event handlers
function handleImportOPML(event) {
    const file = event.target.files[0];
    if (!file) return;
    
    const reader = new FileReader();
    reader.onload = (e) => {
        const content = e.target.result;
        fetch('/api/opml/import', {
            method: 'POST',
            headers: {
                'Content-Type': 'text/xml'
            },
            body: content
        }).then(async res => {
            if (res.ok) {
                window.showToast(store.i18n.t('opmlImportedSuccess'), 'success');
                store.fetchFeeds();
            } else {
                const text = await res.text();
                window.showToast(store.i18n.t('importFailed', { error: text }), 'error');
            }
        });
    };
    reader.readAsText(file);
}

function handleExportOPML() {
    window.location.href = '/api/opml/export';
}

async function handleCleanupDatabase() {
    const confirmed = await window.showConfirm({
        title: store.i18n.t('cleanDatabaseTitle'),
        message: store.i18n.t('cleanDatabaseMessage'),
        confirmText: store.i18n.t('clean'),
        cancelText: store.i18n.t('cancel'),
        isDanger: true
    });
    if (!confirmed) return;
    
    try {
        const res = await fetch('/api/articles/cleanup', { method: 'POST' });
        if (res.ok) {
            const result = await res.json();
            window.showToast(store.i18n.t('databaseCleanedSuccess', { count: result.deleted }), 'success');
            store.fetchArticles();
        } else {
            window.showToast(store.i18n.t('errorCleaningDatabase'), 'error');
        }
    } catch (e) {
        console.error(e);
        window.showToast(store.i18n.t('errorCleaningDatabase'), 'error');
    }
}

function handleAddFeed() {
    window.dispatchEvent(new CustomEvent('show-add-feed'));
}

function handleEditFeed(feed) {
    window.dispatchEvent(new CustomEvent('show-edit-feed', { detail: feed }));
}

async function handleDeleteFeed(id) {
    const confirmed = await window.showConfirm({
        title: store.i18n.t('deleteFeedTitle'),
        message: store.i18n.t('deleteFeedMessage'),
        confirmText: store.i18n.t('delete'),
        cancelText: store.i18n.t('cancel'),
        isDanger: true
    });
    if (!confirmed) return;
    
    await fetch(`/api/feeds/delete?id=${id}`, { method: 'POST' });
    store.fetchFeeds();
    window.showToast(store.i18n.t('feedDeletedSuccess'), 'success');
}

async function handleBatchDelete(selectedIds) {
    const confirmed = await window.showConfirm({
        title: store.i18n.t('deleteMultipleFeedsTitle'),
        message: store.i18n.t('deleteMultipleFeedsMessage', { count: selectedIds.length }),
        confirmText: store.i18n.t('delete'),
        cancelText: store.i18n.t('cancel'),
        isDanger: true
    });
    if (!confirmed) return;

    const promises = selectedIds.map(id => fetch(`/api/feeds/delete?id=${id}`, { method: 'POST' }));
    await Promise.all(promises);
    store.fetchFeeds();
    window.showToast(store.i18n.t('feedsDeletedSuccess'), 'success');
}

async function handleBatchMove(selectedIds) {
    if (!store.feeds) return;
    
    const newCategory = await window.showInput({
        title: store.i18n.t('moveFeeds'),
        message: store.i18n.t('enterCategoryName'),
        placeholder: store.i18n.t('categoryPlaceholder'),
        confirmText: store.i18n.t('move'),
        cancelText: store.i18n.t('cancel')
    });
    if (newCategory === null) return;

    const promises = selectedIds.map(id => {
        const feed = store.feeds.find(f => f.id === id);
        if (!feed) return Promise.resolve();
        return fetch('/api/feeds/update', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ 
                id: feed.id, 
                title: feed.title, 
                url: feed.url, 
                category: newCategory 
            })
        });
    });

    await Promise.all(promises);
    store.fetchFeeds();
    window.showToast(store.i18n.t('feedsMovedSuccess'), 'success');
}

function handleDiscoverAll() {
    showDiscoverAllModal.value = true;
}

// About tab event handlers
async function handleCheckUpdates() {
    checkingUpdates.value = true;
    updateInfo.value = null;
    
    try {
        const res = await fetch('/api/check-updates');
        if (res.ok) {
            const data = await res.json();
            updateInfo.value = data;
            
            if (data.error) {
                window.showToast(store.i18n.t('errorCheckingUpdates'), 'error');
            } else if (data.has_update) {
                window.showToast(store.i18n.t('updateAvailable'), 'info');
            } else {
                window.showToast(store.i18n.t('upToDate'), 'success');
            }
        } else {
            window.showToast(store.i18n.t('errorCheckingUpdates'), 'error');
        }
    } catch (e) {
        console.error(e);
        window.showToast(store.i18n.t('errorCheckingUpdates'), 'error');
    } finally {
        checkingUpdates.value = false;
    }
}

async function handleDownloadInstallUpdate() {
    if (!updateInfo.value || !updateInfo.value.download_url) {
        window.showToast(store.i18n.t('errorCheckingUpdates'), 'error');
        return;
    }

    downloadingUpdate.value = true;
    downloadProgress.value = 0;

    // Simulate progress while downloading
    const progressInterval = setInterval(() => {
        if (downloadProgress.value < 90) {
            downloadProgress.value += 10;
        }
    }, 500);

    try {
        // Download the update
        const downloadRes = await fetch('/api/download-update', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                download_url: updateInfo.value.download_url,
                asset_name: updateInfo.value.asset_name
            })
        });

        clearInterval(progressInterval);

        if (!downloadRes.ok) {
            const errorText = await downloadRes.text();
            console.error('Download error:', errorText);
            throw new Error('DOWNLOAD_ERROR: ' + errorText);
        }

        const downloadData = await downloadRes.json();
        if (!downloadData.success || !downloadData.file_path) {
            throw new Error('DOWNLOAD_ERROR: Invalid response from server');
        }
        
        downloadingUpdate.value = false;
        downloadProgress.value = 100;

        // Show notification
        window.showToast(store.i18n.t('downloadComplete'), 'success');

        // Wait a moment to ensure file is fully written
        await new Promise(resolve => setTimeout(resolve, 500));

        // Install the update
        installingUpdate.value = true;
        window.showToast(store.i18n.t('installingUpdate'), 'info');

        const installRes = await fetch('/api/install-update', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                file_path: downloadData.file_path
            })
        });

        if (!installRes.ok) {
            const errorText = await installRes.text();
            console.error('Install error:', errorText);
            throw new Error('INSTALL_ERROR: ' + errorText);
        }

        const installData = await installRes.json();
        if (!installData.success) {
            throw new Error('INSTALL_ERROR: Installation failed');
        }

        // Show final message - app will close automatically from backend
        window.showToast(store.i18n.t('updateWillRestart'), 'info');

    } catch (e) {
        console.error('Update error:', e);
        clearInterval(progressInterval);
        downloadingUpdate.value = false;
        installingUpdate.value = false;
        
        // Use error codes for more reliable error classification
        const errorMessage = e.message || '';
        if (errorMessage.includes('DOWNLOAD_ERROR')) {
            window.showToast(store.i18n.t('downloadFailed'), 'error');
        } else if (errorMessage.includes('INSTALL_ERROR')) {
            window.showToast(store.i18n.t('installFailed'), 'error');
        } else {
            window.showToast(store.i18n.t('errorCheckingUpdates'), 'error');
        }
    }
}

</script>

<template>
    <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-2 sm:p-4">
        <div class="bg-bg-primary w-full max-w-4xl h-full sm:h-[700px] sm:max-h-[90vh] flex flex-col rounded-none sm:rounded-2xl shadow-2xl border border-border overflow-hidden animate-fade-in">
            <div class="p-3 sm:p-5 border-b border-border flex justify-between items-center shrink-0">
                <h3 class="text-base sm:text-lg font-semibold m-0 flex items-center gap-2">
                    <PhGear :size="20" class="sm:w-6 sm:h-6" />
                    {{ store.i18n.t('settingsTitle') }}
                </h3>
                <span @click="emit('close')" class="text-2xl cursor-pointer text-text-secondary hover:text-text-primary">&times;</span>
            </div>
            
            <div class="flex border-b border-border bg-bg-secondary shrink-0 overflow-x-auto scrollbar-hide">
                <button @click="activeTab = 'general'" :class="['tab-btn', activeTab === 'general' ? 'active' : '']">{{ store.i18n.t('general') }}</button>
                <button @click="activeTab = 'feeds'" :class="['tab-btn', activeTab === 'feeds' ? 'active' : '']">{{ store.i18n.t('feeds') }}</button>
                <button @click="activeTab = 'shortcuts'" :class="['tab-btn', activeTab === 'shortcuts' ? 'active' : '']">{{ store.i18n.t('shortcuts') }}</button>
                <button @click="activeTab = 'about'" :class="['tab-btn', activeTab === 'about' ? 'active' : '']">{{ store.i18n.t('about') }}</button>
            </div>

            <div class="flex-1 overflow-y-auto p-3 sm:p-6 min-h-0">
                <GeneralTab v-if="activeTab === 'general'" :settings="settings" />
                
                <FeedsTab 
                    v-if="activeTab === 'feeds'"
                    @import-opml="handleImportOPML"
                    @export-opml="handleExportOPML"
                    @cleanup-database="handleCleanupDatabase"
                    @add-feed="handleAddFeed"
                    @edit-feed="handleEditFeed"
                    @delete-feed="handleDeleteFeed"
                    @batch-delete="handleBatchDelete"
                    @batch-move="handleBatchMove"
                    @discover-all="handleDiscoverAll"
                />
                
                <ShortcutsTab v-if="activeTab === 'shortcuts'" :settings="settings" />
                
                <AboutTab 
                    v-if="activeTab === 'about'"
                    :update-info="updateInfo"
                    :checking-updates="checkingUpdates"
                    :downloading-update="downloadingUpdate"
                    :installing-update="installingUpdate"
                    :download-progress="downloadProgress"
                    @check-updates="handleCheckUpdates"
                    @download-install-update="handleDownloadInstallUpdate"
                />
            </div>
        </div>
        
        <!-- Discover All Feeds Modal -->
        <DiscoverAllFeedsModal 
            :show="showDiscoverAllModal"
            @close="showDiscoverAllModal = false"
        />
    </div>
</template>

<style scoped>
.tab-btn {
    @apply px-3 sm:px-5 py-2 sm:py-3 bg-transparent border-b-2 border-transparent text-text-secondary font-semibold cursor-pointer hover:text-text-primary transition-all relative whitespace-nowrap text-sm sm:text-base;
}
.tab-btn:hover {
    background-color: rgba(128, 128, 128, 0.1);
}
.tab-btn.active {
    @apply text-accent border-accent;
    background-color: rgba(128, 128, 128, 0.05);
}
.tab-btn.active::after {
    content: '';
    position: absolute;
    bottom: -2px;
    left: 0;
    right: 0;
    height: 2px;
    background: linear-gradient(90deg, transparent, var(--accent-color), transparent);
    animation: shimmer 2s ease-in-out infinite;
}
@keyframes shimmer {
    0%, 100% { opacity: 0.5; }
    50% { opacity: 1; }
}
.btn-primary {
    @apply bg-accent text-white border-none px-5 py-2.5 rounded-lg cursor-pointer font-semibold hover:bg-accent-hover transition-colors;
}
.animate-fade-in {
    animation: modalFadeIn 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}
@keyframes modalFadeIn {
    from { transform: translateY(-20px); opacity: 0; }
    to { transform: translateY(0); opacity: 1; }
}
.scrollbar-hide {
    -ms-overflow-style: none;
    scrollbar-width: none;
}
.scrollbar-hide::-webkit-scrollbar {
    display: none;
}
</style>
