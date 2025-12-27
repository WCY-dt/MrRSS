<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { PhArrowCircleUp, PhDownloadSimple, PhX, PhCircleNotch, PhGear } from '@phosphor-icons/vue';

interface UpdateInfo {
  has_update: boolean;
  current_version: string;
  latest_version: string;
  download_url?: string;
  error?: string;
}

interface Props {
  updateInfo: UpdateInfo;
  downloadingUpdate?: boolean;
  installingUpdate?: boolean;
  downloadProgress?: number;
}

withDefaults(defineProps<Props>(), {
  downloadingUpdate: false,
  installingUpdate: false,
  downloadProgress: 0,
});

const emit = defineEmits<{
  close: [];
  update: [];
}>();

const { t } = useI18n();

function handleClose() {
  emit('close');
}

function handleUpdate() {
  emit('update');
}
</script>

<template>
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4"
    @click.self="handleClose"
  >
    <div
      class="bg-bg-primary w-full max-w-md rounded-2xl shadow-2xl border border-border overflow-hidden animate-fade-in"
    >
      <!-- Header -->
      <div class="flex items-center justify-between p-4 sm:p-6 border-b border-border">
        <div class="flex items-center gap-3">
          <div class="bg-green-500/20 rounded-full p-2">
            <PhArrowCircleUp :size="28" class="text-green-500" />
          </div>
          <h3 class="text-lg sm:text-xl font-bold">{{ t('updateAvailable') }}</h3>
        </div>
        <button
          class="text-text-secondary hover:text-text-primary transition-colors p-1 rounded-md hover:bg-bg-secondary"
          @click="handleClose"
        >
          <PhX :size="24" />
        </button>
      </div>

      <!-- Content -->
      <div class="p-4 sm:p-6">
        <p class="text-text-secondary text-sm mb-4">
          {{ t('newVersionAvailable', { version: updateInfo.latest_version }) }}
        </p>

        <div class="bg-bg-secondary rounded-lg p-3 sm:p-4 space-y-2 text-sm">
          <div class="flex justify-between items-center">
            <span class="text-text-secondary">{{ t('currentVersion') }}:</span>
            <span class="font-mono font-medium">{{ updateInfo.current_version }}</span>
          </div>
          <div class="flex justify-between items-center">
            <span class="text-text-secondary">{{ t('latestVersion') }}:</span>
            <span class="font-mono font-medium text-green-500">{{
              updateInfo.latest_version
            }}</span>
          </div>
        </div>

        <p v-if="!updateInfo.download_url" class="text-text-secondary text-xs mt-4">
          {{ t('noInstallerAvailable') }}
          <a
            href="https://github.com/WCY-dt/MrRSS/releases/latest"
            target="_blank"
            class="text-accent hover:underline"
          >
            {{ t('viewOnGitHub') }}
          </a>
        </p>
      </div>

      <!-- Footer -->
      <div
        class="p-4 sm:p-6 border-t border-border flex flex-col-reverse sm:flex-row gap-3 justify-end"
      >
        <button
          class="btn-secondary w-full sm:w-auto justify-center"
          :disabled="downloadingUpdate || installingUpdate"
          @click="handleClose"
        >
          {{ t('notNow') }}
        </button>
        <button
          v-if="updateInfo.download_url"
          class="btn-primary w-full sm:w-auto justify-center"
          :disabled="downloadingUpdate || installingUpdate"
          @click="handleUpdate"
        >
          <PhCircleNotch v-if="downloadingUpdate" :size="20" class="animate-spin" />
          <PhGear v-else-if="installingUpdate" :size="20" class="animate-spin" />
          <PhDownloadSimple v-else :size="20" />
          <span v-if="downloadingUpdate">{{ t('downloading') }} {{ downloadProgress }}%</span>
          <span v-else-if="installingUpdate">{{ t('installingUpdate') }}</span>
          <span v-else>{{ t('updateNow') }}</span>
        </button>
      </div>

      <!-- Progress bar -->
      <div v-if="downloadingUpdate" class="px-4 sm:px-6 pb-4 sm:pb-6">
        <div class="w-full bg-bg-tertiary rounded-full h-2 overflow-hidden">
          <div
            class="bg-accent h-full transition-all duration-300"
            :style="{ width: downloadProgress + '%' }"
          ></div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "../../../style.css";

.btn-secondary {
  @apply bg-bg-tertiary border border-border text-text-primary px-4 py-2 rounded-lg cursor-pointer font-medium hover:bg-bg-secondary transition-colors flex items-center gap-2;
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
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
.animate-fade-in {
  animation: fadeIn 0.2s ease-out;
}
@keyframes fadeIn {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}
</style>
