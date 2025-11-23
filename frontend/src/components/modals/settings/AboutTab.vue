<script setup>
import { store } from '../../../store.js';
import { ref, onMounted } from 'vue';

const emit = defineEmits(['check-updates', 'download-install-update']);

const props = defineProps({
    updateInfo: { type: Object, default: null },
    checkingUpdates: { type: Boolean, default: false },
    downloadingUpdate: { type: Boolean, default: false },
    installingUpdate: { type: Boolean, default: false },
    downloadProgress: { type: Number, default: 0 }
});

const appVersion = ref('1.1.4');

onMounted(async () => {
    // Fetch current version from API
    try {
        const versionRes = await fetch('/api/version');
        if (versionRes.ok) {
            const versionData = await versionRes.json();
            appVersion.value = versionData.version;
        }
    } catch (e) {
        console.error('Error fetching version:', e);
    }
});

function handleCheckUpdates() {
    emit('check-updates');
}

function handleDownloadInstall() {
    emit('download-install-update');
}
</script>

<template>
    <div class="text-center py-10">
        <img src="/assets/logo.svg" alt="Logo" class="h-16 w-auto mb-4 mx-auto">
        <h3 class="text-xl font-bold mb-2">{{ store.i18n.t('appName') }}</h3>
        <p class="text-text-secondary">{{ store.i18n.t('aboutApp') }}</p>
        <p class="text-text-secondary text-sm mt-2">{{ store.i18n.t('version') }} {{ appVersion }}</p>
        
        <div class="mt-6 mb-6 flex justify-center">
            <button @click="handleCheckUpdates" :disabled="checkingUpdates" class="btn-secondary justify-center">
                <i class="ph ph-arrows-clockwise" :class="{'animate-spin': checkingUpdates}"></i>
                {{ checkingUpdates ? store.i18n.t('checking') : store.i18n.t('checkForUpdates') }}
            </button>
        </div>

        <div v-if="updateInfo && !updateInfo.error" class="mt-4 mx-auto max-w-md text-left bg-bg-secondary p-4 rounded-lg border border-border">
            <div class="flex items-start gap-3">
                <i v-if="updateInfo.has_update" class="ph ph-arrow-circle-up text-2xl text-green-500 mt-0.5"></i>
                <i v-else class="ph ph-check-circle text-2xl text-accent mt-0.5"></i>
                <div class="flex-1">
                    <h4 class="font-semibold mb-1">
                        {{ updateInfo.has_update ? store.i18n.t('updateAvailable') : store.i18n.t('upToDate') }}
                    </h4>
                    <div class="text-sm text-text-secondary space-y-1">
                        <div>{{ store.i18n.t('currentVersion') }}: {{ updateInfo.current_version }}</div>
                        <div v-if="updateInfo.has_update">{{ store.i18n.t('latestVersion') }}: {{ updateInfo.latest_version }}</div>
                    </div>
                    
                    <!-- Download and Install Button -->
                    <div v-if="updateInfo.has_update && updateInfo.download_url" class="mt-3">
                        <button 
                            @click="handleDownloadInstall" 
                            :disabled="downloadingUpdate || installingUpdate"
                            class="btn-primary w-full justify-center">
                            <i v-if="downloadingUpdate" class="ph ph-circle-notch animate-spin"></i>
                            <i v-else-if="installingUpdate" class="ph ph-gear animate-spin"></i>
                            <i v-else class="ph ph-download-simple"></i>
                            <span v-if="downloadingUpdate">{{ store.i18n.t('downloading') }} {{ downloadProgress }}%</span>
                            <span v-else-if="installingUpdate">{{ store.i18n.t('installingUpdate') }}</span>
                            <span v-else>{{ store.i18n.t('downloadUpdate') }}</span>
                        </button>
                        
                        <!-- Progress bar -->
                        <div v-if="downloadingUpdate" class="mt-2 w-full bg-bg-tertiary rounded-full h-2 overflow-hidden">
                            <div class="bg-accent h-full transition-all duration-300" :style="{ width: downloadProgress + '%' }"></div>
                        </div>
                    </div>
                    
                    <!-- Fallback to GitHub if no download URL -->
                    <div v-else-if="updateInfo.has_update && !updateInfo.download_url" class="mt-3 text-xs text-text-secondary">
                        <p class="mb-2">No installer available for your platform. Please download manually:</p>
                        <a href="https://github.com/WCY-dt/MrRSS/releases/latest" target="_blank" class="text-accent hover:underline">
                            View on GitHub
                        </a>
                    </div>
                </div>
            </div>
        </div>

        <div class="mt-6">
            <a href="https://github.com/WCY-dt/MrRSS" target="_blank" class="inline-flex items-center gap-2 text-accent hover:text-accent-hover transition-colors text-sm font-medium">
                <i class="ph ph-github-logo text-xl"></i>
                {{ store.i18n.t('viewOnGitHub') }}
            </a>
        </div>
    </div>
</template>

<style scoped>
.btn-secondary {
    @apply bg-transparent border border-border text-text-primary px-4 py-2 rounded-md cursor-pointer flex items-center gap-2 font-medium hover:bg-bg-tertiary transition-colors;
}
.btn-secondary:disabled {
    @apply opacity-50 cursor-not-allowed;
}
.btn-primary {
    @apply bg-accent text-white border-none px-5 py-2.5 rounded-lg cursor-pointer font-semibold hover:bg-accent-hover transition-colors flex items-center gap-2;
}
.btn-primary:disabled {
    @apply opacity-50 cursor-not-allowed;
}
.animate-spin {
    animation: spin 1s linear infinite;
}
@keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
}
</style>
