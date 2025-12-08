import { useState, useEffect, useMemo, useCallback, Fragment } from 'react';
import { DragDropContext, Droppable, Draggable } from '@hello-pangea/dnd';
import { QRCodeSVG } from 'qrcode.react';
import md5 from 'md5';

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
import Typography from '@mui/material/Typography';
import Autocomplete from '@mui/material/Autocomplete';
import Tooltip from '@mui/material/Tooltip';
import InputAdornment from '@mui/material/InputAdornment';
import Collapse from '@mui/material/Collapse';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemText from '@mui/material/ListItemText';
import ListItemIcon from '@mui/material/ListItemIcon';
import Divider from '@mui/material/Divider';
import Grid from '@mui/material/Grid';
import Tabs from '@mui/material/Tabs';
import Tab from '@mui/material/Tab';
import Fade from '@mui/material/Fade';

// icons
import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import RefreshIcon from '@mui/icons-material/Refresh';
import QrCode2Icon from '@mui/icons-material/QrCode2';
import HistoryIcon from '@mui/icons-material/History';
import SortIcon from '@mui/icons-material/Sort';
import CheckIcon from '@mui/icons-material/Check';
import CloseIcon from '@mui/icons-material/Close';
import KeyboardArrowDownIcon from '@mui/icons-material/KeyboardArrowDown';
import KeyboardArrowUpIcon from '@mui/icons-material/KeyboardArrowUp';
import DragIndicatorIcon from '@mui/icons-material/DragIndicator';
import ChevronRightIcon from '@mui/icons-material/ChevronRight';
import ChevronLeftIcon from '@mui/icons-material/ChevronLeft';
import SearchIcon from '@mui/icons-material/Search';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import { getSubscriptions, addSubscription, updateSubscription, deleteSubscription, sortSubscription } from 'api/subscriptions';
import { getNodes, getNodeCountries } from 'api/nodes';
import { getTemplates } from 'api/templates';
import { getScripts } from 'api/scripts';

// ISO国家代码转换为国旗emoji
const isoToFlag = (isoCode) => {
  if (!isoCode || isoCode.length !== 2) return '';
  const codePoints = isoCode
    .toUpperCase()
    .split('')
    .map((char) => 0x1f1e6 + char.charCodeAt(0) - 65);
  return String.fromCodePoint(...codePoints);
};

// 格式化国家显示 (国旗emoji + 代码)
const formatCountry = (linkCountry) => {
  if (!linkCountry) return '';
  const flag = isoToFlag(linkCountry);
  return flag ? `${flag}${linkCountry}` : linkCountry;
};

// ==============================|| 订阅管理 ||============================== //

export default function SubscriptionList() {
  const theme = useTheme();
  const matchDownMd = useMediaQuery(theme.breakpoints.down('md'));

  const [subscriptions, setSubscriptions] = useState([]);
  const [allNodes, setAllNodes] = useState([]);
  const [templates, setTemplates] = useState([]);
  const [scripts, setScripts] = useState([]);
  const [loading, setLoading] = useState(false);

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

  // 表单对话框
  const [dialogOpen, setDialogOpen] = useState(false);
  const [isEdit, setIsEdit] = useState(false);
  const [currentSub, setCurrentSub] = useState(null);

  // 表单数据
  const [formData, setFormData] = useState({
    name: '',
    clash: './template/clash.yaml',
    surge: './template/surge.conf',
    udp: false,
    cert: false,
    selectionMode: 'nodes', // nodes, groups, mixed
    selectedNodes: [],
    selectedGroups: [],
    selectedScripts: [],
    IPWhitelist: '',
    IPBlacklist: '',
    DelayTime: 0,
    MinSpeed: 0,
    CountryWhitelist: [],
    CountryBlacklist: [],
    nodeNameRule: ''
  });

  // 预览节点名称
  const previewNodeName = (rule) => {
    if (!rule) return '';
    return rule
      .replace(/\$Name/g, '香港节点-备注')
      .replace(/\$LinkName/g, '香港01')
      .replace(/\$LinkCountry/g, 'HK')
      .replace(/\$Speed/g, '1.50MB/s')
      .replace(/\$Delay/g, '125ms')
      .replace(/\$Group/g, 'Premium')
      .replace(/\$Source/g, '机场A')
      .replace(/\$Index/g, '1')
      .replace(/\$Protocol/g, 'VMess');
  };

  // 节点过滤
  const [nodeGroupFilter, setNodeGroupFilter] = useState('all');
  const [nodeSourceFilter, setNodeSourceFilter] = useState('all');
  const [nodeSearchQuery, setNodeSearchQuery] = useState('');
  const [nodeCountryFilter, setNodeCountryFilter] = useState([]);
  const [countryOptions, setCountryOptions] = useState([]);

  // QR码对话框
  const [qrDialogOpen, setQrDialogOpen] = useState(false);
  const [qrUrl, setQrUrl] = useState('');
  const [qrTitle, setQrTitle] = useState('');

  // 客户端对话框
  const [clientDialogOpen, setClientDialogOpen] = useState(false);
  const [clientUrls, setClientUrls] = useState({});

  // 访问记录对话框
  const [logsDialogOpen, setLogsDialogOpen] = useState(false);
  const [currentLogs, setCurrentLogs] = useState([]);

  // 排序模式
  const [sortingSubId, setSortingSubId] = useState(null);
  const [tempSortData, setTempSortData] = useState([]);

  // 展开行
  const [expandedRows, setExpandedRows] = useState({});

  // 穿梭框状态
  const [checkedAvailable, setCheckedAvailable] = useState([]);
  const [checkedSelected, setCheckedSelected] = useState([]);
  const [mobileTab, setMobileTab] = useState(0);
  const [selectedNodeSearch, setSelectedNodeSearch] = useState('');

  // 分页
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(10);

  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });

  // 获取分组列表
  const groupOptions = useMemo(() => {
    const groups = new Set();
    allNodes.forEach((node) => {
      if (node.Group) groups.add(node.Group);
    });
    return Array.from(groups).sort();
  }, [allNodes]);

  // 获取来源列表
  const sourceOptions = useMemo(() => {
    const sources = new Set();
    allNodes.forEach((node) => {
      if (node.Source) sources.add(node.Source);
    });
    return Array.from(sources).sort();
  }, [allNodes]);

  // 按分组统计节点数量
  const groupNodeCounts = useMemo(() => {
    const counts = {};
    allNodes.forEach((node) => {
      const group = node.Group || '未分组';
      counts[group] = (counts[group] || 0) + 1;
    });
    return counts;
  }, [allNodes]);

  // 过滤后的节点列表
  const filteredNodes = useMemo(() => {
    return allNodes.filter((node) => {
      if (nodeGroupFilter !== 'all' && node.Group !== nodeGroupFilter) return false;
      if (nodeSourceFilter !== 'all' && node.Source !== nodeSourceFilter) return false;
      if (nodeSearchQuery) {
        const query = nodeSearchQuery.toLowerCase();
        if (!node.Name?.toLowerCase().includes(query) && !node.Group?.toLowerCase().includes(query)) {
          return false;
        }
      }
      // 国家代码过滤
      if (nodeCountryFilter.length > 0) {
        if (!node.LinkCountry || !nodeCountryFilter.includes(node.LinkCountry)) {
          return false;
        }
      }
      return true;
    });
  }, [allNodes, nodeGroupFilter, nodeSourceFilter, nodeSearchQuery, nodeCountryFilter]);

  // 可选节点（排除已选）
  const availableNodes = useMemo(() => {
    return filteredNodes.filter((node) => !formData.selectedNodes.includes(node.Name));
  }, [filteredNodes, formData.selectedNodes]);

  // 已选节点
  const selectedNodesList = useMemo(() => {
    return allNodes.filter((node) => formData.selectedNodes.includes(node.Name));
  }, [allNodes, formData.selectedNodes]);

  // 获取数据
  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const [subRes, nodesRes, templatesRes, scriptsRes, countriesRes] = await Promise.all([
        getSubscriptions(),
        getNodes(),
        getTemplates(),
        getScripts(),
        getNodeCountries()
      ]);
      setSubscriptions(subRes.data || []);
      setAllNodes(nodesRes.data || []);
      setTemplates(templatesRes.data || []);
      setScripts(scriptsRes.data || []);
      setCountryOptions(countriesRes.data || []);
    } catch (error) {
      showMessage('获取数据失败', 'error');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const showMessage = (message, severity = 'success') => {
    setSnackbar({ open: true, message, severity });
  };

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
    showMessage('已复制到剪贴板');
  };

  const getServerUrl = () => {
    return `${window.location.protocol}//${window.location.hostname}${window.location.port ? ':' + window.location.port : ''}`;
  };

  // === 订阅操作 ===
  const handleAdd = () => {
    setIsEdit(false);
    setCurrentSub(null);
    setFormData({
      name: '',
      clash: './template/clash.yaml',
      surge: './template/surge.conf',
      udp: false,
      cert: false,
      selectionMode: 'nodes',
      selectedNodes: [],
      selectedGroups: [],
      selectedScripts: [],
      IPWhitelist: '',
      IPBlacklist: '',
      DelayTime: 0,
      MinSpeed: 0,
      CountryWhitelist: [],
      CountryBlacklist: [],
      nodeNameRule: ''
    });
    setNodeGroupFilter('all');
    setNodeSourceFilter('all');
    setNodeSearchQuery('');
    setNodeCountryFilter([]);
    setDialogOpen(true);
  };

  const handleEdit = (sub) => {
    setIsEdit(true);
    setCurrentSub(sub);
    const config = typeof sub.Config === 'string' ? JSON.parse(sub.Config) : sub.Config;

    // 解析已选节点
    const nodes = sub.Nodes?.map((n) => n.Name) || [];
    const groups = (sub.Groups || []).map((g) => (typeof g === 'string' ? g : g.Name));
    const scriptIds = (sub.Scripts || []).map((s) => s.id);

    // 判断选择模式
    let mode = 'nodes';
    if (nodes.length > 0 && groups.length > 0) {
      mode = 'mixed';
    } else if (groups.length > 0) {
      mode = 'groups';
    }

    setFormData({
      name: sub.Name,
      clash: config?.clash || './template/clash.yaml',
      surge: config?.surge || './template/surge.conf',
      udp: config?.udp || false,
      cert: config?.cert || false,
      selectionMode: mode,
      selectedNodes: nodes,
      selectedGroups: groups,
      selectedScripts: scriptIds,
      IPWhitelist: sub.IPWhitelist || '',
      IPBlacklist: sub.IPBlacklist || '',
      DelayTime: sub.DelayTime || 0,
      MinSpeed: sub.MinSpeed || 0,
      CountryWhitelist: sub.CountryWhitelist ? sub.CountryWhitelist.split(',').filter((c) => c.trim()) : [],
      CountryBlacklist: sub.CountryBlacklist ? sub.CountryBlacklist.split(',').filter((c) => c.trim()) : [],
      nodeNameRule: sub.NodeNameRule || ''
    });
    setNodeGroupFilter('all');
    setNodeSourceFilter('all');
    setNodeSearchQuery('');
    setNodeCountryFilter([]);
    setDialogOpen(true);
  };

  const handleDelete = async (sub) => {
    openConfirm('删除订阅', `确定要删除订阅 "${sub.Name}" 吗？`, async () => {
      try {
        await deleteSubscription({ id: sub.ID });
        showMessage('删除成功');
        fetchData();
      } catch (error) {
        showMessage('删除失败', 'error');
      }
    });
  };

  const handleSubmit = async () => {
    if (!formData.name.trim()) {
      showMessage('请输入订阅名称', 'warning');
      return;
    }

    try {
      const config = JSON.stringify({
        clash: formData.clash,
        surge: formData.surge,
        udp: formData.udp,
        cert: formData.cert
      });

      const requestData = {
        name: formData.name.trim(),
        config,
        IPWhitelist: formData.IPWhitelist,
        IPBlacklist: formData.IPBlacklist,
        DelayTime: formData.DelayTime,
        MinSpeed: formData.MinSpeed,
        scripts: formData.selectedScripts.join(','),
        CountryWhitelist: formData.CountryWhitelist.join(','),
        CountryBlacklist: formData.CountryBlacklist.join(','),
        NodeNameRule: formData.nodeNameRule
      };

      if (formData.selectionMode === 'nodes') {
        requestData.nodes = formData.selectedNodes.join(',');
        requestData.groups = '';
      } else if (formData.selectionMode === 'groups') {
        requestData.nodes = '';
        requestData.groups = formData.selectedGroups.join(',');
      } else {
        requestData.nodes = formData.selectedNodes.join(',');
        requestData.groups = formData.selectedGroups.join(',');
      }

      if (isEdit) {
        requestData.oldname = currentSub.Name;
        await updateSubscription(requestData);
        showMessage('更新成功');
      } else {
        await addSubscription(requestData);
        showMessage('添加成功');
      }
      setDialogOpen(false);
      fetchData();
    } catch (error) {
      showMessage(isEdit ? '更新失败' : '添加失败', 'error');
    }
  };

  // 节点选择操作
  const handleAddNode = (nodeName) => {
    setFormData({ ...formData, selectedNodes: [...formData.selectedNodes, nodeName] });
    setCheckedAvailable(checkedAvailable.filter((n) => n !== nodeName));
  };

  const handleRemoveNode = (nodeName) => {
    setFormData({ ...formData, selectedNodes: formData.selectedNodes.filter((n) => n !== nodeName) });
    setCheckedSelected(checkedSelected.filter((n) => n !== nodeName));
  };

  const handleAddAllVisible = () => {
    const newNodes = [...formData.selectedNodes, ...availableNodes.map((n) => n.Name)];
    setFormData({ ...formData, selectedNodes: newNodes });
    setCheckedAvailable([]);
  };

  const handleRemoveAll = () => {
    setFormData({ ...formData, selectedNodes: [] });
    setCheckedSelected([]);
  };

  // 多选操作
  const handleToggleAvailable = (nodeName) => {
    if (checkedAvailable.includes(nodeName)) {
      setCheckedAvailable(checkedAvailable.filter((n) => n !== nodeName));
    } else {
      setCheckedAvailable([...checkedAvailable, nodeName]);
    }
  };

  const handleToggleSelected = (nodeName) => {
    if (checkedSelected.includes(nodeName)) {
      setCheckedSelected(checkedSelected.filter((n) => n !== nodeName));
    } else {
      setCheckedSelected([...checkedSelected, nodeName]);
    }
  };

  const handleAddChecked = () => {
    const newNodes = [...formData.selectedNodes, ...checkedAvailable];
    setFormData({ ...formData, selectedNodes: newNodes });
    setCheckedAvailable([]);
  };

  const handleRemoveChecked = () => {
    const newNodes = formData.selectedNodes.filter((n) => !checkedSelected.includes(n));
    setFormData({ ...formData, selectedNodes: newNodes });
    setCheckedSelected([]);
  };

  const handleToggleAllAvailable = () => {
    if (checkedAvailable.length === availableNodes.length) {
      setCheckedAvailable([]);
    } else {
      setCheckedAvailable(availableNodes.map((n) => n.Name));
    }
  };

  const handleToggleAllSelected = () => {
    if (checkedSelected.length === selectedNodesList.length) {
      setCheckedSelected([]);
    } else {
      setCheckedSelected(selectedNodesList.map((n) => n.Name));
    }
  };

  // 筛选已选节点
  const filteredSelectedNodes = useMemo(() => {
    if (!selectedNodeSearch) return selectedNodesList;
    const query = selectedNodeSearch.toLowerCase();
    return selectedNodesList.filter((node) => node.Name?.toLowerCase().includes(query) || node.Group?.toLowerCase().includes(query));
  }, [selectedNodesList, selectedNodeSearch]);

  // === 客户端/QR码 ===
  const handleClient = (name) => {
    const serverUrl = getServerUrl();
    const token = md5(name);
    const baseUrl = `${serverUrl}/c/?token=${token}`;
    setClientUrls({
      自动识别: baseUrl,
      v2ray: baseUrl,
      clash: baseUrl,
      surge: baseUrl
    });
    setClientDialogOpen(true);
  };

  const handleQrcode = (url, title) => {
    setQrUrl(url);
    setQrTitle(title);
    setQrDialogOpen(true);
  };

  // === 访问记录 ===
  const handleLogs = (sub) => {
    setCurrentLogs(sub.SubLogs || []);
    setLogsDialogOpen(true);
  };

  // === 排序功能 ===
  const handleStartSort = (sub) => {
    setSortingSubId(sub.ID);
    // 合并节点和分组
    const sortData = [];
    (sub.Nodes || []).forEach((node, idx) => {
      sortData.push({
        Name: node.Name,
        Sort: node.Sort !== undefined ? node.Sort : idx,
        IsGroup: false
      });
    });
    (sub.Groups || []).forEach((group, idx) => {
      const g = typeof group === 'string' ? { Name: group, Sort: sub.Nodes?.length + idx } : group;
      sortData.push({
        Name: g.Name,
        Sort: g.Sort !== undefined ? g.Sort : sub.Nodes?.length + idx,
        IsGroup: true
      });
    });
    sortData.sort((a, b) => a.Sort - b.Sort);
    setTempSortData(sortData);
    showMessage('已进入排序模式，拖动项目进行排序', 'info');
  };

  const handleConfirmSort = async (sub) => {
    // 重新分配Sort值
    const newSortData = tempSortData.map((item, idx) => ({ ...item, Sort: idx }));
    try {
      await sortSubscription({
        ID: sub.ID,
        NodeSort: newSortData
      });
      showMessage('排序已更新');
      setSortingSubId(null);
      setTempSortData([]);
      fetchData();
    } catch (error) {
      showMessage('排序保存失败', 'error');
    }
  };

  const handleCancelSort = () => {
    setSortingSubId(null);
    setTempSortData([]);
    showMessage('已取消排序', 'info');
  };

  const onDragEnd = (result) => {
    if (!result.destination) return;
    const items = Array.from(tempSortData);
    const [reorderedItem] = items.splice(result.source.index, 1);
    items.splice(result.destination.index, 0, reorderedItem);
    setTempSortData(items);
  };

  // 展开/折叠行
  const toggleRow = (subId) => {
    setExpandedRows({ ...expandedRows, [subId]: !expandedRows[subId] });
  };

  const getSortedItems = (sub) => {
    const items = [];
    (sub.Nodes || []).forEach((node, idx) => {
      items.push({
        ...node,
        _type: 'node',
        _sort: node.Sort !== undefined ? node.Sort : idx
      });
    });
    (sub.Groups || []).forEach((group, idx) => {
      const g = typeof group === 'string' ? { Name: group } : group;
      items.push({
        ...g,
        _type: 'group',
        _sort: g.Sort !== undefined ? g.Sort : (sub.Nodes?.length || 0) + idx
      });
    });
    return items.sort((a, b) => a._sort - b._sort);
  };

  return (
    <MainCard
      title="订阅管理"
      secondary={
        matchDownMd ? (
          <Tooltip title="添加订阅/刷新">
            <Button variant="contained" startIcon={<AddIcon />} onClick={handleAdd}>
              添加
            </Button>
          </Tooltip>
        ) : (
          <Stack direction="row" spacing={1}>
            <Button variant="contained" startIcon={<AddIcon />} onClick={handleAdd}>
              添加订阅
            </Button>
            <IconButton onClick={fetchData} disabled={loading}>
              <RefreshIcon />
            </IconButton>
          </Stack>
        )
      }
    >
      {matchDownMd && (
        <Stack direction="row" justifyContent="flex-end" sx={{ mb: 2 }}>
          <IconButton onClick={fetchData} disabled={loading} size="small">
            <RefreshIcon />
          </IconButton>
        </Stack>
      )}

      {matchDownMd ? (
        <Stack spacing={2}>
          {subscriptions.slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage).map((sub) => (
            <MainCard key={sub.ID} content={false} border shadow={theme.shadows[1]}>
              <Box p={2}>
                <Stack direction="row" justifyContent="space-between" alignItems="center" mb={1} onClick={() => toggleRow(sub.ID)}>
                  <Stack direction="row" spacing={1} alignItems="center">
                    <Chip label={sub.Name} color="primary" />
                    {sortingSubId === sub.ID && <Chip label="排序中" color="warning" size="small" />}
                  </Stack>
                  {expandedRows[sub.ID] || sortingSubId === sub.ID ? <KeyboardArrowUpIcon /> : <KeyboardArrowDownIcon />}
                </Stack>

                <Typography variant="body2" sx={{ mb: 1 }}>
                  {sub.Nodes?.length || 0} 个节点, {sub.Groups?.length || 0} 个分组
                </Typography>

                <Divider sx={{ my: 1 }} />

                <Stack direction="row" justifyContent="space-between" alignItems="center">
                  <Typography variant="caption" color="textSecondary">
                    {sub.CreateDate}
                  </Typography>
                  <Stack direction="row" spacing={0}>
                    <IconButton size="small" onClick={() => handleClient(sub.Name)}>
                      <QrCode2Icon fontSize="small" />
                    </IconButton>
                    <IconButton size="small" onClick={() => handleLogs(sub)}>
                      <HistoryIcon fontSize="small" />
                    </IconButton>
                    <IconButton size="small" onClick={() => handleEdit(sub)}>
                      <EditIcon fontSize="small" />
                    </IconButton>
                    <IconButton size="small" color="error" onClick={() => handleDelete(sub)}>
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                    {sortingSubId !== sub.ID ? (
                      <IconButton
                        size="small"
                        color="warning"
                        onClick={(e) => {
                          e.stopPropagation();
                          handleStartSort(sub);
                        }}
                      >
                        <SortIcon fontSize="small" />
                      </IconButton>
                    ) : (
                      <>
                        <IconButton
                          size="small"
                          color="success"
                          onClick={(e) => {
                            e.stopPropagation();
                            handleConfirmSort(sub);
                          }}
                        >
                          <CheckIcon fontSize="small" />
                        </IconButton>
                        <IconButton
                          size="small"
                          onClick={(e) => {
                            e.stopPropagation();
                            handleCancelSort();
                          }}
                        >
                          <CloseIcon fontSize="small" />
                        </IconButton>
                      </>
                    )}
                  </Stack>
                </Stack>

                {/* Expandable Content for Sort or Details */}
                <Collapse in={expandedRows[sub.ID] || sortingSubId === sub.ID} timeout="auto" unmountOnExit>
                  <Box sx={{ mt: 2 }}>
                    {sortingSubId === sub.ID ? (
                      <DragDropContext onDragEnd={onDragEnd}>
                        <Droppable droppableId="sortList">
                          {(provided) => (
                            <List {...provided.droppableProps} ref={provided.innerRef} dense>
                              {tempSortData.map((item, index) => (
                                <Draggable key={item.Name} draggableId={item.Name} index={index}>
                                  {(provided, snapshot) => (
                                    <ListItem
                                      ref={provided.innerRef}
                                      {...provided.draggableProps}
                                      {...provided.dragHandleProps}
                                      sx={{
                                        bgcolor: snapshot.isDragging ? 'action.selected' : 'background.paper',
                                        border: '1px solid',
                                        borderColor: 'divider',
                                        borderRadius: 1,
                                        mb: 0.5
                                      }}
                                    >
                                      <DragIndicatorIcon sx={{ mr: 1, color: 'text.secondary' }} />
                                      <Chip
                                        label={item.IsGroup ? `📁 ${item.Name} (分组)` : item.Name}
                                        color={item.IsGroup ? 'warning' : 'success'}
                                        variant="outlined"
                                        size="small"
                                      />
                                    </ListItem>
                                  )}
                                </Draggable>
                              ))}
                              {provided.placeholder}
                            </List>
                          )}
                        </Droppable>
                      </DragDropContext>
                    ) : (
                      <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                        {getSortedItems(sub).map((item, idx) =>
                          item._type === 'node' ? (
                            <Chip
                              key={item._type + item.ID}
                              label={item.Name}
                              size="small"
                              variant="outlined"
                              color="success"
                              onClick={() => copyToClipboard(item.Link)}
                              sx={{ mb: 1 }}
                            />
                          ) : (
                            <Chip
                              key={item._type + idx}
                              label={`📁 ${item.Name}`}
                              size="small"
                              variant="outlined"
                              color="warning"
                              sx={{ mb: 1 }}
                            />
                          )
                        )}
                      </Stack>
                    )}
                  </Box>
                </Collapse>
              </Box>
            </MainCard>
          ))}
        </Stack>
      ) : (
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell width={50} />
                <TableCell>订阅名称</TableCell>
                <TableCell>节点/分组</TableCell>
                <TableCell>创建时间</TableCell>
                <TableCell align="right" width={350}>
                  操作
                </TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {subscriptions.slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage).map((sub) => (
                <Fragment key={sub.ID}>
                  <TableRow hover>
                    <TableCell>
                      <IconButton size="small" onClick={() => toggleRow(sub.ID)}>
                        {expandedRows[sub.ID] ? <KeyboardArrowUpIcon /> : <KeyboardArrowDownIcon />}
                      </IconButton>
                    </TableCell>
                    <TableCell>
                      <Chip label={sub.Name} color="primary" />
                      {sortingSubId === sub.ID && <Chip label="排序中" color="warning" size="small" sx={{ ml: 1 }} />}
                    </TableCell>
                    <TableCell>
                      <Typography variant="body2">
                        {sub.Nodes?.length || 0} 个节点, {sub.Groups?.length || 0} 个分组
                      </Typography>
                    </TableCell>
                    <TableCell>{sub.CreateDate}</TableCell>
                    <TableCell align="right">
                      <Stack direction="row" spacing={0.5} justifyContent="flex-end">
                        <Tooltip title="客户端">
                          <IconButton size="small" onClick={() => handleClient(sub.Name)}>
                            <QrCode2Icon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="访问记录">
                          <IconButton size="small" onClick={() => handleLogs(sub)}>
                            <HistoryIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="编辑">
                          <IconButton size="small" onClick={() => handleEdit(sub)}>
                            <EditIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="删除">
                          <IconButton size="small" color="error" onClick={() => handleDelete(sub)}>
                            <DeleteIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                        {sortingSubId !== sub.ID ? (
                          <Tooltip title="排序">
                            <IconButton size="small" color="warning" onClick={() => handleStartSort(sub)}>
                              <SortIcon fontSize="small" />
                            </IconButton>
                          </Tooltip>
                        ) : (
                          <>
                            <Tooltip title="确定">
                              <IconButton size="small" color="success" onClick={() => handleConfirmSort(sub)}>
                                <CheckIcon fontSize="small" />
                              </IconButton>
                            </Tooltip>
                            <Tooltip title="取消">
                              <IconButton size="small" onClick={handleCancelSort}>
                                <CloseIcon fontSize="small" />
                              </IconButton>
                            </Tooltip>
                          </>
                        )}
                      </Stack>
                    </TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell style={{ paddingBottom: 0, paddingTop: 0 }} colSpan={6}>
                      <Collapse in={expandedRows[sub.ID] || sortingSubId === sub.ID} timeout="auto" unmountOnExit>
                        <Box sx={{ margin: 2 }}>
                          {sortingSubId === sub.ID ? (
                            <DragDropContext onDragEnd={onDragEnd}>
                              <Droppable droppableId="sortList">
                                {(provided) => (
                                  <List {...provided.droppableProps} ref={provided.innerRef} dense>
                                    {tempSortData.map((item, index) => (
                                      <Draggable key={item.Name} draggableId={item.Name} index={index}>
                                        {(provided, snapshot) => (
                                          <ListItem
                                            ref={provided.innerRef}
                                            {...provided.draggableProps}
                                            {...provided.dragHandleProps}
                                            sx={{
                                              bgcolor: snapshot.isDragging ? 'action.selected' : 'background.paper',
                                              border: '1px solid',
                                              borderColor: 'divider',
                                              borderRadius: 1,
                                              mb: 0.5
                                            }}
                                          >
                                            <DragIndicatorIcon sx={{ mr: 1, color: 'text.secondary' }} />
                                            <Chip
                                              label={item.IsGroup ? `📁 ${item.Name} (分组)` : item.Name}
                                              color={item.IsGroup ? 'warning' : 'success'}
                                              variant="outlined"
                                              size="small"
                                            />
                                          </ListItem>
                                        )}
                                      </Draggable>
                                    ))}
                                    {provided.placeholder}
                                  </List>
                                )}
                              </Droppable>
                            </DragDropContext>
                          ) : (
                            <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                              {getSortedItems(sub).map((item, idx) =>
                                item._type === 'node' ? (
                                  <Chip
                                    key={item._type + item.ID}
                                    label={item.Name}
                                    size="small"
                                    variant="outlined"
                                    color="success"
                                    onClick={() => copyToClipboard(item.Link)}
                                  />
                                ) : (
                                  <Chip key={item._type + idx} label={`📁 ${item.Name}`} size="small" variant="outlined" color="warning" />
                                )
                              )}
                            </Stack>
                          )}
                        </Box>
                      </Collapse>
                    </TableCell>
                  </TableRow>
                </Fragment>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      <TablePagination
        component="div"
        count={subscriptions.length}
        page={page}
        onPageChange={(e, newPage) => setPage(newPage)}
        rowsPerPage={rowsPerPage}
        onRowsPerPageChange={(e) => {
          setRowsPerPage(parseInt(e.target.value, 10));
          setPage(0);
        }}
        labelRowsPerPage="每页行数:"
      />

      {/* 添加/编辑对话框 */}
      <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)} maxWidth="lg" fullWidth>
        <DialogTitle>{isEdit ? '编辑订阅' : '添加订阅'}</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <TextField
              fullWidth
              label="订阅名称"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            />

            <Grid container spacing={2}>
              <Grid item xs={6}>
                <FormControl fullWidth>
                  <InputLabel>Clash 模板</InputLabel>
                  <Select value={formData.clash} label="Clash 模板" onChange={(e) => setFormData({ ...formData, clash: e.target.value })}>
                    {templates.map((t) => (
                      <MenuItem key={t.file} value={`./template/${t.file}`}>
                        {t.file}
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>
              </Grid>
              <Grid item xs={6}>
                <FormControl fullWidth>
                  <InputLabel>Surge 模板</InputLabel>
                  <Select value={formData.surge} label="Surge 模板" onChange={(e) => setFormData({ ...formData, surge: e.target.value })}>
                    {templates.map((t) => (
                      <MenuItem key={t.file} value={`./template/${t.file}`}>
                        {t.file}
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>
              </Grid>
            </Grid>

            <Stack direction="row" spacing={2}>
              <FormControlLabel
                control={<Checkbox checked={formData.udp} onChange={(e) => setFormData({ ...formData, udp: e.target.checked })} />}
                label="强制开启 UDP"
              />
              <FormControlLabel
                control={<Checkbox checked={formData.cert} onChange={(e) => setFormData({ ...formData, cert: e.target.checked })} />}
                label="跳过证书验证"
              />
            </Stack>

            <Divider />

            {/* 选择模式 */}
            <Typography variant="subtitle1" fontWeight="bold">
              选择节点
            </Typography>
            <RadioGroup row value={formData.selectionMode} onChange={(e) => setFormData({ ...formData, selectionMode: e.target.value })}>
              <FormControlLabel value="nodes" control={<Radio />} label="手动选择节点" />
              <FormControlLabel value="groups" control={<Radio />} label="动态选择分组" />
              <FormControlLabel value="mixed" control={<Radio />} label="混合模式" />
            </RadioGroup>
            <Typography variant="caption" color="textSecondary">
              {formData.selectionMode === 'nodes' && '手动选择具体节点，节点不会随分组变化自动更新'}
              {formData.selectionMode === 'groups' && '选择分组，自动包含该分组下的所有节点，节点会随分组变化自动更新'}
              {formData.selectionMode === 'mixed' && '同时支持手动选择节点和动态选择分组'}
            </Typography>

            {/* 分组选择 */}
            {(formData.selectionMode === 'groups' || formData.selectionMode === 'mixed') && (
              <Autocomplete
                multiple
                options={groupOptions}
                value={formData.selectedGroups}
                onChange={(e, newValue) => setFormData({ ...formData, selectedGroups: newValue })}
                renderInput={(params) => <TextField {...params} label="选择分组（动态）" />}
                renderOption={(props, option) => (
                  <li {...props}>
                    {option} ({groupNodeCounts[option] || 0} 个节点)
                  </li>
                )}
              />
            )}

            {/* 节点选择 */}
            {(formData.selectionMode === 'nodes' || formData.selectionMode === 'mixed') && (
              <>
                <Grid container spacing={2}>
                  <Grid item xs={3}>
                    <FormControl fullWidth size="small">
                      <InputLabel>分组过滤</InputLabel>
                      <Select value={nodeGroupFilter} label="分组过滤" onChange={(e) => setNodeGroupFilter(e.target.value)}>
                        <MenuItem value="all">全部分组 ({allNodes.length})</MenuItem>
                        {groupOptions.map((g) => (
                          <MenuItem key={g} value={g}>
                            {g} ({groupNodeCounts[g] || 0})
                          </MenuItem>
                        ))}
                      </Select>
                    </FormControl>
                  </Grid>
                  <Grid item xs={3}>
                    <FormControl fullWidth size="small">
                      <InputLabel>来源过滤</InputLabel>
                      <Select value={nodeSourceFilter} label="来源过滤" onChange={(e) => setNodeSourceFilter(e.target.value)}>
                        <MenuItem value="all">全部来源</MenuItem>
                        {sourceOptions.map((s) => (
                          <MenuItem key={s} value={s}>
                            {s}
                          </MenuItem>
                        ))}
                      </Select>
                    </FormControl>
                  </Grid>
                  <Grid item xs={3}>
                    <Autocomplete
                      multiple
                      size="small"
                      options={countryOptions}
                      value={nodeCountryFilter}
                      onChange={(e, newValue) => setNodeCountryFilter(newValue)}
                      getOptionLabel={(option) => formatCountry(option)}
                      renderInput={(params) => <TextField {...params} label="国家过滤" />}
                      renderOption={(props, option) => <li {...props}>{formatCountry(option)}</li>}
                      limitTags={2}
                    />
                  </Grid>
                  <Grid item xs={3}>
                    <TextField
                      fullWidth
                      size="small"
                      label="搜索节点"
                      value={nodeSearchQuery}
                      onChange={(e) => setNodeSearchQuery(e.target.value)}
                    />
                  </Grid>
                </Grid>

                {/* 移动端穿梭框 - 使用标签页 */}
                {matchDownMd ? (
                  <Box sx={{ mt: 2 }}>
                    <Tabs
                      value={mobileTab}
                      onChange={(e, v) => setMobileTab(v)}
                      variant="fullWidth"
                      sx={{
                        mb: 2,
                        '& .MuiTab-root': {
                          fontWeight: 600,
                          borderRadius: 2,
                          mx: 0.5,
                          transition: 'all 0.2s'
                        },
                        '& .Mui-selected': {
                          bgcolor: 'primary.light',
                          color: 'primary.contrastText'
                        }
                      }}
                    >
                      <Tab label={`可选节点 (${availableNodes.length})`} icon={<ChevronRightIcon />} iconPosition="end" />
                      <Tab label={`已选节点 (${formData.selectedNodes.length})`} icon={<ChevronLeftIcon />} iconPosition="start" />
                    </Tabs>

                    {/* 可选节点面板 */}
                    <Fade in={mobileTab === 0}>
                      <Box sx={{ display: mobileTab === 0 ? 'block' : 'none' }}>
                        <Paper
                          elevation={0}
                          sx={{
                            p: 2,
                            maxHeight: 350,
                            overflow: 'auto',
                            background: `linear-gradient(145deg, ${theme.palette.mode === 'dark' ? '#1a1a2e' : '#f8f9fa'} 0%, ${theme.palette.mode === 'dark' ? '#16213e' : '#ffffff'} 100%)`,
                            border: '1px solid',
                            borderColor: 'divider',
                            borderRadius: 3
                          }}
                        >
                          <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 1.5 }}>
                            <FormControlLabel
                              control={
                                <Checkbox
                                  checked={checkedAvailable.length === availableNodes.length && availableNodes.length > 0}
                                  indeterminate={checkedAvailable.length > 0 && checkedAvailable.length < availableNodes.length}
                                  onChange={handleToggleAllAvailable}
                                  size="small"
                                />
                              }
                              label={
                                <Typography variant="body2" fontWeight={600}>
                                  全选
                                </Typography>
                              }
                            />
                            <Chip
                              label={availableNodes.length > 100 ? `显示前100/${availableNodes.length}` : `${availableNodes.length}个`}
                              size="small"
                              color="primary"
                              variant="outlined"
                            />
                          </Stack>
                          <List dense sx={{ pt: 0 }}>
                            {availableNodes.slice(0, 100).map((node) => (
                              <ListItem
                                key={node.ID}
                                sx={{
                                  py: 0.75,
                                  px: 1,
                                  mb: 0.5,
                                  borderRadius: 2,
                                  bgcolor: checkedAvailable.includes(node.Name) ? 'action.selected' : 'transparent',
                                  border: '1px solid',
                                  borderColor: checkedAvailable.includes(node.Name) ? 'primary.main' : 'transparent',
                                  transition: 'all 0.15s ease-in-out',
                                  '&:hover': {
                                    bgcolor: 'action.hover',
                                    transform: 'translateX(4px)'
                                  }
                                }}
                                secondaryAction={
                                  <IconButton
                                    size="small"
                                    color="primary"
                                    onClick={() => handleAddNode(node.Name)}
                                    sx={{
                                      bgcolor: 'primary.main',
                                      color: 'white',
                                      '&:hover': { bgcolor: 'primary.dark' }
                                    }}
                                  >
                                    <AddIcon fontSize="small" />
                                  </IconButton>
                                }
                              >
                                <ListItemIcon sx={{ minWidth: 36 }}>
                                  <Checkbox
                                    edge="start"
                                    checked={checkedAvailable.includes(node.Name)}
                                    onChange={() => handleToggleAvailable(node.Name)}
                                    size="small"
                                  />
                                </ListItemIcon>
                                <ListItemText
                                  primary={node.Name}
                                  secondary={
                                    <Chip
                                      label={node.Group || '未分组'}
                                      size="small"
                                      variant="outlined"
                                      sx={{ mt: 0.5, height: 20, fontSize: '0.7rem' }}
                                    />
                                  }
                                  primaryTypographyProps={{
                                    noWrap: true,
                                    fontWeight: 500,
                                    sx: { maxWidth: 'calc(100% - 60px)' }
                                  }}
                                />
                              </ListItem>
                            ))}
                          </List>
                        </Paper>

                        {/* 移动端底部操作栏 */}
                        <Paper
                          elevation={3}
                          sx={{
                            mt: 2,
                            p: 1.5,
                            borderRadius: 2,
                            display: 'flex',
                            gap: 1,
                            justifyContent: 'center',
                            background: `linear-gradient(90deg, ${theme.palette.primary.main} 0%, ${theme.palette.primary.dark} 100%)`
                          }}
                        >
                          <Button
                            variant="contained"
                            color="inherit"
                            size="small"
                            startIcon={<AddIcon />}
                            onClick={handleAddChecked}
                            disabled={checkedAvailable.length === 0}
                            sx={{
                              bgcolor: 'white',
                              color: 'primary.dark',
                              fontWeight: 700,
                              boxShadow: '0 2px 4px rgba(0,0,0,0.2)',
                              '&:hover': { bgcolor: '#f5f5f5' },
                              '&:disabled': { bgcolor: '#e0e0e0', color: '#999' }
                            }}
                          >
                            添加选中 ({checkedAvailable.length})
                          </Button>
                          <Button
                            variant="outlined"
                            size="small"
                            onClick={handleAddAllVisible}
                            sx={{
                              borderColor: 'white',
                              borderWidth: 2,
                              color: 'white',
                              fontWeight: 700,
                              textShadow: '0 1px 2px rgba(0,0,0,0.3)',
                              '&:hover': { bgcolor: 'rgba(255,255,255,0.15)', borderColor: 'white' }
                            }}
                          >
                            全部添加
                          </Button>
                        </Paper>
                      </Box>
                    </Fade>

                    {/* 已选节点面板 */}
                    <Fade in={mobileTab === 1}>
                      <Box sx={{ display: mobileTab === 1 ? 'block' : 'none' }}>
                        <TextField
                          fullWidth
                          size="small"
                          placeholder="搜索已选节点..."
                          value={selectedNodeSearch}
                          onChange={(e) => setSelectedNodeSearch(e.target.value)}
                          InputProps={{
                            startAdornment: (
                              <InputAdornment position="start">
                                <SearchIcon color="action" />
                              </InputAdornment>
                            )
                          }}
                          sx={{ mb: 2 }}
                        />
                        <Paper
                          elevation={0}
                          sx={{
                            p: 2,
                            maxHeight: 350,
                            overflow: 'auto',
                            background: `linear-gradient(145deg, ${theme.palette.mode === 'dark' ? '#1e3a2f' : '#f0fff4'} 0%, ${theme.palette.mode === 'dark' ? '#1a3330' : '#ffffff'} 100%)`,
                            border: '1px solid',
                            borderColor: 'success.light',
                            borderRadius: 3
                          }}
                        >
                          <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 1.5 }}>
                            <FormControlLabel
                              control={
                                <Checkbox
                                  checked={checkedSelected.length === filteredSelectedNodes.length && filteredSelectedNodes.length > 0}
                                  indeterminate={checkedSelected.length > 0 && checkedSelected.length < filteredSelectedNodes.length}
                                  onChange={handleToggleAllSelected}
                                  size="small"
                                  color="success"
                                />
                              }
                              label={
                                <Typography variant="body2" fontWeight={600}>
                                  全选
                                </Typography>
                              }
                            />
                            <Chip label={`${formData.selectedNodes.length}个已选`} size="small" color="success" />
                          </Stack>
                          <List dense sx={{ pt: 0 }}>
                            {filteredSelectedNodes.map((node) => (
                              <ListItem
                                key={node.ID}
                                sx={{
                                  py: 0.75,
                                  px: 1,
                                  mb: 0.5,
                                  borderRadius: 2,
                                  bgcolor: checkedSelected.includes(node.Name) ? 'error.lighter' : 'transparent',
                                  border: '1px solid',
                                  borderColor: checkedSelected.includes(node.Name) ? 'error.main' : 'transparent',
                                  transition: 'all 0.15s ease-in-out',
                                  '&:hover': {
                                    bgcolor: 'action.hover',
                                    transform: 'translateX(-4px)'
                                  }
                                }}
                                secondaryAction={
                                  <IconButton
                                    size="small"
                                    color="error"
                                    onClick={() => handleRemoveNode(node.Name)}
                                    sx={{
                                      bgcolor: 'error.main',
                                      color: 'white',
                                      '&:hover': { bgcolor: 'error.dark' }
                                    }}
                                  >
                                    <DeleteIcon fontSize="small" />
                                  </IconButton>
                                }
                              >
                                <ListItemIcon sx={{ minWidth: 36 }}>
                                  <Checkbox
                                    edge="start"
                                    checked={checkedSelected.includes(node.Name)}
                                    onChange={() => handleToggleSelected(node.Name)}
                                    size="small"
                                    color="error"
                                  />
                                </ListItemIcon>
                                <ListItemText
                                  primary={node.Name}
                                  secondary={
                                    <Chip
                                      label={node.Group || '未分组'}
                                      size="small"
                                      color="success"
                                      variant="outlined"
                                      sx={{ mt: 0.5, height: 20, fontSize: '0.7rem' }}
                                    />
                                  }
                                  primaryTypographyProps={{
                                    noWrap: true,
                                    fontWeight: 500,
                                    sx: { maxWidth: 'calc(100% - 60px)' }
                                  }}
                                />
                              </ListItem>
                            ))}
                          </List>
                          {filteredSelectedNodes.length === 0 && (
                            <Typography color="textSecondary" align="center" sx={{ py: 4 }}>
                              {selectedNodeSearch ? '未找到匹配的节点' : '暂无已选节点'}
                            </Typography>
                          )}
                        </Paper>

                        {/* 移动端底部操作栏 */}
                        <Paper
                          elevation={3}
                          sx={{
                            mt: 2,
                            p: 1.5,
                            borderRadius: 2,
                            display: 'flex',
                            gap: 1,
                            justifyContent: 'center',
                            background: `linear-gradient(90deg, ${theme.palette.error.main} 0%, ${theme.palette.error.dark} 100%)`
                          }}
                        >
                          <Button
                            variant="contained"
                            color="inherit"
                            size="small"
                            startIcon={<DeleteIcon />}
                            onClick={handleRemoveChecked}
                            disabled={checkedSelected.length === 0}
                            sx={{
                              bgcolor: 'white',
                              color: 'error.dark',
                              fontWeight: 700,
                              boxShadow: '0 2px 4px rgba(0,0,0,0.2)',
                              '&:hover': { bgcolor: '#f5f5f5' },
                              '&:disabled': { bgcolor: '#e0e0e0', color: '#999' }
                            }}
                          >
                            移除选中 ({checkedSelected.length})
                          </Button>
                          <Button
                            variant="outlined"
                            size="small"
                            onClick={handleRemoveAll}
                            sx={{
                              borderColor: 'white',
                              borderWidth: 2,
                              color: 'white',
                              fontWeight: 700,
                              textShadow: '0 1px 2px rgba(0,0,0,0.3)',
                              '&:hover': { bgcolor: 'rgba(255,255,255,0.15)', borderColor: 'white' }
                            }}
                          >
                            全部移除
                          </Button>
                        </Paper>
                      </Box>
                    </Fade>
                  </Box>
                ) : (
                  /* 桌面端穿梭框 */
                  <Grid container spacing={2} sx={{ mt: 1 }}>
                    {/* 可选节点 */}
                    <Grid item xs={5}>
                      <Paper
                        elevation={0}
                        sx={{
                          p: 2,
                          height: 380,
                          overflow: 'hidden',
                          display: 'flex',
                          flexDirection: 'column',
                          background: `linear-gradient(145deg, ${theme.palette.mode === 'dark' ? '#1a1a2e' : '#f8f9fa'} 0%, ${theme.palette.mode === 'dark' ? '#16213e' : '#ffffff'} 100%)`,
                          border: '2px solid',
                          borderColor: 'primary.light',
                          borderRadius: 3,
                          transition: 'all 0.3s ease',
                          '&:hover': {
                            borderColor: 'primary.main',
                            boxShadow: `0 4px 20px ${theme.palette.primary.main}20`
                          }
                        }}
                      >
                        <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 1.5, flexShrink: 0 }}>
                          <Stack direction="row" alignItems="center" spacing={1}>
                            <FormControlLabel
                              control={
                                <Checkbox
                                  checked={checkedAvailable.length === availableNodes.length && availableNodes.length > 0}
                                  indeterminate={checkedAvailable.length > 0 && checkedAvailable.length < availableNodes.length}
                                  onChange={handleToggleAllAvailable}
                                  size="small"
                                />
                              }
                              label=""
                              sx={{ mr: 0 }}
                            />
                            <Typography variant="subtitle1" fontWeight={700} color="primary">
                              可选节点
                            </Typography>
                          </Stack>
                          <Chip
                            label={availableNodes.length > 100 ? `前100/${availableNodes.length}` : availableNodes.length}
                            size="small"
                            color="primary"
                          />
                        </Stack>
                        <Box sx={{ flexGrow: 1, overflow: 'auto', pr: 1 }}>
                          <List dense>
                            {availableNodes.slice(0, 100).map((node) => (
                              <ListItem
                                key={node.ID}
                                sx={{
                                  py: 0.5,
                                  px: 1,
                                  mb: 0.5,
                                  borderRadius: 2,
                                  cursor: 'pointer',
                                  bgcolor: checkedAvailable.includes(node.Name) ? 'primary.lighter' : 'transparent',
                                  border: '1px solid',
                                  borderColor: checkedAvailable.includes(node.Name) ? 'primary.main' : 'divider',
                                  transition: 'all 0.15s ease-in-out',
                                  '&:hover': {
                                    bgcolor: 'action.hover',
                                    transform: 'translateX(4px)',
                                    borderColor: 'primary.light'
                                  }
                                }}
                                onClick={() => handleToggleAvailable(node.Name)}
                                onDoubleClick={() => handleAddNode(node.Name)}
                              >
                                <ListItemIcon sx={{ minWidth: 32 }}>
                                  <Checkbox
                                    edge="start"
                                    checked={checkedAvailable.includes(node.Name)}
                                    tabIndex={-1}
                                    disableRipple
                                    size="small"
                                  />
                                </ListItemIcon>
                                <ListItemText
                                  primary={node.Name}
                                  secondary={node.Group}
                                  primaryTypographyProps={{
                                    noWrap: true,
                                    fontSize: '0.875rem',
                                    fontWeight: 500
                                  }}
                                  secondaryTypographyProps={{
                                    noWrap: true,
                                    fontSize: '0.75rem'
                                  }}
                                />
                              </ListItem>
                            ))}
                            {availableNodes.length > 100 && (
                              <Typography variant="caption" color="textSecondary" sx={{ display: 'block', textAlign: 'center', py: 1 }}>
                                还有 {availableNodes.length - 100} 个节点未显示
                              </Typography>
                            )}
                          </List>
                        </Box>
                      </Paper>
                    </Grid>

                    {/* 中间操作按钮 */}
                    <Grid
                      item
                      xs={2}
                      sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: 1 }}
                    >
                      <Button
                        variant="contained"
                        size="small"
                        onClick={handleAddChecked}
                        disabled={checkedAvailable.length === 0}
                        endIcon={<ChevronRightIcon />}
                        sx={{
                          minWidth: 120,
                          background: 'linear-gradient(45deg, #2196F3 30%, #21CBF3 90%)',
                          boxShadow: '0 3px 5px 2px rgba(33, 150, 243, .3)',
                          color: '#fff',
                          fontWeight: 700,
                          textShadow: '0 1px 2px rgba(0,0,0,0.2)',
                          '&:disabled': {
                            background: '#ccc',
                            color: '#888'
                          }
                        }}
                      >
                        添加 ({checkedAvailable.length})
                      </Button>
                      <Button
                        variant="outlined"
                        size="small"
                        onClick={handleAddAllVisible}
                        endIcon={<ChevronRightIcon />}
                        sx={{ minWidth: 120, fontWeight: 600 }}
                      >
                        全部添加
                      </Button>
                      <Divider sx={{ width: '60%', my: 1 }} />
                      <Button
                        variant="outlined"
                        size="small"
                        color="error"
                        onClick={handleRemoveAll}
                        startIcon={<ChevronLeftIcon />}
                        sx={{ minWidth: 120, fontWeight: 600 }}
                      >
                        全部移除
                      </Button>
                      <Button
                        variant="contained"
                        size="small"
                        color="error"
                        onClick={handleRemoveChecked}
                        disabled={checkedSelected.length === 0}
                        startIcon={<ChevronLeftIcon />}
                        sx={{
                          minWidth: 120,
                          background: 'linear-gradient(45deg, #FE6B8B 30%, #FF8E53 90%)',
                          boxShadow: '0 3px 5px 2px rgba(254, 107, 139, .3)',
                          color: '#fff',
                          fontWeight: 700,
                          textShadow: '0 1px 2px rgba(0,0,0,0.2)',
                          '&:disabled': {
                            background: '#ccc',
                            color: '#888'
                          }
                        }}
                      >
                        移除 ({checkedSelected.length})
                      </Button>
                    </Grid>

                    {/* 已选节点 */}
                    <Grid item xs={5}>
                      <Paper
                        elevation={0}
                        sx={{
                          p: 2,
                          height: 380,
                          overflow: 'hidden',
                          display: 'flex',
                          flexDirection: 'column',
                          background: `linear-gradient(145deg, ${theme.palette.mode === 'dark' ? '#1e3a2f' : '#f0fff4'} 0%, ${theme.palette.mode === 'dark' ? '#1a3330' : '#ffffff'} 100%)`,
                          border: '2px solid',
                          borderColor: 'success.light',
                          borderRadius: 3,
                          transition: 'all 0.3s ease',
                          '&:hover': {
                            borderColor: 'success.main',
                            boxShadow: `0 4px 20px ${theme.palette.success.main}20`
                          }
                        }}
                      >
                        <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 1, flexShrink: 0 }}>
                          <Stack direction="row" alignItems="center" spacing={1}>
                            <FormControlLabel
                              control={
                                <Checkbox
                                  checked={checkedSelected.length === filteredSelectedNodes.length && filteredSelectedNodes.length > 0}
                                  indeterminate={checkedSelected.length > 0 && checkedSelected.length < filteredSelectedNodes.length}
                                  onChange={handleToggleAllSelected}
                                  size="small"
                                  color="success"
                                />
                              }
                              label=""
                              sx={{ mr: 0 }}
                            />
                            <Typography variant="subtitle1" fontWeight={700} color="success.main">
                              已选节点
                            </Typography>
                          </Stack>
                          <Chip label={formData.selectedNodes.length} size="small" color="success" />
                        </Stack>
                        <TextField
                          fullWidth
                          size="small"
                          placeholder="搜索已选节点..."
                          value={selectedNodeSearch}
                          onChange={(e) => setSelectedNodeSearch(e.target.value)}
                          InputProps={{
                            startAdornment: (
                              <InputAdornment position="start">
                                <SearchIcon fontSize="small" color="action" />
                              </InputAdornment>
                            )
                          }}
                          sx={{ mb: 1, flexShrink: 0, '& .MuiOutlinedInput-root': { borderRadius: 2 } }}
                        />
                        <Box sx={{ flexGrow: 1, overflow: 'auto', pr: 1 }}>
                          <List dense>
                            {filteredSelectedNodes.map((node) => (
                              <ListItem
                                key={node.ID}
                                sx={{
                                  py: 0.5,
                                  px: 1,
                                  mb: 0.5,
                                  borderRadius: 2,
                                  cursor: 'pointer',
                                  bgcolor: checkedSelected.includes(node.Name) ? 'error.lighter' : 'transparent',
                                  border: '1px solid',
                                  borderColor: checkedSelected.includes(node.Name) ? 'error.main' : 'divider',
                                  transition: 'all 0.15s ease-in-out',
                                  '&:hover': {
                                    bgcolor: 'action.hover',
                                    transform: 'translateX(-4px)',
                                    borderColor: 'error.light'
                                  }
                                }}
                                onClick={() => handleToggleSelected(node.Name)}
                                onDoubleClick={() => handleRemoveNode(node.Name)}
                              >
                                <ListItemIcon sx={{ minWidth: 32 }}>
                                  <Checkbox
                                    edge="start"
                                    checked={checkedSelected.includes(node.Name)}
                                    tabIndex={-1}
                                    disableRipple
                                    size="small"
                                    color="error"
                                  />
                                </ListItemIcon>
                                <ListItemText
                                  primary={node.Name}
                                  secondary={node.Group}
                                  primaryTypographyProps={{
                                    noWrap: true,
                                    fontSize: '0.875rem',
                                    fontWeight: 500
                                  }}
                                  secondaryTypographyProps={{
                                    noWrap: true,
                                    fontSize: '0.75rem'
                                  }}
                                />
                              </ListItem>
                            ))}
                          </List>
                          {filteredSelectedNodes.length === 0 && (
                            <Typography color="textSecondary" align="center" sx={{ py: 4 }}>
                              {selectedNodeSearch ? '未找到匹配的节点' : '暂无已选节点'}
                            </Typography>
                          )}
                        </Box>
                      </Paper>
                    </Grid>
                  </Grid>
                )}
              </>
            )}

            <Divider />

            {/* 延迟和速度过滤 */}
            <Grid container spacing={2}>
              <Grid item xs={6}>
                <TextField
                  fullWidth
                  label="最大延迟"
                  type="number"
                  value={formData.DelayTime}
                  onChange={(e) => setFormData({ ...formData, DelayTime: Number(e.target.value) })}
                  InputProps={{ endAdornment: <InputAdornment position="end">ms</InputAdornment> }}
                  helperText="设置筛选节点的延迟阈值，0表示不限制"
                />
              </Grid>
              <Grid item xs={6}>
                <TextField
                  fullWidth
                  label="最小速度"
                  type="number"
                  value={formData.MinSpeed}
                  onChange={(e) => setFormData({ ...formData, MinSpeed: Number(e.target.value) })}
                  InputProps={{ endAdornment: <InputAdornment position="end">MB/s</InputAdornment> }}
                  helperText="设置筛选节点的最小下载速度，0表示不限制"
                />
              </Grid>
            </Grid>

            {/* 落地IP国家过滤 */}
            <Grid container spacing={2}>
              <Grid item xs={6}>
                <Autocomplete
                  multiple
                  options={countryOptions}
                  value={formData.CountryWhitelist}
                  onChange={(e, newValue) => setFormData({ ...formData, CountryWhitelist: newValue })}
                  getOptionLabel={(option) => formatCountry(option)}
                  renderInput={(params) => (
                    <TextField {...params} label="落地IP国家白名单" helperText="只保留这些国家的节点，不选则不限制" />
                  )}
                  renderOption={(props, option) => <li {...props}>{formatCountry(option)}</li>}
                />
              </Grid>
              <Grid item xs={6}>
                <Autocomplete
                  multiple
                  options={countryOptions}
                  value={formData.CountryBlacklist}
                  onChange={(e, newValue) => setFormData({ ...formData, CountryBlacklist: newValue })}
                  getOptionLabel={(option) => formatCountry(option)}
                  renderInput={(params) => (
                    <TextField {...params} label="落地IP国家黑名单" helperText="排除这些国家的节点（优先级高于白名单）" />
                  )}
                  renderOption={(props, option) => <li {...props}>{formatCountry(option)}</li>}
                />
              </Grid>
            </Grid>

            {/* 脚本选择 */}
            <Autocomplete
              multiple
              options={scripts}
              getOptionLabel={(option) => `${option.name} (${option.version})`}
              value={scripts.filter((s) => formData.selectedScripts.includes(s.id))}
              onChange={(e, newValue) => setFormData({ ...formData, selectedScripts: newValue.map((s) => s.id) })}
              renderInput={(params) => (
                <TextField {...params} label="数据处理脚本" helperText="脚本将在查询到节点数据后运行，多个脚本按顺序执行" />
              )}
              renderOption={(props, option) => (
                <li {...props}>
                  <Box sx={{ display: 'flex', flexDirection: 'column' }}>
                    <Typography variant="body1">{option.name}</Typography>
                    <Typography variant="caption" color="textSecondary">
                      版本: {option.version}
                    </Typography>
                  </Box>
                </li>
              )}
            />

            <Divider />

            {/* 节点命名规则 */}
            <Box>
              <Typography variant="subtitle1" fontWeight="bold" gutterBottom>
                节点命名规则
              </Typography>
              <TextField
                fullWidth
                label="命名规则模板"
                value={formData.nodeNameRule}
                onChange={(e) => setFormData({ ...formData, nodeNameRule: e.target.value })}
                placeholder="例如: $LinkCountry - $LinkName ($Speed)"
                helperText="留空则使用原始名称，仅在访问订阅链接时生效"
              />
              <Box sx={{ mt: 1, p: 1.5, bgcolor: 'action.hover', borderRadius: 1 }}>
                <Typography variant="caption" color="textSecondary" component="div">
                  <strong>可用变量：</strong>
                  <br />• <code>$Name</code> - 系统备注名称 &nbsp;&nbsp; • <code>$LinkName</code> - 原始节点名称
                  <br />• <code>$LinkCountry</code> - 落地IP国家代码 &nbsp;&nbsp; • <code>$Speed</code> - 下载速度
                  <br />• <code>$Delay</code> - 延迟 &nbsp;&nbsp; • <code>$Group</code> - 分组名称
                  <br />• <code>$Source</code> - 来源 &nbsp;&nbsp; • <code>$Index</code> - 序号 &nbsp;&nbsp; • <code>$Protocol</code> -
                  协议类型
                </Typography>
              </Box>
              {formData.nodeNameRule && (
                <Alert severity="info" sx={{ mt: 1 }}>
                  <Typography variant="body2">
                    <strong>预览：</strong> {previewNodeName(formData.nodeNameRule)}
                  </Typography>
                </Alert>
              )}
            </Box>

            <Divider />

            {/* IP 白名单/黑名单 */}
            <TextField
              fullWidth
              label="IP 黑名单（优先级高于白名单），不允许指定IP访问订阅链接"
              multiline
              rows={2}
              value={formData.IPBlacklist}
              onChange={(e) => setFormData({ ...formData, IPBlacklist: e.target.value })}
              helperText="每行一个 IP 或 CIDR"
            />
            <TextField
              fullWidth
              label="IP 白名单，只允许指定IP访问订阅链接"
              multiline
              rows={2}
              value={formData.IPWhitelist}
              onChange={(e) => setFormData({ ...formData, IPWhitelist: e.target.value })}
              helperText="每行一个 IP 或 CIDR"
            />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDialogOpen(false)}>关闭</Button>
          <Button variant="contained" onClick={handleSubmit}>
            确定
          </Button>
        </DialogActions>
      </Dialog>

      {/* 客户端对话框 */}
      <Dialog open={clientDialogOpen} onClose={() => setClientDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>客户端（点击二维码获取地址）</DialogTitle>
        <DialogContent>
          <Stack spacing={2}>
            {Object.entries(clientUrls).map(([name, url]) => (
              <Stack key={name} direction="row" alignItems="center" spacing={2}>
                <Chip label={name} color="success" sx={{ minWidth: 100 }} />
                <Button variant="outlined" onClick={() => handleQrcode(name === '自动识别' ? url : `${url}&client=${name}`, name)}>
                  二维码
                </Button>
                <IconButton size="small" onClick={() => copyToClipboard(name === '自动识别' ? url : `${url}&client=${name}`)}>
                  <ContentCopyIcon fontSize="small" />
                </IconButton>
              </Stack>
            ))}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setClientDialogOpen(false)}>关闭</Button>
        </DialogActions>
      </Dialog>

      {/* QR码对话框 */}
      <Dialog open={qrDialogOpen} onClose={() => setQrDialogOpen(false)}>
        <DialogTitle>{qrTitle}</DialogTitle>
        <DialogContent sx={{ textAlign: 'center', pt: 2 }}>
          <QRCodeSVG value={qrUrl} size={200} />
          <TextField fullWidth value={qrUrl} sx={{ mt: 2 }} size="small" InputProps={{ readOnly: true }} />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => copyToClipboard(qrUrl)}>复制</Button>
          <Button onClick={() => window.open(qrUrl)}>打开</Button>
          <Button onClick={() => setQrDialogOpen(false)}>关闭</Button>
        </DialogActions>
      </Dialog>

      {/* 访问记录对话框 */}
      <Dialog open={logsDialogOpen} onClose={() => setLogsDialogOpen(false)} maxWidth="md" fullWidth>
        <DialogTitle>访问记录</DialogTitle>
        <DialogContent>
          <TableContainer>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>IP</TableCell>
                  <TableCell>来源</TableCell>
                  <TableCell>总访问次数</TableCell>
                  <TableCell>最近时间</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {currentLogs.map((log) => (
                  <TableRow key={log.ID}>
                    <TableCell>{log.IP}</TableCell>
                    <TableCell>{log.Addr || '-'}</TableCell>
                    <TableCell>{log.Count}</TableCell>
                    <TableCell>{log.Date}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
          {currentLogs.length === 0 && (
            <Typography color="textSecondary" align="center" sx={{ py: 4 }}>
              暂无访问记录
            </Typography>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setLogsDialogOpen(false)}>关闭</Button>
        </DialogActions>
      </Dialog>

      {/* 提示消息 */}
      <Snackbar
        open={snackbar.open}
        autoHideDuration={3000}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
      >
        <Alert severity={snackbar.severity}>{snackbar.message}</Alert>
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
    </MainCard>
  );
}
