import { useState, useCallback, useEffect, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { DragDropContext, Droppable, Draggable } from '@hello-pangea/dnd';
import { darken, lighten, useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Chip from '@mui/material/Chip';
import Typography from '@mui/material/Typography';
import IconButton from '@mui/material/IconButton';
import ButtonGroup from '@mui/material/ButtonGroup';
import Button from '@mui/material/Button';
import Tooltip from '@mui/material/Tooltip';
import Fade from '@mui/material/Fade';
import TextField from '@mui/material/TextField';
import Paper from '@mui/material/Paper';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import List from '@mui/material/List';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemText from '@mui/material/ListItemText';
import CircularProgress from '@mui/material/CircularProgress';
import Alert from '@mui/material/Alert';
import AddIcon from '@mui/icons-material/Add';
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline';
import DragIndicatorIcon from '@mui/icons-material/DragIndicator';
import ClearAllIcon from '@mui/icons-material/ClearAll';
import { getTagGroups } from 'api/tags';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { withAlpha } from '../../../utils/colorUtils';
import { getReadableTextTokens, getSurfaceTokens } from '../../../themes/surfaceTokens';
import { getUnlockRenameVariables } from 'views/nodes/utils';

const AVAILABLE_VARIABLES = [
  { key: '$Protocol', labelKey: 'subscriptions.rename.vars.protocol', color: '#9c27b0', descKey: 'subscriptions.rename.vars.protocolDesc' },
  {
    key: '$LinkCountry',
    labelKey: 'subscriptions.rename.vars.country',
    color: '#2196f3',
    descKey: 'subscriptions.rename.vars.countryDesc'
  },
  {
    key: '$LinkCountryName',
    labelKey: 'subscriptions.rename.vars.countryName',
    color: '#1976d2',
    descKey: 'subscriptions.rename.vars.countryNameDesc'
  },
  { key: '$Flag', labelKey: 'subscriptions.rename.vars.flag', color: '#f44336', descKey: 'subscriptions.rename.vars.flagDesc' },
  { key: '$Name', labelKey: 'subscriptions.rename.vars.name', color: '#4caf50', descKey: 'subscriptions.rename.vars.nameDesc' },
  { key: '$LinkName', labelKey: 'subscriptions.rename.vars.linkName', color: '#ff9800', descKey: 'subscriptions.rename.vars.linkNameDesc' },
  {
    key: '$SpeedIcon',
    labelKey: 'subscriptions.rename.vars.speedIcon',
    color: '#ec407a',
    descKey: 'subscriptions.rename.vars.speedIconDesc'
  },
  { key: '$Speed', labelKey: 'subscriptions.rename.vars.speed', color: '#e91e63', descKey: 'subscriptions.rename.vars.speedDesc' },
  {
    key: '$DelayIcon',
    labelKey: 'subscriptions.rename.vars.delayIcon',
    color: '#26c6da',
    descKey: 'subscriptions.rename.vars.delayIconDesc'
  },
  { key: '$Delay', labelKey: 'subscriptions.rename.vars.delay', color: '#00bcd4', descKey: 'subscriptions.rename.vars.delayDesc' },
  { key: '$IpType', labelKey: 'subscriptions.rename.vars.ipType', color: '#3f51b5', descKey: 'subscriptions.rename.vars.ipTypeDesc' },
  {
    key: '$Residential',
    labelKey: 'subscriptions.rename.vars.residential',
    color: '#009688',
    descKey: 'subscriptions.rename.vars.residentialDesc'
  },
  {
    key: '$FraudScoreIcon',
    labelKey: 'subscriptions.rename.vars.fraudScoreIcon',
    color: '#ff7043',
    descKey: 'subscriptions.rename.vars.fraudScoreIconDesc'
  },
  {
    key: '$FraudScore',
    labelKey: 'subscriptions.rename.vars.fraudScore',
    color: '#ff5722',
    descKey: 'subscriptions.rename.vars.fraudScoreDesc'
  },
  { key: '$Group', labelKey: 'subscriptions.rename.vars.group', color: '#795548', descKey: 'subscriptions.rename.vars.groupDesc' },
  { key: '$Source', labelKey: 'subscriptions.rename.vars.source', color: '#607d8b', descKey: 'subscriptions.rename.vars.sourceDesc' },
  { key: '$Index', labelKey: 'subscriptions.rename.vars.index', color: '#9e9e9e', descKey: 'subscriptions.rename.vars.indexDesc' },
  {
    key: '$DuplicateIndex',
    labelKey: 'subscriptions.rename.vars.duplicateIndex',
    color: '#78909c',
    descKey: 'subscriptions.rename.vars.duplicateIndexDesc'
  },
  { key: '$Tags', labelKey: 'subscriptions.rename.vars.tags', color: '#673ab7', descKey: 'subscriptions.rename.vars.tagsDesc' },
  { key: '$TagGroup', labelKey: 'subscriptions.rename.vars.tagGroup', color: '#8bc34a', descKey: 'subscriptions.rename.vars.tagGroupDesc' }
];

const QUICK_SEPARATORS = [
  { key: '-', label: '-' },
  { key: '_', label: '_' },
  { key: '|', label: '|' },
  { key: ' ', labelKey: 'common.space' },
  { key: '[', label: '[' },
  { key: ']', label: ']' },
  { key: '(', label: '(' },
  { key: ')', label: ')' }
];

const UNLOCK_VARIABLE_COLOR_POOL = [
  '#1e88e5',
  '#8e24aa',
  '#00acc1',
  '#43a047',
  '#fb8c00',
  '#e53935',
  '#3949ab',
  '#00897b',
  '#6d4c41',
  '#5e35b1'
];

const getStableAccentFromKey = (key, palette) => {
  if (!key) return palette[0];
  let hash = 0;
  for (let i = 0; i < key.length; i += 1) {
    hash = (hash * 31 + key.charCodeAt(i)) >>> 0;
  }
  return palette[hash % palette.length];
};

const PREVIEW_DATA = {
  $Protocol: 'VMess',
  $LinkCountry: 'HK',
  $LinkCountryName: '香港',
  $Flag: '🇭🇰',
  $Name: 'HK-Node-Remark',
  $LinkName: 'HK-01',
  $SpeedIcon: '🟢',
  $Speed: '12.5MB/s',
  $DelayIcon: '🟡',
  $Delay: '125ms',
  $IpType: 'Native',
  $Residential: 'Residential',
  $FraudScoreIcon: '🟢',
  $FraudScore: '15',
  $Unlock: 'Netflix-US',
  $Group: 'VIP',
  $Source: 'ProviderA',
  $Index: '01',
  $DuplicateIndex: '1',
  $Tags: 'Fast|HK'
};

/**
 */
const parseRule = (rule) => {
  if (!rule) return [];

  const items = [];
  let id = 0;

  const varRegex =
    /\$(Name|LinkName|LinkCountry|Flag|SpeedIcon|Speed|DelayIcon|Delay|IpType|Residential|FraudScoreIcon|FraudScore|Unlock\([^)]+\)|Unlock|Group|Source|DuplicateIndex|Index|Protocol|Tags|TagGroup\([^)]+\))/g;

  let match;
  let lastIndex = 0;

  while ((match = varRegex.exec(rule)) !== null) {
    if (match.index > lastIndex) {
      const sep = rule.substring(lastIndex, match.index);
      items.push({ id: `sep-${id++}`, type: 'separator', value: sep });
    }
    items.push({ id: `var-${id++}`, type: 'variable', value: match[0] });
    lastIndex = match.index + match[0].length;
  }

  if (lastIndex < rule.length) {
    items.push({ id: `sep-${id++}`, type: 'separator', value: rule.substring(lastIndex) });
  }

  return items;
};

/**
 */
const buildRule = (items) => {
  return items.map((item) => item.value).join('');
};

/**
 */
export default function NodeRenameBuilder({ value, onChange }) {
  const { t } = useTranslation();
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));
  const { isDark } = useResolvedColorScheme();
  const palette = theme.vars?.palette || theme.palette;
  const { dialogSurface, dialogSurfaceGradient, mutedPanelSurface, nestedPanelSurface, panelBorder } = getSurfaceTokens(theme, isDark);
  const { primaryText, secondaryText, tertiaryText } = getReadableTextTokens(theme, isDark);
  const insetHighlight = isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.03)}` : 'none';
  const emphasisInsetHighlight = isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.05)}` : 'none';
  const subtleDivider = withAlpha(palette.divider, isDark ? 0.72 : 0.9);
  const accentRing = withAlpha(palette.primary.main, isDark ? 0.16 : 0.08);
  const builderBorderColor = withAlpha(palette.primary.main, isDark ? 0.34 : 0.18);
  const builderHoverBorderColor = withAlpha(palette.primary.main, isDark ? 0.46 : 0.28);
  const dragOverSurface = withAlpha(palette.primary.main, isDark ? 0.12 : 0.05);
  const previewInlineSurface = isDark ? withAlpha(palette.background.default, 0.9) : nestedPanelSurface;

  const getAccentTone = useCallback(
    (accent) => ({
      subtleSurface: withAlpha(accent, isDark ? 0.16 : 0.1),
      hoverSurface: withAlpha(accent, isDark ? 0.24 : 0.16),
      strongSurface: withAlpha(accent, isDark ? 0.2 : 0.14),
      softBorder: withAlpha(accent, isDark ? 0.34 : 0.24),
      strongBorder: withAlpha(accent, isDark ? 0.46 : 0.34),
      textColor: isDark ? lighten(accent, 0.12) : darken(accent, 0.22)
    }),
    [isDark]
  );

  const getSectionAccent = (section) => {
    switch (section) {
      case 'quality':
        return theme.palette.warning.main;
      case 'unlock':
        return theme.palette.info.main;
      case 'basic':
      default:
        return theme.palette.primary.main;
    }
  };
  const buildSectionChipSx = (accent) => {
    const tone = getAccentTone(accent);

    return {
      bgcolor: tone.subtleSurface,
      color: tone.textColor,
      fontWeight: 700,
      border: '1px solid',
      borderColor: tone.softBorder,
      cursor: 'pointer',
      boxShadow: insetHighlight,
      transition: 'background-color 0.2s ease, border-color 0.2s ease, box-shadow 0.2s ease',
      '&:hover': {
        bgcolor: tone.hoverSurface,
        borderColor: tone.strongBorder,
        boxShadow: isDark ? `0 0 0 1px ${withAlpha(accent, 0.16)}` : theme.shadows[1]
      }
    };
  };

  const [ruleItems, setRuleItems] = useState([]);
  const [customSeparator, setCustomSeparator] = useState('');
  const [idCounter, setIdCounter] = useState(0);

  const [tagGroupDialogOpen, setTagGroupDialogOpen] = useState(false);
  const [tagGroups, setTagGroups] = useState([]);
  const [tagGroupsLoading, setTagGroupsLoading] = useState(false);
  const dynamicUnlockVariables = getUnlockRenameVariables().map((item) => ({
    key: item.key,
    label: item.label,
    color: item.color || getStableAccentFromKey(item.key, UNLOCK_VARIABLE_COLOR_POOL),
    description: item.description,
    section: 'unlock'
  }));
  const availableVariables = useMemo(
    () => [
      ...AVAILABLE_VARIABLES.map((item) => ({
        ...item,
        section:
          item.key === '$FraudScoreIcon' || item.key === '$FraudScore' || item.key === '$IpType' || item.key === '$Residential'
            ? 'quality'
            : 'basic'
      })),
      ...dynamicUnlockVariables
    ],
    [dynamicUnlockVariables]
  );
  const variableSections = useMemo(
    () => [
      { key: 'basic', labelKey: 'subscriptions.rename.categories.basic' },
      { key: 'quality', labelKey: 'subscriptions.rename.categories.quality' },
      { key: 'unlock', labelKey: 'subscriptions.rename.categories.unlock' }
    ],
    []
  );

  const sectionPaperSx = {
    p: 2,
    mb: 2,
    bgcolor: dialogSurface,
    backgroundImage: dialogSurfaceGradient,
    border: '1px solid',
    borderColor: panelBorder,
    borderRadius: 2.5,
    boxShadow: insetHighlight
  };

  const sectionHeaderSx = {
    mb: 1.5,
    fontWeight: 700,
    color: secondaryText
  };

  const outlinedInputSx = {
    '& .MuiOutlinedInput-root': {
      borderRadius: 2,
      bgcolor: mutedPanelSurface,
      color: primaryText,
      border: '1px solid',
      borderColor: subtleDivider,
      boxShadow: insetHighlight,
      transition: 'border-color 0.2s ease, background-color 0.2s ease, box-shadow 0.2s ease',
      '& fieldset': {
        borderColor: 'transparent'
      },
      '&:hover': {
        bgcolor: dialogSurface,
        '& fieldset': {
          borderColor: builderBorderColor
        }
      },
      '&.Mui-focused': {
        bgcolor: dialogSurface,
        boxShadow: isDark ? `0 0 0 1px ${accentRing}` : 'none',
        '& fieldset': {
          borderColor: builderHoverBorderColor
        }
      }
    },
    '& .MuiInputBase-input::placeholder': {
      color: tertiaryText,
      opacity: 1
    }
  };

  useEffect(() => {
    const items = parseRule(value);
    setRuleItems(items);
    setIdCounter(items.length + 1);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const syncRule = useCallback(
    (items) => {
      const rule = buildRule(items);
      onChange(rule);
    },
    [onChange]
  );

  const openTagGroupDialog = async () => {
    setTagGroupDialogOpen(true);
    setTagGroupsLoading(true);
    try {
      const res = await getTagGroups();
      setTagGroups(res.data || []);
    } catch (error) {
      console.error('Failed to get tag groups:', error);
      setTagGroups([]);
    } finally {
      setTagGroupsLoading(false);
    }
  };

  const handleSelectTagGroup = (groupName) => {
    const varValue = `$TagGroup(${groupName})`;
    const newItem = { id: `var-${idCounter}`, type: 'variable', value: varValue };
    const newItems = [...ruleItems, newItem];
    setRuleItems(newItems);
    setIdCounter(idCounter + 1);
    syncRule(newItems);
    setTagGroupDialogOpen(false);
  };

  const handleAddVariable = (varKey) => {
    if (varKey === '$TagGroup') {
      openTagGroupDialog();
      return;
    }
    const newItem = { id: `var-${idCounter}`, type: 'variable', value: varKey };
    const newItems = [...ruleItems, newItem];
    setRuleItems(newItems);
    setIdCounter(idCounter + 1);
    syncRule(newItems);
  };

  const handleAddSeparator = (sep) => {
    if (!sep) return;
    const newItem = { id: `sep-${idCounter}`, type: 'separator', value: sep };
    const newItems = [...ruleItems, newItem];
    setRuleItems(newItems);
    setIdCounter(idCounter + 1);
    syncRule(newItems);
    setCustomSeparator('');
  };

  const handleRemoveItem = (itemId) => {
    const newItems = ruleItems.filter((item) => item.id !== itemId);
    setRuleItems(newItems);
    syncRule(newItems);
  };

  const handleClearAll = () => {
    setRuleItems([]);
    syncRule([]);
  };

  const onDragEnd = (result) => {
    if (!result.destination) return;

    const items = Array.from(ruleItems);
    const [reorderedItem] = items.splice(result.source.index, 1);
    items.splice(result.destination.index, 0, reorderedItem);

    setRuleItems(items);
    syncRule(items);
  };

  const getVariableColor = (varKey) => {
    const variable = varKey.startsWith('$TagGroup(')
      ? availableVariables.find((v) => v.key === '$TagGroup')
      : availableVariables.find((v) => v.key === varKey);
    return variable?.color || getSectionAccent(variable?.section);
  };

  const getVariableLabel = (varKey) => {
    const tagGroupMatch = varKey.match(/\$TagGroup\(([^)]+)\)/);
    if (tagGroupMatch) {
      return `${t('subscriptions.rename.vars.tagGroup')}:${tagGroupMatch[1]}`;
    }
    const variable = availableVariables.find((v) => v.key === varKey);
    return variable?.labelKey ? t(variable.labelKey) : variable?.label || varKey;
  };

  const getVariableTone = (varKey) => getAccentTone(getVariableColor(varKey));

  const preview = ruleItems
    .map((item) => {
      if (item.type === 'variable') {
        const tagGroupMatch = item.value.match(/\$TagGroup\(([^)]+)\)/);
        if (tagGroupMatch) {
          return 'Fast';
        }
        if (item.value === '$Unlock') {
          return PREVIEW_DATA.$Unlock;
        }
        if (item.value.startsWith('$Unlock(')) {
          return PREVIEW_DATA[item.value] || 'Unlock-US';
        }
        return PREVIEW_DATA[item.value] || item.value;
      }
      return item.value;
    })
    .join('');

  const qualityVariableKeys = useMemo(
    () => new Set(availableVariables.filter((item) => item.section === 'quality').map((item) => item.key)),
    [availableVariables]
  );

  return (
    <Box>
      <Paper elevation={0} sx={sectionPaperSx}>
        <Typography variant="subtitle2" sx={sectionHeaderSx}>
          🏷️ {t('subscriptions.rename.availableVars')}
        </Typography>
        <Stack spacing={2.5}>
          {variableSections.map((section) => {
            const sectionVariables = availableVariables.filter((v) => v.section === section.key);
            if (sectionVariables.length === 0) return null;
            return (
              <Box key={section.key}>
                <Typography variant="caption" sx={{ color: secondaryText, fontWeight: 700, mb: 1, display: 'block' }}>
                  {t(section.labelKey)}
                </Typography>
                <Stack direction="row" flexWrap="wrap" gap={1}>
                  {sectionVariables.map((variable) => (
                    <Tooltip key={variable.key} title={variable.descKey ? t(variable.descKey) : variable.description} arrow placement="top">
                      <Chip
                        label={`${variable.labelKey ? t(variable.labelKey) : variable.label} ${variable.key}`}
                        onClick={() => handleAddVariable(variable.key)}
                        sx={buildSectionChipSx(getVariableColor(variable.key))}
                      />
                    </Tooltip>
                  ))}
                </Stack>
              </Box>
            );
          })}
        </Stack>
      </Paper>

      <Paper elevation={0} sx={sectionPaperSx}>
        <Typography variant="subtitle2" sx={sectionHeaderSx}>
          ✂️ {t('subscriptions.rename.separators')}
        </Typography>
        <Stack direction="row" alignItems="center" flexWrap="wrap" gap={1}>
          <ButtonGroup
            size="small"
            variant="outlined"
            sx={{
              '& .MuiButton-root': {
                minWidth: isMobile ? 36 : 44,
                fontWeight: 700,
                fontFamily: 'monospace',
                color: secondaryText,
                bgcolor: nestedPanelSurface,
                borderColor: subtleDivider,
                boxShadow: insetHighlight,
                '&:hover': {
                  bgcolor: mutedPanelSurface,
                  borderColor: builderHoverBorderColor,
                  color: primaryText
                }
              }
            }}
          >
            {QUICK_SEPARATORS.map((sep) => (
              <Button key={sep.key} onClick={() => handleAddSeparator(sep.key)}>
                {sep.labelKey ? t(sep.labelKey) : sep.label}
              </Button>
            ))}
          </ButtonGroup>
          <Stack direction="row" alignItems="center" spacing={1} sx={{ ml: isMobile ? 0 : 1, mt: isMobile ? 1 : 0 }}>
            <TextField
              size="small"
              placeholder={t('common.custom')}
              value={customSeparator}
              onChange={(e) => setCustomSeparator(e.target.value)}
              sx={{ width: 90, ...outlinedInputSx }}
            />
            <IconButton
              size="small"
              color="primary"
              onClick={() => handleAddSeparator(customSeparator)}
              disabled={!customSeparator}
              sx={{
                bgcolor: withAlpha(palette.primary.main, isDark ? 0.16 : 0.08),
                border: '1px solid',
                borderColor: withAlpha(palette.primary.main, isDark ? 0.34 : 0.18),
                boxShadow: insetHighlight,
                '&:hover': {
                  bgcolor: withAlpha(palette.primary.main, isDark ? 0.22 : 0.12),
                  borderColor: withAlpha(palette.primary.main, isDark ? 0.46 : 0.28)
                },
                '&.Mui-disabled': {
                  bgcolor: 'action.disabledBackground',
                  borderColor: 'action.disabledBackground'
                }
              }}
            >
              <AddIcon />
            </IconButton>
          </Stack>
        </Stack>
      </Paper>

      <Paper
        elevation={0}
        sx={{
          p: 2,
          mb: 2,
          minHeight: 80,
          bgcolor: dialogSurface,
          backgroundImage: dialogSurfaceGradient,
          boxShadow: insetHighlight,
          border: '2px dashed',
          borderColor: ruleItems.length > 0 ? builderBorderColor : panelBorder,
          borderRadius: 2.5,
          transition: 'border-color 0.3s ease, background-color 0.3s ease, box-shadow 0.3s ease'
        }}
      >
        <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 1.5 }}>
          <Typography variant="subtitle2" sx={{ fontWeight: 700, color: secondaryText }}>
            📝 {t('subscriptions.rename.namingRule')}
          </Typography>
          {ruleItems.length > 0 && (
            <Tooltip title={t('common.clearAll')}>
              <IconButton
                size="small"
                color="error"
                onClick={handleClearAll}
                sx={{
                  bgcolor: nestedPanelSurface,
                  border: '1px solid',
                  borderColor: subtleDivider,
                  boxShadow: insetHighlight,
                  '&:hover': {
                    bgcolor: withAlpha(palette.error.main, isDark ? 0.14 : 0.08),
                    borderColor: withAlpha(palette.error.main, isDark ? 0.42 : 0.24)
                  }
                }}
              >
                <ClearAllIcon fontSize="small" />
              </IconButton>
            </Tooltip>
          )}
        </Stack>

        <DragDropContext onDragEnd={onDragEnd}>
          <Droppable droppableId="ruleBuilder" direction="horizontal">
            {(provided, snapshot) => (
              <Box
                ref={provided.innerRef}
                {...provided.droppableProps}
                sx={{
                  display: 'flex',
                  flexWrap: 'wrap',
                  gap: 1,
                  minHeight: 44,
                  p: 1,
                  borderRadius: 2,
                  bgcolor: snapshot.isDraggingOver ? dragOverSurface : mutedPanelSurface,
                  border: '1px solid',
                  borderColor: snapshot.isDraggingOver ? builderHoverBorderColor : subtleDivider,
                  boxShadow: emphasisInsetHighlight,
                  transition: 'background-color 0.2s ease, border-color 0.2s ease, box-shadow 0.2s ease'
                }}
              >
                {ruleItems.length === 0 ? (
                  <Typography
                    variant="body2"
                    sx={{
                      color: tertiaryText,
                      fontStyle: 'italic',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      width: '100%'
                    }}
                  >
                    {t('subscriptions.rename.clickToAdd')}
                  </Typography>
                ) : (
                  ruleItems.map((item, index) => (
                    <Draggable key={item.id} draggableId={item.id} index={index}>
                      {(provided, snapshot) => (
                        <Fade in>
                          <Box
                            ref={provided.innerRef}
                            {...provided.draggableProps}
                            {...provided.dragHandleProps}
                            sx={{ display: 'inline-flex' }}
                          >
                            <Chip
                              icon={<DragIndicatorIcon sx={{ fontSize: 16 }} />}
                              label={item.type === 'variable' ? getVariableLabel(item.value) : `"${item.value}"`}
                              onDelete={() => handleRemoveItem(item.id)}
                              deleteIcon={<DeleteOutlineIcon sx={{ fontSize: 16 }} />}
                              sx={(() => {
                                const accent = item.type === 'variable' ? getVariableColor(item.value) : null;
                                const tone = item.type === 'variable' ? getVariableTone(item.value) : null;

                                return {
                                  bgcolor: tone ? tone.strongSurface : nestedPanelSurface,
                                  color: tone ? tone.textColor : primaryText,
                                  fontWeight: 600,
                                  border: '1px solid',
                                  borderColor: tone ? tone.softBorder : subtleDivider,
                                  boxShadow: snapshot.isDragging
                                    ? isDark
                                      ? `0 0 0 1px ${tone ? withAlpha(accent, 0.2) : accentRing}, 0 10px 24px ${withAlpha(palette.common.black, 0.28)}`
                                      : theme.shadows[4]
                                    : insetHighlight,
                                  transform: snapshot.isDragging ? 'scale(1.03)' : 'scale(1)',
                                  transition: 'transform 0.1s ease, box-shadow 0.1s ease, border-color 0.2s ease',
                                  '& .MuiChip-icon': {
                                    color: 'inherit',
                                    opacity: 0.72,
                                    cursor: 'grab'
                                  },
                                  '& .MuiChip-deleteIcon': {
                                    color: 'inherit',
                                    opacity: 0.68,
                                    '&:hover': {
                                      opacity: 1,
                                      color: 'error.main'
                                    }
                                  }
                                };
                              })()}
                            />
                          </Box>
                        </Fade>
                      )}
                    </Draggable>
                  ))
                )}
                {provided.placeholder}
              </Box>
            )}
          </Droppable>
        </DragDropContext>
      </Paper>

      {ruleItems.length > 0 && (
        <Fade in>
          <Alert
            variant={'standard'}
            severity="info"
            sx={{
              bgcolor: dialogSurface,
              color: primaryText,
              border: '1px solid',
              borderColor: withAlpha(palette.info.main, isDark ? 0.34 : 0.18),
              boxShadow: insetHighlight,
              '& .MuiAlert-icon': {
                color: 'info.main'
              },
              '& .MuiAlert-message': {
                width: '100%'
              }
            }}
          >
            <Stack direction="row" alignItems="center" spacing={1}>
              <Typography variant="body2" fontWeight={600}>
                {t('common.preview')}:
              </Typography>
              <Typography
                variant="body2"
                sx={{
                  fontFamily: 'monospace',
                  bgcolor: previewInlineSurface,
                  border: '1px solid',
                  borderColor: subtleDivider,
                  color: primaryText,
                  px: 1,
                  py: 0.5,
                  borderRadius: 1,
                  wordBreak: 'break-all'
                }}
              >
                {preview || `(${t('common.empty')})`}
              </Typography>
            </Stack>
            {qualityVariableKeys.size > 0 && (
              <Stack direction="row" flexWrap="wrap" useFlexGap spacing={0.75} sx={{ mt: 1.25 }}>
                {availableVariables
                  .filter((variable) => variable.section === 'quality')
                  .map((variable) => {
                    const tone = getVariableTone(variable.key);
                    return (
                      <Chip
                        key={variable.key}
                        label={variable.label}
                        size="small"
                        sx={{
                          height: 22,
                          fontSize: 11,
                          fontWeight: 700,
                          bgcolor: tone.subtleSurface,
                          color: tone.textColor,
                          border: '1px solid',
                          borderColor: tone.softBorder,
                          boxShadow: insetHighlight,
                          '& .MuiChip-label': { px: 0.9 }
                        }}
                      />
                    );
                  })}
              </Stack>
            )}
          </Alert>
        </Fade>
      )}
      <Dialog
        open={tagGroupDialogOpen}
        onClose={() => setTagGroupDialogOpen(false)}
        maxWidth="xs"
        fullWidth
        slotProps={{
          paper: {
            sx: {
              borderRadius: 2.5,
              bgcolor: dialogSurface,
              backgroundImage: dialogSurfaceGradient,
              border: '1px solid',
              borderColor: panelBorder,
              boxShadow: emphasisInsetHighlight
            }
          }
        }}
      >
        <DialogTitle
          sx={{
            pb: 1,
            bgcolor: mutedPanelSurface,
            color: primaryText,
            borderBottom: '1px solid',
            borderColor: panelBorder
          }}
        >
          {t('subscriptions.rename.selectTagGroup')}
        </DialogTitle>
        <DialogContent sx={{ pt: 1.5, bgcolor: 'transparent' }}>
          {tagGroupsLoading ? (
            <Box
              sx={{
                display: 'flex',
                justifyContent: 'center',
                py: 3,
                bgcolor: mutedPanelSurface,
                border: '1px solid',
                borderColor: subtleDivider,
                borderRadius: 2,
                boxShadow: insetHighlight
              }}
            >
              <CircularProgress size={32} />
            </Box>
          ) : tagGroups.length === 0 ? (
            <Box
              sx={{
                py: 2,
                px: 1.5,
                textAlign: 'center',
                bgcolor: mutedPanelSurface,
                border: '1px solid',
                borderColor: subtleDivider,
                borderRadius: 2,
                boxShadow: insetHighlight
              }}
            >
              <Typography sx={{ color: secondaryText }}>{t('subscriptions.rename.noTagGroup')}</Typography>
            </Box>
          ) : (
            <List
              sx={{
                py: 0,
                px: 0.5,
                bgcolor: mutedPanelSurface,
                border: '1px solid',
                borderColor: subtleDivider,
                borderRadius: 2,
                boxShadow: insetHighlight
              }}
            >
              {tagGroups.map((group) => (
                <ListItemButton
                  key={group}
                  onClick={() => handleSelectTagGroup(group)}
                  sx={{
                    borderRadius: 1.5,
                    mb: 0.5,
                    color: primaryText,
                    bgcolor: nestedPanelSurface,
                    border: '1px solid',
                    borderColor: 'transparent',
                    '&:hover': {
                      bgcolor: withAlpha(palette.primary.main, isDark ? 0.12 : 0.06),
                      borderColor: withAlpha(palette.primary.main, isDark ? 0.34 : 0.18)
                    }
                  }}
                >
                  <ListItemText primary={group} />
                </ListItemButton>
              ))}
            </List>
          )}
        </DialogContent>
      </Dialog>
    </Box>
  );
}
