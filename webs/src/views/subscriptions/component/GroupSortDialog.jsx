import { useState, useEffect, useCallback } from 'react';
import { useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';

// material-ui
import Box from '@mui/material/Box';
import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import Alert from '@mui/material/Alert';
import Snackbar from '@mui/material/Snackbar';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemText from '@mui/material/ListItemText';
import TextField from '@mui/material/TextField';
import InputAdornment from '@mui/material/InputAdornment';
import CircularProgress from '@mui/material/CircularProgress';
import Stack from '@mui/material/Stack';
import Paper from '@mui/material/Paper';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';

// icons
import SearchIcon from '@mui/icons-material/Search';
import ClearIcon from '@mui/icons-material/Clear';
import DragIndicatorIcon from '@mui/icons-material/DragIndicator';
import SaveIcon from '@mui/icons-material/Save';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';

// drag and drop
import { DragDropContext, Droppable, Draggable } from '@hello-pangea/dnd';

// project imports
import { getGroupSortGroups, getGroupSortDetail, saveGroupAirportSort } from 'api/groupSort';

/**
 * 分组排序对话框
 * 管理同一分组内不同机场的节点输出排序
 */
export default function GroupSortDialog({ open, onClose, showMessage }) {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));

  const [groups, setGroups] = useState([]);
  const [selectedGroup, setSelectedGroup] = useState('');
  const [airports, setAirports] = useState([]);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });

  // 加载分组列表
  const loadGroups = useCallback(async () => {
    try {
      const res = await getGroupSortGroups();
      setGroups(res.data || []);
    } catch (err) {
      handleSnackbar('加载分组列表失败', 'error');
    }
  }, []);

  // 加载分组详情
  const loadGroupDetail = useCallback(async (groupName) => {
    if (!groupName) return;
    setLoading(true);
    try {
      const res = await getGroupSortDetail(groupName);
      setAirports(res.data?.airports || []);
    } catch (err) {
      handleSnackbar('加载分组详情失败', 'error');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (open) {
      loadGroups();
      // 重置选中状态
      setSelectedGroup('');
      setAirports([]);
      setSearchText('');
    }
  }, [open, loadGroups]);

  const handleSnackbar = (message, severity) => {
    if (showMessage) {
      showMessage(message, severity);
    } else {
      setSnackbar({ open: true, message, severity });
    }
  };

  // 选择分组
  const handleSelectGroup = (groupName) => {
    setSelectedGroup(groupName);
    loadGroupDetail(groupName);
  };

  // 拖拽结束
  const handleDragEnd = (result) => {
    if (!result.destination) return;
    const items = Array.from(airports);
    const [reorderedItem] = items.splice(result.source.index, 1);
    items.splice(result.destination.index, 0, reorderedItem);
    const updated = items.map((item, index) => ({ ...item, sort: index }));
    setAirports(updated);
  };

  // 保存排序
  const handleSave = async () => {
    if (!selectedGroup) return;
    setSaving(true);
    try {
      const airportSorts = airports.map((a, index) => ({
        airportId: a.airportId,
        sort: index
      }));
      await saveGroupAirportSort({ groupName: selectedGroup, airportSorts });
      handleSnackbar('保存成功', 'success');
      loadGroups();
    } catch (err) {
      handleSnackbar('保存失败', 'error');
    } finally {
      setSaving(false);
    }
  };

  // 过滤分组列表
  const filteredGroups = groups.filter(
    (g) => !searchText || g.groupName.toLowerCase().includes(searchText.toLowerCase())
  );

  return (
    <>
      <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth fullScreen={isMobile}>
        <DialogTitle>
          <Stack direction="row" alignItems="center" justifyContent="space-between">
            <Typography variant="h6">分组排序</Typography>
            {selectedGroup && (
              <Button
                variant="contained"
                size="small"
                startIcon={<SaveIcon />}
                onClick={handleSave}
                disabled={saving || airports.length === 0}
              >
                {saving ? '保存中...' : '保存排序'}
              </Button>
            )}
          </Stack>
        </DialogTitle>

        <DialogContent dividers>
          <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 1.5 }}>
            管理同一分组内不同机场的节点输出顺序，排在前面的机场节点在订阅输出中也排在前面
          </Typography>

          <Grid container spacing={2} sx={{ height: isMobile ? 'auto' : 'calc(70vh - 120px)' }}>
            {/* 左侧：分组列表 */}
            <Grid item xs={12} md={4}>
              <Paper variant="outlined" sx={{ height: isMobile ? 'auto' : '100%', display: 'flex', flexDirection: 'column' }}>
                <Box sx={{ p: 1.5, borderBottom: `1px solid ${theme.palette.divider}` }}>
                  <TextField
                    fullWidth
                    size="small"
                    placeholder="搜索分组..."
                    value={searchText}
                    onChange={(e) => setSearchText(e.target.value)}
                    InputProps={{
                      startAdornment: (
                        <InputAdornment position="start">
                          <SearchIcon fontSize="small" />
                        </InputAdornment>
                      ),
                      endAdornment: searchText && (
                        <InputAdornment position="end" sx={{ cursor: 'pointer' }} onClick={() => setSearchText('')}>
                          <ClearIcon fontSize="small" />
                        </InputAdornment>
                      )
                    }}
                  />
                </Box>
                <List sx={{ overflow: 'auto', flex: 1, py: 0, maxHeight: isMobile ? 200 : 'none' }} dense>
                  {filteredGroups.length === 0 ? (
                    <Box sx={{ p: 3, textAlign: 'center' }}>
                      <Typography variant="body2" color="text.secondary">
                        {groups.length === 0 ? '暂无分组' : '无匹配结果'}
                      </Typography>
                    </Box>
                  ) : (
                    filteredGroups.map((group) => (
                      <ListItem key={group.groupName} disablePadding divider>
                        <ListItemButton
                          selected={selectedGroup === group.groupName}
                          onClick={() => handleSelectGroup(group.groupName)}
                        >
                          <ListItemText
                            primary={
                              <Stack direction="row" alignItems="center" spacing={1}>
                                <Typography variant="body2" noWrap sx={{ flex: 1 }}>
                                  {group.groupName}
                                </Typography>
                                {group.hasSortConfig && (
                                  <CheckCircleIcon sx={{ fontSize: 16, color: 'success.main' }} />
                                )}
                              </Stack>
                            }
                            secondary={`${group.airportCount} 个机场 · ${group.nodeCount} 个节点`}
                          />
                        </ListItemButton>
                      </ListItem>
                    ))
                  )}
                </List>
              </Paper>
            </Grid>

            {/* 右侧：机场排序管理 */}
            <Grid item xs={12} md={8}>
              <Paper variant="outlined" sx={{ height: isMobile ? 'auto' : '100%', display: 'flex', flexDirection: 'column' }}>
                {!selectedGroup ? (
                  <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', flex: 1, p: 4 }}>
                    <Typography variant="body1" color="text.secondary">
                      请从左侧选择一个分组来管理机场排序
                    </Typography>
                  </Box>
                ) : loading ? (
                  <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', flex: 1 }}>
                    <CircularProgress />
                  </Box>
                ) : airports.length === 0 ? (
                  <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', flex: 1, p: 4 }}>
                    <Typography variant="body1" color="text.secondary">
                      该分组下没有机场
                    </Typography>
                  </Box>
                ) : (
                  <Box sx={{ flex: 1, overflow: 'auto' }}>
                    <Box sx={{ p: 2, borderBottom: `1px solid ${theme.palette.divider}` }}>
                      <Typography variant="subtitle1" fontWeight={600}>
                        {selectedGroup}
                      </Typography>
                      <Typography variant="caption" color="text.secondary">
                        拖拽调整机场排序，排在前面的机场节点在订阅输出中也排在前面
                      </Typography>
                    </Box>
                    <DragDropContext onDragEnd={handleDragEnd}>
                      <Droppable droppableId="airportSortList">
                        {(provided) => (
                          <List {...provided.droppableProps} ref={provided.innerRef} sx={{ py: 0 }}>
                            {airports.map((airport, index) => (
                              <Draggable
                                key={`airport-${airport.airportId}`}
                                draggableId={`airport-${airport.airportId}`}
                                index={index}
                              >
                                {(provided, snapshot) => (
                                  <ListItem
                                    ref={provided.innerRef}
                                    {...provided.draggableProps}
                                    {...provided.dragHandleProps}
                                    divider
                                    sx={{
                                      backgroundColor: snapshot.isDragging ? 'action.hover' : 'transparent',
                                      '&:hover': { backgroundColor: 'action.hover' }
                                    }}
                                  >
                                    <Stack direction="row" alignItems="center" spacing={1.5} sx={{ width: '100%' }}>
                                      <DragIndicatorIcon sx={{ color: 'text.secondary', cursor: 'grab' }} />
                                      <Chip
                                        label={index + 1}
                                        size="small"
                                        sx={{ minWidth: 32, fontWeight: 600 }}
                                      />
                                      <Typography variant="body2" sx={{ flex: 1 }} noWrap>
                                        {airport.airportName}
                                      </Typography>
                                      <Chip
                                        label={`${airport.nodeCount} 节点`}
                                        size="small"
                                        variant="outlined"
                                        color="primary"
                                      />
                                    </Stack>
                                  </ListItem>
                                )}
                              </Draggable>
                            ))}
                            {provided.placeholder}
                          </List>
                        )}
                      </Droppable>
                    </DragDropContext>
                  </Box>
                )}
              </Paper>
            </Grid>
          </Grid>
        </DialogContent>

        <DialogActions>
          <Button onClick={onClose}>关闭</Button>
        </DialogActions>
      </Dialog>

      {/* 仅在未传入 showMessage 时使用内部 Snackbar */}
      {!showMessage && (
        <Snackbar
          open={snackbar.open}
          autoHideDuration={3000}
          onClose={() => setSnackbar({ ...snackbar, open: false })}
          anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
        >
          <Alert onClose={() => setSnackbar({ ...snackbar, open: false })} severity={snackbar.severity} sx={{ width: '100%' }}>
            {snackbar.message}
          </Alert>
        </Snackbar>
      )}
    </>
  );
}
