import { useState, useEffect } from 'react';

// material-ui
import Box from '@mui/material/Box';
import Grid from '@mui/material/Grid';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import Chip from '@mui/material/Chip';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import TextField from '@mui/material/TextField';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';
import Switch from '@mui/material/Switch';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import Alert from '@mui/material/Alert';
import Snackbar from '@mui/material/Snackbar';
import Tabs from '@mui/material/Tabs';
import Tab from '@mui/material/Tab';
import Divider from '@mui/material/Divider';

// icons
import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import LocalOfferIcon from '@mui/icons-material/LocalOffer';
import RuleIcon from '@mui/icons-material/Rule';

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

// components
import TagDialog from './component/TagDialog';
import RuleDialog from './component/RuleDialog';

// ==============================|| TAG MANAGEMENT ||============================== //

export default function TagManagement() {
  const [tabValue, setTabValue] = useState(0);
  const [tags, setTags] = useState([]);
  const [rules, setRules] = useState([]);
  const [existingGroups, setExistingGroups] = useState([]);
  const [loading, setLoading] = useState(false);
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });

  // Dialog states
  const [tagDialogOpen, setTagDialogOpen] = useState(false);
  const [ruleDialogOpen, setRuleDialogOpen] = useState(false);
  const [editingTag, setEditingTag] = useState(null);
  const [editingRule, setEditingRule] = useState(null);

  // Fetch data
  const fetchTags = async () => {
    try {
      const res = await getTags();
      if (res.code === 200) {
        setTags(res.data || []);
      }
    } catch (error) {
      showMessage('获取标签列表失败', 'error');
    }
  };

  const fetchRules = async () => {
    try {
      const res = await getTagRules();
      if (res.code === 200) {
        setRules(res.data || []);
      }
    } catch (error) {
      showMessage('获取规则列表失败', 'error');
    }
  };

  const fetchGroups = async () => {
    try {
      const res = await getTagGroups();
      if (res.code === 200) {
        setExistingGroups(res.data || []);
      }
    } catch (error) {
      // Silent fail for groups
    }
  };

  useEffect(() => {
    fetchTags();
    fetchRules();
    fetchGroups();
  }, []);

  const showMessage = (message, severity = 'success') => {
    setSnackbar({ open: true, message, severity });
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
    if (!window.confirm(`确定删除标签 "${tag.name}" 吗？相关规则也会被删除。`)) return;
    try {
      const res = await deleteTag(tag.name);
      if (res.code === 200) {
        showMessage('删除成功');
        fetchTags();
        fetchRules();
      } else {
        showMessage(res.msg || '删除失败', 'error');
      }
    } catch (error) {
      showMessage('删除失败', 'error');
    }
  };

  const handleSaveTag = async (tagData) => {
    try {
      let res;
      if (editingTag) {
        res = await updateTag({ ...tagData, name: editingTag.name });
      } else {
        res = await addTag(tagData);
      }
      if (res.code === 200) {
        showMessage(editingTag ? '更新成功' : '添加成功');
        setTagDialogOpen(false);
        fetchTags();
      } else {
        showMessage(res.msg || '操作失败', 'error');
      }
    } catch (error) {
      showMessage('操作失败', 'error');
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
    if (!window.confirm(`确定删除规则 "${rule.name}" 吗？`)) return;
    try {
      const res = await deleteTagRule(rule.id);
      if (res.code === 200) {
        showMessage('删除成功');
        fetchRules();
      } else {
        showMessage(res.msg || '删除失败', 'error');
      }
    } catch (error) {
      showMessage('删除失败', 'error');
    }
  };

  const handleSaveRule = async (ruleData) => {
    try {
      let res;
      if (editingRule) {
        res = await updateTagRule({ ...ruleData, id: editingRule.id });
      } else {
        res = await addTagRule(ruleData);
      }
      if (res.code === 200) {
        showMessage(editingRule ? '更新成功' : '添加成功');
        setRuleDialogOpen(false);
        fetchRules();
      } else {
        showMessage(res.msg || '操作失败', 'error');
      }
    } catch (error) {
      showMessage('操作失败', 'error');
    }
  };

  const handleTriggerRule = async (rule) => {
    try {
      const res = await triggerTagRule(rule.id);
      if (res.code === 200) {
        showMessage('规则已开始执行');
      } else {
        showMessage(res.msg || '执行失败', 'error');
      }
    } catch (error) {
      showMessage('执行失败', 'error');
    }
  };

  const getTagByName = (tagName) => {
    return tags.find((t) => t.name === tagName);
  };

  return (
    <MainCard title="标签管理">
      <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 2 }}>
        <Tabs value={tabValue} onChange={(e, v) => setTabValue(v)}>
          <Tab icon={<LocalOfferIcon sx={{ mr: 1 }} />} iconPosition="start" label="标签列表" />
          <Tab icon={<RuleIcon sx={{ mr: 1 }} />} iconPosition="start" label="自动规则" />
        </Tabs>
      </Box>

      {/* 标签列表 */}
      {tabValue === 0 && (
        <Box>
          <Box sx={{ mb: 2, display: 'flex', justifyContent: 'flex-end' }}>
            <Button variant="contained" startIcon={<AddIcon />} onClick={handleAddTag}>
              添加标签
            </Button>
          </Box>
          <Grid container spacing={2}>
            {tags.map((tag) => (
              <Grid item xs={12} sm={6} md={4} lg={3} key={tag.name}>
                <Card
                  sx={{
                    borderLeft: `4px solid ${tag.color}`,
                    '&:hover': { boxShadow: 3 }
                  }}
                >
                  <CardContent>
                    <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                        <Box
                          sx={{
                            width: 16,
                            height: 16,
                            borderRadius: '50%',
                            backgroundColor: tag.color
                          }}
                        />
                        <Typography variant="h5">{tag.name}</Typography>
                      </Box>
                      <Box>
                        <IconButton size="small" onClick={() => handleEditTag(tag)}>
                          <EditIcon fontSize="small" />
                        </IconButton>
                        <IconButton size="small" color="error" onClick={() => handleDeleteTag(tag)}>
                          <DeleteIcon fontSize="small" />
                        </IconButton>
                      </Box>
                    </Box>
                    {tag.groupName && (
                      <Chip label={`组: ${tag.groupName}`} size="small" variant="outlined" sx={{ mt: 1, fontSize: '0.7rem' }} />
                    )}
                    {tag.description && (
                      <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                        {tag.description}
                      </Typography>
                    )}
                  </CardContent>
                </Card>
              </Grid>
            ))}
            {tags.length === 0 && (
              <Grid item xs={12}>
                <Typography color="text.secondary" align="center" sx={{ py: 4 }}>
                  暂无标签，点击"添加标签"创建第一个标签
                </Typography>
              </Grid>
            )}
          </Grid>
        </Box>
      )}

      {/* 自动规则 */}
      {tabValue === 1 && (
        <Box>
          <Box sx={{ mb: 2, display: 'flex', justifyContent: 'flex-end' }}>
            <Button variant="contained" startIcon={<AddIcon />} onClick={handleAddRule} disabled={tags.length === 0}>
              添加规则
            </Button>
          </Box>
          {tags.length === 0 && (
            <Alert severity="info" sx={{ mb: 2 }}>
              请先创建标签后再添加自动规则
            </Alert>
          )}
          <TableContainer component={Paper}>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>规则名称</TableCell>
                  <TableCell>关联标签</TableCell>
                  <TableCell>触发时机</TableCell>
                  <TableCell>状态</TableCell>
                  <TableCell align="right">操作</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {rules.map((rule) => {
                  const tag = getTagByName(rule.tagName);
                  return (
                    <TableRow key={rule.id}>
                      <TableCell>{rule.name}</TableCell>
                      <TableCell>
                        {tag ? (
                          <Chip label={tag.name} size="small" sx={{ backgroundColor: tag.color, color: '#fff' }} />
                        ) : (
                          <Typography color="text.secondary">未知标签</Typography>
                        )}
                      </TableCell>
                      <TableCell>
                        {rule.triggerType === 'subscription_update' && '订阅更新后'}
                        {rule.triggerType === 'speed_test' && '测速完成后'}
                      </TableCell>
                      <TableCell>
                        <Chip label={rule.enabled ? '启用' : '禁用'} size="small" color={rule.enabled ? 'success' : 'default'} />
                      </TableCell>
                      <TableCell align="right">
                        <IconButton size="small" onClick={() => handleTriggerRule(rule)} title="手动执行">
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
                {rules.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={5} align="center" sx={{ py: 4 }}>
                      <Typography color="text.secondary">暂无自动规则</Typography>
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </TableContainer>
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
