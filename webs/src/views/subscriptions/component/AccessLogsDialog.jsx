import { useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import Button from '@mui/material/Button';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Stack from '@mui/material/Stack';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import AccessTimeIcon from '@mui/icons-material/AccessTime';
import LocationOnIcon from '@mui/icons-material/LocationOn';
import TouchAppIcon from '@mui/icons-material/TouchApp';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';

export default function AccessLogsDialog({ open, logs, onClose, loading = false, title = '访问记录' }) {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));
  const { isDark } = useResolvedColorScheme();
  const { palette, dialogSurface, dialogSurfaceGradient, mutedPanelSurface, nestedPanelSurface, panelBorder } = getSurfaceTokens(
    theme,
    isDark
  );
  const { primaryText, secondaryText, tertiaryText } = getReadableTextTokens(theme, isDark);

  const rowHoverSurface = withAlpha(palette.primary.main, isDark ? 0.12 : 0.05);
  const rowBorder = isDark ? withAlpha(palette.divider, 0.58) : withAlpha(palette.divider, 0.78);
  const ipSurface = isDark ? withAlpha(palette.primary.main, 0.14) : withAlpha(palette.primary.main, 0.06);
  const ipBorder = withAlpha(palette.primary.main, isDark ? 0.26 : 0.16);
  const countChipSurface = withAlpha(palette.primary.main, isDark ? 0.18 : 0.08);
  const countChipBorder = withAlpha(palette.primary.main, isDark ? 0.34 : 0.18);

  const dialogPaperSx = {
    borderRadius: isMobile ? 0 : 3,
    overflow: 'hidden',
    bgcolor: dialogSurface,
    backgroundImage: dialogSurfaceGradient,
    border: '1px solid',
    borderColor: panelBorder
  };

  const titleSx = {
    px: 2.5,
    py: 2,
    bgcolor: mutedPanelSurface,
    borderBottom: '1px solid',
    borderColor: panelBorder,
    boxShadow: `inset 0 -1px 0 ${withAlpha(palette.divider, 0.42)}`
  };

  const actionsSx = {
    px: 2.5,
    py: 1.5,
    bgcolor: mutedPanelSurface,
    borderTop: '1px solid',
    borderColor: panelBorder
  };

  const countChipSx = {
    ml: 1,
    bgcolor: countChipSurface,
    color: palette.primary.main,
    border: '1px solid',
    borderColor: countChipBorder,
    fontWeight: 600
  };

  const renderIpBlock = (ip) => (
    <Box
      sx={{
        display: 'inline-flex',
        alignItems: 'center',
        maxWidth: '100%',
        px: 1,
        py: 0.5,
        borderRadius: 1.25,
        bgcolor: ipSurface,
        border: '1px solid',
        borderColor: ipBorder
      }}
    >
      <Typography
        variant="body2"
        sx={{
          fontFamily: 'monospace',
          color: palette.primary.main,
          fontWeight: 600,
          wordBreak: 'break-all'
        }}
      >
        {ip}
      </Typography>
    </Box>
  );

  const MobileLogCard = ({ log }) => (
    <Card
      sx={{
        mb: 1.5,
        borderRadius: 2.5,
        bgcolor: nestedPanelSurface,
        border: '1px solid',
        borderColor: rowBorder,
        transition: 'all 0.2s ease',
        '&:hover': {
          bgcolor: rowHoverSurface,
          borderColor: withAlpha(palette.primary.main, isDark ? 0.24 : 0.14)
        }
      }}
    >
      <CardContent sx={{ py: 1.5, px: 2, '&:last-child': { pb: 1.5 } }}>
        <Stack spacing={1}>
          <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 1 }}>
            <Box sx={{ flex: 1, minWidth: 0 }}>{renderIpBlock(log.IP)}</Box>
            <Chip
              size="small"
              label={`${log.Count} 次`}
              icon={<TouchAppIcon sx={{ fontSize: 14 }} />}
              sx={{
                height: 24,
                bgcolor: countChipSurface,
                color: palette.primary.main,
                border: '1px solid',
                borderColor: countChipBorder,
                '& .MuiChip-label': { px: 1 },
                flexShrink: 0
              }}
            />
          </Box>

          <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
            <LocationOnIcon sx={{ fontSize: 16, color: tertiaryText, flexShrink: 0 }} />
            <Typography
              variant="body2"
              sx={{
                color: secondaryText,
                overflow: 'hidden',
                textOverflow: 'ellipsis',
                whiteSpace: 'nowrap'
              }}
            >
              {log.Addr || '未知来源'}
            </Typography>
          </Box>

          <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
            <AccessTimeIcon sx={{ fontSize: 16, color: tertiaryText, flexShrink: 0 }} />
            <Typography variant="body2" sx={{ color: secondaryText }}>
              {log.Date}
            </Typography>
          </Box>
        </Stack>
      </CardContent>
    </Card>
  );

  const DesktopTable = () => (
    <TableContainer
      sx={{
        borderRadius: 2.5,
        bgcolor: nestedPanelSurface,
        border: '1px solid',
        borderColor: rowBorder,
        overflow: 'hidden'
      }}
    >
      <Table size="small">
        <TableHead>
          <TableRow
            sx={{
              bgcolor: mutedPanelSurface,
              '& .MuiTableCell-root': {
                borderColor: rowBorder
              }
            }}
          >
            <TableCell sx={{ fontWeight: 600, minWidth: 140, color: secondaryText }}>IP 地址</TableCell>
            <TableCell sx={{ fontWeight: 600, minWidth: 120, color: secondaryText }}>来源地区</TableCell>
            <TableCell sx={{ fontWeight: 600, width: 100, color: secondaryText }} align="center">
              访问次数
            </TableCell>
            <TableCell sx={{ fontWeight: 600, minWidth: 160, color: secondaryText }}>最近访问</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {logs.map((log) => (
            <TableRow
              key={log.ID}
              sx={{
                transition: 'background-color 0.2s ease',
                '&:hover': { bgcolor: rowHoverSurface },
                '& .MuiTableCell-root': {
                  borderColor: rowBorder
                }
              }}
            >
              <TableCell>{renderIpBlock(log.IP)}</TableCell>
              <TableCell>
                <Typography variant="body2" sx={{ color: secondaryText }}>
                  {log.Addr || '-'}
                </Typography>
              </TableCell>
              <TableCell align="center">
                <Chip size="small" label={log.Count} sx={{ ...countChipSx, minWidth: 50 }} />
              </TableCell>
              <TableCell>
                <Typography variant="body2" sx={{ color: secondaryText }}>
                  {log.Date}
                </Typography>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="md"
      fullWidth
      fullScreen={isMobile}
      slotProps={{
        paper: {
          sx: dialogPaperSx
        }
      }}
    >
      <DialogTitle sx={titleSx}>
        <Stack direction="row" alignItems="center" spacing={1}>
          <TouchAppIcon sx={{ color: palette.primary.main }} />
          <Typography variant="h6" sx={{ color: primaryText }}>
            {title}
          </Typography>
          {!loading && logs.length > 0 && <Chip size="small" label={`共 ${logs.length} 条`} sx={countChipSx} />}
        </Stack>
      </DialogTitle>
      <DialogContent
        sx={{
          p: isMobile ? 1.5 : 2,
          bgcolor: dialogSurface
        }}
      >
        {loading ? (
          <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', py: 8 }}>
            <CircularProgress size={28} />
          </Box>
        ) : logs.length === 0 ? (
          <Box
            sx={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              py: 8,
              borderRadius: 2.5,
              bgcolor: nestedPanelSurface,
              border: '1px solid',
              borderColor: rowBorder,
              color: secondaryText
            }}
          >
            <TouchAppIcon sx={{ fontSize: 48, mb: 2, opacity: 0.5, color: tertiaryText }} />
            <Typography sx={{ color: secondaryText }}>暂无访问记录</Typography>
          </Box>
        ) : isMobile ? (
          <Box sx={{ mt: 1 }}>
            {logs.map((log) => (
              <MobileLogCard key={log.ID} log={log} />
            ))}
          </Box>
        ) : (
          <DesktopTable />
        )}
      </DialogContent>
      <DialogActions sx={actionsSx}>
        <Button onClick={onClose} variant="outlined">
          关闭
        </Button>
      </DialogActions>
    </Dialog>
  );
}
