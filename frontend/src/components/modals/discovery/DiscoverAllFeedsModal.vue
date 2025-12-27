<script setup lang="ts">
import { watch, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhX, PhCircleNotch } from '@phosphor-icons/vue';
import { useDiscoverAllFeeds } from '@/composables/discovery/useDiscoverAllFeeds';
import DiscoveryProgress from './DiscoveryProgress.vue';
import DiscoveryResults from './DiscoveryResults.vue';
import { useModalClose } from '@/composables/ui/useModalClose';

const { t } = useI18n();

interface Props {
  show: boolean;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  close: [];
}>();

const {
  isDiscovering,
  discoveredFeeds,
  selectedFeeds,
  errorMessage,
  progressMessage,
  progressDetail,
  progressCounts,
  isSubscribing,
  hasSelection,
  allSelected,
  startDiscovery,
  toggleFeedSelection,
  selectAll,
  subscribeSelected,
  cleanup,
} = useDiscoverAllFeeds();

// Modal close handling
useModalClose(() => close());

function close() {
  cleanup();
  emit('close');
}

// Auto-start discovery when component is mounted and shown
onMounted(() => {
  if (props.show) {
    startDiscovery();
  }
});

// Watch for modal opening and trigger discovery (for when modal is reused)
watch(
  () => props.show,
  (newShow, oldShow) => {
    if (newShow && !oldShow) {
      startDiscovery();
    }
  }
);
</script>

<template>
  <div
    v-if="show"
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-2 sm:p-4"
    data-modal-open="true"
    style="will-change: transform; transform: translateZ(0)"
    @click.self="close"
  >
    <div
      class="bg-bg-primary w-full max-w-4xl h-full sm:h-auto sm:max-h-[90vh] rounded-none sm:rounded-2xl shadow-2xl border border-border flex flex-col"
    >
      <!-- Header -->
      <div
        class="flex justify-between items-center p-4 sm:p-6 border-b border-border bg-gradient-to-r from-accent/5 to-transparent shrink-0"
      >
        <div class="min-w-0 flex-1">
          <h2 class="text-base sm:text-xl font-bold text-text-primary">
            {{ t('discoverAllFeeds') }}
          </h2>
          <p class="text-xs sm:text-sm text-text-secondary mt-1">{{ t('discoverAllFeedsDesc') }}</p>
        </div>
        <button
          class="p-1.5 sm:p-2 hover:bg-bg-tertiary rounded-lg transition-colors shrink-0 ml-2"
          @click="close"
        >
          <PhX :size="20" class="sm:w-6 sm:h-6 text-text-secondary" />
        </button>
      </div>

      <!-- Content -->
      <div class="flex-1 overflow-y-auto p-4 sm:p-6">
        <!-- Loading State -->
        <DiscoveryProgress
          v-if="isDiscovering"
          :progress-message="progressMessage"
          :progress-detail="progressDetail"
          :progress-counts="progressCounts"
        />

        <!-- Error State -->
        <div
          v-else-if="errorMessage"
          class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-3 sm:p-4 text-red-600 dark:text-red-400 text-sm sm:text-base"
        >
          {{ errorMessage }}
        </div>

        <!-- Results -->
        <DiscoveryResults
          v-if="discoveredFeeds.length > 0"
          :discovered-feeds="discoveredFeeds"
          :selected-feeds="selectedFeeds"
          :all-selected="allSelected"
          @toggle-feed-selection="toggleFeedSelection"
          @select-all="selectAll"
        />

        <!-- Initial State (should not be visible as discovery auto-starts) -->
        <div v-else class="text-center py-12 sm:py-16">
          <PhCircleNotch
            :size="48"
            class="sm:w-16 sm:h-16 text-accent mx-auto mb-3 sm:mb-4 animate-spin"
          />
          <p class="text-text-secondary text-base sm:text-lg">{{ t('preparing') }}...</p>
        </div>
      </div>

      <!-- Footer -->
      <div
        class="flex flex-col-reverse sm:flex-row sm:justify-between items-stretch sm:items-center gap-2 sm:gap-3 p-4 sm:p-6 border-t border-border bg-bg-secondary/50 shrink-0"
      >
        <button class="btn-secondary text-sm sm:text-base" :disabled="isSubscribing" @click="close">
          {{ t('cancel') }}
        </button>
        <button
          :disabled="!hasSelection || isSubscribing"
          :class="[
            'btn-primary flex items-center justify-center gap-2 text-sm sm:text-base',
            (!hasSelection || isSubscribing) && 'opacity-50 cursor-not-allowed',
          ]"
          @click="subscribeSelected"
        >
          <PhCircleNotch v-if="isSubscribing" :size="16" class="animate-spin" />
          {{ isSubscribing ? t('subscribing') : t('subscribeSelected') }}
          <span
            v-if="hasSelection && !isSubscribing"
            class="bg-white/20 px-1.5 sm:px-2 py-0.5 rounded-full text-xs sm:text-sm"
            >({{ selectedFeeds.size }})</span
          >
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "../../../style.css";

.btn-primary {
  @apply px-4 sm:px-6 py-2 sm:py-2.5 bg-accent text-white rounded-lg hover:bg-accent-hover transition-all font-medium shadow-sm hover:shadow-md;
}

.btn-secondary {
  @apply px-4 sm:px-6 py-2 sm:py-2.5 bg-bg-tertiary text-text-primary rounded-lg hover:opacity-80 transition-all font-medium;
}
</style>
