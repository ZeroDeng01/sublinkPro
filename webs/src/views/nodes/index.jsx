import { useState, useEffect, useMemo, useCallback, useRef } from 'react';

// material-ui
import { useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogActions from '@mui/material/DialogActions';
import TextField from '@mui/material/TextField';
import IconButton from '@mui/material/IconButton';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import TablePagination from '@mui/material/TablePagination';
import TableSortLabel from '@mui/material/TableSortLabel';
import Paper from '@mui/material/Paper';
import Chip from '@mui/material/Chip';
import Stack from '@mui/material/Stack';
import Alert from '@mui/material/Alert';
import Snackbar from '@mui/material/Snackbar';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Checkbox from '@mui/material/Checkbox';
import FormControlLabel from '@mui/material/FormControlLabel';
import Radio from '@mui/material/Radio';
import RadioGroup from '@mui/material/RadioGroup';
import Switch from '@mui/material/Switch';
import Typography from '@mui/material/Typography';
import Autocomplete from '@mui/material/Autocomplete';
import Tooltip from '@mui/material/Tooltip';
import InputAdornment from '@mui/material/InputAdornment';

// icons
import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import RefreshIcon from '@mui/icons-material/Refresh';
import SpeedIcon from '@mui/icons-material/Speed';
import SettingsIcon from '@mui/icons-material/Settings';
import DownloadIcon from '@mui/icons-material/Download';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import {
  getNodes,
  addNodes,
  updateNode,
  deleteNode,
  deleteNodesBatch,
  getSpeedTestConfig,
  updateSpeedTestConfig,
  runSpeedTest,
  getNodeCountries,
  getNodeGroups,
  getNodeSources
} from 'api/nodes';
import { getSubSchedulers, addSubScheduler, updateSubScheduler, deleteSubScheduler, pullSubScheduler } from 'api/scheduler';

// Cron 表达式预设 - 包含友好的说明
const CRON_OPTIONS = [
  { label: '每小时 - 每个整点执行', value: '0 * * * *', description: '每小时0分执行' },
  { label: '每2小时 - 每隔2小时执行', value: '0 */2 * * *', description: '0点、2点、4点...' },
  { label: '每6小时 - 每隔6小时执行', value: '0 */6 * * *', description: '0点、6点、12点、18点' },
  { label: '每12小时 - 每天2次', value: '0 */12 * * *', description: '0点、12点' },
  { label: '每天0点 - 每天凌晨执行', value: '0 0 * * *', description: '每天午夜0点整' },
  { label: '每天3点 - 每天凌晨3点执行', value: '0 3 * * *', description: '每天凌晨3点' },
  { label: '每周一 - 每周一凌晨执行', value: '0 0 * * 1', description: '每周一凌晨0点' }
];

// 测速URL选项 - TCP模式 (延迟测试用204响应)
const SPEED_TEST_TCP_OPTIONS = [
  { label: 'Cloudflare (cp.cloudflare.com)', value: 'http://cp.cloudflare.com/generate_204' },
  { label: 'Google (clients3.google.com)', value: 'http://clients3.google.com/generate_204' },
  { label: 'Google (android.clients.google.com)', value: 'http://android.clients.google.com/generate_204' },
  { label: 'Gstatic (www.gstatic.com)', value: 'http://www.gstatic.com/generate_204' }
];

// 测速URL选项 - Mihomo模式 (真速度测试用下载)
const SPEED_TEST_MIHOMO_OPTIONS = [
  { label: '10MB (Cloudflare)', value: 'https://speed.cloudflare.com/__down?bytes=10000000' },
  { label: '50MB (Cloudflare)', value: 'https://speed.cloudflare.com/__down?bytes=50000000' },
  { label: '100MB (Cloudflare)', value: 'https://speed.cloudflare.com/__down?bytes=100000000' }
];

// 格式化日期时间
const formatDateTime = (dateTimeString) => {
  if (!dateTimeString) return '-';
  try {
    const date = new Date(dateTimeString);
    if (isNaN(date.getTime())) return '-';
    // 检测 Go 零时间 (0001-01-01) 或无效日期
    if (date.getFullYear() <= 1) return '-';
    return date.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    });
  } catch (error) {
    console.error(error);
    return '-';
  }
};

// ISO国家代码转换为国旗emoji
const isoToFlag = (isoCode) => {
  if (!isoCode || isoCode.length !== 2) return '';
  const code = isoCode.toUpperCase() === 'TW' ? 'CN' : isoCode.toUpperCase();
  const codePoints = code.split('').map((char) => 0x1f1e6 + char.charCodeAt(0) - 65);
  return String.fromCodePoint(...codePoints);
};

// 格式化国家显示 (国旗emoji + 代码)
const formatCountry = (linkCountry) => {
  if (!linkCountry) return null;
  const flag = isoToFlag(linkCountry);
  return flag ? `${flag}${linkCountry}` : linkCountry;
};

// Cron 表达式验证
const validateCronExpression = (cron) => {
  if (!cron) return false;
  const parts = cron.trim().split(/\s+/);
  if (parts.length !== 5) return false;
  const ranges = [59, 23, 31, 12, 6];
  for (let i = 0; i < parts.length; i++) {
    const part = parts[i];
    const maxVal = ranges[i];
    if (part === '*' || part === '?') continue;
    if (part.includes('-')) {
      const [start, end] = part.split('-').map(Number);
      if (isNaN(start) || isNaN(end) || start < 0 || end > maxVal || start > end) return false;
      continue;
    }
    if (part.includes('/')) {
      const [base, step] = part.split('/');
      if (isNaN(Number(step)) || Number(step) <= 0) return false;
      if (base !== '*' && !base.includes('-')) {
        const num = Number(base);
        if (isNaN(num) || num < 0 || num > maxVal) return false;
      }
      continue;
    }
    if (part.includes(',')) {
      const values = part.split(',').map(Number);
      for (const val of values) {
        if (isNaN(val) || val < 0 || val > maxVal) return false;
      }
      continue;
    }
    const num = Number(part);
    if (isNaN(num) || num < 0 || num > maxVal) return false;
  }
  return true;
};

// ==============================|| 节点管理 ||============================== //

export default function NodeList() {
  const theme = useTheme();
  const matchDownMd = useMediaQuery(theme.breakpoints.down('md'));

  const [nodes, setNodes] = useState([]);
  const [loading, setLoading] = useState(false);
  const [selectedNodes, setSelectedNodes] = useState([]);

  // 确认对话框
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [confirmInfo, setConfirmInfo] = useState({
    title: '',
    content: '',
    action: null
  });

  const openConfirm = (title, content, action) => {
    setConfirmInfo({ title, content, action });
    setConfirmOpen(true);
  };

  const handleConfirmClose = () => {
    setConfirmOpen(false);
  };

  const handleConfirmAction = async () => {
    if (confirmInfo.action) {
      await confirmInfo.action();
    }
    setConfirmOpen(false);
  };

  // 节点表单
  const [nodeDialogOpen, setNodeDialogOpen] = useState(false);
  const [isEditNode, setIsEditNode] = useState(false);
  const [currentNode, setCurrentNode] = useState(null);
  const [nodeForm, setNodeForm] = useState({
    name: '',
    link: '',
    dialerProxyName: '',
    group: '',
    mergeMode: '1' // 1=合并, 2=分开
  });

  // 过滤器
  const [searchQuery, setSearchQuery] = useState('');
  const [groupFilter, setGroupFilter] = useState('');
  const [sourceFilter, setSourceFilter] = useState('');
  const [maxDelay, setMaxDelay] = useState('');
  const [minSpeed, setMinSpeed] = useState('');

  // 排序
  const [sortBy, setSortBy] = useState(''); // 'delay' | 'speed' | ''
  const [sortOrder, setSortOrder] = useState('asc'); // 'asc' | 'desc'

  // 分页
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(10);

  // 订阅调度器
  const [schedulers, setSchedulers] = useState([]);
  const [schedulerDialogOpen, setSchedulerDialogOpen] = useState(false);
  const [schedulerFormOpen, setSchedulerFormOpen] = useState(false);
  const [isEditScheduler, setIsEditScheduler] = useState(false);
  const [schedulerForm, setSchedulerForm] = useState({
    name: '',
    url: '',
    cron_expr: '',
    enabled: true,
    group: '',
    download_with_proxy: false,
    proxy_link: ''
  });

  // 订阅删除对话框状态
  const [deleteSchedulerDialogOpen, setDeleteSchedulerDialogOpen] = useState(false);
  const [deleteSchedulerTarget, setDeleteSchedulerTarget] = useState(null);
  const [deleteSchedulerWithNodes, setDeleteSchedulerWithNodes] = useState(true);

  // 测速配置
  const [speedTestDialogOpen, setSpeedTestDialogOpen] = useState(false);
  const [speedTestForm, setSpeedTestForm] = useState({
    cron: '',
    enabled: false,
    mode: 'tcp',
    url: '',
    timeout: 5,
    groups: [],
    detect_country: false
  });

  // 国家筛选
  const [countryFilter, setCountryFilter] = useState([]);
  const [countryOptions, setCountryOptions] = useState([]);
  // 从后端获取的分组和来源选项
  const [groupOptions, setGroupOptions] = useState([]);
  const [sourceOptions, setSourceOptions] = useState([]);

  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });

  // 后端已完成过滤和排序，直接使用 nodes 数组
  const filteredNodes = nodes;

  // 防抖定时器引用
  const debounceTimerRef = useRef(null);

  // 获取节点列表（支持过滤参数）
  const fetchNodes = useCallback(async (filterParams = {}) => {
    setLoading(true);
    try {
      // 构建过滤参数
      const params = {};
      if (filterParams.search) params.search = filterParams.search;
      if (filterParams.group) params.group = filterParams.group;
      if (filterParams.source) params.source = filterParams.source;
      if (filterParams.maxDelay) params.maxDelay = filterParams.maxDelay;
      if (filterParams.minSpeed) params.minSpeed = filterParams.minSpeed;
      if (filterParams.countries && filterParams.countries.length > 0) {
        params['countries[]'] = filterParams.countries;
      }
      if (filterParams.sortBy) params.sortBy = filterParams.sortBy;
      if (filterParams.sortOrder) params.sortOrder = filterParams.sortOrder;

      const response = await getNodes(params);
      setNodes(response.data || []);
    } catch (error) {
      console.error(error);
      showMessage('获取节点列表失败', 'error');
    } finally {
      setLoading(false);
    }
  }, []);

  // 获取订阅调度器列表
  const fetchSchedulers = useCallback(async () => {
    try {
      const response = await getSubSchedulers();
      setSchedulers(response.data || []);
    } catch (error) {
      console.error('获取订阅调度器失败:', error);
    }
  }, []);

  // 初始化加载
  useEffect(() => {
    fetchNodes();
    // 请求国家代码列表
    getNodeCountries()
      .then((res) => {
        setCountryOptions(res.data || []);
      })
      .catch(console.error);
    // 请求分组列表
    getNodeGroups()
      .then((res) => {
        setGroupOptions((res.data || []).sort());
      })
      .catch(console.error);
    // 请求来源列表
    getNodeSources()
      .then((res) => {
        setSourceOptions((res.data || []).sort());
      })
      .catch(console.error);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // 监听过滤条件变化，带防抖发送请求到后端
  useEffect(() => {
    // 清除之前的定时器
    if (debounceTimerRef.current) {
      clearTimeout(debounceTimerRef.current);
    }

    // 设置防抖延迟
    debounceTimerRef.current = setTimeout(() => {
      const filterParams = {
        search: searchQuery,
        group: groupFilter,
        source: sourceFilter,
        maxDelay: maxDelay,
        minSpeed: minSpeed,
        countries: countryFilter,
        sortBy: sortBy,
        sortOrder: sortOrder
      };
      fetchNodes(filterParams);
    }, 300); // 300ms 防抖延迟

    // 清理函数
    return () => {
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current);
      }
    };
  }, [searchQuery, groupFilter, sourceFilter, maxDelay, minSpeed, countryFilter, sortBy, sortOrder, fetchNodes]);

  const showMessage = (message, severity = 'success') => {
    setSnackbar({ open: true, message, severity });
  };

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
    showMessage('已复制到剪贴板');
  };

  // 重置过滤
  const resetFilters = () => {
    setSearchQuery('');
    setGroupFilter('');
    setSourceFilter('');
    setMaxDelay('');
    setMinSpeed('');
    setCountryFilter([]);
    setSortBy('');
    setSortOrder('asc');
  };

  // 获取当前过滤参数
  const getCurrentFilters = () => ({
    search: searchQuery,
    group: groupFilter,
    source: sourceFilter,
    maxDelay: maxDelay,
    minSpeed: minSpeed,
    countries: countryFilter,
    sortBy: sortBy,
    sortOrder: sortOrder
  });

  // 手动刷新（保留搜索条件）
  const handleRefresh = () => {
    fetchNodes(getCurrentFilters());
  };

  // === 节点操作 ===
  const handleAddNode = () => {
    setIsEditNode(false);
    setCurrentNode(null);
    setNodeForm({ name: '', link: '', dialerProxyName: '', group: '', mergeMode: '1' });
    setNodeDialogOpen(true);
  };

  const handleEditNode = (node) => {
    setIsEditNode(true);
    setCurrentNode(node);
    setNodeForm({
      name: node.Name,
      link: node.Link?.split(',').join('\n') || '',
      dialerProxyName: node.DialerProxyName || '',
      group: node.Group || '',
      mergeMode: '1'
    });
    setNodeDialogOpen(true);
  };

  const handleDeleteNode = async (node) => {
    openConfirm('删除节点', `确定要删除节点 "${node.Name}" 吗？`, async () => {
      try {
        await deleteNode({ id: node.ID });
        showMessage('删除成功');
        fetchNodes();
      } catch (error) {
        console.error(error);
        showMessage('删除失败', 'error');
      }
    });
  };

  const handleBatchDelete = async () => {
    if (selectedNodes.length === 0) {
      showMessage('请选择要删除的节点', 'warning');
      return;
    }
    openConfirm('批量删除', `确定要删除选中的 ${selectedNodes.length} 个节点吗？`, async () => {
      try {
        const ids = selectedNodes.map((node) => node.ID);
        await deleteNodesBatch(ids);
        showMessage(`成功删除 ${selectedNodes.length} 个节点`);
        setSelectedNodes([]);
        fetchNodes();
      } catch (error) {
        console.error(error);
        showMessage('批量删除失败', 'error');
      }
    });
  };

  const handleSubmitNode = async () => {
    const nodeLinks = nodeForm.link
      .split(/[\r\n,]/)
      .map((item) => item.trim())
      .filter((item) => item !== '');

    if (nodeLinks.length === 0) {
      showMessage('请输入节点链接', 'warning');
      return;
    }

    try {
      if (isEditNode) {
        const processedLink = nodeLinks.join(',');
        await updateNode({
          oldname: currentNode.Name,
          oldlink: currentNode.Link,
          link: processedLink,
          name: nodeForm.name.trim(),
          dialerProxyName: nodeForm.dialerProxyName.trim(),
          group: nodeForm.group.trim()
        });
        showMessage('更新成功');
      } else {
        if (nodeForm.mergeMode === '1') {
          // 合并模式
          if (!nodeForm.name.trim()) {
            showMessage('备注不能为空', 'warning');
            return;
          }
          const processedLink = nodeLinks.join(',');
          await addNodes({
            link: processedLink,
            name: nodeForm.name.trim(),
            dialerProxyName: nodeForm.dialerProxyName.trim(),
            group: nodeForm.group.trim()
          });
        } else {
          // 分开模式
          for (const link of nodeLinks) {
            await addNodes({
              link,
              name: '',
              dialerProxyName: nodeForm.dialerProxyName.trim(),
              group: nodeForm.group.trim()
            });
          }
        }
        showMessage('添加成功');
      }
      setNodeDialogOpen(false);
      fetchNodes();
    } catch (error) {
      console.error(error);
      showMessage(isEditNode ? '更新失败' : '添加失败', 'error');
    }
  };

  // === 订阅调度器操作 ===
  const handleOpenSchedulerDialog = () => {
    fetchSchedulers();
    setSchedulerDialogOpen(true);
  };

  const handleAddScheduler = () => {
    setIsEditScheduler(false);
    setSchedulerForm({
      name: '',
      url: '',
      cron_expr: '',
      enabled: true,
      group: '',
      download_with_proxy: false,
      proxy_link: ''
    });
    setSchedulerFormOpen(true);
  };

  const handleEditScheduler = (scheduler) => {
    setIsEditScheduler(true);
    setSchedulerForm({
      id: scheduler.ID,
      name: scheduler.Name,
      url: scheduler.URL,
      cron_expr: scheduler.CronExpr,
      enabled: scheduler.Enabled,
      group: scheduler.Group || '',
      download_with_proxy: scheduler.DownloadWithProxy || false,
      proxy_link: scheduler.ProxyLink || ''
    });
    setSchedulerFormOpen(true);
  };

  const handleDeleteScheduler = (scheduler) => {
    setDeleteSchedulerTarget(scheduler);
    setDeleteSchedulerWithNodes(true);
    setDeleteSchedulerDialogOpen(true);
  };

  const handleConfirmDeleteScheduler = async () => {
    if (!deleteSchedulerTarget) return;
    try {
      await deleteSubScheduler(deleteSchedulerTarget.ID, deleteSchedulerWithNodes);
      showMessage(deleteSchedulerWithNodes ? '已删除订阅及关联节点' : '已删除订阅（保留节点）');
      fetchSchedulers();
      fetchNodes();
    } catch (error) {
      console.error(error);
      showMessage('删除失败', 'error');
    }
    setDeleteSchedulerDialogOpen(false);
    setDeleteSchedulerTarget(null);
  };

  const handlePullScheduler = async (scheduler) => {
    openConfirm('立即更新', `确定要立即更新订阅 "${scheduler.Name}" 吗？`, async () => {
      try {
        await pullSubScheduler({
          id: scheduler.ID,
          name: scheduler.Name,
          url: scheduler.URL,
          cron_expr: scheduler.CronExpr,
          enabled: scheduler.Enabled,
          group: scheduler.Group,
          download_with_proxy: scheduler.DownloadWithProxy,
          proxy_link: scheduler.ProxyLink
        });
        showMessage('提交更新任务成功，请稍后刷新查看结果');
        fetchSchedulers();
        fetchNodes();
      } catch (error) {
        console.error(error);
        showMessage('提交更新任务失败', 'error');
      }
    });
  };

  const handleSubmitScheduler = async () => {
    if (!schedulerForm.name.trim()) {
      showMessage('请输入名称', 'warning');
      return;
    }
    if (!schedulerForm.url.trim()) {
      showMessage('请输入URL', 'warning');
      return;
    }
    // Simple URL validation regex
    const urlPattern = /^(https?|ftp):\/\/[^\s/$.?#].[^\s]*$/i;
    if (!urlPattern.test(schedulerForm.url.trim())) {
      showMessage('请输入有效的URL', 'warning');
      return;
    }
    if (!schedulerForm.cron_expr.trim()) {
      showMessage('请输入Cron表达式', 'warning');
      return;
    }
    if (!validateCronExpression(schedulerForm.cron_expr.trim())) {
      showMessage('Cron表达式格式不正确，格式为：分 时 日 月 周', 'error');
      return;
    }

    try {
      if (isEditScheduler) {
        await updateSubScheduler(schedulerForm);
        showMessage('更新成功');
      } else {
        await addSubScheduler(schedulerForm);
        showMessage('添加成功');
      }
      setSchedulerFormOpen(false);
      fetchSchedulers();
    } catch (error) {
      console.error(error);
      showMessage(isEditScheduler ? '更新失败' : '添加失败', 'error');
    }
  };

  // === 测速配置 ===
  const handleOpenSpeedTest = async () => {
    try {
      const response = await getSpeedTestConfig();
      setSpeedTestForm(response.data || { cron: '', enabled: false, mode: 'tcp', url: '', timeout: 5, groups: [] });
      setSpeedTestDialogOpen(true);
    } catch (error) {
      console.error(error);
      showMessage('获取测速配置失败', 'error');
    }
  };

  const handleSpeedModeChange = (mode) => {
    const newUrl = mode === 'mihomo' ? SPEED_TEST_MIHOMO_OPTIONS[0].value : SPEED_TEST_TCP_OPTIONS[0].value;
    setSpeedTestForm({ ...speedTestForm, mode, url: newUrl });
  };

  const handleSubmitSpeedTest = async () => {
    if (speedTestForm.enabled && !speedTestForm.cron) {
      showMessage('启用时Cron表达式不能为空', 'warning');
      return;
    }
    if (speedTestForm.enabled && !validateCronExpression(speedTestForm.cron)) {
      showMessage('Cron表达式格式不正确', 'error');
      return;
    }
    try {
      await updateSpeedTestConfig(speedTestForm);
      showMessage('保存成功');
      setSpeedTestDialogOpen(false);
    } catch (error) {
      console.error(error);
      showMessage('保存测速配置失败', 'error');
    }
  };

  const handleRunSpeedTest = async () => {
    try {
      await runSpeedTest();
      showMessage('测速任务已在后台启动，请稍后刷新查看结果');
    } catch (error) {
      console.error(error);
      showMessage('启动测速任务失败', 'error');
    }
  };

  const handleBatchSpeedTest = async () => {
    if (selectedNodes.length === 0) {
      showMessage('请选择要测速的节点', 'warning');
      return;
    }
    try {
      const ids = selectedNodes.map((node) => node.ID);
      await runSpeedTest(ids);
      showMessage(`已启动 ${ids.length} 个节点的测速任务`);
    } catch (error) {
      console.error(error);
      showMessage('启动批量测速任务失败', 'error');
    }
  };

  const handleSingleSpeedTest = async (node) => {
    try {
      await runSpeedTest([node.ID]);
      showMessage(`节点 ${node.Name} 测速任务已启动`);
    } catch (error) {
      console.error(error);
      showMessage('启动测速任务失败', 'error');
    }
  };

  // 延迟颜色
  const getDelayColor = (delay) => {
    if (delay <= 0) return 'default';
    if (delay < 100) return 'success';
    if (delay < 500) return 'warning';
    return 'error';
  };

  // 选择所有
  const handleSelectAll = (event) => {
    if (event.target.checked) {
      setSelectedNodes(filteredNodes);
    } else {
      setSelectedNodes([]);
    }
  };

  const handleSelectNode = (node) => {
    const isSelected = selectedNodes.some((n) => n.ID === node.ID);
    if (isSelected) {
      setSelectedNodes(selectedNodes.filter((n) => n.ID !== node.ID));
    } else {
      setSelectedNodes([...selectedNodes, node]);
    }
  };

  const isSelected = (node) => selectedNodes.some((n) => n.ID === node.ID);

  // 排序处理
  const handleSort = (field) => {
    if (sortBy === field) {
      // 如果点击同一列，切换排序顺序或清除排序
      if (sortOrder === 'asc') {
        setSortOrder('desc');
      } else {
        setSortBy('');
        setSortOrder('asc');
      }
    } else {
      // 如果点击不同列，设置新的排序
      setSortBy(field);
      setSortOrder('asc');
    }
  };

  return (
    <MainCard
      title="节点管理"
      secondary={
        matchDownMd ? (
          <Tooltip title="添加节点/更多操作">
            <Button variant="contained" startIcon={<AddIcon />} onClick={handleAddNode}>
              添加
            </Button>
          </Tooltip>
        ) : (
          <Stack direction="row" spacing={1}>
            <Button variant="contained" startIcon={<AddIcon />} onClick={handleAddNode}>
              添加节点
            </Button>
            <Button variant="outlined" color="primary" startIcon={<DownloadIcon />} onClick={handleOpenSchedulerDialog}>
              导入订阅
            </Button>
            <Button variant="outlined" color="info" startIcon={<SettingsIcon />} onClick={handleOpenSpeedTest}>
              测速设置
            </Button>
            <Button variant="outlined" startIcon={<SpeedIcon />} onClick={handleBatchSpeedTest}>
              批量测速
            </Button>
            <IconButton onClick={handleRefresh} disabled={loading}>
              <RefreshIcon
                sx={
                  loading
                    ? {
                      animation: "spin 1s linear infinite",
                      "@keyframes spin": { from: { transform: "rotate(0deg)" }, to: { transform: "rotate(360deg)" } }
                    }
                    : {}
                }
              />
            </IconButton>
          </Stack>
        )
      }
    >
      {/* 移动端顶部额外按钮栏 */}
      {matchDownMd && (
        <Stack direction="row" spacing={1} sx={{ mb: 2, overflowX: 'auto', pb: 1 }} className="hide-scrollbar">
          <Button
            size="small"
            variant="outlined"
            color="primary"
            startIcon={<DownloadIcon />}
            onClick={handleOpenSchedulerDialog}
            sx={{ whiteSpace: 'nowrap' }}
          >
            导入
          </Button>
          <Button
            size="small"
            variant="outlined"
            color="info"
            startIcon={<SettingsIcon />}
            onClick={handleOpenSpeedTest}
            sx={{ whiteSpace: 'nowrap' }}
          >
            测速设置
          </Button>
          <Button size="small" variant="outlined" startIcon={<SpeedIcon />} onClick={handleBatchSpeedTest} sx={{ whiteSpace: 'nowrap' }}>
            批量测速
          </Button>
          <IconButton size="small" onClick={handleRefresh} disabled={loading}>
            <RefreshIcon
              sx={
                loading
                  ? {
                    animation: "spin 1s linear infinite",
                    "@keyframes spin": { from: { transform: "rotate(0deg)" }, to: { transform: "rotate(360deg)" } }
                  }
                  : {}
              }
            />
          </IconButton>
        </Stack>
      )}
      {/* 过滤器 */}
      <Stack direction="row" spacing={2} sx={{ mb: 2 }} flexWrap="wrap" useFlexGap>
        <FormControl size="small" sx={{ minWidth: 120 }}>
          <InputLabel>分组</InputLabel>
          <Select value={groupFilter} label="分组" onChange={(e) => setGroupFilter(e.target.value)}
                  variant={"outlined"}>
            <MenuItem value="">全部</MenuItem>
            <MenuItem value="未分组">未分组</MenuItem>
            {groupOptions.map((group) => (
              <MenuItem key={group} value={group}>
                {group}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
        <TextField
          size="small"
          placeholder="搜索节点备注或链接"
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          sx={{ minWidth: 200 }}
        />
        <FormControl size="small" sx={{ minWidth: 120 }}>
          <InputLabel>来源</InputLabel>
          <Select value={sourceFilter} label="来源" onChange={(e) => setSourceFilter(e.target.value)}
                  variant={"outlined"}>
            <MenuItem value="">全部</MenuItem>
            <MenuItem value="手动添加">手动添加</MenuItem>
            {sourceOptions.map((source) => (
              <MenuItem key={source} value={source}>
                {source === 'manual' ? '手动添加' : source}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
        <TextField
          size="small"
          placeholder="最大延迟"
          type="number"
          value={maxDelay}
          onChange={(e) => setMaxDelay(e.target.value)}
          sx={{ width: 150 }}
          InputProps={{ endAdornment: <InputAdornment position="end">ms</InputAdornment> }}
        />
        <TextField
          size="small"
          placeholder="最低速度"
          type="number"
          value={minSpeed}
          onChange={(e) => setMinSpeed(e.target.value)}
          sx={{ width: 150 }}
          InputProps={{ endAdornment: <InputAdornment position="end">MB/s</InputAdornment> }}
        />
        {countryOptions.length > 0 && (
          <Autocomplete
            multiple
            size="small"
            options={countryOptions}
            value={countryFilter}
            onChange={(e, newValue) => setCountryFilter(newValue)}
            sx={{ minWidth: 150 }}
            getOptionLabel={(option) => `${isoToFlag(option)} ${option}`}
            renderOption={(props, option) => {
              const { key, ...otherProps } = props;
              return (
                <li key={key} {...otherProps}>
                  {isoToFlag(option)} {option}
                </li>
              );
            }}
            renderTags={(value, getTagProps) =>
              value.map((option, index) => {
                const { key, ...tagProps } = getTagProps({ index });
                return <Chip key={key} label={`${isoToFlag(option)} ${option}`} size="small" {...tagProps} />;
              })
            }
            renderInput={(params) => <TextField {...params} label="国家代码" placeholder="选择国家" />}
          />
        )}
        <Button onClick={resetFilters}>重置</Button>
      </Stack>

      {/* 批量操作 */}
      {selectedNodes.length > 0 && (
        <Stack direction="row" spacing={1} sx={{ mb: 2 }}>
          <Typography variant="body2" sx={{ alignSelf: 'center' }}>
            已选择 {selectedNodes.length} 个节点
          </Typography>
          <Button size="small" color="error" startIcon={<DeleteIcon />} onClick={handleBatchDelete}>
            批量删除
          </Button>
        </Stack>
      )}

      {/* 节点列表 */}
      {matchDownMd ? (
        <Stack spacing={2}>
          {filteredNodes.length === 0 && (
            <Typography variant="body2" color="textSecondary" align="center" sx={{ py: 3 }}>
              暂无节点
            </Typography>
          )}
          {filteredNodes.slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage).map((node) => (
            <MainCard key={node.ID} content={false} border shadow={theme.shadows[1]}>
              <Box p={2}>
                {/* Header: Checkbox, Name, Delay */}
                <Stack direction="row" justifyContent="space-between" alignItems="flex-start" mb={1.5}>
                  <Stack direction="row" spacing={1} alignItems="center" sx={{ flex: 1, minWidth: 0 }}>
                    <Checkbox checked={isSelected(node)} onChange={() => handleSelectNode(node)} sx={{ p: 0.5, flexShrink: 0 }} />
                    <Tooltip title={node.Name} placement="top">
                      <Typography
                        variant="subtitle1"
                        fontWeight="bold"
                        sx={{
                          overflow: 'hidden',
                          textOverflow: 'ellipsis',
                          whiteSpace: 'nowrap',
                          maxWidth: '200px'
                        }}
                      >
                        {node.Name}
                      </Typography>
                    </Tooltip>
                  </Stack>
                  <Box sx={{ flexShrink: 0, ml: 1 }}>
                    {node.DelayTime > 0 ? (
                      <Chip label={`${node.DelayTime}ms`} color={getDelayColor(node.DelayTime)} size="small" />
                    ) : node.DelayTime === -1 ? (
                      <Chip label="超时" color="error" size="small" />
                    ) : (
                      <Chip label="未测速" variant="outlined" size="small" />
                    )}
                  </Box>
                </Stack>

                {/* Info Section: Chips for Group, Source, Speed */}
                <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap sx={{ mb: 1.5 }}>
                  <Tooltip title={`分组: ${node.Group || '未分组'}`}>
                    <Chip
                      icon={<span style={{ fontSize: '12px', marginLeft: '8px' }}>📁</span>}
                      label={node.Group || '未分组'}
                      color="warning"
                      variant="outlined"
                      size="small"
                      sx={{ maxWidth: '120px', '& .MuiChip-label': { overflow: 'hidden', textOverflow: 'ellipsis' } }}
                    />
                  </Tooltip>
                  <Chip
                    icon={<span style={{ fontSize: '12px', marginLeft: '8px' }}>📡</span>}
                    label={node.Source === 'manual' ? '手动添加' : node.Source || '未知'}
                    color={node.Source === 'manual' ? 'success' : 'info'}
                    variant="outlined"
                    size="small"
                    sx={{ maxWidth: '100px', '& .MuiChip-label': { overflow: 'hidden', textOverflow: 'ellipsis' } }}
                  />
                  <Chip
                    icon={<span style={{ fontSize: '12px', marginLeft: '8px' }}>⚡</span>}
                    label={node.Speed > 0 ? `${node.Speed.toFixed(2)}MB/s` : '未测速'}
                    color={node.Speed > 0 ? 'primary' : 'default'}
                    variant={node.Speed > 0 ? 'filled' : 'outlined'}
                    size="small"
                  />
                  {node.DialerProxyName && (
                    <Tooltip title={`前置代理: ${node.DialerProxyName}`}>
                      <Chip
                        icon={<span style={{ fontSize: '12px', marginLeft: '8px' }}>🔗</span>}
                        label={node.DialerProxyName}
                        variant="outlined"
                        size="small"
                        sx={{ maxWidth: '100px', '& .MuiChip-label': { overflow: 'hidden', textOverflow: 'ellipsis' } }}
                      />
                    </Tooltip>
                  )}
                  {node.LinkCountry && (
                    <Tooltip title={`国家: ${node.LinkCountry}`}>
                      <Chip label={formatCountry(node.LinkCountry)} color="secondary" variant="outlined" size="small" />
                    </Tooltip>
                  )}
                </Stack>

                {/* Time Info Section */}
                <Box sx={{ bgcolor: 'action.hover', borderRadius: 1, p: 1, mb: 1.5 }}>
                  <Stack spacing={0.5}>
                    <Box>
                      <Typography variant="caption" color="textSecondary" display="block">
                        创建时间
                      </Typography>
                      <Typography variant="caption" fontWeight="medium">
                        {node.CreatedAt ? formatDateTime(node.CreatedAt) : '-'}
                      </Typography>
                    </Box>
                    <Box>
                      <Typography variant="caption" color="textSecondary" display="block">
                        更新时间
                      </Typography>
                      <Typography variant="caption" fontWeight="medium">
                        {node.UpdatedAt ? formatDateTime(node.UpdatedAt) : '-'}
                      </Typography>
                    </Box>
                    <Box>
                      <Typography variant="caption" color="textSecondary" display="block">
                        最后测速
                      </Typography>
                      <Typography variant="caption" fontWeight="medium" color="primary">
                        {node.LastCheck ? formatDateTime(node.LastCheck) : '-'}
                      </Typography>
                    </Box>
                  </Stack>
                </Box>

                {/* Action Buttons */}
                <Stack direction="row" spacing={0.5} justifyContent="flex-end">
                  <Tooltip title="测速">
                    <IconButton size="small" onClick={() => handleSingleSpeedTest(node)}>
                      <SpeedIcon fontSize="small" />
                    </IconButton>
                  </Tooltip>
                  <Tooltip title="复制链接">
                    <IconButton size="small" onClick={() => copyToClipboard(node.Link)}>
                      <ContentCopyIcon fontSize="small" />
                    </IconButton>
                  </Tooltip>
                  <Tooltip title="编辑">
                    <IconButton size="small" onClick={() => handleEditNode(node)}>
                      <EditIcon fontSize="small" />
                    </IconButton>
                  </Tooltip>
                  <Tooltip title="删除">
                    <IconButton size="small" color="error" onClick={() => handleDeleteNode(node)}>
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  </Tooltip>
                </Stack>
              </Box>
            </MainCard>
          ))}
        </Stack>
      ) : (
        <TableContainer component={Paper}>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell padding="checkbox">
                  <Checkbox
                    indeterminate={selectedNodes.length > 0 && selectedNodes.length < filteredNodes.length}
                    checked={filteredNodes.length > 0 && selectedNodes.length === filteredNodes.length}
                    onChange={handleSelectAll}
                  />
                </TableCell>
                <TableCell>备注</TableCell>
                <TableCell>分组</TableCell>
                <TableCell>来源</TableCell>
                <TableCell>节点名称</TableCell>
                <TableCell>前置代理</TableCell>
                <TableCell sortDirection={sortBy === 'delay' ? sortOrder : false}>
                  <TableSortLabel
                    active={sortBy === 'delay'}
                    direction={sortBy === 'delay' ? sortOrder : 'asc'}
                    onClick={() => handleSort('delay')}
                  >
                    延迟
                  </TableSortLabel>
                </TableCell>
                <TableCell sortDirection={sortBy === 'speed' ? sortOrder : false}>
                  <TableSortLabel
                    active={sortBy === 'speed'}
                    direction={sortBy === 'speed' ? sortOrder : 'asc'}
                    onClick={() => handleSort('speed')}
                  >
                    速度
                  </TableSortLabel>
                </TableCell>
                <TableCell sx={{ minWidth: 100, whiteSpace: 'nowrap' }}>国家</TableCell>
                <TableCell sx={{ minWidth: 160, whiteSpace: 'nowrap' }}>创建时间</TableCell>
                <TableCell sx={{ minWidth: 160, whiteSpace: 'nowrap' }}>更新时间</TableCell>
                <TableCell align="right">操作</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {filteredNodes.slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage).map((node) => (
                <TableRow key={node.ID} hover selected={isSelected(node)}>
                  <TableCell padding="checkbox">
                    <Checkbox checked={isSelected(node)} onChange={() => handleSelectNode(node)} />
                  </TableCell>
                  <TableCell>
                    <Tooltip title={node.Name}>
                      <Chip
                        label={node.Name}
                        color="success"
                        variant="outlined"
                        size="small"
                        sx={{ maxWidth: '150px', '& .MuiChip-label': { overflow: 'hidden', textOverflow: 'ellipsis' } }}
                      />
                    </Tooltip>
                  </TableCell>
                  <TableCell>
                    {node.Group ? (
                      <Tooltip title={node.Group}>
                        <Chip
                          label={node.Group}
                          color="warning"
                          variant="outlined"
                          size="small"
                          sx={{ maxWidth: '120px', '& .MuiChip-label': { overflow: 'hidden', textOverflow: 'ellipsis' } }}
                        />
                      </Tooltip>
                    ) : (
                      <Typography variant="caption" color="textSecondary">
                        未分组
                      </Typography>
                    )}
                  </TableCell>
                  <TableCell>
                    <Chip
                      label={node.Source === 'manual' ? '手动添加' : node.Source}
                      color={node.Source === 'manual' ? 'success' : 'warning'}
                      variant="outlined"
                      size="small"
                    />
                  </TableCell>
                  <TableCell>
                    <Tooltip title={node.LinkName || ''}>
                      <Typography sx={{ maxWidth: 200, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                        {node.LinkName || '-'}
                      </Typography>
                    </Tooltip>
                  </TableCell>
                  <TableCell>
                    <Tooltip title={node.DialerProxyName || ''}>
                      <Typography sx={{ minWidth: 100, maxWidth: 150, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                        {node.DialerProxyName || '-'}
                      </Typography>
                    </Tooltip>
                  </TableCell>
                  <TableCell>
                    <Box>
                      {node.DelayTime > 0 ? (
                        <Chip label={`${node.DelayTime}ms`} color={getDelayColor(node.DelayTime)} size="small" />
                      ) : node.DelayTime === -1 ? (
                        <Chip label="超时" color="error" size="small" />
                      ) : (
                        <Chip label="未测速" variant="outlined" size="small" />
                      )}
                      {node.LastCheck && (
                        <Typography variant="caption" color="textSecondary" sx={{ display: 'block', fontSize: '10px', mt: 0.5 }}>
                          {formatDateTime(node.LastCheck)}
                        </Typography>
                      )}
                    </Box>
                  </TableCell>
                  <TableCell>{node.Speed > 0 ? `${node.Speed.toFixed(2)}MB/s` : '-'}</TableCell>
                  <TableCell>
                    {node.LinkCountry ? (
                      <Chip label={formatCountry(node.LinkCountry)} color="secondary" variant="outlined" size="small" />
                    ) : (
                      '-'
                    )}
                  </TableCell>
                  <TableCell sx={{ minWidth: 160, whiteSpace: 'nowrap' }}>
                    <Typography variant="caption">{formatDateTime(node.CreatedAt)}</Typography>
                  </TableCell>
                  <TableCell sx={{ minWidth: 160, whiteSpace: 'nowrap' }}>
                    <Typography variant="caption">{formatDateTime(node.UpdatedAt)}</Typography>
                  </TableCell>
                  <TableCell align="right" sx={{ minWidth: 160 }}>
                    <Tooltip title="测速">
                      <IconButton size="small" onClick={() => handleSingleSpeedTest(node)}>
                        <SpeedIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="复制链接">
                      <IconButton size="small" onClick={() => copyToClipboard(node.Link)}>
                        <ContentCopyIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="编辑">
                      <IconButton size="small" onClick={() => handleEditNode(node)}>
                        <EditIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="删除">
                      <IconButton size="small" color="error" onClick={() => handleDeleteNode(node)}>
                        <DeleteIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      <TablePagination
        component="div"
        count={filteredNodes.length}
        page={page}
        onPageChange={(e, newPage) => setPage(newPage)}
        rowsPerPage={rowsPerPage}
        onRowsPerPageChange={(e) => {
          setRowsPerPage(parseInt(e.target.value, 10));
          setPage(0);
        }}
        labelRowsPerPage="每页行数:"
      />

      {/* 添加/编辑节点对话框 */}
      <Dialog open={nodeDialogOpen} onClose={() => setNodeDialogOpen(false)} maxWidth="md" fullWidth>
        <DialogTitle>{isEditNode ? '编辑节点' : '添加节点'}</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <TextField
              fullWidth
              multiline
              rows={4}
              label="节点链接"
              value={nodeForm.link}
              onChange={(e) => setNodeForm({ ...nodeForm, link: e.target.value })}
              placeholder="请输入节点，多行使用回车或逗号分开，支持base64格式的url订阅"
            />
            {!isEditNode && (
              <RadioGroup row value={nodeForm.mergeMode} onChange={(e) => setNodeForm({ ...nodeForm, mergeMode: e.target.value })}>
                <FormControlLabel value="1" control={<Radio />} label="合并" />
                <FormControlLabel value="2" control={<Radio />} label="分开" />
              </RadioGroup>
            )}
            {(isEditNode || nodeForm.mergeMode === '1') && (
              <TextField
                fullWidth
                label="备注"
                value={nodeForm.name}
                onChange={(e) => setNodeForm({ ...nodeForm, name: e.target.value })}
              />
            )}
            <TextField
              fullWidth
              label="前置代理节点名称或策略组名称"
              value={nodeForm.dialerProxyName}
              onChange={(e) => setNodeForm({ ...nodeForm, dialerProxyName: e.target.value })}
              helperText="仅Clash-Meta内核可用"
            />
            <Autocomplete
              freeSolo
              options={groupOptions}
              value={nodeForm.group}
              onChange={(e, newValue) => setNodeForm({ ...nodeForm, group: newValue || '' })}
              onInputChange={(e, newValue) => setNodeForm({ ...nodeForm, group: newValue || '' })}
              renderInput={(params) => <TextField {...params} label="分组" placeholder="请选择或输入分组名称" />}
            />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setNodeDialogOpen(false)}>关闭</Button>
          <Button variant="contained" onClick={handleSubmitNode}>
            确定
          </Button>
        </DialogActions>
      </Dialog>

      {/* 订阅调度器对话框 */}
      <Dialog open={schedulerDialogOpen} onClose={() => setSchedulerDialogOpen(false)} maxWidth="lg" fullWidth>
        <DialogTitle>
          导入订阅
          <Button sx={{ ml: 2 }} variant="contained" size="small" startIcon={<AddIcon />} onClick={handleAddScheduler}>
            添加订阅
          </Button>
        </DialogTitle>
        <DialogContent>
          <TableContainer>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>名称</TableCell>
                  <TableCell>URL</TableCell>
                  <TableCell>节点数量</TableCell>
                  <TableCell>上次运行</TableCell>
                  <TableCell>下次运行</TableCell>
                  <TableCell>Cron表达式</TableCell>
                  <TableCell>分组</TableCell>
                  <TableCell>状态</TableCell>
                  <TableCell align="right">操作</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {schedulers.map((scheduler) => (
                  <TableRow key={scheduler.ID}>
                    <TableCell>{scheduler.Name}</TableCell>
                    <TableCell sx={{ maxWidth: 200, overflow: 'hidden', textOverflow: 'ellipsis' }}>{scheduler.URL}</TableCell>
                    <TableCell>
                      <Chip label={scheduler.node_count || 0} color="primary" variant="outlined" size="small" />
                    </TableCell>
                    <TableCell>{formatDateTime(scheduler.LastRunTime)}</TableCell>
                    <TableCell>{formatDateTime(scheduler.NextRunTime)}</TableCell>
                    <TableCell>{scheduler.CronExpr}</TableCell>
                    <TableCell>{scheduler.Group || '-'}</TableCell>
                    <TableCell>
                      <Chip label={scheduler.Enabled ? '启用' : '禁用'} color={scheduler.Enabled ? 'success' : 'default'} size="small" />
                    </TableCell>
                    <TableCell align="right">
                      <Tooltip title="立即拉取">
                        <IconButton size="small" onClick={() => handlePullScheduler(scheduler)}>
                          <PlayArrowIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                      <IconButton size="small" onClick={() => handleEditScheduler(scheduler)}>
                        <EditIcon fontSize="small" />
                      </IconButton>
                      <IconButton size="small" color="error" onClick={() => handleDeleteScheduler(scheduler)}>
                        <DeleteIcon fontSize="small" />
                      </IconButton>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setSchedulerDialogOpen(false)}>关闭</Button>
        </DialogActions>
      </Dialog>

      {/* 添加/编辑订阅表单对话框 */}
      <Dialog open={schedulerFormOpen} onClose={() => setSchedulerFormOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>{isEditScheduler ? '编辑订阅' : '添加订阅'}</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <TextField
              fullWidth
              label="名称"
              value={schedulerForm.name}
              onChange={(e) => setSchedulerForm({ ...schedulerForm, name: e.target.value })}
            />
            <TextField
              fullWidth
              label="URL"
              value={schedulerForm.url}
              onChange={(e) => setSchedulerForm({ ...schedulerForm, url: e.target.value })}
            />
            <Autocomplete
              freeSolo
              options={CRON_OPTIONS}
              getOptionLabel={(option) => (typeof option === 'string' ? option : option.value)}
              value={schedulerForm.cron_expr}
              onChange={(e, newValue) => {
                const value = typeof newValue === 'string' ? newValue : newValue?.value || '';
                setSchedulerForm({ ...schedulerForm, cron_expr: value });
              }}
              onInputChange={(e, newValue) => setSchedulerForm({ ...schedulerForm, cron_expr: newValue || '' })}
              renderOption={(props, option) => (
                <Box component="li" {...props} key={option.value}>
                  <Box>
                    <Typography variant="body2">{option.label}</Typography>
                    <Typography variant="caption" color="textSecondary">
                      {option.value}
                    </Typography>
                  </Box>
                </Box>
              )}
              renderInput={(params) => (
                <TextField
                  {...params}
                  label="Cron表达式"
                  placeholder="分 时 日 月 周"
                  helperText="格式: 分 时 日 月 周，如 0 */6 * * * 表示每6小时"
                />
              )}
            />
            <Autocomplete
              freeSolo
              options={groupOptions}
              value={schedulerForm.group}
              onChange={(e, newValue) => setSchedulerForm({ ...schedulerForm, group: newValue || '' })}
              onInputChange={(e, newValue) => setSchedulerForm({ ...schedulerForm, group: newValue || '' })}
              renderInput={(params) => (
                <TextField {...params} label="分组" helperText="设置分组后，从此订阅导入的所有节点将自动归属到此分组" />
              )}
            />
            <FormControlLabel
              control={
                <Switch
                  checked={schedulerForm.enabled}
                  onChange={(e) => setSchedulerForm({ ...schedulerForm, enabled: e.target.checked })}
                />
              }
              label="启用"
            />
            <FormControlLabel
              control={
                <Switch
                  checked={schedulerForm.download_with_proxy}
                  onChange={(e) => setSchedulerForm({ ...schedulerForm, download_with_proxy: e.target.checked })}
                />
              }
              label="使用代理下载"
            />
            {schedulerForm.download_with_proxy && (
              <Box>
                <Autocomplete
                  options={nodes}
                  getOptionLabel={(option) => option.Name || ''}
                  value={nodes.find((n) => n.Link === schedulerForm.proxy_link) || null}
                  onChange={(e, newValue) => setSchedulerForm({ ...schedulerForm, proxy_link: newValue?.Link || '' })}
                  renderOption={(props, option) => (
                    <Box component="li" {...props} key={option.ID}>
                      <Box sx={{ display: 'flex', justifyContent: 'space-between', width: '100%' }}>
                        <Typography variant="body2">{option.Name}</Typography>
                        <Typography variant="caption" color="textSecondary" sx={{ ml: 2 }}>
                          {option.Group || '未分组'}
                        </Typography>
                      </Box>
                    </Box>
                  )}
                  renderInput={(params) => <TextField {...params} label="选择代理节点" placeholder="留空则自动选择最佳节点" />}
                />
                <Typography variant="caption" color="textSecondary" sx={{ mt: 0.5, display: 'block' }}>
                  如果未选择具体代理，系统将自动选择延迟最低且速度最快的节点作为下载代理
                </Typography>
              </Box>
            )}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setSchedulerFormOpen(false)}>取消</Button>
          <Button variant="contained" onClick={handleSubmitScheduler}>
            确定
          </Button>
        </DialogActions>
      </Dialog>

      {/* 测速设置对话框 */}
      <Dialog open={speedTestDialogOpen} onClose={() => setSpeedTestDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>测速设置</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <FormControlLabel
              control={
                <Switch
                  checked={speedTestForm.enabled}
                  onChange={(e) => setSpeedTestForm({ ...speedTestForm, enabled: e.target.checked })}
                />
              }
              label="启用自动测速"
            />
            <Autocomplete
              freeSolo
              options={CRON_OPTIONS}
              getOptionLabel={(option) => (typeof option === 'string' ? option : option.value)}
              value={speedTestForm.cron}
              onChange={(e, newValue) => {
                const value = typeof newValue === 'string' ? newValue : newValue?.value || '';
                setSpeedTestForm({ ...speedTestForm, cron: value });
              }}
              onInputChange={(e, newValue) => setSpeedTestForm({ ...speedTestForm, cron: newValue || '' })}
              renderOption={(props, option) => (
                <Box component="li" {...props} key={option.value}>
                  <Box>
                    <Typography variant="body2">{option.label}</Typography>
                    <Typography variant="caption" color="textSecondary">
                      {option.value}
                    </Typography>
                  </Box>
                </Box>
              )}
              renderInput={(params) => (
                <TextField
                  {...params}
                  label="Cron表达式"
                  placeholder="分 时 日 月 周"
                  helperText="格式: 分 时 日 月 周 (例如: 0 */1 * * * 表示每小时执行一次)"
                />
              )}
            />
            <FormControl fullWidth>
              <InputLabel>测速模式</InputLabel>
              <Select
                variant={"outlined"}
                value={speedTestForm.mode}
                label="测速模式"
                onChange={(e) => handleSpeedModeChange(e.target.value)}
              >
                <MenuItem value="tcp">Mihomo - 仅延迟测试 (更快)</MenuItem>
                <MenuItem value="mihomo">Mihomo - 真速度测试 (延迟+下载速度)</MenuItem>
              </Select>
            </FormControl>
            <Box>
              <Autocomplete
                freeSolo
                options={speedTestForm.mode === 'mihomo' ? SPEED_TEST_MIHOMO_OPTIONS : SPEED_TEST_TCP_OPTIONS}
                getOptionLabel={(option) => (typeof option === 'string' ? option : option.value)}
                value={speedTestForm.url}
                onChange={(e, newValue) => {
                  const value = typeof newValue === 'string' ? newValue : newValue?.value || '';
                  setSpeedTestForm({ ...speedTestForm, url: value });
                }}
                onInputChange={(e, newValue) => setSpeedTestForm({ ...speedTestForm, url: newValue || '' })}
                renderOption={(props, option) => (
                  <Box component="li" {...props} key={option.value}>
                    <Box>
                      <Typography variant="body2">{option.label}</Typography>
                      <Typography variant="caption" color="textSecondary" sx={{ wordBreak: 'break-all' }}>
                        {option.value}
                      </Typography>
                    </Box>
                  </Box>
                )}
                renderInput={(params) => (
                  <TextField
                    {...params}
                    label="测速URL"
                    placeholder={speedTestForm.mode === 'mihomo' ? '请选择或输入下载测速URL' : '请选择或输入204测速URL'}
                  />
                )}
              />
              <Typography variant="caption" color="textSecondary" sx={{ mt: 0.5, display: 'block' }}>
                可以自定义测速URL。
                {speedTestForm.mode === 'mihomo'
                  ? '真速度测试使用可下载资源地址，例如: https://speed.cloudflare.com/__down?bytes=10000000'
                  : '延迟测试使用更轻量的204测试地址，例如: http://cp.cloudflare.com/generate_204'}
              </Typography>
            </Box>
            <TextField
              fullWidth
              label="超时时间"
              type="number"
              value={speedTestForm.timeout}
              onChange={(e) => setSpeedTestForm({ ...speedTestForm, timeout: Number(e.target.value) })}
              InputProps={{ endAdornment: <InputAdornment position="end">秒</InputAdornment> }}
            />
            <Autocomplete
              multiple
              freeSolo
              options={groupOptions}
              value={speedTestForm.groups || []}
              onChange={(e, newValue) => setSpeedTestForm({ ...speedTestForm, groups: newValue })}
              renderInput={(params) => <TextField {...params} label="测速分组" placeholder="留空则测试全部分组" />}
            />
            <FormControlLabel
              control={
                <Switch
                  checked={speedTestForm.detect_country}
                  onChange={(e) => setSpeedTestForm({ ...speedTestForm, detect_country: e.target.checked })}
                />
              }
              label="检测落地IP国家"
            />
            <Typography variant="caption" color="textSecondary" sx={{ mt: -1 }}>
              开启后，测速时会通过代理获取落地IP并解析对应的国家代码，会降低测速效率。IP通过https://api.ip.sb/ip获取。
            </Typography>
            <Button variant="outlined" startIcon={<PlayArrowIcon />} onClick={handleRunSpeedTest}>
              立即测速
            </Button>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setSpeedTestDialogOpen(false)}>取消</Button>
          <Button variant="contained" onClick={handleSubmitSpeedTest}>
            保存
          </Button>
        </DialogActions>
      </Dialog>

      {/* 提示消息 */}
      <Snackbar
        open={snackbar.open}
        autoHideDuration={3000}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
      >
        <Alert severity={snackbar.severity} variant="filled" sx={{ width: '100%' }}>
          {snackbar.message}
        </Alert>
      </Snackbar>
      {/* 确认对话框 */}
      <Dialog
        open={confirmOpen}
        onClose={handleConfirmClose}
        aria-labelledby="alert-dialog-title"
        aria-describedby="alert-dialog-description"
      >
        <DialogTitle id="alert-dialog-title">{confirmInfo.title}</DialogTitle>
        <DialogContent>
          <DialogContentText id="alert-dialog-description">{confirmInfo.content}</DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleConfirmClose}>取消</Button>
          <Button onClick={handleConfirmAction} color="primary" autoFocus>
            确定
          </Button>
        </DialogActions>
      </Dialog>

      {/* 删除订阅对话框 */}
      <Dialog open={deleteSchedulerDialogOpen} onClose={() => setDeleteSchedulerDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>删除订阅</DialogTitle>
        <DialogContent>
          <Typography variant="body1" gutterBottom>
            确定要删除订阅 "{deleteSchedulerTarget?.Name}" 吗？
          </Typography>
          {(deleteSchedulerTarget?.node_count || 0) > 0 && (
            <>
              <Typography variant="body2" color="textSecondary" sx={{ mb: 2 }}>
                该订阅关联了 {deleteSchedulerTarget?.node_count || 0} 个节点
              </Typography>
              <FormControlLabel
                control={<Checkbox checked={deleteSchedulerWithNodes} onChange={(e) => setDeleteSchedulerWithNodes(e.target.checked)} />}
                label="同时删除关联的节点"
              />
              {!deleteSchedulerWithNodes && (
                <Alert severity="info" sx={{ mt: 1 }}>
                  保留的节点将变为手动添加的节点，不再与此订阅关联
                </Alert>
              )}
            </>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteSchedulerDialogOpen(false)}>取消</Button>
          <Button onClick={handleConfirmDeleteScheduler} color="error" variant="contained">
            确认删除
          </Button>
        </DialogActions>
      </Dialog>
    </MainCard>
  );
}
