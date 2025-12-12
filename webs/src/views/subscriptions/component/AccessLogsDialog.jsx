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
import AccessTimeIcon from '@mui/icons-material/AccessTime';
import LocationOnIcon from '@mui/icons-material/LocationOn';
import TouchAppIcon from '@mui/icons-material/TouchApp';

/**
 * 访问记录对话框 - 响应式设计
 * 桌面端显示表格，移动端显示卡片
 */
export default function AccessLogsDialog({ open, logs, onClose }) {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));

  // 移动端卡片布局
  const MobileLogCard = ({ log }) => (
    <Card
      variant="outlined"
      sx={{
        mb: 1.5,
        borderRadius: 2,
        transition: 'all 0.2s ease',
        '&:hover': {
          boxShadow: 2
        }
      }}
    >
      <CardContent sx={{ py: 1.5, px: 2, '&:last-child': { pb: 1.5 } }}>
        <Stack spacing={1}>
          {/* 第一行: IP 地址和访问次数 */}
          <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 1 }}>
            <Typography
              variant="subtitle2"
              sx={{
                fontFamily: 'monospace',
                fontSize: '0.85rem',
                fontWeight: 600,
                color: 'primary.main',
                wordBreak: 'break-all',
                lineHeight: 1.4,
                flex: 1,
                minWidth: 0
              }}
            >
              {log.IP}
            </Typography>
            <Chip
              size="small"
              label={`${log.Count} 次`}
              color="primary"
              variant="outlined"
              icon={<TouchAppIcon sx={{ fontSize: 14 }} />}
              sx={{
                height: 24,
                '& .MuiChip-label': { px: 1 },
                flexShrink: 0
              }}
            />
          </Box>

          {/* 第二行: 来源地区 */}
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
            <LocationOnIcon sx={{ fontSize: 16, color: 'text.secondary', flexShrink: 0 }} />
            <Typography
              variant="body2"
              color="text.secondary"
              sx={{
                overflow: 'hidden',
                textOverflow: 'ellipsis',
                whiteSpace: 'nowrap'
              }}
            >
              {log.Addr || '未知来源'}
            </Typography>
          </Box>

          {/* 第三行: 访问时间 */}
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
            <AccessTimeIcon sx={{ fontSize: 16, color: 'text.secondary', flexShrink: 0 }} />
            <Typography variant="body2" color="text.secondary">
              {log.Date}
            </Typography>
          </Box>
        </Stack>
      </CardContent>
    </Card>
  );

  // 桌面端表格布局
  const DesktopTable = () => (
    <TableContainer>
      <Table size="small">
        <TableHead>
          <TableRow>
            <TableCell sx={{ fontWeight: 600, minWidth: 140 }}>IP 地址</TableCell>
            <TableCell sx={{ fontWeight: 600, minWidth: 120 }}>来源地区</TableCell>
            <TableCell sx={{ fontWeight: 600, width: 100 }} align="center">
              访问次数
            </TableCell>
            <TableCell sx={{ fontWeight: 600, minWidth: 160 }}>最近访问</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {logs.map((log) => (
            <TableRow
              key={log.ID}
              sx={{
                '&:hover': { bgcolor: 'action.hover' },
                transition: 'background-color 0.2s'
              }}
            >
              <TableCell>
                <Typography
                  variant="body2"
                  sx={{
                    fontFamily: 'monospace',
                    color: 'primary.main',
                    fontWeight: 500
                  }}
                >
                  {log.IP}
                </Typography>
              </TableCell>
              <TableCell>
                <Typography variant="body2" color="text.secondary">
                  {log.Addr || '-'}
                </Typography>
              </TableCell>
              <TableCell align="center">
                <Chip size="small" label={log.Count} color="primary" variant="outlined" sx={{ minWidth: 50 }} />
              </TableCell>
              <TableCell>
                <Typography variant="body2" color="text.secondary">
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
      PaperProps={{
        sx: isMobile ? { borderRadius: 0 } : { borderRadius: 2 }
      }}
    >
      <DialogTitle
        sx={{
          pb: 1,
          borderBottom: '1px solid',
          borderColor: 'divider'
        }}
      >
        <Stack direction="row" alignItems="center" spacing={1}>
          <TouchAppIcon color="primary" />
          <Typography variant="h6">访问记录</Typography>
          {logs.length > 0 && <Chip size="small" label={`共 ${logs.length} 条`} sx={{ ml: 1 }} />}
        </Stack>
      </DialogTitle>
      <DialogContent sx={{ p: isMobile ? 1.5 : 2 }}>
        {logs.length === 0 ? (
          <Box
            sx={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              py: 8,
              color: 'text.secondary'
            }}
          >
            <TouchAppIcon sx={{ fontSize: 48, mb: 2, opacity: 0.5 }} />
            <Typography>暂无访问记录</Typography>
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
      <DialogActions sx={{ borderTop: '1px solid', borderColor: 'divider', px: 2, py: 1.5 }}>
        <Button onClick={onClose} variant="outlined">
          关闭
        </Button>
      </DialogActions>
    </Dialog>
  );
}
