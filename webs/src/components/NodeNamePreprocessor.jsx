import { useState, useCallback, useEffect } from 'react';
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
import Chip from '@mui/material/Chip';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';
import AddIcon from '@mui/icons-material/Add';
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline';
import DragIndicatorIcon from '@mui/icons-material/DragIndicator';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';
import TextFieldsIcon from '@mui/icons-material/TextFields';

const PREVIEW_LINK_NAME = 'github-香港节点-01-Premium';

export default function NodeNamePreprocessor({ value, onChange }) {
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));
  const { palette, dialogSurface, dialogSurfaceGradient, mutedPanelSurface, nestedPanelSurface, panelBorder } = getSurfaceTokens(
    theme,
    isDark
  );
  const { primaryText, secondaryText, tertiaryText } = getReadableTextTokens(theme, isDark);
  const insetHighlight = isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.03)}` : 'none';

  const [rules, setRules] = useState([]);
  const [expanded, setExpanded] = useState(true);
  const [idCounter, setIdCounter] = useState(0);

  useEffect(() => {
    if (value) {
      try {
        const parsed = JSON.parse(value);
        if (Array.isArray(parsed)) {
          const rulesWithId = parsed.map((rule, idx) => ({
            ...rule,
            id: `rule-${idx}`
          }));
          setRules(rulesWithId);
          setIdCounter(parsed.length);
        }
      } catch {
        setRules([]);
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const syncRules = useCallback(
    (newRules) => {
      const rulesForSave = newRules.map(({ id, ...rest }) => rest);
      onChange(JSON.stringify(rulesForSave));
    },
    [onChange]
  );

  const handleAddRule = () => {
    const newRule = {
      id: `rule-${idCounter}`,
      matchMode: 'text',
      pattern: '',
      replacement: '',
      enabled: true
    };
    const newRules = [...rules, newRule];
    setRules(newRules);
    setIdCounter(idCounter + 1);
    syncRules(newRules);
  };

  const handleUpdateRule = (id, field, val) => {
    const newRules = rules.map((rule) => (rule.id === id ? { ...rule, [field]: val } : rule));
    setRules(newRules);
    syncRules(newRules);
  };

  const handleDeleteRule = (id) => {
    const newRules = rules.filter((rule) => rule.id !== id);
    setRules(newRules);
    syncRules(newRules);
  };

  const onDragEnd = (result) => {
    if (!result.destination) return;
    const items = Array.from(rules);
    const [reorderedItem] = items.splice(result.source.index, 1);
    items.splice(result.destination.index, 0, reorderedItem);
    setRules(items);
    syncRules(items);
  };

  const getPreviewResult = () => {
    let result = PREVIEW_LINK_NAME;
    for (const rule of rules) {
      if (!rule.enabled || !rule.pattern) continue;
      try {
        if (rule.matchMode === 'regex') {
          const regex = new RegExp(rule.pattern, 'g');
          result = result.replace(regex, rule.replacement);
        } else {
          result = result.replaceAll(rule.pattern, rule.replacement);
        }
      } catch {}
    }
    return result;
  };

  const hasRules = rules.length > 0;
  const previewResult = getPreviewResult();
  const hasChanges = previewResult !== PREVIEW_LINK_NAME;
  const enabledRulesCount = rules.filter((r) => r.enabled).length;
  const headerHoverSurface = isDark ? withAlpha(palette.background.paper, 0.2) : withAlpha(palette.primary.main, 0.04);
  const contentSurface = isDark
    ? `linear-gradient(180deg, ${withAlpha(palette.background.paper, 0.08)} 0%, ${dialogSurface} 100%)`
    : 'none';

  const sectionCardSx = {
    p: 1.75,
    borderRadius: 2,
    bgcolor: nestedPanelSurface,
    border: '1px solid',
    borderColor: panelBorder,
    boxShadow: insetHighlight
  };

  const codeTokenSx = {
    display: 'inline-flex',
    alignItems: 'center',
    px: 0.75,
    py: 0.35,
    ml: 0.5,
    borderRadius: 1,
    bgcolor: withAlpha(palette.background.default, isDark ? 0.72 : 0.92),
    fontFamily: 'monospace',
    wordBreak: 'break-all'
  };

  return (
    <Paper
      elevation={0}
      sx={{
        mb: 0,
        border: '1px solid',
        borderColor: panelBorder,
        borderRadius: 2,
        overflow: 'hidden',
        bgcolor: dialogSurface,
        backgroundImage: dialogSurfaceGradient,
        boxShadow: insetHighlight
      }}
    >
      <Box
        sx={{
          px: 1.75,
          py: 1.5,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          bgcolor: expanded ? nestedPanelSurface : mutedPanelSurface,
          borderBottom: expanded ? '1px solid' : 'none',
          borderColor: panelBorder,
          cursor: 'pointer',
          transition: 'background-color 0.2s ease, border-color 0.2s ease',
          '&:hover': {
            bgcolor: expanded ? nestedPanelSurface : headerHoverSurface
          }
        }}
        onClick={() => setExpanded(!expanded)}
      >
        <Stack direction="row" alignItems="center" spacing={1.25} sx={{ minWidth: 0, flex: 1 }}>
          <TextFieldsIcon color="primary" fontSize="small" />
          <Typography variant="subtitle2" fontWeight={600} sx={{ color: primaryText }}>
            原名预处理
          </Typography>
          {hasRules && (
            <Chip
              size="small"
              variant="outlined"
              label={`${enabledRulesCount}/${rules.length} 条生效`}
              sx={{
                height: 22,
                color: enabledRulesCount > 0 ? palette.primary.main : tertiaryText,
                bgcolor: withAlpha(palette.primary.main, isDark ? 0.14 : 0.06),
                borderColor: withAlpha(palette.primary.main, isDark ? 0.28 : 0.16)
              }}
            />
          )}
        </Stack>
        <Stack direction="row" alignItems="center" spacing={0.5} sx={{ ml: 1 }}>
          <Tooltip title="添加规则">
            <IconButton
              size="small"
              color="primary"
              onClick={(e) => {
                e.stopPropagation();
                handleAddRule();
                setExpanded(true);
              }}
            >
              <AddIcon fontSize="small" />
            </IconButton>
          </Tooltip>
          <Box sx={{ display: 'flex', alignItems: 'center', color: tertiaryText }}>
            {expanded ? <ExpandLessIcon fontSize="small" /> : <ExpandMoreIcon fontSize="small" />}
          </Box>
        </Stack>
      </Box>

      <Collapse in={expanded} timeout="auto">
        <Box sx={{ px: 2.25, py: 2.25, bgcolor: dialogSurface, backgroundImage: contentSurface }}>
          <Stack spacing={2.25}>
            <Typography variant="body2" sx={{ color: secondaryText }}>
              先按顺序处理节点原始名称，再把结果继续交给下游命名规则。支持文本匹配和正则表达式两种模式。
            </Typography>

            {hasRules ? (
              <Box sx={sectionCardSx}>
                <DragDropContext onDragEnd={onDragEnd}>
                  <Droppable droppableId="preprocessRules">
                    {(provided) => (
                      <Box ref={provided.innerRef} {...provided.droppableProps}>
                        <Stack spacing={1.25}>
                          {rules.map((rule, index) => (
                            <Draggable key={rule.id} draggableId={rule.id} index={index}>
                              {(provided, snapshot) => {
                                const regexError =
                                  rule.matchMode === 'regex' &&
                                  rule.pattern &&
                                  (() => {
                                    try {
                                      new RegExp(rule.pattern);
                                      return false;
                                    } catch {
                                      return true;
                                    }
                                  })();

                                return (
                                  <Fade in>
                                    <Paper
                                      ref={provided.innerRef}
                                      {...provided.draggableProps}
                                      elevation={snapshot.isDragging ? 4 : 0}
                                      sx={{
                                        p: isMobile ? 1.5 : 1.75,
                                        border: '1px solid',
                                        borderColor: rule.enabled ? withAlpha(palette.primary.main, isDark ? 0.3 : 0.18) : panelBorder,
                                        borderRadius: 2,
                                        bgcolor: snapshot.isDragging
                                          ? withAlpha(palette.primary.main, isDark ? 0.18 : 0.08)
                                          : rule.enabled
                                            ? dialogSurface
                                            : withAlpha(palette.action.disabledBackground, isDark ? 0.36 : 0.72),
                                        opacity: rule.enabled ? 1 : 0.68,
                                        transition: 'background-color 0.2s ease, border-color 0.2s ease, opacity 0.2s ease',
                                        boxShadow: snapshot.isDragging
                                          ? `0 10px 24px ${withAlpha(palette.common.black, isDark ? 0.28 : 0.12)}`
                                          : insetHighlight
                                      }}
                                    >
                                      <Stack
                                        direction={isMobile ? 'column' : 'row'}
                                        spacing={1.25}
                                        alignItems={isMobile ? 'stretch' : 'center'}
                                      >
                                        <Box
                                          {...provided.dragHandleProps}
                                          sx={{
                                            display: 'flex',
                                            alignItems: 'center',
                                            justifyContent: 'center',
                                            cursor: 'grab',
                                            color: tertiaryText,
                                            alignSelf: isMobile ? 'flex-start' : 'center',
                                            pt: isMobile ? 0.25 : 0
                                          }}
                                        >
                                          <DragIndicatorIcon fontSize="small" />
                                        </Box>

                                        <Switch
                                          size="small"
                                          checked={rule.enabled}
                                          onChange={(e) => handleUpdateRule(rule.id, 'enabled', e.target.checked)}
                                          sx={{ alignSelf: isMobile ? 'flex-start' : 'center' }}
                                        />

                                        <FormControl size="small" sx={{ minWidth: isMobile ? '100%' : 96 }}>
                                          <Select
                                            value={rule.matchMode}
                                            onChange={(e) => handleUpdateRule(rule.id, 'matchMode', e.target.value)}
                                          >
                                            <MenuItem value="text">文本</MenuItem>
                                            <MenuItem value="regex">正则</MenuItem>
                                          </Select>
                                        </FormControl>

                                        <TextField
                                          size="small"
                                          placeholder={rule.matchMode === 'regex' ? '正则表达式' : '查找文本'}
                                          value={rule.pattern}
                                          onChange={(e) => handleUpdateRule(rule.id, 'pattern', e.target.value)}
                                          sx={{ flex: 1, minWidth: isMobile ? '100%' : 140 }}
                                          error={regexError}
                                          helperText={regexError ? '无效正则' : ' '}
                                        />

                                        <Typography sx={{ display: isMobile ? 'none' : 'block', color: tertiaryText, fontWeight: 600 }}>
                                          →
                                        </Typography>

                                        <TextField
                                          size="small"
                                          placeholder="替换为 (留空删除)"
                                          value={rule.replacement}
                                          onChange={(e) => handleUpdateRule(rule.id, 'replacement', e.target.value)}
                                          sx={{ flex: 1, minWidth: isMobile ? '100%' : 120 }}
                                        />

                                        <Tooltip title="删除规则">
                                          <IconButton
                                            size="small"
                                            color="error"
                                            onClick={() => handleDeleteRule(rule.id)}
                                            sx={{ alignSelf: isMobile ? 'flex-end' : 'center' }}
                                          >
                                            <DeleteOutlineIcon fontSize="small" />
                                          </IconButton>
                                        </Tooltip>
                                      </Stack>
                                    </Paper>
                                  </Fade>
                                );
                              }}
                            </Draggable>
                          ))}
                        </Stack>
                        {provided.placeholder}
                      </Box>
                    )}
                  </Droppable>
                </DragDropContext>
              </Box>
            ) : (
              <Box
                sx={{
                  py: 3,
                  px: 2,
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: 'center',
                  gap: 1,
                  color: secondaryText,
                  borderRadius: 2,
                  bgcolor: nestedPanelSurface,
                  border: '1px dashed',
                  borderColor: withAlpha(palette.primary.main, isDark ? 0.24 : 0.18),
                  boxShadow: insetHighlight
                }}
              >
                <Typography variant="body2">暂无预处理规则</Typography>
                <Button variant="outlined" size="small" startIcon={<AddIcon />} onClick={handleAddRule}>
                  添加规则
                </Button>
              </Box>
            )}

            {hasRules && (
              <Fade in>
                <Alert
                  variant="outlined"
                  severity={hasChanges ? 'success' : 'info'}
                  sx={{
                    borderColor: withAlpha(hasChanges ? palette.success.main : palette.info.main, isDark ? 0.3 : 0.18),
                    bgcolor: withAlpha(hasChanges ? palette.success.main : palette.info.main, isDark ? 0.12 : 0.05),
                    boxShadow: insetHighlight
                  }}
                >
                  <Stack spacing={0.75}>
                    <Typography variant="body2" sx={{ color: secondaryText }}>
                      实时预览会按当前顺序连续应用所有启用规则。
                    </Typography>
                    <Typography variant="body2" sx={{ color: primaryText }}>
                      <strong>原名：</strong>
                      <Box component="code" sx={{ ...codeTokenSx, color: secondaryText }}>
                        {PREVIEW_LINK_NAME}
                      </Box>
                    </Typography>
                    <Typography variant="body2" sx={{ color: primaryText }}>
                      <strong>结果：</strong>
                      <Box
                        component="code"
                        sx={{
                          ...codeTokenSx,
                          color: hasChanges ? palette.success.main : secondaryText,
                          fontWeight: hasChanges ? 600 : 400
                        }}
                      >
                        {previewResult || '(空)'}
                      </Box>
                    </Typography>
                  </Stack>
                </Alert>
              </Fade>
            )}
          </Stack>
        </Box>
      </Collapse>
    </Paper>
  );
}
