import { useState, useEffect, useCallback } from 'react';

// material-ui
import { useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Alert from '@mui/material/Alert';
import Autocomplete from '@mui/material/Autocomplete';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import FormControl from '@mui/material/FormControl';
import IconButton from '@mui/material/IconButton';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import Snackbar from '@mui/material/Snackbar';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import ToggleButton from '@mui/material/ToggleButton';
import ToggleButtonGroup from '@mui/material/ToggleButtonGroup';
import Tooltip from '@mui/material/Tooltip';

// icons
import AddIcon from '@mui/icons-material/Add';
import CloudSyncIcon from '@mui/icons-material/CloudSync';
import RefreshIcon from '@mui/icons-material/Refresh';
import ViewModuleIcon from '@mui/icons-material/ViewModule';
import ViewListIcon from '@mui/icons-material/ViewList';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import Pagination from 'components/Pagination';
import ConfirmDialog from 'components/ConfirmDialog';
import TaskProgressPanel from 'components/TaskProgressPanel';
import { getAirports, addAirport, updateAirport, deleteAirport, pullAirport, pullAllAirports, refreshAirportUsage } from 'api/airports';
import { useTaskProgress } from 'contexts/TaskProgressContext';
import { getNodeGroups, getNodes, getNodeProtocols } from 'api/nodes';

// local components
import { AirportTable, AirportListView, AirportMobileList, AirportFormDialog, DeleteAirportDialog } from './component';

// utils
import { validateCronExpression } from './utils';

// ==============================|| 机场管理 ||============================== //

export default function AirportList() {
  const theme = useTheme();
  const matchDownMd = useMediaQuery(theme.breakpoints.down('md'));

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

    const boolKeys = ['nodeNameUniquify'];
    for (const key of boolKeys) {
      if (!!before[key] !== !!after[key]) return true;
    }

    return false;
  };

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

  // 表单状态
  const [formOpen, setFormOpen] = useState(false);
  const [isEdit, setIsEdit] = useState(false);
  const [airportForm, setAirportForm] = useState({
    id: 0,
    name: '',
    url: '',
    cronExpr: '0 */12 * * *',
    enabled: true,
    group: '',
    downloadWithProxy: false,
    proxyLink: '',
    userAgent: '',
    fetchUsageInfo: false,
    skipTLSVerify: false,
    remark: '',
    logo: '',
    nodeNameWhitelist: '',
    nodeNameBlacklist: '',
    protocolWhitelist: '',
    protocolBlacklist: '',
    nodeNamePreprocess: '',
    deduplicationRule: '',
    nodeNameUniquify: false,
    nodeNamePrefix: ''
  });
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

  // 消息提示
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });

  // 显示消息
  const showMessage = useCallback((message, severity = 'success') => {
    setSnackbar({ open: true, message, severity });
  }, []);

  // 获取机场列表
  const fetchAirports = useCallback(async () => {
    setLoading(true);
    try {
      const params = {
        page: page + 1,
        pageSize: rowsPerPage
      };
      // 添加搜索参数
      if (searchKeyword) params.keyword = searchKeyword;
      if (searchGroup) params.group = searchGroup;
      if (searchEnabled !== '') params.enabled = searchEnabled;

      const response = await getAirports(params);
      if (response.data?.items) {
        setAirports(response.data.items);
        setTotalItems(response.data.total || 0);
      } else if (Array.isArray(response.data)) {
        setAirports(response.data);
        setTotalItems(response.data.length);
      }
    } catch (error) {
      console.error('获取机场列表失败:', error);
      showMessage(error.message || '获取机场列表失败', 'error');
    } finally {
      setLoading(false);
    }
  }, [page, rowsPerPage, searchKeyword, searchGroup, searchEnabled, showMessage]);

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
      const response = await getNodeProtocols();
      setProtocolOptions(response.data || []);
    } catch (error) {
      console.error('获取协议列表失败:', error);
    }
  }, []);

  // 初始化
  useEffect(() => {
    fetchAirports();
    fetchGroupOptions();
    fetchProtocolOptions();
  }, [fetchAirports, fetchGroupOptions, fetchProtocolOptions]);

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
    const newForm = {
      id: 0,
      name: '',
      url: '',
      cronExpr: '0 */12 * * *',
      enabled: true,
      group: '',
      downloadWithProxy: false,
      proxyLink: '',
      userAgent: '',
      fetchUsageInfo: false,
      skipTLSVerify: false,
      remark: '',
      logo: '',
      nodeNameWhitelist: '',
      nodeNameBlacklist: '',
      protocolWhitelist: '',
      protocolBlacklist: '',
      nodeNamePreprocess: '',
      deduplicationRule: '',
      nodeNameUniquify: false,
      nodeNamePrefix: ''
    };
    setIsEdit(false);
    setAirportForm(newForm);
    setAirportFormSnapshot(newForm);
    setFormOpen(true);
  };

  // 编辑机场
  const handleEdit = (airport) => {
    const editForm = {
      id: airport.id,
      name: airport.name,
      url: airport.url,
      cronExpr: airport.cronExpr,
      enabled: airport.enabled,
      group: airport.group || '',
      downloadWithProxy: airport.downloadWithProxy || false,
      proxyLink: airport.proxyLink || '',
      userAgent: airport.userAgent || '',
      fetchUsageInfo: airport.fetchUsageInfo || false,
      skipTLSVerify: airport.skipTLSVerify || false,
      remark: airport.remark || '',
      logo: airport.logo || '',
      nodeNameWhitelist: airport.nodeNameWhitelist || '',
      nodeNameBlacklist: airport.nodeNameBlacklist || '',
      protocolWhitelist: airport.protocolWhitelist || '',
      protocolBlacklist: airport.protocolBlacklist || '',
      nodeNamePreprocess: airport.nodeNamePreprocess || '',
      deduplicationRule: airport.deduplicationRule || '',
      nodeNameUniquify: airport.nodeNameUniquify || false,
      nodeNamePrefix: airport.nodeNamePrefix || ''
    };
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
      showMessage(deleteWithNodes ? '已删除机场及关联节点' : '已删除机场（保留节点）');
      fetchAirports();
    } catch (error) {
      console.error('删除失败:', error);
      showMessage(error.message || '删除失败', 'error');
    }
    setDeleteDialogOpen(false);
    setDeleteTarget(null);
  };

  // 拉取机场
  const handlePull = (airport) => {
    openConfirm('立即更新', `确定要立即更新机场 "${airport.name}" 的订阅吗？`, async () => {
      try {
        await pullAirport(airport.id);
        showMessage('已提交更新任务，请稍后刷新查看结果');
        // 任务完成后会自动触发刷新
      } catch (error) {
        console.error('拉取失败:', error);
        showMessage(error.message || '提交更新任务失败', 'error');
      }
    });
  };

  // 批量拉取所有机场
  const handlePullAll = async () => {
    // 这里不要用当前页 airports 来统计启用数量：列表是分页的，会导致误判“没有已启用机场”
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
      showMessage('没有已启用的机场', 'warning');
      return;
    }

    const confirmContent =
      typeof enabledTotal === 'number'
        ? `确定要立即拉取所有已启用的机场订阅吗？（共 ${enabledTotal} 个）`
        : '确定要立即拉取所有已启用的机场订阅吗？';

    openConfirm('拉取所有机场', confirmContent, async () => {
      try {
        const res = await pullAllAirports();
        const count = res.data?.count;

        if (typeof count === 'number') {
          if (count > 0) {
            showMessage(`已提交 ${count} 个机场的拉取任务，请稍后刷新查看结果`);
          } else {
            showMessage('没有已启用的机场', 'warning');
          }
          return;
        }

        // 兼容后端返回 OkWithMsg 的场景（data 为空）
        if (res.msg && res.msg !== '操作成功') {
          showMessage(res.msg, 'warning');
          return;
        }

        showMessage('批量拉取任务已提交，请稍后刷新查看结果');
      } catch (error) {
        console.error('批量拉取失败:', error);
        showMessage(error.message || '批量拉取失败', 'error');
      }
    });
  };

  // 刷新用量信息
  const handleRefreshUsage = async (airport) => {
    try {
      await refreshAirportUsage(airport.id);
      showMessage('用量信息已更新');
      fetchAirports(); // 刷新列表
    } catch (error) {
      console.error('刷新用量失败:', error);
      showMessage(error.message || '刷新用量失败', 'error');
    }
  };

  // 提交表单
  const handleSubmit = async () => {
    // 验证
    if (!airportForm.name.trim()) {
      showMessage('请输入机场名称', 'warning');
      return;
    }
    if (!airportForm.url.trim()) {
      showMessage('请输入订阅地址', 'warning');
      return;
    }
    const urlPattern = /^(https?|ftp):\/\/[^\s/$.?#].[^\s]*$/i;
    if (!urlPattern.test(airportForm.url.trim())) {
      showMessage('请输入有效的订阅地址', 'warning');
      return;
    }
    if (!airportForm.cronExpr.trim()) {
      showMessage('请输入Cron表达式', 'warning');
      return;
    }
    if (!validateCronExpression(airportForm.cronExpr.trim())) {
      showMessage('Cron表达式格式不正确，格式为：分 时 日 月 周', 'error');
      return;
    }

    try {
      // 在提交前计算配置变更状态（提交后 snapshot 会被清空）
      const needPullToApply = isEdit && hasNodeProcessConfigChanged(airportFormSnapshot, airportForm);
      const savedId = airportForm.id;
      const savedName = airportForm.name;

      if (isEdit) {
        await updateAirport(airportForm.id, airportForm);
        showMessage(needPullToApply ? '更新成功（节点处理配置需重新拉取后生效）' : '更新成功');
      } else {
        await addAirport(airportForm);
        showMessage('添加成功');
      }
      setFormOpen(false);
      setAirportFormSnapshot(null);
      fetchAirports();

      // 节点处理配置变更后提示立即拉取，使其对”已存在节点”生效
      if (needPullToApply) {
        openConfirm(
          '立即拉取以生效',
          `节点处理配置已变更，需要重新拉取订阅才能应用到已存在的节点。是否立即拉取机场 “${savedName}”？`,
          async () => {
            try {
              await pullAirport(savedId);
              showMessage('已提交更新任务，请稍后刷新查看结果');
            } catch (error) {
              console.error('拉取失败:', error);
              showMessage(error.message || '提交更新任务失败', 'error');
            }
          }
        );
      }
    } catch (error) {
      console.error('提交失败:', error);
      showMessage(error.message || (isEdit ? '更新失败' : '添加失败'), 'error');
    }
  };

  return (
    <MainCard
      title="机场管理"
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
            添加机场
          </Button>
          <Tooltip title="拉取所有已启用机场的订阅" arrow>
            <IconButton onClick={handlePullAll} color="primary">
              <CloudSyncIcon />
            </IconButton>
          </Tooltip>
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
            label="关键字"
            placeholder="搜索名称或备注"
            value={searchKeyword}
            onChange={(e) => setSearchKeyword(e.target.value)}
            sx={{ minWidth: 150 }}
          />
          <Autocomplete
            size="small"
            options={groupOptions}
            value={searchGroup || null}
            onChange={(e, newValue) => setSearchGroup(newValue || '')}
            renderInput={(params) => <TextField {...params} label="分组" />}
            sx={{ minWidth: 150 }}
          />
          <FormControl size="small" sx={{ minWidth: 100 }}>
            <InputLabel>状态</InputLabel>
            <Select value={searchEnabled} label="状态" onChange={(e) => setSearchEnabled(e.target.value)}>
              <MenuItem value="">全部</MenuItem>
              <MenuItem value="true">启用</MenuItem>
              <MenuItem value="false">禁用</MenuItem>
            </Select>
          </FormControl>
          <Button
            variant="outlined"
            size="small"
            onClick={() => {
              setSearchKeyword('');
              setSearchGroup('');
              setSearchEnabled('');
              setPage(0);
            }}
          >
            重置
          </Button>
          <Button
            variant="contained"
            size="small"
            onClick={() => {
              setPage(0);
              fetchAirports();
            }}
          >
            搜索
          </Button>
        </Stack>
      </Box>

      {/* 任务进度显示 */}
      <TaskProgressPanel />

      {/* 机场列表 */}
      {matchDownMd ? (
        <AirportMobileList
          airports={airports}
          onEdit={handleEdit}
          onDelete={handleDelete}
          onPull={handlePull}
          onRefreshUsage={handleRefreshUsage}
        />
      ) : viewMode === 'list' ? (
        <AirportListView
          airports={airports}
          onEdit={handleEdit}
          onDelete={handleDelete}
          onPull={handlePull}
          onRefreshUsage={handleRefreshUsage}
        />
      ) : (
        <AirportTable
          airports={airports}
          onEdit={handleEdit}
          onDelete={handleDelete}
          onPull={handlePull}
          onRefreshUsage={handleRefreshUsage}
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
        onClose={() => {
          setFormOpen(false);
          setAirportFormSnapshot(null);
        }}
        onSubmit={handleSubmit}
        onFetchProxyNodes={fetchProxyNodes}
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
