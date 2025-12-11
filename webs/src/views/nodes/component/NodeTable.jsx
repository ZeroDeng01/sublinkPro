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

// utils
import { formatDateTime, formatCountry, getDelayColor } from '../utils';

/**
 * 桌面端节点表格
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
  onDelete
}) {
  const isSelected = (node) => selectedNodes.some((n) => n.ID === node.ID);
  // 后端分页：nodes 已经是当前页数据，无需客户端切片
  // const paginatedNodes = nodes.slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage);

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
            <TableCell>备注</TableCell>
            <TableCell>分组</TableCell>
            <TableCell>来源</TableCell>
            <TableCell sx={{ minWidth: 100, whiteSpace: 'nowrap' }}>标签</TableCell>
            <TableCell>节点名称</TableCell>
            <TableCell>前置代理</TableCell>
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
            <TableCell sx={{ minWidth: 100, whiteSpace: 'nowrap' }}>国家</TableCell>
            <TableCell sx={{ minWidth: 160, whiteSpace: 'nowrap' }}>创建时间</TableCell>
            <TableCell sx={{ minWidth: 160, whiteSpace: 'nowrap' }}>更新时间</TableCell>
            <TableCell align="right">操作</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {nodes.map((node) => (
            <TableRow key={node.ID} hover selected={isSelected(node)}>
              <TableCell padding="checkbox">
                <Checkbox checked={isSelected(node)} onChange={() => onSelect(node)} />
              </TableCell>
              <TableCell>
                <Tooltip title={node.Name}>
                  <Chip
                    label={node.Name}
                    color="success"
                    variant="outlined"
                    size="small"
                    sx={{ maxWidth: '150px', '& .MuiChip-label': { overflow: 'hidden', textOverflow: 'ellipsis' } }}
                  />
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
                <Chip
                  label={node.Source === 'manual' ? '手动添加' : node.Source}
                  color={node.Source === 'manual' ? 'success' : 'warning'}
                  variant="outlined"
                  size="small"
                />
              </TableCell>
              <TableCell>
                {node.Tags ? (
                  <Box sx={{ display: 'flex', gap: 0.5, flexWrap: 'wrap', maxWidth: 150 }}>
                    {node.Tags.split(',')
                      .filter((t) => t.trim())
                      .slice(0, 3)
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
                    {node.Tags.split(',').filter((t) => t.trim()).length > 3 && (
                      <Chip
                        label={`+${node.Tags.split(',').filter((t) => t.trim()).length - 3}`}
                        size="small"
                        sx={{ fontSize: '10px', height: 20 }}
                      />
                    )}
                  </Box>
                ) : (
                  <Typography variant="caption" color="textSecondary">
                    -
                  </Typography>
                )}
              </TableCell>
              <TableCell>
                <Tooltip title={node.LinkName || ''}>
                  <Typography sx={{ maxWidth: 200, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                    {node.LinkName || '-'}
                  </Typography>
                </Tooltip>
              </TableCell>
              <TableCell>
                <Tooltip title={node.DialerProxyName || ''}>
                  <Typography
                    sx={{
                      minWidth: 100,
                      maxWidth: 150,
                      whiteSpace: 'nowrap',
                      overflow: 'hidden',
                      textOverflow: 'ellipsis'
                    }}
                  >
                    {node.DialerProxyName || '-'}
                  </Typography>
                </Tooltip>
              </TableCell>
              <TableCell>
                <Box>
                  {node.DelayTime > 0 ? (
                    <Chip label={`${node.DelayTime}ms`} color={getDelayColor(node.DelayTime)} size="small" />
                  ) : node.DelayTime === -1 ? (
                    <Chip label="超时" color="error" size="small" />
                  ) : (
                    <Chip label="未测速" variant="outlined" size="small" />
                  )}
                  {node.LastCheck && (
                    <Typography variant="caption" color="textSecondary" sx={{ display: 'block', fontSize: '10px', mt: 0.5 }}>
                      {formatDateTime(node.LastCheck)}
                    </Typography>
                  )}
                </Box>
              </TableCell>
              <TableCell>{node.Speed > 0 ? `${node.Speed.toFixed(2)}MB/s` : '-'}</TableCell>
              <TableCell>
                {node.LinkCountry ? (
                  <Chip label={formatCountry(node.LinkCountry)} color="secondary" variant="outlined" size="small" />
                ) : (
                  '-'
                )}
              </TableCell>
              <TableCell sx={{ minWidth: 160, whiteSpace: 'nowrap' }}>
                <Typography variant="caption">{formatDateTime(node.CreatedAt)}</Typography>
              </TableCell>
              <TableCell sx={{ minWidth: 160, whiteSpace: 'nowrap' }}>
                <Typography variant="caption">{formatDateTime(node.UpdatedAt)}</Typography>
              </TableCell>
              <TableCell align="right" sx={{ minWidth: 160 }}>
                <Tooltip title="测速">
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
  onDelete: PropTypes.func.isRequired
};
