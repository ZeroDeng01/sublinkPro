import { useState, useEffect, useMemo, useCallback } from 'react';

// material-ui
import { useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import Stack from '@mui/material/Stack';
import Alert from '@mui/material/Alert';
import Snackbar from '@mui/material/Snackbar';
import Tooltip from '@mui/material/Tooltip';

// icons
import AddIcon from '@mui/icons-material/Add';
import RefreshIcon from '@mui/icons-material/Refresh';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import Pagination from 'components/Pagination';
import {
  getSubscriptions,
  addSubscription,
  updateSubscription,
  deleteSubscription,
  sortSubscription,
  batchSortSubscription,
  copySubscription,
  previewSubscriptionNodes
} from 'api/subscriptions';
import { getNodes, getNodeCountries, getNodeGroups, getNodeSources, getNodeProtocols } from 'api/nodes';
import { getTemplates } from 'api/templates';
import { getScripts } from 'api/scripts';
import { getTags } from 'api/tags';

// components
import {
  ConfirmDialog,
  QrCodeDialog,
  ClientUrlsDialog,
  AccessLogsDialog,
  SubscriptionMobileCard,
  SubscriptionTable,
  SubscriptionFormDialog,
  NodePreviewDialog,
  ShareManageDialog,
  ChainProxyDialog
} from './component';

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
    replaceServerWithHost: false,
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
    nodeNameRule: '',
    nodeNamePreprocess: '',
    nodeNameWhitelist: '',
    nodeNameBlacklist: '',
    tagWhitelist: '',
    tagBlacklist: '',
    protocolWhitelist: '',
    protocolBlacklist: '',
    protocolOptions: [],
    deduplicationRule: '',
    refreshUsageOnRequest: true // 默认开启实时获取用量信息
  });

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
  const [clientUrls] = useState({});

  // 访问记录对话框
  const [logsDialogOpen, setLogsDialogOpen] = useState(false);
  const [currentLogs, setCurrentLogs] = useState([]);

  // 排序模式
  const [sortingSubId, setSortingSubId] = useState(null);
  const [tempSortData, setTempSortData] = useState([]);
  const [selectedSortItems, setSelectedSortItems] = useState([]); // 多选排序项

  // 展开行
  const [expandedRows, setExpandedRows] = useState({});

  // 穿梭框状态
  const [checkedAvailable, setCheckedAvailable] = useState([]);
  const [checkedSelected, setCheckedSelected] = useState([]);
  const [mobileTab, setMobileTab] = useState(0);
  const [selectedNodeSearch, setSelectedNodeSearch] = useState('');
  const [namingMode, setNamingMode] = useState('builder');

  // 预览状态
  const [previewOpen, setPreviewOpen] = useState(false);
  const [previewLoading, setPreviewLoading] = useState(false);
  const [previewData, setPreviewData] = useState(null);

  // 分页
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(() => {
    const saved = localStorage.getItem('subscriptions_rowsPerPage');
    return saved ? parseInt(saved, 10) : 10;
  });
  const [totalItems, setTotalItems] = useState(0);

  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });

  // 从后端获取的分组和来源选项
  const [groupOptions, setGroupOptions] = useState([]);
  const [sourceOptions, setSourceOptions] = useState([]);
  const [tagOptions, setTagOptions] = useState([]);
  const [protocolOptions, setProtocolOptions] = useState([]);

  // 获取订阅列表（分页）
  const fetchSubscriptions = async (currentPage, currentPageSize) => {
    setLoading(true);
    try {
      const response = await getSubscriptions({ page: currentPage + 1, pageSize: currentPageSize });
      // 处理分页响应
      if (response.data && response.data.items !== undefined) {
        setSubscriptions(response.data.items || []);
        setTotalItems(response.data.total || 0);
      } else {
        // 向后兼容：老格式直接返回数组
        setSubscriptions(response.data || []);
        setTotalItems((response.data || []).length);
      }
    } catch (error) {
      console.error(error);
      showMessage(error.message || '获取订阅列表失败', 'error');
    } finally {
      setLoading(false);
    }
  };

  // 获取其他数据（不分页）
  const fetchOtherData = useCallback(async () => {
    try {
      const [nodesRes, templatesRes, scriptsRes, countriesRes, groupsRes, sourcesRes, tagsRes, protocolsRes] = await Promise.all([
        getNodes(),
        getTemplates(),
        getScripts(),
        getNodeCountries(),
        getNodeGroups(),
        getNodeSources(),
        getTags(),
        getNodeProtocols()
      ]);
      setAllNodes(nodesRes.data || []);
      setTemplates(templatesRes.data || []);
      setScripts(scriptsRes.data || []);
      setCountryOptions(countriesRes.data || []);
      setGroupOptions((groupsRes.data || []).sort());
      setSourceOptions((sourcesRes.data || []).sort());
      setTagOptions(tagsRes.data || []);
      setProtocolOptions(protocolsRes.data || []);
    } catch (error) {
      console.error(error);
    }
  }, []);

  // 初始加载
  useEffect(() => {
    fetchSubscriptions(0, rowsPerPage);
    fetchOtherData();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const showMessage = (message, severity = 'success') => {
    setSnackbar({ open: true, message, severity });
  };

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
    showMessage('已复制到剪贴板');
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
      replaceServerWithHost: false,
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
      nodeNameRule: '',
      nodeNamePreprocess: '',
      nodeNameWhitelist: '',
      nodeNameBlacklist: '',
      tagWhitelist: '',
      tagBlacklist: '',
      protocolWhitelist: '',
      protocolBlacklist: '',
      protocolOptions: protocolOptions,
      deduplicationRule: '',
      refreshUsageOnRequest: true
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

    const nodes = sub.Nodes?.map((n) => n.ID) || [];
    const groups = (sub.Groups || []).map((g) => (typeof g === 'string' ? g : g.Name));
    const scriptIds = (sub.Scripts || []).map((s) => s.id);

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
      replaceServerWithHost: config?.replaceServerWithHost || false,
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
      nodeNameRule: sub.NodeNameRule || '',
      nodeNamePreprocess: sub.NodeNamePreprocess || '',
      nodeNameWhitelist: sub.NodeNameWhitelist || '',
      nodeNameBlacklist: sub.NodeNameBlacklist || '',
      tagWhitelist: sub.TagWhitelist || '',
      tagBlacklist: sub.TagBlacklist || '',
      protocolWhitelist: sub.ProtocolWhitelist || '',
      protocolBlacklist: sub.ProtocolBlacklist || '',
      protocolOptions: protocolOptions,
      deduplicationRule: sub.DeduplicationRule || '',
      refreshUsageOnRequest: sub.RefreshUsageOnRequest !== false // 默认 true
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
        fetchSubscriptions(page, rowsPerPage);
      } catch (error) {
        console.error(error);
        showMessage(error.message || '删除失败', 'error');
      }
    });
  };

  // 复制订阅
  const handleCopy = async (sub) => {
    openConfirm('复制订阅', `确定要复制订阅 "${sub.Name}" 吗？`, async () => {
      try {
        await copySubscription(sub.ID);
        showMessage('复制成功');
        fetchSubscriptions(page, rowsPerPage);
      } catch (error) {
        console.error(error);
        showMessage(error.message || '复制失败', 'error');
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
        cert: formData.cert,
        replaceServerWithHost: formData.replaceServerWithHost
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
        NodeNameRule: formData.nodeNameRule,
        NodeNamePreprocess: formData.nodeNamePreprocess,
        NodeNameWhitelist: formData.nodeNameWhitelist,
        NodeNameBlacklist: formData.nodeNameBlacklist,
        TagWhitelist: formData.tagWhitelist,
        TagBlacklist: formData.tagBlacklist,
        ProtocolWhitelist: formData.protocolWhitelist,
        ProtocolBlacklist: formData.protocolBlacklist,
        DeduplicationRule: formData.deduplicationRule || '',
        RefreshUsageOnRequest: formData.refreshUsageOnRequest
      };

      if (formData.selectionMode === 'nodes') {
        requestData.nodeIds = formData.selectedNodes.join(',');
        requestData.groups = '';
      } else if (formData.selectionMode === 'groups') {
        requestData.nodeIds = '';
        requestData.groups = formData.selectedGroups.join(',');
      } else {
        requestData.nodeIds = formData.selectedNodes.join(',');
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
      fetchSubscriptions(page, rowsPerPage);
    } catch (error) {
      console.error(error);
      showMessage(error.message || (isEdit ? '更新失败' : '添加失败'), 'error');
    }
  };

  // 节点选择操作（使用 node.ID）
  const handleAddNode = (nodeId) => {
    setFormData({ ...formData, selectedNodes: [...formData.selectedNodes, nodeId] });
    setCheckedAvailable(checkedAvailable.filter((id) => id !== nodeId));
  };

  const handleRemoveNode = (nodeId) => {
    setFormData({ ...formData, selectedNodes: formData.selectedNodes.filter((id) => id !== nodeId) });
    setCheckedSelected(checkedSelected.filter((id) => id !== nodeId));
  };

  // 过滤后的节点
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
      if (nodeCountryFilter.length > 0) {
        if (!node.LinkCountry || !nodeCountryFilter.includes(node.LinkCountry)) {
          return false;
        }
      }
      return true;
    });
  }, [allNodes, nodeGroupFilter, nodeSourceFilter, nodeSearchQuery, nodeCountryFilter]);

  // 可选节点（使用 ID 过滤）
  const availableNodes = useMemo(() => {
    return filteredNodes.filter((node) => !formData.selectedNodes.includes(node.ID));
  }, [filteredNodes, formData.selectedNodes]);

  const handleAddAllVisible = () => {
    const newNodes = [...formData.selectedNodes, ...availableNodes.map((n) => n.ID)];
    setFormData({ ...formData, selectedNodes: newNodes });
    setCheckedAvailable([]);
  };

  const handleRemoveAll = () => {
    setFormData({ ...formData, selectedNodes: [] });
    setCheckedSelected([]);
  };

  // 多选操作（使用 node.ID）
  const handleToggleAvailable = (nodeId) => {
    if (checkedAvailable.includes(nodeId)) {
      setCheckedAvailable(checkedAvailable.filter((id) => id !== nodeId));
    } else {
      setCheckedAvailable([...checkedAvailable, nodeId]);
    }
  };

  const handleToggleSelected = (nodeId) => {
    if (checkedSelected.includes(nodeId)) {
      setCheckedSelected(checkedSelected.filter((id) => id !== nodeId));
    } else {
      setCheckedSelected([...checkedSelected, nodeId]);
    }
  };

  const handleAddChecked = () => {
    const newNodes = [...formData.selectedNodes, ...checkedAvailable];
    setFormData({ ...formData, selectedNodes: newNodes });
    setCheckedAvailable([]);
  };

  const handleRemoveChecked = () => {
    const newNodes = formData.selectedNodes.filter((id) => !checkedSelected.includes(id));
    setFormData({ ...formData, selectedNodes: newNodes });
    setCheckedSelected([]);
  };

  const handleToggleAllAvailable = () => {
    if (checkedAvailable.length === availableNodes.length) {
      setCheckedAvailable([]);
    } else {
      setCheckedAvailable(availableNodes.map((n) => n.ID));
    }
  };

  // 已选节点列表（使用 ID 过滤）
  const selectedNodesList = useMemo(() => {
    return allNodes.filter((node) => formData.selectedNodes.includes(node.ID));
  }, [allNodes, formData.selectedNodes]);

  const handleToggleAllSelected = () => {
    if (checkedSelected.length === selectedNodesList.length) {
      setCheckedSelected([]);
    } else {
      setCheckedSelected(selectedNodesList.map((n) => n.ID));
    }
  };

  // 预览节点
  const handlePreview = async () => {
    setPreviewLoading(true);
    try {
      // 构建预览请求数据
      const previewRequest = {
        Nodes: formData.selectionMode !== 'groups' ? formData.selectedNodes : [],
        Groups: formData.selectionMode !== 'nodes' ? formData.selectedGroups : [],
        Scripts: formData.selectedScripts || [],
        DelayTime: formData.DelayTime || 0,
        MinSpeed: formData.MinSpeed || 0,
        CountryWhitelist: formData.CountryWhitelist.join(','),
        CountryBlacklist: formData.CountryBlacklist.join(','),
        TagWhitelist: formData.tagWhitelist || '',
        TagBlacklist: formData.tagBlacklist || '',
        ProtocolWhitelist: formData.protocolWhitelist || '',
        ProtocolBlacklist: formData.protocolBlacklist || '',
        NodeNameWhitelist: formData.nodeNameWhitelist || '',
        NodeNameBlacklist: formData.nodeNameBlacklist || '',
        NodeNamePreprocess: formData.nodeNamePreprocess || '',
        NodeNameRule: formData.nodeNameRule || '',
        DeduplicationRule: formData.deduplicationRule || ''
      };

      const response = await previewSubscriptionNodes(previewRequest);
      // 成功（code === 200 时返回，否则被拦截器 reject）
      setPreviewData(response.data);
      setPreviewOpen(true);
    } catch (error) {
      console.error(error);
      showMessage(error.message || '预览请求失败', 'error');
    } finally {
      setPreviewLoading(false);
    }
  };

  // 预览已保存的订阅（从列表触发）
  const handlePreviewSubscription = async (sub) => {
    setPreviewLoading(true);
    try {
      // 解析订阅配置
      const nodes = sub.Nodes?.map((n) => n.Name) || [];
      const groups = (sub.Groups || []).map((g) => (typeof g === 'string' ? g : g.Name));

      const previewRequest = {
        Nodes: nodes,
        Groups: groups,
        Scripts: (sub.Scripts || []).map((s) => s.id),
        DelayTime: sub.DelayTime || 0,
        MinSpeed: sub.MinSpeed || 0,
        CountryWhitelist: sub.CountryWhitelist || '',
        CountryBlacklist: sub.CountryBlacklist || '',
        TagWhitelist: sub.TagWhitelist || '',
        TagBlacklist: sub.TagBlacklist || '',
        ProtocolWhitelist: sub.ProtocolWhitelist || '',
        ProtocolBlacklist: sub.ProtocolBlacklist || '',
        NodeNameWhitelist: sub.NodeNameWhitelist || '',
        NodeNameBlacklist: sub.NodeNameBlacklist || '',
        NodeNamePreprocess: sub.NodeNamePreprocess || '',
        NodeNameRule: sub.NodeNameRule || '',
        DeduplicationRule: sub.DeduplicationRule || ''
      };

      const response = await previewSubscriptionNodes(previewRequest);
      // 成功（code === 200 时返回，否则被拦截器 reject）
      setPreviewData(response.data);
      setPreviewOpen(true);
    } catch (error) {
      console.error(error);
      showMessage(error.message || '预览请求失败', 'error');
    } finally {
      setPreviewLoading(false);
    }
  };

  // === 分享管理 ===
  const [shareDialogOpen, setShareDialogOpen] = useState(false);
  const [shareDialogSub, setShareDialogSub] = useState(null);

  // === 链式代理 ===
  const [chainProxyOpen, setChainProxyOpen] = useState(false);
  const [chainProxySub, setChainProxySub] = useState(null);

  const handleChainProxy = (sub) => {
    setChainProxySub(sub);
    setChainProxyOpen(true);
  };

  const handleClient = (sub) => {
    setShareDialogSub(sub);
    setShareDialogOpen(true);
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
    setSelectedSortItems([]); // 重置多选
    const sortData = [];
    (sub.Nodes || []).forEach((node, idx) => {
      sortData.push({
        ID: node.ID,
        Name: node.Name,
        Sort: node.Sort !== undefined ? node.Sort : idx,
        IsGroup: false
      });
    });
    (sub.Groups || []).forEach((group, idx) => {
      const g = typeof group === 'string' ? { Name: group, Sort: sub.Nodes?.length + idx } : group;
      sortData.push({
        ID: 0,
        Name: g.Name,
        Sort: g.Sort !== undefined ? g.Sort : sub.Nodes?.length + idx,
        IsGroup: true
      });
    });
    sortData.sort((a, b) => a.Sort - b.Sort);
    setTempSortData(sortData);
    showMessage('已进入排序模式，拖动或多选批量操作', 'info');
  };

  const handleConfirmSort = async (sub) => {
    const newSortData = tempSortData.map((item, idx) => ({ ...item, Sort: idx }));
    try {
      await sortSubscription({
        ID: sub.ID,
        NodeSort: newSortData
      });
      showMessage('排序已更新');
      setSortingSubId(null);
      setTempSortData([]);
      fetchSubscriptions(page, rowsPerPage);
    } catch (error) {
      console.error(error);
      showMessage(error.message || '排序保存失败', 'error');
    }
  };

  const handleCancelSort = () => {
    setSortingSubId(null);
    setTempSortData([]);
    setSelectedSortItems([]);
    showMessage('已取消排序', 'info');
  };

  const onDragEnd = (result) => {
    if (!result.destination) return;
    const items = Array.from(tempSortData);
    const [reorderedItem] = items.splice(result.source.index, 1);
    items.splice(result.destination.index, 0, reorderedItem);
    setTempSortData(items);
  };

  // === 多选排序功能 ===
  const handleToggleSortSelect = (name) => {
    setSelectedSortItems((prev) => (prev.includes(name) ? prev.filter((n) => n !== name) : [...prev, name]));
  };

  const handleSelectAllSort = () => {
    setSelectedSortItems(tempSortData.map((item) => item.Name));
  };

  const handleClearSortSelection = () => {
    setSelectedSortItems([]);
  };

  // 批量排序（调用后端API）
  const handleBatchSort = async (sortBy, sortOrder) => {
    if (!sortingSubId) return;
    try {
      await batchSortSubscription({
        ID: sortingSubId,
        sortBy,
        sortOrder
      });
      showMessage('批量排序成功');
      // 重新加载订阅数据并刷新排序列表
      const response = await getSubscriptions({ page: page + 1, pageSize: rowsPerPage });
      const subs = response.data?.items || response.data || [];
      setSubscriptions(subs);
      // 找到当前排序的订阅并刷新排序数据
      const currentSub = subs.find((s) => s.ID === sortingSubId);
      if (currentSub) {
        handleStartSort(currentSub);
      }
    } catch (error) {
      console.error(error);
      showMessage(error.message || '批量排序失败', 'error');
    }
  };

  // 批量移动（前端本地操作，需要用户确认后保存）
  const handleBatchMove = (targetIndex) => {
    if (selectedSortItems.length === 0) return;

    const selected = tempSortData.filter((item) => selectedSortItems.includes(item.Name));
    const remaining = tempSortData.filter((item) => !selectedSortItems.includes(item.Name));

    // 插入到目标位置
    const newData = [...remaining];
    const insertAt = Math.min(Math.max(0, targetIndex), newData.length);
    newData.splice(insertAt, 0, ...selected);

    setTempSortData(newData);
    setSelectedSortItems([]);
    showMessage(`已移动 ${selected.length} 项到位置 ${insertAt + 1}`);
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
            <IconButton onClick={() => fetchSubscriptions(page, rowsPerPage)} disabled={loading}>
              <RefreshIcon
                sx={
                  loading
                    ? {
                        animation: 'spin 1s linear infinite',
                        '@keyframes spin': { from: { transform: 'rotate(0deg)' }, to: { transform: 'rotate(360deg)' } }
                      }
                    : {}
                }
              />
            </IconButton>
          </Stack>
        )
      }
    >
      {matchDownMd && (
        <Stack direction="row" justifyContent="flex-end" sx={{ mb: 2 }}>
          <IconButton onClick={() => fetchSubscriptions(page, rowsPerPage)} disabled={loading} size="small">
            <RefreshIcon
              sx={
                loading
                  ? {
                      animation: 'spin 1s linear infinite',
                      '@keyframes spin': { from: { transform: 'rotate(0deg)' }, to: { transform: 'rotate(360deg)' } }
                    }
                  : {}
              }
            />
          </IconButton>
        </Stack>
      )}

      {matchDownMd ? (
        <SubscriptionMobileCard
          subscriptions={subscriptions}
          page={page}
          rowsPerPage={rowsPerPage}
          expandedRows={expandedRows}
          sortingSubId={sortingSubId}
          tempSortData={tempSortData}
          selectedSortItems={selectedSortItems}
          theme={theme}
          onToggleRow={toggleRow}
          onClient={handleClient}
          onLogs={handleLogs}
          onEdit={handleEdit}
          onDelete={handleDelete}
          onCopy={handleCopy}
          onPreview={handlePreviewSubscription}
          onChainProxy={handleChainProxy}
          onStartSort={handleStartSort}
          onConfirmSort={handleConfirmSort}
          onCancelSort={handleCancelSort}
          onDragEnd={onDragEnd}
          onCopyToClipboard={copyToClipboard}
          getSortedItems={getSortedItems}
          onToggleSortSelect={handleToggleSortSelect}
          onSelectAllSort={handleSelectAllSort}
          onClearSortSelection={handleClearSortSelection}
          onBatchSort={handleBatchSort}
          onBatchMove={handleBatchMove}
        />
      ) : (
        <SubscriptionTable
          subscriptions={subscriptions}
          page={page}
          rowsPerPage={rowsPerPage}
          expandedRows={expandedRows}
          sortingSubId={sortingSubId}
          tempSortData={tempSortData}
          selectedSortItems={selectedSortItems}
          onToggleRow={toggleRow}
          onClient={handleClient}
          onLogs={handleLogs}
          onEdit={handleEdit}
          onDelete={handleDelete}
          onCopy={handleCopy}
          onPreview={handlePreviewSubscription}
          onChainProxy={handleChainProxy}
          onStartSort={handleStartSort}
          onConfirmSort={handleConfirmSort}
          onCancelSort={handleCancelSort}
          onDragEnd={onDragEnd}
          onCopyToClipboard={copyToClipboard}
          getSortedItems={getSortedItems}
          onToggleSortSelect={handleToggleSortSelect}
          onSelectAllSort={handleSelectAllSort}
          onClearSortSelection={handleClearSortSelection}
          onBatchSort={handleBatchSort}
          onBatchMove={handleBatchMove}
        />
      )}

      <Pagination
        page={page}
        pageSize={rowsPerPage}
        totalItems={totalItems}
        onPageChange={(e, newPage) => {
          setPage(newPage);
          fetchSubscriptions(newPage, rowsPerPage);
        }}
        onPageSizeChange={(e) => {
          const newValue = parseInt(e.target.value, 10);
          setRowsPerPage(newValue);
          localStorage.setItem('subscriptions_rowsPerPage', newValue);
          setPage(0);
          fetchSubscriptions(0, newValue);
        }}
        pageSizeOptions={[10, 20, 50, 100]}
      />

      {/* 添加/编辑对话框 */}
      <SubscriptionFormDialog
        open={dialogOpen}
        isEdit={isEdit}
        formData={formData}
        setFormData={setFormData}
        templates={templates}
        scripts={scripts}
        allNodes={allNodes}
        groupOptions={groupOptions}
        sourceOptions={sourceOptions}
        countryOptions={countryOptions}
        tagOptions={tagOptions}
        nodeGroupFilter={nodeGroupFilter}
        setNodeGroupFilter={setNodeGroupFilter}
        nodeSourceFilter={nodeSourceFilter}
        setNodeSourceFilter={setNodeSourceFilter}
        nodeSearchQuery={nodeSearchQuery}
        setNodeSearchQuery={setNodeSearchQuery}
        nodeCountryFilter={nodeCountryFilter}
        setNodeCountryFilter={setNodeCountryFilter}
        checkedAvailable={checkedAvailable}
        checkedSelected={checkedSelected}
        mobileTab={mobileTab}
        setMobileTab={setMobileTab}
        selectedNodeSearch={selectedNodeSearch}
        setSelectedNodeSearch={setSelectedNodeSearch}
        namingMode={namingMode}
        setNamingMode={setNamingMode}
        onClose={() => setDialogOpen(false)}
        onSubmit={handleSubmit}
        onAddNode={handleAddNode}
        onRemoveNode={handleRemoveNode}
        onAddAllVisible={handleAddAllVisible}
        onRemoveAll={handleRemoveAll}
        onToggleAvailable={handleToggleAvailable}
        onToggleSelected={handleToggleSelected}
        onAddChecked={handleAddChecked}
        onRemoveChecked={handleRemoveChecked}
        onToggleAllAvailable={handleToggleAllAvailable}
        onToggleAllSelected={handleToggleAllSelected}
        onPreview={handlePreview}
        previewLoading={previewLoading}
      />

      {/* 客户端对话框 */}
      <ClientUrlsDialog
        open={clientDialogOpen}
        clientUrls={clientUrls}
        onClose={() => setClientDialogOpen(false)}
        onQrCode={handleQrcode}
        onCopy={copyToClipboard}
      />

      {/* QR码对话框 */}
      <QrCodeDialog open={qrDialogOpen} title={qrTitle} url={qrUrl} onClose={() => setQrDialogOpen(false)} onCopy={copyToClipboard} />

      {/* 访问记录对话框 */}
      <AccessLogsDialog open={logsDialogOpen} logs={currentLogs} onClose={() => setLogsDialogOpen(false)} />

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
      <ConfirmDialog
        open={confirmOpen}
        title={confirmInfo.title}
        content={confirmInfo.content}
        onClose={handleConfirmClose}
        onConfirm={handleConfirmAction}
      />

      {/* 节点预览对话框 */}
      <NodePreviewDialog
        open={previewOpen}
        loading={previewLoading}
        data={previewData}
        tagColorMap={tagOptions.reduce((acc, tag) => {
          acc[tag.Name] = tag.Color;
          return acc;
        }, {})}
        onClose={() => setPreviewOpen(false)}
      />

      {/* 分享管理对话框 */}
      <ShareManageDialog
        open={shareDialogOpen}
        subscription={shareDialogSub}
        onClose={() => setShareDialogOpen(false)}
        showMessage={showMessage}
      />

      {/* 链式代理配置对话框 */}
      <ChainProxyDialog open={chainProxyOpen} subscription={chainProxySub} onClose={() => setChainProxyOpen(false)} />
    </MainCard>
  );
}
