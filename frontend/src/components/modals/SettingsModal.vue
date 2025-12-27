<script setup lang="ts">
import { useAppStore } from '@/stores/app';
import { useI18n } from 'vue-i18n';
import { ref, onMounted, onUnmounted, type Ref } from 'vue';
import GeneralTab from './settings/general/GeneralTab.vue';
import FeedsTab from './settings/feeds/FeedsTab.vue';
import ContentTab from './settings/content/ContentTab.vue';
import AITab from './settings/ai/AITab.vue';
import NetworkTab from './settings/network/NetworkTab.vue';
import PluginsTab from './settings/plugins/PluginsTab.vue';
import ShortcutsTab from './settings/shortcuts/ShortcutsTab.vue';
import RulesTab from './settings/rules/RulesTab.vue';
import AboutTab from './settings/about/AboutTab.vue';
import DiscoverAllFeedsModal from './discovery/DiscoverAllFeedsModal.vue';
import { PhGear } from '@phosphor-icons/vue';
import type { TabName } from '@/types/settings';
import type { ThemePreference } from '@/stores/app';
import { useSettings } from '@/composables/core/useSettings';
import { useAppUpdates } from '@/composables/core/useAppUpdates';
import { useFeedManagement } from '@/composables/feed/useFeedManagement';
import { useModalClose } from '@/composables/ui/useModalClose';

const store = useAppStore();
const { t } = useI18n();

// Use composables
const { settings, fetchSettings, applySettings } = useSettings();
const {
  updateInfo,
  checkingUpdates,
  downloadingUpdate,
  installingUpdate,
  downloadProgress,
  checkForUpdates: handleCheckUpdates,
  downloadAndInstallUpdate: handleDownloadInstallUpdate,
} = useAppUpdates();
const {
  handleImportOPML,
  handleExportOPML,
  handleCleanupDatabase,
  handleAddFeed,
  handleEditFeed,
  handleDeleteFeed,
  handleBatchDelete,
  handleBatchMove,
} = useFeedManagement();

const emit = defineEmits<{
  close: [];
}>();

const activeTab: Ref<TabName> = ref('general');
const showDiscoverAllModal = ref(false);
const tabsContainer = ref<HTMLElement>();

// Modal close handling
useModalClose(() => emit('close'));

onMounted(async () => {
  try {
    const data = await fetchSettings();
    applySettings(data, (theme: string) => store.setTheme(theme as ThemePreference));
  } catch (e) {
    console.error('Error loading settings:', e);
  }

  // Add wheel event listener for horizontal scrolling
  if (tabsContainer.value) {
    const handleWheel = (e: WheelEvent) => {
      e.preventDefault();
      tabsContainer.value!.scrollLeft += e.deltaY * 0.1; // Reduce scroll speed
    };
    tabsContainer.value.addEventListener('wheel', handleWheel);

    // Store cleanup function
    const cleanup = () => {
      if (tabsContainer.value) {
        tabsContainer.value.removeEventListener('wheel', handleWheel);
      }
    };

    // Cleanup on unmount
    onUnmounted(cleanup);
  }
});

function handleDiscoverAll() {
  showDiscoverAllModal.value = true;
}
</script>

<template>
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-2 sm:p-4"
    data-modal-open="true"
    style="will-change: transform; transform: translateZ(0)"
    @click.self="emit('close')"
  >
    <div
      class="bg-bg-primary w-full max-w-4xl h-full sm:h-[900px] sm:max-h-[90vh] flex flex-col rounded-none sm:rounded-2xl shadow-2xl border border-border overflow-hidden animate-fade-in"
    >
      <div class="p-3 sm:p-5 border-b border-border flex justify-between items-center shrink-0">
        <h3 class="text-text-secondary sm:text-lg font-semibold m-0 flex items-center gap-2">
          <PhGear :size="20" :weight="'fill'" class="sm:w-6 sm:h-6" />
          {{ t('settingsTitle') }}
        </h3>
        <span
          class="text-2xl cursor-pointer text-text-secondary hover:text-text-primary"
          @click="emit('close')"
          >&times;</span
        >
      </div>

      <div
        ref="tabsContainer"
        class="flex border-b border-border bg-bg-secondary shrink-0 overflow-x-auto scrollbar-hide"
      >
        <button
          :class="['tab-btn', activeTab === 'general' ? 'active' : '']"
          @click="activeTab = 'general'"
        >
          {{ t('general') }}
        </button>
        <button
          :class="['tab-btn', activeTab === 'feeds' ? 'active' : '']"
          @click="activeTab = 'feeds'"
        >
          {{ t('feeds') }}
        </button>
        <button
          :class="['tab-btn', activeTab === 'content' ? 'active' : '']"
          @click="activeTab = 'content'"
        >
          {{ t('content') }}
        </button>
        <button :class="['tab-btn', activeTab === 'ai' ? 'active' : '']" @click="activeTab = 'ai'">
          {{ t('ai') }}
        </button>
        <button
          :class="['tab-btn', activeTab === 'rules' ? 'active' : '']"
          @click="activeTab = 'rules'"
        >
          {{ t('rules') }}
        </button>
        <button
          :class="['tab-btn', activeTab === 'network' ? 'active' : '']"
          @click="activeTab = 'network'"
        >
          {{ t('network') }}
        </button>
        <button
          :class="['tab-btn', activeTab === 'plugins' ? 'active' : '']"
          @click="activeTab = 'plugins'"
        >
          {{ t('plugins') }}
        </button>
        <button
          :class="['tab-btn', activeTab === 'shortcuts' ? 'active' : '']"
          @click="activeTab = 'shortcuts'"
        >
          {{ t('shortcuts') }}
        </button>
        <button
          :class="['tab-btn', activeTab === 'about' ? 'active' : '']"
          @click="activeTab = 'about'"
        >
          {{ t('about') }}
        </button>
      </div>

      <div class="flex-1 overflow-y-auto p-3 sm:p-6 min-h-0">
        <GeneralTab
          v-if="activeTab === 'general'"
          :settings="settings"
          @update:settings="settings = $event"
        />

        <FeedsTab
          v-if="activeTab === 'feeds'"
          :settings="settings"
          @import-opml="handleImportOPML"
          @export-opml="handleExportOPML"
          @cleanup-database="handleCleanupDatabase"
          @add-feed="handleAddFeed"
          @edit-feed="handleEditFeed"
          @delete-feed="handleDeleteFeed"
          @batch-delete="handleBatchDelete"
          @batch-move="handleBatchMove"
          @discover-all="handleDiscoverAll"
          @update:settings="settings = $event"
        />

        <ContentTab
          v-if="activeTab === 'content'"
          :settings="settings"
          @update:settings="settings = $event"
        />

        <AITab
          v-if="activeTab === 'ai'"
          :settings="settings"
          @update:settings="settings = $event"
        />

        <NetworkTab
          v-if="activeTab === 'network'"
          :settings="settings"
          @update:settings="settings = $event"
        />

        <PluginsTab
          v-if="activeTab === 'plugins'"
          :settings="settings"
          @update:settings="settings = $event"
        />

        <RulesTab
          v-if="activeTab === 'rules'"
          :settings="settings"
          @update:settings="settings = $event"
        />

        <ShortcutsTab
          v-if="activeTab === 'shortcuts'"
          :settings="settings"
          @update:settings="settings = $event"
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

    <!-- Discover All Feeds Modal -->
    <DiscoverAllFeedsModal :show="showDiscoverAllModal" @close="showDiscoverAllModal = false" />
  </div>
</template>

<style scoped>
@reference "../../style.css";

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
  background: var(--accent-color);
  box-shadow: 0 0 8px var(--accent-color);
  opacity: 0.8;
}
.btn-primary {
  @apply bg-accent text-white border-none px-5 py-2.5 rounded-lg cursor-pointer font-semibold hover:bg-accent-hover transition-colors;
}
.animate-fade-in {
  animation: modalFadeIn 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}
@keyframes modalFadeIn {
  from {
    transform: translateY(-20px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}
.scrollbar-hide {
  -ms-overflow-style: none;
  scrollbar-width: none;
}
.scrollbar-hide::-webkit-scrollbar {
  display: none;
}
</style>
