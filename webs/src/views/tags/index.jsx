import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';

// material-ui
import Box from '@mui/material/Box';
import Grid from '@mui/material/Grid';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import Chip from '@mui/material/Chip';
import Alert from '@mui/material/Alert';
import Snackbar from '@mui/material/Snackbar';
import Tabs from '@mui/material/Tabs';
import Tab from '@mui/material/Tab';
import Stack from '@mui/material/Stack';
import Divider from '@mui/material/Divider';
import Paper from '@mui/material/Paper';
import Tooltip from '@mui/material/Tooltip';
import CircularProgress from '@mui/material/CircularProgress';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TextField from '@mui/material/TextField';
import InputAdornment from '@mui/material/InputAdornment';
import FormControl from '@mui/material/FormControl';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import TableRow from '@mui/material/TableRow';

// icons
import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import LocalOfferIcon from '@mui/icons-material/LocalOffer';
import RuleIcon from '@mui/icons-material/Rule';
import RefreshIcon from '@mui/icons-material/Refresh';
import SearchIcon from '@mui/icons-material/Search';
import ClearIcon from '@mui/icons-material/Clear';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import {
  getTags,
  addTag,
  updateTag,
  deleteTag,
  getTagRules,
  addTagRule,
  updateTagRule,
  deleteTagRule,
  triggerTagRule,
  getTagGroups
} from 'api/tags';
import { getNodeCheckMeta } from 'api/nodeCheck';
import { setUnlockMeta } from 'views/nodes/utils';

// components
import TagDialog from './component/TagDialog';
import RuleDialog from './component/RuleDialog';

// ==============================|| TAG MANAGEMENT ||============================== //

export default function TagManagement() {
  const theme = useTheme();
  const { t } = useTranslation();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));

  const [tabValue, setTabValue] = useState(0);
  const [tags, setTags] = useState([]);
  const [rules, setRules] = useState([]);
  const [existingGroups, setExistingGroups] = useState([]);
  const [refreshing, setRefreshing] = useState(false);
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });

  // Dialog states
  const [tagDialogOpen, setTagDialogOpen] = useState(false);
  const [ruleDialogOpen, setRuleDialogOpen] = useState(false);
  const [editingTag, setEditingTag] = useState(null);
  const [editingRule, setEditingRule] = useState(null);

  // Search/Filter states
  const [tagSearch, setTagSearch] = useState('');
  const [ruleSearch, setRuleSearch] = useState('');
  const [ruleTagFilter, setRuleTagFilter] = useState('');
  const [ruleTriggerFilter, setRuleTriggerFilter] = useState('');
  const [ruleStatusFilter, setRuleStatusFilter] = useState('');

  // Fetch data
  const fetchTags = async (showRefreshing = false) => {
    if (showRefreshing) setRefreshing(true);
    try {
      const res = await getTags();
      // 成功（code === 200 时返回，否则被拦截器 reject）
      setTags(res.data || []);
    } catch (error) {
      showMessage(error.message || t('tags.messages.loadTagsFailed'), 'error');
    } finally {
      if (showRefreshing) setRefreshing(false);
    }
  };

  const fetchRules = async (showRefreshing = false) => {
    if (showRefreshing) setRefreshing(true);
    try {
      const res = await getTagRules();
      // 成功（code === 200 时返回，否则被拦截器 reject）
      setRules(res.data || []);
    } catch (error) {
      showMessage(error.message || t('tags.messages.loadRulesFailed'), 'error');
    } finally {
      if (showRefreshing) setRefreshing(false);
    }
  };

  const fetchGroups = async () => {
    try {
      const res = await getTagGroups();
      // 成功（code === 200 时返回，否则被拦截器 reject）
      setExistingGroups(res.data || []);
    } catch {
      // Silent fail for groups
    }
  };

  const fetchUnlockMeta = async () => {
    try {
      const res = await getNodeCheckMeta();
      setUnlockMeta(res.data || {});
    } catch {
      console.error('加载解锁元数据失败');
    }
  };

  useEffect(() => {
    fetchTags();
    fetchRules();
    fetchGroups();
    fetchUnlockMeta();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const showMessage = (message, severity = 'success') => {
    setSnackbar({ open: true, message, severity });
  };

  // Refresh handler
  const handleRefresh = async () => {
    setRefreshing(true);
    try {
      if (tabValue === 0) {
        await fetchTags();
        await fetchGroups();
      } else {
        await fetchRules();
      }
      showMessage(t('tags.messages.refreshSuccess'));
    } catch (error) {
      showMessage(error.message || t('tags.messages.refreshFailed'), 'error');
    } finally {
      setRefreshing(false);
    }
  };

  // Tag operations
  const handleAddTag = () => {
    setEditingTag(null);
    setTagDialogOpen(true);
  };

  const handleEditTag = (tag) => {
    setEditingTag(tag);
    setTagDialogOpen(true);
  };

  const handleDeleteTag = async (tag) => {
    if (!window.confirm(t('tags.confirm.deleteTag', { name: tag.name }))) return;
    try {
      await deleteTag(tag.name);
      // 成功（code === 200 时返回，否则被拦截器 reject）
      showMessage(t('tags.messages.deleteSuccess'));
      fetchTags();
      fetchRules();
    } catch (error) {
      showMessage(error.message || t('tags.messages.deleteFailed'), 'error');
    }
  };

  const handleSaveTag = async (tagData) => {
    try {
      if (editingTag) {
        await updateTag({ ...tagData, name: editingTag.name });
      } else {
        await addTag(tagData);
      }
      // 成功（code === 200 时返回，否则被拦截器 reject）
      showMessage(editingTag ? t('tags.messages.updateSuccess') : t('tags.messages.addSuccess'));
      setTagDialogOpen(false);
      fetchTags();
    } catch (error) {
      showMessage(error.message || t('tags.messages.operationFailed'), 'error');
    }
  };

  // Rule operations
  const handleAddRule = () => {
    setEditingRule(null);
    setRuleDialogOpen(true);
  };

  const handleEditRule = (rule) => {
    setEditingRule(rule);
    setRuleDialogOpen(true);
  };

  const handleDeleteRule = async (rule) => {
    if (!window.confirm(t('tags.confirm.deleteRule', { name: rule.name }))) return;
    try {
      await deleteTagRule(rule.id);
      // 成功（code === 200 时返回，否则被拦截器 reject）
      showMessage(t('tags.messages.deleteSuccess'));
      fetchRules();
    } catch (error) {
      showMessage(error.message || t('tags.messages.deleteFailed'), 'error');
    }
  };

  const handleSaveRule = async (ruleData) => {
    try {
      if (editingRule) {
        await updateTagRule({ ...ruleData, id: editingRule.id });
      } else {
        await addTagRule(ruleData);
      }
      // 成功（code === 200 时返回，否则被拦截器 reject）
      showMessage(editingRule ? t('tags.messages.updateSuccess') : t('tags.messages.addSuccess'));
      setRuleDialogOpen(false);
      fetchRules();
    } catch (error) {
      showMessage(error.message || t('tags.messages.operationFailed'), 'error');
    }
  };

  const handleTriggerRule = async (rule) => {
    try {
      await triggerTagRule(rule.id);
      // 成功（code === 200 时返回，否则被拦截器 reject）
      showMessage(t('tags.messages.ruleStarted'));
    } catch (error) {
      showMessage(error.message || t('tags.messages.executeFailed'), 'error');
    }
  };

  const getTagByName = (tagName) => {
    return tags.find((t) => t.name === tagName);
  };

  const getTriggerLabel = (triggerType) =>
    triggerType === 'subscription_update' ? t('tags.trigger.subscriptionUpdate') : t('tags.trigger.speedTest');

  const getStatusLabel = (enabled) => (enabled ? t('tags.status.enabled') : t('tags.status.disabled'));

  // 过滤标签
  const filteredTags = tags.filter((tag) => {
    if (!tagSearch) return true;
    const search = tagSearch.toLowerCase();
    return (
      tag.name.toLowerCase().includes(search) ||
      (tag.groupName && tag.groupName.toLowerCase().includes(search)) ||
      (tag.description && tag.description.toLowerCase().includes(search))
    );
  });

  // 过滤规则
  const filteredRules = rules.filter((rule) => {
    // 名称搜索
    if (ruleSearch && !rule.name.toLowerCase().includes(ruleSearch.toLowerCase())) {
      return false;
    }
    // 标签筛选
    if (ruleTagFilter && rule.tagName !== ruleTagFilter) {
      return false;
    }
    // 触发时机筛选
    if (ruleTriggerFilter && rule.triggerType !== ruleTriggerFilter) {
      return false;
    }
    // 状态筛选
    if (ruleStatusFilter !== '') {
      const isEnabled = ruleStatusFilter === 'enabled';
      if (rule.enabled !== isEnabled) {
        return false;
      }
    }
    return true;
  });

  // 获取规则中使用的标签列表（去重）
  const usedTagNames = [...new Set(rules.map((r) => r.tagName))];

  // 移动端规则卡片
  const MobileRuleCard = ({ rule }) => {
    const tag = getTagByName(rule.tagName);
    return (
      <Card
        variant="outlined"
        sx={{
          mb: 1.5,
          borderRadius: 2,
          transition: 'all 0.2s ease',
          '&:hover': { boxShadow: 2 }
        }}
      >
        <CardContent sx={{ py: 1.5, px: 2, '&:last-child': { pb: 1.5 } }}>
          <Stack spacing={1.5}>
            {/* 规则名称和操作按钮 */}
            <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <Typography variant="subtitle1" fontWeight={600}>
                {rule.name}
              </Typography>
              <Stack direction="row" spacing={0.5}>
                <Tooltip title={t('tags.actions.execute')}>
                  <IconButton size="small" onClick={() => handleTriggerRule(rule)} color="primary">
                    <PlayArrowIcon fontSize="small" />
                  </IconButton>
                </Tooltip>
                <Tooltip title={t('tags.actions.edit')}>
                  <IconButton size="small" onClick={() => handleEditRule(rule)}>
                    <EditIcon fontSize="small" />
                  </IconButton>
                </Tooltip>
                <Tooltip title={t('tags.actions.delete')}>
                  <IconButton size="small" color="error" onClick={() => handleDeleteRule(rule)}>
                    <DeleteIcon fontSize="small" />
                  </IconButton>
                </Tooltip>
              </Stack>
            </Box>

            <Divider />

            {/* 标签、触发时机、状态 */}
            <Stack direction="row" spacing={1} flexWrap="wrap" alignItems="center">
              {tag ? (
                <Chip label={tag.name} size="small" sx={{ backgroundColor: tag.color, color: '#fff' }} />
              ) : (
                <Typography variant="body2" color="text.secondary">
                  {t('tags.unknownTag')}
                </Typography>
              )}
              <Chip label={getTriggerLabel(rule.triggerType)} size="small" variant="outlined" />
              <Chip label={getStatusLabel(rule.enabled)} size="small" color={rule.enabled ? 'success' : 'default'} />
            </Stack>
          </Stack>
        </CardContent>
      </Card>
    );
  };

  return (
    <MainCard title={t('tags.title')}>
      <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 2 }}>
        <Stack direction="row" justifyContent="space-between" alignItems="center">
          <Tabs value={tabValue} onChange={(e, v) => setTabValue(v)}>
            <Tab icon={<LocalOfferIcon sx={{ mr: isMobile ? 0 : 1 }} />} iconPosition="start" label={isMobile ? '' : t('tags.tabs.tags')} />
            <Tab icon={<RuleIcon sx={{ mr: isMobile ? 0 : 1 }} />} iconPosition="start" label={isMobile ? '' : t('tags.tabs.rules')} />
          </Tabs>
          <Tooltip title={t('tags.actions.refresh')}>
            <IconButton onClick={handleRefresh} disabled={refreshing} color="primary">
              {refreshing ? <CircularProgress size={20} /> : <RefreshIcon />}
            </IconButton>
          </Tooltip>
        </Stack>
      </Box>

      {/* 标签列表 */}
      {tabValue === 0 && (
        <Box>
          {/* 搜索和添加按钮 */}
          <Box sx={{ mb: 2, display: 'flex', gap: 2, flexWrap: 'wrap', alignItems: 'center', justifyContent: 'space-between' }}>
            <TextField
              placeholder={t('tags.search.tagPlaceholder')}
              value={tagSearch}
              onChange={(e) => setTagSearch(e.target.value)}
              size="small"
              sx={{ minWidth: { xs: '100%', sm: 280 }, flex: { xs: '1 1 100%', sm: '0 1 auto' } }}
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <SearchIcon color="action" fontSize="small" />
                  </InputAdornment>
                ),
                endAdornment: tagSearch && (
                  <InputAdornment position="end">
                    <IconButton size="small" onClick={() => setTagSearch('')}>
                      <ClearIcon fontSize="small" />
                    </IconButton>
                  </InputAdornment>
                )
              }}
            />
            <Button variant="contained" startIcon={<AddIcon />} onClick={handleAddTag} size={isMobile ? 'small' : 'medium'}>
              {t('tags.actions.addTag')}
            </Button>
          </Box>

          {/* 搜索结果统计 */}
          {tagSearch && (
            <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
              {t('tags.search.tagMatched', { count: filteredTags.length })}
            </Typography>
          )}

          <Grid container spacing={isMobile ? 1.5 : 2}>
            {filteredTags.map((tag) => (
              <Grid item xs={6} sm={6} md={4} lg={3} key={tag.name}>
                <Card
                  sx={{
                    borderLeft: `4px solid ${tag.color}`,
                    transition: 'all 0.2s ease',
                    '&:hover': { boxShadow: 3 }
                  }}
                >
                  <CardContent sx={{ py: isMobile ? 1.5 : 2, px: isMobile ? 1.5 : 2, '&:last-child': { pb: isMobile ? 1.5 : 2 } }}>
                    <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between' }}>
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, minWidth: 0, flex: 1 }}>
                        <Box
                          sx={{
                            width: isMobile ? 12 : 16,
                            height: isMobile ? 12 : 16,
                            borderRadius: '50%',
                            backgroundColor: tag.color,
                            flexShrink: 0
                          }}
                        />
                        <Typography
                          variant={isMobile ? 'body1' : 'h5'}
                          sx={{
                            fontWeight: 600,
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                            whiteSpace: 'nowrap'
                          }}
                        >
                          {tag.name}
                        </Typography>
                      </Box>
                      <Box sx={{ flexShrink: 0, ml: 0.5 }}>
                        <IconButton size="small" onClick={() => handleEditTag(tag)} sx={{ p: isMobile ? 0.5 : 1 }}>
                          <EditIcon fontSize="small" />
                        </IconButton>
                        <IconButton size="small" color="error" onClick={() => handleDeleteTag(tag)} sx={{ p: isMobile ? 0.5 : 1 }}>
                          <DeleteIcon fontSize="small" />
                        </IconButton>
                      </Box>
                    </Box>
                    {tag.groupName && (
                      <Chip
                        label={t('tags.groupLabel', { group: tag.groupName })}
                        size="small"
                        variant="outlined"
                        sx={{ mt: 1, fontSize: isMobile ? '0.65rem' : '0.7rem', height: isMobile ? 20 : 24 }}
                      />
                    )}
                    {tag.description && !isMobile && (
                      <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                        {tag.description}
                      </Typography>
                    )}
                  </CardContent>
                </Card>
              </Grid>
            ))}
            {filteredTags.length === 0 && tags.length > 0 && (
              <Grid item xs={12}>
                <Typography color="text.secondary" align="center" sx={{ py: 4 }}>
                  {t('tags.empty.noTagMatch')}
                </Typography>
              </Grid>
            )}
            {tags.length === 0 && (
              <Grid item xs={12}>
                <Typography color="text.secondary" align="center" sx={{ py: 4 }}>
                  {t('tags.empty.noTags')}
                </Typography>
              </Grid>
            )}
          </Grid>
        </Box>
      )}

      {/* 自动规则 */}
      {tabValue === 1 && (
        <Box>
          {/* 搜索、筛选和添加按钮 */}
          <Box sx={{ mb: 2 }}>
            <Stack
              direction={{ xs: 'column', md: 'row' }}
              spacing={1.5}
              alignItems={{ xs: 'stretch', md: 'center' }}
              justifyContent="space-between"
            >
              {/* 搜索和筛选 */}
              <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1} sx={{ flex: 1 }} flexWrap="wrap">
                {/* 名称搜索 */}
                <TextField
                  placeholder={t('tags.search.rulePlaceholder')}
                  value={ruleSearch}
                  onChange={(e) => setRuleSearch(e.target.value)}
                  size="small"
                  sx={{ minWidth: { xs: '100%', sm: 160 } }}
                  InputProps={{
                    startAdornment: (
                      <InputAdornment position="start">
                        <SearchIcon color="action" fontSize="small" />
                      </InputAdornment>
                    ),
                    endAdornment: ruleSearch && (
                      <InputAdornment position="end">
                        <IconButton size="small" onClick={() => setRuleSearch('')}>
                          <ClearIcon fontSize="small" />
                        </IconButton>
                      </InputAdornment>
                    )
                  }}
                />

                {/* 标签筛选 */}
                <FormControl size="small" sx={{ minWidth: 120 }}>
                  <Select value={ruleTagFilter} onChange={(e) => setRuleTagFilter(e.target.value)} displayEmpty>
                    <MenuItem value="">{t('tags.filters.allTags')}</MenuItem>
                    {usedTagNames.map((tagName) => {
                      const tag = getTagByName(tagName);
                      return (
                        <MenuItem key={tagName} value={tagName}>
                          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                            <Box
                              sx={{
                                width: 12,
                                height: 12,
                                borderRadius: '50%',
                                backgroundColor: tag?.color || '#ccc'
                              }}
                            />
                            {tagName}
                          </Box>
                        </MenuItem>
                      );
                    })}
                  </Select>
                </FormControl>

                {/* 触发时机筛选 */}
                <FormControl size="small" sx={{ minWidth: 120 }}>
                  <Select value={ruleTriggerFilter} onChange={(e) => setRuleTriggerFilter(e.target.value)} displayEmpty>
                    <MenuItem value="">{t('tags.filters.allTriggers')}</MenuItem>
                    <MenuItem value="subscription_update">{t('tags.trigger.subscriptionUpdate')}</MenuItem>
                    <MenuItem value="speed_test">{t('tags.trigger.speedTest')}</MenuItem>
                  </Select>
                </FormControl>

                {/* 状态筛选 */}
                <FormControl size="small" sx={{ minWidth: 100 }}>
                  <Select value={ruleStatusFilter} onChange={(e) => setRuleStatusFilter(e.target.value)} displayEmpty>
                    <MenuItem value="">{t('tags.filters.allStatuses')}</MenuItem>
                    <MenuItem value="enabled">{t('tags.status.enabled')}</MenuItem>
                    <MenuItem value="disabled">{t('tags.status.disabled')}</MenuItem>
                  </Select>
                </FormControl>
              </Stack>

              {/* 添加按钮 */}
              <Button
                variant="contained"
                startIcon={<AddIcon />}
                onClick={handleAddRule}
                disabled={tags.length === 0}
                size={isMobile ? 'small' : 'medium'}
                sx={{ flexShrink: 0 }}
              >
                {t('tags.actions.addRule')}
              </Button>
            </Stack>

            {/* 筛选结果统计 */}
            {(ruleSearch || ruleTagFilter || ruleTriggerFilter || ruleStatusFilter) && (
              <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                {t('tags.search.ruleMatched', { count: filteredRules.length })}
              </Typography>
            )}
          </Box>

          {tags.length === 0 && (
            <Alert severity="info" sx={{ mb: 2 }}>
              {t('tags.empty.createTagFirst')}
            </Alert>
          )}

          {/* 移动端使用卡片，桌面端使用表格 */}
          {isMobile ? (
            <Box>
              {filteredRules.map((rule) => (
                <MobileRuleCard key={rule.id} rule={rule} />
              ))}
              {filteredRules.length === 0 && rules.length > 0 && (
                <Typography color="text.secondary" align="center" sx={{ py: 4 }}>
                  {t('tags.empty.noRuleMatch')}
                </Typography>
              )}
              {rules.length === 0 && (
                <Typography color="text.secondary" align="center" sx={{ py: 4 }}>
                  {t('tags.empty.noRules')}
                </Typography>
              )}
            </Box>
          ) : (
            <TableContainer component={Paper}>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>{t('tags.fields.ruleName')}</TableCell>
                    <TableCell>{t('tags.fields.tag')}</TableCell>
                    <TableCell>{t('tags.fields.trigger')}</TableCell>
                    <TableCell>{t('tags.fields.status')}</TableCell>
                    <TableCell align="right">{t('tags.fields.actions')}</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {filteredRules.map((rule) => {
                    const tag = getTagByName(rule.tagName);
                    return (
                      <TableRow key={rule.id}>
                        <TableCell>{rule.name}</TableCell>
                        <TableCell>
                          {tag ? (
                            <Chip label={tag.name} size="small" sx={{ backgroundColor: tag.color, color: '#fff' }} />
                          ) : (
                            <Typography color="text.secondary">{t('tags.unknownTag')}</Typography>
                          )}
                        </TableCell>
                        <TableCell>{getTriggerLabel(rule.triggerType)}</TableCell>
                        <TableCell>
                          <Chip label={getStatusLabel(rule.enabled)} size="small" color={rule.enabled ? 'success' : 'default'} />
                        </TableCell>
                        <TableCell align="right">
                          <IconButton size="small" onClick={() => handleTriggerRule(rule)} title={t('tags.actions.execute')}>
                            <PlayArrowIcon fontSize="small" />
                          </IconButton>
                          <IconButton size="small" onClick={() => handleEditRule(rule)}>
                            <EditIcon fontSize="small" />
                          </IconButton>
                          <IconButton size="small" color="error" onClick={() => handleDeleteRule(rule)}>
                            <DeleteIcon fontSize="small" />
                          </IconButton>
                        </TableCell>
                      </TableRow>
                    );
                  })}
                  {filteredRules.length === 0 && rules.length > 0 && (
                    <TableRow>
                      <TableCell colSpan={5} align="center" sx={{ py: 4 }}>
                        <Typography color="text.secondary">{t('tags.empty.noRuleMatch')}</Typography>
                      </TableCell>
                    </TableRow>
                  )}
                  {rules.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={5} align="center" sx={{ py: 4 }}>
                        <Typography color="text.secondary">{t('tags.empty.noRules')}</Typography>
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </TableContainer>
          )}
        </Box>
      )}

      {/* Dialogs */}
      <TagDialog
        open={tagDialogOpen}
        onClose={() => setTagDialogOpen(false)}
        onSave={handleSaveTag}
        editingTag={editingTag}
        existingGroups={existingGroups}
      />
      <RuleDialog
        open={ruleDialogOpen}
        onClose={() => setRuleDialogOpen(false)}
        onSave={handleSaveRule}
        editingRule={editingRule}
        tags={tags}
      />

      {/* Snackbar */}
      <Snackbar
        open={snackbar.open}
        autoHideDuration={3000}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
      >
        <Alert severity={snackbar.severity}>{snackbar.message}</Alert>
      </Snackbar>
    </MainCard>
  );
}
