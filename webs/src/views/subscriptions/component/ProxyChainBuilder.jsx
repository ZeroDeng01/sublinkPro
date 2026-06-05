import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { useTheme } from '@mui/material/styles';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Paper from '@mui/material/Paper';
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
import Collapse from '@mui/material/Collapse';
import Divider from '@mui/material/Divider';

import DeleteIcon from '@mui/icons-material/Delete';
import AddIcon from '@mui/icons-material/Add';
import ArrowForwardIcon from '@mui/icons-material/ArrowForward';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';

import ConditionBuilder from './ConditionBuilder';
import { getChainProxyFieldControlSx, getChainProxyIconButtonSx, getChainProxyThemeTokens } from './chainProxyTheme';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';

export default function ProxyChainBuilder({ value = [], onChange, nodes = [], fields = [], operators = [], groupTypes = [] }) {
  const theme = useTheme();
  const { t } = useTranslation();
  const { isDark } = useResolvedColorScheme();
  const tokens = getChainProxyThemeTokens(theme, isDark);
  const fieldControlSx = getChainProxyFieldControlSx(tokens);
  const iconButtonSx = getChainProxyIconButtonSx(tokens);
  const errorIconButtonSx = getChainProxyIconButtonSx(tokens, theme.palette.error.main);
  const [chainItems, setChainItems] = useState(value || []);
  const [expandedIndex, setExpandedIndex] = useState(0);

  useEffect(() => {
    if (value && Array.isArray(value)) {
      setChainItems(value);
    }
  }, [value]);

  const notifyChange = (newItems) => {
    onChange?.(newItems);
  };

  const handleAddItem = () => {
    const newItems = [
      ...chainItems,
      {
        type: 'template_group',
        groupName: ''
      }
    ];
    setChainItems(newItems);
    setExpandedIndex(newItems.length - 1);
    notifyChange(newItems);
  };

  const handleRemoveItem = (index) => {
    const newItems = chainItems.filter((_, i) => i !== index);
    setChainItems(newItems);
    notifyChange(newItems);
  };

  const handleItemChange = (index, updates) => {
    const newItems = chainItems.map((item, i) => {
      if (i === index) {
        return { ...item, ...updates };
      }
      return item;
    });
    setChainItems(newItems);
    notifyChange(newItems);
  };

  const toggleExpand = (index) => {
    setExpandedIndex(expandedIndex === index ? -1 : index);
  };

  const getTypeLabel = (type) => {
    const labels = t('subscriptions.chain.proxyTypes', { returnObjects: true });
    return labels[type] || type;
  };

  const getTypeColor = (type) => {
    const colors = {
      template_group: 'primary',
      custom_group: 'secondary',
      dynamic_node: 'warning',
      specified_node: 'success'
    };
    return colors[type] || 'default';
  };

  const renderItemConfig = (item, index) => {
    const isExpanded = expandedIndex === index;

    return (
      <Paper
        key={index}
        variant="outlined"
        sx={{
          p: 2,
          position: 'relative',
          borderColor: isExpanded ? tokens.primaryStrongBorder : tokens.panelBorder,
          backgroundColor: isExpanded ? tokens.containerSurface : tokens.nestedPanelSurface,
          backgroundImage: isExpanded ? tokens.dialogSurfaceGradient : 'none',
          boxShadow: tokens.cardShadow,
          borderRadius: 2
        }}
      >
        <Stack direction="row" alignItems="center" spacing={1}>
          <Chip
            label={getTypeLabel(item.type)}
            color={getTypeColor(item.type)}
            size="small"
            sx={{ borderRadius: 1.5, fontWeight: 600, boxShadow: tokens.insetHighlight }}
          />
          <Typography variant="body2" sx={{ flex: 1, color: item.groupName || item.nodeId ? tokens.primaryText : tokens.tertiaryText }}>
            {item.groupName || item.nodeId ? (
              item.type === 'specified_node' ? (
                nodes.find((n) => n.id === item.nodeId)?.name || t('subscriptions.chain.nodeNumber', { id: item.nodeId })
              ) : (
                item.groupName
              )
            ) : (
              <Box component="em" sx={{ color: tokens.tertiaryText, fontStyle: 'italic' }}>
                {t('subscriptions.chain.unconfigured')}
              </Box>
            )}
          </Typography>
          <IconButton size="small" onClick={() => toggleExpand(index)} sx={iconButtonSx}>
            {isExpanded ? <ExpandLessIcon /> : <ExpandMoreIcon />}
          </IconButton>
          <IconButton size="small" onClick={() => handleRemoveItem(index)} sx={{ ...errorIconButtonSx, color: theme.palette.error.main }}>
            <DeleteIcon fontSize="small" />
          </IconButton>
        </Stack>

        <Collapse in={isExpanded}>
          <Box sx={{ mt: 2, pt: 2, borderTop: `1px solid ${tokens.softBorder}` }}>
            <Stack spacing={2}>
              <FormControl size="small" fullWidth sx={fieldControlSx}>
                <InputLabel>{t('subscriptions.chain.proxyType')}</InputLabel>
                <Select
                  value={item.type}
                  label={t('subscriptions.chain.proxyType')}
                  onChange={(e) =>
                    handleItemChange(index, {
                      type: e.target.value,
                      groupName: '',
                      nodeId: undefined,
                      nodeConditions: undefined
                    })
                  }
                >
                  <MenuItem value="template_group">{t('subscriptions.chain.proxyTypes.template_group')}</MenuItem>
                  <MenuItem value="custom_group">{t('subscriptions.chain.proxyTypes.custom_group')}</MenuItem>
                  <MenuItem value="dynamic_node">{t('subscriptions.chain.proxyTypes.dynamic_node')}</MenuItem>
                  <MenuItem value="specified_node">{t('subscriptions.chain.proxyTypes.specified_node')}</MenuItem>
                </Select>
              </FormControl>

              {item.type === 'template_group' && (
                <TextField
                  size="small"
                  fullWidth
                  label={t('subscriptions.chain.groupName')}
                  placeholder={t('subscriptions.chain.templateGroupPlaceholder')}
                  value={item.groupName || ''}
                  onChange={(e) => handleItemChange(index, { groupName: e.target.value })}
                  helperText={t('subscriptions.chain.templateGroupHelper')}
                  sx={fieldControlSx}
                />
              )}

              {item.type === 'custom_group' && (
                <>
                  <TextField
                    size="small"
                    fullWidth
                    label={t('subscriptions.chain.groupName')}
                    placeholder={t('subscriptions.chain.customGroupPlaceholder')}
                    value={item.groupName || ''}
                    onChange={(e) => handleItemChange(index, { groupName: e.target.value })}
                    sx={fieldControlSx}
                  />
                  <FormControl size="small" fullWidth sx={fieldControlSx}>
                    <InputLabel>{t('subscriptions.chain.groupType')}</InputLabel>
                    <Select
                      value={item.groupType || 'select'}
                      label={t('subscriptions.chain.groupType')}
                      onChange={(e) => handleItemChange(index, { groupType: e.target.value })}
                    >
                      {groupTypes.map((gt) => (
                        <MenuItem key={gt.value} value={gt.value}>
                          {gt.label}
                        </MenuItem>
                      ))}
                    </Select>
                  </FormControl>
                  {(item.groupType === 'url-test' || item.groupType === 'fallback') && (
                    <Stack spacing={1.5}>
                      <TextField
                        size="small"
                        fullWidth
                        label={t('subscriptions.chain.testUrl')}
                        value={item.urlTestConfig?.url || ''}
                        onChange={(e) =>
                          handleItemChange(index, {
                            urlTestConfig: {
                              ...item.urlTestConfig,
                              url: e.target.value
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
                          value={item.urlTestConfig?.interval ?? 300}
                          onChange={(e) =>
                            handleItemChange(index, {
                              urlTestConfig: {
                                ...item.urlTestConfig,
                                interval: parseInt(e.target.value) || 300
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
                          value={item.urlTestConfig?.tolerance ?? 50}
                          onChange={(e) =>
                            handleItemChange(index, {
                              urlTestConfig: {
                                ...item.urlTestConfig,
                                tolerance: parseInt(e.target.value) || 50
                              }
                            })
                          }
                          sx={{ ...fieldControlSx, flex: 1 }}
                          helperText={
                            item.groupType === 'url-test'
                              ? t('subscriptions.chain.toleranceHelper')
                              : t('subscriptions.chain.fallbackThreshold')
                          }
                        />
                      </Stack>
                    </Stack>
                  )}
                  {item.groupType === 'load-balance' && (
                    <Stack spacing={1.5}>
                      <TextField
                        size="small"
                        fullWidth
                        label={t('subscriptions.chain.testUrl')}
                        value={item.urlTestConfig?.url || ''}
                        onChange={(e) =>
                          handleItemChange(index, {
                            urlTestConfig: {
                              ...item.urlTestConfig,
                              url: e.target.value
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
                          value={item.urlTestConfig?.interval ?? 300}
                          onChange={(e) =>
                            handleItemChange(index, {
                              urlTestConfig: {
                                ...item.urlTestConfig,
                                interval: parseInt(e.target.value) || 300
                              }
                            })
                          }
                          sx={{ ...fieldControlSx, flex: 1 }}
                          helperText={t('subscriptions.chain.healthCheckInterval')}
                        />
                        <FormControl size="small" sx={{ ...fieldControlSx, flex: 1 }}>
                          <InputLabel>{t('subscriptions.chain.loadBalanceStrategy')}</InputLabel>
                          <Select
                            value={item.urlTestConfig?.strategy || 'consistent-hashing'}
                            label={t('subscriptions.chain.loadBalanceStrategy')}
                            onChange={(e) =>
                              handleItemChange(index, {
                                urlTestConfig: {
                                  ...item.urlTestConfig,
                                  strategy: e.target.value
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
                    value={item.nodeConditions}
                    onChange={(conds) => handleItemChange(index, { nodeConditions: conds })}
                    fields={fields}
                    operators={operators}
                  />
                </>
              )}

              {item.type === 'dynamic_node' && (
                <>
                  <FormControl size="small" fullWidth sx={fieldControlSx}>
                    <InputLabel>{t('subscriptions.chain.selectMode')}</InputLabel>
                    <Select
                      value={item.selectMode || 'first'}
                      label={t('subscriptions.chain.selectMode')}
                      onChange={(e) => handleItemChange(index, { selectMode: e.target.value })}
                    >
                      <MenuItem value="first">{t('subscriptions.chain.selectModes.first')}</MenuItem>
                      <MenuItem value="random">{t('subscriptions.chain.selectModes.random')}</MenuItem>
                      <MenuItem value="fastest">{t('subscriptions.chain.selectModes.fastest')}</MenuItem>
                    </Select>
                  </FormControl>
                  <ConditionBuilder
                    title={t('subscriptions.chain.nodeMatchConditions')}
                    value={item.nodeConditions}
                    onChange={(conds) => handleItemChange(index, { nodeConditions: conds })}
                    fields={fields}
                    operators={operators}
                  />
                </>
              )}

              {item.type === 'specified_node' && (
                <Autocomplete
                  size="small"
                  options={nodes}
                  sx={fieldControlSx}
                  getOptionLabel={(option) => `${option.name || option.linkName} (${option.linkCountry || t('common.unknown')})`}
                  value={nodes.find((n) => n.id === item.nodeId) || null}
                  onChange={(_event, newValue) => handleItemChange(index, { nodeId: newValue?.id })}
                  renderInput={(params) => <TextField {...params} label={t('subscriptions.chain.selectNode')} />}
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
            </Stack>
          </Box>
        </Collapse>
      </Paper>
    );
  };

  return (
    <Box>
      <Stack spacing={2}>
        {chainItems.length > 0 && (
          <Stack
            direction="row"
            spacing={1}
            alignItems="center"
            flexWrap="wrap"
            useFlexGap
            sx={{ p: 1.5, borderRadius: 2, backgroundColor: tokens.mutedPanelSurface, border: `1px solid ${tokens.softBorder}` }}
          >
            {chainItems.map((item, index) => (
              <Stack key={index} direction="row" alignItems="center" spacing={1}>
                <Chip
                  label={
                    item.type === 'specified_node'
                      ? nodes.find((n) => n.id === item.nodeId)?.name || t('subscriptions.chain.nodeNumber', { id: item.nodeId })
                      : item.groupName || getTypeLabel(item.type)
                  }
                  color={getTypeColor(item.type)}
                  onClick={() => toggleExpand(index)}
                  sx={{ boxShadow: tokens.insetHighlight, cursor: 'pointer' }}
                />
                {index < chainItems.length - 1 && <ArrowForwardIcon color="action" fontSize="small" />}
              </Stack>
            ))}
            <ArrowForwardIcon color="action" fontSize="small" />
            <Chip
              label={t('subscriptions.chain.targetNode')}
              variant="outlined"
              sx={{ borderColor: tokens.softBorder, color: tokens.secondaryText }}
            />
          </Stack>
        )}

        <Divider sx={{ borderColor: tokens.softBorder }} />

        <Typography variant="subtitle2" sx={{ color: tokens.secondaryText }}>
          {t('subscriptions.chain.entryProxyConfig')}
        </Typography>

        {chainItems.map((item, index) => renderItemConfig(item, index))}

        <Button
          variant="outlined"
          startIcon={<AddIcon />}
          onClick={handleAddItem}
          sx={{
            alignSelf: 'flex-start',
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
          {t('subscriptions.chain.addEntryProxy')}
        </Button>

        {chainItems.length === 0 && (
          <Typography variant="body2" sx={{ fontStyle: 'italic', color: tokens.secondaryText }}>
            {t('subscriptions.chain.emptyEntryProxy')}
          </Typography>
        )}
      </Stack>
    </Box>
  );
}
