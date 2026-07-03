import React, { useState, useEffect, useCallback, useMemo, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import Button from '@mui/material/Button';
import Stack from '@mui/material/Stack';
import Box from '@mui/material/Box';
import TextField from '@mui/material/TextField';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import IconButton from '@mui/material/IconButton';
import { formatDateTime } from 'i18n/locales';
import Chip from '@mui/material/Chip';
import Typography from '@mui/material/Typography';
import Switch from '@mui/material/Switch';
import FormControlLabel from '@mui/material/FormControlLabel';
import Alert from '@mui/material/Alert';
import CircularProgress from '@mui/material/CircularProgress';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Tooltip from '@mui/material/Tooltip';
import useMediaQuery from '@mui/material/useMediaQuery';
import { useTheme } from '@mui/material/styles';

import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import LinkIcon from '@mui/icons-material/Link';
import RefreshIcon from '@mui/icons-material/Refresh';
import HistoryIcon from '@mui/icons-material/History';
import DownloadIcon from '@mui/icons-material/Download';
import SortIcon from '@mui/icons-material/Sort';
import Checkbox from '@mui/material/Checkbox';
import Divider from '@mui/material/Divider';

import {
  getShares,
  createShare,
  updateShare,
  deleteShare,
  getShareLogs,
  refreshShareToken,
  batchCreateShares,
  batchDeleteShares,
  batchUpdateShares
} from '../../../api/shares';
import { getSubStoreSettings, getSystemDomain } from '../../../api/settings';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';
import AccessLogsDialog from './AccessLogsDialog';
import ClientUrlsDialog from './ClientUrlsDialog';
import QrCodeDialog from './QrCodeDialog';
import ConfirmDialog from './ConfirmDialog';
import ShareBatchCreateDialog from './ShareBatchCreateDialog';
import ShareBatchUpdateDialog from './ShareBatchUpdateDialog';
import ShareExportDialog from './ShareExportDialog';

const EXPIRE_TYPE_NEVER = 0;
const EXPIRE_TYPE_DAYS = 1;
const EXPIRE_TYPE_DATETIME = 2;
const SHARE_PAGE_SIZE = 50;

const NATIVE_CLIENT_LINKS = [
  { key: 'clash', client: 'clash' },
  { key: 'mihomo', client: 'mihomo' },
  { key: 'surge', client: 'surge' },
  { key: 'v2ray', client: 'v2ray' }
];

const EXPANDED_CLIENT_LINKS = [
  { key: 'loon', client: 'loon' },
  { key: 'egern', client: 'egern' },
  { key: 'stash', client: 'stash' },
  { key: 'surfboard', client: 'surfboard' },
  { key: 'shadowrocket', client: 'shadowrocket' },
  { key: 'quantumultX', client: 'quanx' },
  { key: 'singBox', client: 'sing-box' },
  { key: 'uri', client: 'uri' },
  { key: 'json', client: 'json' }
];

export default function ShareManageDialog({ open, subscription, onClose, showMessage }) {
  const theme = useTheme();
  const { t, i18n } = useTranslation();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const { isDark } = useResolvedColorScheme();
  const { palette, dialogSurface, dialogSurfaceGradient, mutedPanelSurface, nestedPanelSurface, panelBorder } = getSurfaceTokens(
    theme,
    isDark
  );
  const { primaryText, secondaryText, tertiaryText } = getReadableTextTokens(theme, isDark);

  const [shares, setShares] = useState([]);
  const [loading, setLoading] = useState(false);
  const [searching, setSearching] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');

  // 排序和筛选状态
  const [sortBy, setSortBy] = useState('');
  const [sortOrder, setSortOrder] = useState('desc');
  const [ipFilter, setIpFilter] = useState('');

  // 分页状态
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);

  // 批量操作状态
  const [selectedShares, setSelectedShares] = useState([]);
  const [batchCreateOpen, setBatchCreateOpen] = useState(false);
  const [batchUpdateOpen, setBatchUpdateOpen] = useState(false);
  const [batchUpdateMode, setBatchUpdateMode] = useState('expire'); // 'expire' or 'enabled'
  const [exportDialogOpen, setExportDialogOpen] = useState(false);

  const [systemDomainConfig, setSystemDomainConfig] = useState('');
  const [subStoreTargets, setSubStoreTargets] = useState([]);

  const [detailOpen, setDetailOpen] = useState(false);
  const [detailShare, setDetailShare] = useState(null);

  const [formOpen, setFormOpen] = useState(false);
  const [editingShare, setEditingShare] = useState(null);
  const [formData, setFormData] = useState({
    name: '',
    token: '',
    expire_type: EXPIRE_TYPE_NEVER,
    expire_days: 30,
    expire_at: '',
    enabled: true
  });

  const [qrOpen, setQrOpen] = useState(false);
  const [qrUrl, setQrUrl] = useState('');
  const [qrTitle, setQrTitle] = useState('');

  const [logsOpen, setLogsOpen] = useState(false);
  const [logsLoading, setLogsLoading] = useState(false);
  const [logs, setLogs] = useState([]);
  const [logsShareName, setLogsShareName] = useState('');

  const [confirmOpen, setConfirmOpen] = useState(false);
  const [confirmInfo, setConfirmInfo] = useState({ title: '', content: '', onConfirm: null });
  const dialogContentRef = useRef(null);

  const getServerUrl = useCallback(() => {
    if (systemDomainConfig) {
      return systemDomainConfig.replace(/\/+$/, '');
    }
    return `${window.location.protocol}//${window.location.hostname}${window.location.port ? ':' + window.location.port : ''}`;
  }, [systemDomainConfig]);

  const fetchSystemDomain = async () => {
    try {
      const res = await getSystemDomain();
      if (res.data?.systemDomain) {
        setSystemDomainConfig(res.data.systemDomain);
      }
    } catch (error) {
      console.error('Failed to get system domain config:', error);
    }
  };

  const fetchShares = useCallback(
    async (keyword = '', isSearch = false, customSortBy = null, customSortOrder = null) => {
      if (!subscription?.ID) return;
      if (isSearch) {
        setSearching(true);
      } else {
        setLoading(true);
      }
      try {
        // 检测是否为IP地址
        const isIP = isIPAddress(keyword);
        const ipFilterValue = isIP ? keyword : '';
        const keywordValue = isIP ? '' : keyword;

        // 使用传入的排序参数,如果没有则使用当前状态
        const actualSortBy = customSortBy !== null ? customSortBy : sortBy;
        const actualSortOrder = customSortOrder !== null ? customSortOrder : sortOrder;

        const res = await getShares(subscription.ID, 1, SHARE_PAGE_SIZE, keywordValue, ipFilterValue, actualSortBy, actualSortOrder);
        if (res.data?.items) {
          // 分页响应
          setShares(res.data.items);
          setHasMore(res.data.hasMore || false);
          setPage(1);
          setIpFilter(ipFilterValue); // 保存IP筛选用于分页
        } else {
          // 兼容旧版本响应（返回所有数据）
          setShares(res.data || []);
          setHasMore(false);
        }
      } catch (error) {
        console.error('Failed to get share list:', error);
      } finally {
        if (isSearch) {
          setSearching(false);
        } else {
          setLoading(false);
        }
      }
    },
    [subscription?.ID, sortBy, sortOrder] // 添加sortBy和sortOrder依赖,但实际使用customSortBy/customSortOrder参数
  );

  const loadMoreShares = useCallback(async () => {
    if (loadingMore || !hasMore || !subscription?.ID) return;
    setLoadingMore(true);
    try {
      const nextPage = page + 1;
      const isIP = isIPAddress(searchQuery);
      const res = await getShares(
        subscription.ID,
        nextPage,
        SHARE_PAGE_SIZE,
        isIP ? '' : searchQuery,
        isIP ? searchQuery : ipFilter,
        sortBy,
        sortOrder
      );
      if (res.data?.items) {
        setShares((prev) => [...prev, ...res.data.items]);
        setHasMore(res.data.hasMore || false);
        setPage(nextPage);
      }
    } catch (error) {
      console.error('Failed to load more shares:', error);
    } finally {
      setLoadingMore(false);
    }
  }, [loadingMore, hasMore, subscription?.ID, page, searchQuery, ipFilter, sortBy, sortOrder]);

  const handleSort = (field) => {
    let newSortBy = sortBy;
    let newSortOrder = sortOrder;
    const defaultSortOrder = field === 'name' ? 'asc' : 'desc';

    if (sortBy === field) {
      if (sortOrder === defaultSortOrder) {
        newSortBy = field;
        newSortOrder = defaultSortOrder === 'asc' ? 'desc' : 'asc';
      } else {
        newSortBy = '';
        newSortOrder = 'desc';
      }
    } else {
      newSortBy = field;
      newSortOrder = defaultSortOrder;
    }

    // 先更新状态
    setSortBy(newSortBy);
    setSortOrder(newSortOrder);

    // 直接用新值调用fetchShares
    fetchShares(searchQuery, true, newSortBy, newSortOrder);
  };

  // 移除fetchSharesWithSort函数
  // const fetchSharesWithSort = async (sortByParam, sortOrderParam) => {
  //   ...
  // };

  // 移除排序相关的useEffect注释

  const fetchSubStoreTargets = useCallback(async () => {
    try {
      const res = await getSubStoreSettings();
      const settings = res.data || {};
      const enabled = Boolean(settings.enabled?.value);
      setSubStoreTargets(settings.configured && enabled ? settings.allowedTargets?.value || [] : []);
    } catch (error) {
      console.error('Failed to get Sub-Store settings:', error);
      setSubStoreTargets([]);
    }
  }, []);

  useEffect(() => {
    if (open && subscription?.ID) {
      // 重置状态
      setSearchQuery('');
      setSelectedShares([]);
      setPage(1);
      setHasMore(true);
      setSortBy('');
      setSortOrder('desc');

      fetchSystemDomain();
      fetchShares('', false, '', 'desc'); // 明确传入初始排序参数
      fetchSubStoreTargets();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, subscription?.ID]);

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
    showMessage?.(t('common.copied'), 'success');
  };

  const handleOpenDetail = (share) => {
    setDetailShare(share);
    setDetailOpen(true);
  };

  const handleAdd = () => {
    setEditingShare(null);
    setFormData({
      name: '',
      token: '',
      expire_type: EXPIRE_TYPE_NEVER,
      expire_days: 30,
      expire_at: '',
      enabled: true
    });
    setFormOpen(true);
  };

  const handleEdit = (share, e) => {
    e?.stopPropagation();
    setEditingShare(share);
    setFormData({
      name: share.name || '',
      token: share.token || '',
      expire_type: share.expire_type || EXPIRE_TYPE_NEVER,
      expire_days: share.expire_days || 30,
      expire_at: share.expire_at ? share.expire_at.substring(0, 16) : '',
      enabled: share.enabled !== false
    });
    setFormOpen(true);
  };

  const handleSave = async () => {
    try {
      const data = {
        ...formData,
        subscription_id: subscription.ID
      };

      if (editingShare) {
        data.id = editingShare.id;
        await updateShare(data);
        showMessage?.(t('subscriptions.share.messages.updateSuccess'), 'success');
      } else {
        await createShare(data);
        showMessage?.(t('subscriptions.share.messages.createSuccess'), 'success');
      }
      setFormOpen(false);
      fetchShares();
    } catch (error) {
      console.error('Failed to save:', error);
      showMessage?.(error.response?.data?.msg || t('subscriptions.share.messages.saveFailed'), 'error');
    }
  };

  const handleDelete = (share, e) => {
    e?.stopPropagation();
    setConfirmInfo({
      title: t('subscriptions.share.confirm.deleteTitle'),
      content: t('subscriptions.share.confirm.deleteContent', { name: share.name || share.token }),
      onConfirm: async () => {
        try {
          await deleteShare(share.id);
          showMessage?.(t('subscriptions.share.messages.deleteSuccess'), 'success');
          fetchShares();
          if (detailShare?.id === share.id) {
            setDetailOpen(false);
          }
        } catch (error) {
          console.error('Failed to delete:', error);
          showMessage?.(error.response?.data?.msg || t('subscriptions.share.messages.deleteFailed'), 'error');
        }
        setConfirmOpen(false);
      }
    });
    setConfirmOpen(true);
  };

  const handleRefreshToken = (share, e) => {
    e?.stopPropagation();
    setConfirmInfo({
      title: t('subscriptions.share.confirm.refreshTitle'),
      content: t('subscriptions.share.confirm.refreshContent'),
      onConfirm: async () => {
        try {
          await refreshShareToken(share.id);
          showMessage?.(t('subscriptions.share.messages.tokenRefreshed'), 'success');
          fetchShares();
          if (detailShare?.id === share.id) {
            setDetailOpen(false);
          }
        } catch (error) {
          console.error('Failed to refresh:', error);
          showMessage?.(error.response?.data?.msg || t('subscriptions.share.messages.refreshFailed'), 'error');
        }
        setConfirmOpen(false);
      }
    });
    setConfirmOpen(true);
  };

  const handleViewLogs = async (share, e) => {
    e?.stopPropagation();
    setLogsShareName(share.name || t('subscriptions.share.unnamed'));
    setLogsLoading(true);
    setLogsOpen(true);
    try {
      const res = await getShareLogs(share.id);
      setLogs(res.data || []);
    } catch (error) {
      console.error('Failed to get logs:', error);
      setLogs([]);
    } finally {
      setLogsLoading(false);
    }
  };

  const handleQrCode = (url, title) => {
    setQrUrl(url);
    setQrTitle(title);
    setQrOpen(true);
  };

  // 批量操作处理函数
  const handleBatchCreate = async (data) => {
    try {
      await batchCreateShares(subscription.ID, data);
      showMessage?.(t('subscriptions.share.batch.createSuccess'), 'success');
      fetchShares();
      setSelectedShares([]);
      setBatchCreateOpen(false);
    } catch (error) {
      console.error('Failed to batch create:', error);
      showMessage?.(error.response?.data?.msg || t('subscriptions.share.batch.createError'), 'error');
    }
  };

  const handleBatchDelete = () => {
    if (selectedShares.length === 0) {
      showMessage?.(t('subscriptions.share.batch.noSelection'), 'warning');
      return;
    }

    setConfirmInfo({
      title: t('subscriptions.share.batch.deleteTitle'),
      content: t('subscriptions.share.batch.deleteConfirm', { count: selectedShares.length }),
      onConfirm: async () => {
        try {
          await batchDeleteShares(selectedShares);
          showMessage?.(t('subscriptions.share.batch.deleteSuccess'), 'success');
          fetchShares();
          setSelectedShares([]);
        } catch (error) {
          console.error('Failed to batch delete:', error);
          showMessage?.(error.response?.data?.msg || t('subscriptions.share.batch.deleteError'), 'error');
        }
        setConfirmOpen(false);
      }
    });
    setConfirmOpen(true);
  };

  const handleBatchUpdateExpire = () => {
    if (selectedShares.length === 0) {
      showMessage?.(t('subscriptions.share.batch.noSelection'), 'warning');
      return;
    }
    setBatchUpdateMode('expire');
    setBatchUpdateOpen(true);
  };

  const handleBatchToggleEnabled = () => {
    if (selectedShares.length === 0) {
      showMessage?.(t('subscriptions.share.batch.noSelection'), 'warning');
      return;
    }
    setBatchUpdateMode('enabled');
    setBatchUpdateOpen(true);
  };

  const handleBatchUpdate = async (data) => {
    try {
      const updates = {};

      if (batchUpdateMode === 'expire') {
        updates.expire_type = data.expireType;
        updates.expire_days = data.expireDays;
        updates.expire_at = data.expireAt;
      } else if (batchUpdateMode === 'enabled') {
        if (data.action === 'enable') {
          updates.enabled = true;
        } else if (data.action === 'disable') {
          updates.enabled = false;
        } else if (data.action === 'toggle') {
          // 反转状态需要分别处理每个分享
          const enableIds = [];
          const disableIds = [];
          selectedShares.forEach((id) => {
            const share = shares.find((s) => s.id === id);
            if (share) {
              if (share.enabled) {
                disableIds.push(id);
              } else {
                enableIds.push(id);
              }
            }
          });

          if (enableIds.length > 0) {
            await batchUpdateShares(enableIds, { enabled: true });
          }
          if (disableIds.length > 0) {
            await batchUpdateShares(disableIds, { enabled: false });
          }

          showMessage?.(t('subscriptions.share.batch.updateSuccess'), 'success');
          fetchShares();
          setSelectedShares([]);
          setBatchUpdateOpen(false);
          return;
        }
      }

      await batchUpdateShares(selectedShares, updates);
      showMessage?.(t('subscriptions.share.batch.updateSuccess'), 'success');
      fetchShares();
      setSelectedShares([]);
      setBatchUpdateOpen(false);
    } catch (error) {
      console.error('Failed to batch update:', error);
      showMessage?.(error.response?.data?.msg || t('subscriptions.share.batch.updateError'), 'error');
    }
  };

  const handleSelectAll = () => {
    setSelectedShares(filteredShares.map((s) => s.id));
  };

  const handleClearSelection = () => {
    setSelectedShares([]);
  };

  const handleToggleShare = (shareId) => {
    setSelectedShares((prev) => (prev.includes(shareId) ? prev.filter((id) => id !== shareId) : [...prev, shareId]));
  };

  const getExpireText = (share) => {
    if (!share.enabled) return t('common.disabled');
    switch (share.expire_type) {
      case EXPIRE_TYPE_NEVER:
        return t('subscriptions.share.expire.never');
      case EXPIRE_TYPE_DAYS:
        return t('subscriptions.share.expire.days', { count: share.expire_days });
      case EXPIRE_TYPE_DATETIME:
        return share.expire_at
          ? formatDateTime(share.expire_at, i18n.resolvedLanguage || i18n.language)
          : t('subscriptions.share.expire.datetime');
      default:
        return t('subscriptions.share.expire.never');
    }
  };

  const isExpired = (share) => {
    if (!share.enabled) return true;
    if (share.expire_type === EXPIRE_TYPE_DAYS && share.expire_days > 0) {
      const created = new Date(share.created_at);
      const expireDate = new Date(created.getTime() + share.expire_days * 24 * 60 * 60 * 1000);
      return new Date() > expireDate;
    }
    if (share.expire_type === EXPIRE_TYPE_DATETIME && share.expire_at) {
      return new Date() > new Date(share.expire_at);
    }
    return false;
  };

  const filteredShares = useMemo(() => shares, [shares]);

  // 搜索处理（防抖后调用后端API）
  const searchTimeoutRef = useRef(null);

  /**
   * 检测输入是否为IP地址
   */
  const isIPAddress = (input) => {
    if (!input || typeof input !== 'string') return false;
    const trimmed = input.trim();

    // IPv4 pattern
    const ipv4Pattern = /^(\d{1,3}\.){3}\d{1,3}$/;
    if (ipv4Pattern.test(trimmed)) {
      // Validate octets are 0-255
      const octets = trimmed.split('.');
      return octets.every((octet) => {
        const num = parseInt(octet, 10);
        return num >= 0 && num <= 255;
      });
    }

    // IPv6 pattern (simplified)
    const ipv6Pattern = /^([0-9a-fA-F]{0,4}:){2,7}[0-9a-fA-F]{0,4}$/;
    return ipv6Pattern.test(trimmed);
  };

  /**
   * 从URL中提取token参数
   * @param {string} input - 用户输入
   * @returns {string} - 提取的token或原始输入
   */
  const extractTokenFromInput = useCallback((input) => {
    if (!input || typeof input !== 'string') return input;

    const trimmed = input.trim();

    // 检测是否为URL（包含协议或域名模式）
    try {
      // 尝试作为完整URL解析
      const url = new URL(trimmed);
      const token = url.searchParams.get('token');
      if (token) {
        return token;
      }
    } catch {
      // 不是完整URL，尝试检测是否包含查询参数
      const tokenMatch = trimmed.match(/[?&]token=([^&\s]+)/);
      if (tokenMatch && tokenMatch[1]) {
        return tokenMatch[1];
      }
    }

    // 如果不是URL或未找到token，返回原始输入（用于名称模糊搜索）
    return trimmed;
  }, []);

  const handleManualRefresh = useCallback(async () => {
    const scrollTop = dialogContentRef.current?.scrollTop || 0;
    const searchKeyword = extractTokenFromInput(searchQuery);

    setRefreshing(true);
    try {
      await fetchShares(searchKeyword, true);
      requestAnimationFrame(() => {
        if (dialogContentRef.current) {
          dialogContentRef.current.scrollTop = scrollTop;
        }
      });
    } finally {
      setRefreshing(false);
    }
  }, [extractTokenFromInput, fetchShares, searchQuery]);

  const handleSearchChange = useCallback(
    (value) => {
      setSearchQuery(value);
      // 防抖后重新获取数据
      if (searchTimeoutRef.current) {
        clearTimeout(searchTimeoutRef.current);
      }
      searchTimeoutRef.current = setTimeout(() => {
        // 尝试从输入中提取token（如果是链接）
        const searchKeyword = extractTokenFromInput(value);
        fetchShares(searchKeyword, true); // isSearch = true，使用 searching 状态而非 loading
        setSelectedShares([]); // 搜索时清空选择
      }, 500);
    },
    [extractTokenFromInput, fetchShares]
  );

  useEffect(() => {
    return () => {
      if (searchTimeoutRef.current) {
        clearTimeout(searchTimeoutRef.current);
      }
    };
  }, []);

  const getDialogPaperSx = (fullScreen = false) => ({
    borderRadius: fullScreen ? 0 : 3,
    overflow: 'hidden',
    bgcolor: dialogSurface,
    backgroundImage: dialogSurfaceGradient,
    border: fullScreen ? 'none' : '1px solid',
    borderColor: panelBorder
  });

  const iconButtonBaseSx = {
    color: secondaryText,
    bgcolor: nestedPanelSurface,
    border: '1px solid',
    borderColor: panelBorder,
    boxShadow: isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.04)}` : 'none',
    transition: 'all 0.2s ease',
    '&:hover': {
      color: primaryText,
      bgcolor: withAlpha(palette.primary.main, isDark ? 0.14 : 0.06),
      borderColor: withAlpha(palette.primary.main, isDark ? 0.34 : 0.2)
    }
  };

  const actionIconButtonSx = {
    ...iconButtonBaseSx,
    width: 32,
    height: 32
  };

  const legacyChipSx = {
    height: 20,
    fontSize: '0.68rem',
    fontWeight: 700,
    bgcolor: withAlpha(palette.primary.main, isDark ? 0.18 : 0.1),
    color: palette.primary.main,
    border: '1px solid',
    borderColor: withAlpha(palette.primary.main, isDark ? 0.38 : 0.22),
    '& .MuiChip-label': {
      px: 0.9
    }
  };

  const renderShareCard = (share) => {
    const expired = isExpired(share);
    const accentColor = share.is_legacy ? palette.primary.main : expired ? palette.error.main : palette.info.main;
    const accentSurface = share.is_legacy
      ? withAlpha(palette.primary.main, isDark ? 0.16 : 0.06)
      : expired
        ? withAlpha(palette.error.main, isDark ? 0.14 : 0.05)
        : nestedPanelSurface;
    const accentBorder = share.is_legacy
      ? withAlpha(palette.primary.main, isDark ? 0.38 : 0.22)
      : expired
        ? withAlpha(palette.error.main, isDark ? 0.34 : 0.2)
        : panelBorder;

    return (
      <Card
        sx={{
          borderRadius: 2.5,
          bgcolor: accentSurface,
          backgroundImage: share.is_legacy
            ? `linear-gradient(180deg, ${withAlpha(palette.primary.main, isDark ? 0.1 : 0.04)} 0%, ${accentSurface} 100%)`
            : 'none',
          border: '1px solid',
          borderColor: accentBorder,
          boxShadow: isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.04)}` : 'none',
          opacity: expired ? 0.72 : 1,
          transition: 'all 0.2s ease',
          '&:hover': {
            borderColor: withAlpha(accentColor, isDark ? 0.48 : 0.28),
            bgcolor: share.is_legacy ? withAlpha(palette.primary.main, isDark ? 0.2 : 0.08) : mutedPanelSurface
          }
        }}
      >
        <CardContent sx={{ px: 2, py: 1.75, '&:last-child': { pb: 1.75 } }}>
          <Stack direction="row" alignItems="center" spacing={1.25}>
            {/* 复选框 */}
            <Checkbox
              checked={selectedShares.includes(share.id)}
              onChange={() => handleToggleShare(share.id)}
              onClick={(e) => e.stopPropagation()}
              sx={{ p: 0 }}
            />

            <Box
              onClick={() => handleOpenDetail(share)}
              sx={{
                display: 'flex',
                alignItems: 'center',
                flex: 1,
                minWidth: 0,
                cursor: 'pointer',
                gap: 1,
                '&:hover': { opacity: 0.92 }
              }}
            >
              <Box
                sx={{
                  width: 34,
                  height: 34,
                  borderRadius: 1.75,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  flexShrink: 0,
                  bgcolor: withAlpha(accentColor, isDark ? 0.16 : 0.08),
                  color: expired ? tertiaryText : accentColor,
                  border: '1px solid',
                  borderColor: withAlpha(accentColor, isDark ? 0.32 : 0.18)
                }}
              >
                <LinkIcon fontSize="small" />
              </Box>
              <Box sx={{ flex: 1, minWidth: 0 }}>
                <Stack direction="row" alignItems="center" spacing={0.75} sx={{ mb: 0.35 }}>
                  <Typography variant="body2" fontWeight={600} noWrap sx={{ color: primaryText }}>
                    {share.name || t('subscriptions.share.unnamed')}
                  </Typography>
                  {share.is_legacy && <Chip label={t('subscriptions.share.defaultChip')} size="small" sx={legacyChipSx} />}
                </Stack>
                <Typography variant="caption" sx={{ color: expired ? tertiaryText : secondaryText }}>
                  {t('subscriptions.share.cardMeta', { expire: getExpireText(share), count: share.access_count || 0 })}
                </Typography>
              </Box>
            </Box>
            <Stack direction="row" spacing={0.5}>
              <Tooltip title={t('subscriptions.share.actions.accessLogs')}>
                <IconButton size="small" onClick={(e) => handleViewLogs(share, e)} sx={actionIconButtonSx}>
                  <HistoryIcon fontSize="small" />
                </IconButton>
              </Tooltip>
              <Tooltip title={t('common.edit')}>
                <IconButton size="small" onClick={(e) => handleEdit(share, e)} sx={actionIconButtonSx}>
                  <EditIcon fontSize="small" />
                </IconButton>
              </Tooltip>
              {share.is_legacy ? (
                <Tooltip title={t('subscriptions.share.actions.refreshToken')}>
                  <IconButton
                    size="small"
                    onClick={(e) => handleRefreshToken(share, e)}
                    sx={{
                      ...actionIconButtonSx,
                      color: palette.warning.main,
                      '&:hover': {
                        ...actionIconButtonSx['&:hover'],
                        color: palette.warning.main,
                        bgcolor: withAlpha(palette.warning.main, isDark ? 0.16 : 0.08),
                        borderColor: withAlpha(palette.warning.main, isDark ? 0.36 : 0.2)
                      }
                    }}
                  >
                    <RefreshIcon fontSize="small" />
                  </IconButton>
                </Tooltip>
              ) : (
                <Tooltip title={t('common.delete')}>
                  <IconButton
                    size="small"
                    onClick={(e) => handleDelete(share, e)}
                    sx={{
                      ...actionIconButtonSx,
                      color: palette.error.main,
                      '&:hover': {
                        ...actionIconButtonSx['&:hover'],
                        color: palette.error.main,
                        bgcolor: withAlpha(palette.error.main, isDark ? 0.16 : 0.08),
                        borderColor: withAlpha(palette.error.main, isDark ? 0.34 : 0.2)
                      }
                    }}
                  >
                    <DeleteIcon fontSize="small" />
                  </IconButton>
                </Tooltip>
              )}
            </Stack>
          </Stack>
        </CardContent>
      </Card>
    );
  };

  const visibleClientLinks = useMemo(() => {
    const allowedTargets = new Set(subStoreTargets);
    return [...NATIVE_CLIENT_LINKS, ...EXPANDED_CLIENT_LINKS.filter((item) => allowedTargets.has(item.client))];
  }, [subStoreTargets]);

  const detailClientUrls = detailShare
    ? visibleClientLinks.reduce(
        (urls, item) => ({
          ...urls,
          [t(`subscriptions.share.clients.${item.key}`)]: `${getServerUrl()}/c/?token=${detailShare.token}&client=${item.client}`
        }),
        {
          [t('subscriptions.share.clients.auto')]: `${getServerUrl()}/c/?token=${detailShare.token}`
        }
      )
    : {};

  return (
    <>
      <Dialog
        open={open}
        onClose={onClose}
        maxWidth="sm"
        fullWidth
        fullScreen={isMobile}
        slotProps={{
          paper: {
            sx: getDialogPaperSx(isMobile)
          }
        }}
      >
        <DialogTitle
          sx={{
            px: 2.5,
            py: 1.75,
            bgcolor: mutedPanelSurface,
            borderBottom: '1px solid',
            borderColor: panelBorder,
            boxShadow: `inset 0 -1px 0 ${withAlpha(palette.divider, 0.4)}`
          }}
        >
          <Stack spacing={1.5}>
            <Stack direction="row" alignItems="center" justifyContent="space-between">
              <Typography variant="h6">{t('subscriptions.share.title', { name: subscription?.Name })}</Typography>
              <Stack direction="row" spacing={1}>
                <IconButton size="small" onClick={handleManualRefresh} disabled={loading || searching || refreshing} sx={iconButtonBaseSx}>
                  {refreshing ? <CircularProgress size={18} /> : <RefreshIcon fontSize="small" />}
                </IconButton>
                <Button variant="contained" size="small" startIcon={<AddIcon />} onClick={handleAdd}>
                  {t('subscriptions.share.addSingle')}
                </Button>
                <Button
                  variant="outlined"
                  size="small"
                  startIcon={<AddIcon />}
                  onClick={() => {
                    setSearchQuery('');
                    setSelectedShares([]);
                    setBatchCreateOpen(true);
                  }}
                >
                  {t('subscriptions.share.batch.create')}
                </Button>
              </Stack>
            </Stack>

            {/* 批量操作工具栏 */}
            {selectedShares.length > 0 && (
              <Box sx={{ p: 1.5, bgcolor: 'action.selected', borderRadius: 1 }}>
                <Stack direction="row" spacing={1.5} alignItems="center" flexWrap="wrap">
                  <Typography variant="body2" sx={{ color: primaryText }}>
                    {t('subscriptions.share.batch.selectedCount', { count: selectedShares.length })}
                  </Typography>
                  <Button size="small" onClick={handleSelectAll}>
                    {t('subscriptions.share.batch.selectAll')}
                  </Button>
                  <Button size="small" onClick={handleClearSelection}>
                    {t('subscriptions.share.batch.clearSelection')}
                  </Button>
                  <Divider orientation="vertical" flexItem />
                  <Button size="small" color="error" startIcon={<DeleteIcon />} onClick={handleBatchDelete}>
                    {t('subscriptions.share.batch.delete')}
                  </Button>
                  <Button size="small" onClick={handleBatchUpdateExpire}>
                    {t('subscriptions.share.batch.updateExpire')}
                  </Button>
                  <Button size="small" onClick={handleBatchToggleEnabled}>
                    {t('subscriptions.share.batch.toggleEnabled')}
                  </Button>
                  <Button size="small" startIcon={<DownloadIcon />} onClick={() => setExportDialogOpen(true)}>
                    {t('subscriptions.share.batch.export')}
                  </Button>
                </Stack>
              </Box>
            )}

            <Stack direction="row" spacing={1} alignItems="center">
              <TextField
                size="small"
                fullWidth
                placeholder={t('subscriptions.share.searchPlaceholder')}
                value={searchQuery}
                onChange={(e) => handleSearchChange(e.target.value)}
                sx={{
                  '& .MuiOutlinedInput-root': {
                    bgcolor: nestedPanelSurface,
                    '&:hover': {
                      bgcolor: withAlpha(palette.primary.main, isDark ? 0.08 : 0.04)
                    }
                  }
                }}
              />
            </Stack>

            {/* 排序控件 */}
            {shares.length > 0 && (
              <Box
                sx={{
                  mt: 1.5,
                  display: 'flex',
                  alignItems: 'center',
                  gap: 1
                }}
              >
                <Typography variant="caption" sx={{ color: secondaryText, fontSize: '0.8rem' }}>
                  {t('subscriptions.share.sortLabel')}:
                </Typography>
                {['access_count', 'name'].map((field) => {
                  const active = sortBy === field;
                  const labelKey =
                    field === 'access_count'
                      ? active
                        ? sortOrder === 'desc'
                          ? 'sort.accessCountDesc'
                          : 'sort.accessCountAsc'
                        : 'sort.accessCount'
                      : active
                        ? sortOrder === 'desc'
                          ? 'sort.nameDesc'
                          : 'sort.nameAsc'
                        : 'sort.name';

                  return (
                    <Button
                      key={field}
                      size="small"
                      startIcon={<SortIcon />}
                      onClick={() => handleSort(field)}
                      sx={{
                        minWidth: 'auto',
                        px: 1.5,
                        py: 0.5,
                        fontSize: '0.75rem',
                        textTransform: 'none',
                        color: active ? palette.primary.main : secondaryText,
                        bgcolor: active ? withAlpha(palette.primary.main, isDark ? 0.16 : 0.08) : nestedPanelSurface,
                        border: '1px solid',
                        borderColor: active ? withAlpha(palette.primary.main, isDark ? 0.38 : 0.22) : panelBorder,
                        '&:hover': {
                          bgcolor: active
                            ? withAlpha(palette.primary.main, isDark ? 0.22 : 0.12)
                            : withAlpha(palette.primary.main, isDark ? 0.08 : 0.04),
                          borderColor: withAlpha(palette.primary.main, isDark ? 0.42 : 0.28)
                        }
                      }}
                    >
                      {t(`subscriptions.share.${labelKey}`)}
                    </Button>
                  );
                })}
              </Box>
            )}
          </Stack>
        </DialogTitle>

        <DialogContent
          ref={dialogContentRef}
          sx={{
            px: 2.5,
            pt: 2.5,
            pb: 2,
            bgcolor: dialogSurface
          }}
        >
          {loading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', py: 4.25 }}>
              <CircularProgress />
            </Box>
          ) : shares.length === 0 ? (
            <Alert
              variant="outlined"
              severity="info"
              sx={{
                mt: 1.5,
                bgcolor: withAlpha(palette.info.main, isDark ? 0.12 : 0.05),
                borderColor: withAlpha(palette.info.main, isDark ? 0.3 : 0.18)
              }}
            >
              {searching ? t('subscriptions.share.searching') : t('subscriptions.share.empty')}
            </Alert>
          ) : filteredShares.length === 0 ? (
            <Alert
              variant="outlined"
              severity="info"
              sx={{
                mt: 1.5,
                bgcolor: withAlpha(palette.info.main, isDark ? 0.12 : 0.05),
                borderColor: withAlpha(palette.info.main, isDark ? 0.3 : 0.18)
              }}
            >
              {t('subscriptions.share.noResults')}
            </Alert>
          ) : (
            <Stack spacing={1.5} sx={{ mt: 1.5, position: 'relative' }}>
              {searching && (
                <Box
                  sx={{
                    position: 'absolute',
                    top: 0,
                    left: 0,
                    right: 0,
                    bottom: 0,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    bgcolor: withAlpha(dialogSurface, 0.8),
                    zIndex: 1,
                    borderRadius: 1
                  }}
                >
                  <CircularProgress size={32} />
                </Box>
              )}
              {filteredShares.map((share) => (
                <React.Fragment key={share.id}>{renderShareCard(share)}</React.Fragment>
              ))}

              {/* 无限滚动加载指示器 */}
              {hasMore && (
                <Box
                  ref={(el) => {
                    if (!el || !hasMore) return;
                    const observer = new IntersectionObserver(
                      (entries) => {
                        if (entries[0].isIntersecting) {
                          loadMoreShares();
                        }
                      },
                      { threshold: 0.1 }
                    );
                    observer.observe(el);
                    return () => observer.disconnect();
                  }}
                  sx={{ textAlign: 'center', py: 2 }}
                >
                  {loadingMore ? (
                    <CircularProgress size={24} />
                  ) : (
                    <Typography variant="body2" color="textSecondary">
                      {t('subscriptions.share.loadMore')}
                    </Typography>
                  )}
                </Box>
              )}

              {!hasMore && shares.length > 0 && (
                <Typography variant="body2" color="textSecondary" sx={{ textAlign: 'center', py: 2 }}>
                  {t('subscriptions.share.noMore')}
                </Typography>
              )}
            </Stack>
          )}
        </DialogContent>

        <DialogActions sx={{ px: 2.5, py: 1.5, bgcolor: mutedPanelSurface, borderTop: '1px solid', borderColor: panelBorder }}>
          <Button onClick={onClose} variant="outlined">
            {t('common.close')}
          </Button>
        </DialogActions>
      </Dialog>

      <ClientUrlsDialog
        open={detailOpen}
        title={detailShare?.name || t('subscriptions.share.detailTitle')}
        legacy={Boolean(detailShare?.is_legacy)}
        clientUrls={detailClientUrls}
        onClose={() => setDetailOpen(false)}
        onQrCode={handleQrCode}
        onCopy={copyToClipboard}
      />

      <Dialog
        open={formOpen}
        onClose={() => setFormOpen(false)}
        maxWidth="xs"
        fullWidth
        slotProps={{
          paper: {
            sx: getDialogPaperSx(false)
          }
        }}
      >
        <DialogTitle
          sx={{
            px: 2.5,
            py: 2,
            bgcolor: mutedPanelSurface,
            borderBottom: '1px solid',
            borderColor: panelBorder
          }}
        >
          {editingShare ? t('subscriptions.share.form.editTitle') : t('subscriptions.share.form.addTitle')}
        </DialogTitle>
        <DialogContent sx={{ bgcolor: dialogSurface }}>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <TextField
              label={t('subscriptions.share.form.name')}
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder={t('subscriptions.share.form.namePlaceholder')}
              size="small"
              fullWidth
            />

            <TextField
              label={t('subscriptions.share.form.token')}
              value={formData.token}
              onChange={(e) => setFormData({ ...formData, token: e.target.value })}
              placeholder={t('subscriptions.share.form.tokenPlaceholder')}
              size="small"
              fullWidth
              helperText={t('subscriptions.share.form.tokenHelper')}
            />

            <FormControl size="small" fullWidth>
              <InputLabel>{t('subscriptions.share.form.expirePolicy')}</InputLabel>
              <Select
                value={formData.expire_type}
                label={t('subscriptions.share.form.expirePolicy')}
                onChange={(e) => setFormData({ ...formData, expire_type: e.target.value })}
              >
                <MenuItem value={EXPIRE_TYPE_NEVER}>{t('subscriptions.share.expire.never')}</MenuItem>
                <MenuItem value={EXPIRE_TYPE_DAYS}>{t('subscriptions.share.form.expireByDays')}</MenuItem>
                <MenuItem value={EXPIRE_TYPE_DATETIME}>{t('subscriptions.share.form.expireAt')}</MenuItem>
              </Select>
            </FormControl>

            {formData.expire_type === EXPIRE_TYPE_DAYS && (
              <TextField
                label={t('subscriptions.share.form.expireDays')}
                type="number"
                value={formData.expire_days}
                onChange={(e) => setFormData({ ...formData, expire_days: parseInt(e.target.value) || 0 })}
                size="small"
                fullWidth
                inputProps={{ min: 1 }}
              />
            )}

            {formData.expire_type === EXPIRE_TYPE_DATETIME && (
              <TextField
                label={t('subscriptions.share.form.expireTime')}
                type="datetime-local"
                value={formData.expire_at}
                onChange={(e) => setFormData({ ...formData, expire_at: e.target.value })}
                size="small"
                fullWidth
                InputLabelProps={{ shrink: true }}
              />
            )}

            {editingShare && (
              <FormControlLabel
                control={<Switch checked={formData.enabled} onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })} />}
                label={t('subscriptions.share.form.enabled')}
              />
            )}
          </Stack>
        </DialogContent>
        <DialogActions sx={{ px: 2.5, py: 1.75, bgcolor: mutedPanelSurface, borderTop: '1px solid', borderColor: panelBorder }}>
          <Button onClick={() => setFormOpen(false)}>{t('common.cancel')}</Button>
          <Button variant="contained" onClick={handleSave}>
            {t('common.save')}
          </Button>
        </DialogActions>
      </Dialog>

      <AccessLogsDialog
        open={logsOpen}
        logs={logs}
        loading={logsLoading}
        title={t('subscriptions.share.logsTitle', { name: logsShareName })}
        onClose={() => setLogsOpen(false)}
      />

      <QrCodeDialog open={qrOpen} title={qrTitle} url={qrUrl} onClose={() => setQrOpen(false)} onCopy={copyToClipboard} />

      <ConfirmDialog
        open={confirmOpen}
        title={confirmInfo.title}
        content={confirmInfo.content}
        onClose={() => setConfirmOpen(false)}
        onConfirm={confirmInfo.onConfirm}
      />

      <ShareBatchCreateDialog
        open={batchCreateOpen}
        subscription={subscription}
        existingNames={shares.map((s) => s.name)}
        onClose={() => setBatchCreateOpen(false)}
        onSubmit={handleBatchCreate}
      />

      <ShareBatchUpdateDialog
        open={batchUpdateOpen}
        mode={batchUpdateMode}
        shares={shares.filter((s) => selectedShares.includes(s.id))}
        onClose={() => setBatchUpdateOpen(false)}
        onSubmit={handleBatchUpdate}
      />

      <ShareExportDialog
        open={exportDialogOpen}
        onClose={() => setExportDialogOpen(false)}
        selectedShares={shares.filter((s) => selectedShares.includes(s.id))}
        serverUrl={getServerUrl()}
        subStoreTargets={subStoreTargets}
      />
    </>
  );
}
