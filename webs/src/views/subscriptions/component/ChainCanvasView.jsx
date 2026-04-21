import { useState, useMemo, useEffect, memo, useCallback } from 'react';
import PropTypes from 'prop-types';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Chip from '@mui/material/Chip';
import IconButton from '@mui/material/IconButton';
import { alpha, useTheme } from '@mui/material/styles';
import { ReactFlow, Controls, MiniMap, useNodesState, useEdgesState, Handle, Position, getBezierPath } from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import './ChainCanvasView.css';

import PersonIcon from '@mui/icons-material/Person';
import PublicIcon from '@mui/icons-material/Public';
import HubIcon from '@mui/icons-material/Hub';
import MemoryIcon from '@mui/icons-material/Memory';
import RouterIcon from '@mui/icons-material/Router';
import FlagIcon from '@mui/icons-material/Flag';
import GroupWorkIcon from '@mui/icons-material/GroupWork';
import DeviceHubIcon from '@mui/icons-material/DeviceHub';
import AutoAwesomeIcon from '@mui/icons-material/AutoAwesome';
import BlockIcon from '@mui/icons-material/Block';
import WarningAmberIcon from '@mui/icons-material/WarningAmber';
import CloseIcon from '@mui/icons-material/Close';
import { withAlpha } from '../../../utils/colorUtils';

// 国旗转换
const getCountryFlag = (code) => {
  if (!code) return '🌐';
  const codeUpper = code.toUpperCase();
  const offset = 127397;
  return [...codeUpper].map((c) => String.fromCodePoint(c.charCodeAt(0) + offset)).join('');
};

// 类型标签
const getTypeLabel = (type) => {
  const labels = {
    template_group: '模板组',
    custom_group: '自定义组',
    dynamic_node: '动态节点',
    specified_node: '指定节点'
  };
  return labels[type] || type;
};

// 获取节点图标
const getNodeIcon = (type, theme) => {
  const icons = {
    template_group: <GroupWorkIcon sx={{ color: 'info.main', fontSize: 18 }} />,
    custom_group: <DeviceHubIcon sx={{ color: 'secondary.main', fontSize: 18 }} />,
    dynamic_node: <AutoAwesomeIcon sx={{ color: 'warning.main', fontSize: 18 }} />,
    specified_node: <RouterIcon sx={{ color: 'success.main', fontSize: 18 }} />
  };
  return icons[type] || <HubIcon sx={{ color: theme.palette.info.main, fontSize: 18 }} />;
};

// 格式化延迟
const formatLatency = (latency) => {
  if (!latency || latency <= 0) return '-';
  return `${latency}ms`;
};

// 格式化速度 - 后端返回的单位已经是 MB/s
const formatSpeed = (speed) => {
  if (!speed || speed <= 0) return '-';
  // speed 已经是 MB/s 单位的 float64
  return `${speed.toFixed(2)} MB/s`;
};

// 延迟颜色
const getLatencyColor = (latency, theme) => {
  if (!latency || latency <= 0) return theme.palette.text.secondary;
  if (latency < 100) return theme.palette.success.main;
  if (latency < 300) return theme.palette.warning.main;
  return theme.palette.error.main;
};

// 速度颜色 - speed 单位是 MB/s
const getSpeedColor = (speed, theme) => {
  if (!speed || speed <= 0) return theme.palette.text.secondary;
  if (speed >= 10) return theme.palette.success.main;
  if (speed >= 3) return theme.palette.warning.main;
  return theme.palette.error.light;
};

// 详情弹窗组件 - 智能定位确保不超出屏幕
const NodeDetailPanel = memo(({ data, position, onClose }) => {
  const theme = useTheme();
  const hasNodes = data.nodes && data.nodes.length > 0;
  const panelRef = useState(null);

  // 计算安全的显示位置
  const safePosition = useMemo(() => {
    const panelWidth = 400;
    const panelHeight = Math.min(500, 100 + (data.nodes?.length || 0) * 36);
    const padding = 20;

    let x = position.x;
    let y = position.y;

    // 防止超出右边界
    if (x + panelWidth > window.innerWidth - padding) {
      x = position.x - panelWidth - 20;
    }

    // 防止超出底边界
    if (y + panelHeight > window.innerHeight - padding) {
      y = window.innerHeight - panelHeight - padding;
    }

    // 防止超出顶边界
    if (y < padding) {
      y = padding;
    }

    return { x, y };
  }, [position, data.nodes?.length]);

  // 阻止事件冒泡
  const handleClick = (e) => {
    e.stopPropagation();
  };

  return (
    <div
      ref={panelRef}
      className="node-detail-panel"
      style={{
        left: safePosition.x,
        top: safePosition.y
      }}
      onClick={handleClick}
      onMouseDown={handleClick}
    >
      <div className="panel-header">
        <div className="panel-icon">{getNodeIcon(data.type, theme)}</div>
        <div className="panel-title">
          <h4>{data.label || '未配置'}</h4>
          <span>{getTypeLabel(data.type)}</span>
        </div>
        <IconButton size="small" onClick={onClose} sx={{ color: 'text.secondary' }}>
          <CloseIcon sx={{ fontSize: 16 }} />
        </IconButton>
      </div>

      <div className="panel-stats">
        <div className="stat-item">
          <span className="stat-label">包含节点</span>
          <span className="stat-value">{data.nodes?.length || 0}</span>
        </div>
      </div>

      {hasNodes && (
        <div className="panel-nodes-list">
          <div className="list-header">
            <span className="col-name">节点名称</span>
            <span className="col-latency">延迟</span>
            <span className="col-speed">速度</span>
          </div>
          <div className="list-body">
            {data.nodes.map((node, idx) => (
              <div key={idx} className="node-row">
                <span className="col-name">
                  <span className="flag">{getCountryFlag(node.linkCountry)}</span>
                  <span className="name" title={node.name}>
                    {node.name}
                  </span>
                </span>
                <span className="col-latency" style={{ color: getLatencyColor(node.delayTime, theme) }}>
                  {formatLatency(node.delayTime)}
                </span>
                <span className="col-speed" style={{ color: getSpeedColor(node.speed, theme) }}>
                  {formatSpeed(node.speed)}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

      {!hasNodes && <div className="panel-empty">暂无节点信息</div>}
    </div>
  );
});
NodeDetailPanel.displayName = 'NodeDetailPanel';

// 用户节点组件
const UserNode = memo(({ data }) => {
  const isDisabled = data.disabled;
  const isCovered = data.covered;

  return (
    <div className={`sci-fi-node user-node ${isDisabled ? 'disabled' : ''} ${isCovered ? 'covered' : ''}`}>
      <Handle type="source" position={Position.Right} />

      <div className={`rule-label ${isDisabled ? 'disabled' : ''} ${isCovered ? 'covered' : ''}`}>
        {isDisabled && <BlockIcon sx={{ fontSize: 12, mr: 0.5 }} />}
        {isCovered && !isDisabled && <WarningAmberIcon sx={{ fontSize: 12, mr: 0.5, color: 'warning.main' }} />}
        {data.ruleLabel}
        {isCovered && <span className="covered-tag">已被覆盖</span>}
        {isDisabled && <span className="disabled-tag">已禁用</span>}
      </div>

      <div className="node-icon">
        <PersonIcon sx={{ color: 'primary.main' }} />
      </div>
      <div className="node-label">用户</div>
    </div>
  );
});
UserNode.displayName = 'UserNode';

// 代理节点组件 - 点击展开详情
const ProxyNode = memo(({ data }) => {
  const theme = useTheme();
  const isVariant = data.index % 2 === 1;

  const handleClick = useCallback(
    (e) => {
      e.stopPropagation();
      if (data.onShowDetail) {
        const rect = e.currentTarget.getBoundingClientRect();
        data.onShowDetail(data, { x: rect.right + 10, y: rect.top });
      }
    },
    [data]
  );

  return (
    <div className={`sci-fi-node proxy-node ${isVariant ? 'variant' : ''}`} onClick={handleClick}>
      <Handle type="target" position={Position.Left} />
      <Handle type="source" position={Position.Right} />

      <div className="node-icon">{getNodeIcon(data.type, theme)}</div>
      <div className="node-label" title={data.label}>
        {data.label}
      </div>
      <div className="node-type-label">{getTypeLabel(data.type)}</div>
      {data.nodeCount > 0 && (
        <div className="node-count-badge clickable">
          {data.nodeCount} 节点
          <span className="click-hint">点击查看</span>
        </div>
      )}
    </div>
  );
});
ProxyNode.displayName = 'ProxyNode';

// 落地节点组件 - 点击展开详情
const TargetNode = memo(({ data }) => {
  const handleClick = useCallback(
    (e) => {
      e.stopPropagation();
      if (data.onShowDetail) {
        const rect = e.currentTarget.getBoundingClientRect();
        data.onShowDetail(data, { x: rect.right + 10, y: rect.top });
      }
    },
    [data]
  );

  return (
    <div className="sci-fi-node target-node" onClick={handleClick}>
      <Handle type="target" position={Position.Left} />
      <Handle type="source" position={Position.Right} />

      <div className="node-icon">
        <FlagIcon sx={{ color: 'warning.main' }} />
      </div>
      <div className="node-label">落地节点</div>
      <div className="node-type-label">{data.targetInfo || '全部节点'}</div>
      {data.nodeCount > 0 && (
        <div className="node-count-badge clickable">
          {data.nodeCount} 节点
          <span className="click-hint">点击查看</span>
        </div>
      )}
    </div>
  );
});
TargetNode.displayName = 'TargetNode';

// 互联网节点组件
const InternetNode = memo(() => {
  return (
    <div className="sci-fi-node internet-node">
      <Handle type="target" position={Position.Left} />
      <div className="node-icon">
        <PublicIcon sx={{ color: 'success.main' }} />
      </div>
      <div className="node-label">🌐 互联网</div>
    </div>
  );
});
InternetNode.displayName = 'InternetNode';

// 动态粒子边
const AnimatedEdge = ({ id, sourceX, sourceY, targetX, targetY, data }) => {
  const theme = useTheme();
  const [edgePath] = getBezierPath({
    sourceX,
    sourceY,
    targetX,
    targetY,
    curvature: 0.25
  });

  const color = data?.color || theme.palette.primary.main;

  return (
    <>
      <defs>
        <filter id={`glow-${id}`} x="-50%" y="-50%" width="200%" height="200%">
          <feGaussianBlur stdDeviation="3" result="coloredBlur" />
          <feMerge>
            <feMergeNode in="coloredBlur" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>
      </defs>
      <path d={edgePath} fill="none" stroke={color} strokeWidth={4} strokeOpacity={0.08} filter={`url(#glow-${id})`} />
      <path d={edgePath} fill="none" stroke={color} strokeWidth={1.75} strokeOpacity={0.52} />
      <path d={edgePath} fill="none" stroke={color} strokeWidth={2} strokeDasharray="4 16" className="particle-flow" />
    </>
  );
};

// 节点类型定义
const nodeTypes = {
  userNode: UserNode,
  proxyNode: ProxyNode,
  targetNode: TargetNode,
  internetNode: InternetNode
};

// 边类型定义
const edgeTypes = {
  animated: AnimatedEdge
};

// 布局配置
const NODE_H_GAP = 200;
const NODE_V_GAP = 180;
const START_X = 50;
const START_Y = 60;

export default function ChainCanvasView({ rules = [], fullscreen = false }) {
  const theme = useTheme();
  const palette = theme.vars?.palette || theme.palette;
  const isDark = theme.palette.mode === 'dark';
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);

  // 详情面板状态
  const [detailPanel, setDetailPanel] = useState(null);

  // 显示详情面板
  const handleShowDetail = useCallback((nodeData, position) => {
    setDetailPanel({ data: nodeData, position });
  }, []);

  // 关闭详情面板
  const handleCloseDetail = useCallback(() => {
    setDetailPanel(null);
  }, []);

  // 点击画布空白处关闭面板
  const handlePaneClick = useCallback(() => {
    setDetailPanel(null);
  }, []);

  // 构建节点和边
  const buildNodesAndEdges = useCallback(
    (rules) => {
      const nodes = [];
      const edges = [];

      if (!rules || rules.length === 0) return { nodes, edges };

      const edgeColors = [
        theme.palette.primary.main,
        theme.palette.info.main,
        theme.palette.secondary.main,
        theme.palette.success.main,
        theme.palette.warning.main,
        theme.palette.error.main
      ];

      let currentY = START_Y;

      rules.forEach((rule, ruleIndex) => {
        const edgeColor = edgeColors[ruleIndex % edgeColors.length];
        let xOffset = START_X;

        const isDisabled = !rule.enabled;
        const isCovered = rule.enabled && rule.fullyCovered;

        // 用户节点
        const userId = `user-${ruleIndex}`;
        nodes.push({
          id: userId,
          type: 'userNode',
          position: { x: xOffset, y: currentY },
          draggable: true,
          data: {
            label: '用户',
            ruleLabel: rule.ruleName || `规则 ${ruleIndex + 1}`,
            disabled: isDisabled,
            covered: isCovered,
            effectiveNodes: rule.effectiveNodes,
            coveredNodes: rule.coveredNodes
          }
        });
        xOffset += NODE_H_GAP;

        let prevNodeId = userId;

        // 链路节点
        if (rule.links && rule.links.length > 0) {
          rule.links.forEach((link, linkIndex) => {
            const proxyId = `proxy-${ruleIndex}-${linkIndex}`;
            nodes.push({
              id: proxyId,
              type: 'proxyNode',
              position: { x: xOffset, y: currentY },
              draggable: true,
              data: {
                label: link.name || '未配置',
                type: link.type,
                index: linkIndex,
                nodes: link.nodes || [],
                nodeCount: link.nodes?.length || 0,
                onShowDetail: handleShowDetail
              }
            });

            edges.push({
              id: `edge-${prevNodeId}-${proxyId}`,
              source: prevNodeId,
              target: proxyId,
              type: 'animated',
              data: { color: edgeColor }
            });

            prevNodeId = proxyId;
            xOffset += NODE_H_GAP;
          });
        }

        // 落地节点
        const targetId = `target-${ruleIndex}`;
        nodes.push({
          id: targetId,
          type: 'targetNode',
          position: { x: xOffset, y: currentY },
          draggable: true,
          data: {
            label: '落地节点',
            targetInfo: rule.targetInfo || '全部节点',
            type: 'target',
            nodes: rule.targetNodes || [],
            nodeCount: rule.targetNodes?.length || 0,
            onShowDetail: handleShowDetail
          }
        });

        edges.push({
          id: `edge-${prevNodeId}-${targetId}`,
          source: prevNodeId,
          target: targetId,
          type: 'animated',
          data: { color: theme.palette.warning.main }
        });

        xOffset += NODE_H_GAP;

        // 互联网节点
        const internetId = `internet-${ruleIndex}`;
        nodes.push({
          id: internetId,
          type: 'internetNode',
          position: { x: xOffset, y: currentY },
          draggable: true,
          data: {}
        });

        edges.push({
          id: `edge-${targetId}-${internetId}`,
          source: targetId,
          target: internetId,
          type: 'animated',
          data: { color: theme.palette.success.main }
        });

        currentY += NODE_V_GAP;
      });

      return { nodes, edges };
    },
    [handleShowDetail, theme]
  );

  // 规则变化时重新生成
  useEffect(() => {
    const { nodes: newNodes, edges: newEdges } = buildNodesAndEdges(rules);
    setNodes(newNodes);
    setEdges(newEdges);
  }, [rules, buildNodesAndEdges, setNodes, setEdges]);

  const fitViewOptions = useMemo(
    () => ({
      padding: 0.3,
      minZoom: 0.3,
      maxZoom: 1.5
    }),
    []
  );

  if (rules.length === 0) {
    return (
      <Box
        className="chain-canvas-container"
        sx={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          flexDirection: 'column',
          gap: 2
        }}
      >
        <MemoryIcon sx={{ fontSize: 64, color: alpha(theme.palette.primary.main, 0.28) }} />
        <Typography sx={{ color: 'text.secondary', textAlign: 'center' }}>暂无链式代理规则</Typography>
      </Box>
    );
  }

  return (
    <Box
      className={`chain-canvas-container ${fullscreen ? 'fullscreen' : ''}`}
      sx={{
        '--canvas-bg': palette.background.default,
        '--canvas-surface': isDark ? withAlpha(palette.background.default, 0.84) : palette.background.paper,
        '--canvas-surface-strong': isDark ? withAlpha(palette.background.paper, 0.3) : withAlpha(palette.background.paper, 0.98),
        '--canvas-border': isDark ? withAlpha(palette.divider, 0.82) : withAlpha(palette.divider, 0.9),
        '--canvas-grid': isDark ? withAlpha(palette.divider, 0.12) : withAlpha(palette.divider, 0.38),
        '--canvas-muted': palette.text.secondary,
        '--canvas-text': palette.text.primary,
        '--canvas-hover': isDark ? withAlpha(palette.background.paper, 0.22) : withAlpha(palette.background.default, 0.92),
        '--canvas-shadow': alpha(theme.palette.text.primary, isDark ? 0.24 : 0.12),
        '--canvas-primary-soft': alpha(theme.palette.primary.main, 0.1),
        '--canvas-primary-border': alpha(theme.palette.primary.main, 0.28),
        '--canvas-primary-strong': palette.primary.main,
        '--canvas-secondary-soft': alpha(theme.palette.secondary.main, 0.1),
        '--canvas-secondary-border': alpha(theme.palette.secondary.main, 0.28),
        '--canvas-secondary-strong': palette.secondary.main,
        '--canvas-warning-soft': alpha(theme.palette.warning.main, 0.1),
        '--canvas-warning-border': alpha(theme.palette.warning.main, 0.28),
        '--canvas-warning-strong': palette.warning.main,
        '--canvas-warning-main': palette.warning.main,
        '--canvas-success-soft': alpha(theme.palette.success.main, 0.1),
        '--canvas-success-border': alpha(theme.palette.success.main, 0.28),
        '--canvas-success-strong': palette.success.main,
        '--canvas-disabled-soft': alpha(theme.palette.text.secondary, 0.08),
        '--canvas-disabled-border': alpha(theme.palette.text.secondary, 0.22),
        '--canvas-covered-soft': alpha(theme.palette.warning.main, 0.12),
        '--canvas-covered-border': alpha(theme.palette.warning.main, 0.34),
        '--canvas-chip-bg': isDark ? withAlpha(palette.background.paper, 0.22) : withAlpha(palette.background.default, 0.92),
        '--canvas-handle': palette.primary.main,
        '--canvas-panel-header': isDark ? withAlpha(palette.background.default, 0.92) : withAlpha(palette.background.default, 0.9),
        '--canvas-primary-main': palette.primary.main,
        '--canvas-shadow-soft': alpha(theme.palette.primary.main, isDark ? 0.16 : 0.08)
      }}
    >
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        nodeTypes={nodeTypes}
        edgeTypes={edgeTypes}
        fitView
        fitViewOptions={fitViewOptions}
        nodesDraggable={true}
        nodesConnectable={false}
        elementsSelectable={true}
        panOnScroll
        zoomOnScroll
        minZoom={0.2}
        maxZoom={2}
        onPaneClick={handlePaneClick}
        defaultEdgeOptions={{ type: 'animated' }}
        proOptions={{ hideAttribution: true }}
      >
        <Controls showInteractive={false} />
        <MiniMap
          nodeColor={(node) => {
            if (node.type === 'userNode') return theme.palette.primary.main;
            if (node.type === 'proxyNode') return theme.palette.secondary.main;
            if (node.type === 'targetNode') return theme.palette.warning.main;
            if (node.type === 'internetNode') return theme.palette.success.main;
            return theme.palette.text.secondary;
          }}
          maskColor={isDark ? withAlpha(palette.background.default, 0.46) : withAlpha(palette.background.default, 0.16)}
          style={{ background: isDark ? withAlpha(palette.background.paper, 0.72) : withAlpha(palette.background.paper, 0.96) }}
        />
      </ReactFlow>

      {/* 详情面板 - 渲染在最外层确保最高z-index */}
      {detailPanel && <NodeDetailPanel data={detailPanel.data} position={detailPanel.position} onClose={handleCloseDetail} />}

      {/* 图例 */}
      <Box
        sx={{
          position: 'absolute',
          bottom: 16,
          left: 16,
          display: 'flex',
          gap: 1,
          flexWrap: 'wrap',
          zIndex: 5
        }}
      >
        <Chip
          size="small"
          icon={<PersonIcon sx={{ fontSize: 14 }} />}
          label="用户"
          sx={{
            bgcolor: alpha(theme.palette.primary.main, 0.1),
            color: 'primary.main',
            border: '1px solid',
            borderColor: alpha(theme.palette.primary.main, 0.24),
            fontSize: 11
          }}
        />
        <Chip
          size="small"
          icon={<HubIcon sx={{ fontSize: 14 }} />}
          label="代理链"
          sx={{
            bgcolor: alpha(theme.palette.secondary.main, 0.1),
            color: 'secondary.main',
            border: '1px solid',
            borderColor: alpha(theme.palette.secondary.main, 0.24),
            fontSize: 11
          }}
        />
        <Chip
          size="small"
          icon={<FlagIcon sx={{ fontSize: 14 }} />}
          label="落地"
          sx={{
            bgcolor: alpha(theme.palette.warning.main, 0.1),
            color: 'warning.main',
            border: '1px solid',
            borderColor: alpha(theme.palette.warning.main, 0.24),
            fontSize: 11
          }}
        />
        <Chip
          size="small"
          icon={<PublicIcon sx={{ fontSize: 14 }} />}
          label="互联网"
          sx={{
            bgcolor: alpha(theme.palette.success.main, 0.1),
            color: 'success.main',
            border: '1px solid',
            borderColor: alpha(theme.palette.success.main, 0.24),
            fontSize: 11
          }}
        />
      </Box>
    </Box>
  );
}

ChainCanvasView.propTypes = {
  rules: PropTypes.array,
  fullscreen: PropTypes.bool
};
