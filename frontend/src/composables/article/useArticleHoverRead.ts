import { onBeforeUnmount, watch } from 'vue';
import { useSettings } from '@/composables/core/useSettings';
import { useAppStore } from '@/stores/app';
import type { Article } from '@/types/models';

export function useArticleHoverRead(
  getArticle: () => Article,
  onRead: (id: number) => void,
  isDisabled: () => boolean = () => false
) {
  const { settings } = useSettings();
  const store = useAppStore();
  let timer: ReturnType<typeof setTimeout> | undefined;
  let pending = false;

  function leave(): void {
    clearTimeout(timer);
    timer = undefined;
  }

  function enter(): void {
    leave();
    const article = getArticle();
    if (
      isDisabled() ||
      !settings.value.hover_mark_as_read ||
      article.is_read ||
      article.is_read_later ||
      pending
    )
      return;
    timer = setTimeout(async () => {
      timer = undefined;
      if (
        isDisabled() ||
        !settings.value.hover_mark_as_read ||
        article.is_read ||
        article.is_read_later ||
        getArticle().id !== article.id
      )
        return;
      pending = true;
      try {
        const response = await fetch(`/api/articles/read?id=${article.id}&read=true`, {
          method: 'POST',
        });
        if (!response.ok) throw new Error(`Mark as read failed: ${response.status}`);
        onRead(article.id);
        await Promise.all([store.fetchUnreadCounts(), store.fetchFilterCounts()]);
      } catch (error) {
        console.error('Error marking as read on hover:', error);
      } finally {
        pending = false;
      }
    }, 300);
  }

  onBeforeUnmount(leave);
  watch(isDisabled, (disabled) => {
    if (disabled) leave();
  });
  return { enter, leave };
}
