import { useState, useCallback, useEffect, useMemo } from 'react';
import { DragDropContext, Droppable, Draggable } from '@hello-pangea/dnd';
import { darken, lighten, useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Box from '@mui/material/Box';
import Paper from '@mui/material/Paper';
import Typography from '@mui/material/Typography';
import Stack from '@mui/material/Stack';
import Chip from '@mui/material/Chip';
import Button from '@mui/material/Button';
import ButtonGroup from '@mui/material/ButtonGroup';
import IconButton from '@mui/material/IconButton';
import TextField from '@mui/material/TextField';
import Alert from '@mui/material/Alert';
import Tooltip from '@mui/material/Tooltip';
import Fade from '@mui/material/Fade';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import List from '@mui/material/List';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemText from '@mui/material/ListItemText';
import CircularProgress from '@mui/material/CircularProgress';
import Divider from '@mui/material/Divider';
import AddIcon from '@mui/icons-material/Add';
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline';
import DragIndicatorIcon from '@mui/icons-material/DragIndicator';
import ClearAllIcon from '@mui/icons-material/ClearAll';
import { getTagGroups } from 'api/tags';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getFraudScoreIcon } from 'utils/fraudScore';
import { getDelayIcon, getSpeedIcon } from 'utils/nodeMetricIcons';
import { withAlpha } from '../../../utils/colorUtils';
import { getReadableTextTokens, getSurfaceTokens } from '../../../themes/surfaceTokens';
import { getUnlockRenameVariables } from 'views/nodes/utils';

// 将国家ISO代码转换为国旗emoji
const isoToFlag = (isoCode) => {
  if (!isoCode || typeof isoCode !== 'string') {
    return '🏳️'; // 未知国旗使用白旗
  }
  const code = isoCode.toUpperCase().trim();
  if (code.length !== 2) {
    return '🏳️';
  }
  // TW使用中国国旗
  const finalCode = code === 'TW' ? 'CN' : code;
  try {
    const codePoints = [...finalCode].map((char) => 0x1f1e6 + char.charCodeAt(0) - 65);
    return String.fromCodePoint(...codePoints);
  } catch {
    return '🏳️';
  }
};

// 可用变量定义
const AVAILABLE_VARIABLES = [
  { key: '$Protocol', label: '协议', color: '#9c27b0', description: '协议类型 (VMess/VLESS等)' },
  { key: '$LinkCountry', label: '国家', color: '#2196f3', description: '落地IP国家代码' },
  { key: '$Flag', label: '国旗', color: '#f44336', description: '落地IP国旗' },
  { key: '$Name', label: '备注', color: '#4caf50', description: '系统备注名称' },
  { key: '$LinkName', label: '原名', color: '#ff9800', description: '原始节点名称' },
  { key: '$SpeedIcon', label: '速度图标', color: '#ec407a', description: '按速度结果输出图标 (🟢🟡🔴/❌⏱️⛔️)' },
  { key: '$Speed', label: '速度', color: '#e91e63', description: '下载速度' },
  { key: '$DelayIcon', label: '延迟图标', color: '#26c6da', description: '按延迟结果输出图标 (🟢🟡🔴/❌⏱️⛔️)' },
  { key: '$Delay', label: '延迟', color: '#00bcd4', description: '延迟时间' },
  { key: '$IpType', label: 'IP类型', color: '#3f51b5', description: 'IP类型 (原生IP/广播IP)' },
  { key: '$Residential', label: '住宅属性', color: '#009688', description: '住宅属性 (住宅IP/机房IP)' },
  { key: '$FraudScoreIcon', label: '欺诈图标', color: '#ff7043', description: '按欺诈评分输出风险图标 (⚪🟢🟡🟠🔴⚫/⛔️)' },
  { key: '$FraudScore', label: '欺诈评分', color: '#ff5722', description: 'IP欺诈评分' },
  { key: '$Group', label: '分组', color: '#795548', description: '分组名称' },
  { key: '$Source', label: '来源', color: '#607d8b', description: '节点来源' },
  { key: '$Index', label: '序号', color: '#9e9e9e', description: '节点序号' },
  { key: '$Tags', label: '标签', color: '#673ab7', description: '所有标签(竖线｜分隔)' },
  { key: '$TagGroup', label: '标签组', color: '#8bc34a', description: '指定标签组中的标签(点击选择标签组)' }
];

// 快捷分隔符
const QUICK_SEPARATORS = [
  { key: '-', label: '-' },
  { key: '_', label: '_' },
  { key: '|', label: '|' },
  { key: ' ', label: '空格' },
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

// 预览用的示例数据
const PREVIEW_DATA = {
  $Name: '香港节点-备注',
  $LinkName: '香港01',
  $LinkCountry: 'HK',
  $Flag: isoToFlag('HK'),
  $SpeedIcon: getSpeedIcon(1.5, 'success'),
  $Speed: '1.50MB/s',
  $DelayIcon: getDelayIcon(125, 'success'),
  $Delay: '125ms',
  $IpType: '原生IP',
  $Residential: '住宅IP',
  $FraudScoreIcon: getFraudScoreIcon(12, 'success'),
  $FraudScore: '12',
  $Unlock: 'Netflix-解锁-US-+2',
  $Group: 'Premium',
  $Source: '机场A',
  $Index: '1',
  $Protocol: 'VMess',
  $Tags: '速度优秀|香港节点'
};

/**
 * 解析规则字符串为元素数组
 */
const parseRule = (rule) => {
  if (!rule) return [];

  const items = [];
  let id = 0;

  // 匹配普通变量和 $TagGroup(xxx) 格式
  const varRegex =
    /\$(Name|LinkName|LinkCountry|Flag|SpeedIcon|Speed|DelayIcon|Delay|IpType|Residential|FraudScoreIcon|FraudScore|Unlock\([^)]+\)|Unlock|Group|Source|Index|Protocol|Tags|TagGroup\([^)]+\))/g;

  let match;
  let lastIndex = 0;

  while ((match = varRegex.exec(rule)) !== null) {
    // 添加变量前的文本（分隔符）
    if (match.index > lastIndex) {
      const sep = rule.substring(lastIndex, match.index);
      items.push({ id: `sep-${id++}`, type: 'separator', value: sep });
    }
    // 添加变量
    items.push({ id: `var-${id++}`, type: 'variable', value: match[0] });
    lastIndex = match.index + match[0].length;
  }

  // 添加剩余文本
  if (lastIndex < rule.length) {
    items.push({ id: `sep-${id++}`, type: 'separator', value: rule.substring(lastIndex) });
  }

  return items;
};

/**
 * 将元素数组转换为规则字符串
 */
const buildRule = (items) => {
  return items.map((item) => item.value).join('');
};

/**
 * 节点命名规则拖拽构建器
 */
export default function NodeRenameBuilder({ value, onChange }) {
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

  // 标签组选择弹窗状态
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
      { key: 'basic', label: '基础变量' },
      { key: 'quality', label: '质量变量' },
      { key: 'unlock', label: '解锁变量' }
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

  // 初始化：从传入的 value 解析规则
  useEffect(() => {
    const items = parseRule(value);
    setRuleItems(items);
    setIdCounter(items.length + 1);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // 同步规则到父组件
  const syncRule = useCallback(
    (items) => {
      const rule = buildRule(items);
      onChange(rule);
    },
    [onChange]
  );

  // 打开标签组选择弹窗
  const openTagGroupDialog = async () => {
    setTagGroupDialogOpen(true);
    setTagGroupsLoading(true);
    try {
      const res = await getTagGroups();
      setTagGroups(res.data || []);
    } catch (error) {
      console.error('获取标签组失败:', error);
      setTagGroups([]);
    } finally {
      setTagGroupsLoading(false);
    }
  };

  // 选择标签组
  const handleSelectTagGroup = (groupName) => {
    const varValue = `$TagGroup(${groupName})`;
    const newItem = { id: `var-${idCounter}`, type: 'variable', value: varValue };
    const newItems = [...ruleItems, newItem];
    setRuleItems(newItems);
    setIdCounter(idCounter + 1);
    syncRule(newItems);
    setTagGroupDialogOpen(false);
  };

  // 添加变量
  const handleAddVariable = (varKey) => {
    // $TagGroup 需要打开选择弹窗
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

  // 添加分隔符
  const handleAddSeparator = (sep) => {
    if (!sep) return;
    const newItem = { id: `sep-${idCounter}`, type: 'separator', value: sep };
    const newItems = [...ruleItems, newItem];
    setRuleItems(newItems);
    setIdCounter(idCounter + 1);
    syncRule(newItems);
    setCustomSeparator('');
  };

  // 删除元素
  const handleRemoveItem = (itemId) => {
    const newItems = ruleItems.filter((item) => item.id !== itemId);
    setRuleItems(newItems);
    syncRule(newItems);
  };

  // 清空所有
  const handleClearAll = () => {
    setRuleItems([]);
    syncRule([]);
  };

  // 拖拽结束
  const onDragEnd = (result) => {
    if (!result.destination) return;

    const items = Array.from(ruleItems);
    const [reorderedItem] = items.splice(result.source.index, 1);
    items.splice(result.destination.index, 0, reorderedItem);

    setRuleItems(items);
    syncRule(items);
  };

  // 获取变量的颜色
  const getVariableColor = (varKey) => {
    const variable = varKey.startsWith('$TagGroup(')
      ? availableVariables.find((v) => v.key === '$TagGroup')
      : availableVariables.find((v) => v.key === varKey);
    return variable?.color || getSectionAccent(variable?.section);
  };

  // 获取变量的标签
  const getVariableLabel = (varKey) => {
    // 处理 $TagGroup(xxx) 格式 - 显示为 "标签组:xxx"
    const tagGroupMatch = varKey.match(/\$TagGroup\(([^)]+)\)/);
    if (tagGroupMatch) {
      return `标签组:${tagGroupMatch[1]}`;
    }
    const variable = availableVariables.find((v) => v.key === varKey);
    return variable?.label || varKey;
  };

  const getVariableTone = (varKey) => getAccentTone(getVariableColor(varKey));

  // 生成预览
  const preview = ruleItems
    .map((item) => {
      if (item.type === 'variable') {
        // 处理 $TagGroup(xxx) - 预览使用示例标签名
        const tagGroupMatch = item.value.match(/\$TagGroup\(([^)]+)\)/);
        if (tagGroupMatch) {
          return '速度优秀'; // 示例标签名
        }
        if (item.value === '$Unlock') {
          return PREVIEW_DATA.$Unlock;
        }
        if (item.value.startsWith('$Unlock(')) {
          return PREVIEW_DATA[item.value] || '解锁-US';
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
      {/* 可用变量区 */}
      <Paper elevation={0} sx={sectionPaperSx}>
        <Typography variant="subtitle2" sx={sectionHeaderSx}>
          🏷️ 可用变量 (点击添加)
        </Typography>
        <Stack spacing={1.5}>
          {variableSections.map((section, sectionIndex) => {
            const sectionVariables = availableVariables.filter((item) => item.section === section.key);
            if (sectionVariables.length === 0) return null;
            return (
              <Box key={section.key}>
                {sectionIndex > 0 && <Divider sx={{ mb: 1.5, borderColor: subtleDivider }} />}
                <Typography variant="caption" sx={{ display: 'block', mb: 1, fontWeight: 700, color: secondaryText }}>
                  {section.label}
                </Typography>
                <Stack direction="row" flexWrap="wrap" gap={1}>
                  {sectionVariables.map((variable) => (
                    <Tooltip key={variable.key} title={variable.description} arrow placement="top">
                      <Chip
                        label={`${variable.label} ${variable.key}`}
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

      {/* 分隔符快捷按钮 */}
      <Paper elevation={0} sx={sectionPaperSx}>
        <Typography variant="subtitle2" sx={sectionHeaderSx}>
          ✂️ 分隔符
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
                {sep.label}
              </Button>
            ))}
          </ButtonGroup>
          <Stack direction="row" alignItems="center" spacing={1} sx={{ ml: isMobile ? 0 : 1, mt: isMobile ? 1 : 0 }}>
            <TextField
              size="small"
              placeholder="自定义"
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

      {/* 规则构建区 */}
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
            📝 命名规则 (拖拽排序)
          </Typography>
          {ruleItems.length > 0 && (
            <Tooltip title="清空所有">
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
                    点击上方变量和分隔符添加到这里
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

      {/* 实时预览 */}
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
                预览：
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
                {preview || '(空)'}
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
      {/* 标签组选择弹窗 */}
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
          选择标签组
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
              <Typography sx={{ color: secondaryText }}>暂无标签组，请先在标签管理中创建标签组</Typography>
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
