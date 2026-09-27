/**
 * Font detection utility to check which fonts are available on the system
 */

// Common fonts to check for different platforms and languages
const COMMON_FONTS = {
  // Chinese fonts
  chinese: [
    'PingFang SC',
    'Microsoft YaHei',
    'SimHei',
    'SimSun',
    'KaiTi',
    'FangSong',
    'STHeiti',
    'STSong',
    'STKaiti',
    'STFangsong',
    'Noto Sans CJK SC',
    'Noto Sans SC',
    'Noto Serif CJK SC',
    'Noto Serif SC',
    'Source Han Sans SC',
    'Source Han Sans CN',
    'Source Han Serif SC',
    'Source Han Serif CN',
    'Sarasa Gothic SC',
    'Sarasa UI SC',
    'LXGW WenKai',
    'LXGW WenKai GB',
    'LXGW WenKai Lite',
    'LXGW WenKai Screen',
    'WenQuanYi Micro Hei',
    'WenQuanYi Zen Hei',
    '微软雅黑',
    '宋体',
    '黑体',
    '楷体',
    '仿宋',
    '等线',
    '方正书宋',
    '霞鹜文楷',
    '思源黑体',
    '思源宋体',
  ],
  // Japanese fonts
  japanese: [
    'Hiragino Kaku Gothic ProN',
    'Hiragino Mincho ProN',
    'Yu Gothic',
    'Yu Mincho',
    'Meiryo',
    'MS Gothic',
    'MS Mincho',
    'Noto Sans CJK JP',
    'Noto Serif CJK JP',
  ],
  // Korean fonts
  korean: [
    'Malgun Gothic',
    'Apple SD Gothic Neo',
    'Dotum',
    'Batang',
    'Noto Sans CJK KR',
    'Noto Serif CJK KR',
  ],
  // Western fonts
  western: [
    'Arial',
    'Arial Black',
    'Arial Narrow',
    'Calibri',
    'Cambria',
    'Cambria Math',
    'Cambria',
    'Century Gothic',
    'Comic Sans MS',
    'Consolas',
    'Constantia',
    'Corbel',
    'Courier New',
    'Georgia',
    'Helvetica',
    'Impact',
    'Lucida Console',
    'Lucida Sans Unicode',
    'Microsoft Sans Serif',
    'Palatino Linotype',
    'Segoe UI',
    'Tahoma',
    'Times New Roman',
    'Trebuchet MS',
    'Verdana',
    'Monaco',
    'Menlo',
    'PT Sans',
    'PT Serif',
    'Open Sans',
    'Roboto',
    'Ubuntu',
    'Oxygen',
    'Liberation Sans',
    'Nimbus Sans',
    'Cantarell',
    'DejaVu Sans',
    'Noto Sans',
    'Inter',
    'Source Sans Pro',
    'Source Serif Pro',
    'Fira Sans',
    'Fira Code',
    'JetBrains Mono',
  ],
  // Monospace fonts
  monospace: [
    'Consolas',
    'Monaco',
    'Menlo',
    'Courier New',
    'Lucida Console',
    'Source Code Pro',
    'Fira Code',
    'JetBrains Mono',
    'Ubuntu Mono',
    'DejaVu Sans Mono',
    'Liberation Mono',
    'Nimbus Mono',
    'Cascadia Code',
    'PT Mono',
  ],
};

export const SYSTEM_FONT_STACK =
  'Inter, "Noto Sans CJK SC", system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif';

export interface RecommendedFonts {
  serif: string[];
  sansSerif: string[];
  monospace: string[];
}

let cachedRecommendedFonts: RecommendedFonts | null = null;

/**
 * Resolve a saved font setting to a complete CSS font-family value.
 * Returning an explicit stack for "system" prevents article content from
 * inheriting a separately configured interface font.
 */
export function resolveFontFamily(fontFamily: unknown): string {
  const value = typeof fontFamily === 'string' ? fontFamily.trim() : '';

  switch (value) {
    case 'serif':
      return 'Georgia, "Times New Roman", Times, serif';
    case 'sans-serif':
      return '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif';
    case 'monospace':
      return '"Courier New", Courier, monospace';
    case 'system':
    case '':
      return SYSTEM_FONT_STACK;
    default: {
      const escapedValue = value.replace(/["\\]/g, '\\$&');
      return `"${escapedValue}", ${SYSTEM_FONT_STACK}`;
    }
  }
}

/**
 * Check if a specific font is available on the system
 */
export function isFontAvailable(fontName: string): boolean {
  // Create a test canvas context
  const canvas = document.createElement('canvas');
  const context = canvas.getContext('2d');
  if (!context) return false;
  canvas.width = 200;
  canvas.height = 48;

  function glyphsDiffer(fallback: string): boolean {
    // CJK faces often have identical advance widths. Compare their actual
    // glyphs as well, otherwise an installed Chinese face can still be missed.
    try {
      const pixels = (family: string) => {
        context!.clearRect(0, 0, canvas.width, canvas.height);
        context!.font = `32px ${family}`;
        context!.fillText('中文阅读', 0, 36);
        return context!.getImageData(0, 0, canvas.width, canvas.height).data;
      };
      const baseline = pixels(fallback);
      const candidate = pixels(`"${escapedName}", ${fallback}`);
      return candidate.some((value, index) => value !== baseline[index]);
    } catch {
      // Some webviews restrict canvas pixel access. Custom names still work.
      return false;
    }
  }

  // Chinese-only fonts can share Latin fallback glyphs. Compare CJK and Latin
  // samples against several generic families, including the system default.
  const escapedName = fontName.replace(/["\\]/g, '\\$&');
  return ['sans-serif', 'serif', 'monospace'].some(
    (fallback) =>
      ['mmmmmmmmmmlli', '中文字体阅读测试，汉字排版。'].some((sample) => {
        context.font = `100px ${fallback}`;
        const baseline = context.measureText(sample).width;
        context.font = `100px "${escapedName}", ${fallback}`;
        return context.measureText(sample).width !== baseline;
      }) || glyphsDiffer(fallback)
  );
}

/**
 * Get all available fonts from a list
 */
export function getAvailableFonts(fontList: string[]): string[] {
  return fontList.filter((font) => isFontAvailable(font));
}

/**
 * Get all common system fonts available on the user's system
 */
export function getSystemFonts(): {
  chinese: string[];
  japanese: string[];
  korean: string[];
  western: string[];
  monospace: string[];
  all: string[];
} {
  const result = {
    chinese: getAvailableFonts(COMMON_FONTS.chinese),
    japanese: getAvailableFonts(COMMON_FONTS.japanese),
    korean: getAvailableFonts(COMMON_FONTS.korean),
    western: getAvailableFonts(COMMON_FONTS.western),
    monospace: getAvailableFonts(COMMON_FONTS.monospace),
    all: [] as string[],
  };

  // Combine all unique fonts
  result.all = [
    ...result.chinese,
    ...result.japanese,
    ...result.korean,
    ...result.western,
    ...result.monospace,
  ].filter((value, index, self) => self.indexOf(value) === index);

  return result;
}

/**
 * Get recommended fonts based on system availability
 */
export function getRecommendedFonts(): RecommendedFonts {
  if (cachedRecommendedFonts) {
    return cachedRecommendedFonts;
  }

  const systemFonts = getSystemFonts();

  // Categorize fonts (simplified categorization)
  const serifFonts = new Set<string>();
  const sansSerifFonts = new Set<string>();
  const monospaceFonts = new Set<string>();

  // Add monospace fonts
  systemFonts.monospace.forEach((font) => monospaceFonts.add(font));

  // Categorize known fonts
  const knownSerif = [
    'Georgia',
    'Times New Roman',
    'Palatino Linotype',
    'Cambria',
    'PT Serif',
    'Source Serif Pro',
    'Noto Serif',
    'SimSun',
    'STSong',
    'Hiragino Mincho ProN',
    'Yu Mincho',
    'Batang',
    'Noto Serif CJK SC',
    'Noto Serif SC',
    'Noto Serif CJK JP',
    'Noto Serif CJK KR',
    'Source Han Serif',
    'LXGW WenKai',
    '宋体',
    '楷体',
    '仿宋',
    '方正书宋',
    '霞鹜文楷',
  ];

  const knownSansSerif = [
    'Arial',
    'Arial Black',
    'Calibri',
    'Century Gothic',
    'Consolas',
    'Corbel',
    'Helvetica',
    'Segoe UI',
    'Tahoma',
    'Trebuchet MS',
    'Verdana',
    'PT Sans',
    'Open Sans',
    'Roboto',
    'Ubuntu',
    'Inter',
    'Source Sans Pro',
    'Fira Sans',
    'Microsoft YaHei',
    'SimHei',
    'PingFang SC',
    'STHeiti',
    'Hiragino Kaku Gothic ProN',
    'Yu Gothic',
    'Meiryo',
    'Malgun Gothic',
    'Apple SD Gothic Neo',
    'Noto Sans',
    'Noto Sans CJK SC',
    'Noto Sans SC',
    'Noto Sans CJK JP',
    'Noto Sans CJK KR',
    'Source Han Sans',
    'Sarasa Gothic',
    'Sarasa UI',
  ];

  systemFonts.all.forEach((font) => {
    if (knownSerif.some((name) => font.includes(name))) {
      serifFonts.add(font);
    } else if (knownSansSerif.some((name) => font.includes(name))) {
      sansSerifFonts.add(font);
    } else if (monospaceFonts.has(font)) {
      // Already in monospace
    } else {
      // Default to sans-serif for unknown fonts
      sansSerifFonts.add(font);
    }
  });

  cachedRecommendedFonts = {
    serif: Array.from(serifFonts),
    sansSerif: Array.from(sansSerifFonts),
    monospace: Array.from(monospaceFonts),
  };

  return cachedRecommendedFonts;
}
