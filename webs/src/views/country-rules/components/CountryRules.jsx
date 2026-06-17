import { useState, useEffect, useCallback } from 'react';
import { useTranslation, Trans } from 'react-i18next';
import { useTheme, alpha } from '@mui/material/styles';
import PropTypes from 'prop-types';

// material-ui
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Paper from '@mui/material/Paper';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import TableSortLabel from '@mui/material/TableSortLabel';
import IconButton from '@mui/material/IconButton';
import Chip from '@mui/material/Chip';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import TextField from '@mui/material/TextField';
import Switch from '@mui/material/Switch';
import FormControlLabel from '@mui/material/FormControlLabel';
import Alert from '@mui/material/Alert';
import Typography from '@mui/material/Typography';
import Tooltip from '@mui/material/Tooltip';
import Stack from '@mui/material/Stack';
import CircularProgress from '@mui/material/CircularProgress';
import ToggleButton from '@mui/material/ToggleButton';
import ToggleButtonGroup from '@mui/material/ToggleButtonGroup';
import InputAdornment from '@mui/material/InputAdornment';

// icons
import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import RefreshIcon from '@mui/icons-material/Refresh';
import TableViewIcon from '@mui/icons-material/TableView';
import CodeIcon from '@mui/icons-material/Code';
import SaveIcon from '@mui/icons-material/Save';
import UndoIcon from '@mui/icons-material/Undo';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import SearchIcon from '@mui/icons-material/Search';

// project imports
import {
  getCountryRules,
  createCountryRule,
  updateCountryRule,
  deleteCountryRule,
  testCountryRule,
  exportCountryRules,
  syncCountryRules
} from 'api/countryRules';
import Pagination from 'components/Pagination';
import { withAlpha } from 'utils/colorUtils';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';

// ==============================|| UTILITY FUNCTIONS ||============================== //

/**
 * Convert country code to flag emoji
 * @param {string} countryCode - Two-letter country code (e.g., 'US', 'CN', 'HK')
 * @returns {string} Flag emoji (e.g., '🇺🇸', '🇨🇳', '🇭🇰')
 */
const getCountryFlag = (countryCode) => {
  if (!countryCode || countryCode.length !== 2) return '';

  // 特殊处理：TW 使用中国国旗
  if (countryCode.toUpperCase() === 'TW') {
    return '🇨🇳';
  }

  const codePoints = countryCode
    .toUpperCase()
    .split('')
    .map((char) => 127397 + char.charCodeAt());
  return String.fromCodePoint(...codePoints);
};

// ==============================|| COUNTRY RULES MANAGEMENT ||============================== //

const CountryRules = ({ showMessage }) => {
  const theme = useTheme();
  const { t } = useTranslation();
  const { isDark } = useResolvedColorScheme();
  const palette = theme.vars?.palette || theme.palette;

  // 模式切换：table（表单模式）或 text（文本编辑模式）
  const [viewMode, setViewMode] = useState('table');

  const [rules, setRules] = useState([]);
  const [loading, setLoading] = useState(false);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [currentRule, setCurrentRule] = useState(null);
  const [deleteId, setDeleteId] = useState(null);
  const [orderBy, setOrderBy] = useState('priority');
  const [order, setOrder] = useState('desc');

  // 分页状态
  const [page, setPage] = useState(0); // 0-indexed for MUI
  const [rowsPerPage, setRowsPerPage] = useState(() => {
    const saved = localStorage.getItem('country_rules_rowsPerPage');
    return saved ? parseInt(saved, 10) : 20;
  });
  const [totalItems, setTotalItems] = useState(0);

  // 搜索状态
  const [searchKeyword, setSearchKeyword] = useState('');

  // Form state
  const [formData, setFormData] = useState({
    countryCode: '',
    countryName: '',
    pattern: '',
    priority: 0,
    enabled: true
  });

  // Validation errors
  const [errors, setErrors] = useState({
    countryCode: ''
  });

  // Test state
  const [testName, setTestName] = useState('');
  const [testResult, setTestResult] = useState(null);
  const [testing, setTesting] = useState(false);

  // 文本编辑模式相关状态
  const [textContent, setTextContent] = useState('');
  const [originalText, setOriginalText] = useState('');
  const [syncing, setSyncing] = useState(false);

  // Load rules list
  const loadRules = useCallback(
    async (options = {}) => {
      const { page: p, pageSize: ps, keyword: kw } = options;
      setLoading(true);
      try {
        const params = {
          page: p !== undefined ? p : page + 1, // 后端是 1-indexed
          pageSize: ps !== undefined ? ps : rowsPerPage
        };
        const keyword = kw !== undefined ? kw : searchKeyword;
        if (keyword) {
          params.keyword = keyword;
        }

        const response = await getCountryRules(params);
        if (response.data?.items) {
          setRules(response.data.items);
          setTotalItems(response.data.total || 0);
        } else {
          // 兼容旧格式（无分页）
          setRules(response.data || []);
          setTotalItems((response.data || []).length);
        }
      } catch {
        showMessage(t('countryRules.messages.loadFailed'), 'error');
      } finally {
        setLoading(false);
      }
    },
    [page, rowsPerPage, searchKeyword, showMessage, t]
  );

  // 加载文本内容
  const loadTextContent = async () => {
    try {
      const res = await exportCountryRules();
      const text = res.data?.text || '';
      setTextContent(text);
      setOriginalText(text);
    } catch (error) {
      showMessage(error.message || t('countryRules.messages.loadTextFailed'), 'error');
    }
  };

  useEffect(() => {
    loadRules();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // 搜索触发（实时搜索，带防抖）
  useEffect(() => {
    const timer = setTimeout(() => {
      setPage(0); // 搜索时重置到第一页
      loadRules({ page: 1, pageSize: rowsPerPage, keyword: searchKeyword });
    }, 300); // 300ms 防抖

    return () => clearTimeout(timer);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchKeyword]);

  // 切换到文本模式时加载内容
  useEffect(() => {
    if (viewMode === 'text') {
      loadTextContent();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [viewMode]);

  // Handle sorting
  const handleRequestSort = (property) => {
    const isAsc = orderBy === property && order === 'asc';
    setOrder(isAsc ? 'desc' : 'asc');
    setOrderBy(property);
  };

  // Sort comparator (仅用于客户端排序显示，后端已按优先级排序)
  const getComparator = (order, orderBy) => {
    return order === 'desc' ? (a, b) => (b[orderBy] < a[orderBy] ? -1 : 1) : (a, b) => (a[orderBy] < b[orderBy] ? -1 : 1);
  };

  // 获取状态 Chip 样式（统一为机场列表样式）
  const getStatusChipSx = (enabled) => ({
    height: 20,
    fontSize: '0.7rem',
    fontWeight: 600,
    color: enabled ? (isDark ? palette.success.light : palette.success.dark) : palette.text.secondary,
    bgcolor: enabled
      ? alpha(theme.palette.success.main, isDark ? 0.18 : 0.12)
      : isDark
        ? 'background.default'
        : alpha(theme.palette.grey[500], 0.08),
    border: `1px solid ${enabled ? alpha(theme.palette.success.main, isDark ? 0.34 : 0.18) : alpha(theme.palette.divider, isDark ? 0.56 : 0.24)}`,
    '& .MuiChip-label': {
      px: 1
    }
  });

  // Validate country code uniqueness
  const validateCountryCode = (code, editingId) => {
    const upperCode = code.toUpperCase();
    const exists = rules.some((rule) => rule.countryCode === upperCode && rule.id !== editingId);
    return !exists;
  };

  // Open add/edit dialog
  const handleOpenDialog = (rule = null) => {
    if (rule) {
      setCurrentRule(rule);
      setFormData({
        countryCode: rule.countryCode,
        countryName: rule.countryName,
        pattern: rule.pattern,
        priority: rule.priority,
        enabled: rule.enabled
      });
    } else {
      setCurrentRule(null);
      setFormData({
        countryCode: '',
        countryName: '',
        pattern: '',
        priority: 0,
        enabled: true
      });
    }
    setErrors({ countryCode: '' });
    setTestName('');
    setTestResult(null);
    setDialogOpen(true);
  };

  // Close dialog
  const handleCloseDialog = () => {
    setDialogOpen(false);
    setCurrentRule(null);
    setErrors({ countryCode: '' });
    setTestName('');
    setTestResult(null);
  };

  // Save rule
  const handleSave = async () => {
    // Validation
    if (!formData.countryCode || !formData.countryName || !formData.pattern) {
      showMessage(t('settings.countryRulesPanel.messages.incompleteData'), 'warning');
      return;
    }

    // Convert country code to uppercase
    const upperCountryCode = formData.countryCode.toUpperCase();

    // Validate country code uniqueness
    if (!validateCountryCode(upperCountryCode, currentRule?.id)) {
      setErrors({ countryCode: t('settings.countryRulesPanel.messages.countryCodeExists') });
      showMessage(t('settings.countryRulesPanel.messages.countryCodeExists'), 'warning');
      return;
    }

    // Always use regex pattern type
    const data = {
      countryCode: upperCountryCode,
      countryName: formData.countryName,
      pattern: formData.pattern,
      patternType: 'regex',
      priority: formData.priority,
      enabled: formData.enabled
    };

    try {
      if (currentRule) {
        // Update
        await updateCountryRule(currentRule.id, data);
        showMessage(t('settings.countryRulesPanel.messages.saveSuccess'), 'success');
      } else {
        // Create
        await createCountryRule(data);
        showMessage(t('settings.countryRulesPanel.messages.saveSuccess'), 'success');
      }
      handleCloseDialog();
      loadRules();
    } catch {
      showMessage(t('settings.countryRulesPanel.messages.saveFailed'), 'error');
    }
  };

  // Delete rule
  const handleDelete = async () => {
    try {
      await deleteCountryRule(deleteId);
      showMessage(t('settings.countryRulesPanel.messages.deleteSuccess'), 'success');
      setDeleteDialogOpen(false);
      setDeleteId(null);
      loadRules();
    } catch {
      showMessage(t('settings.countryRulesPanel.messages.deleteFailed'), 'error');
    }
  };

  // Test rule
  const handleTest = async () => {
    if (!testName) {
      showMessage(t('settings.countryRulesPanel.messages.testNameRequired'), 'warning');
      return;
    }

    setTesting(true);
    try {
      const response = await testCountryRule({
        pattern: formData.pattern,
        patternType: 'regex',
        testName: testName
      });
      setTestResult(response.data || { matched: false });
    } catch {
      showMessage(t('settings.countryRulesPanel.messages.testFailed'), 'error');
    } finally {
      setTesting(false);
    }
  };

  // ========== 文本编辑模式操作 ==========

  const handleTextSync = async () => {
    setSyncing(true);
    try {
      const res = await syncCountryRules(textContent);
      const { added, updated, deleted } = res.data || {};
      showMessage(t('countryRules.messages.syncSuccess', { added: added || 0, updated: updated || 0, deleted: deleted || 0 }));
      setOriginalText(textContent);
      // 刷新列表以同步状态
      loadRules();
    } catch (error) {
      showMessage(error.message || t('countryRules.messages.syncFailed'), 'error');
    } finally {
      setSyncing(false);
    }
  };

  const handleTextReset = () => {
    setTextContent(originalText);
    showMessage(t('countryRules.messages.restored'));
  };

  const handleCopyText = () => {
    navigator.clipboard.writeText(textContent).then(() => {
      showMessage(t('countryRules.messages.copied'));
    });
  };

  const hasTextChanged = textContent !== originalText;

  const sortedRules = [...rules].sort(getComparator(order, orderBy));

  return (
    <Box>
      <Box sx={{ mb: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: 1 }}>
        <Stack direction="row" spacing={1} alignItems="center">
          <Typography variant="h6">{t('settings.countryRulesPanel.title')}</Typography>

          {/* 模式切换 */}
          <ToggleButtonGroup
            value={viewMode}
            exclusive
            onChange={(e, v) => v && setViewMode(v)}
            size="small"
            sx={{
              '& .MuiToggleButton-root': {
                px: 1.5,
                py: 0.5,
                fontSize: '0.8rem',
                '&.Mui-selected': {
                  backgroundColor: theme.palette.primary.main,
                  color: '#fff',
                  '&:hover': {
                    backgroundColor: theme.palette.primary.dark
                  }
                }
              }
            }}
          >
            <ToggleButton value="table">
              <TableViewIcon sx={{ mr: 0.5, fontSize: '1rem' }} />
              {t('countryRules.mode.form')}
            </ToggleButton>
            <ToggleButton value="text">
              <CodeIcon sx={{ mr: 0.5, fontSize: '1rem' }} />
              {t('countryRules.mode.text')}
            </ToggleButton>
          </ToggleButtonGroup>
        </Stack>

        <Box sx={{ display: 'flex', gap: 1 }}>
          {viewMode === 'table' && (
            <>
              <Button variant="outlined" startIcon={<RefreshIcon />} onClick={() => loadRules()}>
                {t('common.refresh')}
              </Button>
              <Button variant="contained" startIcon={<AddIcon />} onClick={() => handleOpenDialog()}>
                {t('settings.countryRulesPanel.actions.addRule')}
              </Button>
            </>
          )}
          {viewMode === 'text' && (
            <Button variant="outlined" startIcon={<RefreshIcon />} onClick={loadTextContent} disabled={loading}>
              {t('common.refresh')}
            </Button>
          )}
        </Box>
      </Box>

      {/* 表单模式 */}
      {viewMode === 'table' && (
        <>
          <Alert severity="info" sx={{ mb: 2 }}>
            <Typography component="div" variant="body2" dangerouslySetInnerHTML={{ __html: t('settings.countryRulesPanel.description') }} />
          </Alert>

          {/* 搜索框 */}
          <Box sx={{ mb: 2 }}>
            <TextField
              size="small"
              label={t('countryRules.search.label')}
              placeholder={t('countryRules.search.placeholder')}
              value={searchKeyword}
              onChange={(e) => setSearchKeyword(e.target.value)}
              sx={{ minWidth: 250 }}
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <SearchIcon />
                  </InputAdornment>
                )
              }}
            />
          </Box>

          <TableContainer component={Paper}>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell sx={{ whiteSpace: 'nowrap' }}>
                    <TableSortLabel
                      active={orderBy === 'countryCode'}
                      direction={orderBy === 'countryCode' ? order : 'asc'}
                      onClick={() => handleRequestSort('countryCode')}
                    >
                      {t('settings.countryRulesPanel.table.countryCode')}
                    </TableSortLabel>
                  </TableCell>
                  <TableCell sx={{ whiteSpace: 'nowrap' }}>
                    <TableSortLabel
                      active={orderBy === 'countryName'}
                      direction={orderBy === 'countryName' ? order : 'asc'}
                      onClick={() => handleRequestSort('countryName')}
                    >
                      {t('settings.countryRulesPanel.table.countryName')}
                    </TableSortLabel>
                  </TableCell>
                  <TableCell sx={{ whiteSpace: 'nowrap' }}>{t('settings.countryRulesPanel.table.pattern')}</TableCell>
                  <TableCell align="center" sx={{ whiteSpace: 'nowrap' }}>
                    <TableSortLabel
                      active={orderBy === 'priority'}
                      direction={orderBy === 'priority' ? order : 'asc'}
                      onClick={() => handleRequestSort('priority')}
                    >
                      {t('settings.countryRulesPanel.table.priority')}
                    </TableSortLabel>
                  </TableCell>
                  <TableCell align="center" sx={{ whiteSpace: 'nowrap' }}>
                    {t('settings.countryRulesPanel.table.status')}
                  </TableCell>
                  <TableCell align="center" sx={{ whiteSpace: 'nowrap' }}>
                    {t('settings.countryRulesPanel.table.actions')}
                  </TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {loading ? (
                  <TableRow>
                    <TableCell colSpan={6} align="center">
                      <CircularProgress size={24} sx={{ my: 2 }} />
                    </TableCell>
                  </TableRow>
                ) : sortedRules.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={6} align="center">
                      <Typography color="textSecondary">{t('settings.countryRulesPanel.empty.noRules')}</Typography>
                    </TableCell>
                  </TableRow>
                ) : (
                  sortedRules.map((rule) => (
                    <TableRow
                      key={rule.id}
                      sx={{
                        '&:hover': {
                          backgroundColor: theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.02)'
                        }
                      }}
                    >
                      <TableCell sx={{ whiteSpace: 'nowrap' }}>
                        <Chip
                          label={`${rule.countryCode} ${getCountryFlag(rule.countryCode)}`}
                          size="small"
                          sx={{
                            backgroundColor: withAlpha(theme.palette.primary.main, 0.1),
                            color: theme.palette.primary.main,
                            fontWeight: 500
                          }}
                        />
                      </TableCell>
                      <TableCell sx={{ maxWidth: 150 }}>
                        <Tooltip title={rule.countryName}>
                          <Typography
                            variant="body2"
                            sx={{
                              overflow: 'hidden',
                              textOverflow: 'ellipsis',
                              whiteSpace: 'nowrap'
                            }}
                          >
                            {rule.countryName}
                          </Typography>
                        </Tooltip>
                      </TableCell>
                      <TableCell sx={{ maxWidth: 300 }}>
                        <Tooltip title={rule.pattern}>
                          <Typography
                            variant="body2"
                            sx={{
                              overflow: 'hidden',
                              textOverflow: 'ellipsis',
                              whiteSpace: 'nowrap'
                            }}
                          >
                            {rule.pattern}
                          </Typography>
                        </Tooltip>
                      </TableCell>
                      <TableCell align="center" sx={{ whiteSpace: 'nowrap' }}>
                        {rule.priority}
                      </TableCell>
                      <TableCell align="center" sx={{ whiteSpace: 'nowrap' }}>
                        <Chip
                          label={
                            rule.enabled ? t('settings.countryRulesPanel.status.enabled') : t('settings.countryRulesPanel.status.disabled')
                          }
                          size="small"
                          sx={getStatusChipSx(rule.enabled)}
                        />
                      </TableCell>
                      <TableCell align="center" sx={{ whiteSpace: 'nowrap' }}>
                        <Tooltip title={t('settings.countryRulesPanel.actions.edit')}>
                          <IconButton size="small" onClick={() => handleOpenDialog(rule)} color="primary">
                            <EditIcon />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title={t('settings.countryRulesPanel.actions.delete')}>
                          <IconButton
                            size="small"
                            onClick={() => {
                              setDeleteId(rule.id);
                              setDeleteDialogOpen(true);
                            }}
                            color="error"
                          >
                            <DeleteIcon />
                          </IconButton>
                        </Tooltip>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </TableContainer>

          {/* 分页组件 */}
          <Box sx={{ mt: 2 }}>
            <Pagination
              page={page}
              pageSize={rowsPerPage}
              totalItems={totalItems}
              onPageChange={(e, newPage) => {
                setPage(newPage);
                loadRules({ page: newPage + 1, pageSize: rowsPerPage, keyword: searchKeyword });
              }}
              onPageSizeChange={(e) => {
                const newValue = parseInt(e.target.value, 10);
                setRowsPerPage(newValue);
                localStorage.setItem('country_rules_rowsPerPage', newValue);
                setPage(0);
                loadRules({ page: 1, pageSize: newValue, keyword: searchKeyword });
              }}
              pageSizeOptions={[10, 20, 50, 100]}
            />
          </Box>
        </>
      )}

      {/* 文本编辑模式 */}
      {viewMode === 'text' && (
        <Box>
          <Alert severity="info" sx={{ mb: 2 }}>
            <Trans i18nKey="countryRules.textMode.description" components={{ code: <code /> }} />
          </Alert>

          <TextField
            multiline
            fullWidth
            rows={18}
            value={textContent}
            onChange={(e) => setTextContent(e.target.value)}
            placeholder={t('countryRules.textMode.placeholder')}
            sx={{
              mb: 2,
              '& .MuiInputBase-root': {
                fontFamily: 'monospace',
                fontSize: '0.9rem',
                lineHeight: 1.6,
                backgroundColor: isDark ? theme.palette.grey[900] : theme.palette.grey[50]
              }
            }}
          />

          <Stack direction="row" spacing={2} alignItems="center">
            <Button
              variant="contained"
              startIcon={syncing ? <CircularProgress size={16} color="inherit" /> : <SaveIcon />}
              onClick={handleTextSync}
              disabled={!hasTextChanged || syncing}
            >
              {t('countryRules.actions.saveSync')}
            </Button>
            <Button variant="outlined" startIcon={<UndoIcon />} onClick={handleTextReset} disabled={!hasTextChanged}>
              {t('countryRules.actions.restore')}
            </Button>
            <Button variant="text" startIcon={<ContentCopyIcon />} onClick={handleCopyText}>
              {t('countryRules.actions.copy')}
            </Button>
          </Stack>

          {hasTextChanged && (
            <Typography variant="body2" color="warning.main" sx={{ mt: 1 }}>
              {t('countryRules.textMode.changed')}
            </Typography>
          )}
        </Box>
      )}

      {/* Add/Edit Dialog */}
      <Dialog open={dialogOpen} onClose={handleCloseDialog} maxWidth="sm" fullWidth>
        <DialogTitle>
          {currentRule ? t('settings.countryRulesPanel.dialog.titleEdit') : t('settings.countryRulesPanel.dialog.titleAdd')}
        </DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <TextField
              label={t('settings.countryRulesPanel.dialog.countryCode')}
              value={formData.countryCode}
              onChange={(e) => {
                const newCode = e.target.value.toUpperCase();
                setFormData({ ...formData, countryCode: newCode });
                // Clear error when user types
                if (errors.countryCode) {
                  setErrors({ countryCode: '' });
                }
              }}
              placeholder={t('settings.countryRulesPanel.dialog.countryCodePlaceholder')}
              fullWidth
              required
              inputProps={{ maxLength: 10 }}
              error={!!errors.countryCode}
              helperText={errors.countryCode}
            />
            <TextField
              label={t('settings.countryRulesPanel.dialog.countryName')}
              value={formData.countryName}
              onChange={(e) => setFormData({ ...formData, countryName: e.target.value })}
              placeholder={t('settings.countryRulesPanel.dialog.countryNamePlaceholder')}
              fullWidth
              required
            />
            <TextField
              label={t('settings.countryRulesPanel.dialog.pattern')}
              value={formData.pattern}
              onChange={(e) => setFormData({ ...formData, pattern: e.target.value })}
              placeholder={t('settings.countryRulesPanel.dialog.patternPlaceholder')}
              multiline
              rows={3}
              fullWidth
              required
              helperText={t('settings.countryRulesPanel.dialog.regexHelper')}
            />
            <TextField
              label={t('settings.countryRulesPanel.dialog.priority')}
              type="number"
              value={formData.priority}
              onChange={(e) => setFormData({ ...formData, priority: parseInt(e.target.value) || 0 })}
              fullWidth
              helperText={t('settings.countryRulesPanel.dialog.priorityHelper')}
              inputProps={{ min: 0, max: 1000 }}
            />
            <FormControlLabel
              control={<Switch checked={formData.enabled} onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })} />}
              label={t('settings.countryRulesPanel.dialog.enabled')}
            />

            {/* Test Section */}
            <Box sx={{ borderTop: 1, borderColor: 'divider', pt: 2 }}>
              <Typography variant="subtitle2" gutterBottom>
                {t('settings.countryRulesPanel.dialog.testSection')}
              </Typography>
              <Stack direction="row" spacing={1}>
                <TextField
                  label={t('settings.countryRulesPanel.dialog.testNodeName')}
                  value={testName}
                  onChange={(e) => setTestName(e.target.value)}
                  placeholder={t('settings.countryRulesPanel.dialog.testNodePlaceholder')}
                  size="small"
                  fullWidth
                />
                <Button
                  variant="outlined"
                  startIcon={testing ? <CircularProgress size={16} /> : <PlayArrowIcon />}
                  onClick={handleTest}
                  disabled={testing}
                  sx={{ minWidth: 100 }}
                >
                  {t('settings.countryRulesPanel.dialog.testButton')}
                </Button>
              </Stack>
              {testResult && (
                <Alert severity={testResult.matched ? 'success' : 'info'} sx={{ mt: 1 }}>
                  {testResult.matched
                    ? t('settings.countryRulesPanel.dialog.testMatched')
                    : t('settings.countryRulesPanel.dialog.testNotMatched')}
                </Alert>
              )}
            </Box>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseDialog}>{t('settings.countryRulesPanel.actions.cancel')}</Button>
          <Button onClick={handleSave} variant="contained">
            {t('settings.countryRulesPanel.actions.save')}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteDialogOpen} onClose={() => setDeleteDialogOpen(false)}>
        <DialogTitle>{t('settings.countryRulesPanel.deleteDialog.title')}</DialogTitle>
        <DialogContent>
          <Typography>{t('settings.countryRulesPanel.deleteDialog.content')}</Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteDialogOpen(false)}>{t('settings.countryRulesPanel.actions.cancel')}</Button>
          <Button onClick={handleDelete} color="error" variant="contained">
            {t('settings.countryRulesPanel.actions.delete')}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

CountryRules.propTypes = {
  showMessage: PropTypes.func.isRequired
};

export default CountryRules;
