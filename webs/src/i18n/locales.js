import { enUS, zhCN, ruRU } from '@mui/material/locale';

export const DEFAULT_LANGUAGE = 'zh-CN';

export const SUPPORTED_LANGUAGES = ['zh-CN', 'en-US', 'ru-RU'];

export const LANGUAGE_OPTIONS = [{ value: 'zh-CN' }, { value: 'en-US' }, { value: 'ru-RU' }];

export function normalizeLanguage(language) {
  const normalized = String(language || '').replace('_', '-');
  const lower = normalized.toLowerCase();

  if (lower === 'zh' || lower === 'zh-cn' || lower === 'zh-hans' || lower.startsWith('zh-hans-')) {
    return 'zh-CN';
  }

  if (lower === 'en' || lower.startsWith('en-')) {
    return 'en-US';
  }

  if (lower === 'ru' || lower.startsWith('ru-')) {
    return 'ru-RU';
  }

  return DEFAULT_LANGUAGE;
}

export function getMuiLocale(language) {
  switch (normalizeLanguage(language)) {
    case 'zh-CN':
      return zhCN;
    case 'ru-RU':
      return ruRU;
    default:
      return enUS;
  }
}

export function getTurnstileLanguage(language) {
  const normalized = normalizeLanguage(language);
  if (normalized === 'zh-CN') {
    return 'zh-CN';
  }
  return normalized === 'ru-RU' ? 'ru' : 'en-US';
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
