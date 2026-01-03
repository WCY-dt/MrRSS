<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { PhCaretDown, PhCaretRight } from '@phosphor-icons/vue';
import type { Feed } from '@/types/models';
import { useModalClose } from '@/composables/ui/useModalClose';
import { useFeedForm } from '@/composables/feed/useFeedForm';
import UrlInput from './parts/UrlInput.vue';
import ScriptSelector from './parts/ScriptSelector.vue';
import XPathConfig from './parts/XPathConfig.vue';
import EmailConfig from './parts/EmailConfig.vue';
import CategorySelector from './parts/CategorySelector.vue';
import AdvancedSettings from './parts/AdvancedSettings.vue';

interface Props {
  mode: 'add' | 'edit';
  feed?: Feed;
}

const props = defineProps<Props>();

const { t } = useI18n();

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
  articleViewMode,
  proxyMode,
  proxyType,
  proxyHost,
  proxyPort,
  proxyUsername,
  proxyPassword,
  refreshMode,
  refreshInterval,
  autoExpandContent,
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
  // Email fields
  emailAddress,
  imapServer,
  imapPort,
  emailUsername,
  emailPassword,
  emailFolder,
} = useFeedForm(props.feed);

const emit = defineEmits<{
  close: [];
  added: [];
  updated: [];
}>();

// Modal close handling
useModalClose(() => close());

function close() {
  emit('close');
}

function insertRSSHubPrefix() {
  url.value = 'rsshub://';
}

async function submit() {
  if (!isFormValid.value) {
    return;
  }
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
    } else if (feedType.value === 'email') {
      body.type = 'email';
      body.email_address = emailAddress.value;
      body.email_imap_server = imapServer.value;
      body.email_imap_port = imapPort.value;
      body.email_username = emailUsername.value;
      body.email_password = emailPassword.value;
      body.email_folder = emailFolder.value;
    }

    // Add article view mode
    body.article_view_mode = articleViewMode.value;

    // Add auto expand content mode
    body.auto_expand_content = autoExpandContent.value;

    if (props.mode === 'edit') {
      body.id = props.feed!.id;
    }

    // Special handling for RSSHub URLs - use dedicated endpoint
    if (body.url && typeof body.url === 'string' && body.url.startsWith('rsshub://')) {
      const route = body.url.replace('rsshub://', '');

      try {
        // Use the specialized RSSHub add endpoint
        const rsshubResp = await fetch('/api/rsshub/add', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            route: route,
            category: category.value,
            title: title.value,
          }),
        });

        if (rsshubResp.ok) {
          const rsshubResult = await rsshubResp.json();
          if (rsshubResult.success) {
            emit('added');
            resetForm();
            window.showToast(t('feedAddedSuccess'), 'success');
            close();
            return;
          }
        }

        // If RSSHub endpoint failed, try the generic add endpoint as fallback
        const errorText = await rsshubResp.text();
        throw new Error(errorText || 'RSSHub add failed');
      } catch (e) {
        console.error('RSSHub add failed:', e);
        window.showToast(
          `${t('errorAddingFeed')}: ${e instanceof Error ? e.message : 'Unknown error'}`,
          'error'
        );
        isSubmitting.value = false;
        return;
      }
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
      // Read error message from response
      const errorText = await res.text();

      // Check if it's an XPath error for better display
      if (feedType.value === 'xpath' && errorText.includes('XPath')) {
        // For XPath errors, show a more detailed toast
        const errorKey = props.mode === 'add' ? 'errorAddingFeed' : 'errorUpdatingFeed';
        const title = t(errorKey);
        let formattedMessage = `${title}:\n${errorText}`;

        window.showToast(formattedMessage, 'error', 8000); // Show for 8 seconds for XPath errors
      } else {
        // For other errors, show the standard format
        const errorKey = props.mode === 'add' ? 'errorAddingFeed' : 'errorUpdatingFeed';
        window.showToast(`${t(errorKey)}: ${errorText}`, 'error');
      }
    }
  } catch {
    const errorKey = props.mode === 'add' ? 'errorAddingFeed' : 'errorUpdatingFeed';
    window.showToast(t(errorKey), 'error');
  } finally {
    isSubmitting.value = false;
  }
}
</script>

<template>
  <div
    class="fixed inset-0 z-[60] flex items-center justify-center bg-black/50 backdrop-blur-sm p-2 sm:p-4"
    data-modal-open="true"
    style="will-change: transform; transform: translateZ(0)"
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
      <div class="flex-1 overflow-y-auto p-4 sm:p-6 scroll-smooth">
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
        <!-- URL Input (default mode) -->
        <div v-if="feedType === 'url'" key="url-mode" class="mb-3 sm:mb-4">
          <UrlInput v-model="url" :mode="mode" :is-invalid="mode === 'add' && isUrlInvalid" />

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
              {{ t('or') }}
              <button
                type="button"
                class="text-xs text-accent hover:underline mx-1"
                @click="feedType = 'email'"
              >
                {{ t('emailNewsletter') }}
              </button>
              {{ t('or') }}
              <button
                type="button"
                class="text-xs text-accent hover:underline mx-1"
                @click="insertRSSHubPrefix"
              >
                RSSHub
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
              {{ t('or') }}
              <button
                type="button"
                class="text-xs text-accent hover:underline mx-1"
                @click="feedType = 'email'"
              >
                {{ t('emailNewsletter') }}
              </button>
              {{ t('or') }}
              <button
                type="button"
                class="text-xs text-accent hover:underline mx-1"
                @click="insertRSSHubPrefix"
              >
                RSSHub
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
              {{ t('or') }}
              <button
                type="button"
                class="text-xs text-accent hover:underline mx-1"
                @click="feedType = 'email'"
              >
                {{ t('emailNewsletter') }}
              </button>
            </div>
          </div>
        </div>

        <!-- Email Configuration (newsletter mode) -->
        <div v-else-if="feedType === 'email'" key="email-mode" class="mb-3 sm:mb-4">
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

          <!-- Email Configuration Component -->
          <EmailConfig
            :mode="mode"
            :email-address="emailAddress"
            :imap-server="imapServer"
            :imap-port="imapPort"
            :username="emailUsername"
            :password="emailPassword"
            :folder="emailFolder"
            @update:email-address="emailAddress = $event"
            @update:imap-server="imapServer = $event"
            @update:imap-port="imapPort = $event"
            @update:username="emailUsername = $event"
            @update:password="emailPassword = $event"
            @update:folder="emailFolder = $event"
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
          :article-view-mode="articleViewMode"
          :auto-expand-content="autoExpandContent"
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
          @update:article-view-mode="articleViewMode = $event"
          @update:auto-expand-content="autoExpandContent = $event"
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
