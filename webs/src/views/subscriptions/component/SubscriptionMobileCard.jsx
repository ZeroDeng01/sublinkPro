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
import { getSubscriptionNameChipSx } from './subscriptionNameChipStyles';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';

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
  showPreview = false,
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
  const { isDark } = useResolvedColorScheme();
  const { palette, dialogSurface, dialogSurfaceGradient, mutedPanelSurface, nestedPanelSurface, panelBorder } = getSurfaceTokens(
    theme,
    isDark
  );
  const { primaryText, secondaryText, tertiaryText } = getReadableTextTokens(theme, isDark);

  // 更多菜单状态
  const [menuAnchor, setMenuAnchor] = useState(null);
  const [menuSubId, setMenuSubId] = useState(null);

  const shellInset = isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.03)}` : 'none';
  const cardSurface = isDark ? withAlpha(palette.background.paper, 0.94) : palette.background.paper;
  const expandedCardSurface = isDark
    ? `linear-gradient(180deg, ${withAlpha(palette.primary.main, 0.08)} 0%, ${cardSurface} 100%)`
    : withAlpha(palette.primary.main, 0.03);
  const neutralBorder = withAlpha(palette.divider, isDark ? 0.66 : 0.84);
  const interactiveBorder = withAlpha(palette.primary.main, isDark ? 0.28 : 0.18);
  const activeInteractiveBorder = withAlpha(palette.primary.main, isDark ? 0.4 : 0.24);
  const infoBlockSurface = isDark ? withAlpha(mutedPanelSurface, 0.92) : withAlpha(palette.background.default, 0.68);
  const actionStripSurface = isDark ? withAlpha(palette.background.default, 0.64) : mutedPanelSurface;
  const expandedSectionSurface = isDark ? withAlpha(palette.background.default, 0.76) : withAlpha(mutedPanelSurface, 0.88);
  const menuItemHoverSurface = withAlpha(palette.primary.main, isDark ? 0.12 : 0.08);

  const getActionIconSx = (accentColor) => ({
    width: 38,
    height: 38,
    borderRadius: 2.5,
    flexShrink: 0,
    color: accentColor || secondaryText,
    bgcolor: accentColor ? withAlpha(accentColor, isDark ? 0.12 : 0.06) : nestedPanelSurface,
    border: '1px solid',
    borderColor: accentColor ? withAlpha(accentColor, isDark ? 0.28 : 0.18) : neutralBorder,
    boxShadow: shellInset,
    transition: 'all 0.2s ease',
    '&:hover': {
      color: accentColor || primaryText,
      bgcolor: accentColor ? withAlpha(accentColor, isDark ? 0.18 : 0.1) : withAlpha(palette.primary.main, isDark ? 0.12 : 0.08),
      borderColor: accentColor ? withAlpha(accentColor, isDark ? 0.4 : 0.24) : interactiveBorder
    }
  });

  const getInfoChipSx = (accentColor, clickable = false) => ({
    height: 24,
    maxWidth: '100%',
    borderRadius: 1.5,
    bgcolor: withAlpha(accentColor, isDark ? 0.14 : 0.08),
    color: accentColor,
    border: '1px solid',
    borderColor: withAlpha(accentColor, isDark ? 0.28 : 0.18),
    boxShadow: clickable ? shellInset : 'none',
    transition: 'all 0.2s ease',
    cursor: clickable ? 'pointer' : 'default',
    '& .MuiChip-label': {
      px: 1,
      overflow: 'hidden',
      textOverflow: 'ellipsis',
      whiteSpace: 'nowrap',
      fontWeight: 600
    },
    ...(clickable
      ? {
          '&:hover': {
            bgcolor: withAlpha(accentColor, isDark ? 0.2 : 0.12),
            borderColor: withAlpha(accentColor, isDark ? 0.4 : 0.24)
          }
        }
      : null)
  });

  const menuItemSx = {
    borderRadius: 1.75,
    mx: 0.75,
    my: 0.25,
    px: 1.25,
    py: 0.875,
    color: primaryText,
    '& .MuiListItemIcon-root': {
      minWidth: 34,
      color: secondaryText
    },
    '& .MuiListItemText-primary': {
      color: 'inherit'
    },
    '&:hover': {
      bgcolor: menuItemHoverSurface
    }
  };

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
        {subscriptions.map((sub) => {
          const isSorting = sortingSubId === sub.ID;
          const isExpanded = expandedRows[sub.ID] || isSorting;

          return (
            <MainCard
              key={sub.ID}
              content={false}
              border
              sx={{
                borderRadius: 3,
                overflow: 'hidden',
                borderColor: isExpanded ? activeInteractiveBorder : panelBorder,
                bgcolor: isExpanded ? expandedCardSurface : cardSurface,
                boxShadow: isDark
                  ? `0 6px 16px ${withAlpha(palette.common.black, 0.22)}`
                  : `0 6px 18px ${withAlpha(palette.primary.main, isExpanded ? 0.12 : 0.08)}`,
                transition: 'border-color 0.2s ease, box-shadow 0.2s ease, transform 0.2s ease',
                '&:hover': {
                  transform: 'translateY(-1px)',
                  boxShadow: isDark
                    ? `0 10px 24px ${withAlpha(palette.common.black, 0.28)}`
                    : `0 10px 26px ${withAlpha(palette.primary.main, 0.14)}`
                }
              }}
            >
              <Box p={2}>
                <Stack spacing={1.25}>
                  {/* 头部：订阅名称和展开按钮 */}
                  <Stack
                    direction="row"
                    justifyContent="space-between"
                    alignItems="center"
                    onClick={() => onToggleRow(sub.ID)}
                    sx={{ cursor: 'pointer' }}
                  >
                    <Stack direction="row" spacing={1} alignItems="center" sx={{ flex: 1, minWidth: 0 }}>
                      <Chip label={sub.Name} sx={[...getSubscriptionNameChipSx(theme), { minWidth: 0, flexShrink: 1 }]} />
                      {isSorting && (
                        <Chip
                          label="排序中"
                          size="small"
                          sx={{
                            flexShrink: 0,
                            height: 22,
                            color: isDark ? palette.warning.light : palette.warning.dark,
                            bgcolor: withAlpha(palette.warning.main, isDark ? 0.18 : 0.12),
                            border: '1px solid',
                            borderColor: withAlpha(palette.warning.main, isDark ? 0.34 : 0.2),
                            '& .MuiChip-label': {
                              px: 1,
                              fontWeight: 700
                            }
                          }}
                        />
                      )}
                    </Stack>
                    <IconButton size="small" sx={getActionIconSx(isExpanded ? palette.primary.main : null)}>
                      {isExpanded ? <KeyboardArrowUpIcon /> : <KeyboardArrowDownIcon />}
                    </IconButton>
                  </Stack>

                  {/* 统计信息 */}
                  <Box
                    sx={{
                      px: 1.25,
                      py: 1,
                      borderRadius: 2.5,
                      bgcolor: infoBlockSurface,
                      border: '1px solid',
                      borderColor: neutralBorder,
                      boxShadow: shellInset
                    }}
                  >
                    <Stack direction="row" spacing={0.75} flexWrap="wrap" useFlexGap alignItems="center" justifyContent="space-between">
                      <Stack direction="row" spacing={0.75} flexWrap="wrap" useFlexGap>
                        <Chip label={`${sub.Nodes?.length || 0} 个节点`} size="small" sx={getInfoChipSx(palette.success.main)} />
                        <Chip label={`${sub.Groups?.length || 0} 个分组`} size="small" sx={getInfoChipSx(palette.warning.main)} />
                      </Stack>
                      <Typography variant="caption" sx={{ color: tertiaryText }}>
                        轻触头部展开更多内容
                      </Typography>
                    </Stack>
                  </Box>

                  <Divider sx={{ borderColor: withAlpha(palette.divider, isDark ? 0.7 : 1) }} />

                  {/* 操作区域 - 移动端优化 */}
                  <Box
                    sx={{
                      px: 1.25,
                      py: 1,
                      borderRadius: 2.5,
                      bgcolor: actionStripSurface,
                      border: '1px solid',
                      borderColor: neutralBorder,
                      boxShadow: shellInset
                    }}
                  >
                    {isSorting ? (
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
                          sx={{
                            borderColor: neutralBorder,
                            bgcolor: nestedPanelSurface,
                            color: secondaryText,
                            fontWeight: 700,
                            '&:hover': {
                              borderColor: interactiveBorder,
                              bgcolor: withAlpha(palette.primary.main, isDark ? 0.12 : 0.06),
                              color: primaryText
                            }
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
                          sx={{
                            fontWeight: 700,
                            boxShadow: 'none',
                            '&:hover': {
                              boxShadow: 'none'
                            }
                          }}
                        >
                          确认排序
                        </Button>
                      </Stack>
                    ) : (
                      // 正常模式：显示操作按钮
                      <Stack direction="row" justifyContent="space-between" alignItems="center" spacing={1}>
                        <Typography variant="caption" sx={{ color: tertiaryText, fontWeight: 500 }}>
                          {sub.CreateDate}
                        </Typography>

                        {/* 快捷操作按钮 - 使用较大的图标 */}
                        <Stack direction="row" spacing={0.75} alignItems="center">
                          {/* 预览 - 常用 */}
                          {showPreview && (
                            <IconButton
                              size="medium"
                              onClick={(e) => {
                                e.stopPropagation();
                                onPreview(sub);
                              }}
                              sx={getActionIconSx(palette.info.main)}
                            >
                              <VisibilityIcon />
                            </IconButton>
                          )}

                          {/* 客户端链接 - 常用 */}
                          <IconButton
                            size="medium"
                            onClick={(e) => {
                              e.stopPropagation();
                              onClient(sub);
                            }}
                            sx={getActionIconSx(palette.primary.main)}
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
                            sx={getActionIconSx(palette.secondary.main)}
                          >
                            <EditIcon />
                          </IconButton>

                          {/* 更多操作菜单 */}
                          <IconButton size="medium" onClick={(e) => handleOpenMenu(e, sub.ID)} sx={getActionIconSx()}>
                            <MoreVertIcon />
                          </IconButton>
                        </Stack>
                      </Stack>
                    )}
                  </Box>

                  {/* 可展开内容 */}
                  <Collapse in={isExpanded} timeout="auto" unmountOnExit>
                    <Box sx={{ pt: 0.25 }}>
                      <Box
                        sx={{
                          p: 1.25,
                          borderRadius: 2.5,
                          bgcolor: expandedSectionSurface,
                          border: '1px solid',
                          borderColor: neutralBorder,
                          boxShadow: shellInset
                        }}
                      >
                        {isSorting ? (
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
                          <Stack direction="row" spacing={0.75} flexWrap="wrap" useFlexGap>
                            {getSortedItems(sub).map((item, idx) =>
                              item._type === 'node' ? (
                                <Chip
                                  key={item._type + item.ID}
                                  label={item.Name}
                                  size="small"
                                  onClick={() => onCopyToClipboard(item.Link)}
                                  sx={getInfoChipSx(palette.success.main, true)}
                                />
                              ) : (
                                <Chip
                                  key={item._type + idx}
                                  label={`📁 ${item.Name}`}
                                  size="small"
                                  sx={getInfoChipSx(palette.warning.main)}
                                />
                              )
                            )}
                          </Stack>
                        )}
                      </Box>
                    </Box>
                  </Collapse>
                </Stack>
              </Box>
            </MainCard>
          );
        })}
      </Stack>

      {/* 更多操作菜单 - 共享单个 Menu 组件 */}
      <Menu
        anchorEl={menuAnchor}
        open={Boolean(menuAnchor)}
        onClose={handleCloseMenu}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
        transformOrigin={{ vertical: 'top', horizontal: 'right' }}
        slotProps={{
          paper: {
            sx: {
              mt: 0.75,
              minWidth: 196,
              borderRadius: 2.5,
              bgcolor: dialogSurface,
              backgroundImage: dialogSurfaceGradient,
              border: '1px solid',
              borderColor: panelBorder,
              boxShadow: isDark
                ? `0 12px 28px ${withAlpha(palette.common.black, 0.34)}`
                : `0 12px 28px ${withAlpha(palette.primary.main, 0.14)}`,
              '& .MuiList-root': {
                p: 0.75
              },
              '& .MuiDivider-root': {
                my: 0.5,
                borderColor: withAlpha(palette.divider, isDark ? 0.7 : 1)
              }
            }
          }
        }}
      >
        <MenuItem
          sx={menuItemSx}
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
          sx={menuItemSx}
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
          sx={menuItemSx}
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
          sx={menuItemSx}
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
          sx={{
            ...menuItemSx,
            color: 'error.main',
            '& .MuiListItemIcon-root': {
              minWidth: 34,
              color: 'error.main'
            },
            '&:hover': {
              bgcolor: withAlpha(palette.error.main, isDark ? 0.14 : 0.08)
            }
          }}
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
