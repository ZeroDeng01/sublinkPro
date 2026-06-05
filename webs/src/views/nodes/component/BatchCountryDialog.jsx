import PropTypes from 'prop-types';
import { useMemo } from 'react';

// material-ui
import { useTheme } from '@mui/material/styles';
import Autocomplete from '@mui/material/Autocomplete';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { useTranslation } from 'react-i18next';

// utils
import { isoToFlag } from '../utils';
import { getNodeDialogPaperSx, getNodeFieldControlSx, getNodeThemeTokens } from '../nodeTheme';

// 常用国家代码列表
const COMMON_COUNTRY_CODES = [
  { code: 'HK', key: 'hk' },
  { code: 'TW', key: 'tw' },
  { code: 'SG', key: 'sg' },
  { code: 'JP', key: 'jp' },
  { code: 'KR', key: 'kr' },
  { code: 'US', key: 'us' },
  { code: 'GB', key: 'gb' },
  { code: 'DE', key: 'de' },
  { code: 'FR', key: 'fr' },
  { code: 'NL', key: 'nl' },
  { code: 'AU', key: 'au' },
  { code: 'CA', key: 'ca' },
  { code: 'RU', key: 'ru' },
  { code: 'IN', key: 'in' },
  { code: 'BR', key: 'br' },
  { code: 'TR', key: 'tr' },
  { code: 'AR', key: 'ar' },
  { code: 'PH', key: 'ph' },
  { code: 'MY', key: 'my' },
  { code: 'TH', key: 'th' },
  { code: 'VN', key: 'vn' },
  { code: 'ID', key: 'id' }
];

/**
 * 批量修改国家代码对话框
 */
export default function BatchCountryDialog({ open, selectedCount, value, setValue, countryOptions, onClose, onSubmit }) {
  const theme = useTheme();
  const { t } = useTranslation();
  const { isDark } = useResolvedColorScheme();
  const tokens = getNodeThemeTokens(theme, isDark);
  const fieldControlSx = getNodeFieldControlSx(tokens, tokens.palette.secondary.main);
  // 合并已有国家列表和常用国家列表，去重
  const allOptions = useMemo(() => {
    const existingCodes = new Set(countryOptions || []);
    const combined = [...(countryOptions || [])];

    // 添加常用国家代码（如果不存在）
    COMMON_COUNTRY_CODES.forEach((item) => {
      if (!existingCodes.has(item.code)) {
        combined.push(item.code);
      }
    });

    return combined.sort();
  }, [countryOptions]);

  // 获取国家代码的显示名称
  const getCountryLabel = (code) => {
    if (!code) return '';
    const flag = isoToFlag(code);
    const commonItem = COMMON_COUNTRY_CODES.find((item) => item.code === code.toUpperCase());
    const name = commonItem ? t(`nodes.batch.countryDialog.countries.${commonItem.key}`) : code;
    return flag ? `${flag} ${code} - ${name}` : `${code} - ${name}`;
  };

  // 预览当前输入的国家代码
  const previewFlag = value ? isoToFlag(value.toUpperCase()) : '';
  const previewCountry = value ? COMMON_COUNTRY_CODES.find((item) => item.code === value.toUpperCase()) : null;
  const previewName = previewCountry ? t(`nodes.batch.countryDialog.countries.${previewCountry.key}`) : '';

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="sm"
      fullWidth
      PaperProps={{ sx: getNodeDialogPaperSx(theme, tokens, tokens.palette.secondary.main) }}
    >
      <DialogTitle
        sx={{ color: tokens.primaryText, bgcolor: tokens.mutedPanelSurface, borderBottom: '1px solid', borderColor: tokens.panelBorder }}
      >
        {t('nodes.batch.countryDialog.title')}
      </DialogTitle>
      <DialogContent dividers sx={{ bgcolor: 'transparent', borderColor: tokens.panelBorder }}>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
          {t('nodes.batch.countryDialog.description', { count: selectedCount })}
        </Typography>
        <Autocomplete
          freeSolo
          options={allOptions}
          value={value}
          onChange={(e, newValue) => setValue(newValue ? newValue.toUpperCase() : '')}
          onInputChange={(e, newInputValue) => setValue(newInputValue ? newInputValue.toUpperCase() : '')}
          getOptionLabel={(option) => option}
          sx={fieldControlSx}
          renderOption={(props, option) => {
            const { key, ...otherProps } = props;
            return (
              <Box component="li" key={key} {...otherProps}>
                {getCountryLabel(option)}
              </Box>
            );
          }}
          renderInput={(params) => (
            <TextField
              {...params}
              label={t('nodes.batch.countryDialog.fieldLabel')}
              placeholder={t('nodes.batch.countryDialog.placeholder')}
              fullWidth
              sx={fieldControlSx}
            />
          )}
        />

        {/* 国旗预览 */}
        {value && (
          <Box
            sx={{
              mt: 2,
              p: 2,
              bgcolor: tokens.nestedPanelSurface,
              borderRadius: 2,
              border: '1px solid',
              borderColor: tokens.softBorder,
              textAlign: 'center'
            }}
          >
            <Typography variant="h3" sx={{ mb: 0.5 }}>
              {previewFlag || '🏳️'}
            </Typography>
            <Typography variant="body1" fontWeight="medium">
              {value.toUpperCase()}
              {previewName && ` - ${previewName}`}
            </Typography>
          </Box>
        )}

        <Typography variant="caption" color="text.secondary" sx={{ mt: 2, display: 'block' }}>
          {t('nodes.batch.countryDialog.helper')}
        </Typography>
      </DialogContent>
      <DialogActions sx={{ bgcolor: tokens.mutedPanelSurface, borderTop: '1px solid', borderColor: tokens.panelBorder }}>
        <Button onClick={onClose}>{t('common.cancel')}</Button>
        <Button variant="contained" onClick={onSubmit}>
          {t('nodes.batch.countryDialog.confirm')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

BatchCountryDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  selectedCount: PropTypes.number.isRequired,
  value: PropTypes.string.isRequired,
  setValue: PropTypes.func.isRequired,
  countryOptions: PropTypes.array,
  onClose: PropTypes.func.isRequired,
  onSubmit: PropTypes.func.isRequired
};
