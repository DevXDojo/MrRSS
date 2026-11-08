import { Call } from '@wailsio/runtime';

// Native capabilities are never fetched from the loopback/server HTTP API.
export async function updateRequestHeaders(): Promise<Record<string, string>> {
  const token = await Call.ByName('main.DesktopCapabilities.UpdateAuthorization');
  if (typeof token !== 'string' || !token) throw new Error('Desktop update capability unavailable');
  return { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` };
}
