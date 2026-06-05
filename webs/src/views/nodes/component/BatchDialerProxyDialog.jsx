import PropTypes from 'prop-types';

// material-ui
import { useTheme } from '@mui/material/styles';
import Alert from '@mui/material/Alert';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import Typography from '@mui/material/Typography';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { useTranslation } from 'react-i18next';

// project imports
import SearchableNodeSelect from 'components/SearchableNodeSelect';
import { getNodeDialogPaperSx, getNodeFieldControlSx, getNodeThemeTokens } from '../nodeTheme';

/**
 * 批量修改前置代理对话框
 */
export default function BatchDialerProxyDialog({
  open,
  selectedCount,
  value,
  setValue,
  proxyNodeOptions,
  loadingProxyNodes,
  onClose,
  onSubmit
}) {
  const theme = useTheme();
  const { t } = useTranslation();
  const { isDark } = useResolvedColorScheme();
  const tokens = getNodeThemeTokens(theme, isDark);
  const fieldControlSx = getNodeFieldControlSx(tokens);

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth PaperProps={{ sx: getNodeDialogPaperSx(theme, tokens) }}>
      <DialogTitle
        sx={{ color: tokens.primaryText, bgcolor: tokens.mutedPanelSurface, borderBottom: '1px solid', borderColor: tokens.panelBorder }}
      >
        {t('nodes.batch.dialerProxyDialog.title')}
      </DialogTitle>
      <DialogContent dividers sx={{ bgcolor: 'transparent', borderColor: tokens.panelBorder }}>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
          {t('nodes.batch.dialerProxyDialog.description', { count: selectedCount })}
        </Typography>
        <SearchableNodeSelect
          nodes={proxyNodeOptions}
          loading={loadingProxyNodes}
          value={value}
          onChange={(newValue) => {
            const name = typeof newValue === 'string' ? newValue : newValue?.Name || '';
            setValue(name);
          }}
          displayField="Name"
          valueField="Name"
          label={t('nodes.batch.dialerProxyDialog.fieldLabel')}
          placeholder={t('nodes.batch.dialerProxyDialog.placeholder')}
          helperText={t('nodes.batch.dialerProxyDialog.helper')}
          freeSolo={true}
          limit={50}
          sx={fieldControlSx}
        />
        <Alert severity="warning" sx={{ mt: 1, borderColor: tokens.softBorder, bgcolor: tokens.nestedPanelSurface }}>
          {t('nodes.batch.dialerProxyDialog.warning')}
        </Alert>
      </DialogContent>
      <DialogActions sx={{ bgcolor: tokens.mutedPanelSurface, borderTop: '1px solid', borderColor: tokens.panelBorder }}>
        <Button onClick={onClose}>{t('common.cancel')}</Button>
        <Button variant="contained" onClick={onSubmit}>
          {t('nodes.batch.dialerProxyDialog.confirm')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

BatchDialerProxyDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  selectedCount: PropTypes.number.isRequired,
  value: PropTypes.string.isRequired,
  setValue: PropTypes.func.isRequired,
  proxyNodeOptions: PropTypes.array.isRequired,
  loadingProxyNodes: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  onSubmit: PropTypes.func.isRequired
};
