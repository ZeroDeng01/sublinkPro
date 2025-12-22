import { useState, useCallback, useMemo, useRef, useEffect } from 'react';
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
import Tooltip from '@mui/material/Tooltip';
import Autocomplete from '@mui/material/Autocomplete';
import Divider from '@mui/material/Divider';
import ToggleButton from '@mui/material/ToggleButton';
import ToggleButtonGroup from '@mui/material/ToggleButtonGroup';
import Fade from '@mui/material/Fade';

import { ReactFlow, Controls, useNodesState, useEdgesState, Handle, Position, MarkerType } from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import './ChainFlowBuilder.css';

import DeleteIcon from '@mui/icons-material/Delete';
import AddIcon from '@mui/icons-material/Add';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import StopIcon from '@mui/icons-material/Stop';
import GroupWorkIcon from '@mui/icons-material/GroupWork';
import FilterAltIcon from '@mui/icons-material/FilterAlt';
import DeviceHubIcon from '@mui/icons-material/DeviceHub';
import CloseIcon from '@mui/icons-material/Close';

import ConditionBuilder from './ConditionBuilder';

// 自定义节点样式 - 深色科幻风格
const nodeStyles = {
  start: {
    background: 'linear-gradient(135deg, rgba(139, 92, 246, 0.9) 0%, rgba(91, 33, 182, 0.9) 100%)',
    color: 'white',
    borderRadius: 30,
    minWidth: 90,
    padding: '0 16px',
    height: 40,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontSize: 14,
    fontWeight: 'bold',
    boxShadow: '0 0 20px rgba(139, 92, 246, 0.4), 0 4px 15px rgba(102, 126, 234, 0.3)',
    border: '1px solid rgba(255, 255, 255, 0.2)'
  },
  end: {
    background: 'linear-gradient(135deg, rgba(34, 197, 94, 0.9) 0%, rgba(22, 163, 74, 0.9) 100%)',
    color: 'white',
    borderRadius: 8,
    minWidth: 100,
    padding: '8px 16px',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontSize: 12,
    fontWeight: 'bold',
    boxShadow: '0 0 20px rgba(34, 197, 94, 0.4), 0 4px 15px rgba(17, 153, 142, 0.3)',
    border: '1px solid rgba(255, 255, 255, 0.2)',
    cursor: 'pointer'
  },
  proxy: {
    background: 'linear-gradient(135deg, rgba(6, 182, 212, 0.15) 0%, rgba(8, 145, 178, 0.1) 100%)',
    border: '2px solid rgba(6, 182, 212, 0.5)',
    borderRadius: 12,
    padding: '8px 14px',
    minWidth: 120,
    boxShadow: '0 0 15px rgba(6, 182, 212, 0.2), inset 0 1px 0 rgba(255, 255, 255, 0.1)',
    backdropFilter: 'blur(8px)',
    cursor: 'pointer',
    position: 'relative'
  }
};

// 开始节点组件
function StartNode({ data }) {
  return (
    <div style={nodeStyles.start}>
      <PlayArrowIcon fontSize="small" sx={{ mr: 0.5 }} />
      <span>{data?.label || '入口'}</span>
      <Handle type="source" position={Position.Right} style={{ background: '#764ba2' }} />
    </div>
  );
}

// 结束节点组件（目标节点 - 可配置）
function EndNode({ data, selected }) {
  // 根据目标类型显示不同标签
  const getTargetLabel = () => {
    switch (data.targetType) {
      case 'all':
        return '所有节点';
      case 'specified_node':
        return data.nodeName || '指定节点';
      case 'conditions':
        return `${data.conditionCount || 0} 个条件`;
      default:
        return '所有节点';
    }
  };

  return (
    <div
      style={{
        ...nodeStyles.end,
        boxShadow: selected ? '0 4px 20px rgba(17, 153, 142, 0.6)' : nodeStyles.end.boxShadow
      }}
    >
      <Handle type="target" position={Position.Left} style={{ background: '#38ef7d' }} />
      <Stack direction="row" spacing={0.5} alignItems="center">
        <StopIcon fontSize="small" />
        <Box>
          <Typography variant="caption" sx={{ color: 'rgba(255,255,255,0.9)', fontSize: 10 }}>
            {getTargetLabel()}
          </Typography>
        </Box>
      </Stack>
    </div>
  );
}

// 代理组节点组件 - 支持悬停删除
function ProxyNode({ data, selected }) {
  const [hovered, setHovered] = useState(false);

  const getIcon = () => {
    const iconStyles = { fontSize: 18 };
    switch (data.proxyType) {
      case 'template_group':
        return <GroupWorkIcon sx={{ ...iconStyles, color: '#06b6d4' }} />;
      case 'custom_group':
        return <DeviceHubIcon sx={{ ...iconStyles, color: '#8b5cf6' }} />;
      case 'dynamic_node':
        return <FilterAltIcon sx={{ ...iconStyles, color: '#eab308' }} />;
      case 'specified_node':
        return <DeviceHubIcon sx={{ ...iconStyles, color: '#22c55e' }} />;
      default:
        return <GroupWorkIcon sx={{ ...iconStyles, color: '#06b6d4' }} />;
    }
  };

  const getTypeLabel = () => {
    const labels = {
      template_group: '模板组',
      custom_group: '自定义组',
      dynamic_node: '动态节点',
      specified_node: '指定节点'
    };
    return labels[data.proxyType] || '代理';
  };

  // 处理删除点击
  const handleDeleteClick = (e) => {
    e.stopPropagation();
    if (data.onDelete) {
      data.onDelete(data.nodeIndex);
    }
  };

  return (
    <div
      style={{
        ...nodeStyles.proxy,
        borderColor: selected ? '#06b6d4' : 'rgba(6, 182, 212, 0.5)',
        boxShadow: selected ? '0 0 25px rgba(6, 182, 212, 0.5), 0 8px 32px rgba(6, 182, 212, 0.2)' : nodeStyles.proxy.boxShadow
      }}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
    >
      <Handle type="target" position={Position.Left} style={{ background: '#06b6d4', width: 8, height: 8, border: '2px solid #1e3a5f' }} />

      {/* 删除按钮 - 悬停显示 */}
      {hovered && (
        <Tooltip title="删除此节点" arrow placement="top">
          <IconButton
            size="small"
            onClick={handleDeleteClick}
            sx={{
              position: 'absolute',
              top: -8,
              right: -8,
              width: 20,
              height: 20,
              background: 'linear-gradient(135deg, #ef4444 0%, #dc2626 100%)',
              border: '2px solid rgba(255, 255, 255, 0.3)',
              boxShadow: '0 2px 8px rgba(239, 68, 68, 0.4)',
              '&:hover': {
                background: 'linear-gradient(135deg, #f87171 0%, #ef4444 100%)',
                transform: 'scale(1.1)'
              },
              zIndex: 10
            }}
          >
            <CloseIcon sx={{ fontSize: 12, color: 'white' }} />
          </IconButton>
        </Tooltip>
      )}

      <Stack direction="row" spacing={0.5} alignItems="center">
        {getIcon()}
        <Box>
          <Typography variant="caption" sx={{ display: 'block', fontSize: 10, color: '#94a3b8' }}>
            {getTypeLabel()}
          </Typography>
          <Typography variant="body2" fontWeight="medium" sx={{ fontSize: 12, color: '#e2e8f0' }}>
            {data.label || '未配置'}
          </Typography>
        </Box>
      </Stack>
      <Handle type="source" position={Position.Right} style={{ background: '#06b6d4', width: 8, height: 8, border: '2px solid #1e3a5f' }} />
    </div>
  );
}

// 节点类型定义
const nodeTypes = {
  start: StartNode,
  end: EndNode,
  proxy: ProxyNode
};

// 默认边样式
const defaultEdgeOptions = {
  type: 'smoothstep',
  animated: true,
  style: { stroke: '#b1b1b7', strokeWidth: 2 },
  markerEnd: {
    type: MarkerType.ArrowClosed,
    color: '#b1b1b7'
  }
};

/**
 * 基于 React Flow 的画板式链式代理配置器
 * 配置面板在流程图右侧内联显示
 */
export default function ChainFlowBuilder({
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
  // 配置面板状态
  const [panelOpen, setPanelOpen] = useState(false);
  const [panelType, setPanelType] = useState(null); // 'proxy' | 'target'
  const [selectedNodeId, setSelectedNodeId] = useState(null);
  const [editingProxyConfig, setEditingProxyConfig] = useState(null);
  const [editingTargetConfig, setEditingTargetConfig] = useState(null);

  // 获取代理标签
  const getProxyLabel = useCallback(
    (item) => {
      if (!item) return '未配置';
      if (item.type === 'specified_node') {
        const node = availableNodes.find((n) => n.id === item.nodeId);
        return node?.name || node?.linkName || `节点 #${item.nodeId}`;
      }
      // 动态节点显示条件数量
      if (item.type === 'dynamic_node') {
        const condCount = item.nodeConditions?.conditions?.length || 0;
        if (condCount > 0) {
          return `配置${condCount}`;
        }
        return '未配置';
      }
      return item.groupName || '未配置';
    },
    [availableNodes]
  );

  // 直接删除代理节点（从节点悬停按钮）
  const handleDeleteProxyDirect = useCallback(
    (nodeIndex) => {
      const newChainConfig = chainConfig.filter((_, i) => i !== nodeIndex);
      onChainConfigChange?.(newChainConfig);
      // 如果删除的是当前打开面板的节点，关闭面板
      if (selectedNodeId === `proxy-${nodeIndex}`) {
        setPanelOpen(false);
      }
    },
    [chainConfig, onChainConfigChange, selectedNodeId]
  );

  // 构建流程节点
  const flowNodes = useMemo(() => {
    const nodes = [
      {
        id: 'start',
        type: 'start',
        position: { x: 30, y: 80 },
        data: { label: '入口' },
        draggable: false
      }
    ];

    // 添加代理节点
    chainConfig.forEach((item, index) => {
      // 使用保存的位置或计算默认位置，节点间距加大到200px
      const defaultX = 150 + index * 200;
      const defaultY = 100;
      nodes.push({
        id: `proxy-${index}`,
        type: 'proxy',
        position: item.position || { x: defaultX, y: defaultY },
        data: {
          label: getProxyLabel(item),
          proxyType: item.type,
          config: item,
          nodeIndex: index,
          onDelete: handleDeleteProxyDirect
        },
        draggable: true // 代理节点允许拖拽
      });
    });

    // 计算结束节点位置，如果有保存的位置则使用
    const endX = chainConfig.length > 0 ? 150 + chainConfig.length * 200 : 200;
    const conditionCount = targetConfig?.conditions?.conditions?.length || 0;

    // 获取指定节点的名称
    let nodeName = '';
    if (targetConfig?.type === 'specified_node' && targetConfig?.nodeId) {
      const targetNode = availableNodes.find((n) => n.id === targetConfig.nodeId);
      nodeName = targetNode?.name || targetNode?.linkName || `节点 #${targetConfig.nodeId}`;
    }

    nodes.push({
      id: 'end',
      type: 'end',
      position: targetConfig?.endPosition || { x: endX, y: 100 },
      data: {
        label: '目标节点',
        targetType: targetConfig?.type || 'specified_node',
        conditionCount,
        nodeName
      },
      draggable: true // 结束节点也允许拖拽
    });

    return nodes;
  }, [chainConfig, targetConfig, getProxyLabel, availableNodes, handleDeleteProxyDirect]);

  // 构建边
  const flowEdges = useMemo(() => {
    const edges = [];

    if (chainConfig.length === 0) {
      edges.push({
        id: 'start-end',
        source: 'start',
        target: 'end',
        ...defaultEdgeOptions
      });
    } else {
      edges.push({
        id: 'start-proxy-0',
        source: 'start',
        target: 'proxy-0',
        ...defaultEdgeOptions
      });

      for (let i = 0; i < chainConfig.length - 1; i++) {
        edges.push({
          id: `proxy-${i}-proxy-${i + 1}`,
          source: `proxy-${i}`,
          target: `proxy-${i + 1}`,
          ...defaultEdgeOptions
        });
      }

      edges.push({
        id: `proxy-${chainConfig.length - 1}-end`,
        source: `proxy-${chainConfig.length - 1}`,
        target: 'end',
        ...defaultEdgeOptions
      });
    }

    return edges;
  }, [chainConfig]);

  const [nodes, setNodes, onNodesChange] = useNodesState(flowNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(flowEdges);

  // 同步外部配置变化到内部节点
  const prevChainConfigRef = useRef(chainConfig);
  const prevTargetConfigRef = useRef(targetConfig);

  if (
    JSON.stringify(prevChainConfigRef.current) !== JSON.stringify(chainConfig) ||
    JSON.stringify(prevTargetConfigRef.current) !== JSON.stringify(targetConfig)
  ) {
    prevChainConfigRef.current = chainConfig;
    prevTargetConfigRef.current = targetConfig;
    setTimeout(() => {
      setNodes(flowNodes);
      setEdges(flowEdges);
    }, 0);
  }

  // 添加代理节点
  const handleAddProxy = useCallback(() => {
    const isEntryNode = chainConfig.length === 0;
    // 入口节点默认使用模板代理组，中间节点默认使用自定义代理组
    const defaultType = isEntryNode ? 'template_group' : 'custom_group';
    const newConfig = { type: defaultType, groupName: '' };
    const newChainConfig = [...chainConfig, newConfig];
    onChainConfigChange?.(newChainConfig);

    // 打开配置面板
    setSelectedNodeId(`proxy-${chainConfig.length}`);
    setEditingProxyConfig(newConfig);
    setPanelType('proxy');
    setPanelOpen(true);
  }, [chainConfig, onChainConfigChange]);

  // 删除代理节点（从面板）
  const handleDeleteProxy = useCallback(() => {
    if (!selectedNodeId || !selectedNodeId.startsWith('proxy-')) return;
    const nodeIndex = parseInt(selectedNodeId.replace('proxy-', ''), 10);
    const newChainConfig = chainConfig.filter((_, i) => i !== nodeIndex);
    onChainConfigChange?.(newChainConfig);
    setPanelOpen(false);
  }, [selectedNodeId, chainConfig, onChainConfigChange]);

  // 节点点击
  const onNodeClick = useCallback(
    (event, node) => {
      if (node.type === 'proxy') {
        const nodeIndex = parseInt(node.id.replace('proxy-', ''), 10);
        const config = { ...chainConfig[nodeIndex] };
        // 中间节点（索引 > 0）不支持模板代理组，自动修正为自定义代理组
        if (nodeIndex > 0 && config.type === 'template_group') {
          config.type = 'custom_group';
        }
        setSelectedNodeId(node.id);
        setEditingProxyConfig(config);
        setPanelType('proxy');
        setPanelOpen(true);
      } else if (node.type === 'end') {
        setSelectedNodeId('end');
        setEditingTargetConfig({ ...targetConfig });
        setPanelType('target');
        setPanelOpen(true);
      }
    },
    [chainConfig, targetConfig]
  );

  // 保存代理配置（不关闭面板）
  const saveProxyConfig = useCallback(() => {
    if (!selectedNodeId || !editingProxyConfig) return;
    const nodeIndex = parseInt(selectedNodeId.replace('proxy-', ''), 10);
    const newChainConfig = [...chainConfig];
    newChainConfig[nodeIndex] = editingProxyConfig;
    onChainConfigChange?.(newChainConfig);
  }, [selectedNodeId, editingProxyConfig, chainConfig, onChainConfigChange]);

  // 保存目标配置（不关闭面板）
  const saveTargetConfig = useCallback(() => {
    if (!editingTargetConfig) return;
    onTargetConfigChange?.(editingTargetConfig);
  }, [editingTargetConfig, onTargetConfigChange]);

  // 实时保存 - 代理配置变化时自动保存
  useEffect(() => {
    if (panelOpen && panelType === 'proxy' && editingProxyConfig) {
      const timer = setTimeout(() => {
        saveProxyConfig();
      }, 300); // 300ms防抖
      return () => clearTimeout(timer);
    }
  }, [panelOpen, panelType, editingProxyConfig, saveProxyConfig]);

  // 实时保存 - 目标配置变化时自动保存
  useEffect(() => {
    if (panelOpen && panelType === 'target' && editingTargetConfig) {
      const timer = setTimeout(() => {
        saveTargetConfig();
      }, 300); // 300ms防抖
      return () => clearTimeout(timer);
    }
  }, [panelOpen, panelType, editingTargetConfig, saveTargetConfig]);

  // 渲染代理配置面板
  const renderProxyConfigPanel = () => {
    if (!editingProxyConfig) return null;

    // 计算当前编辑节点的索引
    const nodeIndex = selectedNodeId ? parseInt(selectedNodeId.replace('proxy-', ''), 10) : 0;
    // 入口节点（索引0）可选所有类型，后续中间节点只能选择指定节点或动态条件节点
    const isEntryNode = nodeIndex === 0;

    return (
      <Stack spacing={2} sx={{ pt: 0.5 }}>
        <Stack direction="row" alignItems="center" justifyContent="space-between">
          <Typography variant="subtitle1" fontWeight="bold" sx={{ color: '#f1f5f9' }}>
            {isEntryNode ? '入口代理配置' : '中间节点配置'}
          </Typography>
          <IconButton size="small" onClick={() => setPanelOpen(false)} sx={{ color: '#94a3b8' }}>
            <CloseIcon fontSize="small" />
          </IconButton>
        </Stack>

        <FormControl size="small" fullWidth>
          <InputLabel sx={{ color: '#94a3b8' }}>代理类型</InputLabel>
          <Select
            value={editingProxyConfig.type || (isEntryNode ? 'template_group' : 'specified_node')}
            label="代理类型"
            onChange={(e) =>
              setEditingProxyConfig({
                type: e.target.value,
                groupName: '',
                nodeId: undefined,
                nodeConditions: undefined
              })
            }
            sx={{
              color: '#e2e8f0',
              '& .MuiOutlinedInput-notchedOutline': { borderColor: 'rgba(59, 130, 246, 0.3)' },
              '&:hover .MuiOutlinedInput-notchedOutline': { borderColor: 'rgba(59, 130, 246, 0.5)' },
              '&.Mui-focused .MuiOutlinedInput-notchedOutline': { borderColor: '#3b82f6' },
              '& .MuiSelect-icon': { color: '#64748b' }
            }}
          >
            {/* 入口节点可选模板代理组 */}
            {isEntryNode && <MenuItem value="template_group">模板代理组</MenuItem>}
            {/* 所有节点都可选自定义代理组（中间节点的组内节点会自动设置 dialer-proxy） */}
            <MenuItem value="custom_group">自定义代理组</MenuItem>
            <MenuItem value="dynamic_node">动态条件节点</MenuItem>
            <MenuItem value="specified_node">指定节点</MenuItem>
          </Select>
          {!isEntryNode && (
            <Typography variant="caption" sx={{ mt: 0.5, color: '#64748b' }}>
              中间节点的自定义代理组内所有节点会自动设置 dialer-proxy 指向上一级
            </Typography>
          )}
        </FormControl>

        {editingProxyConfig.type === 'template_group' && (
          <Autocomplete
            freeSolo
            size="small"
            fullWidth
            options={templateGroups || []}
            value={editingProxyConfig.groupName || ''}
            onChange={(e, newValue) => setEditingProxyConfig({ ...editingProxyConfig, groupName: newValue || '' })}
            onInputChange={(e, newValue) => setEditingProxyConfig({ ...editingProxyConfig, groupName: newValue || '' })}
            renderInput={(params) => (
              <TextField {...params} label="代理组名称" placeholder="选择或输入代理组名称" helperText="从模板中选择或手动输入代理组名称" />
            )}
          />
        )}

        {editingProxyConfig.type === 'custom_group' && (
          <>
            <TextField
              size="small"
              fullWidth
              label="代理组名称"
              placeholder="自定义代理组名称"
              value={editingProxyConfig.groupName || ''}
              onChange={(e) => setEditingProxyConfig({ ...editingProxyConfig, groupName: e.target.value })}
            />
            <FormControl size="small" fullWidth>
              <InputLabel>组类型</InputLabel>
              <Select
                value={editingProxyConfig.groupType || 'select'}
                label="组类型"
                onChange={(e) => setEditingProxyConfig({ ...editingProxyConfig, groupType: e.target.value })}
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
              value={editingProxyConfig.nodeConditions}
              onChange={(conds) => setEditingProxyConfig({ ...editingProxyConfig, nodeConditions: conds })}
              fields={fields}
              operators={operators}
            />
          </>
        )}

        {editingProxyConfig.type === 'dynamic_node' && (
          <>
            <FormControl size="small" fullWidth>
              <InputLabel>选择模式</InputLabel>
              <Select
                value={editingProxyConfig.selectMode || 'first'}
                label="选择模式"
                onChange={(e) => setEditingProxyConfig({ ...editingProxyConfig, selectMode: e.target.value })}
              >
                <MenuItem value="first">第一个匹配</MenuItem>
                <MenuItem value="random">随机</MenuItem>
                <MenuItem value="fastest">最快节点</MenuItem>
              </Select>
            </FormControl>
            <ConditionBuilder
              title="节点匹配条件"
              value={editingProxyConfig.nodeConditions}
              onChange={(conds) => setEditingProxyConfig({ ...editingProxyConfig, nodeConditions: conds })}
              fields={fields}
              operators={operators}
            />
          </>
        )}

        {editingProxyConfig.type === 'specified_node' && (
          <Autocomplete
            size="small"
            options={availableNodes || []}
            getOptionLabel={(option) => `${option.name || option.linkName} (${option.linkCountry || '未知'})`}
            value={(availableNodes || []).find((n) => n.id === editingProxyConfig.nodeId) || null}
            onChange={(e, newValue) => setEditingProxyConfig({ ...editingProxyConfig, nodeId: newValue?.id })}
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

        <Divider sx={{ borderColor: 'rgba(59, 130, 246, 0.2)' }} />

        {/* 删除按钮 - 实时保存无需确定按钮 */}
        <Stack direction="row" justifyContent="flex-start">
          <Button
            size="small"
            color="error"
            startIcon={<DeleteIcon />}
            onClick={handleDeleteProxy}
            sx={{
              color: '#f87171',
              '&:hover': { bgcolor: 'rgba(248, 113, 113, 0.1)' }
            }}
          >
            删除此节点
          </Button>
        </Stack>
      </Stack>
    );
  };

  // 渲染目标节点配置面板
  const renderTargetConfigPanel = () => {
    if (!editingTargetConfig) return null;

    return (
      <Stack spacing={2} sx={{ pt: 0.5 }}>
        <Stack direction="row" alignItems="center" justifyContent="space-between">
          <Typography variant="subtitle1" fontWeight="bold" sx={{ color: '#f1f5f9' }}>
            目标节点配置
          </Typography>
          <IconButton size="small" onClick={() => setPanelOpen(false)} sx={{ color: '#94a3b8' }}>
            <CloseIcon fontSize="small" />
          </IconButton>
        </Stack>

        <Typography variant="body2" sx={{ color: '#64748b' }}>
          选择应用此规则的节点范围
        </Typography>

        <ToggleButtonGroup
          value={editingTargetConfig.type || 'specified_node'}
          exclusive
          onChange={(e, newType) => {
            if (newType !== null) {
              setEditingTargetConfig({ ...editingTargetConfig, type: newType, nodeId: undefined, conditions: undefined });
            }
          }}
          size="small"
          fullWidth
          sx={{
            '& .MuiToggleButton-root': {
              color: '#94a3b8',
              borderColor: 'rgba(59, 130, 246, 0.3)',
              '&.Mui-selected': {
                color: '#3b82f6',
                bgcolor: 'rgba(59, 130, 246, 0.15)',
                borderColor: 'rgba(59, 130, 246, 0.5)'
              },
              '&:hover': {
                bgcolor: 'rgba(59, 130, 246, 0.1)'
              }
            }
          }}
        >
          <Tooltip title="手动指定唯一一个目标节点" arrow>
            <ToggleButton value="specified_node">指定节点</ToggleButton>
          </Tooltip>
          <Tooltip title="规则应用于所有节点" arrow>
            <ToggleButton value="all">所有节点</ToggleButton>
          </Tooltip>
          <Tooltip title="根据条件筛选节点" arrow>
            <ToggleButton value="conditions">按条件</ToggleButton>
          </Tooltip>
        </ToggleButtonGroup>

        {/* 指定节点选择 */}
        {editingTargetConfig.type === 'specified_node' && (
          <Autocomplete
            size="small"
            options={availableNodes || []}
            getOptionLabel={(option) => `${option.name || option.linkName} (${option.linkCountry || '未知'})`}
            value={(availableNodes || []).find((n) => n.id === editingTargetConfig.nodeId) || null}
            onChange={(e, newValue) => setEditingTargetConfig({ ...editingTargetConfig, nodeId: newValue?.id })}
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

        {/* 条件筛选 */}
        {editingTargetConfig.type === 'conditions' && (
          <ConditionBuilder
            title="目标节点筛选条件"
            value={editingTargetConfig.conditions}
            onChange={(conds) => setEditingTargetConfig({ ...editingTargetConfig, conditions: conds })}
            fields={fields}
            operators={operators}
          />
        )}

        <Divider sx={{ borderColor: 'rgba(59, 130, 246, 0.2)' }} />

        {/* 实时保存，无需确定按钮 */}
        <Typography variant="caption" sx={{ color: '#64748b', textAlign: 'center' }}>
          配置已自动保存
        </Typography>
      </Stack>
    );
  };

  // 处理节点拖拽结束事件，保存位置
  const onNodeDragStop = useCallback(
    (event, node) => {
      if (node.id.startsWith('proxy-')) {
        const nodeIndex = parseInt(node.id.replace('proxy-', ''), 10);
        const newChainConfig = [...chainConfig];
        if (newChainConfig[nodeIndex]) {
          newChainConfig[nodeIndex] = {
            ...newChainConfig[nodeIndex],
            position: node.position
          };
          onChainConfigChange?.(newChainConfig);
        }
      } else if (node.id === 'end') {
        // 保存结束节点位置
        onTargetConfigChange?.({
          ...targetConfig,
          endPosition: node.position
        });
      }
    },
    [chainConfig, targetConfig, onChainConfigChange, onTargetConfigChange]
  );

  return (
    <Box className="chain-flow-container" sx={{ height: 450, width: '100%', display: 'flex', overflow: 'hidden' }}>
      {/* 流程图区域 */}
      <Box sx={{ flex: 1, position: 'relative', minWidth: 0 }}>
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onNodeClick={onNodeClick}
          onNodeDragStop={onNodeDragStop}
          nodeTypes={nodeTypes}
          defaultEdgeOptions={defaultEdgeOptions}
          fitView
          fitViewOptions={{ padding: 0.3, minZoom: 0.5, maxZoom: 1.2 }}
          minZoom={0.3}
          maxZoom={1.5}
          proOptions={{ hideAttribution: true }}
          nodesConnectable={false}
          elementsSelectable={true}
        >
          <Controls showInteractive={false} />
        </ReactFlow>

        {/* 添加代理按钮 */}
        <Box className="chain-flow-toolbar">
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={handleAddProxy}
            size="small"
            disabled={chainConfig.length >= 4}
            sx={{
              background: 'rgba(15, 23, 42, 0.9)',
              border: '1px solid rgba(59, 130, 246, 0.4)',
              backdropFilter: 'blur(8px)',
              '&:hover': { background: 'rgba(59, 130, 246, 0.3)' },
              '&.Mui-disabled': { color: '#64748b', borderColor: 'rgba(100, 116, 139, 0.3)' }
            }}
          >
            {chainConfig.length >= 4 ? '已达最大层级' : '添加代理节点'}
          </Button>
          {chainConfig.length >= 2 && (
            <Typography variant="caption" sx={{ ml: 1, color: '#f87171' }}>
              {chainConfig.length} 级链路，延迟可能较高
            </Typography>
          )}
        </Box>

        {/* 提示文字 */}
        <Box className="chain-flow-hint">
          <Typography variant="caption" sx={{ color: '#64748b' }}>
            点击节点进行配置，悬停节点可快速删除
          </Typography>
        </Box>
      </Box>

      {/* 配置面板 - 深色玻璃效果 */}
      {panelOpen && (
        <Fade in={panelOpen}>
          <Paper
            className="chain-flow-panel"
            elevation={0}
            sx={{
              width: 480,
              minWidth: 480,
              borderLeft: '1px solid rgba(59, 130, 246, 0.3)',
              borderRadius: 0,
              p: 2.5,
              overflow: 'auto',
              background: 'linear-gradient(135deg, rgba(15, 23, 42, 0.98) 0%, rgba(30, 41, 59, 0.98) 100%)',
              backdropFilter: 'blur(20px)',
              boxShadow: '-4px 0 20px rgba(0, 0, 0, 0.3)'
            }}
          >
            {panelType === 'proxy' ? renderProxyConfigPanel() : renderTargetConfigPanel()}
          </Paper>
        </Fade>
      )}
    </Box>
  );
}
