const REGIONAL_INDICATOR_OFFSET = 0x1f1e6 - 65;
const COUNTRY_CODE_PATTERN = /^[A-Z]{2}$/;

export const COUNTRY_FALLBACK_EMOJI = String.fromCodePoint(0x1f310);

const UNKNOWN_COUNTRY_VALUES = new Set(['UNKNOWN', 'UNKNOWN COUNTRY', '未知', 'N/A', 'NA', '-', '--']);

export const normalizeCountryCode = (value) => {
  if (typeof value !== 'string') return '';

  const code = value.trim().toUpperCase();
  if (!code || UNKNOWN_COUNTRY_VALUES.has(code)) return '';

  return code;
};

export const isoToFlag = (isoCode, fallback = '') => {
  const code = normalizeCountryCode(isoCode);
  if (!COUNTRY_CODE_PATTERN.test(code)) return fallback;

  const flagCode = code === 'TW' ? 'CN' : code;
  const codePoints = flagCode.split('').map((char) => REGIONAL_INDICATOR_OFFSET + char.charCodeAt(0));
  return String.fromCodePoint(...codePoints);
};

export const formatCountry = (linkCountry) => {
  const code = normalizeCountryCode(linkCountry);
  if (!code) return '';

  const flag = isoToFlag(code);
  return flag ? `${flag} ${code}` : code;
};

export const getCountryDisplay = (linkCountry, { unknownLabel = 'Unknown', fallbackFlag = COUNTRY_FALLBACK_EMOJI } = {}) => {
  const code = normalizeCountryCode(linkCountry);

  if (!code) {
    return {
      flag: fallbackFlag,
      code: '',
      label: unknownLabel,
      text: `${fallbackFlag} ${unknownLabel}`,
      isUnknown: true
    };
  }

  const flag = isoToFlag(code, fallbackFlag);

  return {
    flag,
    code,
    label: code,
    text: `${flag} ${code}`,
    isUnknown: false
  };
};
