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
import RestartAltIcon from '@mui/icons-material/RestartAlt';
import SaveIcon from '@mui/icons-material/Save';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';

// drag and drop
import { DragDropContext, Droppable, Draggable } from '@hello-pangea/dnd';

// project imports
import { getGroupSortGroups, getGroupSortDetail, saveGroupAirportSort, resetGroupAirportSort } from 'api/groupSort';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';
import { useTranslation } from 'react-i18next';

export default function GroupSortDialog({ open, onClose, showMessage }) {
  const theme = useTheme();
  const { t } = useTranslation();
  const { isDark } = useResolvedColorScheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));
  const { palette, dialogSurface, dialogSurfaceGradient, mutedPanelSurface, nestedPanelSurface, panelBorder } = getSurfaceTokens(
    theme,
    isDark
  );
  const { primaryText, secondaryText, tertiaryText } = getReadableTextTokens(theme, isDark);

  const [groups, setGroups] = useState([]);
  const [selectedGroup, setSelectedGroup] = useState('');
  const [airports, setAirports] = useState([]);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [resetting, setResetting] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });

  const handleSnackbar = useCallback(
    (message, severity) => {
      if (showMessage) {
        showMessage(message, severity);
      } else {
        setSnackbar({ open: true, message, severity });
      }
    },
    [showMessage]
  );

  const loadGroups = useCallback(async () => {
    try {
      const res = await getGroupSortGroups();
      setGroups(res.data || []);
    } catch (err) {
      console.error('Failed to load groups:', err);
      handleSnackbar(t('subscriptions.sort.messages.loadGroupsFailed'), 'error');
    }
  }, [handleSnackbar, t]);

  const loadGroupDetail = useCallback(
    async (groupName) => {
      if (!groupName) return;
      setLoading(true);
      try {
        const res = await getGroupSortDetail(groupName);
        setAirports(res.data?.airports || []);
      } catch (err) {
        console.error('Failed to load group detail:', err);
        handleSnackbar(t('subscriptions.sort.messages.loadGroupDetailFailed'), 'error');
      } finally {
        setLoading(false);
      }
    },
    [handleSnackbar, t]
  );

  useEffect(() => {
    if (open) {
      loadGroups();
      setSelectedGroup('');
      setAirports([]);
      setSearchText('');
    }
  }, [open, loadGroups]);

  const handleSelectGroup = (groupName) => {
    setSelectedGroup(groupName);
    loadGroupDetail(groupName);
  };

  const handleDragEnd = (result) => {
    if (!result.destination) return;
    const items = Array.from(airports);
    const [reorderedItem] = items.splice(result.source.index, 1);
    items.splice(result.destination.index, 0, reorderedItem);
    const updated = items.map((item, index) => ({ ...item, sort: index }));
    setAirports(updated);
  };

  const handleSave = async () => {
    if (!selectedGroup) return;
    setSaving(true);
    try {
      const airportSorts = airports.map((a, index) => ({
        airportId: a.airportId,
        sort: index
      }));
      await saveGroupAirportSort({ groupName: selectedGroup, airportSorts });
      handleSnackbar(t('subscriptions.sort.messages.saveSuccess'), 'success');
      loadGroups();
    } catch {
      handleSnackbar(t('subscriptions.sort.messages.saveFailed'), 'error');
    } finally {
      setSaving(false);
    }
  };

  const handleReset = async () => {
    if (!selectedGroup) return;
    setResetting(true);
    try {
      await resetGroupAirportSort(selectedGroup);
      await loadGroupDetail(selectedGroup);
      await loadGroups();
      handleSnackbar(t('subscriptions.sort.messages.resetSuccess'), 'success');
    } catch {
      handleSnackbar(t('subscriptions.sort.messages.resetFailed'), 'error');
    } finally {
      setResetting(false);
    }
  };

  const filteredGroups = groups.filter((g) => !searchText || g.groupName.toLowerCase().includes(searchText.toLowerCase()));
  const isSortActionDisabled = loading || saving || resetting || airports.length === 0;

  const listItemHoverSurface = withAlpha(palette.primary.main, isDark ? 0.12 : 0.06);
  const listItemSelectedSurface = withAlpha(palette.primary.main, isDark ? 0.18 : 0.1);
  const listItemSelectedHoverSurface = withAlpha(palette.primary.main, isDark ? 0.24 : 0.14);
  const listItemSelectedBorder = withAlpha(palette.primary.main, isDark ? 0.38 : 0.24);
  const listItemHoverBorder = withAlpha(palette.primary.main, isDark ? 0.2 : 0.12);
  const listItemDivider = isDark ? withAlpha(palette.divider, 0.52) : panelBorder;
  const subtleBorder = withAlpha(palette.divider, isDark ? 0.68 : 0.86);
  const neutralSurface = isDark ? withAlpha(palette.background.default, 0.76) : palette.background.paper;
  const searchInsetHighlight = isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.03)}` : 'none';
  const listShellSx = {
    height: isMobile ? 'auto' : '100%',
    display: 'flex',
    flexDirection: 'column',
    borderRadius: 3,
    overflow: 'hidden',
    bgcolor: nestedPanelSurface,
    border: '1px solid',
    borderColor: panelBorder,
    boxShadow: searchInsetHighlight
  };
  const sectionHeaderSx = {
    p: 2,
    bgcolor: mutedPanelSurface,
    borderBottom: '1px solid',
    borderColor: panelBorder,
    boxShadow: `inset 0 -1px 0 ${withAlpha(palette.divider, 0.34)}`
  };
  const searchFieldSx = {
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
          borderColor: withAlpha(palette.primary.main, isDark ? 0.28 : 0.18)
        }
      },
      '&.Mui-focused': {
        bgcolor: dialogSurface,
        boxShadow: isDark ? `0 0 0 1px ${withAlpha(palette.primary.main, 0.16)}` : 'none',
        '& fieldset': {
          borderColor: withAlpha(palette.primary.main, isDark ? 0.38 : 0.24)
        }
      }
    },
    '& .MuiInputBase-input::placeholder': {
      color: isDark ? withAlpha(palette.text.secondary, 0.92) : undefined,
      opacity: 1
    }
  };
  const dragRowSurface = withAlpha(palette.primary.main, isDark ? 0.12 : 0.06);
  const draggingSurface = withAlpha(palette.primary.main, isDark ? 0.18 : 0.1);
  const draggingBorder = withAlpha(palette.primary.main, isDark ? 0.42 : 0.28);
  const indexChipSx = {
    minWidth: 32,
    fontWeight: 700,
    bgcolor: withAlpha(palette.primary.main, isDark ? 0.14 : 0.08),
    color: palette.primary.main,
    border: '1px solid',
    borderColor: withAlpha(palette.primary.main, isDark ? 0.28 : 0.18)
  };

  return (
    <>
      <Dialog
        open={open}
        onClose={onClose}
        maxWidth="md"
        fullWidth
        fullScreen={isMobile}
        slotProps={{
          paper: {
            sx: {
              borderRadius: isMobile ? 0 : 3,
              overflow: 'hidden',
              m: isMobile ? 0 : 3,
              bgcolor: dialogSurface,
              backgroundImage: dialogSurfaceGradient,
              border: '1px solid',
              borderColor: panelBorder
            }
          }
        }}
      >
        <DialogTitle sx={sectionHeaderSx}>
          <Stack direction="row" alignItems="center" justifyContent="space-between" spacing={2} sx={{ flexWrap: 'wrap', gap: 1 }}>
            <Typography variant="h6" sx={{ color: primaryText, fontWeight: 700 }}>
              {t('subscriptions.sort.groupSort')}
            </Typography>
            {selectedGroup && (
              <Stack direction="row" spacing={1} alignItems="center">
                <Button
                  variant="outlined"
                  size="small"
                  startIcon={<RestartAltIcon />}
                  onClick={handleReset}
                  disabled={isSortActionDisabled}
                  sx={{ borderColor: panelBorder, color: secondaryText }}
                >
                  {resetting ? t('subscriptions.sort.resetting') : t('subscriptions.sort.resetSort')}
                </Button>
                <Button variant="contained" size="small" startIcon={<SaveIcon />} onClick={handleSave} disabled={isSortActionDisabled}>
                  {saving ? t('subscriptions.sort.saving') : t('subscriptions.sort.saveSort')}
                </Button>
              </Stack>
            )}
          </Stack>
        </DialogTitle>

        <DialogContent sx={{ p: 2, bgcolor: dialogSurface }}>
          <Typography variant="caption" sx={{ display: 'block', mb: 1.5, color: tertiaryText }}>
            {t('subscriptions.sort.groupSortDesc')}
          </Typography>

          <Grid container spacing={2} sx={{ height: isMobile ? 'auto' : 'calc(70vh - 120px)' }}>
            <Grid item xs={12} md={4}>
              <Paper variant="outlined" sx={listShellSx}>
                <Box sx={{ ...sectionHeaderSx, p: 1.5 }}>
                  <TextField
                    fullWidth
                    size="small"
                    placeholder={t('subscriptions.sort.searchGroup')}
                    value={searchText}
                    onChange={(e) => setSearchText(e.target.value)}
                    sx={searchFieldSx}
                    slotProps={{
                      input: {
                        startAdornment: (
                          <InputAdornment position="start">
                            <SearchIcon fontSize="small" sx={{ color: secondaryText }} />
                          </InputAdornment>
                        ),
                        endAdornment: searchText && (
                          <InputAdornment position="end" sx={{ cursor: 'pointer', color: secondaryText }} onClick={() => setSearchText('')}>
                            <ClearIcon fontSize="small" />
                          </InputAdornment>
                        )
                      }
                    }}
                  />
                </Box>
                <List sx={{ overflow: 'auto', flex: 1, py: 0, maxHeight: isMobile ? 200 : 'none', bgcolor: nestedPanelSurface }} dense>
                  {filteredGroups.length === 0 ? (
                    <Box sx={{ p: 3, textAlign: 'center' }}>
                      <Typography variant="body2" sx={{ color: secondaryText }}>
                        {groups.length === 0 ? t('subscriptions.sort.noGroups') : t('common.noMatch')}
                      </Typography>
                    </Box>
                  ) : (
                    filteredGroups.map((group) => (
                      <ListItem key={group.groupName} disablePadding divider sx={{ borderColor: listItemDivider }}>
                        <ListItemButton
                          selected={selectedGroup === group.groupName}
                          onClick={() => handleSelectGroup(group.groupName)}
                          sx={{
                            alignItems: 'flex-start',
                            px: 2,
                            py: 1.75,
                            color: primaryText,
                            borderLeft: '2px solid transparent',
                            transition: 'background-color 0.2s ease, border-color 0.2s ease',
                            '& .MuiTypography-root': {
                              color: 'inherit'
                            },
                            '& .MuiListItemText-primary': {
                              color: primaryText,
                              fontWeight: selectedGroup === group.groupName ? 700 : 600
                            },
                            '& .MuiListItemText-secondary': {
                              color: selectedGroup === group.groupName ? secondaryText : tertiaryText
                            },
                            '&:hover': {
                              bgcolor: listItemHoverSurface,
                              borderLeftColor: listItemHoverBorder,
                              color: primaryText,
                              '& .MuiListItemText-primary': {
                                color: primaryText
                              },
                              '& .MuiListItemText-secondary': {
                                color: secondaryText
                              }
                            },
                            '&.Mui-selected': {
                              bgcolor: listItemSelectedSurface,
                              borderLeftColor: listItemSelectedBorder,
                              color: primaryText,
                              '& .MuiListItemText-primary': {
                                color: primaryText
                              },
                              '& .MuiListItemText-secondary': {
                                color: secondaryText
                              }
                            },
                            '&.Mui-selected:hover': {
                              bgcolor: listItemSelectedHoverSurface,
                              color: primaryText,
                              '& .MuiListItemText-primary': {
                                color: primaryText
                              },
                              '& .MuiListItemText-secondary': {
                                color: secondaryText
                              }
                            }
                          }}
                        >
                          <ListItemText
                            primary={
                              <Stack direction="row" alignItems="center" spacing={1}>
                                <Typography variant="body2" noWrap sx={{ flex: 1, color: 'inherit', fontWeight: 'inherit' }}>
                                  {group.groupName}
                                </Typography>
                                {group.hasSortConfig && <CheckCircleIcon sx={{ fontSize: 16, color: 'success.main' }} />}
                              </Stack>
                            }
                            secondary={t('subscriptions.sort.airportNodeCount', {
                              airportCount: group.airportCount,
                              nodeCount: group.nodeCount
                            })}
                          />
                        </ListItemButton>
                      </ListItem>
                    ))
                  )}
                </List>
              </Paper>
            </Grid>

            <Grid item xs={12} md={8}>
              <Paper variant="outlined" sx={listShellSx}>
                {!selectedGroup ? (
                  <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', flex: 1, p: 4 }}>
                    <Typography variant="body1" sx={{ color: secondaryText }}>
                      {t('subscriptions.sort.selectGroupPrompt')}
                    </Typography>
                  </Box>
                ) : loading ? (
                  <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', flex: 1 }}>
                    <CircularProgress />
                  </Box>
                ) : airports.length === 0 ? (
                  <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', flex: 1, p: 4 }}>
                    <Typography variant="body1" sx={{ color: secondaryText }}>
                      {t('subscriptions.sort.noAirportsInGroup')}
                    </Typography>
                  </Box>
                ) : (
                  <Box sx={{ flex: 1, overflow: 'auto' }}>
                    <Box sx={sectionHeaderSx}>
                      <Typography variant="subtitle1" sx={{ fontWeight: 700, color: primaryText }}>
                        {selectedGroup}
                      </Typography>
                      <Typography variant="caption" sx={{ color: secondaryText }}>
                        {t('subscriptions.sort.dragAirportHint')}
                      </Typography>
                    </Box>
                    <DragDropContext onDragEnd={handleDragEnd}>
                      <Droppable droppableId="airportSortList">
                        {(provided) => (
                          <List {...provided.droppableProps} ref={provided.innerRef} sx={{ py: 0, bgcolor: nestedPanelSurface }}>
                            {airports.map((airport, index) => (
                              <Draggable key={`airport-${airport.airportId}`} draggableId={`airport-${airport.airportId}`} index={index}>
                                {(provided, snapshot) => (
                                  <ListItem
                                    ref={provided.innerRef}
                                    {...provided.draggableProps}
                                    {...provided.dragHandleProps}
                                    disableGutters
                                    sx={{
                                      mx: 1,
                                      mt: 1,
                                      mb: 0,
                                      px: 2,
                                      py: 1.5,
                                      bgcolor: snapshot.isDragging ? draggingSurface : neutralSurface,
                                      border: '1px solid',
                                      borderColor: snapshot.isDragging ? draggingBorder : subtleBorder,
                                      borderRadius: 2,
                                      transition:
                                        'background-color 0.2s ease, border-color 0.2s ease, box-shadow 0.2s ease, transform 0.2s ease',
                                      ...(snapshot.isDragging && {
                                        boxShadow: theme.shadows[6],
                                        transform: 'rotate(0.25deg)'
                                      }),
                                      '&:hover': {
                                        bgcolor: snapshot.isDragging ? draggingSurface : dragRowSurface,
                                        borderColor: snapshot.isDragging ? draggingBorder : listItemHoverBorder
                                      }
                                    }}
                                  >
                                    <Stack direction="row" alignItems="center" spacing={1.5} sx={{ width: '100%' }}>
                                      <DragIndicatorIcon sx={{ color: secondaryText, cursor: 'grab' }} />
                                      <Chip label={index + 1} size="small" sx={indexChipSx} />
                                      <Typography variant="body2" sx={{ flex: 1, color: primaryText, fontWeight: 600 }} noWrap>
                                        {airport.airportName}
                                      </Typography>
                                      <Chip
                                        label={t('subscriptions.sort.nodeCount', { count: airport.nodeCount })}
                                        size="small"
                                        variant="outlined"
                                        color="primary"
                                        sx={{
                                          bgcolor: withAlpha(palette.primary.main, isDark ? 0.1 : 0.05),
                                          borderColor: withAlpha(palette.primary.main, isDark ? 0.28 : 0.18)
                                        }}
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

        <DialogActions
          sx={{
            px: 2,
            py: 1.5,
            bgcolor: mutedPanelSurface,
            borderTop: '1px solid',
            borderColor: panelBorder
          }}
        >
          <Button onClick={onClose} variant="outlined" sx={{ borderColor: panelBorder, color: secondaryText }}>
            {t('common.close')}
          </Button>
        </DialogActions>
      </Dialog>

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
