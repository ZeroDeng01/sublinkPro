import { useState, useEffect, useCallback, useRef } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

// material-ui
import { useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import Snackbar from '@mui/material/Snackbar';
import Stack from '@mui/material/Stack';
import Tooltip from '@mui/material/Tooltip';

// icons
import AddIcon from '@mui/icons-material/Add';
import FlightIcon from '@mui/icons-material/Flight';
import RefreshIcon from '@mui/icons-material/Refresh';
import SettingsIcon from '@mui/icons-material/Settings';
import SpeedIcon from '@mui/icons-material/Speed';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import TaskProgressPanel from 'components/TaskProgressPanel';
import { useTaskProgress } from 'contexts/TaskProgressContext';
import ConfirmDialog from 'components/ConfirmDialog';
import Pagination from 'components/Pagination';
import IPDetailsDialog from 'components/IPDetailsDialog';

// api
import {
  getNodes,
  getNodeIds,
  addNodes,
  updateNode,
  deleteNode,
  deleteNodesBatch,
  runSpeedTest,
  getNodeCountries,
  getNodeGroups,
  getNodeSources,
  getNodeProtocols,
  batchUpdateNodeGroup,
  batchUpdateNodeDialerProxy,
  batchUpdateNodeSource,
  batchUpdateNodeCountry,
  getProtocolUIMeta
} from 'api/nodes';
import { getNodeCheckMeta } from 'api/nodeCheck';
import { getTags, batchSetNodeTags, batchRemoveNodeTags } from 'api/tags';

// local components
import {
  NodeDialog,
  SpeedTestDialog,
  ProfileSelectDialog,
  NodeCheckProfilesDrawer,
  BatchGroupDialog,
  BatchDialerProxyDialog,
  BatchTagDialog,
  BatchRemoveTagDialog,
  BatchSourceDialog,
  BatchCountryDialog,
  NodeAddResultDialog,
  NodeDetailsPanel,
  NodeRawProtocolDialog,
  NodeFilters,
  BatchActions,
  NodeMobileList,
  NodeTable
} from './component';

// utils
import { buildUnlockRulesPayload, setUnlockMeta, SPEED_TEST_TCP_OPTIONS, SPEED_TEST_MIHOMO_OPTIONS } from './utils';

// 列宽默认配置
const DEFAULT_COLUMN_WIDTHS = {
  checkbox: 48,
  remark: 180,
  protocol: 92,
  group: 120,
  source: 120,
  tags: 120,
  country: 80,
  delay: 130,
  speed: 130,
  ipFeatures: 240,
  actions: 140
};

// ==============================|| 节点管理 ||============================== //

export default function NodeList() {
  const theme = useTheme();
  const { t } = useTranslation();
  const matchDownMd = useMediaQuery(theme.breakpoints.down('md'));
  const location = useLocation();
  const navigate = useNavigate();

  const getSourceFilterFromQuery = useCallback((search) => {
    try {
      const params = new URLSearchParams(search);
      return params.get('source')?.trim() || '';
    } catch (error) {
      console.error('解析节点来源筛选参数失败:', error);
      return '';
    }
  }, []);

  const syncSourceFilterToQuery = useCallback(
    (nextSource) => {
      const params = new URLSearchParams(location.search);
      const currentSource = params.get('source')?.trim() || '';
      const normalizedSource = nextSource?.trim() || '';

      if (currentSource === normalizedSource) {
        return;
      }

      if (normalizedSource) {
        params.set('source', normalizedSource);
      } else {
        params.delete('source');
      }

      const search = params.toString();
      navigate(
        {
          pathname: location.pathname,
          search: search ? `?${search}` : ''
        },
        { replace: true }
      );
    },
    [location.pathname, location.search, navigate]
  );

  // Task progress for auto-refresh
  const { registerOnComplete, unregisterOnComplete } = useTaskProgress();

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
  // 节点表单
  const [nodeDialogOpen, setNodeDialogOpen] = useState(false);
  const [isEditNode, setIsEditNode] = useState(false);
  const [currentNode, setCurrentNode] = useState(null);
  const [nodeForm, setNodeForm] = useState({
    name: '',
    nameMode: 'link',
    link: '',
    dialerProxyName: '',
    group: '',
    mergeMode: '2', // 分开模式
    tags: [] // 标签列表
  });

  // 添加结果汇总弹窗
  const [addResultDialogOpen, setAddResultDialogOpen] = useState(false);
  const [addResult, setAddResult] = useState(null);

  // 原始协议对话框
  const [rawProtocolDialogOpen, setRawProtocolDialogOpen] = useState(false);
  const [rawProtocolNode, setRawProtocolNode] = useState(null);

  // 过滤器
  const [searchQuery, setSearchQuery] = useState('');
  const [groupFilter, setGroupFilter] = useState('');
  const [sourceFilter, setSourceFilter] = useState(() => getSourceFilterFromQuery(window.location.search));
  const [maxDelay, setMaxDelay] = useState('');
  const [minSpeed, setMinSpeed] = useState('');
  const [maxFraudScore, setMaxFraudScore] = useState('');
  const [speedStatusFilter, setSpeedStatusFilter] = useState('');
  const [delayStatusFilter, setDelayStatusFilter] = useState('');
  const [protocolFilter, setProtocolFilter] = useState('');
  const [residentialType, setResidentialType] = useState('');
  const [ipType, setIpType] = useState('');
  const [qualityStatus, setQualityStatus] = useState('');
  const [unlockRules, setUnlockRules] = useState([]);
  const [unlockRuleMode, setUnlockRuleMode] = useState('or');

  // 排序
  const [sortBy, setSortBy] = useState(''); // 'delay' | 'speed' | ''
  const [sortOrder, setSortOrder] = useState('asc'); // 'asc' | 'desc'

  // 列宽配置
  const [columnWidths, setColumnWidths] = useState(() => {
    try {
      const saved = localStorage.getItem('nodes_columnWidths');
      if (saved) {
        const parsed = JSON.parse(saved);
        // 迁移旧格式：将delaySpeed拆分为delay和speed
        if (parsed.delaySpeed && !parsed.delay && !parsed.speed) {
          const halfWidth = Math.floor(parsed.delaySpeed / 2);
          parsed.delay = halfWidth;
          parsed.speed = halfWidth;
          delete parsed.delaySpeed;
        }
        return { ...DEFAULT_COLUMN_WIDTHS, ...parsed };
      }
      return DEFAULT_COLUMN_WIDTHS;
    } catch (error) {
      console.error('Failed to load column widths:', error);
      return DEFAULT_COLUMN_WIDTHS;
    }
  });

  // 分页
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(() => {
    const saved = localStorage.getItem('nodes_rowsPerPage');
    return saved ? parseInt(saved, 10) : 20;
  });
  const [totalItems, setTotalItems] = useState(0);

  // 测速配置
  const [speedTestDialogOpen, setSpeedTestDialogOpen] = useState(false);
  // 策略管理
  const [profilesDrawerOpen, setProfilesDrawerOpen] = useState(false);
  const [profileSelectOpen, setProfileSelectOpen] = useState(false);
  const [profileSelectNodeIds, setProfileSelectNodeIds] = useState([]);
  const [speedTestForm, setSpeedTestForm] = useState({
    cron: '',
    enabled: false,
    mode: 'tcp',
    url: '',
    timeout: 5,
    groups: [],
    detect_country: false,
    landing_ip_url: '',
    latency_concurrency: 0,
    speed_concurrency: 1,

    traffic_by_group: true,
    traffic_by_source: true,
    traffic_by_node: false,
    include_handshake: true,
    persist_host: false,
    dns_server: '',
    dns_presets: []
  });

  // 国家筛选
  const [countryFilter, setCountryFilter] = useState([]);
  const [countryOptions, setCountryOptions] = useState([]);
  // 从后端获取的分组和来源选项
  const [groupOptions, setGroupOptions] = useState([]);
  const [sourceOptions, setSourceOptions] = useState([]);
  const [protocolOptions, setProtocolOptions] = useState([]);

  // 代理节点选择 - 从后端获取的完整节点列表
  const [proxyNodeOptions, setProxyNodeOptions] = useState([]);
  const [loadingProxyNodes, setLoadingProxyNodes] = useState(false);

  // 批量修改分组
  const [batchGroupDialogOpen, setBatchGroupDialogOpen] = useState(false);
  const [batchGroupValue, setBatchGroupValue] = useState('');

  // 批量修改前置代理
  const [batchDialerProxyDialogOpen, setBatchDialerProxyDialogOpen] = useState(false);
  const [batchDialerProxyValue, setBatchDialerProxyValue] = useState('');

  // 批量修改来源
  const [batchSourceDialogOpen, setBatchSourceDialogOpen] = useState(false);
  const [batchSourceValue, setBatchSourceValue] = useState('');

  // 批量修改国家
  const [batchCountryDialogOpen, setBatchCountryDialogOpen] = useState(false);
  const [batchCountryValue, setBatchCountryValue] = useState('');

  // 标签相关状态
  const [tagFilter, setTagFilter] = useState([]);
  const [tagOptions, setTagOptions] = useState([]);
  const [tagColorMap, setTagColorMap] = useState({});
  const [batchTagDialogOpen, setBatchTagDialogOpen] = useState(false);
  const [batchTagValue, setBatchTagValue] = useState([]);
  const [batchRemoveTagDialogOpen, setBatchRemoveTagDialogOpen] = useState(false);
  const [batchRemoveTagValue, setBatchRemoveTagValue] = useState([]);

  // IP详情弹窗
  const [ipDialogOpen, setIpDialogOpen] = useState(false);
  const [selectedIP, setSelectedIP] = useState('');

  // 节点详情面板
  const [detailsPanelOpen, setDetailsPanelOpen] = useState(false);
  const [detailsNode, setDetailsNode] = useState(null);

  // 协议 UI 元数据
  const [protocolMeta, setProtocolMeta] = useState([]);

  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });

  // 后端已完成过滤和排序，直接使用 nodes 数组
  const filteredNodes = nodes;

  // 防抖定时器引用
  const debounceTimerRef = useRef(null);

  // 获取节点列表（支持过滤和分页参数）
  // 注意：不依赖 page/rowsPerPage，而是通过参数传递，避免触发 filter useEffect 循环
  const fetchNodes = useCallback(
    async (filterParams = {}) => {
      setLoading(true);
      try {
        // 构建过滤参数
        const params = {};
        if (filterParams.search) params.search = filterParams.search;
        if (filterParams.group) params.group = filterParams.group;
        if (filterParams.source) params.source = filterParams.source;
        if (filterParams.maxDelay) params.maxDelay = filterParams.maxDelay;
        if (filterParams.minSpeed) params.minSpeed = filterParams.minSpeed;
        if (filterParams.maxFraudScore) params.maxFraudScore = filterParams.maxFraudScore;
        if (filterParams.speedStatus) params.speedStatus = filterParams.speedStatus;
        if (filterParams.delayStatus) params.delayStatus = filterParams.delayStatus;
        if (filterParams.protocol) params.protocol = filterParams.protocol;
        if (filterParams.residentialType) params.residentialType = filterParams.residentialType;
        if (filterParams.ipType) params.ipType = filterParams.ipType;
        if (filterParams.qualityStatus) params.qualityStatus = filterParams.qualityStatus;
        if (filterParams.unlockRules?.some((rule) => rule.provider || rule.status || rule.keyword)) {
          params.unlockRules = buildUnlockRulesPayload(filterParams.unlockRules);
          params.unlockRuleMode = filterParams.unlockRuleMode || 'or';
        }
        if (filterParams.countries && filterParams.countries.length > 0) {
          params['countries[]'] = filterParams.countries;
        }
        if (filterParams.tags && filterParams.tags.length > 0) {
          params['tags[]'] = filterParams.tags.map((t) => t.name || t);
        }
        if (filterParams.sortBy) params.sortBy = filterParams.sortBy;
        if (filterParams.sortOrder) params.sortOrder = filterParams.sortOrder;

        // 分页参数必须通过 filterParams 传递
        params.page = (filterParams.page ?? 0) + 1; // 后端是1-indexed
        params.pageSize = filterParams.pageSize ?? 20;

        const response = await getNodes(params);
        // 处理分页响应
        if (response.data && response.data.items !== undefined) {
          setNodes(response.data.items || []);
          setTotalItems(response.data.total || 0);
        } else {
          // 向后兼容：老格式直接返回数组
          setNodes(response.data || []);
          setTotalItems((response.data || []).length);
        }
      } catch (error) {
        console.error(error);
        showMessage(error.message || t('nodes.page.messages.loadFailed'), 'error');
      } finally {
        setLoading(false);
      }
    },
    [t]
  );

  // 获取代理节点选项（用于订阅下载代理选择）
  const fetchProxyNodes = useCallback(async () => {
    if (proxyNodeOptions.length > 0) return; // 已加载过则不重复加载
    setLoadingProxyNodes(true);
    try {
      const response = await getNodes({});
      setProxyNodeOptions(response.data || []);
    } catch (error) {
      console.error('获取代理节点列表失败:', error);
    } finally {
      setLoadingProxyNodes(false);
    }
  }, [proxyNodeOptions.length]);

  // 初始化加载
  useEffect(() => {
    fetchNodes({ page: 0, pageSize: rowsPerPage, source: getSourceFilterFromQuery(location.search) });
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
    // 请求标签列表
    getTags()
      .then((res) => {
        const tags = res.data || [];
        setTagOptions(tags);
        // 构建标签颜色映射
        const colorMap = {};
        tags.forEach((tag) => {
          colorMap[tag.name] = tag.color || '#1976d2';
        });
        setTagColorMap(colorMap);
      })
      .catch(console.error);
    // 请求协议 UI 元数据
    getProtocolUIMeta()
      .then((res) => {
        setProtocolMeta(res.data || []);
      })
      .catch(console.error);
    // 请求协议列表
    getNodeProtocols()
      .then((res) => {
        setProtocolOptions(res.data || []);
      })
      .catch(console.error);
    getNodeCheckMeta()
      .then((res) => {
        setUnlockMeta(res.data || {});
      })
      .catch(console.error);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    const nextSourceFilter = getSourceFilterFromQuery(location.search);
    setSourceFilter((prev) => (prev === nextSourceFilter ? prev : nextSourceFilter));
  }, [getSourceFilterFromQuery, location.search]);

  useEffect(() => {
    syncSourceFilterToQuery(sourceFilter);
  }, [sourceFilter, syncSourceFilterToQuery]);

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
        maxFraudScore: maxFraudScore,
        speedStatus: speedStatusFilter,
        delayStatus: delayStatusFilter,
        protocol: protocolFilter,
        residentialType: residentialType,
        ipType: ipType,
        qualityStatus: qualityStatus,
        unlockRules: unlockRules,
        unlockRuleMode: unlockRuleMode,
        countries: countryFilter,
        tags: tagFilter,
        sortBy: sortBy,
        sortOrder: sortOrder,
        page: 0, // 筛选条件变化时重置到第一页
        pageSize: rowsPerPage
      };
      setPage(0); // 重置分页
      fetchNodes(filterParams);
    }, 300); // 300ms 防抖延迟

    // 清理函数
    return () => {
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current);
      }
    };
  }, [
    searchQuery,
    groupFilter,
    sourceFilter,
    maxDelay,
    minSpeed,
    maxFraudScore,
    speedStatusFilter,
    delayStatusFilter,
    protocolFilter,
    residentialType,
    ipType,
    qualityStatus,
    unlockRules,
    unlockRuleMode,
    countryFilter,
    tagFilter,
    sortBy,
    sortOrder,
    fetchNodes,
    rowsPerPage
  ]);

  const showMessage = (message, severity = 'success') => {
    setSnackbar({ open: true, message, severity });
  };

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
    showMessage(t('common.copied'));
  };

  const resetFilters = () => {
    setSearchQuery('');
    setGroupFilter('');
    setSourceFilter('');
    setMaxDelay('');
    setMinSpeed('');
    setMaxFraudScore('');
    setSpeedStatusFilter('');
    setDelayStatusFilter('');
    setCountryFilter([]);
    setTagFilter([]);
    setProtocolFilter('');
    setResidentialType('');
    setIpType('');
    setQualityStatus('');
    setUnlockRules([]);
    setUnlockRuleMode('or');
    setSortBy('');
    setSortOrder('asc');
  };

  const getCurrentFilters = () => ({
    search: searchQuery,
    group: groupFilter,
    source: sourceFilter,
    maxDelay: maxDelay,
    minSpeed: minSpeed,
    maxFraudScore: maxFraudScore,
    speedStatus: speedStatusFilter,
    delayStatus: delayStatusFilter,
    protocol: protocolFilter,
    residentialType: residentialType,
    ipType: ipType,
    qualityStatus: qualityStatus,
    unlockRules: unlockRules,
    unlockRuleMode: unlockRuleMode,
    countries: countryFilter,
    tags: tagFilter,
    sortBy: sortBy,
    sortOrder: sortOrder,
    page: page,
    pageSize: rowsPerPage
  });

  // 刷新下拉框选项数据
  const refreshFilterOptions = useCallback(() => {
    // 刷新国家选项
    getNodeCountries()
      .then((res) => {
        setCountryOptions(res.data || []);
      })
      .catch(console.error);
    // 刷新分组选项
    getNodeGroups()
      .then((res) => {
        setGroupOptions((res.data || []).sort());
      })
      .catch(console.error);
    // 刷新来源选项
    getNodeSources()
      .then((res) => {
        setSourceOptions((res.data || []).sort());
      })
      .catch(console.error);
    // 刷新标签选项
    getTags()
      .then((res) => {
        const tags = res.data || [];
        setTagOptions(tags);
        const colorMap = {};
        tags.forEach((tag) => {
          colorMap[tag.name] = tag.color || '#1976d2';
        });
        setTagColorMap(colorMap);
      })
      .catch(console.error);
    // 刷新协议选项
    getNodeProtocols()
      .then((res) => {
        setProtocolOptions(res.data || []);
      })
      .catch(console.error);
  }, []);

  const handleRefresh = useCallback(() => {
    fetchNodes(getCurrentFilters());
    refreshFilterOptions();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [
    fetchNodes,
    refreshFilterOptions,
    searchQuery,
    groupFilter,
    sourceFilter,
    maxDelay,
    minSpeed,
    maxFraudScore,
    speedStatusFilter,
    delayStatusFilter,
    protocolFilter,
    countryFilter,
    tagFilter,
    residentialType,
    ipType,
    qualityStatus,
    unlockRules,
    unlockRuleMode,
    sortBy,
    sortOrder,
    page,
    rowsPerPage
  ]);

  // 监听任务完成，自动刷新节点列表
  useEffect(() => {
    const onTaskComplete = ({ taskType, status }) => {
      // 当测速或订阅更新任务完成时，自动刷新列表
      if (status === 'completed' && (taskType === 'speed_test' || taskType === 'sub_update')) {
        handleRefresh();
      }
    };
    registerOnComplete(onTaskComplete);
    return () => {
      unregisterOnComplete(onTaskComplete);
    };
  }, [registerOnComplete, unregisterOnComplete, handleRefresh]);

  // === 节点操作 ===
  const handleAddNode = () => {
    setIsEditNode(false);
    setCurrentNode(null);
    setNodeForm({ name: '', nameMode: 'link', link: '', dialerProxyName: '', group: '', mergeMode: '2', tags: [] });
    setNodeDialogOpen(true);
  };

  const handleEditNode = (node) => {
    setIsEditNode(true);
    setCurrentNode(node);
    // 解析节点的标签，将字符串转换为标签对象数组
    let nodeTags = [];
    if (node.Tags) {
      const tagNames = node.Tags.split(',')
        .map((t) => t.trim())
        .filter((t) => t);
      nodeTags = tagNames.map((name) => {
        const found = tagOptions.find((t) => t.name === name);
        return found || { name, color: '#1976d2' };
      });
    }
    setNodeForm({
      name: node.Name,
      nameMode: node.NameMode || 'link',
      link: node.Link?.split(',').join('\n') || '',
      dialerProxyName: node.DialerProxyName || '',
      group: node.Group || '',
      mergeMode: '1',
      tags: nodeTags
    });
    setNodeDialogOpen(true);
  };

  const handleOpenRawProtocol = (node) => {
    setRawProtocolNode(node);
    setRawProtocolDialogOpen(true);
  };

  const handleCloseRawProtocol = () => {
    setRawProtocolDialogOpen(false);
    setRawProtocolNode(null);
  };

  const handleRawProtocolUpdate = () => {
    loadNodes(); // 刷新节点列表
  };

  const handleDeleteNode = async (node) => {
    const displayName = node.EffectiveName || node.Name || node.LinkName;
    openConfirm(t('nodes.page.confirm.deleteTitle'), t('nodes.page.confirm.deleteOne', { name: displayName }), async () => {
      try {
        await deleteNode({ id: node.ID });
        showMessage(t('nodes.page.messages.deleteSuccess'));
        handleRefresh();
      } catch (error) {
        console.error(error);
        showMessage(error.message || t('nodes.page.messages.deleteFailed'), 'error');
      }
    });
  };

  const handleBatchDelete = async () => {
    if (selectedNodes.length === 0) {
      showMessage(t('nodes.page.messages.selectDeleteRequired'), 'warning');
      return;
    }
    openConfirm(
      t('nodes.page.confirm.batchDeleteTitle'),
      t('nodes.page.confirm.batchDelete', { count: selectedNodes.length }),
      async () => {
        try {
          const ids = selectedNodes.map((node) => node.ID);
          await deleteNodesBatch(ids);
          showMessage(t('nodes.page.messages.batchDeleteSuccess', { count: selectedNodes.length }));
          setSelectedNodes([]);
          handleRefresh();
        } catch (error) {
          console.error(error);
          showMessage(error.message || t('nodes.page.messages.batchDeleteFailed'), 'error');
        }
      }
    );
  };

  // 批量修改分组
  const handleBatchGroup = () => {
    if (selectedNodes.length === 0) {
      showMessage(t('nodes.page.messages.selectModifyRequired'), 'warning');
      return;
    }
    setBatchGroupValue('');
    setBatchGroupDialogOpen(true);
  };

  const handleSubmitBatchGroup = async () => {
    try {
      const ids = selectedNodes.map((node) => node.ID);
      await batchUpdateNodeGroup(ids, batchGroupValue);
      showMessage(t('nodes.page.messages.batchGroupSuccess', { count: selectedNodes.length }));
      setSelectedNodes([]);
      setBatchGroupDialogOpen(false);
      fetchNodes(getCurrentFilters());
      // 刷新分组选项
      getNodeGroups().then((res) => {
        setGroupOptions((res.data || []).sort());
      });
    } catch (error) {
      console.error(error);
      showMessage(error.message || t('nodes.page.messages.batchGroupFailed'), 'error');
    }
  };

  // 批量修改前置代理
  const handleBatchDialerProxy = () => {
    if (selectedNodes.length === 0) {
      showMessage(t('nodes.page.messages.selectModifyRequired'), 'warning');
      return;
    }
    setBatchDialerProxyValue('');
    fetchProxyNodes();
    setBatchDialerProxyDialogOpen(true);
  };

  const handleSubmitBatchDialerProxy = async () => {
    try {
      const ids = selectedNodes.map((node) => node.ID);
      await batchUpdateNodeDialerProxy(ids, batchDialerProxyValue);
      showMessage(t('nodes.page.messages.batchDialerProxySuccess', { count: selectedNodes.length }));
      setSelectedNodes([]);
      setBatchDialerProxyDialogOpen(false);
      fetchNodes(getCurrentFilters());
    } catch (error) {
      console.error(error);
      showMessage(error.message || t('nodes.page.messages.batchDialerProxyFailed'), 'error');
    }
  };

  // 批量修改来源
  const handleBatchSource = () => {
    if (selectedNodes.length === 0) {
      showMessage(t('nodes.page.messages.selectModifyRequired'), 'warning');
      return;
    }
    setBatchSourceValue('');
    setBatchSourceDialogOpen(true);
  };

  const handleSubmitBatchSource = async () => {
    try {
      const ids = selectedNodes.map((node) => node.ID);
      // 如果值为空，设置为 manual
      const source = batchSourceValue.trim() || 'manual';
      await batchUpdateNodeSource(ids, source);
      showMessage(t('nodes.page.messages.batchSourceSuccess', { count: selectedNodes.length }));
      setSelectedNodes([]);
      setBatchSourceDialogOpen(false);
      fetchNodes(getCurrentFilters());
      // 刷新来源选项
      getNodeSources().then((res) => {
        setSourceOptions((res.data || []).sort());
      });
    } catch (error) {
      console.error(error);
      showMessage(error.message || t('nodes.page.messages.batchSourceFailed'), 'error');
    }
  };

  // 批量修改国家
  const handleBatchCountry = () => {
    if (selectedNodes.length === 0) {
      showMessage(t('nodes.page.messages.selectModifyRequired'), 'warning');
      return;
    }
    setBatchCountryValue('');
    setBatchCountryDialogOpen(true);
  };

  const handleSubmitBatchCountry = async () => {
    try {
      const ids = selectedNodes.map((node) => node.ID);
      await batchUpdateNodeCountry(ids, batchCountryValue);
      showMessage(t('nodes.page.messages.batchCountrySuccess', { count: selectedNodes.length }));
      setSelectedNodes([]);
      setBatchCountryDialogOpen(false);
      fetchNodes(getCurrentFilters());
      // 刷新国家选项
      getNodeCountries().then((res) => {
        setCountryOptions(res.data || []);
      });
    } catch (error) {
      console.error(error);
      showMessage(error.message || t('nodes.page.messages.batchCountryFailed'), 'error');
    }
  };

  // 批量设置标签
  const handleBatchTag = () => {
    if (selectedNodes.length === 0) {
      showMessage(t('nodes.page.messages.selectTagRequired'), 'warning');
      return;
    }
    setBatchTagValue([]);
    setBatchTagDialogOpen(true);
  };

  const handleSubmitBatchTag = async () => {
    try {
      const ids = selectedNodes.map((node) => node.ID);
      const tagNames = batchTagValue.map((t) => t.name || t);
      // 使用新的批量设置标签API（覆盖模式）
      await batchSetNodeTags({ nodeIds: ids, tagNames: tagNames });
      showMessage(t('nodes.page.messages.batchTagSuccess', { count: selectedNodes.length }));
      setSelectedNodes([]);
      setBatchTagDialogOpen(false);
      fetchNodes(getCurrentFilters());
    } catch (error) {
      console.error(error);
      showMessage(error.message || t('nodes.page.messages.batchTagFailed'), 'error');
    }
  };

  // 批量移除标签
  const handleBatchRemoveTag = () => {
    if (selectedNodes.length === 0) {
      showMessage(t('nodes.page.messages.selectRemoveTagRequired'), 'warning');
      return;
    }
    setBatchRemoveTagValue([]);
    setBatchRemoveTagDialogOpen(true);
  };

  const handleSubmitBatchRemoveTag = async () => {
    try {
      const ids = selectedNodes.map((node) => node.ID);
      const tagNames = batchRemoveTagValue.map((t) => t.name || t);
      await batchRemoveNodeTags({ nodeIds: ids, tagNames: tagNames });
      showMessage(t('nodes.page.messages.batchRemoveTagSuccess', { count: selectedNodes.length }));
      setSelectedNodes([]);
      setBatchRemoveTagDialogOpen(false);
      fetchNodes(getCurrentFilters());
    } catch (error) {
      console.error(error);
      showMessage(error.message || t('nodes.page.messages.batchRemoveTagFailed'), 'error');
    }
  };

  const handleSubmitNode = async () => {
    // 检测是否是 WireGuard 配置文件格式（包含 [Interface] 和 [Peer]）
    const isWireGuardConfig = nodeForm.link.includes('[Interface]') && nodeForm.link.includes('[Peer]');
    // 检测是否是 Clash YAML 配置格式（包含 proxies: 关键字）
    const isClashYamlConfig = nodeForm.link.includes('proxies:');

    let nodeLinks;
    if (isWireGuardConfig || isClashYamlConfig) {
      // WireGuard 或 Clash YAML 配置文件格式，保持原样不分割
      nodeLinks = [nodeForm.link.trim()];
    } else {
      // 常规链接格式，按换行符和逗号分割
      nodeLinks = nodeForm.link
        .split(/[\r\n,]/)
        .map((item) => item.trim())
        .filter((item) => item !== '');
    }

    if (nodeLinks.length === 0) {
      showMessage(t('nodes.page.messages.nodeLinkRequired'), 'warning');
      return;
    }

    // 提取标签名称
    const tagNames = (nodeForm.tags || []).map((t) => t.name || t).join(',');

    try {
      if (isEditNode) {
        const processedLink = nodeLinks.join(',');
        await updateNode({
          oldname: currentNode.Name,
          oldlink: currentNode.Link,
          link: processedLink,
          name: nodeForm.name.trim(),
          nameMode: nodeForm.nameMode || 'link',
          dialerProxyName: nodeForm.dialerProxyName.trim(),
          group: nodeForm.group.trim(),
          tags: tagNames
        });
        showMessage(t('nodes.page.messages.updateSuccess'));
        setNodeDialogOpen(false);
        handleRefresh();
      } else {
        // 收集所有链接的添加结果
        const result = { added: 0, skipped: [], failed: [] };
        for (const link of nodeLinks) {
          try {
            const res = await addNodes({
              link,
              name: '',
              dialerProxyName: nodeForm.dialerProxyName.trim(),
              group: nodeForm.group.trim(),
              tags: tagNames
            });
            // 检查后端返回的跳过标记
            if (res.data?.skipped) {
              result.skipped.push(res.data.duplicateInfo);
            } else {
              result.added++;
            }
          } catch (error) {
            result.failed.push({ link, error: error.message || t('nodes.page.messages.addFailed') });
          }
        }
        setNodeDialogOpen(false);
        // 单条且全部成功时，仅显示简单提示
        if (nodeLinks.length === 1 && result.added === 1) {
          showMessage(t('nodes.page.messages.addSuccess'));
        } else {
          // 多条链接或有跳过/失败时，弹出结果汇总面板
          setAddResult(result);
          setAddResultDialogOpen(true);
        }
        handleRefresh();
      }
    } catch (error) {
      console.error(error);
      showMessage(error.message || (isEditNode ? t('nodes.page.messages.updateFailed') : t('nodes.page.messages.addFailed')), 'error');
    }
  };

  // === 测速配置 ===
  const handleOpenSpeedTest = () => {
    // 跳转到节点检测策略页面
    navigate('/subscription/node-check');
  };

  const handleSpeedModeChange = (mode) => {
    const newUrl = mode === 'mihomo' ? SPEED_TEST_MIHOMO_OPTIONS[0].value : SPEED_TEST_TCP_OPTIONS[0].value;
    setSpeedTestForm({ ...speedTestForm, mode, url: newUrl });
  };

  const handleSubmitSpeedTest = async () => {
    // 原测速配置API已迁移到策略管理
    // 关闭旧对话框并打开策略管理抽屉
    setSpeedTestDialogOpen(false);
    setProfilesDrawerOpen(true);
    showMessage(t('nodes.page.messages.speedConfigMoved'), 'info');
  };

  const handleRunSpeedTest = async () => {
    try {
      await runSpeedTest();
      showMessage(t('nodes.page.messages.speedTaskStarted'));
    } catch (error) {
      console.error(error);
      showMessage(error.message || t('nodes.page.messages.speedTaskFailed'), 'error');
    }
  };

  const handleBatchSpeedTest = () => {
    if (selectedNodes.length === 0) {
      showMessage(t('nodes.page.messages.selectCheckRequired'), 'warning');
      return;
    }
    const ids = selectedNodes.map((node) => node.ID);
    setProfileSelectNodeIds(ids);
    setProfileSelectOpen(true);
  };

  const handleSingleSpeedTest = (node) => {
    setProfileSelectNodeIds([node.ID]);
    setProfileSelectOpen(true);
  };

  // 选择所有（获取符合当前筛选条件的所有节点ID）
  const handleSelectAll = async (event) => {
    if (event.target.checked) {
      try {
        // 从后端获取所有符合筛选条件的节点ID
        const filters = getCurrentFilters();
        const params = {};
        if (filters.search) params.search = filters.search;
        if (filters.group) params.group = filters.group;
        if (filters.source) params.source = filters.source;
        if (filters.maxDelay) params.maxDelay = filters.maxDelay;
        if (filters.minSpeed) params.minSpeed = filters.minSpeed;
        if (filters.maxFraudScore) params.maxFraudScore = filters.maxFraudScore;
        if (filters.speedStatus) params.speedStatus = filters.speedStatus;
        if (filters.delayStatus) params.delayStatus = filters.delayStatus;
        if (filters.protocol) params.protocol = filters.protocol;
        if (filters.residentialType) params.residentialType = filters.residentialType;
        if (filters.ipType) params.ipType = filters.ipType;
        if (filters.qualityStatus) params.qualityStatus = filters.qualityStatus;
        if (filters.unlockRules?.some((rule) => rule.provider || rule.status || rule.keyword)) {
          params.unlockRules = buildUnlockRulesPayload(filters.unlockRules);
          params.unlockRuleMode = filters.unlockRuleMode || 'or';
        }
        if (filters.countries && filters.countries.length > 0) {
          params['countries[]'] = filters.countries;
        }
        if (filters.tags && filters.tags.length > 0) {
          params['tags[]'] = filters.tags.map((t) => t.name || t);
        }

        const response = await getNodeIds(params);
        const allIds = response.data || [];

        // 将ID转换为节点对象（只包含ID，用于后续操作）
        const selectedObjs = allIds.map((id) => ({ ID: id }));
        setSelectedNodes(selectedObjs);
        showMessage(t('nodes.page.messages.selectedAll', { count: allIds.length }));
      } catch (error) {
        console.error('获取所有节点ID失败:', error);
        // 回退方案：只选择当前页
        setSelectedNodes(filteredNodes);
      }
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

  // 列宽调整处理
  const handleColumnResize = useCallback((columnKey, newWidth) => {
    setColumnWidths((prev) => {
      const updated = { ...prev, [columnKey]: newWidth };
      try {
        localStorage.setItem('nodes_columnWidths', JSON.stringify(updated));
      } catch (error) {
        console.error('Failed to save column widths:', error);
      }
      return updated;
    });
  }, []);

  // 重置列宽
  const handleResetColumnWidths = useCallback(() => {
    setColumnWidths(DEFAULT_COLUMN_WIDTHS);
    try {
      localStorage.removeItem('nodes_columnWidths');
      showMessage(t('nodes.page.messages.columnWidthsReset'));
    } catch (error) {
      console.error('Failed to reset column widths:', error);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <MainCard
      title={t('nodes.page.title')}
      secondary={
        matchDownMd ? (
          <Tooltip title={t('nodes.page.actions.addMore')}>
            <Button variant="contained" startIcon={<AddIcon />} onClick={handleAddNode}>
              {t('common.add')}
            </Button>
          </Tooltip>
        ) : (
          <Stack direction="row" spacing={1}>
            <Button variant="contained" startIcon={<AddIcon />} onClick={handleAddNode}>
              {t('nodes.page.actions.addNode')}
            </Button>
            <Button variant="outlined" color="primary" startIcon={<FlightIcon />} onClick={() => navigate('/subscription/airports')}>
              {t('nodes.page.actions.airports')}
            </Button>
            <Button variant="outlined" color="info" startIcon={<SettingsIcon />} onClick={handleOpenSpeedTest}>
              {t('nodes.page.actions.checkSettings')}
            </Button>
            <Button variant="outlined" startIcon={<SpeedIcon />} onClick={handleBatchSpeedTest}>
              {t('nodes.page.actions.batchCheck')}
            </Button>
            <IconButton onClick={handleRefresh} disabled={loading}>
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
      {/* 任务进度显示 */}
      <TaskProgressPanel />

      {/* 移动端顶部额外按钮栏 */}
      {matchDownMd && (
        <Stack direction="row" spacing={1} sx={{ mb: 2, overflowX: 'auto', pb: 1 }} className="hide-scrollbar">
          <Button
            size="small"
            variant="outlined"
            color="primary"
            startIcon={<FlightIcon />}
            onClick={() => navigate('/subscription/airports')}
            sx={{ whiteSpace: 'nowrap' }}
          >
            {t('nodes.page.actions.airportsShort')}
          </Button>
          <Button
            size="small"
            variant="outlined"
            color="info"
            startIcon={<SettingsIcon />}
            onClick={handleOpenSpeedTest}
            sx={{ whiteSpace: 'nowrap' }}
          >
            {t('nodes.page.actions.checkSettings')}
          </Button>
          <Button size="small" variant="outlined" startIcon={<SpeedIcon />} onClick={handleBatchSpeedTest} sx={{ whiteSpace: 'nowrap' }}>
            {t('nodes.page.actions.batchCheck')}
          </Button>
          <IconButton size="small" onClick={handleRefresh} disabled={loading}>
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

      <NodeFilters
        searchQuery={searchQuery}
        setSearchQuery={setSearchQuery}
        groupFilter={groupFilter}
        setGroupFilter={setGroupFilter}
        sourceFilter={sourceFilter}
        setSourceFilter={setSourceFilter}
        maxDelay={maxDelay}
        setMaxDelay={setMaxDelay}
        minSpeed={minSpeed}
        setMinSpeed={setMinSpeed}
        maxFraudScore={maxFraudScore}
        setMaxFraudScore={setMaxFraudScore}
        speedStatusFilter={speedStatusFilter}
        setSpeedStatusFilter={setSpeedStatusFilter}
        delayStatusFilter={delayStatusFilter}
        setDelayStatusFilter={setDelayStatusFilter}
        residentialType={residentialType}
        setResidentialType={setResidentialType}
        ipType={ipType}
        setIpType={setIpType}
        qualityStatus={qualityStatus}
        setQualityStatus={setQualityStatus}
        unlockRules={unlockRules}
        setUnlockRules={setUnlockRules}
        unlockRuleMode={unlockRuleMode}
        setUnlockRuleMode={setUnlockRuleMode}
        countryFilter={countryFilter}
        setCountryFilter={setCountryFilter}
        tagFilter={tagFilter}
        setTagFilter={setTagFilter}
        protocolFilter={protocolFilter}
        setProtocolFilter={setProtocolFilter}
        groupOptions={groupOptions}
        sourceOptions={sourceOptions}
        countryOptions={countryOptions}
        tagOptions={tagOptions}
        protocolOptions={protocolOptions}
        onReset={resetFilters}
      />

      {/* 批量操作 */}
      <BatchActions
        selectedCount={selectedNodes.length}
        totalCount={totalItems}
        onSelectAll={handleSelectAll}
        onClearSelection={() => setSelectedNodes([])}
        onDelete={handleBatchDelete}
        onGroup={handleBatchGroup}
        onSource={handleBatchSource}
        onCountry={handleBatchCountry}
        onDialerProxy={handleBatchDialerProxy}
        onTag={handleBatchTag}
        onRemoveTag={handleBatchRemoveTag}
      />

      {/* 桌面端表格工具栏 */}
      {!matchDownMd && (
        <Box sx={{ mb: 1, display: 'flex', justifyContent: 'flex-end' }}>
          <Button size="small" variant="text" onClick={handleResetColumnWidths}>
            {t('nodes.page.actions.resetColumnWidths')}
          </Button>
        </Box>
      )}

      {/* 节点列表 */}
      {matchDownMd ? (
        <NodeMobileList
          nodes={filteredNodes}
          page={page}
          rowsPerPage={rowsPerPage}
          selectedNodes={selectedNodes}
          tagColorMap={tagColorMap}
          protocolMeta={protocolMeta}
          onSelect={handleSelectNode}
          onViewDetails={(node) => {
            setDetailsNode(node);
            setDetailsPanelOpen(true);
          }}
        />
      ) : (
        <NodeTable
          nodes={filteredNodes}
          page={page}
          rowsPerPage={rowsPerPage}
          selectedNodes={selectedNodes}
          sortBy={sortBy}
          sortOrder={sortOrder}
          tagColorMap={tagColorMap}
          protocolMeta={protocolMeta}
          columnWidths={columnWidths}
          onSelectAll={handleSelectAll}
          onSelect={handleSelectNode}
          onSort={handleSort}
          onSpeedTest={handleSingleSpeedTest}
          onCopy={copyToClipboard}
          onEdit={handleEditNode}
          onDelete={handleDeleteNode}
          onViewDetails={(node) => {
            setDetailsNode(node);
            setDetailsPanelOpen(true);
          }}
          onColumnResize={handleColumnResize}
          onOpenRawProtocol={handleOpenRawProtocol}
        />
      )}

      <Pagination
        page={page}
        pageSize={rowsPerPage}
        totalItems={totalItems}
        onPageChange={(e, newPage) => {
          setPage(newPage);
          // 触发数据重新加载
          fetchNodes({ ...getCurrentFilters(), page: newPage, pageSize: rowsPerPage });
        }}
        onPageSizeChange={(e) => {
          const newValue = parseInt(e.target.value, 10);
          setRowsPerPage(newValue);
          localStorage.setItem('nodes_rowsPerPage', newValue);
          setPage(0);
          // 触发数据重新加载
          fetchNodes({ ...getCurrentFilters(), page: 0, pageSize: newValue });
        }}
        pageSizeOptions={[10, 20, 50, 100]}
      />

      {/* 添加/编辑节点对话框 */}
      <NodeDialog
        open={nodeDialogOpen}
        isEdit={isEditNode}
        nodeForm={nodeForm}
        setNodeForm={setNodeForm}
        groupOptions={groupOptions}
        proxyNodeOptions={proxyNodeOptions}
        loadingProxyNodes={loadingProxyNodes}
        tagOptions={tagOptions}
        onClose={() => setNodeDialogOpen(false)}
        onSubmit={handleSubmitNode}
        onFetchProxyNodes={fetchProxyNodes}
      />

      {/* 测速设置对话框 */}
      <SpeedTestDialog
        open={speedTestDialogOpen}
        speedTestForm={speedTestForm}
        setSpeedTestForm={setSpeedTestForm}
        groupOptions={groupOptions}
        tagOptions={tagOptions}
        onClose={() => setSpeedTestDialogOpen(false)}
        onSubmit={handleSubmitSpeedTest}
        onRunSpeedTest={handleRunSpeedTest}
        onModeChange={handleSpeedModeChange}
      />

      {/* 策略管理抽屉 */}
      <NodeCheckProfilesDrawer
        open={profilesDrawerOpen}
        onClose={() => setProfilesDrawerOpen(false)}
        groupOptions={groupOptions}
        tagOptions={tagOptions}
        onMessage={showMessage}
      />

      {/* 策略选择对话框 */}
      <ProfileSelectDialog
        open={profileSelectOpen}
        onClose={() => setProfileSelectOpen(false)}
        nodeIds={profileSelectNodeIds}
        onSuccess={showMessage}
        onOpenSettings={() => setProfilesDrawerOpen(true)}
      />

      {/* 批量修改分组对话框 */}
      <BatchGroupDialog
        open={batchGroupDialogOpen}
        selectedCount={selectedNodes.length}
        value={batchGroupValue}
        setValue={setBatchGroupValue}
        groupOptions={groupOptions}
        onClose={() => setBatchGroupDialogOpen(false)}
        onSubmit={handleSubmitBatchGroup}
      />

      {/* 批量修改前置代理对话框 */}
      <BatchDialerProxyDialog
        open={batchDialerProxyDialogOpen}
        selectedCount={selectedNodes.length}
        value={batchDialerProxyValue}
        setValue={setBatchDialerProxyValue}
        proxyNodeOptions={proxyNodeOptions}
        loadingProxyNodes={loadingProxyNodes}
        onClose={() => setBatchDialerProxyDialogOpen(false)}
        onSubmit={handleSubmitBatchDialerProxy}
      />

      {/* 批量设置标签对话框 */}
      <BatchTagDialog
        open={batchTagDialogOpen}
        selectedCount={selectedNodes.length}
        value={batchTagValue}
        setValue={setBatchTagValue}
        tagOptions={tagOptions}
        onClose={() => setBatchTagDialogOpen(false)}
        onSubmit={handleSubmitBatchTag}
      />

      {/* 批量移除标签对话框 */}
      <BatchRemoveTagDialog
        open={batchRemoveTagDialogOpen}
        selectedCount={selectedNodes.length}
        value={batchRemoveTagValue}
        setValue={setBatchRemoveTagValue}
        tagOptions={tagOptions}
        onClose={() => setBatchRemoveTagDialogOpen(false)}
        onSubmit={handleSubmitBatchRemoveTag}
      />

      {/* 批量修改来源对话框 */}
      <BatchSourceDialog
        open={batchSourceDialogOpen}
        selectedCount={selectedNodes.length}
        value={batchSourceValue}
        setValue={setBatchSourceValue}
        sourceOptions={sourceOptions}
        onClose={() => setBatchSourceDialogOpen(false)}
        onSubmit={handleSubmitBatchSource}
      />

      {/* 批量修改国家对话框 */}
      <BatchCountryDialog
        open={batchCountryDialogOpen}
        selectedCount={selectedNodes.length}
        value={batchCountryValue}
        setValue={setBatchCountryValue}
        countryOptions={countryOptions}
        onClose={() => setBatchCountryDialogOpen(false)}
        onSubmit={handleSubmitBatchCountry}
      />

      {/* IP详情弹窗 */}
      <IPDetailsDialog open={ipDialogOpen} onClose={() => setIpDialogOpen(false)} ip={selectedIP} onCopy={copyToClipboard} />

      {/* 节点详情面板 */}
      <NodeDetailsPanel
        open={detailsPanelOpen}
        node={detailsNode}
        tagColorMap={tagColorMap}
        protocolMeta={protocolMeta}
        onClose={() => setDetailsPanelOpen(false)}
        onSpeedTest={handleSingleSpeedTest}
        onCopy={copyToClipboard}
        onEdit={handleEditNode}
        onDelete={handleDeleteNode}
        onIPClick={(ip) => {
          setSelectedIP(ip);
          setIpDialogOpen(true);
        }}
        onNodeUpdate={() => {
          // 节点原始信息更新后刷新列表
          fetchNodes(getCurrentFilters());
        }}
        showMessage={showMessage}
        onOpenRawProtocol={handleOpenRawProtocol}
      />

      {/* 原始协议对话框 */}
      <NodeRawProtocolDialog
        open={rawProtocolDialogOpen}
        node={rawProtocolNode}
        protocolMeta={protocolMeta}
        onClose={handleCloseRawProtocol}
        onUpdate={handleRawProtocolUpdate}
        showMessage={showMessage}
      />

      {/* 添加结果汇总弹窗 */}
      <NodeAddResultDialog
        open={addResultDialogOpen}
        result={addResult}
        onClose={() => {
          setAddResultDialogOpen(false);
          setAddResult(null);
        }}
      />

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
        onConfirm={confirmInfo.action}
      />
    </MainCard>
  );
}
