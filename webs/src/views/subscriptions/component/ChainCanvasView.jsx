import { useState, useMemo, useEffect, memo, useCallback, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import PropTypes from 'prop-types';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Chip from '@mui/material/Chip';
import IconButton from '@mui/material/IconButton';
import { alpha, useTheme } from '@mui/material/styles';
import { ReactFlow, Controls, MiniMap, useNodesState, useEdgesState, Handle, Position, getBezierPath } from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import './ChainCanvasView.css';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';

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
import { getChainProxyCanvasCssVars, getChainProxyThemeTokens } from './chainProxyTheme';

const getCountryFlag = (code) => {
  if (!code) return '🌐';
  const codeUpper = code.toUpperCase();
  const offset = 127397;
  return [...codeUpper].map((c) => String.fromCodePoint(c.charCodeAt(0) + offset)).join('');
};

const getTypeLabel = (type, t) => {
  const key = `subscriptions.chain.proxyTypeShort.${type}`;
  return t(key, type);
};

const getNodeIcon = (type, theme) => {
  const icons = {
    template_group: <GroupWorkIcon sx={{ color: 'info.main', fontSize: 18 }} />,
    custom_group: <DeviceHubIcon sx={{ color: 'secondary.main', fontSize: 18 }} />,
    dynamic_node: <AutoAwesomeIcon sx={{ color: 'warning.main', fontSize: 18 }} />,
    specified_node: <RouterIcon sx={{ color: 'success.main', fontSize: 18 }} />
  };
  return icons[type] || <HubIcon sx={{ color: theme.palette.info.main, fontSize: 18 }} />;
};

const formatLatency = (latency) => {
  if (!latency || latency <= 0) return '-';
  return `${latency}ms`;
};

const formatSpeed = (speed) => {
  if (!speed || speed <= 0) return '-';
  return `${speed.toFixed(2)} MB/s`;
};

const getLatencyColor = (latency, theme) => {
  if (!latency || latency <= 0) return theme.palette.text.secondary;
  if (latency < 100) return theme.palette.success.main;
  if (latency < 300) return theme.palette.warning.main;
  return theme.palette.error.main;
};

const getSpeedColor = (speed, theme) => {
  if (!speed || speed <= 0) return theme.palette.text.secondary;
  if (speed >= 10) return theme.palette.success.main;
  if (speed >= 3) return theme.palette.warning.main;
  return theme.palette.error.light;
};

const NodeDetailPanel = memo(({ data, position, onClose }) => {
  const { t } = useTranslation();
  const theme = useTheme();
  const hasNodes = data.nodes && data.nodes.length > 0;
  const panelRef = useRef(null);

  const safePosition = useMemo(() => {
    const bounds = position.bounds || { width: window.innerWidth, height: window.innerHeight };
    const isNarrow = bounds.width <= 768;
    const panelWidth = isNarrow ? Math.max(280, bounds.width - 40) : Math.min(400, Math.max(280, bounds.width - 40));
    const panelHeight = Math.min(500, 100 + (data.nodes?.length || 0) * 36);
    const padding = 20;

    let x = position.x;
    let y = position.y;

    if (isNarrow) {
      x = padding;
    } else if (x + panelWidth > bounds.width - padding) {
      x = (position.left ?? position.x) - panelWidth - 10;
    }

    if (x < padding) {
      x = padding;
    }

    if (y + panelHeight > bounds.height - padding) {
      y = bounds.height - panelHeight - padding;
    }

    if (y < padding) {
      y = padding;
    }

    return { x, y, width: panelWidth };
  }, [position, data.nodes?.length]);

  const handleClick = (e) => {
    e.stopPropagation();
  };

  return (
    <div
      ref={panelRef}
      className="node-detail-panel"
      style={{
        left: safePosition.x,
        top: safePosition.y,
        width: safePosition.width
      }}
      onClick={handleClick}
      onMouseDown={handleClick}
    >
      <div className="panel-header">
        <div className="panel-icon">{getNodeIcon(data.type, theme)}</div>
        <div className="panel-title">
          <h4>{data.label || t('subscriptions.chain.unconfigured')}</h4>
          <span>{getTypeLabel(data.type, t)}</span>
        </div>
        <IconButton size="small" onClick={onClose} sx={{ color: 'text.secondary' }}>
          <CloseIcon sx={{ fontSize: 16 }} />
        </IconButton>
      </div>

      <div className="panel-stats">
        <div className="stat-item">
          <span className="stat-label">{t('subscriptions.chain.preview.containsNodes')}</span>
          <span className="stat-value">{data.nodes?.length || 0}</span>
        </div>
      </div>

      {hasNodes && (
        <div className="panel-nodes-list">
          <div className="list-header">
            <span className="col-name">{t('subscriptions.chain.preview.table.node')}</span>
            <span className="col-latency">{t('subscriptions.chain.preview.latency')}</span>
            <span className="col-speed">{t('subscriptions.chain.preview.speed')}</span>
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

      {!hasNodes && <div className="panel-empty">{t('subscriptions.chain.preview.noNodeInfo')}</div>}
    </div>
  );
});
NodeDetailPanel.displayName = 'NodeDetailPanel';

const UserNode = memo(({ data }) => {
  const { t } = useTranslation();
  const isDisabled = data.disabled;
  const isCovered = data.covered;

  return (
    <div className={`sci-fi-node user-node ${isDisabled ? 'disabled' : ''} ${isCovered ? 'covered' : ''}`}>
      <Handle type="source" position={Position.Right} />

      <div className={`rule-label ${isDisabled ? 'disabled' : ''} ${isCovered ? 'covered' : ''}`}>
        {isDisabled && <BlockIcon sx={{ fontSize: 12, mr: 0.5 }} />}
        {isCovered && !isDisabled && <WarningAmberIcon sx={{ fontSize: 12, mr: 0.5, color: 'warning.main' }} />}
        {data.ruleLabel}
        {isCovered && <span className="covered-tag">{t('subscriptions.chain.preview.covered')}</span>}
        {isDisabled && <span className="disabled-tag">{t('common.disabled')}</span>}
      </div>

      <div className="node-icon">
        <PersonIcon sx={{ color: 'primary.main' }} />
      </div>
      <div className="node-label">{t('subscriptions.chain.preview.user')}</div>
    </div>
  );
});
UserNode.displayName = 'UserNode';

const ProxyNode = memo(({ data }) => {
  const { t } = useTranslation();
  const theme = useTheme();
  const isVariant = data.index % 2 === 1;

  const handleClick = useCallback(
    (e) => {
      e.stopPropagation();
      if (data.onShowDetail) {
        const rect = e.currentTarget.getBoundingClientRect();
        data.onShowDetail(data, { x: rect.right + 10, y: rect.top, left: rect.left });
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
      <div className="node-type-label">{getTypeLabel(data.type, t)}</div>
      {data.nodeCount > 0 && (
        <div className="node-count-badge clickable">
          {t('subscriptions.chain.preview.nodeCount', { count: data.nodeCount })}
          <span className="click-hint">{t('subscriptions.chain.preview.clickToView')}</span>
        </div>
      )}
    </div>
  );
});
ProxyNode.displayName = 'ProxyNode';

const TargetNode = memo(({ data }) => {
  const { t } = useTranslation();
  const handleClick = useCallback(
    (e) => {
      e.stopPropagation();
      if (data.onShowDetail) {
        const rect = e.currentTarget.getBoundingClientRect();
        data.onShowDetail(data, { x: rect.right + 10, y: rect.top, left: rect.left });
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
      <div className="node-label">{t('subscriptions.chain.preview.landingNode')}</div>
      <div className="node-type-label">{data.targetInfo || t('subscriptions.chain.targetTypes.all')}</div>
      {data.nodeCount > 0 && (
        <div className="node-count-badge clickable">
          {t('subscriptions.chain.preview.nodeCount', { count: data.nodeCount })}
          <span className="click-hint">{t('subscriptions.chain.preview.clickToView')}</span>
        </div>
      )}
    </div>
  );
});
TargetNode.displayName = 'TargetNode';

const InternetNode = memo(() => {
  const { t } = useTranslation();
  return (
    <div className="sci-fi-node internet-node">
      <Handle type="target" position={Position.Left} />
      <div className="node-icon">
        <PublicIcon sx={{ color: 'success.main' }} />
      </div>
      <div className="node-label">{t('subscriptions.chain.preview.internet')}</div>
    </div>
  );
});
InternetNode.displayName = 'InternetNode';

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

const nodeTypes = {
  userNode: UserNode,
  proxyNode: ProxyNode,
  targetNode: TargetNode,
  internetNode: InternetNode
};

const edgeTypes = {
  animated: AnimatedEdge
};

const NODE_H_GAP = 200;
const NODE_V_GAP = 180;
const START_X = 50;
const START_Y = 60;

export default function ChainCanvasView({ rules = [], fullscreen = false }) {
  const { t } = useTranslation();
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const tokens = getChainProxyThemeTokens(theme, isDark);
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const canvasRef = useRef(null);

  const [detailPanel, setDetailPanel] = useState(null);

  const handleShowDetail = useCallback((nodeData, position) => {
    const canvasRect = canvasRef.current?.getBoundingClientRect();

    if (!canvasRect) {
      setDetailPanel({ data: nodeData, position });
      return;
    }

    setDetailPanel({
      data: nodeData,
      position: {
        x: position.x - canvasRect.left,
        y: position.y - canvasRect.top,
        left: position.left - canvasRect.left,
        bounds: {
          width: canvasRect.width,
          height: canvasRect.height
        }
      }
    });
  }, []);

  const handleCloseDetail = useCallback(() => {
    setDetailPanel(null);
  }, []);

  const handlePaneClick = useCallback(() => {
    setDetailPanel(null);
  }, []);

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

        const userId = `user-${ruleIndex}`;
        nodes.push({
          id: userId,
          type: 'userNode',
          position: { x: xOffset, y: currentY },
          draggable: true,
          data: {
            label: t('subscriptions.chain.preview.user'),
            ruleLabel: rule.ruleName || t('subscriptions.chain.preview.ruleNumber', { number: ruleIndex + 1 }),
            disabled: isDisabled,
            covered: isCovered,
            effectiveNodes: rule.effectiveNodes,
            coveredNodes: rule.coveredNodes
          }
        });
        xOffset += NODE_H_GAP;

        let prevNodeId = userId;

        if (rule.links && rule.links.length > 0) {
          rule.links.forEach((link, linkIndex) => {
            const proxyId = `proxy-${ruleIndex}-${linkIndex}`;
            nodes.push({
              id: proxyId,
              type: 'proxyNode',
              position: { x: xOffset, y: currentY },
              draggable: true,
              data: {
                label: link.name || t('subscriptions.chain.unconfigured'),
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

        const targetId = `target-${ruleIndex}`;
        nodes.push({
          id: targetId,
          type: 'targetNode',
          position: { x: xOffset, y: currentY },
          draggable: true,
          data: {
            label: t('subscriptions.chain.preview.landingNode'),
            targetInfo: rule.targetInfo || t('subscriptions.chain.targetTypes.all'),
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
    [handleShowDetail, theme, t]
  );

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
          ...getChainProxyCanvasCssVars(tokens),
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          flexDirection: 'column',
          gap: 2
        }}
      >
        <MemoryIcon sx={{ fontSize: 64, color: alpha(theme.palette.primary.main, 0.28) }} />
        <Typography sx={{ color: 'text.secondary', textAlign: 'center' }}>{t('subscriptions.chain.emptyRules')}</Typography>
      </Box>
    );
  }

  return (
    <Box
      className={`chain-canvas-container ${fullscreen ? 'fullscreen' : ''}`}
      ref={canvasRef}
      sx={{
        ...getChainProxyCanvasCssVars(tokens)
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
          maskColor={tokens.containerSurface}
          style={{ background: tokens.canvasElevatedSurface }}
        />
      </ReactFlow>

      {detailPanel && <NodeDetailPanel data={detailPanel.data} position={detailPanel.position} onClose={handleCloseDetail} />}

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
          label={t('subscriptions.chain.preview.user')}
          sx={{
            bgcolor: tokens.primarySurface,
            color: 'primary.main',
            border: '1px solid',
            borderColor: tokens.primarySoftBorder,
            fontSize: 11
          }}
        />
        <Chip
          size="small"
          icon={<HubIcon sx={{ fontSize: 14 }} />}
          label={t('subscriptions.chain.preview.proxyChain')}
          sx={{
            bgcolor: tokens.secondarySurface,
            color: 'secondary.main',
            border: '1px solid',
            borderColor: tokens.secondarySoftBorder,
            fontSize: 11
          }}
        />
        <Chip
          size="small"
          icon={<FlagIcon sx={{ fontSize: 14 }} />}
          label={t('subscriptions.chain.preview.landing')}
          sx={{
            bgcolor: tokens.warningSurface,
            color: 'warning.main',
            border: '1px solid',
            borderColor: tokens.warningSoftBorder,
            fontSize: 11
          }}
        />
        <Chip
          size="small"
          icon={<PublicIcon sx={{ fontSize: 14 }} />}
          label={t('subscriptions.chain.preview.internetPlain')}
          sx={{
            bgcolor: tokens.successSurface,
            color: 'success.main',
            border: '1px solid',
            borderColor: tokens.successSoftBorder,
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
