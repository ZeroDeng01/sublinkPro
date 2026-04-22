import { useMemo } from 'react';
import { useTheme } from '@mui/material/styles';
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
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getSurfaceTokens } from 'themes/surfaceTokens';
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
  selectorNodesTotal,
  selectorNodesLoading,
  selectedNodes,
  selectedNodesList,
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
  const { isDark } = useResolvedColorScheme();
  const { palette, dialogSurface, dialogSurfaceGradient, mutedPanelSurface, nestedPanelSurface, panelBorder } = getSurfaceTokens(
    theme,
    isDark
  );
  const insetHighlight = isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.03)}` : 'none';
  const emphasisInsetHighlight = isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.04)}` : 'none';
  const searchInsetHighlight = isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.02)}` : 'none';
  const listItemNeutralBorder = withAlpha(palette.divider, isDark ? 0.72 : 0.48);
  const listItemSurface = isDark ? withAlpha(palette.background.default, 0.62) : 'transparent';
  const listItemHoverSurface = isDark ? withAlpha(palette.background.paper, 0.5) : palette.action.hover;
  const actionStripSurface = isDark ? withAlpha(palette.background.default, 0.9) : palette.background.paper;

  const getAccentTone = (colorKey) => ({
    main: palette[colorKey].main,
    subtleSurface: withAlpha(palette[colorKey].main, isDark ? 0.14 : 0.08),
    subtleHoverSurface: withAlpha(palette[colorKey].main, isDark ? 0.2 : 0.12),
    strongSurface: withAlpha(palette[colorKey].main, isDark ? 0.18 : 0.1),
    strongHoverSurface: withAlpha(palette[colorKey].main, isDark ? 0.26 : 0.16),
    faintSurface: withAlpha(palette[colorKey].main, isDark ? 0.12 : 0.06),
    selectedSurface: withAlpha(palette[colorKey].main, isDark ? 0.16 : 0.08),
    selectedBorder: withAlpha(palette[colorKey].main, isDark ? 0.38 : 0.24),
    softBorder: withAlpha(palette[colorKey].main, isDark ? 0.32 : 0.18),
    mediumBorder: withAlpha(palette[colorKey].main, isDark ? 0.34 : 0.22),
    hoverBorder: withAlpha(palette[colorKey].main, isDark ? 0.28 : 0.18),
    strongBorder: withAlpha(palette[colorKey].main, isDark ? 0.42 : 0.28),
    ring: withAlpha(palette[colorKey].main, 0.18)
  });

  const primaryTone = getAccentTone('primary');
  const errorTone = getAccentTone('error');

  const buildPanelSx = (colorKey) => ({
    backgroundColor: dialogSurface,
    backgroundImage: dialogSurfaceGradient,
    border: '1px solid',
    borderColor: withAlpha(palette[colorKey].main, isDark ? 0.28 : 0.18),
    borderRadius: 3,
    boxShadow: insetHighlight,
    transition: 'border-color 0.2s ease, box-shadow 0.2s ease',
    '&:hover': {
      borderColor: `${colorKey}.main`,
      boxShadow: isDark ? `0 0 0 1px ${withAlpha(palette[colorKey].main, 0.18)}` : theme.shadows[2]
    }
  });

  const listSurfaceSx = {
    bgcolor: nestedPanelSurface,
    border: '1px solid',
    borderColor: panelBorder,
    borderRadius: 2,
    boxShadow: insetHighlight
  };

  const availableNodesTotal = selectorNodesTotal || availableNodes.length;

  const searchFieldSx = {
    mb: 2,
    flexShrink: 0,
    '& .MuiOutlinedInput-root': {
      borderRadius: 2.5,
      bgcolor: mutedPanelSurface,
      border: '1px solid',
      borderColor: panelBorder,
      boxShadow: searchInsetHighlight,
      transition: 'border-color 0.2s ease, background-color 0.2s ease, box-shadow 0.2s ease',
      '& fieldset': { borderColor: 'transparent' },
      '&:hover': {
        bgcolor: isDark ? withAlpha(palette.background.default, 0.92) : nestedPanelSurface,
        '& fieldset': {
          borderColor: primaryTone.hoverBorder
        }
      },
      '&.Mui-focused': {
        bgcolor: dialogSurface,
        boxShadow: isDark ? `0 0 0 1px ${withAlpha(primaryTone.main, 0.16)}` : 'none',
        '& fieldset': {
          borderColor: primaryTone.softBorder
        }
      }
    },
    '& .MuiInputBase-input::placeholder': {
      color: isDark ? withAlpha(palette.text.secondary, 0.92) : undefined,
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
    boxShadow: isDark ? insetHighlight : theme.shadows[1]
  };

  const actionStripButtonSx = (colorKey) => {
    const tone = getAccentTone(colorKey);
    return {
      bgcolor: tone.subtleSurface,
      color: `${colorKey}.main`,
      border: '1px solid',
      borderColor: tone.softBorder,
      fontWeight: 700,
      boxShadow: 'none',
      '&:hover': { bgcolor: tone.subtleHoverSurface },
      '&:disabled': {
        bgcolor: 'action.disabledBackground',
        color: 'text.disabled'
      }
    };
  };
  const actionStripOutlinedButtonSx = (colorKey) => {
    const tone = getAccentTone(colorKey);
    return {
      bgcolor: nestedPanelSurface,
      borderColor: tone.softBorder,
      borderWidth: 1,
      color: tone.main,
      fontWeight: 700,
      '&:hover': {
        bgcolor: tone.faintSurface,
        borderColor: tone.strongBorder
      }
    };
  };
  const desktopActionButtonSx = (colorKey) => {
    const tone = getAccentTone(colorKey);
    return {
      minWidth: 120,
      bgcolor: tone.strongSurface,
      border: '1px solid',
      borderColor: tone.mediumBorder,
      boxShadow: 'none',
      color: tone.main,
      fontWeight: 700,
      '&:hover': {
        bgcolor: tone.strongHoverSurface,
        borderColor: tone.main
      },
      '&:disabled': {
        background: palette.action.disabledBackground,
        color: palette.text.disabled
      }
    };
  };
  const desktopSecondaryActionButtonSx = (colorKey) => {
    const tone = getAccentTone(colorKey);
    return {
      minWidth: 120,
      bgcolor: nestedPanelSurface,
      border: '1px solid',
      borderColor: tone.softBorder,
      color: tone.main,
      fontWeight: 600,
      '&:hover': {
        bgcolor: tone.faintSurface,
        borderColor: tone.main
      }
    };
  };
  const desktopActionStripSx = {
    width: '100%',
    p: 1.5,
    borderRadius: 3,
    bgcolor: mutedPanelSurface,
    border: '1px solid',
    borderColor: panelBorder,
    boxShadow: isDark ? insetHighlight : theme.shadows[1]
  };

  const mobileTabsSx = {
    mb: 2,
    p: 0.5,
    bgcolor: mutedPanelSurface,
    border: '1px solid',
    borderColor: panelBorder,
    borderRadius: 3,
    boxShadow: insetHighlight,
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
        bgcolor: primaryTone.faintSurface
      }
    },
    '& .Mui-selected': {
      bgcolor: nestedPanelSurface,
      color: 'text.primary',
      border: '1px solid',
      borderColor: primaryTone.hoverBorder,
      boxShadow: emphasisInsetHighlight
    }
  };

  const getTransferListItemSx = (tone, checked, translateX) => ({
    py: 0.75,
    px: 1,
    mb: 0.5,
    borderRadius: 2,
    bgcolor: checked ? tone.selectedSurface : listItemSurface,
    border: '1px solid',
    borderColor: checked ? tone.selectedBorder : listItemNeutralBorder,
    boxShadow: checked ? emphasisInsetHighlight : 'none',
    transition: 'all 0.15s ease-in-out',
    '&:hover': {
      bgcolor: checked ? tone.subtleHoverSurface : listItemHoverSurface,
      transform: `translateX(${translateX}px)`,
      borderColor: tone.hoverBorder
    }
  });

  const getTransferIconButtonSx = (tone) => ({
    bgcolor: tone.selectedSurface,
    color: tone.main,
    border: '1px solid',
    borderColor: tone.softBorder,
    '&:hover': { bgcolor: tone.subtleHoverSurface }
  });

  // 筛选已选节点
  const filteredSelectedNodes = useMemo(() => {
    if (!selectedNodeSearch) return selectedNodesList;
    const query = selectedNodeSearch.toLowerCase();
    return selectedNodesList.filter((node) => node.Name?.toLowerCase().includes(query) || node.Group?.toLowerCase().includes(query));
  }, [selectedNodesList, selectedNodeSearch]);

  const handleMobileTabChange = (_event, nextTab) => onMobileTabChange(nextTab);
  const handleSelectedNodeSearchChange = (event) => onSelectedNodeSearchChange(event.target.value);

  // 移动端穿梭框
  if (matchDownMd) {
    return (
      <Box sx={{ mt: 2 }}>
        <Tabs value={mobileTab} onChange={handleMobileTabChange} variant="fullWidth" sx={mobileTabsSx}>
          <Tab label={`可选节点 (${availableNodesTotal})`} icon={<ChevronRightIcon />} iconPosition="end" />
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
                      checked={checkedAvailable.length === availableNodesTotal && availableNodesTotal > 0}
                      indeterminate={checkedAvailable.length > 0 && checkedAvailable.length < availableNodesTotal}
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
                  label={availableNodesTotal > 100 ? `显示前100/${availableNodesTotal}` : `${availableNodesTotal}个`}
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
                      ...getTransferListItemSx(primaryTone, checkedAvailable.includes(node.ID), 4)
                    }}
                    secondaryAction={
                      <IconButton size="small" color="primary" onClick={() => onAddNode(node.ID)} sx={getTransferIconButtonSx(primaryTone)}>
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
              onChange={handleSelectedNodeSearchChange}
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
                      ...getTransferListItemSx(errorTone, checkedSelected.includes(node.ID), -4)
                    }}
                    secondaryAction={
                      <IconButton size="small" color="error" onClick={() => onRemoveNode(node.ID)} sx={getTransferIconButtonSx(errorTone)}>
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
                    checked={checkedAvailable.length === availableNodesTotal && availableNodesTotal > 0}
                    indeterminate={checkedAvailable.length > 0 && checkedAvailable.length < availableNodesTotal}
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
            <Chip label={availableNodesTotal > 100 ? `前100/${availableNodesTotal}` : availableNodesTotal} size="small" color="primary" />
          </Stack>
          <Box sx={{ flexGrow: 1, overflow: 'auto', pr: 1 }}>
            <List dense sx={{ ...listSurfaceSx, p: 1 }}>
              {availableNodes.slice(0, 100).map((node) => (
                <ListItem
                  key={node.ID}
                  sx={{
                    ...getTransferListItemSx(primaryTone, checkedAvailable.includes(node.ID), 4),
                    cursor: 'pointer',
                    py: 0.5
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
              {availableNodesTotal > 100 && (
                <Typography variant="caption" color="text.secondary" sx={{ display: 'block', textAlign: 'center', py: 1 }}>
                  还有 {availableNodesTotal - 100} 个节点未显示
                </Typography>
              )}
              {selectorNodesLoading && (
                <Typography variant="caption" color="text.secondary" sx={{ display: 'block', textAlign: 'center', py: 1 }}>
                  节点列表加载中...
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
            onChange={handleSelectedNodeSearchChange}
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
                    ...getTransferListItemSx(errorTone, checkedSelected.includes(node.ID), -4),
                    cursor: 'pointer',
                    py: 0.5
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
