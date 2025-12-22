import { useState } from 'react';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Typography from '@mui/material/Typography';
import IconButton from '@mui/material/IconButton';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import Chip from '@mui/material/Chip';
import Autocomplete from '@mui/material/Autocomplete';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';

import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import ArrowForwardIcon from '@mui/icons-material/ArrowForward';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import StopIcon from '@mui/icons-material/Stop';
import GroupWorkIcon from '@mui/icons-material/GroupWork';
import FilterAltIcon from '@mui/icons-material/FilterAlt';
import DeviceHubIcon from '@mui/icons-material/DeviceHub';

import ConditionBuilder from './ConditionBuilder';

/**
 * 移动端链式代理配置器
 * 使用简化的卡片式布局替代流程图画板
 */
export default function MobileChainBuilder({
  chainConfig = [],
  targetConfig = { type: 'all', conditions: null },
  onChainConfigChange,
  onTargetConfigChange,
  nodes: availableNodes = [],
  fields = [],
  operators = [],
  groupTypes = [],
  templateGroups = []
}) {
  // 编辑对话框状态
  const [proxyDialogOpen, setProxyDialogOpen] = useState(false);
  const [targetDialogOpen, setTargetDialogOpen] = useState(false);
  const [editingProxyConfig, setEditingProxyConfig] = useState(null);
  const [editingTargetConfig, setEditingTargetConfig] = useState(null);

  // 获取代理类型标签
  const getTypeLabel = (type) => {
    const labels = {
      template_group: '模板组',
      custom_group: '自定义组',
      dynamic_node: '动态节点',
      specified_node: '指定节点'
    };
    return labels[type] || '代理';
  };

  // 获取代理类型图标
  const getTypeIcon = (type) => {
    switch (type) {
      case 'template_group':
        return <GroupWorkIcon color="primary" />;
      case 'custom_group':
        return <DeviceHubIcon color="secondary" />;
      case 'dynamic_node':
        return <FilterAltIcon color="warning" />;
      case 'specified_node':
        return <DeviceHubIcon color="success" />;
      default:
        return <GroupWorkIcon color="primary" />;
    }
  };

  // 获取代理显示名称
  const getProxyLabel = (item) => {
    if (!item) return '未配置';
    if (item.type === 'specified_node') {
      const node = availableNodes.find((n) => n.id === item.nodeId);
      return node?.name || node?.linkName || `节点 #${item.nodeId}`;
    }
    if (item.type === 'dynamic_node') {
      const condCount = item.nodeConditions?.conditions?.length || 0;
      return condCount > 0 ? `${condCount} 个条件` : '未配置';
    }
    return item.groupName || '未配置';
  };

  // 获取目标配置标签
  const getTargetLabel = () => {
    switch (targetConfig?.type) {
      case 'all':
        return '所有节点';
      case 'specified_node':
        if (targetConfig?.nodeId) {
          const node = availableNodes.find((n) => n.id === targetConfig.nodeId);
          return node?.name || node?.linkName || `节点 #${targetConfig.nodeId}`;
        }
        return '未选择节点';
      case 'conditions':
        const condCount = targetConfig?.conditions?.conditions?.length || 0;
        return condCount > 0 ? `${condCount} 个条件` : '未配置';
      default:
        return '未配置';
    }
  };

  // 添加代理节点
  const handleAddProxy = () => {
    const isEntryNode = chainConfig.length === 0;
    // 入口节点默认使用模板代理组，中间节点默认使用自定义代理组
    const defaultType = isEntryNode ? 'template_group' : 'custom_group';
    const newConfig = { type: defaultType, groupName: '' };
    setEditingProxyConfig({ isNew: true, index: chainConfig.length, config: newConfig });
    setProxyDialogOpen(true);
  };

  // 编辑代理节点
  const handleEditProxy = (index) => {
    const config = { ...chainConfig[index] };
    // 中间节点（索引 > 0）不支持模板代理组，自动修正为自定义代理组
    if (index > 0 && config.type === 'template_group') {
      config.type = 'custom_group';
    }
    setEditingProxyConfig({ isNew: false, index, config });
    setProxyDialogOpen(true);
  };

  // 删除代理节点
  const handleDeleteProxy = (index) => {
    const newConfig = chainConfig.filter((_, i) => i !== index);
    onChainConfigChange?.(newConfig);
  };

  // 保存代理配置
  const handleSaveProxy = () => {
    if (!editingProxyConfig) return;
    const newChainConfig = [...chainConfig];
    if (editingProxyConfig.isNew) {
      newChainConfig.push(editingProxyConfig.config);
    } else {
      newChainConfig[editingProxyConfig.index] = editingProxyConfig.config;
    }
    onChainConfigChange?.(newChainConfig);
    setProxyDialogOpen(false);
    setEditingProxyConfig(null);
  };

  // 编辑目标配置
  const handleEditTarget = () => {
    setEditingTargetConfig({ ...targetConfig });
    setTargetDialogOpen(true);
  };

  // 保存目标配置
  const handleSaveTarget = () => {
    if (!editingTargetConfig) return;
    onTargetConfigChange?.(editingTargetConfig);
    setTargetDialogOpen(false);
    setEditingTargetConfig(null);
  };

  return (
    <Box>
      {/* 流程可视化 */}
      <Stack spacing={1.5}>
        {/* 入口节点 */}
        <Card
          variant="outlined"
          sx={{
            background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
            color: 'white',
            borderRadius: 2
          }}
        >
          <CardContent sx={{ py: 1.5, px: 2, '&:last-child': { pb: 1.5 } }}>
            <Stack direction="row" alignItems="center" spacing={1}>
              <PlayArrowIcon />
              <Typography variant="subtitle2" fontWeight={600} color="inherit">
                入口
              </Typography>
            </Stack>
          </CardContent>
        </Card>

        {/* 箭头连接线 */}
        {(chainConfig.length > 0 || true) && (
          <Box sx={{ display: 'flex', justifyContent: 'center' }}>
            <ArrowForwardIcon sx={{ color: 'text.secondary', transform: 'rotate(90deg)' }} />
          </Box>
        )}

        {/* 代理节点列表 */}
        {chainConfig.map((item, index) => (
          <Box key={index}>
            <Card
              variant="outlined"
              sx={{
                borderRadius: 2,
                borderColor: 'primary.main',
                borderWidth: 2
              }}
            >
              <CardContent sx={{ py: 1.5, px: 2, '&:last-child': { pb: 1.5 } }}>
                <Stack direction="row" alignItems="center" justifyContent="space-between">
                  <Stack direction="row" alignItems="center" spacing={1.5}>
                    {getTypeIcon(item.type)}
                    <Box>
                      <Typography variant="caption" color="text.secondary">
                        {getTypeLabel(item.type)}
                      </Typography>
                      <Typography variant="body2" fontWeight={600}>
                        {getProxyLabel(item)}
                      </Typography>
                    </Box>
                  </Stack>
                  <Stack direction="row" spacing={0.5}>
                    <IconButton size="small" onClick={() => handleEditProxy(index)}>
                      <EditIcon fontSize="small" />
                    </IconButton>
                    <IconButton size="small" color="error" onClick={() => handleDeleteProxy(index)}>
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  </Stack>
                </Stack>
              </CardContent>
            </Card>
            <Box sx={{ display: 'flex', justifyContent: 'center', my: 0.5 }}>
              <ArrowForwardIcon sx={{ color: 'text.secondary', transform: 'rotate(90deg)' }} />
            </Box>
          </Box>
        ))}

        {/* 添加代理按钮 - 支持多级链式代理 */}
        {chainConfig.length < 4 && (
          <>
            <Button variant="outlined" startIcon={<AddIcon />} onClick={handleAddProxy} fullWidth sx={{ borderStyle: 'dashed', py: 1.5 }}>
              {chainConfig.length === 0 ? '添加入口代理' : '添加中间代理'}
            </Button>
            {chainConfig.length >= 2 && (
              <Typography variant="caption" color="warning.main" sx={{ textAlign: 'center', display: 'block', mt: 0.5 }}>
                当前 {chainConfig.length} 级链路，延迟可能较高
              </Typography>
            )}
            <Box sx={{ display: 'flex', justifyContent: 'center' }}>
              <ArrowForwardIcon sx={{ color: 'text.secondary', transform: 'rotate(90deg)' }} />
            </Box>
          </>
        )}
        {chainConfig.length >= 4 && (
          <Typography variant="caption" color="text.secondary" sx={{ textAlign: 'center', display: 'block', py: 1 }}>
            已达最大 4 级链路
          </Typography>
        )}

        {/* 目标节点 */}
        <Card
          variant="outlined"
          onClick={handleEditTarget}
          sx={{
            background: 'linear-gradient(135deg, #00897b 0%, #43a047 100%)',
            color: 'white',
            borderRadius: 2,
            cursor: 'pointer',
            transition: 'transform 0.2s',
            '&:active': { transform: 'scale(0.98)' }
          }}
        >
          <CardContent sx={{ py: 1.5, px: 2, '&:last-child': { pb: 1.5 } }}>
            <Stack direction="row" alignItems="center" justifyContent="space-between">
              <Stack direction="row" alignItems="center" spacing={1}>
                <StopIcon />
                <Box>
                  <Typography variant="caption" sx={{ opacity: 0.9 }} color="inherit">
                    目标节点
                  </Typography>
                  <Typography variant="subtitle2" fontWeight={600} color="inherit">
                    {getTargetLabel()}
                  </Typography>
                </Box>
              </Stack>
              <EditIcon fontSize="small" sx={{ opacity: 0.8 }} />
            </Stack>
          </CardContent>
        </Card>
      </Stack>

      {/* 代理节点编辑对话框 */}
      <Dialog open={proxyDialogOpen} onClose={() => setProxyDialogOpen(false)} fullWidth maxWidth="sm">
        <DialogTitle>
          {editingProxyConfig?.isNew
            ? editingProxyConfig?.index === 0
              ? '添加入口代理'
              : '添加中间代理'
            : editingProxyConfig?.index === 0
              ? '编辑入口代理'
              : '编辑中间代理'}
        </DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ pt: 1 }}>
            {/* 入口节点（索引0）可选所有类型，后续中间节点只能选择指定节点或动态条件节点 */}
            {(() => {
              const isEntryNode = editingProxyConfig?.index === 0;
              return (
                <FormControl size="small" fullWidth>
                  <InputLabel>代理类型</InputLabel>
                  <Select
                    value={editingProxyConfig?.config?.type || (isEntryNode ? 'template_group' : 'specified_node')}
                    label="代理类型"
                    onChange={(e) =>
                      setEditingProxyConfig({
                        ...editingProxyConfig,
                        config: {
                          type: e.target.value,
                          groupName: '',
                          nodeId: undefined,
                          nodeConditions: undefined
                        }
                      })
                    }
                  >
                    {/* 入口节点可选模板代理组 */}
                    {isEntryNode && <MenuItem value="template_group">模板代理组</MenuItem>}
                    {/* 所有节点都可选自定义代理组（中间节点的组内节点会自动设置 dialer-proxy） */}
                    <MenuItem value="custom_group">自定义代理组</MenuItem>
                    <MenuItem value="dynamic_node">动态条件节点</MenuItem>
                    <MenuItem value="specified_node">指定节点</MenuItem>
                  </Select>
                  {!isEntryNode && (
                    <Typography variant="caption" color="text.secondary" sx={{ mt: 0.5 }}>
                      中间节点的组内节点会自动设置 dialer-proxy
                    </Typography>
                  )}
                </FormControl>
              );
            })()}

            {editingProxyConfig?.config?.type === 'template_group' && (
              <Autocomplete
                freeSolo
                size="small"
                fullWidth
                options={templateGroups || []}
                value={editingProxyConfig?.config?.groupName || ''}
                onChange={(e, newValue) =>
                  setEditingProxyConfig({
                    ...editingProxyConfig,
                    config: { ...editingProxyConfig.config, groupName: newValue || '' }
                  })
                }
                onInputChange={(e, newValue) =>
                  setEditingProxyConfig({
                    ...editingProxyConfig,
                    config: { ...editingProxyConfig.config, groupName: newValue || '' }
                  })
                }
                renderInput={(params) => <TextField {...params} label="代理组名称" placeholder="选择或输入代理组名称" />}
              />
            )}

            {editingProxyConfig?.config?.type === 'custom_group' && (
              <>
                <TextField
                  size="small"
                  fullWidth
                  label="代理组名称"
                  placeholder="自定义代理组名称"
                  value={editingProxyConfig?.config?.groupName || ''}
                  onChange={(e) =>
                    setEditingProxyConfig({
                      ...editingProxyConfig,
                      config: { ...editingProxyConfig.config, groupName: e.target.value }
                    })
                  }
                />
                <FormControl size="small" fullWidth>
                  <InputLabel>组类型</InputLabel>
                  <Select
                    value={editingProxyConfig?.config?.groupType || 'select'}
                    label="组类型"
                    onChange={(e) =>
                      setEditingProxyConfig({
                        ...editingProxyConfig,
                        config: { ...editingProxyConfig.config, groupType: e.target.value }
                      })
                    }
                  >
                    {(groupTypes || []).map((gt) => (
                      <MenuItem key={gt.value} value={gt.value}>
                        {gt.label}
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>
                <ConditionBuilder
                  title="节点筛选条件"
                  value={editingProxyConfig?.config?.nodeConditions}
                  onChange={(conds) =>
                    setEditingProxyConfig({
                      ...editingProxyConfig,
                      config: { ...editingProxyConfig.config, nodeConditions: conds }
                    })
                  }
                  fields={fields}
                  operators={operators}
                />
              </>
            )}

            {editingProxyConfig?.config?.type === 'dynamic_node' && (
              <>
                <FormControl size="small" fullWidth>
                  <InputLabel>选择模式</InputLabel>
                  <Select
                    value={editingProxyConfig?.config?.selectMode || 'first'}
                    label="选择模式"
                    onChange={(e) =>
                      setEditingProxyConfig({
                        ...editingProxyConfig,
                        config: { ...editingProxyConfig.config, selectMode: e.target.value }
                      })
                    }
                  >
                    <MenuItem value="first">第一个匹配</MenuItem>
                    <MenuItem value="random">随机</MenuItem>
                    <MenuItem value="fastest">最快节点</MenuItem>
                  </Select>
                </FormControl>
                <ConditionBuilder
                  title="节点匹配条件"
                  value={editingProxyConfig?.config?.nodeConditions}
                  onChange={(conds) =>
                    setEditingProxyConfig({
                      ...editingProxyConfig,
                      config: { ...editingProxyConfig.config, nodeConditions: conds }
                    })
                  }
                  fields={fields}
                  operators={operators}
                />
              </>
            )}

            {editingProxyConfig?.config?.type === 'specified_node' && (
              <Autocomplete
                size="small"
                options={availableNodes || []}
                getOptionLabel={(option) => `${option.name || option.linkName} (${option.linkCountry || '未知'})`}
                value={(availableNodes || []).find((n) => n.id === editingProxyConfig?.config?.nodeId) || null}
                onChange={(e, newValue) =>
                  setEditingProxyConfig({
                    ...editingProxyConfig,
                    config: { ...editingProxyConfig.config, nodeId: newValue?.id }
                  })
                }
                renderInput={(params) => <TextField {...params} label="选择节点" />}
                renderOption={(props, option) => (
                  <li {...props} key={option.id}>
                    <Stack direction="row" spacing={1} alignItems="center">
                      <Typography variant="body2">{option.name || option.linkName}</Typography>
                      <Chip label={option.linkCountry || '未知'} size="small" variant="outlined" />
                    </Stack>
                  </li>
                )}
              />
            )}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setProxyDialogOpen(false)}>取消</Button>
          <Button variant="contained" onClick={handleSaveProxy}>
            确定
          </Button>
        </DialogActions>
      </Dialog>

      {/* 目标节点编辑对话框 */}
      <Dialog open={targetDialogOpen} onClose={() => setTargetDialogOpen(false)} fullWidth maxWidth="sm">
        <DialogTitle>配置目标节点</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ pt: 1 }}>
            <Typography variant="body2" color="text.secondary">
              选择应用此规则的节点范围
            </Typography>

            <FormControl size="small" fullWidth>
              <InputLabel>目标类型</InputLabel>
              <Select
                value={editingTargetConfig?.type || 'specified_node'}
                label="目标类型"
                onChange={(e) =>
                  setEditingTargetConfig({
                    ...editingTargetConfig,
                    type: e.target.value,
                    nodeId: undefined,
                    conditions: undefined
                  })
                }
              >
                <MenuItem value="specified_node">指定节点</MenuItem>
                <MenuItem value="all">所有节点</MenuItem>
                <MenuItem value="conditions">按条件筛选</MenuItem>
              </Select>
            </FormControl>

            {editingTargetConfig?.type === 'specified_node' && (
              <Autocomplete
                size="small"
                options={availableNodes || []}
                getOptionLabel={(option) => `${option.name || option.linkName} (${option.linkCountry || '未知'})`}
                value={(availableNodes || []).find((n) => n.id === editingTargetConfig?.nodeId) || null}
                onChange={(e, newValue) =>
                  setEditingTargetConfig({
                    ...editingTargetConfig,
                    nodeId: newValue?.id
                  })
                }
                renderInput={(params) => <TextField {...params} label="选择目标节点" placeholder="搜索节点..." />}
                renderOption={(props, option) => (
                  <li {...props} key={option.id}>
                    <Stack direction="row" spacing={1} alignItems="center">
                      <Typography variant="body2">{option.name || option.linkName}</Typography>
                      <Chip label={option.linkCountry || '未知'} size="small" variant="outlined" />
                      <Chip label={option.protocol || 'unknown'} size="small" color="info" variant="outlined" />
                    </Stack>
                  </li>
                )}
              />
            )}

            {editingTargetConfig?.type === 'conditions' && (
              <ConditionBuilder
                title="目标节点筛选条件"
                value={editingTargetConfig?.conditions}
                onChange={(conds) =>
                  setEditingTargetConfig({
                    ...editingTargetConfig,
                    conditions: conds
                  })
                }
                fields={fields}
                operators={operators}
              />
            )}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setTargetDialogOpen(false)}>取消</Button>
          <Button variant="contained" onClick={handleSaveTarget}>
            确定
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
