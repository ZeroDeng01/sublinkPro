import { useState, useEffect, useCallback } from 'react';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import Typography from '@mui/material/Typography';
import Paper from '@mui/material/Paper';
import Chip from '@mui/material/Chip';
import Switch from '@mui/material/Switch';
import CircularProgress from '@mui/material/CircularProgress';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import useMediaQuery from '@mui/material/useMediaQuery';
import { useTheme } from '@mui/material/styles';

import CloseIcon from '@mui/icons-material/Close';
import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import DragIndicatorIcon from '@mui/icons-material/DragIndicator';
import AccountTreeIcon from '@mui/icons-material/AccountTree';
import ArrowForwardIcon from '@mui/icons-material/ArrowForward';
import TouchAppIcon from '@mui/icons-material/TouchApp';
import VisibilityIcon from '@mui/icons-material/Visibility';

import { DragDropContext, Droppable, Draggable } from '@hello-pangea/dnd';
import {
  getChainRules,
  createChainRule,
  updateChainRule,
  deleteChainRule,
  toggleChainRule,
  sortChainRules,
  getChainOptions,
  previewChainLinks
} from '../../../api/subscriptions';
import ChainPreviewDialog from './ChainPreviewDialog';
import ChainRuleEditor from './ChainRuleEditor';

/**
 * 链式代理配置主对话框
 * 支持移动端响应式布局
 */
export default function ChainProxyDialog({ open, onClose, subscription }) {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));

  const [loading, setLoading] = useState(false);
  const [rules, setRules] = useState([]);
  const [options, setOptions] = useState({ nodes: [], conditionFields: [], operators: [], groupTypes: [], templateGroups: [] });
  const [editingRule, setEditingRule] = useState(null);
  const [editMode, setEditMode] = useState(false); // false: 列表模式, true: 编辑模式
  const [previewOpen, setPreviewOpen] = useState(false);
  const [previewLoading, setPreviewLoading] = useState(false);
  const [previewData, setPreviewData] = useState(null);

  // 加载数据
  const loadData = useCallback(async () => {
    if (!subscription?.ID) return;
    setLoading(true);
    try {
      const [rulesRes, optionsRes] = await Promise.all([getChainRules(subscription.ID), getChainOptions(subscription.ID)]);
      // request 拦截器已返回 response.data，所以这里直接取 .data
      setRules(rulesRes?.data || []);
      setOptions(optionsRes?.data || { nodes: [], conditionFields: [], operators: [], groupTypes: [], templateGroups: [] });
    } catch (err) {
      console.error('加载链式代理数据失败:', err);
    } finally {
      setLoading(false);
    }
  }, [subscription?.ID]);

  useEffect(() => {
    if (open && subscription?.ID) {
      loadData();
    }
  }, [open, subscription?.ID, loadData]);

  // 添加规则
  const handleAdd = () => {
    setEditingRule({
      name: '',
      enabled: true,
      chainConfig: '[]',
      targetConfig: '{"type":"specified_node"}'
    });
    setEditMode(true);
  };

  // 编辑规则
  const handleEdit = (rule) => {
    setEditingRule(rule);
    setEditMode(true);
  };

  // 保存规则
  const handleSave = async () => {
    if (!editingRule) return;

    setLoading(true);
    try {
      if (editingRule.id) {
        // 更新
        await updateChainRule(subscription.ID, editingRule.id, editingRule);
      } else {
        // 创建
        await createChainRule(subscription.ID, editingRule);
      }
      await loadData();
      setEditMode(false);
      setEditingRule(null);
    } catch (err) {
      console.error('保存规则失败:', err);
    } finally {
      setLoading(false);
    }
  };

  // 删除规则
  const handleDelete = async (rule) => {
    if (!window.confirm(`确定删除规则「${rule.name}」吗？`)) return;

    setLoading(true);
    try {
      await deleteChainRule(subscription.ID, rule.id);
      await loadData();
    } catch (err) {
      console.error('删除规则失败:', err);
    } finally {
      setLoading(false);
    }
  };

  // 切换启用状态
  const handleToggle = async (rule) => {
    try {
      await toggleChainRule(subscription.ID, rule.id);
      await loadData();
    } catch (err) {
      console.error('切换规则状态失败:', err);
    }
  };

  // 拖拽排序
  const handleDragEnd = async (result) => {
    if (!result.destination) return;

    const items = Array.from(rules);
    const [reorderedItem] = items.splice(result.source.index, 1);
    items.splice(result.destination.index, 0, reorderedItem);
    setRules(items);

    // 保存排序
    try {
      await sortChainRules(
        subscription.ID,
        items.map((r) => r.id)
      );
    } catch (err) {
      console.error('保存排序失败:', err);
      await loadData(); // 恢复原顺序
    }
  };

  // 规则编辑器数据变化
  const handleRuleChange = (data) => {
    setEditingRule({ ...editingRule, ...data });
  };

  // 返回列表
  const handleBack = () => {
    setEditMode(false);
    setEditingRule(null);
  };

  // 预览整体链路
  const handlePreview = async () => {
    setPreviewOpen(true);
    setPreviewLoading(true);
    try {
      const res = await previewChainLinks(subscription.ID);
      setPreviewData(res?.data || null);
    } catch (err) {
      console.error('预览链路失败:', err);
      setPreviewData(null);
    } finally {
      setPreviewLoading(false);
    }
  };

  // 关闭预览
  const handleClosePreview = () => {
    setPreviewOpen(false);
    setPreviewData(null);
  };

  // 获取节点类型的友好显示名称
  const getTypeFriendlyName = (type) => {
    const labels = {
      template_group: '模板组',
      custom_group: '自定义组',
      dynamic_node: '动态节点',
      specified_node: '指定节点'
    };
    return labels[type] || type;
  };

  // 解析代理链配置用于显示
  const parseChainConfig = (configStr) => {
    try {
      const config = JSON.parse(configStr || '[]');
      return config
        .map((item) => {
          // 指定节点：显示实际选择的节点名称
          if (item.type === 'specified_node') {
            if (item.nodeId) {
              const node = (options.nodes || []).find((n) => n.id === item.nodeId);
              if (node) {
                return node.name || node.linkName || `节点 #${item.nodeId}`;
              }
              return `节点 #${item.nodeId}`;
            }
            return '指定节点';
          }
          // 动态节点：显示"动态节点"
          if (item.type === 'dynamic_node') {
            return '动态节点';
          }
          // 自定义组和模板组：优先显示组名，否则显示类型名称
          if (item.type === 'custom_group' || item.type === 'template_group') {
            return item.groupName || getTypeFriendlyName(item.type);
          }
          // 其他情况：显示组名或类型的友好名称
          return item.groupName || getTypeFriendlyName(item.type);
        })
        .filter(Boolean);
    } catch {
      return [];
    }
  };

  // 解析目标配置用于显示
  const parseTargetConfig = (configStr) => {
    try {
      const config = JSON.parse(configStr || '{}');
      if (config.type === 'all') return '所有节点';
      if (config.type === 'specified_node') {
        if (config.nodeId) {
          // 尝试从 options.nodes 查找节点名称
          const node = (options.nodes || []).find((n) => n.id === config.nodeId);
          if (node) {
            return node.name || node.linkName || `节点 #${config.nodeId}`;
          }
          return `节点 #${config.nodeId}`;
        }
        return '未选择节点';
      }
      if (config.type === 'conditions' && config.conditions?.conditions?.length > 0) {
        return `${config.conditions.conditions.length} 个条件`;
      }
      return '未配置';
    } catch {
      return '未配置';
    }
  };

  return (
    <>
      <Dialog
        open={open}
        onClose={onClose}
        maxWidth="lg"
        fullWidth
        fullScreen={isMobile}
        PaperProps={{
          sx: isMobile ? { borderRadius: 0 } : { minHeight: '80vh', borderRadius: 2 }
        }}
      >
        <DialogTitle sx={{ pb: 1.5 }}>
          <Stack direction="row" alignItems="center" justifyContent="space-between">
            <Stack direction="row" alignItems="center" spacing={1}>
              <AccountTreeIcon color="primary" />
              <Box>
                <Typography variant={isMobile ? 'subtitle1' : 'h6'} fontWeight={600}>
                  链式代理配置
                  <Chip size="small" label="Beta" color="error" variant="outlined" sx={{ ml: 1 }} />
                </Typography>
                {isMobile && (
                  <Typography variant="caption" color="text.secondary">
                    {subscription?.Name}
                  </Typography>
                )}
                {!isMobile && (
                  <Typography component="span" variant="body2" color="text.secondary" sx={{ ml: 1 }}>
                    - {subscription?.Name}
                  </Typography>
                )}
              </Box>
            </Stack>
            <IconButton onClick={onClose} size="small">
              <CloseIcon />
            </IconButton>
          </Stack>
        </DialogTitle>

        <DialogContent dividers sx={{ p: isMobile ? 2 : 3 }}>
          {loading && (
            <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
              <CircularProgress />
            </Box>
          )}

          {!loading && !editMode && (
            <Box>
              {/* 说明文字 */}
              <Typography variant="body2" color="text.secondary" sx={{ mb: 1.5 }}>
                链式代理规则用于配置节点的前置代理（入口代理）。每个落地节点独立匹配规则：按规则顺序检查，应用第一个匹配规则的出口配置。
              </Typography>
              <Typography
                variant="caption"
                color="info.main"
                sx={{ mb: 2, display: 'block', backgroundColor: 'info.lighter', p: 1, borderRadius: 1 }}
              >
                💡
                提示：若需不同落地节点使用不同出口，请确保各规则的「目标节点」配置不重叠（使用「指定节点」或「按条件筛选」精确定义目标范围）。
              </Typography>

              {/* 规则列表 - 移动端使用卡片布局 */}
              {isMobile ? (
                // 移动端卡片布局
                <Stack spacing={1.5}>
                  {rules.map((rule) => (
                    <Card
                      key={rule.id}
                      variant="outlined"
                      sx={{
                        borderRadius: 2,
                        opacity: rule.enabled ? 1 : 0.6,
                        transition: 'all 0.2s ease',
                        '&:active': { transform: 'scale(0.98)' }
                      }}
                    >
                      <CardContent sx={{ py: 1.5, px: 2, '&:last-child': { pb: 1.5 } }}>
                        <Stack spacing={1.5}>
                          {/* 规则名称和状态 */}
                          <Stack direction="row" alignItems="center" justifyContent="space-between">
                            <Stack direction="row" alignItems="center" spacing={1}>
                              <Typography variant="subtitle1" fontWeight={600}>
                                {rule.name || '未命名规则'}
                              </Typography>
                              {!rule.enabled && <Chip label="已禁用" size="small" color="default" />}
                            </Stack>
                            <Switch checked={rule.enabled} onChange={() => handleToggle(rule)} size="small" />
                          </Stack>

                          {/* 代理链显示 */}
                          <Stack direction="row" alignItems="center" spacing={0.5} flexWrap="wrap" sx={{ gap: 0.5 }}>
                            {parseChainConfig(rule.chainConfig).map((name, i) => (
                              <Chip key={i} label={name} size="small" color="primary" variant="outlined" sx={{ borderRadius: 1.5 }} />
                            ))}
                            <ArrowForwardIcon sx={{ fontSize: 16, color: 'text.secondary', mx: 0.5 }} />
                            <Typography variant="body2" color="text.secondary">
                              {parseTargetConfig(rule.targetConfig)}
                            </Typography>
                          </Stack>

                          {/* 操作按钮 - 移动端更大的触摸区域 */}
                          <Stack direction="row" spacing={1} justifyContent="flex-end">
                            <Button
                              size="small"
                              variant="outlined"
                              startIcon={<EditIcon />}
                              onClick={() => handleEdit(rule)}
                              sx={{ minWidth: 80 }}
                            >
                              编辑
                            </Button>
                            <Button
                              size="small"
                              variant="outlined"
                              color="error"
                              startIcon={<DeleteIcon />}
                              onClick={() => handleDelete(rule)}
                              sx={{ minWidth: 80 }}
                            >
                              删除
                            </Button>
                          </Stack>
                        </Stack>
                      </CardContent>
                    </Card>
                  ))}
                </Stack>
              ) : (
                // 桌面端拖拽布局
                <DragDropContext onDragEnd={handleDragEnd}>
                  <Droppable droppableId="chain-rules">
                    {(provided) => (
                      <Stack spacing={1} {...provided.droppableProps} ref={provided.innerRef}>
                        {rules.map((rule, index) => (
                          <Draggable key={rule.id} draggableId={String(rule.id)} index={index}>
                            {(provided, snapshot) => (
                              <Paper
                                ref={provided.innerRef}
                                {...provided.draggableProps}
                                variant="outlined"
                                sx={{
                                  p: 2,
                                  backgroundColor: snapshot.isDragging ? 'action.hover' : 'background.paper',
                                  opacity: rule.enabled ? 1 : 0.6
                                }}
                              >
                                <Stack direction="row" alignItems="center" spacing={2}>
                                  {/* 拖拽手柄 */}
                                  <Box {...provided.dragHandleProps} sx={{ cursor: 'grab', color: 'text.secondary' }}>
                                    <DragIndicatorIcon />
                                  </Box>

                                  {/* 规则信息 */}
                                  <Box sx={{ flex: 1 }}>
                                    <Stack direction="row" alignItems="center" spacing={1}>
                                      <Typography variant="subtitle2">{rule.name || '未命名规则'}</Typography>
                                      {!rule.enabled && <Chip label="已禁用" size="small" color="default" />}
                                    </Stack>
                                    <Stack direction="row" spacing={1} sx={{ mt: 0.5 }}>
                                      {parseChainConfig(rule.chainConfig).map((name, i) => (
                                        <Chip key={i} label={name} size="small" color="primary" variant="outlined" />
                                      ))}
                                      <Typography variant="caption" color="text.secondary">
                                        → {parseTargetConfig(rule.targetConfig)}
                                      </Typography>
                                    </Stack>
                                  </Box>

                                  {/* 操作按钮 */}
                                  <Switch checked={rule.enabled} onChange={() => handleToggle(rule)} size="small" />
                                  <IconButton size="small" onClick={() => handleEdit(rule)}>
                                    <EditIcon fontSize="small" />
                                  </IconButton>
                                  <IconButton size="small" color="error" onClick={() => handleDelete(rule)}>
                                    <DeleteIcon fontSize="small" />
                                  </IconButton>
                                </Stack>
                              </Paper>
                            )}
                          </Draggable>
                        ))}
                        {provided.placeholder}
                      </Stack>
                    )}
                  </Droppable>
                </DragDropContext>
              )}

              {/* 空状态 */}
              {rules.length === 0 && (
                <Paper
                  variant="outlined"
                  sx={{
                    p: isMobile ? 3 : 4,
                    textAlign: 'center',
                    borderRadius: 2
                  }}
                >
                  <TouchAppIcon sx={{ fontSize: 48, color: 'text.secondary', mb: 1 }} />
                  <Typography color="text.secondary" gutterBottom>
                    暂无链式代理规则
                  </Typography>
                  <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                    链式代理可以为节点配置入口代理，实现流量中转
                  </Typography>
                  <Button variant="contained" startIcon={<AddIcon />} onClick={handleAdd} size={isMobile ? 'large' : 'medium'}>
                    添加规则
                  </Button>
                </Paper>
              )}
            </Box>
          )}

          {/* 编辑模式 */}
          {!loading && editMode && editingRule && (
            <ChainRuleEditor
              value={editingRule}
              onChange={handleRuleChange}
              nodes={options.nodes || []}
              fields={options.conditionFields || []}
              operators={options.operators || []}
              groupTypes={options.groupTypes || []}
              templateGroups={options.templateGroups || []}
              isMobile={isMobile}
            />
          )}
        </DialogContent>

        <DialogActions sx={{ px: isMobile ? 2 : 3, py: 1.5 }}>
          {!editMode ? (
            <>
              <Button onClick={onClose}>关闭</Button>
              {rules.length > 0 && (
                <>
                  <Button variant="outlined" color="info" startIcon={<VisibilityIcon />} onClick={handlePreview}>
                    预览整体链路
                  </Button>
                  <Button variant="contained" startIcon={<AddIcon />} onClick={handleAdd}>
                    添加规则
                  </Button>
                </>
              )}
            </>
          ) : (
            <>
              <Button onClick={handleBack}>返回列表</Button>
              <Button variant="contained" onClick={handleSave} disabled={loading}>
                保存规则
              </Button>
            </>
          )}
        </DialogActions>
      </Dialog>

      {/* 链路预览对话框 */}
      <ChainPreviewDialog open={previewOpen} onClose={handleClosePreview} loading={previewLoading} data={previewData} />
    </>
  );
}
