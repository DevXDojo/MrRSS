import { mount, flushPromises } from '@vue/test-utils';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { createI18n } from 'vue-i18n';
import NotificationsTab from './NotificationsTab.vue';
import ChannelEditor from './ChannelEditor.vue';
import RuleEditor from './RuleEditor.vue';
import { enNotifications, zhNotifications } from '@/i18n/locales/notifications';
import { newChannel, newRule } from '@/types/notification';
import type { NotificationConfig, NotificationPreview } from '@/types/notification';

const i18n = () =>
  createI18n({
    legacy: false,
    locale: 'en',
    messages: {
      en: { setting: { notifications: enNotifications }, common: { action: { cancel: 'Cancel' } } },
    },
  });
function config(): NotificationConfig {
  return {
    revision: 0,
    enabled: false,
    timezone: 'UTC',
    quiet_enabled: false,
    quiet_start: '22:00',
    quiet_end: '08:00',
    channels: [],
    rules: [],
  };
}
function button(wrapper: ReturnType<typeof mount>, text: string) {
  const result = wrapper.findAll('button').find((b) => b.text() === text);
  if (!result) throw new Error(`Missing button: ${text}`);
  return result;
}
afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('notification settings', () => {
  it('shows three local logos and does not allow rules before a destination exists', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async (url: string) => Response.json(url.endsWith('config') ? config() : []))
    );
    const wrapper = mount(NotificationsTab, { global: { plugins: [i18n()] } });
    await flushPromises();
    expect(wrapper.findAll('img').map((img) => img.attributes('alt'))).toEqual([
      'Telegram',
      'Discord',
      'Feishu / Lark',
    ]);
    expect(button(wrapper, 'Add alert').attributes('disabled')).toBeDefined();
    expect(wrapper.text()).toContain('First connect a destination');
    wrapper.unmount();
  });

  it('keeps incomplete channel edits local and retains them on save failure', async () => {
    window.showToast = vi.fn();
    const fetcher = vi.fn(async (url: string, options?: RequestInit) => {
      if (options?.method === 'PUT')
        return Response.json({ error: { message: 'telegram_credentials' } }, { status: 400 });
      return Response.json(url.endsWith('config') ? config() : []);
    });
    vi.stubGlobal('fetch', fetcher);
    const wrapper = mount(NotificationsTab, { global: { plugins: [i18n()] } });
    await flushPromises();
    await wrapper.findAll('.push-provider-heading')[0].trigger('click');
    await button(wrapper, 'Add destination').trigger('click');
    const editor = wrapper.findComponent(ChannelEditor);
    await editor.find('input[type="password"]').setValue('invalid');
    expect(fetcher.mock.calls.filter((c) => c[1]?.method === 'PUT')).toHaveLength(0);
    await editor.find('form').trigger('submit');
    await flushPromises();
    expect(wrapper.findComponent(ChannelEditor).exists()).toBe(true);
    expect(wrapper.text()).toContain('Check the bot token');
    await button(wrapper, 'Cancel').trigger('click');
    expect(wrapper.findComponent(ChannelEditor).exists()).toBe(false);
    wrapper.unmount();
  });

  it('does not overwrite other destinations when saving a new one', async () => {
    window.showToast = vi.fn();
    const initial = config();
    initial.channels = [{ ...newChannel('discord', 'Team'), webhook: '••••••••' }];
    let payload: NotificationConfig | undefined;
    vi.stubGlobal(
      'fetch',
      vi.fn(async (url: string, options?: RequestInit) => {
        if (options?.method === 'PUT') {
          payload = JSON.parse(String(options.body)) as NotificationConfig;
          return Response.json({ ...payload, revision: 1 });
        }
        return Response.json(url.endsWith('config') ? initial : []);
      })
    );
    const wrapper = mount(NotificationsTab, { global: { plugins: [i18n()] } });
    await flushPromises();
    await wrapper.findAll('.push-provider-heading')[0].trigger('click');
    await button(wrapper, 'Add destination').trigger('click');
    const editor = wrapper.findComponent(ChannelEditor);
    editor.vm.$emit('save', {
      ...newChannel('telegram', 'Personal'),
      token: '123:test',
      chat_id: '1',
    });
    await flushPromises();
    expect(payload?.channels.map((c) => c.name)).toEqual(['Team', 'Personal']);
    expect(payload?.channels[0].webhook).toBe('••••••••');
    expect(wrapper.findComponent(ChannelEditor).exists()).toBe(false);
    wrapper.unmount();
  });

  it('previews drafts without saving or sending and discards stale preview responses', async () => {
    let resolve!: (value: NotificationPreview) => void;
    const preview = vi.fn(
      () =>
        new Promise<NotificationPreview>((r) => {
          resolve = r;
        })
    );
    const rule = newRule('digest', 'Morning', ['channel']);
    const wrapper = mount(RuleEditor, {
      props: { rule, channels: [], saving: false, preview, errorText: (s: string) => s },
      global: { plugins: [i18n()] },
    });
    await button(wrapper, 'Preview matches').trigger('click');
    await wrapper.find('input').setValue('Evening');
    resolve({ matched: 1, scanned: 1, messages: ['stale private article'] });
    await flushPromises();
    expect(wrapper.text()).not.toContain('stale private article');
    expect(wrapper.emitted('save')).toBeUndefined();
    expect(rule.name).toBe('Morning');
    wrapper.unmount();
  });

  it('shows matched text as text, including hostile feed HTML', async () => {
    const wrapper = mount(RuleEditor, {
      props: {
        rule: newRule('alert', 'Alert', []),
        channels: [],
        saving: false,
        preview: async () => ({
          matched: 1,
          scanned: 1,
          messages: ['<img src=x onerror=alert(1)>'],
        }),
        errorText: (s: string) => s,
      },
      global: { plugins: [i18n()] },
    });
    await button(wrapper, 'Preview matches').trigger('click');
    await flushPromises();
    expect(wrapper.find('.push-preview img').exists()).toBe(false);
    expect(wrapper.find('pre').text()).toContain('<img');
    wrapper.unmount();
  });

  it('has matching English and Chinese translation keys', () => {
    function keys(value: object, prefix = ''): string[] {
      return Object.entries(value)
        .flatMap(([k, v]) =>
          typeof v === 'object' && v !== null ? keys(v, `${prefix}${k}.`) : [`${prefix}${k}`]
        )
        .sort();
    }
    expect(keys(enNotifications)).toEqual(keys(zhNotifications));
  });
});
