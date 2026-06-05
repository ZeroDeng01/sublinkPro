import PropTypes from 'prop-types';

// material-ui
import { useTheme } from '@mui/material/styles';
import Autocomplete from '@mui/material/Autocomplete';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { useTranslation } from 'react-i18next';
import { getNodeColorChipSx, getNodeDialogPaperSx, getNodeFieldControlSx, getNodeThemeTokens } from '../nodeTheme';

/**
 * 批量移除标签对话框
 */
export default function BatchRemoveTagDialog({ open, selectedCount, value, setValue, tagOptions, onClose, onSubmit }) {
  const theme = useTheme();
  const { t } = useTranslation();
  const { isDark } = useResolvedColorScheme();
  const tokens = getNodeThemeTokens(theme, isDark);
  const fieldControlSx = getNodeFieldControlSx(tokens, tokens.palette.error.main);

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="sm"
      fullWidth
      PaperProps={{ sx: getNodeDialogPaperSx(theme, tokens, tokens.palette.error.main) }}
    >
      <DialogTitle
        sx={{ color: tokens.primaryText, bgcolor: tokens.mutedPanelSurface, borderBottom: '1px solid', borderColor: tokens.panelBorder }}
      >
        {t('nodes.batch.removeTagDialog.title')}
      </DialogTitle>
      <DialogContent dividers sx={{ bgcolor: 'transparent', borderColor: tokens.panelBorder }}>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
          {t('nodes.batch.removeTagDialog.description', { count: selectedCount })}
        </Typography>
        <Autocomplete
          multiple
          options={tagOptions}
          value={value}
          onChange={(e, newValue) => setValue(newValue)}
          getOptionLabel={(option) => option.name || option}
          isOptionEqualToValue={(option, val) => option.name === (val.name || val)}
          renderOption={(props, option) => {
            const { key, ...otherProps } = props;
            return (
              <li key={key} {...otherProps}>
                <Box
                  sx={{
                    width: 12,
                    height: 12,
                    borderRadius: '50%',
                    backgroundColor: option.color || tokens.palette.primary.main,
                    mr: 1,
                    flexShrink: 0
                  }}
                />
                {option.name}
              </li>
            );
          }}
          renderTags={(val, getTagProps) =>
            val.map((option, index) => {
              const { key, ...tagProps } = getTagProps({ index });
              return (
                <Chip
                  key={key}
                  label={option.name || option}
                  size="small"
                  sx={getNodeColorChipSx(theme, tokens, option.color || theme.palette.primary.main, { deletable: true })}
                  {...tagProps}
                />
              );
            })
          }
          sx={fieldControlSx}
          renderInput={(params) => (
            <TextField
              {...params}
              label={t('nodes.batch.removeTagDialog.fieldLabel')}
              placeholder={t('nodes.batch.removeTagDialog.placeholder')}
              fullWidth
              sx={fieldControlSx}
            />
          )}
        />
        <Typography variant="caption" color="text.secondary" sx={{ mt: 1, display: 'block' }}>
          {t('nodes.batch.removeTagDialog.helper')}
        </Typography>
      </DialogContent>
      <DialogActions sx={{ bgcolor: tokens.mutedPanelSurface, borderTop: '1px solid', borderColor: tokens.panelBorder }}>
        <Button onClick={onClose}>{t('common.cancel')}</Button>
        <Button variant="contained" color="warning" onClick={onSubmit} disabled={value.length === 0}>
          {t('nodes.batch.removeTagDialog.confirm')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

BatchRemoveTagDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  selectedCount: PropTypes.number.isRequired,
  value: PropTypes.array.isRequired,
  setValue: PropTypes.func.isRequired,
  tagOptions: PropTypes.array.isRequired,
  onClose: PropTypes.func.isRequired,
  onSubmit: PropTypes.func.isRequired
};
