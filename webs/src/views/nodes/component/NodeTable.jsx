import PropTypes from 'prop-types';
import { useState, useEffect, useCallback, useRef } from 'react';

// material-ui
import { alpha, useTheme } from '@mui/material/styles';
import Box from '@mui/material/Box';
import Checkbox from '@mui/material/Checkbox';
import Chip from '@mui/material/Chip';
import IconButton from '@mui/material/IconButton';
import Paper from '@mui/material/Paper';
import Stack from '@mui/material/Stack';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import TableSortLabel from '@mui/material/TableSortLabel';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { useTranslation } from 'react-i18next';

// icons
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import DeleteIcon from '@mui/icons-material/Delete';
import EditIcon from '@mui/icons-material/Edit';
import LockOpenIcon from '@mui/icons-material/LockOpen';
import SpeedIcon from '@mui/icons-material/Speed';

// project imports
import NodeProtocolChip from './NodeProtocolChip';

// utils
import {
  formatDateTime,
  getCountryDisplay,
  getDelayDisplay,
  getFraudScoreDisplay,
  getIpTypeDisplay,
  getNodeUnlockSummaryDisplay,
  getQualityStatusDisplay,
  getResidentialDisplay,
  getSpeedDisplay
} from '../utils';
import { getNodeTagChipSx, getNodeThemeTokens } from '../nodeTheme';

/**
 * 桌面端节点表格（精简版）
 * 只显示核心信息，详细信息通过详情面板查看
 */
export default function NodeTable({
  nodes,
  selectedNodes,
  sortBy,
  sortOrder,
  tagColorMap,
  protocolMeta,
  columnWidths,
  onSelect,
  onSort,
  onSpeedTest,
  onCopy,
  onEdit,
  onDelete,
  onViewDetails,
  onColumnResize
}) {
  const theme = useTheme();
  const { t } = useTranslation();
  const { isDark } = useResolvedColorScheme();
  const tokens = getNodeThemeTokens(theme, isDark);
  const isSelected = (node) => selectedNodes.some((n) => n.ID === node.ID);

  // 列宽调整状态
  const [resizing, setResizing] = useState(null);
  const [ghostLine, setGhostLine] = useState(null); // { x: number, columnKey: string }
  const tableContainerRef = useRef(null);
  const rafRef = useRef(null);

  // 开始调整列宽
  const handleResizeStart = useCallback(
    (e, columnKey) => {
      e.preventDefault();
      e.stopPropagation();

      // 获取表格容器的位置信息
      const tableContainer = tableContainerRef.current;
      const tableRect = tableContainer?.getBoundingClientRect();

      setResizing({
        columnKey,
        startX: e.clientX,
        startWidth: columnWidths[columnKey],
        tableTop: tableRect?.top || 0,
        tableBottom: tableRect?.bottom || window.innerHeight
      });
      setGhostLine({ x: e.clientX, columnKey });
    },
    [columnWidths]
  );

  // 处理鼠标移动和释放
  useEffect(() => {
    if (!resizing) return;

    const handleMouseMove = (e) => {
      // 使用 requestAnimationFrame 优化性能
      if (rafRef.current) {
        cancelAnimationFrame(rafRef.current);
      }

      rafRef.current = requestAnimationFrame(() => {
        const delta = e.clientX - resizing.startX;
        const newWidth = Math.max(60, resizing.startWidth + delta);

        // 如果达到最小宽度限制，ghost line 停在最小位置
        const ghostX = resizing.startX + Math.max(60 - resizing.startWidth, delta);

        setGhostLine({
          x: ghostX,
          columnKey: resizing.columnKey,
          newWidth
        });
      });
    };

    const handleMouseUp = () => {
      if (rafRef.current) {
        cancelAnimationFrame(rafRef.current);
        rafRef.current = null;
      }

      if (ghostLine?.newWidth && ghostLine.newWidth !== resizing.startWidth) {
        onColumnResize(resizing.columnKey, ghostLine.newWidth);
      }

      setResizing(null);
      setGhostLine(null);
    };

    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp);
    document.body.style.cursor = 'col-resize';
    document.body.style.userSelect = 'none';

    return () => {
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
      document.body.style.cursor = '';
      document.body.style.userSelect = '';
      if (rafRef.current) {
        cancelAnimationFrame(rafRef.current);
      }
    };
  }, [resizing, ghostLine, onColumnResize]);

  // 调整手柄组件
  const ResizeHandle = ({ columnKey }) => {
    const isResizing = resizing?.columnKey === columnKey;

    return (
      <Box
        sx={{
          position: 'absolute',
          right: -8, // 居中 16px 宽度的可点击区域
          top: 0,
          bottom: 0,
          width: 16, // 更宽的可点击区域
          cursor: 'col-resize',
          zIndex: 2,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          // 默认状态显示淡色分隔线
          '&::before': {
            content: '""',
            width: '2px',
            height: isResizing ? '100%' : '60%',
            bgcolor: isResizing ? 'primary.main' : tokens.softBorder,
            borderRadius: '1px',
            transition: isResizing ? 'none' : 'all 0.15s ease',
            boxShadow: isResizing ? `0 0 6px ${alpha(theme.palette.primary.main, 0.5)}` : 'none'
          },
          '&:hover::before': {
            height: '80%',
            width: '3px',
            bgcolor: 'primary.main',
            boxShadow: `0 0 4px ${alpha(theme.palette.primary.main, 0.3)}`
          },
          // hover 时提升 z-index 确保在最上层
          '&:hover': {
            zIndex: 3
          }
        }}
        onMouseDown={(e) => handleResizeStart(e, columnKey)}
      />
    );
  };

  const getTableRowSx = (selected = false) => ({
    bgcolor: selected ? tokens.selectedSurface : 'transparent',
    cursor: 'pointer',
    transition: 'background-color 0.2s ease',
    '&:hover': {
      bgcolor: selected ? tokens.selectedHoverSurface : tokens.hoverSurface
    },
    '& td, & .MuiTableCell-root': {
      borderBottomColor: tokens.softBorder
    }
  });

  return (
    <>
      {/* Ghost line - 跟随鼠标的幽灵线 */}
      {ghostLine && (
        <Box
          sx={{
            position: 'fixed',
            left: ghostLine.x,
            top: resizing?.tableTop || 0,
            bottom: `calc(100vh - ${resizing?.tableBottom || window.innerHeight}px)`,
            width: '2px',
            bgcolor: theme.palette.primary.main,
            opacity: 0.8,
            zIndex: 9999,
            pointerEvents: 'none',
            boxShadow: `0 0 8px ${alpha(theme.palette.primary.main, 0.6)}`,
            transition: 'opacity 0.15s ease'
          }}
        />
      )}

      <TableContainer
        ref={tableContainerRef}
        component={Paper}
        sx={{
          bgcolor: tokens.cardSurface,
          backgroundImage: 'none',
          border: '1px solid',
          borderColor: tokens.softBorder,
          boxShadow: tokens.isDark
            ? `0 12px 24px ${alpha(theme.palette.common.black, 0.16)}, inset 0 1px 0 ${alpha(theme.palette.common.white, 0.03)}`
            : `0 6px 18px ${alpha(theme.palette.common.black, 0.06)}`,
          borderRadius: 2.5,
          width: '100%',
          overflowX: 'auto',
          overflowY: 'hidden'
        }}
      >
        <Table
          size="small"
          sx={{
            tableLayout: 'auto',
            width: '100%',
            minWidth: 900,
            '& .MuiTableCell-root': {
              px: 0.75,
              py: 0.75,
              whiteSpace: 'nowrap',
              verticalAlign: 'middle',
              overflow: 'hidden',
              textOverflow: 'ellipsis'
            },
            '& .MuiTableCell-root.actions-cell': {
              overflow: 'visible',
              position: 'sticky',
              right: 0,
              bgcolor: tokens.cardSurface,
              zIndex: 1,
              boxShadow: `-2px 0 4px ${alpha(tokens.isDark ? theme.palette.common.black : theme.palette.common.black, 0.08)}`
            },
            '& .MuiTableCell-paddingCheckbox': { px: 0.5, py: 0.5, verticalAlign: 'middle', width: 48 },
            '& .MuiTableCell-paddingCheckbox .MuiCheckbox-root': { p: 0.5, display: 'flex', alignItems: 'center' },
            '& .MuiChip-root': { height: 22 },
            '& .MuiChip-label': { px: 0.75 },
            '& .MuiIconButton-root': { p: 0.5 }
          }}
        >
          <TableHead
            sx={{
              bgcolor: tokens.cardSurface,
              '& .MuiTableCell-root': {
                color: tokens.primaryText,
                fontWeight: 600,
                fontSize: '0.75rem',
                borderBottomColor: tokens.softBorder
              }
            }}
          >
            <TableRow>
              <TableCell padding="checkbox" sx={{ position: 'relative' }} />
              <TableCell
                sx={{
                  width: columnWidths.remark,
                  minWidth: 60,
                  position: 'relative',
                  bgcolor: resizing?.columnKey === 'remark' ? alpha(theme.palette.primary.main, 0.08) : 'transparent',
                  transition: 'background-color 0.15s ease'
                }}
              >
                <TableSortLabel active={sortBy === 'name'} direction={sortBy === 'name' ? sortOrder : 'asc'} onClick={() => onSort('name')}>
                  {t('nodes.table.remark')}
                </TableSortLabel>
                <ResizeHandle columnKey="remark" />
              </TableCell>
              <TableCell
                sx={{
                  width: columnWidths.protocol,
                  minWidth: 60,
                  whiteSpace: 'nowrap',
                  position: 'relative',
                  bgcolor: resizing?.columnKey === 'protocol' ? alpha(theme.palette.primary.main, 0.08) : 'transparent',
                  transition: 'background-color 0.15s ease'
                }}
              >
                <TableSortLabel
                  active={sortBy === 'protocol'}
                  direction={sortBy === 'protocol' ? sortOrder : 'asc'}
                  onClick={() => onSort('protocol')}
                >
                  {t('nodes.table.protocol')}
                </TableSortLabel>
                <ResizeHandle columnKey="protocol" />
              </TableCell>
              <TableCell
                sx={{
                  width: columnWidths.group,
                  minWidth: 60,
                  position: 'relative',
                  bgcolor: resizing?.columnKey === 'group' ? alpha(theme.palette.primary.main, 0.08) : 'transparent',
                  transition: 'background-color 0.15s ease'
                }}
              >
                <TableSortLabel
                  active={sortBy === 'group'}
                  direction={sortBy === 'group' ? sortOrder : 'asc'}
                  onClick={() => onSort('group')}
                >
                  {t('nodes.table.group')}
                </TableSortLabel>
                <ResizeHandle columnKey="group" />
              </TableCell>
              <TableCell
                sx={{
                  width: columnWidths.source,
                  minWidth: 60,
                  position: 'relative',
                  bgcolor: resizing?.columnKey === 'source' ? alpha(theme.palette.primary.main, 0.08) : 'transparent',
                  transition: 'background-color 0.15s ease'
                }}
              >
                <TableSortLabel
                  active={sortBy === 'source'}
                  direction={sortBy === 'source' ? sortOrder : 'asc'}
                  onClick={() => onSort('source')}
                >
                  {t('nodes.table.source')}
                </TableSortLabel>
                <ResizeHandle columnKey="source" />
              </TableCell>
              <TableCell
                sx={{
                  width: columnWidths.tags,
                  minWidth: 60,
                  whiteSpace: 'nowrap',
                  position: 'relative',
                  bgcolor: resizing?.columnKey === 'tags' ? alpha(theme.palette.primary.main, 0.08) : 'transparent',
                  transition: 'background-color 0.15s ease'
                }}
              >
                {t('nodes.table.tags')}
                <ResizeHandle columnKey="tags" />
              </TableCell>
              <TableCell
                sx={{
                  width: columnWidths.country,
                  minWidth: 60,
                  whiteSpace: 'nowrap',
                  position: 'relative',
                  bgcolor: resizing?.columnKey === 'country' ? alpha(theme.palette.primary.main, 0.08) : 'transparent',
                  transition: 'background-color 0.15s ease'
                }}
              >
                <TableSortLabel
                  active={sortBy === 'country'}
                  direction={sortBy === 'country' ? sortOrder : 'asc'}
                  onClick={() => onSort('country')}
                >
                  {t('nodes.table.country')}
                </TableSortLabel>
                <ResizeHandle columnKey="country" />
              </TableCell>
              <TableCell
                sx={{
                  width: columnWidths.delay,
                  minWidth: 60,
                  position: 'relative',
                  bgcolor: resizing?.columnKey === 'delay' ? alpha(theme.palette.primary.main, 0.08) : 'transparent',
                  transition: 'background-color 0.15s ease'
                }}
                sortDirection={sortBy === 'delay' ? sortOrder : false}
              >
                <TableSortLabel
                  active={sortBy === 'delay'}
                  direction={sortBy === 'delay' ? sortOrder : 'asc'}
                  onClick={() => onSort('delay')}
                >
                  {t('nodes.table.delay')}
                </TableSortLabel>
                <ResizeHandle columnKey="delay" />
              </TableCell>
              <TableCell
                sx={{
                  width: columnWidths.speed,
                  minWidth: 60,
                  position: 'relative',
                  bgcolor: resizing?.columnKey === 'speed' ? alpha(theme.palette.primary.main, 0.08) : 'transparent',
                  transition: 'background-color 0.15s ease'
                }}
                sortDirection={sortBy === 'speed' ? sortOrder : false}
              >
                <TableSortLabel
                  active={sortBy === 'speed'}
                  direction={sortBy === 'speed' ? sortOrder : 'asc'}
                  onClick={() => onSort('speed')}
                >
                  {t('nodes.table.speed')}
                </TableSortLabel>
                <ResizeHandle columnKey="speed" />
              </TableCell>
              <TableCell
                sx={{
                  width: columnWidths.ipFeatures,
                  minWidth: 60,
                  whiteSpace: 'nowrap',
                  position: 'relative',
                  bgcolor: resizing?.columnKey === 'ipFeatures' ? alpha(theme.palette.primary.main, 0.08) : 'transparent',
                  transition: 'background-color 0.15s ease'
                }}
              >
                {t('nodes.table.ipFeatures')}
                <ResizeHandle columnKey="ipFeatures" />
              </TableCell>
              <TableCell align="right" className="actions-cell" sx={{ width: 140, minWidth: 140, pr: 0.5 }}>
                {t('nodes.table.actions')}
              </TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {nodes.map((node) => {
              const effectiveName = node.EffectiveName || node.Name || node.LinkName;
              const secondaryName = node.NameMode === 'remark' ? node.LinkName : node.Name;
              const showSecondaryName = secondaryName && secondaryName !== effectiveName;

              return (
                <TableRow
                  key={node.ID}
                  hover
                  selected={isSelected(node)}
                  sx={getTableRowSx(isSelected(node))}
                  onClick={(e) => {
                    // 点击复选框或操作按钮时不触发详情
                    if (e.target.closest('button') || e.target.closest('input[type="checkbox"]')) return;
                    onViewDetails(node);
                  }}
                >
                  <TableCell padding="checkbox">
                    <Checkbox checked={isSelected(node)} onChange={() => onSelect(node)} />
                  </TableCell>
                  <TableCell>
                    <Tooltip title={effectiveName}>
                      <Stack spacing={0.25} sx={{ width: '100%' }}>
                        <Typography
                          variant="body2"
                          fontWeight="medium"
                          sx={{
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                            whiteSpace: 'nowrap'
                          }}
                        >
                          {effectiveName}
                        </Typography>
                        {showSecondaryName && (
                          <Typography variant="caption" sx={{ color: tokens.secondaryText, overflow: 'hidden', textOverflow: 'ellipsis' }}>
                            {node.NameMode === 'remark'
                              ? t('nodes.table.originalName', { name: secondaryName })
                              : t('nodes.table.remarkName', { name: secondaryName })}
                          </Typography>
                        )}
                      </Stack>
                    </Tooltip>
                  </TableCell>
                  <TableCell>
                    <NodeProtocolChip link={node.Link} protocolMeta={protocolMeta} maxWidth="100%" />
                  </TableCell>
                  <TableCell>
                    {node.Group ? (
                      <Tooltip title={node.Group}>
                        <Chip
                          label={node.Group}
                          color="warning"
                          variant="outlined"
                          size="small"
                          sx={{ maxWidth: '100%', '& .MuiChip-label': { overflow: 'hidden', textOverflow: 'ellipsis' } }}
                        />
                      </Tooltip>
                    ) : (
                      <Typography variant="caption" color="text.secondary">
                        {t('nodes.table.ungrouped')}
                      </Typography>
                    )}
                  </TableCell>
                  <TableCell>
                    {node.Source ? (
                      <Tooltip title={node.Source === 'manual' ? t('nodes.table.manualSource') : node.Source}>
                        <Chip
                          label={node.Source === 'manual' ? t('nodes.table.manualSource') : node.Source}
                          color="info"
                          variant="outlined"
                          size="small"
                          sx={{ maxWidth: '100%', '& .MuiChip-label': { overflow: 'hidden', textOverflow: 'ellipsis' } }}
                        />
                      </Tooltip>
                    ) : (
                      <Typography variant="caption" color="text.secondary">
                        {t('nodes.table.manualSource')}
                      </Typography>
                    )}
                  </TableCell>
                  <TableCell>
                    {node.Tags ? (
                      <Box sx={{ display: 'flex', gap: 0.375, flexWrap: 'wrap', width: '100%' }}>
                        {node.Tags.split(',')
                          .filter((t) => t.trim())
                          .map((tag, idx) => {
                            const tagName = tag.trim();
                            const tagColor = tagColorMap?.[tagName] || tokens.palette.primary.main;
                            return (
                              <Chip
                                key={idx}
                                label={tagName}
                                size="small"
                                sx={{ fontSize: '10px', height: 18, ...getNodeTagChipSx(theme, tokens, tagColor) }}
                              />
                            );
                          })}
                      </Box>
                    ) : (
                      <Typography variant="caption" color="text.secondary">
                        -
                      </Typography>
                    )}
                  </TableCell>
                  <TableCell>
                    {(() => {
                      const countryDisplay = getCountryDisplay(node.LinkCountry, { unknownLabel: t('common.unknown') });
                      return (
                        <Chip
                          label={countryDisplay.text}
                          color={countryDisplay.isUnknown ? 'default' : 'secondary'}
                          variant="outlined"
                          size="small"
                        />
                      );
                    })()}
                  </TableCell>
                  <TableCell>
                    <Box sx={{ display: 'flex', flexDirection: 'column', minWidth: 0, alignItems: 'flex-start' }}>
                      {(() => {
                        const d = getDelayDisplay(node.DelayTime, node.DelayStatus);
                        return <Chip label={d.label} color={d.color} variant={d.variant} size="small" sx={{ maxWidth: 'fit-content' }} />;
                      })()}
                      {node.LatencyCheckAt && (
                        <Typography
                          variant="caption"
                          color="text.secondary"
                          sx={{
                            display: 'block',
                            fontSize: '10px',
                            mt: 0.25,
                            lineHeight: 1.2,
                            whiteSpace: 'nowrap',
                            overflow: 'hidden',
                            textOverflow: 'ellipsis'
                          }}
                        >
                          {formatDateTime(node.LatencyCheckAt)}
                        </Typography>
                      )}
                    </Box>
                  </TableCell>
                  <TableCell>
                    <Box sx={{ display: 'flex', flexDirection: 'column', minWidth: 0, alignItems: 'flex-start' }}>
                      {(() => {
                        const s = getSpeedDisplay(node.Speed, node.SpeedStatus);
                        return <Chip label={s.label} color={s.color} variant={s.variant} size="small" sx={{ maxWidth: 'fit-content' }} />;
                      })()}
                      {node.SpeedCheckAt && node.Speed > 0 && (
                        <Typography
                          variant="caption"
                          color="text.secondary"
                          sx={{
                            display: 'block',
                            fontSize: '10px',
                            mt: 0.25,
                            lineHeight: 1.2,
                            whiteSpace: 'nowrap',
                            overflow: 'hidden',
                            textOverflow: 'ellipsis'
                          }}
                        >
                          {formatDateTime(node.SpeedCheckAt)}
                        </Typography>
                      )}
                    </Box>
                  </TableCell>
                  <TableCell>
                    {(() => {
                      const ipTypeDisplay = getIpTypeDisplay(node.IsBroadcast, node.QualityStatus, node.QualityFamily);
                      const residentialDisplay = getResidentialDisplay(node.IsResidential, node.QualityStatus, node.QualityFamily);
                      const fraudScoreDisplay = getFraudScoreDisplay(node.FraudScore, node.QualityStatus, node.QualityFamily);
                      const qualityStatusDisplay = getQualityStatusDisplay(node.QualityStatus, node.QualityFamily);
                      const unlockDisplay = getNodeUnlockSummaryDisplay(node, { limit: 2 });
                      const isUntested =
                        ipTypeDisplay.state === 'untested' &&
                        residentialDisplay.state === 'untested' &&
                        fraudScoreDisplay.state === 'untested';
                      const shouldMergeQualityTags =
                        node.QualityStatus !== 'success' &&
                        ipTypeDisplay.label === residentialDisplay.label &&
                        residentialDisplay.label === fraudScoreDisplay.label;

                      return (
                        <Box sx={{ display: 'flex', gap: 0.375, flexWrap: 'wrap', minWidth: 0, maxWidth: '100%' }}>
                          {isUntested ? (
                            <Chip label={t('nodes.table.untested')} color="default" variant="outlined" size="small" />
                          ) : shouldMergeQualityTags ? (
                            qualityStatusDisplay.tooltip ? (
                              <Tooltip title={qualityStatusDisplay.tooltip}>
                                <Chip
                                  label={qualityStatusDisplay.label}
                                  color={qualityStatusDisplay.color}
                                  variant={qualityStatusDisplay.variant}
                                  size="small"
                                />
                              </Tooltip>
                            ) : (
                              <Chip
                                label={qualityStatusDisplay.label}
                                color={qualityStatusDisplay.color}
                                variant={qualityStatusDisplay.variant}
                                size="small"
                              />
                            )
                          ) : (
                            <>
                              {ipTypeDisplay.tooltip ? (
                                <Tooltip title={ipTypeDisplay.tooltip}>
                                  <Chip
                                    label={ipTypeDisplay.label}
                                    color={ipTypeDisplay.color}
                                    variant={ipTypeDisplay.variant}
                                    size="small"
                                  />
                                </Tooltip>
                              ) : (
                                <Chip
                                  label={ipTypeDisplay.label}
                                  color={ipTypeDisplay.color}
                                  variant={ipTypeDisplay.variant}
                                  size="small"
                                />
                              )}
                              {residentialDisplay.tooltip ? (
                                <Tooltip title={residentialDisplay.tooltip}>
                                  <Chip
                                    label={residentialDisplay.label}
                                    color={residentialDisplay.color}
                                    variant={residentialDisplay.variant}
                                    size="small"
                                  />
                                </Tooltip>
                              ) : (
                                <Chip
                                  label={residentialDisplay.label}
                                  color={residentialDisplay.color}
                                  variant={residentialDisplay.variant}
                                  size="small"
                                />
                              )}
                              {fraudScoreDisplay.tooltip ? (
                                <Tooltip title={fraudScoreDisplay.tooltip}>
                                  <Chip
                                    label={
                                      node.QualityStatus === 'success'
                                        ? fraudScoreDisplay.label
                                        : fraudScoreDisplay.detailLabel || fraudScoreDisplay.label
                                    }
                                    color={fraudScoreDisplay.color}
                                    variant={fraudScoreDisplay.variant}
                                    size="small"
                                    sx={fraudScoreDisplay.sx}
                                  />
                                </Tooltip>
                              ) : (
                                <Chip
                                  label={
                                    node.QualityStatus === 'success'
                                      ? fraudScoreDisplay.label
                                      : fraudScoreDisplay.detailLabel || fraudScoreDisplay.label
                                  }
                                  color={fraudScoreDisplay.color}
                                  variant={fraudScoreDisplay.variant}
                                  size="small"
                                  sx={fraudScoreDisplay.sx}
                                />
                              )}
                            </>
                          )}
                          {unlockDisplay?.compactItems.map((item) => {
                            const chip = (
                              <Chip
                                key={`unlock-${item.provider}`}
                                icon={<LockOpenIcon sx={{ fontSize: '12px !important' }} />}
                                label={item.compactLabel}
                                color={item.color}
                                variant={item.variant}
                                size="small"
                              />
                            );
                            return item.tooltip ? (
                              <Tooltip key={`unlock-tip-${item.provider}`} title={item.tooltip}>
                                {chip}
                              </Tooltip>
                            ) : (
                              chip
                            );
                          })}
                          {unlockDisplay?.extraCount > 0 && (
                            <Chip label={`+${unlockDisplay.extraCount}`} color="default" variant="outlined" size="small" />
                          )}
                        </Box>
                      );
                    })()}
                  </TableCell>
                  <TableCell align="right" className="actions-cell">
                    <Tooltip title={t('nodes.table.speedTest')}>
                      <IconButton size="small" onClick={() => onSpeedTest(node)}>
                        <SpeedIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title={t('nodes.table.copyLink')}>
                      <IconButton size="small" onClick={() => onCopy(node.Link)}>
                        <ContentCopyIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title={t('nodes.table.edit')}>
                      <IconButton size="small" onClick={() => onEdit(node)}>
                        <EditIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title={t('nodes.table.delete')}>
                      <IconButton size="small" color="error" onClick={() => onDelete(node)}>
                        <DeleteIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </TableContainer>
    </>
  );
}

NodeTable.propTypes = {
  nodes: PropTypes.array.isRequired,
  selectedNodes: PropTypes.array.isRequired,
  sortBy: PropTypes.string.isRequired,
  sortOrder: PropTypes.string.isRequired,
  tagColorMap: PropTypes.object,
  protocolMeta: PropTypes.array,
  columnWidths: PropTypes.object.isRequired,
  onSelect: PropTypes.func.isRequired,
  onSort: PropTypes.func.isRequired,
  onSpeedTest: PropTypes.func.isRequired,
  onCopy: PropTypes.func.isRequired,
  onEdit: PropTypes.func.isRequired,
  onDelete: PropTypes.func.isRequired,
  onViewDetails: PropTypes.func.isRequired,
  onColumnResize: PropTypes.func.isRequired
};
