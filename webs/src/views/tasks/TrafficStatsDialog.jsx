import { useState, useCallback, useMemo } from 'react';
import PropTypes from 'prop-types';
import { useTheme, alpha } from '@mui/material/styles';
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
import TablePagination from '@mui/material/TablePagination';
import Paper from '@mui/material/Paper';
import LinearProgress from '@mui/material/LinearProgress';
import Stack from '@mui/material/Stack';
import Grid from '@mui/material/Grid';
import TextField from '@mui/material/TextField';
import IconButton from '@mui/material/IconButton';
import CircularProgress from '@mui/material/CircularProgress';
import useMediaQuery from '@mui/material/useMediaQuery';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import {
  getTaskActionButtonSx,
  getTaskCardSx,
  getTaskCenterTokens,
  getTaskDialogPaperSx,
  getTaskProgressSx,
  getTaskTableContainerSx,
  getTaskTableRowSx,
  TASK_CLUSTER_ACCENT
} from 'components/taskCenterTheme';

import DownloadIcon from '@mui/icons-material/Download';
import SearchIcon from '@mui/icons-material/Search';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';

import { getTaskTrafficDetails } from 'api/tasks';

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
  const { isDark } = useResolvedColorScheme();
  const tokens = getTaskCenterTokens(theme, isDark);
  const fullScreen = useMediaQuery(theme.breakpoints.down('md'));
  const [tabValue, setTabValue] = useState(0);

  // Drill-down state
  const [drillFilter, setDrillFilter] = useState(null);
  const [drillNodes, setDrillNodes] = useState([]);
  const [drillLoading, setDrillLoading] = useState(false);
  const [drillTotal, setDrillTotal] = useState(0);
  const [drillPage, setDrillPage] = useState(0);
  const [drillPageSize, setDrillPageSize] = useState(50);
  const [drillSearch, setDrillSearch] = useState('');
  const [drillSearchInput, setDrillSearchInput] = useState('');

  // Parse traffic data - memoized to avoid re-parsing
  const trafficData = useMemo(() => {
    if (!task) return null;
    try {
      const result = typeof task.result === 'string' ? JSON.parse(task.result) : task.result;
      return result?.traffic || null;
    } catch (e) {
      console.error('Failed to parse task result:', e);
      return null;
    }
  }, [task]);

  // Load drill-down node data - useCallback MUST be before any returns
  const loadDrillNodes = useCallback(
    async (filterType, filterValue, page = 0, search = '', pageSize = 50) => {
      if (!task?.id) return;
      setDrillLoading(true);
      try {
        const params = {
          page: page + 1,
          pageSize: pageSize,
          search: search
        };
        if (filterType === 'group') {
          params.group = filterValue;
        } else if (filterType === 'source') {
          params.source = filterValue;
        }
        const res = await getTaskTrafficDetails(task.id, params);
        if (res.code === 200 || res.code === 0) {
          setDrillNodes(res.data.nodes || []);
          setDrillTotal(res.data.total || 0);
        }
      } catch (error) {
        console.error('Failed to load node traffic:', error);
      } finally {
        setDrillLoading(false);
      }
    },
    [task?.id]
  );

  // Helper functions
  const calculatePercent = useCallback(
    (bytes) => {
      if (!trafficData?.totalBytes) return 0;
      return (bytes / trafficData.totalBytes) * 100;
    },
    [trafficData?.totalBytes]
  );

  // Check which tabs have data
  const hasGroupData = useMemo(() => trafficData?.byGroup && Object.keys(trafficData.byGroup).length > 0, [trafficData]);
  const hasSourceData = useMemo(() => trafficData?.bySource && Object.keys(trafficData.bySource).length > 0, [trafficData]);
  const hasNodeData = useMemo(() => trafficData?.byNode && Object.keys(trafficData.byNode).length > 0, [trafficData]);

  // NOW we can have early returns - all hooks have been called
  if (!task || !trafficData) {
    return (
      <Dialog open={open} onClose={onClose} fullWidth maxWidth="sm" PaperProps={{ sx: getTaskDialogPaperSx(theme, tokens) }}>
        <DialogTitle sx={{ color: tokens.primaryText }}>流量统计</DialogTitle>
        <DialogContent>
          <Typography sx={{ color: tokens.secondaryText }} textAlign="center" py={4}>
            无流量数据
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button
            onClick={onClose}
            variant="contained"
            sx={getTaskActionButtonSx(theme, tokens, TASK_CLUSTER_ACCENT, { variant: 'solid' })}
          >
            关闭
          </Button>
        </DialogActions>
      </Dialog>
    );
  }

  const handleTabChange = (event, newValue) => {
    setTabValue(newValue);
    setDrillFilter(null);
    setDrillNodes([]);
  };

  // Handle row click to drill-down
  const handleDrillDown = async (type, value) => {
    setDrillFilter({ type, value });
    setDrillPage(0);
    setDrillSearch('');
    setDrillSearchInput('');
    await loadDrillNodes(type, value, 0, '', drillPageSize);
  };

  // Handle back from drill-down
  const handleBackFromDrill = () => {
    setDrillFilter(null);
    setDrillNodes([]);
  };

  // Handle drill-down pagination
  const handleDrillPageChange = (event, newPage) => {
    setDrillPage(newPage);
    if (drillFilter) {
      loadDrillNodes(drillFilter.type, drillFilter.value, newPage, drillSearch, drillPageSize);
    }
  };

  // Handle drill-down search
  const handleDrillSearch = () => {
    setDrillSearch(drillSearchInput);
    setDrillPage(0);
    if (drillFilter) {
      loadDrillNodes(drillFilter.type, drillFilter.value, 0, drillSearchInput, drillPageSize);
    }
  };

  // Export to CSV
  const exportToCSV = (data, filename, isNodeData = false) => {
    let csvContent = '';
    if (isNodeData) {
      csvContent = '节点名称,原始名称,分组,来源,流量(字节),流量\n';
      data.forEach((node) => {
        csvContent += `"${node.name}","${node.originName}","${node.group}","${node.source}",${node.bytes},"${node.formatted}"\n`;
      });
    } else {
      csvContent = '名称,流量(字节),流量,占比\n';
      Object.entries(data)
        .sort((a, b) => b[1].bytes - a[1].bytes)
        .forEach(([name, info]) => {
          const percent = calculatePercent(info.bytes);
          csvContent += `"${name}",${info.bytes},"${info.formatted}",${percent.toFixed(2)}%\n`;
        });
    }

    const blob = new Blob(['\ufeff' + csvContent], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement('a');
    link.href = URL.createObjectURL(blob);
    link.download = `${filename}_${new Date().toISOString().slice(0, 10)}.csv`;
    link.click();
  };

  // Render summary stats table (group or source)
  const renderStatsTable = (statsMap, labelHeader, filterType) => {
    if (!statsMap) return <Typography sx={{ color: tokens.secondaryText }}>无数据</Typography>;

    const entries = Object.entries(statsMap).sort((a, b) => b[1].bytes - a[1].bytes);

    return (
      <Box>
        {hasNodeData && (
          <Typography variant="caption" sx={{ display: 'block', mb: 1, color: tokens.secondaryText }}>
            点击行可查看该{filterType === 'group' ? '分组' : '来源'}下的节点流量详情
          </Typography>
        )}
        <TableContainer component={Paper} variant="outlined" sx={getTaskTableContainerSx(theme, tokens)}>
          <Table size="small">
            <TableHead sx={{ bgcolor: tokens.tableHeaderSurface }}>
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
                  <TableRow
                    key={name}
                    hover
                    sx={getTaskTableRowSx(tokens, hasNodeData)}
                    onClick={() => hasNodeData && handleDrillDown(filterType, name)}
                  >
                    <TableCell component="th" scope="row">
                      <Stack direction="row" alignItems="center" spacing={1}>
                        <Typography variant="body2" fontWeight={500} sx={{ color: tokens.primaryText }}>
                          {name}
                        </Typography>
                        {hasNodeData && <ExpandMoreIcon fontSize="small" color="action" />}
                      </Stack>
                    </TableCell>
                    <TableCell align="right">
                      <Typography variant="body2" color="primary">
                        {data.formatted}
                      </Typography>
                    </TableCell>
                    <TableCell align="right">
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                        <Box sx={{ width: '100%', mr: 1 }}>
                          <LinearProgress
                            variant="determinate"
                            value={percent}
                            sx={getTaskProgressSx(tokens, theme.palette.primary.main)}
                          />
                        </Box>
                        <Typography variant="caption" sx={{ minWidth: 35, color: tokens.secondaryText }}>
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
        <Box sx={{ mt: 1, display: 'flex', justifyContent: 'flex-end' }}>
          <Button
            size="small"
            startIcon={<DownloadIcon />}
            onClick={() => exportToCSV(statsMap, `traffic_by_${filterType}`)}
            sx={getTaskActionButtonSx(theme, tokens, theme.palette.primary.main)}
          >
            导出CSV
          </Button>
        </Box>
      </Box>
    );
  };

  // Render drill-down node list
  const renderDrillDownNodes = () => (
    <Box>
      <Stack direction="row" alignItems="center" spacing={2} mb={2}>
        <IconButton
          onClick={handleBackFromDrill}
          size="small"
          sx={{
            bgcolor: alpha(theme.palette.primary.main, tokens.isDark ? 0.14 : 0.08),
            border: '1px solid',
            borderColor: alpha(theme.palette.primary.main, tokens.isDark ? 0.24 : 0.16),
            color: theme.palette.primary.main,
            '&:hover': {
              bgcolor: alpha(theme.palette.primary.main, tokens.isDark ? 0.2 : 0.12)
            }
          }}
        >
          <ArrowBackIcon />
        </IconButton>
        <Typography variant="subtitle1" fontWeight={500} sx={{ color: tokens.primaryText }}>
          {drillFilter?.type === 'group' ? '分组' : '来源'}: {drillFilter?.value}
        </Typography>
      </Stack>

      <Stack direction="row" spacing={1} mb={2}>
        <TextField
          size="small"
          placeholder="搜索节点名称..."
          value={drillSearchInput}
          onChange={(e) => setDrillSearchInput(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleDrillSearch()}
          sx={{
            flex: 1,
            '& .MuiOutlinedInput-root': {
              bgcolor: tokens.nestedInteractiveSurface
            }
          }}
        />
        <Button
          variant="outlined"
          startIcon={<SearchIcon />}
          onClick={handleDrillSearch}
          sx={getTaskActionButtonSx(theme, tokens, theme.palette.primary.main)}
        >
          搜索
        </Button>
      </Stack>

      {drillLoading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
          <CircularProgress />
        </Box>
      ) : (
        <>
          <TableContainer component={Paper} variant="outlined" sx={getTaskTableContainerSx(theme, tokens)}>
            <Table size="small">
              <TableHead sx={{ bgcolor: tokens.tableHeaderSurface }}>
                <TableRow>
                  <TableCell>节点名称</TableCell>
                  <TableCell>原始名称</TableCell>
                  <TableCell>分组</TableCell>
                  <TableCell>来源</TableCell>
                  <TableCell align="right">流量</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {drillNodes.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={5} align="center">
                      <Typography sx={{ color: tokens.secondaryText }}>无数据</Typography>
                    </TableCell>
                  </TableRow>
                ) : (
                  drillNodes.map((node) => (
                    <TableRow key={node.nodeId} hover sx={getTaskTableRowSx(tokens, false)}>
                      <TableCell>
                        <Typography variant="body2" fontWeight={500} noWrap sx={{ maxWidth: 150, color: tokens.primaryText }}>
                          {node.name}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2" noWrap sx={{ maxWidth: 150, color: tokens.secondaryText }}>
                          {node.originName || '-'}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2" sx={{ color: tokens.primaryText }}>
                          {node.group || '-'}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2" sx={{ color: tokens.primaryText }}>
                          {node.source || '-'}
                        </Typography>
                      </TableCell>
                      <TableCell align="right">
                        <Typography variant="body2" color="primary" fontWeight={500}>
                          {node.formatted}
                        </Typography>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </TableContainer>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mt: 1 }}>
            <Button
              size="small"
              startIcon={<DownloadIcon />}
              onClick={() => exportToCSV(drillNodes, `traffic_nodes_${drillFilter?.value}`, true)}
              sx={getTaskActionButtonSx(theme, tokens, theme.palette.primary.main)}
            >
              导出CSV
            </Button>
            <TablePagination
              component="div"
              count={drillTotal}
              page={drillPage}
              onPageChange={handleDrillPageChange}
              rowsPerPage={drillPageSize}
              onRowsPerPageChange={(e) => {
                const newSize = parseInt(e.target.value, 10);
                setDrillPageSize(newSize);
                setDrillPage(0);
                if (drillFilter) {
                  loadDrillNodes(drillFilter.type, drillFilter.value, 0, drillSearch, newSize);
                }
              }}
              rowsPerPageOptions={[20, 50, 100]}
              labelRowsPerPage="每页"
            />
          </Box>
        </>
      )}
    </Box>
  );

  return (
    <Dialog
      open={open}
      onClose={onClose}
      fullWidth
      maxWidth="md"
      fullScreen={fullScreen}
      PaperProps={{ sx: getTaskDialogPaperSx(theme, tokens) }}
    >
      <DialogTitle sx={{ color: tokens.primaryText }}>
        <Stack direction="column" spacing={1}>
          <Typography variant="h4">流量统计详情</Typography>
          <Typography variant="subtitle2" sx={{ color: tokens.secondaryText }}>
            任务: {task.name}
          </Typography>
        </Stack>
      </DialogTitle>

      <DialogContent dividers>
        <Grid container spacing={2} sx={{ mb: 3 }}>
          <Grid item xs={12}>
            <Paper
              variant="outlined"
              sx={{ ...getTaskCardSx(theme, tokens, theme.palette.primary.main, { interactive: false }), p: 2, textAlign: 'center' }}
            >
              <Typography variant="subtitle2" sx={{ color: tokens.secondaryText }} gutterBottom>
                总消耗流量
              </Typography>
              <Typography variant="h2" color="primary">
                {trafficData.totalFormatted}
              </Typography>
            </Paper>
          </Grid>
        </Grid>

        {!hasGroupData && !hasSourceData && !hasNodeData ? (
          <Typography variant="body2" sx={{ color: tokens.secondaryText }} textAlign="center">
            未开启详细流量统计，可在测速设置中开启
          </Typography>
        ) : drillFilter ? (
          renderDrillDownNodes()
        ) : (
          <>
            <Box sx={{ borderBottom: '1px solid', borderColor: tokens.softBorder }}>
              <Tabs
                value={tabValue}
                onChange={handleTabChange}
                aria-label="traffic stats tabs"
                sx={{
                  '& .MuiTab-root': {
                    color: tokens.secondaryText,
                    minHeight: 44
                  },
                  '& .Mui-selected': {
                    color: theme.palette.primary.main
                  }
                }}
              >
                {hasGroupData && <Tab label="分组统计" />}
                {hasSourceData && <Tab label="来源统计" />}
              </Tabs>
            </Box>
            {hasGroupData && (
              <TabPanel value={tabValue} index={0}>
                {renderStatsTable(trafficData.byGroup, '分组名称', 'group')}
              </TabPanel>
            )}
            {hasSourceData && (
              <TabPanel value={tabValue} index={hasGroupData ? 1 : 0}>
                {renderStatsTable(trafficData.bySource, '来源名称', 'source')}
              </TabPanel>
            )}
          </>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} variant="contained" sx={getTaskActionButtonSx(theme, tokens, TASK_CLUSTER_ACCENT, { variant: 'solid' })}>
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
