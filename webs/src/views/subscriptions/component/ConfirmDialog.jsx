import { useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogActions from '@mui/material/DialogActions';
import Button from '@mui/material/Button';

import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';

/**
 * 通用确认对话框组件
 */
export default function ConfirmDialog({ open, title, content, onClose, onConfirm }) {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const { isDark } = useResolvedColorScheme();
  const { palette, dialogSurface, dialogSurfaceGradient, mutedPanelSurface, panelBorder } = getSurfaceTokens(theme, isDark);
  const { primaryText, secondaryText } = getReadableTextTokens(theme, isDark);

  const cancelButtonSx = {
    borderColor: withAlpha(palette.divider, isDark ? 0.76 : 0.9),
    color: secondaryText,
    bgcolor: mutedPanelSurface,
    '&:hover': {
      borderColor: withAlpha(palette.primary.main, isDark ? 0.28 : 0.18),
      color: primaryText,
      bgcolor: withAlpha(palette.primary.main, isDark ? 0.12 : 0.05)
    }
  };

  const confirmButtonSx = {
    px: 2,
    backgroundColor: palette.error.main,
    boxShadow: 'none',
    '&:hover': {
      backgroundColor: palette.error.dark,
      boxShadow: 'none'
    }
  };

  return (
    <Dialog
      open={open}
      onClose={onClose}
      aria-labelledby="confirm-dialog-title"
      aria-describedby="confirm-dialog-description"
      slotProps={{
        paper: {
          sx: {
            borderRadius: isMobile ? 2.5 : 3,
            overflow: 'hidden',
            bgcolor: dialogSurface,
            backgroundImage: dialogSurfaceGradient,
            border: '1px solid',
            borderColor: panelBorder
          }
        }
      }}
    >
      <DialogTitle
        id="confirm-dialog-title"
        sx={{
          px: 2.5,
          py: 2,
          bgcolor: mutedPanelSurface,
          borderBottom: '1px solid',
          borderColor: panelBorder,
          boxShadow: `inset 0 -1px 0 ${withAlpha(palette.divider, 0.42)}`,
          color: primaryText,
          fontWeight: 700
        }}
      >
        {title}
      </DialogTitle>
      <DialogContent
        sx={{
          px: 2.5,
          pt: 2.25,
          pb: 2,
          bgcolor: dialogSurface
        }}
      >
        <DialogContentText
          id="confirm-dialog-description"
          sx={{
            m: 0,
            color: secondaryText,
            lineHeight: 1.65
          }}
        >
          {content}
        </DialogContentText>
      </DialogContent>
      <DialogActions
        sx={{
          px: 2.5,
          py: 1.5,
          gap: 1,
          bgcolor: mutedPanelSurface,
          borderTop: '1px solid',
          borderColor: panelBorder
        }}
      >
        <Button onClick={onClose} variant="outlined" sx={cancelButtonSx}>
          取消
        </Button>
        <Button onClick={onConfirm} color="error" variant="contained" autoFocus sx={confirmButtonSx}>
          确定
        </Button>
      </DialogActions>
    </Dialog>
  );
}
