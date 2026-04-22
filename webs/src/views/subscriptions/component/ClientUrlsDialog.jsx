import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import Button from '@mui/material/Button';
import Stack from '@mui/material/Stack';
import Chip from '@mui/material/Chip';
import IconButton from '@mui/material/IconButton';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import useMediaQuery from '@mui/material/useMediaQuery';
import { useTheme } from '@mui/material/styles';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import QrCodeIcon from '@mui/icons-material/QrCode';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';

export default function ClientUrlsDialog({
  open,
  title = '客户端（点击二维码获取地址）',
  subtitle = '',
  legacy = false,
  clientUrls,
  onClose,
  onQrCode,
  onCopy
}) {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const { isDark } = useResolvedColorScheme();
  const { palette, dialogSurface, dialogSurfaceGradient, mutedPanelSurface, nestedPanelSurface, panelBorder } = getSurfaceTokens(
    theme,
    isDark
  );
  const { primaryText, secondaryText } = getReadableTextTokens(theme, isDark);

  const rowBorder = isDark ? withAlpha(palette.divider, 0.58) : withAlpha(palette.divider, 0.82);
  const chipSurface = withAlpha(palette.success.main, isDark ? 0.16 : 0.08);
  const chipBorder = withAlpha(palette.success.main, isDark ? 0.32 : 0.16);
  const actionButtonSx = {
    color: secondaryText,
    border: '1px solid',
    borderColor: rowBorder,
    bgcolor: mutedPanelSurface,
    '&:hover': {
      color: primaryText,
      bgcolor: withAlpha(palette.primary.main, isDark ? 0.14 : 0.06),
      borderColor: withAlpha(palette.primary.main, isDark ? 0.26 : 0.14)
    }
  };

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="sm"
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
        <Stack direction="row" spacing={1} alignItems="center" useFlexGap flexWrap="wrap">
          <Typography variant="h6" component="span" sx={{ color: 'inherit' }}>
            {title}
          </Typography>
          {legacy && (
            <Chip
              label="默认"
              size="small"
              sx={{
                height: 20,
                fontSize: '0.68rem',
                fontWeight: 700,
                bgcolor: withAlpha(palette.primary.main, isDark ? 0.18 : 0.1),
                color: palette.primary.main,
                border: '1px solid',
                borderColor: withAlpha(palette.primary.main, isDark ? 0.38 : 0.22),
                '& .MuiChip-label': { px: 0.9 }
              }}
            />
          )}
        </Stack>
      </DialogTitle>
      <DialogContent
        sx={{
          px: 2.5,
          py: 2,
          bgcolor: dialogSurface
        }}
      >
        <Stack spacing={1.5}>
          {subtitle ? (
            <Typography variant="body2" sx={{ color: secondaryText }}>
              {subtitle}
            </Typography>
          ) : null}
          {Object.entries(clientUrls).map(([name, url]) => {
            const resolvedUrl = url;

            return (
              <Stack
                key={name}
                direction={isMobile ? 'column' : 'row'}
                alignItems={isMobile ? 'stretch' : 'center'}
                spacing={1.25}
                sx={{
                  p: 1.5,
                  borderRadius: 2.5,
                  bgcolor: nestedPanelSurface,
                  border: '1px solid',
                  borderColor: rowBorder
                }}
              >
                <Chip
                  label={name}
                  sx={{
                    minWidth: isMobile ? 0 : 100,
                    alignSelf: isMobile ? 'flex-start' : 'center',
                    bgcolor: chipSurface,
                    color: palette.success.main,
                    border: '1px solid',
                    borderColor: chipBorder,
                    fontWeight: 600
                  }}
                />
                <Box sx={{ flex: 1, minWidth: 0 }}>
                  <Typography
                    variant="body2"
                    sx={{
                      fontSize: '0.8rem',
                      color: secondaryText,
                      wordBreak: 'break-all'
                    }}
                  >
                    {resolvedUrl}
                  </Typography>
                </Box>
                <Stack direction="row" spacing={1} justifyContent={isMobile ? 'flex-end' : 'flex-start'}>
                  <Button variant="outlined" startIcon={<QrCodeIcon />} onClick={() => onQrCode(resolvedUrl, name)}>
                    二维码
                  </Button>
                  <IconButton size="small" onClick={() => onCopy(resolvedUrl)} sx={actionButtonSx}>
                    <ContentCopyIcon fontSize="small" />
                  </IconButton>
                </Stack>
              </Stack>
            );
          })}
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
        <Button onClick={onClose} variant="outlined">
          关闭
        </Button>
      </DialogActions>
    </Dialog>
  );
}
