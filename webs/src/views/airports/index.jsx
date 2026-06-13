import { useState, useEffect, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';

// material-ui
import { useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Alert from '@mui/material/Alert';
import Autocomplete from '@mui/material/Autocomplete';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Checkbox from '@mui/material/Checkbox';
import Chip from '@mui/material/Chip';
import FormControl from '@mui/material/FormControl';
import IconButton from '@mui/material/IconButton';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import Paper from '@mui/material/Paper';
import Select from '@mui/material/Select';
import Snackbar from '@mui/material/Snackbar';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import ToggleButton from '@mui/material/ToggleButton';
import ToggleButtonGroup from '@mui/material/ToggleButtonGroup';
import Typography from '@mui/material/Typography';

// icons
import AddIcon from '@mui/icons-material/Add';
import CloudSyncIcon from '@mui/icons-material/CloudSync';
import RefreshIcon from '@mui/icons-material/Refresh';
import TuneIcon from '@mui/icons-material/Tune';
import ViewModuleIcon from '@mui/icons-material/ViewModule';
import ViewListIcon from '@mui/icons-material/ViewList';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import Pagination from 'components/Pagination';
import ConfirmDialog from 'components/ConfirmDialog';
import TaskProgressPanel from 'components/TaskProgressPanel';
import {
  getAirports,
  addAirport,
  updateAirport,
  batchUpdateAirports,
  deleteAirport,
  pullAirport,
  pullAllAirports,
  refreshAirportUsage
} from 'api/airports';
import { getNodeCheckProfiles } from 'api/nodeCheck';
import { useTaskProgress } from 'contexts/TaskProgressContext';
import { getNodeGroups, getNodeIds, getNodes, getProtocolUIMeta } from 'api/nodes';
import ProfileSelectDialog from 'views/nodes/component/ProfileSelectDialog';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getRegisteredProtocolNames } from 'utils/protocolPresentation';

// local components
import {
  AirportTable,
  AirportListView,
  AirportMobileList,
  AirportFormDialog,
  DeleteAirportDialog,
  AirportBatchEditDialog
} from './component';

// utils
import { withAlpha } from 'utils/colorUtils';
import { validateCronExpression } from './utils';

// ==============================|| 机场管理 ||============================== //

const createEmptyRequestHeader = () => ({ key: '', value: '' });

const normalizeRequestHeadersForForm = (requestHeaders) => {
  if (!Array.isArray(requestHeaders)) {
    return [];
  }

  return requestHeaders
    .map((header) => ({
      key: typeof header?.key === 'string' ? header.key : `${header?.key ?? ''}`,
      value: typeof header?.value === 'string' ? header.value : `${header?.value ?? ''}`
    }))
    .filter((header) => header.key.trim() || header.value.trim());
};

const getRequestHeaderValidationMessage = (requestHeader, t) => {
  const key = `${requestHeader?.key ?? ''}`.trim();
  const value = `${requestHeader?.value ?? ''}`.trim();

  if (!key && !value) {
    return '';
  }

  if (!key && value) {
    return t('airports.form.requestHeaders.errors.missingKey');
  }

  if (key.toLowerCase() === 'user-agent') {
    return t('airports.form.requestHeaders.errors.userAgentDedicated');
  }

  return '';
};

const normalizeRequestHeadersForSubmit = (requestHeaders, t) => {
  const normalizedHeaders = Array.isArray(requestHeaders)
    ? requestHeaders
        .map((header) => ({
          key: `${header?.key ?? ''}`.trim(),
          value: `${header?.value ?? ''}`.trim()
        }))
        .filter((header) => header.key || header.value)
    : [];

  const invalidHeaderIndex = normalizedHeaders.findIndex((header) => getRequestHeaderValidationMessage(header, t));

  if (invalidHeaderIndex !== -1) {
    return {
      requestHeaders: normalizedHeaders,
      error: t('airports.page.messages.invalidRequestHeaderRow', {
        row: invalidHeaderIndex + 1,
        message: getRequestHeaderValidationMessage(normalizedHeaders[invalidHeaderIndex], t)
      })
    };
  }

  return {
    requestHeaders: normalizedHeaders,
    error: ''
  };
};

const createAirportFormState = (overrides = {}) => ({
  id: 0,
  name: '',
  url: '',
  cronExpr: '0 */12 * * *',
  enabled: true,
  group: '',
  downloadWithProxy: false,
  proxyLink: '',
  userAgent: '',
  requestHeaders: [],
  fetchUsageInfo: false,
  skipTLSVerify: false,
  updateAfterDetect: false,
  updateAfterDetectProfileId: 0,
  updateAfterDetectChangedOnly: false,
  remark: '',
  logo: '',
  nodeNameWhitelist: '',
  nodeNameBlacklist: '',
  protocolWhitelist: '',
  protocolBlacklist: '',
  nodeNamePreprocess: '',
  deduplicationRule: '',
  nodeNameUniquify: false,
  nodeNamePrefix: '',
  nodeNameIntraUniquify: false,
  ...overrides
});

export default function AirportList() {
  const { t } = useTranslation();
  const theme = useTheme();
  const palette = theme.vars?.palette || theme.palette;
  const { isDark } = useResolvedColorScheme();
  const matchDownMd = useMediaQuery(theme.breakpoints.down('md'));
  const navigate = useNavigate();

  // 判断是否需要重新拉取才能使节点处理配置生效
  const hasNodeProcessConfigChanged = (before, after) => {
    if (!before || !after) return false;

    const normalizeStr = (v) => (v ?? '').toString().trim();

    const stringKeys = [
      'nodeNameWhitelist',
      'nodeNameBlacklist',
      'protocolWhitelist',
      'protocolBlacklist',
      'nodeNamePreprocess',
      'deduplicationRule',
      'nodeNamePrefix'
    ];
    for (const key of stringKeys) {
      if (normalizeStr(before[key]) !== normalizeStr(after[key])) return true;
    }

    const boolKeys = ['nodeNameUniquify', 'nodeNameIntraUniquify'];
    for (const key of boolKeys) {
      if (!!before[key] !== !!after[key]) return true;
    }

    return false;
  };

  const createBatchFormState = () => ({
    applyGroup: false,
    group: '',
    applySchedule: false,
    cronExpr: '0 */12 * * *'
  });

  // 数据状态
  const [airports, setAirports] = useState([]);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(() => {
    const saved = localStorage.getItem('airports_rowsPerPage');
    return saved ? parseInt(saved, 10) : 10;
  });
  const [totalItems, setTotalItems] = useState(0);

  // 视图模式状态（card: 卡片，list: 列表）
  const [viewMode, setViewMode] = useState(() => {
    const saved = localStorage.getItem('airports_viewMode');
    return saved || 'card';
  });

  // 批量选择与编辑状态
  const [selectedAirportIds, setSelectedAirportIds] = useState([]);
  const [batchDialogOpen, setBatchDialogOpen] = useState(false);
  const [batchSubmitting, setBatchSubmitting] = useState(false);
  const [selectingFiltered, setSelectingFiltered] = useState(false);
  const [batchForm, setBatchForm] = useState(createBatchFormState);

  // 表单状态
  const [formOpen, setFormOpen] = useState(false);
  const [isEdit, setIsEdit] = useState(false);
  const [airportForm, setAirportForm] = useState(() => createAirportFormState());
  const [airportFormSnapshot, setAirportFormSnapshot] = useState(null);

  // 搜索筛选状态
  const [searchKeyword, setSearchKeyword] = useState('');
  const [searchGroup, setSearchGroup] = useState('');
  const [searchEnabled, setSearchEnabled] = useState('');

  // 删除对话框状态
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState(null);
  const [deleteWithNodes, setDeleteWithNodes] = useState(true);

  // 确认对话框状态
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [confirmInfo, setConfirmInfo] = useState({ title: '', content: '', action: null });

  // 选项数据
  const [groupOptions, setGroupOptions] = useState([]);
  const [proxyNodeOptions, setProxyNodeOptions] = useState([]);
  const [loadingProxyNodes, setLoadingProxyNodes] = useState(false);
  const [protocolOptions, setProtocolOptions] = useState([]);
  const [nodeCheckProfiles, setNodeCheckProfiles] = useState([]);

  const [profileSelectOpen, setProfileSelectOpen] = useState(false);
  const [profileSelectNodeIds, setProfileSelectNodeIds] = useState([]);

  // 消息提示
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });

  // 显示消息
  const showMessage = useCallback((message, severity = 'success') => {
    setSnackbar({ open: true, message, severity });
  }, []);

  // 构建机场查询参数
  const buildAirportQueryParams = useCallback(
    ({ includePagination = false } = {}) => {
      const params = {};

      if (includePagination) {
        params.page = page + 1;
        params.pageSize = rowsPerPage;
      }
      if (searchKeyword) params.keyword = searchKeyword;
      if (searchGroup) params.group = searchGroup;
      if (searchEnabled !== '') params.enabled = searchEnabled;

      return params;
    },
    [page, rowsPerPage, searchKeyword, searchGroup, searchEnabled]
  );

  // 获取机场列表
  const fetchAirports = useCallback(async () => {
    setLoading(true);
    try {
      const response = await getAirports(buildAirportQueryParams({ includePagination: true }));
      if (response.data?.items) {
        setAirports(response.data.items);
        setTotalItems(response.data.total || 0);
      } else if (Array.isArray(response.data)) {
        setAirports(response.data);
        setTotalItems(response.data.length);
      }
    } catch (error) {
      console.error('获取机场列表失败:', error);
      showMessage(error.message || t('airports.page.messages.loadFailed'), 'error');
    } finally {
      setLoading(false);
    }
  }, [buildAirportQueryParams, showMessage, t]);

  // 获取分组选项
  const fetchGroupOptions = useCallback(async () => {
    try {
      const response = await getNodeGroups();
      setGroupOptions((response.data || []).sort());
    } catch (error) {
      console.error('获取分组选项失败:', error);
    }
  }, []);

  // 获取代理节点
  const fetchProxyNodes = useCallback(async () => {
    setLoadingProxyNodes(true);
    try {
      const response = await getNodes({});
      setProxyNodeOptions(response.data || []);
    } catch (error) {
      console.error('获取代理节点失败:', error);
    } finally {
      setLoadingProxyNodes(false);
    }
  }, []);

  // 获取协议列表
  const fetchProtocolOptions = useCallback(async () => {
    try {
      const response = await getProtocolUIMeta();
      setProtocolOptions(getRegisteredProtocolNames(response.data || []));
    } catch (error) {
      console.error('获取协议列表失败:', error);
    }
  }, []);

  // 获取节点检测策略
  const fetchNodeCheckProfiles = useCallback(async () => {
    try {
      const response = await getNodeCheckProfiles();
      setNodeCheckProfiles(response.data || []);
    } catch (error) {
      console.error('获取节点检测策略失败:', error);
    }
  }, []);

  // 初始化
  useEffect(() => {
    fetchAirports();
    fetchGroupOptions();
    fetchProtocolOptions();
    fetchNodeCheckProfiles();
  }, [fetchAirports, fetchGroupOptions, fetchNodeCheckProfiles, fetchProtocolOptions]);

  // 筛选条件变化时清空选择，避免对隐藏项误做批量操作
  useEffect(() => {
    setSelectedAirportIds([]);
  }, [searchKeyword, searchGroup, searchEnabled]);

  // 任务进度钩子
  const { registerOnComplete, unregisterOnComplete } = useTaskProgress();

  // 监听任务完成
  useEffect(() => {
    const handleTaskComplete = (task) => {
      // 当订阅更新任务完成时，刷新列表以获取最新状态
      if (task.taskType === 'sub_update') {
        fetchAirports();
      }
    };

    registerOnComplete(handleTaskComplete);
    return () => unregisterOnComplete(handleTaskComplete);
  }, [registerOnComplete, unregisterOnComplete, fetchAirports]);

  // 刷新
  const handleRefresh = () => {
    fetchAirports();
    fetchGroupOptions();
  };

  // 切换单个机场选择状态
  const handleToggleAirportSelection = (airportId) => {
    setSelectedAirportIds((prev) => (prev.includes(airportId) ? prev.filter((id) => id !== airportId) : [...prev, airportId]));
  };

  // 切换当前页全选状态
  const handleToggleCurrentPageSelection = (checked) => {
    const pageIds = airports.map((airport) => airport.id);
    setSelectedAirportIds((prev) => {
      if (checked) {
        const merged = [...prev];
        pageIds.forEach((id) => {
          if (!merged.includes(id)) {
            merged.push(id);
          }
        });
        return merged;
      }
      return prev.filter((id) => !pageIds.includes(id));
    });
  };

  // 清空机场选择
  const handleClearSelection = () => {
    setSelectedAirportIds([]);
  };

  // 选择当前筛选结果中的全部机场
  const handleSelectFilteredAirports = async () => {
    if (totalItems === 0) {
      showMessage(t('airports.page.messages.noFilteredAirports'), 'warning');
      return;
    }

    setSelectingFiltered(true);
    try {
      const response = await getAirports(buildAirportQueryParams());
      const items = response.data?.items || (Array.isArray(response.data) ? response.data : []);
      const ids = items.map((airport) => airport.id);
      setSelectedAirportIds(ids);
      showMessage(t('airports.page.messages.selectedFiltered', { count: ids.length }));
    } catch (error) {
      console.error('选择筛选结果失败:', error);
      showMessage(error.message || t('airports.page.messages.selectFilteredFailed'), 'error');
    } finally {
      setSelectingFiltered(false);
    }
  };

  // 打开批量设置对话框
  const handleOpenBatchDialog = () => {
    if (selectedAirportIds.length === 0) {
      showMessage(t('airports.page.messages.selectAirportsFirst'), 'warning');
      return;
    }
    setBatchForm(createBatchFormState());
    setBatchDialogOpen(true);
  };

  // 提交批量设置
  const handleBatchSubmit = async () => {
    if (selectedAirportIds.length === 0) {
      showMessage(t('airports.page.messages.selectAirportsFirst'), 'warning');
      return;
    }
    if (!batchForm.applyGroup && !batchForm.applySchedule) {
      showMessage(t('airports.page.messages.selectFieldFirst'), 'warning');
      return;
    }

    const cronExpr = batchForm.cronExpr.trim();
    if (batchForm.applySchedule) {
      if (!cronExpr) {
        showMessage(t('airports.page.messages.cronRequired'), 'warning');
        return;
      }
      if (!validateCronExpression(cronExpr)) {
        showMessage(t('airports.page.messages.cronInvalid'), 'error');
        return;
      }
    }

    setBatchSubmitting(true);
    try {
      const response = await batchUpdateAirports({
        ids: selectedAirportIds,
        applyGroup: batchForm.applyGroup,
        group: batchForm.applyGroup ? batchForm.group.trim() : '',
        applySchedule: batchForm.applySchedule,
        cronExpr: batchForm.applySchedule ? cronExpr : ''
      });
      const count = response.data?.count || selectedAirportIds.length;
      showMessage(t('airports.page.messages.batchUpdated', { count }));
      setBatchDialogOpen(false);
      setBatchForm(createBatchFormState());
      setSelectedAirportIds([]);
      fetchAirports();
      fetchGroupOptions();
    } catch (error) {
      console.error('批量更新机场失败:', error);
      showMessage(error.message || t('airports.page.messages.batchUpdateFailed'), 'error');
    } finally {
      setBatchSubmitting(false);
    }
  };

  // 打开确认对话框
  const openConfirm = (title, content, action) => {
    setConfirmInfo({ title, content, action });
    setConfirmOpen(true);
  };

  // 关闭确认对话框
  const handleConfirmClose = () => {
    setConfirmOpen(false);
  };

  // === 机场操作 ===

  // 添加机场
  const handleAdd = () => {
    const newForm = createAirportFormState({ requestHeaders: [createEmptyRequestHeader()] });
    setIsEdit(false);
    setAirportForm(newForm);
    setAirportFormSnapshot(newForm);
    setFormOpen(true);
  };

  // 编辑机场
  const handleEdit = (airport) => {
    const editForm = createAirportFormState({
      id: airport.id,
      name: airport.name,
      url: airport.url,
      cronExpr: airport.cronExpr,
      enabled: airport.enabled,
      group: airport.group || '',
      downloadWithProxy: airport.downloadWithProxy || false,
      proxyLink: airport.proxyLink || '',
      userAgent: airport.userAgent || '',
      requestHeaders: normalizeRequestHeadersForForm(airport.requestHeaders),
      fetchUsageInfo: airport.fetchUsageInfo || false,
      skipTLSVerify: airport.skipTLSVerify || false,
      updateAfterDetect: airport.updateAfterDetect || false,
      updateAfterDetectProfileId: airport.updateAfterDetectProfileId || 0,
      updateAfterDetectChangedOnly: airport.updateAfterDetectChangedOnly || false,
      remark: airport.remark || '',
      logo: airport.logo || '',
      nodeNameWhitelist: airport.nodeNameWhitelist || '',
      nodeNameBlacklist: airport.nodeNameBlacklist || '',
      protocolWhitelist: airport.protocolWhitelist || '',
      protocolBlacklist: airport.protocolBlacklist || '',
      nodeNamePreprocess: airport.nodeNamePreprocess || '',
      deduplicationRule: airport.deduplicationRule || '',
      nodeNameUniquify: airport.nodeNameUniquify || false,
      nodeNamePrefix: airport.nodeNamePrefix || '',
      nodeNameIntraUniquify: airport.nodeNameIntraUniquify || false
    });
    setIsEdit(true);
    setAirportForm(editForm);
    setAirportFormSnapshot(editForm);
    if (airport.downloadWithProxy) {
      fetchProxyNodes();
    }
    setFormOpen(true);
  };

  // 删除机场
  const handleDelete = (airport) => {
    setDeleteTarget(airport);
    setDeleteWithNodes(true);
    setDeleteDialogOpen(true);
  };

  // 确认删除
  const handleConfirmDelete = async () => {
    if (!deleteTarget) return;
    try {
      await deleteAirport(deleteTarget.id, deleteWithNodes);
      showMessage(t(deleteWithNodes ? 'airports.page.messages.deletedWithNodes' : 'airports.page.messages.deletedKeepNodes'));
      setSelectedAirportIds((prev) => prev.filter((id) => id !== deleteTarget.id));
      fetchAirports();
    } catch (error) {
      console.error('删除失败:', error);
      showMessage(error.message || t('airports.page.messages.deleteFailed'), 'error');
    }
    setDeleteDialogOpen(false);
    setDeleteTarget(null);
  };

  // 拉取机场
  const handlePull = (airport) => {
    openConfirm(t('airports.page.confirm.pullTitle'), t('airports.page.confirm.pullContent', { name: airport.name }), async () => {
      try {
        await pullAirport(airport.id);
        showMessage(t('airports.page.messages.pullSubmitted'));
        // 任务完成后会自动触发刷新
      } catch (error) {
        console.error('拉取失败:', error);
        showMessage(error.message || t('airports.page.messages.pullSubmitFailed'), 'error');
      }
    });
  };

  // 批量拉取机场
  const handlePullAll = async () => {
    const hasSelection = selectedAirportIds.length > 0;

    if (hasSelection) {
      // 模式1：拉取选中的机场
      const selectedCount = selectedAirportIds.length;
      const confirmContent = t('airports.page.confirm.pullSelectedContentWithCount', { count: selectedCount });

      openConfirm(t('airports.page.confirm.pullSelectedTitle'), confirmContent, async () => {
        try {
          const res = await pullAllAirports(selectedAirportIds);
          const count = res.data?.count;

          if (typeof count === 'number') {
            if (count > 0) {
              showMessage(t('airports.page.messages.pullSelectedSubmittedWithCount', { count }));
              // 清空选择状态，提升用户体验
              setSelectedAirportIds([]);
            } else {
              showMessage(t('airports.page.messages.noValidAirports'), 'warning');
            }
            return;
          }

          // 兼容后端返回 OkWithMsg 的场景（data 为空）
          if (res.msg && res.msg !== '操作成功') {
            showMessage(res.msg, 'warning');
            return;
          }

          showMessage(t('airports.page.messages.pullSelectedSubmitted'));
          setSelectedAirportIds([]);
        } catch (error) {
          console.error('批量拉取失败:', error);
          showMessage(error.message || t('airports.page.messages.pullSelectedFailed'), 'error');
        }
      });
    } else {
      // 模式2：拉取所有已启用的机场
      // 这里不要用当前页 airports 来统计启用数量：列表是分页的，会导致误判”没有已启用机场”
      // 使用后端分页接口返回的 total 作为全量启用机场数（仅用于提示/确认）
      let enabledTotal = null;
      try {
        const res = await getAirports({ page: 1, pageSize: 1, enabled: true });
        if (typeof res.data?.total === 'number') {
          enabledTotal = res.data.total;
        }
      } catch (error) {
        console.error('获取已启用机场数量失败:', error);
        enabledTotal = null;
      }

      if (enabledTotal === 0) {
        showMessage(t('airports.page.messages.noEnabledAirports'), 'warning');
        return;
      }

      const confirmContent =
        typeof enabledTotal === 'number'
          ? t('airports.page.confirm.pullAllContentWithCount', { count: enabledTotal })
          : t('airports.page.confirm.pullAllContent');

      openConfirm(t('airports.page.confirm.pullAllTitle'), confirmContent, async () => {
        try {
          const res = await pullAllAirports();
          const count = res.data?.count;

          if (typeof count === 'number') {
            if (count > 0) {
              showMessage(t('airports.page.messages.pullAllSubmittedWithCount', { count }));
            } else {
              showMessage(t('airports.page.messages.noEnabledAirports'), 'warning');
            }
            return;
          }

          // 兼容后端返回 OkWithMsg 的场景（data 为空）
          if (res.msg && res.msg !== '操作成功') {
            showMessage(res.msg, 'warning');
            return;
          }

          showMessage(t('airports.page.messages.pullAllSubmitted'));
        } catch (error) {
          console.error('批量拉取失败:', error);
          showMessage(error.message || t('airports.page.messages.pullAllFailed'), 'error');
        }
      });
    }
  };

  // 刷新用量信息
  const handleRefreshUsage = async (airport) => {
    try {
      await refreshAirportUsage(airport.id);
      showMessage(t('airports.page.messages.usageRefreshed'));
      fetchAirports(); // 刷新列表
    } catch (error) {
      console.error('刷新用量失败:', error);
      showMessage(error.message || t('airports.page.messages.usageRefreshFailed'), 'error');
    }
  };

  const handleOpenNodeManagement = useCallback(
    (airport) => {
      const source = airport?.name?.trim();
      if (!source) {
        showMessage(t('airports.page.messages.emptyAirportNameForNodes'), 'warning');
        return;
      }

      navigate(`/subscription/nodes?source=${encodeURIComponent(source)}`);
    },
    [navigate, showMessage, t]
  );

  const handleQuickCheck = useCallback(
    async (airport) => {
      const source = airport?.name?.trim();
      if (!source) {
        showMessage(t('airports.page.messages.emptyAirportNameForQuickCheck'), 'warning');
        return;
      }

      try {
        const response = await getNodeIds({ source });
        const nodeIds = Array.isArray(response.data) ? response.data : [];

        if (nodeIds.length === 0) {
          showMessage(t('airports.page.messages.noNodesForSource', { source }), 'warning');
          return;
        }

        setProfileSelectNodeIds(nodeIds);
        setProfileSelectOpen(true);
      } catch (error) {
        console.error('获取机场节点失败:', error);
        showMessage(error.message || t('airports.page.messages.loadAirportNodesFailed'), 'error');
      }
    },
    [showMessage, t]
  );

  // 提交表单
  const handleSubmit = async () => {
    // 验证
    if (!airportForm.name.trim()) {
      showMessage(t('airports.page.messages.nameRequired'), 'warning');
      return;
    }
    if (!airportForm.url.trim()) {
      showMessage(t('airports.page.messages.urlRequired'), 'warning');
      return;
    }
    const urlPattern = /^(https?|ftp):\/\/[^\s/$.?#].[^\s]*$/i;
    if (!urlPattern.test(airportForm.url.trim())) {
      showMessage(t('airports.page.messages.urlInvalid'), 'warning');
      return;
    }
    if (!airportForm.cronExpr.trim()) {
      showMessage(t('airports.page.messages.cronRequired'), 'warning');
      return;
    }
    if (!validateCronExpression(airportForm.cronExpr.trim())) {
      showMessage(t('airports.page.messages.cronInvalid'), 'error');
      return;
    }
    if (airportForm.updateAfterDetect && !airportForm.updateAfterDetectProfileId) {
      showMessage(t('airports.page.messages.detectProfileRequired'), 'warning');
      return;
    }

    const { requestHeaders, error: requestHeadersError } = normalizeRequestHeadersForSubmit(airportForm.requestHeaders, t);
    if (requestHeadersError) {
      showMessage(requestHeadersError, 'warning');
      return;
    }

    try {
      const normalizedAirportForm = {
        ...airportForm,
        requestHeaders
      };

      // 在提交前计算配置变更状态（提交后 snapshot 会被清空）
      const needPullToApply = isEdit && hasNodeProcessConfigChanged(airportFormSnapshot, airportForm);
      const savedId = airportForm.id;
      const savedName = airportForm.name;

      if (isEdit) {
        await updateAirport(airportForm.id, normalizedAirportForm);
        showMessage(t(needPullToApply ? 'airports.page.messages.updateSuccessNeedPull' : 'airports.page.messages.updateSuccess'));
      } else {
        await addAirport(normalizedAirportForm);
        showMessage(t('airports.page.messages.addSuccess'));
      }
      setFormOpen(false);
      setAirportFormSnapshot(null);
      fetchAirports();

      // 节点处理配置变更后提示立即拉取，使其对”已存在节点”生效
      if (needPullToApply) {
        openConfirm(
          t('airports.page.confirm.pullToApplyTitle'),
          t('airports.page.confirm.pullToApplyContent', { name: savedName }),
          async () => {
            try {
              await pullAirport(savedId);
              showMessage(t('airports.page.messages.pullSubmitted'));
            } catch (error) {
              console.error('拉取失败:', error);
              showMessage(error.message || t('airports.page.messages.pullSubmitFailed'), 'error');
            }
          }
        );
      }
    } catch (error) {
      console.error('提交失败:', error);
      showMessage(error.message || t(isEdit ? 'airports.page.messages.updateFailed' : 'airports.page.messages.addFailed'), 'error');
    }
  };

  const currentPageIds = airports.map((airport) => airport.id);
  const selectedOnCurrentPage = currentPageIds.filter((id) => selectedAirportIds.includes(id)).length;
  const allCurrentPageSelected = currentPageIds.length > 0 && selectedOnCurrentPage === currentPageIds.length;
  const currentPageIndeterminate = selectedOnCurrentPage > 0 && selectedOnCurrentPage < currentPageIds.length;
  const hasSelection = selectedAirportIds.length > 0;
  const allFilteredSelected = totalItems > 0 && selectedAirportIds.length === totalItems;

  const selectionToolbarSx = {
    mb: 2,
    p: 1.5,
    borderRadius: 2.5,
    borderColor: isDark ? withAlpha(palette.primary.main, 0.24) : withAlpha(palette.primary.main, 0.16),
    backgroundColor: isDark ? withAlpha(palette.background.paper, 0.96) : withAlpha(palette.primary.main, 0.03)
  };

  const selectionRegionSx = {
    display: 'flex',
    alignItems: 'center',
    minWidth: 0,
    gap: 1,
    px: 1,
    py: 0.5,
    borderRadius: 2,
    bgcolor: isDark ? withAlpha(palette.background.default, 0.88) : withAlpha(palette.background.paper, 0.82),
    border: `1px solid ${hasSelection ? withAlpha(palette.primary.main, isDark ? 0.34 : 0.2) : withAlpha(palette.divider, isDark ? 0.42 : 0.18)}`
  };

  const getSelectionChipSx = (active) => ({
    height: 24,
    borderRadius: 999,
    fontWeight: 600,
    color: active ? (isDark ? palette.primary.light : palette.primary.main) : palette.text.secondary,
    bgcolor: active
      ? isDark
        ? withAlpha(palette.primary.main, 0.18)
        : withAlpha(palette.primary.main, 0.08)
      : isDark
        ? withAlpha(palette.background.default, 0.92)
        : withAlpha(palette.background.default, 0.72),
    border: `1px solid ${active ? withAlpha(palette.primary.main, isDark ? 0.34 : 0.2) : withAlpha(palette.divider, isDark ? 0.42 : 0.18)}`,
    '& .MuiChip-label': {
      px: 1.1
    }
  });

  const selectionActionButtonBaseSx = {
    minHeight: 34,
    borderRadius: 2,
    fontWeight: 600,
    whiteSpace: 'nowrap',
    boxShadow: 'none',
    '&.Mui-disabled': {
      color: palette.action.disabled,
      borderColor: withAlpha(palette.action.disabledBackground, isDark ? 0.68 : 1),
      backgroundColor: isDark ? withAlpha(palette.action.disabledBackground, 0.42) : withAlpha(palette.action.disabledBackground, 0.92)
    }
  };

  const selectFilteredButtonSx = {
    ...selectionActionButtonBaseSx,
    borderColor: withAlpha(palette.primary.main, isDark ? 0.34 : 0.18),
    bgcolor: isDark ? withAlpha(palette.primary.main, 0.1) : withAlpha(palette.primary.main, 0.03),
    color: isDark ? palette.primary.light : palette.primary.main,
    '&:hover': {
      borderColor: withAlpha(palette.primary.main, isDark ? 0.44 : 0.24),
      bgcolor: isDark ? withAlpha(palette.primary.main, 0.16) : withAlpha(palette.primary.main, 0.06)
    }
  };

  const clearSelectionButtonSx = {
    ...selectionActionButtonBaseSx,
    borderColor: withAlpha(palette.divider, isDark ? 0.42 : 0.18),
    bgcolor: isDark ? withAlpha(palette.background.default, 0.9) : withAlpha(palette.background.paper, 0.82),
    color: palette.text.primary,
    '&:hover': {
      borderColor: withAlpha(palette.primary.main, isDark ? 0.24 : 0.14),
      bgcolor: isDark ? withAlpha(palette.primary.main, 0.08) : withAlpha(palette.primary.main, 0.04)
    }
  };

  const batchActionButtonSx = {
    ...selectionActionButtonBaseSx,
    border: `1px solid ${withAlpha(palette.primary.main, isDark ? 0.42 : 0.22)}`,
    bgcolor: isDark ? withAlpha(palette.primary.main, 0.18) : palette.primary.main,
    color: isDark ? palette.primary.light : palette.primary.contrastText,
    '&:hover': {
      boxShadow: 'none',
      borderColor: withAlpha(palette.primary.main, isDark ? 0.5 : 0.28),
      bgcolor: isDark ? withAlpha(palette.primary.main, 0.24) : palette.primary.dark
    }
  };

  return (
    <MainCard
      title={t('airports.title')}
      secondary={
        <Stack direction="row" spacing={1} alignItems="center">
          {/* 视图切换按钮（仅桌面端显示） */}
          {!matchDownMd && (
            <ToggleButtonGroup
              value={viewMode}
              exclusive
              onChange={(e, newMode) => {
                if (newMode) {
                  setViewMode(newMode);
                  localStorage.setItem('airports_viewMode', newMode);
                }
              }}
              size="small"
              sx={{ mr: 1 }}
            >
              <ToggleButton value="card" sx={{ px: 1, py: 0.5 }}>
                <ViewModuleIcon fontSize="small" />
              </ToggleButton>
              <ToggleButton value="list" sx={{ px: 1, py: 0.5 }}>
                <ViewListIcon fontSize="small" />
              </ToggleButton>
            </ToggleButtonGroup>
          )}
          <Button variant="contained" startIcon={<AddIcon />} onClick={handleAdd}>
            {t('airports.page.actions.addAirport')}
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
      }
    >
      {/* 搜索筛选栏 */}
      <Box sx={{ mb: 2 }}>
        <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2} alignItems={{ xs: 'stretch', sm: 'center' }}>
          <TextField
            size="small"
            label={t('airports.page.filters.keyword')}
            placeholder={t('airports.page.filters.keywordPlaceholder')}
            value={searchKeyword}
            onChange={(e) => setSearchKeyword(e.target.value)}
            sx={{ minWidth: 150 }}
          />
          <Autocomplete
            size="small"
            options={groupOptions}
            value={searchGroup || null}
            onChange={(e, newValue) => setSearchGroup(newValue || '')}
            renderInput={(params) => <TextField {...params} label={t('airports.page.filters.group')} />}
            sx={{ minWidth: 150 }}
          />
          <FormControl size="small" sx={{ minWidth: 100 }}>
            <InputLabel>{t('airports.page.filters.status')}</InputLabel>
            <Select value={searchEnabled} label={t('airports.page.filters.status')} onChange={(e) => setSearchEnabled(e.target.value)}>
              <MenuItem value="">{t('common.all')}</MenuItem>
              <MenuItem value="true">{t('common.enabled')}</MenuItem>
              <MenuItem value="false">{t('common.disabled')}</MenuItem>
            </Select>
          </FormControl>
          <Button
            variant="outlined"
            size="small"
            onClick={() => {
              setSelectedAirportIds([]);
              setSearchKeyword('');
              setSearchGroup('');
              setSearchEnabled('');
              setPage(0);
            }}
          >
            {t('common.reset')}
          </Button>
          <Button
            variant="contained"
            size="small"
            onClick={() => {
              setSelectedAirportIds([]);
              setPage(0);
              fetchAirports();
            }}
          >
            {t('airports.page.actions.search')}
          </Button>
        </Stack>
      </Box>

      {/* 任务进度显示 */}
      <TaskProgressPanel />

      {/* 批量操作栏 */}
      {(airports.length > 0 || selectedAirportIds.length > 0) && (
        <Paper variant="outlined" sx={selectionToolbarSx}>
          {matchDownMd ? (
            <Stack spacing={1.5}>
              <Stack direction="row" alignItems="center" justifyContent="space-between" sx={{ gap: 1 }}>
                <Box sx={selectionRegionSx}>
                  <Checkbox
                    checked={allCurrentPageSelected}
                    indeterminate={currentPageIndeterminate}
                    onChange={(e) => handleToggleCurrentPageSelection(e.target.checked)}
                    disabled={currentPageIds.length === 0}
                    size="small"
                    sx={{
                      p: 0.5,
                      color: hasSelection ? 'primary.main' : 'text.secondary'
                    }}
                  />
                  <Typography variant="body2" sx={{ whiteSpace: 'nowrap', color: 'text.primary', fontWeight: 500 }}>
                    {t('airports.page.selection.currentPage')}
                  </Typography>
                </Box>
                <Stack direction="row" spacing={0.75} sx={{ flexWrap: 'wrap', justifyContent: 'flex-end', gap: 0.75 }}>
                  <Chip
                    size="small"
                    label={t('airports.page.selection.selectedCount', { count: selectedAirportIds.length })}
                    sx={getSelectionChipSx(hasSelection)}
                  />
                  <Chip
                    variant="outlined"
                    size="small"
                    label={t('airports.page.selection.filteredCount', { count: totalItems })}
                    sx={getSelectionChipSx(false)}
                  />
                </Stack>
              </Stack>

              <Box
                sx={{
                  display: 'grid',
                  gridTemplateColumns: 'repeat(2, minmax(0, 1fr))',
                  gap: 1
                }}
              >
                <Button
                  fullWidth
                  size="small"
                  variant="contained"
                  startIcon={<TuneIcon />}
                  onClick={handleOpenBatchDialog}
                  disabled={!hasSelection}
                  sx={{
                    ...batchActionButtonSx,
                    gridColumn: '1 / -1'
                  }}
                >
                  {t('airports.page.actions.batchSettings')}
                </Button>
                <Button
                  fullWidth
                  size="small"
                  variant="outlined"
                  onClick={handleSelectFilteredAirports}
                  disabled={selectingFiltered || totalItems === 0 || allFilteredSelected}
                  sx={selectFilteredButtonSx}
                >
                  {selectingFiltered ? t('airports.page.actions.selecting') : t('airports.page.actions.selectFiltered')}
                </Button>
                <Button
                  fullWidth
                  size="small"
                  variant="outlined"
                  onClick={handleClearSelection}
                  disabled={!hasSelection}
                  sx={clearSelectionButtonSx}
                >
                  {t('airports.page.actions.clearSelection')}
                </Button>
                <Button
                  fullWidth
                  size="small"
                  variant="outlined"
                  startIcon={<CloudSyncIcon />}
                  onClick={handlePullAll}
                  sx={{
                    ...selectionActionButtonBaseSx,
                    gridColumn: '1 / -1',
                    borderColor: withAlpha(palette.primary.main, isDark ? 0.4 : 0.3),
                    bgcolor: isDark ? withAlpha(palette.primary.main, 0.08) : withAlpha(palette.primary.main, 0.04),
                    color: isDark ? palette.primary.light : palette.primary.main,
                    fontWeight: 600,
                    '&:hover': {
                      borderColor: withAlpha(palette.primary.main, isDark ? 0.5 : 0.4),
                      bgcolor: isDark ? withAlpha(palette.primary.main, 0.12) : withAlpha(palette.primary.main, 0.08)
                    }
                  }}
                >
                  {t('airports.page.actions.pullAllShort')}
                </Button>
              </Box>
            </Stack>
          ) : (
            <Stack direction={{ xs: 'column', lg: 'row' }} spacing={1.5} alignItems={{ xs: 'stretch', lg: 'center' }}>
              <Stack direction="row" spacing={1} alignItems="center" sx={{ flexWrap: 'wrap', gap: 0.5 }}>
                <Box sx={selectionRegionSx}>
                  <Checkbox
                    checked={allCurrentPageSelected}
                    indeterminate={currentPageIndeterminate}
                    onChange={(e) => handleToggleCurrentPageSelection(e.target.checked)}
                    disabled={currentPageIds.length === 0}
                    size="small"
                    sx={{
                      p: 0.5,
                      color: hasSelection ? 'primary.main' : 'text.secondary'
                    }}
                  />
                  <Typography variant="body2" sx={{ whiteSpace: 'nowrap', color: 'text.primary', fontWeight: 500 }}>
                    {t('airports.page.selection.currentPage')}
                  </Typography>
                </Box>
                <Chip
                  size="small"
                  label={t('airports.page.selection.selectedCount', { count: selectedAirportIds.length })}
                  sx={getSelectionChipSx(hasSelection)}
                />
                <Chip
                  variant="outlined"
                  size="small"
                  label={t('airports.page.selection.filteredCount', { count: totalItems })}
                  sx={getSelectionChipSx(false)}
                />
              </Stack>

              <Stack direction="row" spacing={1} sx={{ flexWrap: 'wrap', gap: 1 }}>
                <Button
                  size="small"
                  variant="outlined"
                  onClick={handleSelectFilteredAirports}
                  disabled={selectingFiltered || totalItems === 0 || allFilteredSelected}
                  sx={selectFilteredButtonSx}
                >
                  {selectingFiltered ? t('airports.page.actions.selecting') : t('airports.page.actions.selectFiltered')}
                </Button>
                <Button size="small" variant="outlined" onClick={handleClearSelection} disabled={!hasSelection} sx={clearSelectionButtonSx}>
                  {t('airports.page.actions.clearSelection')}
                </Button>
                <Button
                  size="small"
                  variant="contained"
                  startIcon={<TuneIcon />}
                  onClick={handleOpenBatchDialog}
                  disabled={!hasSelection}
                  sx={batchActionButtonSx}
                >
                  {t('airports.page.actions.batchSettings')}
                </Button>
                <Button
                  size="small"
                  variant="outlined"
                  startIcon={<CloudSyncIcon />}
                  onClick={handlePullAll}
                  sx={{
                    ...selectionActionButtonBaseSx,
                    borderColor: withAlpha(palette.primary.main, isDark ? 0.4 : 0.3),
                    bgcolor: isDark ? withAlpha(palette.primary.main, 0.08) : withAlpha(palette.primary.main, 0.04),
                    color: isDark ? palette.primary.light : palette.primary.main,
                    fontWeight: 600,
                    '&:hover': {
                      borderColor: withAlpha(palette.primary.main, isDark ? 0.5 : 0.4),
                      bgcolor: isDark ? withAlpha(palette.primary.main, 0.12) : withAlpha(palette.primary.main, 0.08)
                    }
                  }}
                >
                  {t('airports.page.actions.pullAll')}
                </Button>
              </Stack>
            </Stack>
          )}
        </Paper>
      )}

      {/* 机场列表 */}
      {matchDownMd ? (
        <AirportMobileList
          airports={airports}
          selectedIds={selectedAirportIds}
          onToggleSelect={handleToggleAirportSelection}
          onEdit={handleEdit}
          onDelete={handleDelete}
          onPull={handlePull}
          onOpenNodes={handleOpenNodeManagement}
          onQuickCheck={handleQuickCheck}
          onRefreshUsage={handleRefreshUsage}
          nodeCheckProfiles={nodeCheckProfiles}
        />
      ) : viewMode === 'list' ? (
        <AirportListView
          airports={airports}
          selectedIds={selectedAirportIds}
          onToggleSelect={handleToggleAirportSelection}
          onEdit={handleEdit}
          onDelete={handleDelete}
          onPull={handlePull}
          onOpenNodes={handleOpenNodeManagement}
          onQuickCheck={handleQuickCheck}
          onRefreshUsage={handleRefreshUsage}
          nodeCheckProfiles={nodeCheckProfiles}
        />
      ) : (
        <AirportTable
          airports={airports}
          selectedIds={selectedAirportIds}
          onToggleSelect={handleToggleAirportSelection}
          onEdit={handleEdit}
          onDelete={handleDelete}
          onPull={handlePull}
          onOpenNodes={handleOpenNodeManagement}
          onQuickCheck={handleQuickCheck}
          onRefreshUsage={handleRefreshUsage}
          nodeCheckProfiles={nodeCheckProfiles}
        />
      )}

      {/* 分页 */}
      <Pagination
        page={page}
        pageSize={rowsPerPage}
        totalItems={totalItems}
        onPageChange={(e, newPage) => {
          setPage(newPage);
        }}
        onPageSizeChange={(e) => {
          const newValue = parseInt(e.target.value, 10);
          setRowsPerPage(newValue);
          localStorage.setItem('airports_rowsPerPage', newValue);
          setPage(0);
        }}
        pageSizeOptions={[10, 20, 50]}
      />

      {/* 添加/编辑对话框 */}
      <AirportFormDialog
        open={formOpen}
        isEdit={isEdit}
        airportForm={airportForm}
        setAirportForm={setAirportForm}
        groupOptions={groupOptions}
        proxyNodeOptions={proxyNodeOptions}
        loadingProxyNodes={loadingProxyNodes}
        protocolOptions={protocolOptions}
        nodeCheckProfiles={nodeCheckProfiles}
        onClose={() => {
          setFormOpen(false);
          setAirportFormSnapshot(null);
        }}
        onSubmit={handleSubmit}
        onFetchProxyNodes={fetchProxyNodes}
      />

      {/* 批量设置对话框 */}
      <AirportBatchEditDialog
        open={batchDialogOpen}
        selectedCount={selectedAirportIds.length}
        batchForm={batchForm}
        setBatchForm={setBatchForm}
        groupOptions={groupOptions}
        onClose={() => {
          if (!batchSubmitting) {
            setBatchDialogOpen(false);
            setBatchForm(createBatchFormState());
          }
        }}
        onSubmit={handleBatchSubmit}
        submitting={batchSubmitting}
      />

      {/* 删除确认对话框 */}
      <DeleteAirportDialog
        open={deleteDialogOpen}
        airport={deleteTarget}
        withNodes={deleteWithNodes}
        setWithNodes={setDeleteWithNodes}
        onClose={() => setDeleteDialogOpen(false)}
        onConfirm={handleConfirmDelete}
      />

      <ProfileSelectDialog
        open={profileSelectOpen}
        onClose={() => setProfileSelectOpen(false)}
        nodeIds={profileSelectNodeIds}
        onSuccess={showMessage}
        onOpenSettings={() => navigate('/subscription/node-check')}
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
