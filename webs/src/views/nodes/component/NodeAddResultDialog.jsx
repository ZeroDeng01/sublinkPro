import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';

// material-ui
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import Divider from '@mui/material/Divider';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import useMediaQuery from '@mui/material/useMediaQuery';
import { useTheme } from '@mui/material/styles';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getNodeDialogPaperSx, getNodePanelSx, getNodeThemeTokens } from '../nodeTheme';

// icons
import CheckCircleOutlineIcon from '@mui/icons-material/CheckCircleOutline';
import ErrorOutlineIcon from '@mui/icons-material/ErrorOutline';
import SwapHorizIcon from '@mui/icons-material/SwapHoriz';

/**
 * 节点添加结果汇总弹窗
 * 展示添加成功、重复跳过、失败的统计和详情
 */
export default function NodeAddResultDialog({ open, result, onClose }) {
  const { t } = useTranslation();
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const tokens = getNodeThemeTokens(theme, isDark);

  if (!result) return null;

  const { added = 0, skipped = [], failed = [] } = result;
  const total = added + skipped.length + failed.length;
  const allSuccess = skipped.length === 0 && failed.length === 0;

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="sm"
      fullWidth
      fullScreen={isMobile}
      PaperProps={{ sx: getNodeDialogPaperSx(theme, tokens) }}
    >
      <DialogTitle
        sx={{
          pb: 1,
          display: 'flex',
          alignItems: 'center',
          gap: 1,
          color: tokens.primaryText,
          bgcolor: tokens.mutedPanelSurface,
          borderBottom: '1px solid',
          borderColor: tokens.panelBorder
        }}
      >
        {allSuccess ? (
          <CheckCircleOutlineIcon sx={{ color: 'success.main' }} />
        ) : (
          <ErrorOutlineIcon sx={{ color: skipped.length > 0 ? 'warning.main' : 'error.main' }} />
        )}
        {t('nodes.addResult.title')}
      </DialogTitle>

      <DialogContent sx={{ pt: 1, bgcolor: 'transparent' }}>
        {/* 统计概览 */}
        <Stack
          direction="row"
          spacing={1.5}
          sx={{
            mb: 2,
            p: 1.5,
            borderRadius: 2,
            backgroundColor: tokens.nestedPanelSurface,
            border: '1px solid',
            borderColor: tokens.softBorder,
            flexWrap: 'wrap',
            gap: 1
          }}
        >
          <Chip
            icon={<CheckCircleOutlineIcon />}
            label={t('nodes.addResult.stats.added', { count: added })}
            color="success"
            variant={added > 0 ? 'filled' : 'outlined'}
            size={isMobile ? 'small' : 'medium'}
          />
          {skipped.length > 0 && (
            <Chip
              icon={<SwapHorizIcon />}
              label={t('nodes.addResult.stats.skipped', { count: skipped.length })}
              color="warning"
              variant="filled"
              size={isMobile ? 'small' : 'medium'}
            />
          )}
          {failed.length > 0 && (
            <Chip
              icon={<ErrorOutlineIcon />}
              label={t('nodes.addResult.stats.failed', { count: failed.length })}
              color="error"
              variant="filled"
              size={isMobile ? 'small' : 'medium'}
            />
          )}
          <Typography variant="body2" color="text.secondary" sx={{ alignSelf: 'center', ml: 'auto' }}>
            {t('nodes.addResult.stats.total', { count: total })}
          </Typography>
        </Stack>

        {/* 全部成功提示 */}
        {allSuccess && added > 0 && (
          <Alert severity="success" sx={{ borderRadius: 2 }}>
            {t('nodes.addResult.successAll', { count: added })}
          </Alert>
        )}

        {/* 重复跳过详情 */}
        {skipped.length > 0 && (
          <Box sx={{ mb: 2 }}>
            <Typography variant="subtitle2" color="warning.main" sx={{ mb: 1, fontWeight: 600 }}>
              {t('nodes.addResult.skippedTitle')}
            </Typography>
            <List
              dense
              sx={{
                ...getNodePanelSx(theme, tokens, tokens.palette.warning.main, { compact: true }),
                borderRadius: 2,
                overflow: 'hidden',
                '& .MuiListItem-root:not(:last-child)': { borderBottom: `1px solid ${tokens.softBorder}` }
              }}
            >
              {skipped.map((item, index) => (
                <ListItem
                  key={index}
                  sx={{
                    flexDirection: isMobile ? 'column' : 'row',
                    alignItems: isMobile ? 'flex-start' : 'center',
                    py: 1.5,
                    px: 2,
                    gap: isMobile ? 0.5 : 0
                  }}
                >
                  <ListItemIcon sx={{ minWidth: 32, display: isMobile ? 'none' : 'flex' }}>
                    <SwapHorizIcon fontSize="small" color="warning" />
                  </ListItemIcon>
                  <ListItemText
                    primary={
                      <Typography variant="body2" sx={{ fontWeight: 500, wordBreak: 'break-all' }}>
                        {item.name || t('nodes.addResult.unknownNode')}
                      </Typography>
                    }
                    secondary={
                      <Stack direction="row" spacing={0.5} sx={{ mt: 0.5, flexWrap: 'wrap', gap: 0.5 }}>
                        <Typography variant="caption" color="text.secondary">
                          {t('nodes.addResult.duplicatePrefix')}
                        </Typography>
                        <Chip label={item.existingName} size="small" variant="outlined" sx={{ height: 20, fontSize: '0.7rem' }} />
                        <Chip
                          label={t('nodes.addResult.sourceLabel', { source: item.source })}
                          size="small"
                          variant="outlined"
                          color="info"
                          sx={{ height: 20, fontSize: '0.7rem' }}
                        />
                        <Chip
                          label={t('nodes.addResult.groupLabel', { group: item.group })}
                          size="small"
                          variant="outlined"
                          color="secondary"
                          sx={{ height: 20, fontSize: '0.7rem' }}
                        />
                      </Stack>
                    }
                    sx={{ m: 0 }}
                  />
                </ListItem>
              ))}
            </List>
          </Box>
        )}

        {/* 失败详情 */}
        {failed.length > 0 && (
          <Box>
            {skipped.length > 0 && <Divider sx={{ mb: 2 }} />}
            <Typography variant="subtitle2" color="error.main" sx={{ mb: 1, fontWeight: 600 }}>
              {t('nodes.addResult.failedTitle')}
            </Typography>
            <List
              dense
              sx={{
                ...getNodePanelSx(theme, tokens, tokens.palette.error.main, { compact: true }),
                borderRadius: 2,
                overflow: 'hidden',
                '& .MuiListItem-root:not(:last-child)': { borderBottom: `1px solid ${tokens.softBorder}` }
              }}
            >
              {failed.map((item, index) => (
                <ListItem key={index} sx={{ py: 1.5, px: 2 }}>
                  <ListItemIcon sx={{ minWidth: 32 }}>
                    <ErrorOutlineIcon fontSize="small" color="error" />
                  </ListItemIcon>
                  <ListItemText
                    primary={
                      <Typography variant="body2" sx={{ wordBreak: 'break-all' }}>
                        {item.link}
                      </Typography>
                    }
                    secondary={item.error}
                    sx={{ m: 0 }}
                  />
                </ListItem>
              ))}
            </List>
          </Box>
        )}
      </DialogContent>

      <DialogActions sx={{ px: 3, pb: 2, bgcolor: tokens.mutedPanelSurface, borderTop: '1px solid', borderColor: tokens.panelBorder }}>
        <Button variant="contained" onClick={onClose} fullWidth={isMobile}>
          {t('common.confirm')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

NodeAddResultDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  result: PropTypes.shape({
    added: PropTypes.number,
    skipped: PropTypes.arrayOf(
      PropTypes.shape({
        name: PropTypes.string,
        source: PropTypes.string,
        group: PropTypes.string,
        existingName: PropTypes.string
      })
    ),
    failed: PropTypes.arrayOf(
      PropTypes.shape({
        link: PropTypes.string,
        error: PropTypes.string
      })
    )
  }),
  onClose: PropTypes.func.isRequired
};
