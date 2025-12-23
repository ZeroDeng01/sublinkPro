import { useState } from 'react';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Chip from '@mui/material/Chip';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import TextField from '@mui/material/TextField';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';
import Divider from '@mui/material/Divider';
import Menu from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';

// Icons
import SortIcon from '@mui/icons-material/Sort';
import ArrowUpwardIcon from '@mui/icons-material/ArrowUpward';
import ArrowDownwardIcon from '@mui/icons-material/ArrowDownward';
import VerticalAlignTopIcon from '@mui/icons-material/VerticalAlignTop';
import VerticalAlignBottomIcon from '@mui/icons-material/VerticalAlignBottom';
import MoveDownIcon from '@mui/icons-material/MoveDown';
import ClearAllIcon from '@mui/icons-material/ClearAll';
import SourceIcon from '@mui/icons-material/Source';
import AbcIcon from '@mui/icons-material/Abc';
import RouterIcon from '@mui/icons-material/Router';
import SpeedIcon from '@mui/icons-material/Speed';
import TimerIcon from '@mui/icons-material/Timer';
import PublicIcon from '@mui/icons-material/Public';

/**
 * 排序工具栏组件
 * 提供快速排序和批量移动功能
 */
export default function SortToolbar({ selectedItems = [], onBatchSort, onBatchMove, onClearSelection, totalItems = 0 }) {
  // 排序方向
  const [sortOrder, setSortOrder] = useState('asc');
  // 移动位置输入
  const [movePosition, setMovePosition] = useState('');
  // 排序菜单锚点
  const [sortMenuAnchor, setSortMenuAnchor] = useState(null);

  // 排序选项
  const sortOptions = [
    { value: 'source', label: '按来源', icon: <SourceIcon fontSize="small" /> },
    { value: 'name', label: '按名称', icon: <AbcIcon fontSize="small" /> },
    { value: 'protocol', label: '按协议', icon: <RouterIcon fontSize="small" /> },
    { value: 'delay', label: '按延迟', icon: <TimerIcon fontSize="small" /> },
    { value: 'speed', label: '按速度', icon: <SpeedIcon fontSize="small" /> },
    { value: 'country', label: '按地区', icon: <PublicIcon fontSize="small" /> }
  ];

  // 处理批量排序
  const handleBatchSort = (sortBy) => {
    if (onBatchSort) {
      onBatchSort(sortBy, sortOrder);
    }
    setSortMenuAnchor(null);
  };

  // 处理批量移动
  const handleBatchMove = (position) => {
    if (onBatchMove && selectedItems.length > 0) {
      onBatchMove(position);
    }
  };

  // 处理移动到指定位置
  const handleMoveToPosition = () => {
    const pos = parseInt(movePosition, 10);
    if (!isNaN(pos) && pos >= 1 && pos <= totalItems) {
      handleBatchMove(pos - 1); // 转为0-indexed
      setMovePosition('');
    }
  };

  const hasSelection = selectedItems.length > 0;

  return (
    <Box
      sx={{
        p: 1.5,
        mb: 1,
        borderRadius: 1,
        bgcolor: 'background.default',
        border: '1px solid',
        borderColor: 'divider'
      }}
    >
      {/* 快速排序区域 */}
      <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" sx={{ mb: hasSelection ? 1 : 0 }}>
        <Tooltip title="快速排序">
          <Chip
            icon={<SortIcon />}
            label="快速排序"
            size="small"
            onClick={(e) => setSortMenuAnchor(e.currentTarget)}
            sx={{ cursor: 'pointer' }}
          />
        </Tooltip>

        {/* 排序方向切换 */}
        <Tooltip title={sortOrder === 'asc' ? '升序' : '降序'}>
          <IconButton
            size="small"
            onClick={() => setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc')}
            color={sortOrder === 'asc' ? 'primary' : 'error'}
          >
            {sortOrder === 'asc' ? <ArrowUpwardIcon fontSize="small" /> : <ArrowDownwardIcon fontSize="small" />}
          </IconButton>
        </Tooltip>

        {/* 排序菜单 */}
        <Menu anchorEl={sortMenuAnchor} open={Boolean(sortMenuAnchor)} onClose={() => setSortMenuAnchor(null)}>
          {sortOptions.map((opt) => (
            <MenuItem key={opt.value} onClick={() => handleBatchSort(opt.value)}>
              <ListItemIcon>{opt.icon}</ListItemIcon>
              <ListItemText>{opt.label}</ListItemText>
            </MenuItem>
          ))}
        </Menu>

        {hasSelection && (
          <>
            <Divider orientation="vertical" flexItem sx={{ mx: 0.5 }} />
            <Typography variant="caption" color="text.secondary">
              已选 {selectedItems.length} 项
            </Typography>
          </>
        )}
      </Stack>

      {/* 批量移动区域（仅在有选中项时显示） */}
      {hasSelection && (
        <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap">
          <Tooltip title="移动到顶部">
            <Button size="small" variant="outlined" startIcon={<VerticalAlignTopIcon />} onClick={() => handleBatchMove(0)}>
              顶部
            </Button>
          </Tooltip>

          <Tooltip title="移动到底部">
            <Button size="small" variant="outlined" startIcon={<VerticalAlignBottomIcon />} onClick={() => handleBatchMove(totalItems - 1)}>
              底部
            </Button>
          </Tooltip>

          <Divider orientation="vertical" flexItem />

          {/* 移动到指定位置 */}
          <TextField
            size="small"
            type="number"
            placeholder="位置"
            value={movePosition}
            onChange={(e) => setMovePosition(e.target.value)}
            sx={{ width: 70 }}
            inputProps={{ min: 1, max: totalItems }}
          />
          <Tooltip title="移动到指定位置">
            <IconButton size="small" color="primary" onClick={handleMoveToPosition} disabled={!movePosition}>
              <MoveDownIcon />
            </IconButton>
          </Tooltip>

          <Divider orientation="vertical" flexItem />

          <Tooltip title="取消选择">
            <Button size="small" color="inherit" startIcon={<ClearAllIcon />} onClick={onClearSelection}>
              取消选择
            </Button>
          </Tooltip>
        </Stack>
      )}
    </Box>
  );
}
