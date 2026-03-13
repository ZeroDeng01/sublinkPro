import { useState, useCallback, useEffect } from 'react';
import { DragDropContext, Droppable, Draggable } from '@hello-pangea/dnd';
import { useTheme } from '@mui/material/styles';
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
import AddIcon from '@mui/icons-material/Add';
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline';
import DragIndicatorIcon from '@mui/icons-material/DragIndicator';
import ClearAllIcon from '@mui/icons-material/ClearAll';
import { getTagGroups } from 'api/tags';

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
  { key: '$Speed', label: '速度', color: '#e91e63', description: '下载速度' },
  { key: '$Delay', label: '延迟', color: '#00bcd4', description: '延迟时间' },
  { key: '$IpType', label: 'IP类型', color: '#3f51b5', description: 'IP类型 (原生IP/广播IP)' },
  { key: '$Residential', label: '住宅属性', color: '#009688', description: '住宅属性 (住宅IP/机房IP)' },
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

// 预览用的示例数据
const PREVIEW_DATA = {
  $Name: '香港节点-备注',
  $LinkName: '香港01',
  $LinkCountry: 'HK',
  $Flag: isoToFlag('HK'),
  $Speed: '1.50MB/s',
  $Delay: '125ms',
  $IpType: '原生IP',
  $Residential: '住宅IP',
  $FraudScore: '12',
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
    /\$(Name|LinkName|LinkCountry|Flag|Speed|Delay|IpType|Residential|FraudScore|Group|Source|Index|Protocol|Tags|TagGroup\([^)]+\))/g;

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

  const [ruleItems, setRuleItems] = useState([]);
  const [customSeparator, setCustomSeparator] = useState('');
  const [idCounter, setIdCounter] = useState(0);

  // 标签组选择弹窗状态
  const [tagGroupDialogOpen, setTagGroupDialogOpen] = useState(false);
  const [tagGroups, setTagGroups] = useState([]);
  const [tagGroupsLoading, setTagGroupsLoading] = useState(false);

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
    // 处理 $TagGroup(xxx) 格式
    if (varKey.startsWith('$TagGroup(')) {
      return AVAILABLE_VARIABLES.find((v) => v.key === '$TagGroup')?.color || '#8bc34a';
    }
    const variable = AVAILABLE_VARIABLES.find((v) => v.key === varKey);
    return variable?.color || '#9e9e9e';
  };

  // 获取变量的标签
  const getVariableLabel = (varKey) => {
    // 处理 $TagGroup(xxx) 格式 - 显示为 "标签组:xxx"
    const tagGroupMatch = varKey.match(/\$TagGroup\(([^)]+)\)/);
    if (tagGroupMatch) {
      return `标签组:${tagGroupMatch[1]}`;
    }
    const variable = AVAILABLE_VARIABLES.find((v) => v.key === varKey);
    return variable?.label || varKey;
  };

  // 生成预览
  const preview = ruleItems
    .map((item) => {
      if (item.type === 'variable') {
        // 处理 $TagGroup(xxx) - 预览使用示例标签名
        const tagGroupMatch = item.value.match(/\$TagGroup\(([^)]+)\)/);
        if (tagGroupMatch) {
          return '速度优秀'; // 示例标签名
        }
        return PREVIEW_DATA[item.value] || item.value;
      }
      return item.value;
    })
    .join('');

  return (
    <Box>
      {/* 可用变量区 */}
      <Paper
        elevation={0}
        sx={{
          p: 2,
          mb: 2,
          background: `linear-gradient(145deg, ${theme.palette.mode === 'dark' ? '#1a2027' : '#f8f9fa'} 0%, ${theme.palette.mode === 'dark' ? '#121417' : '#ffffff'} 100%)`,
          border: '1px solid',
          borderColor: 'divider',
          borderRadius: 2
        }}
      >
        <Typography variant="subtitle2" color="textSecondary" sx={{ mb: 1.5, fontWeight: 600 }}>
          🏷️ 可用变量 (点击添加)
        </Typography>
        <Stack direction="row" flexWrap="wrap" gap={1}>
          {AVAILABLE_VARIABLES.map((variable) => (
            <Tooltip key={variable.key} title={variable.description} arrow placement="top">
              <Chip
                label={`${variable.label} ${variable.key}`}
                onClick={() => handleAddVariable(variable.key)}
                sx={{
                  bgcolor: `${variable.color}20`,
                  color: variable.color,
                  fontWeight: 600,
                  border: `1px solid ${variable.color}40`,
                  cursor: 'pointer',
                  transition: 'all 0.2s ease',
                  '&:hover': {
                    bgcolor: `${variable.color}30`,
                    transform: 'translateY(-2px)',
                    boxShadow: `0 4px 12px ${variable.color}40`
                  },
                  '&:active': {
                    transform: 'translateY(0)'
                  }
                }}
              />
            </Tooltip>
          ))}
        </Stack>
      </Paper>

      {/* 分隔符快捷按钮 */}
      <Paper
        elevation={0}
        sx={{
          p: 2,
          mb: 2,
          background: theme.palette.mode === 'dark' ? '#1a2027' : '#f8f9fa',
          border: '1px solid',
          borderColor: 'divider',
          borderRadius: 2
        }}
      >
        <Typography variant="subtitle2" color="textSecondary" sx={{ mb: 1.5, fontWeight: 600 }}>
          ✂️ 分隔符
        </Typography>
        <Stack direction="row" alignItems="center" flexWrap="wrap" gap={1}>
          <ButtonGroup size="small" variant="outlined">
            {QUICK_SEPARATORS.map((sep) => (
              <Button
                key={sep.key}
                onClick={() => handleAddSeparator(sep.key)}
                sx={{
                  minWidth: isMobile ? 36 : 44,
                  fontWeight: 700,
                  fontFamily: 'monospace'
                }}
              >
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
              sx={{ width: 90, '& .MuiOutlinedInput-root': { borderRadius: 1 } }}
            />
            <IconButton size="small" color="primary" onClick={() => handleAddSeparator(customSeparator)} disabled={!customSeparator}>
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
          background: `linear-gradient(145deg, ${theme.palette.mode === 'dark' ? '#1e2a35' : '#e3f2fd'} 0%, ${theme.palette.mode === 'dark' ? '#1a2027' : '#ffffff'} 100%)`,
          border: '2px dashed',
          borderColor: ruleItems.length > 0 ? 'primary.main' : 'divider',
          borderRadius: 2,
          transition: 'all 0.3s ease'
        }}
      >
        <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 1.5 }}>
          <Typography variant="subtitle2" color="textSecondary" sx={{ fontWeight: 600 }}>
            📝 命名规则 (拖拽排序)
          </Typography>
          {ruleItems.length > 0 && (
            <Tooltip title="清空所有">
              <IconButton size="small" color="error" onClick={handleClearAll}>
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
                  borderRadius: 1,
                  bgcolor: snapshot.isDraggingOver ? 'action.hover' : 'transparent',
                  transition: 'background-color 0.2s ease'
                }}
              >
                {ruleItems.length === 0 ? (
                  <Typography
                    variant="body2"
                    color="textSecondary"
                    sx={{
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
                          <Chip
                            ref={provided.innerRef}
                            {...provided.draggableProps}
                            {...provided.dragHandleProps}
                            icon={<DragIndicatorIcon sx={{ fontSize: 16 }} />}
                            label={item.type === 'variable' ? getVariableLabel(item.value) : `"${item.value}"`}
                            onDelete={() => handleRemoveItem(item.id)}
                            deleteIcon={<DeleteOutlineIcon sx={{ fontSize: 16 }} />}
                            sx={{
                              bgcolor:
                                item.type === 'variable'
                                  ? `${getVariableColor(item.value)}20`
                                  : theme.palette.mode === 'dark'
                                    ? '#333'
                                    : '#e0e0e0',
                              color: item.type === 'variable' ? getVariableColor(item.value) : 'text.primary',
                              fontWeight: 600,
                              border: '1px solid',
                              borderColor: item.type === 'variable' ? `${getVariableColor(item.value)}40` : 'divider',
                              transform: snapshot.isDragging ? 'scale(1.05)' : 'scale(1)',
                              boxShadow: snapshot.isDragging ? 4 : 0,
                              transition: 'transform 0.1s ease, box-shadow 0.1s ease',
                              '& .MuiChip-icon': {
                                color: 'inherit',
                                opacity: 0.6,
                                cursor: 'grab'
                              },
                              '& .MuiChip-deleteIcon': {
                                color: 'inherit',
                                opacity: 0.6,
                                '&:hover': {
                                  opacity: 1,
                                  color: 'error.main'
                                }
                              }
                            }}
                          />
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
                  bgcolor: 'action.hover',
                  px: 1,
                  py: 0.5,
                  borderRadius: 1,
                  wordBreak: 'break-all'
                }}
              >
                {preview || '(空)'}
              </Typography>
            </Stack>
          </Alert>
        </Fade>
      )}
      {/* 标签组选择弹窗 */}
      <Dialog
        open={tagGroupDialogOpen}
        onClose={() => setTagGroupDialogOpen(false)}
        maxWidth="xs"
        fullWidth
        PaperProps={{
          sx: { borderRadius: 2 }
        }}
      >
        <DialogTitle sx={{ pb: 1 }}>选择标签组</DialogTitle>
        <DialogContent sx={{ pt: 1 }}>
          {tagGroupsLoading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', py: 3 }}>
              <CircularProgress size={32} />
            </Box>
          ) : tagGroups.length === 0 ? (
            <Typography color="textSecondary" sx={{ py: 2, textAlign: 'center' }}>
              暂无标签组，请先在标签管理中创建标签组
            </Typography>
          ) : (
            <List sx={{ py: 0 }}>
              {tagGroups.map((group) => (
                <ListItemButton
                  key={group}
                  onClick={() => handleSelectTagGroup(group)}
                  sx={{
                    borderRadius: 1,
                    mb: 0.5,
                    '&:hover': {
                      bgcolor: 'primary.main',
                      color: 'primary.contrastText'
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
