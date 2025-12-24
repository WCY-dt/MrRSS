/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue';
  const component: DefineComponent<Record<string, unknown>, Record<string, unknown>, unknown>;
  export default component;
}

// Wails3 runtime type definitions
declare global {
  interface Window {
    _wails?: {
      dispatchWailsEvent?: (event: { name: string; data: any }) => void;
      [key: string]: any;
    };
  }
}
