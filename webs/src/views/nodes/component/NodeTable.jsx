import PropTypes from 'prop-types';

// material-ui
import Box from '@mui/material/Box';
import Checkbox from '@mui/material/Checkbox';
import Chip from '@mui/material/Chip';
import IconButton from '@mui/material/IconButton';
import Paper from '@mui/material/Paper';
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
import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined';

// utils
import { formatDateTime, formatCountry, getDelayDisplay, getSpeedDisplay } from '../utils';

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

  return (
    <TableContainer component={Paper}>
      <Table size="small">
        <TableHead>
          <TableRow>
            <TableCell padding="checkbox">
              <Checkbox
                indeterminate={selectedNodes.length > 0 && selectedNodes.length < nodes.length}
                checked={nodes.length > 0 && selectedNodes.length >= nodes.length}
                onChange={onSelectAll}
              />
            </TableCell>
            <TableCell sx={{ minWidth: 150 }}>备注</TableCell>
            <TableCell sx={{ minWidth: 100 }}>分组</TableCell>
            <TableCell sx={{ minWidth: 100 }}>来源</TableCell>
            <TableCell sx={{ minWidth: 100, whiteSpace: 'nowrap' }}>标签</TableCell>
            <TableCell sortDirection={sortBy === 'delay' ? sortOrder : false}>
              <TableSortLabel
                active={sortBy === 'delay'}
                direction={sortBy === 'delay' ? sortOrder : 'asc'}
                onClick={() => onSort('delay')}
              >
                延迟
              </TableSortLabel>
            </TableCell>
            <TableCell sortDirection={sortBy === 'speed' ? sortOrder : false}>
              <TableSortLabel
                active={sortBy === 'speed'}
                direction={sortBy === 'speed' ? sortOrder : 'asc'}
                onClick={() => onSort('speed')}
              >
                速度
              </TableSortLabel>
            </TableCell>
            <TableCell sx={{ minWidth: 80, whiteSpace: 'nowrap' }}>国家</TableCell>
            <TableCell align="right" sx={{ minWidth: 180 }}>
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
                      maxWidth: '200px',
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
                      sx={{ maxWidth: '120px', '& .MuiChip-label': { overflow: 'hidden', textOverflow: 'ellipsis' } }}
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
                      sx={{ maxWidth: '120px', '& .MuiChip-label': { overflow: 'hidden', textOverflow: 'ellipsis' } }}
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
                  <Box sx={{ display: 'flex', gap: 0.5, flexWrap: 'wrap', maxWidth: 200 }}>
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
                              height: 20,
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
                <Box>
                  {(() => {
                    const d = getDelayDisplay(node.DelayTime, node.DelayStatus);
                    return <Chip label={d.label} color={d.color} variant={d.variant} size="small" />;
                  })()}
                  {node.LatencyCheckAt && (
                    <Typography variant="caption" color="textSecondary" sx={{ display: 'block', fontSize: '10px', mt: 0.5 }}>
                      {formatDateTime(node.LatencyCheckAt)}
                    </Typography>
                  )}
                </Box>
              </TableCell>
              <TableCell>
                <Box>
                  {(() => {
                    const s = getSpeedDisplay(node.Speed, node.SpeedStatus);
                    return <Chip label={s.label} color={s.color} variant={s.variant} size="small" />;
                  })()}
                  {node.SpeedCheckAt && node.Speed > 0 && (
                    <Typography variant="caption" color="textSecondary" sx={{ display: 'block', fontSize: '10px', mt: 0.5 }}>
                      {formatDateTime(node.SpeedCheckAt)}
                    </Typography>
                  )}
                </Box>
              </TableCell>
              <TableCell>
                {node.LinkCountry ? (
                  <Chip label={formatCountry(node.LinkCountry)} color="secondary" variant="outlined" size="small" />
                ) : (
                  '-'
                )}
              </TableCell>
              <TableCell align="right">
                <Tooltip title="详情">
                  <IconButton size="small" onClick={() => onViewDetails(node)} color="info">
                    <InfoOutlinedIcon fontSize="small" />
                  </IconButton>
                </Tooltip>
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
