/**
 * Media proxy utilities for handling anti-hotlinking and caching
 */

// Cache for media cache enabled setting to avoid repeated API calls
let mediaCacheEnabledCache: boolean | null = null;
let mediaCachePromise: Promise<boolean> | null = null;

/**
 * Convert a media URL to use the proxy endpoint
 * @param url Original media URL
 * @param referer Optional referer URL for anti-hotlinking
 * @returns Proxied URL
 */
export function getProxiedMediaUrl(url: string, referer?: string): string {
  if (!url) return '';

  // Don't proxy data URLs or blob URLs
  if (url.startsWith('data:') || url.startsWith('blob:')) {
    return url;
  }

  // Don't proxy local URLs
  if (
    url.startsWith('/') ||
    url.startsWith('http://localhost') ||
    url.startsWith('http://127.0.0.1')
  ) {
    return url;
  }

  // Build proxy URL
  const params = new URLSearchParams();
  params.set('url', url);
  if (referer) {
    params.set('referer', referer);
  }

  return `/api/media/proxy?${params.toString()}`;
}

/**
 * Check if media caching is enabled (with caching to avoid repeated API calls)
 * @returns Promise<boolean>
 */
export async function isMediaCacheEnabled(): Promise<boolean> {
  // Return cached value if available
  if (mediaCacheEnabledCache !== null) {
    return mediaCacheEnabledCache;
  }

  // If a request is already in flight, wait for it
  if (mediaCachePromise) {
    return mediaCachePromise;
  }

  // Start a new request
  mediaCachePromise = (async () => {
    try {
      const response = await fetch('/api/settings');
      if (response.ok) {
        const settings = await response.json();
        mediaCacheEnabledCache =
          settings.media_cache_enabled === 'true' || settings.media_cache_enabled === true;
        return mediaCacheEnabledCache;
      }
    } catch (error) {
      console.error('Failed to check media cache status:', error);
    }
    mediaCacheEnabledCache = false;
    return false;
  })();

  const result = await mediaCachePromise;
  mediaCachePromise = null; // Clear the promise after completion
  return result;
}

/**
 * Clear the media cache enabled cache (call this when settings change)
 */
export function clearMediaCacheEnabledCache(): void {
  mediaCacheEnabledCache = null;
}

/**
 * Process HTML content to proxy image URLs
 * @param html HTML content
 * @param referer Optional referer URL
 * @returns HTML with proxied image URLs
 * @note Unquoted src attributes are supported but must not contain spaces (per HTML spec)
 */
export function proxyImagesInHtml(html: string, referer?: string): string {
  if (!html) return html;

  // Enhanced regex to handle img src attributes with better pattern matching
  // Handles double quotes, single quotes, and unquoted values
  // Note: Unquoted values cannot contain spaces per HTML specification
  const imgRegex = /<img([^>]+)src\s*=\s*(['"]?)([^"'\s>]+)\2/gi;

  return html.replace(imgRegex, (match, _attrs, quote, src) => {
    // CRITICAL FIX: Decode HTML entities before processing the URL
    // HTML attributes contain &amp; which should be decoded to & before URL encoding
    // For example: &amp; becomes &, then gets properly URL-encoded as %26
    const decodedSrc = decodeHTMLEntities(src);
    const proxiedUrl = getProxiedMediaUrl(decodedSrc, referer);

    // If proxying failed or returned the same URL, keep original
    if (!proxiedUrl || proxiedUrl === decodedSrc) {
      return match;
    }

    // Replace the src attribute, preserving the original quote style
    const newSrc = `src=${quote}${proxiedUrl}${quote}`;
    const srcRegex = new RegExp(
      `src\\s*=\\s*${quote}${src.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}${quote}`,
      'i'
    );

    return match.replace(srcRegex, newSrc);
  });
}

/**
 * Decode HTML entities in a string
 * Handles common entities like &amp;, &lt;, &gt;, &quot;, &#39;, etc.
 */
function decodeHTMLEntities(text: string): string {
  const textarea = document.createElement('textarea');
  textarea.innerHTML = text;
  return textarea.value;
}
