import { enUS, zhCN } from '@mui/material/locale';

export const DEFAULT_LANGUAGE = 'zh-CN';

export const SUPPORTED_LANGUAGES = ['zh-CN', 'en-US'];

export const LANGUAGE_OPTIONS = [{ value: 'zh-CN' }, { value: 'en-US' }];

export function normalizeLanguage(language) {
  const normalized = String(language || '').replace('_', '-');
  const lower = normalized.toLowerCase();

  if (lower === 'zh' || lower === 'zh-cn' || lower === 'zh-hans' || lower.startsWith('zh-hans-')) {
    return 'zh-CN';
  }

  if (lower === 'en' || lower.startsWith('en-')) {
    return 'en-US';
  }

  return DEFAULT_LANGUAGE;
}

export function getMuiLocale(language) {
  return normalizeLanguage(language) === 'zh-CN' ? zhCN : enUS;
}

export function getTurnstileLanguage(language) {
  return normalizeLanguage(language) === 'zh-CN' ? 'zh-CN' : 'en-US';
}

export function formatDateTime(value, language, options) {
  const date = value instanceof Date ? value : new Date(value);

  if (Number.isNaN(date.getTime())) {
    return '';
  }

  return new Intl.DateTimeFormat(normalizeLanguage(language), options).format(date);
}

export function formatNumber(value, language, options) {
  return new Intl.NumberFormat(normalizeLanguage(language), options).format(value);
}
