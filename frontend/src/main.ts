import { createApp } from 'vue';
import { createPinia } from 'pinia';
import PhosphorIcons from '@phosphor-icons/vue';
import i18n, { locale } from './i18n';
import './style.css';
import App from './App.vue';

const app = createApp(App);
const pinia = createPinia();

app.use(pinia);
app.use(i18n);
app.use(PhosphorIcons);

// Polyfill for _wails.dispatchWailsEvent if not available
// This prevents errors when Wails3 internal events try to dispatch before runtime is ready
function setupWailsPolyfill() {
  if (typeof window === 'undefined') {
    return;
  }

  const wailsWindow = window as typeof window & {
    _wails?: {
      dispatchWailsEvent?: (event: { name: string; data: any }) => void;
      [key: string]: any;
    };
  };

  // Initialize _wails object if it doesn't exist
  if (!wailsWindow._wails) {
    wailsWindow._wails = {};
  }

  // Add dispatchWailsEvent polyfill if it doesn't exist
  if (!wailsWindow._wails.dispatchWailsEvent) {
    wailsWindow._wails.dispatchWailsEvent = (event: { name: string; data: any }) => {
      // Silently ignore events if runtime is not ready
      // This prevents console errors during initialization
      if (
        wailsWindow._wails &&
        typeof (wailsWindow._wails as any).__dispatchWailsEvent === 'function'
      ) {
        (wailsWindow._wails as any).__dispatchWailsEvent(event);
      }
      // Otherwise, just ignore the event (it's likely a timing issue)
    };
  }
}

// Wait for Wails runtime to be available
function waitForWailsRuntime(): Promise<void> {
  return new Promise((resolve) => {
    // Setup polyfill first to prevent errors
    setupWailsPolyfill();

    const wailsWindow = window as typeof window & {
      _wails?: {
        dispatchWailsEvent?: (event: { name: string; data: any }) => void;
        [key: string]: any;
      };
    };

    // Check if Wails runtime is already available
    if (typeof window !== 'undefined' && wailsWindow._wails) {
      resolve();
      return;
    }

    // Wait for Wails runtime to load (max 5 seconds)
    let attempts = 0;
    const maxAttempts = 50;
    const checkInterval = setInterval(() => {
      attempts++;
      if (typeof window !== 'undefined' && wailsWindow._wails) {
        clearInterval(checkInterval);
        resolve();
      } else if (attempts >= maxAttempts) {
        clearInterval(checkInterval);
        // Resolve anyway to avoid blocking app startup
        console.warn('Wails runtime not detected after timeout, continuing anyway');
        resolve();
      }
    }, 100);
  });
}

// Initialize language setting before mounting
async function initializeApp() {
  // Wait for Wails runtime to be available
  await waitForWailsRuntime();

  try {
    const res = await fetch('/api/settings');
    const data = await res.json();
    if (data.language) {
      locale.value = data.language;
    }
  } catch (e) {
    console.error('Error loading language setting:', e);
  }

  app.mount('#app');
}

// Initialize and mount
initializeApp();
