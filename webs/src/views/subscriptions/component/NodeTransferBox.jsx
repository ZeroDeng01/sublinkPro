import { useMemo } from 'react';
import { alpha, useTheme } from '@mui/material/styles';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import IconButton from '@mui/material/IconButton';
import Paper from '@mui/material/Paper';
import Chip from '@mui/material/Chip';
import Stack from '@mui/material/Stack';
import Checkbox from '@mui/material/Checkbox';
import FormControlLabel from '@mui/material/FormControlLabel';
import Typography from '@mui/material/Typography';
import InputAdornment from '@mui/material/InputAdornment';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemText from '@mui/material/ListItemText';
import ListItemIcon from '@mui/material/ListItemIcon';
import Divider from '@mui/material/Divider';
import Grid from '@mui/material/Grid';
import Tabs from '@mui/material/Tabs';
import Tab from '@mui/material/Tab';
import Fade from '@mui/material/Fade';
import { withAlpha } from '../../../utils/colorUtils';

// icons
import AddIcon from '@mui/icons-material/Add';
import DeleteIcon from '@mui/icons-material/Delete';
import ChevronRightIcon from '@mui/icons-material/ChevronRight';
import ChevronLeftIcon from '@mui/icons-material/ChevronLeft';
import SearchIcon from '@mui/icons-material/Search';

/**
 * 节点穿梭框组件
 * 支持移动端Tab模式和桌面端双栏布局
 */
export default function NodeTransferBox({
  // 数据
  availableNodes,
  selectedNodes,
  selectedNodesList,
  allNodes,
  // 选中状态
  checkedAvailable,
  checkedSelected,
  // 搜索
  selectedNodeSearch,
  onSelectedNodeSearchChange,
  // 移动端Tab
  mobileTab,
  onMobileTabChange,
  matchDownMd,
  // 操作回调
  onAddNode,
  onRemoveNode,
  onAddAllVisible,
  onRemoveAll,
  onToggleAvailable,
  onToggleSelected,
  onAddChecked,
  onRemoveChecked,
  onToggleAllAvailable,
  onToggleAllSelected
}) {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const palette = theme.vars?.palette || theme.palette;
  const dialogSurface = isDark ? withAlpha(palette.background.default, 0.96) : palette.background.paper;
  const dialogSurfaceGradient = isDark
    ? `linear-gradient(180deg, ${withAlpha(palette.background.paper, 0.16)} 0%, ${dialogSurface} 100%)`
    : 'none';
  const mutedPanelSurface = isDark ? withAlpha(palette.background.default, 0.84) : palette.background.default;
  const nestedPanelSurface = isDark ? withAlpha(palette.background.paper, 0.42) : palette.background.paper;
  const panelBorder = isDark ? withAlpha(palette.divider, 0.82) : withAlpha(palette.divider, 0.9);
  const listItemSurface = isDark ? withAlpha(palette.background.default, 0.62) : 'transparent';
  const listItemHoverSurface = isDark ? withAlpha(palette.background.paper, 0.5) : theme.palette.action.hover;
  const actionStripSurface = isDark ? withAlpha(palette.background.default, 0.9) : palette.background.paper;

  const buildPanelSx = (colorKey) => ({
    backgroundColor: dialogSurface,
    backgroundImage: dialogSurfaceGradient,
    border: '1px solid',
    borderColor: alpha(theme.palette[colorKey].main, isDark ? 0.28 : 0.18),
    borderRadius: 3,
    boxShadow: isDark ? `inset 0 1px 0 ${alpha(theme.palette.common.white, 0.03)}` : 'none',
    transition: 'border-color 0.2s ease, box-shadow 0.2s ease',
    '&:hover': {
      borderColor: `${colorKey}.main`,
      boxShadow: isDark ? `0 0 0 1px ${alpha(theme.palette[colorKey].main, 0.18)}` : theme.shadows[2]
    }
  });

  const listSurfaceSx = {
    bgcolor: nestedPanelSurface,
    border: '1px solid',
    borderColor: panelBorder,
    borderRadius: 2,
    boxShadow: isDark ? `inset 0 1px 0 ${alpha(theme.palette.common.white, 0.03)}` : 'none'
  };

  const searchFieldSx = {
    mb: 2,
    flexShrink: 0,
    '& .MuiOutlinedInput-root': {
      borderRadius: 2.5,
      bgcolor: mutedPanelSurface,
      border: '1px solid',
      borderColor: panelBorder,
      boxShadow: isDark ? `inset 0 1px 0 ${alpha(theme.palette.common.white, 0.02)}` : 'none',
      transition: 'border-color 0.2s ease, background-color 0.2s ease, box-shadow 0.2s ease',
      '& fieldset': { borderColor: 'transparent' },
      '&:hover': {
        bgcolor: isDark ? withAlpha(palette.background.default, 0.92) : palette.background.paper,
        '& fieldset': {
          borderColor: alpha(theme.palette.primary.main, isDark ? 0.22 : 0.14)
        }
      },
      '&.Mui-focused': {
        bgcolor: dialogSurface,
        boxShadow: isDark ? `0 0 0 1px ${alpha(theme.palette.primary.main, 0.16)}` : 'none',
        '& fieldset': {
          borderColor: alpha(theme.palette.primary.main, isDark ? 0.32 : 0.22)
        }
      }
    },
    '& .MuiInputBase-input::placeholder': {
      color: isDark ? alpha(theme.palette.text.secondary, 0.92) : undefined,
      opacity: 1
    }
  };

  const actionStripSx = {
    mt: 2,
    p: 1.5,
    borderRadius: 2.5,
    display: 'flex',
    gap: 1,
    justifyContent: 'center',
    bgcolor: actionStripSurface,
    border: '1px solid',
    borderColor: panelBorder,
    boxShadow: isDark ? `inset 0 1px 0 ${alpha(theme.palette.common.white, 0.03)}` : theme.shadows[1]
  };

  const actionStripButtonSx = (colorKey) => ({
    bgcolor: alpha(theme.palette[colorKey].main, isDark ? 0.14 : 0.08),
    color: `${colorKey}.main`,
    border: '1px solid',
    borderColor: alpha(theme.palette[colorKey].main, isDark ? 0.32 : 0.18),
    fontWeight: 700,
    boxShadow: 'none',
    '&:hover': { bgcolor: alpha(theme.palette[colorKey].main, isDark ? 0.2 : 0.12) },
    '&:disabled': {
      bgcolor: 'action.disabledBackground',
      color: 'text.disabled'
    }
  });
  const actionStripOutlinedButtonSx = (colorKey) => ({
    bgcolor: nestedPanelSurface,
    borderColor: alpha(theme.palette[colorKey].main, isDark ? 0.28 : 0.18),
    borderWidth: 1,
    color: theme.palette[colorKey].main,
    fontWeight: 700,
    '&:hover': {
      bgcolor: isDark ? withAlpha(palette.background.paper, 0.52) : alpha(theme.palette[colorKey].main, 0.06),
      borderColor: alpha(theme.palette[colorKey].main, isDark ? 0.42 : 0.28)
    }
  });
  const desktopActionButtonSx = (colorKey) => ({
    minWidth: 120,
    bgcolor: alpha(theme.palette[colorKey].main, isDark ? 0.18 : 0.1),
    border: '1px solid',
    borderColor: alpha(theme.palette[colorKey].main, isDark ? 0.34 : 0.22),
    boxShadow: 'none',
    color: theme.palette[colorKey].main,
    fontWeight: 700,
    '&:hover': {
      bgcolor: alpha(theme.palette[colorKey].main, isDark ? 0.26 : 0.16),
      borderColor: theme.palette[colorKey].main
    },
    '&:disabled': {
      background: theme.palette.action.disabledBackground,
      color: theme.palette.text.disabled
    }
  });
  const desktopSecondaryActionButtonSx = (colorKey) => ({
    minWidth: 120,
    bgcolor: isDark ? withAlpha(palette.background.paper, 0.46) : alpha(theme.palette[colorKey].main, 0.04),
    border: '1px solid',
    borderColor: alpha(theme.palette[colorKey].main, isDark ? 0.28 : 0.18),
    color: theme.palette[colorKey].main,
    fontWeight: 600,
    '&:hover': {
      bgcolor: alpha(theme.palette[colorKey].main, isDark ? 0.14 : 0.08),
      borderColor: theme.palette[colorKey].main
    }
  });
  const desktopActionStripSx = {
    width: '100%',
    p: 1.5,
    borderRadius: 3,
    bgcolor: mutedPanelSurface,
    border: '1px solid',
    borderColor: panelBorder,
    boxShadow: isDark ? `inset 0 1px 0 ${alpha(theme.palette.common.white, 0.03)}` : theme.shadows[1]
  };

  // 筛选已选节点
  const filteredSelectedNodes = useMemo(() => {
    if (!selectedNodeSearch) return selectedNodesList;
    const query = selectedNodeSearch.toLowerCase();
    return selectedNodesList.filter((node) => node.Name?.toLowerCase().includes(query) || node.Group?.toLowerCase().includes(query));
  }, [selectedNodesList, selectedNodeSearch]);

  // 移动端穿梭框
  if (matchDownMd) {
    return (
      <Box sx={{ mt: 2 }}>
        <Tabs
          value={mobileTab}
          onChange={(e, v) => onMobileTabChange(v)}
          variant="fullWidth"
          sx={{
            mb: 2,
            p: 0.5,
            bgcolor: mutedPanelSurface,
            border: '1px solid',
            borderColor: panelBorder,
            borderRadius: 3,
            boxShadow: isDark ? `inset 0 1px 0 ${alpha(theme.palette.common.white, 0.03)}` : 'none',
            '& .MuiTabs-indicator': {
              display: 'none'
            },
            '& .MuiTab-root': {
              fontWeight: 600,
              borderRadius: 2,
              mx: 0.5,
              minHeight: 44,
              color: 'text.secondary',
              transition: 'all 0.2s',
              '&:hover': {
                bgcolor: isDark ? withAlpha(palette.background.paper, 0.26) : alpha(theme.palette.primary.main, 0.05)
              }
            },
            '& .Mui-selected': {
              bgcolor: nestedPanelSurface,
              color: 'text.primary',
              border: '1px solid',
              borderColor: alpha(theme.palette.primary.main, isDark ? 0.24 : 0.18),
              boxShadow: isDark ? `inset 0 1px 0 ${alpha(theme.palette.common.white, 0.04)}` : 'none'
            }
          }}
        >
          <Tab label={`可选节点 (${availableNodes.length})`} icon={<ChevronRightIcon />} iconPosition="end" />
          <Tab label={`已选节点 (${selectedNodes.length})`} icon={<ChevronLeftIcon />} iconPosition="start" />
        </Tabs>

        {/* 可选节点面板 */}
        <Fade in={mobileTab === 0}>
          <Box sx={{ display: mobileTab === 0 ? 'block' : 'none' }}>
            <Paper
              elevation={0}
              sx={{
                p: 2,
                maxHeight: 350,
                overflow: 'auto',
                ...buildPanelSx('primary')
              }}
            >
              <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 1.5 }}>
                <FormControlLabel
                  control={
                    <Checkbox
                      checked={checkedAvailable.length === availableNodes.length && availableNodes.length > 0}
                      indeterminate={checkedAvailable.length > 0 && checkedAvailable.length < availableNodes.length}
                      onChange={onToggleAllAvailable}
                      size="small"
                    />
                  }
                  label={
                    <Typography variant="body2" fontWeight={600}>
                      全选
                    </Typography>
                  }
                />
                <Chip
                  label={availableNodes.length > 100 ? `显示前100/${availableNodes.length}` : `${availableNodes.length}个`}
                  size="small"
                  color="primary"
                  variant="outlined"
                />
              </Stack>
              <List dense sx={{ pt: 0, ...listSurfaceSx, p: 1 }}>
                {availableNodes.slice(0, 100).map((node) => (
                  <ListItem
                    key={node.ID}
                    sx={{
                      py: 0.75,
                      px: 1,
                      mb: 0.5,
                      borderRadius: 2,
                      bgcolor: checkedAvailable.includes(node.ID)
                        ? alpha(theme.palette.primary.main, isDark ? 0.16 : 0.08)
                        : listItemSurface,
                      border: '1px solid',
                      borderColor: checkedAvailable.includes(node.ID)
                        ? alpha(theme.palette.primary.main, isDark ? 0.38 : 0.24)
                        : alpha(theme.palette.divider, isDark ? 0.72 : 0.4),
                      transition: 'all 0.15s ease-in-out',
                      '&:hover': {
                        bgcolor: listItemHoverSurface,
                        transform: 'translateX(4px)',
                        borderColor: alpha(theme.palette.primary.main, isDark ? 0.28 : 0.18)
                      }
                    }}
                    secondaryAction={
                      <IconButton
                        size="small"
                        color="primary"
                        onClick={() => onAddNode(node.ID)}
                        sx={{
                          bgcolor: alpha(theme.palette.primary.main, isDark ? 0.16 : 0.08),
                          color: 'primary.main',
                          border: '1px solid',
                          borderColor: alpha(theme.palette.primary.main, isDark ? 0.3 : 0.18),
                          '&:hover': { bgcolor: alpha(theme.palette.primary.main, isDark ? 0.22 : 0.12) }
                        }}
                      >
                        <AddIcon fontSize="small" />
                      </IconButton>
                    }
                  >
                    <ListItemIcon sx={{ minWidth: 36 }}>
                      <Checkbox
                        edge="start"
                        checked={checkedAvailable.includes(node.ID)}
                        onChange={() => onToggleAvailable(node.ID)}
                        size="small"
                      />
                    </ListItemIcon>
                    <ListItemText
                      primary={node.Name}
                      secondary={
                        <Chip
                          label={node.Group || '未分组'}
                          size="small"
                          variant="outlined"
                          sx={{ mt: 0.5, height: 20, fontSize: '0.7rem' }}
                        />
                      }
                      primaryTypographyProps={{
                        noWrap: true,
                        fontWeight: 500,
                        sx: { maxWidth: 'calc(100% - 60px)' }
                      }}
                    />
                  </ListItem>
                ))}
              </List>
            </Paper>

            {/* 移动端底部操作栏 */}
            <Paper elevation={3} sx={actionStripSx}>
              <Button
                variant="contained"
                color="inherit"
                size="small"
                startIcon={<AddIcon />}
                onClick={onAddChecked}
                disabled={checkedAvailable.length === 0}
                sx={actionStripButtonSx('primary')}
              >
                添加选中 ({checkedAvailable.length})
              </Button>
              <Button variant="outlined" size="small" onClick={onAddAllVisible} sx={actionStripOutlinedButtonSx('primary')}>
                全部添加
              </Button>
            </Paper>
          </Box>
        </Fade>

        {/* 已选节点面板 */}
        <Fade in={mobileTab === 1}>
          <Box sx={{ display: mobileTab === 1 ? 'block' : 'none' }}>
            <TextField
              fullWidth
              size="small"
              placeholder="搜索已选节点..."
              value={selectedNodeSearch}
              onChange={(e) => onSelectedNodeSearchChange(e.target.value)}
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <SearchIcon color="action" />
                  </InputAdornment>
                )
              }}
              sx={searchFieldSx}
            />
            <Paper
              elevation={0}
              sx={{
                p: 2,
                maxHeight: 350,
                overflow: 'auto',
                ...buildPanelSx('success')
              }}
            >
              <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 1.5 }}>
                <FormControlLabel
                  control={
                    <Checkbox
                      checked={checkedSelected.length === filteredSelectedNodes.length && filteredSelectedNodes.length > 0}
                      indeterminate={checkedSelected.length > 0 && checkedSelected.length < filteredSelectedNodes.length}
                      onChange={onToggleAllSelected}
                      size="small"
                      color="success"
                    />
                  }
                  label={
                    <Typography variant="body2" fontWeight={600}>
                      全选
                    </Typography>
                  }
                />
                <Chip label={`${selectedNodes.length}个已选`} size="small" color="success" />
              </Stack>
              <List dense sx={{ pt: 0, ...listSurfaceSx, p: 1 }}>
                {filteredSelectedNodes.map((node) => (
                  <ListItem
                    key={node.ID}
                    sx={{
                      py: 0.75,
                      px: 1,
                      mb: 0.5,
                      borderRadius: 2,
                      bgcolor: checkedSelected.includes(node.ID) ? alpha(theme.palette.error.main, isDark ? 0.14 : 0.08) : listItemSurface,
                      border: '1px solid',
                      borderColor: checkedSelected.includes(node.ID)
                        ? alpha(theme.palette.error.main, isDark ? 0.34 : 0.22)
                        : alpha(theme.palette.divider, isDark ? 0.72 : 0.4),
                      transition: 'all 0.15s ease-in-out',
                      '&:hover': {
                        bgcolor: listItemHoverSurface,
                        transform: 'translateX(-4px)',
                        borderColor: alpha(theme.palette.error.main, isDark ? 0.26 : 0.18)
                      }
                    }}
                    secondaryAction={
                      <IconButton
                        size="small"
                        color="error"
                        onClick={() => onRemoveNode(node.ID)}
                        sx={{
                          bgcolor: alpha(theme.palette.error.main, isDark ? 0.16 : 0.08),
                          color: 'error.main',
                          border: '1px solid',
                          borderColor: alpha(theme.palette.error.main, isDark ? 0.3 : 0.18),
                          '&:hover': { bgcolor: alpha(theme.palette.error.main, isDark ? 0.22 : 0.12) }
                        }}
                      >
                        <DeleteIcon fontSize="small" />
                      </IconButton>
                    }
                  >
                    <ListItemIcon sx={{ minWidth: 36 }}>
                      <Checkbox
                        edge="start"
                        checked={checkedSelected.includes(node.ID)}
                        onChange={() => onToggleSelected(node.ID)}
                        size="small"
                        color="error"
                      />
                    </ListItemIcon>
                    <ListItemText
                      primary={node.Name}
                      secondary={
                        <Chip
                          label={node.Group || '未分组'}
                          size="small"
                          color="success"
                          variant="outlined"
                          sx={{ mt: 0.5, height: 20, fontSize: '0.7rem' }}
                        />
                      }
                      primaryTypographyProps={{
                        noWrap: true,
                        fontWeight: 500,
                        sx: { maxWidth: 'calc(100% - 60px)' }
                      }}
                    />
                  </ListItem>
                ))}
              </List>
              {filteredSelectedNodes.length === 0 && (
                <Typography color="text.secondary" align="center" sx={{ py: 4 }}>
                  {selectedNodeSearch ? '未找到匹配的节点' : '暂无已选节点'}
                </Typography>
              )}
            </Paper>

            {/* 移动端底部操作栏 */}
            <Paper elevation={3} sx={actionStripSx}>
              <Button
                variant="contained"
                color="inherit"
                size="small"
                startIcon={<DeleteIcon />}
                onClick={onRemoveChecked}
                disabled={checkedSelected.length === 0}
                sx={actionStripButtonSx('error')}
              >
                移除选中 ({checkedSelected.length})
              </Button>
              <Button variant="outlined" size="small" onClick={onRemoveAll} sx={actionStripOutlinedButtonSx('error')}>
                全部移除
              </Button>
            </Paper>
          </Box>
        </Fade>
      </Box>
    );
  }

  // 桌面端穿梭框
  return (
    <Grid container spacing={2} sx={{ mt: 1 }}>
      {/* 可选节点 */}
      <Grid item xs={5}>
        <Paper
          elevation={0}
          sx={{
            p: 2,
            height: 380,
            overflow: 'hidden',
            display: 'flex',
            flexDirection: 'column',
            ...buildPanelSx('primary')
          }}
        >
          <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 1.5, flexShrink: 0 }}>
            <Stack direction="row" alignItems="center" spacing={1}>
              <FormControlLabel
                control={
                  <Checkbox
                    checked={checkedAvailable.length === availableNodes.length && availableNodes.length > 0}
                    indeterminate={checkedAvailable.length > 0 && checkedAvailable.length < availableNodes.length}
                    onChange={onToggleAllAvailable}
                    size="small"
                  />
                }
                label=""
                sx={{ mr: 0 }}
              />
              <Typography variant="subtitle1" fontWeight={700} color="primary">
                可选节点
              </Typography>
            </Stack>
            <Chip
              label={availableNodes.length > 100 ? `前100/${availableNodes.length}` : availableNodes.length}
              size="small"
              color="primary"
            />
          </Stack>
          <Box sx={{ flexGrow: 1, overflow: 'auto', pr: 1 }}>
            <List dense sx={{ ...listSurfaceSx, p: 1 }}>
              {availableNodes.slice(0, 100).map((node) => (
                <ListItem
                  key={node.ID}
                  sx={{
                    py: 0.5,
                    px: 1,
                    mb: 0.5,
                    borderRadius: 2,
                    cursor: 'pointer',
                    bgcolor: checkedAvailable.includes(node.ID) ? alpha(theme.palette.primary.main, isDark ? 0.16 : 0.1) : listItemSurface,
                    border: '1px solid',
                    borderColor: checkedAvailable.includes(node.ID)
                      ? alpha(theme.palette.primary.main, isDark ? 0.38 : 0.24)
                      : alpha(theme.palette.divider, isDark ? 0.72 : 0.5),
                    transition: 'all 0.15s ease-in-out',
                    '&:hover': {
                      bgcolor: listItemHoverSurface,
                      transform: 'translateX(4px)',
                      borderColor: alpha(theme.palette.primary.main, isDark ? 0.28 : 0.18)
                    }
                  }}
                  onClick={() => onToggleAvailable(node.ID)}
                  onDoubleClick={() => onAddNode(node.ID)}
                >
                  <ListItemIcon sx={{ minWidth: 32 }}>
                    <Checkbox edge="start" checked={checkedAvailable.includes(node.ID)} tabIndex={-1} disableRipple size="small" />
                  </ListItemIcon>
                  <ListItemText
                    primary={node.Name}
                    secondary={node.Group}
                    primaryTypographyProps={{
                      noWrap: true,
                      fontSize: '0.875rem',
                      fontWeight: 500
                    }}
                    secondaryTypographyProps={{
                      noWrap: true,
                      fontSize: '0.75rem'
                    }}
                  />
                </ListItem>
              ))}
              {availableNodes.length > 100 && (
                <Typography variant="caption" color="text.secondary" sx={{ display: 'block', textAlign: 'center', py: 1 }}>
                  还有 {availableNodes.length - 100} 个节点未显示
                </Typography>
              )}
            </List>
          </Box>
        </Paper>
      </Grid>

      {/* 中间操作按钮 */}
      <Grid item xs={2} sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <Box sx={desktopActionStripSx}>
          <Stack spacing={1} alignItems="center">
            <Button
              variant="contained"
              size="small"
              onClick={onAddChecked}
              disabled={checkedAvailable.length === 0}
              endIcon={<ChevronRightIcon />}
              sx={desktopActionButtonSx('primary')}
            >
              添加 ({checkedAvailable.length})
            </Button>
            <Button
              variant="outlined"
              size="small"
              onClick={onAddAllVisible}
              endIcon={<ChevronRightIcon />}
              sx={desktopSecondaryActionButtonSx('primary')}
            >
              全部添加
            </Button>
            <Divider sx={{ width: '60%', my: 0.5, borderColor: panelBorder }} />
            <Button
              variant="outlined"
              size="small"
              color="error"
              onClick={onRemoveAll}
              startIcon={<ChevronLeftIcon />}
              sx={desktopSecondaryActionButtonSx('error')}
            >
              全部移除
            </Button>
            <Button
              variant="contained"
              size="small"
              color="error"
              onClick={onRemoveChecked}
              disabled={checkedSelected.length === 0}
              startIcon={<ChevronLeftIcon />}
              sx={desktopActionButtonSx('error')}
            >
              移除 ({checkedSelected.length})
            </Button>
          </Stack>
        </Box>
      </Grid>

      {/* 已选节点 */}
      <Grid item xs={5}>
        <Paper
          elevation={0}
          sx={{
            p: 2,
            height: 380,
            overflow: 'hidden',
            display: 'flex',
            flexDirection: 'column',
            ...buildPanelSx('success')
          }}
        >
          <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 1, flexShrink: 0 }}>
            <Stack direction="row" alignItems="center" spacing={1}>
              <FormControlLabel
                control={
                  <Checkbox
                    checked={checkedSelected.length === filteredSelectedNodes.length && filteredSelectedNodes.length > 0}
                    indeterminate={checkedSelected.length > 0 && checkedSelected.length < filteredSelectedNodes.length}
                    onChange={onToggleAllSelected}
                    size="small"
                    color="success"
                  />
                }
                label=""
                sx={{ mr: 0 }}
              />
              <Typography variant="subtitle1" fontWeight={700} color="success.main">
                已选节点
              </Typography>
            </Stack>
            <Chip label={selectedNodes.length} size="small" color="success" />
          </Stack>
          <TextField
            fullWidth
            size="small"
            placeholder="搜索已选节点..."
            value={selectedNodeSearch}
            onChange={(e) => onSelectedNodeSearchChange(e.target.value)}
            InputProps={{
              startAdornment: (
                <InputAdornment position="start">
                  <SearchIcon fontSize="small" color="action" />
                </InputAdornment>
              )
            }}
            sx={{ ...searchFieldSx, mb: 1 }}
          />
          <Box sx={{ flexGrow: 1, overflow: 'auto', pr: 1 }}>
            <List dense sx={{ ...listSurfaceSx, p: 1 }}>
              {filteredSelectedNodes.map((node) => (
                <ListItem
                  key={node.ID}
                  sx={{
                    py: 0.5,
                    px: 1,
                    mb: 0.5,
                    borderRadius: 2,
                    cursor: 'pointer',
                    bgcolor: checkedSelected.includes(node.ID) ? alpha(theme.palette.error.main, isDark ? 0.14 : 0.08) : listItemSurface,
                    border: '1px solid',
                    borderColor: checkedSelected.includes(node.ID)
                      ? alpha(theme.palette.error.main, isDark ? 0.34 : 0.22)
                      : alpha(theme.palette.divider, isDark ? 0.72 : 0.5),
                    transition: 'all 0.15s ease-in-out',
                    '&:hover': {
                      bgcolor: listItemHoverSurface,
                      transform: 'translateX(-4px)',
                      borderColor: alpha(theme.palette.error.main, isDark ? 0.26 : 0.18)
                    }
                  }}
                  onClick={() => onToggleSelected(node.ID)}
                  onDoubleClick={() => onRemoveNode(node.ID)}
                >
                  <ListItemIcon sx={{ minWidth: 32 }}>
                    <Checkbox
                      edge="start"
                      checked={checkedSelected.includes(node.ID)}
                      tabIndex={-1}
                      disableRipple
                      size="small"
                      color="error"
                    />
                  </ListItemIcon>
                  <ListItemText
                    primary={node.Name}
                    secondary={node.Group}
                    primaryTypographyProps={{
                      noWrap: true,
                      fontSize: '0.875rem',
                      fontWeight: 500
                    }}
                    secondaryTypographyProps={{
                      noWrap: true,
                      fontSize: '0.75rem'
                    }}
                  />
                </ListItem>
              ))}
            </List>
            {filteredSelectedNodes.length === 0 && (
              <Typography color="text.secondary" align="center" sx={{ py: 4 }}>
                {selectedNodeSearch ? '未找到匹配的节点' : '暂无已选节点'}
              </Typography>
            )}
          </Box>
        </Paper>
      </Grid>
    </Grid>
  );
}
