import { Fragment } from 'react';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Chip from '@mui/material/Chip';
import Typography from '@mui/material/Typography';
import IconButton from '@mui/material/IconButton';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';
import Collapse from '@mui/material/Collapse';
import Tooltip from '@mui/material/Tooltip';
import SortableNodeList from './SortableNodeList';

// icons
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import QrCode2Icon from '@mui/icons-material/QrCode2';
import HistoryIcon from '@mui/icons-material/History';
import SortIcon from '@mui/icons-material/Sort';
import CheckIcon from '@mui/icons-material/Check';
import CloseIcon from '@mui/icons-material/Close';
import KeyboardArrowDownIcon from '@mui/icons-material/KeyboardArrowDown';
import KeyboardArrowUpIcon from '@mui/icons-material/KeyboardArrowUp';
import VisibilityIcon from '@mui/icons-material/Visibility';
import AccountTreeIcon from '@mui/icons-material/AccountTree';

/**
 * 桌面端订阅表格组件
 */
export default function SubscriptionTable({
  subscriptions,
  expandedRows,
  sortingSubId,
  tempSortData,
  selectedSortItems = [],
  onToggleRow,
  onClient,
  onLogs,
  onEdit,
  onDelete,
  onCopy,
  onPreview,
  onChainProxy,
  onStartSort,
  onConfirmSort,
  onCancelSort,
  onDragEnd,
  onCopyToClipboard,
  getSortedItems,
  onToggleSortSelect,
  onSelectAllSort,
  onClearSortSelection,
  onBatchSort,
  onBatchMove
}) {
  return (
    <TableContainer component={Paper}>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell width={50} />
            <TableCell>订阅名称</TableCell>
            <TableCell>节点/分组</TableCell>
            <TableCell>创建时间</TableCell>
            <TableCell align="right" width={350}>
              操作
            </TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {subscriptions.map((sub) => (
            <Fragment key={sub.ID}>
              <TableRow hover>
                <TableCell>
                  <IconButton size="small" onClick={() => onToggleRow(sub.ID)}>
                    {expandedRows[sub.ID] ? <KeyboardArrowUpIcon /> : <KeyboardArrowDownIcon />}
                  </IconButton>
                </TableCell>
                <TableCell>
                  <Chip label={sub.Name} color="primary" />
                  {sortingSubId === sub.ID && <Chip label="排序中" color="warning" size="small" sx={{ ml: 1 }} />}
                </TableCell>
                <TableCell>
                  <Typography variant="body2">
                    {sub.Nodes?.length || 0} 个节点, {sub.Groups?.length || 0} 个分组
                  </Typography>
                </TableCell>
                <TableCell>{sub.CreateDate}</TableCell>
                <TableCell align="right">
                  <Stack direction="row" spacing={0.5} justifyContent="flex-end">
                    <Tooltip title="预览节点">
                      <IconButton size="small" color="info" onClick={() => onPreview(sub)}>
                        <VisibilityIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="复制本配置为新副本">
                      <IconButton size="small" color="secondary" onClick={() => onCopy(sub)}>
                        <ContentCopyIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="编辑">
                      <IconButton size="small" onClick={() => onEdit(sub)}>
                        <EditIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="分享管理">
                      <IconButton size="small" onClick={() => onClient(sub)}>
                        <QrCode2Icon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="访问记录">
                      <IconButton size="small" onClick={() => onLogs(sub)}>
                        <HistoryIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="链式代理">
                      <IconButton size="small" color="warning" onClick={() => onChainProxy(sub)}>
                        <AccountTreeIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                    {sortingSubId !== sub.ID ? (
                      <Tooltip title="排序">
                        <IconButton size="small" onClick={() => onStartSort(sub)}>
                          <SortIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                    ) : (
                      <>
                        <Tooltip title="确定">
                          <IconButton size="small" color="success" onClick={() => onConfirmSort(sub)}>
                            <CheckIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="取消">
                          <IconButton size="small" onClick={onCancelSort}>
                            <CloseIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                      </>
                    )}
                    <Tooltip title="删除">
                      <IconButton size="small" color="error" onClick={() => onDelete(sub)}>
                        <DeleteIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                  </Stack>
                </TableCell>
              </TableRow>
              <TableRow>
                <TableCell style={{ paddingBottom: 0, paddingTop: 0 }} colSpan={6}>
                  <Collapse in={expandedRows[sub.ID] || sortingSubId === sub.ID} timeout="auto" unmountOnExit>
                    <Box sx={{ margin: 2 }}>
                      {sortingSubId === sub.ID ? (
                        <SortableNodeList
                          items={tempSortData}
                          onDragEnd={onDragEnd}
                          selectedItems={selectedSortItems}
                          onToggleSelect={onToggleSortSelect}
                          onSelectAll={onSelectAllSort}
                          onClearSelection={onClearSortSelection}
                          onBatchSort={onBatchSort}
                          onBatchMove={onBatchMove}
                        />
                      ) : (
                        <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                          {getSortedItems(sub).map((item, idx) =>
                            item._type === 'node' ? (
                              <Chip
                                key={item._type + item.ID}
                                label={item.Name}
                                size="small"
                                variant="outlined"
                                color="success"
                                onClick={() => onCopyToClipboard(item.Link)}
                              />
                            ) : (
                              <Chip key={item._type + idx} label={`📁 ${item.Name}`} size="small" variant="outlined" color="warning" />
                            )
                          )}
                        </Stack>
                      )}
                    </Box>
                  </Collapse>
                </TableCell>
              </TableRow>
            </Fragment>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
}
