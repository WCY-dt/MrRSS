import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue';
import { useAppStore } from '@/stores/app';
import { useI18n } from 'vue-i18n';
import { openInBrowser } from '@/utils/browser';
import type { Article } from '@/types/models';
import { proxyImagesInHtml, isMediaCacheEnabled } from '@/utils/mediaProxy';

type ViewMode = 'original' | 'rendered';
type RenderAction = 'showContent' | 'showOriginal' | null;

interface ViewModeChangeEvent extends Event {
  detail: {
    mode: ViewMode;
  };
}

interface RenderActionEvent extends Event {
  detail: {
    action: RenderAction;
  };
}

export function useArticleDetail() {
  const store = useAppStore();
  const { t, locale } = useI18n();

  const article = computed<Article | undefined>(() =>
    store.articles.find((a) => a.id === store.currentArticleId)
  );

  // Get effective view mode based on feed settings and global settings
  function getEffectiveViewMode(): ViewMode {
    if (!article.value) return defaultViewMode.value;

    // Find the feed for this article
    const feed = store.feeds.find((f) => f.id === article.value!.feed_id);
    if (!feed) return defaultViewMode.value;

    // Check feed's article_view_mode
    if (feed.article_view_mode === 'webpage') {
      return 'original';
    } else if (feed.article_view_mode === 'rendered') {
      return 'rendered';
    } else {
      // 'global' or undefined - use global setting
      return defaultViewMode.value;
    }
  }

  const showContent = ref(false);
  const articleContent = ref('');
  const isLoadingContent = ref(false);
  const currentArticleId = ref<number | null>(null);
  const defaultViewMode = ref<ViewMode>('original');
  const pendingRenderAction = ref<RenderAction>(null);
  const userPreferredMode = ref<ViewMode | null>(null); // Remember user's manual choice
  const imageViewerSrc = ref<string | null>(null);
  const imageViewerAlt = ref('');

  // Watch for article changes and apply view mode
  watch(
    () => store.currentArticleId,
    async (newId, oldId) => {
      if (newId && newId !== oldId) {
        // Reset content when switching articles
        articleContent.value = '';
        currentArticleId.value = null;

        // Always fetch article content for AI chat and translation features
        await fetchArticleContent();

        // Check if there's a pending render action from context menu
        if (pendingRenderAction.value) {
          // Apply the explicit action instead of default
          if (pendingRenderAction.value === 'showContent') {
            showContent.value = true;
            userPreferredMode.value = 'rendered';
          } else if (pendingRenderAction.value === 'showOriginal') {
            showContent.value = false;
            userPreferredMode.value = 'original';
          }
          pendingRenderAction.value = null; // Clear the pending action
        } else {
          // Apply user's preferred mode or determine from feed/global settings
          const preferredMode = userPreferredMode.value || getEffectiveViewMode();
          showContent.value = preferredMode === 'rendered';
        }
      }
    }
  );

  // Listen for default view mode changes from settings
  window.addEventListener('default-view-mode-changed', (e: Event) => {
    const event = e as ViewModeChangeEvent;
    defaultViewMode.value = event.detail.mode;
    // Reset user preference when default changes
    userPreferredMode.value = null;
  });

  function close() {
    store.currentArticleId = null;
    showContent.value = false;
    articleContent.value = '';
    currentArticleId.value = null;
  }

  function toggleRead() {
    if (!article.value) return;
    const newState = !article.value.is_read;
    article.value.is_read = newState;
    fetch(`/api/articles/read?id=${article.value.id}&read=${newState}`, { method: 'POST' });
  }

  function toggleFavorite() {
    if (!article.value) return;
    const newState = !article.value.is_favorite;
    article.value.is_favorite = newState;
    fetch(`/api/articles/favorite?id=${article.value.id}`, { method: 'POST' });
  }

  async function toggleReadLater() {
    if (!article.value) return;
    const newState = !article.value.is_read_later;
    article.value.is_read_later = newState;
    // When adding to read later, also mark as unread
    if (newState) {
      article.value.is_read = false;
    }
    try {
      await fetch(`/api/articles/toggle-read-later?id=${article.value.id}`, { method: 'POST' });
      store.fetchUnreadCounts();
    } catch (e) {
      console.error('Error toggling read later:', e);
      // Revert on error
      article.value.is_read_later = !newState;
    }
  }

  function openOriginal() {
    if (article.value) openInBrowser(article.value.url);
  }

  async function toggleContentView() {
    if (!showContent.value) {
      // Switching to content view - fetch content if needed
      if (!article.value) return;
      // Check if we need to fetch content (different article or no content yet)
      if (currentArticleId.value !== article.value.id) {
        await fetchArticleContent();
      }
    }
    showContent.value = !showContent.value;
    // Remember user's preference
    userPreferredMode.value = showContent.value ? 'rendered' : 'original';
  }

  async function fetchArticleContent() {
    if (!article.value) return;

    isLoadingContent.value = true;
    currentArticleId.value = article.value.id; // Track which article we're loading
    try {
      const res = await fetch(`/api/articles/content?id=${article.value.id}`);
      if (res.ok) {
        const data = await res.json();
        let content = data.content || '';

        // Proxy images if media cache is enabled
        const cacheEnabled = await isMediaCacheEnabled();
        if (cacheEnabled && content) {
          // Use feed URL as referer for anti-hotlinking (more reliable than article URL)
          const feedUrl = data.feed_url || article.value.url;
          content = proxyImagesInHtml(content, feedUrl);
        }

        articleContent.value = content;
        // Don't attach event listeners here - let ArticleContent.vue handle it
        // after enhanceRendering is complete
      } else {
        console.error('Failed to fetch article content');
        articleContent.value = '';
      }
    } catch (e) {
      console.error('Error fetching article content:', e);
      articleContent.value = '';
    } finally {
      isLoadingContent.value = false;
    }
  }

  // Handle retry loading content
  function handleRetryLoadContent() {
    if (article.value && showContent.value) {
      fetchArticleContent();
    }
  }

  // Unwrap images from hyperlinks
  // This ensures images can be clicked directly without triggering link navigation
  // Works on both main content and translated content
  function unwrapImagesFromLinks() {
    // Process all links in prose content (both main content and translations)
    const links = document.querySelectorAll<HTMLAnchorElement>('.prose-content a, .prose a');
    const linksToProcess: HTMLAnchorElement[] = [];

    // Collect links that contain images (check both direct children and nested)
    links.forEach((link) => {
      const images = link.querySelectorAll('img');
      if (images.length > 0) {
        linksToProcess.push(link);
      }
    });

    // Process collected links
    linksToProcess.forEach((link) => {
      try {
        // Check if parent exists and is valid
        if (!link.parentNode) {
          return;
        }

        // Extract all child nodes from the link
        const fragment = document.createDocumentFragment();
        while (link.firstChild) {
          fragment.appendChild(link.firstChild);
        }

        // Replace the link with its contents
        link.parentNode.replaceChild(fragment, link);
      } catch (error) {
        console.error('Error unwrapping image from link:', error);
      }
    });
  }

  // Attach event listeners to images in rendered content
  // Can be called multiple times (e.g., after translations modify the DOM)
  function attachImageEventListeners() {
    // First, unwrap any images that are inside hyperlinks
    unwrapImagesFromLinks();

    // Get all images in prose content (use more specific selector)
    const proseContainers = document.querySelectorAll('.prose-content, .prose');

    if (proseContainers.length === 0) {
      return;
    }

    const images = document.querySelectorAll<HTMLImageElement>('.prose-content img, .prose img');

    // Process images if there are any
    if (images.length > 0) {
      images.forEach((img) => {
        try {
          // Verify the image has a valid parent
          if (!img.parentNode) {
            return;
          }

          // Skip images that are very small (likely icons/emojis)
          const isSmallIcon = img.height <= 24 && img.height > 0;
          if (isSmallIcon) {
            return;
          }

          // Ensure the image and its parents can receive pointer events
          img.style.cursor = 'pointer';
          img.style.pointerEvents = 'auto';

          // Remove old listeners by replacing with clone
          const newImg = img.cloneNode(true) as HTMLImageElement;
          img.parentNode.replaceChild(newImg, img);

          // Ensure cloned image maintains pointer interaction styles
          newImg.style.cursor = 'pointer';
          newImg.style.pointerEvents = 'auto';

          // Left click - open image viewer
          newImg.addEventListener(
            'click',
            (e: Event) => {
              e.preventDefault();
              e.stopPropagation(); // Prevent event bubbling to parent elements

              // Verify image src exists
              if (!newImg.src) {
                return;
              }

              imageViewerSrc.value = newImg.src;
              imageViewerAlt.value = newImg.alt || '';
            },
            { capture: true }
          ); // Use capture phase to ensure we get the event first

          // Right click - show context menu for saving
          newImg.addEventListener(
            'contextmenu',
            (e: MouseEvent) => {
              e.preventDefault();
              e.stopPropagation(); // Prevent event bubbling to parent elements

              // Verify image src exists
              if (!newImg.src) {
                return;
              }

              // Use global context menu system
              window.dispatchEvent(
                new CustomEvent('open-context-menu', {
                  detail: {
                    x: e.clientX,
                    y: e.clientY,
                    items: [
                      {
                        label: t('viewImage'),
                        action: 'view',
                        icon: 'PhMagnifyingGlassPlus',
                      },
                      {
                        label: t('downloadImage'),
                        action: 'download',
                        icon: 'PhDownloadSimple',
                      },
                    ],
                    data: { src: newImg.src },
                    callback: (action: string, data: { src: string }) => {
                      if (action === 'view') {
                        imageViewerSrc.value = data.src;
                        imageViewerAlt.value = '';
                      } else if (action === 'download') {
                        downloadImage(data.src);
                      }
                    },
                  },
                })
              );
            },
            { capture: true }
          ); // Use capture phase to ensure we get the event first
        } catch (error) {
          console.error('Error attaching event listeners to image:', error);
        }
      });
    }

    // Always attach link event listeners (even if there are no images)
    attachLinkEventListeners();
  }

  // Attach event listeners to links in rendered content
  // Uses the same strategy as image handling for consistency
  // Works for dynamically added content (e.g., translations)
  function attachLinkEventListeners() {
    // Get all text-only links (no images) in prose content
    const links = document.querySelectorAll<HTMLAnchorElement>('.prose-content a, .prose a');

    links.forEach((link) => {
      try {
        // Verify the link has a valid parent
        if (!link.parentNode) {
          return;
        }

        // Skip if the link contains an image (handled separately by image handlers)
        if (link.querySelector('img')) {
          return;
        }

        // Check if already processed (has our marker)
        if (link.dataset.linkHandlerAttached === 'true') {
          return;
        }

        // Mark as processed
        link.dataset.linkHandlerAttached = 'true';

        // Add event listener using capture phase
        link.addEventListener(
          'click',
          (e: Event) => {
            // Prevent default behavior and stop propagation
            e.preventDefault();
            e.stopPropagation();
            e.stopImmediatePropagation();

            // Get the href
            let href = link.getAttribute('href');
            if (href) {
              // Convert relative URLs to absolute URLs
              // If it starts with / or is a relative path, convert to absolute using article URL
              if (
                href.startsWith('/') ||
                (!href.startsWith('http://') &&
                  !href.startsWith('https://') &&
                  !href.startsWith('mailto:') &&
                  !href.startsWith('#'))
              ) {
                if (article.value?.url) {
                  try {
                    const articleUrl = new URL(article.value.url);
                    // For relative paths like "path/to/page"
                    if (!href.startsWith('/')) {
                      href = new URL(href, article.value.url).href;
                    } else {
                      // For absolute paths like "/path/to/page"
                      href = `${articleUrl.origin}${href}`;
                    }
                  } catch (error) {
                    console.error('Error converting relative URL:', error);
                  }
                }
              }

              try {
                openInBrowser(href);
              } catch (error) {
                console.error('Error opening URL:', href, error);
              }
            }
          },
          { capture: true } // Use capture phase to intercept clicks early
        );
      } catch (error) {
        console.error('Error attaching event listeners to link:', error);
      }
    });
  }

  function closeImageViewer() {
    imageViewerSrc.value = null;
    imageViewerAlt.value = '';
  }

  // Download image from URL
  async function downloadImage(src: string) {
    try {
      const response = await fetch(src);
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const blob = await response.blob();

      // Extract and sanitize filename from URL
      let filename = 'image';
      try {
        const url = new URL(src);
        const pathname = url.pathname;
        const pathSegments = pathname.split('/').filter((segment) => segment.length > 0);
        if (pathSegments.length > 0) {
          const lastSegment = pathSegments[pathSegments.length - 1];
          filename = lastSegment.split('?')[0].replace(/[^a-zA-Z0-9._-]/g, '_') || 'image';
        }
      } catch {
        filename = 'image';
      }

      // Ensure it has a valid extension based on MIME type
      if (!filename.match(/\.(jpg|jpeg|png|gif|webp|svg|bmp)$/i)) {
        const mimeType = blob.type;
        const ext = mimeType.split('/')[1]?.replace('jpeg', 'jpg') || 'png';
        filename = `${filename}.${ext}`;
      }

      // Create download link
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = filename;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Failed to download image:', error);
      window.open(src, '_blank');
    }
  }

  // Export article to Obsidian
  async function exportToObsidian() {
    if (!article.value) return;

    try {
      window.showToast(t('exportingToObsidian'), 'info');

      const response = await fetch('/api/articles/export/obsidian', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          article_id: article.value.id,
        }),
      });

      if (!response.ok) {
        const error = await response.text();
        throw new Error(error);
      }

      const data = await response.json();

      // Show success message with file path
      const message = data.message || t('exportedToObsidian');
      const filePath = data.file_path ? ` (${data.file_path})` : '';
      window.showToast(message + filePath, 'success');
    } catch (error) {
      console.error('Failed to export to Obsidian:', error);
      const message = error instanceof Error ? error.message : t('obsidianExportFailed');
      window.showToast(message, 'error');
    }
  }

  // Listen for render content event from context menu
  async function handleRenderContent(e: Event) {
    const event = e as RenderActionEvent;
    if (!article.value) return;

    const action = event.detail?.action || 'showContent';

    // Mark as read when rendering content
    if (!article.value.is_read) {
      article.value.is_read = true;
      fetch(`/api/articles/read?id=${article.value.id}&read=true`, { method: 'POST' });
    }

    if (action === 'showContent') {
      // Check if we need to fetch content for this article
      if (currentArticleId.value !== article.value.id) {
        await fetchArticleContent();
      }
      showContent.value = true;
      userPreferredMode.value = 'rendered';
    } else if (action === 'showOriginal') {
      showContent.value = false;
      userPreferredMode.value = 'original';
    }
  }

  // Listen for explicit render action from context menu (before article selection)
  function handleExplicitRenderAction(e: Event) {
    const event = e as RenderActionEvent;
    pendingRenderAction.value = event.detail?.action;
  }

  // Handle toggle content view from keyboard shortcut
  function handleToggleContentView() {
    if (article.value) {
      toggleContentView();
    }
  }

  // Handle reset user preference from normal article selection
  function handleResetUserPreference() {
    userPreferredMode.value = null;
  }

  onMounted(async () => {
    window.addEventListener('render-article-content', handleRenderContent);
    window.addEventListener('explicit-render-action', handleExplicitRenderAction);
    window.addEventListener('toggle-content-view', handleToggleContentView);
    window.addEventListener('reset-user-view-preference', handleResetUserPreference);

    // Load default view mode from settings
    try {
      const res = await fetch('/api/settings');
      const data = await res.json();
      defaultViewMode.value = data.default_view_mode || 'original';
    } catch (e) {
      console.error('Error loading settings:', e);
    }
  });

  onBeforeUnmount(() => {
    window.removeEventListener('render-article-content', handleRenderContent);
    window.removeEventListener('explicit-render-action', handleExplicitRenderAction);
    window.removeEventListener('toggle-content-view', handleToggleContentView);
    window.removeEventListener('reset-user-view-preference', handleResetUserPreference);
  });

  return {
    // Reactive state
    article,
    showContent,
    articleContent,
    isLoadingContent,
    imageViewerSrc,
    imageViewerAlt,
    locale,

    // Functions
    close,
    toggleRead,
    toggleFavorite,
    toggleReadLater,
    openOriginal,
    toggleContentView,
    closeImageViewer,
    downloadImage,
    exportToObsidian,
    attachImageEventListeners, // Expose for re-attaching after content modifications
    handleRetryLoadContent,

    // Translations
    t,
  };
}
