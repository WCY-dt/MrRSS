<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { PhYoutubeLogo } from '@phosphor-icons/vue';
import { useI18n } from 'vue-i18n';

interface Props {
  videoUrl: string;
  articleTitle: string;
}

const props = defineProps<Props>();

const { t } = useI18n();

const iframeRef = ref<HTMLIFrameElement | null>(null);
const isLoading = ref(true);

function onLoad() {
  isLoading.value = false;
}

function onError() {
  isLoading.value = false;
  window.showToast(t('videoLoadError'), 'error');
}

// Open video in new tab
function openInNewTab() {
  // Convert embed URL back to watch URL
  const watchURL = props.videoUrl.replace('/embed/', '/watch?v=');
  window.open(watchURL, '_blank');
}

onMounted(() => {
  if (iframeRef.value) {
    iframeRef.value.addEventListener('load', onLoad);
    iframeRef.value.addEventListener('error', onError);
  }
});

onUnmounted(() => {
  if (iframeRef.value) {
    iframeRef.value.removeEventListener('load', onLoad);
    iframeRef.value.removeEventListener('error', onError);
  }
});
</script>

<template>
  <div class="bg-bg-secondary border border-border rounded-lg overflow-hidden mb-4 sm:mb-6">
    <!-- Header -->
    <div class="flex items-center justify-between p-3 border-b border-border">
      <div class="flex items-center gap-2">
        <PhYoutubeLogo :size="20" class="text-red-600 flex-shrink-0" />
        <span class="text-sm font-medium text-text-primary">{{ t('youtubeVideo') }}</span>
      </div>
      <button
        @click="openInNewTab"
        class="text-xs text-accent hover:underline"
        :title="t('openInYouTube')"
      >
        {{ t('openInYouTube') }}
      </button>
    </div>

    <!-- Video Player -->
    <div class="relative w-full" style="padding-bottom: 56.25%">
      <!-- 16:9 Aspect Ratio -->
      <iframe
        ref="iframeRef"
        :src="videoUrl"
        :title="articleTitle"
        class="absolute top-0 left-0 w-full h-full border-none"
        allow="
          accelerometer;
          autoplay;
          clipboard-write;
          encrypted-media;
          gyroscope;
          picture-in-picture;
          web-share;
        "
        allowfullscreen
      />

      <!-- Loading indicator -->
      <div
        v-if="isLoading"
        class="absolute inset-0 flex items-center justify-center bg-bg-tertiary"
      >
        <div
          class="animate-spin rounded-full h-12 w-12 border-4 border-accent border-t-transparent"
        ></div>
      </div>
    </div>
  </div>
</template>
