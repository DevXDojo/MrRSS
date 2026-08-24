export interface TranslationMessages {
  [key: string]: string | TranslationMessages | any;
}

export type SupportedLocale = 'en-US' | 'zh-CN';
