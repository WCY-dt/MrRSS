<script setup>
import { store } from '../../store.js';
import { ref, onMounted } from 'vue';
import GeneralTab from './settings/GeneralTab.vue';
import FeedsTab from './settings/FeedsTab.vue';
import AboutTab from './settings/AboutTab.vue';

const emit = defineEmits(['close']);
const activeTab = ref('general');

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
    theme: 'auto'
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
            theme: data.theme || 'auto'
        };
        // Apply the saved language
        if (data.language) {
            store.i18n.setLocale(data.language);
        }
        // Apply the saved theme
        if (data.theme) {
            store.setTheme(data.theme);
        }
    } catch (e) {
        console.error(e);
    }
});

async function saveSettings() {
    try {
        await fetch('/api/settings', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                update_interval: settings.value.update_interval.toString(),
                translation_enabled: settings.value.translation_enabled.toString(),
                target_language: settings.value.target_language,
                translation_provider: settings.value.translation_provider,
                deepl_api_key: settings.value.deepl_api_key,
                auto_cleanup_enabled: settings.value.auto_cleanup_enabled.toString(),
                max_cache_size_mb: settings.value.max_cache_size_mb.toString(),
                max_article_age_days: settings.value.max_article_age_days.toString(),
                language: settings.value.language,
                theme: settings.value.theme
            })
        });
        store.i18n.setLocale(settings.value.language);
        store.setTheme(settings.value.theme);
        store.startAutoRefresh(settings.value.update_interval);
        emit('close');
    } catch (e) {
        window.showToast(store.i18n.t('errorSavingSettings'), 'error');
    }
}

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
                window.showToast('OPML Imported. Feeds will appear shortly.', 'success');
                store.fetchFeeds();
            } else {
                const text = await res.text();
                window.showToast('Import failed: ' + text, 'error');
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
        title: 'Clean Database',
        message: 'This will delete all articles except read and favorited ones. Continue?',
        confirmText: 'Clean',
        cancelText: 'Cancel',
        isDanger: true
    });
    if (!confirmed) return;
    
    try {
        const res = await fetch('/api/articles/cleanup', { method: 'POST' });
        if (res.ok) {
            const result = await res.json();
            window.showToast(`Database cleaned up successfully. ${result.deleted} articles deleted.`, 'success');
            store.fetchArticles();
        } else {
            window.showToast('Error cleaning up database', 'error');
        }
    } catch (e) {
        console.error(e);
        window.showToast('Error cleaning up database', 'error');
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
        title: 'Delete Feed',
        message: 'Are you sure you want to delete this feed?',
        confirmText: 'Delete',
        cancelText: 'Cancel',
        isDanger: true
    });
    if (!confirmed) return;
    
    await fetch(`/api/feeds/delete?id=${id}`, { method: 'POST' });
    store.fetchFeeds();
    window.showToast('Feed deleted successfully', 'success');
}

async function handleBatchDelete(selectedIds) {
    const confirmed = await window.showConfirm({
        title: 'Delete Multiple Feeds',
        message: `Are you sure you want to delete ${selectedIds.length} feeds?`,
        confirmText: 'Delete',
        cancelText: 'Cancel',
        isDanger: true
    });
    if (!confirmed) return;

    const promises = selectedIds.map(id => fetch(`/api/feeds/delete?id=${id}`, { method: 'POST' }));
    await Promise.all(promises);
    store.fetchFeeds();
    window.showToast('Feeds deleted successfully', 'success');
}

async function handleBatchMove(selectedIds) {
    if (!store.feeds) return;
    
    // TODO: Replace with custom input dialog
    const newCategory = prompt('Enter new category name:');
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
    window.showToast('Feeds moved successfully', 'success');
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
            throw new Error('DOWNLOAD_ERROR');
        }

        const downloadData = await downloadRes.json();
        if (!downloadData.success || !downloadData.file_path) {
            throw new Error('DOWNLOAD_ERROR');
        }
        
        downloadingUpdate.value = false;
        downloadProgress.value = 100;

        // Show notification
        window.showToast(store.i18n.t('downloadComplete'), 'success');

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
            throw new Error('INSTALL_ERROR');
        }

        const installData = await installRes.json();
        if (!installData.success) {
            throw new Error('INSTALL_ERROR');
        }

        // Show final message - app will close automatically from backend
        window.showToast(store.i18n.t('updateWillRestart'), 'info');

    } catch (e) {
        console.error('Update error:', e);
        clearInterval(progressInterval);
        downloadingUpdate.value = false;
        installingUpdate.value = false;
        
        // Use error codes for more reliable error classification
        if (e.message === 'DOWNLOAD_ERROR') {
            window.showToast(store.i18n.t('downloadFailed'), 'error');
        } else if (e.message === 'INSTALL_ERROR') {
            window.showToast(store.i18n.t('installFailed'), 'error');
        } else {
            window.showToast(store.i18n.t('errorCheckingUpdates'), 'error');
        }
    }
}

</script>

<template>
    <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
        <div class="bg-bg-primary w-full max-w-4xl h-[700px] flex flex-col rounded-2xl shadow-2xl border border-border overflow-hidden animate-fade-in">
            <div class="p-5 border-b border-border flex justify-between items-center shrink-0">
                <h3 class="text-lg font-semibold m-0 flex items-center gap-2">
                    <i class="ph ph-gear text-xl"></i>
                    {{ store.i18n.t('settingsTitle') }}
                </h3>
                <span @click="emit('close')" class="text-2xl cursor-pointer text-text-secondary hover:text-text-primary">&times;</span>
            </div>
            
            <div class="flex border-b border-border bg-bg-secondary shrink-0 overflow-x-auto">
                <button @click="activeTab = 'general'" :class="['tab-btn', activeTab === 'general' ? 'active' : '']">{{ store.i18n.t('general') }}</button>
                <button @click="activeTab = 'feeds'" :class="['tab-btn', activeTab === 'feeds' ? 'active' : '']">{{ store.i18n.t('feeds') }}</button>
                <button @click="activeTab = 'about'" :class="['tab-btn', activeTab === 'about' ? 'active' : '']">{{ store.i18n.t('about') }}</button>
            </div>

            <div class="flex-1 overflow-y-auto p-6">
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
                />
                
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
    </div>
</template>

<style scoped>
.tab-btn {
    @apply px-5 py-3 bg-transparent border-b-2 border-transparent text-text-secondary font-semibold cursor-pointer hover:text-text-primary transition-colors;
}
.tab-btn.active {
    @apply text-accent border-accent;
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
</style>
