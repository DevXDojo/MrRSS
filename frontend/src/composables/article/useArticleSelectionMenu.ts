import type { Ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { copyToClipboard } from '@/utils/clipboard';
import { openInBrowser } from '@/utils/browser';

const searchEngines = {
  Google: 'https://www.google.com/search?q=',
  Bing: 'https://www.bing.com/search?q=',
  Baidu: 'https://www.baidu.com/s?wd=',
  DuckDuckGo: 'https://duckduckgo.com/?q=',
};

export function useArticleSelectionMenu(container: Ref<HTMLElement | null>) {
  const { t } = useI18n();

  function onContextMenu(event: MouseEvent) {
    const target = event.target;
    if (!(target instanceof Element) || target.closest('input,textarea,[contenteditable],a,img,video,audio,button')) return;
    const selection = window.getSelection();
    if (!container.value || !selection || selection.isCollapsed || !selection.rangeCount) return;
    if (!container.value.contains(selection.anchorNode) || !container.value.contains(selection.focusNode)) return;
    const text = selection.toString().trim();
    if (!text) return;

    event.preventDefault();
    event.stopPropagation();
    window.dispatchEvent(new CustomEvent('open-context-menu', {
      detail: {
        x: event.clientX, y: event.clientY,
        items: [
          { label: t('common.copy'), action: 'copy', icon: 'PhCopy' },
          { separator: true },
          ...Object.keys(searchEngines).map((engine) => ({
            label: t('article.action.searchWith', { engine }), action: engine, icon: 'PhMagnifyingGlass',
          })),
        ],
        callback: async (action: string) => {
          if (action === 'copy') {
            const copied = await copyToClipboard(text);
            window.showToast(t(copied ? 'common.toast.copiedToClipboard' : 'common.errors.failedToCopy'), copied ? 'success' : 'error');
          } else if (Object.hasOwn(searchEngines, action)) {
            await openInBrowser(searchEngines[action as keyof typeof searchEngines] + encodeURIComponent(text));
          }
        },
      },
    }));
  }

  return { onContextMenu };
}
