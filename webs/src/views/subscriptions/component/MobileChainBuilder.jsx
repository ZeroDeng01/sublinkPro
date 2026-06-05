import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useTheme } from '@mui/material/styles';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Typography from '@mui/material/Typography';
import IconButton from '@mui/material/IconButton';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import Chip from '@mui/material/Chip';
import Autocomplete from '@mui/material/Autocomplete';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import ToggleButton from '@mui/material/ToggleButton';
import ToggleButtonGroup from '@mui/material/ToggleButtonGroup';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import {
  getChainProxyFieldControlSx,
  getChainProxyIconButtonSx,
  getChainProxyThemeTokens,
  getChainProxyToggleButtonGroupSx
} from './chainProxyTheme';
import { withAlpha } from '../../../utils/colorUtils';

import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import ArrowForwardIcon from '@mui/icons-material/ArrowForward';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import StopIcon from '@mui/icons-material/Stop';
import GroupWorkIcon from '@mui/icons-material/GroupWork';
import FilterAltIcon from '@mui/icons-material/FilterAlt';
import DeviceHubIcon from '@mui/icons-material/DeviceHub';

import ConditionBuilder from './ConditionBuilder';

export default function MobileChainBuilder({
  chainConfig = [],
  targetConfig = { type: 'all', conditions: null },
  onChainConfigChange,
  onTargetConfigChange,
  nodes: availableNodes = [],
  fields = [],
  operators = [],
  groupTypes = [],
  templateGroups = []
}) {
  const theme = useTheme();
  const { t } = useTranslation();
  const { isDark } = useResolvedColorScheme();
  const tokens = getChainProxyThemeTokens(theme, isDark);
  const { dialogSurface, dialogSurfaceGradient, mutedPanelSurface, nestedPanelSurface, panelBorder } = tokens;
  const fieldControlSx = getChainProxyFieldControlSx(tokens);
  const iconButtonSx = getChainProxyIconButtonSx(tokens);
  const errorIconButtonSx = getChainProxyIconButtonSx(tokens, theme.palette.error.main);
  const targetToggleSx = getChainProxyToggleButtonGroupSx(tokens, theme.palette.success.main);
  const [proxyDialogOpen, setProxyDialogOpen] = useState(false);
  const [targetDialogOpen, setTargetDialogOpen] = useState(false);
  const [editingProxyConfig, setEditingProxyConfig] = useState(null);
  const [editingTargetConfig, setEditingTargetConfig] = useState(null);

  const getTypeLabel = (type) => {
    const labels = t('subscriptions.chain.proxyTypeShort', { returnObjects: true });
    return labels[type] || t('subscriptions.chain.proxy');
  };

  const getTypeIcon = (type) => {
    switch (type) {
      case 'template_group':
        return <GroupWorkIcon color="primary" />;
      case 'custom_group':
        return <DeviceHubIcon color="secondary" />;
      case 'dynamic_node':
        return <FilterAltIcon color="warning" />;
      case 'specified_node':
        return <DeviceHubIcon color="success" />;
      default:
        return <GroupWorkIcon color="primary" />;
    }
  };

  const getProxyLabel = (item) => {
    if (!item) return t('subscriptions.chain.unconfigured');
    if (item.type === 'specified_node') {
      const node = availableNodes.find((n) => n.id === item.nodeId);
      return node?.name || node?.linkName || t('subscriptions.chain.nodeNumber', { id: item.nodeId });
    }
    if (item.type === 'dynamic_node') {
      const condCount = item.nodeConditions?.conditions?.length || 0;
      return condCount > 0 ? t('subscriptions.chain.conditionCount', { count: condCount }) : t('subscriptions.chain.unconfigured');
    }
    return item.groupName || t('subscriptions.chain.unconfigured');
  };

  const getTargetLabel = () => {
    switch (targetConfig?.type) {
      case 'all':
        return t('subscriptions.chain.targetTypes.all');
      case 'specified_node':
        if (targetConfig?.nodeId) {
          const node = availableNodes.find((n) => n.id === targetConfig.nodeId);
          return node?.name || node?.linkName || t('subscriptions.chain.nodeNumber', { id: targetConfig.nodeId });
        }
        return t('subscriptions.chain.noNodeSelected');
      case 'conditions':
        const condCount = targetConfig?.conditions?.conditions?.length || 0;
        return condCount > 0 ? t('subscriptions.chain.conditionCount', { count: condCount }) : t('subscriptions.chain.unconfigured');
      default:
        return t('subscriptions.chain.unconfigured');
    }
  };

  const handleAddProxy = () => {
    const isEntryNode = chainConfig.length === 0;
    const defaultType = isEntryNode ? 'template_group' : 'custom_group';
    const newConfig = { type: defaultType, groupName: '' };
    setEditingProxyConfig({ isNew: true, index: chainConfig.length, config: newConfig });
    setProxyDialogOpen(true);
  };

  const handleEditProxy = (index) => {
    const config = { ...chainConfig[index] };
    if (index > 0 && config.type === 'template_group') {
      config.type = 'custom_group';
    }
    setEditingProxyConfig({ isNew: false, index, config });
    setProxyDialogOpen(true);
  };

  const handleDeleteProxy = (index) => {
    const newConfig = chainConfig.filter((_, i) => i !== index);
    onChainConfigChange?.(newConfig);
  };

  const handleSaveProxy = () => {
    if (!editingProxyConfig) return;
    const newChainConfig = [...chainConfig];
    if (editingProxyConfig.isNew) {
      newChainConfig.push(editingProxyConfig.config);
    } else {
      newChainConfig[editingProxyConfig.index] = editingProxyConfig.config;
    }
    onChainConfigChange?.(newChainConfig);
    setProxyDialogOpen(false);
    setEditingProxyConfig(null);
  };

  const handleEditTarget = () => {
    setEditingTargetConfig({ ...targetConfig });
    setTargetDialogOpen(true);
  };

  const handleSaveTarget = () => {
    if (!editingTargetConfig) return;
    onTargetConfigChange?.(editingTargetConfig);
    setTargetDialogOpen(false);
    setEditingTargetConfig(null);
  };

  return (
    <Box>
      <Stack spacing={1.5}>
        <Card
          variant="outlined"
          sx={{
            backgroundColor: tokens.elevatedSurface,
            backgroundImage: tokens.dialogSurfaceGradient,
            color: 'primary.main',
            borderRadius: 2,
            borderColor: tokens.primarySoftBorder,
            boxShadow: tokens.insetHighlight
          }}
        >
          <CardContent sx={{ py: 1.5, px: 2, '&:last-child': { pb: 1.5 } }}>
            <Stack direction="row" alignItems="center" spacing={1}>
              <PlayArrowIcon />
              <Typography variant="subtitle2" fontWeight={600} color="inherit">
                {t('subscriptions.chain.entry')}
              </Typography>
            </Stack>
          </CardContent>
        </Card>

        {(chainConfig.length > 0 || true) && (
          <Box sx={{ display: 'flex', justifyContent: 'center' }}>
            <ArrowForwardIcon sx={{ color: 'text.secondary', transform: 'rotate(90deg)' }} />
          </Box>
        )}

        {chainConfig.map((item, index) => (
          <Box key={index}>
            <Card
              variant="outlined"
              sx={{
                borderRadius: 2,
                borderColor: tokens.primarySoftBorder,
                bgcolor: nestedPanelSurface,
                boxShadow: tokens.cardShadow
              }}
            >
              <CardContent sx={{ py: 1.5, px: 2, '&:last-child': { pb: 1.5 } }}>
                <Stack direction="row" alignItems="center" justifyContent="space-between">
                  <Stack direction="row" alignItems="center" spacing={1.5}>
                    {getTypeIcon(item.type)}
                    <Box>
                      <Typography variant="caption" sx={{ color: tokens.secondaryText }}>
                        {getTypeLabel(item.type)}
                      </Typography>
                      <Typography variant="body2" fontWeight={600} sx={{ color: tokens.primaryText }}>
                        {getProxyLabel(item)}
                      </Typography>
                    </Box>
                  </Stack>
                  <Stack direction="row" spacing={0.5}>
                    <IconButton size="small" onClick={() => handleEditProxy(index)} sx={iconButtonSx}>
                      <EditIcon fontSize="small" />
                    </IconButton>
                    <IconButton
                      size="small"
                      onClick={() => handleDeleteProxy(index)}
                      sx={{ ...errorIconButtonSx, color: theme.palette.error.main }}
                    >
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  </Stack>
                </Stack>
              </CardContent>
            </Card>
            <Box sx={{ display: 'flex', justifyContent: 'center', my: 0.5 }}>
              <ArrowForwardIcon sx={{ color: 'text.secondary', transform: 'rotate(90deg)' }} />
            </Box>
          </Box>
        ))}

        {chainConfig.length < 4 && (
          <>
            <Button
              variant="outlined"
              startIcon={<AddIcon />}
              onClick={handleAddProxy}
              fullWidth
              sx={{
                borderStyle: 'dashed',
                py: 1.5,
                borderColor: tokens.primarySoftBorder,
                backgroundColor: tokens.fieldSurface,
                boxShadow: tokens.insetHighlight,
                color: tokens.primaryText,
                '&:hover': {
                  borderColor: tokens.primaryStrongBorder,
                  backgroundColor: tokens.hoverSurface
                }
              }}
            >
              {chainConfig.length === 0 ? t('subscriptions.chain.addEntryProxy') : t('subscriptions.chain.addMiddleProxy')}
            </Button>
            {chainConfig.length >= 2 && (
              <Typography variant="caption" sx={{ textAlign: 'center', display: 'block', mt: 0.5, color: tokens.warningSoftText }}>
                {t('subscriptions.chain.currentLevelWarning', { count: chainConfig.length })}
              </Typography>
            )}
            <Box sx={{ display: 'flex', justifyContent: 'center' }}>
              <ArrowForwardIcon sx={{ color: 'text.secondary', transform: 'rotate(90deg)' }} />
            </Box>
          </>
        )}
        {chainConfig.length >= 4 && (
          <Typography variant="caption" sx={{ textAlign: 'center', display: 'block', py: 1, color: tokens.secondaryText }}>
            {t('subscriptions.chain.maxFourLevels')}
          </Typography>
        )}

        <Card
          variant="outlined"
          onClick={handleEditTarget}
          sx={{
            color: 'success.main',
            borderRadius: 2,
            cursor: 'pointer',
            borderColor: withAlpha(theme.palette.success.main, isDark ? 0.34 : 0.22),
            bgcolor: nestedPanelSurface,
            boxShadow: tokens.cardShadow,
            transition: 'transform 0.2s',
            '&:active': { transform: 'scale(0.98)' },
            '&:hover': {
              backgroundColor: tokens.successSurface,
              borderColor: withAlpha(theme.palette.success.main, isDark ? 0.44 : 0.3)
            }
          }}
        >
          <CardContent sx={{ py: 1.5, px: 2, '&:last-child': { pb: 1.5 } }}>
            <Stack direction="row" alignItems="center" justifyContent="space-between">
              <Stack direction="row" alignItems="center" spacing={1}>
                <StopIcon />
                <Box>
                  <Typography variant="caption" sx={{ opacity: 0.9, color: tokens.secondaryText }}>
                    {t('subscriptions.chain.targetNode')}
                  </Typography>
                  <Typography variant="subtitle2" fontWeight={600} sx={{ color: tokens.primaryText }}>
                    {getTargetLabel()}
                  </Typography>
                </Box>
              </Stack>
              <EditIcon fontSize="small" sx={{ opacity: 0.8, color: tokens.secondaryText }} />
            </Stack>
          </CardContent>
        </Card>
      </Stack>

      <Dialog
        open={proxyDialogOpen}
        onClose={() => setProxyDialogOpen(false)}
        fullWidth
        maxWidth="sm"
        PaperProps={{
          sx: {
            border: '1px solid',
            borderColor: panelBorder,
            bgcolor: dialogSurface,
            backgroundImage: dialogSurfaceGradient
          }
        }}
      >
        <DialogTitle sx={{ bgcolor: mutedPanelSurface, borderBottom: '1px solid', borderColor: panelBorder, color: tokens.primaryText }}>
          {editingProxyConfig?.isNew
            ? editingProxyConfig?.index === 0
              ? t('subscriptions.chain.addEntryProxy')
              : t('subscriptions.chain.addMiddleProxy')
            : editingProxyConfig?.index === 0
              ? t('subscriptions.chain.editEntryProxy')
              : t('subscriptions.chain.editMiddleProxy')}
        </DialogTitle>
        <DialogContent sx={{ bgcolor: dialogSurface }}>
          <Stack spacing={2} sx={{ pt: 1 }}>
            {(() => {
              const isEntryNode = editingProxyConfig?.index === 0;
              return (
                <FormControl size="small" fullWidth sx={fieldControlSx}>
                  <InputLabel>{t('subscriptions.chain.proxyType')}</InputLabel>
                  <Select
                    value={editingProxyConfig?.config?.type || (isEntryNode ? 'template_group' : 'specified_node')}
                    label={t('subscriptions.chain.proxyType')}
                    onChange={(e) =>
                      setEditingProxyConfig({
                        ...editingProxyConfig,
                        config: {
                          type: e.target.value,
                          groupName: '',
                          nodeId: undefined,
                          nodeConditions: undefined
                        }
                      })
                    }
                  >
                    {isEntryNode && <MenuItem value="template_group">{t('subscriptions.chain.proxyTypes.template_group')}</MenuItem>}
                    <MenuItem value="custom_group">{t('subscriptions.chain.proxyTypes.custom_group')}</MenuItem>
                    <MenuItem value="dynamic_node">{t('subscriptions.chain.proxyTypes.dynamic_node')}</MenuItem>
                    <MenuItem value="specified_node">{t('subscriptions.chain.proxyTypes.specified_node')}</MenuItem>
                  </Select>
                  {!isEntryNode && (
                    <Typography variant="caption" sx={{ mt: 0.5, color: tokens.tertiaryText }}>
                      {t('subscriptions.chain.mobileDialerProxyHint')}
                    </Typography>
                  )}
                </FormControl>
              );
            })()}

            {editingProxyConfig?.config?.type === 'template_group' && (
              <Autocomplete
                freeSolo
                size="small"
                fullWidth
                sx={fieldControlSx}
                options={templateGroups || []}
                value={editingProxyConfig?.config?.groupName || ''}
                onChange={(_event, newValue) =>
                  setEditingProxyConfig({
                    ...editingProxyConfig,
                    config: { ...editingProxyConfig.config, groupName: newValue || '' }
                  })
                }
                onInputChange={(_event, newValue) =>
                  setEditingProxyConfig({
                    ...editingProxyConfig,
                    config: { ...editingProxyConfig.config, groupName: newValue || '' }
                  })
                }
                renderInput={(params) => (
                  <TextField
                    {...params}
                    label={t('subscriptions.chain.groupName')}
                    placeholder={t('subscriptions.chain.selectOrInputGroupPlaceholder')}
                  />
                )}
              />
            )}

            {editingProxyConfig?.config?.type === 'custom_group' && (
              <>
                <TextField
                  size="small"
                  fullWidth
                  label={t('subscriptions.chain.groupName')}
                  placeholder={t('subscriptions.chain.customGroupPlaceholder')}
                  value={editingProxyConfig?.config?.groupName || ''}
                  onChange={(e) =>
                    setEditingProxyConfig({
                      ...editingProxyConfig,
                      config: { ...editingProxyConfig.config, groupName: e.target.value }
                    })
                  }
                  sx={fieldControlSx}
                />
                <FormControl size="small" fullWidth sx={fieldControlSx}>
                  <InputLabel>{t('subscriptions.chain.groupType')}</InputLabel>
                  <Select
                    value={editingProxyConfig?.config?.groupType || 'select'}
                    label={t('subscriptions.chain.groupType')}
                    onChange={(e) =>
                      setEditingProxyConfig({
                        ...editingProxyConfig,
                        config: { ...editingProxyConfig.config, groupType: e.target.value }
                      })
                    }
                  >
                    {(groupTypes || []).map((gt) => (
                      <MenuItem key={gt.value} value={gt.value}>
                        {gt.label}
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>
                {(editingProxyConfig?.config?.groupType === 'url-test' || editingProxyConfig?.config?.groupType === 'fallback') && (
                  <Stack spacing={1.5}>
                    <TextField
                      size="small"
                      fullWidth
                      label={t('subscriptions.chain.testUrl')}
                      value={editingProxyConfig?.config?.urlTestConfig?.url || ''}
                      onChange={(e) =>
                        setEditingProxyConfig({
                          ...editingProxyConfig,
                          config: {
                            ...editingProxyConfig.config,
                            urlTestConfig: { ...editingProxyConfig.config.urlTestConfig, url: e.target.value }
                          }
                        })
                      }
                      placeholder="http://www.gstatic.com/generate_204"
                      helperText={t('subscriptions.chain.testUrlHelper')}
                      sx={fieldControlSx}
                    />
                    <Stack direction="row" spacing={1}>
                      <TextField
                        size="small"
                        label={t('subscriptions.chain.intervalSeconds')}
                        type="number"
                        value={editingProxyConfig?.config?.urlTestConfig?.interval ?? 300}
                        onChange={(e) =>
                          setEditingProxyConfig({
                            ...editingProxyConfig,
                            config: {
                              ...editingProxyConfig.config,
                              urlTestConfig: { ...editingProxyConfig.config.urlTestConfig, interval: parseInt(e.target.value) || 300 }
                            }
                          })
                        }
                        sx={{ ...fieldControlSx, flex: 1 }}
                        helperText={t('subscriptions.chain.healthCheckInterval')}
                      />
                      <TextField
                        size="small"
                        label={t('subscriptions.chain.toleranceMs')}
                        type="number"
                        value={editingProxyConfig?.config?.urlTestConfig?.tolerance ?? 50}
                        onChange={(e) =>
                          setEditingProxyConfig({
                            ...editingProxyConfig,
                            config: {
                              ...editingProxyConfig.config,
                              urlTestConfig: { ...editingProxyConfig.config.urlTestConfig, tolerance: parseInt(e.target.value) || 50 }
                            }
                          })
                        }
                        sx={{ ...fieldControlSx, flex: 1 }}
                        helperText={
                          editingProxyConfig?.config?.groupType === 'url-test'
                            ? t('subscriptions.chain.toleranceHelper')
                            : t('subscriptions.chain.fallbackThreshold')
                        }
                      />
                    </Stack>
                  </Stack>
                )}
                {editingProxyConfig?.config?.groupType === 'load-balance' && (
                  <Stack spacing={1.5}>
                    <TextField
                      size="small"
                      fullWidth
                      label={t('subscriptions.chain.testUrl')}
                      value={editingProxyConfig?.config?.urlTestConfig?.url || ''}
                      onChange={(e) =>
                        setEditingProxyConfig({
                          ...editingProxyConfig,
                          config: {
                            ...editingProxyConfig.config,
                            urlTestConfig: { ...editingProxyConfig.config.urlTestConfig, url: e.target.value }
                          }
                        })
                      }
                      placeholder="http://www.gstatic.com/generate_204"
                      helperText={t('subscriptions.chain.testUrlHelper')}
                      sx={fieldControlSx}
                    />
                    <Stack direction="row" spacing={1}>
                      <TextField
                        size="small"
                        label={t('subscriptions.chain.intervalSeconds')}
                        type="number"
                        value={editingProxyConfig?.config?.urlTestConfig?.interval ?? 300}
                        onChange={(e) =>
                          setEditingProxyConfig({
                            ...editingProxyConfig,
                            config: {
                              ...editingProxyConfig.config,
                              urlTestConfig: { ...editingProxyConfig.config.urlTestConfig, interval: parseInt(e.target.value) || 300 }
                            }
                          })
                        }
                        sx={{ ...fieldControlSx, flex: 1 }}
                        helperText={t('subscriptions.chain.healthCheckInterval')}
                      />
                      <FormControl size="small" sx={{ ...fieldControlSx, flex: 1 }}>
                        <InputLabel>{t('subscriptions.chain.loadBalanceStrategy')}</InputLabel>
                        <Select
                          value={editingProxyConfig?.config?.urlTestConfig?.strategy || 'consistent-hashing'}
                          label={t('subscriptions.chain.loadBalanceStrategy')}
                          onChange={(e) =>
                            setEditingProxyConfig({
                              ...editingProxyConfig,
                              config: {
                                ...editingProxyConfig.config,
                                urlTestConfig: { ...editingProxyConfig.config.urlTestConfig, strategy: e.target.value }
                              }
                            })
                          }
                        >
                          <MenuItem value="consistent-hashing">{t('subscriptions.chain.strategies.consistentHashing')}</MenuItem>
                          <MenuItem value="round-robin">{t('subscriptions.chain.strategies.roundRobin')}</MenuItem>
                          <MenuItem value="sticky-sessions">{t('subscriptions.chain.strategies.stickySessions')}</MenuItem>
                        </Select>
                      </FormControl>
                    </Stack>
                  </Stack>
                )}
                <ConditionBuilder
                  title={t('subscriptions.chain.nodeFilterConditions')}
                  value={editingProxyConfig?.config?.nodeConditions}
                  onChange={(conds) =>
                    setEditingProxyConfig({
                      ...editingProxyConfig,
                      config: { ...editingProxyConfig.config, nodeConditions: conds }
                    })
                  }
                  fields={fields}
                  operators={operators}
                />
              </>
            )}

            {editingProxyConfig?.config?.type === 'dynamic_node' && (
              <>
                <FormControl size="small" fullWidth sx={fieldControlSx}>
                  <InputLabel>{t('subscriptions.chain.selectMode')}</InputLabel>
                  <Select
                    value={editingProxyConfig?.config?.selectMode || 'first'}
                    label={t('subscriptions.chain.selectMode')}
                    onChange={(e) =>
                      setEditingProxyConfig({
                        ...editingProxyConfig,
                        config: { ...editingProxyConfig.config, selectMode: e.target.value }
                      })
                    }
                  >
                    <MenuItem value="first">{t('subscriptions.chain.selectModes.first')}</MenuItem>
                    <MenuItem value="random">{t('subscriptions.chain.selectModes.random')}</MenuItem>
                    <MenuItem value="fastest">{t('subscriptions.chain.selectModes.fastest')}</MenuItem>
                  </Select>
                </FormControl>
                <ConditionBuilder
                  title={t('subscriptions.chain.nodeMatchConditions')}
                  value={editingProxyConfig?.config?.nodeConditions}
                  onChange={(conds) =>
                    setEditingProxyConfig({
                      ...editingProxyConfig,
                      config: { ...editingProxyConfig.config, nodeConditions: conds }
                    })
                  }
                  fields={fields}
                  operators={operators}
                />
              </>
            )}

            {editingProxyConfig?.config?.type === 'specified_node' && (
              <Autocomplete
                size="small"
                options={availableNodes || []}
                sx={fieldControlSx}
                getOptionLabel={(option) => `${option.name || option.linkName} (${option.linkCountry || t('common.unknown')})`}
                value={(availableNodes || []).find((n) => n.id === editingProxyConfig?.config?.nodeId) || null}
                onChange={(_event, newValue) =>
                  setEditingProxyConfig({
                    ...editingProxyConfig,
                    config: { ...editingProxyConfig.config, nodeId: newValue?.id }
                  })
                }
                renderInput={(params) => <TextField {...params} label={t('subscriptions.chain.selectNode')} />}
                renderOption={(props, option) => (
                  <li {...props} key={option.id}>
                    <Stack direction="row" spacing={1} alignItems="center">
                      <Typography variant="body2">{option.name || option.linkName}</Typography>
                      <Chip label={option.linkCountry || t('common.unknown')} size="small" variant="outlined" />
                    </Stack>
                  </li>
                )}
              />
            )}
          </Stack>
        </DialogContent>
        <DialogActions sx={{ bgcolor: mutedPanelSurface, borderTop: '1px solid', borderColor: panelBorder }}>
          <Button onClick={() => setProxyDialogOpen(false)}>{t('common.cancel')}</Button>
          <Button variant="contained" onClick={handleSaveProxy}>
            {t('common.confirm')}
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog
        open={targetDialogOpen}
        onClose={() => setTargetDialogOpen(false)}
        fullWidth
        maxWidth="sm"
        PaperProps={{
          sx: {
            border: '1px solid',
            borderColor: panelBorder,
            bgcolor: dialogSurface,
            backgroundImage: dialogSurfaceGradient
          }
        }}
      >
        <DialogTitle sx={{ bgcolor: mutedPanelSurface, borderBottom: '1px solid', borderColor: panelBorder, color: tokens.primaryText }}>
          {t('subscriptions.chain.targetNodeConfig')}
        </DialogTitle>
        <DialogContent sx={{ bgcolor: dialogSurface }}>
          <Stack spacing={2} sx={{ pt: 1 }}>
            <Typography variant="body2" sx={{ color: tokens.secondaryText }}>
              {t('subscriptions.chain.targetScopeDescription')}
            </Typography>

            <ToggleButtonGroup
              value={editingTargetConfig?.type || 'specified_node'}
              exclusive
              fullWidth
              size="small"
              onChange={(_event, newType) => {
                if (newType !== null) {
                  setEditingTargetConfig({
                    ...editingTargetConfig,
                    type: newType,
                    nodeId: undefined,
                    conditions: undefined
                  });
                }
              }}
              sx={{ ...targetToggleSx, '& .MuiToggleButton-root': { minHeight: 40 } }}
            >
              <ToggleButton value="specified_node">{t('subscriptions.chain.targetTypes.specified_node')}</ToggleButton>
              <ToggleButton value="all">{t('subscriptions.chain.targetTypes.all')}</ToggleButton>
              <ToggleButton value="conditions">{t('subscriptions.chain.targetTypes.conditions')}</ToggleButton>
            </ToggleButtonGroup>

            {editingTargetConfig?.type === 'specified_node' && (
              <Autocomplete
                size="small"
                options={availableNodes || []}
                sx={fieldControlSx}
                getOptionLabel={(option) => `${option.name || option.linkName} (${option.linkCountry || t('common.unknown')})`}
                value={(availableNodes || []).find((n) => n.id === editingTargetConfig?.nodeId) || null}
                onChange={(_event, newValue) =>
                  setEditingTargetConfig({
                    ...editingTargetConfig,
                    nodeId: newValue?.id
                  })
                }
                renderInput={(params) => (
                  <TextField
                    {...params}
                    label={t('subscriptions.chain.selectTargetNode')}
                    placeholder={t('subscriptions.chain.searchNodes')}
                  />
                )}
                renderOption={(props, option) => (
                  <li {...props} key={option.id}>
                    <Stack direction="row" spacing={1} alignItems="center">
                      <Typography variant="body2">{option.name || option.linkName}</Typography>
                      <Chip label={option.linkCountry || t('common.unknown')} size="small" variant="outlined" />
                      <Chip label={option.protocol || 'unknown'} size="small" color="info" variant="outlined" />
                    </Stack>
                  </li>
                )}
              />
            )}

            {editingTargetConfig?.type === 'conditions' && (
              <ConditionBuilder
                title={t('subscriptions.chain.targetNodeFilterConditions')}
                value={editingTargetConfig?.conditions}
                onChange={(conds) =>
                  setEditingTargetConfig({
                    ...editingTargetConfig,
                    conditions: conds
                  })
                }
                fields={fields}
                operators={operators}
              />
            )}
          </Stack>
        </DialogContent>
        <DialogActions sx={{ bgcolor: mutedPanelSurface, borderTop: '1px solid', borderColor: panelBorder }}>
          <Button onClick={() => setTargetDialogOpen(false)}>{t('common.cancel')}</Button>
          <Button variant="contained" onClick={handleSaveTarget}>
            {t('common.confirm')}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
