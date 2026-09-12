import { afterEach, describe, expect, it } from 'vitest';
import { mount, type VueWrapper } from '@vue/test-utils';
import { nextTick } from 'vue';
import type { Feed } from '@/types/models';
import FeedIcon from './FeedIcon.vue';

const feed: Feed = {
  id: 1,
  title: 'News',
  url: 'rsshub://news/latest',
  link: 'https://news.example.org/',
  image_url: 'https://news.example.org/logo.png',
  category: '',
  last_fetched_at: '',
  last_updated: '2026-09-12T10:00:00Z',
};

describe('recoverable feed icons', () => {
  let wrapper: VueWrapper;
  afterEach(() => wrapper?.unmount());

  it('falls back to the website favicon and then a local icon', async () => {
    wrapper = mount(FeedIcon, { props: { feed } });
    await wrapper.get('img').trigger('error');
    expect(wrapper.get('img').attributes('src')).toBe(
      'https://www.google.com/s2/favicons?domain=news.example.org'
    );
    await wrapper.get('img').trigger('error');
    expect(wrapper.find('img').exists()).toBe(false);
    expect(wrapper.find('[data-testid="feed-icon-fallback"]').exists()).toBe(true);
  });

  it('restores a failed icon when its source changes without remounting the feed', async () => {
    wrapper = mount(FeedIcon, { props: { feed } });
    const oldImage = wrapper.get('img').element;
    await wrapper.get('img').trigger('error');
    await wrapper.get('img').trigger('error');
    await wrapper.setProps({ feed: { ...feed, image_url: 'https://news.example.org/new.png' } });
    oldImage.dispatchEvent(new Event('error'));
    await nextTick();
    expect(wrapper.get('img').attributes('src')).toBe('https://news.example.org/new.png');
    expect((wrapper.get('img').element as HTMLImageElement).style.display).not.toBe('none');
  });

  it('retries failed sources on feed refresh but does not reload healthy icons', async () => {
    wrapper = mount(FeedIcon, { props: { feed } });
    const original = wrapper.get('img').element;
    await wrapper.setProps({ feed: { ...feed, title: 'Updated name' } });
    expect(wrapper.get('img').element).toBe(original);
    await wrapper.get('img').trigger('error');
    await wrapper.setProps({ feed: { ...feed, last_updated: '2026-09-12T11:00:00Z' } });
    expect(wrapper.get('img').attributes('src')).toBe(feed.image_url);
    const retry = wrapper.get('img').element;
    await wrapper.setProps({ feed: { ...feed, last_updated: '2026-09-12T12:00:00Z' } });
    expect(wrapper.get('img').element).toBe(retry);
  });

  it('retries after connectivity returns without automatic retry loops', async () => {
    wrapper = mount(FeedIcon, { props: { feed } });
    await wrapper.get('img').trigger('error');
    await wrapper.get('img').trigger('error');
    await nextTick();
    expect(wrapper.find('img').exists()).toBe(false);
    window.dispatchEvent(new Event('online'));
    await nextTick();
    expect(wrapper.get('img').attributes('src')).toBe(feed.image_url);
    const retry = wrapper.get('img').element;
    window.dispatchEvent(new Event('online'));
    await nextTick();
    expect(wrapper.get('img').element).toBe(retry);
  });

  it('does not request empty URLs or RSSHub route names as website domains', () => {
    wrapper = mount(FeedIcon, {
      props: { feed: { ...feed, image_url: '', link: '', website_url: '' } },
    });
    expect(wrapper.find('img').exists()).toBe(false);
    expect(wrapper.find('[data-testid="feed-icon-fallback"]').exists()).toBe(true);
  });

  it('only attempts a source once when the primary URL is already the favicon', async () => {
    wrapper = mount(FeedIcon, {
      props: {
        feed: {
          ...feed,
          image_url: 'https://www.google.com/s2/favicons?domain=news.example.org',
        },
      },
    });
    await wrapper.get('img').trigger('error');
    expect(wrapper.find('img').exists()).toBe(false);
  });
});
