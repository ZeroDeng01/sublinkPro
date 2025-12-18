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
            '& .MuiTab-root': {
              fontWeight: 600,
              borderRadius: 2,
              mx: 0.5,
              transition: 'all 0.2s'
            },
            '& .Mui-selected': {
              bgcolor: 'primary.light',
              color: 'primary.contrastText'
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
                background: `linear-gradient(145deg, ${theme.palette.mode === 'dark' ? '#1a1a2e' : '#f8f9fa'} 0%, ${theme.palette.mode === 'dark' ? '#16213e' : '#ffffff'} 100%)`,
                border: '1px solid',
                borderColor: 'divider',
                borderRadius: 3
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
              <List dense sx={{ pt: 0 }}>
                {availableNodes.slice(0, 100).map((node) => (
                  <ListItem
                    key={node.ID}
                    sx={{
                      py: 0.75,
                      px: 1,
                      mb: 0.5,
                      borderRadius: 2,
                      bgcolor: checkedAvailable.includes(node.ID) ? 'action.selected' : 'transparent',
                      border: '1px solid',
                      borderColor: checkedAvailable.includes(node.ID) ? 'primary.main' : 'transparent',
                      transition: 'all 0.15s ease-in-out',
                      '&:hover': {
                        bgcolor: 'action.hover',
                        transform: 'translateX(4px)'
                      }
                    }}
                    secondaryAction={
                      <IconButton
                        size="small"
                        color="primary"
                        onClick={() => onAddNode(node.ID)}
                        sx={{
                          bgcolor: 'primary.main',
                          color: 'white',
                          '&:hover': { bgcolor: 'primary.dark' }
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
            <Paper
              elevation={3}
              sx={{
                mt: 2,
                p: 1.5,
                borderRadius: 2,
                display: 'flex',
                gap: 1,
                justifyContent: 'center',
                background: `linear-gradient(90deg, ${theme.palette.primary.main} 0%, ${theme.palette.primary.dark} 100%)`
              }}
            >
              <Button
                variant="contained"
                color="inherit"
                size="small"
                startIcon={<AddIcon />}
                onClick={onAddChecked}
                disabled={checkedAvailable.length === 0}
                sx={{
                  bgcolor: 'white',
                  color: 'primary.dark',
                  fontWeight: 700,
                  boxShadow: '0 2px 4px rgba(0,0,0,0.2)',
                  '&:hover': { bgcolor: '#f5f5f5' },
                  '&:disabled': { bgcolor: '#e0e0e0', color: '#999' }
                }}
              >
                添加选中 ({checkedAvailable.length})
              </Button>
              <Button
                variant="outlined"
                size="small"
                onClick={onAddAllVisible}
                sx={{
                  borderColor: 'white',
                  borderWidth: 2,
                  color: 'white',
                  fontWeight: 700,
                  textShadow: '0 1px 2px rgba(0,0,0,0.3)',
                  '&:hover': { bgcolor: 'rgba(255,255,255,0.15)', borderColor: 'white' }
                }}
              >
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
              sx={{ mb: 2 }}
            />
            <Paper
              elevation={0}
              sx={{
                p: 2,
                maxHeight: 350,
                overflow: 'auto',
                background: `linear-gradient(145deg, ${theme.palette.mode === 'dark' ? '#1e3a2f' : '#f0fff4'} 0%, ${theme.palette.mode === 'dark' ? '#1a3330' : '#ffffff'} 100%)`,
                border: '1px solid',
                borderColor: 'success.light',
                borderRadius: 3
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
              <List dense sx={{ pt: 0 }}>
                {filteredSelectedNodes.map((node) => (
                  <ListItem
                    key={node.ID}
                    sx={{
                      py: 0.75,
                      px: 1,
                      mb: 0.5,
                      borderRadius: 2,
                      bgcolor: checkedSelected.includes(node.ID) ? 'error.lighter' : 'transparent',
                      border: '1px solid',
                      borderColor: checkedSelected.includes(node.ID) ? 'error.main' : 'transparent',
                      transition: 'all 0.15s ease-in-out',
                      '&:hover': {
                        bgcolor: 'action.hover',
                        transform: 'translateX(-4px)'
                      }
                    }}
                    secondaryAction={
                      <IconButton
                        size="small"
                        color="error"
                        onClick={() => onRemoveNode(node.ID)}
                        sx={{
                          bgcolor: 'error.main',
                          color: 'white',
                          '&:hover': { bgcolor: 'error.dark' }
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
                <Typography color="textSecondary" align="center" sx={{ py: 4 }}>
                  {selectedNodeSearch ? '未找到匹配的节点' : '暂无已选节点'}
                </Typography>
              )}
            </Paper>

            {/* 移动端底部操作栏 */}
            <Paper
              elevation={3}
              sx={{
                mt: 2,
                p: 1.5,
                borderRadius: 2,
                display: 'flex',
                gap: 1,
                justifyContent: 'center',
                background: `linear-gradient(90deg, ${theme.palette.error.main} 0%, ${theme.palette.error.dark} 100%)`
              }}
            >
              <Button
                variant="contained"
                color="inherit"
                size="small"
                startIcon={<DeleteIcon />}
                onClick={onRemoveChecked}
                disabled={checkedSelected.length === 0}
                sx={{
                  bgcolor: 'white',
                  color: 'error.dark',
                  fontWeight: 700,
                  boxShadow: '0 2px 4px rgba(0,0,0,0.2)',
                  '&:hover': { bgcolor: '#f5f5f5' },
                  '&:disabled': { bgcolor: '#e0e0e0', color: '#999' }
                }}
              >
                移除选中 ({checkedSelected.length})
              </Button>
              <Button
                variant="outlined"
                size="small"
                onClick={onRemoveAll}
                sx={{
                  borderColor: 'white',
                  borderWidth: 2,
                  color: 'white',
                  fontWeight: 700,
                  textShadow: '0 1px 2px rgba(0,0,0,0.3)',
                  '&:hover': { bgcolor: 'rgba(255,255,255,0.15)', borderColor: 'white' }
                }}
              >
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
            background: `linear-gradient(145deg, ${theme.palette.mode === 'dark' ? '#1a1a2e' : '#f8f9fa'} 0%, ${theme.palette.mode === 'dark' ? '#16213e' : '#ffffff'} 100%)`,
            border: '2px solid',
            borderColor: 'primary.light',
            borderRadius: 3,
            transition: 'all 0.3s ease',
            '&:hover': {
              borderColor: 'primary.main',
              boxShadow: `0 4px 20px ${theme.palette.primary.main}20`
            }
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
            <List dense>
              {availableNodes.slice(0, 100).map((node) => (
                <ListItem
                  key={node.ID}
                  sx={{
                    py: 0.5,
                    px: 1,
                    mb: 0.5,
                    borderRadius: 2,
                    cursor: 'pointer',
                    bgcolor: checkedAvailable.includes(node.ID) ? 'primary.lighter' : 'transparent',
                    border: '1px solid',
                    borderColor: checkedAvailable.includes(node.ID) ? 'primary.main' : 'divider',
                    transition: 'all 0.15s ease-in-out',
                    '&:hover': {
                      bgcolor: 'action.hover',
                      transform: 'translateX(4px)',
                      borderColor: 'primary.light'
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
                <Typography variant="caption" color="textSecondary" sx={{ display: 'block', textAlign: 'center', py: 1 }}>
                  还有 {availableNodes.length - 100} 个节点未显示
                </Typography>
              )}
            </List>
          </Box>
        </Paper>
      </Grid>

      {/* 中间操作按钮 */}
      <Grid item xs={2} sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: 1 }}>
        <Button
          variant="contained"
          size="small"
          onClick={onAddChecked}
          disabled={checkedAvailable.length === 0}
          endIcon={<ChevronRightIcon />}
          sx={{
            minWidth: 120,
            background: 'linear-gradient(45deg, #2196F3 30%, #21CBF3 90%)',
            boxShadow: '0 3px 5px 2px rgba(33, 150, 243, .3)',
            color: '#fff',
            fontWeight: 700,
            textShadow: '0 1px 2px rgba(0,0,0,0.2)',
            '&:disabled': {
              background: '#ccc',
              color: '#888'
            }
          }}
        >
          添加 ({checkedAvailable.length})
        </Button>
        <Button
          variant="outlined"
          size="small"
          onClick={onAddAllVisible}
          endIcon={<ChevronRightIcon />}
          sx={{ minWidth: 120, fontWeight: 600 }}
        >
          全部添加
        </Button>
        <Divider sx={{ width: '60%', my: 1 }} />
        <Button
          variant="outlined"
          size="small"
          color="error"
          onClick={onRemoveAll}
          startIcon={<ChevronLeftIcon />}
          sx={{ minWidth: 120, fontWeight: 600 }}
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
          sx={{
            minWidth: 120,
            background: 'linear-gradient(45deg, #FE6B8B 30%, #FF8E53 90%)',
            boxShadow: '0 3px 5px 2px rgba(254, 107, 139, .3)',
            color: '#fff',
            fontWeight: 700,
            textShadow: '0 1px 2px rgba(0,0,0,0.2)',
            '&:disabled': {
              background: '#ccc',
              color: '#888'
            }
          }}
        >
          移除 ({checkedSelected.length})
        </Button>
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
            background: `linear-gradient(145deg, ${theme.palette.mode === 'dark' ? '#1e3a2f' : '#f0fff4'} 0%, ${theme.palette.mode === 'dark' ? '#1a3330' : '#ffffff'} 100%)`,
            border: '2px solid',
            borderColor: 'success.light',
            borderRadius: 3,
            transition: 'all 0.3s ease',
            '&:hover': {
              borderColor: 'success.main',
              boxShadow: `0 4px 20px ${theme.palette.success.main}20`
            }
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
            sx={{ mb: 1, flexShrink: 0, '& .MuiOutlinedInput-root': { borderRadius: 2 } }}
          />
          <Box sx={{ flexGrow: 1, overflow: 'auto', pr: 1 }}>
            <List dense>
              {filteredSelectedNodes.map((node) => (
                <ListItem
                  key={node.ID}
                  sx={{
                    py: 0.5,
                    px: 1,
                    mb: 0.5,
                    borderRadius: 2,
                    cursor: 'pointer',
                    bgcolor: checkedSelected.includes(node.ID) ? 'error.lighter' : 'transparent',
                    border: '1px solid',
                    borderColor: checkedSelected.includes(node.ID) ? 'error.main' : 'divider',
                    transition: 'all 0.15s ease-in-out',
                    '&:hover': {
                      bgcolor: 'action.hover',
                      transform: 'translateX(-4px)',
                      borderColor: 'error.light'
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
              <Typography color="textSecondary" align="center" sx={{ py: 4 }}>
                {selectedNodeSearch ? '未找到匹配的节点' : '暂无已选节点'}
              </Typography>
            )}
          </Box>
        </Paper>
      </Grid>
    </Grid>
  );
}
