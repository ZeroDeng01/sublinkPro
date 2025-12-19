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
import AddIcon from '@mui/icons-material/Add';
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline';
import DragIndicatorIcon from '@mui/icons-material/DragIndicator';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';

// 示例原名用于预览
const PREVIEW_LINK_NAME = 'github-香港节点-01-Premium';

/**
 * 原名预处理规则编辑器
 * 允许用户通过纯文本或正则表达式匹配并替换/删除原名中的内容
 */
export default function NodeNamePreprocessor({ value, onChange }) {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));

  const [rules, setRules] = useState([]);
  const [expanded, setExpanded] = useState(true);
  const [idCounter, setIdCounter] = useState(0);

  // 解析传入的JSON规则
  useEffect(() => {
    if (value) {
      try {
        const parsed = JSON.parse(value);
        if (Array.isArray(parsed)) {
          // 添加唯一ID
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

  // 同步规则到父组件
  const syncRules = useCallback(
    (newRules) => {
      // 移除id字段后序列化
      const rulesForSave = newRules.map(({ id, ...rest }) => rest);
      onChange(JSON.stringify(rulesForSave));
    },
    [onChange]
  );

  // 添加新规则
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

  // 更新规则
  const handleUpdateRule = (id, field, val) => {
    const newRules = rules.map((rule) => (rule.id === id ? { ...rule, [field]: val } : rule));
    setRules(newRules);
    syncRules(newRules);
  };

  // 删除规则
  const handleDeleteRule = (id) => {
    const newRules = rules.filter((rule) => rule.id !== id);
    setRules(newRules);
    syncRules(newRules);
  };

  // 拖拽结束
  const onDragEnd = (result) => {
    if (!result.destination) return;
    const items = Array.from(rules);
    const [reorderedItem] = items.splice(result.source.index, 1);
    items.splice(result.destination.index, 0, reorderedItem);
    setRules(items);
    syncRules(items);
  };

  // 计算预览结果
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
      } catch {
        // 忽略无效正则
      }
    }
    return result;
  };

  const hasRules = rules.length > 0;
  const previewResult = getPreviewResult();
  const hasChanges = previewResult !== PREVIEW_LINK_NAME;

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
      {/* 标题栏 */}
      <Box
        sx={{
          p: 1.5,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          background: `linear-gradient(145deg, ${theme.palette.mode === 'dark' ? '#1a2027' : '#f5f5f5'} 0%, ${theme.palette.mode === 'dark' ? '#121417' : '#fafafa'} 100%)`,
          cursor: 'pointer',
          '&:hover': {
            bgcolor: 'action.hover'
          }
        }}
        onClick={() => setExpanded(!expanded)}
      >
        <Stack direction="row" alignItems="center" spacing={1}>
          <Typography variant="subtitle2" fontWeight={600}>
            ✏️ 原名预处理
          </Typography>
          {hasRules && (
            <Typography variant="caption" color="textSecondary">
              ({rules.filter((r) => r.enabled).length} 条规则)
            </Typography>
          )}
        </Stack>
        <Stack direction="row" alignItems="center" spacing={0.5}>
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
          {expanded ? <ExpandLessIcon /> : <ExpandMoreIcon />}
        </Stack>
      </Box>

      <Collapse in={expanded} timeout="auto">
        <Box sx={{ p: 2, pt: 1 }}>
          {/* 规则列表 */}
          {hasRules ? (
            <DragDropContext onDragEnd={onDragEnd}>
              <Droppable droppableId="preprocessRules">
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
                                borderColor: rule.enabled ? 'primary.light' : 'divider',
                                borderRadius: 1.5,
                                bgcolor: snapshot.isDragging
                                  ? 'action.selected'
                                  : rule.enabled
                                    ? 'transparent'
                                    : 'action.disabledBackground',
                                opacity: rule.enabled ? 1 : 0.6,
                                transition: 'all 0.2s ease'
                              }}
                            >
                              <Stack direction={isMobile ? 'column' : 'row'} spacing={1} alignItems={isMobile ? 'stretch' : 'center'}>
                                {/* 拖拽手柄 */}
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

                                {/* 启用开关 */}
                                <Switch
                                  size="small"
                                  checked={rule.enabled}
                                  onChange={(e) => handleUpdateRule(rule.id, 'enabled', e.target.checked)}
                                />

                                {/* 匹配模式 */}
                                <FormControl size="small" sx={{ minWidth: isMobile ? '100%' : 80 }}>
                                  <Select value={rule.matchMode} onChange={(e) => handleUpdateRule(rule.id, 'matchMode', e.target.value)}>
                                    <MenuItem value="text">文本</MenuItem>
                                    <MenuItem value="regex">正则</MenuItem>
                                  </Select>
                                </FormControl>

                                {/* 查找内容 */}
                                <TextField
                                  size="small"
                                  placeholder={rule.matchMode === 'regex' ? '正则表达式' : '查找文本'}
                                  value={rule.pattern}
                                  onChange={(e) => handleUpdateRule(rule.id, 'pattern', e.target.value)}
                                  sx={{ flex: 1, minWidth: isMobile ? '100%' : 120 }}
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
                                        return '无效正则';
                                      }
                                    })()
                                  }
                                />

                                {/* 箭头 */}
                                <Typography color="textSecondary" sx={{ display: isMobile ? 'none' : 'block' }}>
                                  →
                                </Typography>

                                {/* 替换内容 */}
                                <TextField
                                  size="small"
                                  placeholder="替换为 (留空删除)"
                                  value={rule.replacement}
                                  onChange={(e) => handleUpdateRule(rule.id, 'replacement', e.target.value)}
                                  sx={{ flex: 1, minWidth: isMobile ? '100%' : 100 }}
                                />

                                {/* 删除按钮 */}
                                <Tooltip title="删除规则">
                                  <IconButton size="small" color="error" onClick={() => handleDeleteRule(rule.id)}>
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
            <Box
              sx={{
                py: 3,
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                gap: 1,
                color: 'text.secondary'
              }}
            >
              <Typography variant="body2">暂无预处理规则</Typography>
              <Button variant="outlined" size="small" startIcon={<AddIcon />} onClick={handleAddRule}>
                添加规则
              </Button>
            </Box>
          )}

          {/* 实时预览 */}
          {hasRules && (
            <Fade in>
              <Alert variant={'standard'} severity={hasChanges ? 'success' : 'info'} sx={{ mt: 1 }}>
                <Stack spacing={0.5}>
                  <Typography variant="body2">
                    <strong>原名：</strong>
                    <code style={{ marginLeft: 4 }}>{PREVIEW_LINK_NAME}</code>
                  </Typography>
                  <Typography variant="body2">
                    <strong>结果：</strong>
                    <code
                      style={{
                        marginLeft: 4,
                        color: hasChanges ? theme.palette.success.main : 'inherit',
                        fontWeight: hasChanges ? 600 : 400
                      }}
                    >
                      {previewResult || '(空)'}
                    </code>
                  </Typography>
                </Stack>
              </Alert>
            </Fade>
          )}
        </Box>
      </Collapse>
    </Paper>
  );
}
