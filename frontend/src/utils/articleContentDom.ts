const ARTICLE_CONTENT_SELECTOR = '[data-article-content] .prose-content';

export function queryArticleContentImages(root: ParentNode = document): HTMLImageElement[] {
  return Array.from(root.querySelectorAll<HTMLImageElement>(`${ARTICLE_CONTENT_SELECTOR} img`));
}

export function queryArticleContentLinks(root: ParentNode = document): HTMLAnchorElement[] {
  return Array.from(root.querySelectorAll<HTMLAnchorElement>(`${ARTICLE_CONTENT_SELECTOR} a`));
}

export function hasArticleContent(root: ParentNode = document): boolean {
  return root.querySelector(ARTICLE_CONTENT_SELECTOR) !== null;
}
