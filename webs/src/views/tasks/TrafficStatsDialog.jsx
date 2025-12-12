import { useState } from 'react';
import PropTypes from 'prop-types';
import { useTheme } from '@mui/material/styles';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import Button from '@mui/material/Button';
import Tabs from '@mui/material/Tabs';
import Tab from '@mui/material/Tab';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';
import LinearProgress from '@mui/material/LinearProgress';
import Stack from '@mui/material/Stack';
import Chip from '@mui/material/Chip';
import Grid from '@mui/material/Grid';
import useMediaQuery from '@mui/material/useMediaQuery';

// ==============================|| TAB PANEL ||============================== //

function TabPanel({ children, value, index, ...other }) {
  return (
    <div role="tabpanel" hidden={value !== index} id={`traffic-tabpanel-${index}`} aria-labelledby={`traffic-tab-${index}`} {...other}>
      {value === index && <Box sx={{ pt: 2 }}>{children}</Box>}
    </div>
  );
}

TabPanel.propTypes = {
  children: PropTypes.node,
  index: PropTypes.number.isRequired,
  value: PropTypes.number.isRequired
};

// ==============================|| TRAFFIC STATS DIALOG ||============================== //

export default function TrafficStatsDialog({ open, onClose, task }) {
  const theme = useTheme();
  const fullScreen = useMediaQuery(theme.breakpoints.down('md'));
  const [tabValue, setTabValue] = useState(0);

  if (!task) return null;

  let trafficData = null;
  try {
    const result = typeof task.result === 'string' ? JSON.parse(task.result) : task.result;
    trafficData = result?.traffic;
  } catch (e) {
    console.error('Failed to parse task result:', e);
  }

  if (!trafficData) return null;

  const handleTabChange = (event, newValue) => {
    setTabValue(newValue);
  };

  // Helper to calculate percentage
  const calculatePercent = (bytes) => {
    if (!trafficData.totalBytes) return 0;
    return (bytes / trafficData.totalBytes) * 100;
  };

  // Render Stats Table
  const renderStatsTable = (statsMap, labelHeader) => {
    if (!statsMap) return <Typography color="textSecondary">无数据</Typography>;

    const entries = Object.entries(statsMap).sort((a, b) => b[1].bytes - a[1].bytes);

    return (
      <TableContainer component={Paper} variant="outlined">
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>{labelHeader}</TableCell>
              <TableCell align="right">流量</TableCell>
              <TableCell align="right" width="40%">
                占比
              </TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {entries.map(([name, data]) => {
              const percent = calculatePercent(data.bytes);
              return (
                <TableRow key={name} hover>
                  <TableCell component="th" scope="row">
                    <Typography variant="body2" fontWeight={500}>
                      {name}
                    </Typography>
                  </TableCell>
                  <TableCell align="right">
                    <Typography variant="body2" color="primary">
                      {data.formatted}
                    </Typography>
                  </TableCell>
                  <TableCell align="right">
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                      <Box sx={{ width: '100%', mr: 1 }}>
                        <LinearProgress variant="determinate" value={percent} sx={{ height: 6, borderRadius: 1 }} />
                      </Box>
                      <Typography variant="caption" color="textSecondary" sx={{ minWidth: 35 }}>
                        {Math.round(percent)}%
                      </Typography>
                    </Box>
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </TableContainer>
    );
  };

  return (
    <Dialog open={open} onClose={onClose} fullWidth maxWidth="sm" fullScreen={fullScreen}>
      <DialogTitle>
        <Stack direction="column" spacing={1}>
          <Typography variant="h4">流量统计详情</Typography>
          <Typography variant="subtitle2" color="textSecondary">
            任务: {task.name}
          </Typography>
        </Stack>
      </DialogTitle>

      <DialogContent dividers>
        {/* Summary Cards */}
        <Grid container spacing={2} sx={{ mb: 3 }}>
          <Grid item xs={12}>
            <Paper
              variant="outlined"
              sx={{ p: 2, textAlign: 'center', bgcolor: theme.palette.mode === 'dark' ? 'background.default' : 'primary.lighter' }}
            >
              <Typography variant="subtitle2" color="textSecondary" gutterBottom>
                总消耗流量
              </Typography>
              <Typography variant="h2" color="primary">
                {trafficData.totalFormatted}
              </Typography>
            </Paper>
          </Grid>
        </Grid>

        <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
          <Tabs value={tabValue} onChange={handleTabChange} aria-label="traffic stats tabs">
            <Tab label="分组统计" />
            <Tab label="来源统计" />
          </Tabs>
        </Box>
        <TabPanel value={tabValue} index={0}>
          {renderStatsTable(trafficData.byGroup, '分组名称')}
        </TabPanel>
        <TabPanel value={tabValue} index={1}>
          {renderStatsTable(trafficData.bySource, '来源名称')}
        </TabPanel>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} variant="contained">
          关闭
        </Button>
      </DialogActions>
    </Dialog>
  );
}

TrafficStatsDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  task: PropTypes.object
};
