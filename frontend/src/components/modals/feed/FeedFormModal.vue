<script setup lang="ts">
import { ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhCaretDown, PhCaretRight } from '@phosphor-icons/vue';
import type { Feed } from '@/types/models';
import type { RouteMatch } from '@/types/rsshub';
import { useModalClose } from '@/composables/ui/useModalClose';
import { useFeedForm } from '@/composables/feed/useFeedForm';
import { useAppStore } from '@/stores/app';
import UrlInput from './parts/UrlInput.vue';
import ScriptSelector from './parts/ScriptSelector.vue';
import XPathConfig from './parts/XPathConfig.vue';
import CategorySelector from './parts/CategorySelector.vue';
import AdvancedSettings from './parts/AdvancedSettings.vue';
import RSSHubRouteSelectorModal from './RSSHubRouteSelectorModal.vue';

interface Props {
  mode: 'add' | 'edit';
  feed?: Feed;
}

const props = defineProps<Props>();

const { t } = useI18n();
const appStore = useAppStore();

// Use the shared feed form composable
const {
  imageGalleryEnabled,
  feedType,
  title,
  url,
  category,
  categorySelection,
  showCustomCategory,
  scriptPath,
  hideFromTimeline,
  isImageMode,
  xpathType,
  xpathItem,
  xpathItemTitle,
  xpathItemContent,
  xpathItemUri,
  xpathItemAuthor,
  xpathItemTimestamp,
  xpathItemTimeFormat,
  xpathItemThumbnail,
  xpathItemCategories,
  xpathItemUid,
  proxyMode,
  proxyType,
  proxyHost,
  proxyPort,
  proxyUsername,
  proxyPassword,
  refreshMode,
  refreshInterval,
  isSubmitting,
  showAdvancedSettings,
  availableScripts,
  scriptsDir,
  existingCategories,
  isFormValid,
  isUrlInvalid,
  isScriptInvalid,
  isXpathItemInvalid,
  handleCategoryChange,
  buildProxyUrl,
  getRefreshInterval,
  resetForm,
  openScriptsFolder,
} = useFeedForm(props.feed);

const emit = defineEmits<{
  close: [];
  added: [];
  updated: [];
}>();

// RSSHub route selector state
const showRouteSelector = ref(false);
const rsshubMatches = ref<RouteMatch[]>([]);
const rsshubOriginalUrl = ref('');
const suggestedRoutes = ref<RouteMatch[]>([]); // Routes to show below URL input
const isLoadingRoutes = ref(false);
let routeMatchTimer: ReturnType<typeof setTimeout> | null = null;

// Modal close handling
useModalClose(() => close());

function close() {
  emit('close');
}

// Check for RSSHub routes when URL changes
async function checkRSSHubRoutes() {
  // Only check in URL mode for new subscriptions
  if (props.mode !== 'add' || feedType.value !== 'url' || !url.value) {
    suggestedRoutes.value = [];
    return;
  }

  // Check if RSSHub is enabled
  const settings = appStore.settings;

  // Guard against undefined settings
  if (!settings) {
    suggestedRoutes.value = [];
    return;
  }

  console.log('[RSSHub] Checking routes, settings:', {
    enabled: settings.rsshub_enabled,
    fallback: settings.rsshub_fallback_enabled,
    url: url.value,
  });

  if (!settings.rsshub_enabled || !settings.rsshub_fallback_enabled) {
    suggestedRoutes.value = [];
    return;
  }

  // Debounce the API call
  if (routeMatchTimer) {
    clearTimeout(routeMatchTimer);
  }

  routeMatchTimer = setTimeout(async () => {
    try {
      isLoadingRoutes.value = true;
      console.log('[RSSHub] Fetching routes for:', url.value);

      const res = await fetch('/api/rsshub/routes/match', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ url: url.value }),
      });

      console.log('[RSSHub] Response status:', res.status);

      if (res.ok) {
        const data = await res.json();
        console.log('[RSSHub] Response data:', data);

        if (data.matches && data.matches.length > 0) {
          suggestedRoutes.value = data.matches;
          console.log('[RSSHub] Found', data.matches.length, 'routes');
        } else {
          suggestedRoutes.value = [];
          console.log('[RSSHub] No matches found');
        }
      } else {
        suggestedRoutes.value = [];
        console.log('[RSSHub] Request failed:', res.status);
      }
    } catch (e) {
      console.error('[RSSHub] Failed to check routes:', e);
      suggestedRoutes.value = [];
    } finally {
      isLoadingRoutes.value = false;
    }
  }, 500); // 500ms debounce
}

// Watch URL changes to check for RSSHub routes
watch([url, feedType], () => {
  checkRSSHubRoutes();
});

async function submit() {
  if (!isFormValid.value) return;
  isSubmitting.value = true;

  try {
    const body: Record<string, string | boolean | number> = {
      category: category.value,
      title: title.value,
      hide_from_timeline: hideFromTimeline.value,
      is_image_mode: isImageMode.value,
      refresh_interval: getRefreshInterval(),
    };

    // Handle proxy settings
    if (proxyMode.value === 'custom') {
      body.proxy_enabled = true;
      body.proxy_url = buildProxyUrl();
    } else if (proxyMode.value === 'global') {
      body.proxy_enabled = true;
      body.proxy_url = '';
    } else {
      body.proxy_enabled = false;
      body.proxy_url = '';
    }

    if (feedType.value === 'url') {
      body.url = url.value;
      if (props.mode === 'edit') {
        body.script_path = '';
      }
    } else if (feedType.value === 'script') {
      if (props.mode === 'add') {
        body.script_path = scriptPath.value;
      } else {
        body.url = scriptPath.value ? 'script://' + scriptPath.value : props.feed!.url;
        body.script_path = scriptPath.value;
      }
    } else if (feedType.value === 'xpath') {
      body.url = url.value;
      if (props.mode === 'edit') {
        body.script_path = '';
      }
      body.type = xpathType.value;
      body.xpath_item = xpathItem.value;
      body.xpath_item_title = xpathItemTitle.value;
      body.xpath_item_content = xpathItemContent.value;
      body.xpath_item_uri = xpathItemUri.value;
      body.xpath_item_author = xpathItemAuthor.value;
      body.xpath_item_timestamp = xpathItemTimestamp.value;
      body.xpath_item_time_format = xpathItemTimeFormat.value;
      body.xpath_item_thumbnail = xpathItemThumbnail.value;
      body.xpath_item_categories = xpathItemCategories.value;
      body.xpath_item_uid = xpathItemUid.value;
    }

    if (props.mode === 'edit') {
      body.id = props.feed!.id;
    }

    const endpoint = props.mode === 'add' ? '/api/feeds/add' : '/api/feeds/update';
    const res = await fetch(endpoint, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });

    if (res.ok) {
      if (props.mode === 'add') {
        emit('added');
        resetForm();
        window.showToast(t('feedAddedSuccess'), 'success');
      } else {
        emit('updated');
        window.showToast(t('feedUpdatedSuccess'), 'success');
      }
      close();
    } else {
      // Try to parse error as JSON first
      const contentType = res.headers.get('content-type');
      if (contentType && contentType.includes('application/json')) {
        try {
          const errorData: {
            error: string;
            code: string;
            original_url: string;
            matches: RouteMatch[];
          } = await res.json();

          // Check if this is an RSSHub match error
          if (errorData.code === 'RSSHUB_MATCH_FOUND' && errorData.matches) {
            // Show route selector modal
            rsshubOriginalUrl.value = errorData.original_url;
            rsshubMatches.value = errorData.matches;
            showRouteSelector.value = true;
            return;
          }
        } catch (e) {
          console.error('Failed to parse error JSON:', e);
        }
      }

      // Fallback to text error
      const errorText = await res.text();
      const errorKey = props.mode === 'add' ? 'errorAddingFeed' : 'errorUpdatingFeed';
      window.showToast(`${t(errorKey)}: ${errorText}`, 'error');
    }
  } catch (e) {
    console.error(e);
    const errorKey = props.mode === 'add' ? 'errorAddingFeed' : 'errorUpdatingFeed';
    window.showToast(t(errorKey), 'error');
  } finally {
    isSubmitting.value = false;
  }
}

// Handle RSSHub route selection
async function handleRouteSelect(match: RouteMatch) {
  // Guard against undefined settings
  const settings = appStore.settings;
  if (!settings) {
    window.showToast('Settings not loaded', 'error');
    return;
  }

  // Create RSSHub protocol URL instead of full URL
  // This allows the backend to dynamically build the RSSHub URL
  const rssHubProtocolURL = 'rsshub://' + rsshubOriginalUrl.value;

  // Build request body with RSSHub protocol URL
  isSubmitting.value = true;
  try {
    const body: Record<string, string | boolean | number> = {
      url: rssHubProtocolURL, // Use rsshub:// protocol
      category: category.value,
      title: title.value || match.title,
      hide_from_timeline: hideFromTimeline.value,
      is_image_mode: isImageMode.value,
      refresh_interval: getRefreshInterval(),
      // RSSHub metadata
      is_rsshub: true,
      rsshub_route: match.route,
      original_url: rsshubOriginalUrl.value,
    };

    // Handle proxy settings
    if (proxyMode.value === 'custom') {
      body.proxy_enabled = true;
      body.proxy_url = buildProxyUrl();
    } else if (proxyMode.value === 'global') {
      body.proxy_enabled = true;
      body.proxy_url = '';
    } else {
      body.proxy_enabled = false;
      body.proxy_url = '';
    }

    const res = await fetch('/api/feeds/add', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });

    if (res.ok) {
      showRouteSelector.value = false;
      emit('added');
      resetForm();
      window.showToast(t('feedAddedSuccess'), 'success');
      close();
    } else {
      const errorText = await res.text();
      window.showToast(`${t('errorAddingFeed')}: ${errorText}`, 'error');
    }
  } catch (e) {
    console.error(e);
    window.showToast(t('errorAddingFeed'), 'error');
  } finally {
    isSubmitting.value = false;
  }
}

function closeRouteSelector() {
  showRouteSelector.value = false;
}
</script>

<template>
  <div
    class="fixed inset-0 z-[60] flex items-center justify-center bg-black/50 backdrop-blur-sm p-2 sm:p-4"
    data-modal-open="true"
    @click.self="close"
  >
    <div
      class="bg-bg-primary w-full max-w-md h-full sm:h-auto sm:max-h-[90vh] flex flex-col rounded-none sm:rounded-2xl shadow-2xl border border-border overflow-hidden animate-fade-in"
    >
      <div class="p-3 sm:p-5 border-b border-border flex justify-between items-center shrink-0">
        <h3 class="text-base sm:text-lg font-semibold m-0">
          {{ mode === 'add' ? t('addNewFeed') : t('editFeed') }}
        </h3>
        <span
          class="text-2xl cursor-pointer text-text-secondary hover:text-text-primary"
          @click="close"
          >&times;</span
        >
      </div>
      <div class="flex-1 overflow-y-auto p-4 sm:p-6">
        <div class="mb-3 sm:mb-4">
          <label
            class="block mb-1 sm:mb-1.5 font-semibold text-xs sm:text-sm text-text-secondary"
            >{{ t('title') }}</label
          >
          <input
            v-model="title"
            type="text"
            :placeholder="mode === 'add' ? t('titlePlaceholder') : ''"
            class="input-field"
          />
        </div>

        <!-- Content switching with different modes -->
        <Transition
          name="mode-transition"
          mode="out-in"
          enter-active-class="transition-all duration-300 ease-out"
          leave-active-class="transition-all duration-200 ease-in"
          enter-from-class="opacity-0 transform translate-y-4"
          enter-to-class="opacity-100 transform translate-y-0"
          leave-from-class="opacity-100 transform translate-y-0"
          leave-to-class="opacity-0 transform -translate-y-4"
        >
          <!-- URL Input (default mode) -->
          <div v-if="feedType === 'url'" key="url-mode">
            <UrlInput v-model="url" :mode="mode" :is-invalid="mode === 'add' && isUrlInvalid" />

            <!-- RSSHub suggested routes -->
            <div v-if="suggestedRoutes.length > 0 || isLoadingRoutes" class="mt-3">
              <div v-if="isLoadingRoutes" class="text-xs text-text-tertiary text-center">
                {{ t('rsshub.loading_routes') }}
              </div>
              <div v-else class="space-y-2">
                <div class="text-xs text-text-secondary font-medium">
                  {{ t('rsshub.available_routes') }}
                </div>
                <div
                  v-for="match in suggestedRoutes"
                  :key="match.route"
                  class="p-2 sm:p-2.5 bg-bg-secondary rounded-lg border border-border hover:border-accent cursor-pointer transition-colors"
                  :class="{ 'ring-2 ring-accent': match.score >= 0.8 }"
                  @click="handleRouteSelect(match)"
                >
                  <div class="flex items-center justify-between">
                    <div class="flex-1 min-w-0">
                      <div class="flex items-center gap-2">
                        <span class="text-sm font-medium text-text-primary truncate">{{
                          match.title
                        }}</span>
                        <span
                          v-if="match.score >= 0.8"
                          class="px-1.5 py-0.5 text-xs bg-green-500/20 text-green-600 rounded shrink-0"
                        >
                          {{ t('rsshub.recommended') }}
                        </span>
                      </div>
                      <code class="text-xs text-text-tertiary break-all">{{ match.route }}</code>
                    </div>
                    <PhCaretRight :size="16" class="text-text-tertiary shrink-0 ml-2" />
                  </div>
                </div>
              </div>
            </div>

            <!-- Mode switching links -->
            <div class="mt-3 text-center">
              <div class="text-xs text-text-tertiary">
                {{ mode === 'add' ? t('orTry') : t('switchTo') }}
                <button
                  type="button"
                  class="text-xs text-accent hover:underline mx-1"
                  @click="feedType = 'script'"
                >
                  {{ t('customScript') }}
                </button>
                {{ t('or') }}
                <button
                  type="button"
                  class="text-xs text-accent hover:underline mx-1"
                  @click="feedType = 'xpath'"
                >
                  {{ t('xpath') }}
                </button>
              </div>
            </div>
          </div>

          <!-- Script Selection (advanced mode) -->
          <div v-else-if="feedType === 'script'" key="script-mode" class="mb-3 sm:mb-4">
            <!-- Back to URL link -->
            <div class="mb-3 text-center">
              <button
                type="button"
                class="text-xs text-accent hover:underline transition-colors"
                @click="feedType = 'url'"
              >
                ← {{ t('backToUrl') }}
              </button>
            </div>

            <!-- Script Selection Component -->
            <ScriptSelector
              v-model="scriptPath"
              :mode="mode"
              :is-invalid="mode === 'add' && isScriptInvalid"
              :available-scripts="availableScripts"
              :scripts-dir="scriptsDir"
              @open-scripts-folder="openScriptsFolder"
            />

            <!-- Switch to other mode links -->
            <div class="mt-3 text-center">
              <div class="text-xs text-text-tertiary">
                {{ mode === 'add' ? t('orTry') : t('switchTo') }}
                <button
                  type="button"
                  class="text-xs text-accent hover:underline mx-1"
                  @click="feedType = 'url'"
                >
                  {{ t('rssUrl') }}
                </button>
                {{ t('or') }}
                <button
                  type="button"
                  class="text-xs text-accent hover:underline mx-1"
                  @click="feedType = 'xpath'"
                >
                  {{ t('xpath') }}
                </button>
              </div>
            </div>
          </div>

          <!-- XPath Configuration (advanced mode) -->
          <div v-else-if="feedType === 'xpath'" key="xpath-mode" class="mb-3 sm:mb-4">
            <!-- Back to URL link -->
            <div class="mb-3 text-center">
              <button
                type="button"
                class="text-xs text-accent hover:underline transition-colors"
                @click="feedType = 'url'"
              >
                ← {{ t('backToUrl') }}
              </button>
            </div>

            <!-- XPath Configuration Component -->
            <XPathConfig
              :mode="mode"
              :url="url"
              :xpath-type="xpathType"
              :xpath-item="xpathItem"
              :xpath-item-title="xpathItemTitle"
              :xpath-item-content="xpathItemContent"
              :xpath-item-uri="xpathItemUri"
              :xpath-item-author="xpathItemAuthor"
              :xpath-item-timestamp="xpathItemTimestamp"
              :xpath-item-time-format="xpathItemTimeFormat"
              :xpath-item-thumbnail="xpathItemThumbnail"
              :xpath-item-categories="xpathItemCategories"
              :xpath-item-uid="xpathItemUid"
              :is-xpath-item-invalid="mode === 'add' && isXpathItemInvalid"
              @update:url="url = $event"
              @update:xpath-type="xpathType = $event as 'HTML+XPath' | 'XML+XPath'"
              @update:xpath-item="xpathItem = $event"
              @update:xpath-item-title="xpathItemTitle = $event"
              @update:xpath-item-content="xpathItemContent = $event"
              @update:xpath-item-uri="xpathItemUri = $event"
              @update:xpath-item-author="xpathItemAuthor = $event"
              @update:xpath-item-timestamp="xpathItemTimestamp = $event"
              @update:xpath-item-time-format="xpathItemTimeFormat = $event"
              @update:xpath-item-thumbnail="xpathItemThumbnail = $event"
              @update:xpath-item-categories="xpathItemCategories = $event"
              @update:xpath-item-uid="xpathItemUid = $event"
            />

            <!-- Switch to other mode links -->
            <div class="mt-3 text-center">
              <div class="text-xs text-text-tertiary">
                {{ mode === 'add' ? t('orTry') : t('switchTo') }}
                <button
                  type="button"
                  class="text-xs text-accent hover:underline mx-1"
                  @click="feedType = 'url'"
                >
                  {{ t('rssUrl') }}
                </button>
                {{ t('or') }}
                <button
                  type="button"
                  class="text-xs text-accent hover:underline mx-1"
                  @click="feedType = 'script'"
                >
                  {{ t('customScript') }}
                </button>
              </div>
            </div>
          </div>
        </Transition>

        <CategorySelector
          :category="category"
          :category-selection="categorySelection"
          :show-custom-category="showCustomCategory"
          :existing-categories="existingCategories"
          @update:category="category = $event"
          @update:category-selection="categorySelection = $event"
          @update:show-custom-category="showCustomCategory = $event"
          @handle-category-change="handleCategoryChange"
        />

        <!-- Advanced Settings Toggle -->
        <div class="mb-3 sm:mb-4">
          <button
            type="button"
            class="flex items-center gap-1 text-xs sm:text-sm text-accent hover:text-accent-hover transition-colors"
            @click="showAdvancedSettings = !showAdvancedSettings"
          >
            <PhCaretRight
              v-if="!showAdvancedSettings"
              :size="12"
              class="transition-transform duration-200"
            />
            <PhCaretDown v-else :size="12" class="transition-transform duration-200" />
            <span class="hover:underline">
              {{ showAdvancedSettings ? t('hideAdvancedSettings') : t('showAdvancedSettings') }}
            </span>
          </button>
        </div>

        <!-- Advanced Settings Section (Collapsible) -->
        <AdvancedSettings
          v-if="showAdvancedSettings"
          :image-gallery-enabled="imageGalleryEnabled"
          :is-image-mode="isImageMode"
          :hide-from-timeline="hideFromTimeline"
          :proxy-mode="proxyMode"
          :proxy-type="proxyType"
          :proxy-host="proxyHost"
          :proxy-port="proxyPort"
          :proxy-username="proxyUsername"
          :proxy-password="proxyPassword"
          :refresh-mode="refreshMode"
          :refresh-interval="refreshInterval"
          @update:is-image-mode="isImageMode = $event"
          @update:hide-from-timeline="hideFromTimeline = $event"
          @update:proxy-mode="proxyMode = $event"
          @update:proxy-type="proxyType = $event"
          @update:proxy-host="proxyHost = $event"
          @update:proxy-port="proxyPort = $event"
          @update:proxy-username="proxyUsername = $event"
          @update:proxy-password="proxyPassword = $event"
          @update:refresh-mode="refreshMode = $event"
          @update:refresh-interval="refreshInterval = $event"
        />
      </div>
      <div class="p-3 sm:p-5 border-t border-border bg-bg-secondary text-right shrink-0">
        <button
          :disabled="isSubmitting || !isFormValid"
          class="btn-primary text-sm sm:text-base"
          @click="submit"
        >
          {{
            isSubmitting
              ? mode === 'add'
                ? t('adding')
                : t('saving')
              : mode === 'add'
                ? t('addSubscription')
                : t('saveChanges')
          }}
        </button>
      </div>
    </div>
  </div>

  <!-- RSSHub Route Selector Modal -->
  <RSSHubRouteSelectorModal
    :visible="showRouteSelector"
    :original-url="rsshubOriginalUrl"
    :matches="rsshubMatches"
    @close="closeRouteSelector"
    @select="handleRouteSelect"
  />
</template>

<style scoped>
@reference "../../../style.css";

.input-field {
  @apply w-full p-2 sm:p-2.5 border border-border rounded-md bg-bg-tertiary text-text-primary text-xs sm:text-sm focus:border-accent focus:outline-none transition-colors;
}
.btn-primary {
  @apply bg-accent text-white border-none px-4 sm:px-5 py-2 sm:py-2.5 rounded-lg cursor-pointer font-semibold hover:bg-accent-hover transition-colors disabled:opacity-70;
}
.toggle {
  @apply w-10 h-5 appearance-none bg-bg-tertiary rounded-full relative cursor-pointer border border-border transition-colors checked:bg-accent checked:border-accent shrink-0;
}
.toggle::after {
  content: '';
  @apply absolute top-0.5 left-0.5 w-3.5 h-3.5 bg-white rounded-full shadow-sm transition-transform;
}
.toggle:checked::after {
  transform: translateX(20px);
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
</style>
