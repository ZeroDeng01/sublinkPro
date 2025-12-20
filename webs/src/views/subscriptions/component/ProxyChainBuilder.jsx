import { useState, useEffect } from 'react';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Paper from '@mui/material/Paper';
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
import Collapse from '@mui/material/Collapse';
import Divider from '@mui/material/Divider';

import DeleteIcon from '@mui/icons-material/Delete';
import AddIcon from '@mui/icons-material/Add';
import ArrowForwardIcon from '@mui/icons-material/ArrowForward';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';

import ConditionBuilder from './ConditionBuilder';

/**
 * 代理链可视化构建器
 * 用于配置入口代理的类型和参数
 */
export default function ProxyChainBuilder({ value = [], onChange, nodes = [], fields = [], operators = [], groupTypes = [] }) {
  const [chainItems, setChainItems] = useState(value || []);
  const [expandedIndex, setExpandedIndex] = useState(0);

  // 当外部 value 变化时更新内部状态
  useEffect(() => {
    if (value && Array.isArray(value)) {
      setChainItems(value);
    }
  }, [value]);

  // 通知父组件数据变化
  const notifyChange = (newItems) => {
    onChange?.(newItems);
  };

  // 添加代理项
  const handleAddItem = () => {
    const newItems = [
      ...chainItems,
      {
        type: 'template_group',
        groupName: ''
      }
    ];
    setChainItems(newItems);
    setExpandedIndex(newItems.length - 1);
    notifyChange(newItems);
  };

  // 删除代理项
  const handleRemoveItem = (index) => {
    const newItems = chainItems.filter((_, i) => i !== index);
    setChainItems(newItems);
    notifyChange(newItems);
  };

  // 更新代理项
  const handleItemChange = (index, updates) => {
    const newItems = chainItems.map((item, i) => {
      if (i === index) {
        return { ...item, ...updates };
      }
      return item;
    });
    setChainItems(newItems);
    notifyChange(newItems);
  };

  // 切换展开状态
  const toggleExpand = (index) => {
    setExpandedIndex(expandedIndex === index ? -1 : index);
  };

  // 获取代理类型标签
  const getTypeLabel = (type) => {
    const labels = {
      template_group: '模板代理组',
      custom_group: '自定义代理组',
      dynamic_node: '动态条件节点',
      specified_node: '指定节点'
    };
    return labels[type] || type;
  };

  // 获取代理类型颜色
  const getTypeColor = (type) => {
    const colors = {
      template_group: 'primary',
      custom_group: 'secondary',
      dynamic_node: 'warning',
      specified_node: 'success'
    };
    return colors[type] || 'default';
  };

  // 渲染单个代理项配置
  const renderItemConfig = (item, index) => {
    const isExpanded = expandedIndex === index;

    return (
      <Paper
        key={index}
        variant="outlined"
        sx={{
          p: 2,
          position: 'relative',
          borderColor: isExpanded ? 'primary.main' : 'divider'
        }}
      >
        {/* 代理项头部 */}
        <Stack direction="row" alignItems="center" spacing={1}>
          <Chip label={getTypeLabel(item.type)} color={getTypeColor(item.type)} size="small" />
          <Typography variant="body2" sx={{ flex: 1 }}>
            {item.groupName || item.nodeId ? (
              item.type === 'specified_node' ? (
                nodes.find((n) => n.id === item.nodeId)?.name || `节点 #${item.nodeId}`
              ) : (
                item.groupName
              )
            ) : (
              <em style={{ color: 'gray' }}>未配置</em>
            )}
          </Typography>
          <IconButton size="small" onClick={() => toggleExpand(index)}>
            {isExpanded ? <ExpandLessIcon /> : <ExpandMoreIcon />}
          </IconButton>
          <IconButton size="small" color="error" onClick={() => handleRemoveItem(index)}>
            <DeleteIcon fontSize="small" />
          </IconButton>
        </Stack>

        {/* 展开的配置区域 */}
        <Collapse in={isExpanded}>
          <Box sx={{ mt: 2 }}>
            <Stack spacing={2}>
              {/* 代理类型选择 */}
              <FormControl size="small" fullWidth>
                <InputLabel>代理类型</InputLabel>
                <Select
                  value={item.type}
                  label="代理类型"
                  onChange={(e) =>
                    handleItemChange(index, {
                      type: e.target.value,
                      groupName: '',
                      nodeId: undefined,
                      nodeConditions: undefined
                    })
                  }
                >
                  <MenuItem value="template_group">模板代理组</MenuItem>
                  <MenuItem value="custom_group">自定义代理组</MenuItem>
                  <MenuItem value="dynamic_node">动态条件节点</MenuItem>
                  <MenuItem value="specified_node">指定节点</MenuItem>
                </Select>
              </FormControl>

              {/* 模板代理组配置 */}
              {item.type === 'template_group' && (
                <TextField
                  size="small"
                  fullWidth
                  label="代理组名称"
                  placeholder="输入模板中的代理组名称"
                  value={item.groupName || ''}
                  onChange={(e) => handleItemChange(index, { groupName: e.target.value })}
                  helperText="输入 Clash 模板中已存在的代理组名称"
                />
              )}

              {/* 自定义代理组配置 */}
              {item.type === 'custom_group' && (
                <>
                  <TextField
                    size="small"
                    fullWidth
                    label="代理组名称"
                    placeholder="自定义代理组名称"
                    value={item.groupName || ''}
                    onChange={(e) => handleItemChange(index, { groupName: e.target.value })}
                  />
                  <FormControl size="small" fullWidth>
                    <InputLabel>组类型</InputLabel>
                    <Select
                      value={item.groupType || 'select'}
                      label="组类型"
                      onChange={(e) => handleItemChange(index, { groupType: e.target.value })}
                    >
                      {groupTypes.map((gt) => (
                        <MenuItem key={gt.value} value={gt.value}>
                          {gt.label}
                        </MenuItem>
                      ))}
                    </Select>
                  </FormControl>
                  {item.groupType === 'url-test' && (
                    <Stack direction="row" spacing={1}>
                      <TextField
                        size="small"
                        label="测速 URL"
                        value={item.urlTestConfig?.url || ''}
                        onChange={(e) =>
                          handleItemChange(index, {
                            urlTestConfig: {
                              ...item.urlTestConfig,
                              url: e.target.value
                            }
                          })
                        }
                        sx={{ flex: 2 }}
                        placeholder="http://www.gstatic.com/generate_204"
                      />
                      <TextField
                        size="small"
                        label="间隔(秒)"
                        type="number"
                        value={item.urlTestConfig?.interval || 300}
                        onChange={(e) =>
                          handleItemChange(index, {
                            urlTestConfig: {
                              ...item.urlTestConfig,
                              interval: parseInt(e.target.value) || 300
                            }
                          })
                        }
                        sx={{ flex: 1 }}
                      />
                      <TextField
                        size="small"
                        label="容差(ms)"
                        type="number"
                        value={item.urlTestConfig?.tolerance || 50}
                        onChange={(e) =>
                          handleItemChange(index, {
                            urlTestConfig: {
                              ...item.urlTestConfig,
                              tolerance: parseInt(e.target.value) || 50
                            }
                          })
                        }
                        sx={{ flex: 1 }}
                      />
                    </Stack>
                  )}
                  <ConditionBuilder
                    title="节点筛选条件"
                    value={item.nodeConditions}
                    onChange={(conds) => handleItemChange(index, { nodeConditions: conds })}
                    fields={fields}
                    operators={operators}
                  />
                </>
              )}

              {/* 动态条件节点配置 */}
              {item.type === 'dynamic_node' && (
                <>
                  <FormControl size="small" fullWidth>
                    <InputLabel>选择模式</InputLabel>
                    <Select
                      value={item.selectMode || 'first'}
                      label="选择模式"
                      onChange={(e) => handleItemChange(index, { selectMode: e.target.value })}
                    >
                      <MenuItem value="first">第一个匹配</MenuItem>
                      <MenuItem value="random">随机</MenuItem>
                      <MenuItem value="fastest">最快节点</MenuItem>
                    </Select>
                  </FormControl>
                  <ConditionBuilder
                    title="节点匹配条件"
                    value={item.nodeConditions}
                    onChange={(conds) => handleItemChange(index, { nodeConditions: conds })}
                    fields={fields}
                    operators={operators}
                  />
                </>
              )}

              {/* 指定节点配置 */}
              {item.type === 'specified_node' && (
                <Autocomplete
                  size="small"
                  options={nodes}
                  getOptionLabel={(option) => `${option.name || option.linkName} (${option.linkCountry || '未知'})`}
                  value={nodes.find((n) => n.id === item.nodeId) || null}
                  onChange={(e, newValue) => handleItemChange(index, { nodeId: newValue?.id })}
                  renderInput={(params) => <TextField {...params} label="选择节点" />}
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
            </Stack>
          </Box>
        </Collapse>
      </Paper>
    );
  };

  return (
    <Box>
      <Stack spacing={2}>
        {/* 代理链可视化 */}
        {chainItems.length > 0 && (
          <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap>
            {chainItems.map((item, index) => (
              <Stack key={index} direction="row" alignItems="center" spacing={1}>
                <Chip
                  label={
                    item.type === 'specified_node'
                      ? nodes.find((n) => n.id === item.nodeId)?.name || `节点 #${item.nodeId}`
                      : item.groupName || getTypeLabel(item.type)
                  }
                  color={getTypeColor(item.type)}
                  onClick={() => toggleExpand(index)}
                />
                {index < chainItems.length - 1 && <ArrowForwardIcon color="action" fontSize="small" />}
              </Stack>
            ))}
            <ArrowForwardIcon color="action" fontSize="small" />
            <Chip label="目标节点" variant="outlined" />
          </Stack>
        )}

        <Divider />

        {/* 代理项配置列表 */}
        <Typography variant="subtitle2" color="text.secondary">
          入口代理配置
        </Typography>

        {chainItems.map((item, index) => renderItemConfig(item, index))}

        {/* 添加代理按钮 */}
        <Button variant="outlined" startIcon={<AddIcon />} onClick={handleAddItem} sx={{ alignSelf: 'flex-start' }}>
          添加入口代理
        </Button>

        {/* 空状态提示 */}
        {chainItems.length === 0 && (
          <Typography variant="body2" color="text.secondary" sx={{ fontStyle: 'italic' }}>
            点击上方按钮添加入口代理配置
          </Typography>
        )}
      </Stack>
    </Box>
  );
}
