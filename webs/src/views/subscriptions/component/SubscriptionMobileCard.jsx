import { useState } from 'react';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Chip from '@mui/material/Chip';
import Typography from '@mui/material/Typography';
import Divider from '@mui/material/Divider';
import IconButton from '@mui/material/IconButton';
import Button from '@mui/material/Button';
import Collapse from '@mui/material/Collapse';
import Menu from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';

import MainCard from 'ui-component/cards/MainCard';
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
import MoreVertIcon from '@mui/icons-material/MoreVert';

/**
 * 移动端订阅卡片组件
 * 优化触摸交互：常用操作使用大按钮，其余放入更多菜单
 */
export default function SubscriptionMobileCard({
  subscriptions,
  expandedRows,
  sortingSubId,
  tempSortData,
  selectedSortItems = [],
  theme,
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
  // 更多菜单状态
  const [menuAnchor, setMenuAnchor] = useState(null);
  const [menuSubId, setMenuSubId] = useState(null);

  const handleOpenMenu = (event, subId) => {
    event.stopPropagation();
    setMenuAnchor(event.currentTarget);
    setMenuSubId(subId);
  };

  const handleCloseMenu = () => {
    setMenuAnchor(null);
    setMenuSubId(null);
  };

  const handleMenuAction = (action, sub) => {
    handleCloseMenu();
    action(sub);
  };

  return (
    <>
      <Stack spacing={2}>
        {subscriptions.map((sub) => (
          <MainCard key={sub.ID} content={false} border shadow={theme.shadows[1]}>
            <Box p={2}>
              {/* 头部：订阅名称和展开按钮 */}
              <Stack
                direction="row"
                justifyContent="space-between"
                alignItems="center"
                mb={1.5}
                onClick={() => onToggleRow(sub.ID)}
                sx={{ cursor: 'pointer' }}
              >
                <Stack direction="row" spacing={1} alignItems="center" sx={{ flex: 1 }}>
                  <Chip label={sub.Name} color="primary" sx={{ maxWidth: '70%' }} />
                  {sortingSubId === sub.ID && <Chip label="排序中" color="warning" size="small" />}
                </Stack>
                <IconButton size="small">
                  {expandedRows[sub.ID] || sortingSubId === sub.ID ? <KeyboardArrowUpIcon /> : <KeyboardArrowDownIcon />}
                </IconButton>
              </Stack>

              {/* 统计信息 */}
              <Typography variant="body2" color="text.secondary" sx={{ mb: 1.5 }}>
                {sub.Nodes?.length || 0} 个节点, {sub.Groups?.length || 0} 个分组
              </Typography>

              <Divider sx={{ mb: 1.5 }} />

              {/* 操作区域 - 移动端优化 */}
              {sortingSubId === sub.ID ? (
                // 排序模式：显示确认/取消按钮
                <Stack direction="row" spacing={1} justifyContent="flex-end">
                  <Button
                    variant="outlined"
                    size="small"
                    startIcon={<CloseIcon />}
                    onClick={(e) => {
                      e.stopPropagation();
                      onCancelSort();
                    }}
                  >
                    取消
                  </Button>
                  <Button
                    variant="contained"
                    size="small"
                    color="success"
                    startIcon={<CheckIcon />}
                    onClick={(e) => {
                      e.stopPropagation();
                      onConfirmSort(sub);
                    }}
                  >
                    确认排序
                  </Button>
                </Stack>
              ) : (
                // 正常模式：显示操作按钮
                <Stack direction="row" justifyContent="space-between" alignItems="center">
                  <Typography variant="caption" color="text.secondary">
                    {sub.CreateDate}
                  </Typography>

                  {/* 快捷操作按钮 - 使用较大的图标 */}
                  <Stack direction="row" spacing={0.5} alignItems="center">
                    {/* 预览 - 常用 */}
                    <IconButton
                      size="medium"
                      color="info"
                      onClick={(e) => {
                        e.stopPropagation();
                        onPreview(sub);
                      }}
                      sx={{ p: 1.5 }}
                    >
                      <VisibilityIcon />
                    </IconButton>

                    {/* 客户端链接 - 常用 */}
                    <IconButton
                      size="medium"
                      color="primary"
                      onClick={(e) => {
                        e.stopPropagation();
                        onClient(sub);
                      }}
                      sx={{ p: 1.5 }}
                    >
                      <QrCode2Icon />
                    </IconButton>

                    {/* 编辑 - 常用 */}
                    <IconButton
                      size="medium"
                      onClick={(e) => {
                        e.stopPropagation();
                        onEdit(sub);
                      }}
                      sx={{ p: 1.5 }}
                    >
                      <EditIcon />
                    </IconButton>

                    {/* 更多操作菜单 */}
                    <IconButton size="medium" onClick={(e) => handleOpenMenu(e, sub.ID)} sx={{ p: 1.5 }}>
                      <MoreVertIcon />
                    </IconButton>
                  </Stack>
                </Stack>
              )}

              {/* 可展开内容 */}
              <Collapse in={expandedRows[sub.ID] || sortingSubId === sub.ID} timeout="auto" unmountOnExit>
                <Box sx={{ mt: 2 }}>
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
                            sx={{ mb: 1 }}
                          />
                        ) : (
                          <Chip
                            key={item._type + idx}
                            label={`📁 ${item.Name}`}
                            size="small"
                            variant="outlined"
                            color="warning"
                            sx={{ mb: 1 }}
                          />
                        )
                      )}
                    </Stack>
                  )}
                </Box>
              </Collapse>
            </Box>
          </MainCard>
        ))}
      </Stack>

      {/* 更多操作菜单 - 共享单个 Menu 组件 */}
      <Menu
        anchorEl={menuAnchor}
        open={Boolean(menuAnchor)}
        onClose={handleCloseMenu}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
        transformOrigin={{ vertical: 'top', horizontal: 'right' }}
        PaperProps={{
          sx: { minWidth: 180, borderRadius: 2 }
        }}
      >
        <MenuItem
          onClick={() => {
            const sub = subscriptions.find((s) => s.ID === menuSubId);
            if (sub) handleMenuAction(onLogs, sub);
          }}
        >
          <ListItemIcon>
            <HistoryIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText>访问记录</ListItemText>
        </MenuItem>

        <MenuItem
          onClick={() => {
            const sub = subscriptions.find((s) => s.ID === menuSubId);
            if (sub) handleMenuAction(onChainProxy, sub);
          }}
        >
          <ListItemIcon>
            <AccountTreeIcon fontSize="small" color="warning" />
          </ListItemIcon>
          <ListItemText>链式代理</ListItemText>
        </MenuItem>

        <MenuItem
          onClick={() => {
            const sub = subscriptions.find((s) => s.ID === menuSubId);
            if (sub) handleMenuAction(onStartSort, sub);
          }}
        >
          <ListItemIcon>
            <SortIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText>排序节点</ListItemText>
        </MenuItem>

        <MenuItem
          onClick={() => {
            const sub = subscriptions.find((s) => s.ID === menuSubId);
            if (sub) handleMenuAction(onCopy, sub);
          }}
        >
          <ListItemIcon>
            <ContentCopyIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText>复制订阅</ListItemText>
        </MenuItem>

        <Divider />

        <MenuItem
          onClick={() => {
            const sub = subscriptions.find((s) => s.ID === menuSubId);
            if (sub) handleMenuAction(onDelete, sub);
          }}
          sx={{ color: 'error.main' }}
        >
          <ListItemIcon>
            <DeleteIcon fontSize="small" color="error" />
          </ListItemIcon>
          <ListItemText>删除</ListItemText>
        </MenuItem>
      </Menu>
    </>
  );
}
