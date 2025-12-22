import { useState, useMemo } from 'react';
import PropTypes from 'prop-types';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import IconButton from '@mui/material/IconButton';
import Typography from '@mui/material/Typography';
import Chip from '@mui/material/Chip';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Collapse from '@mui/material/Collapse';
import CircularProgress from '@mui/material/CircularProgress';
import Tab from '@mui/material/Tab';
import Tabs from '@mui/material/Tabs';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import useMediaQuery from '@mui/material/useMediaQuery';
import { useTheme, alpha, keyframes } from '@mui/material/styles';

import CloseIcon from '@mui/icons-material/Close';
import PersonIcon from '@mui/icons-material/Person';
import PublicIcon from '@mui/icons-material/Public';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';
import RouteIcon from '@mui/icons-material/Route';
import LayersIcon from '@mui/icons-material/Layers';
import AdjustIcon from '@mui/icons-material/Adjust';
import FlagIcon from '@mui/icons-material/Flag';
import AccountTreeIcon from '@mui/icons-material/AccountTree';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import CancelIcon from '@mui/icons-material/Cancel';

import ChainCanvasView from './ChainCanvasView';

// 箭头脉冲动画
const pulseAnimation = keyframes`
  0%, 100% {
    opacity: 0.4;
    transform: translateX(0);
  }
  50% {
    opacity: 1;
    transform: translateX(4px);
  }
`;

// 国旗转换
const getCountryFlag = (code) => {
  if (!code) return '🌐';
  const codeUpper = code.toUpperCase();
  const offset = 127397;
  return [...codeUpper].map((c) => String.fromCodePoint(c.charCodeAt(0) + offset)).join('');
};

// 类型标签颜色
const getTypeColor = (type, theme) => {
  const colors = {
    template_group: theme.palette.primary.main,
    custom_group: theme.palette.secondary.main,
    dynamic_node: theme.palette.warning.main,
    specified_node: theme.palette.success.main
  };
  return colors[type] || theme.palette.grey[500];
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

// 链路节点卡片（紧凑版）
function ChainNodeCard({ node, index, isLast, isMobile, theme }) {
  const [expanded, setExpanded] = useState(false);
  const hasNodes = node.nodes && node.nodes.length > 0;
  const typeColor = getTypeColor(node.type, theme);

  return (
    <Box sx={{ display: 'flex', flexDirection: isMobile ? 'column' : 'row', alignItems: 'center' }}>
      <Card
        sx={{
          minWidth: isMobile ? '100%' : 140,
          maxWidth: isMobile ? '100%' : 180,
          background: alpha(typeColor, 0.08),
          border: `1px solid ${alpha(typeColor, 0.3)}`,
          borderRadius: 2,
          transition: 'all 0.2s ease',
          '&:hover': {
            borderColor: typeColor,
            boxShadow: `0 2px 8px ${alpha(typeColor, 0.2)}`
          }
        }}
      >
        <CardContent sx={{ p: 1.5, '&:last-child': { pb: 1.5 } }}>
          <Stack direction="row" alignItems="center" spacing={0.5} mb={0.5}>
            <Chip
              size="small"
              label={getTypeLabel(node.type)}
              sx={{
                bgcolor: alpha(typeColor, 0.15),
                color: typeColor,
                fontWeight: 600,
                fontSize: 10,
                height: 18,
                '& .MuiChip-label': { px: 0.75 }
              }}
            />
            <Typography variant="caption" color="text.secondary">
              #{index + 1}
            </Typography>
          </Stack>

          <Typography
            variant="body2"
            fontWeight={600}
            sx={{
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
              mb: hasNodes ? 0.5 : 0
            }}
            title={node.name}
          >
            {node.name || '未配置'}
          </Typography>

          {hasNodes && (
            <Box
              onClick={() => setExpanded(!expanded)}
              sx={{
                display: 'flex',
                alignItems: 'center',
                cursor: 'pointer',
                color: 'primary.main',
                '&:hover': { textDecoration: 'underline' }
              }}
            >
              <Typography variant="caption" fontWeight={600}>
                {node.nodes.length} 节点
              </Typography>
              {expanded ? <ExpandLessIcon sx={{ fontSize: 16 }} /> : <ExpandMoreIcon sx={{ fontSize: 16 }} />}
            </Box>
          )}
          <Collapse in={expanded}>
            <Box
              sx={{
                mt: 0.5,
                maxHeight: 100,
                overflow: 'auto',
                bgcolor: alpha(theme.palette.background.paper, 0.8),
                borderRadius: 1,
                p: 0.5
              }}
            >
              {node.nodes?.map((n, i) => (
                <Typography key={i} variant="caption" display="block" sx={{ py: 0.25 }}>
                  {getCountryFlag(n.linkCountry)} {n.name}
                </Typography>
              ))}
            </Box>
          </Collapse>
        </CardContent>
      </Card>

      {!isLast && (
        <Box
          sx={{
            display: 'flex',
            alignItems: 'center',
            px: isMobile ? 0 : 1,
            py: isMobile ? 0.5 : 0,
            transform: isMobile ? 'rotate(90deg)' : 'none'
          }}
        >
          {[0, 1].map((i) => (
            <Box
              key={i}
              sx={{
                width: 0,
                height: 0,
                borderTop: '5px solid transparent',
                borderBottom: '5px solid transparent',
                borderLeft: `6px solid ${alpha(theme.palette.primary.main, 0.5)}`,
                ml: i > 0 ? -0.3 : 0,
                animation: `${pulseAnimation} 1.5s ease-in-out infinite`,
                animationDelay: `${i * 0.15}s`
              }}
            />
          ))}
        </Box>
      )}
    </Box>
  );
}

ChainNodeCard.propTypes = {
  node: PropTypes.object.isRequired,
  index: PropTypes.number.isRequired,
  isLast: PropTypes.bool.isRequired,
  isMobile: PropTypes.bool.isRequired,
  theme: PropTypes.object.isRequired
};

// 单条规则的链路图
function RuleChainFlow({ rule, isMobile, theme }) {
  const [expanded, setExpanded] = useState(!rule.fullyCovered);

  // 规则是否完全被覆盖（无生效节点）
  const isFullyCovered = rule.enabled && rule.fullyCovered;

  return (
    <Card
      variant="outlined"
      sx={{
        mb: 2,
        borderRadius: 2,
        opacity: rule.enabled ? (isFullyCovered ? 0.6 : 1) : 0.4,
        transition: 'all 0.2s ease',
        borderColor: isFullyCovered ? 'warning.main' : 'divider',
        bgcolor: isFullyCovered ? alpha(theme.palette.warning.main, 0.03) : 'background.paper'
      }}
    >
      <CardContent sx={{ p: 2, '&:last-child': { pb: 2 } }}>
        {/* 规则标题 */}
        <Stack direction="row" alignItems="center" justifyContent="space-between" mb={1.5}>
          <Stack direction="row" alignItems="center" spacing={1} flexWrap="wrap">
            <Typography variant="subtitle1" fontWeight={700} sx={{ textDecoration: isFullyCovered ? 'line-through' : 'none' }}>
              {rule.ruleName || '未命名规则'}
            </Typography>
            {!rule.enabled && <Chip label="已禁用" size="small" color="default" />}
            {rule.enabled && isFullyCovered && <Chip label="已被覆盖" size="small" color="warning" variant="outlined" />}
            {rule.enabled && !isFullyCovered && rule.effectiveNodes > 0 && (
              <Chip label={`生效 ${rule.effectiveNodes} 节点`} size="small" color="success" variant="outlined" />
            )}
            {rule.enabled && rule.coveredNodes > 0 && !isFullyCovered && (
              <Chip label={`${rule.coveredNodes} 被覆盖`} size="small" color="warning" variant="outlined" />
            )}
          </Stack>
          <IconButton size="small" onClick={() => setExpanded(!expanded)}>
            {expanded ? <ExpandLessIcon /> : <ExpandMoreIcon />}
          </IconButton>
        </Stack>

        <Collapse in={expanded}>
          {/* 链路图 */}
          <Box
            sx={{
              display: 'flex',
              flexDirection: isMobile ? 'column' : 'row',
              alignItems: 'center',
              gap: 0,
              py: 1,
              px: isMobile ? 0 : 1,
              overflowX: 'auto'
            }}
          >
            {/* 用户端点 */}
            <Box sx={{ display: 'flex', flexDirection: isMobile ? 'column' : 'row', alignItems: 'center' }}>
              <Box
                sx={{
                  px: 1.5,
                  py: 0.75,
                  borderRadius: 2,
                  bgcolor: alpha(theme.palette.info.main, 0.1),
                  border: `1px dashed ${alpha(theme.palette.info.main, 0.4)}`,
                  textAlign: 'center',
                  minWidth: 60
                }}
              >
                <PersonIcon sx={{ color: theme.palette.info.main, fontSize: 20 }} />
                <Typography variant="caption" display="block" fontWeight={600}>
                  用户
                </Typography>
              </Box>
              <Box
                sx={{
                  display: 'flex',
                  alignItems: 'center',
                  px: isMobile ? 0 : 1,
                  py: isMobile ? 0.5 : 0,
                  transform: isMobile ? 'rotate(90deg)' : 'none'
                }}
              >
                {[0, 1].map((i) => (
                  <Box
                    key={i}
                    sx={{
                      width: 0,
                      height: 0,
                      borderTop: '5px solid transparent',
                      borderBottom: '5px solid transparent',
                      borderLeft: `6px solid ${alpha(theme.palette.info.main, 0.4)}`,
                      ml: i > 0 ? -0.3 : 0
                    }}
                  />
                ))}
              </Box>
            </Box>

            {/* 链路节点 */}
            {rule.links?.map((node, index) => (
              <ChainNodeCard
                key={index}
                node={node}
                index={index}
                isLast={index === rule.links.length - 1}
                isMobile={isMobile}
                theme={theme}
              />
            ))}

            {/* 箭头到目标节点 */}
            {rule.links?.length > 0 && (
              <Box
                sx={{
                  display: 'flex',
                  alignItems: 'center',
                  px: isMobile ? 0 : 1,
                  py: isMobile ? 0.5 : 0,
                  transform: isMobile ? 'rotate(90deg)' : 'none'
                }}
              >
                {[0, 1].map((i) => (
                  <Box
                    key={i}
                    sx={{
                      width: 0,
                      height: 0,
                      borderTop: '5px solid transparent',
                      borderBottom: '5px solid transparent',
                      borderLeft: `6px solid ${alpha(theme.palette.error.main, 0.4)}`,
                      ml: i > 0 ? -0.3 : 0
                    }}
                  />
                ))}
              </Box>
            )}

            {/* 目标节点 */}
            <Box
              sx={{
                px: 1.5,
                py: 0.75,
                borderRadius: 2,
                bgcolor: alpha(theme.palette.error.main, 0.1),
                border: `1px solid ${alpha(theme.palette.error.main, 0.3)}`,
                textAlign: 'center',
                minWidth: isMobile ? '100%' : 100
              }}
            >
              {rule.targetType === 'all' ? (
                <LayersIcon sx={{ color: theme.palette.error.main, fontSize: 20 }} />
              ) : rule.targetType === 'conditions' ? (
                <AdjustIcon sx={{ color: theme.palette.error.main, fontSize: 20 }} />
              ) : (
                <FlagIcon sx={{ color: theme.palette.error.main, fontSize: 20 }} />
              )}
              <Typography variant="caption" display="block" fontWeight={600} color="error.main">
                落地节点
              </Typography>
              <Typography variant="caption" display="block" color="text.secondary">
                {rule.targetInfo}
              </Typography>
              {rule.targetNodes?.length > 0 && (
                <Typography variant="caption" color="text.secondary">
                  ({rule.targetNodes.length} 节点)
                </Typography>
              )}
            </Box>

            {/* 箭头到互联网 */}
            <Box
              sx={{
                display: 'flex',
                alignItems: 'center',
                px: isMobile ? 0 : 1,
                py: isMobile ? 0.5 : 0,
                transform: isMobile ? 'rotate(90deg)' : 'none'
              }}
            >
              {[0, 1].map((i) => (
                <Box
                  key={i}
                  sx={{
                    width: 0,
                    height: 0,
                    borderTop: '5px solid transparent',
                    borderBottom: '5px solid transparent',
                    borderLeft: `6px solid ${alpha(theme.palette.success.main, 0.4)}`,
                    ml: i > 0 ? -0.3 : 0
                  }}
                />
              ))}
            </Box>

            {/* 互联网 */}
            <Box
              sx={{
                px: 1.5,
                py: 0.75,
                borderRadius: 2,
                bgcolor: alpha(theme.palette.success.main, 0.1),
                border: `1px dashed ${alpha(theme.palette.success.main, 0.4)}`,
                textAlign: 'center',
                minWidth: 60
              }}
            >
              <PublicIcon sx={{ color: theme.palette.success.main, fontSize: 20 }} />
              <Typography variant="caption" display="block" fontWeight={600}>
                🌐 互联网
              </Typography>
            </Box>
          </Box>
        </Collapse>
      </CardContent>
    </Card>
  );
}

RuleChainFlow.propTypes = {
  rule: PropTypes.object.isRequired,
  isMobile: PropTypes.bool.isRequired,
  theme: PropTypes.object.isRequired
};

// 节点匹配摘要表格
function NodeMatchTable({ matchSummary, isMobile }) {
  const matchedCount = matchSummary?.filter((n) => !n.unmatched).length || 0;
  const unmatchedCount = matchSummary?.filter((n) => n.unmatched).length || 0;

  return (
    <Box>
      <Stack direction="row" spacing={2} mb={2}>
        <Chip icon={<CheckCircleIcon />} label={`已匹配: ${matchedCount}`} color="success" variant="outlined" size="small" />
        <Chip icon={<CancelIcon />} label={`未匹配: ${unmatchedCount}`} color="default" variant="outlined" size="small" />
      </Stack>

      <Box sx={{ maxHeight: 300, overflow: 'auto' }}>
        <Table size="small" stickyHeader>
          <TableHead>
            <TableRow>
              <TableCell>节点</TableCell>
              <TableCell>匹配规则</TableCell>
              <TableCell>入口代理</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {matchSummary?.map((node) => (
              <TableRow key={node.nodeId} sx={{ opacity: node.unmatched ? 0.5 : 1 }}>
                <TableCell>
                  <Stack direction="row" alignItems="center" spacing={0.5}>
                    <Typography variant="caption">{getCountryFlag(node.linkCountry)}</Typography>
                    <Typography
                      variant="body2"
                      sx={{ maxWidth: isMobile ? 100 : 200, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}
                    >
                      {node.nodeName}
                    </Typography>
                  </Stack>
                </TableCell>
                <TableCell>
                  {node.unmatched ? (
                    <Typography variant="caption" color="text.disabled">
                      无
                    </Typography>
                  ) : (
                    <Typography variant="body2">{node.matchedRule}</Typography>
                  )}
                </TableCell>
                <TableCell>
                  {node.entryProxy ? (
                    <Typography variant="body2">{node.entryProxy}</Typography>
                  ) : (
                    <Typography variant="caption" color="text.disabled">
                      -
                    </Typography>
                  )}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Box>
    </Box>
  );
}

NodeMatchTable.propTypes = {
  matchSummary: PropTypes.array,
  isMobile: PropTypes.bool.isRequired
};

/**
 * 链式代理预览对话框
 * 展示整个订阅的所有规则及节点匹配情况
 */
export default function ChainPreviewDialog({ open, onClose, loading, data }) {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));
  const [tab, setTab] = useState(0);

  const rules = useMemo(() => data?.rules || [], [data?.rules]);
  const matchSummary = useMemo(() => data?.matchSummary || [], [data?.matchSummary]);

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="lg"
      fullWidth
      fullScreen={isMobile}
      PaperProps={{
        sx: {
          borderRadius: isMobile ? 0 : 3,
          minHeight: isMobile ? 'auto' : '70vh'
        }
      }}
    >
      <DialogTitle sx={{ pb: 1 }}>
        <Stack direction="row" alignItems="center" justifyContent="space-between">
          <Stack direction="row" alignItems="center" spacing={1.5}>
            <RouteIcon sx={{ color: 'primary.main' }} />
            <Box>
              <Typography variant={isMobile ? 'subtitle1' : 'h6'} fontWeight={700}>
                🔗 链路预览
              </Typography>
              {data?.subscriptionName && (
                <Typography variant="caption" color="text.secondary">
                  订阅：{data.subscriptionName} | 节点总数：{data.totalNodes}
                </Typography>
              )}
            </Box>
          </Stack>
          <IconButton onClick={onClose} size="small">
            <CloseIcon />
          </IconButton>
        </Stack>
      </DialogTitle>

      <DialogContent sx={{ pt: 0 }}>
        {loading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 200 }}>
            <CircularProgress />
          </Box>
        ) : rules.length === 0 ? (
          <Box sx={{ textAlign: 'center', py: 6 }}>
            <AccountTreeIcon sx={{ fontSize: 56, color: 'text.disabled', mb: 2 }} />
            <Typography variant="h6" color="text.secondary">
              暂无链式代理规则
            </Typography>
            <Typography variant="body2" color="text.disabled">
              请先添加规则
            </Typography>
          </Box>
        ) : (
          <>
            <Tabs value={tab} onChange={(e, v) => setTab(v)} sx={{ borderBottom: 1, borderColor: 'divider', mb: 2 }}>
              <Tab label={`规则链路 (${rules.length})`} />
              <Tab label={`节点匹配 (${matchSummary.length})`} />
            </Tabs>

            {tab === 0 && (
              <Box>
                <Typography variant="caption" color="text.secondary" display="block" mb={2}>
                  💡 规则按顺序匹配，每个节点只会应用第一个匹配的规则 · 鼠标滚轮缩放，拖拽平移画布
                </Typography>
                <ChainCanvasView rules={rules} fullscreen={isMobile} />
              </Box>
            )}

            {tab === 1 && <NodeMatchTable matchSummary={matchSummary} isMobile={isMobile} />}
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}

ChainPreviewDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  loading: PropTypes.bool,
  data: PropTypes.shape({
    subscriptionName: PropTypes.string,
    totalNodes: PropTypes.number,
    rules: PropTypes.array,
    matchSummary: PropTypes.array
  })
};
