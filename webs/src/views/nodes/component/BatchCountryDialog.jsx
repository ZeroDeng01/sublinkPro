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

// utils
import { isoToFlag } from '../utils';
import { getNodeDialogPaperSx, getNodeFieldControlSx, getNodeThemeTokens } from '../nodeTheme';

// 常用国家代码列表
const COMMON_COUNTRY_CODES = [
  { code: 'HK', name: '香港' },
  { code: 'TW', name: '台湾' },
  { code: 'SG', name: '新加坡' },
  { code: 'JP', name: '日本' },
  { code: 'KR', name: '韩国' },
  { code: 'US', name: '美国' },
  { code: 'GB', name: '英国' },
  { code: 'DE', name: '德国' },
  { code: 'FR', name: '法国' },
  { code: 'NL', name: '荷兰' },
  { code: 'AU', name: '澳大利亚' },
  { code: 'CA', name: '加拿大' },
  { code: 'RU', name: '俄罗斯' },
  { code: 'IN', name: '印度' },
  { code: 'BR', name: '巴西' },
  { code: 'TR', name: '土耳其' },
  { code: 'AR', name: '阿根廷' },
  { code: 'PH', name: '菲律宾' },
  { code: 'MY', name: '马来西亚' },
  { code: 'TH', name: '泰国' },
  { code: 'VN', name: '越南' },
  { code: 'ID', name: '印度尼西亚' }
];

/**
 * 批量修改国家代码对话框
 */
export default function BatchCountryDialog({ open, selectedCount, value, setValue, countryOptions, onClose, onSubmit }) {
  const theme = useTheme();
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
    const name = commonItem ? commonItem.name : code;
    return flag ? `${flag} ${code} - ${name}` : `${code} - ${name}`;
  };

  // 预览当前输入的国家代码
  const previewFlag = value ? isoToFlag(value.toUpperCase()) : '';
  const previewName = value ? COMMON_COUNTRY_CODES.find((item) => item.code === value.toUpperCase())?.name || '' : '';

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
        批量修改国家/地区
      </DialogTitle>
      <DialogContent dividers sx={{ bgcolor: 'transparent', borderColor: tokens.panelBorder }}>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
          将为选中的 {selectedCount} 个节点设置相同的国家代码
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
            <TextField {...params} label="国家代码" placeholder="输入或选择国家代码，如 US、HK、JP" fullWidth sx={fieldControlSx} />
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
          提示：留空将清除所选节点的国家标记。国家代码使用 ISO 3166-1 alpha-2 标准（如 US、CN、JP）。
        </Typography>
      </DialogContent>
      <DialogActions sx={{ bgcolor: tokens.mutedPanelSurface, borderTop: '1px solid', borderColor: tokens.panelBorder }}>
        <Button onClick={onClose}>取消</Button>
        <Button variant="contained" onClick={onSubmit}>
          确认修改
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
