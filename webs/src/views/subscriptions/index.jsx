import { useState, useEffect, useMemo, useCallback, useRef } from 'react';
import { useTranslation } from 'react-i18next';

// material-ui
import { useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import Stack from '@mui/material/Stack';
import Alert from '@mui/material/Alert';
import Snackbar from '@mui/material/Snackbar';
import Tooltip from '@mui/material/Tooltip';
import TextField from '@mui/material/TextField';
import InputAdornment from '@mui/material/InputAdornment';
import CircularProgress from '@mui/material/CircularProgress';

// icons
import AddIcon from '@mui/icons-material/Add';
import RefreshIcon from '@mui/icons-material/Refresh';
import CategoryIcon from '@mui/icons-material/Category';
import SearchIcon from '@mui/icons-material/Search';
import ClearIcon from '@mui/icons-material/Clear';

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
import { getNodeCheckMeta } from 'api/nodeCheck';
import {
  getNodeSelector,
  getNodeSelectorByIds,
  getNodeGroupStats,
  getNodeCountries,
  getNodeGroups,
  getNodeSources,
  getNodeIds,
  getProtocolUIMeta
} from 'api/nodes';
import { getTemplates } from 'api/templates';
import { getScripts } from 'api/scripts';
import { getTags } from 'api/tags';
import { getShares } from 'api/shares';
import { getAirports } from 'api/airports';
import { buildUnlockRulesPayload, normalizeUnlockRules, setUnlockMeta } from 'views/nodes/utils';
import { getRegisteredProtocolNames } from 'utils/protocolPresentation';
import { getNodeDisplayName } from 'utils/nodeDisplayName';

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
  ChainProxyDialog,
  GroupSortDialog
} from './component';

// ==============================|| 订阅管理 ||============================== //

export default function SubscriptionList() {
  const theme = useTheme();
  const { t } = useTranslation();
  const matchDownMd = useMediaQuery(theme.breakpoints.down('md'));

  const [subscriptions, setSubscriptions] = useState([]);
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
    selectedAirports: [],
    selectedScripts: [],
    IPWhitelist: '',
    IPBlacklist: '',
    DelayTime: 0,
    MinSpeed: 0,
    UpdateInterval: 0,
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
    MaxFraudScore: 0,
    OnlyResidential: false,
    OnlyNative: false,
    ResidentialType: '',
    IPType: '',
    QualityStatus: '',
    UnlockProvider: '',
    UnlockStatus: '',
    UnlockKeyword: '',
    UnlockRuleMode: 'or',
    unlockRules: [],
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

  const [availableNodes, setAvailableNodes] = useState([]);
  const [availableNodesTotal, setAvailableNodesTotal] = useState(0);
  const [availableNodesLoading, setAvailableNodesLoading] = useState(false);
  const [selectedNodeMap, setSelectedNodeMap] = useState({});
  const [groupNodeCounts, setGroupNodeCounts] = useState({});
  const [allNodeTotal, setAllNodeTotal] = useState(0);

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
  const [subscriptionSearch, setSubscriptionSearch] = useState('');
  const [shareSearchResults, setShareSearchResults] = useState([]);
  const [shareSearching, setShareSearching] = useState(false);
  const [shareSearchActive, setShareSearchActive] = useState(false);
  const shareSearchRequestRef = useRef(0);

  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });

  const [groupOptions, setGroupOptions] = useState([]);
  const [airportOptions, setAirportOptions] = useState([]);
  const [sourceOptions, setSourceOptions] = useState([]);
  const [tagOptions, setTagOptions] = useState([]);
  const [protocolOptions, setProtocolOptions] = useState([]);

  const getAirportId = useCallback((airport) => Number(airport?.id ?? airport?.ID), []);
  const getAirportName = useCallback((airport) => airport?.name || airport?.Name || '', []);
  const normalizeAirportList = useCallback((data) => {
    const items = data?.items || (Array.isArray(data) ? data : []);
    return items
      .map((airport) => ({ ...airport, id: Number(airport?.id ?? airport?.ID), name: airport?.name || airport?.Name || '' }))
      .filter((airport) => Number.isInteger(airport.id) && airport.id > 0 && airport.name);
  }, []);

  const buildNodeFilterParams = useCallback(
    () => ({
      search: nodeSearchQuery,
      group: nodeGroupFilter === 'all' ? '' : nodeGroupFilter,
      source: nodeSourceFilter === 'all' ? '' : nodeSourceFilter,
      'countries[]': nodeCountryFilter,
      'excludeIds[]': formData.selectedNodes
    }),
    [formData.selectedNodes, nodeCountryFilter, nodeGroupFilter, nodeSearchQuery, nodeSourceFilter]
  );

  const buildSelectorParams = useCallback(
    () => ({
      ...buildNodeFilterParams(),
      page: 1,
      pageSize: 100
    }),
    [buildNodeFilterParams]
  );

  const hydrateSelectedNodeMap = useCallback((items) => {
    if (!Array.isArray(items) || items.length === 0) return;
    setSelectedNodeMap((prev) => {
      const next = { ...prev };
      items.forEach((item) => {
        next[item.ID] = item;
      });
      return next;
    });
  }, []);

  const fetchNodeSelector = useCallback(async () => {
    setAvailableNodesLoading(true);
    try {
      const response = await getNodeSelector(buildSelectorParams());
      const items = response.data?.items || [];
      setAvailableNodes(items);
      setAvailableNodesTotal(response.data?.total || items.length);
      hydrateSelectedNodeMap(items);
    } catch (error) {
      console.error(error);
    } finally {
      setAvailableNodesLoading(false);
    }
  }, [buildSelectorParams, hydrateSelectedNodeMap]);

  const extractShareSearchKeyword = useCallback((input) => {
    const trimmed = input.trim();
    if (!trimmed) return '';

    const tokenMatch = trimmed.match(/[?&]token=([^&\s]+)/);
    if (tokenMatch?.[1]) return tokenMatch[1];

    const pathTokenMatch = trimmed.match(/(?:^|\/)c\/([^/?&#\s]+)/);
    return pathTokenMatch?.[1] || trimmed;
  }, []);

  const isShareLookupQuery = useCallback((input) => {
    const trimmed = input.trim();
    if (!trimmed) return false;
    if (/^https?:\/\//i.test(trimmed) || trimmed.includes('/c/') || /[?&]token=/.test(trimmed)) return true;
    return /^[A-Za-z0-9_-]{8,}$/.test(trimmed);
  }, []);

  const normalizeSubscriptionResponse = useCallback((response) => {
    if (response.data && response.data.items !== undefined) {
      return response.data.items || [];
    }
    return response.data || [];
  }, []);

  const fetchAllSubscriptionsForSearch = useCallback(async () => {
    if (totalItems <= subscriptions.length) return subscriptions;
    const response = await getSubscriptions({ page: 1, pageSize: Math.max(totalItems, rowsPerPage) });
    return normalizeSubscriptionResponse(response);
  }, [normalizeSubscriptionResponse, rowsPerPage, subscriptions, totalItems]);

  const refreshNodeSelector = useCallback(() => {
    if (!dialogOpen || formData.selectionMode === 'groups') return;
    void fetchNodeSelector();
  }, [dialogOpen, fetchNodeSelector, formData.selectionMode]);

  const trimmedSubscriptionSearch = subscriptionSearch.trim();

  const nameFilteredSubscriptions = useMemo(() => {
    if (!trimmedSubscriptionSearch) return subscriptions;
    const keyword = trimmedSubscriptionSearch.toLowerCase();
    return subscriptions.filter((sub) => (sub.Name || '').toLowerCase().includes(keyword));
  }, [subscriptions, trimmedSubscriptionSearch]);

  const displayedSubscriptions = trimmedSubscriptionSearch && shareSearchActive ? shareSearchResults : nameFilteredSubscriptions;

  const fetchSelectedNodeDetails = useCallback(
    async (ids) => {
      const validIds = Array.from(new Set((ids || []).map((id) => Number(id)).filter((id) => Number.isInteger(id) && id > 0)));
      if (validIds.length === 0) return;
      const missingIds = validIds.filter((id) => !selectedNodeMap[id]);
      if (missingIds.length === 0) return;
      try {
        const response = await getNodeSelectorByIds({ 'ids[]': missingIds });
        hydrateSelectedNodeMap(response.data || []);
      } catch (error) {
        console.error(error);
      }
    },
    [hydrateSelectedNodeMap, selectedNodeMap]
  );

  const syncSelectedNodeMapFromList = useCallback((nodes) => {
    if (!Array.isArray(nodes) || nodes.length === 0) return;
    setSelectedNodeMap((prev) => {
      const next = { ...prev };
      nodes.forEach((node) => {
        next[node.ID] = {
          ID: node.ID,
          Name: getNodeDisplayName(node),
          Group: node.Group,
          Source: node.Source,
          LinkCountry: node.LinkCountry,
          UnlockSummary: node.UnlockSummary,
          UnlockCheckAt: node.UnlockCheckAt
        };
      });
      return next;
    });
  }, []);

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
      showMessage(error.message || t('subscriptions.page.messages.loadFailed'), 'error');
    } finally {
      setLoading(false);
    }
  };

  // 获取其他数据（分层加载）
  const fetchOtherData = useCallback(async () => {
    try {
      const [
        templatesRes,
        scriptsRes,
        countriesRes,
        groupsRes,
        airportsRes,
        sourcesRes,
        tagsRes,
        protocolMetaRes,
        nodeCheckMetaRes,
        groupStatsRes
      ] = await Promise.all([
        getTemplates(),
        getScripts(),
        getNodeCountries(),
        getNodeGroups(),
        getAirports(),
        getNodeSources(),
        getTags(),
        getProtocolUIMeta(),
        getNodeCheckMeta(),
        getNodeGroupStats()
      ]);
      setTemplates(templatesRes.data || []);
      setScripts(scriptsRes.data || []);
      setCountryOptions(countriesRes.data || []);
      setGroupOptions((groupsRes.data || []).sort());
      setAirportOptions(normalizeAirportList(airportsRes.data));
      setSourceOptions((sourcesRes.data || []).sort());
      setTagOptions(tagsRes.data || []);
      setProtocolOptions(getRegisteredProtocolNames(protocolMetaRes.data || []));
      setUnlockMeta(nodeCheckMetaRes.data || {});
      const counts = groupStatsRes.data || {};
      setGroupNodeCounts(counts);
      setAllNodeTotal(Object.values(counts).reduce((sum, value) => sum + Number(value || 0), 0));
    } catch (error) {
      console.error(error);
    }
  }, [normalizeAirportList]);

  // 初始加载
  useEffect(() => {
    fetchSubscriptions(0, rowsPerPage);
    fetchOtherData();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    refreshNodeSelector();
  }, [refreshNodeSelector, nodeGroupFilter, nodeSourceFilter, nodeSearchQuery, nodeCountryFilter, formData.selectedNodes]);

  const showMessage = useCallback((message, severity = 'success') => {
    setSnackbar({ open: true, message, severity });
  }, []);

  useEffect(() => {
    const query = subscriptionSearch.trim();
    const requestId = shareSearchRequestRef.current + 1;
    shareSearchRequestRef.current = requestId;

    if (!query) {
      setShareSearchActive(false);
      setShareSearchResults([]);
      setShareSearching(false);
      return undefined;
    }

    const shouldSearchShares = isShareLookupQuery(query) || nameFilteredSubscriptions.length === 0;
    if (!shouldSearchShares) {
      setShareSearchActive(false);
      setShareSearchResults([]);
      setShareSearching(false);
      return undefined;
    }

    setShareSearchActive(true);
    setShareSearchResults([]);
    setShareSearching(true);

    const timeoutId = window.setTimeout(async () => {
      const keyword = extractShareSearchKeyword(query);
      try {
        const allSubscriptions = await fetchAllSubscriptionsForSearch();
        const checks = await Promise.all(
          allSubscriptions.map(async (sub) => {
            const response = await getShares(sub.ID, 1, 100, keyword);
            const shares = response.data?.items || response.data || [];
            return shares.length > 0 ? sub : null;
          })
        );

        if (shareSearchRequestRef.current === requestId) {
          setShareSearchResults(checks.filter(Boolean));
        }
      } catch (error) {
        console.error(error);
        if (shareSearchRequestRef.current === requestId) {
          showMessage(error.message || t('subscriptions.page.messages.shareSearchFailed'), 'error');
          setShareSearchResults([]);
        }
      } finally {
        if (shareSearchRequestRef.current === requestId) {
          setShareSearching(false);
        }
      }
    }, 500);

    return () => window.clearTimeout(timeoutId);
  }, [
    extractShareSearchKeyword,
    fetchAllSubscriptionsForSearch,
    isShareLookupQuery,
    nameFilteredSubscriptions.length,
    showMessage,
    subscriptionSearch,
    t
  ]);

  const copyToClipboard = async (text) => {
    try {
      // 优先使用现代 Clipboard API
      if (navigator.clipboard && window.isSecureContext) {
        await navigator.clipboard.writeText(text);
        showMessage(t('common.copied'));
        return;
      }
      // 备用方案：使用传统的 execCommand
      const textArea = document.createElement('textarea');
      textArea.value = text;
      textArea.style.position = 'fixed';
      textArea.style.left = '-999999px';
      textArea.style.top = '-999999px';
      document.body.appendChild(textArea);
      textArea.focus();
      textArea.select();
      const successful = document.execCommand('copy');
      document.body.removeChild(textArea);
      if (successful) {
        showMessage(t('common.copied'));
      } else {
        showMessage(t('common.copyFailedManual'), 'error');
      }
    } catch (error) {
      console.error('复制失败:', error);
      showMessage(t('common.copyFailedManual'), 'error');
    }
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
      selectedAirports: [],
      selectedScripts: [],
      IPWhitelist: '',
      IPBlacklist: '',
      DelayTime: 0,
      MinSpeed: 0,
      UpdateInterval: 0,
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
      MaxFraudScore: 0,
      OnlyResidential: false,
      OnlyNative: false,
      ResidentialType: '',
      IPType: '',
      QualityStatus: '',
      UnlockProvider: '',
      UnlockStatus: '',
      UnlockKeyword: '',
      UnlockRuleMode: 'or',
      unlockRules: [],
      refreshUsageOnRequest: true
    });
    setNodeGroupFilter('all');
    setNodeSourceFilter('all');
    setNodeSearchQuery('');
    setNodeCountryFilter([]);
    setAvailableNodes([]);
    setAvailableNodesTotal(0);
    setSelectedNodeMap({});
    setDialogOpen(true);
    refreshNodeSelector();
  };

  const handleEdit = (sub) => {
    setIsEdit(true);
    setCurrentSub(sub);
    const config = typeof sub.Config === 'string' ? JSON.parse(sub.Config) : sub.Config;
    const parsedUnlockRules = (() => {
      if (!sub.UnlockRules) return [];
      try {
        return normalizeUnlockRules(typeof sub.UnlockRules === 'string' ? JSON.parse(sub.UnlockRules) : sub.UnlockRules);
      } catch {
        return [];
      }
    })();

    const nodes = sub.Nodes?.map((n) => n.ID) || [];
    const groups = (sub.Groups || []).map((g) => (typeof g === 'string' ? g : g.Name));
    const airports = (sub.Airports || []).map((airport) => getAirportId(airport)).filter((id) => Number.isInteger(id) && id > 0);
    const scriptIds = (sub.Scripts || []).map((s) => s.id);

    let mode = 'nodes';
    if (nodes.length > 0 && (groups.length > 0 || airports.length > 0)) {
      mode = 'mixed';
    } else if (groups.length > 0 || airports.length > 0) {
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
      selectedAirports: airports,
      selectedScripts: scriptIds,
      IPWhitelist: sub.IPWhitelist || '',
      IPBlacklist: sub.IPBlacklist || '',
      DelayTime: sub.DelayTime || 0,
      MinSpeed: sub.MinSpeed || 0,
      UpdateInterval: sub.UpdateInterval || 0,
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
      MaxFraudScore: sub.MaxFraudScore || 0,
      OnlyResidential: sub.OnlyResidential || false,
      OnlyNative: sub.OnlyNative || false,
      ResidentialType: sub.ResidentialType || (sub.OnlyResidential ? 'residential' : ''),
      IPType: sub.IPType || (sub.OnlyNative ? 'native' : ''),
      QualityStatus: sub.QualityStatus || '',
      UnlockProvider: sub.UnlockProvider || '',
      UnlockStatus: sub.UnlockStatus || '',
      UnlockKeyword: sub.UnlockKeyword || '',
      UnlockRuleMode: sub.UnlockRuleMode || 'or',
      unlockRules:
        parsedUnlockRules.length > 0
          ? parsedUnlockRules
          : sub.UnlockProvider || sub.UnlockStatus || sub.UnlockKeyword
            ? [{ provider: sub.UnlockProvider || '', status: sub.UnlockStatus || '', keyword: sub.UnlockKeyword || '' }]
            : [],
      refreshUsageOnRequest: sub.RefreshUsageOnRequest !== false // 默认 true
    });
    setNodeGroupFilter('all');
    setNodeSourceFilter('all');
    setNodeSearchQuery('');
    setNodeCountryFilter([]);
    syncSelectedNodeMapFromList(sub.Nodes || []);
    setDialogOpen(true);
    refreshNodeSelector();
  };

  const handleDelete = async (sub) => {
    openConfirm(t('subscriptions.page.confirm.deleteTitle'), t('subscriptions.page.confirm.deleteOne', { name: sub.Name }), async () => {
      try {
        await deleteSubscription({ id: sub.ID });
        showMessage(t('subscriptions.page.messages.deleteSuccess'));
        fetchSubscriptions(page, rowsPerPage);
      } catch (error) {
        console.error(error);
        showMessage(error.message || t('subscriptions.page.messages.deleteFailed'), 'error');
      }
    });
  };

  // 复制订阅
  const handleCopy = async (sub) => {
    openConfirm(t('subscriptions.page.confirm.copyTitle'), t('subscriptions.page.confirm.copyOne', { name: sub.Name }), async () => {
      try {
        await copySubscription(sub.ID);
        showMessage(t('subscriptions.page.messages.copySuccess'));
        fetchSubscriptions(page, rowsPerPage);
      } catch (error) {
        console.error(error);
        showMessage(error.message || t('subscriptions.page.messages.copyFailed'), 'error');
      }
    });
  };

  const handleSubmit = async () => {
    if (!formData.name.trim()) {
      showMessage(t('subscriptions.page.messages.nameRequired'), 'warning');
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
        UpdateInterval: Math.min(8760, Math.max(0, Number(formData.UpdateInterval) || 0)),
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
        MaxFraudScore: formData.MaxFraudScore,
        OnlyResidential: formData.ResidentialType === 'residential',
        OnlyNative: formData.IPType === 'native',
        ResidentialType: formData.ResidentialType || '',
        IPType: formData.IPType || '',
        QualityStatus: formData.QualityStatus || '',
        UnlockProvider: '',
        UnlockStatus: '',
        UnlockKeyword: '',
        UnlockRuleMode: formData.UnlockRuleMode || 'or',
        UnlockRules: buildUnlockRulesPayload(formData.unlockRules),
        RefreshUsageOnRequest: formData.refreshUsageOnRequest
      };

      if (formData.selectionMode === 'nodes') {
        requestData.nodeIds = formData.selectedNodes.join(',');
        requestData.groups = '';
        requestData.airports = '';
      } else if (formData.selectionMode === 'groups') {
        requestData.nodeIds = '';
        requestData.groups = formData.selectedGroups.join(',');
        requestData.airports = formData.selectedAirports.join(',');
      } else {
        requestData.nodeIds = formData.selectedNodes.join(',');
        requestData.groups = formData.selectedGroups.join(',');
        requestData.airports = formData.selectedAirports.join(',');
      }

      if (isEdit) {
        requestData.oldname = currentSub.Name;
        await updateSubscription(requestData);
        showMessage(t('subscriptions.page.messages.updateSuccess'));
      } else {
        await addSubscription(requestData);
        showMessage(t('subscriptions.page.messages.addSuccess'));
      }
      setDialogOpen(false);
      fetchSubscriptions(page, rowsPerPage);
    } catch (error) {
      console.error(error);
      showMessage(
        error.message || (isEdit ? t('subscriptions.page.messages.updateFailed') : t('subscriptions.page.messages.addFailed')),
        'error'
      );
    }
  };

  // 节点选择操作（使用 node.ID）
  const handleAddNode = (nodeId) => {
    if (formData.selectedNodes.includes(nodeId)) return;
    const nextSelected = [...formData.selectedNodes, nodeId];
    setFormData({ ...formData, selectedNodes: nextSelected });
    setCheckedAvailable(checkedAvailable.filter((id) => id !== nodeId));
    void fetchSelectedNodeDetails(nextSelected);
  };

  const handleRemoveNode = (nodeId) => {
    setFormData({ ...formData, selectedNodes: formData.selectedNodes.filter((id) => id !== nodeId) });
    setCheckedSelected(checkedSelected.filter((id) => id !== nodeId));
  };

  const selectedNodesList = useMemo(() => {
    return formData.selectedNodes.map((id) => selectedNodeMap[id]).filter(Boolean);
  }, [formData.selectedNodes, selectedNodeMap]);

  const availableNodeCount = availableNodesTotal || availableNodes.length;

  const handleAddAllVisible = async () => {
    try {
      const response = await getNodeIds(buildNodeFilterParams());
      const matchedIds = response.data || [];
      const newNodes = Array.from(new Set([...formData.selectedNodes, ...matchedIds]));
      setFormData({ ...formData, selectedNodes: newNodes });
      setCheckedAvailable([]);
      await fetchSelectedNodeDetails(newNodes);
    } catch (error) {
      console.error(error);
      showMessage(error.message || '批量添加节点失败', 'error');
    }
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
    const newNodes = Array.from(new Set([...formData.selectedNodes, ...checkedAvailable]));
    setFormData({ ...formData, selectedNodes: newNodes });
    setCheckedAvailable([]);
    void fetchSelectedNodeDetails(newNodes);
  };

  const handleRemoveChecked = () => {
    const newNodes = formData.selectedNodes.filter((id) => !checkedSelected.includes(id));
    setFormData({ ...formData, selectedNodes: newNodes });
    setCheckedSelected([]);
  };

  const handleToggleAllAvailable = async () => {
    if (checkedAvailable.length === availableNodeCount && availableNodeCount > 0) {
      setCheckedAvailable([]);
      return;
    }
    try {
      const response = await getNodeIds(buildNodeFilterParams());
      setCheckedAvailable(response.data || []);
    } catch (error) {
      console.error(error);
      showMessage(error.message || '批量选择节点失败', 'error');
    }
  };

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
        NodeIDs: formData.selectionMode !== 'groups' ? formData.selectedNodes : [],
        Groups: formData.selectionMode !== 'nodes' ? formData.selectedGroups : [],
        AirportIDs: formData.selectionMode !== 'nodes' ? formData.selectedAirports : [],
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
        MaxFraudScore: formData.MaxFraudScore || 0,
        OnlyResidential: formData.ResidentialType === 'residential',
        OnlyNative: formData.IPType === 'native',
        ResidentialType: formData.ResidentialType || '',
        IPType: formData.IPType || '',
        QualityStatus: formData.QualityStatus || '',
        UnlockProvider: '',
        UnlockStatus: '',
        UnlockKeyword: '',
        UnlockRuleMode: formData.UnlockRuleMode || 'or',
        UnlockRules: buildUnlockRulesPayload(formData.unlockRules),
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
  // 使用 SubscriptionID 让后端直接调用 GetSub 逻辑，确保预览与实际拉取结果一致
  const handlePreviewSubscription = async (sub) => {
    setPreviewLoading(true);
    try {
      const previewRequest = {
        SubscriptionID: sub.ID // 使用订阅ID，后端会调用 GetSub 获取完整节点列表
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

  // === 分组排序 ===
  const [groupSortOpen, setGroupSortOpen] = useState(false);

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
        Name: getNodeDisplayName(node),
        SortKey: `node:${node.ID}`,
        Sort: node.Sort !== undefined ? node.Sort : idx,
        IsGroup: false,
        IsAirport: false
      });
    });
    (sub.Groups || []).forEach((group, idx) => {
      const g = typeof group === 'string' ? { Name: group, Sort: sub.Nodes?.length + idx } : group;
      sortData.push({
        ID: 0,
        Name: g.Name,
        SortKey: `group:${g.Name}`,
        Sort: g.Sort !== undefined ? g.Sort : sub.Nodes?.length + idx,
        IsGroup: true,
        IsAirport: false
      });
    });
    (sub.Airports || []).forEach((airport, idx) => {
      const airportId = getAirportId(airport);
      const fallbackSort = (sub.Nodes?.length || 0) + (sub.Groups?.length || 0) + idx;
      sortData.push({
        ID: airportId,
        Name: getAirportName(airport),
        SortKey: `airport:${airportId}`,
        Sort: airport.Sort !== undefined ? airport.Sort : fallbackSort,
        IsGroup: false,
        IsAirport: true
      });
    });
    sortData.sort((a, b) => a.Sort - b.Sort);
    setTempSortData(sortData);
    showMessage(t('subscriptions.page.messages.sortModeStarted'), 'info');
  };

  const handleConfirmSort = async (sub) => {
    const newSortData = tempSortData.map((item, idx) => ({ ...item, Sort: idx }));
    try {
      await sortSubscription({
        ID: sub.ID,
        NodeSort: newSortData
      });
      showMessage(t('subscriptions.page.messages.sortUpdated'));
      setSortingSubId(null);
      setTempSortData([]);
      fetchSubscriptions(page, rowsPerPage);
    } catch (error) {
      console.error(error);
      showMessage(error.message || t('subscriptions.page.messages.sortSaveFailed'), 'error');
    }
  };

  const handleCancelSort = () => {
    setSortingSubId(null);
    setTempSortData([]);
    setSelectedSortItems([]);
    showMessage(t('subscriptions.page.messages.sortCancelled'), 'info');
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
    setSelectedSortItems(tempSortData.map((item) => item.SortKey || item.Name));
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
      showMessage(t('subscriptions.page.messages.batchSortSuccess'));
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
      showMessage(error.message || t('subscriptions.page.messages.batchSortFailed'), 'error');
    }
  };

  // 批量移动（前端本地操作，需要用户确认后保存）
  const handleBatchMove = (targetIndex) => {
    if (selectedSortItems.length === 0) return;

    const selected = tempSortData.filter((item) => selectedSortItems.includes(item.SortKey || item.Name));
    const remaining = tempSortData.filter((item) => !selectedSortItems.includes(item.SortKey || item.Name));

    // 插入到目标位置
    const newData = [...remaining];
    const insertAt = Math.min(Math.max(0, targetIndex), newData.length);
    newData.splice(insertAt, 0, ...selected);

    setTempSortData(newData);
    setSelectedSortItems([]);
    showMessage(t('subscriptions.page.messages.movedItems', { count: selected.length, position: insertAt + 1 }));
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
    (sub.Airports || []).forEach((airport, idx) => {
      const fallbackSort = (sub.Nodes?.length || 0) + (sub.Groups?.length || 0) + idx;
      items.push({
        ...airport,
        ID: getAirportId(airport),
        Name: getAirportName(airport),
        _type: 'airport',
        _sort: airport.Sort !== undefined ? airport.Sort : fallbackSort
      });
    });
    return items.sort((a, b) => a._sort - b._sort);
  };

  return (
    <MainCard
      title={t('subscriptions.page.title')}
      secondary={
        matchDownMd ? (
          <Stack direction="row" spacing={1}>
            <Tooltip title={t('subscriptions.page.actions.groupSort')}>
              <IconButton onClick={() => setGroupSortOpen(true)} size="small">
                <CategoryIcon />
              </IconButton>
            </Tooltip>
            <Button variant="contained" startIcon={<AddIcon />} onClick={handleAdd}>
              {t('common.add')}
            </Button>
          </Stack>
        ) : (
          <Stack direction="row" spacing={1}>
            <Button variant="outlined" startIcon={<CategoryIcon />} onClick={() => setGroupSortOpen(true)}>
              {t('subscriptions.page.actions.groupSort')}
            </Button>
            <Button variant="contained" startIcon={<AddIcon />} onClick={handleAdd}>
              {t('subscriptions.page.actions.addSubscription')}
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

      <Stack spacing={1} sx={{ mb: 2 }}>
        <TextField
          size="small"
          fullWidth={matchDownMd}
          label={t('subscriptions.page.search.label')}
          placeholder={t('subscriptions.page.search.placeholder')}
          value={subscriptionSearch}
          onChange={(event) => setSubscriptionSearch(event.target.value)}
          sx={{ width: { xs: '100%', md: 460 } }}
          InputProps={{
            startAdornment: (
              <InputAdornment position="start">
                <SearchIcon color="action" fontSize="small" />
              </InputAdornment>
            ),
            endAdornment: trimmedSubscriptionSearch && (
              <InputAdornment position="end">
                {shareSearching ? (
                  <CircularProgress size={18} />
                ) : (
                  <IconButton size="small" onClick={() => setSubscriptionSearch('')} edge="end" aria-label={t('common.clear')}>
                    <ClearIcon fontSize="small" />
                  </IconButton>
                )}
              </InputAdornment>
            )
          }}
        />
      </Stack>

      {matchDownMd ? (
        <SubscriptionMobileCard
          subscriptions={displayedSubscriptions}
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
          subscriptions={displayedSubscriptions}
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

      {!trimmedSubscriptionSearch && (
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
      )}

      {/* 添加/编辑对话框 */}
      <SubscriptionFormDialog
        open={dialogOpen}
        isEdit={isEdit}
        formData={formData}
        setFormData={setFormData}
        templates={templates}
        scripts={scripts}
        selectorNodes={availableNodes}
        selectorNodesTotal={availableNodeCount}
        selectorNodesLoading={availableNodesLoading}
        selectedNodesList={selectedNodesList}
        selectedNodeSearch={selectedNodeSearch}
        setSelectedNodeSearch={setSelectedNodeSearch}
        groupNodeCounts={groupNodeCounts}
        allNodeTotal={allNodeTotal}
        groupOptions={groupOptions}
        airportOptions={airportOptions}
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

      {/* 分组排序对话框 */}
      <GroupSortDialog open={groupSortOpen} onClose={() => setGroupSortOpen(false)} showMessage={showMessage} />

      {/* 链式代理配置对话框 */}
      <ChainProxyDialog open={chainProxyOpen} subscription={chainProxySub} onClose={() => setChainProxyOpen(false)} />
    </MainCard>
  );
}
