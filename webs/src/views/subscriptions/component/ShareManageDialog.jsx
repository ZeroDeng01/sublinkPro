import { useState, useEffect, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import Button from '@mui/material/Button';
import Stack from '@mui/material/Stack';
import Box from '@mui/material/Box';
import TextField from '@mui/material/TextField';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import IconButton from '@mui/material/IconButton';
import { formatDateTime } from 'i18n/locales';
import Chip from '@mui/material/Chip';
import Typography from '@mui/material/Typography';
import Switch from '@mui/material/Switch';
import FormControlLabel from '@mui/material/FormControlLabel';
import Alert from '@mui/material/Alert';
import CircularProgress from '@mui/material/CircularProgress';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Tooltip from '@mui/material/Tooltip';
import useMediaQuery from '@mui/material/useMediaQuery';
import { useTheme } from '@mui/material/styles';

import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import LinkIcon from '@mui/icons-material/Link';
import RefreshIcon from '@mui/icons-material/Refresh';
import HistoryIcon from '@mui/icons-material/History';

import { getShares, createShare, updateShare, deleteShare, getShareLogs, refreshShareToken } from '../../../api/shares';
import { getSystemDomain } from '../../../api/settings';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';
import AccessLogsDialog from './AccessLogsDialog';
import ClientUrlsDialog from './ClientUrlsDialog';
import QrCodeDialog from './QrCodeDialog';
import ConfirmDialog from './ConfirmDialog';

const EXPIRE_TYPE_NEVER = 0;
const EXPIRE_TYPE_DAYS = 1;
const EXPIRE_TYPE_DATETIME = 2;

/**
 */
export default function ShareManageDialog({ open, subscription, onClose, showMessage }) {
  const theme = useTheme();
  const { t, i18n } = useTranslation();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const { isDark } = useResolvedColorScheme();
  const { palette, dialogSurface, dialogSurfaceGradient, mutedPanelSurface, nestedPanelSurface, panelBorder } = getSurfaceTokens(
    theme,
    isDark
  );
  const { primaryText, secondaryText, tertiaryText } = getReadableTextTokens(theme, isDark);

  const [shares, setShares] = useState([]);
  const [loading, setLoading] = useState(false);

  const [systemDomainConfig, setSystemDomainConfig] = useState('');

  const [detailOpen, setDetailOpen] = useState(false);
  const [detailShare, setDetailShare] = useState(null);

  const [formOpen, setFormOpen] = useState(false);
  const [editingShare, setEditingShare] = useState(null);
  const [formData, setFormData] = useState({
    name: '',
    token: '',
    expire_type: EXPIRE_TYPE_NEVER,
    expire_days: 30,
    expire_at: '',
    enabled: true
  });

  const [qrOpen, setQrOpen] = useState(false);
  const [qrUrl, setQrUrl] = useState('');
  const [qrTitle, setQrTitle] = useState('');

  const [logsOpen, setLogsOpen] = useState(false);
  const [logsLoading, setLogsLoading] = useState(false);
  const [logs, setLogs] = useState([]);
  const [logsShareName, setLogsShareName] = useState('');

  const [confirmOpen, setConfirmOpen] = useState(false);
  const [confirmInfo, setConfirmInfo] = useState({ title: '', content: '', onConfirm: null });

  const getServerUrl = useCallback(() => {
    if (systemDomainConfig) {
      return systemDomainConfig.replace(/\/+$/, '');
    }
    return `${window.location.protocol}//${window.location.hostname}${window.location.port ? ':' + window.location.port : ''}`;
  }, [systemDomainConfig]);

  const fetchSystemDomain = async () => {
    try {
      const res = await getSystemDomain();
      if (res.data?.systemDomain) {
        setSystemDomainConfig(res.data.systemDomain);
      }
    } catch (error) {
      console.error('Failed to get system domain config:', error);
    }
  };

  const fetchShares = useCallback(async () => {
    if (!subscription?.ID) return;
    setLoading(true);
    try {
      const res = await getShares(subscription.ID);
      setShares(res.data || []);
    } catch (error) {
      console.error('Failed to get share list:', error);
    } finally {
      setLoading(false);
    }
  }, [subscription?.ID]);

  useEffect(() => {
    if (open && subscription?.ID) {
      fetchSystemDomain();
      fetchShares();
    }
  }, [open, subscription?.ID, fetchShares]);

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
    showMessage?.(t('common.copied'), 'success');
  };

  const handleOpenDetail = (share) => {
    setDetailShare(share);
    setDetailOpen(true);
  };

  const handleAdd = () => {
    setEditingShare(null);
    setFormData({
      name: '',
      token: '',
      expire_type: EXPIRE_TYPE_NEVER,
      expire_days: 30,
      expire_at: '',
      enabled: true
    });
    setFormOpen(true);
  };

  const handleEdit = (share, e) => {
    e?.stopPropagation();
    setEditingShare(share);
    setFormData({
      name: share.name || '',
      token: share.token || '',
      expire_type: share.expire_type || EXPIRE_TYPE_NEVER,
      expire_days: share.expire_days || 30,
      expire_at: share.expire_at ? share.expire_at.substring(0, 16) : '',
      enabled: share.enabled !== false
    });
    setFormOpen(true);
  };

  const handleSave = async () => {
    try {
      const data = {
        ...formData,
        subscription_id: subscription.ID
      };

      if (editingShare) {
        data.id = editingShare.id;
        await updateShare(data);
        showMessage?.(t('subscriptions.share.messages.updateSuccess'), 'success');
      } else {
        await createShare(data);
        showMessage?.(t('subscriptions.share.messages.createSuccess'), 'success');
      }
      setFormOpen(false);
      fetchShares();
    } catch (error) {
      console.error('Failed to save:', error);
      showMessage?.(error.response?.data?.msg || t('subscriptions.share.messages.saveFailed'), 'error');
    }
  };

  const handleDelete = (share, e) => {
    e?.stopPropagation();
    setConfirmInfo({
      title: t('subscriptions.share.confirm.deleteTitle'),
      content: t('subscriptions.share.confirm.deleteContent', { name: share.name || share.token }),
      onConfirm: async () => {
        try {
          await deleteShare(share.id);
          showMessage?.(t('subscriptions.share.messages.deleteSuccess'), 'success');
          fetchShares();
          if (detailShare?.id === share.id) {
            setDetailOpen(false);
          }
        } catch (error) {
          console.error('Failed to delete:', error);
          showMessage?.(error.response?.data?.msg || t('subscriptions.share.messages.deleteFailed'), 'error');
        }
        setConfirmOpen(false);
      }
    });
    setConfirmOpen(true);
  };

  const handleRefreshToken = (share, e) => {
    e?.stopPropagation();
    setConfirmInfo({
      title: t('subscriptions.share.confirm.refreshTitle'),
      content: t('subscriptions.share.confirm.refreshContent'),
      onConfirm: async () => {
        try {
          await refreshShareToken(share.id);
          showMessage?.(t('subscriptions.share.messages.tokenRefreshed'), 'success');
          fetchShares();
          if (detailShare?.id === share.id) {
            setDetailOpen(false);
          }
        } catch (error) {
          console.error('Failed to refresh:', error);
          showMessage?.(error.response?.data?.msg || t('subscriptions.share.messages.refreshFailed'), 'error');
        }
        setConfirmOpen(false);
      }
    });
    setConfirmOpen(true);
  };

  const handleViewLogs = async (share, e) => {
    e?.stopPropagation();
    setLogsShareName(share.name || t('subscriptions.share.unnamed'));
    setLogsLoading(true);
    setLogsOpen(true);
    try {
      const res = await getShareLogs(share.id);
      setLogs(res.data || []);
    } catch (error) {
      console.error('Failed to get logs:', error);
      setLogs([]);
    } finally {
      setLogsLoading(false);
    }
  };

  const handleQrCode = (url, title) => {
    setQrUrl(url);
    setQrTitle(title);
    setQrOpen(true);
  };

  const getExpireText = (share) => {
    if (!share.enabled) return t('common.disabled');
    switch (share.expire_type) {
      case EXPIRE_TYPE_NEVER:
        return t('subscriptions.share.expire.never');
      case EXPIRE_TYPE_DAYS:
        return t('subscriptions.share.expire.days', { count: share.expire_days });
      case EXPIRE_TYPE_DATETIME:
        return share.expire_at
          ? formatDateTime(share.expire_at, i18n.resolvedLanguage || i18n.language)
          : t('subscriptions.share.expire.datetime');
      default:
        return t('subscriptions.share.expire.never');
    }
  };

  const isExpired = (share) => {
    if (!share.enabled) return true;
    if (share.expire_type === EXPIRE_TYPE_DAYS && share.expire_days > 0) {
      const created = new Date(share.created_at);
      const expireDate = new Date(created.getTime() + share.expire_days * 24 * 60 * 60 * 1000);
      return new Date() > expireDate;
    }
    if (share.expire_type === EXPIRE_TYPE_DATETIME && share.expire_at) {
      return new Date() > new Date(share.expire_at);
    }
    return false;
  };

  const getDialogPaperSx = (fullScreen = false) => ({
    borderRadius: fullScreen ? 0 : 3,
    overflow: 'hidden',
    bgcolor: dialogSurface,
    backgroundImage: dialogSurfaceGradient,
    border: fullScreen ? 'none' : '1px solid',
    borderColor: panelBorder
  });

  const iconButtonBaseSx = {
    color: secondaryText,
    bgcolor: nestedPanelSurface,
    border: '1px solid',
    borderColor: panelBorder,
    boxShadow: isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.04)}` : 'none',
    transition: 'all 0.2s ease',
    '&:hover': {
      color: primaryText,
      bgcolor: withAlpha(palette.primary.main, isDark ? 0.14 : 0.06),
      borderColor: withAlpha(palette.primary.main, isDark ? 0.34 : 0.2)
    }
  };

  const actionIconButtonSx = {
    ...iconButtonBaseSx,
    width: 32,
    height: 32
  };

  const legacyChipSx = {
    height: 20,
    fontSize: '0.68rem',
    fontWeight: 700,
    bgcolor: withAlpha(palette.primary.main, isDark ? 0.18 : 0.1),
    color: palette.primary.main,
    border: '1px solid',
    borderColor: withAlpha(palette.primary.main, isDark ? 0.38 : 0.22),
    '& .MuiChip-label': {
      px: 0.9
    }
  };

  const renderShareCard = (share) => {
    const expired = isExpired(share);
    const accentColor = share.is_legacy ? palette.primary.main : expired ? palette.error.main : palette.info.main;
    const accentSurface = share.is_legacy
      ? withAlpha(palette.primary.main, isDark ? 0.16 : 0.06)
      : expired
        ? withAlpha(palette.error.main, isDark ? 0.14 : 0.05)
        : nestedPanelSurface;
    const accentBorder = share.is_legacy
      ? withAlpha(palette.primary.main, isDark ? 0.38 : 0.22)
      : expired
        ? withAlpha(palette.error.main, isDark ? 0.34 : 0.2)
        : panelBorder;

    return (
      <Card
        key={share.id}
        sx={{
          borderRadius: 2.5,
          bgcolor: accentSurface,
          backgroundImage: share.is_legacy
            ? `linear-gradient(180deg, ${withAlpha(palette.primary.main, isDark ? 0.1 : 0.04)} 0%, ${accentSurface} 100%)`
            : 'none',
          border: '1px solid',
          borderColor: accentBorder,
          boxShadow: isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.04)}` : 'none',
          opacity: expired ? 0.72 : 1,
          transition: 'all 0.2s ease',
          '&:hover': {
            borderColor: withAlpha(accentColor, isDark ? 0.48 : 0.28),
            bgcolor: share.is_legacy ? withAlpha(palette.primary.main, isDark ? 0.2 : 0.08) : mutedPanelSurface
          }
        }}
      >
        <CardContent sx={{ px: 2, py: 1.75, '&:last-child': { pb: 1.75 } }}>
          <Stack direction="row" alignItems="center" spacing={1.25}>
            <Box
              onClick={() => handleOpenDetail(share)}
              sx={{
                display: 'flex',
                alignItems: 'center',
                flex: 1,
                minWidth: 0,
                cursor: 'pointer',
                gap: 1,
                '&:hover': { opacity: 0.92 }
              }}
            >
              <Box
                sx={{
                  width: 34,
                  height: 34,
                  borderRadius: 1.75,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  flexShrink: 0,
                  bgcolor: withAlpha(accentColor, isDark ? 0.16 : 0.08),
                  color: expired ? tertiaryText : accentColor,
                  border: '1px solid',
                  borderColor: withAlpha(accentColor, isDark ? 0.32 : 0.18)
                }}
              >
                <LinkIcon fontSize="small" />
              </Box>
              <Box sx={{ flex: 1, minWidth: 0 }}>
                <Stack direction="row" alignItems="center" spacing={0.75} sx={{ mb: 0.35 }}>
                  <Typography variant="body2" fontWeight={600} noWrap sx={{ color: primaryText }}>
                    {share.name || t('subscriptions.share.unnamed')}
                  </Typography>
                  {share.is_legacy && <Chip label={t('subscriptions.share.defaultChip')} size="small" sx={legacyChipSx} />}
                </Stack>
                <Typography variant="caption" sx={{ color: expired ? tertiaryText : secondaryText }}>
                  {t('subscriptions.share.cardMeta', { expire: getExpireText(share), count: share.access_count || 0 })}
                </Typography>
              </Box>
            </Box>
            <Stack direction="row" spacing={0.5}>
              <Tooltip title={t('subscriptions.share.actions.accessLogs')}>
                <IconButton size="small" onClick={(e) => handleViewLogs(share, e)} sx={actionIconButtonSx}>
                  <HistoryIcon fontSize="small" />
                </IconButton>
              </Tooltip>
              <Tooltip title={t('common.edit')}>
                <IconButton size="small" onClick={(e) => handleEdit(share, e)} sx={actionIconButtonSx}>
                  <EditIcon fontSize="small" />
                </IconButton>
              </Tooltip>
              {share.is_legacy ? (
                <Tooltip title={t('subscriptions.share.actions.refreshToken')}>
                  <IconButton
                    size="small"
                    onClick={(e) => handleRefreshToken(share, e)}
                    sx={{
                      ...actionIconButtonSx,
                      color: palette.warning.main,
                      '&:hover': {
                        ...actionIconButtonSx['&:hover'],
                        color: palette.warning.main,
                        bgcolor: withAlpha(palette.warning.main, isDark ? 0.16 : 0.08),
                        borderColor: withAlpha(palette.warning.main, isDark ? 0.36 : 0.2)
                      }
                    }}
                  >
                    <RefreshIcon fontSize="small" />
                  </IconButton>
                </Tooltip>
              ) : (
                <Tooltip title={t('common.delete')}>
                  <IconButton
                    size="small"
                    onClick={(e) => handleDelete(share, e)}
                    sx={{
                      ...actionIconButtonSx,
                      color: palette.error.main,
                      '&:hover': {
                        ...actionIconButtonSx['&:hover'],
                        color: palette.error.main,
                        bgcolor: withAlpha(palette.error.main, isDark ? 0.16 : 0.08),
                        borderColor: withAlpha(palette.error.main, isDark ? 0.34 : 0.2)
                      }
                    }}
                  >
                    <DeleteIcon fontSize="small" />
                  </IconButton>
                </Tooltip>
              )}
            </Stack>
          </Stack>
        </CardContent>
      </Card>
    );
  };

  const detailClientUrls = detailShare
    ? {
        [t('subscriptions.share.clients.auto')]: `${getServerUrl()}/c/?token=${detailShare.token}`,
        Clash: `${getServerUrl()}/c/?token=${detailShare.token}&client=clash`,
        Surge: `${getServerUrl()}/c/?token=${detailShare.token}&client=surge`,
        V2ray: `${getServerUrl()}/c/?token=${detailShare.token}&client=v2ray`
      }
    : {};

  return (
    <>
      <Dialog
        open={open}
        onClose={onClose}
        maxWidth="sm"
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
          <Stack direction="row" alignItems="center" justifyContent="space-between">
            <Typography variant="h6">{t('subscriptions.share.title', { name: subscription?.Name })}</Typography>
            <Stack direction="row" spacing={1}>
              <IconButton size="small" onClick={fetchShares} disabled={loading} sx={iconButtonBaseSx}>
                <RefreshIcon fontSize="small" />
              </IconButton>
              <Button variant="contained" size="small" startIcon={<AddIcon />} onClick={handleAdd}>
                {t('common.add')}
              </Button>
            </Stack>
          </Stack>
        </DialogTitle>

        <DialogContent
          sx={{
            px: 2.5,
            pt: 2.5,
            pb: 2,
            bgcolor: dialogSurface
          }}
        >
          {loading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', py: 4.25 }}>
              <CircularProgress />
            </Box>
          ) : shares.length === 0 ? (
            <Alert
              variant="outlined"
              severity="info"
              sx={{
                mt: 1.5,
                bgcolor: withAlpha(palette.info.main, isDark ? 0.12 : 0.05),
                borderColor: withAlpha(palette.info.main, isDark ? 0.3 : 0.18)
              }}
            >
              {t('subscriptions.share.empty')}
            </Alert>
          ) : (
            <Stack spacing={1.5} sx={{ mt: 1.5 }}>
              {shares.map((share) => renderShareCard(share))}
            </Stack>
          )}
        </DialogContent>

        <DialogActions sx={{ px: 2.5, py: 1.5, bgcolor: mutedPanelSurface, borderTop: '1px solid', borderColor: panelBorder }}>
          <Button onClick={onClose} variant="outlined">
            {t('common.close')}
          </Button>
        </DialogActions>
      </Dialog>

      <ClientUrlsDialog
        open={detailOpen}
        title={detailShare?.name || t('subscriptions.share.detailTitle')}
        subtitle={t('subscriptions.share.detailSubtitle')}
        legacy={Boolean(detailShare?.is_legacy)}
        clientUrls={detailClientUrls}
        onClose={() => setDetailOpen(false)}
        onQrCode={handleQrCode}
        onCopy={copyToClipboard}
      />

      <Dialog
        open={formOpen}
        onClose={() => setFormOpen(false)}
        maxWidth="xs"
        fullWidth
        slotProps={{
          paper: {
            sx: getDialogPaperSx(false)
          }
        }}
      >
        <DialogTitle
          sx={{
            px: 2.5,
            py: 2,
            bgcolor: mutedPanelSurface,
            borderBottom: '1px solid',
            borderColor: panelBorder
          }}
        >
          {editingShare ? t('subscriptions.share.form.editTitle') : t('subscriptions.share.form.addTitle')}
        </DialogTitle>
        <DialogContent sx={{ bgcolor: dialogSurface }}>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <TextField
              label={t('subscriptions.share.form.name')}
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder={t('subscriptions.share.form.namePlaceholder')}
              size="small"
              fullWidth
            />

            <TextField
              label={t('subscriptions.share.form.token')}
              value={formData.token}
              onChange={(e) => setFormData({ ...formData, token: e.target.value })}
              placeholder={t('subscriptions.share.form.tokenPlaceholder')}
              size="small"
              fullWidth
              helperText={t('subscriptions.share.form.tokenHelper')}
            />

            <FormControl size="small" fullWidth>
              <InputLabel>{t('subscriptions.share.form.expirePolicy')}</InputLabel>
              <Select
                value={formData.expire_type}
                label={t('subscriptions.share.form.expirePolicy')}
                onChange={(e) => setFormData({ ...formData, expire_type: e.target.value })}
              >
                <MenuItem value={EXPIRE_TYPE_NEVER}>{t('subscriptions.share.expire.never')}</MenuItem>
                <MenuItem value={EXPIRE_TYPE_DAYS}>{t('subscriptions.share.form.expireByDays')}</MenuItem>
                <MenuItem value={EXPIRE_TYPE_DATETIME}>{t('subscriptions.share.form.expireAt')}</MenuItem>
              </Select>
            </FormControl>

            {formData.expire_type === EXPIRE_TYPE_DAYS && (
              <TextField
                label={t('subscriptions.share.form.expireDays')}
                type="number"
                value={formData.expire_days}
                onChange={(e) => setFormData({ ...formData, expire_days: parseInt(e.target.value) || 0 })}
                size="small"
                fullWidth
                inputProps={{ min: 1 }}
              />
            )}

            {formData.expire_type === EXPIRE_TYPE_DATETIME && (
              <TextField
                label={t('subscriptions.share.form.expireTime')}
                type="datetime-local"
                value={formData.expire_at}
                onChange={(e) => setFormData({ ...formData, expire_at: e.target.value })}
                size="small"
                fullWidth
                InputLabelProps={{ shrink: true }}
              />
            )}

            {editingShare && (
              <FormControlLabel
                control={<Switch checked={formData.enabled} onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })} />}
                label={t('subscriptions.share.form.enabled')}
              />
            )}
          </Stack>
        </DialogContent>
        <DialogActions sx={{ px: 2.5, py: 1.75, bgcolor: mutedPanelSurface, borderTop: '1px solid', borderColor: panelBorder }}>
          <Button onClick={() => setFormOpen(false)}>{t('common.cancel')}</Button>
          <Button variant="contained" onClick={handleSave}>
            {t('common.save')}
          </Button>
        </DialogActions>
      </Dialog>

      <AccessLogsDialog
        open={logsOpen}
        logs={logs}
        loading={logsLoading}
        title={t('subscriptions.share.logsTitle', { name: logsShareName })}
        onClose={() => setLogsOpen(false)}
      />

      <QrCodeDialog open={qrOpen} title={qrTitle} url={qrUrl} onClose={() => setQrOpen(false)} onCopy={copyToClipboard} />

      <ConfirmDialog
        open={confirmOpen}
        title={confirmInfo.title}
        content={confirmInfo.content}
        onClose={() => setConfirmOpen(false)}
        onConfirm={confirmInfo.onConfirm}
      />
    </>
  );
}
