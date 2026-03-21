import PropTypes from 'prop-types';

// material-ui
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

// icons
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import DeleteIcon from '@mui/icons-material/Delete';
import EditIcon from '@mui/icons-material/Edit';
import SpeedIcon from '@mui/icons-material/Speed';

// utils
import {
  formatDateTime,
  formatCountry,
  getDelayDisplay,
  getFraudScoreDisplay,
  getIpTypeDisplay,
  getResidentialDisplay,
  getSpeedDisplay
} from '../utils';

/**
 * 桌面端节点表格（精简版）
 * 只显示核心信息，详细信息通过详情面板查看
 */
export default function NodeTable({
  nodes,
  page,
  rowsPerPage,
  selectedNodes,
  sortBy,
  sortOrder,
  tagColorMap,
  onSelectAll,
  onSelect,
  onSort,
  onSpeedTest,
  onCopy,
  onEdit,
  onDelete,
  onViewDetails
}) {
  const isSelected = (node) => selectedNodes.some((n) => n.ID === node.ID);
  const denseCellSx = {
    px: 0.75,
    py: 0.75,
    whiteSpace: 'nowrap',
    verticalAlign: 'top'
  };

  return (
    <TableContainer component={Paper}>
      <Table
        size="small"
        sx={{
          '& .MuiTableCell-root': denseCellSx,
          '& .MuiTableCell-paddingCheckbox': { px: 0.5, py: 0.5 },
          '& .MuiChip-root': { height: 22 },
          '& .MuiChip-label': { px: 0.75 },
          '& .MuiIconButton-root': { p: 0.5 }
        }}
      >
        <TableHead>
          <TableRow>
            <TableCell padding="checkbox">
              <Checkbox
                indeterminate={selectedNodes.length > 0 && selectedNodes.length < nodes.length}
                checked={nodes.length > 0 && selectedNodes.length >= nodes.length}
                onChange={onSelectAll}
              />
            </TableCell>
            <TableCell sx={{ minWidth: 132 }}>备注</TableCell>
            <TableCell sx={{ minWidth: 88 }}>分组</TableCell>
            <TableCell sx={{ minWidth: 88 }}>来源</TableCell>
            <TableCell sx={{ minWidth: 92, whiteSpace: 'nowrap' }}>标签</TableCell>
            <TableCell sx={{ minWidth: 64, whiteSpace: 'nowrap' }}>国家</TableCell>
            <TableCell sx={{ minWidth: 168 }} sortDirection={sortBy === 'delay' || sortBy === 'speed' ? sortOrder : false}>
              <Stack direction="row" spacing={1.5} alignItems="center" sx={{ whiteSpace: 'nowrap' }}>
                <TableSortLabel
                  active={sortBy === 'delay'}
                  direction={sortBy === 'delay' ? sortOrder : 'asc'}
                  onClick={() => onSort('delay')}
                >
                  延迟
                </TableSortLabel>
                <TableSortLabel
                  active={sortBy === 'speed'}
                  direction={sortBy === 'speed' ? sortOrder : 'asc'}
                  onClick={() => onSort('speed')}
                >
                  速度
                </TableSortLabel>
              </Stack>
            </TableCell>
            <TableCell sx={{ minWidth: 128, whiteSpace: 'nowrap' }}>IP特征</TableCell>
            <TableCell align="right" sx={{ minWidth: 104, pr: 0.5 }}>
              操作
            </TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {nodes.map((node) => (
            <TableRow
              key={node.ID}
              hover
              selected={isSelected(node)}
              sx={{ cursor: 'pointer' }}
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
                <Tooltip title={node.Name}>
                  <Typography
                    variant="body2"
                    fontWeight="medium"
                    sx={{
                      maxWidth: '180px',
                      overflow: 'hidden',
                      textOverflow: 'ellipsis',
                      whiteSpace: 'nowrap'
                    }}
                  >
                    {node.Name}
                  </Typography>
                </Tooltip>
              </TableCell>
              <TableCell>
                {node.Group ? (
                  <Tooltip title={node.Group}>
                    <Chip
                      label={node.Group}
                      color="warning"
                      variant="outlined"
                      size="small"
                      sx={{ maxWidth: '104px', '& .MuiChip-label': { overflow: 'hidden', textOverflow: 'ellipsis' } }}
                    />
                  </Tooltip>
                ) : (
                  <Typography variant="caption" color="textSecondary">
                    未分组
                  </Typography>
                )}
              </TableCell>
              <TableCell>
                {node.Source ? (
                  <Tooltip title={node.Source === 'manual' ? '手动添加' : node.Source}>
                    <Chip
                      label={node.Source === 'manual' ? '手动添加' : node.Source}
                      color="info"
                      variant="outlined"
                      size="small"
                      sx={{ maxWidth: '104px', '& .MuiChip-label': { overflow: 'hidden', textOverflow: 'ellipsis' } }}
                    />
                  </Tooltip>
                ) : (
                  <Typography variant="caption" color="textSecondary">
                    手动添加
                  </Typography>
                )}
              </TableCell>
              <TableCell>
                {node.Tags ? (
                  <Box sx={{ display: 'flex', gap: 0.375, flexWrap: 'wrap', maxWidth: 180 }}>
                    {node.Tags.split(',')
                      .filter((t) => t.trim())
                      .map((tag, idx) => {
                        const tagName = tag.trim();
                        const tagColor = tagColorMap?.[tagName] || '#1976d2';
                        return (
                          <Chip
                            key={idx}
                            label={tagName}
                            size="small"
                            sx={{
                              fontSize: '10px',
                              height: 18,
                              backgroundColor: tagColor,
                              color: '#fff'
                            }}
                          />
                        );
                      })}
                  </Box>
                ) : (
                  <Typography variant="caption" color="textSecondary">
                    -
                  </Typography>
                )}
              </TableCell>
              <TableCell>
                {node.LinkCountry ? (
                  <Chip label={formatCountry(node.LinkCountry)} color="secondary" variant="outlined" size="small" />
                ) : (
                  '-'
                )}
              </TableCell>
              <TableCell>
                <Stack spacing={0.75} sx={{ minWidth: 0 }}>
                  <Stack direction="row" spacing={0.75} alignItems="flex-start" flexWrap="wrap" useFlexGap>
                    <Box>
                      {(() => {
                        const d = getDelayDisplay(node.DelayTime, node.DelayStatus);
                        return <Chip label={d.label} color={d.color} variant={d.variant} size="small" />;
                      })()}
                      {node.LatencyCheckAt && (
                        <Typography
                          variant="caption"
                          color="textSecondary"
                          sx={{ display: 'block', fontSize: '10px', mt: 0.25, lineHeight: 1.2 }}
                        >
                          {formatDateTime(node.LatencyCheckAt)}
                        </Typography>
                      )}
                    </Box>
                    <Box>
                      {(() => {
                        const s = getSpeedDisplay(node.Speed, node.SpeedStatus);
                        return <Chip label={s.label} color={s.color} variant={s.variant} size="small" />;
                      })()}
                      {node.SpeedCheckAt && node.Speed > 0 && (
                        <Typography
                          variant="caption"
                          color="textSecondary"
                          sx={{ display: 'block', fontSize: '10px', mt: 0.25, lineHeight: 1.2 }}
                        >
                          {formatDateTime(node.SpeedCheckAt)}
                        </Typography>
                      )}
                    </Box>
                  </Stack>
                </Stack>
              </TableCell>
              <TableCell>
                {(() => {
                  const ipTypeDisplay = getIpTypeDisplay(node.IsBroadcast, node.FraudScore);
                  const residentialDisplay = getResidentialDisplay(node.IsResidential, node.FraudScore);
                  const fraudScoreDisplay = getFraudScoreDisplay(node.FraudScore);
                  const isUntested =
                    ipTypeDisplay.label === '未检测' && residentialDisplay.label === '未检测' && fraudScoreDisplay.label === '未检测';

                  return (
                    <Box sx={{ display: 'flex', gap: 0.375, flexWrap: 'wrap', minWidth: 0, maxWidth: 160 }}>
                      {isUntested ? (
                        <Chip label="未检测" color="default" variant="outlined" size="small" />
                      ) : (
                        <>
                          <Chip label={ipTypeDisplay.label} color={ipTypeDisplay.color} variant={ipTypeDisplay.variant} size="small" />
                          <Chip
                            label={residentialDisplay.label}
                            color={residentialDisplay.color}
                            variant={residentialDisplay.variant}
                            size="small"
                          />
                          <Chip
                            label={fraudScoreDisplay.label}
                            color={fraudScoreDisplay.color}
                            variant={fraudScoreDisplay.variant}
                            size="small"
                            sx={fraudScoreDisplay.sx}
                          />
                        </>
                      )}
                    </Box>
                  );
                })()}
              </TableCell>
              <TableCell align="right" sx={{ pr: 0.5 }}>
                <Tooltip title="检测">
                  <IconButton size="small" onClick={() => onSpeedTest(node)}>
                    <SpeedIcon fontSize="small" />
                  </IconButton>
                </Tooltip>
                <Tooltip title="复制链接">
                  <IconButton size="small" onClick={() => onCopy(node.Link)}>
                    <ContentCopyIcon fontSize="small" />
                  </IconButton>
                </Tooltip>
                <Tooltip title="编辑">
                  <IconButton size="small" onClick={() => onEdit(node)}>
                    <EditIcon fontSize="small" />
                  </IconButton>
                </Tooltip>
                <Tooltip title="删除">
                  <IconButton size="small" color="error" onClick={() => onDelete(node)}>
                    <DeleteIcon fontSize="small" />
                  </IconButton>
                </Tooltip>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
}

NodeTable.propTypes = {
  nodes: PropTypes.array.isRequired,
  page: PropTypes.number.isRequired,
  rowsPerPage: PropTypes.number.isRequired,
  selectedNodes: PropTypes.array.isRequired,
  sortBy: PropTypes.string.isRequired,
  sortOrder: PropTypes.string.isRequired,
  tagColorMap: PropTypes.object,
  onSelectAll: PropTypes.func.isRequired,
  onSelect: PropTypes.func.isRequired,
  onSort: PropTypes.func.isRequired,
  onSpeedTest: PropTypes.func.isRequired,
  onCopy: PropTypes.func.isRequired,
  onEdit: PropTypes.func.isRequired,
  onDelete: PropTypes.func.isRequired,
  onViewDetails: PropTypes.func.isRequired
};
