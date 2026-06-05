import { useState, useCallback, useMemo, useRef, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { alpha, useTheme } from '@mui/material/styles';
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
import {
  getChainProxyFieldControlSx,
  getChainProxyIconButtonSx,
  getChainProxyThemeTokens,
  getChainProxyToggleButtonGroupSx
} from './chainProxyTheme';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { withAlpha } from '../../../utils/colorUtils';

const getProxyTypeColor = (theme, proxyType) => {
  switch (proxyType) {
    case 'template_group':
      return theme.palette.info.main;
    case 'custom_group':
      return theme.palette.secondary.main;
    case 'dynamic_node':
      return theme.palette.warning.main;
    case 'specified_node':
      return theme.palette.success.main;
    default:
      return theme.palette.info.main;
  }
};

const getNodeStyles = (theme, tokens) => {
  const { mutedPanelSurface, nestedPanelSurface, cardShadow, subtleBorder } = tokens;

  return {
    start: {
      background: mutedPanelSurface,
      color: theme.palette.secondary.main,
      borderRadius: 30,
      minWidth: 90,
      padding: '0 16px',
      height: 40,
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      fontSize: 14,
      fontWeight: 'bold',
      boxShadow: cardShadow,
      border: `1px solid ${alpha(theme.palette.secondary.main, 0.28)}`
    },
    end: {
      background: nestedPanelSurface,
      color: theme.palette.success.main,
      borderRadius: 8,
      minWidth: 100,
      padding: '8px 16px',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      fontSize: 12,
      fontWeight: 'bold',
      boxShadow: cardShadow,
      border: `1px solid ${alpha(theme.palette.success.main, 0.28)}`,
      cursor: 'pointer'
    },
    proxy: {
      background: nestedPanelSurface,
      border: `1px solid ${subtleBorder}`,
      borderRadius: 12,
      padding: '8px 14px',
      minWidth: 120,
      boxShadow: cardShadow,
      cursor: 'pointer',
      position: 'relative'
    }
  };
};

function StartNode({ data }) {
  const theme = useTheme();
  const { t } = useTranslation();
  const { isDark } = useResolvedColorScheme();
  const tokens = getChainProxyThemeTokens(theme, isDark);
  const nodeStyles = getNodeStyles(theme, tokens);
  return (
    <div style={nodeStyles.start}>
      <PlayArrowIcon fontSize="small" sx={{ mr: 0.5 }} />
      <span>{data?.label || t('subscriptions.chain.entry')}</span>
      <Handle
        type="source"
        position={Position.Right}
        style={{ background: theme.palette.secondary.main, border: `2px solid ${nodeStyles.start.background}` }}
      />
    </div>
  );
}

function EndNode({ data, selected }) {
  const theme = useTheme();
  const { t } = useTranslation();
  const { isDark } = useResolvedColorScheme();
  const tokens = getChainProxyThemeTokens(theme, isDark);
  const nodeStyles = getNodeStyles(theme, tokens);
  const getTargetLabel = () => {
    switch (data.targetType) {
      case 'all':
        return t('subscriptions.chain.targetTypes.all');
      case 'specified_node':
        return data.nodeName || t('subscriptions.chain.targetTypes.specified_node');
      case 'conditions':
        return t('subscriptions.chain.conditionCount', { count: data.conditionCount || 0 });
      default:
        return t('subscriptions.chain.targetTypes.all');
    }
  };

  return (
    <div
      style={{
        ...nodeStyles.end,
        boxShadow: selected
          ? isDark
            ? `0 0 0 1px ${alpha(theme.palette.success.main, 0.32)}`
            : `0 4px 16px ${alpha(theme.palette.success.main, 0.22)}`
          : nodeStyles.end.boxShadow
      }}
    >
      <Handle
        type="target"
        position={Position.Left}
        style={{ background: theme.palette.success.main, border: `2px solid ${nodeStyles.end.background}` }}
      />
      <Stack direction="row" spacing={0.5} alignItems="center">
        <StopIcon fontSize="small" />
        <Box>
          <Typography variant="caption" sx={{ color: 'text.secondary', fontSize: 10 }}>
            {getTargetLabel()}
          </Typography>
        </Box>
      </Stack>
    </div>
  );
}

function ProxyNode({ data, selected }) {
  const theme = useTheme();
  const { t } = useTranslation();
  const { isDark } = useResolvedColorScheme();
  const tokens = getChainProxyThemeTokens(theme, isDark);
  const nodeStyles = getNodeStyles(theme, tokens);
  const [hovered, setHovered] = useState(false);

  const getIcon = () => {
    const iconStyles = { fontSize: 18 };
    const color = getProxyTypeColor(theme, data.proxyType);
    switch (data.proxyType) {
      case 'template_group':
        return <GroupWorkIcon sx={{ ...iconStyles, color }} />;
      case 'custom_group':
        return <DeviceHubIcon sx={{ ...iconStyles, color }} />;
      case 'dynamic_node':
        return <FilterAltIcon sx={{ ...iconStyles, color }} />;
      case 'specified_node':
        return <DeviceHubIcon sx={{ ...iconStyles, color }} />;
      default:
        return <GroupWorkIcon sx={{ ...iconStyles, color }} />;
    }
  };

  const getTypeLabel = () => {
    const labels = t('subscriptions.chain.proxyTypeShort', { returnObjects: true });
    return labels[data.proxyType] || t('subscriptions.chain.proxy');
  };

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
        borderColor: selected ? theme.palette.info.main : alpha(theme.palette.info.main, 0.42),
        boxShadow: selected
          ? isDark
            ? `0 0 0 1px ${alpha(theme.palette.info.main, 0.28)}`
            : `0 6px 18px ${alpha(theme.palette.info.main, 0.16)}`
          : nodeStyles.proxy.boxShadow
      }}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
    >
      <Handle
        type="target"
        position={Position.Left}
        style={{ background: theme.palette.info.main, width: 8, height: 8, border: `2px solid ${nodeStyles.proxy.background}` }}
      />

      {hovered && (
        <Tooltip title={t('subscriptions.chain.deleteNode')} arrow placement="top">
          <IconButton
            size="small"
            onClick={handleDeleteClick}
            sx={{
              position: 'absolute',
              top: -8,
              right: -8,
              width: 20,
              height: 20,
              background: theme.palette.error.main,
              border: `1px solid ${alpha(theme.palette.error.main, 0.24)}`,
              boxShadow: tokens.insetHighlight,
              '&:hover': {
                background: withAlpha(theme.palette.error.main, 0.9),
                transform: 'scale(1.1)'
              },
              zIndex: 10
            }}
          >
            <CloseIcon sx={{ fontSize: 12, color: 'error.contrastText' }} />
          </IconButton>
        </Tooltip>
      )}

      <Stack direction="row" spacing={0.5} alignItems="center">
        {getIcon()}
        <Box>
          <Typography variant="caption" sx={{ display: 'block', fontSize: 10, color: 'text.secondary' }}>
            {getTypeLabel()}
          </Typography>
          <Typography variant="body2" fontWeight="medium" sx={{ fontSize: 12, color: 'text.primary' }}>
            {data.label || t('subscriptions.chain.unconfigured')}
          </Typography>
        </Box>
      </Stack>
      <Handle
        type="source"
        position={Position.Right}
        style={{ background: theme.palette.info.main, width: 8, height: 8, border: `2px solid ${nodeStyles.proxy.background}` }}
      />
    </div>
  );
}

const nodeTypes = {
  start: StartNode,
  end: EndNode,
  proxy: ProxyNode
};

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
  const theme = useTheme();
  const { t } = useTranslation();
  const { isDark } = useResolvedColorScheme();
  const tokens = getChainProxyThemeTokens(theme, isDark);
  const palette = tokens.palette;
  const fieldControlSx = getChainProxyFieldControlSx(tokens);
  const toggleButtonSx = getChainProxyToggleButtonGroupSx(tokens);
  const iconButtonSx = getChainProxyIconButtonSx(tokens);
  const defaultEdgeOptions = useMemo(
    () => ({
      type: 'smoothstep',
      animated: true,
      style: { stroke: withAlpha(palette.text.secondary, isDark ? 0.76 : 0.72), strokeWidth: 2 },
      markerEnd: {
        type: MarkerType.ArrowClosed,
        color: withAlpha(palette.text.secondary, isDark ? 0.76 : 0.72)
      }
    }),
    [isDark, palette.text.secondary]
  );
  const [panelOpen, setPanelOpen] = useState(false);
  const [panelType, setPanelType] = useState(null); // 'proxy' | 'target'
  const [selectedNodeId, setSelectedNodeId] = useState(null);
  const [editingProxyConfig, setEditingProxyConfig] = useState(null);
  const [editingTargetConfig, setEditingTargetConfig] = useState(null);

  const getProxyLabel = useCallback(
    (item) => {
      if (!item) return t('subscriptions.chain.unconfigured');
      if (item.type === 'specified_node') {
        const node = availableNodes.find((n) => n.id === item.nodeId);
        return node?.name || node?.linkName || t('subscriptions.chain.nodeNumber', { id: item.nodeId });
      }
      if (item.type === 'dynamic_node') {
        const condCount = item.nodeConditions?.conditions?.length || 0;
        if (condCount > 0) {
          return t('subscriptions.chain.configuredConditions', { count: condCount });
        }
        return t('subscriptions.chain.unconfigured');
      }
      return item.groupName || t('subscriptions.chain.unconfigured');
    },
    [availableNodes, t]
  );

  const handleDeleteProxyDirect = useCallback(
    (nodeIndex) => {
      const newChainConfig = chainConfig.filter((_, i) => i !== nodeIndex);
      onChainConfigChange?.(newChainConfig);
      if (selectedNodeId === `proxy-${nodeIndex}`) {
        setPanelOpen(false);
      }
    },
    [chainConfig, onChainConfigChange, selectedNodeId]
  );

  const flowNodes = useMemo(() => {
    const nodes = [
      {
        id: 'start',
        type: 'start',
        position: { x: 30, y: 80 },
        data: { label: t('subscriptions.chain.entry') },
        draggable: false
      }
    ];

    chainConfig.forEach((item, index) => {
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
        draggable: true
      });
    });

    const endX = chainConfig.length > 0 ? 150 + chainConfig.length * 200 : 200;
    const conditionCount = targetConfig?.conditions?.conditions?.length || 0;

    let nodeName = '';
    if (targetConfig?.type === 'specified_node' && targetConfig?.nodeId) {
      const targetNode = availableNodes.find((n) => n.id === targetConfig.nodeId);
      nodeName = targetNode?.name || targetNode?.linkName || t('subscriptions.chain.nodeNumber', { id: targetConfig.nodeId });
    }

    nodes.push({
      id: 'end',
      type: 'end',
      position: targetConfig?.endPosition || { x: endX, y: 100 },
      data: {
        label: t('subscriptions.chain.targetNode'),
        targetType: targetConfig?.type || 'specified_node',
        conditionCount,
        nodeName
      },
      draggable: true
    });

    return nodes;
  }, [chainConfig, targetConfig, getProxyLabel, availableNodes, handleDeleteProxyDirect, t]);

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
  }, [chainConfig, defaultEdgeOptions]);

  const [nodes, setNodes, onNodesChange] = useNodesState(flowNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(flowEdges);

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

  const handleAddProxy = useCallback(() => {
    const isEntryNode = chainConfig.length === 0;
    const defaultType = isEntryNode ? 'template_group' : 'custom_group';
    const newConfig = { type: defaultType, groupName: '' };
    const newChainConfig = [...chainConfig, newConfig];
    onChainConfigChange?.(newChainConfig);

    setSelectedNodeId(`proxy-${chainConfig.length}`);
    setEditingProxyConfig(newConfig);
    setPanelType('proxy');
    setPanelOpen(true);
  }, [chainConfig, onChainConfigChange]);

  const handleDeleteProxy = useCallback(() => {
    if (!selectedNodeId || !selectedNodeId.startsWith('proxy-')) return;
    const nodeIndex = parseInt(selectedNodeId.replace('proxy-', ''), 10);
    const newChainConfig = chainConfig.filter((_, i) => i !== nodeIndex);
    onChainConfigChange?.(newChainConfig);
    setPanelOpen(false);
  }, [selectedNodeId, chainConfig, onChainConfigChange]);

  const onNodeClick = useCallback(
    (_event, node) => {
      if (node.type === 'proxy') {
        const nodeIndex = parseInt(node.id.replace('proxy-', ''), 10);
        const config = { ...chainConfig[nodeIndex] };
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

  const saveProxyConfig = useCallback(() => {
    if (!selectedNodeId || !editingProxyConfig) return;
    const nodeIndex = parseInt(selectedNodeId.replace('proxy-', ''), 10);
    const newChainConfig = [...chainConfig];
    newChainConfig[nodeIndex] = editingProxyConfig;
    onChainConfigChange?.(newChainConfig);
  }, [selectedNodeId, editingProxyConfig, chainConfig, onChainConfigChange]);

  const saveTargetConfig = useCallback(() => {
    if (!editingTargetConfig) return;
    onTargetConfigChange?.(editingTargetConfig);
  }, [editingTargetConfig, onTargetConfigChange]);

  useEffect(() => {
    if (panelOpen && panelType === 'proxy' && editingProxyConfig) {
      const timer = setTimeout(() => {
        saveProxyConfig();
      }, 300);
      return () => clearTimeout(timer);
    }
  }, [panelOpen, panelType, editingProxyConfig, saveProxyConfig]);

  useEffect(() => {
    if (panelOpen && panelType === 'target' && editingTargetConfig) {
      const timer = setTimeout(() => {
        saveTargetConfig();
      }, 300);
      return () => clearTimeout(timer);
    }
  }, [panelOpen, panelType, editingTargetConfig, saveTargetConfig]);

  const renderProxyConfigPanel = () => {
    if (!editingProxyConfig) return null;

    const nodeIndex = selectedNodeId ? parseInt(selectedNodeId.replace('proxy-', ''), 10) : 0;
    const isEntryNode = nodeIndex === 0;

    return (
      <Stack spacing={2} sx={{ pt: 0.5 }}>
        <Stack direction="row" alignItems="center" justifyContent="space-between">
          <Typography variant="subtitle1" fontWeight="bold" color={tokens.primaryText}>
            {isEntryNode ? t('subscriptions.chain.entryProxyConfig') : t('subscriptions.chain.middleNodeConfig')}
          </Typography>
          <IconButton size="small" onClick={() => setPanelOpen(false)} sx={iconButtonSx}>
            <CloseIcon fontSize="small" />
          </IconButton>
        </Stack>

        <FormControl size="small" fullWidth sx={fieldControlSx}>
          <InputLabel color="primary">{t('subscriptions.chain.proxyType')}</InputLabel>
          <Select
            value={editingProxyConfig.type || (isEntryNode ? 'template_group' : 'specified_node')}
            label={t('subscriptions.chain.proxyType')}
            onChange={(e) =>
              setEditingProxyConfig({
                type: e.target.value,
                groupName: '',
                nodeId: undefined,
                nodeConditions: undefined
              })
            }
          >
            {isEntryNode && <MenuItem value="template_group">{t('subscriptions.chain.proxyTypes.template_group')}</MenuItem>}
            <MenuItem value="custom_group">{t('subscriptions.chain.proxyTypes.custom_group')}</MenuItem>
            <MenuItem value="dynamic_node">{t('subscriptions.chain.proxyTypes.dynamic_node')}</MenuItem>
            <MenuItem value="specified_node">{t('subscriptions.chain.proxyTypes.specified_node')}</MenuItem>
          </Select>
          {!isEntryNode && (
            <Typography variant="caption" sx={{ mt: 0.5, color: tokens.tertiaryText }}>
              {t('subscriptions.chain.middleNodeDialerProxyHint')}
            </Typography>
          )}
        </FormControl>

        {editingProxyConfig.type === 'template_group' && (
          <Autocomplete
            freeSolo
            size="small"
            fullWidth
            sx={fieldControlSx}
            options={templateGroups || []}
            value={editingProxyConfig.groupName || ''}
            onChange={(_event, newValue) => setEditingProxyConfig({ ...editingProxyConfig, groupName: newValue || '' })}
            onInputChange={(_event, newValue) => setEditingProxyConfig({ ...editingProxyConfig, groupName: newValue || '' })}
            renderInput={(params) => (
              <TextField
                {...params}
                label={t('subscriptions.chain.groupName')}
                placeholder={t('subscriptions.chain.selectOrInputGroupPlaceholder')}
                helperText={t('subscriptions.chain.selectOrInputGroupHelper')}
              />
            )}
          />
        )}

        {editingProxyConfig.type === 'custom_group' && (
          <>
            <TextField
              size="small"
              fullWidth
              label={t('subscriptions.chain.groupName')}
              placeholder={t('subscriptions.chain.customGroupPlaceholder')}
              value={editingProxyConfig.groupName || ''}
              onChange={(e) => setEditingProxyConfig({ ...editingProxyConfig, groupName: e.target.value })}
              sx={fieldControlSx}
            />
            <FormControl size="small" fullWidth sx={fieldControlSx}>
              <InputLabel>{t('subscriptions.chain.groupType')}</InputLabel>
              <Select
                value={editingProxyConfig.groupType || 'select'}
                label={t('subscriptions.chain.groupType')}
                onChange={(e) => setEditingProxyConfig({ ...editingProxyConfig, groupType: e.target.value })}
              >
                {(groupTypes || []).map((gt) => (
                  <MenuItem key={gt.value} value={gt.value}>
                    {gt.label}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
            {(editingProxyConfig.groupType === 'url-test' || editingProxyConfig.groupType === 'fallback') && (
              <Stack spacing={1.5}>
                <TextField
                  size="small"
                  fullWidth
                  label={t('subscriptions.chain.testUrl')}
                  value={editingProxyConfig.urlTestConfig?.url || ''}
                  onChange={(e) =>
                    setEditingProxyConfig({
                      ...editingProxyConfig,
                      urlTestConfig: { ...editingProxyConfig.urlTestConfig, url: e.target.value }
                    })
                  }
                  placeholder="http://www.gstatic.com/generate_204"
                  helperText={t('subscriptions.chain.testUrlHelper')}
                  sx={fieldControlSx}
                />
                <Stack direction="row" spacing={1}>
                  <TextField
                    size="small"
                    label={t('subscriptions.chain.intervalSeconds')}
                    type="number"
                    value={editingProxyConfig.urlTestConfig?.interval ?? 300}
                    onChange={(e) =>
                      setEditingProxyConfig({
                        ...editingProxyConfig,
                        urlTestConfig: { ...editingProxyConfig.urlTestConfig, interval: parseInt(e.target.value) || 300 }
                      })
                    }
                    sx={{ ...fieldControlSx, flex: 1 }}
                    helperText={t('subscriptions.chain.healthCheckInterval')}
                  />
                  <TextField
                    size="small"
                    label={t('subscriptions.chain.toleranceMs')}
                    type="number"
                    value={editingProxyConfig.urlTestConfig?.tolerance ?? 50}
                    onChange={(e) =>
                      setEditingProxyConfig({
                        ...editingProxyConfig,
                        urlTestConfig: { ...editingProxyConfig.urlTestConfig, tolerance: parseInt(e.target.value) || 50 }
                      })
                    }
                    sx={{ ...fieldControlSx, flex: 1 }}
                    helperText={
                      editingProxyConfig.groupType === 'url-test'
                        ? t('subscriptions.chain.toleranceHelper')
                        : t('subscriptions.chain.fallbackThreshold')
                    }
                  />
                </Stack>
              </Stack>
            )}
            {editingProxyConfig.groupType === 'load-balance' && (
              <Stack spacing={1.5}>
                <TextField
                  size="small"
                  fullWidth
                  label={t('subscriptions.chain.testUrl')}
                  value={editingProxyConfig.urlTestConfig?.url || ''}
                  onChange={(e) =>
                    setEditingProxyConfig({
                      ...editingProxyConfig,
                      urlTestConfig: { ...editingProxyConfig.urlTestConfig, url: e.target.value }
                    })
                  }
                  placeholder="http://www.gstatic.com/generate_204"
                  helperText={t('subscriptions.chain.testUrlHelper')}
                  sx={fieldControlSx}
                />
                <Stack direction="row" spacing={1}>
                  <TextField
                    size="small"
                    label={t('subscriptions.chain.intervalSeconds')}
                    type="number"
                    value={editingProxyConfig.urlTestConfig?.interval ?? 300}
                    onChange={(e) =>
                      setEditingProxyConfig({
                        ...editingProxyConfig,
                        urlTestConfig: { ...editingProxyConfig.urlTestConfig, interval: parseInt(e.target.value) || 300 }
                      })
                    }
                    sx={{ ...fieldControlSx, flex: 1 }}
                    helperText={t('subscriptions.chain.healthCheckInterval')}
                  />
                  <FormControl size="small" sx={{ ...fieldControlSx, flex: 1 }}>
                    <InputLabel>{t('subscriptions.chain.loadBalanceStrategy')}</InputLabel>
                    <Select
                      value={editingProxyConfig.urlTestConfig?.strategy || 'consistent-hashing'}
                      label={t('subscriptions.chain.loadBalanceStrategy')}
                      onChange={(e) =>
                        setEditingProxyConfig({
                          ...editingProxyConfig,
                          urlTestConfig: { ...editingProxyConfig.urlTestConfig, strategy: e.target.value }
                        })
                      }
                    >
                      <MenuItem value="consistent-hashing">{t('subscriptions.chain.strategies.consistentHashing')}</MenuItem>
                      <MenuItem value="round-robin">{t('subscriptions.chain.strategies.roundRobin')}</MenuItem>
                      <MenuItem value="sticky-sessions">{t('subscriptions.chain.strategies.stickySessions')}</MenuItem>
                    </Select>
                  </FormControl>
                </Stack>
              </Stack>
            )}
            <ConditionBuilder
              title={t('subscriptions.chain.nodeFilterConditions')}
              value={editingProxyConfig.nodeConditions}
              onChange={(conds) => setEditingProxyConfig({ ...editingProxyConfig, nodeConditions: conds })}
              fields={fields}
              operators={operators}
            />
          </>
        )}

        {editingProxyConfig.type === 'dynamic_node' && (
          <>
            <FormControl size="small" fullWidth sx={fieldControlSx}>
              <InputLabel>{t('subscriptions.chain.selectMode')}</InputLabel>
              <Select
                value={editingProxyConfig.selectMode || 'first'}
                label={t('subscriptions.chain.selectMode')}
                onChange={(e) => setEditingProxyConfig({ ...editingProxyConfig, selectMode: e.target.value })}
              >
                <MenuItem value="first">{t('subscriptions.chain.selectModes.first')}</MenuItem>
                <MenuItem value="random">{t('subscriptions.chain.selectModes.random')}</MenuItem>
                <MenuItem value="fastest">{t('subscriptions.chain.selectModes.fastest')}</MenuItem>
              </Select>
            </FormControl>
            <ConditionBuilder
              title={t('subscriptions.chain.nodeMatchConditions')}
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
            sx={fieldControlSx}
            getOptionLabel={(option) => `${option.name || option.linkName} (${option.linkCountry || t('common.unknown')})`}
            value={(availableNodes || []).find((n) => n.id === editingProxyConfig.nodeId) || null}
            onChange={(_event, newValue) => setEditingProxyConfig({ ...editingProxyConfig, nodeId: newValue?.id })}
            renderInput={(params) => <TextField {...params} label={t('subscriptions.chain.selectNode')} />}
            renderOption={(props, option) => (
              <li {...props} key={option.id}>
                <Stack direction="row" spacing={1} alignItems="center">
                  <Typography variant="body2">{option.name || option.linkName}</Typography>
                  <Chip label={option.linkCountry || t('common.unknown')} size="small" variant="outlined" />
                </Stack>
              </li>
            )}
          />
        )}

        <Divider sx={{ borderColor: tokens.softBorder }} />

        <Stack direction="row" justifyContent="flex-start">
          <Button
            size="small"
            color="error"
            startIcon={<DeleteIcon />}
            onClick={handleDeleteProxy}
            sx={{
              color: theme.palette.error.main,
              border: `1px solid ${withAlpha(theme.palette.error.main, isDark ? 0.32 : 0.2)}`,
              backgroundColor: tokens.fieldSurface,
              boxShadow: tokens.insetHighlight,
              '&:hover': {
                bgcolor: tokens.errorSurface,
                borderColor: withAlpha(theme.palette.error.main, isDark ? 0.42 : 0.28)
              }
            }}
            variant="outlined"
          >
            {t('subscriptions.chain.deleteNode')}
          </Button>
        </Stack>
      </Stack>
    );
  };

  const renderTargetConfigPanel = () => {
    if (!editingTargetConfig) return null;

    return (
      <Stack spacing={2} sx={{ pt: 0.5 }}>
        <Stack direction="row" alignItems="center" justifyContent="space-between">
          <Typography variant="subtitle1" fontWeight="bold" color={tokens.primaryText}>
            {t('subscriptions.chain.targetNodeConfig')}
          </Typography>
          <IconButton size="small" onClick={() => setPanelOpen(false)} sx={iconButtonSx}>
            <CloseIcon fontSize="small" />
          </IconButton>
        </Stack>

        <Typography variant="body2" color={tokens.secondaryText}>
          {t('subscriptions.chain.targetScopeDescription')}
        </Typography>

        <ToggleButtonGroup
          value={editingTargetConfig.type || 'specified_node'}
          exclusive
          onChange={(_event, newType) => {
            if (newType !== null) {
              setEditingTargetConfig({ ...editingTargetConfig, type: newType, nodeId: undefined, conditions: undefined });
            }
          }}
          size="small"
          fullWidth
          sx={toggleButtonSx}
        >
          <Tooltip title={t('subscriptions.chain.targetTooltips.specified')} arrow>
            <ToggleButton value="specified_node">{t('subscriptions.chain.targetTypes.specified_node')}</ToggleButton>
          </Tooltip>
          <Tooltip title={t('subscriptions.chain.targetTooltips.all')} arrow>
            <ToggleButton value="all">{t('subscriptions.chain.targetTypes.all')}</ToggleButton>
          </Tooltip>
          <Tooltip title={t('subscriptions.chain.targetTooltips.conditions')} arrow>
            <ToggleButton value="conditions">{t('subscriptions.chain.targetTypes.conditions')}</ToggleButton>
          </Tooltip>
        </ToggleButtonGroup>

        {editingTargetConfig.type === 'specified_node' && (
          <Autocomplete
            size="small"
            options={availableNodes || []}
            sx={fieldControlSx}
            getOptionLabel={(option) => `${option.name || option.linkName} (${option.linkCountry || t('common.unknown')})`}
            value={(availableNodes || []).find((n) => n.id === editingTargetConfig.nodeId) || null}
            onChange={(_event, newValue) => setEditingTargetConfig({ ...editingTargetConfig, nodeId: newValue?.id })}
            renderInput={(params) => (
              <TextField {...params} label={t('subscriptions.chain.selectTargetNode')} placeholder={t('subscriptions.chain.searchNodes')} />
            )}
            renderOption={(props, option) => (
              <li {...props} key={option.id}>
                <Stack direction="row" spacing={1} alignItems="center">
                  <Typography variant="body2">{option.name || option.linkName}</Typography>
                  <Chip label={option.linkCountry || t('common.unknown')} size="small" variant="outlined" />
                  <Chip label={option.protocol || 'unknown'} size="small" color="info" variant="outlined" />
                </Stack>
              </li>
            )}
          />
        )}

        {editingTargetConfig.type === 'conditions' && (
          <ConditionBuilder
            title={t('subscriptions.chain.targetNodeFilterConditions')}
            value={editingTargetConfig.conditions}
            onChange={(conds) => setEditingTargetConfig({ ...editingTargetConfig, conditions: conds })}
            fields={fields}
            operators={operators}
          />
        )}

        <Divider sx={{ borderColor: tokens.softBorder }} />

        <Typography variant="caption" sx={{ color: tokens.tertiaryText, textAlign: 'center' }}>
          {t('subscriptions.chain.autoSaved')}
        </Typography>
      </Stack>
    );
  };

  const onNodeDragStop = useCallback(
    (_event, node) => {
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
        onTargetConfigChange?.({
          ...targetConfig,
          endPosition: node.position
        });
      }
    },
    [chainConfig, targetConfig, onChainConfigChange, onTargetConfigChange]
  );

  return (
    <Box
      className="chain-flow-container"
      sx={{
        height: 450,
        width: '100%',
        display: 'flex',
        overflow: 'hidden',
        '--flow-bg': tokens.mutedPanelSurface,
        '--flow-border': tokens.panelBorder,
        '--flow-border-muted': tokens.subtleBorder,
        '--flow-grid': isDark ? withAlpha(palette.divider, 0.12) : withAlpha(palette.divider, 0.4),
        '--flow-overlay': isDark ? withAlpha(palette.background.paper, 0.1) : withAlpha(palette.background.default, 0.16),
        '--flow-surface-strong': tokens.containerSurface,
        '--flow-shadow': withAlpha(palette.text.primary, isDark ? 0.24 : 0.12),
        '--flow-text': tokens.primaryText,
        '--flow-muted': tokens.secondaryText,
        '--flow-primary': theme.palette.primary.main,
        '--flow-primary-dark': theme.palette.primary.dark,
        '--flow-primary-contrast': theme.palette.primary.contrastText,
        '--flow-hover': tokens.hoverSurface,
        '--flow-toolbar-bg': tokens.elevatedSurface,
        '--flow-handle': theme.palette.primary.main,
        '--flow-handle-border': tokens.nestedPanelSurface,
        '--flow-handle-shadow': alpha(theme.palette.primary.main, 0.18),
        '--flow-panel-bg': tokens.dialogSurface,
        '--flow-panel-input-bg': tokens.fieldSurface,
        '--flow-panel-input-bg-active': tokens.fieldSurfaceActive,
        '--flow-panel-input-highlight': isDark ? withAlpha(theme.palette.common.white, 0.05) : withAlpha(theme.palette.common.black, 0.02),
        '--flow-panel-border-soft': tokens.softBorder,
        '--flow-panel-button-bg': theme.palette.primary.main,
        '--flow-panel-button-bg-hover': theme.palette.primary.dark
      }}
    >
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

        <Box className="chain-flow-toolbar">
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={handleAddProxy}
            size="small"
            disabled={chainConfig.length >= 4}
            sx={{
              backgroundColor: tokens.elevatedSurface,
              border: `1px solid ${tokens.primarySoftBorder}`,
              color: tokens.primaryText,
              boxShadow: tokens.panelShadow,
              '&:hover': { backgroundColor: tokens.hoverSurface, borderColor: tokens.primaryStrongBorder },
              '&.Mui-disabled': { color: palette.text.disabled, borderColor: tokens.subtleBorder }
            }}
          >
            {chainConfig.length >= 4 ? t('subscriptions.chain.maxLevelReached') : t('subscriptions.chain.addProxyNode')}
          </Button>
          {chainConfig.length >= 2 && (
            <Typography variant="caption" sx={{ ml: 1, color: tokens.warningSoftText }}>
              {t('subscriptions.chain.levelWarning', { count: chainConfig.length })}
            </Typography>
          )}
        </Box>

        <Box className="chain-flow-hint">
          <Typography variant="caption" color="text.secondary">
            {t('subscriptions.chain.flowHint')}
          </Typography>
        </Box>
      </Box>

      {panelOpen && (
        <Fade in={panelOpen}>
          <Paper
            className="chain-flow-panel"
            elevation={0}
            sx={{
              width: 480,
              minWidth: 480,
              borderLeft: `1px solid ${tokens.softBorder}`,
              borderRadius: 0,
              p: 2.5,
              overflow: 'auto',
              backgroundColor: tokens.dialogSurface,
              backgroundImage: tokens.dialogSurfaceGradient,
              boxShadow: isDark ? 'none' : `-4px 0 16px ${withAlpha(palette.common.black, 0.1)}`
            }}
          >
            {panelType === 'proxy' ? renderProxyConfigPanel() : renderTargetConfigPanel()}
          </Paper>
        </Fade>
      )}
    </Box>
  );
}
