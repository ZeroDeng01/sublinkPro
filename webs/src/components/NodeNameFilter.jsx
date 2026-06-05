import { useState, useCallback, useEffect } from 'react';
import { Trans, useTranslation } from 'react-i18next';
import { DragDropContext, Droppable, Draggable } from '@hello-pangea/dnd';
import { useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Box from '@mui/material/Box';
import Paper from '@mui/material/Paper';
import Typography from '@mui/material/Typography';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import FormControl from '@mui/material/FormControl';
import Alert from '@mui/material/Alert';
import Tooltip from '@mui/material/Tooltip';
import Fade from '@mui/material/Fade';
import Switch from '@mui/material/Switch';
import Collapse from '@mui/material/Collapse';
import Divider from '@mui/material/Divider';
import Chip from '@mui/material/Chip';
import AddIcon from '@mui/icons-material/Add';
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline';
import DragIndicatorIcon from '@mui/icons-material/DragIndicator';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';
import FilterAltIcon from '@mui/icons-material/FilterAlt';
import BlockIcon from '@mui/icons-material/Block';
import CheckCircleOutlineIcon from '@mui/icons-material/CheckCircleOutline';

const PREVIEW_NODE_NAMES = ['HongKong-Node-01-Premium', 'US-Test-Node-02', 'Japan-Tokyo-03', 'Singapore-Node-04', 'Taiwan-Premium-05'];

export default function NodeNameFilter({ whitelistValue, blacklistValue, onWhitelistChange, onBlacklistChange }) {
  const { t } = useTranslation();
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));

  const [whitelistRules, setWhitelistRules] = useState([]);
  const [blacklistRules, setBlacklistRules] = useState([]);
  const [expanded, setExpanded] = useState(false);
  const [whitelistIdCounter, setWhitelistIdCounter] = useState(0);
  const [blacklistIdCounter, setBlacklistIdCounter] = useState(0);

  useEffect(() => {
    if (whitelistValue) {
      try {
        const parsed = JSON.parse(whitelistValue);
        if (Array.isArray(parsed)) {
          const rulesWithId = parsed.map((rule, idx) => ({
            ...rule,
            id: `whitelist-${idx}`
          }));
          setWhitelistRules(rulesWithId);
          setWhitelistIdCounter(parsed.length);
        }
      } catch {
        setWhitelistRules([]);
      }
    }
    if (blacklistValue) {
      try {
        const parsed = JSON.parse(blacklistValue);
        if (Array.isArray(parsed)) {
          const rulesWithId = parsed.map((rule, idx) => ({
            ...rule,
            id: `blacklist-${idx}`
          }));
          setBlacklistRules(rulesWithId);
          setBlacklistIdCounter(parsed.length);
        }
      } catch {
        setBlacklistRules([]);
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const syncWhitelistRules = useCallback(
    (newRules) => {
      const rulesForSave = newRules.map(({ id, ...rest }) => rest);
      onWhitelistChange(JSON.stringify(rulesForSave));
    },
    [onWhitelistChange]
  );

  const syncBlacklistRules = useCallback(
    (newRules) => {
      const rulesForSave = newRules.map(({ id, ...rest }) => rest);
      onBlacklistChange(JSON.stringify(rulesForSave));
    },
    [onBlacklistChange]
  );

  const handleAddWhitelistRule = () => {
    const newRule = {
      id: `whitelist-${whitelistIdCounter}`,
      matchMode: 'text',
      pattern: '',
      enabled: true
    };
    const newRules = [...whitelistRules, newRule];
    setWhitelistRules(newRules);
    setWhitelistIdCounter(whitelistIdCounter + 1);
    syncWhitelistRules(newRules);
    setExpanded(true);
  };

  const handleAddBlacklistRule = () => {
    const newRule = {
      id: `blacklist-${blacklistIdCounter}`,
      matchMode: 'text',
      pattern: '',
      enabled: true
    };
    const newRules = [...blacklistRules, newRule];
    setBlacklistRules(newRules);
    setBlacklistIdCounter(blacklistIdCounter + 1);
    syncBlacklistRules(newRules);
    setExpanded(true);
  };

  const handleUpdateRule = (listType, id, field, val) => {
    if (listType === 'whitelist') {
      const newRules = whitelistRules.map((rule) => (rule.id === id ? { ...rule, [field]: val } : rule));
      setWhitelistRules(newRules);
      syncWhitelistRules(newRules);
    } else {
      const newRules = blacklistRules.map((rule) => (rule.id === id ? { ...rule, [field]: val } : rule));
      setBlacklistRules(newRules);
      syncBlacklistRules(newRules);
    }
  };

  const handleDeleteRule = (listType, id) => {
    if (listType === 'whitelist') {
      const newRules = whitelistRules.filter((rule) => rule.id !== id);
      setWhitelistRules(newRules);
      syncWhitelistRules(newRules);
    } else {
      const newRules = blacklistRules.filter((rule) => rule.id !== id);
      setBlacklistRules(newRules);
      syncBlacklistRules(newRules);
    }
  };

  const onDragEnd = (listType) => (result) => {
    if (!result.destination) return;
    const rules = listType === 'whitelist' ? whitelistRules : blacklistRules;
    const setRules = listType === 'whitelist' ? setWhitelistRules : setBlacklistRules;
    const syncRules = listType === 'whitelist' ? syncWhitelistRules : syncBlacklistRules;

    const items = Array.from(rules);
    const [reorderedItem] = items.splice(result.source.index, 1);
    items.splice(result.destination.index, 0, reorderedItem);
    setRules(items);
    syncRules(items);
  };

  const matchesRules = (nodeName, rules) => {
    for (const rule of rules) {
      if (!rule.enabled || !rule.pattern) continue;
      try {
        if (rule.matchMode === 'regex') {
          const regex = new RegExp(rule.pattern);
          if (regex.test(nodeName)) return true;
        } else {
          if (nodeName.includes(rule.pattern)) return true;
        }
      } catch {}
    }
    return false;
  };

  const getFilteredNodes = () => {
    const hasWhitelist = whitelistRules.some((r) => r.enabled && r.pattern);
    const hasBlacklist = blacklistRules.some((r) => r.enabled && r.pattern);

    return PREVIEW_NODE_NAMES.map((name) => {
      const inBlacklist = matchesRules(name, blacklistRules);
      const inWhitelist = matchesRules(name, whitelistRules);

      if (hasBlacklist && inBlacklist) {
        return { name, status: 'excluded', reason: t('components.nodeNameFilter.blacklist') };
      }
      if (hasWhitelist && !inWhitelist) {
        return { name, status: 'excluded', reason: t('components.nodeNameFilter.notInWhitelist') };
      }
      return { name, status: 'included', reason: '' };
    });
  };

  const activeWhitelistRules = whitelistRules.filter((r) => r.enabled && r.pattern);
  const activeBlacklistRules = blacklistRules.filter((r) => r.enabled && r.pattern);
  const hasActiveWhitelist = activeWhitelistRules.length > 0;
  const hasActiveBlacklist = activeBlacklistRules.length > 0;
  const hasWhitelistRules = whitelistRules.length > 0;
  const hasBlacklistRules = blacklistRules.length > 0;
  const hasAnyRules = hasWhitelistRules || hasBlacklistRules;
  const hasAnyActiveRules = hasActiveWhitelist || hasActiveBlacklist;
  const filteredPreview = getFilteredNodes();
  const includedCount = filteredPreview.filter((n) => n.status === 'included').length;

  const renderRuleList = (listType, rules, titleKey, icon, color) => {
    const title = t(titleKey);
    return (
      <Box sx={{ mb: 2 }}>
        <Stack direction="row" alignItems="center" spacing={1} sx={{ mb: 1.5 }}>
          {icon}
          <Typography variant="subtitle2" fontWeight={600} color={`${color}.main`}>
            {title}
          </Typography>
          {rules.length > 0 && (
            <Chip label={`${rules.filter((r) => r.enabled).length}/${rules.length}`} size="small" color={color} variant="outlined" />
          )}
          <Box sx={{ flex: 1 }} />
          <Button
            size="small"
            startIcon={<AddIcon />}
            onClick={listType === 'whitelist' ? handleAddWhitelistRule : handleAddBlacklistRule}
            color={color}
          >
            {t('common.add')}
          </Button>
        </Stack>

        {rules.length > 0 ? (
          <DragDropContext onDragEnd={onDragEnd(listType)}>
            <Droppable droppableId={`${listType}Rules`}>
              {(provided) => (
                <Box ref={provided.innerRef} {...provided.droppableProps}>
                  {rules.map((rule, index) => (
                    <Draggable key={rule.id} draggableId={rule.id} index={index}>
                      {(provided, snapshot) => (
                        <Fade in>
                          <Paper
                            ref={provided.innerRef}
                            {...provided.draggableProps}
                            elevation={snapshot.isDragging ? 4 : 0}
                            sx={{
                              p: isMobile ? 1.5 : 1,
                              mb: 1,
                              border: '1px solid',
                              borderColor: rule.enabled ? `${color}.light` : 'divider',
                              borderRadius: 1.5,
                              bgcolor: snapshot.isDragging ? 'action.selected' : rule.enabled ? 'transparent' : 'action.disabledBackground',
                              opacity: rule.enabled ? 1 : 0.6,
                              transition: 'all 0.2s ease'
                            }}
                          >
                            <Stack direction={isMobile ? 'column' : 'row'} spacing={1} alignItems={isMobile ? 'stretch' : 'center'}>
                              <Box
                                {...provided.dragHandleProps}
                                sx={{
                                  display: 'flex',
                                  alignItems: 'center',
                                  cursor: 'grab',
                                  color: 'text.secondary'
                                }}
                              >
                                <DragIndicatorIcon fontSize="small" />
                              </Box>

                              <Switch
                                size="small"
                                checked={rule.enabled}
                                onChange={(e) => handleUpdateRule(listType, rule.id, 'enabled', e.target.checked)}
                                color={color}
                              />

                              <FormControl size="small" sx={{ minWidth: isMobile ? '100%' : 80 }}>
                                <Select
                                  value={rule.matchMode}
                                  onChange={(e) => handleUpdateRule(listType, rule.id, 'matchMode', e.target.value)}
                                >
                                  <MenuItem value="text">{t('components.nodeNameFilter.matchMode.text')}</MenuItem>
                                  <MenuItem value="regex">{t('components.nodeNameFilter.matchMode.regex')}</MenuItem>
                                </Select>
                              </FormControl>

                              <TextField
                                size="small"
                                placeholder={
                                  rule.matchMode === 'regex'
                                    ? t('components.nodeNameFilter.placeholders.regex')
                                    : t('components.nodeNameFilter.placeholders.keyword')
                                }
                                value={rule.pattern}
                                onChange={(e) => handleUpdateRule(listType, rule.id, 'pattern', e.target.value)}
                                sx={{ flex: 1, minWidth: isMobile ? '100%' : 150 }}
                                error={
                                  rule.matchMode === 'regex' &&
                                  rule.pattern &&
                                  (() => {
                                    try {
                                      new RegExp(rule.pattern);
                                      return false;
                                    } catch {
                                      return true;
                                    }
                                  })()
                                }
                                helperText={
                                  rule.matchMode === 'regex' &&
                                  rule.pattern &&
                                  (() => {
                                    try {
                                      new RegExp(rule.pattern);
                                      return null;
                                    } catch {
                                      return t('components.nodeNameFilter.invalidRegex');
                                    }
                                  })()
                                }
                              />

                              <Tooltip title={t('components.nodeNameFilter.deleteRule')}>
                                <IconButton size="small" color="error" onClick={() => handleDeleteRule(listType, rule.id)}>
                                  <DeleteOutlineIcon fontSize="small" />
                                </IconButton>
                              </Tooltip>
                            </Stack>
                          </Paper>
                        </Fade>
                      )}
                    </Draggable>
                  ))}
                  {provided.placeholder}
                </Box>
              )}
            </Droppable>
          </DragDropContext>
        ) : (
          <Typography variant="body2" color="text.secondary" sx={{ py: 1, textAlign: 'center' }}>
            {t('components.nodeNameFilter.noRules', { title })}
          </Typography>
        )}
      </Box>
    );
  };

  return (
    <Paper
      elevation={0}
      sx={{
        mb: 2,
        border: '1px solid',
        borderColor: 'divider',
        borderRadius: 2,
        overflow: 'hidden'
      }}
    >
      <Box
        sx={{
          p: 1.5,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          bgcolor: 'background.default',
          borderBottom: '1px solid',
          borderColor: 'divider',
          cursor: 'pointer',
          '&:hover': {
            bgcolor: 'action.hover'
          }
        }}
        onClick={() => setExpanded(!expanded)}
      >
        <Stack direction="row" alignItems="center" spacing={1}>
          <FilterAltIcon color="primary" fontSize="small" />
          <Typography variant="subtitle2" fontWeight={600}>
            {t('components.nodeNameFilter.title')}
          </Typography>
          {hasAnyRules && (
            <Typography variant="caption" color="text.secondary">
              {t('components.nodeNameFilter.summary', {
                whitelist: activeWhitelistRules.length,
                blacklist: activeBlacklistRules.length
              })}
            </Typography>
          )}
        </Stack>
        <Stack direction="row" alignItems="center" spacing={0.5}>
          {expanded ? <ExpandLessIcon /> : <ExpandMoreIcon />}
        </Stack>
      </Box>

      <Collapse in={expanded} timeout="auto">
        <Box sx={{ p: 2, pt: 1 }}>
          <Alert variant={'standard'} severity="info" sx={{ mb: 2 }}>
            <Typography variant="body2">
              <Trans i18nKey="components.nodeNameFilter.description" components={{ strong: <strong /> }} />
            </Typography>
          </Alert>

          {renderRuleList(
            'whitelist',
            whitelistRules,
            'components.nodeNameFilter.whitelist',
            <CheckCircleOutlineIcon color="success" fontSize="small" />,
            'success'
          )}

          <Divider sx={{ my: 2 }} />

          {renderRuleList(
            'blacklist',
            blacklistRules,
            'components.nodeNameFilter.blacklist',
            <BlockIcon color="error" fontSize="small" />,
            'error'
          )}

          {hasAnyActiveRules && (
            <Fade in>
              <Alert variant={'standard'} severity={includedCount < PREVIEW_NODE_NAMES.length ? 'warning' : 'success'} sx={{ mt: 2 }}>
                <Typography variant="body2" sx={{ mb: 1 }}>
                  <strong>{t('components.nodeNameFilter.previewTitle')}</strong>
                  {t('components.nodeNameFilter.previewCount', { included: includedCount, total: PREVIEW_NODE_NAMES.length })}
                </Typography>
                <Stack spacing={0.5}>
                  {filteredPreview.map((node, idx) => (
                    <Stack key={idx} direction="row" alignItems="center" spacing={1}>
                      {node.status === 'included' ? (
                        <CheckCircleOutlineIcon fontSize="small" color="success" />
                      ) : (
                        <BlockIcon fontSize="small" color="error" />
                      )}
                      <Typography
                        variant="body2"
                        sx={{
                          textDecoration: node.status === 'excluded' ? 'line-through' : 'none',
                          color: node.status === 'excluded' ? 'text.disabled' : 'text.primary'
                        }}
                      >
                        {node.name}
                        {node.reason && (
                          <Typography component="span" variant="caption" color="error" sx={{ ml: 1 }}>
                            ({node.reason})
                          </Typography>
                        )}
                      </Typography>
                    </Stack>
                  ))}
                </Stack>
              </Alert>
            </Fade>
          )}
        </Box>
      </Collapse>
    </Paper>
  );
}
