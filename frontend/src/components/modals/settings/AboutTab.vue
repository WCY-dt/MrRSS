<script setup>
import { store } from '../../../store.js';
import { ref, onMounted } from 'vue';
import { 
    PhArrowsClockwise, PhArrowCircleUp, PhCheckCircle, PhCircleNotch, 
    PhGear, PhDownloadSimple, PhGithubLogo 
} from "@phosphor-icons/vue";

const emit = defineEmits(['check-updates', 'download-install-update']);

const props = defineProps({
    updateInfo: { type: Object, default: null },
    checkingUpdates: { type: Boolean, default: false },
    downloadingUpdate: { type: Boolean, default: false },
    installingUpdate: { type: Boolean, default: false },
    downloadProgress: { type: Number, default: 0 }
});

const appVersion = ref('1.2.1');

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
    <div class="text-center py-6 sm:py-10 px-2">
        <img src="/assets/logo.svg" alt="Logo" class="h-12 sm:h-16 w-auto mb-3 sm:mb-4 mx-auto">
        <h3 class="text-lg sm:text-xl font-bold mb-2">{{ store.i18n.t('appName') }}</h3>
        <p class="text-text-secondary text-sm sm:text-base">{{ store.i18n.t('aboutApp') }}</p>
        <p class="text-text-secondary text-xs sm:text-sm mt-2">{{ store.i18n.t('version') }} {{ appVersion }}</p>
        
        <div class="mt-4 sm:mt-6 mb-4 sm:mb-6 flex justify-center">
            <button @click="handleCheckUpdates" :disabled="checkingUpdates" class="btn-secondary justify-center text-sm sm:text-base">
                <PhArrowsClockwise :size="18" class="sm:w-5 sm:h-5" :class="{'animate-spin': checkingUpdates}" />
                {{ checkingUpdates ? store.i18n.t('checking') : store.i18n.t('checkForUpdates') }}
            </button>
        </div>

        <div v-if="updateInfo && !updateInfo.error" class="mt-3 sm:mt-4 mx-auto max-w-md text-left bg-bg-secondary p-3 sm:p-4 rounded-lg border border-border">
            <div class="flex items-start gap-2 sm:gap-3">
                <PhArrowCircleUp v-if="updateInfo.has_update" :size="28" class="text-green-500 mt-0.5 shrink-0 sm:w-8 sm:h-8" />
                <PhCheckCircle v-else :size="28" class="text-accent mt-0.5 shrink-0 sm:w-8 sm:h-8" />
                <div class="flex-1 min-w-0">
                    <h4 class="font-semibold mb-1 text-sm sm:text-base">
                        {{ updateInfo.has_update ? store.i18n.t('updateAvailable') : store.i18n.t('upToDate') }}
                    </h4>
                    <div class="text-xs sm:text-sm text-text-secondary space-y-1">
                        <div class="truncate">{{ store.i18n.t('currentVersion') }}: {{ updateInfo.current_version }}</div>
                        <div v-if="updateInfo.has_update" class="truncate">{{ store.i18n.t('latestVersion') }}: {{ updateInfo.latest_version }}</div>
                    </div>
                    
                    <!-- Download and Install Button -->
                    <div v-if="updateInfo.has_update && updateInfo.download_url" class="mt-2 sm:mt-3">
                        <button 
                            @click="handleDownloadInstall" 
                            :disabled="downloadingUpdate || installingUpdate"
                            class="btn-primary w-full justify-center text-sm sm:text-base">
                            <PhCircleNotch v-if="downloadingUpdate" :size="18" class="animate-spin sm:w-5 sm:h-5" />
                            <PhGear v-else-if="installingUpdate" :size="18" class="animate-spin sm:w-5 sm:h-5" />
                            <PhDownloadSimple v-else :size="18" class="sm:w-5 sm:h-5" />
                            <span v-if="downloadingUpdate">{{ store.i18n.t('downloading') }} {{ downloadProgress }}%</span>
                            <span v-else-if="installingUpdate">{{ store.i18n.t('installingUpdate') }}</span>
                            <span v-else>{{ store.i18n.t('downloadUpdate') }}</span>
                        </button>
                        
                        <!-- Progress bar -->
                        <div v-if="downloadingUpdate" class="mt-2 w-full bg-bg-tertiary rounded-full h-1.5 sm:h-2 overflow-hidden">
                            <div class="bg-accent h-full transition-all duration-300" :style="{ width: downloadProgress + '%' }"></div>
                        </div>
                    </div>
                    
                    <!-- Fallback to GitHub if no download URL -->
                    <div v-else-if="updateInfo.has_update && !updateInfo.download_url" class="mt-2 sm:mt-3 text-xs text-text-secondary">
                        <p class="mb-2">No installer available for your platform. Please download manually:</p>
                        <a href="https://github.com/WCY-dt/MrRSS/releases/latest" target="_blank" class="text-accent hover:underline break-all">
                            View on GitHub
                        </a>
                    </div>
                </div>
            </div>
        </div>

        <div class="mt-4 sm:mt-6">
            <a href="https://github.com/WCY-dt/MrRSS" target="_blank" class="inline-flex items-center gap-1.5 sm:gap-2 text-accent hover:text-accent-hover transition-colors text-xs sm:text-sm font-medium">
                <PhGithubLogo :size="20" class="sm:w-6 sm:h-6" />
                {{ store.i18n.t('viewOnGitHub') }}
            </a>
        </div>
    </div>
</template>

<style scoped>
.btn-secondary {
    @apply bg-transparent border border-border text-text-primary px-3 sm:px-4 py-1.5 sm:py-2 rounded-md cursor-pointer flex items-center gap-1.5 sm:gap-2 font-medium hover:bg-bg-tertiary transition-colors;
}
.btn-secondary:disabled {
    @apply opacity-50 cursor-not-allowed;
}
.btn-primary {
    @apply bg-accent text-white border-none px-4 sm:px-5 py-2 sm:py-2.5 rounded-lg cursor-pointer font-semibold hover:bg-accent-hover transition-colors flex items-center gap-1.5 sm:gap-2;
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
