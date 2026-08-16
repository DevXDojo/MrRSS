export interface ReadingTimeEstimate {
  minutes: number;
  latinWords: number;
  cjkCharacters: number;
}

const LATIN_WORDS_PER_MINUTE = 220;
const CJK_CHARACTERS_PER_MINUTE = 500;

/** Estimate reading time from already-sanitized article HTML or plain text. */
export function estimateReadingTime(content: string): ReadingTimeEstimate {
  if (!content.trim()) return { minutes: 1, latinWords: 0, cjkCharacters: 0 };

  const text = content
    .replace(/<script\b[^>]*>[\s\S]*?<\/script>/gi, ' ')
    .replace(/<style\b[^>]*>[\s\S]*?<\/style>/gi, ' ')
    .replace(/<(?:nav|aside|footer|header)\b[^>]*>[\s\S]*?<\/(?:nav|aside|footer|header)>/gi, ' ')
    .replace(/<[^>]+>/g, ' ')
    .replace(/&(?:nbsp|amp|lt|gt|quot|#39);/gi, ' ');
  const cjkCharacters = (
    text.match(/[\p{Script=Han}\p{Script=Hiragana}\p{Script=Katakana}\p{Script=Hangul}]/gu) || []
  ).length;
  const withoutCjk = text.replace(
    /[\p{Script=Han}\p{Script=Hiragana}\p{Script=Katakana}\p{Script=Hangul}]/gu,
    ' '
  );
  const latinWords = withoutCjk.match(/[\p{L}\p{N}]+(?:['’.-][\p{L}\p{N}]+)*/gu)?.length ?? 0;
  const minutes = Math.max(
    1,
    Math.ceil(latinWords / LATIN_WORDS_PER_MINUTE + cjkCharacters / CJK_CHARACTERS_PER_MINUTE)
  );

  return { minutes, latinWords, cjkCharacters };
}
