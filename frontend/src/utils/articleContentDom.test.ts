import { describe, expect, it } from 'vitest';
import {
  hasArticleContent,
  queryArticleContentImages,
  queryArticleContentLinks,
} from './articleContentDom';

describe('article content DOM queries', () => {
  it('excludes images and links rendered by adjacent AI chat UI', () => {
    document.body.innerHTML = `
      <section class="modal-prose-content">
        <div data-article-content>
          <div class="prose-content">
            <a href="https://example.com/article"><img id="article-image" src="article.png"></a>
          </div>
        </div>
        <aside class="article-chat-panel prose">
          <a href="https://example.com/chat"><img id="chat-icon" src="chat.svg"></a>
        </aside>
      </section>
    `;

    const modal = document.querySelector('.modal-prose-content')!;
    expect(hasArticleContent(modal)).toBe(true);
    expect(queryArticleContentImages(modal).map((image) => image.id)).toEqual(['article-image']);
    expect(queryArticleContentLinks(modal).map((link) => link.href)).toEqual([
      'https://example.com/article',
    ]);
  });
});
