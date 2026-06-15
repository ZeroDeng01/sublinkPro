import { useState, useEffect, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import Stack from '@mui/material/Stack';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import useMediaQuery from '@mui/material/useMediaQuery';
import { useTheme } from '@mui/material/styles';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getSurfaceTokens, getReadableTextTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';

const NATIVE_CLIENT_LINKS = [
  { key: 'clash', client: 'clash' },
  { key: 'mihomo', client: 'mihomo' },
  { key: 'surge', client: 'surge' },
  { key: 'v2ray', client: 'v2ray' }
];

const EXPANDED_CLIENT_LINKS = [
  { key: 'loon', client: 'loon' },
  { key: 'egern', client: 'egern' },
  { key: 'stash', client: 'stash' },
  { key: 'surfboard', client: 'surfboard' },
  { key: 'shadowrocket', client: 'shadowrocket' },
  { key: 'quantumultX', client: 'quanx' },
  { key: 'singBox', client: 'sing-box' },
  { key: 'uri', client: 'uri' },
  { key: 'json', client: 'json' }
];

/**
 * ShareExportDialog - Batch export share links to text file
 */
export default function ShareExportDialog({ open, onClose, selectedShares, serverUrl, subStoreTargets }) {
  const theme = useTheme();
  const { t } = useTranslation();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const { isDark } = useResolvedColorScheme();
  const { palette, dialogSurface, dialogSurfaceGradient, mutedPanelSurface, nestedPanelSurface, panelBorder } = getSurfaceTokens(
    theme,
    isDark
  );
  const { primaryText, secondaryText } = getReadableTextTokens(theme, isDark);

  const [selectedClientType, setSelectedClientType] = useState('auto');

  // Generate links based on selected client type
  const generatedLinks = useMemo(() => {
    if (!selectedShares || selectedShares.length === 0) return [];

    const baseUrl = serverUrl.replace(/\/+$/, '');

    return selectedShares.map((share) => {
      if (selectedClientType === 'auto') {
        return `${baseUrl}/c/?token=${share.token}`;
      }
      return `${baseUrl}/c/?token=${share.token}&client=${selectedClientType}`;
    });
  }, [selectedShares, selectedClientType, serverUrl]);

  // Reset to default when dialog opens
  useEffect(() => {
    if (open) {
      setSelectedClientType('auto');
    }
  }, [open]);

  // Handle export/download
  const handleExport = () => {
    const links = generatedLinks.join('\n');
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, -5);
    const clientLabel = selectedClientType === 'auto' ? 'auto' : selectedClientType;
    const filename = `sublinkpro_${clientLabel}_${selectedShares.length}_${timestamp}.txt`;

    // Create blob with UTF-8 BOM
    const blob = new Blob(['﻿' + links], {
      type: 'text/plain;charset=utf-8;'
    });

    // Trigger download
    const link = document.createElement('a');
    link.href = URL.createObjectURL(blob);
    link.download = filename;
    link.click();

    // Cleanup
    URL.revokeObjectURL(link.href);
    onClose();
  };

  // Build available client type options
  const allOptions = [
    { value: 'auto', label: t('subscriptions.share.exportDialog.formatAuto') },
    ...NATIVE_CLIENT_LINKS.map((client) => ({
      value: client.client,
      label: t(`subscriptions.share.exportDialog.format${client.key.charAt(0).toUpperCase() + client.key.slice(1)}`)
    })),
    ...EXPANDED_CLIENT_LINKS.filter((client) => subStoreTargets?.includes(client.client)).map((client) => ({
      value: client.client,
      label: client.key.charAt(0).toUpperCase() + client.key.slice(1)
    }))
  ];

  const getDialogPaperSx = (fullScreen) => ({
    borderRadius: fullScreen ? 0 : 3,
    overflow: 'hidden',
    bgcolor: dialogSurface,
    backgroundImage: dialogSurfaceGradient,
    border: fullScreen ? 'none' : '1px solid',
    borderColor: panelBorder
  });

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="md"
      fullWidth
      fullScreen={isMobile}
      slotProps={{
        paper: {
          sx: getDialogPaperSx(isMobile)
        }
      }}
    >
      <DialogTitle
        sx={{
          px: 2.5,
          py: 1.75,
          bgcolor: mutedPanelSurface,
          borderBottom: '1px solid',
          borderColor: panelBorder,
          boxShadow: `inset 0 -1px 0 ${withAlpha(palette.divider, 0.4)}`
        }}
      >
        {t('subscriptions.share.exportDialog.title')}
      </DialogTitle>

      <DialogContent
        sx={{
          px: 2.5,
          bgcolor: dialogSurface
        }}
      >
        <Stack spacing={3} sx={{ mt: 2 }}>
          {/* Client Type Selection */}
          <FormControl fullWidth size="small">
            <InputLabel id="client-type-select-label">{t('subscriptions.share.exportDialog.selectFormat')}</InputLabel>
            <Select
              labelId="client-type-select-label"
              value={selectedClientType}
              label={t('subscriptions.share.exportDialog.selectFormat')}
              onChange={(e) => setSelectedClientType(e.target.value)}
              sx={{
                bgcolor: nestedPanelSurface,
                '&:hover': {
                  bgcolor: withAlpha(palette.primary.main, isDark ? 0.08 : 0.04)
                }
              }}
            >
              {allOptions.map((option) => (
                <MenuItem key={option.value} value={option.value}>
                  {option.label}
                </MenuItem>
              ))}
            </Select>
          </FormControl>

          {/* Link Preview */}
          <Box>
            <Typography variant="body2" sx={{ color: primaryText, fontWeight: 600, mb: 1 }}>
              {t('subscriptions.share.exportDialog.preview')}
            </Typography>
            <TextField
              multiline
              fullWidth
              rows={12}
              value={generatedLinks.join('\n')}
              slotProps={{
                input: {
                  readOnly: true,
                  sx: {
                    fontFamily: 'monospace',
                    fontSize: '0.75rem',
                    bgcolor: nestedPanelSurface,
                    color: secondaryText,
                    '& .MuiInputBase-input': {
                      cursor: 'text'
                    }
                  }
                }
              }}
              sx={{
                '& .MuiOutlinedInput-root': {
                  bgcolor: nestedPanelSurface,
                  border: '1px solid',
                  borderColor: panelBorder
                }
              }}
            />
          </Box>
        </Stack>
      </DialogContent>

      <DialogActions
        sx={{
          px: 2.5,
          py: 1.5,
          bgcolor: mutedPanelSurface,
          borderTop: '1px solid',
          borderColor: panelBorder
        }}
      >
        <Button onClick={onClose}>{t('common.cancel')}</Button>
        <Button variant="contained" onClick={handleExport}>
          {t('subscriptions.share.exportDialog.download')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
