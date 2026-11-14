import { afterEach, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { Call } from '@wailsio/runtime';
import { useAppStore } from './app';
import { updateRequestHeaders } from '@/utils/updateAuthorization';

vi.mock('vue-i18n', () => ({ useI18n: () => ({ locale: { value: 'en' } }) }));
// Exercise the legacy automatic-install branch; the normal settings UI only checks for updates.
vi.mock('@/composables/core/useSettings', () => ({
  useSettings: () => ({ settings: { value: { auto_update: true } } }),
}));
vi.mock('@wailsio/runtime', () => ({ Call: { ByName: vi.fn(async () => 'desktop-test-token') } }));
afterEach(() => {
  vi.useRealTimers();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

it('authorizes both automatic download and installation through the native capability', async () => {
  setActivePinia(createPinia());
  vi.useFakeTimers();
  const fetchMock = vi.fn(
    async (url: string, _options?: RequestInit) =>
      new Response(
        JSON.stringify(
          url === '/api/check-updates'
            ? {
                has_update: true,
                download_url: 'https://github.com/DevXDojo/MrRSS/releases/download/v9/MrRSS.exe',
                asset_name: 'MrRSS.exe',
              }
            : { success: true, file_path: 'verified-download' }
        )
      )
  );
  vi.stubGlobal('fetch', fetchMock);
  await useAppStore().checkForAppUpdates();
  await vi.advanceTimersByTimeAsync(1000);
  for (const url of ['/api/download-update', '/api/install-update']) {
    const call = fetchMock.mock.calls.find(([request]) => request === url);
    expect(call?.[1]?.headers).toEqual({
      'Content-Type': 'application/json',
      Authorization: 'Bearer desktop-test-token',
    });
  }
  expect(Call.ByName).toHaveBeenCalledWith(
    'MrRSS/internal/desktopapi.DesktopCapabilities.UpdateAuthorization'
  );
});

it('rejects missing native capabilities before an update request is made', async () => {
  vi.mocked(Call.ByName).mockResolvedValueOnce('');
  await expect(updateRequestHeaders()).rejects.toThrow('Desktop update capability unavailable');
});
