import PropTypes from 'prop-types';
import { useState, useEffect, useCallback } from 'react';

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

  // 表格总宽度（各列宽度之和），用于 table-layout: fixed 模式
  const totalWidth = Object.values(columnWidths).reduce((sum, w) => sum + (Number(w) || 0), 0);

  // 开始调整列宽
  const handleResizeStart = useCallback(
    (e, columnKey) => {
      e.preventDefault();
      e.stopPropagation();
      setResizing({
        columnKey,
        startX: e.clientX,
        startWidth: columnWidths[columnKey]
      });
    },
    [columnWidths]
  );

  // 处理鼠标移动和释放
  useEffect(() => {
    if (!resizing) return;

    const handleMouseMove = (e) => {
      const delta = e.clientX - resizing.startX;
      const newWidth = Math.max(60, resizing.startWidth + delta);
      onColumnResize(resizing.columnKey, newWidth);
    };

    const handleMouseUp = () => {
      setResizing(null);
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
    };
  }, [resizing, onColumnResize]);

  // 调整手柄组件
  const ResizeHandle = ({ columnKey }) => (
    <Box
      sx={{
        position: 'absolute',
        right: -2,
        top: 0,
        bottom: 0,
        width: 4,
        cursor: 'col-resize',
        zIndex: 2,
        '&:hover': {
          bgcolor: 'primary.main',
          opacity: 0.5
        },
        '&:active': {
          bgcolor: 'primary.main',
          opacity: 0.7
        }
      }}
      onMouseDown={(e) => handleResizeStart(e, columnKey)}
    />
  );

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
    <TableContainer
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
          tableLayout: 'fixed',
          width: '100%',
          minWidth: Math.max(totalWidth, 900),
          '& .MuiTableCell-root': {
            px: 0.75,
            py: 0.75,
            whiteSpace: 'nowrap',
            verticalAlign: 'middle',
            overflow: 'hidden',
            textOverflow: 'ellipsis'
          },
          '& .MuiTableCell-root.actions-cell': {
            overflow: 'visible'
          },
          '& .MuiTableCell-paddingCheckbox': { px: 0.5, py: 0.5, verticalAlign: 'middle' },
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
            <TableCell padding="checkbox" sx={{ width: columnWidths.checkbox, position: 'relative' }} />
            <TableCell sx={{ width: columnWidths.remark, minWidth: 60, position: 'relative' }}>
              <TableSortLabel active={sortBy === 'name'} direction={sortBy === 'name' ? sortOrder : 'asc'} onClick={() => onSort('name')}>
                {t('nodes.table.remark')}
              </TableSortLabel>
              <ResizeHandle columnKey="remark" />
            </TableCell>
            <TableCell sx={{ width: columnWidths.protocol, minWidth: 60, whiteSpace: 'nowrap', position: 'relative' }}>
              <TableSortLabel
                active={sortBy === 'protocol'}
                direction={sortBy === 'protocol' ? sortOrder : 'asc'}
                onClick={() => onSort('protocol')}
              >
                {t('nodes.table.protocol')}
              </TableSortLabel>
              <ResizeHandle columnKey="protocol" />
            </TableCell>
            <TableCell sx={{ width: columnWidths.group, minWidth: 60, position: 'relative' }}>
              <TableSortLabel
                active={sortBy === 'group'}
                direction={sortBy === 'group' ? sortOrder : 'asc'}
                onClick={() => onSort('group')}
              >
                {t('nodes.table.group')}
              </TableSortLabel>
              <ResizeHandle columnKey="group" />
            </TableCell>
            <TableCell sx={{ width: columnWidths.source, minWidth: 60, position: 'relative' }}>
              <TableSortLabel
                active={sortBy === 'source'}
                direction={sortBy === 'source' ? sortOrder : 'asc'}
                onClick={() => onSort('source')}
              >
                {t('nodes.table.source')}
              </TableSortLabel>
              <ResizeHandle columnKey="source" />
            </TableCell>
            <TableCell sx={{ width: columnWidths.tags, minWidth: 60, whiteSpace: 'nowrap', position: 'relative' }}>
              {t('nodes.table.tags')}
              <ResizeHandle columnKey="tags" />
            </TableCell>
            <TableCell sx={{ width: columnWidths.country, minWidth: 60, whiteSpace: 'nowrap', position: 'relative' }}>
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
              sx={{ width: columnWidths.delay, minWidth: 120, position: 'relative' }}
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
              sx={{ width: columnWidths.speed, minWidth: 120, position: 'relative' }}
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
            <TableCell sx={{ width: columnWidths.ipFeatures, minWidth: 180, whiteSpace: 'nowrap', position: 'relative' }}>
              {t('nodes.table.ipFeatures')}
              <ResizeHandle columnKey="ipFeatures" />
            </TableCell>
            <TableCell align="right" sx={{ width: columnWidths.actions, minWidth: 60, pr: 0.5, position: 'relative' }}>
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
                              <Chip label={ipTypeDisplay.label} color={ipTypeDisplay.color} variant={ipTypeDisplay.variant} size="small" />
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
                <TableCell align="right" className="actions-cell" sx={{ pr: 0.5 }}>
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
