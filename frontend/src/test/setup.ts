// src/test/setup.ts
import { vi } from 'vitest';

// Mock window.matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: (query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: () => {},
    removeListener: () => {},
    addEventListener: () => {},
    removeEventListener: () => {},
    dispatchEvent: () => {},
  }),
});

// Mock fetch only in test environment
const originalFetch = global.fetch;
global.fetch = vi.fn((input: RequestInfo | URL, init?: RequestInit): Promise<Response> => {
  // Convert input to string for URL matching
  const url = typeof input === 'string' ? input : input instanceof URL ? input.href : input.url;

  // Helper to create mock Response
  const createMockResponse = (data: any): Response => ({
    ok: true,
    status: 200,
    statusText: 'OK',
    headers: new Headers({ 'content-type': 'application/json' }),
    json: () => Promise.resolve(data),
    text: () => Promise.resolve(JSON.stringify(data)),
    bytes: () => Promise.resolve(new Uint8Array(new TextEncoder().encode(JSON.stringify(data)))),
    clone: () => createMockResponse(data),
    body: null,
    bodyUsed: false,
    arrayBuffer: () => Promise.resolve(new ArrayBuffer(0)),
    blob: () => Promise.resolve(new Blob()),
    formData: () => Promise.resolve(new FormData()),
    redirected: false,
    type: 'basic' as ResponseType,
    url: url,
  });

  // Mock successful responses for common API calls
  if (url === '/api/settings') {
    return Promise.resolve(
      createMockResponse({
        theme: 'light',
        update_interval: '30',
        last_global_refresh: '',
        auto_update: false,
        shortcuts: '{}',
        image_gallery_enabled: 'false',
        translation_enabled: 'false',
        target_language: 'en',
        show_article_preview_images: 'false',
        default_view_mode: 'original',
      })
    );
  }

  if (url === '/api/progress') {
    return Promise.resolve(
      createMockResponse({
        is_running: false,
        current: 0,
        total: 0,
        message: '',
      })
    );
  }

  if (url === '/api/window/state') {
    return Promise.resolve(
      createMockResponse({
        width: 1200,
        height: 800,
        x: 100,
        y: 100,
        maximized: false,
      })
    );
  }

  // For any other URLs, fall back to original fetch if available
  // This ensures tests can still make real HTTP calls if needed
  if (originalFetch) {
    return originalFetch(input, init);
  }

  // Default mock response for unknown URLs
  return Promise.resolve(createMockResponse({}));
});

// Mock window.showToast, window.showConfirm, etc.
Object.defineProperty(window, 'showToast', {
  writable: true,
  value: () => {},
});

Object.defineProperty(window, 'showConfirm', {
  writable: true,
  value: () => Promise.resolve(true),
});

// Mock ResizeObserver
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};
