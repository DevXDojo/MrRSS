/**
 * Date formatting utilities for MrRSS
 */

export interface DateFormatPreferences {
  date_format?: string;
  time_format?: string;
  relative_time?: boolean;
}

function formatCalendarDate(date: Date, locale: string, format?: string): string {
  const year = String(date.getFullYear()).padStart(4, '0');
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  switch (format) {
    case 'yyyy-mm-dd':
      return `${year}-${month}-${day}`;
    case 'mm/dd/yyyy':
      return `${month}/${day}/${year}`;
    case 'dd/mm/yyyy':
      return `${day}/${month}/${year}`;
    case 'dd.mm.yyyy':
      return `${day}.${month}.${year}`;
    default:
      return date.toLocaleDateString(locale);
  }
}

/**
 * Format a date string as relative time (within 14 days) or absolute date
 * @param dateStr - ISO date string
 * @param locale - Locale code (e.g., 'en-US', 'zh-CN')
 * @param t - Translation function from i18n
 * @returns Formatted date string (relative time if within 14 days, absolute date otherwise)
 */
export function formatDate(
  dateStr: string,
  locale: string = 'en-US',
  t?: (key: string, params?: Record<string, unknown>) => string,
  preferences: DateFormatPreferences = {}
): string {
  if (!dateStr) return '';

  try {
    const date = new Date(dateStr);
    if (Number.isNaN(date.getTime())) return '';
    if (preferences.relative_time === false) {
      return formatExactDateTime(dateStr, locale, preferences);
    }
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();

    // Handle future dates - always show absolute date
    if (diffMs < 0) {
      if (preferences.date_format && preferences.date_format !== 'locale') {
        return formatCalendarDate(date, locale, preferences.date_format);
      }
      if (locale === 'zh-CN') {
        const year = date.getFullYear();
        const month = date.getMonth() + 1;
        const day = date.getDate();
        return `${year}年${month}月${day}日`;
      } else {
        return date.toLocaleDateString(locale);
      }
    }

    const diffSeconds = Math.floor(diffMs / 1000);
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    // Use relative time for articles within 14 days
    if (diffDays < 14) {
      if (!t) {
        // Fallback if no translation function provided
        if (diffSeconds < 60) return `${diffSeconds}s ago`;
        if (diffMins < 60) return `${diffMins}m ago`;
        if (diffHours < 24) return `${diffHours}h ago`;
        return `${diffDays}d ago`;
      }

      // Use translations
      if (diffSeconds < 60) return t('common.time.secondsAgo', { count: diffSeconds });
      if (diffMins < 60) return t('common.time.minutesAgo', { count: diffMins });
      if (diffHours < 24) return t('common.time.hoursAgo', { count: diffHours });
      return t('common.time.daysAgo', { count: diffDays });
    }

    // Use absolute date for articles 14+ days old
    if (preferences.date_format && preferences.date_format !== 'locale') {
      return formatCalendarDate(date, locale, preferences.date_format);
    }
    if (locale === 'zh-CN') {
      // Format as "2023年12月8日" for Chinese
      const year = date.getFullYear();
      const month = date.getMonth() + 1;
      const day = date.getDate();
      return `${year}年${month}月${day}日`;
    } else {
      return date.toLocaleDateString(locale);
    }
  } catch {
    return '';
  }
}

/** Format an article timestamp with date and time, rounded to minute precision. */
export function formatExactDateTime(
  dateStr: string,
  locale: string = 'en-US',
  preferences: DateFormatPreferences = {}
): string {
  if (!dateStr) return '';

  try {
    const date = new Date(dateStr);
    if (Number.isNaN(date.getTime())) return '';
    const timeOptions: Intl.DateTimeFormatOptions = {
      hour: '2-digit',
      minute: '2-digit',
    };
    if (preferences.time_format === '12h') timeOptions.hourCycle = 'h12';
    if (preferences.time_format === '24h') timeOptions.hourCycle = 'h23';
    if (preferences.date_format && preferences.date_format !== 'locale') {
      return `${formatCalendarDate(date, locale, preferences.date_format)} ${new Intl.DateTimeFormat(locale, timeOptions).format(date)}`;
    }
    return new Intl.DateTimeFormat(locale, {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      ...timeOptions,
    }).format(date);
  } catch {
    return '';
  }
}

/**
 * Format a timestamp as an absolute date (year-month-day)
 * @param timestamp - ISO timestamp string
 * @param locale - Current locale for translations
 * @param t - Translation function from i18n
 * @returns Formatted date string in locale-specific format
 */
export function formatRelativeTime(
  timestamp: string,
  locale: string,
  t: (key: string, params?: Record<string, unknown>) => string
): string {
  if (!timestamp) return t('common.time.never');
  try {
    const date = new Date(timestamp);

    // Format date based on locale
    if (locale === 'zh-CN') {
      // Chinese format: "2023年12月8日"
      const year = date.getFullYear();
      const month = date.getMonth() + 1;
      const day = date.getDate();
      return `${year}年${month}月${day}日`;
    } else {
      // English format: "12/8/2023" (using locale's default format)
      return date.toLocaleDateString('en-US');
    }
  } catch {
    return t('common.time.never');
  }
}
