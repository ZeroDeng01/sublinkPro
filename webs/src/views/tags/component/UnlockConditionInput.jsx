import { useTranslation } from 'react-i18next';
import PropTypes from 'prop-types';
import Box from '@mui/material/Box';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import { getUnlockProviderOptions, getUnlockStatusOptions } from '../../nodes/utils';

/**
 * 解锁情况复合输入组件
 * 用于在标签规则条件中输入解锁相关的组合条件
 */
export default function UnlockConditionInput({ value, onChange, isMobile }) {
  const { t } = useTranslation();

  const unlockValue = value || {
    provider: '',
    status: '',
    keyword: '',
    providerOperator: 'equals',
    statusOperator: 'equals',
    keywordOperator: 'contains'
  };

  const handleChange = (field, newValue) => {
    onChange({
      ...unlockValue,
      [field]: newValue
    });
  };

  const providerOptions = getUnlockProviderOptions();
  const statusOptions = getUnlockStatusOptions(false);

  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: isMobile ? 'column' : 'row',
        gap: isMobile ? 1.5 : 1,
        alignItems: isMobile ? 'stretch' : 'center',
        flex: 1,
        p: isMobile ? 1.5 : 0,
        backgroundColor: isMobile ? 'action.hover' : 'transparent',
        borderRadius: isMobile ? 1 : 0
      }}
    >
      {/* Provider 选择器 */}
      <FormControl size="small" sx={{ minWidth: isMobile ? '100%' : 140 }}>
        <InputLabel>{t('tags.dialog.rule.unlockProvider')}</InputLabel>
        <Select
          value={unlockValue.provider}
          label={t('tags.dialog.rule.unlockProvider')}
          onChange={(e) => handleChange('provider', e.target.value)}
        >
          <MenuItem value="">
            <em>{t('common.notSelected')}</em>
          </MenuItem>
          {providerOptions.map((opt) => (
            <MenuItem key={opt.value} value={opt.value}>
              {opt.labelKey ? t(opt.labelKey, opt.label) : opt.label}
            </MenuItem>
          ))}
        </Select>
      </FormControl>

      {/* Status 选择器 */}
      <FormControl size="small" sx={{ minWidth: isMobile ? '100%' : 120 }}>
        <InputLabel>{t('tags.dialog.rule.unlockStatus')}</InputLabel>
        <Select
          value={unlockValue.status}
          label={t('tags.dialog.rule.unlockStatus')}
          onChange={(e) => handleChange('status', e.target.value)}
        >
          <MenuItem value="">
            <em>{t('common.notSelected')}</em>
          </MenuItem>
          {statusOptions.map((opt) => (
            <MenuItem key={opt.value} value={opt.value}>
              {opt.labelKey ? t(opt.labelKey, opt.label) : opt.label}
            </MenuItem>
          ))}
        </Select>
      </FormControl>

      {/* Keyword 输入框（可选） */}
      <TextField
        size="small"
        label={t('tags.dialog.rule.unlockKeyword')}
        value={unlockValue.keyword}
        onChange={(e) => handleChange('keyword', e.target.value)}
        placeholder={t('tags.dialog.rule.unlockKeywordPlaceholder')}
        sx={{ minWidth: isMobile ? '100%' : 140, flex: isMobile ? 0 : 1 }}
      />

      {isMobile && (
        <Typography variant="caption" color="text.secondary" sx={{ mt: 0.5 }}>
          {t('tags.dialog.rule.unlockConditionHint')}
        </Typography>
      )}
    </Box>
  );
}

UnlockConditionInput.propTypes = {
  value: PropTypes.object,
  onChange: PropTypes.func.isRequired,
  isMobile: PropTypes.bool
};
