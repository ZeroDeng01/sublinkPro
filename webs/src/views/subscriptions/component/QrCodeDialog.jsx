import { QRCodeSVG } from 'qrcode.react';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import Stack from '@mui/material/Stack';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import useMediaQuery from '@mui/material/useMediaQuery';
import { useTheme } from '@mui/material/styles';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';

export default function QrCodeDialog({ open, title, url, onClose, onCopy }) {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const { isDark } = useResolvedColorScheme();
  const { palette, dialogSurface, dialogSurfaceGradient, mutedPanelSurface, nestedPanelSurface, panelBorder } = getSurfaceTokens(
    theme,
    isDark
  );
  const { primaryText, secondaryText } = getReadableTextTokens(theme, isDark);

  const qrPanelBorder = isDark ? withAlpha(palette.divider, 0.58) : withAlpha(palette.divider, 0.82);
  const qrCanvasBackground = isDark ? palette.common.white : palette.background.paper;

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="xs"
      fullWidth
      fullScreen={isMobile}
      slotProps={{
        paper: {
          sx: {
            borderRadius: isMobile ? 0 : 3,
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
        sx={{
          px: 2.5,
          py: 2,
          bgcolor: mutedPanelSurface,
          borderBottom: '1px solid',
          borderColor: panelBorder,
          boxShadow: `inset 0 -1px 0 ${withAlpha(palette.divider, 0.42)}`,
          color: primaryText
        }}
      >
        {title}
      </DialogTitle>
      <DialogContent
        sx={{
          px: 2.5,
          py: 2.5,
          bgcolor: dialogSurface
        }}
      >
        <Stack spacing={2.5} alignItems="center">
          <Box
            sx={{
              width: '100%',
              borderRadius: 3,
              bgcolor: nestedPanelSurface,
              border: '1px solid',
              borderColor: qrPanelBorder,
              px: 2,
              py: 2.5,
              display: 'flex',
              justifyContent: 'center'
            }}
          >
            <Box
              sx={{
                p: 1.5,
                borderRadius: 2,
                bgcolor: qrCanvasBackground,
                boxShadow: isDark
                  ? `0 10px 24px ${withAlpha(palette.common.black, 0.28)}`
                  : `0 10px 24px ${withAlpha(palette.common.black, 0.08)}`
              }}
            >
              <QRCodeSVG value={url} size={200} bgColor={qrCanvasBackground} fgColor="#000000" />
            </Box>
          </Box>
          <Box sx={{ width: '100%' }}>
            <Typography variant="caption" sx={{ display: 'block', mb: 0.75, color: secondaryText }}>
              链接地址
            </Typography>
            <TextField fullWidth value={url} size="small" InputProps={{ readOnly: true }} />
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
        <Button onClick={() => onCopy(url)}>复制</Button>
        <Button onClick={() => window.open(url)}>打开</Button>
        <Button onClick={onClose} variant="outlined">
          关闭
        </Button>
      </DialogActions>
    </Dialog>
  );
}
