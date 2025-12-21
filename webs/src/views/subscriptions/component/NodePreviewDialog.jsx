import PropTypes from 'prop-types';
import { useState, useMemo, useCallback, useRef, useEffect } from 'react';

// material-ui
import { useTheme, alpha } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Collapse from '@mui/material/Collapse';
import Dialog from '@mui/material/Dialog';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import Grid from '@mui/material/Grid';
import IconButton from '@mui/material/IconButton';
import InputAdornment from '@mui/material/InputAdornment';
import Skeleton from '@mui/material/Skeleton';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import Chip from '@mui/material/Chip';
import Tooltip from '@mui/material/Tooltip';

// icons
import CloseIcon from '@mui/icons-material/Close';
import SearchIcon from '@mui/icons-material/Search';
import FilterListIcon from '@mui/icons-material/FilterList';
import ArrowForwardIcon from '@mui/icons-material/ArrowForward';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import AccessTimeIcon from '@mui/icons-material/AccessTime';
import SpeedIcon from '@mui/icons-material/Speed';
import EmojiEventsIcon from '@mui/icons-material/EmojiEvents';
import KeyboardArrowDownIcon from '@mui/icons-material/KeyboardArrowDown';
import KeyboardArrowUpIcon from '@mui/icons-material/KeyboardArrowUp';

// project imports
import NodePreviewCard from './NodePreviewCard';
import NodePreviewDetailsPanel from './NodePreviewDetailsPanel';
import IPDetailsDialog from 'components/IPDetailsDialog';
import Alert from '@mui/material/Alert';
import { AlertTitle } from '@mui/material';

// 每次加载的卡片数量
const BATCH_SIZE = 100;

// 格式化字节数
const formatBytes = (bytes) => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

// 格式化到期时间
const formatExpireDate = (timestamp) => {
  if (!timestamp) return '-';
  const date = new Date(timestamp * 1000);
  return date.toLocaleDateString('zh-CN');
};

/**
 * 节点预览对话框组件
 * 展示应用过滤和重命名规则后的节点列表
 * 使用渐进式加载优化大量卡片的性能
 */
export default function NodePreviewDialog({ open, loading, data, tagColorMap, onClose }) {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const contentRef = useRef(null);

  // 搜索状态
  const [searchText, setSearchText] = useState('');
  // 详情面板状态
  const [detailsOpen, setDetailsOpen] = useState(false);
  const [selectedNode, setSelectedNode] = useState(null);
  // IP详情状态
  const [ipDialogOpen, setIpDialogOpen] = useState(false);
  const [selectedIP, setSelectedIP] = useState('');
  // 渐进式加载状态
  const [displayCount, setDisplayCount] = useState(BATCH_SIZE);
  // 统计面板展开状态
  const [statsExpanded, setStatsExpanded] = useState(true);

  // 当对话框关闭或数据变化时重置显示数量
  useEffect(() => {
    if (!open) {
      setDisplayCount(BATCH_SIZE);
      setSearchText('');
    }
  }, [open]);

  // 搜索变化时重置显示数量
  useEffect(() => {
    setDisplayCount(BATCH_SIZE);
  }, [searchText]);

  // 过滤后的节点列表
  const filteredNodes = useMemo(() => {
    if (!data?.Nodes) return [];
    if (!searchText.trim()) return data.Nodes;

    const lowerSearch = searchText.toLowerCase();
    return data.Nodes.filter(
      (node) =>
        node.PreviewName?.toLowerCase().includes(lowerSearch) ||
        node.OriginalName?.toLowerCase().includes(lowerSearch) ||
        node.Protocol?.toLowerCase().includes(lowerSearch) ||
        node.Group?.toLowerCase().includes(lowerSearch) ||
        node.Tags?.toLowerCase().includes(lowerSearch)
    );
  }, [data?.Nodes, searchText]);

  // 当前显示的节点（切片）
  const displayedNodes = useMemo(() => {
    return filteredNodes.slice(0, displayCount);
  }, [filteredNodes, displayCount]);

  // 节点测试统计信息
  const nodeStats = useMemo(() => {
    if (!data?.Nodes || data.Nodes.length === 0) {
      return {
        delayPassCount: 0,
        speedPassCount: 0,
        lowestDelayNode: null,
        highestSpeedNode: null
      };
    }

    const nodes = data.Nodes;

    // 延迟测试通过的节点（DelayStatus 不是 timeout/error 且 DelayTime > 0）
    const delayPassNodes = nodes.filter((node) => {
      const status = node.DelayStatus;
      const isError = status === 'timeout' || status === 'error' || status === 2 || status === 3;
      return !isError && node.DelayTime > 0;
    });

    // 速度测试通过的节点（SpeedStatus 不是 timeout/error 且 Speed > 0）
    const speedPassNodes = nodes.filter((node) => {
      const status = node.SpeedStatus;
      const isError = status === 'timeout' || status === 'error' || status === 2 || status === 3;
      return !isError && node.Speed > 0;
    });

    // 延迟最低的节点（需要速度 > 0，以保证节点可用）
    const validNodesForDelay = nodes.filter((node) => {
      const delayStatus = node.DelayStatus;
      const speedStatus = node.SpeedStatus;
      const isDelayError = delayStatus === 'timeout' || delayStatus === 'error' || delayStatus === 2 || delayStatus === 3;
      const isSpeedError = speedStatus === 'timeout' || speedStatus === 'error' || speedStatus === 2 || speedStatus === 3;
      return !isDelayError && !isSpeedError && node.DelayTime > 0 && node.Speed > 0;
    });

    let lowestDelayNode = null;
    if (validNodesForDelay.length > 0) {
      lowestDelayNode = validNodesForDelay.reduce((min, node) => (node.DelayTime < min.DelayTime ? node : min));
    }

    // 速度最高的节点
    let highestSpeedNode = null;
    if (speedPassNodes.length > 0) {
      highestSpeedNode = speedPassNodes.reduce((max, node) => (node.Speed > max.Speed ? node : max));
    }

    return {
      delayPassCount: delayPassNodes.length,
      speedPassCount: speedPassNodes.length,
      lowestDelayNode,
      highestSpeedNode
    };
  }, [data?.Nodes]);

  // 是否还有更多节点可加载
  const hasMore = displayCount < filteredNodes.length;

  // 加载更多
  const loadMore = useCallback(() => {
    setDisplayCount((prev) => Math.min(prev + BATCH_SIZE, filteredNodes.length));
  }, [filteredNodes.length]);

  // 滚动检测 - 无限滚动
  const handleScroll = useCallback(
    (e) => {
      const { scrollTop, scrollHeight, clientHeight } = e.target;
      // 距离底部 200px 时触发加载
      if (scrollHeight - scrollTop - clientHeight < 200 && hasMore) {
        loadMore();
      }
    },
    [hasMore, loadMore]
  );

  // 打开详情面板
  const handleViewDetails = (node) => {
    setSelectedNode(node);
    setDetailsOpen(true);
  };

  // 关闭详情面板
  const handleCloseDetails = () => {
    setDetailsOpen(false);
    setSelectedNode(null);
  };

  // 查看IP详情
  const handleViewIP = (ip) => {
    setSelectedIP(ip);
    setIpDialogOpen(true);
  };

  // 骨架屏加载状态
  const renderSkeletons = () => (
    <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(140px, 1fr))', gap: 1.5 }}>
      {[1, 2, 3, 4, 5, 6, 7, 8, 9, 10].map((i) => (
        <Skeleton
          key={i}
          variant="rounded"
          height={88}
          sx={{
            borderRadius: 2,
            bgcolor: alpha(theme.palette.action.hover, 0.1)
          }}
        />
      ))}
    </Box>
  );

  // 空状态
  const renderEmpty = () => (
    <Box sx={{ textAlign: 'center', py: 8, px: 4 }}>
      <FilterListIcon sx={{ fontSize: 64, color: 'text.disabled', mb: 2 }} />
      <Typography variant="h6" color="text.secondary" gutterBottom>
        无匹配节点
      </Typography>
      <Typography variant="body2" color="text.disabled">
        {searchText ? '没有找到匹配的节点，请尝试其他搜索条件' : '当前过滤条件下没有可用节点'}
      </Typography>
    </Box>
  );

  return (
    <>
      <Dialog
        open={open}
        onClose={onClose}
        fullScreen={isMobile}
        maxWidth="lg"
        fullWidth
        PaperProps={{
          sx: {
            minHeight: isMobile ? '100%' : '80vh',
            maxHeight: isMobile ? '100%' : '90vh',
            borderRadius: isMobile ? 0 : 4
          }
        }}
      >
        {/* 标题栏 */}
        <DialogTitle
          sx={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            borderBottom: '1px solid',
            borderColor: 'divider',
            pb: 2
          }}
        >
          <Stack direction="row" alignItems="center" spacing={2} flexWrap="wrap">
            <Typography variant="h5" fontWeight="bold">
              节点预览
              <Chip size="small" label="Beta" color="error" variant="outlined" sx={{ ml: 1 }} />
            </Typography>
            {!loading && data && (
              <Stack direction="row" alignItems="center" spacing={1} flexWrap="wrap">
                <Chip label={`共 ${data.TotalCount} 个`} size="small" variant="outlined" sx={{ fontWeight: 600 }} />
                {data.TotalCount !== data.FilteredCount && (
                  <>
                    <ArrowForwardIcon sx={{ fontSize: 14, color: 'text.disabled' }} />
                    <Chip label={`过滤后 ${data.FilteredCount} 个`} size="small" color="primary" sx={{ fontWeight: 600 }} />
                  </>
                )}
                {/* 用量信息 */}
                {data.UsageTotal > 0 && (
                  <Chip
                    label={`${formatBytes(data.UsageUpload + data.UsageDownload)} / ${formatBytes(data.UsageTotal)}`}
                    size="small"
                    color="info"
                    variant="outlined"
                    sx={{ fontWeight: 600 }}
                  />
                )}
                {/* 最近到期时间 */}
                {data.UsageExpire > 0 && (
                  <Chip
                    label={`最近到期时间: ${formatExpireDate(data.UsageExpire)}`}
                    size="small"
                    color="warning"
                    variant="outlined"
                    sx={{ fontWeight: 600 }}
                  />
                )}
              </Stack>
            )}
          </Stack>
          <IconButton onClick={onClose} size="medium">
            <CloseIcon />
          </IconButton>
        </DialogTitle>

        <DialogContent sx={{ p: 0, overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
          <Box
            sx={{
              px: 3,
              py: 2,
              borderBottom: '1px solid',
              borderColor: 'divider',
              bgcolor: alpha(theme.palette.background.paper, 0.5),
              flexShrink: 0
            }}
          >
            <Alert severity="error">
              <AlertTitle>温馨提示</AlertTitle>
              本功能为测试版功能还不稳定。
              预览数据仅供参考，以客户端获取到的实际结果为准，目前部分客户端不支持相关协议的节点，所以节点数量会有出入。
            </Alert>
          </Box>
          {/* 搜索栏 */}
          <Box
            sx={{
              px: 3,
              py: 2,
              borderBottom: '1px solid',
              borderColor: 'divider',
              bgcolor: alpha(theme.palette.background.paper, 0.5),
              flexShrink: 0
            }}
          >
            <TextField
              fullWidth
              size="small"
              placeholder="搜索节点名称、协议、分组或标签..."
              value={searchText}
              onChange={(e) => setSearchText(e.target.value)}
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <SearchIcon color="action" />
                  </InputAdornment>
                ),
                sx: { borderRadius: 2 }
              }}
            />
            {searchText && filteredNodes.length !== data?.Nodes?.length && (
              <Typography variant="caption" color="text.secondary" sx={{ mt: 1, display: 'block' }}>
                找到 {filteredNodes.length} 个匹配结果
              </Typography>
            )}
          </Box>

          {/* 节点测试统计信息卡片 */}
          {!loading && data?.Nodes && data.Nodes.length > 0 && (
            <Box
              sx={{
                px: isMobile ? 1.5 : 3,
                py: 1.5,
                borderBottom: '1px solid',
                borderColor: 'divider',
                bgcolor: alpha(theme.palette.background.paper, 0.3),
                flexShrink: 0
              }}
            >
              {/* 统计区域标题和展开/收起按钮 */}
              <Stack
                direction="row"
                alignItems="center"
                justifyContent="space-between"
                onClick={() => setStatsExpanded(!statsExpanded)}
                sx={{
                  cursor: 'pointer',
                  userSelect: 'none',
                  mb: statsExpanded ? 1.5 : 0,
                  transition: 'margin 0.2s ease'
                }}
              >
                <Stack direction="row" alignItems="center" spacing={1}>
                  <EmojiEventsIcon sx={{ fontSize: 18, color: theme.palette.primary.main }} />
                  <Typography variant="subtitle2" fontWeight={600} color="text.primary">
                    节点测试统计
                  </Typography>
                  <Chip
                    label={`${nodeStats.delayPassCount + nodeStats.speedPassCount > 0 ? '有可用节点' : '暂无数据'}`}
                    size="small"
                    color={nodeStats.delayPassCount + nodeStats.speedPassCount > 0 ? 'success' : 'default'}
                    variant="outlined"
                    sx={{ fontSize: 10, height: 20 }}
                  />
                </Stack>
                <IconButton size="small" sx={{ p: 0.5 }}>
                  {statsExpanded ? <KeyboardArrowUpIcon /> : <KeyboardArrowDownIcon />}
                </IconButton>
              </Stack>

              {/* 统计卡片区域（可折叠） */}
              <Collapse in={statsExpanded} timeout="auto">
                <Grid container spacing={isMobile ? 1 : 1.5}>
                  {/* 延迟测试通过 */}
                  <Grid item xs={6} sm={3}>
                    <Card
                      elevation={0}
                      sx={{
                        background: `linear-gradient(135deg, ${alpha(theme.palette.success.main, 0.1)} 0%, ${alpha(theme.palette.success.light, 0.05)} 100%)`,
                        border: '1px solid',
                        borderColor: alpha(theme.palette.success.main, 0.2),
                        borderRadius: 2,
                        transition: 'all 0.2s ease',
                        '&:hover': {
                          borderColor: alpha(theme.palette.success.main, 0.4),
                          transform: 'translateY(-1px)',
                          boxShadow: `0 4px 12px ${alpha(theme.palette.success.main, 0.15)}`
                        }
                      }}
                    >
                      <CardContent sx={{ p: isMobile ? 1.25 : 1.5, '&:last-child': { pb: isMobile ? 1.25 : 1.5 } }}>
                        <Stack direction="row" alignItems="center" spacing={1} mb={0.5}>
                          <AccessTimeIcon sx={{ fontSize: 16, color: theme.palette.success.main }} />
                          <Typography variant="caption" color="text.secondary" fontWeight={500} sx={{ fontSize: isMobile ? 10 : 11 }}>
                            延迟通过
                          </Typography>
                        </Stack>
                        <Typography variant={isMobile ? 'h6' : 'h5'} fontWeight={700} color="success.main">
                          {nodeStats.delayPassCount}
                          <Typography component="span" variant="caption" color="text.secondary" sx={{ ml: 0.5, fontWeight: 400 }}>
                            / {data.Nodes.length}
                          </Typography>
                        </Typography>
                      </CardContent>
                    </Card>
                  </Grid>

                  {/* 速度测试通过 */}
                  <Grid item xs={6} sm={3}>
                    <Card
                      elevation={0}
                      sx={{
                        background: `linear-gradient(135deg, ${alpha(theme.palette.info.main, 0.1)} 0%, ${alpha(theme.palette.info.light, 0.05)} 100%)`,
                        border: '1px solid',
                        borderColor: alpha(theme.palette.info.main, 0.2),
                        borderRadius: 2,
                        transition: 'all 0.2s ease',
                        '&:hover': {
                          borderColor: alpha(theme.palette.info.main, 0.4),
                          transform: 'translateY(-1px)',
                          boxShadow: `0 4px 12px ${alpha(theme.palette.info.main, 0.15)}`
                        }
                      }}
                    >
                      <CardContent sx={{ p: isMobile ? 1.25 : 1.5, '&:last-child': { pb: isMobile ? 1.25 : 1.5 } }}>
                        <Stack direction="row" alignItems="center" spacing={1} mb={0.5}>
                          <SpeedIcon sx={{ fontSize: 16, color: theme.palette.info.main }} />
                          <Typography variant="caption" color="text.secondary" fontWeight={500} sx={{ fontSize: isMobile ? 10 : 11 }}>
                            速度通过
                          </Typography>
                        </Stack>
                        <Typography variant={isMobile ? 'h6' : 'h5'} fontWeight={700} color="info.main">
                          {nodeStats.speedPassCount}
                          <Typography component="span" variant="caption" color="text.secondary" sx={{ ml: 0.5, fontWeight: 400 }}>
                            / {data.Nodes.length}
                          </Typography>
                        </Typography>
                      </CardContent>
                    </Card>
                  </Grid>

                  {/* 延迟最低节点 */}
                  <Grid item xs={6} sm={3}>
                    <Card
                      elevation={0}
                      onClick={() => nodeStats.lowestDelayNode && handleViewDetails(nodeStats.lowestDelayNode)}
                      sx={{
                        background: `linear-gradient(135deg, ${alpha(theme.palette.warning.main, 0.1)} 0%, ${alpha(theme.palette.warning.light, 0.05)} 100%)`,
                        border: '1px solid',
                        borderColor: alpha(theme.palette.warning.main, 0.2),
                        borderRadius: 2,
                        cursor: nodeStats.lowestDelayNode ? 'pointer' : 'default',
                        transition: 'all 0.2s ease',
                        '&:hover': nodeStats.lowestDelayNode
                          ? {
                              borderColor: alpha(theme.palette.warning.main, 0.4),
                              transform: 'translateY(-1px)',
                              boxShadow: `0 4px 12px ${alpha(theme.palette.warning.main, 0.15)}`
                            }
                          : {}
                      }}
                    >
                      <CardContent sx={{ p: isMobile ? 1.25 : 1.5, '&:last-child': { pb: isMobile ? 1.25 : 1.5 } }}>
                        <Stack direction="row" alignItems="center" spacing={1} mb={0.5}>
                          <AccessTimeIcon sx={{ fontSize: 16, color: theme.palette.warning.main }} />
                          <Typography variant="caption" color="text.secondary" fontWeight={500} sx={{ fontSize: isMobile ? 10 : 11 }}>
                            最低延迟
                          </Typography>
                        </Stack>
                        {nodeStats.lowestDelayNode ? (
                          <>
                            <Tooltip
                              title={nodeStats.lowestDelayNode.PreviewName || nodeStats.lowestDelayNode.OriginalName}
                              placement="top"
                              arrow
                            >
                              <Typography
                                variant="body2"
                                fontWeight={600}
                                color="warning.dark"
                                sx={{
                                  overflow: 'hidden',
                                  textOverflow: 'ellipsis',
                                  whiteSpace: 'nowrap',
                                  fontSize: isMobile ? 11 : 12
                                }}
                              >
                                {nodeStats.lowestDelayNode.PreviewName || nodeStats.lowestDelayNode.OriginalName}
                              </Typography>
                            </Tooltip>
                            <Typography variant="caption" color="text.secondary" sx={{ fontSize: isMobile ? 10 : 11 }}>
                              {nodeStats.lowestDelayNode.DelayTime}ms · {nodeStats.lowestDelayNode.Speed?.toFixed(1)}MB/s
                            </Typography>
                          </>
                        ) : (
                          <Typography variant="body2" color="text.disabled" sx={{ fontSize: isMobile ? 11 : 12 }}>
                            暂无数据
                          </Typography>
                        )}
                      </CardContent>
                    </Card>
                  </Grid>

                  {/* 速度最高节点 */}
                  <Grid item xs={6} sm={3}>
                    <Card
                      elevation={0}
                      onClick={() => nodeStats.highestSpeedNode && handleViewDetails(nodeStats.highestSpeedNode)}
                      sx={{
                        background: `linear-gradient(135deg, ${alpha(theme.palette.primary.main, 0.1)} 0%, ${alpha(theme.palette.primary.light, 0.05)} 100%)`,
                        border: '1px solid',
                        borderColor: alpha(theme.palette.primary.main, 0.2),
                        borderRadius: 2,
                        cursor: nodeStats.highestSpeedNode ? 'pointer' : 'default',
                        transition: 'all 0.2s ease',
                        '&:hover': nodeStats.highestSpeedNode
                          ? {
                              borderColor: alpha(theme.palette.primary.main, 0.4),
                              transform: 'translateY(-1px)',
                              boxShadow: `0 4px 12px ${alpha(theme.palette.primary.main, 0.15)}`
                            }
                          : {}
                      }}
                    >
                      <CardContent sx={{ p: isMobile ? 1.25 : 1.5, '&:last-child': { pb: isMobile ? 1.25 : 1.5 } }}>
                        <Stack direction="row" alignItems="center" spacing={1} mb={0.5}>
                          <SpeedIcon sx={{ fontSize: 16, color: theme.palette.primary.main }} />
                          <Typography variant="caption" color="text.secondary" fontWeight={500} sx={{ fontSize: isMobile ? 10 : 11 }}>
                            最高速度
                          </Typography>
                        </Stack>
                        {nodeStats.highestSpeedNode ? (
                          <>
                            <Tooltip
                              title={nodeStats.highestSpeedNode.PreviewName || nodeStats.highestSpeedNode.OriginalName}
                              placement="top"
                              arrow
                            >
                              <Typography
                                variant="body2"
                                fontWeight={600}
                                color="primary.main"
                                sx={{
                                  overflow: 'hidden',
                                  textOverflow: 'ellipsis',
                                  whiteSpace: 'nowrap',
                                  fontSize: isMobile ? 11 : 12
                                }}
                              >
                                {nodeStats.highestSpeedNode.PreviewName || nodeStats.highestSpeedNode.OriginalName}
                              </Typography>
                            </Tooltip>
                            <Typography variant="caption" color="text.secondary" sx={{ fontSize: isMobile ? 10 : 11 }}>
                              {nodeStats.highestSpeedNode.Speed?.toFixed(1)}MB/s · {nodeStats.highestSpeedNode.DelayTime}ms
                            </Typography>
                          </>
                        ) : (
                          <Typography variant="body2" color="text.disabled" sx={{ fontSize: isMobile ? 11 : 12 }}>
                            暂无数据
                          </Typography>
                        )}
                      </CardContent>
                    </Card>
                  </Grid>
                </Grid>
              </Collapse>
            </Box>
          )}

          {/* 可滚动内容区域 */}
          <Box
            ref={contentRef}
            onScroll={handleScroll}
            sx={{
              flex: 1,
              overflow: 'auto',
              px: 1.5,
              py: 1.5
            }}
          >
            {loading ? (
              renderSkeletons()
            ) : filteredNodes.length === 0 ? (
              renderEmpty()
            ) : (
              <>
                <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(140px, 1fr))', gap: 1 }}>
                  {displayedNodes.map((node, index) => (
                    <NodePreviewCard key={index} node={node} onClick={() => handleViewDetails(node)} />
                  ))}
                </Box>

                {/* 加载更多按钮/提示 */}
                {hasMore && (
                  <Box sx={{ textAlign: 'center', py: 2, mt: 1 }}>
                    <Button variant="outlined" size="small" onClick={loadMore} startIcon={<ExpandMoreIcon />} sx={{ borderRadius: 2 }}>
                      加载更多 ({displayedNodes.length}/{filteredNodes.length})
                    </Button>
                  </Box>
                )}

                {/* 已全部加载提示 */}
                {!hasMore && filteredNodes.length > BATCH_SIZE && (
                  <Typography variant="caption" color="text.disabled" sx={{ display: 'block', textAlign: 'center', py: 2 }}>
                    已加载全部 {filteredNodes.length} 个节点
                  </Typography>
                )}
              </>
            )}
          </Box>

          {/* 底部提示 */}
          {!loading && filteredNodes.length > 0 && (
            <Box
              sx={{
                px: 3,
                py: 1.5,
                borderTop: '1px solid',
                borderColor: 'divider',
                bgcolor: alpha(theme.palette.background.paper, 0.5),
                textAlign: 'center',
                flexShrink: 0
              }}
            >
              <Typography variant="caption" color="text.secondary">
                显示 {displayedNodes.length}/{filteredNodes.length} 个节点 · 点击卡片查看详情
              </Typography>
            </Box>
          )}
        </DialogContent>
      </Dialog>

      {/* 节点详情面板 */}
      <NodePreviewDetailsPanel
        open={detailsOpen}
        node={selectedNode}
        tagColorMap={tagColorMap}
        onClose={handleCloseDetails}
        onViewIP={handleViewIP}
      />

      {/* IP详情对话框 */}
      <IPDetailsDialog open={ipDialogOpen} onClose={() => setIpDialogOpen(false)} ip={selectedIP} />
    </>
  );
}

NodePreviewDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  loading: PropTypes.bool,
  data: PropTypes.shape({
    Nodes: PropTypes.array,
    TotalCount: PropTypes.number,
    FilteredCount: PropTypes.number
  }),
  tagColorMap: PropTypes.object,
  onClose: PropTypes.func.isRequired
};
